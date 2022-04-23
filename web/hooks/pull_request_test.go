package hooks

import "testing"

func TestGetLinkedIssue(t *testing.T) {
	var wantIssueNumber uint64 = 30
	tests := map[string]struct {
		body        string
		expectError bool
	}{
		"Simple":         {body: "Fixes #30", expectError: false},
		"Not a number":   {body: "Fixes #30nan", expectError: true},
		"Invalid body":   {body: "Fixes #30nan #", expectError: true},
		"Invalid body 2": {body: "Fixes", expectError: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotIssueNumber, err := getLinkedIssue(tt.body)
			if err != nil {
				if tt.expectError {
					return
				}
			}
			if gotIssueNumber != wantIssueNumber {
				t.Errorf("getLinkedIssue() = %d, expected %d", gotIssueNumber, wantIssueNumber)
			}
		})
	}
}
