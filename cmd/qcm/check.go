package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

const checkSynopsis = `Usage: qcm check -course ORG [flags]

Check that the course's test environment is working, using the local clones of
the tests and assignments repositories in <dir>/<org>. The checks are:

	run scripts  every run.sh parses, and every assignment has one
	dockerfile   the course's Dockerfile builds, if the course has one
	content      the two repositories are aligned and their json files parse
	skeleton     the handout code scores at most -max-skeleton-score
	solution     the solution code scores 100% (only with -solution)

The skeleton check is the important one: if the handout code students start
from already scores well, the tests are not exercising what the students are
asked to write. The check exits non-zero if any of the checks fail.`

// Result values of a single check.
const (
	pass = "PASS"
	fail = "FAIL"
	skip = "SKIP"
)

const (
	runScriptFile = "run.sh"
	scriptsDir    = "scripts"
	solutionOwner = "solution"
)

// checkResult is one row of the summary table that qcm check prints. The
// details hold one item per entry: the first is printed on the check's own
// row, and the rest on continuation rows indented to the details column, so
// that a list of tests or of repository issues reads one per line.
type checkResult struct {
	name    string
	result  string
	details []string
}

// checker collects the results of the checks and reports progress as they
// complete. The Docker-backed checks take minutes each, and the summary table
// is only printed once all of them are done, so without progress lines the
// command would appear to hang.
type checker struct {
	progress *progressWriter
	results  []checkResult
}

// newChecker returns a checker reporting progress to w, which must also be the
// writer the command's logger uses; see progressWriter.
func newChecker(w io.Writer) *checker {
	return &checker{progress: &progressWriter{w: w}}
}

// start announces a check that is about to do slow work; its result completes
// the line.
func (ck *checker) start(name, what string) {
	ck.progress.begin(name, what)
}

// add records a finished check and reports its result at once; the details
// follow in the summary table.
func (ck *checker) add(res checkResult) {
	ck.results = append(ck.results, res)
	ck.progress.finish(res.name, res.result)
}

// progressWriter writes the progress lines of qcm check to stderr, which the
// command's logger writes to as well. The line announcing a slow check is
// left open, without its newline, so that the check's result can complete
// it:
//
//	skeleton lab1: running the tests against the skeleton code... PASS
//
// A log record written meanwhile would otherwise be glued onto that line, so
// the writer first ends the open line, and the result is then reported on a
// line of its own. The command runs its checks one at a time, so there is no
// concurrent use to guard against.
type progressWriter struct {
	w    io.Writer
	open string // the check whose progress line awaits its result, if any
}

// Write passes log output through, ending an open progress line first.
func (p *progressWriter) Write(b []byte) (int, error) {
	if p.open != "" {
		p.open = ""
		if _, err := io.WriteString(p.w, "\n"); err != nil {
			return 0, err
		}
	}
	return p.w.Write(b)
}

// begin opens the progress line of the named check.
func (p *progressWriter) begin(name, what string) {
	if p.open != "" {
		fmt.Fprintln(p.w)
	}
	fmt.Fprintf(p.w, "%s: %s...", name, what)
	p.open = name
}

// finish reports the named check's result, on its open progress line if it
// has one and otherwise on a line of its own.
func (p *progressWriter) finish(name, result string) {
	if p.open == name {
		p.open = ""
		fmt.Fprintf(p.w, " %s\n", result)
		return
	}
	if p.open != "" {
		p.open = ""
		fmt.Fprintln(p.w)
	}
	fmt.Fprintf(p.w, "%s: %s\n", name, result)
}

func checkCmd(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("check", stderr, checkSynopsis)
	var c commonFlags
	c.register(fs)
	solution := fs.String("solution", "", "local `directory` holding the course's solution code")
	lab := fs.String("lab", "", "check only this assignment `name`; default is every assignment")
	maxSkeleton := fs.Uint("max-skeleton-score", uint(ci.MaxSkeletonScore), "highest total `score` the skeleton code may reach")
	build := fs.Bool("build", true, "build the course's Docker image from the tests repository")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := c.validate(); err != nil {
		return err
	}

	course := c.course()
	parsed, buildContext, issues, err := assignments.ReadTestsRepository(c.testsDir(), 0)
	if err != nil {
		return fmt.Errorf("reading tests repository %s: %w", c.testsDir(), err)
	}
	if *lab != "" {
		// A misspelled assignment name would otherwise only skip the test runs,
		// and the command would report success without having checked anything.
		if _, err := findAssignment(parsed, *lab); err != nil {
			return err
		}
	}
	course.UpdateDockerfile(buildContext[ci.Dockerfile])
	scripts, err := parseRunScripts(c.testsDir(), parsed)
	if err != nil {
		return err
	}

	// The logger shares stderr with the progress lines; see progressWriter.
	ck := newChecker(stderr)
	ctx := qlog.NewContext(context.Background(), c.logger(ck.progress))
	runner, err := newDockerRunner(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = runner.Close() }()

	// The static checks come first, so that their results are known before the
	// slow Docker-backed checks start.
	ck.add(checkRunScripts(scripts))
	ck.add(checkContent(&c, parsed, issues))
	ck.add(checkDockerfile(ctx, runner, course, buildContext, scripts, *build, ck))
	checkSkeleton(ctx, runner, course, parsed, *lab, uint32(*maxSkeleton), ck)
	if *solution != "" {
		checkSolution(ctx, runner, course, parsed, *lab, *solution, ck)
	}
	return printChecks(stdout, ck.results)
}

// runScripts records what the tests repository's run scripts declare.
type runScripts struct {
	images     map[string]string // repository-relative script path -> Docker image
	found      map[string]bool   // repository-relative paths of the scripts that exist
	failures   []string          // scripts that could not be parsed
	missing    []string          // assignments without a run script of their own
	hasDefault bool              // whether the course has a scripts/run.sh
}

// parseRunScripts parses the course's own run script and the run script of
// every top-level folder in the tests repository, and records which of the
// parsed assignments have no run script of their own. Folders are scanned
// rather than only the parsed assignments, so that a broken run script is
// reported even for a folder whose assignment.json is also broken.
func parseRunScripts(testsDir string, parsed []*qf.Assignment) (*runScripts, error) {
	scripts := &runScripts{images: make(map[string]string), found: make(map[string]bool)}
	hasDefault, err := scripts.parse(testsDir, filepath.Join(scriptsDir, runScriptFile))
	if err != nil {
		return nil, err
	}
	scripts.hasDefault = hasDefault

	entries, err := os.ReadDir(testsDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", testsDir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == scriptsDir || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if _, err := scripts.parse(testsDir, filepath.Join(entry.Name(), runScriptFile)); err != nil {
			return nil, err
		}
	}
	for _, assignment := range parsed {
		if !scripts.found[assignment.GetName()+"/"+runScriptFile] {
			scripts.missing = append(scripts.missing, assignment.GetName())
		}
	}
	return scripts, nil
}

// parse reads and parses the run script at the repository-relative path rel,
// recording either its image or the reason it could not be parsed. It reports
// whether the script exists.
func (s *runScripts) parse(testsDir, rel string) (bool, error) {
	content, err := os.ReadFile(filepath.Join(testsDir, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", rel, err)
	}
	s.found[filepath.ToSlash(rel)] = true
	image, _, _, err := ci.ParseRunScript(string(content))
	if err != nil {
		s.failures = append(s.failures, fmt.Sprintf("%s: %v", filepath.ToSlash(rel), err))
		return true, nil
	}
	s.images[filepath.ToSlash(rel)] = image
	return true, nil
}

// usersOf returns the run scripts that name the given Docker image.
func (s *runScripts) usersOf(image string) []string {
	var users []string
	for path, declared := range s.images {
		if declared == image {
			users = append(users, path)
		}
	}
	slices.Sort(users)
	return users
}

// checkRunScripts reports whether every run script parses and whether every
// assignment is covered by a run script.
func checkRunScripts(scripts *runScripts) checkResult {
	problems := slices.Clone(scripts.failures)
	if !scripts.hasDefault && len(scripts.missing) > 0 {
		problems = append(problems, fmt.Sprintf("no %s/%s, and no %s for %s",
			scriptsDir, runScriptFile, runScriptFile, strings.Join(scripts.missing, ", ")))
	}
	if len(problems) > 0 {
		return checkResult{name: "run scripts", result: fail, details: problems}
	}
	return checkResult{
		name:    "run scripts",
		result:  pass,
		details: []string{fmt.Sprintf("parsed %s", strings.Join(slices.Sorted(maps.Keys(scripts.images)), ", "))},
	}
}

// checkDockerfile builds the course's own Docker image, if it has one. A course
// whose run scripts name only prebuilt images needs no Dockerfile, but an empty
// one is a mistake; see courseDockerfile. The build is announced through ck,
// since it can take minutes.
func checkDockerfile(ctx context.Context, runner ci.Runner, course *qf.Course, buildContext map[string]string, scripts *runScripts, build bool, ck *checker) checkResult {
	const name = "dockerfile"
	dockerfile, err := courseDockerfile(buildContext)
	if err != nil {
		return checkResult{name: name, result: fail, details: []string{err.Error()}}
	}
	if dockerfile == "" {
		if users := scripts.usersOf(course.DockerImage()); len(users) > 0 {
			return checkResult{name: name, result: fail, details: []string{fmt.Sprintf(
				"%s name the course image %q, but the tests repository has no %s/%s to build it",
				strings.Join(users, " and "), course.DockerImage(), scriptsDir, ci.Dockerfile)}}
		}
		return checkResult{name: name, result: skip, details: []string{fmt.Sprintf(
			"no %s/%s; the run scripts name prebuilt images", scriptsDir, ci.Dockerfile)}}
	}
	if !build {
		return checkResult{name: name, result: skip, details: []string{"not built because -build=false"}}
	}
	ck.start(name, "building the "+course.DockerImage()+" image from "+scriptsDir+"/"+ci.Dockerfile)
	if err := assignments.BuildDockerImage(ctx, runner, course, buildContext); err != nil {
		return checkResult{name: name, result: fail, details: []string{err.Error()}}
	}
	return checkResult{name: name, result: pass, details: []string{"built the " + course.DockerImage() + " image"}}
}

// checkContent reports the content problems in the tests repository and the
// alignment problems between the two course repositories.
func checkContent(c *commonFlags, parsed []*qf.Assignment, issues []assignments.RepoIssue) checkResult {
	const name = "content"
	problems := make([]string, 0, len(issues))
	for _, issue := range issues {
		problems = append(problems, issue.String())
	}
	note := ""
	if isDir(c.assignmentsDir()) {
		alignment, err := assignments.CourseRepositoryIssues(c.testsDir(), c.assignmentsDir(), parsed)
		if err != nil {
			problems = append(problems, err.Error())
		}
		for _, issue := range alignment {
			problems = append(problems, issue.String())
		}
	} else {
		note = fmt.Sprintf("no %s repository in %s; skipped the cross-repository comparison",
			qf.AssignmentsRepo, c.course().CloneDir())
	}
	result, details := pass, []string{fmt.Sprintf("%d assignments, no issues", len(parsed))}
	if len(problems) > 0 {
		result, details = fail, problems
	}
	if note != "" {
		details = append(details, note)
	}
	return checkResult{name: name, result: result, details: details}
}

// checkSkeleton runs the course's tests against the handout code in the
// assignments repository, once per auto-graded assignment, recording a result
// per assignment through ck. Tests that actually exercise what the students
// are asked to write barely score on the skeleton.
func checkSkeleton(ctx context.Context, runner ci.Runner, course *qf.Course, parsed []*qf.Assignment, lab string, maxScore uint32, ck *checker) {
	selected := autoGraded(parsed, lab)
	if len(selected) == 0 {
		ck.add(checkResult{name: "skeleton", result: skip, details: []string{nothingToRun(lab)}})
		return
	}
	for _, assignment := range selected {
		name := "skeleton " + assignment.GetName()
		ck.start(name, "running the tests against the skeleton code")
		res, err := runTests(ctx, ci.NewSkeletonRun(course, assignment, localCommitID), runner)
		if err != nil {
			ck.add(checkResult{name: name, result: fail, details: []string{err.Error()}})
			continue
		}
		report := ci.CheckSkeleton(res, maxScore)
		if !report.Healthy() {
			ck.add(checkResult{name: name, result: fail, details: withPassingTests(report.Problem, report.PassingTests)})
			continue
		}
		summary := fmt.Sprintf("scored %d%%, at most %d%% allowed", res.Sum(), maxScore)
		ck.add(checkResult{name: name, result: pass, details: withPassingTests(summary, report.PassingTests)})
	}
}

// withPassingTests returns the summary followed by the tests that score on the
// skeleton code, one per line, so that the teaching staff can see which tests
// to look at. Also for a passing check: the tests that award the skeleton its
// few percent are worth knowing about.
func withPassingTests(summary string, passing []string) []string {
	details := []string{summary}
	if len(passing) > 0 {
		details = append(details, "tests passing on the skeleton code:")
		details = append(details, passing...)
	}
	return details
}

// checkSolution runs the course's tests against the solution code in the given
// directory, once per auto-graded assignment, recording a result per assignment
// through ck. The solution must score 100%; a lower score means the tests
// cannot be passed as written.
func checkSolution(ctx context.Context, runner ci.Runner, course *qf.Course, parsed []*qf.Assignment, lab, solutionDir string, ck *checker) {
	selected := autoGraded(parsed, lab)
	if len(selected) == 0 {
		ck.add(checkResult{name: "solution", result: skip, details: []string{nothingToRun(lab)}})
		return
	}
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: course.GetScmOrganizationName()}
	for _, assignment := range selected {
		name := "solution " + assignment.GetName()
		ck.start(name, "running the tests against the solution code")
		res, err := runTests(ctx, &ci.RunData{
			Course:        course,
			Assignment:    assignment,
			SubmissionDir: solutionDir,
			Repo: &qf.Repository{
				HTMLURL:  repo.StudentRepoURL(solutionOwner),
				RepoType: qf.Repository_USER,
			},
			JobOwner: solutionOwner,
			CommitID: localCommitID,
		}, runner)
		if err != nil {
			ck.add(checkResult{name: name, result: fail, details: []string{err.Error()}})
			continue
		}
		ck.add(solutionResult(name, res))
	}
}

// solutionResult judges a solution run: the solution code must score 100%.
func solutionResult(name string, results *score.Results) checkResult {
	if results.Failed() {
		return checkResult{name: name, result: fail, details: []string{fmt.Sprintf(
			"the solution run failed with status %s", results.GetBuildInfo().GetStatus())}}
	}
	sum := results.Sum()
	if sum == 100 {
		return checkResult{name: name, result: pass, details: []string{"scored 100%"}}
	}
	details := []string{fmt.Sprintf("solution code scored %d%%, expected 100%%", sum)}
	var failing []string
	for _, sc := range results.Scores {
		if sc.GetScore() < sc.GetMaxScore() {
			failing = append(failing, fmt.Sprintf("%s (%d/%d)", sc.GetTestName(), sc.GetScore(), sc.GetMaxScore()))
		}
	}
	if len(failing) > 0 {
		details = append(details, "tests below their max score:")
		details = append(details, failing...)
	}
	return checkResult{name: name, result: fail, details: details}
}

// runTests runs one test job with its own timeout, taken from the assignment.
// The local runs started by qcm check never need an SCM client.
func runTests(ctx context.Context, runData *ci.RunData, runner ci.Runner) (*score.Results, error) {
	ctx, cancel := runData.Assignment.WithTimeout(ctx, ci.DefaultContainerTimeout)
	defer cancel()
	var noSCM scm.SCM
	return runData.RunTests(ctx, noSCM, runner)
}

// autoGraded returns the assignments to run tests for: the auto-graded ones, or
// only the named assignment when lab is given.
func autoGraded(parsed []*qf.Assignment, lab string) []*qf.Assignment {
	var selected []*qf.Assignment
	for _, assignment := range parsed {
		if assignment.GradedManually() || (lab != "" && assignment.GetName() != lab) {
			continue
		}
		selected = append(selected, assignment)
	}
	return selected
}

func nothingToRun(lab string) string {
	if lab != "" {
		return fmt.Sprintf("assignment %q is not an auto-graded assignment of this course", lab)
	}
	return "the course has no auto-graded assignments"
}

// printChecks prints the summary table and returns an error if any check failed.
// A check's first detail shares its row; the remaining details follow on rows
// of their own, indented to the details column.
func printChecks(w io.Writer, results []checkResult) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "CHECK\tRESULT\tDETAILS")
	failed := 0
	for _, res := range results {
		if res.result == fail {
			failed++
		}
		lines := detailLines(res.details)
		first := ""
		if len(lines) > 0 {
			first, lines = lines[0], lines[1:]
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", res.name, res.result, first)
		for _, line := range lines {
			fmt.Fprintf(tw, "\t\t%s\n", line)
		}
	}
	tw.Flush()
	if failed > 0 {
		return fmt.Errorf("%d of %d checks failed", failed, len(results))
	}
	return nil
}

// detailLines splits details that span several lines, such as a Docker build
// failure, into one entry per non-blank line, so that every line of the table
// is a line of text.
func detailLines(details []string) []string {
	var lines []string
	for _, detail := range details {
		for line := range strings.Lines(detail) {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				lines = append(lines, trimmed)
			}
		}
	}
	return lines
}

// isDir reports whether path exists and is a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
