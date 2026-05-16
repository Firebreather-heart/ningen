package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ningen/internal/llm"
)

// CriticResponse represents the verdict and feedback from the Critic node.
type CriticResponse struct {
	Verdict  string `json:"verdict"`  // "PASS" or "FAIL"
	Feedback string `json:"feedback"` // Suggestions for improvement if FAIL
}

// Critic creates a node function that performs behavioral fidelity QA.
// It checks the draft review against the user's historical patterns and assigns a verdict.
func Critic(model llm.LLMProvider) func(context.Context, AgentState) (AgentState, error) {
	return func(ctx context.Context, state AgentState) (AgentState, error) {
		if state.UserProfile == nil {
			return state, fmt.Errorf("user profile is missing")
		}
		if state.DraftReview == "" {
			return state, fmt.Errorf("draft review is missing")
		}

		messages := buildCriticPrompt(state.UserProfile, state.DraftReview, state.TargetProduct, state.UserHistory)

		response, err := model.Complete(ctx, messages, llm.WithJSONSchemaResponse("critic_response", buildCriticSchema()))
		if err != nil {
			return state, fmt.Errorf("critic LLM call failed: %w", err)
		}

		verdict, feedback := parseCriticResponse(response, state.DraftReview)

		state.Iterations++

		state.CriticVerdict = verdict
		state.CriticFeedback = feedback

		if verdict == "PASS" {
			state.FinalReview = state.DraftReview
		}

		return state, nil
	}
}

// buildCriticPrompt constructs the prompt for behavioral fidelity QA.
func buildCriticPrompt(profile *UserProfile, draftReview string, product TargetProduct, history []HistoryEntry) []llm.Message {
	historyStr := buildHistorySample(history, 3)

	userPrompt := fmt.Sprintf(`You are a behavioral consistency auditor. Your task is to evaluate whether a generated review authentically matches the behavioral profile and voice of a specific user.

USER BEHAVIORAL PROFILE:
%s

SAMPLE OF USER'S PAST REVIEWS:
%s

GENERATED REVIEW TO EVALUATE:
"%s"

TARGET PRODUCT:
%s

EVALUATION CRITERIA:
1. Tone Consistency: Does the generated review match the user's tone profile?
2. Detail Level: Is the review's length and detail appropriate for the user's style?
3. Language Authenticity: Are there red flags like generic AI phrases ("delve", "tapestry", "curate", "elevate", "journey", "experience")?
4. Topic Alignment: Does the review discuss topics relevant to this user's interests?
5. Behavioral Markers: Are the user's key behavioral markers (price-consciousness, quality focus, etc.) naturally represented?
6. Rating Consistency: Does the target rating align with the user's typical assessment patterns for this category?
7. Emotional Expression: Does the emotional tone match what you'd expect from this user?

VERDICT RULES:
- Return "PASS" if the review is highly authentic and matches the user's voice
- Return "FAIL" if you detect significant inconsistencies, generic AI language, or behavioral mismatches

Respond with a JSON object (no markdown code blocks, just raw JSON):
{
  "verdict": "PASS" or "FAIL",
  "feedback": "If FAIL, provide specific suggestions for improvement. If PASS, leave empty or confirm authenticity."
}

Provide ONLY the JSON response, no additional text.`, formatStructuredProfile(profile), historyStr, draftReview, formatStructuredProduct(&product))

	return buildMessages(
		"You are a behavioral consistency auditor. Decide whether the review authentically matches the user's voice and return only JSON.",
		userPrompt,
	)
}

// buildHistorySample selects a sample of past reviews to include in the prompt.
func buildHistorySample(history []HistoryEntry, sampleSize int) string {
	if len(history) == 0 {
		return "No past reviews available."
	}

	if sampleSize > len(history) {
		sampleSize = len(history)
	}

	// Take the last sampleSize reviews (most recent)
	start := max(len(history)-sampleSize, 0)

	var sb strings.Builder
	for i := start; i < len(history); i++ {
		entry := history[i]
		fmt.Fprintf(&sb, "Review %d (%s):\n", i-start+1, entry.ProductName)
		fmt.Fprintf(&sb, "Rating: %.1f stars\n", entry.StarRating)
		fmt.Fprintf(&sb, "Text: %s\n\n", entry.ReviewText)
	}
	return sb.String()
}

// parseCriticResponse parses the critic's response and performs validation checks.
func parseCriticResponse(responseText string, draftReview string) (string, string) {
	// Try to extract JSON
	jsonStr := extractJSON(responseText)

	var criticResp CriticResponse
	err := jsonUnmarshal(jsonStr, &criticResp)

	// If JSON parsing fails or verdict is unpredictable, use local checks
	if err != nil || !isValidVerdict(criticResp.Verdict) {
		verdict := localCheckForAISpeak(draftReview)
		if verdict == "FAIL" {
			return "FAIL", "Review contains generic AI language. Avoid phrases like: delve, tapestry, curate, elevate, journey, experience. Use authentic, conversational language."
		}
		return "PASS", ""
	}

	// Validate response against schema
	if err := validateCriticResponse(&criticResp); err != nil {
		// On validation error, fall back to local checks
		verdict := localCheckForAISpeak(draftReview)
		if verdict == "FAIL" {
			return "FAIL", "Review contains generic AI language. Avoid phrases like: delve, tapestry, curate, elevate, journey, experience. Use authentic, conversational language."
		}
		return "PASS", ""
	}

	return criticResp.Verdict, criticResp.Feedback
}

// isValidVerdict checks if a verdict is one of the allowed values.
func isValidVerdict(verdict string) bool {
	return verdict == "PASS" || verdict == "FAIL"
}

// validateCriticResponse validates a CriticResponse against the structured schema.
func validateCriticResponse(criticResp *CriticResponse) error {
	if !isValidVerdict(criticResp.Verdict) {
		return fmt.Errorf("verdict must be 'PASS' or 'FAIL', got %q", criticResp.Verdict)
	}

	if criticResp.Verdict == "FAIL" && criticResp.Feedback == "" {
		return fmt.Errorf("feedback is required when verdict is 'FAIL'")
	}

	return nil
}

// localCheckForAISpeak performs a local check for generic AI language patterns.
func localCheckForAISpeak(text string) string {
	lowerText := strings.ToLower(text)

	// Common AI red flags
	redFlags := []string{
		"delve", "tapestry", "curate", "elevate",
		"journey", "seamless", "elevate", "paradigm",
		"game-changer", "must-have", "can't-miss",
		"unlock", "transform", "revolutionize",
	}

	for _, flag := range redFlags {
		if strings.Contains(lowerText, flag) {
			return "FAIL"
		}
	}

	// Check for excessive marketing language
	marketingPhrases := []string{
		"highly recommend", "5/5", "best ever",
		"absolutely amazing", "perfectly perfect",
	}

	flagCount := 0
	for _, phrase := range marketingPhrases {
		if strings.Contains(lowerText, phrase) {
			flagCount++
		}
	}

	if flagCount > 2 {
		return "FAIL"
	}

	return "PASS"
}

// jsonUnmarshal is a helper for unmarshaling JSON using the encoding/json package.
func jsonUnmarshal(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// buildCriticSchema constructs the strict response schema for the Critic LLM call.
func buildCriticSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"verdict": map[string]any{
				"type":        "string",
				"enum":        []string{"PASS", "FAIL"},
				"description": "Whether the draft review authentically matches the user's behavioral profile",
			},
			"feedback": map[string]any{
				"type":        "string",
				"description": "If FAIL, specific suggestions for improvement. If PASS, confirmation of authenticity.",
			},
		},
		"required":             []string{"verdict", "feedback"},
		"additionalProperties": false,
	}
}
