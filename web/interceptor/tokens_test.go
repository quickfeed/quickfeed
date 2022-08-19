package interceptor_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestRefreshTokens(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, scm.TestSCMManager(), web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db, "test")
	if err != nil {
		t.Fatal(err)
	}

	lis := bufconn.Listen(BufSize)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	opt := grpc.ChainUnaryInterceptor(
		interceptor.UnaryUserVerifier(logger, tm),
		interceptor.TokenRefresher(logger, tm),
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
	f := func(t *testing.T, id uint64) context.Context {
		token, err := tm.NewAuthCookie(id)
		if err != nil {
			t.Fatal(err)
		}
		return qtest.WithAuthCookie(ctx, token)
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	user := qtest.CreateFakeUser(t, db, 56)
	adminCtx := f(t, admin.ID)
	userCtx := f(t, user.ID)
	adminClaims := &auth.Claims{
		UserID: admin.ID,
		Admin:  true,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
		},
	}
	userClaims := &auth.Claims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
		},
	}
	if tm.UpdateRequired(adminClaims) || tm.UpdateRequired(userClaims) {
		t.Error("No users should be in the token update list at the start")
	}
	if _, err := client.GetUsers(adminCtx, &qf.Void{}); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(adminClaims) || tm.UpdateRequired(userClaims) {
		t.Error("No users should be in the token update list")
	}
	if _, err := client.UpdateUser(adminCtx, user); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update list after admin has updated the user's information")
	}
	if _, err := client.GetUser(userCtx, &qf.Void{}); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should not be in the token update list after the token has been updated")
	}
	course := &qf.Course{
		ID:               1,
		OrganizationID:   1,
		OrganizationPath: "testorg",
		Provider:         "fake",
	}
	group := &qf.Group{
		ID:       1,
		Name:     "test",
		CourseID: 1,
		Users: []*qf.User{
			user,
		},
	}
	if _, err := client.CreateCourse(adminCtx, course); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(adminClaims) {
		t.Error("Admin must be in the token update list after creating a new course")
	}
	qtest.EnrollStudent(t, db, user, course)
	if _, err := client.CreateGroup(adminCtx, group); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should not be in the token update list after methods that don't affect the user's information")
	}
	if _, err := client.UpdateGroup(adminCtx, group); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update group after changes to the group")
	}
	if _, err := client.GetUser(userCtx, &qf.Void{}); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should be removed from the token update list after the user's token has been updated")
	}
	if _, err := client.DeleteGroup(adminCtx, &qf.GroupRequest{
		GroupID:  group.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update list after the group has been deleted")
	}
	if tm.UpdateRequired(adminClaims) {
		t.Error("Admin should not be in the token update list")
	}
}
