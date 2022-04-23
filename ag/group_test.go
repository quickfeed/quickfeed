package ag_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetUserSubset(t *testing.T) {
	user1 := &pb.User{ID: 1}
	user2 := &pb.User{ID: 2}
	user3 := &pb.User{ID: 3}
	group := &pb.Group{Users: []*pb.User{user1, user2, user3}}
	wantSubset := []*pb.User{user2, user3}
	gotSubset := group.GetUserSubset(1)
	if diff := cmp.Diff(wantSubset, gotSubset, protocmp.Transform()); diff != "" {
		t.Errorf("GetUserSubset() mismatch (-wantSubset, +gotSubset):\n%s", diff)
	}
}
