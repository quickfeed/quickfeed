package assignments

import (
	"maps"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetNextReviewer(t *testing.T) {
	tests := []struct {
		desc              string
		users             []*qf.User
		numAssignments    int
		initialAssignment map[uint64]int // optional: prefill count for each user ID
		wantAssignment    map[uint64]int // expected total count after test
		wantReviewer      []*qf.User
	}{
		{
			desc:           "NoUsers/users=nil/assignments=0",
			users:          nil,
			numAssignments: 0,
			wantAssignment: nil,
			wantReviewer:   nil,
		},
		{
			desc:           "NoUsers/users=nil/assignments=1",
			users:          nil,
			numAssignments: 1,
			wantAssignment: map[uint64]int{},
			wantReviewer:   []*qf.User{nil},
		},
		{
			desc:           "NoAssignments/users=3/assignments=0",
			users:          users(1, 2, 3),
			numAssignments: 0,
			wantAssignment: nil,
			wantReviewer:   nil,
		},
		{
			desc:           "EvenDistribution/users=1/assignments=1",
			users:          users(42),
			numAssignments: 1,
			wantAssignment: map[uint64]int{42: 1},
			wantReviewer:   users(42),
		},
		{
			desc:           "EvenDistribution/users=1/assignments=10",
			users:          users(42),
			numAssignments: 10,
			wantAssignment: map[uint64]int{42: 10},
			wantReviewer:   users(42, 42, 42, 42, 42, 42, 42, 42, 42, 42),
		},
		{
			desc:           "EvenDistribution/users=3/assignments=9",
			users:          users(1, 2, 3),
			numAssignments: 9,
			wantAssignment: map[uint64]int{1: 3, 2: 3, 3: 3},
			wantReviewer:   users(1, 2, 3, 1, 2, 3, 1, 2, 3),
		},
		{
			desc:           "EvenDistribution/users=5/assignments=25",
			users:          users(1, 2, 3, 4, 5),
			numAssignments: 25,
			wantAssignment: map[uint64]int{1: 5, 2: 5, 3: 5, 4: 5, 5: 5},
			wantReviewer:   users(1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5),
		},
		{
			desc:              "SkewedInitialDistribution/users=3/assignments=6",
			users:             users(1, 2, 3),
			numAssignments:    6,                                // 6 new assignments
			initialAssignment: map[uint64]int{1: 2, 2: 0, 3: 1}, // 3 initial assignments
			wantAssignment:    map[uint64]int{1: 3, 2: 3, 3: 3}, // 9 total assignments
			wantReviewer:      users(2, 2, 3, 1, 2, 3, 1, 2, 3),
		},
		{
			desc:              "SkewedInitialDistribution/users=5/assignments=10",
			users:             users(1, 2, 3, 4, 5),
			numAssignments:    10,                                           // 10 new assignments
			initialAssignment: map[uint64]int{1: 1, 2: 1, 3: 1, 4: 0, 5: 0}, // 3 initial assignments
			wantAssignment:    map[uint64]int{1: 3, 2: 3, 3: 3, 4: 2, 5: 2}, // 13 total assignments
			wantReviewer:      users(4, 5, 1, 2, 3, 4, 5, 1, 2, 3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			reviewCounter := make(countMap)
			reviewCounter.initialize(1)

			// initialize with prefilled counts if provided
			if tc.initialAssignment != nil {
				maps.Copy(reviewCounter[1], tc.initialAssignment)
			}

			for i := range tc.numAssignments {
				gotReviewer := getNextReviewer(tc.users, reviewCounter[1])
				if diff := cmp.Diff(tc.wantReviewer[i], gotReviewer, protocmp.Transform()); diff != "" {
					t.Errorf("getNextReviewer() mismatch (-wantReviewer, +gotReviewer):\n%s", diff)
				}
			}

			// verify total count
			totalReviews := sum(reviewCounter[1])
			if totalReviews != tc.numAssignments+sum(tc.initialAssignment) {
				t.Errorf("total reviews = %d; want %d", totalReviews, tc.numAssignments+sum(tc.initialAssignment))
			}

			// verify individual counts
			for id, want := range tc.wantAssignment {
				got := reviewCounter[1][id]
				if got != want {
					t.Errorf("user ID %d: got %d assignments; want %d", id, got, want)
				}
			}
		})
	}
}

func users(ids ...uint64) []*qf.User {
	users := make([]*qf.User, len(ids))
	for i, id := range ids {
		users[i] = &qf.User{ID: id}
	}
	return users
}

func sum(m map[uint64]int) (total int) {
	for _, v := range m {
		total += v
	}
	return total
}
