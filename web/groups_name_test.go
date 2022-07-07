package web

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/types"
	"go.uber.org/zap"
)

func TestBadGroupNames(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	course := &types.Course{
		Name: "Distributed Systems",
		Code: "DAT520",
		Year: 2018,
	}
	qtest.CreateCourse(t, db, admin, course)

	_, scms := qtest.FakeProviderMap(t)
	ags := NewQuickFeedService(zap.NewNop(), db, scms, BaseHookOptions{}, &ci.Local{})
	user1 := qtest.CreateFakeUser(t, db, 2)
	user2 := qtest.CreateFakeUser(t, db, 3)
	// enroll users in course
	qtest.EnrollStudent(t, db, user1, course)
	qtest.EnrollStudent(t, db, user2, course)

	group := &types.Group{
		ID:       1,
		CourseID: course.ID,
		Name:     "DuplicateGroupName",
		Users:    []*types.User{user1, user2},
	}
	// current user1 (in context) must be in group being created
	ctx := qtest.WithUserContext(context.Background(), user1)
	gotGroup, err := ags.CreateGroup(ctx, group)
	if err != nil {
		t.Fatal(err)
	}

	groupNames := []struct {
		name      string
		wantError error
	}{
		{"abcdefghijklmnopqrstuvwxyz", errGroupNameTooLong},
		{"groupNameStillTooLong", errGroupNameTooLong},
		{"groupNameNotTooLong", nil},
		{"a", nil},
		{"a1", nil},
		{"23", nil},
		{"HeinsGroup", nil},
		{"Heins-group", nil},
		{"Heins_group", nil},
		{"Hein's group", errGroupNameInvalid},
		{"a" + string([]byte{0x7f}), errGroupNameInvalid},
		{"a" + string([]byte{0x80}), errGroupNameInvalid},
		{"abc ", errGroupNameInvalid},
		{"æ", errGroupNameInvalid},
		{"ø", errGroupNameInvalid},
		{"å", errGroupNameInvalid},
		{"Æ", errGroupNameInvalid},
		{"Ø", errGroupNameInvalid},
		{"Å", errGroupNameInvalid},
		{gotGroup.GetName(), errGroupNameDuplicate},
	}
	for _, test := range groupNames {
		if err := ags.checkGroupName(course.ID, test.name); err != test.wantError {
			t.Errorf("checkGroupName(%q) = %s, expected %s", test.name, err, test.wantError)
		}
	}
}
