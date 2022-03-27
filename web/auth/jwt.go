package auth

import (
	"fmt"
	"log"

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

func (tm *TokenManager) JWTUpdateRequired(claims *Claims) bool {
	for _, tokenID := range tm.TokensToUpdate {
		if claims.UserID == tokenID {
			log.Printf("Token for UserID %d found in the refresh list", claims.UserID)
			return true
		}
	}
	return false
}

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
