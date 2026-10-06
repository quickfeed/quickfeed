package hookforward

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

const testTimeout = 5 * time.Second

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func validOptions() Options {
	return Options{
		Orgs:   []string{"qf101"},
		URL:    URL("127.0.0.1:443"),
		Events: []string{"push", "installation"},
		Secret: "s3cret",
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		domain string
		want   string
	}{
		{domain: "127.0.0.1:443", want: "https://127.0.0.1:443/hook/"},
		{domain: "localhost:8080", want: "https://localhost:8080/hook/"},
		{domain: "uis.itest.run:443", want: "https://uis.itest.run:443/hook/"},
	}
	for _, test := range tests {
		got := URL(test.domain)
		if got != test.want {
			t.Errorf("URL(%q) = %q, want %q", test.domain, got, test.want)
		}
		// The trailing slash is the whole point: without it the forwarded POST
		// is redirected to a GET and the payload is lost.
		if !strings.HasSuffix(got, "/hook/") {
			t.Errorf("URL(%q) = %q, want suffix %q", test.domain, got, "/hook/")
		}
	}
}

func TestOptionsValidate(t *testing.T) {
	tests := map[string]struct {
		mutate  func(*Options)
		wantErr string
	}{
		"valid":        {mutate: func(*Options) {}},
		"no orgs":      {mutate: func(o *Options) { o.Orgs = nil }, wantErr: "no organizations"},
		"no events":    {mutate: func(o *Options) { o.Events = nil }, wantErr: "no webhook event types"},
		"no secret":    {mutate: func(o *Options) { o.Secret = "" }, wantErr: "QUICKFEED_WEBHOOK_SECRET"},
		"no url":       {mutate: func(o *Options) { o.URL = "" }, wantErr: "must end with"},
		"no slash":     {mutate: func(o *Options) { o.URL = "https://127.0.0.1:443/hook" }, wantErr: "must end with"},
		"wrong route":  {mutate: func(o *Options) { o.URL = "https://127.0.0.1:443/" }, wantErr: "must end with"},
		"hook infix":   {mutate: func(o *Options) { o.URL = "https://127.0.0.1:443/hook/x" }, wantErr: "must end with"},
		"custom route": {mutate: func(o *Options) { o.URL = "https://qf.example.com:443/hook/" }},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opts := validOptions()
			test.mutate(&opts)
			err := opts.validate()
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("validate() = %v, want <nil>", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validate() = <nil>, want error containing %q", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("validate() = %q, want error containing %q", err, test.wantErr)
			}
		})
	}
}

// fakeGh writes an executable named gh to dir, printing the given output,
// and puts dir first on PATH. It is a stand-in for the GitHub CLI; the tests
// never run the real one.
func fakeGh(t *testing.T, output string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake gh is a shell script")
	}
	dir := t.TempDir()
	// The script must use only shell builtins, since dir is the whole PATH.
	script := "#!/bin/sh\necho '" + output + "'\n"
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

func TestCheckMissingGh(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	err := ghRunner{}.Check(t.Context())
	if err == nil {
		t.Fatal("Check() = <nil>, want error")
	}
	if !strings.Contains(err.Error(), setupHint) {
		t.Errorf("Check() = %q, want error containing %q", err, setupHint)
	}
}

func TestCheckMissingExtension(t *testing.T) {
	fakeGh(t, "github/gh-copilot\tv1.0.0")
	err := ghRunner{}.Check(t.Context())
	if err == nil {
		t.Fatal("Check() = <nil>, want error")
	}
	if !strings.Contains(err.Error(), extension) || !strings.Contains(err.Error(), setupHint) {
		t.Errorf("Check() = %q, want error containing %q and %q", err, extension, setupHint)
	}
}

func TestCheckInstalledExtension(t *testing.T) {
	fakeGh(t, "cli/gh-webhook\tv0.3.0")
	if err := (ghRunner{}).Check(t.Context()); err != nil {
		t.Errorf("Check() = %v, want <nil>", err)
	}
}

// fakeRunner records the forwarding commands it is asked to run, and blocks
// each of them until the context is cancelled, as the real forwarder does.
type fakeRunner struct {
	checkErr error
	started  chan []string
	stopped  chan struct{}

	mu   sync.Mutex
	args [][]string
}

func newFakeRunner(orgs int) *fakeRunner {
	return &fakeRunner{
		started: make(chan []string, orgs),
		stopped: make(chan struct{}, orgs),
	}
}

func (f *fakeRunner) Check(context.Context) error {
	return f.checkErr
}

func (f *fakeRunner) Forward(ctx context.Context, _ *slog.Logger, args []string) error {
	f.mu.Lock()
	f.args = append(f.args, args)
	f.mu.Unlock()
	f.started <- args
	<-ctx.Done()
	f.stopped <- struct{}{}
	return ctx.Err()
}

func TestStartCheckFails(t *testing.T) {
	runner := newFakeRunner(1)
	runner.checkErr = errors.New("missing prerequisite")
	opts := validOptions()
	opts.Runner = runner
	if err := Start(t.Context(), testLogger(), opts); !errors.Is(err, runner.checkErr) {
		t.Errorf("Start() = %v, want %v", err, runner.checkErr)
	}
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if len(runner.args) != 0 {
		t.Errorf("Start() ran %d commands, want 0 when the check fails", len(runner.args))
	}
}

func TestStartInvalidOptions(t *testing.T) {
	runner := newFakeRunner(1)
	opts := validOptions()
	opts.Runner = runner
	opts.Orgs = nil
	if err := Start(t.Context(), testLogger(), opts); err == nil {
		t.Error("Start() = <nil>, want error for options without organizations")
	}
}

func TestStartOnePerOrg(t *testing.T) {
	orgs := []string{"qf101", "dat320"}
	runner := newFakeRunner(len(orgs))
	opts := validOptions()
	opts.Orgs = orgs
	opts.Runner = runner

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := Start(ctx, testLogger(), opts); err != nil {
		t.Fatalf("Start() = %v, want <nil>", err)
	}

	want := map[string][]string{}
	for _, org := range orgs {
		want[org] = []string{
			"webhook", "forward",
			"--org=" + org,
			"--events=push,installation",
			"--url=https://127.0.0.1:443/hook/",
			"--secret=s3cret",
		}
	}
	for range orgs {
		select {
		case args := <-runner.started:
			org, ok := orgOf(args)
			if !ok {
				t.Fatalf("Forward(%q): no --org argument", args)
			}
			if !slices.Equal(args, want[org]) {
				t.Errorf("Forward() args = %q, want %q", args, want[org])
			}
			delete(want, org)
		case <-time.After(testTimeout):
			t.Fatal("timed out waiting for a forwarding command to start")
		}
	}
	if len(want) > 0 {
		t.Errorf("no forwarding command started for %v", want)
	}

	// Cancelling the context must stop every forwarder, as Ctrl+C does.
	cancel()
	for range orgs {
		select {
		case <-runner.stopped:
		case <-time.After(testTimeout):
			t.Fatal("timed out waiting for a forwarding command to stop")
		}
	}
}

func orgOf(args []string) (string, bool) {
	for _, arg := range args {
		if org, ok := strings.CutPrefix(arg, "--org="); ok {
			return org, true
		}
	}
	return "", false
}
