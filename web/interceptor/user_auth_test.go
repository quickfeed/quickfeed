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
	"google.golang.org/protobuf/testing/protocmp"
)

func TestUserVerifier(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs(),
		web.WithInterceptors(
			web.UserInterceptorFunc,
		),
	)
	ctx := context.Background()

	adminUser := qtest.CreateFakeUser(t, db)
	student := qtest.CreateFakeUser(t, db)

	adminCookie := client.Cookie(t, adminUser)
	studentCookie := client.Cookie(t, student)

	userTest := []struct {
		code     connect.Code
		cookie   string
		wantUser *qf.User
	}{
		{code: connect.CodeUnauthenticated, cookie: "", wantUser: nil},
		{code: connect.CodeUnauthenticated, cookie: "should fail", wantUser: nil},
		{code: 0, cookie: adminCookie, wantUser: adminUser},
		{code: 0, cookie: studentCookie, wantUser: student},
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
