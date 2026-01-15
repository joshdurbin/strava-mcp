package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/logging"
	"github.com/spf13/cobra"
)

var (
	verbosity            int
	dbPath               string
	mcpPort              int
	syncInterval         time.Duration
	tokenRefreshInterval time.Duration
	noSync               bool
	forceReauth          bool
)

var rootCmd = &cobra.Command{
	Use:   "strava-mcp",
	Short: "Strava MCP Server - expose your Strava activities via Model Context Protocol",
	Long: `Strava MCP Server syncs your Strava activities to a local SQLite database
and exposes them via the Model Context Protocol (MCP) for AI assistants.

The server runs with:
- Automatic authentication via OAuth (prompts on first run)
- Background token refresh to keep authentication valid
- Periodic activity sync from Strava
- MCP server for AI tool access

On first run, you will be prompted for your Strava API credentials.
Get these from https://www.strava.com/settings/api

Use --force-reauth to re-enter credentials and re-authenticate.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logging based on verbosity before any command runs
		logging.Setup(logging.Level(verbosity))
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create runtime config from CLI flags
		rtCfg := &RuntimeConfig{
			DBPath:               dbPath,
			MCPPort:              mcpPort,
			SyncInterval:         syncInterval,
			TokenRefreshInterval: tokenRefreshInterval,
			NoSync:               noSync,
			ForceReauth:          forceReauth,
		}

		return Run(rtCfg)
	},
}

func init() {
	// Logging verbosity
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase verbosity (-v for debug, -vv for trace with HTTP headers)")

	// Runtime settings as CLI flags
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "strava_activities.db", "path to SQLite database file")
	rootCmd.PersistentFlags().IntVarP(&mcpPort, "port", "p", 8080, "MCP server port (0 for stdio mode)")
	rootCmd.PersistentFlags().DurationVar(&syncInterval, "sync-interval", 15*time.Minute, "interval between activity syncs")
	rootCmd.PersistentFlags().DurationVar(&tokenRefreshInterval, "token-refresh-interval", 30*time.Minute, "interval between token refresh checks")

	// Offline mode
	rootCmd.PersistentFlags().BoolVar(&noSync, "no-sync", false, "run MCP server only without Strava API sync (offline mode)")

	// Force re-authentication
	rootCmd.PersistentFlags().BoolVar(&forceReauth, "force-reauth", false, "force OAuth re-authentication, clearing existing tokens")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
