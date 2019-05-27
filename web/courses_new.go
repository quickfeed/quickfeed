package web

import (
	"context"
	"fmt"
	"log"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	private = true
	public  = !private
)

// RepoPaths maps from Autograder repository path names to a boolean indicating
// whether or not the repository should be create as public or private.
var RepoPaths = map[string]bool{
	InfoRepo:       public,
	AssignmentRepo: private,
	TestsRepo:      private,
	SolutionsRepo:  private,
}

var repoNames = fmt.Sprintf("(%s, %s, %s, %s)", InfoRepo, AssignmentRepo, TestsRepo, SolutionsRepo)

// NewCourse creates a new course for the directory specified in the request
// and creates the repositories for the course. Requires that the directory
// does not contain the Autograder repositories that will be created.
//TODO(meling) should have proper logging in these funcs, especially for errors.
func NewCourse(ctx context.Context, request *pb.Course, db database.Database, s scm.SCM, bh BaseHookOptions) (*pb.Course, error) {
	if !validCourse(request) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}

	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	directory, err := s.GetDirectory(ctx, request.DirectoryId)
	if err != nil {
		return nil, err
	}
	repos, err := s.GetRepositories(ctx, directory)
	if err != nil {
		return nil, err
	}
	if isDirty(repos) {
		return nil, status.Errorf(codes.AlreadyExists,
			"%s contains one or more Autograder repositories %s", directory.GetPath(), repoNames)
	}

	for path, private := range RepoPaths {
		repoOptions := &scm.CreateRepositoryOptions{
			Path:      path,
			Directory: directory,
			Private:   private,
		}
		repo, err := s.CreateRepository(ctx, repoOptions)
		if err != nil {
			log.Println("NewCourse: failed to create repository:", path)
			return nil, err
		}
		log.Println("Created repository:", path)

		hookOptions := &scm.CreateHookOptions{
			URL:        GetEventsURL(bh.BaseURL, request.Provider),
			Secret:     bh.Secret,
			Repository: repo,
		}
		if err := s.CreateHook(ctx, hookOptions); err != nil {
			log.Println("NewCourse: Failed to create webhook for repository:", path)
			return nil, err
		}
		log.Println("Created webhook for repository:", path)

		dbRepo := pb.Repository{
			DirectoryId:  directory.Id,
			RepositoryId: repo.ID,
			HtmlUrl:      repo.WebURL,
			RepoType:     repoType(path),
		}
		if err := db.CreateRepository(&dbRepo); err != nil {
			return nil, err
		}
	}

	if err := db.CreateCourse(request.GetCoursecreatorId(), request); err != nil {
		return nil, err
	}
	return request, nil
}

// isDirty returns true if the list of provided repositories contains
// any of the repositories that Autograder wants to create.
func isDirty(repos []*scm.Repository) bool {
	if len(repos) == 0 {
		return false
	}
	for _, repo := range repos {
		if _, exists := RepoPaths[repo.Path]; exists {
			return true
		}
	}
	return false
}
