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

// MetricsQuerier defines the interface for metrics queries
type MetricsQuerier interface {
	GetCaloriesSummary(ctx context.Context) (db.GetCaloriesSummaryRow, error)
	GetCaloriesSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetCaloriesSummaryByTypeRow, error)
	GetCaloriesSummaryInRange(ctx context.Context, arg db.GetCaloriesSummaryInRangeParams) (db.GetCaloriesSummaryInRangeRow, error)
	GetCaloriesSummaryByTypeInRange(ctx context.Context, arg db.GetCaloriesSummaryByTypeInRangeParams) (db.GetCaloriesSummaryByTypeInRangeRow, error)

	GetHeartrateSummary(ctx context.Context) (db.GetHeartrateSummaryRow, error)
	GetHeartrateSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetHeartrateSummaryByTypeRow, error)
	GetHeartrateSummaryInRange(ctx context.Context, arg db.GetHeartrateSummaryInRangeParams) (db.GetHeartrateSummaryInRangeRow, error)
	GetHeartrateSummaryByTypeInRange(ctx context.Context, arg db.GetHeartrateSummaryByTypeInRangeParams) (db.GetHeartrateSummaryByTypeInRangeRow, error)

	GetSpeedSummary(ctx context.Context) (db.GetSpeedSummaryRow, error)
	GetSpeedSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetSpeedSummaryByTypeRow, error)
	GetSpeedSummaryInRange(ctx context.Context, arg db.GetSpeedSummaryInRangeParams) (db.GetSpeedSummaryInRangeRow, error)
	GetSpeedSummaryByTypeInRange(ctx context.Context, arg db.GetSpeedSummaryByTypeInRangeParams) (db.GetSpeedSummaryByTypeInRangeRow, error)

	GetCadenceSummary(ctx context.Context) (db.GetCadenceSummaryRow, error)
	GetCadenceSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetCadenceSummaryByTypeRow, error)
	GetCadenceSummaryInRange(ctx context.Context, arg db.GetCadenceSummaryInRangeParams) (db.GetCadenceSummaryInRangeRow, error)
	GetCadenceSummaryByTypeInRange(ctx context.Context, arg db.GetCadenceSummaryByTypeInRangeParams) (db.GetCadenceSummaryByTypeInRangeRow, error)

	GetDistanceSummary(ctx context.Context) (db.GetDistanceSummaryRow, error)
	GetDistanceSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetDistanceSummaryByTypeRow, error)
	GetDistanceSummaryInRange(ctx context.Context, arg db.GetDistanceSummaryInRangeParams) (db.GetDistanceSummaryInRangeRow, error)
	GetDistanceSummaryByTypeInRange(ctx context.Context, arg db.GetDistanceSummaryByTypeInRangeParams) (db.GetDistanceSummaryByTypeInRangeRow, error)

	GetElevationSummary(ctx context.Context) (db.GetElevationSummaryRow, error)
	GetElevationSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetElevationSummaryByTypeRow, error)
	GetElevationSummaryInRange(ctx context.Context, arg db.GetElevationSummaryInRangeParams) (db.GetElevationSummaryInRangeRow, error)
	GetElevationSummaryByTypeInRange(ctx context.Context, arg db.GetElevationSummaryByTypeInRangeParams) (db.GetElevationSummaryByTypeInRangeRow, error)

	GetDurationSummary(ctx context.Context) (db.GetDurationSummaryRow, error)
	GetDurationSummaryByType(ctx context.Context, activityType sql.NullString) (db.GetDurationSummaryByTypeRow, error)
	GetDurationSummaryInRange(ctx context.Context, arg db.GetDurationSummaryInRangeParams) (db.GetDurationSummaryInRangeRow, error)
	GetDurationSummaryByTypeInRange(ctx context.Context, arg db.GetDurationSummaryByTypeInRangeParams) (db.GetDurationSummaryByTypeInRangeRow, error)
}

// MetricsSummaryInput - input for retrieving detailed metrics summary
type MetricsSummaryInput struct {
	Metrics   []string `json:"metrics,omitempty" jsonschema:"Which metrics to include in the response. Valid values: 'distance', 'duration', 'speed', 'heartrate', 'calories', 'cadence', 'elevation'. Omit or leave empty to include all available metrics."`
	Type      string   `json:"type,omitempty" jsonschema:"Filter metrics to a specific activity type. Common values: Run, Ride, Swim, Walk, Hike. Leave empty to aggregate across all activity types."`
	StartDate string   `json:"start_date,omitempty" jsonschema:"Include activities on or after this date. Format: YYYY-MM-DD. Leave empty for all-time metrics."`
	EndDate   string   `json:"end_date,omitempty" jsonschema:"Include activities on or before this date. Format: YYYY-MM-DD. Leave empty to include up to today."`
}

// Individual metric output types
type DistanceMetrics struct {
	Total   string `json:"total"`
	Average string `json:"average"`
	Min     string `json:"min"`
	Max     string `json:"max"`
}

type DurationMetrics struct {
	Total   string `json:"total"`
	Average string `json:"average"`
	Min     string `json:"min"`
	Max     string `json:"max"`
}

type SpeedMetrics struct {
	AvgPace     string  `json:"avg_pace"`
	FastestPace string  `json:"fastest_pace"`
	SlowestPace string  `json:"slowest_pace"`
	MaxSpeedKmh float64 `json:"max_speed_kmh"`
}

type HeartrateMetrics struct {
	Average    int `json:"average_bpm"`
	MinAverage int `json:"min_avg_bpm"`
	MaxAverage int `json:"max_avg_bpm"`
	OverallMax int `json:"overall_max_bpm"`
}

type CaloriesMetrics struct {
	Total   int `json:"total"`
	Average int `json:"average"`
	Min     int `json:"min"`
	Max     int `json:"max"`
}

type CadenceMetrics struct {
	Average int `json:"average"`
	Min     int `json:"min"`
	Max     int `json:"max"`
}

type ElevationMetrics struct {
	Total   string `json:"total"`
	Average string `json:"average"`
	Min     string `json:"min"`
	Max     string `json:"max"`
}

// Consolidated metrics output type
type MetricsSummaryOutput struct {
	Distance      *DistanceMetrics  `json:"distance,omitempty"`
	Duration      *DurationMetrics  `json:"duration,omitempty"`
	Speed         *SpeedMetrics     `json:"speed,omitempty"`
	Heartrate     *HeartrateMetrics `json:"heartrate,omitempty"`
	Calories      *CaloriesMetrics  `json:"calories,omitempty"`
	Cadence       *CadenceMetrics   `json:"cadence,omitempty"`
	Elevation     *ElevationMetrics `json:"elevation,omitempty"`
	ActivityCount int64             `json:"activity_count"`
	Filter        string            `json:"filter,omitempty"`
}

// registerMetricsTools registers the metrics summary tool
func (s *Server) registerMetricsTools() {
	logging.Debug("Registering tool", "name", "get_metrics_summary")
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_metrics_summary",
		Description: `Get detailed statistics for specific metrics including totals, averages, minimums, and maximums.

Use when:
- User asks about specific metrics like "What's my average heartrate?" or "Total calories burned"
- User wants detailed breakdowns of distance, duration, speed, heartrate, calories, cadence, or elevation
- User needs min/max/average statistics for training analysis

Parameters:
- metrics (array): Which metrics to include: "distance", "duration", "speed", "heartrate", "calories", "cadence", "elevation". Omit for all metrics.
- type (string): Filter by activity type (Run, Ride, Swim, etc.). Leave empty for all types.
- start_date (string): Start date in YYYY-MM-DD format. Leave empty for all time.
- end_date (string): End date in YYYY-MM-DD format. Leave empty for all time.

Returns: For each requested metric - total, average, min, max values plus activity count. Speed returns pace format (min/km).

Example: {"metrics": ["heartrate", "calories"], "type": "Run"} or {"metrics": ["distance", "elevation"]}`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Get Metrics Summary",
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
			DestructiveHint: ptr(false),
		},
	}, s.getMetricsSummary)
}

// Helper to parse date range
func parseDateRange(startDate, endDate string) (sql.NullTime, sql.NullTime, error) {
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

// Helper to build filter description
func buildFilterDesc(activityType, startDate, endDate string) string {
	var parts []string
	if activityType != "" {
		parts = append(parts, "type="+activityType)
	}
	if startDate != "" || endDate != "" {
		if startDate != "" && endDate != "" {
			parts = append(parts, fmt.Sprintf("date=%s to %s", startDate, endDate))
		} else if startDate != "" {
			parts = append(parts, "from="+startDate)
		} else {
			parts = append(parts, "to="+endDate)
		}
	}
	if len(parts) == 0 {
		return "all time"
	}
	return fmt.Sprintf("%v", parts)
}

// Helper to convert interface{} to float64
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case sql.NullFloat64:
		if val.Valid {
			return val.Float64
		}
		return 0
	case sql.NullInt64:
		if val.Valid {
			return float64(val.Int64)
		}
		return 0
	default:
		return 0
	}
}

// Helper to convert interface{} to int64
func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case sql.NullFloat64:
		if val.Valid {
			return int64(val.Float64)
		}
		return 0
	case sql.NullInt64:
		if val.Valid {
			return val.Int64
		}
		return 0
	default:
		return 0
	}
}

// mpsToKmh converts meters per second to km/h
func mpsToKmh(mps float64) float64 {
	return mps * 3.6
}

// mpsToPace converts m/s to min/km pace string
func mpsToPace(mps float64) string {
	if mps <= 0 {
		return "-"
	}
	secPerKm := 1000 / mps
	mins := int(secPerKm) / 60
	secs := int(secPerKm) % 60
	return fmt.Sprintf("%d:%02d/km", mins, secs)
}

// formatDurationHuman converts seconds to human-readable format
func formatDurationHuman(seconds int64) string {
	if seconds <= 0 {
		return "-"
	}
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

// formatDistanceHuman converts meters to human-readable format
func formatDistanceHuman(meters float64) string {
	if meters <= 0 {
		return "-"
	}
	km := meters / 1000
	if km >= 1 {
		return fmt.Sprintf("%.2f km", km)
	}
	return fmt.Sprintf("%.0f m", meters)
}

// formatElevationHuman converts meters to elevation format
func formatElevationHuman(meters float64) string {
	if meters <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.0f m", meters)
}

// Helper to check if a metric is requested
func isMetricRequested(metrics []string, name string) bool {
	if len(metrics) == 0 {
		return true // Return all metrics if none specified
	}
	for _, m := range metrics {
		if m == name {
			return true
		}
	}
	return false
}

// Consolidated metrics summary handler
func (s *Server) getMetricsSummary(ctx context.Context, req *mcp.CallToolRequest, input MetricsSummaryInput) (*mcp.CallToolResult, MetricsSummaryOutput, error) {
	logging.Info("MCP tool call", "tool", "get_metrics_summary", "metrics", input.Metrics, "type", input.Type, "start", input.StartDate, "end", input.EndDate)
	if logging.IsVerbose() {
		logging.Debug("MCP request params", "tool", "get_metrics_summary", "input", logging.ToJSON(input))
	}

	queries := s.queries.(MetricsQuerier)
	filter := buildFilterDesc(input.Type, input.StartDate, input.EndDate)

	hasType := input.Type != ""
	hasDateRange := input.StartDate != "" || input.EndDate != ""

	output := MetricsSummaryOutput{
		Filter: filter,
	}

	var maxActivityCount int64

	// Distance metrics
	if isMetricRequested(input.Metrics, "distance") {
		dist, count, err := s.fetchDistanceMetrics(ctx, queries, input.Type, input.StartDate, input.EndDate, hasType, hasDateRange)
		if err != nil {
			return nil, MetricsSummaryOutput{}, err
		}
		output.Distance = dist
		if count > maxActivityCount {
			maxActivityCount = count
		}
	}

	// Duration metrics
	if isMetricRequested(input.Metrics, "duration") {
		dur, count, err := s.fetchDurationMetrics(ctx, queries, input.Type, input.StartDate, input.EndDate, hasType, hasDateRange)
		if err != nil {
			return nil, MetricsSummaryOutput{}, err
		}
		output.Duration = dur
		if count > maxActivityCount {
			maxActivityCount = count
		}
	}

	// Speed metrics
	if isMetricRequested(input.Metrics, "speed") {
		speed, count, err := s.fetchSpeedMetrics(ctx, queries, input.Type, input.StartDate, input.EndDate, hasType, hasDateRange)
		if err != nil {
			return nil, MetricsSummaryOutput{}, err
		}
		output.Speed = speed
		if count > maxActivityCount {
			maxActivityCount = count
		}
	}

	// Heartrate metrics
	if isMetricRequested(input.Metrics, "heartrate") {
		hr, count, err := s.fetchHeartrateMetrics(ctx, queries, input.Type, input.StartDate, input.EndDate, hasType, hasDateRange)
		if err != nil {
			return nil, MetricsSummaryOutput{}, err
		}
		output.Heartrate = hr
		if count > maxActivityCount {
			maxActivityCount = count
		}
	}

	// Calories metrics
	if isMetricRequested(input.Metrics, "calories") {
		cal, count, err := s.fetchCaloriesMetrics(ctx, queries, input.Type, input.StartDate, input.EndDate, hasType, hasDateRange)
		if err != nil {
			return nil, MetricsSummaryOutput{}, err
		}
		output.Calories = cal
		if count > maxActivityCount {
			maxActivityCount = count
		}
	}

	// Cadence metrics
	if isMetricRequested(input.Metrics, "cadence") {
		cad, count, err := s.fetchCadenceMetrics(ctx, queries, input.Type, input.StartDate, input.EndDate, hasType, hasDateRange)
		if err != nil {
			return nil, MetricsSummaryOutput{}, err
		}
		output.Cadence = cad
		if count > maxActivityCount {
			maxActivityCount = count
		}
	}

	// Elevation metrics
	if isMetricRequested(input.Metrics, "elevation") {
		elev, count, err := s.fetchElevationMetrics(ctx, queries, input.Type, input.StartDate, input.EndDate, hasType, hasDateRange)
		if err != nil {
			return nil, MetricsSummaryOutput{}, err
		}
		output.Elevation = elev
		if count > maxActivityCount {
			maxActivityCount = count
		}
	}

	output.ActivityCount = maxActivityCount

	logging.Info("MCP tool completed", "tool", "get_metrics_summary", "activity_count", output.ActivityCount)
	if logging.IsVerbose() {
		logging.Debug("MCP response", "tool", "get_metrics_summary", "output", logging.ToJSON(output))
	}
	return nil, output, nil
}

// Fetch distance metrics
func (s *Server) fetchDistanceMetrics(ctx context.Context, queries MetricsQuerier, activityType, startDate, endDate string, hasType, hasDateRange bool) (*DistanceMetrics, int64, error) {
	var totalDist, avgDist, minDist, maxDist float64
	var activityCount int64

	if hasType && hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetDistanceSummaryByTypeInRange(ctx, db.GetDistanceSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: activityType, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalDist = toFloat64(row.TotalDistance)
		avgDist = toFloat64(row.AvgDistance)
		minDist = toFloat64(row.MinDistance)
		maxDist = toFloat64(row.MaxDistance)
		activityCount = row.ActivityCount
	} else if hasType {
		row, err := queries.GetDistanceSummaryByType(ctx, sql.NullString{String: activityType, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		totalDist = toFloat64(row.TotalDistance)
		avgDist = toFloat64(row.AvgDistance)
		minDist = toFloat64(row.MinDistance)
		maxDist = toFloat64(row.MaxDistance)
		activityCount = row.ActivityCount
	} else if hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetDistanceSummaryInRange(ctx, db.GetDistanceSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalDist = toFloat64(row.TotalDistance)
		avgDist = toFloat64(row.AvgDistance)
		minDist = toFloat64(row.MinDistance)
		maxDist = toFloat64(row.MaxDistance)
		activityCount = row.ActivityCount
	} else {
		row, err := queries.GetDistanceSummary(ctx)
		if err != nil {
			return nil, 0, err
		}
		totalDist = toFloat64(row.TotalDistance)
		avgDist = toFloat64(row.AvgDistance)
		minDist = toFloat64(row.MinDistance)
		maxDist = toFloat64(row.MaxDistance)
		activityCount = row.ActivityCount
	}

	return &DistanceMetrics{
		Total:   formatDistanceHuman(totalDist),
		Average: formatDistanceHuman(avgDist),
		Min:     formatDistanceHuman(minDist),
		Max:     formatDistanceHuman(maxDist),
	}, activityCount, nil
}

// Fetch duration metrics
func (s *Server) fetchDurationMetrics(ctx context.Context, queries MetricsQuerier, activityType, startDate, endDate string, hasType, hasDateRange bool) (*DurationMetrics, int64, error) {
	var totalDur, avgDur, minDur, maxDur int64
	var activityCount int64

	if hasType && hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetDurationSummaryByTypeInRange(ctx, db.GetDurationSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: activityType, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalDur = toInt64(row.TotalMovingTime)
		avgDur = toInt64(row.AvgMovingTime)
		minDur = toInt64(row.MinMovingTime)
		maxDur = toInt64(row.MaxMovingTime)
		activityCount = row.ActivityCount
	} else if hasType {
		row, err := queries.GetDurationSummaryByType(ctx, sql.NullString{String: activityType, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		totalDur = toInt64(row.TotalMovingTime)
		avgDur = toInt64(row.AvgMovingTime)
		minDur = toInt64(row.MinMovingTime)
		maxDur = toInt64(row.MaxMovingTime)
		activityCount = row.ActivityCount
	} else if hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetDurationSummaryInRange(ctx, db.GetDurationSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalDur = toInt64(row.TotalMovingTime)
		avgDur = toInt64(row.AvgMovingTime)
		minDur = toInt64(row.MinMovingTime)
		maxDur = toInt64(row.MaxMovingTime)
		activityCount = row.ActivityCount
	} else {
		row, err := queries.GetDurationSummary(ctx)
		if err != nil {
			return nil, 0, err
		}
		totalDur = toInt64(row.TotalMovingTime)
		avgDur = toInt64(row.AvgMovingTime)
		minDur = toInt64(row.MinMovingTime)
		maxDur = toInt64(row.MaxMovingTime)
		activityCount = row.ActivityCount
	}

	return &DurationMetrics{
		Total:   formatDurationHuman(totalDur),
		Average: formatDurationHuman(avgDur),
		Min:     formatDurationHuman(minDur),
		Max:     formatDurationHuman(maxDur),
	}, activityCount, nil
}

// Fetch speed metrics
func (s *Server) fetchSpeedMetrics(ctx context.Context, queries MetricsQuerier, activityType, startDate, endDate string, hasType, hasDateRange bool) (*SpeedMetrics, int64, error) {
	var avgSpeed, minSpeed, maxSpeed, overallMax float64
	var activityCount int64

	if hasType && hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetSpeedSummaryByTypeInRange(ctx, db.GetSpeedSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: activityType, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		avgSpeed = toFloat64(row.AvgSpeed)
		minSpeed = toFloat64(row.MinAvgSpeed)
		maxSpeed = toFloat64(row.MaxAvgSpeed)
		overallMax = toFloat64(row.OverallMaxSpeed)
		activityCount = row.ActivityCount
	} else if hasType {
		row, err := queries.GetSpeedSummaryByType(ctx, sql.NullString{String: activityType, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		avgSpeed = toFloat64(row.AvgSpeed)
		minSpeed = toFloat64(row.MinAvgSpeed)
		maxSpeed = toFloat64(row.MaxAvgSpeed)
		overallMax = toFloat64(row.OverallMaxSpeed)
		activityCount = row.ActivityCount
	} else if hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetSpeedSummaryInRange(ctx, db.GetSpeedSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		avgSpeed = toFloat64(row.AvgSpeed)
		minSpeed = toFloat64(row.MinAvgSpeed)
		maxSpeed = toFloat64(row.MaxAvgSpeed)
		overallMax = toFloat64(row.OverallMaxSpeed)
		activityCount = row.ActivityCount
	} else {
		row, err := queries.GetSpeedSummary(ctx)
		if err != nil {
			return nil, 0, err
		}
		avgSpeed = toFloat64(row.AvgSpeed)
		minSpeed = toFloat64(row.MinAvgSpeed)
		maxSpeed = toFloat64(row.MaxAvgSpeed)
		overallMax = toFloat64(row.OverallMaxSpeed)
		activityCount = row.ActivityCount
	}

	return &SpeedMetrics{
		AvgPace:     mpsToPace(avgSpeed),
		FastestPace: mpsToPace(maxSpeed),
		SlowestPace: mpsToPace(minSpeed),
		MaxSpeedKmh: mpsToKmh(overallMax),
	}, activityCount, nil
}

// Fetch heartrate metrics
func (s *Server) fetchHeartrateMetrics(ctx context.Context, queries MetricsQuerier, activityType, startDate, endDate string, hasType, hasDateRange bool) (*HeartrateMetrics, int64, error) {
	var avgHR, minAvgHR, maxAvgHR, overallMaxHR float64
	var activityCount int64

	if hasType && hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetHeartrateSummaryByTypeInRange(ctx, db.GetHeartrateSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: activityType, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		avgHR = toFloat64(row.AvgHeartrate)
		minAvgHR = toFloat64(row.MinAvgHeartrate)
		maxAvgHR = toFloat64(row.MaxAvgHeartrate)
		overallMaxHR = toFloat64(row.OverallMaxHeartrate)
		activityCount = row.ActivityCount
	} else if hasType {
		row, err := queries.GetHeartrateSummaryByType(ctx, sql.NullString{String: activityType, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		avgHR = toFloat64(row.AvgHeartrate)
		minAvgHR = toFloat64(row.MinAvgHeartrate)
		maxAvgHR = toFloat64(row.MaxAvgHeartrate)
		overallMaxHR = toFloat64(row.OverallMaxHeartrate)
		activityCount = row.ActivityCount
	} else if hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetHeartrateSummaryInRange(ctx, db.GetHeartrateSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		avgHR = toFloat64(row.AvgHeartrate)
		minAvgHR = toFloat64(row.MinAvgHeartrate)
		maxAvgHR = toFloat64(row.MaxAvgHeartrate)
		overallMaxHR = toFloat64(row.OverallMaxHeartrate)
		activityCount = row.ActivityCount
	} else {
		row, err := queries.GetHeartrateSummary(ctx)
		if err != nil {
			return nil, 0, err
		}
		avgHR = toFloat64(row.AvgHeartrate)
		minAvgHR = toFloat64(row.MinAvgHeartrate)
		maxAvgHR = toFloat64(row.MaxAvgHeartrate)
		overallMaxHR = toFloat64(row.OverallMaxHeartrate)
		activityCount = row.ActivityCount
	}

	return &HeartrateMetrics{
		Average:    int(avgHR),
		MinAverage: int(minAvgHR),
		MaxAverage: int(maxAvgHR),
		OverallMax: int(overallMaxHR),
	}, activityCount, nil
}

// Fetch calories metrics
func (s *Server) fetchCaloriesMetrics(ctx context.Context, queries MetricsQuerier, activityType, startDate, endDate string, hasType, hasDateRange bool) (*CaloriesMetrics, int64, error) {
	var totalCal, avgCal, minCal, maxCal float64
	var activityCount int64

	if hasType && hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetCaloriesSummaryByTypeInRange(ctx, db.GetCaloriesSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: activityType, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalCal = toFloat64(row.TotalCalories)
		avgCal = toFloat64(row.AvgCalories)
		minCal = toFloat64(row.MinCalories)
		maxCal = toFloat64(row.MaxCalories)
		activityCount = row.ActivityCount
	} else if hasType {
		row, err := queries.GetCaloriesSummaryByType(ctx, sql.NullString{String: activityType, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		totalCal = toFloat64(row.TotalCalories)
		avgCal = toFloat64(row.AvgCalories)
		minCal = toFloat64(row.MinCalories)
		maxCal = toFloat64(row.MaxCalories)
		activityCount = row.ActivityCount
	} else if hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetCaloriesSummaryInRange(ctx, db.GetCaloriesSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalCal = toFloat64(row.TotalCalories)
		avgCal = toFloat64(row.AvgCalories)
		minCal = toFloat64(row.MinCalories)
		maxCal = toFloat64(row.MaxCalories)
		activityCount = row.ActivityCount
	} else {
		row, err := queries.GetCaloriesSummary(ctx)
		if err != nil {
			return nil, 0, err
		}
		totalCal = toFloat64(row.TotalCalories)
		avgCal = toFloat64(row.AvgCalories)
		minCal = toFloat64(row.MinCalories)
		maxCal = toFloat64(row.MaxCalories)
		activityCount = row.ActivityCount
	}

	return &CaloriesMetrics{
		Total:   int(totalCal),
		Average: int(avgCal),
		Min:     int(minCal),
		Max:     int(maxCal),
	}, activityCount, nil
}

// Fetch cadence metrics
func (s *Server) fetchCadenceMetrics(ctx context.Context, queries MetricsQuerier, activityType, startDate, endDate string, hasType, hasDateRange bool) (*CadenceMetrics, int64, error) {
	var avgCad, minCad, maxCad float64
	var activityCount int64

	if hasType && hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetCadenceSummaryByTypeInRange(ctx, db.GetCadenceSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: activityType, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		avgCad = toFloat64(row.AvgCadence)
		minCad = toFloat64(row.MinCadence)
		maxCad = toFloat64(row.MaxCadence)
		activityCount = row.ActivityCount
	} else if hasType {
		row, err := queries.GetCadenceSummaryByType(ctx, sql.NullString{String: activityType, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		avgCad = toFloat64(row.AvgCadence)
		minCad = toFloat64(row.MinCadence)
		maxCad = toFloat64(row.MaxCadence)
		activityCount = row.ActivityCount
	} else if hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetCadenceSummaryInRange(ctx, db.GetCadenceSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		avgCad = toFloat64(row.AvgCadence)
		minCad = toFloat64(row.MinCadence)
		maxCad = toFloat64(row.MaxCadence)
		activityCount = row.ActivityCount
	} else {
		row, err := queries.GetCadenceSummary(ctx)
		if err != nil {
			return nil, 0, err
		}
		avgCad = toFloat64(row.AvgCadence)
		minCad = toFloat64(row.MinCadence)
		maxCad = toFloat64(row.MaxCadence)
		activityCount = row.ActivityCount
	}

	return &CadenceMetrics{
		Average: int(avgCad),
		Min:     int(minCad),
		Max:     int(maxCad),
	}, activityCount, nil
}

// Fetch elevation metrics
func (s *Server) fetchElevationMetrics(ctx context.Context, queries MetricsQuerier, activityType, startDate, endDate string, hasType, hasDateRange bool) (*ElevationMetrics, int64, error) {
	var totalElev, avgElev, minElev, maxElev float64
	var activityCount int64

	if hasType && hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetElevationSummaryByTypeInRange(ctx, db.GetElevationSummaryByTypeInRangeParams{
			Type:        sql.NullString{String: activityType, Valid: true},
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalElev = toFloat64(row.TotalElevation)
		avgElev = toFloat64(row.AvgElevation)
		minElev = toFloat64(row.MinElevation)
		maxElev = toFloat64(row.MaxElevation)
		activityCount = row.ActivityCount
	} else if hasType {
		row, err := queries.GetElevationSummaryByType(ctx, sql.NullString{String: activityType, Valid: true})
		if err != nil {
			return nil, 0, err
		}
		totalElev = toFloat64(row.TotalElevation)
		avgElev = toFloat64(row.AvgElevation)
		minElev = toFloat64(row.MinElevation)
		maxElev = toFloat64(row.MaxElevation)
		activityCount = row.ActivityCount
	} else if hasDateRange {
		start, end, err := parseDateRange(startDate, endDate)
		if err != nil {
			return nil, 0, err
		}
		row, err := queries.GetElevationSummaryInRange(ctx, db.GetElevationSummaryInRangeParams{
			StartDate:   start,
			StartDate_2: end,
		})
		if err != nil {
			return nil, 0, err
		}
		totalElev = toFloat64(row.TotalElevation)
		avgElev = toFloat64(row.AvgElevation)
		minElev = toFloat64(row.MinElevation)
		maxElev = toFloat64(row.MaxElevation)
		activityCount = row.ActivityCount
	} else {
		row, err := queries.GetElevationSummary(ctx)
		if err != nil {
			return nil, 0, err
		}
		totalElev = toFloat64(row.TotalElevation)
		avgElev = toFloat64(row.AvgElevation)
		minElev = toFloat64(row.MinElevation)
		maxElev = toFloat64(row.MaxElevation)
		activityCount = row.ActivityCount
	}

	return &ElevationMetrics{
		Total:   formatElevationHuman(totalElev),
		Average: formatElevationHuman(avgElev),
		Min:     formatElevationHuman(minElev),
		Max:     formatElevationHuman(maxElev),
	}, activityCount, nil
}
