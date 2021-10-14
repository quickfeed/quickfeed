package score_test

import (
	"fmt"
	"log"
	"strings"
	"sync"
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

var scoreRegistry = score.NewRegistry()

func init() {
	scoreRegistry.Add(TestFibonacciMax, len(fibonacciTests), 20)
	scoreRegistry.Add(TestFibonacciMin, len(fibonacciTests), 20)
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
	sc := scoreRegistry.Max()
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

var (
	once sync.Once
	ran  bool
)

func TestFibonacciMin(t *testing.T) {
	if ran {
		// This test cannot be run more than once
		t.SkipNow()
	}
	sc := scoreRegistry.Min()
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
	once.Do(func() { ran = true })
}

func TestFibonacciSubTest(t *testing.T) {
	for _, ft := range fibonacciTests {
		t.Run(subTestName("Max", ft.in), func(t *testing.T) {
			sc := scoreRegistry.MaxByName(t.Name())
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
			sc := scoreRegistry.MinByName(t.Name())
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

func TestStudentAttackCode(t *testing.T) {
	tests := []struct {
		id, name, test string
		fn             func(string) *score.Score
		want           string
	}{
		{id: "1", name: "MaxByName", test: "TestFibonacciMax", fn: scoreRegistry.MaxByName, want: "unauthorized lookup: TestFibonacciMax"},
		{id: "2", name: "MaxByName", test: "TestFibonacciMin", fn: scoreRegistry.MaxByName, want: "unauthorized lookup: TestFibonacciMin"},
		{id: "3", name: "MaxByName", test: "TestFibonacciSubTest", fn: scoreRegistry.MaxByName, want: "unauthorized lookup: TestFibonacciSubTest"},
		{id: "4", name: "MaxByName", test: "TestStudentAttackCode", fn: scoreRegistry.MaxByName, want: "unknown score test: TestStudentAttackCode"},
		{id: "1", name: "MinByName", test: "TestFibonacciMax", fn: scoreRegistry.MinByName, want: "unauthorized lookup: TestFibonacciMax"},
		{id: "2", name: "MinByName", test: "TestFibonacciMin", fn: scoreRegistry.MinByName, want: "unauthorized lookup: TestFibonacciMin"},
		{id: "3", name: "MinByName", test: "TestFibonacciSubTest", fn: scoreRegistry.MinByName, want: "unauthorized lookup: TestFibonacciSubTest"},
		{id: "4", name: "MinByName", test: "TestStudentAttackCode", fn: scoreRegistry.MinByName, want: "unknown score test: TestStudentAttackCode"},
	}
	for _, test := range tests {
		t.Run(test.name+"/"+test.id, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					out := strings.TrimSpace(fmt.Sprintln(r))
					// ignore the file name and line number in the prefix of out
					if !strings.HasSuffix(out, test.want) {
						t.Errorf("%s('%s')='%s', expected '%s'", test.name, test.test, out, test.want)
					}
					if len(test.want) == 0 {
						t.Errorf("%s('%s')='%s', not expected to fail", test.name, test.test, out)
					}
				}
			}()
			sc := test.fn(test.test)
			log.Fatalf("Should never be reached: %v", sc)
		})
	}
}
