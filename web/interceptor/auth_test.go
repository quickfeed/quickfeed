package interceptor_test

import (
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestUserVerifier(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t).Desugar()
	ags := web.NewQuickFeedService(logger, db, &scm.Manager{}, web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db, "test")
	if err != nil {
		t.Fatal(err)
	}

	const (
		bufSize = 1024 * 1024
	)

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

	lis := bufconn.Listen(bufSize)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	opt := grpc.ChainUnaryInterceptor(
		interceptor.UnaryUserVerifier(qtest.Logger(t), tm),
	)
	s := grpc.NewServer(opt) // skipcq: GO-S0902
	qf.RegisterQuickFeedServiceServer(s, ags)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
			return
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := qf.NewQuickFeedServiceClient(conn)

	userTest := []struct {
		code     codes.Code
		metadata bool
		token    string
		wantUser *qf.User
	}{
		{code: codes.Unauthenticated, metadata: false, token: "", wantUser: nil},
		{code: codes.Unauthenticated, metadata: true, token: "should fail", wantUser: nil},
		{code: codes.OK, metadata: true, token: interceptor.AuthTokenString(adminToken.Value), wantUser: adminUser},
		{code: codes.OK, metadata: true, token: interceptor.AuthTokenString(studentToken.Value), wantUser: student},
	}

	for _, user := range userTest {
		if user.metadata {
			meta := metadata.MD{}
			meta.Set(auth.Cookie, user.token)
			ctx = metadata.NewOutgoingContext(ctx, meta)
		}

		gotUser, err := client.GetUser(ctx, &qf.Void{})
		if s, ok := status.FromError(err); ok {
			if s.Code() != user.code {
				t.Errorf("GetUser().Code(): %v, want: %v", s.Code(), user.code)
			}
		}
		if user.wantUser != nil {
			// ignore comparing remote identity
			user.wantUser.RemoteIdentities = nil
		}
		wantUser := user.wantUser
		if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
			t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
		}
	}
}
