package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joshdurbin/strava-mcp/internal/db"
	"github.com/joshdurbin/strava-mcp/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerResources registers all MCP resources for the server
func (s *Server) registerResources() {
	logging.Debug("Registering MCP resources")

	// Static resource: Latest activity
	s.mcp.AddResource(&mcp.Resource{
		URI:         "strava://activities/latest",
		Name:        "latest_activity",
		Description: "The most recent Strava activity with full details",
		MIMEType:    "application/json",
	}, s.readLatestActivity)

	// Static resource: Current week summary
	s.mcp.AddResource(&mcp.Resource{
		URI:         "strava://summary/week/current",
		Name:        "current_week_summary",
		Description: "Training summary for the current week including activities and totals",
		MIMEType:    "application/json",
	}, s.readCurrentWeekSummary)

	// Static resource: Personal records
	s.mcp.AddResource(&mcp.Resource{
		URI:         "strava://records/personal",
		Name:        "personal_records",
		Description: "Personal bests across all record categories",
		MIMEType:    "application/json",
	}, s.readPersonalRecords)

	// Resource template: Activity by ID
	s.mcp.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "strava://activities/{id}",
		Name:        "activity_by_id",
		Description: "Fetch a specific activity by its Strava ID",
		MIMEType:    "application/json",
	}, s.readActivityByID)

	logging.Debug("MCP resources registered", "count", 4)
}

// readLatestActivity returns the most recent activity
func (s *Server) readLatestActivity(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	logging.Info("MCP resource read", "resource", "latest_activity")

	activity, err := s.queries.GetLatestActivity(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      "strava://activities/latest",
						MIMEType: "application/json",
						Text:     `{"error": "No activities found"}`,
					},
				},
			}, nil
		}
		logging.Error("readLatestActivity failed", "error", err)
		return nil, NewDatabaseError(err)
	}

	summary := convertActivity(activity)
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, NewInternalErrorWithCause("failed to marshal activity", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "strava://activities/latest",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// readCurrentWeekSummary returns the current week's training summary
func (s *Server) readCurrentWeekSummary(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	logging.Info("MCP resource read", "resource", "current_week_summary")

	// Calculate current week boundaries
	now := time.Now()
	daysSinceMonday := int(now.Weekday()) - 1
	if daysSinceMonday < 0 {
		daysSinceMonday = 6 // Sunday
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, now.Location())
	weekEnd := weekStart.AddDate(0, 0, 7).Add(-time.Second)

	weekLabel := fmt.Sprintf("%d-W%02d", weekStart.Year(), weekStart.YearDay()/7+1)

	// Get week's activities
	startTime := sql.NullTime{Time: weekStart, Valid: true}
	endTime := sql.NullTime{Time: weekEnd, Valid: true}

	activities, err := s.queries.SearchActivities(ctx, db.SearchActivitiesParams{
		Column1:     sql.NullString{Valid: false},
		Type:        sql.NullString{Valid: false},
		Column3:     startTime,
		StartDate:   startTime,
		Column5:     endTime,
		StartDate_2: endTime,
		Limit:       100,
	})
	if err != nil {
		logging.Error("readCurrentWeekSummary failed", "error", err)
		return nil, NewDatabaseError(err)
	}

	// Calculate totals
	var totalDistance float64
	var totalDuration int64
	var totalCalories float64
	var totalElevation float64

	for _, a := range activities {
		if a.Distance.Valid {
			totalDistance += a.Distance.Float64
		}
		if a.MovingTime.Valid {
			totalDuration += a.MovingTime.Int64
		}
		if a.Calories.Valid {
			totalCalories += a.Calories.Float64
		}
		if a.TotalElevationGain.Valid {
			totalElevation += a.TotalElevationGain.Float64
		}
	}

	summary := WeekSummaryOutput{
		Week:           weekLabel,
		DateRange:      fmt.Sprintf("%s to %s", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02")),
		ActivityCount:  int64(len(activities)),
		TotalDistance:  formatDistance(totalDistance),
		TotalDuration:  formatDuration(totalDuration),
		TotalCalories:  int(totalCalories),
		TotalElevation: fmt.Sprintf("%.0fm", totalElevation),
		Activities:     convertActivities(activities),
	}

	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, NewInternalErrorWithCause("failed to marshal week summary", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "strava://summary/week/current",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// readPersonalRecords returns all personal bests
func (s *Server) readPersonalRecords(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	logging.Info("MCP resource read", "resource", "personal_records")

	queries := s.queries.(RecordsQuerier)
	records := make([]PersonalRecord, 0)

	// Fastest
	if activity, err := queries.GetFastestActivity(ctx); err == nil {
		pace := ""
		if activity.AverageSpeed.Valid && activity.AverageSpeed.Float64 > 0 {
			pace = formatPace(activity.AverageSpeed.Float64)
		}
		records = append(records, PersonalRecord{
			Category:     "fastest",
			Activity:     convertActivity(activity),
			RecordValue:  pace,
			RecordMetric: "pace",
		})
	}

	// Longest distance
	if activity, err := queries.GetLongestDistanceActivity(ctx); err == nil {
		distance := ""
		if activity.Distance.Valid && activity.Distance.Float64 > 0 {
			distance = formatDistance(activity.Distance.Float64)
		}
		records = append(records, PersonalRecord{
			Category:     "longest_distance",
			Activity:     convertActivity(activity),
			RecordValue:  distance,
			RecordMetric: "distance",
		})
	}

	// Longest duration
	if activity, err := queries.GetLongestDurationActivity(ctx); err == nil {
		duration := ""
		if activity.MovingTime.Valid && activity.MovingTime.Int64 > 0 {
			duration = formatDuration(activity.MovingTime.Int64)
		}
		records = append(records, PersonalRecord{
			Category:     "longest_duration",
			Activity:     convertActivity(activity),
			RecordValue:  duration,
			RecordMetric: "duration",
		})
	}

	// Highest elevation
	if activity, err := queries.GetHighestElevationActivity(ctx); err == nil {
		elevation := ""
		if activity.TotalElevationGain.Valid && activity.TotalElevationGain.Float64 > 0 {
			elevation = formatElevationHuman(activity.TotalElevationGain.Float64)
		}
		records = append(records, PersonalRecord{
			Category:     "highest_elevation",
			Activity:     convertActivity(activity),
			RecordValue:  elevation,
			RecordMetric: "elevation",
		})
	}

	// Most calories
	if activity, err := queries.GetMostCaloriesActivity(ctx); err == nil {
		calories := ""
		if activity.Calories.Valid && activity.Calories.Float64 > 0 {
			calories = formatCalories(int(activity.Calories.Float64))
		}
		records = append(records, PersonalRecord{
			Category:     "most_calories",
			Activity:     convertActivity(activity),
			RecordValue:  calories,
			RecordMetric: "calories",
		})
	}

	output := GetPersonalRecordsOutput{
		Records:  records,
		Insights: []Insight{{Type: "achievement", Message: generateRecordsInsight(records, "")}},
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, NewInternalErrorWithCause("failed to marshal records", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "strava://records/personal",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// readActivityByID returns a specific activity by its Strava ID
func (s *Server) readActivityByID(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Parse activity ID from URI
	// URI format: strava://activities/{id}
	uri := req.Params.URI
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return nil, NewInvalidInputError("invalid activity URI format")
	}

	idStr := parts[len(parts)-1]
	activityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, NewInvalidInputErrorWithDetails("invalid activity ID", idStr)
	}

	logging.Info("MCP resource read", "resource", "activity_by_id", "id", activityID)

	activity, err := s.queries.GetActivity(ctx, activityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      uri,
						MIMEType: "application/json",
						Text:     fmt.Sprintf(`{"error": "Activity %d not found"}`, activityID),
					},
				},
			}, nil
		}
		logging.Error("readActivityByID failed", "error", err)
		return nil, NewDatabaseError(err)
	}

	summary := convertActivity(activity)
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, NewInternalErrorWithCause("failed to marshal activity", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      uri,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}
