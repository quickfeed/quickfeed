package web

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/assignments"
	"github.com/autograde/quickfeed/scm"
)

var criteriaFile = "criteria.json"

// getAssignments lists the assignments for the provided course.
func (s *AutograderService) getAssignments(courseID uint64) (*pb.Assignments, error) {
	allAssignments, err := s.db.GetAssignmentsByCourse(courseID, true)
	if err != nil {
		return nil, err
	}
	// Hack to ensure that assignments stored in database with wrong format
	// is displayed correctly in the frontend. This should ideally be removed
	// when the database no longer contains any incorrectly formatted dates.
	for _, assignment := range allAssignments {
		assignment.Deadline = assignments.FixDeadline(assignment.GetDeadline())
	}
	return &pb.Assignments{Assignments: allAssignments}, nil
}

// updateAssignments updates the assignments for the given course.
func (s *AutograderService) updateAssignments(ctx context.Context, sc scm.SCM, courseID uint64) error {
	course, err := s.db.GetCourse(courseID, false)
	if err != nil {
		return err
	}
	assignments, err := assignments.FetchAssignments(ctx, sc, course)
	if err != nil {
		return err
	}
	if err = s.db.UpdateAssignments(assignments); err != nil {
		return err
	}
	return nil
}

func (s *AutograderService) createBenchmark(query *pb.GradingBenchmark) (*pb.GradingBenchmark, error) {
	if _, err := s.db.GetAssignment(&pb.Assignment{
		ID: query.AssignmentID,
	}); err != nil {
		return nil, err
	}
	if err := s.db.CreateBenchmark(query); err != nil {
		return nil, err
	}
	return query, nil
}

func (s *AutograderService) updateBenchmark(query *pb.GradingBenchmark) error {
	return s.db.UpdateBenchmark(query)
}

func (s *AutograderService) deleteBenchmark(query *pb.GradingBenchmark) error {
	return s.db.DeleteBenchmark(query)
}

func (s *AutograderService) createCriterion(query *pb.GradingCriterion) (*pb.GradingCriterion, error) {
	if err := s.db.CreateCriterion(query); err != nil {
		return nil, err
	}
	return query, nil
}

func (s *AutograderService) updateCriterion(query *pb.GradingCriterion) error {
	return s.db.UpdateCriterion(query)
}

func (s *AutograderService) deleteCriterion(query *pb.GradingCriterion) error {
	return s.db.DeleteCriterion(query)
}

func (s *AutograderService) loadCriteria(ctx context.Context, sc scm.SCM, request *pb.LoadCriteriaRequest) ([]*pb.GradingBenchmark, error) {

	// get assignment, check that exists
	assignment, err := s.db.GetAssignment(&pb.Assignment{ID: request.AssignmentID, CourseID: request.CourseID})
	if err != nil {
		return nil, err
	}

	course, err := s.db.GetCourse(request.CourseID, false)
	if err != nil {
		return nil, err
	}

	opts := &scm.FileOptions{
		Path:       filepath.Join(assignment.GetName(), criteriaFile),
		Owner:      course.OrganizationPath,
		Repository: pb.TestsRepo,
	}

	criteriaString, err := sc.GetFileContent(ctx, opts)
	if err != nil || criteriaString == "" {
		return nil, err
	}
	fmt.Printf("Read file content for options: %+v/n", opts)
	fmt.Println("File content is: ", criteriaString)

	// unmarshall, log success
	var benchmarks []*pb.GradingBenchmark
	if err := json.Unmarshal([]byte(criteriaString), &benchmarks); err != nil {
		return nil, err
	}

	fmt.Printf("Unmarshalled %d benchmarks from file\n", len(benchmarks))

	if len(assignment.GradingBenchmarks) > 0 {
		for _, bm := range assignment.GradingBenchmarks {
			for _, c := range bm.Criteria {
				if err := s.db.DeleteCriterion(c); err != nil {
					fmt.Printf("Failed to delete criteria %v: %s\n", c, err)
				}
			}
			if err := s.db.DeleteBenchmark(bm); err != nil {
				fmt.Printf("Failed to delete benchmark %v: %s\n", bm, err)
			}
		}
	}

	for _, bm := range benchmarks {
		bm.AssignmentID = assignment.ID
		if err := s.db.CreateBenchmark(bm); err != nil {
			return nil, err
		}
		for _, c := range bm.Criteria {
			c.BenchmarkID = bm.ID
			if err := s.db.CreateCriterion(c); err != nil {
				return nil, err
			}
		}
	}

	return benchmarks, nil
}

func (s *AutograderService) createReview(query *pb.Review) (*pb.Review, error) {
	if _, err := s.db.GetSubmission(&pb.Submission{ID: query.SubmissionID}); err != nil {
		return nil, err
	}
	if err := s.db.CreateReview(query); err != nil {
		return nil, err
	}
	return query, nil
}

func (s *AutograderService) updateReview(query *pb.Review) error {
	return s.db.UpdateReview(query)
}
