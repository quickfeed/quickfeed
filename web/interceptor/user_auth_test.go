package interceptor_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
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
	client := web.MockClient(t, db, scm.WithMockOrgs(), connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
	))
	ctx := context.Background()

	adminUser := qtest.CreateFakeUser(t, db)
	student := qtest.CreateFakeUser(t, db)

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

	for _, user := range userTest {
		gotUser, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, user.cookie))
		if err != nil {
			// zero codes won't actually reach this check, but that's okay, since zero is CodeOK
			if gotCode := connect.CodeOf(err); gotCode != user.code {
				t.Errorf("GetUser() = %v, want %v", gotCode, user.code)
			}
		}
		wantUser := user.wantUser
		if gotUser == nil {
			if wantUser != nil {
				t.Errorf("GetUser(): %v, want: %v", gotUser, wantUser)
			}
		} else {
			if diff := cmp.Diff(wantUser, gotUser.Msg, protocmp.Transform()); diff != "" {
				t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
			}
		}
	}
}
