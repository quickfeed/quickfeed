package interceptor_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestUserVerifier(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t).Desugar()
	ags := web.NewQuickFeedService(logger, db, scm.TestSCMManager(), web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db, "test")
	if err != nil {
		t.Fatal(err)
	}

	adminUser := qtest.CreateFakeUser(t, db, 1)
	student := qtest.CreateFakeUser(t, db, 56)

	adminToken, err := tm.NewAuthCookie(adminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	studentToken, err := tm.NewAuthCookie(student.ID)
	if err != nil {
		t.Fatal(err)
	}

	router := http.NewServeMux()
	router.Handle(ags.NewQuickFeedHandler(tm))
	muxServer := &http.Server{
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		Addr:              "127.0.0.1:8081",
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}

	go func() {
		if err := muxServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				t.Errorf("Server exited with unexpected error: %v", err)
			}
			return
		}
	}()

	client := qtest.QuickFeedClient("")

	userTest := []struct {
		code     connect.Code
		metadata bool
		token    string
		wantUser *qf.User
	}{
		{code: connect.CodeUnauthenticated, metadata: false, token: "", wantUser: nil},
		{code: connect.CodeUnauthenticated, metadata: true, token: "should fail", wantUser: nil},
		{code: 0, metadata: true, token: auth.TokenString(adminToken), wantUser: adminUser},
		{code: 0, metadata: true, token: auth.TokenString(studentToken), wantUser: student},
	}

	ctx := context.Background()
	for _, user := range userTest {
		req := connect.NewRequest(&qf.Void{})
		if user.metadata {
			ctx = context.WithValue(ctx, auth.Cookie, user.token) // skipcq: GO-W5003
		}

		gotUser, err := client.GetUser(ctx, req)
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
	if err = muxServer.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}
