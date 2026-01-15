package server

import (
	"context"
	"fmt"

	"github.com/joshdurbin/strava-mcp/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPrompts registers all MCP prompts for the server
func (s *Server) registerPrompts() {
	logging.Debug("Registering MCP prompts")

	// Weekly review prompt
	s.mcp.AddPrompt(&mcp.Prompt{
		Name:        "weekly_review",
		Description: "Generate a comprehensive weekly training review with insights and recommendations",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "week",
				Description: "Which week to review: 'current', 'last', or ISO format 'YYYY-Www' (e.g., '2024-W03')",
				Required:    false,
			},
		},
	}, s.weeklyReviewPrompt)

	// Progress check prompt
	s.mcp.AddPrompt(&mcp.Prompt{
		Name:        "progress_check",
		Description: "Analyze training progress and trends over time with actionable insights",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "metric",
				Description: "Primary metric to analyze: 'pace', 'distance', 'duration', or 'elevation'",
				Required:    false,
			},
			{
				Name:        "timeframe",
				Description: "Analysis period: 'last_30_days', 'last_90_days', 'last_6_months', or 'last_year'",
				Required:    false,
			},
		},
	}, s.progressCheckPrompt)

	// Zone analysis prompt
	s.mcp.AddPrompt(&mcp.Prompt{
		Name:        "zone_analysis",
		Description: "Analyze training intensity distribution across heart rate or power zones",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "type",
				Description: "Activity type to analyze (e.g., 'Run', 'Ride'). Leave empty for all types.",
				Required:    false,
			},
		},
	}, s.zoneAnalysisPrompt)

	// Personal records check prompt
	s.mcp.AddPrompt(&mcp.Prompt{
		Name:        "pr_check",
		Description: "Review personal bests and analyze recent achievements",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "type",
				Description: "Activity type to check PRs for (e.g., 'Run', 'Ride'). Leave empty for all types.",
				Required:    false,
			},
		},
	}, s.prCheckPrompt)

	logging.Debug("MCP prompts registered", "count", 4)
}

// weeklyReviewPrompt generates a prompt for comprehensive weekly training review
func (s *Server) weeklyReviewPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	week := "current"
	if req.Params.Arguments != nil {
		if w, ok := req.Params.Arguments["week"]; ok && w != "" {
			week = w
		}
	}

	logging.Info("MCP prompt requested", "prompt", "weekly_review", "week", week)

	promptText := fmt.Sprintf(`Please provide a comprehensive review of my %s week's training.

Use the following tools to gather data:
1. **get_week_summary** with week="%s" to get the weekly overview
2. **find_activities** with the week's date range to see all activities
3. **analyze_zones** to check training intensity distribution (if zone data is available)

Then provide:
- **Summary**: Activities completed, total distance, duration, and elevation
- **Intensity Analysis**: Time spent in different training zones (easy vs hard)
- **Highlights**: Notable achievements or personal bests
- **Recovery Check**: Signs of overtraining or adequate recovery
- **Recommendations**: Suggestions for the coming week based on the data

Please be specific with numbers and use the actual data from the tools.`, week, week)

	return &mcp.GetPromptResult{
		Description: "Weekly training review prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: promptText},
			},
		},
	}, nil
}

// progressCheckPrompt generates a prompt for progress analysis
func (s *Server) progressCheckPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	metric := "pace"
	timeframe := "last_90_days"

	if req.Params.Arguments != nil {
		if m, ok := req.Params.Arguments["metric"]; ok && m != "" {
			metric = m
		}
		if t, ok := req.Params.Arguments["timeframe"]; ok && t != "" {
			timeframe = t
		}
	}

	logging.Info("MCP prompt requested", "prompt", "progress_check", "metric", metric, "timeframe", timeframe)

	promptText := fmt.Sprintf(`Please analyze my training progress focusing on %s over the %s.

Use the following tools to gather data:
1. **analyze_progress** with metric="%s" and timeframe="%s" for the core trend analysis
2. **get_training_summary** to get overall statistics for context
3. **get_personal_records** to see if I'm approaching any PRs

Then provide:
- **Trend Summary**: Is my %s improving, stable, or declining?
- **Percentage Change**: Quantify the improvement or decline
- **Contributing Factors**: What activities or patterns are driving the trend?
- **Comparison to Goals**: How does this progress compare to typical improvement rates?
- **Action Items**: Specific recommendations to continue improving or reverse a decline

Use specific numbers from the data and explain what they mean for my training.`, metric, timeframe, metric, timeframe, metric)

	return &mcp.GetPromptResult{
		Description: "Training progress analysis prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: promptText},
			},
		},
	}, nil
}

// zoneAnalysisPrompt generates a prompt for training zone analysis
func (s *Server) zoneAnalysisPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	activityType := ""
	typeDescription := "all activities"

	if req.Params.Arguments != nil {
		if t, ok := req.Params.Arguments["type"]; ok && t != "" {
			activityType = t
			typeDescription = t + " activities"
		}
	}

	logging.Info("MCP prompt requested", "prompt", "zone_analysis", "type", activityType)

	typeParam := ""
	if activityType != "" {
		typeParam = fmt.Sprintf(`, type="%s"`, activityType)
	}

	promptText := fmt.Sprintf(`Please analyze my training intensity distribution across heart rate zones for %s.

Use the following tools to gather data:
1. **analyze_zones** with zone_type="heartrate"%s for the zone distribution
2. **get_training_summary**%s for overall volume context
3. **check_training_load** to understand recent training patterns

Then provide:
- **Zone Distribution**: Percentage of time in each zone (Z1-Z5)
- **80/20 Rule Check**: Am I following the recommended ~80%% easy / ~20%% hard distribution?
- **Aerobic Base**: Is there sufficient Zone 2 training for aerobic development?
- **High Intensity**: Is there adequate Zone 4-5 work for fitness gains?
- **Recommendations**: How should I adjust my training intensity?

Note: Zone data requires a Strava Summit subscription. If zone data is unavailable, explain this to the user.`, typeDescription, typeParam, typeParam)

	return &mcp.GetPromptResult{
		Description: "Heart rate zone analysis prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: promptText},
			},
		},
	}, nil
}

// prCheckPrompt generates a prompt for personal records review
func (s *Server) prCheckPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	activityType := ""
	typeDescription := "all activities"

	if req.Params.Arguments != nil {
		if t, ok := req.Params.Arguments["type"]; ok && t != "" {
			activityType = t
			typeDescription = t
		}
	}

	logging.Info("MCP prompt requested", "prompt", "pr_check", "type", activityType)

	typeParam := ""
	if activityType != "" {
		typeParam = fmt.Sprintf(`type="%s"`, activityType)
	}

	promptText := fmt.Sprintf(`Please review my personal records and recent achievements for %s.

Use the following tools to gather data:
1. **get_personal_records**%s to get all my PRs
2. **find_activities** with query="fastest"%s to see recent fast activities
3. **analyze_progress** with metric="pace" to see if I'm trending toward new PRs

Then provide:
- **Current PRs**: List my personal bests in each category (fastest, longest, etc.)
- **When Set**: Note when each PR was achieved
- **Trending Toward PRs**: Am I close to breaking any records based on recent performance?
- **Achievement Analysis**: What made those PR performances special?
- **PR Strategy**: Recommendations for targeting specific records

Celebrate achievements and provide motivation for chasing new PRs!`, typeDescription, wrapParam(typeParam), wrapParam(typeParam))

	return &mcp.GetPromptResult{
		Description: "Personal records review prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: promptText},
			},
		},
	}, nil
}

// wrapParam adds " with " prefix if param is not empty
func wrapParam(param string) string {
	if param == "" {
		return ""
	}
	return " with " + param
}
