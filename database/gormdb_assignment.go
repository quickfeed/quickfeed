package database

import (
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/jinzhu/gorm"
)

/// Assignments ///

// CreateAssignment creates a new assignment record.
func (db *GormDB) CreateAssignment(assignment *pb.Assignment) error {
	// Course id and assignment order must be given.
	if assignment.CourseID < 1 || assignment.Order < 1 {
		return gorm.ErrRecordNotFound
	}

	var course uint64
	if err := db.conn.Model(&pb.Course{}).Where(&pb.Course{
		ID: assignment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	return db.conn.
		Where(pb.Assignment{
			CourseID: assignment.CourseID,
			Order:    assignment.Order,
		}).
		Assign(map[string]interface{}{
			"name":              assignment.Name,
			"order":             assignment.Order,
			"script_file":       assignment.ScriptFile,
			"deadline":          assignment.Deadline,
			"auto_approve":      assignment.AutoApprove,
			"score_limit":       assignment.ScoreLimit,
			"is_group_lab":      assignment.IsGroupLab,
			"reviewers":         assignment.Reviewers,
			"container_timeout": assignment.ContainerTimeout,
		}).FirstOrCreate(assignment).Error
}

// GetAssignment returns assignment with the given ID.
func (db *GormDB) GetAssignment(query *pb.Assignment) (*pb.Assignment, error) {
	var assignment pb.Assignment
	if err := db.conn.Where(query).Preload("GradingBenchmarks").Preload("GradingBenchmarks.GradingCriteria").First(&assignment).Error; err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetAssignmentsByCourse fetches all assignments for the given course ID.
func (db *GormDB) GetAssignmentsByCourse(courseID uint64, withGrading bool) ([]*pb.Assignment, error) {
	var course pb.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}
	assignments := course.Assignments
	if withGrading {
		for _, a := range assignments {
			var benchmarks []*pb.GradingBenchmark
			if err := db.conn.Where("assignment_id = ?", a.ID).Find(&benchmarks).Error; err != nil {
				return nil, err
			}
			a.GradingBenchmarks = benchmarks
			for _, b := range a.GradingBenchmarks {
				var criteria []*pb.GradingCriterion
				if err := db.conn.Where("benchmark_id = ?", b.ID).Find(&criteria).Error; err != nil {
					return nil, err
				}
				b.Criteria = criteria
			}
		}
	}
	return assignments, nil
}

// UpdateAssignments updates assignment information.
func (db *GormDB) UpdateAssignments(assignments []*pb.Assignment) error {
	//TODO(meling) Updating the database may need locking?? Or maybe rewrite as a single query or txn.
	for _, v := range assignments {
		// this will create or update an existing assignment
		if err := db.CreateAssignment(v); err != nil {
			return err
		}
	}
	return nil
}

// GetCourseAssignmentsWithSubmissions returns all course assignments
// of requested type with preloaded submissions.
func (db *GormDB) GetCourseAssignmentsWithSubmissions(courseID uint64, submissionType pb.SubmissionsForCourseRequest_Type) ([]*pb.Assignment, error) {
	var assignments []*pb.Assignment

	if err := db.conn.Preload("Submissions").Preload("Submissions.Reviews").Where(&pb.Assignment{CourseID: courseID}).Order("order").Find(&assignments).Error; err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if submissionType == pb.SubmissionsForCourseRequest_ALL {
		return assignments, nil
	}

	wantGroupLabs := submissionType == pb.SubmissionsForCourseRequest_GROUP
	filteredAssignments := make([]*pb.Assignment, 0)
	for _, a := range assignments {
		if a.IsGroupLab == wantGroupLabs {
			filteredAssignments = append(filteredAssignments, a)
		}
	}
	return filteredAssignments, nil
}

// CreateBenchmark creates a new grading benchmark
func (db *GormDB) CreateBenchmark(query *pb.GradingBenchmark) error {
	return db.conn.Create(query).Error
}

// UpdateBenchmark updates the given benchmark
func (db *GormDB) UpdateBenchmark(query *pb.GradingBenchmark) error {
	return db.conn.Model(query).
		Where(&pb.GradingBenchmark{ID: query.ID, AssignmentID: query.AssignmentID}).
		Update(&pb.GradingBenchmark{Heading: query.Heading, Comment: query.Comment}).Error
}

// DeleteBenchmark removes the given benchmark
func (db *GormDB) DeleteBenchmark(query *pb.GradingBenchmark) error {
	db.conn.Where("benchmark_id = ?", query.GetID()).Delete(&pb.GradingCriterion{})
	return db.conn.Delete(query).Error
}

// CreateCriterion creates a new grading criterion
func (db *GormDB) CreateCriterion(query *pb.GradingCriterion) error {
	return db.conn.Create(query).Error
}

// UpdateCriterion updates the given criterion
func (db *GormDB) UpdateCriterion(query *pb.GradingCriterion) error {
	return db.conn.Model(query).
		Where(&pb.GradingCriterion{ID: query.ID, BenchmarkID: query.BenchmarkID}).
		Update(&pb.GradingCriterion{Description: query.Description, Comment: query.Comment, Grade: query.Grade, Score: query.Score}).Error
}

// DeleteCriterion removes the given criterion
func (db *GormDB) DeleteCriterion(query *pb.GradingCriterion) error {
	return db.conn.Delete(query).Error
}
