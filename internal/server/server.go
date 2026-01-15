package server

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ptr returns a pointer to the given value - useful for optional fields in structs
func ptr[T any](v T) *T {
	return &v
}

// Querier defines the interface for database queries
type Querier interface {
	GetActivity(ctx context.Context, id int64) (db.Activity, error)
	GetAllActivities(ctx context.Context) ([]db.Activity, error)
	CountActivities(ctx context.Context) (int64, error)
	CountActivitiesByType(ctx context.Context, activityType sql.NullString) (int64, error)
	CountActivitiesInRange(ctx context.Context, arg db.CountActivitiesInRangeParams) (int64, error)
	GetOldestActivity(ctx context.Context) (db.Activity, error)
	GetLatestActivity(ctx context.Context) (db.Activity, error)
	GetActivityTypeSummary(ctx context.Context) ([]db.GetActivityTypeSummaryRow, error)
	GetActivityTypeSummaryInRange(ctx context.Context, arg db.GetActivityTypeSummaryInRangeParams) ([]db.GetActivityTypeSummaryInRangeRow, error)
	CountActivitiesByTypeInRange(ctx context.Context, arg db.CountActivitiesByTypeInRangeParams) (int64, error)
	GetActivityCountsByMonth(ctx context.Context, arg db.GetActivityCountsByMonthParams) ([]db.GetActivityCountsByMonthRow, error)
	GetActivityCountsByWeek(ctx context.Context, arg db.GetActivityCountsByWeekParams) ([]db.GetActivityCountsByWeekRow, error)
	// Training summary queries
	GetTrainingSummary(ctx context.Context) (db.GetTrainingSummaryRow, error)
	GetTrainingSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetTrainingSummaryByTypeRow, error)
	GetTrainingSummaryInRange(ctx context.Context, arg db.GetTrainingSummaryInRangeParams) (db.GetTrainingSummaryInRangeRow, error)
	GetTrainingSummaryByTypeInRange(ctx context.Context, arg db.GetTrainingSummaryByTypeInRangeParams) (db.GetTrainingSummaryByTypeInRangeRow, error)
	// Period stats queries
	GetPeriodStats(ctx context.Context, arg db.GetPeriodStatsParams) (db.GetPeriodStatsRow, error)
	GetPeriodStatsByType(ctx context.Context, arg db.GetPeriodStatsByTypeParams) (db.GetPeriodStatsByTypeRow, error)
	// Zone queries
	GetActivityZones(ctx context.Context, activityID int64) ([]db.ActivityZone, error)
	GetZoneBuckets(ctx context.Context, activityZoneID int64) ([]db.ZoneBucket, error)
	GetActivitiesWithZones(ctx context.Context, limit int64) ([]db.GetActivitiesWithZonesRow, error)
	CountActivitiesWithZones(ctx context.Context) (int64, error)
	CountActivitiesWithoutZones(ctx context.Context) (int64, error)
	GetHeartRateZoneSummary(ctx context.Context) ([]db.GetHeartRateZoneSummaryRow, error)
	GetHeartRateZoneSummaryByType(ctx context.Context, activityType sql.NullString) ([]db.GetHeartRateZoneSummaryByTypeRow, error)
	GetHeartRateZoneSummaryInRange(ctx context.Context, arg db.GetHeartRateZoneSummaryInRangeParams) ([]db.GetHeartRateZoneSummaryInRangeRow, error)
	GetHeartRateZoneSummaryByTypeInRange(ctx context.Context, arg db.GetHeartRateZoneSummaryByTypeInRangeParams) ([]db.GetHeartRateZoneSummaryByTypeInRangeRow, error)
	GetPowerZoneSummary(ctx context.Context) ([]db.GetPowerZoneSummaryRow, error)
	GetPowerZoneSummaryByType(ctx context.Context, activityType sql.NullString) ([]db.GetPowerZoneSummaryByTypeRow, error)
	GetPowerZoneSummaryInRange(ctx context.Context, arg db.GetPowerZoneSummaryInRangeParams) ([]db.GetPowerZoneSummaryInRangeRow, error)
	// Personal records queries
	GetFastestActivity(ctx context.Context) (db.Activity, error)
	GetFastestActivityByType(ctx context.Context, activityType sql.NullString) (db.Activity, error)
	GetLongestDistanceActivity(ctx context.Context) (db.Activity, error)
	GetLongestDistanceActivityByType(ctx context.Context, activityType sql.NullString) (db.Activity, error)
	GetLongestDurationActivity(ctx context.Context) (db.Activity, error)
	GetLongestDurationActivityByType(ctx context.Context, activityType sql.NullString) (db.Activity, error)
	GetHighestElevationActivity(ctx context.Context) (db.Activity, error)
	GetHighestElevationActivityByType(ctx context.Context, activityType sql.NullString) (db.Activity, error)
	GetMostCaloriesActivity(ctx context.Context) (db.Activity, error)
	GetMostCaloriesActivityByType(ctx context.Context, activityType sql.NullString) (db.Activity, error)
	// Weekly volume queries
	GetWeeklyVolume(ctx context.Context, arg db.GetWeeklyVolumeParams) ([]db.GetWeeklyVolumeRow, error)
	GetWeeklyVolumeByType(ctx context.Context, arg db.GetWeeklyVolumeByTypeParams) ([]db.GetWeeklyVolumeByTypeRow, error)
	// Search queries
	SearchActivities(ctx context.Context, arg db.SearchActivitiesParams) ([]db.Activity, error)
	SearchActivitiesByDistance(ctx context.Context, arg db.SearchActivitiesByDistanceParams) ([]db.Activity, error)
	SearchActivitiesByDuration(ctx context.Context, arg db.SearchActivitiesByDurationParams) ([]db.Activity, error)
	SearchActivitiesBySpeed(ctx context.Context, arg db.SearchActivitiesBySpeedParams) ([]db.Activity, error)
	SearchActivitiesByElevation(ctx context.Context, arg db.SearchActivitiesByElevationParams) ([]db.Activity, error)
}

// Server wraps the MCP server and database queries
type Server struct {
	mcp     *mcp.Server
	queries Querier
}

// MCPServer returns the underlying MCP server (for use with HTTP/SSE transport)
func (s *Server) MCPServer() *mcp.Server {
	return s.mcp
}

// New creates a new MCP server with activity query tools
func New(queries Querier) *Server {
	logging.Info("MCP server initializing", "name", "strava-mcp", "version", "1.0.0")

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "strava-mcp",
		Version: "1.0.0",
	}, nil)

	s := &Server{
		mcp:     mcpServer,
		queries: queries,
	}

	logging.Debug("Registering MCP tools")
	s.registerTools()
	s.registerMetricsTools()
	s.registerZoneTools()
	s.registerProgressTools()
	s.registerRecordsTools()

	logging.Debug("Registering MCP resources")
	s.registerResources()

	logging.Debug("Registering MCP prompts")
	s.registerPrompts()

	logging.Info("MCP server initialized", "tools_registered", 11, "resources_registered", 4, "prompts_registered", 4)
	return s
}

// Run starts the MCP server over stdio transport
func (s *Server) Run(ctx context.Context) error {
	logging.Info("MCP server starting")
	defer logging.Info("MCP server stopped")
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	// Consolidated activity search tool
	logging.Debug("Registering tool", "name", "find_activities")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "find_activities",
		Description: `Search and retrieve Strava activities with flexible filtering and sorting options.

Use when:
- User asks "Show me my latest activity" or "What did I do last week?"
- User wants to find specific activities by type, date, or performance
- User needs activity details by ID

Parameters:
- query (string): Special queries: "latest", "oldest", "fastest", "best", "longest". Overrides other filters.
- id (integer): Get a specific activity by its Strava ID.
- type (string): Filter by activity type (Run, Ride, Swim, Walk, Hike, etc.).
- start_date (string): Start date in YYYY-MM-DD format.
- end_date (string): End date in YYYY-MM-DD format.
- sort_by (string): Sort results by "date", "distance", "duration", "pace", or "elevation". Default: "date".
- limit (integer): Number of activities to return. Default: 20, Max: 100.

Returns: List of activities with id, name, type, date, distance, duration, pace, elevation, heartrate, and calories.

Example: {"query": "latest"} or {"type": "Run", "start_date": "2024-01-01", "sort_by": "distance", "limit": 10}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Find Activities",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.findActivities)

	// Consolidated count activities tool
	logging.Debug("Registering tool", "name", "count_activities")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "count_activities",
		Description: `Count activities with optional filtering and grouping by type, month, or week.

Use when:
- User asks "How many runs have I done?" or "Activity count by month"
- User wants to see training frequency over time
- User needs activity breakdown by type

Parameters:
- type (string): Filter by activity type (Run, Ride, Swim, etc.). Leave empty for all types.
- start_date (string): Start date in YYYY-MM-DD format. Leave empty for all time.
- end_date (string): End date in YYYY-MM-DD format. Leave empty for all time.
- group_by (string): Group results by "type", "month", or "week". Omit for total count only.

Returns: Total count and/or grouped counts with activity type breakdown.

Example: {"group_by": "type"} or {"type": "Run", "start_date": "2024-01-01", "group_by": "month"}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Count Activities",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.countActivitiesConsolidated)

	// Training summary tool
	logging.Debug("Registering tool", "name", "get_training_summary")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_training_summary",
		Description: `Get comprehensive training statistics including distance, duration, pace, heartrate, calories, and elevation.

Use when:
- User asks "What's my total distance?" or "Training overview"
- User wants aggregate stats for all activities or a specific type/period
- User needs to understand their overall training volume

Parameters:
- type (string): Filter by activity type (Run, Ride, Swim, etc.). Leave empty for all types.
- start_date (string): Start date in YYYY-MM-DD format. Leave empty for all time.
- end_date (string): End date in YYYY-MM-DD format. Leave empty for all time.

Returns: Activity count, total distance, total duration, average pace, average heartrate, total calories, and total elevation with insights.

Example: {"type": "Run"} or {"start_date": "2024-01-01", "end_date": "2024-12-31"}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Get Training Summary",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.getTrainingSummary)

	// Compare periods tool
	logging.Debug("Registering tool", "name", "compare_periods")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "compare_periods",
		Description: `Compare training metrics between two time periods with percentage changes and insights.

Use when:
- User asks "Compare this month to last month" or "How does this year compare to last year?"
- User wants to track progress over time
- User needs to see if training volume has increased or decreased

Parameters:
- period1_start (string, required): Start date of first period in YYYY-MM-DD format.
- period1_end (string, required): End date of first period in YYYY-MM-DD format.
- period2_start (string, required): Start date of second period in YYYY-MM-DD format.
- period2_end (string, required): End date of second period in YYYY-MM-DD format.
- type (string): Filter by activity type (Run, Ride, Swim, etc.). Leave empty for all types.

Returns: Side-by-side comparison with activity count, distance, duration, pace for each period plus percentage changes and insights.

Example: {"period1_start": "2024-01-01", "period1_end": "2024-01-31", "period2_start": "2024-02-01", "period2_end": "2024-02-29"}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Compare Periods",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.comparePeriods)

	// Week summary tool
	logging.Debug("Registering tool", "name", "get_week_summary")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_week_summary",
		Description: `Get a detailed breakdown of training for a specific week including all activities and totals.

Use when:
- User asks "How was my week?" or "This week's training"
- User wants to review recent training
- User needs weekly volume breakdown

Parameters:
- week (string): Which week to analyze: "current", "last", or ISO week format "YYYY-Www" (e.g., "2024-W03"). Default: "current".
- type (string): Filter by activity type (Run, Ride, Swim, etc.). Leave empty for all types.

Returns: Week label, date range, activity count, total distance, total duration, total calories, total elevation, list of activities, and comparison to average week.

Example: {"week": "current"} or {"week": "last", "type": "Run"}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Get Week Summary",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.getWeekSummary)
}

// Tool input/output types

// Default and max limits for activity queries
const (
	defaultActivityLimit = 20
	maxActivityLimit     = 100
)

// FindActivitiesInput - consolidated input for activity search
type FindActivitiesInput struct {
	// Special queries (mutually exclusive with filters)
	Query string `json:"query,omitempty" jsonschema:"Special query shortcuts: 'latest' (most recent), 'oldest' (first ever), 'fastest' or 'best' (highest pace), 'longest' (greatest distance). When set, overrides other filter parameters."`
	ID    int64  `json:"id,omitempty" jsonschema:"Get a specific activity by its unique Strava activity ID. When set, overrides other parameters."`

	// Filters
	Type      string `json:"type,omitempty" jsonschema:"Filter by activity type. Common values: Run, Ride, Swim, Walk, Hike, VirtualRide, WeightTraining, Yoga."`
	StartDate string `json:"start_date,omitempty" jsonschema:"Include activities on or after this date. Format: YYYY-MM-DD (e.g., 2024-01-15)."`
	EndDate   string `json:"end_date,omitempty" jsonschema:"Include activities on or before this date. Format: YYYY-MM-DD (e.g., 2024-12-31)."`

	// Sorting and pagination
	SortBy string `json:"sort_by,omitempty" jsonschema:"Sort results by this field. Valid values: date (newest first), distance (longest first), duration (longest first), pace (fastest first), elevation (most climbing first). Default: date."`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of activities to return. Default: 20, Maximum: 100."`
}

// FindActivitiesOutput - output for activity search
type FindActivitiesOutput struct {
	Query            string            `json:"query,omitempty"`
	Activities       []ActivitySummary `json:"activities"`
	TotalMatching    int               `json:"total_matching,omitempty"`
	Insights         []Insight         `json:"insights,omitempty"`
	SuggestedActions []SuggestedAction `json:"suggested_actions,omitempty"`
}

// CountActivitiesConsolidatedInput - input for counting activities with optional grouping
type CountActivitiesConsolidatedInput struct {
	Type      string `json:"type,omitempty" jsonschema:"Filter by activity type. Common values: Run, Ride, Swim, Walk, Hike. Leave empty to count all activity types."`
	StartDate string `json:"start_date,omitempty" jsonschema:"Count activities on or after this date. Format: YYYY-MM-DD. Leave empty to include all historical activities."`
	EndDate   string `json:"end_date,omitempty" jsonschema:"Count activities on or before this date. Format: YYYY-MM-DD. Leave empty to include up to today."`
	GroupBy   string `json:"group_by,omitempty" jsonschema:"How to group results. Valid values: 'type' (by activity type), 'month' (by calendar month), 'week' (by calendar week). Omit for single total count."`
}

// Consolidated count activities output
type CountActivitiesConsolidatedOutput struct {
	Count    int64           `json:"count,omitempty"`
	ByType   []TypeCount     `json:"by_type,omitempty"`
	ByPeriod []PeriodSummary `json:"by_period,omitempty"`
	Filter   string          `json:"filter,omitempty"`
}

type PeriodSummary struct {
	Period string           `json:"period"`
	Total  int64            `json:"total"`
	ByType map[string]int64 `json:"by_type,omitempty"`
}

// TrainingSummaryInput - input for retrieving training summary statistics
type TrainingSummaryInput struct {
	Type      string `json:"type,omitempty" jsonschema:"Filter statistics by activity type. Common values: Run, Ride, Swim, Walk, Hike. Leave empty for summary across all activity types."`
	StartDate string `json:"start_date,omitempty" jsonschema:"Include activities on or after this date. Format: YYYY-MM-DD. Leave empty for all-time statistics."`
	EndDate   string `json:"end_date,omitempty" jsonschema:"Include activities on or before this date. Format: YYYY-MM-DD. Leave empty to include up to today."`
}

type TrainingSummaryOutput struct {
	ActivityCount    int64             `json:"activity_count"`
	TotalDistance    string            `json:"total_distance"`
	TotalDuration    string            `json:"total_duration"`
	AvgPace          string            `json:"avg_pace,omitempty"`
	AvgHeartrate     int               `json:"avg_heartrate,omitempty"`
	TotalCalories    int               `json:"total_calories,omitempty"`
	TotalElevation   string            `json:"total_elevation,omitempty"`
	Filter           string            `json:"filter,omitempty"`
	Insights         []Insight         `json:"insights,omitempty"`
	SuggestedActions []SuggestedAction `json:"suggested_actions,omitempty"`
}

// ComparePeriodInput - input for comparing two time periods (all date fields required)
type ComparePeriodInput struct {
	Period1Start string `json:"period1_start" jsonschema:"Start date of the first comparison period. Format: YYYY-MM-DD (e.g., 2024-01-01). Required."`
	Period1End   string `json:"period1_end" jsonschema:"End date of the first comparison period. Format: YYYY-MM-DD (e.g., 2024-01-31). Required."`
	Period2Start string `json:"period2_start" jsonschema:"Start date of the second comparison period. Format: YYYY-MM-DD (e.g., 2024-02-01). Required."`
	Period2End   string `json:"period2_end" jsonschema:"End date of the second comparison period. Format: YYYY-MM-DD (e.g., 2024-02-29). Required."`
	Type         string `json:"type,omitempty" jsonschema:"Filter comparison to a specific activity type. Common values: Run, Ride, Swim. Leave empty to compare all activities."`
}

type ComparePeriodOutput struct {
	Period1          PeriodStats       `json:"period1"`
	Period2          PeriodStats       `json:"period2"`
	Change           ChangeStats       `json:"change"`
	Insights         []Insight         `json:"insights,omitempty"`
	SuggestedActions []SuggestedAction `json:"suggested_actions,omitempty"`
}

type PeriodStats struct {
	DateRange     string `json:"date_range"`
	ActivityCount int64  `json:"activity_count"`
	TotalDistance string `json:"total_distance"`
	TotalDuration string `json:"total_duration"`
	AvgPace       string `json:"avg_pace,omitempty"`
}

type ChangeStats struct {
	ActivityCount string `json:"activity_count"`
	Distance      string `json:"distance"`
	Duration      string `json:"duration"`
}

// WeekSummaryInput - input for retrieving weekly training summary
type WeekSummaryInput struct {
	Week string `json:"week,omitempty" jsonschema:"Which week to summarize. Valid values: 'current' (this week), 'last' (previous week), or ISO week format 'YYYY-Www' (e.g., '2024-W03' for the 3rd week of 2024). Default: current."`
	Type string `json:"type,omitempty" jsonschema:"Filter summary to a specific activity type. Common values: Run, Ride, Swim, Walk, Hike. Leave empty to include all activity types."`
}

type WeekSummaryOutput struct {
	Week             string            `json:"week"`
	DateRange        string            `json:"date_range"`
	ActivityCount    int64             `json:"activity_count"`
	TotalDistance    string            `json:"total_distance"`
	TotalDuration    string            `json:"total_duration"`
	TotalCalories    int               `json:"total_calories,omitempty"`
	TotalElevation   string            `json:"total_elevation,omitempty"`
	Activities       []ActivitySummary `json:"activities,omitempty"`
	ComparedToAvg    string            `json:"compared_to_avg,omitempty"`
	Insights         []Insight         `json:"insights,omitempty"`
	SuggestedActions []SuggestedAction `json:"suggested_actions,omitempty"`
}

type ActivitySummary struct {
	ID            int64  `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	Date          string `json:"date,omitempty"`
	Distance      string `json:"distance,omitempty"`
	Duration      string `json:"duration,omitempty"`
	Pace          string `json:"pace,omitempty"`
	ElevationGain string `json:"elevation_gain,omitempty"`
	AvgHeartrate  int    `json:"avg_heartrate_bpm,omitempty"`
	MaxHeartrate  int    `json:"max_heartrate_bpm,omitempty"`
	Calories      int    `json:"calories,omitempty"`
}

type TypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// Tool handlers

// findActivities - consolidated activity search handler
func (s *Server) findActivities(ctx context.Context, req *mcp.CallToolRequest, input FindActivitiesInput) (*mcp.CallToolResult, FindActivitiesOutput, error) {
	logging.Info("MCP tool call", "tool", "find_activities", "query", input.Query, "id", input.ID, "type", input.Type, "sort_by", input.SortBy)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "find_activities", "input", logging.ToJSON(input))
	}

	limit := applyLimit(input.Limit)
	output := FindActivitiesOutput{
		Activities: []ActivitySummary{},
	}

	// Handle special queries first
	if input.ID > 0 {
		activity, err := s.queries.GetActivity(ctx, input.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, output, nil
			}
			return nil, FindActivitiesOutput{}, fmt.Errorf("querying activity: %w", err)
		}
		output.Query = fmt.Sprintf("id=%d", input.ID)
		output.Activities = []ActivitySummary{convertActivity(activity)}
		output.TotalMatching = 1
		output.SuggestedActions = SuggestNextActions("activities")
		return nil, output, nil
	}

	if input.Query != "" {
		switch input.Query {
		case "latest":
			activity, err := s.queries.GetLatestActivity(ctx)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, output, nil
				}
				return nil, FindActivitiesOutput{}, fmt.Errorf("querying latest activity: %w", err)
			}
			output.Query = "latest"
			output.Activities = []ActivitySummary{convertActivity(activity)}
			output.TotalMatching = 1
			output.SuggestedActions = SuggestNextActions("activities")
			return nil, output, nil

		case "oldest":
			activity, err := s.queries.GetOldestActivity(ctx)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, output, nil
				}
				return nil, FindActivitiesOutput{}, fmt.Errorf("querying oldest activity: %w", err)
			}
			output.Query = "oldest"
			output.Activities = []ActivitySummary{convertActivity(activity)}
			output.TotalMatching = 1
			output.SuggestedActions = SuggestNextActions("activities")
			return nil, output, nil

		case "fastest", "best":
			var activity db.Activity
			var err error
			if input.Type != "" {
				activity, err = s.queries.GetFastestActivityByType(ctx, sql.NullString{String: input.Type, Valid: true})
			} else {
				activity, err = s.queries.GetFastestActivity(ctx)
			}
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, output, nil
				}
				return nil, FindActivitiesOutput{}, fmt.Errorf("querying fastest activity: %w", err)
			}
			output.Query = "fastest"
			output.Activities = []ActivitySummary{convertActivity(activity)}
			output.TotalMatching = 1
			output.SuggestedActions = SuggestNextActions("activities")
			return nil, output, nil

		case "longest":
			var activity db.Activity
			var err error
			if input.Type != "" {
				activity, err = s.queries.GetLongestDistanceActivityByType(ctx, sql.NullString{String: input.Type, Valid: true})
			} else {
				activity, err = s.queries.GetLongestDistanceActivity(ctx)
			}
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, output, nil
				}
				return nil, FindActivitiesOutput{}, fmt.Errorf("querying longest activity: %w", err)
			}
			output.Query = "longest"
			output.Activities = []ActivitySummary{convertActivity(activity)}
			output.TotalMatching = 1
			output.SuggestedActions = SuggestNextActions("activities")
			return nil, output, nil
		}
	}

	// Regular filtered/sorted search
	var activities []db.Activity
	var err error

	// Determine sort order
	sortBy := input.SortBy
	if sortBy == "" {
		sortBy = "date"
	}

	// Parse date filters
	hasType := input.Type != ""
	hasDateRange := input.StartDate != "" || input.EndDate != ""

	var startTime, endTime sql.NullTime
	if hasDateRange {
		startTime, endTime, err = parseServerDateRange(input.StartDate, input.EndDate)
		if err != nil {
			return nil, FindActivitiesOutput{}, err
		}
	}

	// Build query based on sort and filters
	switch sortBy {
	case "distance":
		activities, err = s.queries.SearchActivitiesByDistance(ctx, db.SearchActivitiesByDistanceParams{
			Column1:   sql.NullString{String: input.Type, Valid: hasType},
			Type:      sql.NullString{String: input.Type, Valid: hasType},
			Column3:   startTime,
			StartDate: startTime,
			Column5:   endTime,
			StartDate_2: endTime,
			Limit:     int64(limit),
		})
	case "duration":
		activities, err = s.queries.SearchActivitiesByDuration(ctx, db.SearchActivitiesByDurationParams{
			Column1:   sql.NullString{String: input.Type, Valid: hasType},
			Type:      sql.NullString{String: input.Type, Valid: hasType},
			Column3:   startTime,
			StartDate: startTime,
			Column5:   endTime,
			StartDate_2: endTime,
			Limit:     int64(limit),
		})
	case "pace", "speed":
		activities, err = s.queries.SearchActivitiesBySpeed(ctx, db.SearchActivitiesBySpeedParams{
			Column1:   sql.NullString{String: input.Type, Valid: hasType},
			Type:      sql.NullString{String: input.Type, Valid: hasType},
			Column3:   startTime,
			StartDate: startTime,
			Column5:   endTime,
			StartDate_2: endTime,
			Limit:     int64(limit),
		})
	case "elevation":
		activities, err = s.queries.SearchActivitiesByElevation(ctx, db.SearchActivitiesByElevationParams{
			Column1:   sql.NullString{String: input.Type, Valid: hasType},
			Type:      sql.NullString{String: input.Type, Valid: hasType},
			Column3:   startTime,
			StartDate: startTime,
			Column5:   endTime,
			StartDate_2: endTime,
			Limit:     int64(limit),
		})
	default: // date
		activities, err = s.queries.SearchActivities(ctx, db.SearchActivitiesParams{
			Column1:   sql.NullString{String: input.Type, Valid: hasType},
			Type:      sql.NullString{String: input.Type, Valid: hasType},
			Column3:   startTime,
			StartDate: startTime,
			Column5:   endTime,
			StartDate_2: endTime,
			Limit:     int64(limit),
		})
	}

	if err != nil {
		return nil, FindActivitiesOutput{}, fmt.Errorf("searching activities: %w", err)
	}

	output.Activities = convertActivities(activities)
	output.TotalMatching = len(activities)
	output.SuggestedActions = SuggestNextActions("activities")

	logging.Info("MCP tool completed", "tool", "find_activities", "returned", len(output.Activities))
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "find_activities", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Consolidated count activities handler
func (s *Server) countActivitiesConsolidated(ctx context.Context, req *mcp.CallToolRequest, input CountActivitiesConsolidatedInput) (*mcp.CallToolResult, CountActivitiesConsolidatedOutput, error) {
	logging.Info("MCP tool call", "tool", "count_activities", "type", input.Type, "group_by", input.GroupBy, "start", input.StartDate, "end", input.EndDate)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "count_activities", "input", logging.ToJSON(input))
	}

	output := CountActivitiesConsolidatedOutput{
		Filter: buildFilterDesc(input.Type, input.StartDate, input.EndDate),
	}

	hasType := input.Type != ""
	hasDateRange := input.StartDate != "" || input.EndDate != ""

	switch input.GroupBy {
	case "type":
		// Group by activity type
		if hasDateRange {
			start, end, err := parseServerDateRange(input.StartDate, input.EndDate)
			if err != nil {
				return nil, CountActivitiesConsolidatedOutput{}, err
			}
			rows, err := s.queries.GetActivityTypeSummaryInRange(ctx, db.GetActivityTypeSummaryInRangeParams{
				StartDate:   start,
				StartDate_2: end,
			})
			if err != nil {
				return nil, CountActivitiesConsolidatedOutput{}, err
			}
			var total int64
			for _, row := range rows {
				typeName := ""
				if row.Type.Valid {
					typeName = row.Type.String
				}
				output.ByType = append(output.ByType, TypeCount{Type: typeName, Count: row.Count})
				total += row.Count
			}
			output.Count = total
		} else {
			rows, err := s.queries.GetActivityTypeSummary(ctx)
			if err != nil {
				return nil, CountActivitiesConsolidatedOutput{}, err
			}
			var total int64
			for _, row := range rows {
				typeName := ""
				if row.Type.Valid {
					typeName = row.Type.String
				}
				output.ByType = append(output.ByType, TypeCount{Type: typeName, Count: row.Count})
				total += row.Count
			}
			output.Count = total
		}

	case "month":
		// Group by month
		start, end, err := parseServerDateRange(input.StartDate, input.EndDate)
		if err != nil {
			return nil, CountActivitiesConsolidatedOutput{}, err
		}
		rows, err := s.queries.GetActivityCountsByMonth(ctx, db.GetActivityCountsByMonthParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, CountActivitiesConsolidatedOutput{}, err
		}
		output.ByPeriod = aggregatePeriodCounts(rows, func(r db.GetActivityCountsByMonthRow) (string, string, int64) {
			month := ""
			if r.Month != nil {
				month = fmt.Sprintf("%v", r.Month)
			}
			typeName := ""
			if r.Type.Valid {
				typeName = r.Type.String
			}
			return month, typeName, r.Count
		})
		// Calculate total
		var total int64
		for _, p := range output.ByPeriod {
			total += p.Total
		}
		output.Count = total

	case "week":
		// Group by week
		start, end, err := parseServerDateRange(input.StartDate, input.EndDate)
		if err != nil {
			return nil, CountActivitiesConsolidatedOutput{}, err
		}
		rows, err := s.queries.GetActivityCountsByWeek(ctx, db.GetActivityCountsByWeekParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, CountActivitiesConsolidatedOutput{}, err
		}
		output.ByPeriod = aggregatePeriodCountsWeek(rows, func(r db.GetActivityCountsByWeekRow) (string, string, int64) {
			week := ""
			if r.Week != nil {
				week = fmt.Sprintf("%v", r.Week)
			}
			typeName := ""
			if r.Type.Valid {
				typeName = r.Type.String
			}
			return week, typeName, r.Count
		})
		// Calculate total
		var total int64
		for _, p := range output.ByPeriod {
			total += p.Total
		}
		output.Count = total

	default:
		// Simple count
		var count int64
		var err error
		if hasType && hasDateRange {
			start, end, parseErr := parseServerDateRange(input.StartDate, input.EndDate)
			if parseErr != nil {
				return nil, CountActivitiesConsolidatedOutput{}, parseErr
			}
			count, err = s.queries.CountActivitiesByTypeInRange(ctx, db.CountActivitiesByTypeInRangeParams{
				Type:        sql.NullString{String: input.Type, Valid: true},
				StartDate:   start,
				StartDate_2: end,
			})
		} else if hasType {
			count, err = s.queries.CountActivitiesByType(ctx, sql.NullString{String: input.Type, Valid: true})
		} else if hasDateRange {
			start, end, parseErr := parseServerDateRange(input.StartDate, input.EndDate)
			if parseErr != nil {
				return nil, CountActivitiesConsolidatedOutput{}, parseErr
			}
			count, err = s.queries.CountActivitiesInRange(ctx, db.CountActivitiesInRangeParams{
				StartDate:   start,
				StartDate_2: end,
			})
		} else {
			count, err = s.queries.CountActivities(ctx)
		}
		if err != nil {
			return nil, CountActivitiesConsolidatedOutput{}, err
		}
		output.Count = count
	}

	logging.Info("MCP tool completed", "tool", "count_activities", "count", output.Count)
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "count_activities", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Training summary handler
func (s *Server) getTrainingSummary(ctx context.Context, req *mcp.CallToolRequest, input TrainingSummaryInput) (*mcp.CallToolResult, TrainingSummaryOutput, error) {
	logging.Info("MCP tool call", "tool", "get_training_summary", "type", input.Type, "start", input.StartDate, "end", input.EndDate)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "get_training_summary", "input", logging.ToJSON(input))
	}

	output := TrainingSummaryOutput{
		Filter: buildFilterDesc(input.Type, input.StartDate, input.EndDate),
	}

	hasType := input.Type != ""
	hasDateRange := input.StartDate != "" || input.EndDate != ""

	var row trainingSummaryData

	if hasType && hasDateRange {
		start, end, err := parseServerDateRange(input.StartDate, input.EndDate)
		if err != nil {
			return nil, TrainingSummaryOutput{}, err
		}
		dbRow, err := s.queries.GetTrainingSummaryByTypeInRange(ctx, db.GetTrainingSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: input.Type, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, TrainingSummaryOutput{}, err
		}
		row = trainingSummaryData{
			ActivityCount:   dbRow.ActivityCount,
			TotalDistance:   toFloat64(dbRow.TotalDistance),
			TotalMovingTime: toInt64(dbRow.TotalMovingTime),
			AvgSpeed:        toFloat64(dbRow.AvgSpeed),
			AvgHeartrate:    toFloat64(dbRow.AvgHeartrate),
			TotalCalories:   toFloat64(dbRow.TotalCalories),
			TotalElevation:  toFloat64(dbRow.TotalElevation),
		}
	} else if hasType {
		dbRow, err := s.queries.GetTrainingSummaryByType(ctx, sql.NullString{String: input.Type, Valid: true})
		if err != nil {
			return nil, TrainingSummaryOutput{}, err
		}
		row = trainingSummaryData{
			ActivityCount:   dbRow.ActivityCount,
			TotalDistance:   toFloat64(dbRow.TotalDistance),
			TotalMovingTime: toInt64(dbRow.TotalMovingTime),
			AvgSpeed:        toFloat64(dbRow.AvgSpeed),
			AvgHeartrate:    toFloat64(dbRow.AvgHeartrate),
			TotalCalories:   toFloat64(dbRow.TotalCalories),
			TotalElevation:  toFloat64(dbRow.TotalElevation),
		}
	} else if hasDateRange {
		start, end, err := parseServerDateRange(input.StartDate, input.EndDate)
		if err != nil {
			return nil, TrainingSummaryOutput{}, err
		}
		dbRow, err := s.queries.GetTrainingSummaryInRange(ctx, db.GetTrainingSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, TrainingSummaryOutput{}, err
		}
		row = trainingSummaryData{
			ActivityCount:   dbRow.ActivityCount,
			TotalDistance:   toFloat64(dbRow.TotalDistance),
			TotalMovingTime: toInt64(dbRow.TotalMovingTime),
			AvgSpeed:        toFloat64(dbRow.AvgSpeed),
			AvgHeartrate:    toFloat64(dbRow.AvgHeartrate),
			TotalCalories:   toFloat64(dbRow.TotalCalories),
			TotalElevation:  toFloat64(dbRow.TotalElevation),
		}
	} else {
		dbRow, err := s.queries.GetTrainingSummary(ctx)
		if err != nil {
			return nil, TrainingSummaryOutput{}, err
		}
		row = trainingSummaryData{
			ActivityCount:   dbRow.ActivityCount,
			TotalDistance:   toFloat64(dbRow.TotalDistance),
			TotalMovingTime: toInt64(dbRow.TotalMovingTime),
			AvgSpeed:        toFloat64(dbRow.AvgSpeed),
			AvgHeartrate:    toFloat64(dbRow.AvgHeartrate),
			TotalCalories:   toFloat64(dbRow.TotalCalories),
			TotalElevation:  toFloat64(dbRow.TotalElevation),
		}
	}

	output.ActivityCount = row.ActivityCount
	output.TotalDistance = formatDistance(row.TotalDistance)
	output.TotalDuration = formatDuration(row.TotalMovingTime)
	if row.AvgSpeed > 0 {
		output.AvgPace = formatPace(row.AvgSpeed)
	}
	if row.AvgHeartrate > 0 {
		output.AvgHeartrate = int(row.AvgHeartrate)
	}
	if row.TotalCalories > 0 {
		output.TotalCalories = int(row.TotalCalories)
	}
	if row.TotalElevation > 0 {
		output.TotalElevation = fmt.Sprintf("%.0fm", row.TotalElevation)
	}

	// Add insights and suggested actions
	output.Insights = []Insight{}
	if row.ActivityCount > 0 {
		avgDist := row.TotalDistance / float64(row.ActivityCount)
		output.Insights = append(output.Insights, Insight{
			Type:    "trend",
			Message: fmt.Sprintf("Average %.1f km per activity", avgDist/1000),
		})
	}
	output.SuggestedActions = SuggestNextActions("training_summary")

	logging.Info("MCP tool completed", "tool", "get_training_summary", "activity_count", output.ActivityCount)
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "get_training_summary", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Compare periods handler
func (s *Server) comparePeriods(ctx context.Context, req *mcp.CallToolRequest, input ComparePeriodInput) (*mcp.CallToolResult, ComparePeriodOutput, error) {
	logging.Info("MCP tool call", "tool", "compare_periods", "p1_start", input.Period1Start, "p1_end", input.Period1End, "p2_start", input.Period2Start, "p2_end", input.Period2End, "type", input.Type)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "compare_periods", "input", logging.ToJSON(input))
	}

	hasType := input.Type != ""

	// Fetch period 1 stats
	p1Start, p1End, err := parseServerDateRange(input.Period1Start, input.Period1End)
	if err != nil {
		return nil, ComparePeriodOutput{}, fmt.Errorf("period 1: %w", err)
	}

	var p1Stats db.GetPeriodStatsRow
	if hasType {
		row, err := s.queries.GetPeriodStatsByType(ctx, db.GetPeriodStatsByTypeParams{
			Type:        sql.NullString{String: input.Type, Valid: true},
			StartDate:   p1Start,
			StartDate_2: p1End,
		})
		if err != nil {
			return nil, ComparePeriodOutput{}, fmt.Errorf("fetching period 1 stats: %w", err)
		}
		p1Stats = db.GetPeriodStatsRow{
			ActivityCount:   row.ActivityCount,
			TotalDistance:   row.TotalDistance,
			TotalMovingTime: row.TotalMovingTime,
			AvgSpeed:        row.AvgSpeed,
		}
	} else {
		p1Stats, err = s.queries.GetPeriodStats(ctx, db.GetPeriodStatsParams{
			StartDate:   p1Start,
			StartDate_2: p1End,
		})
		if err != nil {
			return nil, ComparePeriodOutput{}, fmt.Errorf("fetching period 1 stats: %w", err)
		}
	}

	// Fetch period 2 stats
	p2Start, p2End, err := parseServerDateRange(input.Period2Start, input.Period2End)
	if err != nil {
		return nil, ComparePeriodOutput{}, fmt.Errorf("period 2: %w", err)
	}

	var p2Stats db.GetPeriodStatsRow
	if hasType {
		row, err := s.queries.GetPeriodStatsByType(ctx, db.GetPeriodStatsByTypeParams{
			Type:        sql.NullString{String: input.Type, Valid: true},
			StartDate:   p2Start,
			StartDate_2: p2End,
		})
		if err != nil {
			return nil, ComparePeriodOutput{}, fmt.Errorf("fetching period 2 stats: %w", err)
		}
		p2Stats = db.GetPeriodStatsRow{
			ActivityCount:   row.ActivityCount,
			TotalDistance:   row.TotalDistance,
			TotalMovingTime: row.TotalMovingTime,
			AvgSpeed:        row.AvgSpeed,
		}
	} else {
		p2Stats, err = s.queries.GetPeriodStats(ctx, db.GetPeriodStatsParams{
			StartDate:   p2Start,
			StartDate_2: p2End,
		})
		if err != nil {
			return nil, ComparePeriodOutput{}, fmt.Errorf("fetching period 2 stats: %w", err)
		}
	}

	// Build output
	p1Dist := toFloat64(p1Stats.TotalDistance)
	p2Dist := toFloat64(p2Stats.TotalDistance)
	p1Dur := toInt64(p1Stats.TotalMovingTime)
	p2Dur := toInt64(p2Stats.TotalMovingTime)

	// Generate insights
	generator := NewInsightGenerator()
	insights := generator.GenerateComparisonInsights(
		p1Stats.ActivityCount, p2Stats.ActivityCount,
		p1Dist, p2Dist,
		p1Dur, p2Dur,
	)

	output := ComparePeriodOutput{
		Period1: PeriodStats{
			DateRange:     fmt.Sprintf("%s to %s", input.Period1Start, input.Period1End),
			ActivityCount: p1Stats.ActivityCount,
			TotalDistance: formatDistance(p1Dist),
			TotalDuration: formatDuration(p1Dur),
			AvgPace:       formatPace(toFloat64(p1Stats.AvgSpeed)),
		},
		Period2: PeriodStats{
			DateRange:     fmt.Sprintf("%s to %s", input.Period2Start, input.Period2End),
			ActivityCount: p2Stats.ActivityCount,
			TotalDistance: formatDistance(p2Dist),
			TotalDuration: formatDuration(p2Dur),
			AvgPace:       formatPace(toFloat64(p2Stats.AvgSpeed)),
		},
		Change: ChangeStats{
			ActivityCount: formatChange(p1Stats.ActivityCount, p2Stats.ActivityCount),
			Distance:      formatDistanceChange(p1Dist, p2Dist),
			Duration:      formatDurationChange(p1Dur, p2Dur),
		},
		Insights:         insights,
		SuggestedActions: SuggestNextActions("comparison"),
	}

	logging.Info("MCP tool completed", "tool", "compare_periods")
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "compare_periods", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Week summary handler
func (s *Server) getWeekSummary(ctx context.Context, req *mcp.CallToolRequest, input WeekSummaryInput) (*mcp.CallToolResult, WeekSummaryOutput, error) {
	week := input.Week
	if week == "" {
		week = "current"
	}

	logging.Info("MCP tool call", "tool", "get_week_summary", "week", week, "type", input.Type)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "get_week_summary", "input", logging.ToJSON(input))
	}

	// Calculate week boundaries
	now := time.Now()
	var weekStart, weekEnd time.Time

	switch week {
	case "current":
		// Start of current week (Monday)
		daysSinceMonday := int(now.Weekday()) - 1
		if daysSinceMonday < 0 {
			daysSinceMonday = 6 // Sunday
		}
		weekStart = time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, now.Location())
		weekEnd = weekStart.AddDate(0, 0, 7).Add(-time.Second)
	case "last":
		daysSinceMonday := int(now.Weekday()) - 1
		if daysSinceMonday < 0 {
			daysSinceMonday = 6
		}
		weekStart = time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday-7, 0, 0, 0, 0, now.Location())
		weekEnd = weekStart.AddDate(0, 0, 7).Add(-time.Second)
	default:
		// Try to parse ISO week format (YYYY-Www)
		weekStart = now.AddDate(0, 0, -7) // Fallback
		weekEnd = now
	}

	weekLabel := fmt.Sprintf("%d-W%02d", weekStart.Year(), weekStart.YearDay()/7+1)

	hasType := input.Type != ""

	// Get week's activities and stats
	startTime := sql.NullTime{Time: weekStart, Valid: true}
	endTime := sql.NullTime{Time: weekEnd, Valid: true}

	var activities []db.Activity
	var err error

	activities, err = s.queries.SearchActivities(ctx, db.SearchActivitiesParams{
		Column1:     sql.NullString{String: input.Type, Valid: hasType},
		Type:        sql.NullString{String: input.Type, Valid: hasType},
		Column3:     startTime,
		StartDate:   startTime,
		Column5:     endTime,
		StartDate_2: endTime,
		Limit:       100,
	})
	if err != nil {
		return nil, WeekSummaryOutput{}, fmt.Errorf("fetching week activities: %w", err)
	}

	// Calculate totals
	var totalDistance float64
	var totalDuration int64
	var totalCalories float64
	var totalElevation float64

	for _, a := range activities {
		if a.Distance.Valid {
			totalDistance += a.Distance.Float64
		}
		if a.MovingTime.Valid {
			totalDuration += a.MovingTime.Int64
		}
		if a.Calories.Valid {
			totalCalories += a.Calories.Float64
		}
		if a.TotalElevationGain.Valid {
			totalElevation += a.TotalElevationGain.Float64
		}
	}

	output := WeekSummaryOutput{
		Week:           weekLabel,
		DateRange:      fmt.Sprintf("%s to %s", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02")),
		ActivityCount:  int64(len(activities)),
		TotalDistance:  formatDistance(totalDistance),
		TotalDuration:  formatDuration(totalDuration),
		TotalCalories:  int(totalCalories),
		TotalElevation: fmt.Sprintf("%.0fm", totalElevation),
		Activities:     convertActivities(activities),
		SuggestedActions: SuggestNextActions("week_summary"),
	}

	// Add basic insight
	if len(activities) > 0 {
		output.Insights = []Insight{
			{
				Type:    "trend",
				Message: fmt.Sprintf("%d activities this week covering %s", len(activities), formatDistance(totalDistance)),
			},
		}
	} else {
		output.Insights = []Insight{
			{
				Type:    "suggestion",
				Message: "No activities recorded this week yet",
			},
		}
	}

	logging.Info("MCP tool completed", "tool", "get_week_summary", "activity_count", len(activities))
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "get_week_summary", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Helper types and functions

type trainingSummaryData struct {
	ActivityCount   int64
	TotalDistance   float64
	TotalMovingTime int64
	AvgSpeed        float64
	AvgHeartrate    float64
	TotalCalories   float64
	TotalElevation  float64
}

func applyLimit(limit int) int {
	if limit <= 0 {
		return defaultActivityLimit
	}
	if limit > maxActivityLimit {
		return maxActivityLimit
	}
	return limit
}

func convertActivities(activities []db.Activity) []ActivitySummary {
	result := make([]ActivitySummary, len(activities))
	for i, a := range activities {
		result[i] = convertActivity(a)
	}
	return result
}

func convertActivity(a db.Activity) ActivitySummary {
	summary := ActivitySummary{
		ID:   a.ID,
		Name: a.Name,
	}

	if a.Type.Valid {
		summary.Type = a.Type.String
	}
	if a.StartDate.Valid {
		summary.Date = a.StartDate.Time.Format("2006-01-02")
	}
	if a.Distance.Valid && a.Distance.Float64 > 0 {
		summary.Distance = formatDistance(a.Distance.Float64)
	}
	if a.MovingTime.Valid && a.MovingTime.Int64 > 0 {
		summary.Duration = formatDuration(a.MovingTime.Int64)
	}
	if a.AverageSpeed.Valid && a.AverageSpeed.Float64 > 0 && a.Distance.Valid && a.Distance.Float64 > 0 {
		summary.Pace = formatPace(a.AverageSpeed.Float64)
	}
	if a.TotalElevationGain.Valid && a.TotalElevationGain.Float64 > 0 {
		summary.ElevationGain = fmt.Sprintf("%.0fm", a.TotalElevationGain.Float64)
	}
	if a.AverageHeartrate.Valid && a.AverageHeartrate.Float64 > 0 {
		summary.AvgHeartrate = int(a.AverageHeartrate.Float64)
	}
	if a.MaxHeartrate.Valid && a.MaxHeartrate.Float64 > 0 {
		summary.MaxHeartrate = int(a.MaxHeartrate.Float64)
	}
	if a.Calories.Valid && a.Calories.Float64 > 0 {
		summary.Calories = int(a.Calories.Float64)
	}

	return summary
}

// formatDistance converts meters to human-readable format
func formatDistance(meters float64) string {
	km := meters / 1000
	if km >= 1 {
		return fmt.Sprintf("%.2f km", km)
	}
	return fmt.Sprintf("%.0f m", meters)
}

// formatDuration converts seconds to human-readable format
func formatDuration(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// formatPace converts m/s to min/km pace
func formatPace(mps float64) string {
	if mps <= 0 {
		return ""
	}
	// Convert m/s to seconds per km
	secPerKm := 1000 / mps
	mins := int(secPerKm) / 60
	secs := int(secPerKm) % 60
	return fmt.Sprintf("%d:%02d/km", mins, secs)
}

// parseServerDateRange parses date strings into sql.NullTime values
func parseServerDateRange(startDate, endDate string) (sql.NullTime, sql.NullTime, error) {
	var start, end sql.NullTime

	if startDate != "" {
		t, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return start, end, fmt.Errorf("parsing start date: %w", err)
		}
		start = sql.NullTime{Time: t, Valid: true}
	}

	if endDate != "" {
		t, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return start, end, fmt.Errorf("parsing end date: %w", err)
		}
		// Set to end of day
		t = t.Add(24*time.Hour - time.Second)
		end = sql.NullTime{Time: t, Valid: true}
	}

	return start, end, nil
}

// aggregatePeriodCounts aggregates monthly period counts
func aggregatePeriodCounts(rows []db.GetActivityCountsByMonthRow, extract func(db.GetActivityCountsByMonthRow) (string, string, int64)) []PeriodSummary {
	periodMap := make(map[string]*PeriodSummary)
	periodOrder := []string{}

	for _, r := range rows {
		period, typeName, count := extract(r)
		if _, exists := periodMap[period]; !exists {
			periodMap[period] = &PeriodSummary{
				Period: period,
				ByType: make(map[string]int64),
			}
			periodOrder = append(periodOrder, period)
		}
		periodMap[period].Total += count
		if typeName != "" {
			periodMap[period].ByType[typeName] = count
		}
	}

	result := make([]PeriodSummary, 0, len(periodOrder))
	for _, period := range periodOrder {
		result = append(result, *periodMap[period])
	}
	return result
}

// aggregatePeriodCountsWeek aggregates weekly period counts
func aggregatePeriodCountsWeek(rows []db.GetActivityCountsByWeekRow, extract func(db.GetActivityCountsByWeekRow) (string, string, int64)) []PeriodSummary {
	periodMap := make(map[string]*PeriodSummary)
	periodOrder := []string{}

	for _, r := range rows {
		period, typeName, count := extract(r)
		if _, exists := periodMap[period]; !exists {
			periodMap[period] = &PeriodSummary{
				Period: period,
				ByType: make(map[string]int64),
			}
			periodOrder = append(periodOrder, period)
		}
		periodMap[period].Total += count
		if typeName != "" {
			periodMap[period].ByType[typeName] = count
		}
	}

	result := make([]PeriodSummary, 0, len(periodOrder))
	for _, period := range periodOrder {
		result = append(result, *periodMap[period])
	}
	return result
}

// formatChange formats the change between two values with percentage
func formatChange(old, new int64) string {
	diff := new - old
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	pct := float64(0)
	if old > 0 {
		pct = float64(diff) / float64(old) * 100
	}
	return fmt.Sprintf("%s%d (%.0f%%)", sign, diff, pct)
}

// formatDistanceChange formats distance change
func formatDistanceChange(oldMeters, newMeters float64) string {
	diff := newMeters - oldMeters
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	pct := float64(0)
	if oldMeters > 0 {
		pct = diff / oldMeters * 100
	}
	return fmt.Sprintf("%s%.2f km (%.0f%%)", sign, diff/1000, pct)
}

// formatDurationChange formats duration change
func formatDurationChange(oldSecs, newSecs int64) string {
	diff := newSecs - oldSecs
	sign := "+"
	if diff < 0 {
		sign = ""
		diff = -diff
	}
	pct := float64(0)
	if oldSecs > 0 {
		pct = float64(newSecs-oldSecs) / float64(oldSecs) * 100
	}
	hours := diff / 3600
	mins := (diff % 3600) / 60
	if hours > 0 {
		return fmt.Sprintf("%s%dh %dm (%.0f%%)", sign, hours, mins, pct)
	}
	return fmt.Sprintf("%s%dm (%.0f%%)", sign, mins, pct)
}
