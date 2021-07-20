package score

// GradingScheme for an assignment
type GradingScheme struct {
	ID uint64 `json:"id"`

	Name        string
	GradePoints []uint32
	GradeNames  []string
}

// Grade computes the grade for the given points.
// The points must be in the range [0,100].
func (g *GradingScheme) Grade(points uint32) string {
	if points > 100 {
		return "bad points"
	}
	for i, p := range g.GradePoints {
		if points >= p {
			return g.GradeNames[i]
		}
	}
	return g.GradeNames[len(g.GradeNames)-1]
}
