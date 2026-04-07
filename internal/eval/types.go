package eval

import "time"

// TestCase defines a single evaluation question with ground truth.
type TestCase struct {
	ID           string   `json:"id"`
	Question     string   `json:"question"`
	ExpectedADRs []int    `json:"expected_adrs"`
	KeyFacts     []string `json:"key_facts"`
}

// EvalResult holds the scored output for a single test case.
type EvalResult struct {
	ID                   string  `json:"id"`
	Question             string  `json:"question"`
	Answer               string  `json:"answer"`
	RetrievedADRs        []int   `json:"retrieved_adrs"`
	CitedADRs            []int   `json:"cited_adrs"`
	ReturnedADRs         []int   `json:"returned_adrs"` // deprecated: use CitedADRs
	Precision            float64 `json:"precision"`
	Recall               float64 `json:"recall"`
	F1                   float64 `json:"f1"`
	Accuracy             float64 `json:"accuracy"`
	Completeness         float64 `json:"completeness"`
	AccuracyReason       string  `json:"accuracy_reason"`
	CompletenessReason   string  `json:"completeness_reason"`
}

// Baseline is a snapshot of results from a known-good state.
type Baseline struct {
	CreatedAt      time.Time    `json:"created_at"`
	DeltaThreshold float64      `json:"delta_threshold"`
	Results        []EvalResult `json:"results"`
}

// RunReport is the aggregate output from a single evaluation run.
type RunReport struct {
	Timestamp   time.Time       `json:"timestamp"`
	Results     []EvalResult    `json:"results"`
	Aggregate   AggregateScores `json:"aggregate"`
	Regressions []Regression    `json:"regressions"`
	NewCases    []string        `json:"new_cases"`
}

// AggregateScores holds mean scores across all test cases.
type AggregateScores struct {
	AvgPrecision    float64 `json:"avg_precision"`
	AvgRecall       float64 `json:"avg_recall"`
	AvgF1           float64 `json:"avg_f1"`
	AvgAccuracy     float64 `json:"avg_accuracy"`
	AvgCompleteness float64 `json:"avg_completeness"`
}

// Regression records a score drop for a specific test case and dimension.
type Regression struct {
	ID            string  `json:"id"`
	Dimension     string  `json:"dimension"`
	BaselineScore float64 `json:"baseline_score"`
	CurrentScore  float64 `json:"current_score"`
	Delta         float64 `json:"delta"`
}
