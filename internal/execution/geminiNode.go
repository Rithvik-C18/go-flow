package execution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiNode struct {
	apiKey       string
	model        string
	prompt       string
	temperature  float32
	maxTokens    int32
	systemPrompt string
}

func NewGeminiNode(config map[string]any, apiKey string) (*GeminiNode, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini node: GEMINI_API_KEY not configured")
	}

	prompt, ok := config["prompt"].(string)
	if !ok || prompt == "" {
		return nil, fmt.Errorf("gemini node: missing or invalid prompt")
	}

	node := &GeminiNode{
		apiKey: apiKey,
		prompt: prompt,
		model:  "gemini-3.1-flash-lite",
	}

	if model, ok := config["model"].(string); ok && model != "" {
		node.model = model
	}

	if temp, ok := config["temperature"].(float64); ok {
		node.temperature = float32(temp)
	} else {
		node.temperature = 0.7
	}

	if maxTokens, ok := config["maxTokens"].(float64); ok {
		node.maxTokens = int32(maxTokens)
	} else {
		node.maxTokens = 1024
	}

	if systemPrompt, ok := config["systemPrompt"].(string); ok {
		node.systemPrompt = systemPrompt
	}

	return node, nil
}

func (g *GeminiNode) Execute(ctx *ExecutionContext) error {
	ctxGen, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctxGen, option.WithAPIKey(g.apiKey))
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(g.model)
	model.SetTemperature(g.temperature)
	model.SetMaxOutputTokens(g.maxTokens)

	if g.systemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(g.systemPrompt)},
		}
	}

	resolvedPrompt, err := resolveValue(g.prompt, ctx.Input, ctx.Input, ctx.Params)
	if err != nil {
		return fmt.Errorf("gemini node: failed to resolve prompt: %w", err)
	}

	promptStr := fmt.Sprint(resolvedPrompt)

	resp, err := model.GenerateContent(ctxGen, genai.Text(promptStr))
	if err != nil {
		return fmt.Errorf("gemini node: failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("gemini node: no response generated")
	}

	var responseBuilder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseBuilder.WriteString(string(text))
		}
	}

	ctx.Output = map[string]any{
		"response": responseBuilder.String(),
		"model":    g.model,
	}

	return nil
}
