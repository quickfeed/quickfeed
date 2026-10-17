// Package hookforward forwards GitHub webhook events to a QuickFeed server
// that GitHub cannot reach directly, such as one running on localhost.
//
// Forwarding runs the gh-webhook extension of the GitHub CLI, one process per
// course organization, for the lifetime of the given context. QuickFeed knows
// every argument the command needs, so a developer no longer has to assemble it
// by hand. The URL in particular is built from the server's own domain and
// webhook route, so its trailing slash is correct by construction: without it,
// the POST from the forwarder is redirected (301) to /hook/, the redirect turns
// it into a GET, and the payload is lost with neither an event nor an error.
package hookforward

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/web/auth"
)

const (
	// ghCommand is the GitHub CLI binary; it must be on PATH.
	ghCommand = "gh"
	// extension is the gh extension that does the actual forwarding.
	extension = "gh-webhook"
	// setupHint names the target that installs and authorizes both of the above.
	setupHint = "run 'make webhook-setup'"
)

const (
	// streamLabel names the subprocess output stream a logged line came from.
	streamLabel = "stream"
	// urlLabel is the attribute holding the webhook endpoint events are sent to.
	urlLabel = "url"
	// eventsLabel is the attribute holding the forwarded webhook event types.
	eventsLabel = "events"
)

// URL returns the webhook endpoint of a QuickFeed server served on the given
// domain, e.g., the value of env.DomainWithPort.
func URL(domainWithPort string) string {
	return "https://" + domainWithPort + auth.Hook
}

// Options configures webhook forwarding.
type Options struct {
	// Orgs are the GitHub organizations to forward events from; each gets its
	// own forwarding process.
	Orgs []string
	// URL is the webhook endpoint of the running server; see URL.
	URL string
	// Events are the webhook event types to forward; see hooks.Events.
	Events []string
	// Secret is the shared secret that GitHub signs forwarded payloads with;
	// it must match the server's QUICKFEED_WEBHOOK_SECRET.
	Secret string
	// Runner runs the external commands. Leave nil to run the gh binary on
	// PATH; tests substitute a fake to avoid depending on an installed gh.
	Runner Runner
}

func (o Options) validate() error {
	if len(o.Orgs) == 0 {
		return errors.New("no organizations to forward webhook events from")
	}
	if !strings.HasSuffix(o.URL, auth.Hook) {
		return fmt.Errorf("webhook URL %q must end with %q", o.URL, auth.Hook)
	}
	if len(o.Events) == 0 {
		return errors.New("no webhook event types to forward")
	}
	if o.Secret == "" {
		return errors.New("no webhook secret; set QUICKFEED_WEBHOOK_SECRET in .env")
	}
	return nil
}

// Runner runs the external commands that forwarding depends on.
type Runner interface {
	// Check reports whether the tools needed for forwarding are available,
	// returning an error naming what is missing.
	Check(ctx context.Context) error
	// Forward runs one forwarding command to completion, logging each line it
	// writes on stdout or stderr through logger. It returns when the command
	// exits, which includes when ctx is done.
	Forward(ctx context.Context, logger *slog.Logger, args []string) error
}

// Start starts forwarding webhook events, one process per organization, and
// returns as soon as they are running. Each process is bound to ctx, so
// cancelling it, e.g., on Ctrl+C, tears them all down with the server.
//
// Start returns an error if the options are inconsistent or the prerequisites
// are missing; the processes themselves report through logger.
func Start(ctx context.Context, logger *slog.Logger, opts Options) error {
	if err := opts.validate(); err != nil {
		return err
	}
	runner := opts.Runner
	if runner == nil {
		runner = ghRunner{}
	}
	if err := runner.Check(ctx); err != nil {
		return err
	}
	events := strings.Join(opts.Events, ",")
	for _, org := range opts.Orgs {
		args := []string{
			"webhook", "forward",
			"--org=" + org,
			"--events=" + events,
			"--url=" + opts.URL,
			"--secret=" + opts.Secret,
		}
		logger := logger.With(label.Organization, org)
		logger.Info("forwarding webhook events", urlLabel, opts.URL, eventsLabel, events)
		go func() {
			// A forwarder killed by the context is the expected shutdown path,
			// not a failure worth reporting.
			if err := runner.Forward(ctx, logger, args); err != nil && ctx.Err() == nil {
				logger.Error("webhook forwarding stopped", label.Error, err)
			}
		}()
	}
	return nil
}

// ghRunner runs the GitHub CLI found on PATH.
type ghRunner struct{}

func (ghRunner) Check(ctx context.Context) error {
	if _, err := exec.LookPath(ghCommand); err != nil {
		return fmt.Errorf("%w; %s", err, setupHint)
	}
	// The extension list also fails if gh is not authenticated at all.
	out, err := exec.CommandContext(ctx, ghCommand, "extension", "list").CombinedOutput()
	if err != nil {
		return fmt.Errorf("listing %s extensions: %w; %s", ghCommand, err, setupHint)
	}
	if !strings.Contains(string(out), extension) {
		return fmt.Errorf("the %s extension is not installed; %s", extension, setupHint)
	}
	return nil
}

func (ghRunner) Forward(ctx context.Context, logger *slog.Logger, args []string) error {
	cmd := exec.CommandContext(ctx, ghCommand, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("connecting to stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("connecting to stderr: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %s: %w", ghCommand, err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go logLines(&wg, logger, "stdout", stdout)
	go logLines(&wg, logger, "stderr", stderr)
	// Both pipes must be drained before Wait closes them.
	wg.Wait()
	return cmd.Wait()
}

// maxLineLength bounds a single line logged from the subprocess.
const maxLineLength = 1 << 20

// logLines logs each line read from r as its own record, so that the events the
// forwarder receives stay as visible as they are in a dedicated terminal.
func logLines(wg *sync.WaitGroup, logger *slog.Logger, stream string, r io.Reader) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(nil, maxLineLength)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			logger.Info(line, streamLabel, stream)
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("reading webhook forwarder output", streamLabel, stream, label.Error, err)
	}
}
