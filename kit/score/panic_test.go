package score_test

import (
	"fmt"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
)

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
}

func init() {
	score.Add(TestPanicTriangularBefore, len(triangularTests), 5)
	score.Add(TestPanicTriangularPanic, len(triangularTests), 5)
	score.Add(TestPanicTriangularAfter, len(triangularTests), 5)
	score.Add(TestPanicHandler, len(triangularTests), 5)
}

func TestPanicTriangularBefore(t *testing.T) {
	sc := score.Max()
	defer sc.Print(t)

	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangular(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestPanicTriangularPanic(t *testing.T) {
	// This test aims to emulate that student code submitted code may result in a panic.
	// Hence, we do not run this as part of the CI tests.
	// Comment t.Skip to test that TestPanicTriangularPanic fails with a panic stack trace, which is expected.
	t.Skip("Skipping because it is expected to fail (see comment).")
	sc := score.Max()
	defer sc.Print(t)

	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangularPanic(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestPanicTriangularAfter(t *testing.T) {
	sc := score.Max()
	defer sc.Print(t)

	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangular(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestPanicHandler(t *testing.T) {
	// This test aims to emulate that student code submitted code may result in a panic.
	// Hence, we do not run this as part of the CI tests.
	// Comment t.Skip to test that TestPanicHandler fails with a panic stack trace, which is expected.
	t.Skip("Skipping because it is expected to fail (see comment).")
	sc := score.Max()
	defer sc.Print(t)

	for _, test := range triangularTests {
		t.Run(fmt.Sprintf("triangular(%d)=%d", test.in, test.want), func(t *testing.T) {
			defer sc.PanicHandler(t)
			if diff := cmp.Diff(test.want, triangularPanic(test.in)); diff != "" {
				sc.Dec()
				t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
			}
		})
	}
}

func triangular(n uint) uint {
	return (n * (n + 1)) / 2
}

func triangularPanic(n uint) uint {
	if n > 4 {
		panic("n > 4")
	}
	return (n * (n + 1)) / 2
}
