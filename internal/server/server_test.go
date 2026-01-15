package server

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
)

// MockQuerier implements the Querier interface for testing
type MockQuerier struct {
	activities          []db.Activity
	activityTypeSummary []db.GetActivityTypeSummaryRow
	countByMonth        []db.GetActivityCountsByMonthRow
	countByWeek         []db.GetActivityCountsByWeekRow
}

func (m *MockQuerier) GetActivity(ctx context.Context, id int64) (db.Activity, error) {
	for _, a := range m.activities {
		if a.ID == id {
			return a, nil
		}
	}
	return db.Activity{}, sql.ErrNoRows
}

func (m *MockQuerier) GetAllActivities(ctx context.Context) ([]db.Activity, error) {
	return m.activities, nil
}

func (m *MockQuerier) CountActivities(ctx context.Context) (int64, error) {
	return int64(len(m.activities)), nil
}

func (m *MockQuerier) CountActivitiesByType(ctx context.Context, activityType sql.NullString) (int64, error) {
	var count int64
	for _, a := range m.activities {
		if a.Type.Valid && a.Type.String == activityType.String {
			count++
		}
	}
	return count, nil
}

func (m *MockQuerier) GetOldestActivity(ctx context.Context) (db.Activity, error) {
	if len(m.activities) == 0 {
		return db.Activity{}, sql.ErrNoRows
	}
	oldest := m.activities[0]
	for _, a := range m.activities[1:] {
		if a.StartDate.Valid && oldest.StartDate.Valid && a.StartDate.Time.Before(oldest.StartDate.Time) {
			oldest = a
		}
	}
	return oldest, nil
}

func (m *MockQuerier) GetLatestActivity(ctx context.Context) (db.Activity, error) {
	if len(m.activities) == 0 {
		return db.Activity{}, sql.ErrNoRows
	}
	latest := m.activities[0]
	for _, a := range m.activities[1:] {
		if a.StartDate.Valid && latest.StartDate.Valid && a.StartDate.Time.After(latest.StartDate.Time) {
			latest = a
		}
	}
	return latest, nil
}

func (m *MockQuerier) GetActivityTypeSummary(ctx context.Context) ([]db.GetActivityTypeSummaryRow, error) {
	return m.activityTypeSummary, nil
}

func (m *MockQuerier) CountActivitiesByTypeInRange(ctx context.Context, arg db.CountActivitiesByTypeInRangeParams) (int64, error) {
	var count int64
	for _, a := range m.activities {
		if a.Type.Valid && a.Type.String == arg.Type.String {
			if a.StartDate.Valid && a.StartDate.Time.After(arg.StartDate.Time) && a.StartDate.Time.Before(arg.StartDate_2.Time) {
				count++
			}
		}
	}
	return count, nil
}

func (m *MockQuerier) GetActivityCountsByMonth(ctx context.Context, arg db.GetActivityCountsByMonthParams) ([]db.GetActivityCountsByMonthRow, error) {
	return m.countByMonth, nil
}

func (m *MockQuerier) GetActivityCountsByWeek(ctx context.Context, arg db.GetActivityCountsByWeekParams) ([]db.GetActivityCountsByWeekRow, error) {
	return m.countByWeek, nil
}

// Zone-related stub methods for interface compliance
func (m *MockQuerier) GetActivityZones(ctx context.Context, activityID int64) ([]db.ActivityZone, error) {
	return nil, nil
}
func (m *MockQuerier) GetZoneBuckets(ctx context.Context, activityZoneID int64) ([]db.ZoneBucket, error) {
	return nil, nil
}
func (m *MockQuerier) GetActivitiesWithZones(ctx context.Context, limit int64) ([]db.GetActivitiesWithZonesRow, error) {
	return nil, nil
}
func (m *MockQuerier) CountActivitiesWithZones(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountActivitiesWithoutZones(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) GetHeartRateZoneSummary(ctx context.Context) ([]db.GetHeartRateZoneSummaryRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetHeartRateZoneSummaryByType(ctx context.Context, activityType sql.NullString) ([]db.GetHeartRateZoneSummaryByTypeRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetHeartRateZoneSummaryInRange(ctx context.Context, arg db.GetHeartRateZoneSummaryInRangeParams) ([]db.GetHeartRateZoneSummaryInRangeRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetHeartRateZoneSummaryByTypeInRange(ctx context.Context, arg db.GetHeartRateZoneSummaryByTypeInRangeParams) ([]db.GetHeartRateZoneSummaryByTypeInRangeRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetPowerZoneSummary(ctx context.Context) ([]db.GetPowerZoneSummaryRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetPowerZoneSummaryByType(ctx context.Context, activityType sql.NullString) ([]db.GetPowerZoneSummaryByTypeRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetPowerZoneSummaryInRange(ctx context.Context, arg db.GetPowerZoneSummaryInRangeParams) ([]db.GetPowerZoneSummaryInRangeRow, error) {
	return nil, nil
}

// New interface methods
func (m *MockQuerier) CountActivitiesInRange(ctx context.Context, arg db.CountActivitiesInRangeParams) (int64, error) {
	return int64(len(m.activities)), nil
}

func (m *MockQuerier) GetActivityTypeSummaryInRange(ctx context.Context, arg db.GetActivityTypeSummaryInRangeParams) ([]db.GetActivityTypeSummaryInRangeRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetTrainingSummary(ctx context.Context) (db.GetTrainingSummaryRow, error) {
	return db.GetTrainingSummaryRow{}, nil
}

func (m *MockQuerier) GetTrainingSummaryByType(ctx context.Context, t sql.NullString) (db.GetTrainingSummaryByTypeRow, error) {
	return db.GetTrainingSummaryByTypeRow{}, nil
}

func (m *MockQuerier) GetTrainingSummaryInRange(ctx context.Context, arg db.GetTrainingSummaryInRangeParams) (db.GetTrainingSummaryInRangeRow, error) {
	return db.GetTrainingSummaryInRangeRow{}, nil
}

func (m *MockQuerier) GetTrainingSummaryByTypeInRange(ctx context.Context, arg db.GetTrainingSummaryByTypeInRangeParams) (db.GetTrainingSummaryByTypeInRangeRow, error) {
	return db.GetTrainingSummaryByTypeInRangeRow{}, nil
}

func (m *MockQuerier) GetPeriodStats(ctx context.Context, arg db.GetPeriodStatsParams) (db.GetPeriodStatsRow, error) {
	return db.GetPeriodStatsRow{}, nil
}

func (m *MockQuerier) GetPeriodStatsByType(ctx context.Context, arg db.GetPeriodStatsByTypeParams) (db.GetPeriodStatsByTypeRow, error) {
	return db.GetPeriodStatsByTypeRow{}, nil
}

// Personal records methods
func (m *MockQuerier) GetFastestActivity(ctx context.Context) (db.Activity, error) {
	if len(m.activities) == 0 {
		return db.Activity{}, sql.ErrNoRows
	}
	fastest := m.activities[0]
	for _, a := range m.activities[1:] {
		if a.AverageSpeed.Valid && fastest.AverageSpeed.Valid && a.AverageSpeed.Float64 > fastest.AverageSpeed.Float64 {
			fastest = a
		}
	}
	return fastest, nil
}

func (m *MockQuerier) GetFastestActivityByType(ctx context.Context, t sql.NullString) (db.Activity, error) {
	var fastest db.Activity
	found := false
	for _, a := range m.activities {
		if a.Type.Valid && a.Type.String == t.String {
			if !found || (a.AverageSpeed.Valid && fastest.AverageSpeed.Valid && a.AverageSpeed.Float64 > fastest.AverageSpeed.Float64) {
				fastest = a
				found = true
			}
		}
	}
	if !found {
		return db.Activity{}, sql.ErrNoRows
	}
	return fastest, nil
}

func (m *MockQuerier) GetLongestDistanceActivity(ctx context.Context) (db.Activity, error) {
	if len(m.activities) == 0 {
		return db.Activity{}, sql.ErrNoRows
	}
	longest := m.activities[0]
	for _, a := range m.activities[1:] {
		if a.Distance.Valid && longest.Distance.Valid && a.Distance.Float64 > longest.Distance.Float64 {
			longest = a
		}
	}
	return longest, nil
}

func (m *MockQuerier) GetLongestDistanceActivityByType(ctx context.Context, t sql.NullString) (db.Activity, error) {
	var longest db.Activity
	found := false
	for _, a := range m.activities {
		if a.Type.Valid && a.Type.String == t.String {
			if !found || (a.Distance.Valid && longest.Distance.Valid && a.Distance.Float64 > longest.Distance.Float64) {
				longest = a
				found = true
			}
		}
	}
	if !found {
		return db.Activity{}, sql.ErrNoRows
	}
	return longest, nil
}

func (m *MockQuerier) GetLongestDurationActivity(ctx context.Context) (db.Activity, error) {
	if len(m.activities) == 0 {
		return db.Activity{}, sql.ErrNoRows
	}
	longest := m.activities[0]
	for _, a := range m.activities[1:] {
		if a.MovingTime.Valid && longest.MovingTime.Valid && a.MovingTime.Int64 > longest.MovingTime.Int64 {
			longest = a
		}
	}
	return longest, nil
}

func (m *MockQuerier) GetLongestDurationActivityByType(ctx context.Context, t sql.NullString) (db.Activity, error) {
	var longest db.Activity
	found := false
	for _, a := range m.activities {
		if a.Type.Valid && a.Type.String == t.String {
			if !found || (a.MovingTime.Valid && longest.MovingTime.Valid && a.MovingTime.Int64 > longest.MovingTime.Int64) {
				longest = a
				found = true
			}
		}
	}
	if !found {
		return db.Activity{}, sql.ErrNoRows
	}
	return longest, nil
}

func (m *MockQuerier) GetHighestElevationActivity(ctx context.Context) (db.Activity, error) {
	if len(m.activities) == 0 {
		return db.Activity{}, sql.ErrNoRows
	}
	highest := m.activities[0]
	for _, a := range m.activities[1:] {
		if a.TotalElevationGain.Valid && highest.TotalElevationGain.Valid && a.TotalElevationGain.Float64 > highest.TotalElevationGain.Float64 {
			highest = a
		}
	}
	return highest, nil
}

func (m *MockQuerier) GetHighestElevationActivityByType(ctx context.Context, t sql.NullString) (db.Activity, error) {
	var highest db.Activity
	found := false
	for _, a := range m.activities {
		if a.Type.Valid && a.Type.String == t.String {
			if !found || (a.TotalElevationGain.Valid && highest.TotalElevationGain.Valid && a.TotalElevationGain.Float64 > highest.TotalElevationGain.Float64) {
				highest = a
				found = true
			}
		}
	}
	if !found {
		return db.Activity{}, sql.ErrNoRows
	}
	return highest, nil
}

func (m *MockQuerier) GetMostCaloriesActivity(ctx context.Context) (db.Activity, error) {
	if len(m.activities) == 0 {
		return db.Activity{}, sql.ErrNoRows
	}
	most := m.activities[0]
	for _, a := range m.activities[1:] {
		if a.Calories.Valid && most.Calories.Valid && a.Calories.Float64 > most.Calories.Float64 {
			most = a
		}
	}
	return most, nil
}

func (m *MockQuerier) GetMostCaloriesActivityByType(ctx context.Context, t sql.NullString) (db.Activity, error) {
	var most db.Activity
	found := false
	for _, a := range m.activities {
		if a.Type.Valid && a.Type.String == t.String {
			if !found || (a.Calories.Valid && most.Calories.Valid && a.Calories.Float64 > most.Calories.Float64) {
				most = a
				found = true
			}
		}
	}
	if !found {
		return db.Activity{}, sql.ErrNoRows
	}
	return most, nil
}

// Weekly volume methods
func (m *MockQuerier) GetWeeklyVolume(ctx context.Context, arg db.GetWeeklyVolumeParams) ([]db.GetWeeklyVolumeRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetWeeklyVolumeByType(ctx context.Context, arg db.GetWeeklyVolumeByTypeParams) ([]db.GetWeeklyVolumeByTypeRow, error) {
	return nil, nil
}

// Search methods
func (m *MockQuerier) SearchActivities(ctx context.Context, arg db.SearchActivitiesParams) ([]db.Activity, error) {
	var result []db.Activity
	for _, a := range m.activities {
		// Basic filtering
		if arg.Type.Valid && (!a.Type.Valid || a.Type.String != arg.Type.String) {
			continue
		}
		result = append(result, a)
		if int64(len(result)) >= arg.Limit {
			break
		}
	}
	return result, nil
}

func (m *MockQuerier) SearchActivitiesByDistance(ctx context.Context, arg db.SearchActivitiesByDistanceParams) ([]db.Activity, error) {
	return m.activities, nil
}

func (m *MockQuerier) SearchActivitiesByDuration(ctx context.Context, arg db.SearchActivitiesByDurationParams) ([]db.Activity, error) {
	return m.activities, nil
}

func (m *MockQuerier) SearchActivitiesBySpeed(ctx context.Context, arg db.SearchActivitiesBySpeedParams) ([]db.Activity, error) {
	return m.activities, nil
}

func (m *MockQuerier) SearchActivitiesByElevation(ctx context.Context, arg db.SearchActivitiesByElevationParams) ([]db.Activity, error) {
	return m.activities, nil
}

// Test helpers
func createTestActivity(id int64, name, activityType string, date time.Time) db.Activity {
	return db.Activity{
		ID:        id,
		Name:      name,
		Type:      sql.NullString{String: activityType, Valid: true},
		StartDate: sql.NullTime{Time: date, Valid: true},
		Distance:  sql.NullFloat64{Float64: 5000, Valid: true},
	}
}

func TestConvertActivity(t *testing.T) {
	t.Parallel()

	activity := db.Activity{
		ID:               123,
		Name:             "Morning Run",
		Type:             sql.NullString{String: "Run", Valid: true},
		SportType:        sql.NullString{String: "Run", Valid: true},
		Distance:         sql.NullFloat64{Float64: 5000, Valid: true},
		MovingTime:       sql.NullInt64{Int64: 1800, Valid: true},
		AverageSpeed:     sql.NullFloat64{Float64: 2.78, Valid: true},
		MaxSpeed:         sql.NullFloat64{Float64: 3.5, Valid: true},
		AverageHeartrate: sql.NullFloat64{Float64: 150, Valid: true},
		MaxHeartrate:     sql.NullFloat64{Float64: 175, Valid: true},
		StartDate:        sql.NullTime{Time: time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC), Valid: true},
	}

	summary := convertActivity(activity)

	if summary.ID != 123 {
		t.Errorf("expected ID 123, got %d", summary.ID)
	}
	if summary.Name != "Morning Run" {
		t.Errorf("expected name 'Morning Run', got %q", summary.Name)
	}
	if summary.Type != "Run" {
		t.Errorf("expected type 'Run', got %q", summary.Type)
	}
	if summary.Distance != "5.00 km" {
		t.Errorf("expected distance '5.00 km', got %q", summary.Distance)
	}
	if summary.Duration != "30m 0s" {
		t.Errorf("expected duration '30m 0s', got %q", summary.Duration)
	}
	if summary.AvgHeartrate != 150 {
		t.Errorf("expected avg heartrate 150, got %d", summary.AvgHeartrate)
	}
}

func TestConvertActivityNullFields(t *testing.T) {
	t.Parallel()

	activity := db.Activity{
		ID:   456,
		Name: "Unnamed Activity",
		// All other fields are null
	}

	summary := convertActivity(activity)

	if summary.ID != 456 {
		t.Errorf("expected ID 456, got %d", summary.ID)
	}
	if summary.Type != "" {
		t.Errorf("expected empty type, got %q", summary.Type)
	}
	if summary.Distance != "" {
		t.Errorf("expected empty distance, got %q", summary.Distance)
	}
}

func TestConvertActivities(t *testing.T) {
	t.Parallel()

	activities := []db.Activity{
		createTestActivity(1, "Run 1", "Run", time.Now()),
		createTestActivity(2, "Ride 1", "Ride", time.Now()),
	}

	summaries := convertActivities(activities)

	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ID != 1 {
		t.Errorf("expected first activity ID 1, got %d", summaries[0].ID)
	}
	if summaries[1].ID != 2 {
		t.Errorf("expected second activity ID 2, got %d", summaries[1].ID)
	}
}

func TestServerNew(t *testing.T) {
	t.Parallel()

	mock := &MockQuerier{}
	srv := New(mock)

	if srv == nil {
		t.Fatal("expected non-nil server")
	}
	if srv.mcp == nil {
		t.Error("expected non-nil MCP server")
	}
	if srv.queries == nil {
		t.Error("expected non-nil queries")
	}
}

func TestServerMCPServer(t *testing.T) {
	t.Parallel()

	mock := &MockQuerier{}
	srv := New(mock)

	mcpServer := srv.MCPServer()
	if mcpServer == nil {
		t.Error("expected non-nil MCP server from MCPServer()")
	}
	if mcpServer != srv.mcp {
		t.Error("expected MCPServer() to return the internal mcp server")
	}
}

func TestFormatDistance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		meters   float64
		expected string
	}{
		{5000, "5.00 km"},
		{10500, "10.50 km"},
		{500, "500 m"},
		{0, "0 m"},
	}

	for _, tc := range tests {
		result := formatDistance(tc.meters)
		if result != tc.expected {
			t.Errorf("formatDistance(%f): expected %q, got %q", tc.meters, tc.expected, result)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		seconds  int64
		expected string
	}{
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{1800, "30m 0s"},
		{90, "1m 30s"},
		{45, "45s"},
		{0, "0s"},
	}

	for _, tc := range tests {
		result := formatDuration(tc.seconds)
		if result != tc.expected {
			t.Errorf("formatDuration(%d): expected %q, got %q", tc.seconds, tc.expected, result)
		}
	}
}

func TestFormatPace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mps      float64
		expected string
	}{
		{2.78, "5:59/km"},  // ~6 min/km pace
		{3.33, "5:00/km"},  // 5 min/km pace
		{4.0, "4:10/km"},   // 4 m/s = 4:10/km pace
		{0, ""},
		{-1, ""},
	}

	for _, tc := range tests {
		result := formatPace(tc.mps)
		if result != tc.expected {
			t.Errorf("formatPace(%f): expected %q, got %q", tc.mps, tc.expected, result)
		}
	}
}

func TestInsightGenerator(t *testing.T) {
	t.Parallel()

	gen := NewInsightGenerator()

	// Test progress insights
	insights := gen.GenerateProgressInsights(100, 90, "pace", true)
	if len(insights) == 0 {
		t.Error("expected at least one insight")
	}

	// Test training load insights
	loadInsights := gen.GenerateTrainingLoadInsights(100, 80, 5, 4)
	if len(loadInsights) == 0 {
		t.Error("expected at least one training load insight")
	}

	// Test zone insights
	zonePercentages := map[int]float64{
		1: 10,
		2: 50,
		3: 20,
		4: 15,
		5: 5,
	}
	zoneInsights := gen.GenerateZoneInsights(zonePercentages)
	if len(zoneInsights) == 0 {
		t.Error("expected at least one zone insight")
	}
}
