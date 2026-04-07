package eval

// ComputeRetrieval calculates precision, recall, and F1 for citation sets.
// Both sets are deduplicated before comparison (treats citations as sets, not lists).
func ComputeRetrieval(returned, expected []int) (precision, recall, f1 float64) {
	returnedSet := make(map[int]bool, len(returned))
	for _, n := range returned {
		returnedSet[n] = true
	}

	expectedSet := make(map[int]bool, len(expected))
	for _, n := range expected {
		expectedSet[n] = true
	}

	if len(returnedSet) == 0 && len(expectedSet) == 0 {
		return 1.0, 1.0, 1.0
	}

	var overlap int
	for n := range returnedSet {
		if expectedSet[n] {
			overlap++
		}
	}

	if len(returnedSet) > 0 {
		precision = float64(overlap) / float64(len(returnedSet))
	}
	if len(expectedSet) > 0 {
		recall = float64(overlap) / float64(len(expectedSet))
	}
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}
	return precision, recall, f1
}
