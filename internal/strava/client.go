package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/joshdurbin/strava-mcp/internal/logging"
)

const (
	baseURL        = "https://www.strava.com/api/v3"
	perPage        = 200
	requestTimeout = 30 * time.Second
)

// Default retry settings
const (
	defaultMaxRetries     = 5
	defaultInitialBackoff = 1 * time.Second
	defaultMaxBackoff     = 5 * time.Minute
)

// Activity represents a Strava activity from the API
type Activity struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Distance           float64   `json:"distance"`
	MovingTime         int       `json:"moving_time"`
	ElapsedTime        int       `json:"elapsed_time"`
	TotalElevationGain float64   `json:"total_elevation_gain"`
	Type               string    `json:"type"`
	SportType          string    `json:"sport_type"`
	StartDate          time.Time `json:"start_date"`
	StartDateLocal     time.Time `json:"start_date_local"`
	Timezone           string    `json:"timezone"`
	AverageSpeed       float64   `json:"average_speed"`
	MaxSpeed           float64   `json:"max_speed"`
	AverageCadence     float64   `json:"average_cadence"`
	AverageHeartrate   float64   `json:"average_heartrate"`
	MaxHeartrate       float64   `json:"max_heartrate"`
	Kilojoules         float64   `json:"kilojoules"`
}

// ActivityZone represents zone data from the Strava API
type ActivityZone struct {
	Type                string           `json:"type"` // "heartrate" or "power"
	SensorBased         bool             `json:"sensor_based"`
	DistributionBuckets []TimedZoneRange `json:"distribution_buckets"`
}

// TimedZoneRange represents time spent in a specific zone range
type TimedZoneRange struct {
	Min  int     `json:"min"`
	Max  int     `json:"max"`
	Time float64 `json:"time"` // seconds (API returns float)
}

// RateLimitInfo contains rate limit information from the API
type RateLimitInfo struct {
	Limit15Min    int
	Usage15Min    int
	LimitDaily    int
	UsageDaily    int
	IsRateLimited bool
	// Calculated fields
	TimeUntil15MinReset time.Duration
	TimeUntilDailyReset time.Duration
	RecommendedWait     time.Duration
}

// Buffer to keep from rate limit boundaries (leave room for other operations)
const rateLimitBuffer = 5

// timeUntilNext15MinWindow calculates time until the next 15-minute boundary
// Strava rate limits reset at 0, 15, 30, 45 minutes past each hour
func timeUntilNext15MinWindow(now time.Time) time.Duration {
	// Get current minute within the hour
	minute := now.Minute()
	second := now.Second()
	nano := now.Nanosecond()

	// Find next 15-minute boundary
	nextBoundary := ((minute / 15) + 1) * 15
	if nextBoundary >= 60 {
		nextBoundary = 0
	}

	// Calculate minutes until next boundary
	var minutesUntil int
	if nextBoundary == 0 {
		minutesUntil = 60 - minute
	} else {
		minutesUntil = nextBoundary - minute
	}

	// Convert to duration, subtracting current seconds/nanos
	waitDuration := time.Duration(minutesUntil)*time.Minute -
		time.Duration(second)*time.Second -
		time.Duration(nano)*time.Nanosecond

	// Add a small buffer (2 seconds) to ensure we're past the boundary
	return waitDuration + 2*time.Second
}

// timeUntilMidnightUTC calculates time until midnight UTC (daily reset)
func timeUntilMidnightUTC(now time.Time) time.Duration {
	nowUTC := now.UTC()
	midnight := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day()+1, 0, 0, 0, 0, time.UTC)
	return midnight.Sub(nowUTC) + 2*time.Second // Add 2 second buffer
}

// ShouldWaitForRateLimit checks if we should wait before making more requests
// Returns the recommended wait duration (0 if no wait needed)
func (info *RateLimitInfo) ShouldWaitForRateLimit() time.Duration {
	return info.RecommendedWait
}

// IsApproaching15MinLimit returns true if we're close to the 15-minute limit
func (info *RateLimitInfo) IsApproaching15MinLimit() bool {
	if info.Limit15Min == 0 {
		return false
	}
	return info.Usage15Min >= info.Limit15Min-rateLimitBuffer
}

// IsApproachingDailyLimit returns true if we're close to the daily limit
func (info *RateLimitInfo) IsApproachingDailyLimit() bool {
	if info.LimitDaily == 0 {
		return false
	}
	return info.UsageDaily >= info.LimitDaily-rateLimitBuffer
}

// FetchResult contains the result of a fetch operation
type FetchResult struct {
	Activities   []Activity
	RateLimit    RateLimitInfo
	Page         int
	TotalFetched int
	Error        error
	RetryCount   int
	IsRetrying   bool
}

// ErrPremiumRequired indicates the API endpoint requires Strava Summit subscription
var ErrPremiumRequired = fmt.Errorf("strava summit subscription required")

// ErrRateLimited indicates the API returned a 429 rate limit error
var ErrRateLimited = fmt.Errorf("rate limited")

// Client is a Strava API client with automatic retry and backoff
type Client struct {
	httpClient  *retryablehttp.Client
	accessToken string
	baseURL     string
	rateMu      sync.RWMutex
	rateLimit   RateLimitInfo
}

// RetryConfig holds retry/backoff settings
type RetryConfig struct {
	MaxRetries int
	MinWait    time.Duration
	MaxWait    time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: defaultMaxRetries,
		MinWait:    defaultInitialBackoff,
		MaxWait:    defaultMaxBackoff,
	}
}

// NewClient creates a new Strava API client with automatic retry
func NewClient(accessToken string) *Client {
	return newClientWithConfig(accessToken, baseURL, defaultMaxRetries, defaultInitialBackoff, defaultMaxBackoff)
}

// NewClientWithRetryConfig creates a new Strava API client with custom retry settings
func NewClientWithRetryConfig(accessToken string, cfg RetryConfig) *Client {
	return newClientWithConfig(accessToken, baseURL, cfg.MaxRetries, cfg.MinWait, cfg.MaxWait)
}

// NewClientWithBaseURL creates a new Strava API client with a custom base URL (for testing)
func NewClientWithBaseURL(accessToken, customBaseURL string) *Client {
	return newClientWithConfig(accessToken, customBaseURL, defaultMaxRetries, defaultInitialBackoff, defaultMaxBackoff)
}

func newClientWithConfig(accessToken, baseURL string, maxRetries int, minWait, maxWait time.Duration) *Client {
	log := logging.Logger
	client := retryablehttp.NewClient()
	client.RetryMax = maxRetries
	client.RetryWaitMin = minWait
	client.RetryWaitMax = maxWait
	client.Logger = &logging.LeveledLogger{}

	// Custom retry policy: retry on 429 and 5xx, but not on 402 (premium required)
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// Don't retry on context errors
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		// Retry on connection errors
		if err != nil {
			return true, nil
		}

		// Don't retry on 402 Payment Required (premium feature)
		if resp.StatusCode == http.StatusPaymentRequired {
			return false, nil
		}

		// Don't retry on 404 Not Found
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		// Retry on 429 Too Many Requests (rate limited)
		if resp.StatusCode == http.StatusTooManyRequests {
			return true, nil
		}

		// Retry on 5xx server errors
		if resp.StatusCode >= 500 {
			return true, nil
		}

		return false, nil
	}

	// Custom backoff that waits for rate limit window resets
	client.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		// For rate limit responses, wait until the 15-minute window resets
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			// Check Retry-After header first (if Strava provides it)
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					wait := time.Duration(seconds) * time.Second
					log.Info().
						Dur("wait", wait).
						Int("attempt", attemptNum).
						Msg("rate limited, waiting for Retry-After header")
					return wait
				}
			}

			// Calculate time until next 15-minute window reset
			wait := timeUntilNext15MinWindow(time.Now())
			log.Info().
				Dur("wait", wait).
				Int("attempt", attemptNum).
				Msg("rate limited, waiting for 15-minute window reset")
			return wait
		}

		// For non-rate-limit errors (5xx, connection errors), use exponential backoff
		wait := min * time.Duration(1<<uint(attemptNum))
		if wait > max {
			wait = max
		}
		log.Info().
			Dur("wait", wait).
			Int("attempt", attemptNum).
			Dur("max_wait", max).
			Msg("backing off before retry")
		return wait
	}

	// Hook to log requests
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		if retry > 0 {
			log.Info().
				Str("url", req.URL.Path).
				Int("attempt", retry+1).
				Msg("retrying request")
		}

		// Log request headers at trace level (-vv)
		if logging.IsTraceEnabled() {
			log.Debug().
				Str("method", req.Method).
				Str("url", req.URL.String()).
				Str("headers", formatHeaders(req.Header)).
				Msg("request headers")
		}
	}

	// Hook to log responses and capture rate limit info
	client.ResponseLogHook = func(logger retryablehttp.Logger, resp *http.Response) {
		rateLimit := parseRateLimitHeaders(resp.Header, time.Now())

		// Log response headers at trace level (-vv)
		if logging.IsTraceEnabled() {
			log.Debug().
				Int("status", resp.StatusCode).
				Str("url", resp.Request.URL.Path).
				Str("headers", formatHeaders(resp.Header)).
				Msg("response headers")
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Warn().
				Int("status", resp.StatusCode).
				Str("url", resp.Request.URL.Path).
				Str("15min_usage", fmt.Sprintf("%d/%d", rateLimit.Usage15Min, rateLimit.Limit15Min)).
				Str("daily_usage", fmt.Sprintf("%d/%d", rateLimit.UsageDaily, rateLimit.LimitDaily)).
				Dur("wait_for_reset", rateLimit.TimeUntil15MinReset).
				Msg("rate limited by API")
		}
	}

	return &Client{
		httpClient:  client,
		accessToken: accessToken,
		baseURL:     baseURL,
	}
}

// WithRetryConfig sets custom retry configuration (useful for testing)
func (c *Client) WithRetryConfig(maxRetries int, initialBackoff, maxBackoff time.Duration) *Client {
	c.httpClient.RetryMax = maxRetries
	c.httpClient.RetryWaitMin = initialBackoff
	c.httpClient.RetryWaitMax = maxBackoff
	return c
}

// GetRateLimit returns the current rate limit info (with recalculated reset times)
func (c *Client) GetRateLimit() RateLimitInfo {
	c.rateMu.RLock()
	info := c.rateLimit
	c.rateMu.RUnlock()

	// Recalculate reset times based on current time
	now := time.Now()
	info.TimeUntil15MinReset = timeUntilNext15MinWindow(now)
	info.TimeUntilDailyReset = timeUntilMidnightUTC(now)

	// Recalculate recommended wait
	info.RecommendedWait = 0
	if info.Limit15Min > 0 && info.Usage15Min >= info.Limit15Min {
		info.IsRateLimited = true
		info.RecommendedWait = info.TimeUntil15MinReset
	} else if info.LimitDaily > 0 && info.UsageDaily >= info.LimitDaily {
		info.IsRateLimited = true
		info.RecommendedWait = info.TimeUntilDailyReset
	} else if info.IsApproaching15MinLimit() {
		info.RecommendedWait = info.TimeUntil15MinReset
	} else if info.IsApproachingDailyLimit() {
		info.RecommendedWait = info.TimeUntilDailyReset
	}

	return info
}

// WaitForRateLimit blocks until rate limits allow more requests, or context is cancelled
// Returns nil if ready to proceed, or context error if cancelled
func (c *Client) WaitForRateLimit(ctx context.Context) error {
	log := logging.Logger
	rateLimit := c.GetRateLimit()
	waitDuration := rateLimit.ShouldWaitForRateLimit()

	if waitDuration <= 0 {
		return nil
	}

	log.Info().
		Dur("wait", waitDuration).
		Str("15min_usage", fmt.Sprintf("%d/%d", rateLimit.Usage15Min, rateLimit.Limit15Min)).
		Str("daily_usage", fmt.Sprintf("%d/%d", rateLimit.UsageDaily, rateLimit.LimitDaily)).
		Msg("waiting for rate limit window to reset")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDuration):
		log.Info().Msg("rate limit window reset, resuming")
		return nil
	}
}

// updateRateLimit updates the stored rate limit info from response headers
func (c *Client) updateRateLimit(resp *http.Response) RateLimitInfo {
	rateLimit := parseRateLimitHeaders(resp.Header, time.Now())
	if resp.StatusCode == http.StatusTooManyRequests {
		rateLimit.IsRateLimited = true
	}
	c.rateMu.Lock()
	c.rateLimit = rateLimit
	c.rateMu.Unlock()
	return rateLimit
}

// ProgressCallback is called after each page is fetched
type ProgressCallback func(result FetchResult)

// FetchAllActivities fetches all activities from the authenticated user's account
func (c *Client) FetchAllActivities(ctx context.Context, progress ProgressCallback) ([]Activity, error) {
	var allActivities []Activity
	page := 1

	for {
		activities, rateLimit, err := c.fetchActivitiesPage(ctx, page, 0)

		result := FetchResult{
			Activities:   activities,
			RateLimit:    rateLimit,
			Page:         page,
			TotalFetched: len(allActivities) + len(activities),
		}

		if progress != nil {
			progress(result)
		}

		if err != nil {
			return allActivities, err
		}

		if len(activities) == 0 {
			break
		}

		allActivities = append(allActivities, activities...)
		page++
	}

	return allActivities, nil
}

// FetchActivitiesSince fetches activities since a given timestamp (for delta sync)
func (c *Client) FetchActivitiesSince(ctx context.Context, since time.Time, progress ProgressCallback) ([]Activity, error) {
	var allActivities []Activity
	page := 1
	afterEpoch := since.Unix()

	for {
		activities, rateLimit, err := c.fetchActivitiesPage(ctx, page, afterEpoch)

		result := FetchResult{
			Activities:   activities,
			RateLimit:    rateLimit,
			Page:         page,
			TotalFetched: len(allActivities) + len(activities),
		}

		if progress != nil {
			progress(result)
		}

		if err != nil {
			return allActivities, err
		}

		if len(activities) == 0 {
			break
		}

		allActivities = append(allActivities, activities...)
		page++
	}

	return allActivities, nil
}

// FetchActivityZones fetches zone data for a specific activity
// Note: This endpoint requires Strava Summit (premium) subscription
func (c *Client) FetchActivityZones(ctx context.Context, activityID int64) ([]ActivityZone, error) {
	url := fmt.Sprintf("%s/activities/%d/zones", c.baseURL, activityID)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit info
	rateLimit := c.updateRateLimit(resp)

	// 402 Payment Required - requires Strava Summit subscription
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, ErrPremiumRequired
	}

	// 429 handled by retryablehttp, but if we still get here after retries exhausted
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // No zone data available for this activity
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update rate limit from successful response
	c.updateRateLimit(resp)
	_ = rateLimit // Used above

	var zones []ActivityZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return zones, nil
}

func (c *Client) fetchActivitiesPage(ctx context.Context, page int, after int64) ([]Activity, RateLimitInfo, error) {
	url := fmt.Sprintf("%s/athlete/activities?page=%d&per_page=%d", c.baseURL, page, perPage)
	if after > 0 {
		url += fmt.Sprintf("&after=%d", after)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, RateLimitInfo{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, RateLimitInfo{}, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit info
	rateLimit := c.updateRateLimit(resp)

	if resp.StatusCode == http.StatusTooManyRequests {
		// Retries exhausted
		return nil, rateLimit, ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, rateLimit, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var activities []Activity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, rateLimit, fmt.Errorf("decoding response: %w", err)
	}

	return activities, rateLimit, nil
}

// minPositive returns the minimum of two values, preferring positive values.
// If one value is zero/unset, returns the other. If both are positive, returns the minimum.
func minPositive(a, b int) int {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

func parseRateLimitHeaders(headers http.Header, now time.Time) RateLimitInfo {
	var info RateLimitInfo

	// Strava returns two sets of rate limit headers:
	// 1. X-RateLimit-* - General rate limits (higher)
	// 2. X-ReadRateLimit-* - Read-specific rate limits (lower, more restrictive)
	// Format: "15min_limit,daily_limit" and "15min_usage,daily_usage"
	// We need to use the more restrictive of the two.

	// Parse general rate limits
	var generalLimit15Min, generalLimitDaily int
	if limitHeader := headers.Get("X-RateLimit-Limit"); limitHeader != "" {
		parts := strings.Split(limitHeader, ",")
		if len(parts) >= 1 {
			generalLimit15Min, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 {
			generalLimitDaily, _ = strconv.Atoi(parts[1])
		}
	}

	var generalUsage15Min, generalUsageDaily int
	if usageHeader := headers.Get("X-RateLimit-Usage"); usageHeader != "" {
		parts := strings.Split(usageHeader, ",")
		if len(parts) >= 1 {
			generalUsage15Min, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 {
			generalUsageDaily, _ = strconv.Atoi(parts[1])
		}
	}

	// Parse read-specific rate limits (more restrictive)
	var readLimit15Min, readLimitDaily int
	if limitHeader := headers.Get("X-ReadRateLimit-Limit"); limitHeader != "" {
		parts := strings.Split(limitHeader, ",")
		if len(parts) >= 1 {
			readLimit15Min, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 {
			readLimitDaily, _ = strconv.Atoi(parts[1])
		}
	}

	var readUsage15Min, readUsageDaily int
	if usageHeader := headers.Get("X-ReadRateLimit-Usage"); usageHeader != "" {
		parts := strings.Split(usageHeader, ",")
		if len(parts) >= 1 {
			readUsage15Min, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 {
			readUsageDaily, _ = strconv.Atoi(parts[1])
		}
	}

	// Use the more restrictive limits (minimum of general and read limits)
	// and the higher usage values (maximum, representing worst case)
	info.Limit15Min = minPositive(generalLimit15Min, readLimit15Min)
	info.LimitDaily = minPositive(generalLimitDaily, readLimitDaily)
	info.Usage15Min = max(generalUsage15Min, readUsage15Min)
	info.UsageDaily = max(generalUsageDaily, readUsageDaily)

	// Calculate time until rate limit windows reset
	info.TimeUntil15MinReset = timeUntilNext15MinWindow(now)
	info.TimeUntilDailyReset = timeUntilMidnightUTC(now)

	// Determine recommended wait time based on current usage
	info.RecommendedWait = 0

	// Check if we've hit or are approaching limits
	if info.Limit15Min > 0 && info.Usage15Min >= info.Limit15Min {
		info.IsRateLimited = true
		info.RecommendedWait = info.TimeUntil15MinReset
	} else if info.LimitDaily > 0 && info.UsageDaily >= info.LimitDaily {
		info.IsRateLimited = true
		info.RecommendedWait = info.TimeUntilDailyReset
	} else if info.IsApproaching15MinLimit() {
		// Approaching 15-min limit - recommend waiting for reset
		info.RecommendedWait = info.TimeUntil15MinReset
	} else if info.IsApproachingDailyLimit() {
		// Approaching daily limit - recommend waiting for reset
		info.RecommendedWait = info.TimeUntilDailyReset
	}

	return info
}

// formatHeaders formats HTTP headers for logging, redacting sensitive values
func formatHeaders(headers http.Header) string {
	if len(headers) == 0 {
		return "{}"
	}

	// Get sorted keys for consistent output
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for _, k := range keys {
		if !first {
			sb.WriteString(", ")
		}
		first = false

		// Redact sensitive headers
		value := strings.Join(headers[k], ", ")
		lowerKey := strings.ToLower(k)
		if lowerKey == "authorization" || lowerKey == "cookie" || lowerKey == "set-cookie" {
			value = "[REDACTED]"
		}

		sb.WriteString(fmt.Sprintf("%s: %q", k, value))
	}
	sb.WriteString("}")
	return sb.String()
}
