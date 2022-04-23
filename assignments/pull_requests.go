package assignments

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
)

var (
	// These are used to track how many times someone has been assigned to review a pull request. They map as follows.
	// teacherReviewCounter[courseID][userID] = count
	// groupReviewCounter[groupID][userID] = count
	teacherReviewCounter = make(map[uint64]map[uint64]int)
	groupReviewCounter   = make(map[uint64]map[uint64]int)
)

func AssignReviewers(sc scm.SCM, db database.Database, course *pb.Course, repo *pb.Repository, pullRequest *pb.PullRequest) error {
	teacherReviewer, err := getNextTeacherReviewer(db, course)
	if err != nil {
		return err
	}
	studentReviewer, err := getNextStudentReviewer(db, repo, pullRequest.GetUserID())
	if err != nil {
		return err
	}

	reviewers := []string{}
	reviewers = append(reviewers, teacherReviewer.GetLogin())
	reviewers = append(reviewers, studentReviewer.GetLogin())

	opt := &scm.RequestReviewersOptions{
		Organization: course.GetOrganizationPath(),
		Repository:   repo.Name(),
		Number:       int(pullRequest.GetNumber()),
		Reviewers:    reviewers,
	}

	ctx := context.Background()
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
//
// Returns an error if the list of users is empty
func getNextReviewer(ID uint64, users []*pb.User, reviewCounter map[uint64]map[uint64]int) (*pb.User, error) {
	if len(users) == 0 {
		return nil, errors.New("list of users is empty")
	}
	reviewerMap, ok := reviewCounter[ID]
	if !ok {
		// If a map does not exist for a course we create it,
		// and assign the first user as the reviewer.
		reviewCounter[ID] = make(map[uint64]int)
		reviewCounter[ID][users[0].GetID()] = 1
		return users[0], nil
	}
	userWithLowestCount := users[0]
	lowestCount := reviewerMap[users[0].GetID()]
	for _, user := range users {
		count, ok := reviewerMap[user.GetID()]
		if !ok {
			// If the user is not present in the review map,
			// then they are assigned as the next reviewer.
			reviewerMap[user.GetID()] = 1
			return user, nil
		}
		if count < lowestCount {
			userWithLowestCount = user
			lowestCount = count
		}
	}
	reviewerMap[userWithLowestCount.GetID()]++
	return userWithLowestCount, nil
}

// getNextTeacherReviewer gets the teacher with the least total reviews.
func getNextTeacherReviewer(db database.Database, course *pb.Course) (*pb.User, error) {
	teachers, err := db.GetCourseTeachers(course)
	if err != nil {
		return nil, err
	}
	teacherReviewer, err := getNextReviewer(course.GetID(), teachers, teacherReviewCounter)
	if err != nil {
		return nil, fmt.Errorf("failed to get next teacher reviewer: %w", err)
	}
	return teacherReviewer, nil
}

// getNextStudentReviewer gets the student with the least total reviews.
func getNextStudentReviewer(db database.Database, repo *pb.Repository, ownerID uint64) (*pb.User, error) {
	group, err := db.GetGroup(repo.GetGroupID())
	if err != nil {
		return nil, err
	}
	if len(group.Users) == 0 {
		// This should never happen.
		return nil, errors.New("failed to get next teacher reviewer: no users in group")
	}
	// We exclude the PR owner from the search.
	studentReviewer, err := getNextReviewer(group.GetID(), group.GetUserSubset(ownerID), groupReviewCounter)
	if err != nil {
		return nil, fmt.Errorf("failed to get next teacher reviewer: %w", err)
	}
	return studentReviewer, nil
}
