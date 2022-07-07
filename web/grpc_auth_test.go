package web_test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/types"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	grpcAddr = "127.0.0.1:9090"
	token    = "some-secret-string"
	// same as quickfeed root user
	botUserID = 1
	userName  = "meling"
)

var user *types.User

func TestGrpcAuth(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	fillDatabase(t, db)
	if user.Login != userName {
		t.Errorf("Expected %v, got %v\n", userName, user.Login)
	}

	// start gRPC server in background
	go startGrpcAuthServer(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		t.Fatalf("failed to connect to grpc server: %v", err)
	}
	defer conn.Close()

	client := qf.NewQuickFeedServiceClient(conn)

	// create request context with the helpbot's secret token
	reqCtx := metadata.NewOutgoingContext(ctx,
		metadata.New(map[string]string{auth.Cookie: token}),
	)

	request := &types.CourseUserRequest{
		CourseCode: "DAT320",
		CourseYear: 2021,
		UserLogin:  userName,
	}
	userInfo, err := client.GetUserByCourse(reqCtx, request)
	check(t, err)
	if userInfo.ID != user.ID {
		t.Errorf("expected user id %d, got %d", user.ID, userInfo.ID)
	}
	if userInfo.Login != user.Login {
		t.Errorf("expected user login %s, got %s", user.Login, userInfo.Login)
	}
}

func fillDatabase(t *testing.T, db database.Database) {
	// Add secret token for the helpbot application (to allow it to invoke gRPC methods)
	auth.Add(token, botUserID)

	// Check that token was stored and maps to correct user
	checkCookie := auth.Get(token)
	if checkCookie != botUserID {
		t.Errorf("Expected %v, got %v\n", botUserID, checkCookie)
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	// admin := qtest.CreateUser(t, db, 1, &types.User{Login: "admin"})
	course := &types.Course{
		Code: "DAT320",
		Name: "Operating Systems and Systems Programming",
		Year: 2021,
	}
	qtest.CreateCourse(t, db, admin, course)

	user = qtest.CreateUser(t, db, 11, &types.User{Login: userName})
	qtest.EnrollStudent(t, db, user, course)
}

func startGrpcAuthServer(t *testing.T, db database.Database) {
	logger := qtest.Logger(t)

	_, scms := qtest.FakeProviderMap(t)
	agService := web.NewQuickFeedService(logger.Desugar(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	lis, err := net.Listen("tcp", grpcAddr)
	check(t, err)

	opt := grpc.ChainUnaryInterceptor(
		auth.UserVerifier(),
	)
	grpcServer := grpc.NewServer(opt)

	qf.RegisterQuickFeedServiceServer(grpcServer, agService)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v\n", err)
	}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
