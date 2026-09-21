package score

import (
	"io"
	"os"
	"strings"
	"testing"
)

// attrRecorder stands in for the *testing.T that a test passes to Print.
type attrRecorder struct {
	key, value string
	calls      int
}

func (a *attrRecorder) Attr(key, value string) {
	a.key, a.value, a.calls = key, value, a.calls+1
}

func TestEmitReportsScoreAsAttributeWhenChatty(t *testing.T) {
	// testing.T.Attr emits nothing unless the run is chatty, but when it does,
	// it attributes the score to the test that reported it even when the output
	// of parallel tests is interleaved. That is worth more than the blank line
	// the printed form relies on.
	sc := &Score{TestName: "TestStack", Score: 3, MaxScore: 5, Weight: 1}
	rec := &attrRecorder{}

	printed := capture(t, func() { sc.emit(rec, true) })

	if rec.calls != 1 {
		t.Fatalf("Attr called %d times, want 1", rec.calls)
	}
	if rec.key != scoreAttrKey {
		t.Errorf("Attr key = %q, want %q", rec.key, scoreAttrKey)
	}
	if !strings.Contains(rec.value, `"TestName":"TestStack"`) {
		t.Errorf("Attr value = %q, want the score JSON", rec.value)
	}
	if strings.ContainsAny(rec.value, "\r\n") {
		t.Errorf("Attr value = %q, want no newline; testing.T.Attr rejects one", rec.value)
	}
	if printed != "" {
		t.Errorf("printed %q, want nothing; the attribute already reported the score", printed)
	}
}

func TestEmitPrintsScoreWhenNotChatty(t *testing.T) {
	// A run script that omits -v gets no attributes at all, so the score must
	// still be printed for the run to be scored.
	sc := &Score{TestName: "TestStack", Score: 3, MaxScore: 5, Weight: 1}
	rec := &attrRecorder{}

	printed := capture(t, func() { sc.emit(rec, false) })

	if rec.calls != 0 {
		t.Errorf("Attr called %d times, want 0", rec.calls)
	}
	if !strings.Contains(printed, `"TestName":"TestStack"`) {
		t.Errorf("printed %q, want the score JSON", printed)
	}
	// Scanning a long line of student output for a score is costly, so the
	// score starts on a line of its own.
	if !strings.HasPrefix(printed, "\n") {
		t.Errorf("printed %q, want it to start on a new line", printed)
	}
}

// capture returns what fn wrote to standard output.
func capture(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	os.Stdout = original
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
