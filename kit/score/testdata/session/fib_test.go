package session

import (
	"testing"
)

func init() {
	scores.Add(TestFibonacci, len(fibonacciTests), 5)
}

var fibonacciTests = []struct {
	in, want uint
}{
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

func TestFibonacci(t *testing.T) {
	sc := scores.Max()
	defer scores.Print(t, sc)

	for _, ft := range fibonacciTests {
		got := fibonacci(ft.in)
		if got != ft.want {
			t.Errorf("fibonacci(%d) = %d, want %d", ft.in, got, ft.want)
			sc.Dec()
		}
	}
}
