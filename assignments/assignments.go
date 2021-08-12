package assignments

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

// UpdateFromTestsRepo updates the database record for the course assignments.
func UpdateFromTestsRepo(logger *zap.SugaredLogger, runner ci.Runner, db database.Database, repo *pb.Repository, course *pb.Course) {
	logger.Debugf("Updating %s from '%s' repository", course.GetCode(), pb.TestsRepo)
	s, err := scm.NewSCMClient(logger, course.GetProvider(), course.GetAccessToken())
	if err != nil {
		logger.Errorf("Failed to create SCM Client: %v", err)
		return
	}
	assignments, dockerfile, err := FetchAssignments(context.Background(), s, course)
	if err != nil {
		logger.Errorf("Failed to fetch assignments from '%s' repository: %v", pb.TestsRepo, err)
		return
	}
	for _, assignment := range assignments {
		logger.Debugf("Found assignment in '%s' repository: %s", pb.TestsRepo, assignment.Name)
		updateGradingCriteria(logger, db, assignment)
	}
	if dockerfile != "" && dockerfile != course.Dockerfile {
		logger.Debugf("Saving Dockerfile for course %s", course.Code)
		course.Dockerfile = dockerfile
		if err := db.UpdateCourse(course); err != nil {
			logger.Debugf("Failed to update dockerfile for course %s: %s", course.Code, err)
			return
		}
	}

	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debugf("Failed to update database for: %v", assignment)
		}
		logger.Errorf("Failed to update assignments in database: %v", err)
		return
	}
	logger.Debugf("Assignments for %s successfully updated from '%s' repo", course.GetCode(), pb.TestsRepo)
}

// FetchAssignments returns a list of assignments for the given course, by
// cloning the 'tests' repo for the given course and extracting the assignments
// from the 'assignment.yml' files, one for each assignment.
//
// Note: This will typically be called on a push event to the 'tests' repo,
// which should happen infrequently. It may also be called manually by a
// teacher/admin from the frontend. However, even if multiple invocations
// happen concurrently, the function is idempotent. That is, it only reads
// data from GitHub, processes the yml files and returns the assignments.
// The TempDir() function ensures that cloning is done in distinct temp
// directories, should there be concurrent calls to this function.
func FetchAssignments(c context.Context, sc scm.SCM, course *pb.Course) ([]*pb.Assignment, string, error) {
	ctx, cancel := context.WithTimeout(c, pb.MaxWait)
	defer cancel()

	// ensuring compatibility with the old database:
	// TODO(meling) Check if this is still needed with the new database?
	if course.OrganizationPath == "" {
		org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.OrganizationID})
		if err != nil {
			return nil, "", err
		}
		course.OrganizationPath = org.GetPath()
	}

	log.Printf("org %s\n", course.GetOrganizationPath())

	cloneURL := sc.CreateCloneURL(&scm.URLPathOptions{
		Organization: course.OrganizationPath,
		Repository:   pb.TestsRepo,
	})
	log.Printf("cloneURL %v\n", cloneURL)

	cloneDir, err := ioutil.TempDir("", pb.TestsRepo)
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(cloneDir)

	// clone the tests repository to cloneDir
	job := &ci.Job{
		Commands: []string{
			"cd " + cloneDir,
			"git clone " + cloneURL,
		},
	}
	log.Printf("cd %v\n", cloneDir)
	log.Printf("git clone %v\n", cloneURL)

	runner := ci.Local{}
	_, err = runner.Run(ctx, job)
	if err != nil {
		return nil, "", err
	}

	// parse assignments found in the cloned tests directory
	assignments, dockerfile, err := parseAssignments(cloneDir, course.ID)
	if err != nil {
		return nil, "", err
	}

	// if a Dockerfile added/updated, build docker image locally
	// tag the image with the course code
	if dockerfile != "" && dockerfile != course.Dockerfile {
		job.Commands = []string{
			"cd " + cloneDir + "/tests/scripts",
			fmt.Sprintf("docker build -t %s .", course.Code),
		}

		if out, err := runner.Run(context.Background(), job); err != nil {
			log.Printf("Failed to build image from dockerfile for %s (%s): %s", course.Code, out, err)
		} else {
			log.Println("Built new image from course Dockerfile for ", course.Code)
		}
	}
	return assignments, dockerfile, nil
}

// updateGradingCriteria will remove old grading criteria and related reviews when criteria.json gets updated
func updateGradingCriteria(logger *zap.SugaredLogger, db database.Database, assignment *pb.Assignment) {
	if len(assignment.GetGradingBenchmarks()) > 0 {
		savedAssignment, err := db.GetAssignment(&pb.Assignment{
			CourseID: assignment.CourseID,
			Name:     assignment.Name,
		})
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// a new assignment, no actions required
				return
			}
			logger.Debugf("Failed to fetch assignment %s from database: %s", assignment.Name, err)
			return
		}
		if len(savedAssignment.GetGradingBenchmarks()) > 0 {
			if diff := cmp.Diff(assignment.GradingBenchmarks, savedAssignment.GradingBenchmarks, protocmp.Transform()); diff != "" {
				for _, bm := range assignment.GradingBenchmarks {
					for _, c := range bm.Criteria {
						if err := db.DeleteCriterion(c); err != nil {
							fmt.Printf("Failed to delete criteria %v: %s\n", c, err)
						}
					}
					if err := db.DeleteBenchmark(bm); err != nil {
						fmt.Printf("Failed to delete benchmark %v: %s\n", bm, err)
					}
				}
				submissions, err := db.GetSubmissions(&pb.Submission{AssignmentID: assignment.GetID()})
				if err != nil {
					logger.Debugf("No submissions for assignment %s: %s", assignment.Name, err)
					return
				}
				for _, submission := range submissions {
					if err := db.DeleteReview(&pb.Review{SubmissionID: submission.ID}); err != nil {
						logger.Debugf("Failed to delete reviews for submission %s to assignment %s: %s", submission.ID, assignment.Name, err)
					}
				}
			}
		}
		for _, bm := range assignment.GradingBenchmarks {
			bm.AssignmentID = assignment.ID
			if err := db.CreateBenchmark(bm); err != nil {
				logger.Errorf("Failed to save grading benchmark: %s", err)
				return
			}
			for _, c := range bm.Criteria {
				c.BenchmarkID = bm.ID
				if err := db.CreateCriterion(c); err != nil {
					logger.Errorf("Failed to save grading criterion: %s", err)
					return
				}
			}
		}
	}
}
