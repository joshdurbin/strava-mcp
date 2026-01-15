package sync

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joshdurbin/strava-mcp/internal/logging"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/strava"
)

// FetchProgressCallback is called after each page is fetched
type FetchProgressCallback func(result strava.FetchResult)

// SaveProgressCallback is called after each activity is saved
type SaveProgressCallback func(current, total int, activityName string)

// Service handles syncing activities from Strava to the database
type Service struct {
	queries *db.Queries
	client  *strava.Client
}

// NewService creates a new sync service
func NewService(queries *db.Queries, client *strava.Client) *Service {
	return &Service{
		queries: queries,
		client:  client,
	}
}

// Sync fetches all activities from Strava and saves them to the database
func (s *Service) Sync(ctx context.Context, fetchProgress FetchProgressCallback, saveProgress SaveProgressCallback) error {
	fmt.Println("Fetching activities from Strava...")

	var progressCb strava.ProgressCallback
	if fetchProgress != nil {
		progressCb = func(result strava.FetchResult) {
			fetchProgress(result)
		}
	}

	activities, err := s.client.FetchAllActivities(ctx, progressCb)
	if err != nil {
		return fmt.Errorf("fetching activities: %w", err)
	}

	fmt.Printf("Fetched %d total activities\n", len(activities))
	fmt.Println("Saving activities to database...")

	for i, activity := range activities {
		params := ConvertActivityToParams(activity)
		if err := s.queries.CreateActivity(ctx, params); err != nil {
			return fmt.Errorf("saving activity %d (%s): %w", activity.ID, activity.Name, err)
		}

		if saveProgress != nil {
			saveProgress(i+1, len(activities), activity.Name)
		}
	}

	return nil
}

// SyncDelta fetches only new activities since the last sync
func (s *Service) SyncDelta(ctx context.Context, since time.Time, fetchProgress FetchProgressCallback, saveProgress SaveProgressCallback) (int, error) {
	var progressCb strava.ProgressCallback
	if fetchProgress != nil {
		progressCb = func(result strava.FetchResult) {
			fetchProgress(result)
		}
	}

	activities, err := s.client.FetchActivitiesSince(ctx, since, progressCb)
	if err != nil {
		return 0, fmt.Errorf("fetching activities: %w", err)
	}

	if len(activities) == 0 {
		return 0, nil
	}

	for i, activity := range activities {
		params := ConvertActivityToParams(activity)
		if err := s.queries.CreateActivity(ctx, params); err != nil {
			return i, fmt.Errorf("saving activity %d (%s): %w", activity.ID, activity.Name, err)
		}

		if saveProgress != nil {
			saveProgress(i+1, len(activities), activity.Name)
		}
	}

	return len(activities), nil
}

// ConvertActivityToParams converts a Strava activity to database params
func ConvertActivityToParams(a strava.Activity) db.CreateActivityParams {
	return db.CreateActivityParams{
		ID:                 a.ID,
		Name:               a.Name,
		Distance:           toNullFloat64(a.Distance),
		MovingTime:         toNullInt64(int64(a.MovingTime)),
		ElapsedTime:        toNullInt64(int64(a.ElapsedTime)),
		TotalElevationGain: toNullFloat64(a.TotalElevationGain),
		Type:               toNullString(a.Type),
		SportType:          toNullString(a.SportType),
		StartDate:          toNullTime(a.StartDate),
		StartDateLocal:     toNullTime(a.StartDateLocal),
		Timezone:           toNullString(a.Timezone),
		AverageSpeed:       toNullFloat64(a.AverageSpeed),
		MaxSpeed:           toNullFloat64(a.MaxSpeed),
		AverageCadence:     toNullFloat64(a.AverageCadence),
		AverageHeartrate:   toNullFloat64(a.AverageHeartrate),
		MaxHeartrate:       toNullFloat64(a.MaxHeartrate),
		Calories:           toNullFloat64(a.Kilojoules),
	}
}

func toNullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: v != 0}
}

func toNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func toNullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}

func toNullTime(v time.Time) sql.NullTime {
	return sql.NullTime{Time: v, Valid: !v.IsZero()}
}

// ZoneSyncProgressCallback is called after each activity zone is synced
type ZoneSyncProgressCallback func(current, total int, activityID int64)

// ErrPremiumRequired is returned when Strava Summit subscription is required
var ErrPremiumRequired = strava.ErrPremiumRequired

// ErrRateLimited is returned when rate limited by the API
var ErrRateLimited = strava.ErrRateLimited

// SyncZonesForActivity fetches and stores zone data for a single activity
func (s *Service) SyncZonesForActivity(ctx context.Context, activityID int64) error {
	zones, err := s.client.FetchActivityZones(ctx, activityID)
	if err != nil {
		// Propagate premium required error so caller can handle it
		if err == strava.ErrPremiumRequired {
			return ErrPremiumRequired
		}
		// Propagate rate limit error
		if err == strava.ErrRateLimited {
			return ErrRateLimited
		}
		return fmt.Errorf("fetching zones: %w", err)
	}

	if zones == nil || len(zones) == 0 {
		return nil // No zone data available
	}

	for _, zone := range zones {
		// Create or update activity zone
		var sensorBased int64
		if zone.SensorBased {
			sensorBased = 1
		}

		zoneID, err := s.queries.CreateActivityZone(ctx, db.CreateActivityZoneParams{
			ActivityID:  activityID,
			ZoneType:    zone.Type,
			SensorBased: sensorBased,
		})
		if err != nil {
			return fmt.Errorf("creating activity zone: %w", err)
		}

		// Delete existing buckets and recreate
		if err := s.queries.DeleteZoneBucketsForActivityZone(ctx, zoneID); err != nil {
			return fmt.Errorf("deleting existing zone buckets: %w", err)
		}

		for i, bucket := range zone.DistributionBuckets {
			err := s.queries.CreateZoneBucket(ctx, db.CreateZoneBucketParams{
				ActivityZoneID: zoneID,
				ZoneNumber:     int64(i + 1),
				MinValue:       int64(bucket.Min),
				MaxValue:       int64(bucket.Max),
				TimeSeconds:    int64(bucket.Time), // Convert float64 to int64
			})
			if err != nil {
				return fmt.Errorf("creating zone bucket: %w", err)
			}
		}
	}

	return nil
}

// SyncZones syncs zone data for activities that don't have zones yet
// Returns the number synced and an error. If ErrPremiumRequired is returned,
// the caller should stop trying to sync zones.
func (s *Service) SyncZones(ctx context.Context, batchSize int, progress ZoneSyncProgressCallback) (int, error) {
	// Get activities without zones
	activityIDs, err := s.queries.GetActivitiesWithoutZones(ctx, int64(batchSize))
	if err != nil {
		return 0, fmt.Errorf("getting activities without zones: %w", err)
	}

	if len(activityIDs) == 0 {
		return 0, nil
	}

	synced := 0
	consecutiveRateLimits := 0
	maxConsecutiveRateLimits := 3 // Stop batch if we hit 3 consecutive rate limits

	for i, id := range activityIDs {
		if progress != nil {
			progress(i+1, len(activityIDs), id)
		}

		if err := s.SyncZonesForActivity(ctx, id); err != nil {
			// If premium is required, stop immediately and return the error
			if err == ErrPremiumRequired {
				return synced, ErrPremiumRequired
			}
			// If rate limited multiple times in a row, stop this batch
			if err == ErrRateLimited {
				consecutiveRateLimits++
				if consecutiveRateLimits >= maxConsecutiveRateLimits {
					logging.Warn("stopping zone sync batch due to repeated rate limiting",
						"consecutive_rate_limits", consecutiveRateLimits,
						"synced_so_far", synced)
					return synced, ErrRateLimited
				}
				continue
			}
			// Log but continue - some activities may not have zone data
			logging.Warn("failed to sync zones", "activity_id", id, "error", err)
			consecutiveRateLimits = 0 // Reset on non-rate-limit errors
			continue
		}

		synced++
		consecutiveRateLimits = 0 // Reset on success

		// Small delay to respect rate limits
		select {
		case <-ctx.Done():
			return synced, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	return synced, nil
}
