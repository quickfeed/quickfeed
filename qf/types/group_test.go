package types_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf/types"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetUserSubset(t *testing.T) {
	user1 := &types.User{ID: 1}
	user2 := &types.User{ID: 2}
	user3 := &types.User{ID: 3}
	group := &types.Group{Users: []*types.User{user1, user2, user3}}
	wantSubset := []*types.User{user2, user3}
	gotSubset := group.GetUsersExcept(1)
	if diff := cmp.Diff(wantSubset, gotSubset, protocmp.Transform()); diff != "" {
		t.Errorf("GetUserSubset() mismatch (-wantSubset, +gotSubset):\n%s", diff)
	}
}
