package score

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// T combines a test handle with a score shared by its subtests.
type T struct {
	*testing.T
	st *state
}

type state struct {
	mu      sync.Mutex
	sc      *Score
	printed bool
	noPrint bool
}

// NewUnscored wraps a plain Go test without printing a QuickFeed score.
func NewUnscored(t *testing.T) *T {
	return &T{T: t, st: &state{sc: &Score{TestName: t.Name()}, noPrint: true}}
}

// Score returns the shared score. Callers must not mutate it or read it while
// score methods are running concurrently.
func (t *T) Score() *Score {
	return t.st.sc
}

// Inc increments the score up to its maximum.
func (t *T) Inc() {
	t.st.mu.Lock()
	defer t.st.mu.Unlock()
	t.st.sc.Inc()
}

// IncBy increments the score by n, up to its maximum.
func (t *T) IncBy(n int) {
	t.st.mu.Lock()
	defer t.st.mu.Unlock()
	t.st.sc.IncBy(n)
}

// Dec decrements the score down to zero.
func (t *T) Dec() {
	t.st.mu.Lock()
	defer t.st.mu.Unlock()
	t.st.sc.Dec()
}

// DecBy decrements the score by n, down to zero.
func (t *T) DecBy(n int) {
	t.st.mu.Lock()
	defer t.st.mu.Unlock()
	t.st.sc.DecBy(n)
}

// Fail sets the score to zero.
func (t *T) Fail() {
	t.st.mu.Lock()
	t.st.sc.Fail()
	t.st.mu.Unlock()
}

// Error records a failure at its caller and reports it to the test.
func (t *T) Error(args ...any) {
	t.Helper()
	t.record(fmt.Sprintln(args...))
	t.T.Error(args...)
}

// Errorf records a formatted failure at its caller and reports it to the test.
func (t *T) Errorf(format string, args ...any) {
	t.Helper()
	t.record(fmt.Sprintf(format, args...))
	t.T.Errorf(format, args...)
}

// Fatal records a failure and stops the current test goroutine.
// Like testing.T.Fatal, it must be called from the test goroutine.
func (t *T) Fatal(args ...any) {
	t.Helper()
	t.record(fmt.Sprintln(args...))
	t.T.Fatal(args...)
}

// Fatalf records a formatted failure and stops the current test goroutine.
// Like testing.T.Fatalf, it must be called from the test goroutine.
func (t *T) Fatalf(format string, args ...any) {
	t.Helper()
	t.record(fmt.Sprintf(format, args...))
	t.T.Fatalf(format, args...)
}

func (t *T) record(message string) {
	_, file, line, ok := runtime.Caller(2)
	message = strings.TrimSuffix(message, "\n")
	if strings.Contains(t.Name(), "/") {
		message = t.Name() + ": " + message
	}
	if ok {
		message = fmt.Sprintf("%s:%d: %s", file, line, message)
	}
	t.st.mu.Lock()
	t.st.sc.TestDetails += message + "\n"
	t.st.mu.Unlock()
}

// Run executes a subtest with the same score and recovers its panics.
func (t *T) Run(name string, f func(*T)) bool {
	t.Helper()
	return t.T.Run(name, func(sub *testing.T) {
		child := &T{T: sub, st: t.st}
		defer child.Recover()
		f(child)
	})
}

// Recover handles a panic in the current test goroutine and zeros its score.
// Defer it in a top-level test. Run installs it for each subtest automatically.
func (t *T) Recover(msg ...string) {
	if r := recover(); r != nil {
		t.Fail()
		t.T.Fail()
		printPanicMessage(t.Name(), firstMessage(msg), r)
	}
}

func (t *T) print() {
	t.st.mu.Lock()
	defer t.st.mu.Unlock()
	if t.st.printed || t.st.noPrint {
		return
	}
	t.st.printed = true
	t.st.sc.emit(t.T, testing.Verbose())
}
