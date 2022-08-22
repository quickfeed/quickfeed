package interceptor_test

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		Handler: h2c.NewHandler(router, &http2.Server{}),
		Addr:    "127.0.0.1:8081",
	}

	go func() {
		if err := muxServer.ListenAndServe(); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	client := qfconnect.NewQuickFeedServiceClient(http.DefaultClient, "http://127.0.0.1:8081/")

	userTest := []struct {
		code     codes.Code
		metadata bool
		token    string
		wantUser *qf.User
	}{
		{code: codes.Unauthenticated, metadata: false, token: "", wantUser: nil},
		{code: codes.Unauthenticated, metadata: true, token: "should fail", wantUser: nil},
		{code: codes.OK, metadata: true, token: auth.TokenString(adminToken), wantUser: adminUser},
		{code: codes.OK, metadata: true, token: auth.TokenString(studentToken), wantUser: student},
	}

	ctx := context.Background()
	for _, user := range userTest {
		req := connect.NewRequest(&qf.Void{})
		if user.metadata {
			req.Header().Set(auth.Cookie, user.token)
		}

		gotUser, err := client.GetUser(ctx, req)
		if err, ok := status.FromError(err); ok {
			if err.Code() != user.code {
				t.Fatalf("got code %v, want %v", err.Code(), user.code)
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
}
