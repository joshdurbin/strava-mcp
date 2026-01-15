package workers

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/strava"
	_ "modernc.org/sqlite"
)

func TestFormatDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: "unknown",
		},
		{
			name:     "string",
			input:    "2024-01-15T10:30:00Z",
			expected: "2024-01-15T10:30:00Z",
		},
		{
			name:     "byte slice",
			input:    []byte("2024-01-15T10:30:00Z"),
			expected: "2024-01-15T10:30:00Z",
		},
		{
			name:     "time.Time",
			input:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expected: "2024-01-15T10:30:00Z",
		},
		{
			name:     "sql.NullTime valid",
			input:    sql.NullTime{Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), Valid: true},
			expected: "2024-01-15T10:30:00Z",
		},
		{
			name:     "sql.NullTime invalid",
			input:    sql.NullTime{Valid: false},
			expected: "unknown",
		},
		{
			name:     "unknown type",
			input:    123,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatDate(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewTokenRefresher(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher(nil, 30*time.Minute)

	if refresher.interval != 30*time.Minute {
		t.Errorf("expected interval 30m, got %v", refresher.interval)
	}

	if refresher.storage != nil {
		t.Errorf("expected nil storage, got %v", refresher.storage)
	}
}

func TestNewActivitySyncer(t *testing.T) {
	t.Parallel()

	retryConfig := strava.DefaultRetryConfig()
	syncer := NewActivitySyncer(nil, nil, 15*time.Minute, retryConfig)

	if syncer.interval != 15*time.Minute {
		t.Errorf("expected interval 15m, got %v", syncer.interval)
	}

	if syncer.retryConfig.MaxRetries != retryConfig.MaxRetries {
		t.Errorf("expected retry max %d, got %d", retryConfig.MaxRetries, syncer.retryConfig.MaxRetries)
	}
}

func TestNewZoneSyncer(t *testing.T) {
	t.Parallel()

	retryConfig := strava.DefaultRetryConfig()
	syncer := NewZoneSyncer(nil, nil, 1*time.Hour, retryConfig)

	if syncer.interval != 1*time.Hour {
		t.Errorf("expected interval 1h, got %v", syncer.interval)
	}

	if syncer.batchSize != 25 {
		t.Errorf("expected batch size 25, got %d", syncer.batchSize)
	}
}

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) (*db.Queries, *sql.DB, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "strava-mcp-workers-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open db: %v", err)
	}

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
	CREATE TABLE IF NOT EXISTS activity_zones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		activity_id INTEGER NOT NULL,
		zone_type TEXT NOT NULL,
		sensor_based BOOLEAN,
		UNIQUE(activity_id, zone_type)
	);
	`
	if _, err := sqlDB.Exec(schema); err != nil {
		sqlDB.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create schema: %v", err)
	}

	queries := db.New(sqlDB)
	cleanup := func() {
		sqlDB.Close()
		os.RemoveAll(tmpDir)
	}

	return queries, sqlDB, cleanup
}

func TestGetLatestActivityDate(t *testing.T) {
	t.Parallel()

	queries, sqlDB, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Empty database should return nil
	result, err := queries.GetLatestActivityDate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty db, got %v", result)
	}

	// Insert activities with different dates
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, -3, 0), // oldest
		now.AddDate(0, -2, 0),
		now.AddDate(0, -1, 0), // newest
	}

	for i, d := range dates {
		_, err := sqlDB.Exec(
			"INSERT INTO activities (id, name, start_date) VALUES (?, ?, ?)",
			i+1, "Activity", d,
		)
		if err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
	}

	result, err = queries.GetLatestActivityDate(ctx)
	if err != nil {
		t.Fatalf("failed to get latest date: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestGetOldestActivityDate(t *testing.T) {
	t.Parallel()

	queries, sqlDB, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Empty database
	result, err := queries.GetOldestActivityDate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty db, got %v", result)
	}

	// Insert activities
	now := time.Now()
	for i := 1; i <= 3; i++ {
		_, err := sqlDB.Exec(
			"INSERT INTO activities (id, name, start_date) VALUES (?, ?, ?)",
			i, "Activity", now.AddDate(0, -i, 0),
		)
		if err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
	}

	result, err = queries.GetOldestActivityDate(ctx)
	if err != nil {
		t.Fatalf("failed to get oldest date: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestGetRecentActivities(t *testing.T) {
	t.Parallel()

	queries, sqlDB, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert 10 activities
	now := time.Now()
	for i := 1; i <= 10; i++ {
		_, err := sqlDB.Exec(
			"INSERT INTO activities (id, name, start_date) VALUES (?, ?, ?)",
			i, "Activity", now.AddDate(0, 0, -i),
		)
		if err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
	}

	// Get 5 most recent
	activities, err := queries.GetRecentActivities(ctx, 5)
	if err != nil {
		t.Fatalf("failed to get recent: %v", err)
	}
	if len(activities) != 5 {
		t.Errorf("expected 5 activities, got %d", len(activities))
	}

	// Most recent should be first (ID=1 has most recent date)
	if activities[0].ID != 1 {
		t.Errorf("expected ID=1 first, got %d", activities[0].ID)
	}
}

func TestGetAllActivityIDs(t *testing.T) {
	t.Parallel()

	queries, sqlDB, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Empty
	ids, err := queries.GetAllActivityIDs(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 IDs, got %d", len(ids))
	}

	// Insert activities
	for i := int64(100); i <= 105; i++ {
		_, err := sqlDB.Exec("INSERT INTO activities (id, name) VALUES (?, ?)", i, "Activity")
		if err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
	}

	ids, err = queries.GetAllActivityIDs(ctx)
	if err != nil {
		t.Fatalf("failed to get IDs: %v", err)
	}
	if len(ids) != 6 {
		t.Errorf("expected 6 IDs, got %d", len(ids))
	}
}

func TestCountActivitiesWithZones(t *testing.T) {
	t.Parallel()

	queries, sqlDB, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert 5 activities
	for i := 1; i <= 5; i++ {
		_, err := sqlDB.Exec("INSERT INTO activities (id, name) VALUES (?, ?)", i, "Activity")
		if err != nil {
			t.Fatalf("failed to insert activity: %v", err)
		}
	}

	// Add zones to 3 of them
	for i := 1; i <= 3; i++ {
		_, err := sqlDB.Exec(
			"INSERT INTO activity_zones (activity_id, zone_type) VALUES (?, ?)",
			i, "heartrate",
		)
		if err != nil {
			t.Fatalf("failed to insert zone: %v", err)
		}
	}

	count, err := queries.CountActivitiesWithZones(ctx)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 with zones, got %d", count)
	}
}

func TestCountActivitiesWithoutZones(t *testing.T) {
	t.Parallel()

	queries, sqlDB, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert 5 activities
	for i := 1; i <= 5; i++ {
		_, err := sqlDB.Exec("INSERT INTO activities (id, name) VALUES (?, ?)", i, "Activity")
		if err != nil {
			t.Fatalf("failed to insert activity: %v", err)
		}
	}

	// Add zones to 2 of them
	for i := 1; i <= 2; i++ {
		_, err := sqlDB.Exec(
			"INSERT INTO activity_zones (activity_id, zone_type) VALUES (?, ?)",
			i, "heartrate",
		)
		if err != nil {
			t.Fatalf("failed to insert zone: %v", err)
		}
	}

	count, err := queries.CountActivitiesWithoutZones(ctx)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 without zones, got %d", count)
	}
}

func TestCreateActivity(t *testing.T) {
	t.Parallel()

	queries, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create
	err := queries.CreateActivity(ctx, db.CreateActivityParams{
		ID:       12345,
		Name:     "Test Run",
		Distance: sql.NullFloat64{Float64: 5000, Valid: true},
		Type:     sql.NullString{String: "Run", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Verify
	activity, err := queries.GetActivity(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if activity.Name != "Test Run" {
		t.Errorf("expected 'Test Run', got %q", activity.Name)
	}
}

func TestCreateActivityUpsert(t *testing.T) {
	t.Parallel()

	queries, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create initial
	err := queries.CreateActivity(ctx, db.CreateActivityParams{
		ID:   99999,
		Name: "Original",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Upsert with same ID
	err = queries.CreateActivity(ctx, db.CreateActivityParams{
		ID:   99999,
		Name: "Updated",
	})
	if err != nil {
		t.Fatalf("failed to upsert: %v", err)
	}

	// Verify update
	activity, err := queries.GetActivity(ctx, 99999)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if activity.Name != "Updated" {
		t.Errorf("expected 'Updated', got %q", activity.Name)
	}
}

func TestLogDatabaseStatsWithData(t *testing.T) {
	t.Parallel()

	queries, sqlDB, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Add data
	now := time.Now()
	for i := 1; i <= 5; i++ {
		_, err := sqlDB.Exec(
			"INSERT INTO activities (id, name, start_date) VALUES (?, ?, ?)",
			i, "Activity", now.AddDate(0, 0, -i),
		)
		if err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
	}

	// Add zones
	for i := 1; i <= 2; i++ {
		_, err := sqlDB.Exec(
			"INSERT INTO activity_zones (activity_id, zone_type) VALUES (?, ?)",
			i, "heartrate",
		)
		if err != nil {
			t.Fatalf("failed to insert zone: %v", err)
		}
	}

	// Should not panic
	LogDatabaseStats(ctx, queries)
}

func TestLogDatabaseStatsEmpty(t *testing.T) {
	t.Parallel()

	queries, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Should not panic with empty database
	LogDatabaseStats(ctx, queries)
}
