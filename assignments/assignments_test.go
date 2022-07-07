package assignments

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/types"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
)

// To run this test, please see instructions in the developer guide (dev.md).

func TestFetchAssignments(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &types.Course{
		Name:             "QuickFeed Test Course",
		OrganizationPath: qfTestOrg,
	}

	assignments, _, err := fetchAssignments(context.Background(), zap.NewNop().Sugar(), s, course)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	for _, assignment := range assignments {
		t.Logf("assignment: %v", assignment)
	}
}

// TestUpdateCriteria simulates the behavior of UpdateFromTestsRepo
// where we update the criteria for an assignment.
// Benchmarks and criteria specifically related to a review should not be affected by UpdateFromTestsRepo.
// Neither should reviews
func TestUpdateCriteria(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &types.Course{}
	admin := qtest.CreateFakeUser(t, db, 10)
	user := qtest.CreateFakeUser(t, db, 20)
	qtest.CreateCourse(t, db, admin, course)

	// Assignment that will be updated
	assignment := &types.Assignment{
		CourseID:    course.ID,
		Name:        "Assignment 1",
		ScriptFile:  "go.sh",
		Deadline:    "12.12.2021",
		AutoApprove: false,
		Order:       1,
		IsGroupLab:  false,
	}

	assignment2 := &types.Assignment{
		CourseID:    course.ID,
		Name:        "Assignment 2",
		ScriptFile:  "go.sh",
		Deadline:    "12.01.2022",
		AutoApprove: false,
		Order:       2,
		IsGroupLab:  false,
	}

	for _, a := range []*types.Assignment{assignment, assignment2} {
		if err := db.CreateAssignment(a); err != nil {
			t.Fatal(err)
		}
	}

	benchmarks := []*types.GradingBenchmark{
		{
			ID:           1,
			AssignmentID: assignment.ID,
			Heading:      "Test benchmark 1",
			Criteria: []*types.GradingCriterion{
				{
					Description: "Criterion 1",
					BenchmarkID: 1,
					Points:      5,
				},
				{
					Description: "Criterion 2",
					BenchmarkID: 1,
					Points:      10,
				},
			},
		},
		{
			ID:           2,
			AssignmentID: assignment.ID,
			Heading:      "Test benchmark 2",
			Criteria: []*types.GradingCriterion{
				{
					Description: "Criterion 3",
					BenchmarkID: 2,
					Points:      1,
				},
			},
		},
	}

	benchmarks2 := []*types.GradingBenchmark{
		{
			ID:           3,
			AssignmentID: assignment2.ID,
			Heading:      "Test benchmark 3",
			Criteria: []*types.GradingCriterion{
				{
					Description: "Criterion 4",
					BenchmarkID: 3,
					Points:      2,
				},
			},
		},
	}

	for _, bms := range [][]*types.GradingBenchmark{benchmarks, benchmarks2} {
		for _, bm := range bms {
			if err := db.CreateBenchmark(bm); err != nil {
				t.Fatal(err)
			}
		}
	}

	assignment.GradingBenchmarks = benchmarks

	submission := &types.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
	}

	submission2 := &types.Submission{
		AssignmentID: assignment2.ID,
		UserID:       admin.ID,
	}

	for _, s := range []*types.Submission{submission, submission2} {
		if err := db.CreateSubmission(s); err != nil {
			t.Fatal(err)
		}
	}

	// Review for assignment that will be updated
	review := &types.Review{
		ReviewerID:   admin.ID,
		SubmissionID: submission.ID,
		GradingBenchmarks: []*types.GradingBenchmark{
			{
				AssignmentID: assignment.ID,
				Heading:      "Test benchmark 2",
				Comment:      "This is a comment",
				Criteria: []*types.GradingCriterion{
					{
						Description: "Criterion 3",
						Comment:     "This is a comment",
						Grade:       types.GradingCriterion_PASSED,
						BenchmarkID: 2,
						Points:      1,
					},
				},
			},
		},
	}

	// Review for assignment that will *not* be updated
	review2 := &types.Review{
		ReviewerID:   user.ID,
		SubmissionID: submission2.ID,
		GradingBenchmarks: []*types.GradingBenchmark{
			{
				AssignmentID: assignment2.ID,
				Heading:      "Test benchmark 2",
				Comment:      "This is another comment",
				Criteria: []*types.GradingCriterion{
					{
						Description: "Criterion 3",
						Comment:     "This is another comment",
						Grade:       types.GradingCriterion_PASSED,
						BenchmarkID: 3,
						Points:      1,
					},
				},
			},
		},
	}

	for _, r := range []*types.Review{review, review2} {
		if err := db.CreateReview(r); err != nil {
			t.Fatal(err)
		}
	}

	// If assignment.GradingBenchmarks is empty beyond this point, it means that there were no added / removed benchmarks / criteria
	updateGradingCriteria(zap.NewNop().Sugar(), db, assignment)

	// Assignment has no added or removed benchmarks, expect nil
	if assignment.GradingBenchmarks != nil {
		t.Fatalf("Expected assignment.GradingBenchmarks to be nil, got %v", assignment.GradingBenchmarks)
	}

	// Update assignments. GradingBenchmarks should not be updated
	db.UpdateAssignments([]*types.Assignment{assignment, assignment2})

	for _, wantReview := range []*types.Review{review, review2} {
		gotReview, err := db.GetReview(&types.Review{ID: wantReview.ID})
		if err != nil {
			t.Fatal(err)
		}
		// Review should not have changed
		if diff := cmp.Diff(wantReview, gotReview, protocmp.Transform()); diff != "" {
			t.Fatalf("GetReview() mismatch (-want +got):\n%s", diff)
		}
	}

	gotBenchmarks, err := db.GetBenchmarks(&types.Assignment{ID: assignment.ID, CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(benchmarks, gotBenchmarks, cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(&types.GradingBenchmark{}, "ID", "AssignmentID", "ReviewID"),
		protocmp.IgnoreFields(&types.GradingCriterion{}, "ID", "BenchmarkID"),
		protocmp.IgnoreEnums(),
	}); diff != "" {
		t.Errorf("GetBenchmarks() mismatch (-want +got):\n%s", diff)
	}

	updatedBenchmarks := []*types.GradingBenchmark{
		{
			ID:           1,
			AssignmentID: assignment.ID,
			Heading:      "Test benchmark 1",
			Criteria: []*types.GradingCriterion{
				{
					Description: "Criterion 1",
					BenchmarkID: 1,
					Points:      5,
				},
			},
		},
	}

	assignment.GradingBenchmarks = updatedBenchmarks

	// This should delete the old benchmarks and criteria existing in the database, and return the new benchmarks
	updateGradingCriteria(zap.NewNop().Sugar(), db, assignment)

	gotBenchmarks, err = db.GetBenchmarks(&types.Assignment{ID: assignment.ID, CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}
	// updateGradingCriteria should have deleted the old benchmarks and criteria
	if len(gotBenchmarks) > 0 {
		t.Fatalf("Expected no benchmarks, got %v", gotBenchmarks)
	}

	// Assignment has been modified, expect benchmarks to not be nil
	if assignment.GradingBenchmarks == nil {
		t.Fatal("Expected assignment.GradingBenchmarks to not be nil")
	}

	// Update assignments. GradingBenchmarks should be updated
	db.UpdateAssignments([]*types.Assignment{assignment, assignment2})

	// Benchmarks should have been updated to reflect the removal of a benchmark and a criterion
	gotBenchmarks, err = db.GetBenchmarks(&types.Assignment{ID: assignment.ID, CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(updatedBenchmarks, gotBenchmarks, protocmp.Transform()); diff != "" {
		t.Errorf("GetBenchmarks() mismatch (-want +got):\n%s", diff)
	}

	// Finally check that reviews are unaffected
	for _, wantReview := range []*types.Review{review, review2} {
		gotReview, err := db.GetReview(&types.Review{ID: wantReview.ID})
		if err != nil {
			t.Fatal(err)
		}
		// Review should not have changed
		if diff := cmp.Diff(wantReview, gotReview, protocmp.Transform()); diff != "" {
			t.Fatalf("GetReview() mismatch (-want +got):\n%s", diff)
		}
	}
}
