package qf

// ComputeScore computes the total score for the review and assigns it to r.
// If the grading criteria have predefined points, the score is the sum of these points.
// Otherwise, each criterion is given equal weight, such that the max score is 100.
func (r *Review) ComputeScore() {
	scorePoints := 0
	totalCriteria := 0
	passedCriteria := 0
	for _, bm := range r.GetGradingBenchmarks() {
		for _, c := range bm.GetCriteria() {
			totalCriteria++
			if c.GetGrade() == GradingCriterion_PASSED {
				passedCriteria++
				scorePoints += int(c.GetPoints())
			}
		}
	}
	if totalCriteria == 0 {
		return
	}
	if scorePoints == 0 {
		r.Score = uint32(100 * passedCriteria / totalCriteria)
	} else {
		r.Score = uint32(scorePoints)
	}
}
