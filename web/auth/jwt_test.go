package auth_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
)

func TestNewManager(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user1 := qtest.CreateFakeUser(t, db)
	user2 := qtest.CreateFakeUser(t, db)

	user2.UpdateToken = true
	if err := db.UpdateUser(user2); err != nil {
		t.Fatal(err)
	}
	manager, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	// User 1 should not be in the update list.
	user1claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
		},
		UserID:  user1.GetID(),
		Admin:   true,
		Courses: make(map[uint64]qf.Enrollment_UserStatus, 0),
	}
	cookie, err := manager.UpdateCookie(&user1claims)
	if err != nil {
		t.Error(err)
	}
	if cookie != nil {
		t.Error("expected nil, got updated cookie")
	}

	// But must require update if claims are about to expire.
	user1claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-10 * time.Second))
	cookie, err = manager.UpdateCookie(&user1claims)
	if err != nil {
		t.Error(err)
	}
	if cookie == nil {
		t.Error("expected updated cookie, got nil")
	}

	// User 2 must be in the update list.
	user2claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
		},
		UserID: user2.GetID(),
		Admin:  false,
	}
	cookie, err = manager.UpdateCookie(&user2claims)
	if err != nil {
		t.Error(err)
	}
	if cookie == nil {
		t.Error("expected updated cookie, got nil")
	}
}

func TestNewCookie(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user := qtest.CreateFakeUser(t, db)
	manager, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	cookie, err := manager.NewAuthCookie(user.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if !cookie.Secure || !cookie.HttpOnly {
		t.Error("Cookie not secure")
	}
	if cookie.Name != auth.CookieName {
		t.Errorf("Incorrect cookie name. Expected %s, got %s", auth.CookieName, cookie.Name)
	}
	domain := env.Domain()
	if cookie.Domain != domain {
		t.Errorf("Incorrect cookie domain. Expected %s, got %s", domain, cookie.Domain)
	}
}

func TestUserClaims(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)
	manager, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	adminCookie, err := manager.NewAuthCookie(admin.GetID())
	if err != nil {
		t.Fatal(err)
	}
	adminClaims, err := manager.GetClaims(adminCookie.String())
	if err != nil {
		t.Fatal(err)
	}
	if adminClaims.UserID != admin.GetID() {
		t.Errorf("Incorrect user ID: expected %d, got %d", admin.GetID(), adminClaims.UserID)
	}
	if adminClaims.Issuer != "QuickFeed" {
		t.Errorf("Incorrect claims issuer: expecter 'QuickFeed', got %s", adminClaims.Issuer)
	}
	if !adminClaims.Admin {
		t.Error("No admin status for admin user in claims")
	}
	if len(adminClaims.Courses) != 1 {
		t.Errorf("Incorrect number of user courses: expected 1, got %d", len(adminClaims.Courses))
	}
	status, ok := adminClaims.Courses[1]
	if !ok {
		t.Error("No record for user course in claims")
	}
	if status != qf.Enrollment_TEACHER {
		t.Errorf("Incorrect enrollment status, expected %s, got %s", qf.Enrollment_TEACHER, status)
	}
}

func TestUpdateTokenList(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db)
	manager, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	claims := &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
		},
		UserID: admin.GetID(),
		Admin:  false,
	}
	// Admin should not be in the token update list.
	cookie, err := manager.UpdateCookie(claims)
	if err != nil {
		t.Error(err)
	}
	if cookie != nil {
		t.Error("expected nil, got updated cookie")
	}

	// Adding user must update manager's update list and database record.
	if err := manager.Add(admin.GetID()); err != nil {
		t.Fatal(err)
	}
	// Check database record first.
	updatedUser, err := db.GetUser(admin.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if !updatedUser.GetUpdateToken() {
		t.Error("User's 'UpdateToken' field not updated in the database")
	}
	// UpdateCookie will remove user from token list and update the database record.
	cookie, err = manager.UpdateCookie(claims)
	if err != nil {
		t.Error(err)
	}
	if cookie == nil {
		t.Error("expected updated cookie, got nil")
	}

	// Adding and then removing user from the list.
	if err := manager.Add(admin.GetID()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Remove(admin.GetID()); err != nil {
		t.Fatal(err)
	}
	// Database record should be updated.
	updatedUser, err = db.GetUser(admin.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if updatedUser.GetUpdateToken() {
		t.Error("User's 'UpdateToken' field not updated in the database")
	}
	// UpdateCookie must return nil and not an updated cookie.
	cookie, err = manager.UpdateCookie(claims)
	if err != nil {
		t.Error(err)
	}
	if cookie != nil {
		t.Error("expected nil, got updated cookie")
	}
}

func TestUpdateCookie(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := qtest.CreateFakeUser(t, db)
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	claims := &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Minute)),
		},
		UserID: user.GetID(),
		Admin:  false,
	}
	user.IsAdmin = false
	if err := db.UpdateUser(user); err != nil {
		t.Fatal(err)
	}
	// To trigger cookie update add user to the update list.
	if err := tm.Add(user.GetID()); err != nil {
		t.Error(err)
	}
	newCookie, err := tm.UpdateCookie(claims)
	if err != nil {
		t.Fatal(err)
	}
	if newCookie == nil {
		t.Error("expected updated cookie, got nil")
	}
	newClaims, err := tm.GetClaims(newCookie.String())
	if err != nil {
		t.Fatal(err)
	}
	if newClaims.Admin {
		t.Error("Got admin status in user claims for non-admin user")
	}
}

func TestExpiredTokenAndErrorCodePaths(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := qtest.CreateFakeUser(t, db)
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

	claims := &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-3 * time.Minute)), // token already expired
		},
		UserID: user.GetID(),
	}
	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(env.AuthSecret()))
	if err != nil {
		t.Fatal(fmt.Errorf("failed to sign token: %s", err))
	}

	newCookie := &http.Cookie{
		Name:  auth.CookieName,
		Value: signedToken,
	}
	newClaims, err := tm.GetClaims(newCookie.String())
	if err != nil {
		t.Fatal(err)
	}
	if newClaims.UserID != user.GetID() {
		t.Errorf("Expected user ID %d, got %d", user.GetID(), newClaims.UserID)
	}
	// TODO(meling): I'm a bit confused about GetClaims and how it handles expired tokens.
	// The GetClaims returns a valid claims object even if the token is expired.
	t.Logf("ExpiresAt: %s", newClaims.RegisteredClaims.ExpiresAt.Time)
	if !newClaims.RegisteredClaims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Expected token to be expired, but it is not")
	}

	otherClaims, err := tm.GetClaims("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if otherClaims != nil {
		t.Error("expected nil claims for invalid token, got non-nil claims")
	}

	otherCookie := &http.Cookie{
		Name:  auth.CookieName,
		Value: "not-a-valid-token",
	}
	otherClaims, err = tm.GetClaims(otherCookie.String())
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
	if otherClaims != nil {
		t.Error("expected nil claims for invalid token, got non-nil claims")
	}
}
