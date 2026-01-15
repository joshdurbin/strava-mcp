package sync

import (
	"database/sql"
	"testing"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/strava"
)

func TestConvertActivityToParams(t *testing.T) {
	startDate := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	startDateLocal := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

	activity := strava.Activity{
		ID:                 12345,
		Name:               "Morning Run",
		Distance:           5000.5,
		MovingTime:         1800,
		ElapsedTime:        2000,
		TotalElevationGain: 50.5,
		Type:               "Run",
		SportType:          "Run",
		StartDate:          startDate,
		StartDateLocal:     startDateLocal,
		Timezone:           "(GMT+01:00) Europe/Paris",
		AverageSpeed:       2.78,
		MaxSpeed:           4.5,
		AverageCadence:     85.5,
		AverageHeartrate:   145.0,
		MaxHeartrate:       175.0,
		Kilojoules:         350.0,
	}

	params := ConvertActivityToParams(activity)

	if params.ID != 12345 {
		t.Errorf("expected ID 12345, got %d", params.ID)
	}
	if params.Name != "Morning Run" {
		t.Errorf("expected name 'Morning Run', got '%s'", params.Name)
	}
	if !params.Distance.Valid || params.Distance.Float64 != 5000.5 {
		t.Errorf("expected distance 5000.5, got %+v", params.Distance)
	}
	if !params.MovingTime.Valid || params.MovingTime.Int64 != 1800 {
		t.Errorf("expected moving time 1800, got %+v", params.MovingTime)
	}
	if !params.Type.Valid || params.Type.String != "Run" {
		t.Errorf("expected type 'Run', got %+v", params.Type)
	}
	if !params.StartDate.Valid || !params.StartDate.Time.Equal(startDate) {
		t.Errorf("expected start date %v, got %+v", startDate, params.StartDate)
	}
	if !params.Calories.Valid || params.Calories.Float64 != 350.0 {
		t.Errorf("expected calories 350.0, got %+v", params.Calories)
	}
}

func TestConvertActivityToParams_ZeroValues(t *testing.T) {
	activity := strava.Activity{
		ID:   12345,
		Name: "Test Activity",
	}

	params := ConvertActivityToParams(activity)

	if params.ID != 12345 {
		t.Errorf("expected ID 12345, got %d", params.ID)
	}
	if params.Name != "Test Activity" {
		t.Errorf("expected name 'Test Activity', got '%s'", params.Name)
	}
	if params.Distance.Valid {
		t.Error("expected distance to be invalid for zero value")
	}
	if params.MovingTime.Valid {
		t.Error("expected moving time to be invalid for zero value")
	}
	if params.Type.Valid {
		t.Error("expected type to be invalid for empty string")
	}
	if params.StartDate.Valid {
		t.Error("expected start date to be invalid for zero time")
	}
}

func TestToNullFloat64(t *testing.T) {
	tests := []struct {
		input    float64
		expected sql.NullFloat64
	}{
		{0, sql.NullFloat64{Float64: 0, Valid: false}},
		{123.45, sql.NullFloat64{Float64: 123.45, Valid: true}},
		{-50.5, sql.NullFloat64{Float64: -50.5, Valid: true}},
	}

	for _, tt := range tests {
		result := toNullFloat64(tt.input)
		if result != tt.expected {
			t.Errorf("toNullFloat64(%v) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

func TestToNullInt64(t *testing.T) {
	tests := []struct {
		input    int64
		expected sql.NullInt64
	}{
		{0, sql.NullInt64{Int64: 0, Valid: false}},
		{100, sql.NullInt64{Int64: 100, Valid: true}},
		{-50, sql.NullInt64{Int64: -50, Valid: true}},
	}

	for _, tt := range tests {
		result := toNullInt64(tt.input)
		if result != tt.expected {
			t.Errorf("toNullInt64(%v) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

func TestToNullString(t *testing.T) {
	tests := []struct {
		input    string
		expected sql.NullString
	}{
		{"", sql.NullString{String: "", Valid: false}},
		{"Run", sql.NullString{String: "Run", Valid: true}},
	}

	for _, tt := range tests {
		result := toNullString(tt.input)
		if result != tt.expected {
			t.Errorf("toNullString(%q) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

func TestToNullTime(t *testing.T) {
	validTime := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	zeroTime := time.Time{}

	result := toNullTime(validTime)
	if !result.Valid || !result.Time.Equal(validTime) {
		t.Errorf("expected valid time %v, got %+v", validTime, result)
	}

	result = toNullTime(zeroTime)
	if result.Valid {
		t.Error("expected invalid time for zero value")
	}
}
