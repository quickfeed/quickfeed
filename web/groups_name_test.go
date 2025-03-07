package web_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestBadGroupNames(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{
		Name: "Distributed Systems",
		Code: "DAT520",
		Year: 2018,
	}
	qtest.CreateCourse(t, db, admin, course)

	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	groupNames := []struct {
		name      string
		wantError *connect.Error
	}{
		{"abcdefghijklmnopqrstuvwxyz", web.ErrGroupNameTooLong},
		{"groupNameStillTooLong", web.ErrGroupNameTooLong},
		{"groupNameNotTooLong", nil},
		{"a", nil},
		{"a1", nil},
		{"23", nil},
		{"HeinsGroup", nil},
		{"Heins-group", nil},
		{"Heins_group", nil},
		{"Hein's group", web.ErrGroupNameInvalid},
		{"a" + string([]byte{0x7f}), web.ErrGroupNameInvalid},
		{"abc ", web.ErrGroupNameInvalid},
		{"æ", web.ErrGroupNameInvalid},
		{"ø", web.ErrGroupNameInvalid},
		{"å", web.ErrGroupNameInvalid},
		{"Æ", web.ErrGroupNameInvalid},
		{"Ø", web.ErrGroupNameInvalid},
		{"Å", web.ErrGroupNameInvalid},
		{"DuplicateGroupName", nil},
		{"DuplicateGroupName", web.ErrGroupNameDuplicate},
		{"duplicateGroupName", web.ErrGroupNameDuplicate},
		{"duplicateGroupname", web.ErrGroupNameDuplicate},
	}
	for _, tt := range groupNames {
		t.Run(tt.name, func(t *testing.T) {
			user1 := qtest.CreateFakeUser(t, db)
			user2 := qtest.CreateFakeUser(t, db)
			qtest.EnrollStudent(t, db, user1, course)
			qtest.EnrollStudent(t, db, user2, course)

			group := &qf.Group{
				CourseID: course.GetID(),
				Name:     tt.name,
				Users:    []*qf.User{user1, user2},
			}
			_, err := client.CreateGroup(context.Background(), connect.NewRequest(group))
			if tt.wantError == nil {
				if err != nil {
					t.Errorf("got error %v, want nil", err)
				}
				return
			}
			if err == nil && tt.wantError != nil {
				t.Errorf("got no error, want %v", tt.wantError)
			}
			if connErr, ok := err.(*connect.Error); ok {
				if connErr.Code() != tt.wantError.Code() {
					t.Errorf("got error code %v, want %v", connErr.Code(), tt.wantError.Code())
				}
				if connErr.Error() != tt.wantError.Error() {
					t.Errorf("got error %v, want %v", connErr, tt.wantError)
				}
			}
		})
	}
}
