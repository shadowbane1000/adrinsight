package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const judgeSystemPrompt = `You are an expert evaluator of RAG (Retrieval-Augmented Generation) system answers.

You will receive:
1. A question that was asked
2. The full text of the ADRs that SHOULD have been used to answer (ground truth)
3. The system's actual answer

Score the answer on two dimensions:

**Accuracy (0-5)**: Does the answer correctly represent what the ADRs say?
- 5: Perfectly accurate, no errors or hallucinations
- 4: Mostly accurate, minor imprecisions
- 3: Generally correct but some notable errors or misleading statements
- 2: Several factual errors or hallucinated claims
- 1: Mostly incorrect
- 0: Completely wrong or fabricated

**Completeness (0-5)**: Does the answer address all aspects of the question?
- 5: Thoroughly addresses every aspect of the question
- 4: Addresses most aspects, minor gaps
- 3: Covers the main point but misses significant aspects
- 2: Only partially addresses the question
- 1: Barely touches on the question
- 0: Does not address the question at all

Be strict. Ground your evaluation in the actual ADR content provided.`

// Judge scores answers for accuracy and completeness.
type Judge interface {
	Score(ctx context.Context, question, expectedADRContent, answer string) (JudgeResult, error)
}

// JudgeResult holds the judge's scores and reasoning.
type JudgeResult struct {
	Accuracy           float64
	Completeness       float64
	AccuracyReason     string
	CompletenessReason string
}

type judgeResponse struct {
	Accuracy           int    `json:"accuracy"`
	Completeness       int    `json:"completeness"`
	AccuracyReason     string `json:"accuracy_reason"`
	CompletenessReason string `json:"completeness_reason"`
}

// AnthropicJudge uses Claude to evaluate answer quality.
type AnthropicJudge struct {
	client *anthropic.Client
	model  anthropic.Model
}

// NewAnthropicJudge creates a judge using the Anthropic API.
func NewAnthropicJudge(apiKey, model string) *AnthropicJudge {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicJudge{
		client: &client,
		model:  anthropic.Model(model),
	}
}

// Score evaluates the answer against the expected ADR content.
// Retries once on malformed response before falling back to zero scores.
func (j *AnthropicJudge) Score(ctx context.Context, question, expectedADRContent, answer string) (JudgeResult, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		result, err := j.score(ctx, question, expectedADRContent, answer)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return JudgeResult{
		AccuracyReason:     fmt.Sprintf("Judge scoring failed: %v", lastErr),
		CompletenessReason: fmt.Sprintf("Judge scoring failed: %v", lastErr),
	}, nil
}

func (j *AnthropicJudge) score(ctx context.Context, question, expectedADRContent, answer string) (JudgeResult, error) {
	var userMsg strings.Builder
	userMsg.WriteString("## Question\n\n")
	userMsg.WriteString(question)
	userMsg.WriteString("\n\n## Expected ADR Content (Ground Truth)\n\n")
	userMsg.WriteString(expectedADRContent)
	userMsg.WriteString("\n\n## System's Answer\n\n")
	userMsg.WriteString(answer)

	resp, err := j.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     j.model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: judgeSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg.String())),
		},
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{
				Schema: judgeSchema(),
			},
		},
	})
	if err != nil {
		return JudgeResult{}, fmt.Errorf("judge API call: %w", err)
	}

	return parseJudgeResponse(resp)
}

func parseJudgeResponse(resp *anthropic.Message) (JudgeResult, error) {
	if len(resp.Content) == 0 {
		return JudgeResult{}, fmt.Errorf("judge scoring: empty response from Anthropic API")
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			var jr judgeResponse
			if err := json.Unmarshal([]byte(block.Text), &jr); err != nil {
				return JudgeResult{}, fmt.Errorf("parsing judge response: %w", err)
			}
			return JudgeResult{
				Accuracy:           float64(jr.Accuracy) / 5.0,
				Completeness:       float64(jr.Completeness) / 5.0,
				AccuracyReason:     jr.AccuracyReason,
				CompletenessReason: jr.CompletenessReason,
			}, nil
		}
	}

	return JudgeResult{}, fmt.Errorf("judge scoring: no text content in response (unexpected content types)")
}

func judgeSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"accuracy": map[string]any{
				"type":        "integer",
				"description": "Accuracy score 0-5",
			},
			"completeness": map[string]any{
				"type":        "integer",
				"description": "Completeness score 0-5",
			},
			"accuracy_reason": map[string]any{
				"type":        "string",
				"description": "Brief justification for the accuracy score",
			},
			"completeness_reason": map[string]any{
				"type":        "string",
				"description": "Brief justification for the completeness score",
			},
		},
		"required":             []string{"accuracy", "completeness", "accuracy_reason", "completeness_reason"},
		"additionalProperties": false,
	}
}
