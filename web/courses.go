package web

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/yamlparser"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 10 * time.Minute

// BaseHookOptions contains options shared among all webhooks.
type BaseHookOptions struct {
	BaseURL string
	// Secret is used to verify that the event received is legit. GitHub
	// sends back a signature of the payload, while GitLab just sends back
	// the secret. This is all handled by the
	// gopkg.in/go-playground/webhooks.v3 package.
	Secret string
}

func validCourse(c *pb.Course) bool {
	return c != nil &&
		c.Name != "" &&
		c.Code != "" &&
		(c.Provider == "github" || c.Provider == "gitlab" || c.Provider == "fake") &&
		c.DirectoryId != 0 &&
		c.Year != 0 &&
		c.Tag != ""
}

// EnrollUserRequest represent a request for enrolling a user to a course.
type EnrollUserRequest struct {
	Status uint `json:"status"`
}

func (eur *EnrollUserRequest) valid() bool {
	return eur.Status <= models.Teacher
}

// ListCourses returns a JSON object containing all the courses in the database.
func ListCourses(db database.Database) (*pb.Courses, error) {
	courses, err := db.GetCourses()
	if err != nil {
		return nil, err
	}
	return &pb.Courses{Courses: courses}, nil
}

// ListCoursesWithEnrollment lists all existing courses with the provided users
// enrollment status.
// If status query param is provided, lists only courses of the student filtered by the query param.
func ListCoursesWithEnrollment(request *pb.RecordRequest, db database.Database) (*pb.Courses, error) {
	courses, err := db.GetCoursesByUser(request.Id, request.Statuses...)
	if err != nil {
		return nil, err
	}
	return &pb.Courses{Courses: courses}, nil
}

// ListAssignments lists the assignments for the provided course.
func ListAssignments(request *pb.RecordRequest, db database.Database) (*pb.Assignments, error) {
	assignments, err := db.GetAssignmentsByCourse(request.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Assignments{Assignments: assignments}, nil
}

// NewCourse creates a new course and associates it with a directory (organization in github)
// and creates the repositories for the course.
func NewCourse(ctx context.Context, request *pb.Course, db database.Database, s scm.SCM, bh BaseHookOptions, currentUser *pb.User) (*pb.Course, error) {
	if !validCourse(request) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}

	contextWithTimeout, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	directory, err := s.GetDirectory(contextWithTimeout, request.DirectoryId)
	if err != nil {
		return nil, err
	}

	repos, err := s.GetRepositories(contextWithTimeout, directory)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]*scm.Repository)
	for _, repo := range repos {
		existing[repo.Path] = repo
	}

	var paths = []string{InfoRepo, AssignmentRepo, TestsRepo, SolutionsRepo}
	for _, path := range paths {

		var repo *scm.Repository
		var ok bool
		if repo, ok = existing[path]; !ok {
			privRepo := false
			if path == TestsRepo {
				privRepo = true
			}
			var err error
			repo, err = s.CreateRepository(
				ctx,
				&scm.CreateRepositoryOptions{
					Path:      path,
					Directory: directory,
					Private:   privRepo},
			)

			if err != nil {
				log.Println("NewCourse: failed to create repository")
				return nil, err
			}
			log.Println("Created new repository")
		}
		hooks, err := s.ListHooks(contextWithTimeout, repo)
		if err != nil {
			log.Println("Failed to list hooks for repository")
			return nil, err
		}

		hasAGWebHook := false
		for _, hook := range hooks {
			log.Println("Hook for repository: ", hook.ID, " ", hook.Name, " ", hook.URL)
			// TODO this check is specific for the github implementation ; fix this
			if hook.Name == "web" {
				hasAGWebHook = true
				break
			}
		}
		if !hasAGWebHook {
			if err := s.CreateHook(contextWithTimeout, &scm.CreateHookOptions{
				URL:        GetEventsURL(bh.BaseURL, request.Provider),
				Secret:     bh.Secret,
				Repository: repo,
			}); err != nil {
				log.Println("Failed to create webhook for repository: ", path)
				return nil, err
			}
			log.Println("Created new webhook for repository: ", path)
		}
		var repoType pb.Repository_RepoType
		switch path {
		case InfoRepo:
			repoType = pb.Repository_COURSEINFO
		case AssignmentRepo:
			repoType = pb.Repository_ASSIGNMENT
		case TestsRepo:
			repoType = pb.Repository_TESTS
		case SolutionsRepo:
			repoType = pb.Repository_SOLUTION
		}

		dbRepo := pb.Repository{
			DirectoryId:  directory.Id,
			RepositoryId: repo.ID,
			HtmlUrl:      repo.WebURL,
			RepoType:     repoType,
		}
		if err := db.CreateRepository(&dbRepo); err != nil {
			return nil, err
		}
	}

	request.CoursecreatorId = currentUser.Id
	request.DirectoryId = directory.Id

	if err := db.CreateCourse(currentUser.Id, request); err != nil {
		//TODO(meling) Should we even communicate bad request to the client?
		// We should log errors and debug it on the server side instead.
		// If clients make mistakes, there is nothing it can do with the
		return nil, err
	}
	return request, nil

}

// CreateEnrollment enrolls a user in a course.
func CreateEnrollment(request *pb.ActionRequest, db database.Database) (*pb.StatusCode, error) {

	if request.Status > pb.Enrollment_TEACHER || request.UserId == 0 || request.CourseId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid payload")
	}

	enrollment := pb.Enrollment{
		UserId:   request.UserId,
		CourseId: request.CourseId,
	}

	log.Println(enrollment.UserId)
	if err := db.CreateEnrollment(&enrollment); err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "Record not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	return &pb.StatusCode{StatusCode: int32(codes.OK)}, nil
}

// UpdateEnrollment accepts or rejects a user to enroll in a course.
func UpdateEnrollment(ctx context.Context, request *pb.ActionRequest, db database.Database, s scm.SCM, currentUser *pb.User) (*pb.StatusCode, error) {

	if request.Status > pb.Enrollment_TEACHER || request.UserId == 0 || request.CourseId == 0 {
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, "invalid payload")
	}

	if _, err := db.GetEnrollmentByCourseAndUser(request.CourseId, request.UserId); err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	if !currentUser.IsAdmin {
		return &pb.StatusCode{StatusCode: int32(codes.PermissionDenied)}, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	// TODO If the enrollment is accepted, create repositories and permissions for them with webooks.
	var err error
	switch request.Status {
	case pb.Enrollment_STUDENT:
		// Update enrollment for student in DB.
		err = db.EnrollStudent(request.UserId, request.CourseId)
		if err != nil {
			return nil, err
		}
		course, err := db.GetCourse(request.CourseId)
		if err != nil {
			return nil, err
		}

		contextWithTimeout, cancel := context.WithTimeout(ctx, MaxWait)
		defer cancel()

		dir, err := s.GetDirectory(contextWithTimeout, course.DirectoryId)
		if err != nil {
			return nil, err
		}
		student, err := db.GetUser(request.UserId)
		if err != nil {
			return nil, err
		}
		// Find out what the current plan is to set repo/team as private if not(?) do not create the repo
		// TODO Decide which provider/remoteIdentity is being used,
		gitUserName, err := s.GetUserNameByID(contextWithTimeout, student.RemoteIdentities[0].RemoteId)
		if err != nil {
			return nil, err
		}
		// Creating repository
		studentName := strings.Replace(gitUserName, " ", "", -1)
		pathName := studentName + "-labs"
		repo, err := s.CreateRepository(contextWithTimeout, &scm.CreateRepositoryOptions{
			Directory: dir,
			Path:      pathName,
			Private:   true,
		})
		if err != nil {
			return nil, err
		}
		dbRepo := pb.Repository{
			DirectoryId:  course.DirectoryId,
			RepositoryId: repo.ID,
			HtmlUrl:      repo.WebURL,
			RepoType:     pb.Repository_USER,
			UserId:       request.UserId,
		}
		if err := db.CreateRepository(&dbRepo); err != nil {
			return nil, err
		}
		// Creating team
		team, err := s.CreateTeam(contextWithTimeout, &scm.CreateTeamOptions{
			Directory: &scm.Directory{Path: dir.Path},
			TeamName:  studentName,
			Users:     []string{gitUserName},
		})
		if err != nil {
			return nil, err
		}
		if team != nil {
			err = s.AddTeamRepo(contextWithTimeout, &scm.AddTeamRepoOptions{
				TeamID: team.ID,
				Owner:  repo.Owner,
				Repo:   repo.Path,
			})
			if err != nil {
				return nil, err
			}
		}
	case pb.Enrollment_TEACHER:
		err = db.EnrollTeacher(request.UserId, request.CourseId)
	case pb.Enrollment_REJECTED:
		err = db.RejectEnrollment(request.UserId, request.CourseId)
	case pb.Enrollment_PENDING:
		err = db.SetPendingEnrollment(request.UserId, request.CourseId)
	}
	if err != nil {
		return nil, err
	}
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	return &pb.StatusCode{StatusCode: int32(codes.OK)}, nil
}

// GetCourse find course by id and return JSON object.
func GetCourse(query *pb.RecordRequest, db database.Database) (*pb.Course, error) {

	course, err := db.GetCourse(query.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Course not found")
		}
		return nil, err

	}
	return course, nil
}

// RefreshCourse refreshes the information to a course
func RefreshCourse(ctx context.Context, request *pb.RecordRequest, s scm.SCM, db database.Database, currentUser *pb.User) (*pb.Assignments, error) {

	course, err := db.GetCourse(request.Id)
	if err != nil {
		return nil, err
	}

	remoteID, err := getRemoteIDFor(currentUser, course.Provider)
	if err != nil {
		return nil, err
	}

	if currentUser.IsAdmin {
		updateRepoToPrivate(ctx, db, s, course.DirectoryId)
	}

	assignments, err := RefreshCourseInformation(ctx, db, course, remoteID, s)
	if err != nil {
		return nil, err
	}

	return &pb.Assignments{Assignments: assignments}, nil
}

var runLock sync.Mutex

func RefreshCourseInformation(ctx context.Context, db database.Database, course *pb.Course, remoteID *pb.RemoteIdentity, s scm.SCM) ([]*pb.Assignment, error) {
	runLock.Lock()
	defer runLock.Unlock()

	contextWithTimeout, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	directory, err := s.GetDirectory(contextWithTimeout, course.DirectoryId)
	if err != nil {
		log.Println("Problem fetching Directory")
		return nil, err
	}

	path, err := s.CreateCloneURL(contextWithTimeout, &scm.CreateClonePathOptions{
		Directory:  directory.Path,
		Repository: "tests.git",
		UserToken:  remoteID.AccessToken,
	})
	if err != nil {
		log.Println("Problem Creating clone URL")
		return nil, err
	}

	runner := ci.Local{}

	cloneDirectory := "agclonepath"

	// Clone all tests from tests repositry
	job := &ci.Job{
		Commands: []string{
			"mkdir " + cloneDirectory,
			"cd " + cloneDirectory,
			"git clone " + path,
		},
	}

	log.Println("Running Job ", job)
	_, err = runner.Run(ctx, job)
	if err != nil {
		log.Println("Problem Running CI runner")
		runner.Run(ctx, &ci.Job{
			Commands: []string{
				"yes | rm -r " + cloneDirectory,
			},
		})
		return nil, err
	}
	// Parse assignments in the test directory
	assignments, err := yamlparser.Parse("agclonepath/tests")
	if err != nil {
		log.Println("Problem getting assignments")
		return nil, err
	}

	// Cleanup downloaded
	runner.Run(ctx, &ci.Job{
		Commands: []string{
			"yes | rm -r " + cloneDirectory,
		},
	})

	for _, v := range assignments {
		assignment, err := createAssignment(v, course)
		if err != nil {
			log.Println("Problem createing assignment")
			return nil, err
		}
		// a hack to avoid error messages

		ass := pb.Assignment{
			Id:         assignment.Id,
			CourseId:   assignment.CourseId,
			IsGrouplab: assignment.IsGrouplab,
			Language:   assignment.Language,
		}

		if err := db.CreateAssignment(&ass); err != nil {

			// end of the hack

			log.Println("Problem adding assignment to DB")
			return nil, err
		}

	}

	return assignments, nil
}

func getRemoteIDFor(user *pb.User, provider string) (*pb.RemoteIdentity, error) {
	var remoteID *pb.RemoteIdentity
	for _, v := range user.RemoteIdentities {
		if v.Provider == provider {
			remoteID = v
			break
		}
	}
	if remoteID == nil {
		return nil, echo.ErrNotFound
	}
	return remoteID, nil
}

func createAssignment(request *pb.Assignment, course *pb.Course) (*pb.Assignment, error) {

	return &pb.Assignment{
		AutoApprove: request.AutoApprove,
		CourseId:    course.Id,
		Deadline:    request.Deadline,
		Language:    request.Language,
		Name:        request.Name,
		Order:       uint32(request.Id),
		IsGrouplab:  request.IsGrouplab,
	}, nil
}

// GetSubmission returns a single submission for a assignment and a user
func GetSubmission(request *pb.RecordRequest, db database.Database, currentUser *pb.User) (*pb.Submission, error) {
	submission, err := db.GetSubmissionForUser(request.Id, currentUser.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "not found")
		}
		return nil, err
	}

	return submission, nil

}

// ListSubmissions returns all the latests submissions for a user to a course
func ListSubmissions(request *pb.ActionRequest, db database.Database) (*pb.Submissions, error) {

	submissions, err := db.GetSubmissions(request.UserId, request.CourseId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "not found")
		}
		return nil, err
	}
	return &pb.Submissions{Submissions: submissions}, nil
}

// UpdateCourse updates an existing course
func UpdateCourse(ctx context.Context, request *pb.Course, db database.Database, s scm.SCM) (*pb.StatusCode, error) {
	_, err := db.GetCourse(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "Course not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, err
	}

	if !validCourse(request) {
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, "invalid payload")
	}

	contextWithTimeout, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	// Check that the directory exists.
	_, err = s.GetDirectory(contextWithTimeout, request.DirectoryId)
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	if err := db.UpdateCourse(request); err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	return &pb.StatusCode{StatusCode: int32(codes.OK)}, nil
}

// GetEnrollmentsByCourse get all enrollments for a course.
func GetEnrollmentsByCourse(request *pb.RecordRequest, db database.Database) (*pb.Enrollments, error) {

	enrollments, err := db.GetEnrollmentsByCourse(request.Id, request.Statuses...)

	if err != nil {
		return nil, err
	}

	for _, enrollment := range enrollments {
		enrollment.User, err = db.GetUser(enrollment.UserId)
		if err != nil {
			return nil, err
		}
	}
	return &pb.Enrollments{Enrollments: enrollments}, nil
}

// UpdateSubmission updates a submission
func UpdateSubmission(request *pb.RecordRequest, db database.Database) error {

	err := db.UpdateSubmissionByID(request.Id, true)
	if err != nil {
		return err
	}

	return nil

}

// ListGroupSubmissions fetches all submissions from specific group
func ListGroupSubmissions(request *pb.ActionRequest, db database.Database) (*pb.Submissions, error) {
	submissions, err := db.GetGroupSubmissions(request.CourseId, request.UserId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "not found")
		}
		return nil, err
	}

	return &pb.Submissions{Submissions: submissions}, nil

}

// GetCourseInformationURL returns the course information html as string
func GetCourseInformationURL(request *pb.RecordRequest, db database.Database) (*pb.URLResponse, error) {

	courseInfoRepo, err := db.GetRepositoriesByCourseIDAndType(request.Id, pb.Repository_COURSEINFO)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "No repository found")
	}
	if len(courseInfoRepo) > 1 {
		return nil, status.Errorf(codes.Internal, "Too many information repositories exist")
	} else if len(courseInfoRepo) < 1 {
		return nil, status.Errorf(codes.Internal, "No information repository found")
	}
	return &pb.URLResponse{Url: courseInfoRepo[0].HtmlUrl}, nil
}

// GetRepositoryURL returns the repository information
func GetRepositoryURL(currentUser *pb.User, request *pb.RepositoryRequest, db database.Database) (*pb.URLResponse, error) {
	// One course can have many user repos, but only one of every other repo type
	log.Println("GetRepoURL: requested type is ", request.Type)
	if request.Type == pb.Repository_USER {

		if currentUser == nil {
			return nil, status.Errorf(codes.Unauthenticated, "user not registered")
		}

		userRepo, err := db.GetRepoByCourseIDUserIDandType(request.CourseId, currentUser.Id, request.Type)
		if err != nil {
			log.Println("GetRepoURL fails getRepoByCourseIDUserIDandType")
			return nil, err
		}
		return &pb.URLResponse{Url: userRepo.HtmlUrl}, nil
	}
	repos, err := db.GetRepositoriesByCourseIDAndType(request.CourseId, request.Type)

	if err != nil {
		log.Println("GetRepoURL gets fails getRepoByCourseIDandType")

		return nil, err
	}
	if len(repos) > 1 {
		return nil, status.Errorf(codes.NotFound, "too many course information repositories exists for this course")
	}
	if len(repos) == 0 {
		return nil, status.Errorf(codes.NotFound, "no repositories found")
	}
	return &pb.URLResponse{Url: repos[0].HtmlUrl}, nil
}

func updateRepoToPrivate(ctx context.Context, db database.Database, s scm.SCM, directoryID uint64) {
	repositories, err := db.GetRepositoriesByDirectory(directoryID)
	if err != nil {
		return
	}
	payment, _ := s.GetPaymentPlan(ctx, directoryID)
	// If privaterepos is bigger than 0, we know that the org/team is paid for.
	if payment.PrivateRepos > 0 {
		for _, repo := range repositories {
			if repo.RepoType != pb.Repository_ASSIGNMENT &&
				repo.RepoType != pb.Repository_COURSEINFO &&
				repo.RepoType != pb.Repository_SOLUTION {

				scmRepo := &scm.Repository{
					DirectoryID: repo.DirectoryId,
					ID:          repo.RepositoryId,
				}
				err := s.UpdateRepository(ctx, scmRepo)
				if err != nil {
					return
				}
			}
		}
	}
}
