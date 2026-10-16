package score_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

func TestWrapperOutputAndRecovery(t *testing.T) {
	out := runWrapperTests(t, "-race", "-v", "-run", "Test(TopError|Error|Fatal|Fatalf|SubError|Parallel|TopPanic|SubPanic|AfterPanic|Concurrent|Arithmetic|Fail)$")
	if strings.Contains(out, "WARNING: DATA RACE") {
		t.Fatalf("wrapper tests raced:\n%s", out)
	}
	got := parseWrapperScores(t, out)
	if len(got) != 12 {
		t.Fatalf("got %d scores, want 12:\n%s", len(got), out)
	}
	for name, want := range map[string]struct{ value, max int32 }{
		"TestTopError":   {2, 2},
		"TestError":      {2, 2},
		"TestFatal":      {0, 2},
		"TestFatalf":     {0, 2},
		"TestSubError":   {2, 2},
		"TestParallel":   {4, 4},
		"TestTopPanic":   {0, 2},
		"TestSubPanic":   {0, 2},
		"TestAfterPanic": {1, 2},
		"TestConcurrent": {40, 40},
		"TestArithmetic": {4, 5},
		"TestFail":       {0, 2},
	} {
		sc, ok := got[name]
		if !ok || sc.GetScore() != want.value || sc.GetMaxScore() != want.max {
			t.Errorf("%s score = %v, want %d/%d", name, sc, want.value, want.max)
		}
	}
	for name, marker := range map[string]string{
		"TestTopError": "top failure",
		"TestSubError": "sub failure",
	} {
		callLine := fixtureCallLine(t, marker)
		if name == "TestSubError" {
			marker = "TestSubError/child: " + marker
		}
		want := "wrapper_test.go:" + strconv.Itoa(callLine) + ": " + marker
		if !strings.Contains(got[name].GetTestDetails(), want) {
			t.Errorf("%s details = %q, want %q", name, got[name].GetTestDetails(), want)
		}
	}
	for name, detail := range map[string]string{
		"TestError":  "plain failure",
		"TestFatal":  "fatal failure",
		"TestFatalf": "fatal formatted",
	} {
		if !strings.Contains(got[name].GetTestDetails(), detail) {
			t.Errorf("%s details = %q, want %q", name, got[name].GetTestDetails(), detail)
		}
	}
	if n := strings.Count(got["TestConcurrent"].GetTestDetails(), "concurrent failure"); n != 40 {
		t.Errorf("concurrent details contain %d failures, want 40", n)
	}
	for _, marker := range []string{"TestTopPanic panicked: top panic marker", "TestSubPanic/child panicked: sub panic marker", "--- PASS: TestAfterPanic", "--- PASS: TestArithmetic"} {
		if !strings.Contains(out, marker) {
			t.Errorf("missing %q from subprocess output", marker)
		}
	}
	for name := range got {
		if !strings.Contains(out, "=== ATTR  "+name+" score {\"TestName\":\""+name+"\"") {
			t.Errorf("score for %s was not reported as a test attribute", name)
		}
	}
	parallelScore := strings.Index(out, `"TestName":"TestParallel"`)
	for _, name := range []string{"a", "b", "c", "d"} {
		marker := "parallel done TestParallel/" + name
		if idx := strings.Index(out, marker); idx < 0 || idx > parallelScore {
			t.Errorf("score was printed before %s", marker)
		}
	}
}

// TestWrapperOutputWithoutVerbose checks the fallback used when the framework
// does not emit attributes: a bare JSON line that follows an empty line.
func TestWrapperOutputWithoutVerbose(t *testing.T) {
	out := runWrapperTests(t, "-run", "Test(Error|Arithmetic)$")
	if strings.Contains(out, "=== ATTR") {
		t.Errorf("non-verbose output contains attributes:\n%s", out)
	}
	got := parseWrapperScores(t, out)
	if len(got) != 2 {
		t.Fatalf("got %d scores, want 2:\n%s", len(got), out)
	}
	for name := range got {
		// The first score has no output before it to be separated from.
		if !strings.Contains("\n"+out, "\n\n{\"TestName\":\""+name+"\"") {
			t.Errorf("score for %s was not printed after an empty line", name)
		}
	}
}

func TestMutedWrapperOutput(t *testing.T) {
	out := runWrapperTests(t, "-v", "-run", "^TestMuted$")
	if !strings.Contains(out, "--- PASS: TestMuted") || strings.Contains(out, `"TestName"`) {
		t.Errorf("unscored test output = %q", out)
	}
}

func TestUnregisteredWrapper(t *testing.T) {
	for _, name := range []string{"TestUnregistered", "TestUnregisteredMin"} {
		out := runWrapperTests(t, "-v", "-run", "^"+name+"$")
		if !strings.Contains(out, `test "`+name+`" is not registered; call Add`) {
			t.Errorf("unregistered test output = %q", out)
		}
	}
}

func runWrapperTests(t *testing.T, args ...string) string {
	t.Helper()
	args = append([]string{"test", "-count=1"}, args...)
	args = append(args, "./testdata/wrapper")
	cmd := exec.Command("go", args...)
	out, _ := cmd.CombinedOutput() // fixture tests intentionally fail
	return string(out)
}

func parseWrapperScores(t *testing.T, out string) map[string]*score.Score {
	t.Helper()
	scores := make(map[string]*score.Score)
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimSpace(line)
		// Verbose runs report a score as "=== ATTR  <test> score <json>".
		if _, after, ok := strings.Cut(line, " score "); ok && strings.HasPrefix(line, "=== ATTR") {
			line = after
		}
		if !strings.HasPrefix(line, `{"TestName"`) {
			continue
		}
		var sc score.Score
		if err := json.Unmarshal([]byte(line), &sc); err != nil {
			t.Fatalf("parse score %q: %v", line, err)
		}
		if _, exists := scores[sc.GetTestName()]; exists {
			t.Errorf("duplicate score for %s", sc.GetTestName())
		}
		scores[sc.GetTestName()] = &sc
	}
	return scores
}

func fixtureCallLine(t *testing.T, marker string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "wrapper", "wrapper_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, marker) && strings.Contains(line, "// ") {
			return i + 1
		}
	}
	t.Fatalf("no call site for %q", marker)
	return 0
}
