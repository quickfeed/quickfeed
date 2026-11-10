package ci

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	// dispatcherScript is the generated run script inside the job zip.
	dispatcherScript = ".qf/run.sh"
	// timeoutMessage is what Docker.Run returns together with context.DeadlineExceeded.
	timeoutMessage = "Container timeout. Please check for infinite loops or other slowness."
)

// Dispatcher is a Runner that sends each job to a remote dispatcher, which hands it to
// a worker. The worker extracts the job's files into /quickfeed, and runs its commands.
type Dispatcher struct {
	baseURL string
	token   string
	client  *http.Client
	timeout time.Duration
}

var _ Runner = (*Dispatcher)(nil)

// DispatcherOption configures a Dispatcher.
type DispatcherOption func(*Dispatcher)

// WithHTTPClient sets the HTTP client. It must not have an overall timeout, since a call
// blocks until the job has finished; use the context's deadline instead.
func WithHTTPClient(c *http.Client) DispatcherOption {
	return func(d *Dispatcher) { d.client = c }
}

// WithRunTimeout sets how long the dispatcher lets a script run (its default is 10 minutes).
func WithRunTimeout(t time.Duration) DispatcherOption {
	return func(d *Dispatcher) { d.timeout = t }
}

// NewDispatcher returns a runner that uses the dispatcher at baseURL, authenticated by the RPC token.
func NewDispatcher(baseURL, token string, opts ...DispatcherOption) *Dispatcher {
	d := &Dispatcher{baseURL: strings.TrimRight(baseURL, "/"), token: token, client: &http.Client{}}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Close mirrors Docker.Close; there is nothing to release.
func (*Dispatcher) Close() error { return nil }

// dispatcherRequest is the "details" part of POST /rpc/jobs.
type dispatcherRequest struct {
	ImageType      string `json:"image_type"`
	AssignmentID   string `json:"assignment_id"`
	Secret         string `json:"secret"`
	SubmittedDir   string `json:"submitted_dir"`
	TestsDir       string `json:"tests_dir"`
	AssignmentsDir string `json:"assignments_dir"`
	ScriptPath     string `json:"script_path"`
	RunTimeoutSecs int64  `json:"run_timeout_secs,omitempty"`
	Name           string `json:"name,omitempty"`
}

type dispatcherOutcome struct {
	State  string `json:"state"` // completed | failed
	Reason string `json:"reason"`
	Result *struct {
		Status   string `json:"status"` // completed | timed_out | error
		ExitCode *int64 `json:"exit_code"`
		Output   string `json:"output"`
	} `json:"result"`
}

// Run implements the Runner interface. It blocks until the dispatcher has an outcome for
// the job, or ctx is done, in which case the dispatcher cancels the job.
func (d *Dispatcher) Run(ctx context.Context, job *Job) (string, error) {
	if job.Image == "" {
		return "", fmt.Errorf("no image name specified for '%s'", job.Name)
	}
	if job.BindDir == "" || job.BuildContext[Dockerfile] != "" {
		// Image builds (Dockerfile) and jobs without files are not supported by the dispatcher.
		return "", fmt.Errorf("dispatcher cannot run job '%s': needs BindDir and no Dockerfile build", job.Name)
	}
	env := envMap(job.Env)
	req := dispatcherRequest{
		ImageType:      job.Image,
		AssignmentID:   env["CURRENT"],
		Secret:         env[secretEnvName],
		SubmittedDir:   baseName(env["SUBMITTED"]),
		TestsDir:       baseName(env["TESTS"]),
		AssignmentsDir: baseName(env["ASSIGNMENTS"]),
		ScriptPath:     dispatcherScript,
		RunTimeoutSecs: int64(d.timeout / time.Second),
		Name:           job.Name,
	}
	if req.AssignmentID == "" || req.Secret == "" || env["SUBMITTED"] == "" || env["TESTS"] == "" || env["ASSIGNMENTS"] == "" {
		return "", fmt.Errorf("job '%s': env lacks CURRENT, %s, SUBMITTED, TESTS or ASSIGNMENTS", job.Name, secretEnvName)
	}
	body, contentType, err := jobBody(req, job)
	if err != nil {
		return "", fmt.Errorf("creating job zip: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.baseURL+"/rpc/jobs", body)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Authorization", "Bearer "+d.token)

	resp, err := d.client.Do(httpReq)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return timeoutMessage, ctx.Err()
		}
		return "", fmt.Errorf("dispatcher call failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusConflict:
		return "", ErrConflict
	default:
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("dispatcher returned %s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	var out dispatcherOutcome
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decoding dispatcher response: %w", err)
	}
	return outcomeResult(out)
}

// outcomeResult maps the dispatcher's answer to Run's return values, like Docker.Run does.
func outcomeResult(out dispatcherOutcome) (string, error) {
	switch out.State {
	case "failed":
		if out.Reason == "run_timeout" {
			return timeoutMessage, context.DeadlineExceeded
		}
		// no_worker_available, worker_disconnected: the caller may retry.
		return "", fmt.Errorf("dispatcher job failed: %s", out.Reason)
	case "completed":
		res := out.Result
		if res == nil {
			return "", errors.New("dispatcher response has no result")
		}
		switch res.Status {
		case "timed_out":
			return timeoutMessage, context.DeadlineExceeded
		case "error":
			return res.Output, fmt.Errorf("worker could not run the job: %s", res.Output)
		}
		output := truncateLog(res.Output, maxLogSize, lastSegmentSize, maxToScan)
		if res.ExitCode == nil {
			return output, errors.New("script did not exit normally")
		}
		if *res.ExitCode != 0 {
			return output, &ContainerExitError{Code: *res.ExitCode}
		}
		return output, nil
	}
	return "", fmt.Errorf("unexpected dispatcher state %q", out.State)
}

func envMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			m[k] = v
		}
	}
	return m
}

// jobBody returns the multipart body of the job: its details and the zip.
func jobBody(req dispatcherRequest, job *Job) (io.Reader, string, error) {
	zipBytes, err := jobZip(job)
	if err != nil {
		return nil, "", err
	}
	details, err := json.Marshal(req)
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	parts := []struct {
		name, typ, filename string
		data                []byte
	}{
		{"details", "application/json", "", details},
		{"code", "application/zip", "job.zip", zipBytes},
	}
	for _, part := range parts {
		disposition := fmt.Sprintf("form-data; name=%q", part.name)
		if part.filename != "" {
			disposition += fmt.Sprintf("; filename=%q", part.filename)
		}
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", disposition)
		h.Set("Content-Type", part.typ)
		w, err := mw.CreatePart(h)
		if err != nil {
			return nil, "", err
		}
		if _, err := w.Write(part.data); err != nil {
			return nil, "", err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, "", err
	}
	return &buf, mw.FormDataContentType(), nil
}

// jobZip returns a zip with what the container would see below /quickfeed: the top-level
// entries of job.BindDir, each read-only mount under the base name of its container
// path, and the generated run script. Git directories and symlinks are left out.
func jobZip(job *Job) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	script, err := zw.CreateHeader(&zip.FileHeader{Name: dispatcherScript, Method: zip.Deflate, Modified: time.Now()})
	if err != nil {
		return nil, err
	}
	content := "#!/bin/bash\n" + strings.Join(job.Commands, "\n") + "\n"
	if _, err := io.WriteString(script, content); err != nil {
		return nil, err
	}

	if err := addDir(zw, job.BindDir, ""); err != nil {
		return nil, err
	}
	for _, src := range slices.Sorted(maps.Keys(job.ReadOnlyMounts)) {
		if err := addDir(zw, src, baseName(job.ReadOnlyMounts[src])); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// addDir adds the files below dir to the zip, below prefix ("" for the zip's root).
func addDir(zw *zip.Writer, dir, prefix string) error {
	return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		name := path.Join(prefix, filepath.ToSlash(rel))
		switch {
		case d.IsDir() && d.Name() == ".git":
			return filepath.SkipDir
		case d.IsDir():
			if name != "." {
				_, err = zw.Create(name + "/")
			}
			return err
		case !d.Type().IsRegular(): // symlinks and special files
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name, hdr.Method = name, zip.Deflate
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
}

// baseName returns the last element of a container path. EnvVars and RunTests build these
// paths with filepath.Join, so on a Windows host they hold backslashes.
func baseName(p string) string {
	return path.Base(strings.ReplaceAll(p, `\`, "/"))
}
