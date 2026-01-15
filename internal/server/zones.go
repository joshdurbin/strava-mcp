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

// ZonesQuerier defines the interface for zone queries
type ZonesQuerier interface {
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
	GetActivity(ctx context.Context, id int64) (db.Activity, error)
}

// Zone input types

// GetActivityZonesInput - input for retrieving zone data for a specific activity
type GetActivityZonesInput struct {
	ActivityID int64 `json:"activity_id" jsonschema:"The Strava activity ID to retrieve zones for. This is the unique numeric identifier visible in the activity URL on Strava. Required."`
}

// AnalyzeZonesInput - input for aggregated zone analysis across activities
type AnalyzeZonesInput struct {
	ZoneType  string `json:"zone_type,omitempty" jsonschema:"Type of training zones to analyze. Valid values: 'heartrate' (heart rate zones 1-5) or 'power' (power zones for cycling). Default: heartrate."`
	Type      string `json:"type,omitempty" jsonschema:"Filter analysis to a specific activity type. Common values: Run, Ride, Swim. Leave empty to analyze all activities with zone data."`
	StartDate string `json:"start_date,omitempty" jsonschema:"Include activities on or after this date. Format: YYYY-MM-DD. Leave empty for all-time analysis."`
	EndDate   string `json:"end_date,omitempty" jsonschema:"Include activities on or before this date. Format: YYYY-MM-DD. Leave empty to include up to today."`
}

// Zone output types

type ActivityZonesOutput struct {
	ActivityID       int64             `json:"activity_id"`
	ActivityName     string            `json:"activity_name,omitempty"`
	ActivityType     string            `json:"activity_type,omitempty"`
	Zones            []ZoneData        `json:"zones"`
	Insights         []Insight         `json:"insights,omitempty"`
	SuggestedActions []SuggestedAction `json:"suggested_actions,omitempty"`
}

type ZoneData struct {
	Type        string       `json:"type"` // heartrate or power
	SensorBased bool         `json:"sensor_based"`
	Buckets     []ZoneBucket `json:"buckets"`
	TotalTime   string       `json:"total_time"`
}

type ZoneBucket struct {
	Zone       int     `json:"zone"`        // 1-5
	MinValue   int     `json:"min_value"`   // BPM or watts
	MaxValue   int     `json:"max_value"`   // BPM or watts (-1 for unbounded)
	TimeSpent  string  `json:"time_spent"`  // Human readable
	Percentage float64 `json:"percentage"`  // % of total zone time
}

type AnalyzeZonesOutput struct {
	ZoneType         string            `json:"zone_type"`
	Zones            []ZoneSummaryRow  `json:"zones"`
	TotalTime        string            `json:"total_time"`
	ActivityCount    int64             `json:"activity_count"`
	Filter           string            `json:"filter,omitempty"`
	Insights         []Insight         `json:"insights"`
	SuggestedActions []SuggestedAction `json:"suggested_actions"`
}

type ZoneSummaryRow struct {
	Zone          int     `json:"zone"`
	TotalTime     string  `json:"total_time"`
	AvgTime       string  `json:"avg_time_per_activity"`
	ActivityCount int64   `json:"activity_count"`
	Percentage    float64 `json:"percentage_of_total"`
}

// registerZoneTools registers all zone-related MCP tools
func (s *Server) registerZoneTools() {
	logging.Debug("Registering tool", "name", "get_activity_zones")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_activity_zones",
		Description: `Get heart rate and power zone distribution for a specific activity.

Use when:
- User asks "Show me the zones for my last run" or "What zones was I in during activity X?"
- User wants detailed zone breakdown for a single workout
- User needs to analyze training intensity of a specific session

Parameters:
- activity_id (integer, required): The Strava activity ID to analyze.

Returns: Activity name/type, list of zones (heartrate and/or power) with buckets showing zone number, BPM/watt ranges, time spent, and percentage of total time. Includes training intensity insights.

Note: Requires Strava Summit subscription for zone data to be available.

Example: {"activity_id": 12345678901}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Get Activity Zones",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.getActivityZones)

	logging.Debug("Registering tool", "name", "analyze_zones")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "analyze_zones",
		Description: `Analyze aggregated training zone distribution across multiple activities with intensity insights.

Use when:
- User asks "How much time do I spend in Zone 2?" or "Analyze my heart rate zones"
- User wants to understand their training intensity distribution
- User asks "Am I training at the right intensity?" or about the 80/20 rule

Parameters:
- zone_type (string): Type of zones to analyze: "heartrate" or "power". Default: "heartrate".
- type (string): Filter by activity type (Run, Ride, etc.). Leave empty for all types.
- start_date (string): Start date in YYYY-MM-DD format. Leave empty for all time.
- end_date (string): End date in YYYY-MM-DD format. Leave empty for all time.

Returns: Zone-by-zone breakdown with total time, average time per activity, activity count, and percentage of total training time. Includes insights about training intensity balance (80/20 rule compliance).

Note: Requires Strava Summit subscription for zone data to be available.

Example: {"zone_type": "heartrate", "type": "Run"} or {"zone_type": "power", "start_date": "2024-01-01"}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Analyze Training Zones",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.analyzeZones)
}

// getActivityZones returns zone distribution for a specific activity
func (s *Server) getActivityZones(ctx context.Context, req *mcp.CallToolRequest, input GetActivityZonesInput) (*mcp.CallToolResult, ActivityZonesOutput, error) {
	logging.Info("MCP tool call", "tool", "get_activity_zones", "activity_id", input.ActivityID)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "get_activity_zones", "input", logging.ToJSON(input))
	}

	queries := s.queries.(ZonesQuerier)

	// Get activity info
	activity, err := queries.GetActivity(ctx, input.ActivityID)
	if err != nil {
		if err == sql.ErrNoRows {
			logging.Info("MCP tool completed", "tool", "get_activity_zones", "found", false)
			return nil, ActivityZonesOutput{ActivityID: input.ActivityID, Zones: []ZoneData{}}, nil
		}
		logging.Error("get_activity_zones failed", "error", err)
		return nil, ActivityZonesOutput{}, fmt.Errorf("querying activity: %w", err)
	}

	// Get zones for this activity
	zones, err := queries.GetActivityZones(ctx, input.ActivityID)
	if err != nil {
		logging.Error("get_activity_zones failed", "error", err)
		return nil, ActivityZonesOutput{}, fmt.Errorf("querying zones: %w", err)
	}

	output := ActivityZonesOutput{
		ActivityID:   input.ActivityID,
		ActivityName: activity.Name,
		Zones:        []ZoneData{},
	}
	if activity.Type.Valid {
		output.ActivityType = activity.Type.String
	}

	// Collect zone percentages for insights
	zonePercentages := make(map[int]float64)

	for _, zone := range zones {
		buckets, err := queries.GetZoneBuckets(ctx, zone.ID)
		if err != nil {
			continue
		}

		zoneData := ZoneData{
			Type:        zone.ZoneType,
			SensorBased: zone.SensorBased == 1,
			Buckets:     []ZoneBucket{},
		}

		// Calculate total time for percentage
		var totalSeconds int64
		for _, b := range buckets {
			totalSeconds += b.TimeSeconds
		}

		for _, b := range buckets {
			var pct float64
			if totalSeconds > 0 {
				pct = float64(b.TimeSeconds) / float64(totalSeconds) * 100
			}
			zoneData.Buckets = append(zoneData.Buckets, ZoneBucket{
				Zone:       int(b.ZoneNumber),
				MinValue:   int(b.MinValue),
				MaxValue:   int(b.MaxValue),
				TimeSpent:  formatDurationHuman(b.TimeSeconds),
				Percentage: pct,
			})

			// Collect for insights (heartrate zones only)
			if zone.ZoneType == "heartrate" {
				zonePercentages[int(b.ZoneNumber)] = pct
			}
		}
		zoneData.TotalTime = formatDurationHuman(totalSeconds)

		output.Zones = append(output.Zones, zoneData)
	}

	// Generate insights
	if len(zonePercentages) > 0 {
		generator := NewInsightGenerator()
		output.Insights = generator.GenerateZoneInsights(zonePercentages)
	}
	output.SuggestedActions = SuggestNextActions("zones")

	logging.Info("MCP tool completed", "tool", "get_activity_zones", "zone_count", len(output.Zones))
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "get_activity_zones", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// analyzeZones returns aggregated zone statistics with insights
func (s *Server) analyzeZones(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeZonesInput) (*mcp.CallToolResult, AnalyzeZonesOutput, error) {
	zoneType := input.ZoneType
	if zoneType == "" {
		zoneType = "heartrate"
	}

	logging.Info("MCP tool call", "tool", "analyze_zones", "zone_type", zoneType, "type", input.Type, "start", input.StartDate, "end", input.EndDate)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "analyze_zones", "input", logging.ToJSON(input))
	}

	queries := s.queries.(ZonesQuerier)
	filter := buildFilterDesc(input.Type, input.StartDate, input.EndDate)

	hasType := input.Type != ""
	hasDateRange := input.StartDate != "" || input.EndDate != ""

	var rows []hrZoneRow
	var err error

	if zoneType == "heartrate" {
		if hasType && hasDateRange {
			start, end, parseErr := parseZoneDateRange(input.StartDate, input.EndDate)
			if parseErr != nil {
				return nil, AnalyzeZonesOutput{}, parseErr
			}
			dbRows, dbErr := queries.GetHeartRateZoneSummaryByTypeInRange(ctx, db.GetHeartRateZoneSummaryByTypeInRangeParams{
				Type:        sql.NullString{String: input.Type, Valid: true},
				StartDate:   start,
				StartDate_2: end,
			})
			err = dbErr
			rows = convertHRZoneRowsByTypeInRange(dbRows)
		} else if hasType {
			dbRows, dbErr := queries.GetHeartRateZoneSummaryByType(ctx, sql.NullString{String: input.Type, Valid: true})
			err = dbErr
			rows = convertHRZoneRowsByType(dbRows)
		} else if hasDateRange {
			start, end, parseErr := parseZoneDateRange(input.StartDate, input.EndDate)
			if parseErr != nil {
				return nil, AnalyzeZonesOutput{}, parseErr
			}
			dbRows, dbErr := queries.GetHeartRateZoneSummaryInRange(ctx, db.GetHeartRateZoneSummaryInRangeParams{
				StartDate:   start,
				StartDate_2: end,
			})
			err = dbErr
			rows = convertHRZoneRowsInRange(dbRows)
		} else {
			dbRows, dbErr := queries.GetHeartRateZoneSummary(ctx)
			err = dbErr
			rows = convertHRZoneRows(dbRows)
		}
	} else {
		// Power zones
		if hasType && hasDateRange {
			start, end, parseErr := parseZoneDateRange(input.StartDate, input.EndDate)
			if parseErr != nil {
				return nil, AnalyzeZonesOutput{}, parseErr
			}
			dbRows, dbErr := queries.GetPowerZoneSummaryInRange(ctx, db.GetPowerZoneSummaryInRangeParams{
				StartDate:   start,
				StartDate_2: end,
			})
			err = dbErr
			rows = convertPowerZoneRowsInRange(dbRows)
		} else if hasType {
			dbRows, dbErr := queries.GetPowerZoneSummaryByType(ctx, sql.NullString{String: input.Type, Valid: true})
			err = dbErr
			rows = convertPowerZoneRowsByType(dbRows)
		} else if hasDateRange {
			start, end, parseErr := parseZoneDateRange(input.StartDate, input.EndDate)
			if parseErr != nil {
				return nil, AnalyzeZonesOutput{}, parseErr
			}
			dbRows, dbErr := queries.GetPowerZoneSummaryInRange(ctx, db.GetPowerZoneSummaryInRangeParams{
				StartDate:   start,
				StartDate_2: end,
			})
			err = dbErr
			rows = convertPowerZoneRowsInRange(dbRows)
		} else {
			dbRows, dbErr := queries.GetPowerZoneSummary(ctx)
			err = dbErr
			rows = convertPowerZoneRows(dbRows)
		}
	}

	if err != nil {
		logging.Error("analyze_zones failed", "error", err)
		return nil, AnalyzeZonesOutput{}, fmt.Errorf("querying zone summary: %w", err)
	}

	// Build output with insights
	output := buildAnalyzeZonesOutput(rows, zoneType, filter)

	logging.Info("MCP tool completed", "tool", "analyze_zones", "zone_count", len(output.Zones))
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "analyze_zones", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Helper types and functions for zone processing

type hrZoneRow struct {
	ZoneNumber    int64
	TotalTime     int64
	AvgTime       float64
	ActivityCount int64
}

func parseZoneDateRange(startDate, endDate string) (sql.NullTime, sql.NullTime, error) {
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

func convertHRZoneRows(rows []db.GetHeartRateZoneSummaryRow) []hrZoneRow {
	result := make([]hrZoneRow, len(rows))
	for i, r := range rows {
		result[i] = hrZoneRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     toInt64(r.TotalTime),
			AvgTime:       toFloat64(r.AvgTime),
			ActivityCount: r.ActivityCount,
		}
	}
	return result
}

func convertHRZoneRowsByType(rows []db.GetHeartRateZoneSummaryByTypeRow) []hrZoneRow {
	result := make([]hrZoneRow, len(rows))
	for i, r := range rows {
		result[i] = hrZoneRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     toInt64(r.TotalTime),
			AvgTime:       toFloat64(r.AvgTime),
			ActivityCount: r.ActivityCount,
		}
	}
	return result
}

func convertHRZoneRowsInRange(rows []db.GetHeartRateZoneSummaryInRangeRow) []hrZoneRow {
	result := make([]hrZoneRow, len(rows))
	for i, r := range rows {
		result[i] = hrZoneRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     toInt64(r.TotalTime),
			AvgTime:       toFloat64(r.AvgTime),
			ActivityCount: r.ActivityCount,
		}
	}
	return result
}

func convertHRZoneRowsByTypeInRange(rows []db.GetHeartRateZoneSummaryByTypeInRangeRow) []hrZoneRow {
	result := make([]hrZoneRow, len(rows))
	for i, r := range rows {
		result[i] = hrZoneRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     toInt64(r.TotalTime),
			AvgTime:       toFloat64(r.AvgTime),
			ActivityCount: r.ActivityCount,
		}
	}
	return result
}

func convertPowerZoneRows(rows []db.GetPowerZoneSummaryRow) []hrZoneRow {
	result := make([]hrZoneRow, len(rows))
	for i, r := range rows {
		result[i] = hrZoneRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     toInt64(r.TotalTime),
			AvgTime:       toFloat64(r.AvgTime),
			ActivityCount: r.ActivityCount,
		}
	}
	return result
}

func convertPowerZoneRowsByType(rows []db.GetPowerZoneSummaryByTypeRow) []hrZoneRow {
	result := make([]hrZoneRow, len(rows))
	for i, r := range rows {
		result[i] = hrZoneRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     toInt64(r.TotalTime),
			AvgTime:       toFloat64(r.AvgTime),
			ActivityCount: r.ActivityCount,
		}
	}
	return result
}

func convertPowerZoneRowsInRange(rows []db.GetPowerZoneSummaryInRangeRow) []hrZoneRow {
	result := make([]hrZoneRow, len(rows))
	for i, r := range rows {
		result[i] = hrZoneRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     toInt64(r.TotalTime),
			AvgTime:       toFloat64(r.AvgTime),
			ActivityCount: r.ActivityCount,
		}
	}
	return result
}

func buildAnalyzeZonesOutput(rows []hrZoneRow, zoneType, filter string) AnalyzeZonesOutput {
	output := AnalyzeZonesOutput{
		ZoneType: zoneType,
		Zones:    make([]ZoneSummaryRow, 0, len(rows)),
		Filter:   filter,
	}

	var totalTime int64
	var maxActivityCount int64
	zonePercentages := make(map[int]float64)

	for _, r := range rows {
		totalTime += r.TotalTime
		if r.ActivityCount > maxActivityCount {
			maxActivityCount = r.ActivityCount
		}
	}

	for _, r := range rows {
		var pct float64
		if totalTime > 0 {
			pct = float64(r.TotalTime) / float64(totalTime) * 100
		}
		output.Zones = append(output.Zones, ZoneSummaryRow{
			Zone:          int(r.ZoneNumber),
			TotalTime:     formatDurationHuman(r.TotalTime),
			AvgTime:       formatDurationHuman(int64(r.AvgTime)),
			ActivityCount: r.ActivityCount,
			Percentage:    pct,
		})
		zonePercentages[int(r.ZoneNumber)] = pct
	}

	output.TotalTime = formatDurationHuman(totalTime)
	output.ActivityCount = maxActivityCount

	// Generate insights
	generator := NewInsightGenerator()
	output.Insights = generator.GenerateZoneInsights(zonePercentages)
	output.SuggestedActions = SuggestNextActions("zones")

	return output
}
