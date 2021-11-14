package score_test

import (
	"fmt"
	"testing"
)

func fibonacciIter(n uint) uint {
	a, b := uint(0), uint(1)
	// Iterate until desired position in sequence.
	for i := 0; i < int(n); i++ {
		temp := a
		a = b
		b = temp + a
	}
	return a
}

func init() {
	scoreRegistry.Add(BenchmarkFibonacci, len(fibonacciTests), 20)
	// scoreRegistry.Add(TestFibonacciMin, len(fibonacciTests), 20)
	// for _, ft := range fibonacciTests {
	// 	scoreRegistry.AddSub(TestFibonacciSubTest, subTestName("Max", ft.in), 1, 1)
	// }
	// for _, ft := range fibonacciTests {
	// 	scoreRegistry.AddSub(TestFibonacciSubTest, subTestName("Min", ft.in), 1, 1)
	// }
}

// Q for slack or golang nuts:
// Is there a way to access testing.BenchmarkResult after a benchmark has been executed (from within code)
//
// Not clear how to make further progress on this at the moment.
// Seems like the best approach is to shell out a go test -bench command and parse the benchmark results
// 

func BenchmarkFibonacci(b *testing.B) {
	sc := scoreRegistry.Max()
	for _, ft := range fibonacciTests {
		b.Run(fmt.Sprintf("iterative/n=%d", ft.in), func(b *testing.B) {
			var out uint
			for i := 0; i < b.N; i++ {
				out = fibonacciIter(ft.in)
			}
			if out != ft.want {
				sc.Dec()
			}
		})
		b.Run(fmt.Sprintf("recursive/n=%d", ft.in), func(b *testing.B) {
			var out uint
			for i := 0; i < b.N; i++ {
				out = fibonacci(ft.in)
			}
			if out != ft.want {
				sc.Dec()
			}
		})
	}
}
