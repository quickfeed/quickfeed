package assignments

import (
	"context"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
)

// countMap maps a (courseID/groupID, userID)-pair to the number reviews
// the user has been assigned for the given course/group.
type countMap map[uint64]map[uint64]int

// Creates a new map if none exists for the given course/group id.
func (m countMap) initialize(id uint64) {
	if _, ok := m[id]; !ok {
		m[id] = make(map[uint64]int) // [id][userID] -> count
	}
}

var (
	teacherReviewCounter = make(countMap) // [courseID][userID] -> count
	groupReviewCounter   = make(countMap) // [groupID][userID] -> count
)

// AssignReviewers assigns reviewers to a group repository pull request.
// It assigns one other group member and one course teacher as reviewers.
func AssignReviewers(ctx context.Context, sc scm.SCM, db database.Database, course *pb.Course, repo *pb.Repository, pullRequest *pb.PullRequest) error {
	teacherReviewer, err := getNextTeacherReviewer(db, course)
	if err != nil {
		return err
	}
	studentReviewer, err := getNextStudentReviewer(db, repo.GetGroupID(), pullRequest.GetUserID())
	if err != nil {
		return err
	}

	opt := &scm.RequestReviewersOptions{
		Organization: course.GetOrganizationPath(),
		Repository:   repo.Name(),
		Number:       int(pullRequest.GetNumber()),
		Reviewers: []string{
			teacherReviewer.GetLogin(),
			studentReviewer.GetLogin(),
		},
	}
	if err := sc.RequestReviewers(ctx, opt); err != nil {
		return err
	}
	// Change pull request stage to review
	pullRequest.SetReview()
	return db.UpdatePullRequest(pullRequest)
}

// getNextReviewer gets the next reviewer from either teacherReviewCounter or studentReviewCounter,
// based on whoever in total has been assigned to the least amount of pull requests.
// It is simple, and does not account for how many current review requests any user has.
func getNextReviewer(users []*pb.User, reviewCounter map[uint64]int) *pb.User {
	userWithLowestCount := users[0]
	lowestCount := reviewCounter[userWithLowestCount.GetID()]
	for _, user := range users {
		count, ok := reviewCounter[user.GetID()]
		if !ok {
			// Found user with no prior reviews; assign as the next reviewer.
			reviewCounter[user.GetID()] = 1
			return user
		}
		if count < lowestCount {
			userWithLowestCount = user
			lowestCount = count
		}
	}
	reviewCounter[userWithLowestCount.GetID()]++
	return userWithLowestCount
}

// getNextTeacherReviewer gets the teacher with the least total reviews.
func getNextTeacherReviewer(db database.Database, course *pb.Course) (*pb.User, error) {
	teachers, err := db.GetCourseTeachers(course)
	if err != nil {
		return nil, fmt.Errorf("failed to get teachers from database: %w", err)
	}
	teacherReviewCounter.initialize(course.GetID())
	teacherReviewer := getNextReviewer(teachers, teacherReviewCounter[course.GetID()])
	return teacherReviewer, nil
}

// getNextStudentReviewer gets the student in a group with the least total reviews.
func getNextStudentReviewer(db database.Database, groupID, ownerID uint64) (*pb.User, error) {
	group, err := db.GetGroup(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group from database: %w", err)
	}
	groupReviewCounter.initialize(group.GetID())
	// We exclude the PR owner from the search.
	studentReviewer := getNextReviewer(group.GetUsersExcept(ownerID), groupReviewCounter[group.GetID()])
	return studentReviewer, nil
}
