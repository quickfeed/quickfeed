package main

import (
	"bytes"
	"fmt"
	"html/template"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const (
	readmeTmplFile = "readme_tmpl.md"
	readmeFile     = "README.md"
	assignmentFile = "assignment.yml"
)

// LabHeader contains a string representation of the content to be written as a lab header
type LabHeader struct {
	LabHeader string
	ToC       []string
}

func genReadme() {
	if err := loadEnv(); err != nil {
		exitErr(err, "Error loading environment variables")
	}

	labs, err := findLabsWithReadmeTmpl(gitRoot)
	if err != nil {
		exitErr(err, "Error finding labs with readme_tmpl.md files")
	}
	assignments := generateReadme(gitRoot, labs)

	labPlan := mustExecute(parseTemplate("labplan", labPlanTemplate), assignments)
	err = os.WriteFile(filepath.Join(gitRoot, "info", "lab-plan.md"), []byte(labPlan), 0o644)
	if err != nil {
		exitErr(err, "Error writing lab-plan.md")
	}
}

func generateReadme(repo string, labs map[string][]string) map[int]*AssignmentInfo {
	// process each assignment and corresponding readme_tmpl.md
	assignments := make(map[int]*AssignmentInfo)
	for _, lab := range slices.Sorted(maps.Keys(labs)) {
		for _, readme := range labs[lab] {
			// prepare paths for log output and for output file
			assignPath := strings.Replace(filepath.Join(lab, assignmentFile), repo, course(), 1)
			readmePath := strings.Replace(readme, repo, course(), 1)
			readmeMDPath := strings.Replace(readme, readmeTmplFile, readmeFile, 1)
			readmeMDPathShort := strings.Replace(readmeMDPath, repo, course(), 1)
			fmt.Printf("Combining (%v, %v) -> %v\n", assignPath, readmePath, readmeMDPathShort)

			header := ""
			if filepath.Dir(assignPath) == filepath.Dir(readmePath) {
				// Only include header for labX/readme_tmpl.md files, not those in subdirectories
				header = parseAssignmentHeader(lab, headerTemplate, assignments)
			} else {
				// Only include title for readme_tmpl.md files in subdirectories
				header = parseAssignmentHeader(lab, titleOnlyHeaderTemplate, assignments)
			}

			// load readme_tmpl.md and execute template; saving output to README.md
			b, err := os.ReadFile(readme)
			check(err)
			readme = string(b)
			// compute position in readme without the level one title, if exists
			if !strings.HasPrefix(readme, "# ") {
				panic(readmePath + " expected to start with markdown title (single #)")
			}
			posWithoutTitle := strings.Index(readme, "\n")
			if posWithoutTitle > 0 && readme[posWithoutTitle] == '\n' {
				posWithoutTitle++
			}
			// inject the template lines in the beginning of the readme
			readmeTemplate := fmt.Sprintf("%v\n%v", tocTemplate, readme[posWithoutTitle+1:])
			toc := generateToC(readmeTemplate)
			readmeHeader := &LabHeader{
				LabHeader: header,
				ToC:       toc,
			}
			tocMD := mustExecute(parseTemplate("readme", tocTemplate), readmeHeader)
			readmeMD := tocMD + readme[posWithoutTitle+1:]
			err = os.WriteFile(readmeMDPath, []byte(readmeMD), 0o644)
			check(err)
		}
	}
	return assignments
}

// findLabsWithReadmeTmpl returns a map of labs with assignment.yml files
// and a slice of their corresponding readme_tmpl.md files.
func findLabsWithReadmeTmpl(repo string) (map[string][]string, error) {
	labs := make(map[string][]string)

	// find all labs with assignment.yml files
	err := filepath.WalkDir(repo, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		var emptySlice []string
		if !d.IsDir() && d.Name() == assignmentFile {
			dir := filepath.Dir(path)
			labs[dir] = emptySlice
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// find all corresponding readme_tmpl.md files
	err = filepath.WalkDir(repo, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == readmeTmplFile {
			dir := filepath.Dir(path)
			if _, found := labs[dir]; found {
				labs[dir] = append(labs[dir], path)
			} else {
				for level := 4; !found && level > 0; level-- {
					// traverse up the hierarchy looking for existing lab dir
					// with a previously recorded assignment.yml file;
					// stop when level reach 0
					dir := filepath.Dir(dir)
					if _, found = labs[dir]; found {
						labs[dir] = append(labs[dir], path)
					}
				}
				if !found {
					fmt.Printf("ignoring %s: couldn't find %s in or above %v\n", path, assignmentFile, dir)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return labs, nil
}

// parseAssignmentHeader returns a header by parsing assignment.yml.
func parseAssignmentHeader(lab, headerTemplate string, assignments map[int]*AssignmentInfo) string {
	assignment, err := parseAssignment(filepath.Join(lab, assignmentFile))
	check(err)

	// make sure all assignments has CourseOrg field set
	assignment.CourseOrg = course()
	if _, found := assignments[assignment.Order]; !found {
		// add to assignments only once; this is since the assignment.yml
		// may exist for multiple versions of the same assignment README.md.
		assignments[assignment.Order] = assignment
	}
	return mustExecute(parseTemplate("assignment", headerTemplate), assignment)
}

// generateToC takes a markdown file and generates a table of contents
// of all level two headings.
func generateToC(readme string) []string {
	// this reg exp became a bit nasty since we want to match with backtick
	legalHeadingChars := `\w\s\/:#-` + "`"
	headingRegExp := regexp.MustCompile(`^#{2,3}\s([` + legalHeadingChars + `]+)$`)
	headings := make([]string, 0)
	for line := range strings.SplitSeq(readme, "\n") {
		if headingRegExp.MatchString(line) {
			headingText := headingRegExp.ReplaceAllString(line, "$1")
			headings = append(headings, headingText)
		}
	}
	return headings
}

var funcMap = template.FuncMap{
	"link": func(heading string) string {
		replace := map[string]string{
			" ": "-",
			"#": "",
			"`": "",
			":": "",
			"/": "",
		}
		str := strings.ToLower(heading)
		for old, new := range replace {
			str = strings.ReplaceAll(str, old, new)
		}
		return str
	},
	"inc": func(i int) int {
		return i + 1
	},
}

func parseTemplate(name, tmpl string) *template.Template {
	return template.Must(template.New(name).Funcs(funcMap).Parse(tmpl))
}

func mustExecute(t *template.Template, data any) string {
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		panic(err)
	}
	return b.String()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
