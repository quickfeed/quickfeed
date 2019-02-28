package database

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
)

var (
	// ErrDuplicateIdentity is returned when trying to associate a remote identity
	// with a user account and the identity is already in use.
	ErrDuplicateIdentity = errors.New("remote identity register with another user")
	// ErrDuplicateGroup is returned when trying to create a group with the same
	// name as a previously registered group.
	ErrDuplicateGroup = errors.New("group name already registered")
	// ErrCourseExists is returned when trying to create an association in
	// the database for a DirectoryID that already exists in the database.
	ErrCourseExists = errors.New("course already exists on git provider")
	// ErrInsufficientAccess is returned when trying to update database
	// with insufficient access priviledges.
	ErrInsufficientAccess = errors.New("user must be admin to perform this operation")
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
		&models.User{},
		&models.RemoteIdentity{},
		&models.Course{},
		&models.Enrollment{},
		&models.Assignment{},
		&models.Submission{},
		&models.Group{},
		&models.Repository{},
	).Error; err != nil {
		return nil, err
	}

	return &GormDB{conn}, nil
}

// GetUser implements the Database interface.
func (db *GormDB) GetUser(uid uint64) (*models.User, error) {
	var user models.User
	if err := db.conn.Preload("RemoteIdentities").First(&user, uid).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByRemoteIdentity implements the Database interface.
func (db *GormDB) GetUserByRemoteIdentity(remote *models.RemoteIdentity) (*models.User, error) {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity models.RemoteIdentity
	if err := tx.
		Where(&models.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get the user.
	var user models.User
	if err := tx.Preload("RemoteIdentities").First(&user, remoteIdentity.UserID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateAccessToken implements the Database interface.
func (db *GormDB) UpdateAccessToken(remote *models.RemoteIdentity) error {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity models.RemoteIdentity
	if err := tx.
		Where(&models.RemoteIdentity{
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

// GetUsers implements the Database interface.
func (db *GormDB) GetUsers(withRemoteIDs bool, uids ...uint64) ([]*models.User, error) {
	m := db.conn
	if len(uids) > 0 {
		m = m.Where(uids)
	}
	if withRemoteIDs {
		m = m.Preload("RemoteIdentities")
	}

	var users []*models.User
	if err := m.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser implements the Database interface
func (db *GormDB) UpdateUser(user *models.User) error {
	return db.conn.Model(&models.User{}).Updates(user).Error
}

// SetAdmin implements the Database interface.
func (db *GormDB) SetAdmin(uid uint64) error {
	var user models.User
	if err := db.conn.First(&user, uid).Error; err != nil {
		return err
	}
	//var admin bool
	admin := true
	user.IsAdmin = &admin
	return db.conn.Save(&user).Error
}

// GetRemoteIdentity implements the Database interface.
func (db *GormDB) GetRemoteIdentity(provider string, rid uint64) (*models.RemoteIdentity, error) {
	var remoteIdentity models.RemoteIdentity
	if err := db.conn.Model(&models.RemoteIdentity{}).
		Where(&models.RemoteIdentity{
			Provider: provider,
			RemoteID: rid,
		}).
		First(&remoteIdentity).Error; err != nil {
		return nil, err
	}
	return &remoteIdentity, nil
}

// CreateUserFromRemoteIdentity implements the Database interface.
func (db *GormDB) CreateUserFromRemoteIdentity(user *models.User, remoteIdentity *models.RemoteIdentity) error {
	user.RemoteIdentities = []*models.RemoteIdentity{remoteIdentity}
	if err := db.conn.Create(&user).Error; err != nil {
		return err
	}
	// The first user defaults to admin user.
	if user.ID == 1 {
		if err := db.SetAdmin(1); err != nil {
			return err
		}
		admin := true
		user.IsAdmin = &admin
	}
	return nil
}

// AssociateUserWithRemoteIdentity implements the Database interface.
func (db *GormDB) AssociateUserWithRemoteIdentity(uid uint64, provider string, rid uint64, accessToken string) error {
	var count uint64
	if err := db.conn.
		Model(&models.RemoteIdentity{}).
		Where(&models.RemoteIdentity{
			Provider: provider,
			RemoteID: rid,
		}).
		Not(&models.RemoteIdentity{
			UserID: uid,
		}).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateIdentity
	}

	var remoteIdentity models.RemoteIdentity
	return db.conn.
		Where(models.RemoteIdentity{Provider: provider, RemoteID: rid, UserID: uid}).
		Assign(models.RemoteIdentity{AccessToken: accessToken}).
		FirstOrCreate(&remoteIdentity).Error
}

// CreateCourse creates the given course and enrolls
// the given user (uid) as teacher for the course.
func (db *GormDB) CreateCourse(uid uint64, course *models.Course) error {
	user, err := db.GetUser(uid)
	if err != nil {
		return err
	}
	if user.IsAdmin == nil || !*user.IsAdmin {
		return ErrInsufficientAccess
	}

	var courses uint64
	// check if course already exists in database
	if err := db.conn.Model(&models.Course{}).
		Where(&models.Course{DirectoryID: course.DirectoryID}).
		Count(&courses).Error; err != nil {
		return err
	}
	if courses > 0 {
		return ErrCourseExists
	}
	//TODO(meling) these db updates should be done as a transaction
	if err := db.conn.Create(course).Error; err != nil {
		return err
	}
	if err := db.CreateEnrollment(&models.Enrollment{UserID: uid, CourseID: course.ID}); err != nil {
		return err
	}
	if err := db.EnrollTeacher(uid, course.ID); err != nil {
		return err
	}
	return nil
}

// GetCourses implements the Database interface.
// If one or more course ids are provided, the corresponding courses
// are returned. Otherwise, all courses are returned.
func (db *GormDB) GetCourses(cids ...uint64) ([]*models.Course, error) {
	m := db.conn
	if len(cids) > 0 {
		m = m.Where(cids)
	}
	var courses []*models.Course
	if err := m.Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

// GetAssignmentsByCourse implements the Database interface
func (db *GormDB) GetAssignmentsByCourse(cid uint64) ([]*models.Assignment, error) {
	var course models.Course
	if err := db.conn.Preload("Assignments").First(&course, cid).Error; err != nil {
		return nil, err
	}
	return course.Assignments, nil
}

// GetNextAssignment returns the next assignment to be approved for
// the given course, user, or group if the next assignment is a group lab.
func (db *GormDB) GetNextAssignment(cid uint64, uid uint64, gid uint64) (*models.Assignment, error) {
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
		var sub *models.Submission
		switch {
		case v.IsGroupLab && gid > 0:
			sub, err = db.GetSubmissionForGroup(v.ID, gid)
		case !v.IsGroupLab && uid > 0:
			sub, err = db.GetSubmissionForUser(v.ID, uid)
		default:
			// This is when uid or gid is set to 0, but there is a group or user lab
			// TODO: Fix so uid always is included and gid is optional
			sub = &models.Submission{Approved: true}
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

// CreateSubmission implements the Database interface
// TODO: Also check enrollment to see if the user is
// enrolled in the course the assignment belongs to
func (db *GormDB) CreateSubmission(submission *models.Submission) error {
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
		m = db.conn.First(&models.User{ID: submission.UserID})
	case submission.GroupID > 0:
		m = db.conn.First(&models.Group{ID: submission.GroupID})
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
	if err := db.conn.Model(&models.Assignment{}).Where(&models.Assignment{
		ID: submission.AssignmentID,
	}).Count(&assignment).Error; err != nil {
		return err
	}

	if assignment+group != 2 {
		return gorm.ErrRecordNotFound
	}
	return db.conn.Create(submission).Error
}

// GetSubmissionForUser implements the Database interface
func (db *GormDB) GetSubmissionForUser(aid uint64, uid uint64) (*models.Submission, error) {
	var submission models.Submission
	if err := db.conn.Where(&models.Submission{AssignmentID: aid, UserID: uid}).Last(&submission).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetSubmissionForGroup implements the Database interface
func (db *GormDB) GetSubmissionForGroup(aid uint64, gid uint64) (*models.Submission, error) {
	var submission models.Submission
	if err := db.conn.Where(&models.Submission{AssignmentID: aid, GroupID: gid}).Last(&submission).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetSubmissionsByID implements the Database interface
func (db *GormDB) GetSubmissionsByID(sid uint64) (*models.Submission, error) {
	var submission models.Submission
	if err := db.conn.First(&submission, sid).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// UpdateSubmissionByID implements the Database interface
func (db *GormDB) UpdateSubmissionByID(sid uint64, approved bool) error {
	sub, err := db.GetSubmissionsByID(sid)
	if err != nil {
		return err
	}
	sub.Approved = approved
	return db.conn.Model(&models.Submission{}).Update(sub).Error
}

// GetSubmissions implements the Database interface
func (db *GormDB) GetSubmissions(cid uint64, uid uint64) ([]*models.Submission, error) {
	var course models.Course
	if err := db.conn.Preload("Assignments").First(&course, cid).Error; err != nil {
		return nil, err
	}

	var latestSubs []*models.Submission
	for _, a := range course.Assignments {
		temp, err := db.GetSubmissionForUser(a.ID, uid)
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

// GetGroupSubmissions implements the Database interface
func (db *GormDB) GetGroupSubmissions(cid uint64, gid uint64) ([]*models.Submission, error) {
	var course models.Course
	if err := db.conn.Preload("Assignments").First(&course, cid).Error; err != nil {
		return nil, err
	}

	var latestSubs []*models.Submission
	for _, a := range course.Assignments {
		temp, err := db.GetSubmissionForGroup(a.ID, gid)
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

// CreateAssignment implements the Database interface
func (db *GormDB) CreateAssignment(assignment *models.Assignment) error {
	// Course id and assignment order must be given.
	if assignment.CourseID < 1 || assignment.Order < 1 {
		return gorm.ErrRecordNotFound
	}

	var course uint64
	if err := db.conn.Model(&models.Course{}).Where(&models.Course{
		ID: assignment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	return db.conn.
		Where(models.Assignment{
			CourseID: assignment.CourseID,
			Order:    assignment.Order,
		}).
		Assign(models.Assignment{
			Name:        assignment.Name,
			Language:    assignment.Language,
			Deadline:    assignment.Deadline,
			AutoApprove: assignment.AutoApprove,
		}).FirstOrCreate(assignment).Error
}

// UpdateAssignments implements the Database interface.
func (db *GormDB) UpdateAssignments(assignments []*models.Assignment) error {
	//TODO Updating the database may need locking??
	for _, v := range assignments {
		// this will create or update an existing assignment
		if err := db.CreateAssignment(v); err != nil {
			return err
		}
	}
	return nil
}

// CreateEnrollment implements the Database interface.
// This method will overwrite the status field with models.Pending.
func (db *GormDB) CreateEnrollment(enrollment *models.Enrollment) error {
	var user, course uint64
	if err := db.conn.Model(&models.User{}).Where(&models.User{
		ID: enrollment.UserID,
	}).Count(&user).Error; err != nil {
		return err
	}
	if err := db.conn.Model(&models.Course{}).Where(&models.Course{
		ID: enrollment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if user+course != 2 {
		return gorm.ErrRecordNotFound
	}

	enrollment.Status = models.Pending
	return db.conn.Create(&enrollment).Error
}

// EnrollStudent implements the Database interface.
func (db *GormDB) EnrollStudent(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, models.Student)
}

// RejectEnrollment implements the Database interface.
func (db *GormDB) RejectEnrollment(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, models.Rejected)
}

// EnrollTeacher implements the Database interface.
func (db *GormDB) EnrollTeacher(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, models.Teacher)
}

// SetPendingEnrollment implements the Database interface.
func (db *GormDB) SetPendingEnrollment(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, models.Pending)
}

// GetEnrollmentsByCourse implements the Database interface.
func (db *GormDB) GetEnrollmentsByCourse(cid uint64, statuses ...uint) ([]*models.Enrollment, error) {
	return db.getEnrollments(&models.Course{ID: cid}, statuses...)
}

func (db *GormDB) getEnrollments(model interface{}, statuses ...uint) ([]*models.Enrollment, error) {
	if len(statuses) == 0 {
		statuses = []uint{models.Pending, models.Rejected, models.Student, models.Teacher}
	}
	var enrollments []*models.Enrollment
	if err := db.conn.Model(model).
		Where("status in (?)", statuses).
		Association("Enrollments").
		Find(&enrollments).Error; err != nil {
		return nil, err
	}

	return enrollments, nil
}

// GetEnrollmentByCourseAndUser return a record of Enrollment
func (db *GormDB) GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*models.Enrollment, error) {
	var enrollment models.Enrollment
	if err := db.conn.
		Where(&models.Enrollment{
			CourseID: cid,
			UserID:   uid,
		}).
		First(&enrollment).Error; err != nil {
		return nil, err
	}
	return &enrollment, nil
}

func (db *GormDB) setEnrollment(uid, cid uint64, status uint) error {
	if status > models.Teacher {
		panic("invalid status")
	}
	var enrollment models.Enrollment
	if err := db.conn.
		Model(&models.Enrollment{}).
		Where(&models.Enrollment{CourseID: cid, UserID: uid}).First(&enrollment).Error; err != nil {
		return err
	}
	enrollment.Status = status
	// Update wont allow value 0 (models.Pending), so need to use save.
	return db.conn.Save(&enrollment).Error
}

// GetCoursesByUser returns all courses (with enrollment status)
// for the given user id.
// If enrollment statuses is provided, the set of courses returned
// is filtered according to these enrollment statuses.
func (db *GormDB) GetCoursesByUser(uid uint64, statuses ...uint) ([]*models.Course, error) {
	enrollments, err := db.getEnrollments(&models.User{ID: uid}, statuses...)
	if err != nil {
		return nil, err
	}

	var courseIDs []uint64
	m := make(map[uint64]*models.Enrollment)
	for _, enrollment := range enrollments {
		m[enrollment.CourseID] = enrollment
		courseIDs = append(courseIDs, enrollment.CourseID)
	}

	if len(statuses) == 0 {
		courseIDs = nil
	} else if len(courseIDs) == 0 {
		// No need to query database since user have no enrolled courses.
		return []*models.Course{}, nil
	}
	courses, err := db.GetCourses(courseIDs...)
	if err != nil {
		return nil, err
	}

	for _, course := range courses {
		course.Enrolled = models.None
		if enrollment, ok := m[course.ID]; ok {
			course.Enrolled = int(enrollment.Status)
		}
	}
	return courses, nil
}

// GetCourse implements the Database interface
func (db *GormDB) GetCourse(cid uint64) (*models.Course, error) {
	var course models.Course
	if err := db.conn.First(&course, cid).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourseByDirectoryID implements the Database interface
func (db *GormDB) GetCourseByDirectoryID(did uint64) (*models.Course, error) {
	var course models.Course
	if err := db.conn.First(&course, &models.Course{DirectoryID: did}).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// UpdateCourse implements the Database interface
func (db *GormDB) UpdateCourse(course *models.Course) error {
	return db.conn.Model(&models.Course{}).Updates(course).Error
}

// CreateGroup creates a new group and assign users to newly created group
func (db *GormDB) CreateGroup(group *models.Group) error {
	if group.CourseID == 0 {
		return gorm.ErrRecordNotFound
	}

	tx := db.conn.Begin()
	var course uint64
	if err := db.conn.Model(&models.Course{}).Where(&models.Course{
		ID: group.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	if err := tx.Model(&models.Group{}).Create(group).Error; err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateGroup
		}
		return err
	}
	var userids []uint64
	for _, u := range group.Users {
		userids = append(userids, u.ID)
	}
	query := tx.Model(&models.Enrollment{}).
		Where(&models.Enrollment{
			CourseID: group.CourseID,
		}).
		Where("user_id IN (?) AND status IN (?)", userids, []uint{models.Student, models.Teacher}).
		Updates(&models.Enrollment{
			GroupID: group.ID,
		})

	if query.Error != nil {
		tx.Rollback()
		return query.Error
	}

	if query.RowsAffected != int64(len(userids)) {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	tx.Commit()
	return nil
}

// GetGroup implements the Database interface
func (db *GormDB) GetGroup(withRemoteIDs bool, gid uint64) (*models.Group, error) {
	var group models.Group
	if err := db.conn.Preload("Enrollments").First(&group, gid).Error; err != nil {
		return nil, err
	}
	var userIDs []uint64
	for _, enrollment := range group.Enrollments {
		userIDs = append(userIDs, enrollment.UserID)
	}
	if len(userIDs) > 0 {
		users, err := db.GetUsers(withRemoteIDs, userIDs...)
		if err != nil {
			return nil, err
		}
		group.Users = users
	}
	return &group, nil
}

// UpdateGroupStatus updates status field of a group
func (db *GormDB) UpdateGroupStatus(group *models.Group) error {
	return db.conn.Model(group).Update("status", group.Status).Error
}

// GetGroupsByCourse returns a list of groups
func (db *GormDB) GetGroupsByCourse(cid uint64) ([]*models.Group, error) {
	//TODO(meling) add test for this method
	//TODO(meling) can this also Preload("Users") to avoid the GetUsers below.
	var groups []*models.Group
	if err := db.conn.
		Preload("Enrollments").
		Where(&models.Group{
			CourseID: cid,
		}).
		Find(&groups).Error; err != nil {
		return nil, err
	}

	for _, group := range groups {
		var userIDs []uint64
		for _, enrollment := range group.Enrollments {
			userIDs = append(userIDs, enrollment.UserID)
		}
		if len(userIDs) > 0 {
			users, err := db.GetUsers(false, userIDs...)
			if err != nil {
				return nil, err
			}
			group.Users = users
		}
	}
	return groups, nil
}

// DeleteGroup delete a group
func (db *GormDB) DeleteGroup(gid uint64) error {
	group, err := db.GetGroup(false, gid)
	if err != nil {
		return err
	}

	tx := db.conn.Begin()
	if err := tx.Delete(group).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("UPDATE enrollments SET group_id= ? WHERE group_id= ?", 0, gid).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// CreateRepository implements the Database interface
func (db *GormDB) CreateRepository(repo *models.Repository) error {
	if repo.DirectoryID == 0 || repo.RepositoryID == 0 {
		return gorm.ErrRecordNotFound
	}

	if repo.UserID > 0 {
		err := db.conn.First(&models.User{}, repo.UserID).Error
		if err != nil {
			return err
		}
	}

	return db.conn.Create(repo).Error
}

// GetRepository imlements the Database interface
func (db *GormDB) GetRepository(rid uint64) (*models.Repository, error) {
	// This uses the repositoryid from the provider to search with, and not
	// the id of the entry in the database
	var repo models.Repository

	if err := db.conn.First(&repo, &models.Repository{RepositoryID: rid}).Error; err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepositoriesByCourseIDandUserID Fetches Repo based on courseid, userid and type
func (db *GormDB) GetRepositoriesByCourseIDandUserID(cid uint64, uid uint64) (*models.Repository, error) {
	course, err := db.GetCourse(cid)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	var repo models.Repository
	if err := db.conn.First(&repo, &models.Repository{DirectoryID: course.DirectoryID, UserID: uid}).Error; err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepoByCourseIDUserIDandType Fetches Repo based on courseid, userid and type
func (db *GormDB) GetRepoByCourseIDUserIDandType(cid uint64, uid uint64, repoType models.RepoType) (*models.Repository, error) {
	course, err := db.GetCourse(cid)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	var repo models.Repository
	if err := db.conn.First(&repo, &models.Repository{DirectoryID: course.DirectoryID, UserID: uid, Type: repoType}).Error; err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepositoriesByDirectory implements the database interface
func (db *GormDB) GetRepositoriesByDirectory(did uint64) ([]*models.Repository, error) {

	var repos []*models.Repository
	if err := db.conn.Find(&repos, &models.Repository{DirectoryID: did}).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

// UpdateGroup updates a group
func (db *GormDB) UpdateGroup(group *models.Group) error {
	if group.CourseID == 0 {
		return gorm.ErrRecordNotFound
	}
	tx := db.conn.Begin()
	var course uint64
	if err := db.conn.Model(&models.Course{}).Where(&models.Course{
		ID: group.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}
	if err := tx.Model(&models.Group{}).Updates(group).Error; err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateGroup
		}
		return err
	}
	if err := tx.Exec("UPDATE enrollments SET group_id= ? WHERE group_id= ?", 0, group.ID).Error; err != nil {
		tx.Rollback()
		return err
	}
	var userids []uint64
	for _, u := range group.Users {
		userids = append(userids, u.ID)
	}
	query := tx.Model(&models.Enrollment{}).
		Where(&models.Enrollment{
			CourseID: group.CourseID,
		}).
		Where("user_id IN (?) AND status IN (?)", userids, []uint{models.Student, models.Teacher}).
		Updates(&models.Enrollment{
			GroupID: group.ID,
		})

	if query.Error != nil {
		tx.Rollback()
		return query.Error
	}

	if query.RowsAffected != int64(len(userids)) {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	tx.Commit()
	return nil
}

// GetRepositoriesByCourseIDAndType returns repos beloning to directoryID and with repo type
func (db *GormDB) GetRepositoriesByCourseIDAndType(cid uint64, repoType models.RepoType) ([]*models.Repository, error) {

	course, err := db.GetCourse(cid)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	var repos []*models.Repository
	if err := db.conn.Find(&repos, &models.Repository{DirectoryID: course.DirectoryID, Type: repoType}).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

// Close closes the gorm database.
func (db *GormDB) Close() error {
	return db.conn.Close()
}
