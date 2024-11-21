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

func TestGroup_UserNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		group *qf.Group
		want  []string
	}{
		{
			name:  "Empty group",
			group: &qf.Group{},
			want:  nil,
		},
		{
			name: "Non empty group",
			group: &qf.Group{
				Users: []*qf.User{
					{Login: "adityaa"},
					{Login: "tootsy-tiger"},
					{Login: "rhea"},
				},
			},
			want: []string{"adityaa", "tootsy-tiger", "rhea"},
		},
		{
			name:  "Nil group",
			group: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.group.UserNames()
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Group.Usernames() mismatch (-wantSubset, +gotSubset):\n%s", diff)
			}
		})
	}
}

func TestGroup_Contains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		group *qf.Group
		user  *qf.User
		want  bool
	}{
		{
			name:  "Empty group",
			group: &qf.Group{},
			user:  &qf.User{ID: 1},
			want:  false,
		},
		{
			name: "User not in group",
			group: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			user: &qf.User{ID: 3},
			want: false,
		},
		{
			name: "User in group",
			group: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			user: &qf.User{ID: 2},
			want: true,
		},
		{
			name: "Nil user",
			group: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
			},
			user: nil,
			want: false,
		},
		{
			name:  "Nil group and nil user",
			group: nil,
			user:  nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.group.Contains(tt.user)
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Group.Contains() mismatch (-wantSubset, +gotSubset):\n%s", diff)
			}
		})
	}
}

func TestGroup_ContainsAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		g1   *qf.Group
		g2   *qf.Group
		want bool
	}{
		{
			name: "Two empty groups",
			g1:   &qf.Group{},
			g2:   &qf.Group{},
			want: true,
		},
		{
			name: "Identical groups",
			g1: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			g2: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			want: true,
		},
		{
			name: "Different order same users",
			g1: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			g2: &qf.Group{
				Users: []*qf.User{
					{ID: 2},
					{ID: 1},
				},
			},
			want: true,
		},
		{
			name: "Different users",
			g1: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			g2: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 3},
				},
			},
			want: false,
		},
		{
			name: "Different number of users",
			g1: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			g2: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
				},
			},
			want: false,
		},
		{
			name: "Nil argument group",
			g1: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
			},
			g2:   nil,
			want: false,
		},
		{
			name: "Nil receiver group",
			g1: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
			},
			g2:   nil,
			want: false,
		},
		{
			name: "Nil all group",
			g1:   nil,
			g2:   nil,
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.g1.ContainsAll(tt.g2)
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Group.ContainsAll() mismatch (-wantSubset, +gotSubset):\n%s", diff)
			}
		})
	}
}

func TestGroup_GetUsersExcept(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		group  *qf.Group
		userID uint64
		want   []*qf.User
	}{
		{
			name:   "Empty group",
			group:  &qf.Group{},
			userID: 1,
			want:   []*qf.User{},
		},
		{
			name: "User ID not present",
			group: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			userID: 3,
			want: []*qf.User{
				{ID: 1},
				{ID: 2},
			},
		},
		{
			name: "User ID present",
			group: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
				},
			},
			userID: 2,
			want: []*qf.User{
				{ID: 1},
			},
		},
		{
			name:   "Nil group",
			group:  nil,
			userID: 3,
			want:   []*qf.User{},
		},
		{
			name: "Nil user list in group",
			group: &qf.Group{
				Users: nil,
			},
			userID: 3,
			want:   []*qf.User{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.group.GetUsersExcept(tt.userID)
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Group.GetUsersExcept() mismatch (-wantSubset, +gotSubset):\n%s", diff)
			}
		})
	}
}

func TestGroup_UserIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		group *qf.Group
		want  []uint64
	}{
		{
			name:  "Empty group",
			group: &qf.Group{},
			want:  []uint64{},
		},
		{
			name: "Non empty group",
			group: &qf.Group{
				Users: []*qf.User{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
			},
			want: []uint64{1, 2, 3},
		},
		{
			name:  "Nil group",
			group: nil,
			want:  []uint64{},
		},
		{
			name: "Nil user list in group",
			group: &qf.Group{
				Users: nil,
			},
			want: []uint64{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.group.UserIDs()
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Group.UserIDs() mismatch (-wantSubset, +gotSubset):\n%s", diff)
			}
		})
	}
}
