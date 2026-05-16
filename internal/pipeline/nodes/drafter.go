package nodes

import (
	"context"
	"fmt"
	"strings"

	"ningen/internal/llm"
)

// Drafter creates a node function that generates a persona-driven review.
// It localizes context to Nigerian settings and incorporates feedback from previous iterations.
func Drafter(model llm.LLMProvider) func(context.Context, AgentState) (AgentState, error) {
	return func(ctx context.Context, state AgentState) (AgentState, error) {
		if state.UserProfile == nil {
			return state, fmt.Errorf("user profile is missing")
		}

		localizedProduct := LocalizeContext(&state.TargetProduct)

		messages := buildDrafterPrompt(state.UserProfile, &localizedProduct, state.PredictedRating, state.CriticFeedback)

		response, err := model.Complete(ctx, messages)
		if err != nil {
			return state, fmt.Errorf("drafter LLM call failed: %w", err)
		}

		draftReview := strings.TrimSpace(response)
		state.DraftReview = draftReview

		return state, nil
	}
}

// buildDrafterPrompt constructs the prompt for generating a persona-driven review.
func buildDrafterPrompt(profile *UserProfile, product *TargetProduct, rating float64, criticFeedback string) []llm.Message {
	feedbackSection := ""
	if criticFeedback != "" {
		feedbackSection = fmt.Sprintf("IMPORTANT FEEDBACK FROM PREVIOUS ITERATION:\n%s\n\nPlease address these issues in your revision.", criticFeedback)
	}

	iterationNote := ""
	if criticFeedback != "" {
		iterationNote = "This is a REVISED review addressing feedback. Incorporate the feedback while maintaining authenticity."
	}

	userPrompt := fmt.Sprintf(`You are an AI assistant generating a highly authentic product review that perfectly mimics the behavioral patterns and voice of a specific user.

USER PROFILE TO EMULATE:
%s

PRODUCT TO REVIEW:
%s

REVIEW SPECIFICATIONS:
- Your predicted rating: %.1f stars
- Write in the EXACT voice and style of the user profile
- Match the user's typical review length and detail level
- Incorporate the user's favorite topics and behavioral markers naturally
- Avoid generic AI language (no "delve", "tapestry", "curate", "elevate", etc.)
- Be conversational and authentic, as if the user personally wrote this
- The review should reflect the user's genuine assessment patterns
%s
%s

Generate ONLY the review text, no metadata, no ratings, no markdown formatting.`, formatStructuredProfile(profile), formatStructuredProduct(product), rating, iterationNote, feedbackSection)

	return buildMessages(
		"You write authentic reviews in a specific user's voice and must avoid generic AI language.",
		userPrompt,
	)
}

// boolToString converts a boolean to "use " or "don't use ".
func boolToString(b bool) string {
	if b {
		return "uses "
	}
	return "doesn't use "
}

// formatTopicPreferences formats topic preferences for display in prompt.
func formatTopicPreferences(prefs []TopicPreference) string {
	if len(prefs) == 0 {
		return "None identified"
	}
	var parts []string
	for _, p := range prefs {
		parts = append(parts, fmt.Sprintf("%s (%s, mentioned %d times)", p.Topic, p.Sentiment, p.Frequency))
	}
	return strings.Join(parts, "; ")
}
