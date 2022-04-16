package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/golang-jwt/jwt"
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
	cookie      string
}

func NewTokenManager(db database.Database, expireAfter time.Duration, secret, domain string) (*TokenManager, error) {
	if secret == "" || domain == "" {
		return nil, fmt.Errorf("failed to create a token manager: missing secret or domain")
	}
	manager := &TokenManager{
		db:          db,
		expireAfter: expireAfter,
		secret:      secret,
		domain:      domain,
	}
	if err := manager.Update(); err != nil {
		return nil, err
	}
	return manager, nil
}

// JWTUpdateRequired returns true if JWT update is needed for this user ID
func (tm *TokenManager) UpdateRequired(claims *Claims) bool {
	for _, token := range tm.tokens {
		if claims.UserID == token {
			return true
		}
	}
	return false
}

func (tm *TokenManager) NewTokenCookie(ctx context.Context, token *jwt.Token) (*http.Cookie, error) {
	signed, err := token.SignedString([]byte(tm.secret))
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     tm.cookie,
		Value:    signed,
		Domain:   tm.domain,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(tm.expireAfter),
	}, nil
}

func (tm *TokenManager) NewToken(claims *Claims) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodES256, claims)
}

// NewClaims creates user claims for a JWT token
func (tm *TokenManager) NewClaims(userID uint64) (*Claims, error) {
	usr, err := tm.db.GetUserWithEnrollments(userID)
	if err != nil {
		return nil, err
	}
	newClaims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tm.expireAfter).Unix(),
			IssuedAt:  time.Now().Unix(),
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

func (tm *TokenManager) GetClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("failed to parse token: incorrect signing method")
		}
		return []byte(tm.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("failed to parse token: invalid claims")
}

// Get returns the list with tokens that require update
func (tm *TokenManager) GetTokens() []uint64 {
	return tm.tokens
}

// Update removes user ID from the manager and updates user record in the database
func (tm *TokenManager) Remove(userID uint64) error {
	if !tm.exists(userID) {
		return fmt.Errorf("user with ID %d is not in the list", userID)
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
	// Return if the given ID is already in the list to avoid duplicates
	if tm.exists(userID) {
		return fmt.Errorf("user with ID %d is already in the list", userID)
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
