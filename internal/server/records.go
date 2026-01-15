package server

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RecordsQuerier defines the interface for personal records queries
type RecordsQuerier interface {
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
}

// Input types

// GetPersonalRecordsInput - input for retrieving personal records
type GetPersonalRecordsInput struct {
	Type       string   `json:"type,omitempty" jsonschema:"Filter records to a specific activity type. Common values: Run, Ride, Swim, Walk, Hike. Leave empty for best records across all activity types."`
	Categories []string `json:"categories,omitempty" jsonschema:"Which record categories to include. Valid values: 'fastest' (best pace), 'longest_distance' (furthest), 'longest_duration' (most time), 'highest_elevation' (most climbing), 'most_calories' (highest calorie burn). Omit or leave empty for all categories."`
}

// Output types

type GetPersonalRecordsOutput struct {
	Type             string            `json:"type,omitempty"`
	Records          []PersonalRecord  `json:"records"`
	Insights         []Insight         `json:"insights"`
	SuggestedActions []SuggestedAction `json:"suggested_actions"`
}

type PersonalRecord struct {
	Category     string          `json:"category"`
	Activity     ActivitySummary `json:"activity"`
	RecordValue  string          `json:"record_value"`
	RecordMetric string          `json:"record_metric"`
}

// registerRecordsTools registers the personal records tool
func (s *Server) registerRecordsTools() {
	logging.Debug("Registering tool", "name", "get_personal_records")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_personal_records",
		Description: `Get personal best records across various categories including fastest, longest, and most intense activities.

Use when:
- User asks "What are my PRs?" or "What's my longest run?"
- User wants to see their all-time best performances
- User asks "Show me my fastest activities" or "What's my elevation record?"

Parameters:
- type (string): Filter to a specific activity type (Run, Ride, Swim, etc.). Leave empty for overall bests.
- categories (array): Which record categories to include: "fastest", "longest_distance", "longest_duration", "highest_elevation", "most_calories". Omit for all categories.

Returns: List of personal records, each with category name, the activity details (id, name, date, etc.), the record value, and the metric type. Includes achievement insights.

Example: {"type": "Run"} or {"categories": ["fastest", "longest_distance"]}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Get Personal Records",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.getPersonalRecords)
}

// getPersonalRecords retrieves personal bests across various categories
func (s *Server) getPersonalRecords(ctx context.Context, req *mcp.CallToolRequest, input GetPersonalRecordsInput) (*mcp.CallToolResult, GetPersonalRecordsOutput, error) {
	logging.Info("MCP tool call", "tool", "get_personal_records", "type", input.Type, "categories", input.Categories)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "get_personal_records", "input", logging.ToJSON(input))
	}

	queries := s.queries.(RecordsQuerier)
	hasType := input.Type != ""

	// Determine which categories to fetch
	categories := input.Categories
	if len(categories) == 0 {
		categories = []string{"fastest", "longest_distance", "longest_duration", "highest_elevation", "most_calories"}
	}

	categorySet := make(map[string]bool)
	for _, c := range categories {
		categorySet[c] = true
	}

	records := make([]PersonalRecord, 0)
	var insights []Insight

	// Fastest activity
	if categorySet["fastest"] {
		var activity db.Activity
		var err error
		if hasType {
			activity, err = queries.GetFastestActivityByType(ctx, sql.NullString{String: input.Type, Valid: true})
		} else {
			activity, err = queries.GetFastestActivity(ctx)
		}
		if err == nil {
			pace := ""
			if activity.AverageSpeed.Valid && activity.AverageSpeed.Float64 > 0 {
				pace = formatPace(activity.AverageSpeed.Float64)
			}
			records = append(records, PersonalRecord{
				Category:     "fastest",
				Activity:     convertActivity(activity),
				RecordValue:  pace,
				RecordMetric: "pace",
			})
		}
	}

	// Longest distance activity
	if categorySet["longest_distance"] {
		var activity db.Activity
		var err error
		if hasType {
			activity, err = queries.GetLongestDistanceActivityByType(ctx, sql.NullString{String: input.Type, Valid: true})
		} else {
			activity, err = queries.GetLongestDistanceActivity(ctx)
		}
		if err == nil {
			distance := ""
			if activity.Distance.Valid && activity.Distance.Float64 > 0 {
				distance = formatDistance(activity.Distance.Float64)
			}
			records = append(records, PersonalRecord{
				Category:     "longest_distance",
				Activity:     convertActivity(activity),
				RecordValue:  distance,
				RecordMetric: "distance",
			})
		}
	}

	// Longest duration activity
	if categorySet["longest_duration"] {
		var activity db.Activity
		var err error
		if hasType {
			activity, err = queries.GetLongestDurationActivityByType(ctx, sql.NullString{String: input.Type, Valid: true})
		} else {
			activity, err = queries.GetLongestDurationActivity(ctx)
		}
		if err == nil {
			duration := ""
			if activity.MovingTime.Valid && activity.MovingTime.Int64 > 0 {
				duration = formatDuration(activity.MovingTime.Int64)
			}
			records = append(records, PersonalRecord{
				Category:     "longest_duration",
				Activity:     convertActivity(activity),
				RecordValue:  duration,
				RecordMetric: "duration",
			})
		}
	}

	// Highest elevation activity
	if categorySet["highest_elevation"] {
		var activity db.Activity
		var err error
		if hasType {
			activity, err = queries.GetHighestElevationActivityByType(ctx, sql.NullString{String: input.Type, Valid: true})
		} else {
			activity, err = queries.GetHighestElevationActivity(ctx)
		}
		if err == nil {
			elevation := ""
			if activity.TotalElevationGain.Valid && activity.TotalElevationGain.Float64 > 0 {
				elevation = formatElevationHuman(activity.TotalElevationGain.Float64)
			}
			records = append(records, PersonalRecord{
				Category:     "highest_elevation",
				Activity:     convertActivity(activity),
				RecordValue:  elevation,
				RecordMetric: "elevation",
			})
		}
	}

	// Most calories activity
	if categorySet["most_calories"] {
		var activity db.Activity
		var err error
		if hasType {
			activity, err = queries.GetMostCaloriesActivityByType(ctx, sql.NullString{String: input.Type, Valid: true})
		} else {
			activity, err = queries.GetMostCaloriesActivity(ctx)
		}
		if err == nil {
			calories := ""
			if activity.Calories.Valid && activity.Calories.Float64 > 0 {
				calories = formatCalories(int(activity.Calories.Float64))
			}
			records = append(records, PersonalRecord{
				Category:     "most_calories",
				Activity:     convertActivity(activity),
				RecordValue:  calories,
				RecordMetric: "calories",
			})
		}
	}

	// Generate insights
	if len(records) > 0 {
		insights = append(insights, Insight{
			Type:    "achievement",
			Message: generateRecordsInsight(records, input.Type),
		})
	}

	output := GetPersonalRecordsOutput{
		Type:             input.Type,
		Records:          records,
		Insights:         insights,
		SuggestedActions: SuggestNextActions("records"),
	}

	logging.Info("MCP tool completed", "tool", "get_personal_records", "record_count", len(records))
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "get_personal_records", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// formatCalories formats calories value
func formatCalories(calories int) string {
	if calories >= 1000 {
		return formatWithCommas(calories) + " kcal"
	}
	return formatWithCommas(calories) + " kcal"
}

// formatWithCommas adds thousand separators to a number
func formatWithCommas(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}
	// Add thousand separators for larger numbers
	str := ""
	for n > 0 {
		if len(str) > 0 && len(str)%4 == 3 {
			str = "," + str
		}
		str = string(rune('0'+n%10)) + str
		n /= 10
	}
	return str
}

// generateRecordsInsight creates a summary insight for the records
func generateRecordsInsight(records []PersonalRecord, activityType string) string {
	if len(records) == 0 {
		return "No personal records found"
	}

	typeStr := "across all activities"
	if activityType != "" {
		typeStr = "for " + activityType
	}

	return "Found " + strconv.Itoa(len(records)) + " personal records " + typeStr
}
