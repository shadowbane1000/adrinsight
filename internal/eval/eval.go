package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shadowbane1000/adrinsight/internal/rag"
)

// LoadTestCases reads and validates test cases from a JSON file.
func LoadTestCases(path string) ([]TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading test cases: %w", err)
	}

	var cases []TestCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, fmt.Errorf("parsing test cases: %w", err)
	}

	seen := make(map[string]bool)
	for i, tc := range cases {
		if tc.ID == "" {
			return nil, fmt.Errorf("test case %d: missing id", i)
		}
		if seen[tc.ID] {
			return nil, fmt.Errorf("test case %d: duplicate id %q", i, tc.ID)
		}
		seen[tc.ID] = true
		if tc.Question == "" {
			return nil, fmt.Errorf("test case %q: missing question", tc.ID)
		}
		// expected_adrs may be empty (meaning no ADRs should be cited)
	}

	return cases, nil
}

// LoadBaseline reads a baseline file. Returns nil, nil if the file doesn't exist.
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading baseline: %w", err)
	}

	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		// Corrupted baseline — treat as missing.
		return nil, nil
	}
	return &b, nil
}

// SaveBaseline writes the current results as the new baseline.
func SaveBaseline(path string, report *RunReport, deltaThreshold float64) error {
	b := Baseline{
		CreatedAt:      time.Now().UTC(),
		DeltaThreshold: deltaThreshold,
		Results:        report.Results,
	}

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling baseline: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ComputeAggregates calculates mean scores across all results.
func ComputeAggregates(results []EvalResult) AggregateScores {
	if len(results) == 0 {
		return AggregateScores{}
	}

	var agg AggregateScores
	n := float64(len(results))
	for _, r := range results {
		agg.AvgPrecision += r.Precision
		agg.AvgRecall += r.Recall
		agg.AvgF1 += r.F1
		agg.AvgAccuracy += r.Accuracy
		agg.AvgCompleteness += r.Completeness
	}
	agg.AvgPrecision /= n
	agg.AvgRecall /= n
	agg.AvgF1 /= n
	agg.AvgAccuracy /= n
	agg.AvgCompleteness /= n
	return agg
}

// RunEval executes all test cases against the pipeline and scores them.
func RunEval(ctx context.Context, cases []TestCase, pipeline *rag.Pipeline, judge Judge, adrDir string) (*RunReport, error) {
	report := &RunReport{
		Timestamp: time.Now().UTC(),
	}

	for i, tc := range cases {
		log.Printf("Evaluating [%d/%d]: %s", i+1, len(cases), tc.ID)

		// Per-question timeout: 2 minutes for embedding + retrieval + synthesis + judge.
		qCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)

		// Query the pipeline.
		resp, err := pipeline.Query(qCtx, tc.Question)
		if err != nil {
			cancel()
			log.Printf("Warning: query failed for %q: %v (skipping)", tc.ID, err)
			continue
		}

		// Extract cited ADR numbers from LLM response.
		var citedADRs []int
		for _, c := range resp.Citations {
			citedADRs = append(citedADRs, c.ADRNumber)
		}

		// Compute retrieval scores from deterministic search results.
		precision, recall, f1 := ComputeRetrieval(resp.RetrievedADRs, tc.ExpectedADRs)

		// Score with LLM judge (skip if judge is nil).
		var jr JudgeResult
		if judge != nil {
			adrContent := loadExpectedADRContent(tc.ExpectedADRs, adrDir)
			jr, err = judge.Score(qCtx, tc.Question, adrContent, resp.Answer)
			if err != nil {
				log.Printf("Warning: judge scoring failed for %q: %v (using 0 scores)", tc.ID, err)
				jr = JudgeResult{
					AccuracyReason:     "Judge scoring failed",
					CompletenessReason: "Judge scoring failed",
				}
			}
		} else {
			jr = JudgeResult{
				AccuracyReason:     "skipped",
				CompletenessReason: "skipped",
			}
		}

		result := EvalResult{
			ID:                 tc.ID,
			Question:           tc.Question,
			Answer:             resp.Answer,
			RetrievedADRs:      resp.RetrievedADRs,
			CitedADRs:          citedADRs,
			ReturnedADRs:       citedADRs, // backward compat
			Precision:          precision,
			Recall:             recall,
			F1:                 f1,
			Accuracy:           jr.Accuracy,
			Completeness:       jr.Completeness,
			AccuracyReason:     jr.AccuracyReason,
			CompletenessReason: jr.CompletenessReason,
		}
		report.Results = append(report.Results, result)
		cancel()
	}

	report.Aggregate = ComputeAggregates(report.Results)
	return report, nil
}

// loadExpectedADRContent reads ADR files for the expected ADR numbers.
func loadExpectedADRContent(adrNumbers []int, adrDir string) string {
	entries, err := os.ReadDir(adrDir)
	if err != nil {
		return ""
	}

	// Build a map of ADR number → file path.
	adrFiles := make(map[int]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		var num int
		if _, err := fmt.Sscanf(name, "ADR-%d", &num); err == nil {
			adrFiles[num] = filepath.Join(adrDir, name)
		}
	}

	var content string
	for _, num := range adrNumbers {
		path, ok := adrFiles[num]
		if !ok {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content += fmt.Sprintf("### ADR-%03d\n\n%s\n\n---\n\n", num, string(data))
	}
	return content
}

// DetectRegressions compares current results against a baseline.
func DetectRegressions(report *RunReport, baseline *Baseline, delta float64) {
	if baseline == nil {
		return
	}

	baseMap := make(map[string]EvalResult)
	for _, r := range baseline.Results {
		baseMap[r.ID] = r
	}

	for _, r := range report.Results {
		base, ok := baseMap[r.ID]
		if !ok {
			report.NewCases = append(report.NewCases, r.ID)
			continue
		}

		checkDim := func(dim string, baseVal, curVal float64) {
			if baseVal-curVal > delta {
				report.Regressions = append(report.Regressions, Regression{
					ID:            r.ID,
					Dimension:     dim,
					BaselineScore: baseVal,
					CurrentScore:  curVal,
					Delta:         baseVal - curVal,
				})
			}
		}

		checkDim("precision", base.Precision, r.Precision)
		checkDim("recall", base.Recall, r.Recall)
		checkDim("f1", base.F1, r.F1)
		checkDim("accuracy", base.Accuracy, r.Accuracy)
		checkDim("completeness", base.Completeness, r.Completeness)
	}
}
