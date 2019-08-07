package ag

import (
	"context"
	"reflect"
	"time"

	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MaxWait is the maximum time a request is allowed to stay open before aborting.
const MaxWait = 2 * time.Minute

type validator interface {
	IsValid() bool
}

type idCleaner interface {
	RemoveRemoteID()
}

// Interceptor returns a new unary server interceptor that validates requests
// that implements the validator interface.
// Invalid requests are rejected without logging and before it reaches any
// user-level code and returns an illegal argument to the client.
// In addition, the interceptor also implements a cancel mechanism.
func Interceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(validator); ok {
			if !v.IsValid() {
				return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
			}
		} else {
			// just logging, but still handling the call
			logger.Sugar().Debugf("message type '%s' does not implement validator interface",
				reflect.TypeOf(req).String())
		}
		ctx, cancel := context.WithTimeout(ctx, MaxWait)
		defer cancel()
		// if response has information on remote ID, it will be removed
		resp, err := handler(ctx, req)
		if resp != nil {
			if v, ok := resp.(idCleaner); ok {
				v.RemoveRemoteID()
			}
		}
		return resp, err
	}
}

// IsValid on void message always returns true.
func (v Void) IsValid() bool {
	return true
}

// IsValid checks required fields of a group request
func (grp Group) IsValid() bool {
	return grp.GetName() != "" && grp.GetCourseID() > 0
}

// IsValid checks required fields of a course request
func (c Course) IsValid() bool {
	return c.GetName() != "" &&
		c.GetCode() != "" &&
		(c.GetProvider() == "github" || c.GetProvider() == "gitlab" || c.GetProvider() == "fake") &&
		c.GetOrganizationID() != 0 &&
		c.GetYear() != 0 &&
		c.GetTag() != ""
}

// IsValid chacks required fields of a user request
func (u User) IsValid() bool {
	return u.GetID() > 0
}

// IsValid checks required fields of an enrollment request.
func (req Enrollment) IsValid() bool {
	return req.GetStatus() <= Enrollment_TEACHER &&
		req.GetUserID() > 0 && req.GetCourseID() > 0
}

// IsValid checks whether RecordRequest fields are valid
func (req RecordRequest) IsValid() bool {
	return req.GetID() > 0
}

// IsValid checks that all requested repo types are valid types and course ID field is set
func (req URLRequest) IsValid() bool {
	if req.GetCourseID() < 1 {
		return false
	}
	for _, r := range req.GetRepoTypes() {
		if r <= Repository_NONE {
			return false
		}
	}
	return true
}

// IsValid checks required fields of an action request.
// It must have a positive course ID and
// a positive user ID or group ID but not both.
func (req SubmissionRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return req.GetCourseID() > 0 &&
		(uid == 0 && gid > 0) ||
		(uid > 0 && gid == 0)
}

// IsValid ensures that both submission and course IDs are set
func (req ApproveSubmissionRequest) IsValid() bool {
	return req.GetCourseID() > 0 && req.GetSubmissionID() > 0
}

// IsValid ensures that course ID and group or user IDs are set
func (req GroupRequest) IsValid() bool {
	uid := req.GetUserID()
	gid := req.GetGroupID()
	return (uid > 0 || gid > 0) && req.GetCourseID() > 0
}

// IsValid checks that course ID is positive.
func (req EnrollmentRequest) IsValid() bool {
	return req.GetCourseID() > 0
}

// IsValid ensures that provider string is one of implemented providers
func (req Provider) IsValid() bool {
	provider := req.GetProvider()
	return provider == "github" ||
		provider == "gitlab" ||
		provider == "fake"
}

// IsValidProvider validates provider string coming from front end
func (l Providers) IsValidProvider(provider string) bool {
	isValid := false
	for _, p := range l.GetProviders() {
		if p == provider {
			isValid = true
		}
	}
	return isValid
}
