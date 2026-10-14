package score

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testLogSecret is the session secret embedded in the score lines of every fixture.
const testLogSecret = "Ie7cJP1sYQvqO2Xk"

// fixtures names every captured go test output under testdata.
var fixtures = []string{
	"attr_v",
	"legacy_v",
	"attr_json_go125",
	"attr_json_go127",
	"parallel_v",
	"panic_v",
}

func loadTestLog(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "testlog", name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestScanAttributesOutputToTests(t *testing.T) {
	log := scanTestLog(loadTestLog(t, "attr_v"))

	wantOrder := []string{"TestStack", "TestQueue", "TestQueue/Drain", "TestSkipped"}
	gotOrder := make([]string, len(log.Tests))
	for i, tr := range log.Tests {
		gotOrder[i] = tr.Name
	}
	if strings.Join(gotOrder, ",") != strings.Join(wantOrder, ",") {
		t.Errorf("test order = %q, want %q", gotOrder, wantOrder)
	}

	stack := log.Test("TestStack")
	if stack.Status != TestStatus_PASSED {
		t.Errorf("TestStack status = %v, want %v", stack.Status, TestStatus_PASSED)
	}
	if want := "stack: pushing 3 elements"; strings.Join(stack.Output, "\n") != want {
		t.Errorf("TestStack output = %q, want %q", stack.Output, want)
	}
	if len(stack.Failures) != 0 {
		t.Errorf("TestStack failures = %q, want none", stack.Failures)
	}
	if len(stack.Scores) != 1 {
		t.Fatalf("TestStack scores = %q, want 1", stack.Scores)
	}
	if !strings.Contains(stack.Scores[0], `"TestName":"TestStack"`) {
		t.Errorf("TestStack score = %q, want the score JSON emitted by the test", stack.Scores[0])
	}
}

func TestScanSeparatesFailuresFromOutput(t *testing.T) {
	log := scanTestLog(loadTestLog(t, "attr_v"))

	queue := log.Test("TestQueue")
	if queue.Status != TestStatus_FAILED {
		t.Errorf("TestQueue status = %v, want %v", queue.Status, TestStatus_FAILED)
	}
	// A failing test's indented diagnostics explain the failure, so they belong
	// in Failures; its unindented writes are ordinary output.
	failures := strings.Join(queue.Failures, "\n")
	if !strings.Contains(failures, "queue_test.go:44: Pop() = <nil>, want: x") {
		t.Errorf("TestQueue failures = %q, want the Errorf message", queue.Failures)
	}
	if want := "queue: raw diagnostic from student code"; strings.Join(queue.Output, "\n") != want {
		t.Errorf("TestQueue output = %q, want %q", queue.Output, want)
	}

	// The subtest's diagnostic belongs to the subtest, not to its parent.
	if strings.Contains(failures, "Drain() left 2 items") {
		t.Errorf("TestQueue failures = %q, want the subtest's message attributed to the subtest", queue.Failures)
	}
	drain := log.Test("TestQueue/Drain")
	if drain.Status != TestStatus_FAILED {
		t.Errorf("TestQueue/Drain status = %v, want %v", drain.Status, TestStatus_FAILED)
	}
	if want := "queue_test.go:48: Drain() left 2 items"; strings.Join(drain.Failures, "\n") != want {
		t.Errorf("TestQueue/Drain failures = %q, want %q", drain.Failures, want)
	}
}

func TestScanRecordsSkippedTests(t *testing.T) {
	log := scanTestLog(loadTestLog(t, "attr_v"))

	skipped := log.Test("TestSkipped")
	if skipped.Status != TestStatus_SKIPPED {
		t.Errorf("TestSkipped status = %v, want %v", skipped.Status, TestStatus_SKIPPED)
	}
	// A skip reason is not a failure.
	if len(skipped.Failures) != 0 {
		t.Errorf("TestSkipped failures = %q, want none", skipped.Failures)
	}
	if want := "queue_test.go:55: not implemented for this assignment"; strings.Join(skipped.Output, "\n") != want {
		t.Errorf("TestSkipped output = %q, want %q", skipped.Output, want)
	}
}

func TestScanAttributesBareScoreLines(t *testing.T) {
	// Courses that have not adopted t.Attr print the score with fmt.Println.
	log := scanTestLog(loadTestLog(t, "legacy_v"))

	stack := log.Test("TestStack")
	if len(stack.Scores) != 1 || !strings.Contains(stack.Scores[0], `"TestName":"TestStack"`) {
		t.Errorf("TestStack scores = %q, want the bare score line attributed to the test", stack.Scores)
	}
	if stack.Status != TestStatus_PASSED {
		t.Errorf("TestStack status = %v, want %v", stack.Status, TestStatus_PASSED)
	}
}

func TestScanReadsTest2JSONEvents(t *testing.T) {
	// Go 1.25 and 1.26 emit no OutputType; Go 1.27 tags Errorf output as
	// "error". Attribution must not depend on which of the two ran.
	for _, name := range []string{"attr_json_go125", "attr_json_go127"} {
		t.Run(name, func(t *testing.T) {
			log := scanTestLog(loadTestLog(t, name))

			queue := log.Test("TestQueue")
			if queue == nil {
				t.Fatal("TestQueue not found")
			}
			if queue.Status != TestStatus_FAILED {
				t.Errorf("TestQueue status = %v, want %v", queue.Status, TestStatus_FAILED)
			}
			if !strings.Contains(strings.Join(queue.Failures, "\n"), "Pop() = <nil>, want: x") {
				t.Errorf("TestQueue failures = %q, want the Errorf message", queue.Failures)
			}
			if !strings.Contains(strings.Join(queue.Output, "\n"), "queue: raw diagnostic from student code") {
				t.Errorf("TestQueue output = %q, want the test's own write", queue.Output)
			}
			// The attribute event and the "=== ATTR" line it also prints report
			// the same score; it must be recorded once.
			if len(queue.Scores) != 1 {
				t.Errorf("TestQueue scores = %q, want 1", queue.Scores)
			}
			if got := log.Test("TestStack").Elapsed; got < 0 {
				t.Errorf("TestStack elapsed = %v, want a non-negative duration", got)
			}
		})
	}
}

func TestScanSeparatesLoggingFromFailuresWhenTagged(t *testing.T) {
	// Go 1.27 tags Errorf output as "error" and leaves Logf output untagged,
	// which is the one thing the JSON form knows that the text form cannot: a
	// failing test's ordinary logging stays out of its failures.
	log := scanTestLog(loadTestLog(t, "attr_json_go127"))

	queue := log.Test("TestQueue")
	if strings.Contains(strings.Join(queue.Failures, "\n"), "queue length is 0") {
		t.Errorf("TestQueue failures = %q, want the Logf message left out", queue.Failures)
	}
	if !strings.Contains(strings.Join(queue.Output, "\n"), "queue length is 0") {
		t.Errorf("TestQueue output = %q, want the Logf message", queue.Output)
	}

	// Without the tagging, a failing test's diagnostics are indistinguishable,
	// so they are all treated as explaining the failure.
	text := scanTestLog(loadTestLog(t, "attr_v"))
	if !strings.Contains(strings.Join(text.Test("TestQueue").Failures, "\n"), "queue length is 0") {
		t.Errorf("TestQueue failures = %q, want every diagnostic of a failing test", text.Test("TestQueue").Failures)
	}
}

func TestScanAttributesFailuresUnderParallelism(t *testing.T) {
	// Go interleaves the raw writes of parallel tests, so their attribution is
	// best effort in every mode. Failures and scores go through the framework's
	// per-test framing and must still be attributed correctly.
	log := scanTestLog(loadTestLog(t, "parallel_v"))

	alpha := log.Test("TestAlpha")
	if alpha.Status != TestStatus_FAILED {
		t.Errorf("TestAlpha status = %v, want %v", alpha.Status, TestStatus_FAILED)
	}
	if want := "queue_test.go:33: alpha broke"; strings.Join(alpha.Failures, "\n") != want {
		t.Errorf("TestAlpha failures = %q, want %q", alpha.Failures, want)
	}
	if len(alpha.Scores) != 1 {
		t.Errorf("TestAlpha scores = %q, want 1", alpha.Scores)
	}
	beta := log.Test("TestBeta")
	if beta.Status != TestStatus_PASSED {
		t.Errorf("TestBeta status = %v, want %v", beta.Status, TestStatus_PASSED)
	}
	if len(beta.Scores) != 1 {
		t.Errorf("TestBeta scores = %q, want 1", beta.Scores)
	}
}

func TestScanLeavesPanicTraceUnattributed(t *testing.T) {
	// A panic trace is printed after the test's own framing has ended, so it
	// belongs to the run rather than to any one test.
	log := scanTestLog(loadTestLog(t, "panic_v"))

	unattributed := strings.Join(log.Unattributed, "\n")
	if !strings.Contains(unattributed, "panic: assignment to entry in nil map") {
		t.Errorf("unattributed = %q, want the panic message", unattributed)
	}
	if got := log.Test("TestPanics"); got != nil && strings.Contains(strings.Join(got.Output, "\n"), "goroutine 35") {
		t.Errorf("TestPanics output = %q, want the stack trace left unattributed", got.Output)
	}
}

func TestScanDropsFramingAndScoreLines(t *testing.T) {
	// Framing lines are the framework talking about the test, not the test's
	// own output, and a score line carries the run's session secret, which must
	// never reach a student. Neither may survive in anything the testLogScanner
	// attributes or leaves over.
	for _, name := range fixtures {
		t.Run(name, func(t *testing.T) {
			log := scanTestLog(loadTestLog(t, name))

			var everything []string
			everything = append(everything, log.Unattributed...)
			for _, tr := range log.Tests {
				everything = append(everything, tr.Output...)
				everything = append(everything, tr.Failures...)
			}
			for _, line := range everything {
				if strings.Contains(line, testLogSecret) {
					t.Errorf("line %q leaks the session secret", line)
				}
				for _, framing := range []string{"=== RUN", "=== ATTR", "=== PAUSE", "=== CONT", "=== NAME", "--- PASS:", "--- FAIL:", "--- SKIP:"} {
					if strings.HasPrefix(strings.TrimSpace(line), framing) {
						t.Errorf("line %q is framing, want it dropped", line)
					}
				}
			}
		})
	}
}

func TestScanKeepsRunScriptOutputUnattributed(t *testing.T) {
	// A run script prints before and after the test command, and the build
	// check runs ahead of both. None of it belongs to a test.
	out := "*** Preparing Test Execution for lab1 ***\n" +
		loadTestLog(t, "attr_v") +
		"\n*** Finished Running Tests in 3 seconds ***\n"
	log := scanTestLog(out)

	unattributed := strings.Join(log.Unattributed, "\n")
	for _, want := range []string{"*** Preparing Test Execution for lab1 ***", "*** Finished Running Tests in 3 seconds ***"} {
		if !strings.Contains(unattributed, want) {
			t.Errorf("unattributed = %q, want it to contain %q", unattributed, want)
		}
	}
	if log.Test("TestStack") == nil {
		t.Error("TestStack not found; the surrounding script output must not disturb attribution")
	}
}

func TestScanKeepsScoreLinesOutsideAnyTest(t *testing.T) {
	// A registry prints its zero scores from TestMain, before any test runs.
	// Those lines belong to no test, but dropping them would lose scores.
	zero := `{"Secret":"` + testLogSecret + `","TestName":"TestStack","Score":0,"MaxScore":5,"Weight":1}`
	log := scanTestLog(zero + "\n" + loadTestLog(t, "attr_v"))

	if len(log.Scores) != 1 || log.Scores[0] != zero {
		t.Errorf("unattributed scores = %q, want %q", log.Scores, zero)
	}
	if strings.Contains(strings.Join(log.Unattributed, "\n"), testLogSecret) {
		t.Error("a score line reached the unattributed output, leaking the session secret")
	}
}
