package auth

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
	_ "modernc.org/sqlite"
)

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) (*db.Queries, func()) {
	t.Helper()

	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "strava-mcp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open db: %v", err)
	}

	// Create schema
	schema := `
	CREATE TABLE IF NOT EXISTS auth_config (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		client_id TEXT NOT NULL,
		client_secret TEXT NOT NULL,
		access_token TEXT,
		refresh_token TEXT,
		expires_at INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := sqlDB.Exec(schema); err != nil {
		sqlDB.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create schema: %v", err)
	}

	queries := db.New(sqlDB)

	cleanup := func() {
		sqlDB.Close()
		os.RemoveAll(tmpDir)
	}

	return queries, cleanup
}

func TestNewStorage(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)
	if storage == nil {
		t.Fatal("expected non-nil storage")
	}
	if storage.queries != queries {
		t.Error("storage queries not set correctly")
	}
}

func TestSaveAndLoadClientConfig(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	// Save client config
	err := storage.SaveClientConfig("test_client_id", "test_client_secret")
	if err != nil {
		t.Fatalf("failed to save client config: %v", err)
	}

	// Load client config
	config, err := storage.LoadClientConfig()
	if err != nil {
		t.Fatalf("failed to load client config: %v", err)
	}

	if config.ClientID != "test_client_id" {
		t.Errorf("expected client ID 'test_client_id', got %q", config.ClientID)
	}
	if config.ClientSecret != "test_client_secret" {
		t.Errorf("expected client secret 'test_client_secret', got %q", config.ClientSecret)
	}
}

func TestLoadClientConfigNotFound(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	// Try to load without saving first
	_, err := storage.LoadClientConfig()
	if err == nil {
		t.Error("expected error when loading non-existent config")
	}
}

func TestSaveAndLoadTokens(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	// First save client config (required before saving tokens)
	err := storage.SaveClientConfig("test_client", "test_secret")
	if err != nil {
		t.Fatalf("failed to save client config: %v", err)
	}

	// Save tokens
	tokens := &TokenResponse{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		TokenType:    "Bearer",
	}

	err = storage.SaveTokens(tokens)
	if err != nil {
		t.Fatalf("failed to save tokens: %v", err)
	}

	// Load tokens
	loaded, err := storage.LoadTokens()
	if err != nil {
		t.Fatalf("failed to load tokens: %v", err)
	}

	if loaded.AccessToken != "test_access_token" {
		t.Errorf("expected access token 'test_access_token', got %q", loaded.AccessToken)
	}
	if loaded.RefreshToken != "test_refresh_token" {
		t.Errorf("expected refresh token 'test_refresh_token', got %q", loaded.RefreshToken)
	}
}

func TestSaveTokensWithoutClientConfig(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	// Try to save tokens without client config
	tokens := &TokenResponse{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}

	err := storage.SaveTokens(tokens)
	if err == nil {
		t.Error("expected error when saving tokens without client config")
	}
}

func TestLoadTokensNotFound(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	// Try to load without saving first
	_, err := storage.LoadTokens()
	if err == nil {
		t.Error("expected error when loading non-existent tokens")
	}
}

func TestLoadTokensNoAccessToken(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	// Save only client config (no tokens)
	err := storage.SaveClientConfig("test_client", "test_secret")
	if err != nil {
		t.Fatalf("failed to save client config: %v", err)
	}

	// Try to load tokens - should fail because no access token
	_, err = storage.LoadTokens()
	if err == nil {
		t.Error("expected error when loading config without access token")
	}
}

func TestSaveFullConfig(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	tokens := &TokenResponse{
		AccessToken:  "full_access_token",
		RefreshToken: "full_refresh_token",
		ExpiresAt:    time.Now().Add(2 * time.Hour).Unix(),
		TokenType:    "Bearer",
	}

	err := storage.SaveFullConfig("full_client_id", "full_client_secret", tokens)
	if err != nil {
		t.Fatalf("failed to save full config: %v", err)
	}

	// Verify client config
	clientConfig, err := storage.LoadClientConfig()
	if err != nil {
		t.Fatalf("failed to load client config: %v", err)
	}
	if clientConfig.ClientID != "full_client_id" {
		t.Errorf("expected client ID 'full_client_id', got %q", clientConfig.ClientID)
	}

	// Verify tokens
	loadedTokens, err := storage.LoadTokens()
	if err != nil {
		t.Fatalf("failed to load tokens: %v", err)
	}
	if loadedTokens.AccessToken != "full_access_token" {
		t.Errorf("expected access token 'full_access_token', got %q", loadedTokens.AccessToken)
	}
}

func TestDeleteTokens(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(queries)

	// Save full config
	tokens := &TokenResponse{
		AccessToken:  "delete_access_token",
		RefreshToken: "delete_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}
	err := storage.SaveFullConfig("delete_client", "delete_secret", tokens)
	if err != nil {
		t.Fatalf("failed to save full config: %v", err)
	}

	// Delete
	err = storage.DeleteTokens()
	if err != nil {
		t.Fatalf("failed to delete tokens: %v", err)
	}

	// Verify deleted
	_, err = storage.LoadTokens()
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestGetAuthConfigDirectly(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Initially should not exist
	_, err := queries.GetAuthConfig(ctx)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}

	// Save config
	err = queries.SaveAuthConfig(ctx, db.SaveAuthConfigParams{
		ClientID:     "direct_client",
		ClientSecret: "direct_secret",
		AccessToken:  sql.NullString{String: "direct_token", Valid: true},
		RefreshToken: sql.NullString{String: "direct_refresh", Valid: true},
		ExpiresAt:    sql.NullInt64{Int64: time.Now().Add(1 * time.Hour).Unix(), Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to save auth config: %v", err)
	}

	// Load directly
	config, err := queries.GetAuthConfig(ctx)
	if err != nil {
		t.Fatalf("failed to get auth config: %v", err)
	}

	if config.ClientID != "direct_client" {
		t.Errorf("expected client ID 'direct_client', got %q", config.ClientID)
	}
	if !config.AccessToken.Valid || config.AccessToken.String != "direct_token" {
		t.Error("access token not saved correctly")
	}
}

func TestUpdateTokens(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// First create the config
	err := queries.SaveAuthConfig(ctx, db.SaveAuthConfigParams{
		ClientID:     "update_client",
		ClientSecret: "update_secret",
	})
	if err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	// Update tokens
	newExpiry := time.Now().Add(2 * time.Hour).Unix()
	err = queries.UpdateTokens(ctx, db.UpdateTokensParams{
		AccessToken:  sql.NullString{String: "new_access_token", Valid: true},
		RefreshToken: sql.NullString{String: "new_refresh_token", Valid: true},
		ExpiresAt:    sql.NullInt64{Int64: newExpiry, Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to update tokens: %v", err)
	}

	// Verify update
	config, err := queries.GetAuthConfig(ctx)
	if err != nil {
		t.Fatalf("failed to get auth config: %v", err)
	}

	if config.AccessToken.String != "new_access_token" {
		t.Errorf("expected access token 'new_access_token', got %q", config.AccessToken.String)
	}
	if config.RefreshToken.String != "new_refresh_token" {
		t.Errorf("expected refresh token 'new_refresh_token', got %q", config.RefreshToken.String)
	}
	// Client credentials should be preserved
	if config.ClientID != "update_client" {
		t.Errorf("expected client ID 'update_client', got %q", config.ClientID)
	}
}

func TestDeleteAuthConfig(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create config
	err := queries.SaveAuthConfig(ctx, db.SaveAuthConfigParams{
		ClientID:     "delete_client",
		ClientSecret: "delete_secret",
	})
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify exists
	_, err = queries.GetAuthConfig(ctx)
	if err != nil {
		t.Fatalf("config should exist: %v", err)
	}

	// Delete
	err = queries.DeleteAuthConfig(ctx)
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify deleted
	_, err = queries.GetAuthConfig(ctx)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestOverwriteConfig(t *testing.T) {
	t.Parallel()

	queries, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Save initial config
	err := queries.SaveAuthConfig(ctx, db.SaveAuthConfigParams{
		ClientID:     "initial_client",
		ClientSecret: "initial_secret",
		AccessToken:  sql.NullString{String: "initial_token", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	// Overwrite with new config
	err = queries.SaveAuthConfig(ctx, db.SaveAuthConfigParams{
		ClientID:     "new_client",
		ClientSecret: "new_secret",
		AccessToken:  sql.NullString{String: "new_token", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to save new config: %v", err)
	}

	// Verify overwrite
	config, err := queries.GetAuthConfig(ctx)
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}

	if config.ClientID != "new_client" {
		t.Errorf("expected client ID 'new_client', got %q", config.ClientID)
	}
	if config.AccessToken.String != "new_token" {
		t.Errorf("expected access token 'new_token', got %q", config.AccessToken.String)
	}
}
