package assignments

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/scm"
)

var (
	// These are used to track how many times someone has been assigned to review a pull request. They map as follows.
	// teacherReviewCounter[courseID][userID] = count
	// groupReviewCounter[groupID][userID] = count
	teacherReviewCounter = make(countMap)
	groupReviewCounter   = make(countMap)
)

type countMap map[uint64]map[uint64]int

// Creates a new map if none exists.
func (m countMap) initialize(id uint64) {
	_, ok := m[id]
	if !ok {
		m[id] = make(map[uint64]int)
	}
}

// CreateFeedbackComment formats a feedback comment to be posted on pull requests.
// It uses the test results from a student commit to create a table like the one shown below.
// Only the test scores associated with the supplied task are used to generate this table.
// Table formatting ref: https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-tables
//
//  ## Test results from latest push
//	| Test Name | Score | Weight | % of Total |
//	| :-------- | :---- | :----- | ---------: |
//  | Test 1	| 2/4	| 1		 |	   6.25%  |
//  | Test 2	| 1/4   | 2	     |     6.25%  |
//  | Test 3	| 3/4   | 5      |     46.86% |
//  | Total		|		|		 |	   59.36% |
//
// 	Once a total score of 80% is reached, reviewers are automatically assigned.
//
func CreateFeedbackComment(results *score.Results, taskLocalName string, assignment *pb.Assignment) string {
	body := "## Test results from latest push\n\n" +
		"| Test Name | Score | Weight | % of Total |\n" +
		"| :-------- | :---- | :----- | ---------: |\n"

	for _, testScore := range results.Scores {
		if testScore.TaskName != taskLocalName {
			continue
		}
		percentageScore := score.CalculateWeightedScore(float64(testScore.Score), float64(testScore.MaxScore), float64(testScore.Weight), results.TotalTaskWeight(taskLocalName))
		body += fmt.Sprintf("| %s | %d/%d | %d | %.2f%% |\n", testScore.TestName, testScore.Score, testScore.MaxScore, testScore.Weight, percentageScore*100)
	}
	// TODO(espeland): TaskSum returns an int, while a float is used for individual tests
	body += fmt.Sprintf("| **Total** | | | **%d%%** |\n\n", results.TaskSum(taskLocalName))
	body += fmt.Sprintf("\nOnce a total score of %d%% is reached, reviewers are automatically assigned.\n", assignment.GetScoreLimit())
	return body
}

// AssignReviewers assigns reviewers to a group repository pull request.
// It assigns one other group member and one course teacher as reviewers.
func AssignReviewers(ctx context.Context, sc scm.SCM, db database.Database, course *pb.Course, repo *pb.Repository, pullRequest *pb.PullRequest) error {
	teacherReviewer, err := getNextTeacherReviewer(db, course)
	if err != nil {
		return err
	}
	// TODO(espeland): Remember to uncomment when finished testing
	// studentReviewer, err := getNextStudentReviewer(db, repo.GetGroupID(), pullRequest.GetUserID())
	// if err != nil {
	// 	return err
	// }

	reviewers := []string{
		teacherReviewer.GetLogin(),
		// studentReviewer.GetLogin(),
	}
	opt := &scm.RequestReviewersOptions{
		Organization: course.GetOrganizationPath(),
		Repository:   repo.Name(),
		Number:       int(pullRequest.GetNumber()),
		Reviewers:    reviewers,
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
	lowestCount := reviewCounter[users[0].GetID()]
	for _, user := range users {
		count, ok := reviewCounter[user.GetID()]
		if !ok {
			// If the user is not present in the review map
			// they are returned as the next reviewer.
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
	if len(group.Users) == 0 {
		// This should never happen.
		return nil, errors.New("failed to get next student reviewer: no users in group")
	}
	groupReviewCounter.initialize(group.GetID())
	// We exclude the PR owner from the search.
	studentReviewer := getNextReviewer(group.GetUsersExcept(ownerID), groupReviewCounter[group.GetID()])
	return studentReviewer, nil
}
