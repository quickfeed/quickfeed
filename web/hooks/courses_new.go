package hooks

import (
	"context"
	"fmt"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// createCourse creates a new course for the directory specified in the request
// and creates the repositories for the course. Requires that the directory
// does not contain the QuickFeed repositories that will be created.
func CreateCourse(ctx context.Context, db database.Database, sc scm.SCM, course *qf.Course, courseCreator *qf.User) (*qf.Course, error) {
	repos, err := sc.CreateCourse(ctx, &scm.CourseOptions{
		CourseCreator:  courseCreator.Login,
		OrganizationID: course.ScmOrganizationID,
	})
	if err != nil {
		return nil, err
	}
	for _, repo := range repos {
		dbRepo := qf.Repository{
			ScmOrganizationID: course.ScmOrganizationID,
			ScmRepositoryID:   repo.ID,
			HTMLURL:           repo.HTMLURL,
			RepoType:          qf.RepoType(repo.Repo),
		}
		if dbRepo.IsUserRepo() {
			dbRepo.UserID = courseCreator.ID
		}
		if err := db.CreateRepository(&dbRepo); err != nil {
			return nil, fmt.Errorf("failed to create database record for repository %s: %w", repo.Repo, err)
		}
	}
	// make sure to set as course creator
	course.CourseCreatorID = courseCreator.ID
	if err := db.CreateCourse(course.GetCourseCreatorID(), course); err != nil {
		return nil, fmt.Errorf("failed to create database record for course %s: %w", course.Name, err)
	}
	return course, nil
}
