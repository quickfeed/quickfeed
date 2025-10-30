package web_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"golang.org/x/oauth2"
)

func TestThirdPartyAppAuth(t *testing.T) {
	token := scm.GetAccessToken(t)
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := fillDatabase(t, db, token)

	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithClientOptions(connect.WithInterceptors(
		interceptor.NewTokenAuthClientInterceptor(token),
	)))
	ctx := context.Background()

	userInfo, err := client.GetUser(ctx, connect.NewRequest(&qf.Void{}))
	check(t, err)
	if userInfo.Msg.GetID() != user.GetID() {
		t.Errorf("expected user id %d, got %d", user.GetID(), userInfo.Msg.GetID())
	}
	if userInfo.Msg.GetLogin() != user.GetLogin() {
		t.Errorf("expected user login %s, got %s", user.GetLogin(), userInfo.Msg.GetLogin())
	}
}

func fillDatabase(t *testing.T, db database.Database, token string) *qf.User {
	t.Helper()

	admin := qtest.CreateFakeUser(t, db)
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
	teacher := &qf.User{
		Login:       externalUser.Login,
		ScmRemoteID: externalUser.ID,
	}
	if err := db.CreateUser(teacher); err != nil {
		t.Error(err)
	}
	qtest.EnrollTeacher(t, db, teacher, course)
	return teacher
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
