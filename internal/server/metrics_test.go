package server

import (
	"context"
	"database/sql"
	"testing"

	"github.com/joshdurbin/strava-mcp/internal/db"
)

// MockMetricsQuerier implements both Querier and MetricsQuerier for testing
type MockMetricsQuerier struct {
	MockQuerier
	caloriesSummary  db.GetCaloriesSummaryRow
	heartrateSummary db.GetHeartrateSummaryRow
	speedSummary     db.GetSpeedSummaryRow
	cadenceSummary   db.GetCadenceSummaryRow
	distanceSummary  db.GetDistanceSummaryRow
	elevationSummary db.GetElevationSummaryRow
	durationSummary  db.GetDurationSummaryRow
}

// Implement new query methods required by the updated Querier interface
func (m *MockMetricsQuerier) CountActivitiesInRange(ctx context.Context, arg db.CountActivitiesInRangeParams) (int64, error) {
	return 0, nil
}

func (m *MockMetricsQuerier) GetActivityTypeSummaryInRange(ctx context.Context, arg db.GetActivityTypeSummaryInRangeParams) ([]db.GetActivityTypeSummaryInRangeRow, error) {
	return nil, nil
}

func (m *MockMetricsQuerier) GetTrainingSummary(ctx context.Context) (db.GetTrainingSummaryRow, error) {
	return db.GetTrainingSummaryRow{}, nil
}

func (m *MockMetricsQuerier) GetTrainingSummaryByType(ctx context.Context, t sql.NullString) (db.GetTrainingSummaryByTypeRow, error) {
	return db.GetTrainingSummaryByTypeRow{}, nil
}

func (m *MockMetricsQuerier) GetTrainingSummaryInRange(ctx context.Context, arg db.GetTrainingSummaryInRangeParams) (db.GetTrainingSummaryInRangeRow, error) {
	return db.GetTrainingSummaryInRangeRow{}, nil
}

func (m *MockMetricsQuerier) GetTrainingSummaryByTypeInRange(ctx context.Context, arg db.GetTrainingSummaryByTypeInRangeParams) (db.GetTrainingSummaryByTypeInRangeRow, error) {
	return db.GetTrainingSummaryByTypeInRangeRow{}, nil
}

func (m *MockMetricsQuerier) GetPeriodStats(ctx context.Context, arg db.GetPeriodStatsParams) (db.GetPeriodStatsRow, error) {
	return db.GetPeriodStatsRow{}, nil
}

func (m *MockMetricsQuerier) GetPeriodStatsByType(ctx context.Context, arg db.GetPeriodStatsByTypeParams) (db.GetPeriodStatsByTypeRow, error) {
	return db.GetPeriodStatsByTypeRow{}, nil
}

func (m *MockMetricsQuerier) GetCaloriesSummary(ctx context.Context) (db.GetCaloriesSummaryRow, error) {
	return m.caloriesSummary, nil
}
func (m *MockMetricsQuerier) GetCaloriesSummaryByType(ctx context.Context, t sql.NullString) (db.GetCaloriesSummaryByTypeRow, error) {
	return db.GetCaloriesSummaryByTypeRow{
		TotalCalories: m.caloriesSummary.TotalCalories,
		AvgCalories:   m.caloriesSummary.AvgCalories,
		MinCalories:   m.caloriesSummary.MinCalories,
		MaxCalories:   m.caloriesSummary.MaxCalories,
		ActivityCount: m.caloriesSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetCaloriesSummaryInRange(ctx context.Context, arg db.GetCaloriesSummaryInRangeParams) (db.GetCaloriesSummaryInRangeRow, error) {
	return db.GetCaloriesSummaryInRangeRow{
		TotalCalories: m.caloriesSummary.TotalCalories,
		AvgCalories:   m.caloriesSummary.AvgCalories,
		MinCalories:   m.caloriesSummary.MinCalories,
		MaxCalories:   m.caloriesSummary.MaxCalories,
		ActivityCount: m.caloriesSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetCaloriesSummaryByTypeInRange(ctx context.Context, arg db.GetCaloriesSummaryByTypeInRangeParams) (db.GetCaloriesSummaryByTypeInRangeRow, error) {
	return db.GetCaloriesSummaryByTypeInRangeRow{
		TotalCalories: m.caloriesSummary.TotalCalories,
		AvgCalories:   m.caloriesSummary.AvgCalories,
		MinCalories:   m.caloriesSummary.MinCalories,
		MaxCalories:   m.caloriesSummary.MaxCalories,
		ActivityCount: m.caloriesSummary.ActivityCount,
	}, nil
}

func (m *MockMetricsQuerier) GetHeartrateSummary(ctx context.Context) (db.GetHeartrateSummaryRow, error) {
	return m.heartrateSummary, nil
}
func (m *MockMetricsQuerier) GetHeartrateSummaryByType(ctx context.Context, t sql.NullString) (db.GetHeartrateSummaryByTypeRow, error) {
	return db.GetHeartrateSummaryByTypeRow{
		AvgHeartrate:        m.heartrateSummary.AvgHeartrate,
		MinAvgHeartrate:     m.heartrateSummary.MinAvgHeartrate,
		MaxAvgHeartrate:     m.heartrateSummary.MaxAvgHeartrate,
		OverallMaxHeartrate: m.heartrateSummary.OverallMaxHeartrate,
		ActivityCount:       m.heartrateSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetHeartrateSummaryInRange(ctx context.Context, arg db.GetHeartrateSummaryInRangeParams) (db.GetHeartrateSummaryInRangeRow, error) {
	return db.GetHeartrateSummaryInRangeRow{
		AvgHeartrate:        m.heartrateSummary.AvgHeartrate,
		MinAvgHeartrate:     m.heartrateSummary.MinAvgHeartrate,
		MaxAvgHeartrate:     m.heartrateSummary.MaxAvgHeartrate,
		OverallMaxHeartrate: m.heartrateSummary.OverallMaxHeartrate,
		ActivityCount:       m.heartrateSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetHeartrateSummaryByTypeInRange(ctx context.Context, arg db.GetHeartrateSummaryByTypeInRangeParams) (db.GetHeartrateSummaryByTypeInRangeRow, error) {
	return db.GetHeartrateSummaryByTypeInRangeRow{
		AvgHeartrate:        m.heartrateSummary.AvgHeartrate,
		MinAvgHeartrate:     m.heartrateSummary.MinAvgHeartrate,
		MaxAvgHeartrate:     m.heartrateSummary.MaxAvgHeartrate,
		OverallMaxHeartrate: m.heartrateSummary.OverallMaxHeartrate,
		ActivityCount:       m.heartrateSummary.ActivityCount,
	}, nil
}

func (m *MockMetricsQuerier) GetSpeedSummary(ctx context.Context) (db.GetSpeedSummaryRow, error) {
	return m.speedSummary, nil
}
func (m *MockMetricsQuerier) GetSpeedSummaryByType(ctx context.Context, t sql.NullString) (db.GetSpeedSummaryByTypeRow, error) {
	return db.GetSpeedSummaryByTypeRow{
		AvgSpeed:        m.speedSummary.AvgSpeed,
		MinAvgSpeed:     m.speedSummary.MinAvgSpeed,
		MaxAvgSpeed:     m.speedSummary.MaxAvgSpeed,
		OverallMaxSpeed: m.speedSummary.OverallMaxSpeed,
		ActivityCount:   m.speedSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetSpeedSummaryInRange(ctx context.Context, arg db.GetSpeedSummaryInRangeParams) (db.GetSpeedSummaryInRangeRow, error) {
	return db.GetSpeedSummaryInRangeRow{
		AvgSpeed:        m.speedSummary.AvgSpeed,
		MinAvgSpeed:     m.speedSummary.MinAvgSpeed,
		MaxAvgSpeed:     m.speedSummary.MaxAvgSpeed,
		OverallMaxSpeed: m.speedSummary.OverallMaxSpeed,
		ActivityCount:   m.speedSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetSpeedSummaryByTypeInRange(ctx context.Context, arg db.GetSpeedSummaryByTypeInRangeParams) (db.GetSpeedSummaryByTypeInRangeRow, error) {
	return db.GetSpeedSummaryByTypeInRangeRow{
		AvgSpeed:        m.speedSummary.AvgSpeed,
		MinAvgSpeed:     m.speedSummary.MinAvgSpeed,
		MaxAvgSpeed:     m.speedSummary.MaxAvgSpeed,
		OverallMaxSpeed: m.speedSummary.OverallMaxSpeed,
		ActivityCount:   m.speedSummary.ActivityCount,
	}, nil
}

func (m *MockMetricsQuerier) GetCadenceSummary(ctx context.Context) (db.GetCadenceSummaryRow, error) {
	return m.cadenceSummary, nil
}
func (m *MockMetricsQuerier) GetCadenceSummaryByType(ctx context.Context, t sql.NullString) (db.GetCadenceSummaryByTypeRow, error) {
	return db.GetCadenceSummaryByTypeRow{
		AvgCadence:    m.cadenceSummary.AvgCadence,
		MinCadence:    m.cadenceSummary.MinCadence,
		MaxCadence:    m.cadenceSummary.MaxCadence,
		ActivityCount: m.cadenceSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetCadenceSummaryInRange(ctx context.Context, arg db.GetCadenceSummaryInRangeParams) (db.GetCadenceSummaryInRangeRow, error) {
	return db.GetCadenceSummaryInRangeRow{
		AvgCadence:    m.cadenceSummary.AvgCadence,
		MinCadence:    m.cadenceSummary.MinCadence,
		MaxCadence:    m.cadenceSummary.MaxCadence,
		ActivityCount: m.cadenceSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetCadenceSummaryByTypeInRange(ctx context.Context, arg db.GetCadenceSummaryByTypeInRangeParams) (db.GetCadenceSummaryByTypeInRangeRow, error) {
	return db.GetCadenceSummaryByTypeInRangeRow{
		AvgCadence:    m.cadenceSummary.AvgCadence,
		MinCadence:    m.cadenceSummary.MinCadence,
		MaxCadence:    m.cadenceSummary.MaxCadence,
		ActivityCount: m.cadenceSummary.ActivityCount,
	}, nil
}

func (m *MockMetricsQuerier) GetDistanceSummary(ctx context.Context) (db.GetDistanceSummaryRow, error) {
	return m.distanceSummary, nil
}
func (m *MockMetricsQuerier) GetDistanceSummaryByType(ctx context.Context, t sql.NullString) (db.GetDistanceSummaryByTypeRow, error) {
	return db.GetDistanceSummaryByTypeRow{
		TotalDistance: m.distanceSummary.TotalDistance,
		AvgDistance:   m.distanceSummary.AvgDistance,
		MinDistance:   m.distanceSummary.MinDistance,
		MaxDistance:   m.distanceSummary.MaxDistance,
		ActivityCount: m.distanceSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetDistanceSummaryInRange(ctx context.Context, arg db.GetDistanceSummaryInRangeParams) (db.GetDistanceSummaryInRangeRow, error) {
	return db.GetDistanceSummaryInRangeRow{
		TotalDistance: m.distanceSummary.TotalDistance,
		AvgDistance:   m.distanceSummary.AvgDistance,
		MinDistance:   m.distanceSummary.MinDistance,
		MaxDistance:   m.distanceSummary.MaxDistance,
		ActivityCount: m.distanceSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetDistanceSummaryByTypeInRange(ctx context.Context, arg db.GetDistanceSummaryByTypeInRangeParams) (db.GetDistanceSummaryByTypeInRangeRow, error) {
	return db.GetDistanceSummaryByTypeInRangeRow{
		TotalDistance: m.distanceSummary.TotalDistance,
		AvgDistance:   m.distanceSummary.AvgDistance,
		MinDistance:   m.distanceSummary.MinDistance,
		MaxDistance:   m.distanceSummary.MaxDistance,
		ActivityCount: m.distanceSummary.ActivityCount,
	}, nil
}

func (m *MockMetricsQuerier) GetElevationSummary(ctx context.Context) (db.GetElevationSummaryRow, error) {
	return m.elevationSummary, nil
}
func (m *MockMetricsQuerier) GetElevationSummaryByType(ctx context.Context, t sql.NullString) (db.GetElevationSummaryByTypeRow, error) {
	return db.GetElevationSummaryByTypeRow{
		TotalElevation: m.elevationSummary.TotalElevation,
		AvgElevation:   m.elevationSummary.AvgElevation,
		MinElevation:   m.elevationSummary.MinElevation,
		MaxElevation:   m.elevationSummary.MaxElevation,
		ActivityCount:  m.elevationSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetElevationSummaryInRange(ctx context.Context, arg db.GetElevationSummaryInRangeParams) (db.GetElevationSummaryInRangeRow, error) {
	return db.GetElevationSummaryInRangeRow{
		TotalElevation: m.elevationSummary.TotalElevation,
		AvgElevation:   m.elevationSummary.AvgElevation,
		MinElevation:   m.elevationSummary.MinElevation,
		MaxElevation:   m.elevationSummary.MaxElevation,
		ActivityCount:  m.elevationSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetElevationSummaryByTypeInRange(ctx context.Context, arg db.GetElevationSummaryByTypeInRangeParams) (db.GetElevationSummaryByTypeInRangeRow, error) {
	return db.GetElevationSummaryByTypeInRangeRow{
		TotalElevation: m.elevationSummary.TotalElevation,
		AvgElevation:   m.elevationSummary.AvgElevation,
		MinElevation:   m.elevationSummary.MinElevation,
		MaxElevation:   m.elevationSummary.MaxElevation,
		ActivityCount:  m.elevationSummary.ActivityCount,
	}, nil
}

func (m *MockMetricsQuerier) GetDurationSummary(ctx context.Context) (db.GetDurationSummaryRow, error) {
	return m.durationSummary, nil
}
func (m *MockMetricsQuerier) GetDurationSummaryByType(ctx context.Context, t sql.NullString) (db.GetDurationSummaryByTypeRow, error) {
	return db.GetDurationSummaryByTypeRow{
		TotalMovingTime: m.durationSummary.TotalMovingTime,
		AvgMovingTime:   m.durationSummary.AvgMovingTime,
		MinMovingTime:   m.durationSummary.MinMovingTime,
		MaxMovingTime:   m.durationSummary.MaxMovingTime,
		ActivityCount:   m.durationSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetDurationSummaryInRange(ctx context.Context, arg db.GetDurationSummaryInRangeParams) (db.GetDurationSummaryInRangeRow, error) {
	return db.GetDurationSummaryInRangeRow{
		TotalMovingTime: m.durationSummary.TotalMovingTime,
		AvgMovingTime:   m.durationSummary.AvgMovingTime,
		MinMovingTime:   m.durationSummary.MinMovingTime,
		MaxMovingTime:   m.durationSummary.MaxMovingTime,
		ActivityCount:   m.durationSummary.ActivityCount,
	}, nil
}
func (m *MockMetricsQuerier) GetDurationSummaryByTypeInRange(ctx context.Context, arg db.GetDurationSummaryByTypeInRangeParams) (db.GetDurationSummaryByTypeInRangeRow, error) {
	return db.GetDurationSummaryByTypeInRangeRow{
		TotalMovingTime: m.durationSummary.TotalMovingTime,
		AvgMovingTime:   m.durationSummary.AvgMovingTime,
		MinMovingTime:   m.durationSummary.MinMovingTime,
		MaxMovingTime:   m.durationSummary.MaxMovingTime,
		ActivityCount:   m.durationSummary.ActivityCount,
	}, nil
}

// Helper function tests

func TestToFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"nil", nil, 0},
		{"float64", float64(123.45), 123.45},
		{"int64", int64(100), 100},
		{"int", int(50), 50},
		{"string", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toFloat64(tt.input)
			if result != tt.expected {
				t.Errorf("toFloat64(%v) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"nil", nil, 0},
		{"int64", int64(100), 100},
		{"float64", float64(123.7), 123},
		{"int", int(50), 50},
		{"string", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toInt64(tt.input)
			if result != tt.expected {
				t.Errorf("toInt64(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDateRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		startDate   string
		endDate     string
		expectError bool
	}{
		{"both dates", "2024-01-01", "2024-12-31", false},
		{"start only", "2024-01-01", "", false},
		{"end only", "", "2024-12-31", false},
		{"neither", "", "", false},
		{"invalid start", "invalid", "2024-12-31", true},
		{"invalid end", "2024-01-01", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, end, err := parseDateRange(tt.startDate, tt.endDate)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.startDate != "" && !start.Valid {
				t.Error("expected valid start date")
			}
			if tt.endDate != "" && !end.Valid {
				t.Error("expected valid end date")
			}
		})
	}
}

func TestBuildFilterDesc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		actType   string
		startDate string
		endDate   string
		contains  string
	}{
		{"all time", "", "", "", "all time"},
		{"type only", "Run", "", "", "type=Run"},
		{"date range", "", "2024-01-01", "2024-12-31", "2024-01-01 to 2024-12-31"},
		{"type and range", "Ride", "2024-01-01", "2024-06-30", "type=Ride"},
		{"start only", "", "2024-01-01", "", "from=2024-01-01"},
		{"end only", "", "", "2024-12-31", "to=2024-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := buildFilterDesc(tt.actType, tt.startDate, tt.endDate)
			if result != tt.contains && !containsSubstring(result, tt.contains) {
				t.Errorf("buildFilterDesc(%q, %q, %q) = %q, expected to contain %q",
					tt.actType, tt.startDate, tt.endDate, result, tt.contains)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Consolidated metrics summary tests

func TestGetMetricsSummaryAllMetrics(t *testing.T) {
	t.Parallel()

	mock := &MockMetricsQuerier{
		caloriesSummary: db.GetCaloriesSummaryRow{
			TotalCalories: float64(5000),
			AvgCalories:   float64(500),
			MinCalories:   float64(100),
			MaxCalories:   float64(1000),
			ActivityCount: 10,
		},
		distanceSummary: db.GetDistanceSummaryRow{
			TotalDistance: float64(100000),
			AvgDistance:   float64(5000),
			MinDistance:   float64(1000),
			MaxDistance:   float64(21000),
			ActivityCount: 20,
		},
		durationSummary: db.GetDurationSummaryRow{
			TotalMovingTime: int64(360000),
			AvgMovingTime:   int64(3600),
			MinMovingTime:   int64(1800),
			MaxMovingTime:   int64(7200),
			ActivityCount:   100,
		},
		speedSummary: db.GetSpeedSummaryRow{
			AvgSpeed:        float64(3.5),
			MinAvgSpeed:     float64(2.0),
			MaxAvgSpeed:     float64(5.0),
			OverallMaxSpeed: float64(8.0),
			ActivityCount:   15,
		},
		heartrateSummary: db.GetHeartrateSummaryRow{
			AvgHeartrate:        float64(145),
			MinAvgHeartrate:     float64(120),
			MaxAvgHeartrate:     float64(165),
			OverallMaxHeartrate: float64(185),
			ActivityCount:       20,
		},
		cadenceSummary: db.GetCadenceSummaryRow{
			AvgCadence:    float64(85),
			MinCadence:    float64(70),
			MaxCadence:    float64(95),
			ActivityCount: 10,
		},
		elevationSummary: db.GetElevationSummaryRow{
			TotalElevation: float64(5000),
			AvgElevation:   float64(250),
			MinElevation:   float64(50),
			MaxElevation:   float64(800),
			ActivityCount:  20,
		},
	}

	srv := New(mock)
	ctx := context.Background()

	// Test with no filters - should return all metrics
	_, output, err := srv.getMetricsSummary(ctx, nil, MetricsSummaryInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Distance == nil {
		t.Error("expected distance metrics")
	}
	if output.Duration == nil {
		t.Error("expected duration metrics")
	}
	if output.Speed == nil {
		t.Error("expected speed metrics")
	}
	if output.Heartrate == nil {
		t.Error("expected heartrate metrics")
	}
	if output.Calories == nil {
		t.Error("expected calories metrics")
	}
	if output.Cadence == nil {
		t.Error("expected cadence metrics")
	}
	if output.Elevation == nil {
		t.Error("expected elevation metrics")
	}
}

func TestGetMetricsSummarySelectiveMetrics(t *testing.T) {
	t.Parallel()

	mock := &MockMetricsQuerier{
		distanceSummary: db.GetDistanceSummaryRow{
			TotalDistance: float64(100000),
			ActivityCount: 20,
		},
		durationSummary: db.GetDurationSummaryRow{
			TotalMovingTime: int64(360000),
			ActivityCount:   100,
		},
	}

	srv := New(mock)
	ctx := context.Background()

	// Test with specific metrics selected
	_, output, err := srv.getMetricsSummary(ctx, nil, MetricsSummaryInput{
		Metrics: []string{"distance", "duration"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Distance == nil {
		t.Error("expected distance metrics")
	}
	if output.Duration == nil {
		t.Error("expected duration metrics")
	}
	// These should NOT be included
	if output.Speed != nil {
		t.Error("did not expect speed metrics")
	}
	if output.Heartrate != nil {
		t.Error("did not expect heartrate metrics")
	}
}

func TestGetMetricsSummaryWithTypeFilter(t *testing.T) {
	t.Parallel()

	mock := &MockMetricsQuerier{
		caloriesSummary: db.GetCaloriesSummaryRow{
			TotalCalories: float64(3000),
			ActivityCount: 5,
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.getMetricsSummary(ctx, nil, MetricsSummaryInput{
		Metrics: []string{"calories"},
		Type:    "Run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Calories == nil {
		t.Fatal("expected calories metrics")
	}
	if output.Calories.Total != 3000 {
		t.Errorf("expected total calories 3000, got %d", output.Calories.Total)
	}
	if output.Filter == "" {
		t.Error("expected non-empty filter description")
	}
}

func TestGetMetricsSummaryWithDateRange(t *testing.T) {
	t.Parallel()

	mock := &MockMetricsQuerier{
		distanceSummary: db.GetDistanceSummaryRow{
			TotalDistance: float64(50000),
			ActivityCount: 10,
		},
	}

	srv := New(mock)
	ctx := context.Background()

	_, output, err := srv.getMetricsSummary(ctx, nil, MetricsSummaryInput{
		Metrics:   []string{"distance"},
		StartDate: "2024-01-01",
		EndDate:   "2024-06-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Distance == nil {
		t.Fatal("expected distance metrics")
	}
	if output.ActivityCount != 10 {
		t.Errorf("expected activity count 10, got %d", output.ActivityCount)
	}
}

func TestGetMetricsSummaryWithInvalidDate(t *testing.T) {
	t.Parallel()

	mock := &MockMetricsQuerier{}
	srv := New(mock)
	ctx := context.Background()

	// Test with invalid start date
	_, _, err := srv.getMetricsSummary(ctx, nil, MetricsSummaryInput{
		Metrics:   []string{"distance"},
		StartDate: "invalid-date",
	})
	if err == nil {
		t.Error("expected error for invalid start date")
	}
}

func TestIsMetricRequested(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metrics  []string
		metric   string
		expected bool
	}{
		{"empty list returns all", []string{}, "distance", true},
		{"metric in list", []string{"distance", "duration"}, "distance", true},
		{"metric not in list", []string{"distance", "duration"}, "calories", false},
		{"single metric match", []string{"speed"}, "speed", true},
		{"single metric no match", []string{"speed"}, "heartrate", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isMetricRequested(tt.metrics, tt.metric)
			if result != tt.expected {
				t.Errorf("isMetricRequested(%v, %q) = %v, want %v", tt.metrics, tt.metric, result, tt.expected)
			}
		})
	}
}
