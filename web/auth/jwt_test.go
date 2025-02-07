package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
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
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
		UserID:  user1.ID,
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
	user1claims.StandardClaims.ExpiresAt = time.Now().Unix() - 10
	cookie, err = manager.UpdateCookie(&user1claims)
	if err != nil {
		t.Error(err)
	}
	if cookie == nil {
		t.Error("expected updated cookie, got nil")
	}

	// User 2 must be in the update list.
	user2claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
		UserID: user2.ID,
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
	cookie, err := manager.NewAuthCookie(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !(cookie.Secure && cookie.HttpOnly) {
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
	adminCookie, err := manager.NewAuthCookie(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	adminClaims, err := manager.GetClaims(adminCookie.String())
	if err != nil {
		t.Fatal(err)
	}
	if adminClaims.UserID != admin.ID {
		t.Errorf("Incorrect user ID: expected %d, got %d", admin.ID, adminClaims.UserID)
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
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
		UserID: admin.ID,
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
	if err := manager.Add(admin.ID); err != nil {
		t.Fatal(err)
	}
	// Check database record first.
	updatedUser, err := db.GetUser(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !updatedUser.UpdateToken {
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
	if err := manager.Add(admin.ID); err != nil {
		t.Fatal(err)
	}
	if err := manager.Remove(admin.ID); err != nil {
		t.Fatal(err)
	}
	// Database record should be updated.
	updatedUser, err = db.GetUser(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedUser.UpdateToken {
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
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 3).Unix(),
		},
		UserID: user.ID,
		Admin:  false,
	}
	user.IsAdmin = false
	if err := db.UpdateUser(user); err != nil {
		t.Fatal(err)
	}
	// To trigger cookie update add user to the update list.
	if err := tm.Add(user.ID); err != nil {
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
		t.Error("Admin status in user claims for demoted user")
	}
}
