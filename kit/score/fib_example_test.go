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
	score.Add(TestFibonacci, len(fibonacciTests)*2, 20)
	score.Add(TestFibonacci2, len(fibonacciTests), 20)
	for _, ft := range fibonacciTests {
		score.AddSubtest(score.TestName(TestFibonacciWithRun)+"/"+subTestName(ft.in), 1, 1)
	}
	os.Exit(m.Run())
}

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

func TestFibonacci(t *testing.T) {
	sc := score.GetMax()
	defer sc.Print(t)

	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out != ft.want {
			sc.Dec()
		}
	}
	if sc.Score != 24 {
		t.Errorf("expected 24 tests to pass, but got %v", sc.Score)
	}
	if sc.TestName != "TestFibonacci" {
		t.Errorf("expected TestName=TestFibonacci, but got %v", sc.TestName)
	}
	t.Log(sc.Secret)
}

func TestFibonacci2(t *testing.T) {
	sc := score.GetMax()
	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out != ft.want {
			sc.Dec()
		}
	}
	if sc.Score != 10 {
		t.Errorf("expected 10 tests to pass, but got %v", sc.Score)
	}
	if sc.TestName != "TestFibonacci2" {
		t.Errorf("expected TestName=TestFibonacci2, but got %v", sc.TestName)
	}
}

func TestFibonacciWithRun(t *testing.T) {
	for _, ft := range fibonacciTests {
		t.Run(subTestName(ft.in), func(t *testing.T) {
			sc := score.GMax(t.Name())
			out := fibonacci(ft.in)
			if out != ft.want {
				sc.Dec()
			}
			fmt.Println(sc)
		})
	}
}

// TODO(meling) find good design for interacting with subtests

func subTestName(i uint) string {
	return fmt.Sprintf("Fib/%d", i)
}

// func TestFibonacciWithPanic(t *testing.T) {
// 	// TODO(meling) make this continue
// 	panic("hei")
// }

func TestFibonacciWithAfterPanic(t *testing.T) {
	// TODO(meling) make this continue
	t.Log("hallo")
}
