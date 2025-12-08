package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
)

var GeminiClient *genai.Client

// InitializeGemini initializes the Gemini AI client using GEMINI_API_KEY.
func InitializeGemini() error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil) // reads GEMINI_API_KEY automatically
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	GeminiClient = client
	return nil
}

// GenerateQuestion generates a math question using Gemini API (new GenAI SDK)
func GenerateQuestion(ctx context.Context, category string, difficulty string) (map[string]interface{}, error) {
	if GeminiClient == nil {
		return nil, fmt.Errorf("gemini client not initialized")
	}

	systemTemplate := `You are a math problem generator.`

	humanTemplate := fmt.Sprintf(
		`Create ONE math problem in JSON format for %s difficulty level in %s category.
Keep answer and explanation brief. Return ONLY valid JSON with these exact fields:
{"id":"","title":"","question":"","solution":{"answer":"","explanation":""},"hints":[],"difficulty":"%s","expectedTime":10,"points":50,"category":"%s","subcategory":"","tags":[],"requirements":[],"imageUrl":""}`,
		difficulty, category, difficulty, category,
	)

	// --- POINTER VALUES ---
	var maxTokens int32 = 2048
	var temperature float32 = 0.4
	zero := int32(0) // disable thinking

	resp, err := GeminiClient.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(humanTemplate),
		&genai.GenerateContentConfig{
			MaxOutputTokens:  maxTokens,
			Temperature:      &temperature,
			ResponseMIMEType: "application/json",

			// Disable thinking (fixes empty MaxTokens responses)
			ThinkingConfig: &genai.ThinkingConfig{
				ThinkingBudget: &zero,
			},

			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemTemplate}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("nil response from Gemini API")
	}

	 // Combine / clean output
    text := strings.TrimSpace(resp.Text())
    if text == "" {
        return nil, fmt.Errorf("empty response from Gemini API")
    }

    // Remove possible ```json fences
    text = strings.TrimPrefix(text, "```json")
    text = strings.TrimPrefix(text, "```")
    text = strings.TrimSuffix(text, "```")
    text = strings.TrimSpace(text)

    // If response has extra content, try to isolate JSON
    start := strings.Index(text, "{")
    end := strings.LastIndex(text, "}")
    if start != -1 && end != -1 && end > start {
        text = text[start : end+1]
    }

	fmt.Println("Generated Question JSON:", text)
    var obj map[string]interface{}
    if err := json.Unmarshal([]byte(text), &obj); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w\nraw: %s", err, text)
    }

    return obj, nil
}
