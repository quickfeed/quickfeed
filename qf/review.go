package qf

// rubricTally counts the grading criteria of a review and their points.
type rubricTally struct {
	criteria     uint64 // all criteria
	graded       uint64 // criteria graded as passed or failed
	passed       uint64 // criteria graded as passed
	points       uint64 // points of all criteria
	passedPoints uint64 // points of the passed criteria
}

func (r *Review) tally() rubricTally {
	var t rubricTally
	for _, bm := range r.GetGradingBenchmarks() {
		for _, c := range bm.GetCriteria() {
			t.criteria++
			t.points += c.GetPoints()
			if c.GetGrade() != GradingCriterion_NONE {
				t.graded++
			}
			if c.GetGrade() == GradingCriterion_PASSED {
				t.passed++
				t.passedPoints += c.GetPoints()
			}
		}
	}
	return t
}

// ComputeScore computes the total score for the review and assigns it to r.
// If the grading criteria have predefined points, the score is the sum of these points.
// Otherwise, each criterion is given equal weight, such that the max score is 100.
func (r *Review) ComputeScore() {
	t := r.tally()
	if t.criteria == 0 {
		return
	}
	if t.passedPoints == 0 {
		r.Score = uint32(100 * t.passed / t.criteria)
	} else {
		r.Score = uint32(t.passedPoints)
	}
}

// PercentScore returns the review's score as a percentage: the share of the
// criteria's points that passed, or the share of passed criteria if no
// criterion has points.
func (r *Review) PercentScore() uint32 {
	t := r.tally()
	switch {
	case t.points > 0:
		return uint32(100 * t.passedPoints / t.points)
	case t.criteria > 0:
		return uint32(100 * t.passed / t.criteria)
	default:
		return 0
	}
}

// Complete returns true if every criterion of the review is graded.
func (r *Review) Complete() bool {
	t := r.tally()
	return t.graded == t.criteria
}
