package tokens

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
)

var (
	authCookieName = "auth"
	// Time left till expiration to trigger auto update.
	refreshTime = 1 * time.Minute
	alg         = "HS256"
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
	updateTokens []uint64
	db           database.Database
	expireAfter  time.Duration
	secret       string
	domain       string
	cookieName   string
}

// NewTokenManager starts a new token manager. Will create a list with all tokens that need update.
func NewTokenManager(db database.Database, expireAfter time.Duration, secret, domain string) (*TokenManager, error) {
	if secret == "" || domain == "" {
		return nil, fmt.Errorf("failed to create a new token manager: missing secret (%s) or domain (%s)", secret, domain)
	}
	manager := &TokenManager{
		db:          db,
		expireAfter: expireAfter,
		secret:      secret,
		domain:      domain,
		cookieName:  authCookieName,
	}
	if err := manager.Update(); err != nil {
		return nil, err
	}
	return manager, nil
}

// NewAuthCookie creates a cookie with signed JWT from user ID.
func (tm *TokenManager) NewAuthCookie(userID uint64) (*http.Cookie, error) {
	claims, err := tm.newClaims(userID)
	if err != nil {
		return nil, err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signed, err := token.SignedString([]byte(tm.secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %s", err)
	}
	return &http.Cookie{
		Name:     tm.cookieName,
		Value:    signed,
		Domain:   tm.domain, // TODO(vera): looks like you have to omit this field when working on localhost
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(tm.expireAfter),
		SameSite: http.SameSiteStrictMode,
	}, nil
}

// GetClaims returns user claims after parsing and validating a signed token string
func (tm *TokenManager) GetClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Header["alg"] != alg {
			return nil, fmt.Errorf("incorect signing algorithm, expected %s, got %s", alg, t.Header["alg"])
		}
		return []byte(tm.secret), nil
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("Signing algorithm in header: %+v", token.Header["alg"]) // tmp

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("failed to parse token: validation error")
}

func (tm *TokenManager) GetAuthCookieName() string {
	return tm.cookieName
}

// newClaims creates new JWT claims for user ID
func (tm *TokenManager) newClaims(userID uint64) (*Claims, error) {
	usr, err := tm.db.GetUserWithEnrollments(userID)
	if err != nil {
		return nil, err
	}
	newClaims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tm.expireAfter).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "Quickfeed",
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
