package assignments

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/score"
	pb "github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetNextReviewer(t *testing.T) {
	// We create local versions of the maps
	teacherReviewCounter := make(countMap)
	groupReviewCounter := make(countMap)
	IDs := []uint64{1, 2, 3, 4}
	teachers := []*pb.User{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}
	students := []*pb.User{{ID: 1}, {ID: 2}, {ID: 3}}
	for _, ID := range IDs {
		for i := 0; i < len(teachers)*5; i++ {
			teacherReviewCounter.initialize(ID)
			gotTeacher := getNextReviewer(teachers, teacherReviewCounter[ID])
			wantTeacher := teachers[i%len(teachers)]
			if diff := cmp.Diff(wantTeacher, gotTeacher, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantTeacher, +gotTeacher):\n%s", diff)
			}
		}

		// Adding a new teacher.
		// Teacher is expected to be picked as reviewer len(teachers)-1 times.
		wantTeacher := &pb.User{ID: 6}
		teachers = append(teachers, wantTeacher)
		for i := 0; i < len(teachers)-1; i++ {
			teacherReviewCounter.initialize(ID)
			gotTeacher := getNextReviewer(teachers, teacherReviewCounter[ID])
			if diff := cmp.Diff(wantTeacher, gotTeacher, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantTeacher, +gotTeacher):\n%s", diff)
			}
		}
		teachers = teachers[:len(teachers)-1]

		for i := 0; i < len(students)*3; i++ {
			groupReviewCounter.initialize(ID)
			gotStudent := getNextReviewer(students, groupReviewCounter[ID])
			wantStudent := students[i%len(students)]
			if diff := cmp.Diff(wantStudent, gotStudent, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantStudent, +gotStudent):\n%s", diff)
			}
		}

		// Adding a new student
		// Student is expected to be picked as reviewer len(student)-1 times.
		wantStudent := &pb.User{ID: 4}
		students = append(students, wantStudent)
		for i := 0; i < len(students)-1; i++ {
			groupReviewCounter.initialize(ID)
			gotStudent := getNextReviewer(students, groupReviewCounter[ID])
			if diff := cmp.Diff(wantStudent, gotStudent, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantStudent, +gotStudent):\n%s", diff)
			}
		}
		students = students[:len(students)-1]
	}
}

// TestPublishFeedbackComment tests creating a feedback comment on a pull request, with the given result.
func TestPublishFeedbackComment(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	qfTestUser := scm.GetTestUser(t)
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	results := &score.Results{
		Scores: []*score.Score{
			{TestName: "Test1", TaskName: "1", Score: 5, MaxScore: 7, Weight: 2},
			{TestName: "Test2", TaskName: "1", Score: 3, MaxScore: 9, Weight: 3},
			{TestName: "Test3", TaskName: "1", Score: 8, MaxScore: 8, Weight: 5},
			{TestName: "Test4", TaskName: "1", Score: 2, MaxScore: 5, Weight: 1},
			{TestName: "Test5", TaskName: "1", Score: 5, MaxScore: 7, Weight: 1},
			{TestName: "Test6", TaskName: "2", Score: 5, MaxScore: 7, Weight: 1},
			{TestName: "Test7", TaskName: "3", Score: 5, MaxScore: 7, Weight: 1},
		},
	}
	body := results.MarkdownComment("1", 80)

	// To use this test, the CommentID have to be set manually.
	// TODO(meling) Fix test to CreateIssue and CreateIssueComment to get commentID
	opt := &scm.IssueCommentOptions{
		Organization: qfTestOrg,
		Repository:   pb.StudentRepoName(qfTestUser),
		Body:         body,
		CommentID:    0,
	}
	if err := s.UpdateIssueComment(context.Background(), opt); err != nil {
		t.Fatal(err)
	}
}
