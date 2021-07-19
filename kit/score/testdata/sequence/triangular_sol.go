package sequence

// triangular(n) returns the n-th Triangular number, and is defined by the
// recurrence relation F_n = n + F_n-1 where F_0=0 and F_1=1
//
// Visualization of numbers:
// n = 1:    n = 2:     n = 3:      n = 4:    etc...
//   o         o          o           o
//            o o        o o         o o
//                      o o o       o o o
//                                 o o o o
func triangularRecurrence(n uint) uint {
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	return n + triangular(n-1)
}

// Direct mathematical solution:

func triangularFormula(n uint) uint {
	return (n * (n + 1)) / 2
}

func triangular(n uint) uint {
	sum := uint(0)
	for i := uint(1); i <= n; i++ {
		sum += i
	}
	return sum
}
