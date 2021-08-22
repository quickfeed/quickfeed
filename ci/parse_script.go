package ci

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qutil"
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
	return &AssignmentInfo{
		AssignmentName:     assignment.GetName(),
		Script:             assignment.ScriptFile,
		CreatorAccessToken: course.GetAccessToken(),
		GetURL:             cloneURL,
		TestURL:            testURL,
		RandomSecret:       qutil.RandomString(),
	}
}

// parseScriptTemplate returns a job describing the docker image to use and
// the commands of the job. The job is extracted from a script template file
// provided as input along with assignment metadata for the template.
func parseScriptTemplate(info *AssignmentInfo) (*Job, error) {
	// info.Script is the saved contents of the script, not the file name
	t, err := template.New("scriptfile").Parse(info.Script)
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
	return &Job{Image: parts[1], Commands: s[1:]}, nil
}
