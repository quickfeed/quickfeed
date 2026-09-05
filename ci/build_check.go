package ci

import (
	"regexp"
	"strings"
)

// Markers delimiting the output of the build check phase; see buildCheckCommands.
const (
	buildCheckStartMarker  = "*** QuickFeed: verifying that the submitted code compiles ***"
	buildCheckOKMarker     = "*** QuickFeed: the submitted code compiles ***"
	buildCheckFailedMarker = "*** QuickFeed: the submitted code did not compile ***"
)

// buildCheckCommands returns the commands for a student-only compilation phase,
// which parseTestRunnerScript prepends to the course's run script.
//
// The phase compiles the submitted code on its own, before the run script copies
// the course's tests into the submitted module, so that a compiler diagnostic
// printed between the markers can only originate from the submitted code.
// Without such a phase the run's only compiler summary comes from the test
// command, which compiles the course's tests together with the submitted code;
// a broken test file or an unavailable dependency then looks exactly like a
// submission that does not compile, and would be recorded as a zero score.
//
// The phase never aborts the run: a run script may supply files that the
// submitted code needs, such as generated code or a helper package from the
// tests repository. If the tests still run and produce scores, those scores
// decide the outcome; the build check only attributes a run that produced no
// scores at all. See classifyRun.
//
// A language without a compilation phase here has no trustworthy zero score:
// its failed runs are recorded as NO_SCORES, which keeps the previous scores.
func buildCheckCommands(language string) []string {
	switch language {
	case languageGo:
		return []string{goBuildCheck}
	default:
		return nil
	}
}

// goBuildCheck builds every package in the submitted assignment folder,
// discarding the built binaries. The build runs in a subshell so that the
// run script still starts in the container's working directory.
const goBuildCheck = `printf '\n%s\n' '` + buildCheckStartMarker + `'
if (cd "$SUBMITTED/$CURRENT" && go build -o /dev/null ./...) 2>&1; then
    printf '%s\n\n' '` + buildCheckOKMarker + `'
else
    printf '%s\n\n' '` + buildCheckFailedMarker + `'
fi`

// goCompilerDiagnostic matches a Go compiler diagnostic: a source position
// followed by a message, e.g., "./student.go:8:2: undefined: missing".
var goCompilerDiagnostic = regexp.MustCompile(`(?m)^\S+\.go:\d+(:\d+)?: \S`)

// unattributableErrors name build check failures that the submitted code cannot
// be blamed for: the run script may resolve missing dependencies with go get or
// go mod tidy before it runs the tests, and a download failure is a failure of
// the test environment. Matched against the lowercased build check output.
var unattributableErrors = []string{
	"no required module provides package",
	"cannot find module providing package",
	"missing go.sum entry",
	"go.mod file not found",
	"updates to go.mod needed",
	"dial tcp",
	"connection refused",
	"i/o timeout",
	"tls handshake",
	"permission denied",
}

// buildCheckFailed reports whether the build check blamed the submitted code for
// a compilation failure. It reports false when no build check ran, when the
// submitted code compiled, and when the failure is not attributable to the
// submitted code.
func buildCheckFailed(out string) bool {
	_, section, found := strings.Cut(out, buildCheckStartMarker)
	if !found {
		// No build check ran for this job; see buildCheckCommands.
		return false
	}
	section, _, found = strings.Cut(section, buildCheckFailedMarker)
	if !found {
		// The submitted code compiled, or the phase never finished.
		return false
	}
	lower := strings.ToLower(section)
	for _, unattributable := range unattributableErrors {
		if strings.Contains(lower, unattributable) {
			return false
		}
	}
	return goCompilerDiagnostic.MatchString(section)
}
