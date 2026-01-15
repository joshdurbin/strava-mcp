package workers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/auth"
	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/logging"
	"github.com/joshdurbin/strava-mcp/internal/strava"
	syncsvc "github.com/joshdurbin/strava-mcp/internal/sync"
)

// TokenRefresher keeps auth tokens up to date
type TokenRefresher struct {
	storage  *auth.Storage
	interval time.Duration
}

// NewTokenRefresher creates a new token refresher worker
func NewTokenRefresher(storage *auth.Storage, interval time.Duration) *TokenRefresher {
	return &TokenRefresher{
		storage:  storage,
		interval: interval,
	}
}

// Run starts the token refresh worker
func (t *TokenRefresher) Run(ctx context.Context) {
	log := logging.Logger
	log.Info().Dur("interval", t.interval).Msg("token refresher started")

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	// Do an initial check
	t.checkAndRefresh()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("token refresher stopped")
			return
		case <-ticker.C:
			t.checkAndRefresh()
		}
	}
}

func (t *TokenRefresher) checkAndRefresh() {
	log := logging.Logger
	log.Debug().Msg("checking token validity")

	tokens, err := t.storage.LoadTokens()
	if err != nil {
		log.Error().Err(err).Msg("failed to load tokens for refresh check")
		return
	}

	// Refresh if token expires within 10 minutes
	expiresAt := time.Unix(tokens.ExpiresAt, 0)
	timeUntilExpiry := time.Until(expiresAt)

	if timeUntilExpiry < 10*time.Minute {
		log.Info().Dur("expires_in", timeUntilExpiry).Msg("token expiring soon, refreshing")

		// Load client credentials from storage
		clientConfig, err := t.storage.LoadClientConfig()
		if err != nil {
			log.Error().Err(err).Msg("failed to load client config for refresh")
			return
		}

		newTokens, err := auth.RefreshAccessToken(clientConfig.ClientID, clientConfig.ClientSecret, tokens.RefreshToken)
		if err != nil {
			log.Error().Err(err).Msg("failed to refresh token")
			return
		}

		if err := t.storage.SaveTokens(newTokens); err != nil {
			log.Error().Err(err).Msg("failed to save refreshed tokens")
			return
		}

		log.Info().
			Str("new_expires_at", time.Unix(newTokens.ExpiresAt, 0).Format(time.RFC3339)).
			Msg("token refreshed successfully")
	} else {
		log.Debug().Dur("expires_in", timeUntilExpiry.Round(time.Second)).Msg("token still valid")
	}
}

// ActivitySyncer periodically syncs activities from Strava
type ActivitySyncer struct {
	queries     *db.Queries
	storage     *auth.Storage
	interval    time.Duration
	retryConfig strava.RetryConfig
}

// NewActivitySyncer creates a new activity sync worker
func NewActivitySyncer(queries *db.Queries, storage *auth.Storage, interval time.Duration, retryConfig strava.RetryConfig) *ActivitySyncer {
	return &ActivitySyncer{
		queries:     queries,
		storage:     storage,
		interval:    interval,
		retryConfig: retryConfig,
	}
}

// Run starts the activity sync worker
func (a *ActivitySyncer) Run(ctx context.Context) {
	log := logging.Logger
	log.Info().Dur("interval", a.interval).Msg("activity syncer started")

	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	// Do an initial sync
	a.syncActivities(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("activity syncer stopped")
			return
		case <-ticker.C:
			a.syncActivities(ctx)
		}
	}
}

func (a *ActivitySyncer) syncActivities(ctx context.Context) {
	log := logging.Logger
	log.Info().Msg("starting activity sync")

	// Get valid access token
	accessToken, err := a.storage.GetValidAccessToken()
	if err != nil {
		log.Error().Err(err).Msg("failed to get access token for sync")
		return
	}

	// Create Strava client with retry config
	client := strava.NewClientWithRetryConfig(accessToken, a.retryConfig)

	// Wait for rate limits before starting (in case we're approaching limits from previous sync)
	if err := client.WaitForRateLimit(ctx); err != nil {
		log.Info().Err(err).Msg("activity sync cancelled while waiting for rate limit")
		return
	}

	// Get the latest activity date for delta sync
	latestDate, err := a.getLatestActivityDate(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get latest activity date, doing full sync")
	}

	var activities []strava.Activity

	progressCallback := func(result strava.FetchResult) {
		rl := result.RateLimit
		logEvent := log.Debug()
		// Upgrade to Info level if retrying or rate limited
		if result.IsRetrying || rl.IsRateLimited {
			logEvent = log.Info()
		}
		logEvent.
			Int("page", result.Page).
			Int("activities_on_page", len(result.Activities)).
			Int("total_fetched", result.TotalFetched).
			Str("15min_usage", fmt.Sprintf("%d/%d", rl.Usage15Min, rl.Limit15Min)).
			Str("daily_usage", fmt.Sprintf("%d/%d", rl.UsageDaily, rl.LimitDaily)).
			Bool("is_retrying", result.IsRetrying).
			Bool("rate_limited", rl.IsRateLimited).
			Msg("activity sync progress")
	}

	if !latestDate.IsZero() {
		log.Info().Str("since", latestDate.Format(time.RFC3339)).Msg("performing delta sync")
		activities, err = client.FetchActivitiesSince(ctx, latestDate, progressCallback)
	} else {
		log.Info().Msg("performing full sync")
		activities, err = client.FetchAllActivities(ctx, progressCallback)
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to fetch activities")
		return
	}

	if len(activities) == 0 {
		rl := client.GetRateLimit()
		log.Info().
			Str("15min_usage", fmt.Sprintf("%d/%d", rl.Usage15Min, rl.Limit15Min)).
			Str("daily_usage", fmt.Sprintf("%d/%d", rl.UsageDaily, rl.LimitDaily)).
			Msg("no new activities to sync")
		return
	}

	rl := client.GetRateLimit()
	log.Info().
		Int("count", len(activities)).
		Str("15min_usage", fmt.Sprintf("%d/%d", rl.Usage15Min, rl.Limit15Min)).
		Str("daily_usage", fmt.Sprintf("%d/%d", rl.UsageDaily, rl.LimitDaily)).
		Msg("fetched activities")

	// Save activities to database
	saved := 0
	for _, activity := range activities {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			log.Info().Int("fetched", len(activities)).Int("saved", saved).Msg("sync interrupted")
			return
		default:
		}

		params := syncsvc.ConvertActivityToParams(activity)
		if err := a.queries.CreateActivity(ctx, params); err != nil {
			log.Error().
				Int64("activity_id", activity.ID).
				Str("activity_name", activity.Name).
				Err(err).
				Msg("failed to save activity")
			continue
		}
		saved++
		log.Debug().
			Int64("activity_id", activity.ID).
			Str("activity_name", activity.Name).
			Str("activity_type", activity.Type).
			Msg("saved activity")
	}

	rl = client.GetRateLimit()
	log.Info().
		Int("fetched", len(activities)).
		Int("saved", saved).
		Str("15min_usage", fmt.Sprintf("%d/%d", rl.Usage15Min, rl.Limit15Min)).
		Str("daily_usage", fmt.Sprintf("%d/%d", rl.UsageDaily, rl.LimitDaily)).
		Msg("activity sync completed")
}

func (a *ActivitySyncer) getLatestActivityDate(ctx context.Context) (time.Time, error) {
	activities, err := a.queries.GetRecentActivities(ctx, 1)
	if err != nil {
		return time.Time{}, err
	}

	if len(activities) == 0 {
		return time.Time{}, nil
	}

	if activities[0].StartDate.Valid {
		return activities[0].StartDate.Time, nil
	}

	return time.Time{}, nil
}

// SyncOnce performs a single sync (used for initial sync on startup)
func SyncOnce(ctx context.Context, queries *db.Queries, accessToken string, retryConfig strava.RetryConfig) error {
	log := logging.Logger
	log.Info().Msg("performing initial sync")

	client := strava.NewClientWithRetryConfig(accessToken, retryConfig)

	// Get the latest activity date for delta sync
	var latestDate time.Time
	activities, err := queries.GetRecentActivities(ctx, 1)
	if err == nil && len(activities) > 0 && activities[0].StartDate.Valid {
		latestDate = activities[0].StartDate.Time
	}

	var fetchedActivities []strava.Activity

	progressCallback := func(result strava.FetchResult) {
		log.Debug().
			Int("page", result.Page).
			Int("total_fetched", result.TotalFetched).
			Msg("initial sync progress")
	}

	if !latestDate.IsZero() {
		log.Info().Str("since", latestDate.Format(time.RFC3339)).Msg("performing delta sync")
		fetchedActivities, err = client.FetchActivitiesSince(ctx, latestDate, progressCallback)
	} else {
		log.Info().Msg("performing full sync (no existing activities)")
		fetchedActivities, err = client.FetchAllActivities(ctx, progressCallback)
	}

	if err != nil {
		return err
	}

	if len(fetchedActivities) == 0 {
		log.Info().Msg("no activities to sync")
		return nil
	}

	log.Info().Int("count", len(fetchedActivities)).Msg("saving activities")

	saved := 0
	for _, activity := range fetchedActivities {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			log.Info().Int("fetched", len(fetchedActivities)).Int("saved", saved).Msg("initial sync interrupted")
			return ctx.Err()
		default:
		}

		params := syncsvc.ConvertActivityToParams(activity)
		if err := queries.CreateActivity(ctx, params); err != nil {
			log.Error().Int64("activity_id", activity.ID).Err(err).Msg("failed to save activity")
			continue
		}
		saved++
	}

	log.Info().Int("fetched", len(fetchedActivities)).Int("saved", saved).Msg("initial sync completed")
	return nil
}

// ZoneSyncer periodically syncs activity zones from Strava
type ZoneSyncer struct {
	queries         *db.Queries
	storage         *auth.Storage
	interval        time.Duration
	batchSize       int
	premiumRequired bool // Set to true if we detect premium is required
	retryConfig     strava.RetryConfig
}

// NewZoneSyncer creates a new zone sync worker
func NewZoneSyncer(queries *db.Queries, storage *auth.Storage, interval time.Duration, retryConfig strava.RetryConfig) *ZoneSyncer {
	return &ZoneSyncer{
		queries:     queries,
		storage:     storage,
		interval:    interval,
		batchSize:   25, // Sync 25 activities per batch, but run multiple batches based on rate limit headroom
		retryConfig: retryConfig,
	}
}

// Run starts the zone sync worker
func (z *ZoneSyncer) Run(ctx context.Context) {
	log := logging.Logger
	log.Info().Dur("interval", z.interval).Int("batch_size", z.batchSize).Msg("zone syncer started")

	// Initial delay to let activity sync complete first
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}

	// Do an initial sync (continuous until rate limited or done)
	z.syncZonesContinuously(ctx)

	ticker := time.NewTicker(z.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("zone syncer stopped")
			return
		case <-ticker.C:
			z.syncZonesContinuously(ctx)
		}
	}
}

// syncZonesContinuously syncs zones in batches, continuing as long as we have API rate limit headroom
func (z *ZoneSyncer) syncZonesContinuously(ctx context.Context) {
	log := logging.Logger

	// If we've already detected premium is required, skip
	if z.premiumRequired {
		log.Debug().Msg("zone sync skipped - Strava Summit subscription required")
		return
	}

	// Check how many activities need zones synced
	withoutZones, err := z.queries.CountActivitiesWithoutZones(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to count activities without zones")
		return
	}

	if withoutZones == 0 {
		log.Debug().Msg("all activities have zones synced")
		return
	}

	log.Info().Int64("activities_remaining", withoutZones).Msg("starting zone sync")

	// Get valid access token
	accessToken, err := z.storage.GetValidAccessToken()
	if err != nil {
		log.Error().Err(err).Msg("failed to get access token for zone sync")
		return
	}

	// Create Strava client with retry config and sync service
	client := strava.NewClientWithRetryConfig(accessToken, z.retryConfig)
	syncService := syncsvc.NewService(z.queries, client)

	totalSynced := 0
	batchNum := 0

	for {
		// Check if we need to wait for rate limits before proceeding
		rateLimit := client.GetRateLimit()

		// If approaching daily limit, stop for this session (can't wait that long)
		if rateLimit.IsApproachingDailyLimit() {
			log.Info().
				Int("usage", rateLimit.UsageDaily).
				Int("limit", rateLimit.LimitDaily).
				Dur("reset_in", rateLimit.TimeUntilDailyReset.Round(time.Minute)).
				Int("total_synced", totalSynced).
				Msg("zone sync stopping - approaching daily rate limit")
			break
		}

		// If approaching 15-minute limit, wait for window reset then continue
		if rateLimit.IsApproaching15MinLimit() {
			log.Info().
				Int("usage", rateLimit.Usage15Min).
				Int("limit", rateLimit.Limit15Min).
				Dur("wait", rateLimit.TimeUntil15MinReset.Round(time.Second)).
				Int("total_synced", totalSynced).
				Msg("zone sync waiting for 15-minute rate limit window to reset")

			if err := client.WaitForRateLimit(ctx); err != nil {
				log.Info().Err(err).Msg("zone sync cancelled while waiting for rate limit")
				return
			}
			// After waiting, continue to next iteration to re-check limits
			continue
		}

		batchNum++
		synced, err := syncService.SyncZones(ctx, z.batchSize, nil)
		totalSynced += synced

		if err != nil {
			if err == syncsvc.ErrPremiumRequired {
				z.premiumRequired = true
				log.Warn().
					Str("info", "https://www.strava.com/summit").
					Msg("zone sync disabled - Activity Zones API requires Strava Summit (premium) subscription")
				return
			}
			if err == syncsvc.ErrRateLimited {
				// Rate limited - wait for window reset and continue
				log.Info().
					Int("total_synced", totalSynced).
					Int("batches", batchNum).
					Msg("zone sync hit rate limit, waiting for window reset")
				if err := client.WaitForRateLimit(ctx); err != nil {
					log.Info().Err(err).Msg("zone sync cancelled while waiting for rate limit")
					return
				}
				continue
			}
			log.Error().Err(err).Int("batch", batchNum).Msg("zone sync batch failed")
			return
		}

		// Update remaining count
		withoutZones -= int64(synced)

		// Log progress every batch
		rateLimit = client.GetRateLimit()
		log.Info().
			Int("batch", batchNum).
			Int("synced", synced).
			Int("total_synced", totalSynced).
			Int64("remaining", withoutZones).
			Str("15min_usage", fmt.Sprintf("%d/%d", rateLimit.Usage15Min, rateLimit.Limit15Min)).
			Str("daily_usage", fmt.Sprintf("%d/%d", rateLimit.UsageDaily, rateLimit.LimitDaily)).
			Msg("zone sync batch completed")

		// If we synced less than batch size, we're probably done or hit errors
		if synced < z.batchSize {
			break
		}

		// If no more activities need syncing, we're done
		if withoutZones <= 0 {
			log.Info().Int("total_synced", totalSynced).Msg("zone sync complete - all activities synced")
			break
		}

		// Small delay between batches to be nice to the API
		select {
		case <-ctx.Done():
			return
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// LogDatabaseStats logs current database statistics
func LogDatabaseStats(ctx context.Context, queries *db.Queries) {
	log := logging.Logger

	count, err := queries.CountActivities(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("failed to count activities")
		return
	}

	if count == 0 {
		log.Info().Int64("total_activities", 0).Msg("database statistics")
		return
	}

	newestRaw, _ := queries.GetLatestActivityDate(ctx)
	oldestRaw, _ := queries.GetOldestActivityDate(ctx)
	withZones, _ := queries.CountActivitiesWithZones(ctx)
	withoutZones, _ := queries.CountActivitiesWithoutZones(ctx)

	newest := formatDate(newestRaw)
	oldest := formatDate(oldestRaw)

	log.Info().
		Int64("total_activities", count).
		Str("newest_activity", newest).
		Str("oldest_activity", oldest).
		Int64("with_zones", withZones).
		Int64("without_zones", withoutZones).
		Msg("database statistics")
}

func formatDate(raw interface{}) string {
	if raw == nil {
		return "unknown"
	}
	switch v := raw.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case sql.NullTime:
		if v.Valid {
			return v.Time.Format(time.RFC3339)
		}
	case time.Time:
		return v.Format(time.RFC3339)
	}
	return "unknown"
}
