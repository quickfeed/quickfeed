package ci

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	pb "github.com/autograde/quickfeed/ag"
)

// AssignmentInfo holds metadata needed to fetch student code
// and the test repository for an assignment.
type AssignmentInfo struct {
	AssignmentName     string
	Script             string
	CreatorAccessToken string
	GetURL             string
	TestURL            string
	RandomSecret       string
}

func newAssignmentInfo(course *pb.Course, assignment *pb.Assignment, cloneURL, testURL string) *AssignmentInfo {
	script := assignment.GetScriptFile()
	if strings.Count(script, ".") < 1 {
		script = script + ".sh"
	}

	return &AssignmentInfo{
		AssignmentName:     assignment.GetName(),
		Script:             script,
		CreatorAccessToken: course.GetAccessToken(),
		GetURL:             cloneURL,
		TestURL:            testURL,
		RandomSecret:       randomSecret(),
	}
}

// parseScriptTemplate returns a job describing the docker image to use and
// the commands of the job. The job is extracted from a script template file
// provided as input along with assignment metadata for the template.
func parseScriptTemplate(scriptPath string, info *AssignmentInfo) (*Job, error) {
	tmplFile := filepath.Join(scriptPath, info.Script)
	t, err := template.ParseFiles(tmplFile)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	if err := t.Execute(buffer, info); err != nil {
		return nil, err
	}
	s := strings.Split(buffer.String(), "\n")
	if len(s) < 2 {
		return nil, fmt.Errorf("no script template in %s", tmplFile)
	}
	parts := strings.Split(s[0], "#image/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("no docker image specified in script template %s", tmplFile)
	}
	return &Job{Image: parts[1], Commands: s[1:]}, nil
}
