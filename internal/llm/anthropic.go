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

const keywordPrompt = `Extract the key technical search terms from this Architecture Decision Record. Return ONLY a JSON array of lowercase strings. Include: technology names, product names, library names, protocol names, file formats, architectural patterns, and domain-specific technical concepts. Exclude: common English words, verbs, adjectives, and generic software terms like "system", "application", "server" unless they are part of a proper name.

ADR Title: %s

ADR Content:
%s`

// ExtractKeywords uses the LLM to extract technical search terms from an ADR.
func (a *AnthropicLLM) ExtractKeywords(ctx context.Context, title, body string) ([]string, error) {
	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf(keywordPrompt, title, body))),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("keyword extraction: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	text := resp.Content[0].Text
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array in response: %s", text)
	}

	var keywords []string
	if err := json.Unmarshal([]byte(text[start:end+1]), &keywords); err != nil {
		return nil, fmt.Errorf("parsing keywords: %w", err)
	}
	return keywords, nil
}

const classifyPrompt = `Classify the relationship between these two ADRs.

Source ADR: "%s"
Related ADR bullet: "%s"

Respond with exactly one of: supersedes, superseded_by, depends_on, drives, related_to`

// ClassifyRelationship uses the LLM to classify a relationship type from natural language.
func (a *AnthropicLLM) ClassifyRelationship(ctx context.Context, sourceTitle, bulletText string) (string, error) {
	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 32,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf(classifyPrompt, sourceTitle, bulletText))),
		},
	})
	if err != nil {
		return "", fmt.Errorf("classify relationship: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	result := strings.TrimSpace(strings.ToLower(resp.Content[0].Text))
	// Validate against known types.
	switch result {
	case "supersedes", "superseded_by", "depends_on", "drives", "related_to":
		return result, nil
	default:
		// Try to extract a valid type from a longer response.
		for _, t := range []string{"supersedes", "superseded_by", "depends_on", "drives", "related_to"} {
			if strings.Contains(result, t) {
				return t, nil
			}
		}
		return "related_to", nil // safe fallback
	}
}

// Verify AnthropicLLM implements LLM.
var _ LLM = (*AnthropicLLM)(nil)
