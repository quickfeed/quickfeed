package web

import (
	"sort"
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

func TestOrderSubmissions(t *testing.T) {
	assignments := []*qf.Assignment{
		{
			ID:    1,
			Order: 2,
		},
		{
			ID:    2,
			Order: 3,
		},
		{
			ID:    3,
			Order: 1,
		},
	}

	// Submissions in unsorted order
	// We want to sort them by assignment order
	submissions := []*qf.Submission{
		{
			ID:           1,
			AssignmentID: 1,
		},
		{
			ID:           2,
			AssignmentID: 2,
		},
		{
			ID:           3,
			AssignmentID: 3,
		},
	}

	// Create a map of assignment ID to order
	orderMap := newOrderMap(assignments)

	// Sort the submissions by assignment order
	sort.Slice(submissions, func(i, j int) bool {
		return orderMap.Less(submissions[i].AssignmentID, submissions[j].AssignmentID)
	})

	// Check that the submissions are sorted correctly
	if submissions[0].ID != 3 || submissions[1].ID != 1 || submissions[2].ID != 2 {
		t.Error("Submissions not sorted correctly")
	}
}
