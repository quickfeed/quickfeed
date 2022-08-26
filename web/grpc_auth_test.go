package web_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	grpcAddr = "127.0.0.1:8081"
	token    = "some-secret-string"
	// same as quickfeed root user
	// botUserID = 1
	userName = "meling"
)

var user *qf.User

func TestGrpcAuth(t *testing.T) {
	if os.Getenv("HELPBOT_TEST") == "" {
		t.Skip("Needs update for helpbot compatibility")
	}
	db, cleanup, _, qfService := testQuickFeedService(t)
	defer cleanup()

	fillDatabase(t, db)
	if user.Login != userName {
		t.Errorf("Expected %v, got %v\n", userName, user.Login)
	}

	tm, err := auth.NewTokenManager(db, "test")
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}
	// start gRPC server in background
	serveFn, shutdown := startGrpcAuthServer(t, qfService, tm)
	go serveFn()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		t.Fatalf("failed to connect to grpc server: %v", err)
	}
	defer conn.Close()

	client := qtest.QuickFeedClient("")

	// create request context with the helpbot's secret token
	reqCtx := metadata.NewOutgoingContext(ctx,
		metadata.New(map[string]string{auth.Cookie: token}),
	)

	request := connect.NewRequest(&qf.CourseUserRequest{
		CourseCode: "DAT320",
		CourseYear: 2021,
		UserLogin:  userName,
	})
	userInfo, err := client.GetUserByCourse(reqCtx, request)
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

func startGrpcAuthServer(t *testing.T, qfService *web.QuickFeedService, tm *auth.TokenManager) (func(), func(context.Context)) {
	t.Helper()

	router := http.NewServeMux()
	router.Handle(qfService.NewQuickFeedHandler(tm))
	muxServer := &http.Server{
		Handler: h2c.NewHandler(router, &http2.Server{}),
		Addr:    "127.0.0.1:8081",
	}

	return func() {
			if err := muxServer.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					t.Errorf("Server exited with unexpected error: %v", err)
				}
				return
			}
		}, func(ctx context.Context) {
			if err := muxServer.Shutdown(ctx); err != nil {
				t.Fatal(err)
			}
		}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
