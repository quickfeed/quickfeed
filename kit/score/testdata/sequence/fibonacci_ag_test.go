package sequence

import (
	"os"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

func init() {
	// Reduce max score by 1 since the first test-case ({0, 0}) always passes, which gave free points.
	score.Add(TestFibonacciAG, len(fibonacciTestsAG)-1, 5)
	score.Add(TestFibonacciAttackAG, len(fibonacciTestsAG)-1, 5)
}

var fibonacciTestsAG = []struct {
	in, want uint
}{
	{0, 0},
	{1, 1},
	{2, 1},
	{3, 2},
	{4, 3},
	{5, 5},
	{6, 8},
	{7, 13},
	{8, 21},
	{9, 34},
	{10, 55},
	{12, 144},
	{16, 987},
	{20, 6765},
}

func TestFibonacciAG(t *testing.T) {
	sc := score.Max()
	defer sc.Print(t)

	for _, ft := range fibonacciTestsAG {
		got := fibonacci(ft.in)
		if got != ft.want {
			t.Errorf("fibonacci(%d) = %d, want %d", ft.in, got, ft.want)
			sc.Dec()
		}
	}
}

// To run all these tests, to show stack trace and panic output, use:
//   QUICKFEED_PANIC_TEST=1 go test -v -run TestFibonacciAttackAG
//

const (
	panicTestEnvName = "QUICKFEED_PANIC_TEST"
)

func TestFibonacciAttackAG(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	panicTest := os.Getenv(panicTestEnvName)
	if panicTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", panicTestEnvName, t.Name())
	}

	sc := score.Max()
	defer sc.Print(t)
	for _, ft := range fibonacciTestsAG {
		fibonacciAttack(ft.in)
	}
}
