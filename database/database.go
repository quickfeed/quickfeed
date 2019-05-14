package database

import (
	pb "github.com/autograde/aguis/ag"
)

// Database contains methods for manipulating the database.
type Database interface {
	GetRemoteIdentity(provider string, rid uint64) (*pb.RemoteIdentity, error)

	CreateUserFromRemoteIdentity(*pb.User, *pb.RemoteIdentity) error
	AssociateUserWithRemoteIdentity(uid uint64, provider string, rid uint64, accessToken string) error

	GetUser(uint64) (*pb.User, error)
	// GetUserByRemoteIdentity gets an user by a remote identity and updates the access token.
	// TODO: The update access token functionality should be split into its own method.
	GetUserByRemoteIdentity(provider string, rid uint64, accessToken string) (*pb.User, error)
	GetUsers(...uint64) ([]*pb.User, error)
	UpdateUser(*pb.User) error

	// SetAdmin makes an existing user an administrator. The admin role is allowed to
	// create courses, so it makes sense that teachers are made admins.
	SetAdmin(uint64) error

	CreateCourse(uint64, *pb.Course) error
	GetCourse(uint64) (*pb.Course, error)
	GetCourseByDirectoryID(did uint64) (*pb.Course, error)
	GetCourses(...uint64) ([]*pb.Course, error)
	GetCoursesByUser(uid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Course, error)
	UpdateCourse(*pb.Course) error

	CreateEnrollment(*pb.Enrollment) error
	RejectEnrollment(uid uint64, cid uint64) error
	EnrollStudent(uid uint64, cid uint64) error
	EnrollTeacher(uid uint64, cid uint64) error
	SetPendingEnrollment(uid, cid uint64) error
	GetEnrollmentsByCourse(cid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error)
	GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*pb.Enrollment, error)

	CreateAssignment(*pb.Assignment) error
	GetAssignmentsByCourse(uint64) ([]*pb.Assignment, error)
	GetNextAssignment(cid, uid, gid uint64) (*pb.Assignment, error)

	CreateSubmission(*pb.Submission) error
	GetSubmissionForUser(aid uint64, uid uint64) (*pb.Submission, error)
	GetSubmissionForGroup(aid uint64, gid uint64) (*pb.Submission, error)
	GetSubmissions(cid uint64, uid uint64) ([]*pb.Submission, error)
	GetGroupSubmissions(cid uint64, gid uint64) ([]*pb.Submission, error)
	GetSubmissionsByID(sid uint64) (*pb.Submission, error)
	UpdateSubmissionByID(sid uint64, approved bool) error

	CreateGroup(*pb.Group) error
	GetGroup(uint64) (*pb.Group, error)
	GetGroupsByCourse(cid uint64) ([]*pb.Group, error)
	UpdateGroupStatus(*pb.Group) error
	UpdateGroup(group *pb.Group) error
	DeleteGroup(uint64) error

	CreateRepository(repo *pb.Repository) error
	GetRepository(uint64) (*pb.Repository, error)
	GetRepositoriesByDirectory(uint64) ([]*pb.Repository, error)
	GetRepositoriesByCourseIDAndType(uint64, pb.Repository_RepoType) ([]*pb.Repository, error)
	GetRepositoriesByCourseIDandUserID(uint64, uint64) (*pb.Repository, error)
	GetRepoByCourseIDUserIDandType(uint64, uint64, pb.Repository_RepoType) (*pb.Repository, error)
}
