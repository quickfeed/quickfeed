// Package score provides support for scoring tests.
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
//   TotalWeight     = sum(Weight)
//   TaskScore[i]    = Score[i] / MaxScore[i], gives {0 < TaskScore < 1}
//   TaskWeight[i]   = Weight[i] / TotalWeight
//   TotalScore      = sum(TaskScore[i]*TaskWeight[i]), gives {0 < TotalScore < 1}
//
// QuickFeed expects that tests are initialized in TestMain before test execution.
// This is done via the score.Add() method or the score.AddSub() method as shown below.
//
// func TestMain(m *testing.M) {
//     score.Add(TestFibonacciMax, len(fibonacciTests), 20)
//     score.Add(TestFibonacciMin, len(fibonacciTests), 20)
//     for _, ft := range fibonacciTests {
//         score.AddSub(TestFibonacciSubTest, subTestName("Max", ft.in), 1, 1)
//     }
//     for _, ft := range fibonacciTests {
//         score.AddSub(TestFibonacciSubTest, subTestName("Min", ft.in), 1, 1)
//     }
//     os.Exit(m.Run())
// }
//
// To implement tests with scoring functions is done as follows, where score.Max()
// returns a score object with Score equals to MaxScore, which may be decremented
// for each test failure.
//
// func TestFibonacciMax(t *testing.T) {
//     sc := score.Max()
//     for _, ft := range fibonacciTests {
//         out := fibonacci(ft.in)
//         if out != ft.want {
//             sc.Dec()
//         }
//     }
// }
//
// Similarly, it is also possible to use score.Min() to obtain a score object with
// Score equals to zero, which may be incremented for each test success.
//
// func TestFibonacciMin(t *testing.T) {
//     sc := score.Min()
//     for _, ft := range fibonacciTests {
//         out := fibonacci(ft.in)
//         if out == ft.want {
//             sc.Inc()
//         }
//     }
// }
//
package score
