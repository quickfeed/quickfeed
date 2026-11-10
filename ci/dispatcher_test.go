package ci

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// testJob returns a job laid out like the one RunTests creates.
func testJob(t *testing.T) *Job {
	t.Helper()
	bind, tests, assignments := t.TempDir(), t.TempDir(), t.TempDir()
	writeFile(t, bind, "labs/lab1/main.go", "package lab1")
	writeFile(t, bind, "labs/.git/config", "secret")
	writeFile(t, tests, "scripts/run.sh", "#image/qf104\necho hi")
	writeFile(t, assignments, "lab1/assignment.json", "{}")
	return &Job{
		Name:     "dat520-lab1-student-a1b2c3",
		Image:    "qf104",
		Language: "go",
		BindDir:  bind,
		ReadOnlyMounts: map[string]string{
			tests:       "/quickfeed/tests",
			assignments: "/quickfeed/assignments",
		},
		Env:      EnvVars("s3cr3t", QuickFeedPath, "labs", "lab1"),
		Commands: []string{"echo one", "echo two"},
	}
}

func TestDispatcherRequestShape(t *testing.T) {
	var (
		details dispatcherRequest
		files   = map[string]string{}
		auth    string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if r.URL.Path != "/rpc/jobs" || r.Method != http.MethodPost {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Error(err)
			return
		}
		if err := json.Unmarshal([]byte(r.MultipartForm.Value["details"][0]), &details); err != nil {
			// details may be sent as a file part
			f, err := r.MultipartForm.File["details"][0].Open()
			if err != nil {
				t.Error(err)
				return
			}
			b, _ := io.ReadAll(f)
			if err := json.Unmarshal(b, &details); err != nil {
				t.Error(err)
			}
		}
		fh := r.MultipartForm.File["code"][0]
		f, _ := fh.Open()
		b, _ := io.ReadAll(f)
		zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
		if err != nil {
			t.Error(err)
			return
		}
		for _, zf := range zr.File {
			rc, _ := zf.Open()
			c, _ := io.ReadAll(rc)
			rc.Close()
			files[zf.Name] = string(c)
		}
		_, _ = io.WriteString(w, `{"state":"completed","result":{"status":"completed","exit_code":0,"output":"ok"}}`)
	}))
	defer srv.Close()

	out, err := NewDispatcher(srv.URL+"/", "tok", WithRunTimeout(90*time.Second)).Run(context.Background(), testJob(t))
	if err != nil || out != "ok" {
		t.Fatalf("Run() = %q, %v; want ok, nil", out, err)
	}
	if auth != "Bearer tok" {
		t.Errorf("Authorization = %q", auth)
	}
	want := dispatcherRequest{
		ImageType: "qf104", AssignmentID: "lab1", Secret: "s3cr3t", SubmittedDir: "labs",
		TestsDir: "tests", AssignmentsDir: "assignments", ScriptPath: ".qf/run.sh",
		RunTimeoutSecs: 90, Name: "dat520-lab1-student-a1b2c3",
	}
	if details != want {
		t.Errorf("details = %+v, want %+v", details, want)
	}
	for _, name := range []string{"labs/lab1/main.go", "tests/scripts/run.sh", "assignments/lab1/assignment.json"} {
		if _, ok := files[name]; !ok {
			t.Errorf("zip lacks %s; has %v", name, files)
		}
	}
	for name := range files {
		if strings.Contains(name, ".git") {
			t.Errorf("zip contains %s", name)
		}
	}
	if got, want := files[".qf/run.sh"], "#!/bin/bash\necho one\necho two\n"; got != want {
		t.Errorf("script = %q, want %q", got, want)
	}
}

func TestDispatcherOutcomes(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantOut string
		check   func(error) bool
	}{
		{"ok", 200, `{"state":"completed","result":{"status":"completed","exit_code":0,"output":"out"}}`, "out",
			func(err error) bool { return err == nil }},
		{"exit code", 200, `{"state":"completed","result":{"status":"completed","exit_code":2,"output":"out"}}`, "out",
			func(err error) bool {
				var e *ContainerExitError
				return errors.As(err, &e) && e.Code == 2
			}},
		{"timed out", 200, `{"state":"completed","result":{"status":"timed_out","exit_code":null,"output":"x"}}`, timeoutMessage,
			func(err error) bool { return errors.Is(err, context.DeadlineExceeded) }},
		{"run timeout", 200, `{"state":"failed","reason":"run_timeout"}`, timeoutMessage,
			func(err error) bool { return errors.Is(err, context.DeadlineExceeded) }},
		{"worker error", 200, `{"state":"completed","result":{"status":"error","exit_code":null,"output":"bad zip"}}`, "bad zip",
			func(err error) bool { return err != nil && !errors.Is(err, context.DeadlineExceeded) }},
		{"no worker", 200, `{"state":"failed","reason":"no_worker_available"}`, "",
			func(err error) bool { return err != nil && strings.Contains(err.Error(), "no_worker_available") }},
		{"conflict", 409, `{"error":"dup"}`, "", func(err error) bool { return errors.Is(err, ErrConflict) }},
		{"unauthorized", 401, `{"error":"no"}`, "", func(err error) bool { return err != nil && !errors.Is(err, ErrConflict) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer srv.Close()
			out, err := NewDispatcher(srv.URL, "tok").Run(context.Background(), testJob(t))
			if out != tt.wantOut || !tt.check(err) {
				t.Errorf("Run() = %q, %v; want output %q", out, err, tt.wantOut)
			}
		})
	}
}

func TestDispatcherContextDeadline(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body) // the server notices a closed connection only after the body is read
		<-r.Context().Done()
		close(done)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	out, err := NewDispatcher(srv.URL, "tok").Run(ctx, testJob(t))
	if out != timeoutMessage || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Run() = %q, %v", out, err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("server did not see the request cancelled")
	}
}

func TestDispatcherRejectsUnsupportedJobs(t *testing.T) {
	d := NewDispatcher("http://127.0.0.1:1", "tok")
	// An image build, as in assignments.BuildDockerImage.
	_, err := d.Run(context.Background(), &Job{Name: "x", Image: "x", BuildContext: map[string]string{Dockerfile: "FROM x"}, Commands: []string{"true"}})
	if err == nil {
		t.Error("expected an error for a Dockerfile build job")
	}
}

// TestDispatcherSampleJob runs the sample job (docs/job in the work manager project) through
// a live dispatcher and worker, and validates the output with QuickFeed's own score package.
// Set QF_DISPATCHER_URL, QF_DISPATCHER_TOKEN and QF_SAMPLE_DIR (the docs/job directory).
func TestDispatcherSampleJob(t *testing.T) {
	url, token, sample := os.Getenv("QF_DISPATCHER_URL"), os.Getenv("QF_DISPATCHER_TOKEN"), os.Getenv("QF_SAMPLE_DIR")
	if url == "" || token == "" || sample == "" {
		t.Skip("QF_DISPATCHER_URL, QF_DISPATCHER_TOKEN and QF_SAMPLE_DIR must be set")
	}
	const repoName, assignment, secret = "meling-labs", "lab1", "integration-secret-42"
	bind, tests, assignments := t.TempDir(), t.TempDir(), t.TempDir()
	for src, dst := range map[string]string{
		filepath.Join(sample, repoName):      filepath.Join(bind, repoName),
		filepath.Join(sample, "tests"):       tests,
		filepath.Join(sample, "assignments"): assignments,
	} {
		if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
			t.Fatal(err)
		}
	}
	// Real student repositories have a go.mod; the sample lacks it.
	writeFile(t, filepath.Join(bind, repoName), "go.mod", "module meling-labs\n\ngo 1.22\n")

	content, err := os.ReadFile(filepath.Join(tests, "scripts", RunScriptFile))
	if err != nil {
		t.Fatal(err)
	}
	image, language, commands, err := ParseRunScript(string(content))
	if err != nil {
		t.Fatal(err)
	}
	job := &Job{
		Name:     "sample-lab1-student-abc123",
		Image:    image,
		Language: language,
		BindDir:  bind,
		ReadOnlyMounts: map[string]string{
			tests:       QuickFeedPath + "/tests",
			assignments: QuickFeedPath + "/assignments",
		},
		Env:      EnvVars(secret, QuickFeedPath, repoName, assignment),
		Commands: append(buildCheckCommands(language), commands...),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	out, err := NewDispatcher(url, token).Run(ctx, job)
	if err != nil {
		var exitErr *ContainerExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("Run() error: %v\noutput:\n%s", err, out)
		}
	}
	// ExtractResults keeps only scores of expected tests (Assignment.ZeroScoreTests in QuickFeed).
	expected := []*score.Score{{TestName: "TestFibonacciAG", MaxScore: 13, Weight: 5}}
	results, err := score.ExtractResults(out, secret, time.Second, expected)
	if err != nil {
		t.Logf("ExtractResults: %v", err)
	}
	var found bool
	for _, sc := range results.Scores {
		t.Logf("score: %s %d/%d weight %d", sc.GetTestName(), sc.GetScore(), sc.GetMaxScore(), sc.GetWeight())
		if sc.GetTestName() == "TestFibonacciAG" {
			found = true
			if sc.GetScore() != 13 || sc.GetMaxScore() != 13 || sc.GetWeight() != 5 {
				t.Errorf("TestFibonacciAG = %d/%d weight %d, want 13/13 weight 5", sc.GetScore(), sc.GetMaxScore(), sc.GetWeight())
			}
		}
	}
	if !found {
		t.Errorf("no TestFibonacciAG score in output:\n%s", out)
	}
	if red := score.Redact(out, secret); strings.Contains(red, secret) {
		t.Error("secret still present after Redact")
	}
}
