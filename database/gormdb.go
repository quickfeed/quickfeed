package database

import (
	"errors"
	"fmt"
	"sort"

	pb "github.com/autograde/aguis/ag"
	"github.com/jinzhu/gorm"
)

var (
	// ErrDuplicateIdentity is returned when trying to associate a remote identity
	// with a user account and the identity is already in use.
	ErrDuplicateIdentity = errors.New("remote identity registered with another user")
	// ErrEmptyGroup is returned when trying to create a group without users.
	ErrEmptyGroup = errors.New("cannot create group without users")
	// ErrDuplicateGroup is returned when trying to create a group with the same
	// name as a previously registered group.
	ErrDuplicateGroup = errors.New("group name already registered")
	// ErrUpdateGroup is returned when updating a group's enrollment fails.
	ErrUpdateGroup = errors.New("failed to update group enrollment")
	// ErrCourseExists is returned when trying to create an association in
	// the database for a DirectoryId that already exists in the database.
	ErrCourseExists = errors.New("course already exists on git provider")
	// ErrInsufficientAccess is returned when trying to update database
	// with insufficient access priviledges.
	ErrInsufficientAccess = errors.New("user must be admin to perform this operation")
	// ErrCreateRepo is returned when trying to create repository with wrong argument.
	ErrCreateRepo = errors.New("failed to create repository; invalid arguments")
)

// GormDB implements the Database interface.
type GormDB struct {
	conn *gorm.DB
}

// NewGormDB creates a new gorm database using the provided driver.
func NewGormDB(driver, path string, logger GormLogger) (*GormDB, error) {
	conn, err := gorm.Open(driver, path)
	if err != nil {
		return nil, err
	}

	if logger != nil {
		conn.SetLogger(logger)
	}
	conn.LogMode(logger != nil)

	if err := conn.AutoMigrate(
		&pb.User{},
		&pb.RemoteIdentity{},
		&pb.Course{},
		&pb.Enrollment{},
		&pb.Assignment{},
		&pb.Submission{},
		&pb.Group{},
		&pb.Repository{},
	).Error; err != nil {
		return nil, err
	}

	return &GormDB{conn}, nil
}

// GetUser fetches a user by ID with remote identities.
func (db *GormDB) GetUser(uid uint64) (*pb.User, error) {
	var user pb.User
	if err := db.conn.Preload("RemoteIdentities").First(&user, uid).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByRemoteIdentity fetches user by remote identity.
func (db *GormDB) GetUserByRemoteIdentity(remote *pb.RemoteIdentity) (*pb.User, error) {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity pb.RemoteIdentity
	if err := tx.
		Where(&pb.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get the user.
	var user pb.User
	if err := tx.Preload("RemoteIdentities").First(&user, remoteIdentity.UserID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateAccessToken refreshes the token info for the given remote identity.
func (db *GormDB) UpdateAccessToken(remote *pb.RemoteIdentity) error {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity pb.RemoteIdentity
	if err := tx.
		Where(&pb.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update the access token.
	if err := tx.Model(&remoteIdentity).Update("access_token", remote.AccessToken).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// GetUsers fetches all users by provided IDs.
func (db *GormDB) GetUsers(uids ...uint64) ([]*pb.User, error) {
	m := db.conn
	if len(uids) > 0 {
		m = m.Where(uids)
	}
	m = m.Preload("RemoteIdentities")

	var users []*pb.User
	if err := m.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates user information.
func (db *GormDB) UpdateUser(user *pb.User) error {
	return db.conn.Model(&pb.User{}).Updates(user).Error
}

// SetAdmin promotes user with given ID to admin.
func (db *GormDB) SetAdmin(uid uint64) error {
	var user pb.User
	if err := db.conn.First(&user, uid).Error; err != nil {
		return err
	}
	// TODO(vera): is there a reason we are doing it in two steps?
	admin := true
	user.IsAdmin = admin
	return db.conn.Save(&user).Error
}

// GetRemoteIdentity fetches remote identity by provider and ID.
func (db *GormDB) GetRemoteIdentity(provider string, rid uint64) (*pb.RemoteIdentity, error) {
	var remoteIdentity pb.RemoteIdentity
	if err := db.conn.Model(&pb.RemoteIdentity{}).
		Where(&pb.RemoteIdentity{
			Provider: provider,
			RemoteID: rid,
		}).
		First(&remoteIdentity).Error; err != nil {
		return nil, err
	}
	return &remoteIdentity, nil
}

// CreateUserFromRemoteIdentity creates new user record from remote identity, sets user with ID 1 as admin.
func (db *GormDB) CreateUserFromRemoteIdentity(user *pb.User, remoteIdentity *pb.RemoteIdentity) error {
	user.RemoteIdentities = []*pb.RemoteIdentity{remoteIdentity}
	if err := db.conn.Create(&user).Error; err != nil {
		return err
	}
	// The first user defaults to admin user.
	if user.ID == 1 {
		if err := db.SetAdmin(1); err != nil {
			return err
		}
		admin := true
		user.IsAdmin = admin
	}
	return nil
}

// AssociateUserWithRemoteIdentity associates remote identity with the user with given ID.
func (db *GormDB) AssociateUserWithRemoteIdentity(uid uint64, provider string, rid uint64, accessToken string) error {
	var count uint64
	if err := db.conn.
		Model(&pb.RemoteIdentity{}).
		Where(&pb.RemoteIdentity{
			Provider: provider,
			RemoteID: rid,
		}).
		Not(&pb.RemoteIdentity{
			UserID: uid,
		}).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateIdentity
	}

	var remoteIdentity pb.RemoteIdentity
	return db.conn.
		Where(pb.RemoteIdentity{Provider: provider, RemoteID: rid, UserID: uid}).
		Assign(pb.RemoteIdentity{AccessToken: accessToken}).
		FirstOrCreate(&remoteIdentity).Error
}

// CreateCourse creates a new course if user with given ID is admin, enrolls user as course teacher.
func (db *GormDB) CreateCourse(uid uint64, course *pb.Course) error {
	user, err := db.GetUser(uid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return ErrInsufficientAccess
	}

	var courses uint64
	if err := db.conn.Model(&pb.Course{}).Where(&pb.Course{
		OrganizationID: course.OrganizationID,
	}).Count(&courses).Error; err != nil {
		return err
	}
	if courses > 0 {
		return ErrCourseExists
	}

	//TODO(meling) these db updates should be done as a transaction
	if err := db.conn.Create(course).Error; err != nil {
		return err
	}
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: uid, CourseID: course.ID}); err != nil {
		return err
	}
	if err := db.EnrollTeacher(uid, course.ID); err != nil {
		return err
	}
	return nil
}

// GetCourses returns a list of courses. If one or more course ids are provided,
// the corresponding courses are returned. Otherwise, all courses are returned.
func (db *GormDB) GetCourses(cids ...uint64) ([]*pb.Course, error) {
	m := db.conn
	if len(cids) > 0 {
		m = m.Where(cids)
	}
	var courses []*pb.Course
	if err := m.Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

// GetAssignmentsByCourse fetches all assignments for a course with given ID
func (db *GormDB) GetAssignmentsByCourse(cid uint64) ([]*pb.Assignment, error) {
	var course pb.Course
	if err := db.conn.Preload("Assignments").First(&course, cid).Error; err != nil {
		return nil, err
	}
	return course.Assignments, nil
}

// GetNextAssignment returns the next assignment to be approved for
// the given course, user, or group if the next assignment is a group lab.
func (db *GormDB) GetNextAssignment(cid uint64, uid uint64, gid uint64) (*pb.Assignment, error) {
	assignments, err := db.GetAssignmentsByCourse(cid)
	if err != nil {
		return nil, err
	}
	if len(assignments) < 1 {
		return nil, fmt.Errorf("no assignments found for course %d", cid)
	}
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].Order < assignments[j].Order
	})
	approved := 0
	nxtToApprove := assignments[0]
	for _, v := range assignments {
		var sub *pb.Submission
		switch {
		case v.IsGroupLab && gid > 0:
			query := &pb.Submission{AssignmentID: v.GetID(), GroupID: gid}
			sub, err = db.GetSubmission(query)
		case !v.IsGroupLab && uid > 0:
			query := &pb.Submission{AssignmentID: v.GetID(), UserID: uid}
			sub, err = db.GetSubmission(query)
		default:
			// This is when uid or gid is set to 0, but there is a group or user lab
			// TODO: Fix so uid always is included and gid is optional
			sub = &pb.Submission{Approved: true}
			// return nil, gorm.ErrRecordNotFound
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if sub != nil && sub.Approved {
			approved++
			continue
		}
		nxtToApprove = v
		break
	}
	if len(assignments) == approved {
		return nil, fmt.Errorf("all assignments approved for user %d (group %d) in course %d", uid, gid, cid)
	}
	return nxtToApprove, nil
}

// CreateSubmission creates a new submission record
// TODO: Also check enrollment to see if the user is
// enrolled in the course the assignment belongs to
func (db *GormDB) CreateSubmission(submission *pb.Submission) error {
	// Primary key must be greater than 0.
	if submission.AssignmentID < 1 {
		return gorm.ErrRecordNotFound
	}

	// Either user or group id must be set, but not both.
	var m *gorm.DB
	switch {
	case submission.UserID > 0 && submission.GroupID > 0:
		return gorm.ErrRecordNotFound
	case submission.UserID > 0:
		m = db.conn.First(&pb.User{ID: submission.UserID})
	case submission.GroupID > 0:
		m = db.conn.First(&pb.Group{ID: submission.GroupID})
	default:
		return gorm.ErrRecordNotFound
	}

	// Check that group exists.
	var group uint64
	if err := m.Count(&group).Error; err != nil {
		return err
	}

	// Checks that the assignment exists.
	var assignment uint64
	if err := db.conn.Model(&pb.Assignment{}).Where(&pb.Assignment{
		ID: submission.AssignmentID,
	}).Count(&assignment).Error; err != nil {
		return err
	}

	if assignment+group != 2 {
		return gorm.ErrRecordNotFound
	}
	return db.conn.Create(submission).Error
}

// UpdateSubmission updates submission approved status
func (db *GormDB) UpdateSubmission(sid uint64, approved bool) error {
	//TODO(meling) consider to make this into a transaction
	var submission pb.Submission
	if err := db.conn.First(&submission, sid).Error; err != nil {
		return err
	}
	submission.Approved = approved
	return db.conn.Model(&pb.Submission{}).Update(submission).Error
}

// GetSubmission fetches a submission record
func (db *GormDB) GetSubmission(query *pb.Submission) (*pb.Submission, error) {
	var submission pb.Submission
	if err := db.conn.Where(query).Last(&submission).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetSubmissions returns all submissions for the active assignment for the given course
func (db *GormDB) GetSubmissions(courseID uint64, query *pb.Submission) ([]*pb.Submission, error) {
	var course pb.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}

	// note that, this creates a query with possibly both user and group;
	// it will only limit the number of submissions returned if both are supplied.
	q := &pb.Submission{UserID: query.GetUserID(), GroupID: query.GetGroupID()}
	var latestSubs []*pb.Submission
	for _, a := range course.Assignments {
		q.AssignmentID = a.GetID()
		temp, err := db.GetSubmission(q)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return nil, err
		}
		latestSubs = append(latestSubs, temp)
	}
	return latestSubs, nil
}

// CreateAssignment creates new assignment record
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
		Assign(pb.Assignment{
			Name:        assignment.Name,
			Language:    assignment.Language,
			Deadline:    assignment.Deadline,
			AutoApprove: assignment.AutoApprove,
		}).FirstOrCreate(assignment).Error
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

// CreateEnrollment creates a new pending enrollment.
func (db *GormDB) CreateEnrollment(enrollment *pb.Enrollment) error {
	var user, course uint64
	if err := db.conn.Model(&pb.User{}).Where(&pb.User{
		ID: enrollment.UserID,
	}).Count(&user).Error; err != nil {
		return err
	}
	if err := db.conn.Model(&pb.Course{}).Where(&pb.Course{
		ID: enrollment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if user+course != 2 {
		return gorm.ErrRecordNotFound
	}

	enrollment.Status = pb.Enrollment_PENDING
	return db.conn.Create(&enrollment).Error
}

// EnrollStudent enrolls user as course student.
func (db *GormDB) EnrollStudent(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, pb.Enrollment_STUDENT)
}

// RejectEnrollment rejects user enrollment.
func (db *GormDB) RejectEnrollment(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, pb.Enrollment_REJECTED)
}

// EnrollTeacher enrolls user as course teacher.
func (db *GormDB) EnrollTeacher(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, pb.Enrollment_TEACHER)
}

// SetPendingEnrollment sets enrollment status to pending.
func (db *GormDB) SetPendingEnrollment(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, pb.Enrollment_PENDING)
}

// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
func (db *GormDB) GetEnrollmentsByCourse(cid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error) {
	return db.getEnrollments(&pb.Course{ID: cid}, statuses...)
}

// getEnrollments is generic helper function that return enrollments for either course and user.
func (db *GormDB) getEnrollments(model interface{}, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error) {
	if len(statuses) == 0 {
		statuses = []pb.Enrollment_UserStatus{
			pb.Enrollment_NONE,
			pb.Enrollment_PENDING,
			pb.Enrollment_REJECTED,
			pb.Enrollment_STUDENT,
			pb.Enrollment_TEACHER,
		}
	}
	var enrollments []*pb.Enrollment
	if err := db.conn.Preload("User").Preload("Course").Model(model).
		Where("status in (?)", statuses).
		Association("Enrollments").
		Find(&enrollments).Error; err != nil {
		return nil, err
	}
	return enrollments, nil
}

// GetEnrollmentByCourseAndUser return a record of Enrollment.
func (db *GormDB) GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*pb.Enrollment, error) {
	var enrollment pb.Enrollment
	m := db.conn.Preload("Course").Preload("User")
	if err := m.
		Where(&pb.Enrollment{
			CourseID: cid,
			UserID:   uid,
		}).
		First(&enrollment).Error; err != nil {
		return nil, err
	}
	return &enrollment, nil
}

// UpdateGroupEnrollment sets GroupID of a student enrollment to 0.
func (db *GormDB) UpdateGroupEnrollment(uid, cid uint64) error {
	return db.conn.
		Model(&pb.Enrollment{}).
		Where(&pb.Enrollment{CourseID: cid, UserID: uid}).
		Update("group_id", uint64(0)).Error
}

// setEnrollment updates enrollment status
func (db *GormDB) setEnrollment(uid, cid uint64, status pb.Enrollment_UserStatus) error {
	return db.conn.
		Model(&pb.Enrollment{}).
		Where(&pb.Enrollment{CourseID: cid, UserID: uid}).
		Update(&pb.Enrollment{Status: status}).Error
}

// GetCoursesByUser returns all courses (with enrollment status)
// for the given user id.
// If enrollment statuses is provided, the set of courses returned
// is filtered according to these enrollment statuses.
func (db *GormDB) GetCoursesByUser(uid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Course, error) {
	enrollments, err := db.getEnrollments(&pb.User{ID: uid}, statuses...)
	if err != nil {
		return nil, err
	}

	var courseIDs []uint64
	m := make(map[uint64]*pb.Enrollment)
	for _, enrollment := range enrollments {
		m[enrollment.CourseID] = enrollment
		courseIDs = append(courseIDs, enrollment.CourseID)
	}

	if len(statuses) == 0 {
		courseIDs = nil
	} else if len(courseIDs) == 0 {
		// No need to query database since user have no enrolled courses.
		return []*pb.Course{}, nil
	}
	courses, err := db.GetCourses(courseIDs...)
	if err != nil {
		return nil, err
	}

	for _, course := range courses {
		course.Enrolled = pb.Enrollment_NONE
		if enrollment, ok := m[course.ID]; ok {
			course.Enrolled = enrollment.Status
		}
	}
	return courses, nil
}

// GetCourse fetches course by ID.
func (db *GormDB) GetCourse(cid uint64) (*pb.Course, error) {
	var course pb.Course
	if err := db.conn.First(&course, cid).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourseByOrganizationID fetches course by organization ID.
func (db *GormDB) GetCourseByOrganizationID(did uint64) (*pb.Course, error) {
	var course pb.Course
	if err := db.conn.First(&course, &pb.Course{OrganizationID: did}).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// UpdateCourse updates course information.
func (db *GormDB) UpdateCourse(course *pb.Course) error {
	return db.conn.Model(&pb.Course{}).Updates(course).Error
}

// CreateRepository creates a new repository record.
func (db *GormDB) CreateRepository(repo *pb.Repository) error {
	if repo.OrganizationID == 0 || repo.RepositoryID == 0 {
		// both organization and repository must be non-zero
		return ErrCreateRepo
	}

	switch {
	case repo.UserID > 0:
		// check that user exists before creating repo in database
		err := db.conn.First(&pb.User{}, repo.UserID).Error
		if err != nil {
			return err
		}
	case repo.GroupID > 0:
		// check that group exists before creating repo in database
		err := db.conn.First(&pb.Group{}, repo.GroupID).Error
		if err != nil {
			return err
		}
	case !repo.RepoType.IsCourseRepo():
		// both user and group unset, then repository type must be an autograder repo type
		return ErrCreateRepo
	}

	return db.conn.Create(repo).Error
}

// GetRepository fetches repository by ID.
func (db *GormDB) GetRepository(rid uint64) (*pb.Repository, error) {
	// This uses the repository ID from the provider to search with,
	// and not the id of the entry in the database
	var repo pb.Repository
	if err := db.conn.First(&repo, &pb.Repository{RepositoryID: rid}).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetRepositories fetches all repositories satisfying the given repository fields
func (db *GormDB) GetRepositories(query *pb.Repository) ([]*pb.Repository, error) {
	var repos []*pb.Repository
	if err := db.conn.Find(&repos, query).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

// Close closes the gorm database.
func (db *GormDB) Close() error {
	return db.conn.Close()
}
