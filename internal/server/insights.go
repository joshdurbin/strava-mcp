package server

import (
	"fmt"
	"math"
)

// Insight represents a single AI-friendly insight about the data
type Insight struct {
	Type    string `json:"type"`    // e.g., "trend", "achievement", "warning", "suggestion"
	Message string `json:"message"` // Human-readable insight
}

// SuggestedAction represents a suggested next tool call
type SuggestedAction struct {
	Tool        string `json:"tool"`        // Tool name to call
	Description string `json:"description"` // Why this action is suggested
	Priority    string `json:"priority"`    // "high", "medium", "low"
}

// InsightGenerator provides methods for generating insights from data
type InsightGenerator struct{}

// NewInsightGenerator creates a new insight generator
func NewInsightGenerator() *InsightGenerator {
	return &InsightGenerator{}
}

// GenerateProgressInsights generates insights about progress/trends
func (g *InsightGenerator) GenerateProgressInsights(
	currentValue, previousValue float64,
	metric string,
	higherIsBetter bool,
) []Insight {
	var insights []Insight

	if previousValue == 0 {
		return insights
	}

	changePercent := ((currentValue - previousValue) / previousValue) * 100
	improving := (higherIsBetter && changePercent > 0) || (!higherIsBetter && changePercent < 0)

	absChange := math.Abs(changePercent)

	if absChange < 5 {
		insights = append(insights, Insight{
			Type:    "trend",
			Message: fmt.Sprintf("Your %s is stable (%.1f%% change)", metric, changePercent),
		})
	} else if improving {
		intensity := "improving"
		if absChange > 20 {
			intensity = "significantly improving"
		}
		insights = append(insights, Insight{
			Type:    "achievement",
			Message: fmt.Sprintf("Your %s is %s (%.1f%% better)", metric, intensity, absChange),
		})
	} else {
		intensity := "declining"
		if absChange > 20 {
			intensity = "significantly declining"
		}
		insights = append(insights, Insight{
			Type:    "warning",
			Message: fmt.Sprintf("Your %s is %s (%.1f%% worse)", metric, intensity, absChange),
		})
	}

	return insights
}

// GenerateTrainingLoadInsights generates insights about training load
func (g *InsightGenerator) GenerateTrainingLoadInsights(
	currentWeekVolume, avgWeeklyVolume float64,
	currentWeekActivities, avgWeeklyActivities int64,
) []Insight {
	var insights []Insight

	if avgWeeklyVolume == 0 {
		return insights
	}

	volumeRatio := currentWeekVolume / avgWeeklyVolume

	if volumeRatio > 1.3 {
		insights = append(insights, Insight{
			Type:    "warning",
			Message: fmt.Sprintf("Training volume is %.0f%% above your average - consider recovery", (volumeRatio-1)*100),
		})
	} else if volumeRatio > 1.1 {
		insights = append(insights, Insight{
			Type:    "trend",
			Message: fmt.Sprintf("Training volume is %.0f%% above average - good progressive overload", (volumeRatio-1)*100),
		})
	} else if volumeRatio < 0.7 {
		insights = append(insights, Insight{
			Type:    "suggestion",
			Message: fmt.Sprintf("Training volume is %.0f%% below average - planned recovery or time to ramp up?", (1-volumeRatio)*100),
		})
	} else if volumeRatio < 0.9 {
		insights = append(insights, Insight{
			Type:    "trend",
			Message: fmt.Sprintf("Training volume is slightly below average (%.0f%%)", (1-volumeRatio)*100),
		})
	} else {
		insights = append(insights, Insight{
			Type:    "trend",
			Message: "Training volume is consistent with your average",
		})
	}

	// Activity frequency insight
	if avgWeeklyActivities > 0 {
		activityRatio := float64(currentWeekActivities) / float64(avgWeeklyActivities)
		if activityRatio > 1.5 {
			insights = append(insights, Insight{
				Type:    "trend",
				Message: fmt.Sprintf("Activity frequency is high (%d this week vs %.1f avg)", currentWeekActivities, float64(avgWeeklyActivities)),
			})
		} else if activityRatio < 0.5 && currentWeekActivities > 0 {
			insights = append(insights, Insight{
				Type:    "trend",
				Message: fmt.Sprintf("Activity frequency is low (%d this week vs %.1f avg)", currentWeekActivities, float64(avgWeeklyActivities)),
			})
		}
	}

	return insights
}

// GenerateZoneInsights generates insights about training zone distribution
func (g *InsightGenerator) GenerateZoneInsights(zonePercentages map[int]float64) []Insight {
	var insights []Insight

	// Calculate Zone 2 percentage (aerobic base)
	zone2Pct := zonePercentages[2]
	zone4And5Pct := zonePercentages[4] + zonePercentages[5]

	// 80/20 rule check: ~80% should be easy (Z1-Z2), ~20% hard (Z4-Z5)
	easyZonePct := zonePercentages[1] + zonePercentages[2]

	if easyZonePct < 70 {
		insights = append(insights, Insight{
			Type:    "suggestion",
			Message: fmt.Sprintf("Only %.0f%% of training is in easy zones (Z1-Z2). Consider more aerobic base training.", easyZonePct),
		})
	} else if easyZonePct > 90 {
		insights = append(insights, Insight{
			Type:    "suggestion",
			Message: fmt.Sprintf("%.0f%% of training is in easy zones. Consider adding intensity work for fitness gains.", easyZonePct),
		})
	} else {
		insights = append(insights, Insight{
			Type:    "achievement",
			Message: fmt.Sprintf("Good zone distribution: %.0f%% easy / %.0f%% hard (close to 80/20 rule)", easyZonePct, zone4And5Pct),
		})
	}

	// Zone 2 specific insight
	if zone2Pct > 50 {
		insights = append(insights, Insight{
			Type:    "achievement",
			Message: "Strong aerobic base building with good Zone 2 volume",
		})
	}

	return insights
}

// GenerateComparisonInsights generates insights from period comparisons
func (g *InsightGenerator) GenerateComparisonInsights(
	p1Activities, p2Activities int64,
	p1Distance, p2Distance float64,
	p1Duration, p2Duration int64,
) []Insight {
	var insights []Insight

	// Activity count change
	if p1Activities > 0 {
		activityChange := float64(p2Activities-p1Activities) / float64(p1Activities) * 100
		if activityChange > 20 {
			insights = append(insights, Insight{
				Type:    "achievement",
				Message: fmt.Sprintf("Activity frequency increased by %.0f%%", activityChange),
			})
		} else if activityChange < -20 {
			insights = append(insights, Insight{
				Type:    "warning",
				Message: fmt.Sprintf("Activity frequency decreased by %.0f%%", math.Abs(activityChange)),
			})
		}
	}

	// Distance change
	if p1Distance > 0 {
		distanceChange := (p2Distance - p1Distance) / p1Distance * 100
		if distanceChange > 15 {
			insights = append(insights, Insight{
				Type:    "achievement",
				Message: fmt.Sprintf("Total distance increased by %.0f%%", distanceChange),
			})
		} else if distanceChange < -15 {
			insights = append(insights, Insight{
				Type:    "trend",
				Message: fmt.Sprintf("Total distance decreased by %.0f%%", math.Abs(distanceChange)),
			})
		}
	}

	// Duration change
	if p1Duration > 0 {
		durationChange := float64(p2Duration-p1Duration) / float64(p1Duration) * 100
		if durationChange > 15 {
			insights = append(insights, Insight{
				Type:    "achievement",
				Message: fmt.Sprintf("Training time increased by %.0f%%", durationChange),
			})
		}
	}

	return insights
}

// SuggestNextActions suggests logical next tool calls based on context
func SuggestNextActions(context string) []SuggestedAction {
	suggestions := make([]SuggestedAction, 0)

	switch context {
	case "activities":
		suggestions = append(suggestions,
			SuggestedAction{
				Tool:        "get_training_summary",
				Description: "Get aggregate stats for these activities",
				Priority:    "medium",
			},
			SuggestedAction{
				Tool:        "analyze_zones",
				Description: "See training intensity distribution",
				Priority:    "medium",
			},
		)
	case "training_summary":
		suggestions = append(suggestions,
			SuggestedAction{
				Tool:        "compare_periods",
				Description: "Compare with a previous period",
				Priority:    "high",
			},
			SuggestedAction{
				Tool:        "analyze_progress",
				Description: "Check if you're improving",
				Priority:    "high",
			},
		)
	case "week_summary":
		suggestions = append(suggestions,
			SuggestedAction{
				Tool:        "check_training_load",
				Description: "Analyze training load trends",
				Priority:    "high",
			},
			SuggestedAction{
				Tool:        "compare_periods",
				Description: "Compare with last week or last month",
				Priority:    "medium",
			},
		)
	case "progress":
		suggestions = append(suggestions,
			SuggestedAction{
				Tool:        "get_personal_records",
				Description: "See your all-time bests",
				Priority:    "medium",
			},
			SuggestedAction{
				Tool:        "find_activities",
				Description: "Find specific activities to analyze",
				Priority:    "low",
			},
		)
	case "zones":
		suggestions = append(suggestions,
			SuggestedAction{
				Tool:        "check_training_load",
				Description: "Review overall training load",
				Priority:    "medium",
			},
			SuggestedAction{
				Tool:        "get_activity_zones",
				Description: "Drill into a specific activity's zones",
				Priority:    "low",
			},
		)
	case "records":
		suggestions = append(suggestions,
			SuggestedAction{
				Tool:        "find_activities",
				Description: "Find activities near your records",
				Priority:    "medium",
			},
			SuggestedAction{
				Tool:        "analyze_progress",
				Description: "See if you're trending toward new PRs",
				Priority:    "high",
			},
		)
	case "comparison":
		suggestions = append(suggestions,
			SuggestedAction{
				Tool:        "analyze_progress",
				Description: "Get detailed progress analysis",
				Priority:    "high",
			},
			SuggestedAction{
				Tool:        "check_training_load",
				Description: "Understand training volume changes",
				Priority:    "medium",
			},
		)
	}

	return suggestions
}
