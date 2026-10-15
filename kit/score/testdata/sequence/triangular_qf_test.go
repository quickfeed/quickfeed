package sequence

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/score"
)

func init() {
	// Reduce max score by 1 since the first test-case ({0, 0}) always passes, which gave free points.
	max, weight := len(triangularTests)-1, 5
	scores.Add(TestTriangular, max, weight)
	scores.Add(TestTriangularRecurrence, max, weight)
	scores.Add(TestTriangularFormula, max, weight)
	// Alternative strategy using subtests that share one score.
	scores.Add(TestTriangularSubTest, max*len(funcs), weight)
}

var funcs = map[string]func(uint) uint{
	"triangular":           triangular,
	"triangularRecurrence": triangularRecurrence,
	"triangularFormula":    triangularFormula,
}

var triangularTests = []struct {
	in, want uint
}{
	{0, 0},
	{1, 1},
	{2, 3},
	{3, 6},
	{4, 10},
	{5, 15},
	{6, 21},
	{7, 28},
	{8, 36},
	{9, 45},
	{10, 55},
	{20, 210},
	{42, 903},
	{56, 1596},
	{62, 1953},
	{75, 2850},
	{90, 4095},
}

func TestTriangular(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()

	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangular(test.in)); diff != "" {
			sc.Dec()
			sc.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestTriangularRecurrence(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()

	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangularRecurrence(test.in)); diff != "" {
			sc.Dec()
			sc.Errorf("triangularRecurrence(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestTriangularFormula(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()

	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangularFormula(test.in)); diff != "" {
			sc.Dec()
			sc.Errorf("triangularFormula(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestTriangularSubTest(t *testing.T) {
	s := scores.Max(t)
	defer s.Recover()
	for name, fn := range funcs {
		s.Run(name, func(sc *score.T) {
			for _, test := range triangularTests {
				if diff := cmp.Diff(test.want, fn(test.in)); diff != "" {
					sc.Dec()
					sc.Errorf("%s(%d): (-want +got):\n%s", name, test.in, diff)
				}
			}
		})
	}
}
