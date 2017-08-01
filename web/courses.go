package web

import (
	"context"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 60 * time.Second

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

// EnrollUserRequest represent a request for enrolling a user to a course
type EnrollUserRequest struct {
	UserID   uint64 `json:"userid"`
	CourseID uint64 `json:"courseid"`
	Status   uint   `json:"status"`
}

func (eur *EnrollUserRequest) valid() bool {
	return eur.CourseID != 0 && eur.UserID != 0 && eur.Status <= models.Accepted
}

// NewGroupRequest represents a new group
type NewGroupRequest struct {
	Name     string   `json:"name"`
	CourseID uint64   `json:"courseid"`
	UserIDs  []uint64 `json:"userids"`
}

func (grp *NewGroupRequest) valid() bool {
	return grp != nil &&
		grp.Name != "" &&
		len(grp.UserIDs) > 0
}

// UpdateGroupRequest updates group
type UpdateGroupRequest struct {
	Status uint `json:"status"`
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
			return err
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
			return err
		}
		assignments, err := db.GetAssignmentsByCourse(id)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, assignments, "\t")
	}
}

// Default repository names.
const (
	InfoRepo       = "course-info"
	AssignmentRepo = "assignments"
	TestsRepo      = "tests"
	SolutionsRepo  = "solutions"
)

// NewCourse creates a new course and associates it with a directory (organization in github)
// and creates the repositories for the course.
func NewCourse(logger logrus.FieldLogger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cr NewCourseRequest
		if err := c.Bind(&cr); err != nil {
			return err
		}
		if !cr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if c.Get(cr.Provider) == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "provider "+cr.Provider+" not registered")
		}
		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		s := c.Get(cr.Provider).(scm.SCM)

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		directory, err := s.GetDirectory(ctx, cr.DirectoryID)
		if err != nil {
			return err
		}

		var paths = []string{InfoRepo, AssignmentRepo, TestsRepo, SolutionsRepo}
		for _, path := range paths {
			repo, err := s.CreateRepository(
				ctx,
				&scm.CreateRepositoryOptions{
					Path:      path,
					Directory: directory},
			)
			if err != nil {
				return err
			}
			logger.WithField("repo", repo).Println("Created new repository")
		}

		course := models.Course{
			Name:        cr.Name,
			Code:        cr.Code,
			Year:        cr.Year,
			Tag:         cr.Tag,
			Provider:    cr.Provider,
			DirectoryID: directory.ID,
		}

		if err := db.CreateCourse(&course); err != nil {
			return err
		}

		return c.JSONPretty(http.StatusCreated, &course, "\t")
	}
}

// SetEnrollment enrolls a user in a course.
func SetEnrollment(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courseID, err := parseUint(c.Param("cid"))
		if err != nil {
			return err
		}
		userID, err := parseUint(c.Param("uid"))
		if err != nil {
			return err
		}

		var eur EnrollUserRequest
		if err := c.Bind(&eur); err != nil {
			return err
		}
		if !eur.valid() || eur.UserID != userID || eur.CourseID != courseID {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		enrollment := models.Enrollment{
			UserID:   eur.UserID,
			CourseID: eur.CourseID,
		}
		if err := db.CreateEnrollment(&enrollment); err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		user := c.Get("user").(*models.User)
		if !user.IsAdmin {
			// This means that the request has been accepted for processing,
			// i.e., we need to wait for a teacher to accept the enrollment.
			// TODO: Rename Accepted to Approved to avoid this confusion.
			return c.NoContent(http.StatusAccepted)
		}

		switch eur.Status {
		case models.Accepted:
			if err := db.AcceptEnrollment(enrollment.ID); err != nil {
				return err
			}
		case models.Rejected:
			if err := db.RejectEnrollment(enrollment.ID); err != nil {
				return err
			}
		}
		return c.NoContent(http.StatusOK)
	}
}

// GetCourse find course by id and return JSON object.
func GetCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return err
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

// GetSubmission returns a single submission for a assignment and a user
func GetSubmission(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := parseUint(c.Param("cid"))
		if err != nil {
			return err
		}
		assignmentID, err := parseUint(c.Param("aid"))
		if err != nil {
			return err
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
			return err
		}

		user := c.Get("user").(*models.User)
		submission, err := db.GetSubmissions(courseID, user.ID)
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
			return err
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

		if c.Get(cr.Provider) == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "provider "+cr.Provider+" not registered")
		}
		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		s := c.Get(cr.Provider).(scm.SCM)

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
			return err
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

// NewGroup creates a new group under a course
func NewGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return err
		}

		if _, err := db.GetCourse(cid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}

		var grp NewGroupRequest
		if err := c.Bind(&grp); err != nil {
			return err
		}
		if !grp.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		users, err := db.GetUsers(grp.UserIDs...)
		if err != nil {
			return err
		}
		// check if provided user ids are valid
		if len(users) != len(grp.UserIDs) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		group := models.Group{
			Name:     grp.Name,
			CourseID: cid,
			Users:    users,
		}
		// CreateGroup creates a new group and update group_id in enrollment table
		if err := db.CreateGroup(&group); err != nil {
			return err
		}

		return c.JSONPretty(http.StatusCreated, &group, "\t")
	}
}

// PatchGroup updates status of a group
func PatchGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("gid"))
		if err != nil {
			return err
		}
		oldgrp, err := db.GetGroup(id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		var ngrp UpdateGroupRequest
		if err := c.Bind(&ngrp); err != nil {
			return err
		}
		if ngrp.Status > models.Accepted {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		user := c.Get("user").(*models.User)
		if !user.IsAdmin {
			// Ony Admin i.e Teacher can update status of a group
			return c.NoContent(http.StatusForbidden)
		}

		if err := db.UpdateGroupStatus(&models.Group{
			ID:     oldgrp.ID,
			Status: ngrp.Status,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	}
}

// GetGroups returns all groups under a course
func GetGroups(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return err
		}
		if _, err := db.GetCourse(cid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}
		groups, err := db.GetGroups(cid)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, groups, "\t")
	}
}
