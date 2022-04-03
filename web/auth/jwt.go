package auth

import (
	"fmt"

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
	TokensToUpdate []uint64 // UserID
	DB             database.Database
}

// TODO(vera): probably most of these methods (and struct fields) can be changed to unexported

// JWTUpdateRequired returns true if JWT update is needed for this user ID
func (tm *TokenManager) UpdateRequired(claims *Claims) bool {
	for _, token := range tm.TokensToUpdate {
		fmt.Printf("Comparing %d with %d", claims.UserID, token)
		if claims.UserID == token {
			fmt.Printf("Token for UserID %d found in the refresh list", claims.UserID)
			return true
		}
	}
	return false
}

// UpdateClaims fetches the up-to-date user information from the database and returns
// updated JWT user claims
func (tm *TokenManager) UpdateClaims(userID uint64) (*Claims, error) {
	usr, err := tm.DB.GetUserWithEnrollments(userID)
	if err != nil {
		return nil, err
	}
	newClaims := &Claims{
		UserID: userID,
		Admin:  usr.IsAdmin,
	}
	userCourses := make(map[uint64]pb.Enrollment_UserStatus)
	for _, enrol := range usr.Enrollments {
		fmt.Printf("User %d enrolled into course %d %s with role %d", userID, enrol.GetCourseID(), enrol.GetCourse().GetName(), enrol.GetStatus())
		userCourses[enrol.GetCourseID()] = enrol.GetStatus()
	}
	newClaims.Courses = userCourses
	return newClaims, nil
}

// Update removes user ID from the manager and updates user record in the database
func (tm *TokenManager) Remove(userID uint64) error {
	if !tm.exists(userID) {
		return fmt.Errorf("User ID %d is not in the list", userID)
	}
	if err := tm.update(userID, false); err != nil {
		return err
	}
	var updatedTokenList []uint64
	for _, id := range tm.TokensToUpdate {
		if id != userID {
			updatedTokenList = append(updatedTokenList, id)
		}
	}
	tm.TokensToUpdate = updatedTokenList
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
	tm.TokensToUpdate = append(tm.TokensToUpdate, userID)
	return nil
}

// Update fetches IDs of users who need token updates from the database
func (tm *TokenManager) Update() error {
	users, err := tm.DB.GetUsers()
	if err != nil {
		return fmt.Errorf("cannot fetch token to update from the database: %w", err)
	}
	var tokens []uint64
	for _, user := range users {
		if user.UpdateToken {
			tokens = append(tokens, user.ID)
		}
	}
	tm.TokensToUpdate = tokens
	return nil
}

// update updates user record in the database
func (tm *TokenManager) update(userID uint64, updateToken bool) error {
	user, err := tm.DB.GetUser(userID)
	if err != nil {
		return err
	}
	user.UpdateToken = updateToken
	if err := tm.DB.UpdateUser(user); err != nil {
		return err
	}
	return nil
}

// exists checks if ID is in the list
func (tm *TokenManager) exists(id uint64) bool {
	for _, token := range tm.TokensToUpdate {
		if id == token {
			return true
		}
	}
	return false
}
