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
)

func TestUserVerifier(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs(),
		web.WithInterceptors(
			web.UserInterceptorFunc,
		),
	)

	adminUser := qtest.CreateFakeUser(t, db)
	student := qtest.CreateFakeUser(t, db)

	userTest := []struct {
		code     connect.Code
		ctx      context.Context
		wantUser *qf.User
	}{
		{code: connect.CodeUnauthenticated, ctx: context.Background(), wantUser: nil},
		{code: connect.CodeUnauthenticated, ctx: context.Background(), wantUser: nil},
		{code: 0, ctx: client.Context(t, adminUser), wantUser: adminUser},
		{code: 0, ctx: client.Context(t, student), wantUser: student},
	}

	for _, user := range userTest {
		gotUser, err := client.GetUser(user.ctx, &qf.Void{})
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
			if diff := cmp.Diff(wantUser, gotUser, qtest.UserDiffOptions()); diff != "" {
				t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
			}
		}
	}
}
