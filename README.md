# Strava MCP

MCP server for querying Strava activities via LLMs. Designed with AI-first principles for natural language queries.

## Features

- Syncs activities from Strava API to local SQLite database
- Background workers for automatic token refresh and activity sync
- 11 intent-based MCP tools with AI-friendly insights
- HTTP/SSE transport (LM Studio) or stdio (Claude Desktop)
- Zone data sync (requires Strava Summit subscription)

## Requirements

- Go 1.24+
- Strava API credentials
- Strava Summit subscription (optional, for zone data)

## Setup

### 1. Get Strava API Credentials

1. Go to https://www.strava.com/settings/api
2. Create an application
3. Note the Client ID and Client Secret
4. Set Authorization Callback Domain to `localhost`

### 2. Build

```bash
make build
# or
go build -o strava-mcp .
```

### 3. Run

```bash
./strava-mcp
```

On first run, you'll be prompted to enter your Strava API credentials and authenticate via browser. Credentials and tokens are stored in the SQLite database and refreshed automatically.

## CLI Options

```
Usage:
  strava-mcp [flags]

Flags:
      --db string                    path to SQLite database file (default "strava_activities.db")
      --force-reauth                 force OAuth re-authentication, clearing existing tokens
  -h, --help                         help for strava-mcp
      --no-sync                      run MCP server only without Strava API sync (offline mode)
  -p, --port int                     MCP server port (0 for stdio mode) (default 8080)
      --sync-interval duration       interval between activity syncs (default 15m0s)
      --token-refresh-interval duration   interval between token refresh checks (default 30m0s)
  -v, --verbose count                increase verbosity (-v for debug, -vv for trace with HTTP headers)
```

## MCP Client Configuration

### LM Studio (HTTP/SSE)

Run with default port:
```bash
./strava-mcp
```

Configure in LM Studio:
```json
{
  "mcpServers": {
    "strava": {
      "url": "http://127.0.0.1:8080/"
    }
  }
}
```

### Claude Desktop (stdio)

Run with stdio transport:
```bash
./strava-mcp --port 0
```

Configure in Claude Desktop (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "strava": {
      "command": "/path/to/strava-mcp",
      "args": ["--port", "0"]
    }
  }
}
```

### Offline Mode

Run without Strava API connectivity (uses existing database):
```bash
./strava-mcp --no-sync
```

## Example Questions

Ask your LLM these questions - the MCP tools will be used automatically:

### Progress & Trends
- "Am I getting faster at running?"
- "How has my pace improved over the last 3 months?"
- "What's my distance trend this year?"

### Training Load
- "Am I overtraining?"
- "How does this week compare to my average?"
- "What's my weekly volume?"

### Personal Records
- "What are my PRs for cycling?"
- "What's my longest run ever?"
- "Show me my fastest activities"

### Activity Search
- "Show me my latest activity"
- "Find my longest rides this year"
- "What did I do last week?"

### Zone Analysis (Strava Summit required)
- "How much time do I spend in Zone 2?"
- "Analyze my heart rate zones"
- "Am I training at the right intensity?"

### Comparisons
- "Compare this month to last month"
- "How does this year compare to last year?"

### Weekly Summary
- "How was my week?"
- "What did I train this week?"

## Available Tools

### Activity Search

| Tool | Description |
|------|-------------|
| `find_activities` | Unified activity search with special queries (latest/oldest/fastest/longest), filters (type/date), and sorting |

### Aggregation & Summaries

| Tool | Description |
|------|-------------|
| `count_activities` | Activity counts with optional grouping by type/month/week |
| `get_training_summary` | Comprehensive training stats (distance, duration, pace, heartrate, calories, elevation) |
| `get_week_summary` | Current/last week breakdown with activities and totals |

### Analysis & Insights

| Tool | Description |
|------|-------------|
| `compare_periods` | Side-by-side comparison of two time periods with percentage changes |
| `analyze_progress` | Trend detection - answers "Am I getting faster?" |
| `check_training_load` | Weekly volume analysis - answers "Am I overtraining?" |
| `get_personal_records` | Personal bests across categories (fastest, longest, most calories) |

### Metrics

| Tool | Description |
|------|-------------|
| `get_metrics_summary` | Detailed metrics (distance/duration/speed/heartrate/calories/cadence/elevation) |

### Zones (Requires Strava Summit)

| Tool | Description |
|------|-------------|
| `get_activity_zones` | Heart rate and power zones for a specific activity |
| `analyze_zones` | Aggregated zone statistics with 80/20 training insights |

## Tool Response Format

All tools return structured responses with:
- **Data**: The requested information
- **Insights**: AI-friendly observations about the data
- **Suggested Actions**: Recommended follow-up tool calls

## Development

```bash
make build      # Build binary
make test       # Run tests
make test-cover # Run tests with coverage
make generate   # Regenerate sqlc code
make vet        # Run go vet
make deps       # Tidy dependencies
```

### Database Migrations

Database migrations are managed with [goose](https://github.com/pressly/goose). Migration files are in `sql/migrations/` as SQL files with `-- +goose Up` and `-- +goose Down` annotations.

Migrations run automatically on startup.

## License

MIT
