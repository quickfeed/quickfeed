package hooks

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

const (
	moreThanOneHash = "more than one '#' character in pull request body"
	noIssueFound    = "no issue found in pull request body"
)

func TestFindIssue(t *testing.T) {
	wantIssue := &qf.Issue{ScmIssueNumber: 30}
	issues := []*qf.Issue{
		{ScmIssueNumber: 10},
		{ScmIssueNumber: 20},
		wantIssue,
	}
	tests := map[string]struct {
		body   string
		errStr string
	}{
		"Fixes":            {body: "Fixes #30", errStr: ""},
		"fixes":            {body: "fixes #30", errStr: ""},
		"Closes":           {body: "Closes #30", errStr: ""},
		"closes":           {body: "closes #30", errStr: ""},
		"Resolves":         {body: "Resolves #30", errStr: ""},
		"resolves":         {body: "resolves #30", errStr: ""},
		"Not a number":     {body: "Fixes #30nan", errStr: noIssueFound},
		"Unexpected issue": {body: "Fixes #40", errStr: "unknown issue #40"},
		"Multiple issues":  {body: "Fixes #30 Fixes #20", errStr: moreThanOneHash},
		"Duplicate issues": {body: "Fixes #30 Fixes #30", errStr: moreThanOneHash},
		"Invalid body":     {body: "Fixes #30nan #", errStr: moreThanOneHash},
		"Invalid body 2":   {body: "Fixes", errStr: noIssueFound},
		"Invalid body 3":   {body: "Fixes #", errStr: noIssueFound},
		"Invalid body 4":   {body: "Fixes #30 task-hello_world", errStr: noIssueFound},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotIssue, err := findIssue(tt.body, issues)
			if tt.errStr != "" && err == nil {
				t.Errorf("findIssue() = %v, expected %q", gotIssue, tt.errStr)
				return
			}
			if err != nil {
				if tt.errStr == err.Error() {
					return
				}
				t.Errorf("findIssue() = %q, expected %q", err, tt.errStr)
				return
			}
			if gotIssue != wantIssue {
				t.Errorf("findIssue() = %v, expected %v", gotIssue, wantIssue)
			}
		})
	}
}
