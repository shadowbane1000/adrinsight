package llm

import (
	"context"
	"os"
	"testing"
)

func TestNewAnthropicLLM(t *testing.T) {
	llm := NewAnthropicLLM("test-key", "claude-sonnet-4-5")
	if llm == nil {
		t.Fatal("expected non-nil LLM")
	}
}

func TestAnthropicLLMInterfaceCompliance(t *testing.T) {
	var _ LLM = (*AnthropicLLM)(nil)
}

func TestAnthropicSynthesizeIntegration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	llm := NewAnthropicLLM(apiKey, "claude-sonnet-4-5")
	ctx := context.Background()

	adrs := []ADRContext{
		{
			Number:  1,
			Title:   "Use Go for the Project",
			Content: "# ADR-001: Use Go\n\n## Context\nWe need a language for a cloud-native tool.\n\n## Decision\nUse Go.\n\n## Rationale\nGo produces static binaries and has great concurrency support.",
		},
	}

	resp, err := llm.Synthesize(ctx, "Why did we choose Go?", adrs)
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}

	if resp.Answer == "" {
		t.Error("expected non-empty answer")
	}
	if len(resp.Citations) == 0 {
		t.Error("expected at least one citation")
	}
	for _, c := range resp.Citations {
		if c.ADRNumber != 1 {
			t.Errorf("expected citation to ADR-001, got ADR-%03d", c.ADRNumber)
		}
	}
}
