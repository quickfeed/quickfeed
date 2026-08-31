package score_test

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/score"
)

func TestRelativeScore(t *testing.T) {
	sc := &score.Score{
		TestName: t.Name(),
		MaxScore: 10,
		Score:    3,
	}
	rs := sc.RelativeScore()
	expectedRS := t.Name() + ": score = 3/10 = 0.3"
	if rs != expectedRS {
		t.Errorf(`RelativeScore() = %q, expected %q`, rs, expectedRS)
	}
}

func TestNormalize(t *testing.T) {
	sc := &score.Score{
		TestName: t.Name(),
		MaxScore: 100,
		Score:    33,
	}
	newMaxScore := 50
	sc.Normalize(newMaxScore)
	expectedScore := int32(17)
	if sc.GetScore() != expectedScore {
		t.Errorf("Normalize(%d) = %d, expected %d", newMaxScore, sc.GetScore(), expectedScore)
	}
}

// These tests simulate the behavior of our test flow, i.e., code is tested in
// Docker containers, where the output is captured and parsed. We expect that
// our helper methods (sc.Errorf, sc.Fatalf) will correctly add the test details
// to the score object, and that parsing the test details will be correct.
func TestScoreDetails(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		fatal    int // index of the message that should cause a fatal error
		expected []string
	}{
		{
			name:     "no messages",
			messages: []string{},
			fatal:    -1,
			expected: []string{},
		},
		{
			name:     "one message",
			messages: []string{"first"},
			fatal:    -1,
			expected: []string{"first"},
		},
		{
			name:     "two messages",
			messages: []string{"first", "second"},
			fatal:    1,
			expected: []string{"first", "second"},
		},
		{
			name:     "two messages",
			messages: []string{"first", "second"},
			fatal:    0,
			expected: []string{"first"},
		},
		{
			name:     "three messages",
			messages: []string{"first", "second", "third"},
			fatal:    1,
			expected: []string{"first", "second"},
		},
	}

	r, cleanup := redirectStdout(t)
	defer cleanup()

	var wg sync.WaitGroup
	for _, test := range tests {
		sc := &score.Score{}

		wg.Go(func() {
			// run using the subT mock in a goroutine to avoid Fatalf
			// from exiting the test via t.FailNow (runtime.Goexit)
			subT := &testing.T{}
			for i, m := range test.messages {
				if test.fatal == i {
					sc.Fatalf(subT, "%s", m)
				} else {
					sc.Errorf(subT, "%s", m)
				}
			}
		})
		wg.Wait()

		sc.Print(t)
		out := make([]byte, 1024)
		n, err := r.Read(out)
		if err != nil {
			t.Fatalf("Failed to read from pipe: %v", err)
		}

		parsedScore := &score.Score{}
		if err := json.Unmarshal(out[:n], parsedScore); err != nil {
			t.Fatalf("Failed to unmarshal score: %v", err)
		}

		gotMessages := parseTestDetails(parsedScore.TestDetails)
		if diff := cmp.Diff(test.expected, gotMessages); diff != "" {
			t.Errorf("TestDetails mismatch (-want +got):\n%s", diff)
		}
	}
}

// redirectStdout redirects os.Stdout to a pipe and returns
// the read end of the pipe for capturing output, and
// a cleanup function to restore os.Stdout and close the pipe.
func redirectStdout(t *testing.T) (*os.File, func()) {
	t.Helper()
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	// redirect os.Stdout to the pipe's write end
	os.Stdout = w
	return r, func() {
		os.Stdout = originalStdout
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

// parseTestDetails splits the details string into individual lines and removes the prefix
// that contains the file and line number of the test function that called Errorf or Error.
func parseTestDetails(details string) []string {
	lines := strings.Split(details, "\n")

	// The final line is an empty string, remove it
	lines = lines[:len(lines)-1]

	// Remove test details prefix, e.g. "dir/file.go:123: "
	for i, line := range lines {
		if _, after, ok := strings.Cut(line, ": "); ok {
			lines[i] = after
		}
	}
	return lines
}
