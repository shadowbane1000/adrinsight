package eval

import (
	"fmt"
	"io"
	"strings"
)

// PrintReport writes a human-readable evaluation report to w.
func PrintReport(w io.Writer, report *RunReport, cases []TestCase, baseline *Baseline) {
	p := func(format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }

	p("Evaluation Report\n")
	p("=================\n\n")

	passed := 0
	regressed := 0

	regMap := make(map[string]bool)
	for _, r := range report.Regressions {
		regMap[r.ID] = true
	}
	newSet := make(map[string]bool)
	for _, id := range report.NewCases {
		newSet[id] = true
	}

	caseMap := make(map[string]TestCase)
	for _, tc := range cases {
		caseMap[tc.ID] = tc
	}

	for _, r := range report.Results {
		p("Question: %s\n", truncateStr(r.Question, 70))
		p("  Citations: %s (expected: %s)\n",
			formatADRList(r.ReturnedADRs), formatADRList(caseMap[r.ID].ExpectedADRs))
		p("  Precision: %.2f  Recall: %.2f  F1: %.2f\n", r.Precision, r.Recall, r.F1)
		p("  Accuracy:  %.2f — %q\n", r.Accuracy, truncateStr(r.AccuracyReason, 60))
		p("  Completeness: %.2f — %q\n", r.Completeness, truncateStr(r.CompletenessReason, 60))

		if newSet[r.ID] {
			p("  Status: NEW (no baseline)\n")
		} else if regMap[r.ID] {
			p("  Status: REGRESSED\n")
			regressed++
		} else {
			p("  Status: PASS\n")
			passed++
		}
		p("\n")
	}

	p("Summary\n")
	p("-------\n")
	p("Questions: %d  Passed: %d  Regressed: %d  New: %d\n",
		len(report.Results), passed, regressed, len(report.NewCases))
	p("Avg Precision: %.2f  Avg Recall: %.2f  Avg F1: %.2f\n",
		report.Aggregate.AvgPrecision, report.Aggregate.AvgRecall, report.Aggregate.AvgF1)
	p("Avg Accuracy: %.2f  Avg Completeness: %.2f\n",
		report.Aggregate.AvgAccuracy, report.Aggregate.AvgCompleteness)
	p("\n")

	if regressed > 0 {
		p("RESULT: FAIL (%d regression(s) detected)\n\n", regressed)
		for _, reg := range report.Regressions {
			p("  %s [%s]: %.2f → %.2f (delta: %.2f)\n",
				reg.ID, reg.Dimension, reg.BaselineScore, reg.CurrentScore, reg.Delta)
		}
	} else if baseline == nil {
		p("RESULT: NO BASELINE (run with --save-baseline to create one)\n")
	} else {
		p("RESULT: PASS\n")
	}
}

func formatADRList(nums []int) string {
	if len(nums) == 0 {
		return "[]"
	}
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func truncateStr(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
