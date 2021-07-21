package ag

// CalculateScore returns total score for the review. If grading criteria give
// a predefined amount of points (not necessary summing up to 100), returns the sum of such points.
// Otherwise, gives each criterion an equal weight and sets max score to 100
func (r *Review) CalculateScore() {
	scorePoints := 0
	totalCriteria := 0
	passedCriteria := 0
	for _, bm := range r.GradingBenchmarks {
		for _, c := range bm.Criteria {
			totalCriteria++
			if c.Grade == GradingCriterion_PASSED {
				passedCriteria++
				scorePoints += int(c.Points)
			}
		}
	}
	if scorePoints == 0 {
		r.Score = uint32(100 * passedCriteria / totalCriteria)
	} else {
		r.Score = uint32(scorePoints)
	}
}
