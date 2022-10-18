package web_test

import (
	"context"
	"os"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

const (
	userName = "meling"
)

var user *qf.User

// TODO(meling): Fix this test when support for third-party applications is added
func TestThirdPartyAuth(t *testing.T) {
	if os.Getenv("HELPBOT_TEST") == "" {
		t.Skip("Needs update for helpbot compatibility")
	}
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)

	fillDatabase(t, db)
	if user.Login != userName {
		t.Errorf("Expected %v, got %v\n", userName, user.Login)
	}

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	shutdown, client := MockQuickFeedClient(t, db, connect.WithInterceptors(
		interceptor.NewMetricsInterceptor(),
		interceptor.NewValidationInterceptor(logger),
		interceptor.NewUserInterceptor(logger, tm),
		interceptor.NewAccessControlInterceptor(tm),
		interceptor.NewTokenInterceptor(tm),
	))

	request := connect.NewRequest(&qf.CourseUserRequest{
		CourseCode: "DAT320",
		CourseYear: 2021,
		UserLogin:  userName,
	})
	ctx := context.Background()
	// request.Header().Set(auth.Cookie, firstAdminCookie.String())
	userInfo, err := client.GetUserByCourse(ctx, request)
	check(t, err)
	if userInfo.Msg.ID != user.ID {
		t.Errorf("expected user id %d, got %d", user.ID, userInfo.Msg.ID)
	}
	if userInfo.Msg.Login != user.Login {
		t.Errorf("expected user login %s, got %s", user.Login, userInfo.Msg.Login)
	}
	shutdown(ctx)
}

func fillDatabase(t *testing.T, db database.Database) {
	t.Helper()
	// Add secret token for the helpbot application (to allow it to invoke gRPC methods)
	admin := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{
		Code: "DAT320",
		Name: "Operating Systems and Systems Programming",
		Year: 2021,
	}
	qtest.CreateCourse(t, db, admin, course)

	user = qtest.CreateUser(t, db, 11, &qf.User{Login: userName})
	qtest.EnrollStudent(t, db, user, course)
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
