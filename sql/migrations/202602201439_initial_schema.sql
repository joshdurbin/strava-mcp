-- +goose Up
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

CREATE INDEX IF NOT EXISTS idx_activities_start_date ON activities(start_date);
CREATE INDEX IF NOT EXISTS idx_activities_type ON activities(type);
CREATE INDEX IF NOT EXISTS idx_activities_sport_type ON activities(sport_type);

-- Composite index for type + date range queries (used by all metric summaries)
CREATE INDEX IF NOT EXISTS idx_activities_type_start_date ON activities(type, start_date);

-- Covering index for GROUP BY queries on type
CREATE INDEX IF NOT EXISTS idx_activities_type_count ON activities(type) WHERE type IS NOT NULL;

-- Auth configuration table (stores client credentials and tokens)
CREATE TABLE IF NOT EXISTS auth_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- Singleton row
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Activity zones table (one row per zone type per activity)
CREATE TABLE IF NOT EXISTS activity_zones (
    id INTEGER PRIMARY KEY,
    activity_id INTEGER NOT NULL,
    zone_type TEXT NOT NULL,  -- 'heartrate' or 'power'
    sensor_based INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    UNIQUE(activity_id, zone_type)
);

-- Zone distribution buckets (time spent in each zone range)
CREATE TABLE IF NOT EXISTS zone_buckets (
    id INTEGER PRIMARY KEY,
    activity_zone_id INTEGER NOT NULL,
    zone_number INTEGER NOT NULL,  -- 1-5 typically
    min_value INTEGER NOT NULL,    -- BPM or watts
    max_value INTEGER NOT NULL,    -- BPM or watts (-1 for unbounded)
    time_seconds INTEGER NOT NULL, -- Time spent in zone
    FOREIGN KEY (activity_zone_id) REFERENCES activity_zones(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_activity_zones_activity_id ON activity_zones(activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_zones_type ON activity_zones(zone_type);
CREATE INDEX IF NOT EXISTS idx_zone_buckets_zone_id ON zone_buckets(activity_zone_id);

-- +goose Down
DROP INDEX IF EXISTS idx_zone_buckets_zone_id;
DROP INDEX IF EXISTS idx_activity_zones_type;
DROP INDEX IF EXISTS idx_activity_zones_activity_id;
DROP TABLE IF EXISTS zone_buckets;
DROP TABLE IF EXISTS activity_zones;
DROP TABLE IF EXISTS auth_config;
DROP INDEX IF EXISTS idx_activities_type_count;
DROP INDEX IF EXISTS idx_activities_type_start_date;
DROP INDEX IF EXISTS idx_activities_sport_type;
DROP INDEX IF EXISTS idx_activities_type;
DROP INDEX IF EXISTS idx_activities_start_date;
DROP TABLE IF EXISTS activities;
