package score_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
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

func TestMain(m *testing.M) {
	score.Add(TestFibonacciMax, len(fibonacciTests), 20)
	score.Add(TestFibonacciMin, len(fibonacciTests), 20)
	for _, ft := range fibonacciTests {
		score.AddSub(TestFibonacciSubTest, subTestName("Max", ft.in), 1, 1)
	}
	for _, ft := range fibonacciTests {
		score.AddSub(TestFibonacciSubTest, subTestName("Min", ft.in), 1, 1)
	}
	os.Exit(m.Run())
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
	sc := score.Max()
	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out != ft.want {
			sc.Dec()
		}
	}
	if sc.Score != numCorrect {
		t.Errorf("Score=%d, expected %d", sc.Score, numCorrect)
	}
	if sc.TestName != t.Name() {
		t.Errorf("TestName=%s, expected %s", sc.TestName, t.Name())
	}
}

func TestFibonacciMin(t *testing.T) {
	sc := score.Min()
	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out == ft.want {
			sc.Inc()
		}
	}
	if sc.Score != numCorrect {
		t.Errorf("Score=%d, expected %d", sc.Score, numCorrect)
	}
	if sc.TestName != t.Name() {
		t.Errorf("TestName=%s, expected %s", sc.TestName, t.Name())
	}
}

func TestFibonacciSubTest(t *testing.T) {
	for _, ft := range fibonacciTests {
		t.Run(subTestName("Max", ft.in), func(t *testing.T) {
			sc := score.MaxByName(t.Name())
			out := fibonacci(ft.in)
			if out != ft.want {
				sc.Dec()
			}
			expectedScore := int32(0)
			if ft.in < numCorrect {
				expectedScore = 1
			}
			if sc.Score != expectedScore {
				t.Errorf("Score=%d, expected %d", sc.Score, expectedScore)
			}
			if sc.TestName != t.Name() {
				t.Errorf("TestName=%s, expected %s", sc.TestName, t.Name())
			}
		})
		t.Run(subTestName("Min", ft.in), func(t *testing.T) {
			sc := score.MinByName(t.Name())
			out := fibonacci(ft.in)
			if out == ft.want {
				sc.Inc()
			}
			expectedScore := int32(0)
			if ft.in < numCorrect {
				expectedScore = 1
			}
			if sc.Score != expectedScore {
				t.Errorf("Score=%d, expected %d", sc.Score, expectedScore)
			}
			if sc.TestName != t.Name() {
				t.Errorf("TestName=%s, expected %s", sc.TestName, t.Name())
			}
		})
	}
}

func subTestName(prefix string, i uint) string {
	return fmt.Sprintf("%s/%d", prefix, i)
}

// func TestFibonacciWithPanic(t *testing.T) {
// 	// TODO(meling) make this continue
// 	panic("hei")
// }

func TestFibonacciWithAfterPanic(t *testing.T) {
	// TODO(meling) make this continue
	t.Log("hallo")
}
