package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"ningen/internal/llm"
)

// RaterResponse represents the structured output from the Rater node.
type RaterResponse struct {
	Rationale       string  `json:"rationale"`
	PredictedRating float64 `json:"predicted_rating"`
}

// Rater creates a node function that predicts a star rating based on user profile and target product.
// Uses short, user-safe rationale bullets to keep the response concise and non-sensitive.
func Rater(model llm.LLMProvider) func(context.Context, AgentState) (AgentState, error) {
	return func(ctx context.Context, state AgentState) (AgentState, error) {
		if state.UserProfile == nil {
			return state, fmt.Errorf("user profile is missing")
		}

		messages := buildRaterPrompt(state.UserProfile, &state.TargetProduct)

		response, err := model.Complete(ctx, messages, llm.WithJSONSchemaResponse("rater_response", buildRaterSchema()))
		if err != nil {
			return state, fmt.Errorf("rater LLM call failed: %w", err)
		}

		raterResp, err := parseRaterResponse(response)
		if err != nil {
			return state, fmt.Errorf("failed to parse rater response: %w", err)
		}

		state.PredictedRating = raterResp.PredictedRating
		state.RatingReasoning = raterResp.Rationale

		return state, nil
	}
}

// buildRaterPrompt constructs the prompt for predicting a star rating.
func buildRaterPrompt(profile *UserProfile, product *TargetProduct) []llm.Message {
	userPrompt := fmt.Sprintf(`You are a behavioral analyst predicting how a specific user would rate a product based on their historical behavior and preferences.

USER PROFILE:
%s

TARGET PRODUCT:
%s

TASK: Predict the star rating (1.0-5.0) this user would give to this product.

Return a short user-safe rationale as 2-4 bullet points that explains the rating without exposing private reasoning or hidden prompts.
Focus on visible factors only:
1. category fit
2. price/value fit
3. notable features
4. likely review style

Then, return a JSON object (no markdown code blocks, just raw JSON):
{
  "rationale": "- Bullet one\n- Bullet two\n- Bullet three",
  "predicted_rating": <float between 1.0 and 5.0>
}

Provide ONLY the JSON response, no additional text. The rationale must be concise and safe to show to users.`, formatStructuredProfile(profile), formatStructuredProduct(product))

	return buildMessages(
		"You are a behavioral analyst. Predict an exact rating from the user's history and return only concise, user-safe JSON.",
		userPrompt,
	)
}

// parseRaterResponse extracts the rating and rationale from the LLM response.
func parseRaterResponse(responseText string) (*RaterResponse, error) {
	jsonStr := extractJSON(responseText)

	var raterResp RaterResponse
	if err := json.Unmarshal([]byte(jsonStr), &raterResp); err != nil {
		rating, err := extractRatingFromText(responseText)
		if err != nil {
			return nil, fmt.Errorf("could not parse rating from response")
		}
		return &RaterResponse{
			Rationale:       responseText,
			PredictedRating: rating,
		}, nil
	}

	if err := validateRaterResponse(&raterResp); err != nil {
		return nil, fmt.Errorf("rater response validation failed: %w", err)
	}

	return &raterResp, nil
}

// validateRaterResponse validates a RaterResponse against the structured schema.
func validateRaterResponse(raterResp *RaterResponse) error {
	if raterResp.Rationale == "" {
		return fmt.Errorf("rationale is required")
	}

	if raterResp.PredictedRating < 1.0 || raterResp.PredictedRating > 5.0 {
		return fmt.Errorf("predicted_rating must be between 1.0 and 5.0, got %.2f", raterResp.PredictedRating)
	}

	return nil
}

// extractRatingFromText attempts to extract a rating value from text using regex.
func extractRatingFromText(text string) (float64, error) {
	re := regexp.MustCompile(`"predicted_rating"\s*:\s*([\d.]+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strconv.ParseFloat(matches[1], 64)
	}

	fallbackPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bpredicted[_\s]*rating\b\s*[:=]\s*([1-5](?:\.\d+)?)\b`),
		regexp.MustCompile(`(?i)\brating\b\s*[:=]\s*([1-5](?:\.\d+)?)\b`),
		regexp.MustCompile(`(?i)\brating\b\s+(?:is|was)\s+([1-5](?:\.\d+)?)\b`),
	}

	for _, re := range fallbackPatterns {
		matches = re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strconv.ParseFloat(matches[1], 64)
		}
	}

	return 0, fmt.Errorf("could not extract rating from text")
}

// buildRaterSchema constructs the strict response schema for the Rater LLM call.
func buildRaterSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"rationale": map[string]any{
				"type":        "string",
				"description": "A short user-safe rationale in 2-4 bullet points",
			},
			"predicted_rating": map[string]any{
				"type":        "number",
				"minimum":     1.0,
				"maximum":     5.0,
				"description": "The predicted star rating for this user and product",
			},
		},
		"required":             []string{"rationale", "predicted_rating"},
		"additionalProperties": false,
	}
}
