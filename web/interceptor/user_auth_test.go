package interceptor_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
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
	_, scmMgr := qtest.TestSCMManager(t)
	ags := web.NewQuickFeedService(logger, db, scmMgr, web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

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

	router := http.NewServeMux()
	router.Handle(ags.NewQuickFeedHandler(tm))
	muxServer := &http.Server{
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		Addr:              "127.0.0.1:8081",
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}

	serverReady := make(chan error, 1)
	go func() {
		listener, err := net.Listen("tcp", muxServer.Addr)
		if err != nil {
			serverReady <- err
			return
		}
		serverReady <- nil
		if err := muxServer.Serve(listener); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				t.Errorf("Server exited with unexpected error: %v", err)
			}
			return
		}
	}()

	if err := <-serverReady; err != nil {
		t.Fatal(err)
	}
	client := qtest.QuickFeedClient("")

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
	if err = muxServer.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}
