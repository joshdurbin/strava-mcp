package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/auth"
	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/logging"
	"github.com/joshdurbin/strava-mcp/internal/server"
	"github.com/joshdurbin/strava-mcp/internal/strava"
	"github.com/joshdurbin/strava-mcp/internal/workers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pressly/goose/v3"
	"golang.org/x/sync/errgroup"

	_ "modernc.org/sqlite"
)

// RuntimeConfig holds all runtime configuration from CLI flags
type RuntimeConfig struct {
	DBPath               string
	MCPPort              int
	SyncInterval         time.Duration
	TokenRefreshInterval time.Duration
	NoSync               bool
	ForceReauth          bool
}

// Run is the main entry point for the unified run mode
func Run(cfg *RuntimeConfig) error {
	log := logging.Logger

	log.Info().
		Str("db_path", cfg.DBPath).
		Int("mcp_port", cfg.MCPPort).
		Bool("no_sync", cfg.NoSync).
		Dur("sync_interval", cfg.SyncInterval).
		Dur("token_refresh_interval", cfg.TokenRefreshInterval).
		Msg("starting strava-mcp")

	// Set up context for shutdown handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		cancel()
	}()

	// Open database with SQLite concurrency settings
	log.Info().Str("path", cfg.DBPath).Msg("opening database")
	sqlDB, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer sqlDB.Close()

	// Configure SQLite for concurrent access
	if err := configureSQLite(sqlDB); err != nil {
		return fmt.Errorf("configuring SQLite: %w", err)
	}

	// Check for database lock (another instance running)
	if err := checkDatabaseLock(sqlDB); err != nil {
		return err
	}

	// Run SQL migrations using goose
	gooseProvider, err := goose.NewProvider(goose.DialectSQLite3, sqlDB, os.DirFS("sql/migrations"))
	if err != nil {
		return fmt.Errorf("creating goose provider: %w", err)
	}

	results, err := gooseProvider.Up(ctx)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	for _, r := range results {
		log.Debug().Int64("version", r.Source.Version).Str("path", r.Source.Path).Msg("migration applied")
	}
	log.Debug().Int("applied", len(results)).Msg("database migrations completed")

	// Create queries and storage
	queries := db.New(sqlDB)

	// Log database statistics
	workers.LogDatabaseStats(ctx, queries)

	// Start background workers with errgroup for graceful shutdown
	g, gCtx := errgroup.WithContext(ctx)

	if !cfg.NoSync {
		storage := auth.NewStorage(queries)

		// Check and handle authentication
		accessToken, err := ensureAuthenticated(ctx, storage, cfg)
		if err != nil {
			return fmt.Errorf("authentication: %w", err)
		}

		// Use default retry config (rate limiting is handled by waiting for window resets)
		retryConfig := strava.DefaultRetryConfig()

		// Perform initial sync
		if err := workers.SyncOnce(ctx, queries, accessToken, retryConfig); err != nil {
			log.Warn().Err(err).Msg("initial sync failed")
			// Continue anyway - background worker will retry
		}

		// Log database statistics after initial sync
		workers.LogDatabaseStats(ctx, queries)

		log.Info().Msg("starting background workers")

		// Token refresh worker
		tokenRefresher := workers.NewTokenRefresher(
			storage,
			cfg.TokenRefreshInterval,
		)
		g.Go(func() error {
			tokenRefresher.Run(gCtx)
			return nil
		})

		// Activity sync worker
		activitySyncer := workers.NewActivitySyncer(
			queries,
			storage,
			cfg.SyncInterval,
			retryConfig,
		)
		g.Go(func() error {
			activitySyncer.Run(gCtx)
			return nil
		})

		// Zone sync worker (syncs heart rate and power zone data)
		zoneSyncer := workers.NewZoneSyncer(
			queries,
			storage,
			cfg.SyncInterval, // Use same interval as activity sync
			retryConfig,
		)
		g.Go(func() error {
			zoneSyncer.Run(gCtx)
			return nil
		})
	} else {
		log.Info().Msg("running in offline mode (--no-sync), skipping Strava API sync")
	}

	// Start MCP server
	srv := server.New(queries)

	var serverErr error
	if cfg.MCPPort > 0 {
		serverErr = runHTTPServer(ctx, srv.MCPServer(), cfg.MCPPort)
	} else {
		log.Info().Msg("MCP server running via stdio")
		serverErr = srv.Run(ctx)
	}

	// Wait for workers to finish (only if workers were started)
	if !cfg.NoSync {
		log.Info().Msg("waiting for workers to shut down")
		if err := g.Wait(); err != nil {
			log.Warn().Err(err).Msg("worker error during shutdown")
		} else {
			log.Info().Msg("all workers shut down gracefully")
		}
	}

	return serverErr
}

// ensureAuthenticated checks if we have valid auth tokens, and if not, runs the OAuth flow
func ensureAuthenticated(ctx context.Context, storage *auth.Storage, cfg *RuntimeConfig) (string, error) {
	log := logging.Logger

	// If force reauth is requested, clear existing tokens and credentials, then re-prompt
	if cfg.ForceReauth {
		log.Info().Msg("force re-authentication requested, clearing existing credentials and tokens")
		if err := storage.DeleteTokens(); err != nil {
			log.Debug().Err(err).Msg("failed to delete existing auth config (may not exist)")
		}
	}

	// Check if we have credentials in the database
	clientConfig, err := storage.LoadClientConfig()
	if err != nil || cfg.ForceReauth {
		// Need to prompt for credentials
		clientConfig, err = promptForCredentials()
		if err != nil {
			return "", fmt.Errorf("getting credentials: %w", err)
		}
	}

	// Try to get existing valid token (only if not force reauth)
	if !cfg.ForceReauth {
		accessToken, err := storage.GetValidAccessToken()
		if err == nil {
			log.Info().Msg("using existing authentication")
			return accessToken, nil
		}

		// Check if this was a refresh failure (token exists but refresh failed)
		if strings.Contains(err.Error(), "refreshing token") {
			log.Warn().Err(err).Msg("token refresh failed, re-authentication required")
			fmt.Println("\n=== Token Refresh Failed ===")
			fmt.Println("Your Strava authentication has expired or been revoked.")
			fmt.Println("Re-authentication is required.")
		} else {
			log.Info().Msg("no valid authentication found, starting OAuth flow")
		}
	}

	return runOAuthFlow(ctx, storage, clientConfig)
}

// promptForCredentials prompts the user to enter their Strava API credentials
func promptForCredentials() (*auth.ClientConfig, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Strava API Credentials Required ===")
	fmt.Println("Get your API credentials from: https://www.strava.com/settings/api")
	fmt.Println()

	fmt.Print("Enter your Client ID: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("reading client ID: %w", err)
	}
	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	fmt.Print("Enter your Client Secret: ")
	clientSecret, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("reading client secret: %w", err)
	}
	clientSecret = strings.TrimSpace(clientSecret)

	if clientSecret == "" {
		return nil, fmt.Errorf("client secret is required")
	}

	return &auth.ClientConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

// runOAuthFlow performs the OAuth authentication flow with Strava
func runOAuthFlow(ctx context.Context, storage *auth.Storage, clientConfig *auth.ClientConfig) (string, error) {
	log := logging.Logger

	fmt.Println("\n=== Strava Authentication Required ===")
	fmt.Println("A browser window will open for you to authorize this application.")
	fmt.Println("Press Enter to continue...")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	tokens, err := auth.Authenticate(ctx, clientConfig.ClientID, clientConfig.ClientSecret)
	if err != nil {
		return "", fmt.Errorf("OAuth flow failed: %w", err)
	}

	log.Info().
		Str("expires_at", time.Unix(tokens.ExpiresAt, 0).Format(time.RFC3339)).
		Msg("OAuth authentication successful")

	// Save tokens with client config
	if err := storage.SaveFullConfig(clientConfig.ClientID, clientConfig.ClientSecret, tokens); err != nil {
		return "", fmt.Errorf("saving tokens: %w", err)
	}

	fmt.Printf("\nAuthentication successful! Token expires: %s\n\n",
		time.Unix(tokens.ExpiresAt, 0).Format(time.RFC1123))

	return tokens.AccessToken, nil
}

// runHTTPServer runs the MCP server over HTTP/SSE
func runHTTPServer(ctx context.Context, mcpServer *mcp.Server, port int) error {
	log := logging.Logger

	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)

	addr := fmt.Sprintf(":%d", port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	errChan := make(chan error, 1)
	go func() {
		log.Info().
			Str("address", addr).
			Str("endpoint", fmt.Sprintf("http://localhost%s", addr)).
			Msg("MCP server running via HTTP/SSE")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down HTTP server")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

// configureSQLite sets up SQLite for concurrent access
func configureSQLite(sqlDB *sql.DB) error {
	log := logging.Logger

	// Enable WAL mode for better concurrency (allows concurrent reads)
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("setting WAL mode: %w", err)
	}

	// Set busy timeout to 5 seconds (wait instead of failing immediately)
	if _, err := sqlDB.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return fmt.Errorf("setting busy timeout: %w", err)
	}

	// Synchronous mode - NORMAL is safe with WAL and faster than FULL
	if _, err := sqlDB.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		return fmt.Errorf("setting synchronous mode: %w", err)
	}

	// Limit connection pool - SQLite works best with limited connections
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	log.Debug().
		Str("journal_mode", "WAL").
		Str("busy_timeout", "5000ms").
		Msg("SQLite configured")
	return nil
}

// checkDatabaseLock verifies no other process has the database locked
func checkDatabaseLock(sqlDB *sql.DB) error {
	log := logging.Logger

	// Try to acquire an exclusive lock with immediate timeout
	// This will fail if another process has the database open
	_, err := sqlDB.Exec("PRAGMA locking_mode=EXCLUSIVE")
	if err != nil {
		return fmt.Errorf("another instance may be running (database locked): %w", err)
	}

	// Try to start a transaction to actually acquire the lock
	_, err = sqlDB.Exec("BEGIN EXCLUSIVE")
	if err != nil {
		if strings.Contains(err.Error(), "locked") || strings.Contains(err.Error(), "busy") {
			return fmt.Errorf("another instance is already running (database is locked)")
		}
		return fmt.Errorf("checking database lock: %w", err)
	}

	// Commit the transaction - we've verified no other process has exclusive access
	_, err = sqlDB.Exec("COMMIT")
	if err != nil {
		return fmt.Errorf("releasing lock check: %w", err)
	}

	log.Debug().Msg("database lock check passed")
	return nil
}
