package web_test

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"golang.org/x/oauth2"
)

func TestThirdPartyAuth(t *testing.T) {
	token := scm.GetAccessToken(t)
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := fillDatabase(t, db, token)

	client, _, _ := MockClientWithUser(t, db, connect.WithInterceptors(
		interceptor.NewClientInterceptor(token),
	))
	ctx := context.Background()

	request := connect.NewRequest(&qf.CourseUserRequest{
		CourseCode: "DAT320",
		CourseYear: 2021,
		UserLogin:  user.Login,
	})
	userInfo, err := client.GetUserByCourse(ctx, request)
	check(t, err)
	if userInfo.Msg.ID != user.ID {
		t.Errorf("expected user id %d, got %d", user.ID, userInfo.Msg.ID)
	}
	if userInfo.Msg.Login != user.Login {
		t.Errorf("expected user login %s, got %s", user.Login, userInfo.Msg.Login)
	}
}

func fillDatabase(t *testing.T, db database.Database, token string) *qf.User {
	t.Helper()

	admin := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{
		Code: "DAT320",
		Name: "Operating Systems and Systems Programming",
		Year: 2021,
	}
	qtest.CreateCourse(t, db, admin, course)

	externalUser, err := auth.FetchExternalUser(&oauth2.Token{
		AccessToken: token,
	})
	if err != nil {
		t.Fatalf("Error when fetching user %v", err)
	}
	teacher := qtest.CreateUser(t, db, externalUser.ID, &qf.User{Login: externalUser.Login})
	qtest.EnrollTeacher(t, db, teacher, course)
	return teacher
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
