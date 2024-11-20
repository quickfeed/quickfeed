package qf_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetUserSubset(t *testing.T) {
	user1 := &qf.User{ID: 1}
	user2 := &qf.User{ID: 2}
	user3 := &qf.User{ID: 3}
	group := &qf.Group{Users: []*qf.User{user1, user2, user3}}
	wantSubset := []*qf.User{user2, user3}
	gotSubset := group.GetUsersExcept(1)
	if diff := cmp.Diff(wantSubset, gotSubset, protocmp.Transform()); diff != "" {
		t.Errorf("GetUserSubset() mismatch (-wantSubset, +gotSubset):\n%s", diff)
	}
}

func TestGroupContains(t *testing.T) {
	tests := []struct {
		name       string
		groupUsers []*qf.User
		user       *qf.User
		want       bool
	}{
		{
			name:       "User in group",
			groupUsers: []*qf.User{{ID: 1}, {ID: 2}, {ID: 3}},
			user:       &qf.User{ID: 2},
			want:       true,
		},
		{
			name:       "User not in group",
			groupUsers: []*qf.User{{ID: 1}, {ID: 2}, {ID: 3}},
			user:       &qf.User{ID: 4},
			want:       false,
		},
		{
			name:       "Empty group",
			groupUsers: []*qf.User{},
			user:       &qf.User{ID: 1},
			want:       false,
		},
		{
			name:       "Nil user",
			groupUsers: []*qf.User{{ID: 1}, {ID: 2}, {ID: 3}},
			user:       nil,
			want:       false,
		},
		{
			name:       "Nil group and nil user",
			groupUsers: nil,
			user:       nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &qf.Group{
				Users: tt.groupUsers,
			}
			got := group.Contains(tt.user)
			if got != tt.want {
				t.Errorf("Group.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
