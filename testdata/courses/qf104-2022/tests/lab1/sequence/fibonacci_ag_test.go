package sequence

import (
	"testing"
)

func init() {
	// Reduce max score by 1 since the first test-case ({0, 0}) always passes, which gave free points.
	scores.Add(TestFibonacciAG, len(fibonacciTestsAG)-1, 5)
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
	sc := scores.Max()
	defer sc.Print(t)

	for _, ft := range fibonacciTestsAG {
		got := fibonacci(ft.in)
		if got != ft.want {
			t.Errorf("fibonacci(%d) = %d, want %d", ft.in, got, ft.want)
			sc.Dec()
		}
	}
}
