package auth

import (
	"fmt"
	"strconv"
	"strings"

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
	TokensToUpdate []uint64
}

// TODO: probably most of these methods (and struct fields) can be changed to unexported

// JWTUpdateRequired returns element index if the user ID is in the list,
// otherwise returns -1 to indicate that JWT update is not needed for this user
func (tm *TokenManager) JWTUpdateRequired(claims *Claims) uint64 {
	index := -1
	for i, tokenID := range tm.TokensToUpdate {
		if claims.UserID == tokenID {
			fmt.Printf("Token for UserID %d found in the refresh list", claims.UserID)
			index = i
		}
	}
	return uint64(index)
}

// JWTUpdateClaims fetches the up-to-date user information from the database and returns
// a new JWT user claims with the updated information
func (tm *TokenManager) JWTUpdateClaims(db database.Database, userID uint64) (*Claims, error) {
	usr, err := db.GetUserWithEnrollments(userID)
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

// Update will remove a user ID at the index from the list
func (tm *TokenManager) Remove(index uint64) {
	updatedTokenList := append(tm.TokensToUpdate[:index], tm.TokensToUpdate[index+1:]...)
	tm.TokensToUpdate = updatedTokenList
}

// Add adds a new UserID to the list of users who require an updated JWT token
func (tm *TokenManager) Add(userID uint64) {
	// Return if the given ID is already in the list to avoid duplicates
	for _, id := range tm.TokensToUpdate {
		if userID == id {
			return
		}
	}
	tm.TokensToUpdate = append(tm.TokensToUpdate, userID)
}

// String creates a string of all IDs in the list to store in the database
func (tm *TokenManager) String() string {
	return strings.Trim(strings.Join(strings.Split(fmt.Sprint(tm.TokensToUpdate), " "), ","), "[]")
}

// UpdateFromDatabase splits string of user IDs stored in the database into a list
func (tm *TokenManager) UpdateFromDatabase(stringWithIDs string) {
	stringIDs := strings.Split(stringWithIDs, ", ")
	integerIDs := make([]uint64, len(stringIDs))
	for i, id := range stringIDs {
		var intID int
		intID, err := strconv.Atoi(id)
		if err != nil {
			fmt.Printf("Error updating token manager from database: failed converting ID %s to int: %s", id, err)
		}
		integerIDs[i] = uint64(intID)
	}
	tm.TokensToUpdate = integerIDs
}
