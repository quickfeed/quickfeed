package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/grpc/metadata"
)

// Claims contain the bearer information.
type Claims struct {
	jwt.StandardClaims
	UserID  uint64                              `json:"user_id"`
	Admin   bool                                `json:"admin"`
	Courses map[uint64]qf.Enrollment_UserStatus `json:"courses"`
	Groups  []uint64                            `json:"groups"`
}

// TokenManager creates and updates JWTs.
type TokenManager struct {
	tokensToUpdate []uint64 // User IDs for user who need a token update.
	db             database.Database
	secret         string
	domain         string
}

// NewTokenManager starts a new token manager. Will create a list with all tokens that need update.
func NewTokenManager(db database.Database, domain string) (*TokenManager, error) {
	if domain == "" {
		return nil, errors.New("failed to create a new token manager: missing domain")
	}
	hostname, _, ok := strings.Cut(domain, ":")
	if ok {
		domain = hostname
	}
	manager := &TokenManager{
		db:     db,
		secret: rand.String(),
		domain: domain,
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
		Domain:   tm.domain,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(cookieExpirationTime),
		SameSite: http.SameSiteStrictMode,
	}, nil
}

// GetClaims returns validated user claims.
func (tm *TokenManager) GetClaims(ctx context.Context) (*Claims, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("failed to extract metadata from context")
	}
	tokenString, err := extractToken(meta)
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// It is necessary to check for correct signing algorithm in the header due to JWT vulnerability
		//  (ref https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/).
		if t.Header["alg"] != alg {
			return nil, fmt.Errorf("incorrect signing algorithm, expected %s, got %s", alg, t.Header["alg"])
		}
		return []byte(tm.secret), nil
	})
	if err != nil {
		if tokenExpired(err) {
			// token has expired; if signature is valid, update it.
			if err = tm.validateSignature(token); err == nil {
				return claims, nil
			}
		}
		return nil, err
	}
	claims, ok = token.Claims.(*Claims)
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
	newClaims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenExpirationTime).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "QuickFeed",
		},
		UserID: userID,
		Admin:  usr.IsAdmin,
	}
	userCourses := make(map[uint64]qf.Enrollment_UserStatus)
	for _, enrol := range usr.Enrollments {
		userCourses[enrol.GetCourseID()] = enrol.GetStatus()
	}
	newClaims.Courses = userCourses
	return newClaims, nil
}

// tokenExpired returns true if the given JWT validation error is due to an expired token.
func tokenExpired(err error) bool {
	v, ok := err.(*jwt.ValidationError)
	if ok {
		return v.Errors == jwt.ValidationErrorExpired
	}
	return ok
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

// extractToken extracts a JWT authentication token from metadata.
func extractToken(meta metadata.MD) (string, error) {
	cookies := meta.Get(Cookie)
	for _, cookie := range cookies {
		_, cookieValue, ok := strings.Cut(cookie, CookieName+"=")
		if ok {
			return strings.TrimSpace(cookieValue), nil
		}
	}
	return "", errors.New("failed to get authentication cookie from metadata")
}
