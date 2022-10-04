package interceptor_test

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestUserVerifier(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	shutdown := web.MockQuickFeedServer(t, logger, db, connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
	))

	client := qtest.QuickFeedClient("")

	adminUser := qtest.CreateFakeUser(t, db, 1)
	student := qtest.CreateFakeUser(t, db, 56)

	adminCookie, err := tm.NewAuthCookie(adminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	studentCookie, err := tm.NewAuthCookie(student.ID)
	if err != nil {
		t.Fatal(err)
	}

	userTest := []struct {
		code     connect.Code
		cookie   string
		wantUser *qf.User
	}{
		{code: connect.CodeUnauthenticated, cookie: "", wantUser: nil},
		{code: connect.CodeUnauthenticated, cookie: "should fail", wantUser: nil},
		{code: 0, cookie: adminCookie.String(), wantUser: adminUser},
		{code: 0, cookie: studentCookie.String(), wantUser: student},
	}

	ctx := context.Background()
	for _, user := range userTest {
		gotUser, err := client.GetUser(ctx, requestWithCookie(&qf.Void{}, user.cookie))
		if err != nil {
			// zero codes won't actually reach this check, but that's okay, since zero is CodeOK
			if gotCode := connect.CodeOf(err); gotCode != user.code {
				t.Errorf("GetUser() = %v, want %v", gotCode, user.code)
			}
		}
		wantUser := user.wantUser
		if wantUser != nil {
			// ignore comparing remote identity
			user.wantUser.RemoteIdentities = nil
		}

		if gotUser == nil {
			if wantUser != nil {
				t.Fatalf("GetUser(): %v, want: %v", gotUser, wantUser)
			}
		} else {
			if diff := cmp.Diff(wantUser, gotUser.Msg, protocmp.Transform()); diff != "" {
				t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
			}
		}
	}
	shutdown(ctx)
}
