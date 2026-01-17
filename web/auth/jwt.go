package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
)

// Claims contain the bearer information.
type Claims struct {
	UserID  uint64                              `json:"user_id"`
	Admin   bool                                `json:"admin"`
	Courses map[uint64]qf.Enrollment_UserStatus `json:"courses"`
	Groups  []uint64                            `json:"groups"`
	jwt.RegisteredClaims
}

// TokenManager creates and updates JWTs.
type TokenManager struct {
	tokensToUpdate []uint64 // User IDs for user who need a token update.
	db             database.Database
	secret         string
	parser         *jwt.Parser
}

// NewTokenManager starts a new token manager. Will create a list with all tokens that need update.
func NewTokenManager(db database.Database) (*TokenManager, error) {
	manager := &TokenManager{
		db:     db,
		secret: env.AuthSecret(),
		parser: jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()})),
	}
	if err := manager.updateTokenList(); err != nil {
		return nil, err
	}
	return manager, nil
}

// NewAuthCookie creates a signed JWT cookie from user ID.
func (tm *TokenManager) NewAuthCookie(userID uint64) (*http.Cookie, error) {
	claims, err := tm.newClaims(userID)
	if err != nil {
		return nil, err
	}
	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(tm.secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %s", err)
	}
	return &http.Cookie{
		Name:     CookieName,
		Value:    signedToken,
		Domain:   env.Domain(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(cookieExpirationTime),
		SameSite: http.SameSiteStrictMode,
	}, nil
}

// GetClaims returns validated user claims.
func (tm *TokenManager) GetClaims(cookie string) (*Claims, error) {
	tokenString, err := extractToken(cookie)
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	token, err := tm.parser.ParseWithClaims(tokenString, claims, func(*jwt.Token) (any, error) {
		return []byte(tm.secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			// token has expired; if signature is valid, update it.
			if err = tm.validateSignature(token); err == nil {
				return claims, nil
			}
		}
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("failed to parse token: validation error")
	}
	return claims, nil
}

func (tm *TokenManager) Database() database.Database {
	return tm.db
}

// newClaims creates new JWT claims for user ID.
func (tm *TokenManager) newClaims(userID uint64) (*Claims, error) {
	usr, err := tm.db.GetUserWithEnrollments(userID)
	if err != nil {
		return nil, err
	}
	userCourses := make(map[uint64]qf.Enrollment_UserStatus)
	userGroups := make([]uint64, 0)
	for _, enrol := range usr.GetEnrollments() {
		userCourses[enrol.GetCourseID()] = enrol.GetStatus()
		if enrol.GetGroupID() != 0 {
			userGroups = append(userGroups, enrol.GetGroupID())
		}
	}

	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "QuickFeed",
		},
		UserID:  userID,
		Admin:   usr.GetIsAdmin(),
		Courses: userCourses,
		Groups:  userGroups,
	}, nil
}

// validateSignature checks the validity of the signature.
// This makes it possible to update expired JWTs. The built in methods
// will return an error for an expired JWT before validating the signature.
func (tm *TokenManager) validateSignature(token *jwt.Token) error {
	signingString, err := token.SigningString()
	if err != nil {
		return err
	}
	return token.Method.Verify(signingString, token.Signature, []byte(tm.secret))
}

// extractToken returns a JWT authentication token extracted from the request header's cookie.
func extractToken(cookieString string) (string, error) {
	for cookie := range strings.SplitSeq(cookieString, ";") {
		_, cookieValue, ok := strings.Cut(cookie, CookieName+"=")
		if ok {
			return strings.TrimSpace(cookieValue), nil
		}
	}
	return "", errors.New("failed to extract authentication cookie from request header")
}

// Context returns a new context with the claims as value.
func (c *Claims) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyClaims, c)
}

// ClaimsFromContext returns the claims value from the context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*Claims)
	return claims, ok
}

type (
	userIDProvider  interface{ GetUserID() uint64 }
	groupIDProvider interface{ GetGroupID() uint64 }
)

// IsCourseStudent returns true if the user is a student in the course.
func (c *Claims) IsCourseStudent(courseID uint64) bool {
	return c.Courses[courseID] == qf.Enrollment_STUDENT
}

// IsCourseTeacher returns true if the user is a teacher in the course.
func (c *Claims) IsCourseTeacher(courseID uint64) bool {
	return c.Courses[courseID] == qf.Enrollment_TEACHER
}

// SameUser returns true if user ID in request is the same as in claims.
func (c *Claims) SameUser(req any) bool {
	if uid, ok := req.(userIDProvider); ok {
		return uid.GetUserID() == c.UserID
	}
	return false
}

// IsGroupMember returns true if the request is a group and the claim's user is a member of that group.
func (c *Claims) IsGroupMember(req any) bool {
	if group, ok := req.(*qf.Group); ok {
		return group.Contains(&qf.User{ID: c.UserID})
	}
	return false
}

// IsInGroup returns true if the group ID in the request is in the claims' groups.
func (c *Claims) IsInGroup(req any) bool {
	if gid, ok := req.(groupIDProvider); ok {
		return slices.Contains(c.Groups, gid.GetGroupID())
	}
	return false
}

// UnauthorizedAdminChange returns true if an unauthorized admin change is attempted.
func (c *Claims) UnauthorizedAdminChange(req any) bool {
	if user, ok := req.(*qf.User); ok {
		return user.GetIsAdmin() && !c.Admin
	}
	return false
}

func (c *Claims) String() string {
	admin := ""
	if c.Admin {
		admin = " (admin)"
	}
	return fmt.Sprintf("UserID: %d%s: Courses: %v, Groups: %v", c.UserID, admin, c.Courses, c.Groups)
}
