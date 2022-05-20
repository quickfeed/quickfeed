package auth

import (
	"fmt"
	"net/http"
	"time"

	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/database"
	"github.com/golang-jwt/jwt"
)

var (
	authCookieName = "auth"
	refreshTime    = 1 * time.Minute
)

type Claims struct {
	jwt.StandardClaims
	UserID  uint64                              `json:"user_id"`
	Admin   bool                                `json:"admin"`
	Courses map[uint64]pb.Enrollment_UserStatus `json:"courses"`
}

type TokenManager struct {
	tokens      []uint64
	db          database.Database
	expireAfter time.Duration
	secret      string
	domain      string
	cookieName  string
}

// NewTokenManager creates a new token manager, populating
// the token list with user IDs from the database
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
	// Collect IDs of users who require token update from database
	if err := manager.Update(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (tm *TokenManager) GetAuthCookieName() string {
	return tm.cookieName
}

// JWTUpdateRequired returns true if JWT update is needed for this user ID
// due to updated user role or token expiration time
func (tm *TokenManager) UpdateRequired(claims *Claims) bool {
	for _, token := range tm.tokens {
		if claims.UserID == token {
			return true
		}
	}
	if claims.ExpiresAt-time.Now().Unix() < refreshTime.Milliseconds() {
		fmt.Println("Updating token, expires after ", claims.ExpiresAt-time.Now().Unix() < refreshTime.Milliseconds()) // tmp
		return true
	}
	return false
}

// NewTokenCookie creates a cookie with signed JWT from user ID.
func (tm *TokenManager) NewTokenCookie(userID uint64) (*http.Cookie, error) {
	claims, err := tm.NewClaims(userID)
	if err != nil {
		return nil, err
	}
	token := newToken(claims)
	fmt.Printf("Making new token cookie: secret is %s", tm.secret) // tmp
	signed, err := token.SignedString([]byte(tm.secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %s", err)
	}
	return &http.Cookie{
		Name:  tm.cookieName,
		Value: signed,
		// Domain:   tm.domain,  // TODO(vera): looks like you have to omit this field when working on localhost
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(tm.expireAfter),
		SameSite: http.SameSiteStrictMode, // http.SameSiteLaxMode,
	}, nil
}

// NewClaims creates new JWT claims for user ID
func (tm *TokenManager) NewClaims(userID uint64) (*Claims, error) {
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
	userCourses := make(map[uint64]pb.Enrollment_UserStatus)
	for _, enrol := range usr.Enrollments {
		userCourses[enrol.GetCourseID()] = enrol.GetStatus()
	}
	newClaims.Courses = userCourses
	return newClaims, nil
}

// GetClaims returns user claims after parsing and validating a signed token string
func (tm *TokenManager) GetClaims(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to parse token: incorrect signing method")
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

// Get returns the list with tokens that require update
func (tm *TokenManager) GetTokens() []uint64 {
	return tm.tokens
}

// Update removes user ID from the manager and updates user record in the database
func (tm *TokenManager) Remove(userID uint64) error {
	if !tm.exists(userID) {
		return nil
	}
	if err := tm.update(userID, false); err != nil {
		return err
	}
	var updatedTokenList []uint64
	for _, id := range tm.tokens {
		if id != userID {
			updatedTokenList = append(updatedTokenList, id)
		}
	}
	tm.tokens = updatedTokenList
	return nil
}

// Add adds a new UserID to the manager and updates user record in the database
func (tm *TokenManager) Add(userID uint64) error {
	if tm.exists(userID) {
		return nil
	}
	if err := tm.update(userID, true); err != nil {
		return err
	}
	tm.tokens = append(tm.tokens, userID)
	return nil
}

// Update fetches IDs of users who need token updates from the database
func (tm *TokenManager) Update() error {
	users, err := tm.db.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to update JWT tokens from database: %w", err)
	}
	var tokens []uint64
	for _, user := range users {
		if user.UpdateToken {
			tokens = append(tokens, user.ID)
		}
	}
	tm.tokens = tokens
	return nil
}

// update updates user record in the database
func (tm *TokenManager) update(userID uint64, updateToken bool) error {
	user, err := tm.db.GetUser(userID)
	if err != nil {
		return err
	}
	user.UpdateToken = updateToken
	if err := tm.db.UpdateUser(user); err != nil {
		return err
	}
	return nil
}

// exists checks if ID is in the list
func (tm *TokenManager) exists(id uint64) bool {
	for _, token := range tm.tokens {
		if id == token {
			return true
		}
	}
	return false
}

// NewToken makes a new JWT token with given claims
func newToken(claims *Claims) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
}
