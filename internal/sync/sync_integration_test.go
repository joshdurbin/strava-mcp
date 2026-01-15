//go:build integration

package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/strava"

	_ "modernc.org/sqlite"
)

func TestSyncIntegration(t *testing.T) {
	// Create test activities that the mock server will return
	testActivities := []strava.Activity{
		{
			ID:                 1,
			Name:               "Morning Run",
			Distance:           5000,
			MovingTime:         1800,
			ElapsedTime:        2000,
			TotalElevationGain: 50,
			Type:               "Run",
			SportType:          "Run",
			StartDate:          time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
			StartDateLocal:     time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
			Timezone:           "(GMT+01:00) Europe/Paris",
			AverageSpeed:       2.78,
			MaxSpeed:           4.5,
			AverageHeartrate:   145,
			MaxHeartrate:       175,
			Kilojoules:         350,
		},
		{
			ID:                 2,
			Name:               "Evening Ride",
			Distance:           25000,
			MovingTime:         3600,
			ElapsedTime:        4000,
			TotalElevationGain: 200,
			Type:               "Ride",
			SportType:          "Ride",
			StartDate:          time.Date(2024, 1, 14, 17, 0, 0, 0, time.UTC),
			StartDateLocal:     time.Date(2024, 1, 14, 18, 0, 0, 0, time.UTC),
			Timezone:           "(GMT+01:00) Europe/Paris",
			AverageSpeed:       6.94,
			MaxSpeed:           12.0,
			Kilojoules:         800,
		},
	}

	// Create mock Strava API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "1" {
			json.NewEncoder(w).Encode(testActivities)
		} else {
			json.NewEncoder(w).Encode([]strava.Activity{})
		}
	}))
	defer server.Close()

	// Create temporary database
	tmpFile, err := os.CreateTemp("", "strava_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Open database connection
	sqlDB, err := sql.Open("sqlite", tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer sqlDB.Close()

	// Run schema
	schema := `
		CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			distance REAL,
			moving_time INTEGER,
			elapsed_time INTEGER,
			total_elevation_gain REAL,
			type TEXT,
			sport_type TEXT,
			start_date DATETIME,
			start_date_local DATETIME,
			timezone TEXT,
			average_speed REAL,
			max_speed REAL,
			average_cadence REAL,
			average_heartrate REAL,
			max_heartrate REAL,
			calories REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Create queries and client
	queries := db.New(sqlDB)
	client := strava.NewClientWithBaseURL("test-token", server.URL)

	// Create sync service
	service := NewService(queries, client)

	// Run sync
	ctx := context.Background()
	var fetchCalls, saveCalls int
	err = service.Sync(ctx,
		func(page, activitiesFetched, totalSoFar int) {
			fetchCalls++
		},
		func(current, total int, activityName string) {
			saveCalls++
		},
	)

	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// Verify fetch callback was called
	if fetchCalls != 1 {
		t.Errorf("expected 1 fetch callback, got %d", fetchCalls)
	}

	// Verify save callback was called for each activity
	if saveCalls != 2 {
		t.Errorf("expected 2 save callbacks, got %d", saveCalls)
	}

	// Verify activities in database
	count, err := queries.CountActivities(ctx)
	if err != nil {
		t.Fatalf("failed to count activities: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 activities in database, got %d", count)
	}

	// Verify specific activity
	activity, err := queries.GetActivity(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get activity: %v", err)
	}
	if activity.Name != "Morning Run" {
		t.Errorf("expected activity name 'Morning Run', got '%s'", activity.Name)
	}
}
