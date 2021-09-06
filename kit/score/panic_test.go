package score_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// To run all these tests, to show stack trace and panic output, use:
//   QUICKFEED_PANIC_TEST=1 go test -v -run TestPanic
//

const (
	panicTestEnvName = "QUICKFEED_PANIC_TEST"
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
	scores.Add(TestPanicTriangularBefore, len(triangularTests), 5)
	scores.Add(TestPanicTriangularPanic, len(triangularTests), 5)
	scores.Add(TestPanicTriangularAfter, len(triangularTests), 5)
	scores.Add(TestPanicHandler, len(triangularTests), 5)
	scores.Add(TestPanicTriangularPanicWithMsg, len(triangularTests), 5)
	scores.Add(TestPanicHandlerWithMsg, len(triangularTests), 5)
}

// TestPanicTriangularBefore is meant to run before TestPanicTriangularPanic.
// This test should pass and give max score.
func TestPanicTriangularBefore(t *testing.T) {
	sc := scores.Max()
	defer sc.Print(t)
	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangular(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

// TestPanicTriangularPanic is meant to run after TestPanicTriangularBefore.
// This test is meant to panic, and will give zero score
// since test scores are printed by TestMain() before test execution.
func TestPanicTriangularPanic(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	panicTest := os.Getenv(panicTestEnvName)
	if panicTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", panicTestEnvName, t.Name())
	}

	sc := scores.Max()
	defer sc.Print(t)
	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangularPanic(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

// TestPanicTriangularAfter is meant to run after TestPanicTriangularPanic.
// This test will not run since TestPanicTriangularPanic panicks, and will give zero score
// since test scores are printed by TestMain() before test execution.
func TestPanicTriangularAfter(t *testing.T) {
	sc := scores.Max()
	defer sc.Print(t)
	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangular(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestPanicHandler(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	panicTest := os.Getenv(panicTestEnvName)
	if panicTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", panicTestEnvName, t.Name())
	}

	sc := scores.Max()
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

func TestPanicTriangularPanicWithMsg(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	panicTest := os.Getenv(panicTestEnvName)
	if panicTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", panicTestEnvName, t.Name())
	}

	sc := scores.Max()
	defer sc.Print(t, "This could be a helpful message to explain why something might panic.")
	for _, test := range triangularTests {
		if diff := cmp.Diff(test.want, triangularPanic(test.in)); diff != "" {
			sc.Dec()
			t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
		}
	}
}

func TestPanicHandlerWithMsg(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	panicTest := os.Getenv(panicTestEnvName)
	if panicTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", panicTestEnvName, t.Name())
	}

	sc := scores.Max()
	defer sc.Print(t)
	for _, test := range triangularTests {
		t.Run(fmt.Sprintf("triangular(%d)=%d", test.in, test.want), func(t *testing.T) {
			defer sc.PanicHandler(t, "This could be a helpful message to explain why something might panic.")
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
