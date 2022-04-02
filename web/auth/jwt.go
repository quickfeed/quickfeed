package auth

import (
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.StandardClaims
	UserID  uint64                              `json:"user_id"`
	Admin   bool                                `json:"admin"`
	Courses map[uint64]pb.Enrollment_UserStatus `json:"courses"`
}

type TokenManager struct {
	TokensToUpdate []*pb.UpdateTokenRecord // UserID
}

// TODO: probably most of these methods (and struct fields) can be changed to unexported

// JWTUpdateRequired returns element index if the user ID is in the list,
// otherwise returns -1 to indicate that JWT update is not needed for this user
func (tm *TokenManager) JWTUpdateRequired(claims *Claims) int {
	for _, token := range tm.TokensToUpdate {
		fmt.Printf("Comparing %d with %d", claims.UserID, token.UserID)
		if claims.UserID == token.UserID {
			fmt.Printf("Token for UserID %d found in the refresh list", claims.UserID)
			return int(token.UserID)
		}
	}
	return -1
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

// TODO(vera): needs access to the database to update changes there
// Update will remove a user ID at the index from the list
func (tm *TokenManager) Remove(index uint64, db database.Database) {
	updatedTokenList := append(tm.TokensToUpdate[:index], tm.TokensToUpdate[index+1:]...)
	tm.TokensToUpdate = updatedTokenList
}

// Add adds a new UserID to the list of users who require an updated JWT token
func (tm *TokenManager) Add(userID uint64, logger *zap.SugaredLogger, db database.Database) {
	// Return if the given ID is already in the list to avoid duplicates
	for _, token := range tm.TokensToUpdate {
		if userID == token.UserID {
			return
		}
	}
	tokenQuery := &pb.UpdateTokenRecord{UserID: userID}
	if err := db.CreateTokenRecord(tokenQuery); err != nil {
		logger.Errorf("error updating token record in the database: %w", err)
	}
	// TODO(vera): remove, testing if ID is getting updated, must be tested in the corresponding test
	logger.Debugf("Token query after update in the database: %+v", tokenQuery)
	tm.TokensToUpdate = append(tm.TokensToUpdate, tokenQuery)
}

// UpdateFromDatabase splits string of user IDs stored in the database into a list
func (tm *TokenManager) UpdateFromDatabase(logger *zap.SugaredLogger, db database.Database) {
	tokens, err := db.GetTokenRecords()
	if err != nil {
		logger.Errorf("cannot fetch token to update from the database: %w", err)
	}
	tm.TokensToUpdate = tokens
}
