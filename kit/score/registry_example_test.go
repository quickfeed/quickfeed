package score_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

func fibonacci(n uint) uint {
	if n <= 1 || n == 5 {
		return n
	}
	if n < 5 {
		return n - 1
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

var scoreRegistry = score.NewRegistry()

func init() {
	scoreRegistry.Add(TestFibonacciMax, len(fibonacciTests), 20)
	scoreRegistry.Add(TestFibonacciMin, len(fibonacciTests), 20)
	scoreRegistry.Add(TestFibonacciWrapper, 1, 1)
	for _, ft := range fibonacciTests {
		scoreRegistry.AddSub(TestFibonacciSubTest, subTestName("Max", ft.in), 1, 1)
	}
	for _, ft := range fibonacciTests {
		scoreRegistry.AddSub(TestFibonacciSubTest, subTestName("Min", ft.in), 1, 1)
	}
}

const (
	numCorrect = 10
)

var fibonacciTests = []struct {
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
	{10, 155},   // correct 55
	{12, 154},   // correct 89
	{16, 1987},  // correct 987
	{20, 26765}, // correct 6765
}

func TestFibonacciMax(t *testing.T) {
	sc := scoreRegistry.Max(t)
	defer sc.Recover()
	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out != ft.want {
			sc.Dec()
		}
	}
	if sc.Score().GetScore() != numCorrect {
		t.Errorf("Score=%d, expected %d", sc.Score().GetScore(), numCorrect)
	}
	if sc.Score().GetTestName() != t.Name() {
		t.Errorf("TestName=%s, expected %s", sc.Score().GetTestName(), t.Name())
	}
}

var (
	once sync.Once
	ran  bool
)

func TestFibonacciMin(t *testing.T) {
	if ran {
		// This test cannot be run more than once
		t.SkipNow()
	}
	sc := scoreRegistry.Min(t)
	defer sc.Recover()
	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out == ft.want {
			sc.Inc()
		}
	}
	if sc.Score().GetScore() != numCorrect {
		t.Errorf("Score=%d, expected %d", sc.Score().GetScore(), numCorrect)
	}
	if sc.Score().GetTestName() != t.Name() {
		t.Errorf("TestName=%s, expected %s", sc.Score().GetTestName(), t.Name())
	}
	once.Do(func() { ran = true })
}

func TestFibonacciSubTest(t *testing.T) {
	for _, ft := range fibonacciTests {
		t.Run(subTestName("Max", ft.in), func(t *testing.T) {
			sc := scoreRegistry.MaxByName(t)
			defer sc.Recover()
			out := fibonacci(ft.in)
			if out != ft.want {
				sc.Dec()
			}
			expectedScore := int32(0)
			if ft.in < numCorrect {
				expectedScore = 1
			}
			if sc.Score().GetScore() != expectedScore {
				t.Errorf("Score=%d, expected %d", sc.Score().GetScore(), expectedScore)
			}
			if sc.Score().GetTestName() != t.Name() {
				t.Errorf("TestName=%s, expected %s", sc.Score().GetTestName(), t.Name())
			}
		})
		t.Run(subTestName("Min", ft.in), func(t *testing.T) {
			sc := scoreRegistry.MinByName(t)
			defer sc.Recover()
			out := fibonacci(ft.in)
			if out == ft.want {
				sc.Inc()
			}
			expectedScore := int32(0)
			if ft.in < numCorrect {
				expectedScore = 1
			}
			if sc.Score().GetScore() != expectedScore {
				t.Errorf("Score=%d, expected %d", sc.Score().GetScore(), expectedScore)
			}
			if sc.Score().GetTestName() != t.Name() {
				t.Errorf("TestName=%s, expected %s", sc.Score().GetTestName(), t.Name())
			}
		})
	}
}

func checkFibonacci(t *score.T) {
	t.Helper()
	if got := fibonacci(5); got != 5 {
		t.Errorf("fibonacci(5) = %d, want 5", got)
	} else {
		t.Inc()
	}
}

func TestFibonacciWrapper(t *testing.T) {
	sc := scoreRegistry.Min(t)
	defer sc.Recover()
	checkFibonacci(sc)
}

func TestFibonacciPlain(t *testing.T) {
	checkFibonacci(score.NewUnscored(t))
}

func subTestName(prefix string, i uint) string {
	return fmt.Sprintf("%s/%d", prefix, i)
}

func TestDuplicateAdd(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			out := strings.TrimSpace(fmt.Sprintln(r))
			if !strings.HasSuffix(out, "duplicate score test: TestDuplicateAdd") {
				t.Errorf("Add() unexpected panic: %v", r)
			}
		} else {
			t.Errorf("Add() did not panic")
		}
	}()
	scores.Add(TestDuplicateAdd, 1, 1)
	scores.Add(TestDuplicateAdd, 1, 1) // should panic with duplicate score test
}
