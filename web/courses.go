package web

import (
	"context"
	"net/http"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

// NewCourseRequest represents a request for a new course.
type NewCourseRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Year uint   `json:"year"`
	Tag  string `json:"tag"`

	Provider    string `json:"provider"`
	DirectoryID uint64 `json:"directoryid"`
}

func (cr *NewCourseRequest) valid() bool {
	return cr != nil &&
		cr.Name != "" &&
		cr.Code != "" &&
		(cr.Provider == "github" || cr.Provider == "gitlab" || cr.Provider == "fake") &&
		cr.DirectoryID != 0 &&
		cr.Year != 0 &&
		cr.Tag != ""
}

// EnrollUserRequest represent a request for enrolling a user to a course.
type EnrollUserRequest struct {
	Status uint `json:"status"`
}

func (eur *EnrollUserRequest) valid() bool {
	return eur.Status <= models.Teacher
}

// ListCourses returns a JSON object containing all the courses in the database.
func ListCourses(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courses, err := db.GetCourses()
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, courses, "\t")
	}
}

// ListCoursesWithEnrollment lists all existing courses with the provided users
// enrollment status.
// If status query param is provided, lists only courses of the student filtered by the query param.
func ListCoursesWithEnrollment(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("uid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		statuses, err := parseEnrollmentStatus(c.QueryParam("status"))
		if err != nil {
			return err
		}

		courses, err := db.GetCoursesByUser(id, statuses...)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, courses, "\t")
	}
}

// ListAssignments lists the assignments for the provided course.
func ListAssignments(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		assignments, err := db.GetAssignmentsByCourse(id)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, assignments, "\t")
	}
}

// BaseHookOptions contains options shared among all webhooks.
type BaseHookOptions struct {
	BaseURL string
	// Secret is used to verify that the event received is legit. GitHub
	// sends back a signature of the payload, while GitLab just sends back
	// the secret. This is all handled by the
	// gopkg.in/go-playground/webhooks.v3 package.
	Secret string
}

// NewCourse creates a new course and associates it with a directory (organization in github)
// and creates the repositories for the course.
//TODO(meling) refactor this to separate out business logic
//TODO(meling) remove logger from method, and use c.Logger() instead
// Problem: (the echo.Logger is not compatible with logrus.FieldLogger)
func NewCourse(logger logrus.FieldLogger, db database.Database, bh *BaseHookOptions) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cr NewCourseRequest
		if err := c.Bind(&cr); err != nil {
			return err
		}
		if !cr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		user := c.Get("user").(*models.User)
		// TODO: This check should be performed in AccessControl.
		if !user.IAdmin() {
			// Only teacher with admin rights can create a new course
			return c.NoContent(http.StatusForbidden)
		}

		s, err := getSCM(c, cr.Provider)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		directory, err := s.GetDirectory(ctx, cr.DirectoryID)
		if err != nil {
			return err
		}
		repos, err := s.GetRepositories(ctx, directory)
		if err != nil {
			return err
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
					logger.WithField("repo", path).WithField("private", privRepo).WithError(err).Warn("Failed to create repository")
					return err
				}
				logger.WithField("repo", repo).Println("Created new repository")
			}

			hooks, err := s.ListHooks(ctx, repo)
			if err != nil {
				logger.WithField("repo", path).WithError(err).Warn("Failed to list hooks for repository")
				return err
			}
			hasAGWebHook := false
			for _, hook := range hooks {
				logger.WithField("url", hook.URL).WithField("id", hook.ID).WithField("name", hook.Name).Println("Hook for repository")
				// TODO this check is specific for the github implementation ; fix this
				if hook.Name == "web" {
					hasAGWebHook = true
					break
				}
			}

			if !hasAGWebHook {
				if err := s.CreateHook(ctx, &scm.CreateHookOptions{
					URL:        GetEventsURL(bh.BaseURL, cr.Provider),
					Secret:     bh.Secret,
					Repository: repo,
				}); err != nil {
					logger.WithField("repo", path).WithError(err).Println("Failed to create webhook for repository")
					return err
				}

				logger.WithField("repo", repo).Println("Created new webhook for repository")
			}

			courseRepo := &models.Repository{
				DirectoryID:  directory.ID,
				RepositoryID: repo.ID,
				HTMLURL:      repo.WebURL,
				Type:         repoType(path),
			}
			if err := db.CreateRepository(courseRepo); err != nil {
				return err
			}
		}

		course := models.Course{
			Name:            cr.Name,
			CourseCreatorID: user.ID,
			Code:            cr.Code,
			Year:            cr.Year,
			Tag:             cr.Tag,
			Provider:        cr.Provider,
			DirectoryID:     directory.ID,
		}
		if err := db.CreateCourse(user.ID, &course); err != nil {
			//TODO(meling) Should we even communicate bad request to the client?
			// We should log errors and debug it on the server side instead.
			// If clients make mistakes, there is nothing it can do with the error message.
			if err == database.ErrCourseExists {
				return c.JSONPretty(http.StatusBadRequest, err.Error(), "\t")
			}
			return err
		}
		return c.JSONPretty(http.StatusCreated, &course, "\t")
	}
}

func repoType(path string) (repoType models.RepoType) {
	switch path {
	case InfoRepo:
		repoType = models.CourseInfoRepo
	case AssignmentRepo:
		repoType = models.AssignmentsRepo
	case TestsRepo:
		repoType = models.TestsRepo
	case SolutionsRepo:
		repoType = models.SolutionsRepo
	}
	return
}

// CreateEnrollment enrolls a user in a course.
func CreateEnrollment(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courseID, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		userID, err := parseUint(c.Param("uid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		var eur EnrollUserRequest
		if err := c.Bind(&eur); err != nil {
			return err
		}
		if !eur.valid() || userID == 0 || courseID == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		enrollment := models.Enrollment{
			UserID:   userID,
			CourseID: courseID,
		}
		if err := db.CreateEnrollment(&enrollment); err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}

// UpdateEnrollment accepts or rejects a user to enroll in a course.
func UpdateEnrollment(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courseID, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		userID, err := parseUint(c.Param("uid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		var eur EnrollUserRequest
		if err := c.Bind(&eur); err != nil {
			return err
		}
		if !eur.valid() || userID == 0 || courseID == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// check that userID has enrolled in courseID
		if _, err := db.GetEnrollmentByCourseAndUser(courseID, userID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		user := c.Get("user").(*models.User)
		// TODO: This check should be performed in AccessControl.
		if !user.IAdmin() {
			// Only admin users are allowed to enroll or reject users to a course.
			// TODO we should also allow users of the 'teachers' team to accept/reject users
			return c.NoContent(http.StatusUnauthorized)
		}

		// TODO If the enrollment is accepted, create repositories and permissions for them with webhooks.
		switch eur.Status {
		case models.Student:
			// update enrollment for student in database
			err = db.EnrollStudent(userID, courseID)
			if err != nil {
				return err
			}
			course, err := db.GetCourse(courseID)
			if err != nil {
				return err
			}
			student, err := db.GetUser(userID)
			if err != nil {
				return err
			}

			// create user repo and team on the SCM
			repo, err := createUserRepoAndTeam(c, course, student)
			if err != nil {
				return err
			}
			// logger.WithField("repo", repo).Println("Successfully created new student repository")

			// add student repo to database if SCM interaction was successful
			studentRepo := &models.Repository{
				DirectoryID:  course.DirectoryID,
				RepositoryID: repo.ID,
				HTMLURL:      repo.WebURL,
				Type:         models.UserRepo,
				UserID:       userID,
			}
			if err := db.CreateRepository(studentRepo); err != nil {
				return err
			}

		case models.Teacher:
			err = db.EnrollTeacher(userID, courseID)
		case models.Rejected:
			err = db.RejectEnrollment(userID, courseID)
		case models.Pending:
			err = db.SetPendingEnrollment(userID, courseID)
		}
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	}
}

func createUserRepoAndTeam(c echo.Context, course *models.Course, student *models.User) (*scm.Repository, error) {
	s, err := getSCM(c, course.Provider)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
	defer cancel()

	dir, err := s.GetDirectory(ctx, course.DirectoryID)
	if err != nil {
		return nil, err
	}

	gitUserNames, err := fetchGitUserNames(ctx, s, course, student)
	if err != nil {
		return nil, err
	}
	if len(gitUserNames) > 1 || len(gitUserNames) == 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	// the student's git user name is the same as the team name
	teamName := gitUserNames[0]

	opt := &scm.CreateRepositoryOptions{
		Directory: dir,
		Path:      StudentRepoName(teamName),
		Private:   true,
	}
	return s.CreateRepoAndTeam(ctx, opt, teamName, gitUserNames)
}

// GetCourse find course by id and return JSON object.
func GetCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		course, err := db.GetCourse(id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.JSONPretty(http.StatusOK, course, "\t")
	}
}

// RefreshCourse updates the course assignments (and possibly other course information).
func RefreshCourse(logger logrus.FieldLogger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		course, err := db.GetCourse(cid)
		if err != nil {
			return err
		}
		s, err := getSCM(c, course.Provider)
		if err != nil {
			return err
		}

		user := c.Get("user").(*models.User)
		if user.IAdmin() {
			// Only admin users should be able to update repos to private, if they are public.
			//TODO(meling) remove this; we should prevent creating public repos in the first place
			// and instead only if the teacher specifically requests public repo from the frontend
			updateRepoToPrivate(c.Request().Context(), db, s, course.DirectoryID)
		}

		assignments, err := FetchAssignments(c.Request().Context(), s, course)
		if err != nil {
			return err
		}
		if err = db.UpdateAssignments(assignments); err != nil {
			logger.WithError(err).Error("Failed to update assignments in database")
			return err
		}

		//TODO(meling) Are the assignments (previously it was yamlparser.NewAssignmentRequest)
		// needed by the frontend? Or can we use models.Assignment instead through db.GetAssignmentsByCourse()?
		// Currently the frontend looks faulty, i.e. doesn't use the returned results from this
		// function; see 'refreshCoursesFor(courseID: number): Promise<any>' in ServerProvider.ts,
		// which does a 'return this.makeUserInfo(result.data);', indicating that the result is
		// converted into a UserInfo type, which probably fails??
		return c.JSONPretty(http.StatusOK, &assignments, "\t")
	}
}

// GetSubmission returns a single submission for a assignment and a user
func GetSubmission(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		assignmentID, err := parseUint(c.Param("aid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		user := c.Get("user").(*models.User)

		submission, err := db.GetSubmissionForUser(assignmentID, user.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.JSONPretty(http.StatusOK, submission, "\t")
	}
}

// ListSubmissions returns all the latests submissions for a user to a course
func ListSubmissions(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courseID, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// Check if a user is provided, else used logged in user
		userID, err := parseUint(c.Param("uid"))
		if err != nil {
			userID = c.Get("user").(*models.User).ID
		}

		submission, err := db.GetSubmissions(courseID, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.JSONPretty(http.StatusOK, submission, "\t")
	}
}

// UpdateCourse updates an existing course
func UpdateCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if _, err := db.GetCourse(id); err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		// TODO: Might be better to define a Validate method on models.Course and bind to that.
		var cr NewCourseRequest
		if err := c.Bind(&cr); err != nil {
			return err
		}
		if !cr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		s, err := getSCM(c, cr.Provider)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		// Check that the directory exists.
		_, err = s.GetDirectory(ctx, cr.DirectoryID)
		if err != nil {
			return err
		}

		if err := db.UpdateCourse(&models.Course{
			ID:          id,
			Name:        cr.Name,
			Code:        cr.Code,
			Year:        cr.Year,
			Tag:         cr.Tag,
			Provider:    cr.Provider,
			DirectoryID: cr.DirectoryID,
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)

	}
}

// GetEnrollmentsByCourse get all enrollments for a course.
func GetEnrollmentsByCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		statuses, err := parseEnrollmentStatus(c.QueryParam("status"))
		if err != nil {
			return err
		}

		enrollments, err := db.GetEnrollmentsByCourse(id, statuses...)
		if err != nil {
			return err
		}

		for _, enrollment := range enrollments {
			enrollment.User, err = db.GetUser(enrollment.UserID)
			if err != nil {
				return err
			}
		}

		return c.JSONPretty(http.StatusOK, enrollments, "\t")
	}
}

// UpdateSubmission updates a submission
func UpdateSubmission(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		sid, err := parseUint(c.Param("sid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		err = db.UpdateSubmissionByID(sid, true)
		if err != nil {
			return err
		}

		return nil
	}
}

// ListGroupSubmissions fetches all submissions from specific group
func ListGroupSubmissions(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		submission, err := db.GetGroupSubmissions(cid, gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.JSONPretty(http.StatusOK, submission, "\t")
	}
}

// GetCourseInformationURL returns the course information html as string
//TODO(meling) merge this functionality with func below into a single func.
// Use only one db call as well. Make sure the db can only return one repo.
func GetCourseInformationURL(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse courseID")
		}
		courseInfoRepo, err := db.GetRepositoriesByCourseAndType(cid, models.CourseInfoRepo)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not retrieve any course information repos")
		}
		// There should only exist 1 course info pr course.
		if len(courseInfoRepo) > 1 && len(courseInfoRepo) == 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, "Too many course information repositories exists for this course")
		}

		// Have to be in string array to be able to jsonify so frontend recognize it.
		// See public/src/HttpHelper.ts -> send()
		courseInfoURL := []string{courseInfoRepo[0].HTMLURL}
		return c.JSONPretty(http.StatusOK, &courseInfoURL, "\t")
	}
}

// GetRepositoryURL returns the course information html as string
func GetRepositoryURL(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to parse course ID")
		}
		repoType, err := models.RepoTypeFromString(c.QueryParam("type"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		var repos []*models.Repository
		if repoType.IsStudentRepo() {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "user not registered")
			}
			userRepo, err := db.GetRepositoryByCourseUserType(cid, user.ID, models.UserRepo)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "GetRepoByCourseIDUserIDandType: Could not retrieve any UserRepo")
			}
			repos = append(repos, userRepo)
		} else {
			repos, err = db.GetRepositoriesByCourseAndType(cid, repoType)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "GetRepositoriesByCourseIDAndType: Could not retrieve any repos")
			}
		}

		// There should only exist one of each specific repo pr course.
		// AssignmentRepo, CourseInfoRepo, SolutionRepo, TestRepo
		// There can exist many UserRepo, but only one pr user
		if len(repos) > 1 && len(repos) == 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, "Too many course information repositories exists for this course")
		}

		// Have to be in string array to be able to jsonify so frontend recognize it.
		// See public/src/HttpHelper.ts -> send()
		repoURL := []string{repos[0].HTMLURL}
		return c.JSONPretty(http.StatusOK, &repoURL, "\t")
	}
}

//TODO(meling) there are no error handling here; also add tests
//TODO(meling) this method should probably not be necessary because we shouldn't let the frontend
// create non-private repos unless this action is specifically specified in the CreateRepository call.
func updateRepoToPrivate(ctx context.Context, db database.Database, s scm.SCM, directoryID uint64) {
	repositories, err := db.GetRepositoriesByDirectory(directoryID)
	if err != nil {
		return
	}

	payment, _ := s.GetPaymentPlan(ctx, directoryID)
	// If privaterepos is bigger than 0, we know that the org/team is paid for.
	if payment.PrivateRepos > 0 {
		for _, repo := range repositories {
			if repo.Type != models.AssignmentsRepo &&
				repo.Type != models.CourseInfoRepo &&
				repo.Type != models.SolutionsRepo {

				scmRepo := &scm.Repository{
					DirectoryID: repo.DirectoryID,
					ID:          repo.RepositoryID,
				}
				err := s.UpdateRepository(ctx, scmRepo)
				if err != nil {
					return
				}
			}
		}
	}
}
