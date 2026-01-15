package strava

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")

	if client.accessToken != "test-token" {
		t.Errorf("expected access token 'test-token', got '%s'", client.accessToken)
	}
	if client.baseURL != baseURL {
		t.Errorf("expected base URL '%s', got '%s'", baseURL, client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("expected http client to be initialized")
	}
}

func TestFetchAllActivities(t *testing.T) {
	page1Activities := []Activity{
		{ID: 1, Name: "Morning Run", Distance: 5000, Type: "Run"},
		{ID: 2, Name: "Evening Ride", Distance: 20000, Type: "Ride"},
	}
	page2Activities := []Activity{
		{ID: 3, Name: "Swim", Distance: 1500, Type: "Swim"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}

		// Return activities based on page parameter
		page := r.URL.Query().Get("page")
		var activities []Activity
		switch page {
		case "1":
			activities = page1Activities
		case "2":
			activities = page2Activities
		default:
			activities = []Activity{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Limit", "100,1000")
		w.Header().Set("X-RateLimit-Usage", "5,50")
		json.NewEncoder(w).Encode(activities)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL).
		WithRetryConfig(3, 10*time.Millisecond, 50*time.Millisecond)

	var progressCalls []FetchResult

	ctx := context.Background()
	activities, err := client.FetchAllActivities(ctx, func(result FetchResult) {
		progressCalls = append(progressCalls, result)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(activities) != 3 {
		t.Errorf("expected 3 activities, got %d", len(activities))
	}

	// Progress is called for each page including the final empty page
	if len(progressCalls) != 3 {
		t.Errorf("expected 3 progress calls (2 with data + 1 empty), got %d", len(progressCalls))
	}

	// Verify progress callbacks
	if progressCalls[0].Page != 1 || len(progressCalls[0].Activities) != 2 || progressCalls[0].TotalFetched != 2 {
		t.Errorf("unexpected first progress call: page=%d, activities=%d, total=%d",
			progressCalls[0].Page, len(progressCalls[0].Activities), progressCalls[0].TotalFetched)
	}
	if progressCalls[1].Page != 2 || len(progressCalls[1].Activities) != 1 || progressCalls[1].TotalFetched != 3 {
		t.Errorf("unexpected second progress call: page=%d, activities=%d, total=%d",
			progressCalls[1].Page, len(progressCalls[1].Activities), progressCalls[1].TotalFetched)
	}
	if progressCalls[2].Page != 3 || len(progressCalls[2].Activities) != 0 || progressCalls[2].TotalFetched != 3 {
		t.Errorf("unexpected third progress call: page=%d, activities=%d, total=%d",
			progressCalls[2].Page, len(progressCalls[2].Activities), progressCalls[2].TotalFetched)
	}

	// Verify rate limit info was parsed
	if progressCalls[0].RateLimit.Limit15Min != 100 {
		t.Errorf("expected 15min limit 100, got %d", progressCalls[0].RateLimit.Limit15Min)
	}
	if progressCalls[0].RateLimit.Usage15Min != 5 {
		t.Errorf("expected 15min usage 5, got %d", progressCalls[0].RateLimit.Usage15Min)
	}
}

func TestFetchAllActivitiesUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("invalid-token", server.URL).
		WithRetryConfig(2, 10*time.Millisecond, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.FetchAllActivities(ctx, nil)

	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}

func TestFetchAllActivitiesContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activities := []Activity{{ID: 1, Name: "Test"}}
		json.NewEncoder(w).Encode(activities)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL).
		WithRetryConfig(3, 10*time.Millisecond, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should eventually timeout because the server keeps returning activities
	_, err := client.FetchAllActivities(ctx, nil)

	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestFetchAllActivitiesRateLimited(t *testing.T) {
	rateLimitedCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")

		// Return 429 for first two calls to page 1, then succeed
		if page == "1" {
			rateLimitedCalls++
			if rateLimitedCalls <= 2 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			activities := []Activity{{ID: 1, Name: "Test"}}
			json.NewEncoder(w).Encode(activities)
			return
		}

		// Page 2 returns empty to end pagination
		json.NewEncoder(w).Encode([]Activity{})
	}))
	defer server.Close()

	// Use short backoff times for faster tests
	client := NewClientWithBaseURL("test-token", server.URL).
		WithRetryConfig(5, 10*time.Millisecond, 100*time.Millisecond)

	var progressCalls []FetchResult
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	activities, err := client.FetchAllActivities(ctx, func(result FetchResult) {
		progressCalls = append(progressCalls, result)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(activities) != 1 {
		t.Errorf("expected 1 activity, got %d", len(activities))
	}

	// Should have had retries for page 1
	if rateLimitedCalls < 3 {
		t.Errorf("expected at least 3 calls to page 1 (2 rate-limited + 1 success), got %d", rateLimitedCalls)
	}
}

func TestActivityJSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": 12345,
		"name": "Morning Run",
		"distance": 5000.5,
		"moving_time": 1800,
		"elapsed_time": 2000,
		"total_elevation_gain": 50.5,
		"type": "Run",
		"sport_type": "Run",
		"start_date": "2024-01-15T08:00:00Z",
		"start_date_local": "2024-01-15T09:00:00Z",
		"timezone": "(GMT+01:00) Europe/Paris",
		"average_speed": 2.78,
		"max_speed": 4.5,
		"average_cadence": 85.5,
		"average_heartrate": 145.0,
		"max_heartrate": 175.0,
		"kilojoules": 350.0
	}`

	var activity Activity
	err := json.Unmarshal([]byte(jsonData), &activity)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if activity.ID != 12345 {
		t.Errorf("expected ID 12345, got %d", activity.ID)
	}
	if activity.Name != "Morning Run" {
		t.Errorf("expected name 'Morning Run', got '%s'", activity.Name)
	}
	if activity.Distance != 5000.5 {
		t.Errorf("expected distance 5000.5, got %f", activity.Distance)
	}
	if activity.Kilojoules != 350.0 {
		t.Errorf("expected kilojoules 350.0, got %f", activity.Kilojoules)
	}
}

func TestTimeUntilNext15MinWindow(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		minWait  time.Duration
		maxWait  time.Duration
	}{
		{
			name:    "at minute 0, should wait ~15 minutes",
			time:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			minWait: 14 * time.Minute,
			maxWait: 16 * time.Minute,
		},
		{
			name:    "at minute 14, should wait ~1 minute",
			time:    time.Date(2024, 1, 15, 10, 14, 0, 0, time.UTC),
			minWait: 30 * time.Second,
			maxWait: 2 * time.Minute,
		},
		{
			name:    "at minute 15, should wait ~15 minutes",
			time:    time.Date(2024, 1, 15, 10, 15, 0, 0, time.UTC),
			minWait: 14 * time.Minute,
			maxWait: 16 * time.Minute,
		},
		{
			name:    "at minute 30, should wait ~15 minutes",
			time:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			minWait: 14 * time.Minute,
			maxWait: 16 * time.Minute,
		},
		{
			name:    "at minute 44, should wait ~1 minute",
			time:    time.Date(2024, 1, 15, 10, 44, 0, 0, time.UTC),
			minWait: 30 * time.Second,
			maxWait: 2 * time.Minute,
		},
		{
			name:    "at minute 59, should wait ~1 minute (until :00)",
			time:    time.Date(2024, 1, 15, 10, 59, 0, 0, time.UTC),
			minWait: 30 * time.Second,
			maxWait: 2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wait := timeUntilNext15MinWindow(tt.time)
			if wait < tt.minWait || wait > tt.maxWait {
				t.Errorf("timeUntilNext15MinWindow(%v) = %v, want between %v and %v",
					tt.time.Format("15:04:05"), wait, tt.minWait, tt.maxWait)
			}
		})
	}
}

func TestTimeUntilMidnightUTC(t *testing.T) {
	tests := []struct {
		name    string
		time    time.Time
		minWait time.Duration
		maxWait time.Duration
	}{
		{
			name:    "at midnight UTC, should wait ~24 hours",
			time:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			minWait: 23*time.Hour + 59*time.Minute,
			maxWait: 24*time.Hour + 1*time.Minute,
		},
		{
			name:    "at noon UTC, should wait ~12 hours",
			time:    time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			minWait: 11*time.Hour + 59*time.Minute,
			maxWait: 12*time.Hour + 1*time.Minute,
		},
		{
			name:    "at 23:59 UTC, should wait ~1 minute",
			time:    time.Date(2024, 1, 15, 23, 59, 0, 0, time.UTC),
			minWait: 30 * time.Second,
			maxWait: 2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wait := timeUntilMidnightUTC(tt.time)
			if wait < tt.minWait || wait > tt.maxWait {
				t.Errorf("timeUntilMidnightUTC(%v) = %v, want between %v and %v",
					tt.time.Format("15:04:05"), wait, tt.minWait, tt.maxWait)
			}
		})
	}
}

func TestRateLimitInfoMethods(t *testing.T) {
	// Note: rateLimitBuffer is 5, so "approaching" means usage >= limit - 5
	t.Run("IsApproaching15MinLimit", func(t *testing.T) {
		info := RateLimitInfo{Limit15Min: 100, Usage15Min: 96}
		if !info.IsApproaching15MinLimit() {
			t.Error("expected IsApproaching15MinLimit to be true at 96/100 (4 remaining, buffer is 5)")
		}

		info = RateLimitInfo{Limit15Min: 100, Usage15Min: 95}
		if !info.IsApproaching15MinLimit() {
			t.Error("expected IsApproaching15MinLimit to be true at 95/100 (5 remaining, buffer is 5)")
		}

		info = RateLimitInfo{Limit15Min: 100, Usage15Min: 94}
		if info.IsApproaching15MinLimit() {
			t.Error("expected IsApproaching15MinLimit to be false at 94/100 (6 remaining, buffer is 5)")
		}

		info = RateLimitInfo{Limit15Min: 100, Usage15Min: 50}
		if info.IsApproaching15MinLimit() {
			t.Error("expected IsApproaching15MinLimit to be false at 50/100")
		}

		info = RateLimitInfo{Limit15Min: 0, Usage15Min: 100}
		if info.IsApproaching15MinLimit() {
			t.Error("expected IsApproaching15MinLimit to be false when limit is 0")
		}
	})

	t.Run("IsApproachingDailyLimit", func(t *testing.T) {
		info := RateLimitInfo{LimitDaily: 1000, UsageDaily: 996}
		if !info.IsApproachingDailyLimit() {
			t.Error("expected IsApproachingDailyLimit to be true at 996/1000")
		}

		info = RateLimitInfo{LimitDaily: 1000, UsageDaily: 500}
		if info.IsApproachingDailyLimit() {
			t.Error("expected IsApproachingDailyLimit to be false at 500/1000")
		}
	})
}

func TestFetchActivitiesSince(t *testing.T) {
	activities := []Activity{
		{ID: 1, Name: "Recent Run", Distance: 5000, Type: "Run"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify after parameter is set
		after := r.URL.Query().Get("after")
		if after == "" {
			t.Error("expected 'after' parameter to be set")
		}

		page := r.URL.Query().Get("page")
		if page == "1" {
			json.NewEncoder(w).Encode(activities)
		} else {
			json.NewEncoder(w).Encode([]Activity{})
		}
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL).
		WithRetryConfig(3, 10*time.Millisecond, 50*time.Millisecond)

	since := time.Now().Add(-24 * time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.FetchActivitiesSince(ctx, since, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 activity, got %d", len(result))
	}
}
