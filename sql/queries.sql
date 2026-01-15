-- name: CreateActivity :exec
INSERT INTO activities (
    id, name, distance, moving_time, elapsed_time, total_elevation_gain,
    type, sport_type, start_date, start_date_local, timezone,
    average_speed, max_speed, average_cadence, average_heartrate,
    max_heartrate, calories, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    distance = excluded.distance,
    moving_time = excluded.moving_time,
    elapsed_time = excluded.elapsed_time,
    total_elevation_gain = excluded.total_elevation_gain,
    type = excluded.type,
    sport_type = excluded.sport_type,
    start_date = excluded.start_date,
    start_date_local = excluded.start_date_local,
    timezone = excluded.timezone,
    average_speed = excluded.average_speed,
    max_speed = excluded.max_speed,
    average_cadence = excluded.average_cadence,
    average_heartrate = excluded.average_heartrate,
    max_heartrate = excluded.max_heartrate,
    calories = excluded.calories,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetActivity :one
SELECT * FROM activities WHERE id = ?;

-- name: GetAllActivities :many
SELECT * FROM activities ORDER BY start_date DESC;

-- name: GetActivitiesByType :many
SELECT * FROM activities WHERE type = ? ORDER BY start_date DESC;

-- name: GetActivitiesByDateRange :many
SELECT * FROM activities
WHERE start_date >= ? AND start_date <= ?
ORDER BY start_date DESC;

-- name: CountActivities :one
SELECT COUNT(*) FROM activities;

-- name: CountActivitiesByType :one
SELECT COUNT(*) FROM activities WHERE type = ?;

-- name: GetLatestActivityDate :one
SELECT MAX(start_date) as latest_date FROM activities;

-- name: GetOldestActivityDate :one
SELECT MIN(start_date) as oldest_date FROM activities;

-- name: GetAllActivityIDs :many
SELECT id FROM activities;

-- name: GetRecentActivities :many
SELECT * FROM activities ORDER BY start_date DESC LIMIT ?;

-- name: GetOldestActivity :one
SELECT * FROM activities ORDER BY start_date ASC LIMIT 1;

-- name: GetLatestActivity :one
SELECT * FROM activities ORDER BY start_date DESC LIMIT 1;

-- name: GetActivityTypeSummary :many
SELECT type, COUNT(*) as count FROM activities WHERE type IS NOT NULL GROUP BY type ORDER BY count DESC;

-- name: CountActivitiesByTypeInRange :one
SELECT COUNT(*) FROM activities WHERE type = ? AND start_date >= ? AND start_date <= ?;

-- name: GetActivitiesByTypeAndDateRange :many
SELECT * FROM activities WHERE type = ? AND start_date >= ? AND start_date <= ? ORDER BY start_date DESC;

-- name: GetActivityCountsByMonth :many
SELECT 
    strftime('%Y-%m', start_date) as month,
    type,
    COUNT(*) as count
FROM activities 
WHERE start_date >= ? AND start_date <= ?
GROUP BY strftime('%Y-%m', start_date), type
ORDER BY month DESC, count DESC;

-- name: GetActivityCountsByWeek :many
SELECT 
    strftime('%Y-W%W', start_date) as week,
    type,
    COUNT(*) as count
FROM activities 
WHERE start_date >= ? AND start_date <= ?
GROUP BY strftime('%Y-W%W', start_date), type
ORDER BY week DESC, count DESC;

-- Metrics aggregation queries

-- name: GetCaloriesSummary :one
SELECT 
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(AVG(calories), 0) as avg_calories,
    COALESCE(MIN(calories), 0) as min_calories,
    COALESCE(MAX(calories), 0) as max_calories,
    COUNT(*) as activity_count
FROM activities 
WHERE calories IS NOT NULL AND calories > 0;

-- name: GetCaloriesSummaryByType :one
SELECT 
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(AVG(calories), 0) as avg_calories,
    COALESCE(MIN(calories), 0) as min_calories,
    COALESCE(MAX(calories), 0) as max_calories,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND calories IS NOT NULL AND calories > 0;

-- name: GetCaloriesSummaryInRange :one
SELECT 
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(AVG(calories), 0) as avg_calories,
    COALESCE(MIN(calories), 0) as min_calories,
    COALESCE(MAX(calories), 0) as max_calories,
    COUNT(*) as activity_count
FROM activities 
WHERE start_date >= ? AND start_date <= ? AND calories IS NOT NULL AND calories > 0;

-- name: GetCaloriesSummaryByTypeInRange :one
SELECT 
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(AVG(calories), 0) as avg_calories,
    COALESCE(MIN(calories), 0) as min_calories,
    COALESCE(MAX(calories), 0) as max_calories,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND start_date >= ? AND start_date <= ? AND calories IS NOT NULL AND calories > 0;

-- name: GetHeartrateSummary :one
SELECT 
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(MIN(average_heartrate), 0) as min_avg_heartrate,
    COALESCE(MAX(average_heartrate), 0) as max_avg_heartrate,
    COALESCE(MAX(max_heartrate), 0) as overall_max_heartrate,
    COUNT(*) as activity_count
FROM activities 
WHERE average_heartrate IS NOT NULL AND average_heartrate > 0;

-- name: GetHeartrateSummaryByType :one
SELECT 
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(MIN(average_heartrate), 0) as min_avg_heartrate,
    COALESCE(MAX(average_heartrate), 0) as max_avg_heartrate,
    COALESCE(MAX(max_heartrate), 0) as overall_max_heartrate,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND average_heartrate IS NOT NULL AND average_heartrate > 0;

-- name: GetHeartrateSummaryInRange :one
SELECT 
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(MIN(average_heartrate), 0) as min_avg_heartrate,
    COALESCE(MAX(average_heartrate), 0) as max_avg_heartrate,
    COALESCE(MAX(max_heartrate), 0) as overall_max_heartrate,
    COUNT(*) as activity_count
FROM activities 
WHERE start_date >= ? AND start_date <= ? AND average_heartrate IS NOT NULL AND average_heartrate > 0;

-- name: GetHeartrateSummaryByTypeInRange :one
SELECT 
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(MIN(average_heartrate), 0) as min_avg_heartrate,
    COALESCE(MAX(average_heartrate), 0) as max_avg_heartrate,
    COALESCE(MAX(max_heartrate), 0) as overall_max_heartrate,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND start_date >= ? AND start_date <= ? AND average_heartrate IS NOT NULL AND average_heartrate > 0;

-- name: GetSpeedSummary :one
SELECT 
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(MIN(average_speed), 0) as min_avg_speed,
    COALESCE(MAX(average_speed), 0) as max_avg_speed,
    COALESCE(MAX(max_speed), 0) as overall_max_speed,
    COUNT(*) as activity_count
FROM activities 
WHERE average_speed IS NOT NULL AND average_speed > 0;

-- name: GetSpeedSummaryByType :one
SELECT 
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(MIN(average_speed), 0) as min_avg_speed,
    COALESCE(MAX(average_speed), 0) as max_avg_speed,
    COALESCE(MAX(max_speed), 0) as overall_max_speed,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND average_speed IS NOT NULL AND average_speed > 0;

-- name: GetSpeedSummaryInRange :one
SELECT 
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(MIN(average_speed), 0) as min_avg_speed,
    COALESCE(MAX(average_speed), 0) as max_avg_speed,
    COALESCE(MAX(max_speed), 0) as overall_max_speed,
    COUNT(*) as activity_count
FROM activities 
WHERE start_date >= ? AND start_date <= ? AND average_speed IS NOT NULL AND average_speed > 0;

-- name: GetSpeedSummaryByTypeInRange :one
SELECT 
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(MIN(average_speed), 0) as min_avg_speed,
    COALESCE(MAX(average_speed), 0) as max_avg_speed,
    COALESCE(MAX(max_speed), 0) as overall_max_speed,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND start_date >= ? AND start_date <= ? AND average_speed IS NOT NULL AND average_speed > 0;

-- name: GetCadenceSummary :one
SELECT 
    COALESCE(AVG(average_cadence), 0) as avg_cadence,
    COALESCE(MIN(average_cadence), 0) as min_cadence,
    COALESCE(MAX(average_cadence), 0) as max_cadence,
    COUNT(*) as activity_count
FROM activities 
WHERE average_cadence IS NOT NULL AND average_cadence > 0;

-- name: GetCadenceSummaryByType :one
SELECT 
    COALESCE(AVG(average_cadence), 0) as avg_cadence,
    COALESCE(MIN(average_cadence), 0) as min_cadence,
    COALESCE(MAX(average_cadence), 0) as max_cadence,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND average_cadence IS NOT NULL AND average_cadence > 0;

-- name: GetCadenceSummaryInRange :one
SELECT 
    COALESCE(AVG(average_cadence), 0) as avg_cadence,
    COALESCE(MIN(average_cadence), 0) as min_cadence,
    COALESCE(MAX(average_cadence), 0) as max_cadence,
    COUNT(*) as activity_count
FROM activities 
WHERE start_date >= ? AND start_date <= ? AND average_cadence IS NOT NULL AND average_cadence > 0;

-- name: GetCadenceSummaryByTypeInRange :one
SELECT 
    COALESCE(AVG(average_cadence), 0) as avg_cadence,
    COALESCE(MIN(average_cadence), 0) as min_cadence,
    COALESCE(MAX(average_cadence), 0) as max_cadence,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND start_date >= ? AND start_date <= ? AND average_cadence IS NOT NULL AND average_cadence > 0;

-- name: GetDistanceSummary :one
SELECT 
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(AVG(distance), 0) as avg_distance,
    COALESCE(MIN(distance), 0) as min_distance,
    COALESCE(MAX(distance), 0) as max_distance,
    COUNT(*) as activity_count
FROM activities 
WHERE distance IS NOT NULL AND distance > 0;

-- name: GetDistanceSummaryByType :one
SELECT 
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(AVG(distance), 0) as avg_distance,
    COALESCE(MIN(distance), 0) as min_distance,
    COALESCE(MAX(distance), 0) as max_distance,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND distance IS NOT NULL AND distance > 0;

-- name: GetDistanceSummaryInRange :one
SELECT 
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(AVG(distance), 0) as avg_distance,
    COALESCE(MIN(distance), 0) as min_distance,
    COALESCE(MAX(distance), 0) as max_distance,
    COUNT(*) as activity_count
FROM activities 
WHERE start_date >= ? AND start_date <= ? AND distance IS NOT NULL AND distance > 0;

-- name: GetDistanceSummaryByTypeInRange :one
SELECT 
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(AVG(distance), 0) as avg_distance,
    COALESCE(MIN(distance), 0) as min_distance,
    COALESCE(MAX(distance), 0) as max_distance,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND start_date >= ? AND start_date <= ? AND distance IS NOT NULL AND distance > 0;

-- name: GetElevationSummary :one
SELECT 
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation,
    COALESCE(AVG(total_elevation_gain), 0) as avg_elevation,
    COALESCE(MIN(total_elevation_gain), 0) as min_elevation,
    COALESCE(MAX(total_elevation_gain), 0) as max_elevation,
    COUNT(*) as activity_count
FROM activities 
WHERE total_elevation_gain IS NOT NULL AND total_elevation_gain > 0;

-- name: GetElevationSummaryByType :one
SELECT 
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation,
    COALESCE(AVG(total_elevation_gain), 0) as avg_elevation,
    COALESCE(MIN(total_elevation_gain), 0) as min_elevation,
    COALESCE(MAX(total_elevation_gain), 0) as max_elevation,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND total_elevation_gain IS NOT NULL AND total_elevation_gain > 0;

-- name: GetElevationSummaryInRange :one
SELECT 
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation,
    COALESCE(AVG(total_elevation_gain), 0) as avg_elevation,
    COALESCE(MIN(total_elevation_gain), 0) as min_elevation,
    COALESCE(MAX(total_elevation_gain), 0) as max_elevation,
    COUNT(*) as activity_count
FROM activities 
WHERE start_date >= ? AND start_date <= ? AND total_elevation_gain IS NOT NULL AND total_elevation_gain > 0;

-- name: GetElevationSummaryByTypeInRange :one
SELECT 
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation,
    COALESCE(AVG(total_elevation_gain), 0) as avg_elevation,
    COALESCE(MIN(total_elevation_gain), 0) as min_elevation,
    COALESCE(MAX(total_elevation_gain), 0) as max_elevation,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND start_date >= ? AND start_date <= ? AND total_elevation_gain IS NOT NULL AND total_elevation_gain > 0;

-- name: GetDurationSummary :one
SELECT 
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(moving_time), 0) as avg_moving_time,
    COALESCE(MIN(moving_time), 0) as min_moving_time,
    COALESCE(MAX(moving_time), 0) as max_moving_time,
    COUNT(*) as activity_count
FROM activities 
WHERE moving_time IS NOT NULL AND moving_time > 0;

-- name: GetDurationSummaryByType :one
SELECT 
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(moving_time), 0) as avg_moving_time,
    COALESCE(MIN(moving_time), 0) as min_moving_time,
    COALESCE(MAX(moving_time), 0) as max_moving_time,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND moving_time IS NOT NULL AND moving_time > 0;

-- name: GetDurationSummaryInRange :one
SELECT 
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(moving_time), 0) as avg_moving_time,
    COALESCE(MIN(moving_time), 0) as min_moving_time,
    COALESCE(MAX(moving_time), 0) as max_moving_time,
    COUNT(*) as activity_count
FROM activities 
WHERE start_date >= ? AND start_date <= ? AND moving_time IS NOT NULL AND moving_time > 0;

-- name: GetDurationSummaryByTypeInRange :one
SELECT 
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(moving_time), 0) as avg_moving_time,
    COALESCE(MIN(moving_time), 0) as min_moving_time,
    COALESCE(MAX(moving_time), 0) as max_moving_time,
    COUNT(*) as activity_count
FROM activities 
WHERE type = ? AND start_date >= ? AND start_date <= ? AND moving_time IS NOT NULL AND moving_time > 0;

-- Auth config queries

-- name: SaveAuthConfig :exec
INSERT INTO auth_config (id, client_id, client_secret, access_token, refresh_token, expires_at, updated_at)
VALUES (1, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
    client_id = excluded.client_id,
    client_secret = excluded.client_secret,
    access_token = excluded.access_token,
    refresh_token = excluded.refresh_token,
    expires_at = excluded.expires_at,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetAuthConfig :one
SELECT * FROM auth_config WHERE id = 1;

-- name: UpdateTokens :exec
UPDATE auth_config SET
    access_token = ?,
    refresh_token = ?,
    expires_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 1;

-- name: DeleteAuthConfig :exec
DELETE FROM auth_config WHERE id = 1;

-- Activity zone queries

-- name: CreateActivityZone :one
INSERT INTO activity_zones (activity_id, zone_type, sensor_based)
VALUES (?, ?, ?)
ON CONFLICT(activity_id, zone_type) DO UPDATE SET
    sensor_based = excluded.sensor_based
RETURNING id;

-- name: CreateZoneBucket :exec
INSERT INTO zone_buckets (activity_zone_id, zone_number, min_value, max_value, time_seconds)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteZoneBucketsForActivityZone :exec
DELETE FROM zone_buckets WHERE activity_zone_id = ?;

-- name: GetActivityZoneByActivityAndType :one
SELECT * FROM activity_zones WHERE activity_id = ? AND zone_type = ?;

-- name: GetActivityZones :many
SELECT * FROM activity_zones WHERE activity_id = ?;

-- name: GetZoneBuckets :many
SELECT * FROM zone_buckets WHERE activity_zone_id = ? ORDER BY zone_number;

-- name: GetActivitiesWithZones :many
SELECT DISTINCT a.id, a.name, a.type, a.sport_type, a.start_date
FROM activities a
JOIN activity_zones az ON a.id = az.activity_id
ORDER BY a.start_date DESC
LIMIT ?;

-- name: GetActivitiesWithoutZones :many
SELECT a.id FROM activities a
LEFT JOIN activity_zones az ON a.id = az.activity_id
WHERE az.id IS NULL
ORDER BY a.start_date DESC
LIMIT ?;

-- name: CountActivitiesWithZones :one
SELECT COUNT(DISTINCT activity_id) FROM activity_zones;

-- name: CountActivitiesWithoutZones :one
SELECT COUNT(*) FROM activities a
LEFT JOIN activity_zones az ON a.id = az.activity_id
WHERE az.id IS NULL;

-- name: GetHeartRateZoneSummary :many
SELECT 
    zb.zone_number,
    SUM(zb.time_seconds) as total_time,
    AVG(zb.time_seconds) as avg_time,
    COUNT(*) as activity_count
FROM zone_buckets zb
JOIN activity_zones az ON zb.activity_zone_id = az.id
WHERE az.zone_type = 'heartrate'
GROUP BY zb.zone_number
ORDER BY zb.zone_number;

-- name: GetHeartRateZoneSummaryByType :many
SELECT 
    zb.zone_number,
    SUM(zb.time_seconds) as total_time,
    AVG(zb.time_seconds) as avg_time,
    COUNT(*) as activity_count
FROM zone_buckets zb
JOIN activity_zones az ON zb.activity_zone_id = az.id
JOIN activities a ON az.activity_id = a.id
WHERE az.zone_type = 'heartrate' AND a.type = ?
GROUP BY zb.zone_number
ORDER BY zb.zone_number;

-- name: GetHeartRateZoneSummaryInRange :many
SELECT 
    zb.zone_number,
    SUM(zb.time_seconds) as total_time,
    AVG(zb.time_seconds) as avg_time,
    COUNT(*) as activity_count
FROM zone_buckets zb
JOIN activity_zones az ON zb.activity_zone_id = az.id
JOIN activities a ON az.activity_id = a.id
WHERE az.zone_type = 'heartrate' 
  AND a.start_date >= ? AND a.start_date <= ?
GROUP BY zb.zone_number
ORDER BY zb.zone_number;

-- name: GetHeartRateZoneSummaryByTypeInRange :many
SELECT 
    zb.zone_number,
    SUM(zb.time_seconds) as total_time,
    AVG(zb.time_seconds) as avg_time,
    COUNT(*) as activity_count
FROM zone_buckets zb
JOIN activity_zones az ON zb.activity_zone_id = az.id
JOIN activities a ON az.activity_id = a.id
WHERE az.zone_type = 'heartrate' AND a.type = ?
  AND a.start_date >= ? AND a.start_date <= ?
GROUP BY zb.zone_number
ORDER BY zb.zone_number;

-- name: GetPowerZoneSummary :many
SELECT 
    zb.zone_number,
    SUM(zb.time_seconds) as total_time,
    AVG(zb.time_seconds) as avg_time,
    COUNT(*) as activity_count
FROM zone_buckets zb
JOIN activity_zones az ON zb.activity_zone_id = az.id
WHERE az.zone_type = 'power'
GROUP BY zb.zone_number
ORDER BY zb.zone_number;

-- name: GetPowerZoneSummaryByType :many
SELECT 
    zb.zone_number,
    SUM(zb.time_seconds) as total_time,
    AVG(zb.time_seconds) as avg_time,
    COUNT(*) as activity_count
FROM zone_buckets zb
JOIN activity_zones az ON zb.activity_zone_id = az.id
JOIN activities a ON az.activity_id = a.id
WHERE az.zone_type = 'power' AND a.type = ?
GROUP BY zb.zone_number
ORDER BY zb.zone_number;

-- name: GetPowerZoneSummaryInRange :many
SELECT
    zb.zone_number,
    SUM(zb.time_seconds) as total_time,
    AVG(zb.time_seconds) as avg_time,
    COUNT(*) as activity_count
FROM zone_buckets zb
JOIN activity_zones az ON zb.activity_zone_id = az.id
JOIN activities a ON az.activity_id = a.id
WHERE az.zone_type = 'power'
  AND a.start_date >= ? AND a.start_date <= ?
GROUP BY zb.zone_number
ORDER BY zb.zone_number;

-- Training summary queries

-- name: GetTrainingSummary :one
SELECT
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation
FROM activities;

-- name: GetTrainingSummaryByType :one
SELECT
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation
FROM activities
WHERE type = ?;

-- name: GetTrainingSummaryInRange :one
SELECT
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation
FROM activities
WHERE start_date >= ? AND start_date <= ?;

-- name: GetTrainingSummaryByTypeInRange :one
SELECT
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(average_speed), 0) as avg_speed,
    COALESCE(AVG(average_heartrate), 0) as avg_heartrate,
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation
FROM activities
WHERE type = ? AND start_date >= ? AND start_date <= ?;

-- Period comparison queries with detailed metrics

-- name: GetPeriodStats :one
SELECT
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(average_speed), 0) as avg_speed
FROM activities
WHERE start_date >= ? AND start_date <= ?;

-- name: GetPeriodStatsByType :one
SELECT
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_moving_time,
    COALESCE(AVG(average_speed), 0) as avg_speed
FROM activities
WHERE type = ? AND start_date >= ? AND start_date <= ?;

-- Consolidated counting queries

-- name: CountActivitiesInRange :one
SELECT COUNT(*) FROM activities WHERE start_date >= ? AND start_date <= ?;

-- name: GetActivityTypeSummaryInRange :many
SELECT type, COUNT(*) as count FROM activities
WHERE start_date >= ? AND start_date <= ? AND type IS NOT NULL
GROUP BY type ORDER BY count DESC;

-- Personal records queries

-- name: GetFastestActivity :one
SELECT * FROM activities
WHERE average_speed IS NOT NULL AND average_speed > 0
ORDER BY average_speed DESC
LIMIT 1;

-- name: GetFastestActivityByType :one
SELECT * FROM activities
WHERE type = ? AND average_speed IS NOT NULL AND average_speed > 0
ORDER BY average_speed DESC
LIMIT 1;

-- name: GetLongestDistanceActivity :one
SELECT * FROM activities
WHERE distance IS NOT NULL AND distance > 0
ORDER BY distance DESC
LIMIT 1;

-- name: GetLongestDistanceActivityByType :one
SELECT * FROM activities
WHERE type = ? AND distance IS NOT NULL AND distance > 0
ORDER BY distance DESC
LIMIT 1;

-- name: GetLongestDurationActivity :one
SELECT * FROM activities
WHERE moving_time IS NOT NULL AND moving_time > 0
ORDER BY moving_time DESC
LIMIT 1;

-- name: GetLongestDurationActivityByType :one
SELECT * FROM activities
WHERE type = ? AND moving_time IS NOT NULL AND moving_time > 0
ORDER BY moving_time DESC
LIMIT 1;

-- name: GetHighestElevationActivity :one
SELECT * FROM activities
WHERE total_elevation_gain IS NOT NULL AND total_elevation_gain > 0
ORDER BY total_elevation_gain DESC
LIMIT 1;

-- name: GetHighestElevationActivityByType :one
SELECT * FROM activities
WHERE type = ? AND total_elevation_gain IS NOT NULL AND total_elevation_gain > 0
ORDER BY total_elevation_gain DESC
LIMIT 1;

-- name: GetMostCaloriesActivity :one
SELECT * FROM activities
WHERE calories IS NOT NULL AND calories > 0
ORDER BY calories DESC
LIMIT 1;

-- name: GetMostCaloriesActivityByType :one
SELECT * FROM activities
WHERE type = ? AND calories IS NOT NULL AND calories > 0
ORDER BY calories DESC
LIMIT 1;

-- Weekly volume queries for training load analysis

-- name: GetWeeklyVolume :many
SELECT
    strftime('%Y-W%W', start_date) as week,
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_duration,
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation
FROM activities
WHERE start_date >= ? AND start_date <= ?
GROUP BY strftime('%Y-W%W', start_date)
ORDER BY week DESC;

-- name: GetWeeklyVolumeByType :many
SELECT
    strftime('%Y-W%W', start_date) as week,
    COUNT(*) as activity_count,
    COALESCE(SUM(distance), 0) as total_distance,
    COALESCE(SUM(moving_time), 0) as total_duration,
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(SUM(total_elevation_gain), 0) as total_elevation
FROM activities
WHERE type = ? AND start_date >= ? AND start_date <= ?
GROUP BY strftime('%Y-W%W', start_date)
ORDER BY week DESC;

-- Flexible activity search with sorting

-- name: SearchActivities :many
SELECT * FROM activities
WHERE (? IS NULL OR type = ?)
  AND (? IS NULL OR start_date >= ?)
  AND (? IS NULL OR start_date <= ?)
ORDER BY start_date DESC
LIMIT ?;

-- name: SearchActivitiesByDistance :many
SELECT * FROM activities
WHERE (? IS NULL OR type = ?)
  AND (? IS NULL OR start_date >= ?)
  AND (? IS NULL OR start_date <= ?)
  AND distance IS NOT NULL
ORDER BY distance DESC
LIMIT ?;

-- name: SearchActivitiesByDuration :many
SELECT * FROM activities
WHERE (? IS NULL OR type = ?)
  AND (? IS NULL OR start_date >= ?)
  AND (? IS NULL OR start_date <= ?)
  AND moving_time IS NOT NULL
ORDER BY moving_time DESC
LIMIT ?;

-- name: SearchActivitiesBySpeed :many
SELECT * FROM activities
WHERE (? IS NULL OR type = ?)
  AND (? IS NULL OR start_date >= ?)
  AND (? IS NULL OR start_date <= ?)
  AND average_speed IS NOT NULL
ORDER BY average_speed DESC
LIMIT ?;

-- name: SearchActivitiesByElevation :many
SELECT * FROM activities
WHERE (? IS NULL OR type = ?)
  AND (? IS NULL OR start_date >= ?)
  AND (? IS NULL OR start_date <= ?)
  AND total_elevation_gain IS NOT NULL
ORDER BY total_elevation_gain DESC
LIMIT ?;
