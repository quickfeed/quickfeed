package sequence

import (
	"testing"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
)

func init() {
	// Reduce max score by 1 since the first test-case ({0, 0}) always passes, which gave free points.
	score.Add(TestTriangularAG, len(triangularTestsAG)-1, 5)
	score.Add(TestTriangularRecurrenceAG, len(triangularTestsAG)-1, 5)
	score.Add(TestTriangularFormulaAG, len(triangularTestsAG)-1, 5)
	// Here is alternative strategy using subtests
	for name := range funcs {
		score.AddSub(TestTriangularSubTestAG, name, len(triangularTestsAG)-1, 5)
	}
}

var funcs = map[string]func(uint) uint{
	"triangular":           triangular,
	"triangularRecurrence": triangularRecurrence,
	"triangularFormula":    triangularFormula,
}

var triangularTestsAG = []struct {
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

func TestTriangularAG(t *testing.T) {
	sc := score.Max()
	defer sc.Print(t)

	for _, test := range triangularTestsAG {
		if diff := cmp.Diff(test.want, triangular(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestTriangularRecurrenceAG(t *testing.T) {
	sc := score.Max()
	defer sc.Print(t)

	for _, test := range triangularTestsAG {
		if diff := cmp.Diff(test.want, triangularRecurrence(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangularRecurrence(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestTriangularFormulaAG(t *testing.T) {
	sc := score.Max()
	defer sc.Print(t)

	for _, test := range triangularTestsAG {
		if diff := cmp.Diff(test.want, triangularFormula(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangularFormula(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestTriangularSubTestAG(t *testing.T) {
	for name, fn := range funcs {
		t.Run(name, func(t *testing.T) {
			sc := score.MaxByName(t.Name())
			defer sc.Print(t)
			for _, test := range triangularTestsAG {
				if diff := cmp.Diff(test.want, fn(test.in)); diff != "" {
					sc.Dec()
					t.Errorf("%s(%d): (-want +got):\n%s", name, test.in, diff)
				}
			}
		})
	}
}
