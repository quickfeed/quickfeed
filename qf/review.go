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

// PercentScore returns the review's score as a percentage: the share of the
// criteria's points that passed, or the share of passed criteria if no
// criterion has points.
func (r *Review) PercentScore() uint32 {
	var points, passedPoints, criteria, passed uint64
	for _, bm := range r.GetGradingBenchmarks() {
		for _, c := range bm.GetCriteria() {
			criteria++
			points += c.GetPoints()
			if c.GetGrade() == GradingCriterion_PASSED {
				passed++
				passedPoints += c.GetPoints()
			}
		}
	}
	switch {
	case points > 0:
		return uint32(100 * passedPoints / points)
	case criteria > 0:
		return uint32(100 * passed / criteria)
	default:
		return 0
	}
}

// Complete returns true if every criterion of the review is graded.
func (r *Review) Complete() bool {
	for _, bm := range r.GetGradingBenchmarks() {
		for _, c := range bm.GetCriteria() {
			if c.GetGrade() == GradingCriterion_NONE {
				return false
			}
		}
	}
	return true
}
