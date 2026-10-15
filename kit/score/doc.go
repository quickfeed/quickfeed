// Package score provides support for scoring tests.
//
// It is intended to be used in concert with the QuickFeed web service,
// which automates execution and scoring of student implemented assignments
// aimed to pass a given set of tests.
//
// QuickFeed computes the score according to the formulas below, providing
// a percentage score for a test or a group of tests. The Weight parameter can
// be used to give more or less weight to some Score objects (representing
// different test sets). For example, if TestA has Weight 2 and TestB has Weight 1,
// then passing TestA gives twice the score of passing TestB.
//
// QuickFeed computes the final score as follows:
//
//	TotalWeight     = sum(Weight)
//	TaskScore[i]    = Score[i] / MaxScore[i], gives {0 < TaskScore < 1}
//	TaskWeight[i]   = Weight[i] / TotalWeight
//	TotalScore      = sum(TaskScore[i]*TaskWeight[i]), gives {0 < TotalScore < 1}
//
// QuickFeed expects that tests are initialized in the init() method before test execution.
// This is done via the score.Add() method or the score.AddSub() method as shown below.
// Add() is used for regular tests, and AddSub() is used for subtests with individual scores.
//
//	func init() {
//	    max, weight := len(fibonacciTests), 20
//	    score.Add(TestFibonacciMax, max, weight)
//	    score.Add(TestFibonacciMin, max, weight)
//	    for _, ft := range fibonacciTests {
//	        score.AddSub(TestFibonacciSubTest, subTestName("Max", ft.in), 1, 1)
//	    }
//	    for _, ft := range fibonacciTests {
//	        score.AddSub(TestFibonacciSubTest, subTestName("Min", ft.in), 1, 1)
//	    }
//	}
//
// In addition, TestMain() should call score.PrintTestInfo() before running the tests
// to ensure that all tests are registered and will be picked up by QuickFeed.
//
//	func TestMain(m *testing.M) {
//	    score.PrintTestInfo()
//	    exitCode := m.Run()
//	    if err := score.Validate(); err != nil {
//	        fmt.Println(err)
//	    }
//	    os.Exit(exitCode)
//	}
//
// A graded test obtains a T from the registry. Its score is printed by Cleanup
// after its subtests finish. Defer Recover to isolate a panic in the top-level test.
//
//	func TestFibonacciMax(t *testing.T) {
//	    sc := scores.Max(t)
//	    defer sc.Recover()
//	    for _, ft := range fibonacciTests {
//	        out := fibonacci(ft.in)
//	        if out != ft.want {
//	            sc.Dec()
//	            sc.Errorf("fibonacci(%d) = %d, want %d", ft.in, out, ft.want)
//	        }
//	    }
//	}
//
// A helper can take *T and use its testing and scoring methods directly.
// NewUnscored runs the same helper without printing a QuickFeed score.
//
//	func checkFibonacci(t *score.T) {
//	    if got := fibonacci(5); got != 5 {
//	        t.Errorf("fibonacci(5) = %d, want 5", got)
//	    } else {
//	        t.Inc()
//	    }
//	}
//	func TestFibonacciPlain(t *testing.T) {
//	    checkFibonacci(score.NewUnscored(t))
//	}
//
// A test that reports its failures through the score object rather than through
// t alone tells the student why it failed, in the teacher's own words, above the
// rest of that test's output:
//
//	func TestFibonacci(t *testing.T) {
//	    sc := scores.Max(t)
//	    defer sc.Recover()
//	    for _, ft := range fibonacciTests {
//	        if out := fibonacci(ft.in); out != ft.want {
//	            sc.Dec()
//	            sc.Errorf("fibonacci(%d) = %d, want: %d", ft.in, out, ft.want)
//	        }
//	    }
//	}
//
// The score is reported as a test attribute, which names the test it came
// from even when the output of parallel tests is interleaved. The testing
// package emits attributes only in verbose mode, so a run script should keep
// -v (or -json) on its go test command; without it the score is printed
// instead, and the run is still scored.
//
// Please see package score/testdata/sequence for other usage examples.
package score
