package web

import (
	"context"
	"errors"
	"strconv"

	"github.com/autograde/quickfeed/admin"
	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
	scms "github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web/config"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AdminService struct {
	logger *zap.SugaredLogger
	db     database.Database
	app    *scms.GithubApp
	Config *config.Config // TODO(vera): make unexported again after refactoring the startup method
	admin.UnimplementedAdminServiceServer
}

func NewAdminService(logger *zap.Logger, db database.Database, app *scm.GithubApp, config *config.Config) *AdminService {
	return &AdminService{
		logger: logger.Sugar(),
		db:     db,
		app:    app,
		Config: config,
	}
}

// GetUsers returns a list of all users.
// Frontend note: This method is called from AdminPage.
func (s *AdminService) GetUsers(ctx context.Context, _ *pb.Void) (*pb.Users, error) {
	users, err := s.getUsers()
	if err != nil {
		s.logger.Errorf("GetUsers failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get users")
	}
	return users, nil
}

// CreateCourse creates a new course.
// Access policy: Admin.
// TODO(vera): instead of calling getUserAndSCM here we want to fetch the app installations, choose the correct installation
// for the given course org and create a new scm for this course, because there will be no scm client for the course at this point.
func (s *AdminService) CreateCourse(ctx context.Context, in *pb.Course) (*pb.Course, error) {
	usr, scm, err := s.getUserAndSCM(ctx, in.GetID())
	if err != nil {
		s.logger.Errorf("CreateCourse failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsAdmin {
		s.logger.Error("CreateCourse failed: user is not admin")
		return nil, status.Error(codes.PermissionDenied, "user must be admin to create course")
	}

	// make sure that the current user is set as course creator
	in.CourseCreatorID = usr.GetID()
	course, err := s.createCourse(ctx, scm, in)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: %v", err)
		// errors informing about requested organization state will have code 9: FailedPrecondition
		// error message will be displayed to the user
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == ErrAlreadyExists || err == ErrFreePlan {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to create course")
	}
	return course, nil
}

// TODO(vera): these methods are temporary duplicate and will be removed,
// all services will need to somehow fetch current user and the proper scm client.
// The goal is to do it some other way. User info will be in
// the JWT token, scm client will "belong" to a course, not user

func (s *AdminService) getUserAndSCM(ctx context.Context, courseID uint64) (*pb.User, scm.SCM, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		return nil, nil, err
	}
	scm, err := s.getSCM(courseID)
	if err != nil {
		return nil, nil, err
	}
	return usr, scm, nil
}

func (s *AdminService) getCurrentUser(ctx context.Context) (*pb.User, error) {
	// process user id from context
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("malformed request")
	}
	userValues := meta.Get("user")
	if len(userValues) == 0 {
		return nil, errors.New("no user metadata in context")
	}
	if len(userValues) != 1 || userValues[0] == "" {
		return nil, errors.New("invalid user payload in context")
	}
	userID, err := strconv.ParseUint(userValues[0], 10, 64)
	if err != nil {
		return nil, err
	}
	// return the user corresponding to userID, or an error.
	return s.db.GetUser(userID)
}

// TODO(vera): repurpose for new scm type (or two scm types)
func (s *AdminService) getSCM(courseID uint64) (scm.SCM, error) {
	sc, ok := s.app.GetSCM(courseID)
	if ok {
		return sc, nil
	}
	return nil, errors.New("no SCM found")
}

// TODO(vera): this method must work for any service, all services must share same scm list
// MakeSCMClients creates a new scm client (GitHub app installation based client) for each course in the database.
// This method is called at the server start.
func (s *AdminService) MakeSCMClients(provider string) error {
	ctx := context.Background()
	courses, err := s.db.GetCourses()
	if err != nil {
		return err
	}
	for _, course := range courses {
		courseCreator, err := s.db.GetUser(course.CourseCreatorID)
		if err != nil {
			return err
		}
		ghClient, err := s.app.NewInstallationClient(ctx, course.OrganizationPath)
		if err != nil {
			return err
		}
		token, err := courseCreator.GetAccessToken(provider)
		if err != nil {
			return err
		}

		sc, err := scm.NewSCMClient(s.logger, ghClient, provider, token)
		if err != nil {
			return err
		}
		s.app.AddSCM(sc, course.ID)
	}
	return nil
}
