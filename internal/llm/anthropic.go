package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const systemPrompt = `You are an expert on this project's Architecture Decision Records (ADRs).
Answer the user's question using ONLY the provided ADR content.

Rules:
- Base your answer solely on the provided ADR content
- If the ADRs don't contain relevant information, say so clearly
- Do not invent or hallucinate information not present in the ADRs
- Reference ADRs by their number and title in your answer

You will receive the full content of relevant ADRs as context.`

// AnthropicLLM implements the LLM interface using the Anthropic Claude API.
type AnthropicLLM struct {
	client *anthropic.Client
	model  anthropic.Model
}

// NewAnthropicLLM creates a new Anthropic-backed LLM.
// apiKey is required. model should be a valid Anthropic model string.
func NewAnthropicLLM(apiKey, model string) *AnthropicLLM {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicLLM{
		client: &client,
		model:  anthropic.Model(model),
	}
}

// Synthesize sends the query and ADR context to Claude and returns a structured response.
func (a *AnthropicLLM) Synthesize(ctx context.Context, query string, adrContents []ADRContext) (QueryResponse, error) {
	// Build the user message with ADR context.
	var userMsg strings.Builder
	userMsg.WriteString("## Question\n\n")
	userMsg.WriteString(query)
	userMsg.WriteString("\n\n## ADR Context\n\n")

	for _, adr := range adrContents {
		fmt.Fprintf(&userMsg, "### ADR-%03d: %s\n\n%s\n\n---\n\n", adr.Number, adr.Title, adr.Content)
	}

	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg.String())),
		},
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{
				Schema: outputSchema(),
			},
		},
	})
	if err != nil {
		return QueryResponse{}, fmt.Errorf("anthropic synthesis: %w", err)
	}

	// Extract the text content from the response.
	if len(resp.Content) == 0 {
		return QueryResponse{}, fmt.Errorf("empty response from anthropic")
	}

	var result QueryResponse
	for _, block := range resp.Content {
		if block.Type == "text" {
			if err := json.Unmarshal([]byte(block.Text), &result); err != nil {
				return QueryResponse{}, fmt.Errorf("parsing structured response: %w", err)
			}
			return result, nil
		}
	}

	return QueryResponse{}, fmt.Errorf("no text content in anthropic response")
}

func outputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"answer": map[string]any{
				"type":        "string",
				"description": "A synthesized answer to the user's question based on the ADR content",
			},
			"citations": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"adr_number": map[string]any{
							"type":        "integer",
							"description": "The ADR number (e.g., 1 for ADR-001)",
						},
						"title": map[string]any{
							"type":        "string",
							"description": "The title of the ADR",
						},
						"section": map[string]any{
							"type":        "string",
							"description": "The section of the ADR most relevant to this citation",
						},
					},
					"required":             []string{"adr_number", "title", "section"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"answer", "citations"},
		"additionalProperties": false,
	}
}

// Verify AnthropicLLM implements LLM.
var _ LLM = (*AnthropicLLM)(nil)
