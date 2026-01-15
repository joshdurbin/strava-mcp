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

// ProgressQuerier defines the interface for progress-related queries
type ProgressQuerier interface {
	GetPeriodStats(ctx context.Context, arg db.GetPeriodStatsParams) (db.GetPeriodStatsRow, error)
	GetPeriodStatsByType(ctx context.Context, arg db.GetPeriodStatsByTypeParams) (db.GetPeriodStatsByTypeRow, error)
	GetWeeklyVolume(ctx context.Context, arg db.GetWeeklyVolumeParams) ([]db.GetWeeklyVolumeRow, error)
	GetWeeklyVolumeByType(ctx context.Context, arg db.GetWeeklyVolumeByTypeParams) ([]db.GetWeeklyVolumeByTypeRow, error)
	GetTrainingSummaryInRange(ctx context.Context, arg db.GetTrainingSummaryInRangeParams) (db.GetTrainingSummaryInRangeRow, error)
	GetTrainingSummaryByTypeInRange(ctx context.Context, arg db.GetTrainingSummaryByTypeInRangeParams) (db.GetTrainingSummaryByTypeInRangeRow, error)
}

// Input types

// AnalyzeProgressInput - input for analyzing progress trends over time
type AnalyzeProgressInput struct {
	Metric    string `json:"metric,omitempty" jsonschema:"Which performance metric to analyze for progress. Valid values: 'pace' (speed improvement), 'distance' (volume increase), 'duration' (time increase), 'elevation' (climbing increase). Default: pace."`
	Type      string `json:"type,omitempty" jsonschema:"Filter analysis to a specific activity type. Common values: Run, Ride, Swim. Leave empty to analyze across all activity types."`
	Timeframe string `json:"timeframe,omitempty" jsonschema:"Time period to analyze. Valid values: 'last_30_days', 'last_90_days', 'last_6_months', 'last_year'. Compares this period to the equivalent previous period. Default: last_90_days."`
}

// CheckTrainingLoadInput - input for analyzing training load and volume
type CheckTrainingLoadInput struct {
	Weeks int    `json:"weeks,omitempty" jsonschema:"Number of recent weeks to analyze for training load. Range: 1-12. Default: 4 weeks."`
	Type  string `json:"type,omitempty" jsonschema:"Filter analysis to a specific activity type. Common values: Run, Ride, Swim. Leave empty to analyze total training load across all activities."`
}

// Output types

type AnalyzeProgressOutput struct {
	Metric           string            `json:"metric"`
	Type             string            `json:"type,omitempty"`
	Timeframe        string            `json:"timeframe"`
	CurrentPeriod    PeriodMetrics     `json:"current_period"`
	PreviousPeriod   PeriodMetrics     `json:"previous_period"`
	Trend            string            `json:"trend"` // "improving", "stable", "declining"
	ChangePercent    float64           `json:"change_percent"`
	Insights         []Insight         `json:"insights"`
	SuggestedActions []SuggestedAction `json:"suggested_actions"`
}

type PeriodMetrics struct {
	DateRange     string `json:"date_range"`
	ActivityCount int64  `json:"activity_count"`
	Value         string `json:"value"` // Formatted metric value
	RawValue      float64 `json:"raw_value,omitempty"`
}

type CheckTrainingLoadOutput struct {
	CurrentWeek      WeeklyLoadSummary   `json:"current_week"`
	RecentWeeks      []WeeklyLoadSummary `json:"recent_weeks"`
	AverageWeekly    WeeklyLoadSummary   `json:"average_weekly"`
	LoadStatus       string              `json:"load_status"` // "overreaching", "optimal", "maintaining", "undertraining"
	LoadChangePercent float64            `json:"load_change_percent"`
	Insights         []Insight           `json:"insights"`
	SuggestedActions []SuggestedAction   `json:"suggested_actions"`
}

type WeeklyLoadSummary struct {
	Week          string `json:"week,omitempty"`
	ActivityCount int64  `json:"activity_count"`
	TotalDistance string `json:"total_distance"`
	TotalDuration string `json:"total_duration"`
	TotalCalories int    `json:"total_calories,omitempty"`
	TotalElevation string `json:"total_elevation,omitempty"`
}

// registerProgressTools registers the progress analysis tools
func (s *Server) registerProgressTools() {
	logging.Debug("Registering tool", "name", "analyze_progress")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "analyze_progress",
		Description: `Analyze progress trends by comparing current performance to a previous period.

Use when:
- User asks "Am I getting faster?" or "How has my pace improved?"
- User wants to know if they're making progress on a specific metric
- User asks "What's my distance trend?" or "Am I running more?"

Parameters:
- metric (string): Which metric to analyze: "pace", "distance", "duration", "elevation". Default: "pace".
- type (string): Filter to a specific activity type (Run, Ride, Swim, etc.). Leave empty for all types.
- timeframe (string): Analysis period: "last_30_days", "last_90_days", "last_6_months", "last_year". Default: "last_90_days".

Returns: Current period metrics vs previous period metrics, trend direction (improving/stable/declining), percentage change, and progress insights.

Example: {"metric": "pace", "type": "Run", "timeframe": "last_90_days"}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Analyze Progress",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.analyzeProgress)

	logging.Debug("Registering tool", "name", "check_training_load")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "check_training_load",
		Description: `Analyze weekly training volume to detect overtraining or undertraining patterns.

Use when:
- User asks "Am I overtraining?" or "Am I doing too much?"
- User wants to compare this week to their average
- User asks "What's my weekly volume?" or "How consistent is my training?"

Parameters:
- weeks (integer): Number of weeks to analyze. Range: 1-12. Default: 4.
- type (string): Filter to a specific activity type (Run, Ride, Swim, etc.). Leave empty for total load.

Returns: Current week summary, recent weeks breakdown, average weekly metrics, load status (overreaching/optimal/maintaining/undertraining), percentage change from average, and training load insights.

Example: {"weeks": 4, "type": "Run"} or {"weeks": 8}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Check Training Load",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.checkTrainingLoad)
}

// analyzeProgress compares performance across periods to detect trends
func (s *Server) analyzeProgress(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeProgressInput) (*mcp.CallToolResult, AnalyzeProgressOutput, error) {
	metric := input.Metric
	if metric == "" {
		metric = "pace"
	}
	timeframe := input.Timeframe
	if timeframe == "" {
		timeframe = "last_90_days"
	}

	logging.Info("MCP tool call", "tool", "analyze_progress", "metric", metric, "type", input.Type, "timeframe", timeframe)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "analyze_progress", "input", logging.ToJSON(input))
	}

	// Calculate date ranges based on timeframe
	now := time.Now()
	var periodDays int
	switch timeframe {
	case "last_30_days":
		periodDays = 30
	case "last_90_days":
		periodDays = 90
	case "last_6_months":
		periodDays = 180
	case "last_year":
		periodDays = 365
	default:
		periodDays = 90
	}

	currentEnd := now
	currentStart := now.AddDate(0, 0, -periodDays)
	previousEnd := currentStart.AddDate(0, 0, -1)
	previousStart := previousEnd.AddDate(0, 0, -periodDays)

	queries := s.queries.(ProgressQuerier)
	hasType := input.Type != ""

	// Fetch current period stats
	var currentStats, previousStats periodStatsData

	if hasType {
		current, err := queries.GetPeriodStatsByType(ctx, db.GetPeriodStatsByTypeParams{
			Type:        sql.NullString{String: input.Type, Valid: true},
			StartDate:   sql.NullTime{Time: currentStart, Valid: true},
			StartDate_2: sql.NullTime{Time: currentEnd, Valid: true},
		})
		if err != nil {
			return nil, AnalyzeProgressOutput{}, fmt.Errorf("fetching current period stats: %w", err)
		}
		currentStats = periodStatsData{
			ActivityCount: current.ActivityCount,
			TotalDistance: toFloat64(current.TotalDistance),
			TotalDuration: toInt64(current.TotalMovingTime),
			AvgSpeed:      toFloat64(current.AvgSpeed),
		}

		previous, err := queries.GetPeriodStatsByType(ctx, db.GetPeriodStatsByTypeParams{
			Type:        sql.NullString{String: input.Type, Valid: true},
			StartDate:   sql.NullTime{Time: previousStart, Valid: true},
			StartDate_2: sql.NullTime{Time: previousEnd, Valid: true},
		})
		if err != nil {
			return nil, AnalyzeProgressOutput{}, fmt.Errorf("fetching previous period stats: %w", err)
		}
		previousStats = periodStatsData{
			ActivityCount: previous.ActivityCount,
			TotalDistance: toFloat64(previous.TotalDistance),
			TotalDuration: toInt64(previous.TotalMovingTime),
			AvgSpeed:      toFloat64(previous.AvgSpeed),
		}
	} else {
		current, err := queries.GetPeriodStats(ctx, db.GetPeriodStatsParams{
			StartDate:   sql.NullTime{Time: currentStart, Valid: true},
			StartDate_2: sql.NullTime{Time: currentEnd, Valid: true},
		})
		if err != nil {
			return nil, AnalyzeProgressOutput{}, fmt.Errorf("fetching current period stats: %w", err)
		}
		currentStats = periodStatsData{
			ActivityCount: current.ActivityCount,
			TotalDistance: toFloat64(current.TotalDistance),
			TotalDuration: toInt64(current.TotalMovingTime),
			AvgSpeed:      toFloat64(current.AvgSpeed),
		}

		previous, err := queries.GetPeriodStats(ctx, db.GetPeriodStatsParams{
			StartDate:   sql.NullTime{Time: previousStart, Valid: true},
			StartDate_2: sql.NullTime{Time: previousEnd, Valid: true},
		})
		if err != nil {
			return nil, AnalyzeProgressOutput{}, fmt.Errorf("fetching previous period stats: %w", err)
		}
		previousStats = periodStatsData{
			ActivityCount: previous.ActivityCount,
			TotalDistance: toFloat64(previous.TotalDistance),
			TotalDuration: toInt64(previous.TotalMovingTime),
			AvgSpeed:      toFloat64(previous.AvgSpeed),
		}
	}

	// Extract the relevant metric
	var currentValue, previousValue float64
	var currentFormatted, previousFormatted string
	var higherIsBetter bool

	switch metric {
	case "pace":
		// For pace, we use average speed (higher is better = faster pace)
		currentValue = currentStats.AvgSpeed
		previousValue = previousStats.AvgSpeed
		currentFormatted = formatPace(currentValue)
		previousFormatted = formatPace(previousValue)
		higherIsBetter = true // Higher speed = better
	case "distance":
		currentValue = currentStats.TotalDistance
		previousValue = previousStats.TotalDistance
		currentFormatted = formatDistance(currentValue)
		previousFormatted = formatDistance(previousValue)
		higherIsBetter = true
	case "duration":
		currentValue = float64(currentStats.TotalDuration)
		previousValue = float64(previousStats.TotalDuration)
		currentFormatted = formatDuration(currentStats.TotalDuration)
		previousFormatted = formatDuration(previousStats.TotalDuration)
		higherIsBetter = true
	case "elevation":
		// Would need elevation data - for now use distance as proxy
		currentValue = currentStats.TotalDistance
		previousValue = previousStats.TotalDistance
		currentFormatted = formatDistance(currentValue)
		previousFormatted = formatDistance(previousValue)
		higherIsBetter = true
	default:
		currentValue = currentStats.AvgSpeed
		previousValue = previousStats.AvgSpeed
		currentFormatted = formatPace(currentValue)
		previousFormatted = formatPace(previousValue)
		higherIsBetter = true
	}

	// Calculate change and trend
	var changePercent float64
	var trend string
	if previousValue > 0 {
		changePercent = ((currentValue - previousValue) / previousValue) * 100
		improving := (higherIsBetter && changePercent > 0) || (!higherIsBetter && changePercent < 0)

		if changePercent > -5 && changePercent < 5 {
			trend = "stable"
		} else if improving {
			trend = "improving"
		} else {
			trend = "declining"
		}
	} else {
		trend = "insufficient_data"
	}

	// Generate insights
	generator := NewInsightGenerator()
	insights := generator.GenerateProgressInsights(currentValue, previousValue, metric, higherIsBetter)

	output := AnalyzeProgressOutput{
		Metric:    metric,
		Type:      input.Type,
		Timeframe: timeframe,
		CurrentPeriod: PeriodMetrics{
			DateRange:     fmt.Sprintf("%s to %s", currentStart.Format("2006-01-02"), currentEnd.Format("2006-01-02")),
			ActivityCount: currentStats.ActivityCount,
			Value:         currentFormatted,
			RawValue:      currentValue,
		},
		PreviousPeriod: PeriodMetrics{
			DateRange:     fmt.Sprintf("%s to %s", previousStart.Format("2006-01-02"), previousEnd.Format("2006-01-02")),
			ActivityCount: previousStats.ActivityCount,
			Value:         previousFormatted,
			RawValue:      previousValue,
		},
		Trend:            trend,
		ChangePercent:    changePercent,
		Insights:         insights,
		SuggestedActions: SuggestNextActions("progress"),
	}

	logging.Info("MCP tool completed", "tool", "analyze_progress", "trend", trend, "change_percent", changePercent)
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "analyze_progress", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// checkTrainingLoad analyzes weekly training volume for overtraining detection
func (s *Server) checkTrainingLoad(ctx context.Context, req *mcp.CallToolRequest, input CheckTrainingLoadInput) (*mcp.CallToolResult, CheckTrainingLoadOutput, error) {
	weeks := input.Weeks
	if weeks <= 0 {
		weeks = 4
	}
	if weeks > 12 {
		weeks = 12
	}

	logging.Info("MCP tool call", "tool", "check_training_load", "weeks", weeks, "type", input.Type)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "check_training_load", "input", logging.ToJSON(input))
	}

	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -weeks*7)

	queries := s.queries.(ProgressQuerier)
	hasType := input.Type != ""

	var weeklyData []weeklyVolumeData

	if hasType {
		rows, err := queries.GetWeeklyVolumeByType(ctx, db.GetWeeklyVolumeByTypeParams{
			Type:        sql.NullString{String: input.Type, Valid: true},
			StartDate:   sql.NullTime{Time: startDate, Valid: true},
			StartDate_2: sql.NullTime{Time: endDate, Valid: true},
		})
		if err != nil {
			return nil, CheckTrainingLoadOutput{}, fmt.Errorf("fetching weekly volume: %w", err)
		}
		for _, r := range rows {
			week := ""
			if r.Week != nil {
				week = fmt.Sprintf("%v", r.Week)
			}
			weeklyData = append(weeklyData, weeklyVolumeData{
				Week:          week,
				ActivityCount: r.ActivityCount,
				TotalDistance: toFloat64(r.TotalDistance),
				TotalDuration: toInt64(r.TotalDuration),
				TotalCalories: toFloat64(r.TotalCalories),
				TotalElevation: toFloat64(r.TotalElevation),
			})
		}
	} else {
		rows, err := queries.GetWeeklyVolume(ctx, db.GetWeeklyVolumeParams{
			StartDate:   sql.NullTime{Time: startDate, Valid: true},
			StartDate_2: sql.NullTime{Time: endDate, Valid: true},
		})
		if err != nil {
			return nil, CheckTrainingLoadOutput{}, fmt.Errorf("fetching weekly volume: %w", err)
		}
		for _, r := range rows {
			week := ""
			if r.Week != nil {
				week = fmt.Sprintf("%v", r.Week)
			}
			weeklyData = append(weeklyData, weeklyVolumeData{
				Week:          week,
				ActivityCount: r.ActivityCount,
				TotalDistance: toFloat64(r.TotalDistance),
				TotalDuration: toInt64(r.TotalDuration),
				TotalCalories: toFloat64(r.TotalCalories),
				TotalElevation: toFloat64(r.TotalElevation),
			})
		}
	}

	// Calculate averages and current week
	var totalDistance, totalDuration, totalCalories, totalElevation float64
	var totalActivities int64
	var currentWeek weeklyVolumeData
	recentWeeks := make([]WeeklyLoadSummary, 0)

	for i, w := range weeklyData {
		if i == 0 {
			currentWeek = w
		}
		totalDistance += w.TotalDistance
		totalDuration += float64(w.TotalDuration)
		totalCalories += w.TotalCalories
		totalElevation += w.TotalElevation
		totalActivities += w.ActivityCount

		recentWeeks = append(recentWeeks, WeeklyLoadSummary{
			Week:          w.Week,
			ActivityCount: w.ActivityCount,
			TotalDistance: formatDistance(w.TotalDistance),
			TotalDuration: formatDuration(w.TotalDuration),
			TotalCalories: int(w.TotalCalories),
			TotalElevation: fmt.Sprintf("%.0fm", w.TotalElevation),
		})
	}

	numWeeks := len(weeklyData)
	if numWeeks == 0 {
		numWeeks = 1 // Avoid division by zero
	}

	avgDistance := totalDistance / float64(numWeeks)
	avgDuration := totalDuration / float64(numWeeks)
	avgCalories := totalCalories / float64(numWeeks)
	avgElevation := totalElevation / float64(numWeeks)
	avgActivities := totalActivities / int64(numWeeks)

	// Determine load status based on current week vs average
	var loadStatus string
	var loadChangePercent float64
	if avgDistance > 0 {
		loadChangePercent = ((currentWeek.TotalDistance - avgDistance) / avgDistance) * 100

		if loadChangePercent > 30 {
			loadStatus = "overreaching"
		} else if loadChangePercent > 10 {
			loadStatus = "optimal"
		} else if loadChangePercent > -10 {
			loadStatus = "maintaining"
		} else {
			loadStatus = "undertraining"
		}
	} else {
		loadStatus = "insufficient_data"
	}

	// Generate insights
	generator := NewInsightGenerator()
	insights := generator.GenerateTrainingLoadInsights(
		currentWeek.TotalDistance, avgDistance,
		currentWeek.ActivityCount, avgActivities,
	)

	output := CheckTrainingLoadOutput{
		CurrentWeek: WeeklyLoadSummary{
			Week:          currentWeek.Week,
			ActivityCount: currentWeek.ActivityCount,
			TotalDistance: formatDistance(currentWeek.TotalDistance),
			TotalDuration: formatDuration(currentWeek.TotalDuration),
			TotalCalories: int(currentWeek.TotalCalories),
			TotalElevation: fmt.Sprintf("%.0fm", currentWeek.TotalElevation),
		},
		RecentWeeks: recentWeeks,
		AverageWeekly: WeeklyLoadSummary{
			ActivityCount: avgActivities,
			TotalDistance: formatDistance(avgDistance),
			TotalDuration: formatDuration(int64(avgDuration)),
			TotalCalories: int(avgCalories),
			TotalElevation: fmt.Sprintf("%.0fm", avgElevation),
		},
		LoadStatus:        loadStatus,
		LoadChangePercent: loadChangePercent,
		Insights:          insights,
		SuggestedActions:  SuggestNextActions("week_summary"),
	}

	logging.Info("MCP tool completed", "tool", "check_training_load", "load_status", loadStatus, "change_percent", loadChangePercent)
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "check_training_load", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Helper types
type periodStatsData struct {
	ActivityCount int64
	TotalDistance float64
	TotalDuration int64
	AvgSpeed      float64
}

type weeklyVolumeData struct {
	Week          string
	ActivityCount int64
	TotalDistance float64
	TotalDuration int64
	TotalCalories float64
	TotalElevation float64
}
