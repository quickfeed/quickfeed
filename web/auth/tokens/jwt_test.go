package tokens_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth/tokens"
)

func TestNewManager(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user1 := qtest.CreateFakeUser(t, db, 1)
	user2 := qtest.CreateFakeUser(t, db, 2)

	user2.UpdateToken = true
	if err := db.UpdateUser(user2); err != nil {
		t.Fatal(err)
	}
	// Create manager with missing required parameters.
	manager, err := tokens.NewTokenManager(db, time.Minute*15, "notasecret", "")
	if err == nil {
		t.Fatal("Expected error: missing secret or domain variable")
	}
	manager, err = tokens.NewTokenManager(db, time.Minute*15, "", "localhost")
	if err == nil {
		t.Fatal("Expected error: missing secret or domain variable")
	}
	manager, err = tokens.NewTokenManager(db, time.Minute*15, "notasecret", "test")
	if err != nil {
		t.Fatal(err)
	}
	// User 1 should not be in the update list.
	user1claims := tokens.Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
		UserID:  user1.ID,
		Admin:   true,
		Courses: make(map[uint64]qf.Enrollment_UserStatus, 0),
	}
	if manager.UpdateRequired(&user1claims) {
		t.Error("JWT update required is true, expected false")
	}
	// But must require update if claims are about to expire.
	user1claims.StandardClaims.ExpiresAt = time.Now().Add(time.Second * 10).Unix()
	if !manager.UpdateRequired(&user1claims) {
		t.Error("JWT update required is false for expiring token, expected true")
	}
	// User 2 must be in the update list.
	user2claims := tokens.Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
		UserID: user2.ID,
		Admin:  false,
	}
	if !manager.UpdateRequired(&user2claims) {
		t.Error("JWT update required is false, expected true")
	}
}

func TestNewCookie(t *testing.T) {
	secret := qtest.RandomString(t)
	domain := "test"
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user := qtest.CreateFakeUser(t, db, 1)
	manager, err := tokens.NewTokenManager(db, time.Minute*15, secret, domain)
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
	if cookie.Name != manager.GetAuthCookieName() {
		t.Errorf("Incorrect cookie name. Expected %s, got %s", manager.GetAuthCookieName(), cookie.Name)
	}
	if cookie.Domain != domain {
		t.Errorf("Incorrect cookie domain. Expected %s, got %s", domain, cookie.Domain)
	}
}

func TestUserClaims(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)
	manager, err := tokens.NewTokenManager(db, time.Minute*15, "notasecret", "localhost")
	if err != nil {
		t.Fatal(err)
	}
	adminCookie, err := manager.NewAuthCookie(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	adminClaims, err := manager.GetClaims(adminCookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	if adminClaims.UserID != admin.ID {
		t.Errorf("Incorrect user ID: expexted %d, got %d", admin.ID, adminClaims.UserID)
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
	admin := qtest.CreateFakeUser(t, db, 1)
	manager, err := tokens.NewTokenManager(db, time.Minute*15, "notasecret", "localhost")
	if err != nil {
		t.Fatal(err)
	}
	claims := &tokens.Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
		UserID: admin.ID,
		Admin:  false,
	}
	// Admin should not be in the token update list.
	if manager.UpdateRequired(claims) {
		t.Error("JWT update required is true, expected false")
	}
	// Adding user must update manager's update list and database record.
	manager.Add(admin.ID)
	if !manager.UpdateRequired(claims) {
		t.Error("JWT update required is false, expected true")
	}
	updatedUser, err := db.GetUser(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !updatedUser.UpdateToken {
		t.Error("User's 'UpdateToken' field not updated in the database")
	}
	// Removing user must update token list and user record in the database.
	manager.Remove(admin.ID)
	if manager.UpdateRequired(claims) {
		t.Error("JWT update required is true, expected false")
	}
	updatedUser, err = db.GetUser(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedUser.UpdateToken {
		t.Error("User's 'UpdateToken' field not updated in the database")
	}
}
