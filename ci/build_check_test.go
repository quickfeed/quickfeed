package ci

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoBuildCheck runs the injected build check phase against a submitted
// module to check that the phase reports the compilation result of the
// submitted code, and that buildCheckFailed agrees with it.
func TestGoBuildCheck(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("bash not available: %v", err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go not available: %v", err)
	}
	const current = "lab1"
	submitted := t.TempDir()
	assignmentDir := filepath.Join(submitted, current)
	if err := os.Mkdir(assignmentDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// The module has no dependencies, so the build check needs no network.
	write(t, filepath.Join(submitted, "go.mod"), "module example.com/student\n\ngo 1.24\n")

	commands := buildCheckCommands(languageGo)
	runBuildCheck := func(t *testing.T) string {
		t.Helper()
		cmd := exec.Command(bash, "-c", strings.Join(commands, "\n"))
		cmd.Env = append(os.Environ(), "SUBMITTED="+submitted, "CURRENT="+current)
		out, err := cmd.CombinedOutput()
		if err != nil {
			// The phase must not abort the run; see buildCheckCommands.
			t.Fatalf("build check exited with %v; output:\n%s", err, out)
		}
		return string(out)
	}

	write(t, filepath.Join(assignmentDir, "student.go"), "package lab1\n\nfunc Answer() int { return 42 }\n")
	out := runBuildCheck(t)
	if !strings.Contains(out, buildCheckOKMarker) {
		t.Errorf("build check output = %q, want it to contain %q", out, buildCheckOKMarker)
	}
	if buildCheckFailed(out) {
		t.Errorf("buildCheckFailed(%q) = true, want false", out)
	}

	write(t, filepath.Join(assignmentDir, "student.go"), "package lab1\n\nfunc Answer() int { return missing }\n")
	out = runBuildCheck(t)
	if !strings.Contains(out, buildCheckFailedMarker) {
		t.Errorf("build check output = %q, want it to contain %q", out, buildCheckFailedMarker)
	}
	if !strings.Contains(out, "undefined: missing") {
		t.Errorf("build check output = %q, want it to contain the compiler diagnostic", out)
	}
	if !buildCheckFailed(out) {
		t.Errorf("buildCheckFailed(%q) = false, want true", out)
	}

	// A missing assignment folder is not attributable to the submitted code.
	if err := os.RemoveAll(assignmentDir); err != nil {
		t.Fatal(err)
	}
	out = runBuildCheck(t)
	if buildCheckFailed(out) {
		t.Errorf("buildCheckFailed(%q) = true, want false", out)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
