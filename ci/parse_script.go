package ci

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// AssignmentInfo holds metadata needed to fetch student code
// and the test repository for an assignment.
type AssignmentInfo struct {
	AssignmentName     string
	CreatorAccessToken string
	GetURL             string
	TestURL            string
	RandomSecret       string
}

// parseScriptTemplate returns a job specifying the docker image and commands
// to be executed by the docker image. The job's commands are extracted from
// the script template file (run.sh) associated with the RunData's assignment.
// The script template may use the variables of the AssignmentInfo struct, e.g.,
// {{ .AssignmentName }}, {{ .RandomSecret }}, etc.
func (r RunData) parseScriptTemplate(secret string) (*Job, error) {
	info := &AssignmentInfo{
		AssignmentName:     r.Assignment.GetName(),
		CreatorAccessToken: r.Course.GetAccessToken(),
		GetURL:             r.Repo.GetHTMLURL(),
		TestURL:            r.Repo.GetTestURL(),
		RandomSecret:       secret,
	}
	// TODO(meling) rename ScriptFile field to ScriptTemplate
	// ScriptTemplate contains the script itself with variables in double curly braces
	t, err := template.New("script_template").Parse(r.Assignment.ScriptFile)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	if err := t.Execute(buffer, info); err != nil {
		return nil, err
	}
	s := strings.Split(buffer.String(), "\n")
	if len(s) < 2 {
		return nil, fmt.Errorf("no script template for assignment %s in %s", info.AssignmentName, info.TestURL)
	}
	parts := strings.Split(s[0], "#image/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("no docker image specified in script template for assignment %s in %s", info.AssignmentName, info.TestURL)
	}
	return &Job{Name: r.String(info.RandomSecret[:6]), Image: parts[1], Commands: s[1:]}, nil
}
