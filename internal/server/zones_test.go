package server

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
)

// MockZonesQuerier implements ZonesQuerier for testing
type MockZonesQuerier struct {
	MockQuerier
	activityZones       []db.ActivityZone
	zoneBuckets         map[int64][]db.ZoneBucket
	activitiesWithZones []db.GetActivitiesWithZonesRow
	hrZoneSummary       []db.GetHeartRateZoneSummaryRow
	powerZoneSummary    []db.GetPowerZoneSummaryRow
}

func (m *MockZonesQuerier) GetActivityZones(ctx context.Context, activityID int64) ([]db.ActivityZone, error) {
	var result []db.ActivityZone
	for _, az := range m.activityZones {
		if az.ActivityID == activityID {
			result = append(result, az)
		}
	}
	return result, nil
}

func (m *MockZonesQuerier) GetActivityZoneByActivityAndType(ctx context.Context, arg db.GetActivityZoneByActivityAndTypeParams) (db.ActivityZone, error) {
	for _, az := range m.activityZones {
		if az.ActivityID == arg.ActivityID && az.ZoneType == arg.ZoneType {
			return az, nil
		}
	}
	return db.ActivityZone{}, sql.ErrNoRows
}

func (m *MockZonesQuerier) GetZoneBuckets(ctx context.Context, activityZoneID int64) ([]db.ZoneBucket, error) {
	if buckets, ok := m.zoneBuckets[activityZoneID]; ok {
		return buckets, nil
	}
	return []db.ZoneBucket{}, nil
}

func (m *MockZonesQuerier) GetActivitiesWithZones(ctx context.Context, limit int64) ([]db.GetActivitiesWithZonesRow, error) {
	if int64(len(m.activitiesWithZones)) <= limit {
		return m.activitiesWithZones, nil
	}
	return m.activitiesWithZones[:limit], nil
}

func (m *MockZonesQuerier) CountActivitiesWithZones(ctx context.Context) (int64, error) {
	return int64(len(m.activitiesWithZones)), nil
}

func (m *MockZonesQuerier) CountActivitiesWithoutZones(ctx context.Context) (int64, error) {
	return int64(len(m.activities)) - int64(len(m.activitiesWithZones)), nil
}

func (m *MockZonesQuerier) GetHeartRateZoneSummary(ctx context.Context) ([]db.GetHeartRateZoneSummaryRow, error) {
	return m.hrZoneSummary, nil
}

func (m *MockZonesQuerier) GetHeartRateZoneSummaryByType(ctx context.Context, activityType sql.NullString) ([]db.GetHeartRateZoneSummaryByTypeRow, error) {
	var result []db.GetHeartRateZoneSummaryByTypeRow
	for _, r := range m.hrZoneSummary {
		result = append(result, db.GetHeartRateZoneSummaryByTypeRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     r.TotalTime,
			AvgTime:       r.AvgTime,
			ActivityCount: r.ActivityCount,
		})
	}
	return result, nil
}

func (m *MockZonesQuerier) GetHeartRateZoneSummaryInRange(ctx context.Context, arg db.GetHeartRateZoneSummaryInRangeParams) ([]db.GetHeartRateZoneSummaryInRangeRow, error) {
	var result []db.GetHeartRateZoneSummaryInRangeRow
	for _, r := range m.hrZoneSummary {
		result = append(result, db.GetHeartRateZoneSummaryInRangeRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     r.TotalTime,
			AvgTime:       r.AvgTime,
			ActivityCount: r.ActivityCount,
		})
	}
	return result, nil
}

func (m *MockZonesQuerier) GetHeartRateZoneSummaryByTypeInRange(ctx context.Context, arg db.GetHeartRateZoneSummaryByTypeInRangeParams) ([]db.GetHeartRateZoneSummaryByTypeInRangeRow, error) {
	var result []db.GetHeartRateZoneSummaryByTypeInRangeRow
	for _, r := range m.hrZoneSummary {
		result = append(result, db.GetHeartRateZoneSummaryByTypeInRangeRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     r.TotalTime,
			AvgTime:       r.AvgTime,
			ActivityCount: r.ActivityCount,
		})
	}
	return result, nil
}

func (m *MockZonesQuerier) GetPowerZoneSummary(ctx context.Context) ([]db.GetPowerZoneSummaryRow, error) {
	return m.powerZoneSummary, nil
}

func (m *MockZonesQuerier) GetPowerZoneSummaryByType(ctx context.Context, activityType sql.NullString) ([]db.GetPowerZoneSummaryByTypeRow, error) {
	var result []db.GetPowerZoneSummaryByTypeRow
	for _, r := range m.powerZoneSummary {
		result = append(result, db.GetPowerZoneSummaryByTypeRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     r.TotalTime,
			AvgTime:       r.AvgTime,
			ActivityCount: r.ActivityCount,
		})
	}
	return result, nil
}

func (m *MockZonesQuerier) GetPowerZoneSummaryInRange(ctx context.Context, arg db.GetPowerZoneSummaryInRangeParams) ([]db.GetPowerZoneSummaryInRangeRow, error) {
	var result []db.GetPowerZoneSummaryInRangeRow
	for _, r := range m.powerZoneSummary {
		result = append(result, db.GetPowerZoneSummaryInRangeRow{
			ZoneNumber:    r.ZoneNumber,
			TotalTime:     r.TotalTime,
			AvgTime:       r.AvgTime,
			ActivityCount: r.ActivityCount,
		})
	}
	return result, nil
}

// Test GetActivityZones tool
func TestGetActivityZones(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		MockQuerier: MockQuerier{
			activities: []db.Activity{
				{
					ID:        123,
					Name:      "Morning Run",
					Type:      sql.NullString{String: "Run", Valid: true},
					StartDate: sql.NullTime{Time: time.Now(), Valid: true},
				},
			},
		},
		activityZones: []db.ActivityZone{
			{ID: 1, ActivityID: 123, ZoneType: "heartrate", SensorBased: 1},
			{ID: 2, ActivityID: 123, ZoneType: "power", SensorBased: 0},
		},
		zoneBuckets: map[int64][]db.ZoneBucket{
			1: {
				{ID: 1, ActivityZoneID: 1, ZoneNumber: 1, MinValue: 0, MaxValue: 120, TimeSeconds: 600},
				{ID: 2, ActivityZoneID: 1, ZoneNumber: 2, MinValue: 120, MaxValue: 140, TimeSeconds: 900},
				{ID: 3, ActivityZoneID: 1, ZoneNumber: 3, MinValue: 140, MaxValue: 160, TimeSeconds: 300},
			},
			2: {
				{ID: 4, ActivityZoneID: 2, ZoneNumber: 1, MinValue: 0, MaxValue: 150, TimeSeconds: 400},
				{ID: 5, ActivityZoneID: 2, ZoneNumber: 2, MinValue: 150, MaxValue: 200, TimeSeconds: 500},
			},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.getActivityZones(ctx, nil, GetActivityZonesInput{ActivityID: 123})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ActivityID != 123 {
		t.Errorf("expected activity ID 123, got %d", output.ActivityID)
	}

	if len(output.Zones) != 2 {
		t.Errorf("expected 2 zone types, got %d", len(output.Zones))
	}

	// Check heartrate zone
	found := false
	for _, z := range output.Zones {
		if z.Type == "heartrate" {
			found = true
			if len(z.Buckets) != 3 {
				t.Errorf("expected 3 heartrate buckets, got %d", len(z.Buckets))
			}
			if !z.SensorBased {
				t.Error("expected sensor based to be true for heartrate")
			}
		}
	}
	if !found {
		t.Error("heartrate zone not found")
	}
}

func TestGetActivityZonesNotFound(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		MockQuerier: MockQuerier{
			activities: []db.Activity{},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	// Implementation returns empty result for non-existent activity (no error)
	_, output, err := srv.getActivityZones(ctx, nil, GetActivityZonesInput{ActivityID: 999})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(output.Zones) != 0 {
		t.Errorf("expected 0 zones for non-existent activity, got %d", len(output.Zones))
	}
}

func TestGetActivityZonesNoZoneData(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		MockQuerier: MockQuerier{
			activities: []db.Activity{
				{
					ID:        123,
					Name:      "Morning Run",
					Type:      sql.NullString{String: "Run", Valid: true},
					StartDate: sql.NullTime{Time: time.Now(), Valid: true},
				},
			},
		},
		activityZones: []db.ActivityZone{}, // No zones
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.getActivityZones(ctx, nil, GetActivityZonesInput{ActivityID: 123})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Zones) != 0 {
		t.Errorf("expected 0 zones, got %d", len(output.Zones))
	}

	// No insights when there's no zone data (insights only generated from zone percentages)
	// SuggestedActions are always added though
	if len(output.SuggestedActions) == 0 {
		t.Error("expected suggested actions")
	}
}

// Test AnalyzeZones tool
func TestAnalyzeZonesHeartrate(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		hrZoneSummary: []db.GetHeartRateZoneSummaryRow{
			{ZoneNumber: 1, TotalTime: sql.NullFloat64{Float64: 3600, Valid: true}, AvgTime: sql.NullFloat64{Float64: 600, Valid: true}, ActivityCount: 6},
			{ZoneNumber: 2, TotalTime: sql.NullFloat64{Float64: 7200, Valid: true}, AvgTime: sql.NullFloat64{Float64: 1200, Valid: true}, ActivityCount: 6},
			{ZoneNumber: 3, TotalTime: sql.NullFloat64{Float64: 5400, Valid: true}, AvgTime: sql.NullFloat64{Float64: 900, Valid: true}, ActivityCount: 6},
			{ZoneNumber: 4, TotalTime: sql.NullFloat64{Float64: 1800, Valid: true}, AvgTime: sql.NullFloat64{Float64: 300, Valid: true}, ActivityCount: 6},
			{ZoneNumber: 5, TotalTime: sql.NullFloat64{Float64: 900, Valid: true}, AvgTime: sql.NullFloat64{Float64: 150, Valid: true}, ActivityCount: 6},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{ZoneType: "heartrate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ZoneType != "heartrate" {
		t.Errorf("expected zone type 'heartrate', got %q", output.ZoneType)
	}

	if len(output.Zones) != 5 {
		t.Errorf("expected 5 zones, got %d", len(output.Zones))
	}

	// Should have insights
	if len(output.Insights) == 0 {
		t.Error("expected insights")
	}

	// Should have suggested actions
	if len(output.SuggestedActions) == 0 {
		t.Error("expected suggested actions")
	}
}

func TestAnalyzeZonesPower(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		powerZoneSummary: []db.GetPowerZoneSummaryRow{
			{ZoneNumber: 1, TotalTime: sql.NullFloat64{Float64: 1800, Valid: true}, AvgTime: sql.NullFloat64{Float64: 300, Valid: true}, ActivityCount: 6},
			{ZoneNumber: 2, TotalTime: sql.NullFloat64{Float64: 3600, Valid: true}, AvgTime: sql.NullFloat64{Float64: 600, Valid: true}, ActivityCount: 6},
			{ZoneNumber: 3, TotalTime: sql.NullFloat64{Float64: 2400, Valid: true}, AvgTime: sql.NullFloat64{Float64: 400, Valid: true}, ActivityCount: 6},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{ZoneType: "power"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ZoneType != "power" {
		t.Errorf("expected zone type 'power', got %q", output.ZoneType)
	}

	if len(output.Zones) != 3 {
		t.Errorf("expected 3 zones, got %d", len(output.Zones))
	}
}

func TestAnalyzeZonesDefaultsToHeartrate(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		hrZoneSummary: []db.GetHeartRateZoneSummaryRow{
			{ZoneNumber: 1, TotalTime: sql.NullFloat64{Float64: 1800, Valid: true}, AvgTime: sql.NullFloat64{Float64: 300, Valid: true}, ActivityCount: 6},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	// Empty zone type should default to heartrate
	_, output, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ZoneType != "heartrate" {
		t.Errorf("expected zone type 'heartrate' as default, got %q", output.ZoneType)
	}
}

func TestAnalyzeZonesWithTypeFilter(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		hrZoneSummary: []db.GetHeartRateZoneSummaryRow{
			{ZoneNumber: 1, TotalTime: sql.NullFloat64{Float64: 1800, Valid: true}, AvgTime: sql.NullFloat64{Float64: 300, Valid: true}, ActivityCount: 3},
			{ZoneNumber: 2, TotalTime: sql.NullFloat64{Float64: 3600, Valid: true}, AvgTime: sql.NullFloat64{Float64: 600, Valid: true}, ActivityCount: 3},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{
		ZoneType: "heartrate",
		Type:     "Run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Filter should contain type info
	if output.Filter == "" {
		t.Error("expected non-empty filter with type")
	}
}

func TestAnalyzeZonesWithDateRange(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		hrZoneSummary: []db.GetHeartRateZoneSummaryRow{
			{ZoneNumber: 1, TotalTime: sql.NullFloat64{Float64: 1200, Valid: true}, AvgTime: sql.NullFloat64{Float64: 400, Valid: true}, ActivityCount: 3},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{
		ZoneType:  "heartrate",
		StartDate: "2024-01-01",
		EndDate:   "2024-06-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Zones) != 1 {
		t.Errorf("expected 1 zone, got %d", len(output.Zones))
	}
}

func TestAnalyzeZonesInvalidDateRange(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{}
	srv := New(mock)
	ctx := context.Background()

	_, _, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{
		ZoneType:  "heartrate",
		StartDate: "invalid-date",
	})
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestAnalyzeZonesNoData(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		hrZoneSummary: []db.GetHeartRateZoneSummaryRow{}, // No data
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{ZoneType: "heartrate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Zones) != 0 {
		t.Errorf("expected 0 zones, got %d", len(output.Zones))
	}
}

// Test zone percentage calculations
func TestZonePercentages(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		hrZoneSummary: []db.GetHeartRateZoneSummaryRow{
			{ZoneNumber: 1, TotalTime: sql.NullFloat64{Float64: 1000, Valid: true}, AvgTime: sql.NullFloat64{Float64: 200, Valid: true}, ActivityCount: 5},
			{ZoneNumber: 2, TotalTime: sql.NullFloat64{Float64: 2000, Valid: true}, AvgTime: sql.NullFloat64{Float64: 400, Valid: true}, ActivityCount: 5},
			{ZoneNumber: 3, TotalTime: sql.NullFloat64{Float64: 2000, Valid: true}, AvgTime: sql.NullFloat64{Float64: 400, Valid: true}, ActivityCount: 5},
			{ZoneNumber: 4, TotalTime: sql.NullFloat64{Float64: 4000, Valid: true}, AvgTime: sql.NullFloat64{Float64: 800, Valid: true}, ActivityCount: 5},
			{ZoneNumber: 5, TotalTime: sql.NullFloat64{Float64: 1000, Valid: true}, AvgTime: sql.NullFloat64{Float64: 200, Valid: true}, ActivityCount: 5},
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.analyzeZones(ctx, nil, AnalyzeZonesInput{ZoneType: "heartrate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total should be 10000 seconds
	// Zone 1: 10%, Zone 2: 20%, Zone 3: 20%, Zone 4: 40%, Zone 5: 10%
	expectedPercentages := map[int]float64{
		1: 10.0,
		2: 20.0,
		3: 20.0,
		4: 40.0,
		5: 10.0,
	}

	for _, z := range output.Zones {
		expected := expectedPercentages[z.Zone]
		if z.Percentage < expected-0.1 || z.Percentage > expected+0.1 {
			t.Errorf("zone %d: expected percentage ~%.1f%%, got %.1f%%", z.Zone, expected, z.Percentage)
		}
	}
}

// Test GetActivityZoneByActivityAndType
func TestGetActivityZoneByActivityAndType(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		activityZones: []db.ActivityZone{
			{ID: 1, ActivityID: 123, ZoneType: "heartrate", SensorBased: 1},
			{ID: 2, ActivityID: 123, ZoneType: "power", SensorBased: 0},
			{ID: 3, ActivityID: 456, ZoneType: "heartrate", SensorBased: 1},
		},
	}

	ctx := context.Background()

	// Test finding heartrate zone for activity 123
	zone, err := mock.GetActivityZoneByActivityAndType(ctx, db.GetActivityZoneByActivityAndTypeParams{
		ActivityID: 123,
		ZoneType:   "heartrate",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zone.ID != 1 {
		t.Errorf("expected zone ID 1, got %d", zone.ID)
	}

	// Test finding power zone for activity 123
	zone, err = mock.GetActivityZoneByActivityAndType(ctx, db.GetActivityZoneByActivityAndTypeParams{
		ActivityID: 123,
		ZoneType:   "power",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zone.ID != 2 {
		t.Errorf("expected zone ID 2, got %d", zone.ID)
	}

	// Test not found
	_, err = mock.GetActivityZoneByActivityAndType(ctx, db.GetActivityZoneByActivityAndTypeParams{
		ActivityID: 999,
		ZoneType:   "heartrate",
	})
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

// Test CountActivitiesWithZones and CountActivitiesWithoutZones
func TestCountActivitiesWithAndWithoutZones(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		MockQuerier: MockQuerier{
			activities: []db.Activity{
				{ID: 1, Name: "Activity 1"},
				{ID: 2, Name: "Activity 2"},
				{ID: 3, Name: "Activity 3"},
				{ID: 4, Name: "Activity 4"},
				{ID: 5, Name: "Activity 5"},
			},
		},
		activitiesWithZones: []db.GetActivitiesWithZonesRow{
			{ID: 1, Name: "Activity 1"},
			{ID: 2, Name: "Activity 2"},
			{ID: 3, Name: "Activity 3"},
		},
	}

	ctx := context.Background()

	withZones, err := mock.CountActivitiesWithZones(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withZones != 3 {
		t.Errorf("expected 3 activities with zones, got %d", withZones)
	}

	withoutZones, err := mock.CountActivitiesWithoutZones(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withoutZones != 2 {
		t.Errorf("expected 2 activities without zones, got %d", withoutZones)
	}
}

// Test GetActivitiesWithZones
func TestGetActivitiesWithZones(t *testing.T) {
	t.Parallel()

	mock := &MockZonesQuerier{
		activitiesWithZones: []db.GetActivitiesWithZonesRow{
			{ID: 1, Name: "Activity 1", Type: sql.NullString{String: "Run", Valid: true}},
			{ID: 2, Name: "Activity 2", Type: sql.NullString{String: "Ride", Valid: true}},
			{ID: 3, Name: "Activity 3", Type: sql.NullString{String: "Run", Valid: true}},
		},
	}

	ctx := context.Background()

	// Get all
	activities, err := mock.GetActivitiesWithZones(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(activities) != 3 {
		t.Errorf("expected 3 activities, got %d", len(activities))
	}

	// Get limited
	activities, err = mock.GetActivitiesWithZones(ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(activities) != 2 {
		t.Errorf("expected 2 activities, got %d", len(activities))
	}
}

// Test formatZoneTime helper
func TestFormatZoneTime(t *testing.T) {
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
		{7200, "2h 0m"},
	}

	for _, tc := range tests {
		result := formatDuration(tc.seconds)
		if result != tc.expected {
			t.Errorf("formatDuration(%d): expected %q, got %q", tc.seconds, tc.expected, result)
		}
	}
}
