package ag

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MaxWait is the maximum time a request is allowed to stay open before aborting.
const MaxWait = 2 * time.Minute
const Cookie = "cookie"
const UserKey = "user"

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
func Interceptor(logger *zap.Logger, userMap map[string]uint64) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		methodName := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		AgMethodSuccessRateMetric.WithLabelValues(methodName, "total").Inc()
		responseTimer := prometheus.NewTimer(prometheus.ObserverFunc(
			AgResponseTimeByMethodsMetric.WithLabelValues(methodName).Set),
		)
		defer responseTimer.ObserveDuration().Milliseconds()

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

		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("Could not grab metadata from context")
		}
		meta, err := UserValidation(meta, userMap)
		if err != nil {
			return nil, err
		}

		ctx = metadata.NewOutgoingContext(ctx, meta)

		// if response has information on remote ID, it will be removed
		resp, err := handler(ctx, req)
		if resp != nil {
			AgMethodSuccessRateMetric.WithLabelValues(methodName, "success").Inc()
			if v, ok := resp.(idCleaner); ok {
				v.RemoveRemoteID()
			}
		}
		if err != nil {
			AgFailedMethodsMetric.WithLabelValues(methodName).Inc()
			AgMethodSuccessRateMetric.WithLabelValues(methodName, "error").Inc()
		}
		return resp, err
	}
}

// Returns modified metadata containing a valid user. Returns an error if the user is not authenticated.
func UserValidation(meta metadata.MD, userMap map[string]uint64) (metadata.MD, error) {
	token := meta.Get(Cookie)

	if len(token) > 0 {
		token := token[0]
		user := userMap[token]
		if user == 0 {
			return nil, errors.New("Could not associate token with a user")
		}
		meta.Set(UserKey, strconv.FormatUint(user, 10))
	} else {
		return nil, errors.New("Request does not contain a session token")
	}
	return meta, nil
}

// IsValid on void message always returns true.
func (v *Void) IsValid() bool {
	return true
}

// IsValid checks required fields of a group request
func (grp *Group) IsValid() bool {
	return grp.GetName() != "" && grp.GetCourseID() > 0
}

// IsValid checks required fields of a course request
func (c *Course) IsValid() bool {
	return c.GetName() != "" &&
		c.GetCode() != "" &&
		(c.GetProvider() == "github" || c.GetProvider() == "gitlab" || c.GetProvider() == "fake") &&
		c.GetOrganizationID() != 0 &&
		c.GetYear() != 0 &&
		c.GetTag() != ""
}

// IsValid checks required fields of a user request
func (u *User) IsValid() bool {
	return u.GetID() > 0
}

// IsValid ensures that user ID is set
func (u *UserRequest) IsValid() bool {
	return u.GetUserID() > 0
}

// IsValid checks required fields of an enrollment request.
func (req *Enrollment) IsValid() bool {
	return req.GetStatus() <= Enrollment_TEACHER &&
		req.GetUserID() > 0 && req.GetCourseID() > 0
}

// IsValid ensures that course ID is set
func (req *CourseRequest) IsValid() bool {
	return req.GetCourseID() > 0
}

// IsValid ensures that user ID is set
func (req *EnrollmentStatusRequest) IsValid() bool {
	return req.GetUserID() > 0
}

// IsValid checks whether OrgRequest fields are valid
func (req *OrgRequest) IsValid() bool {
	return req.GetOrgName() != ""
}

// IsValid checks that all requested repo types are valid types and course ID field is set
func (req *URLRequest) IsValid() bool {
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

// IsValid checks that the request has positive course ID
// and either user ID or group ID is set
func (req *RepositoryRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return req.GetCourseID() > 0 &&
		(uid == 0 && gid > 0) ||
		(uid > 0 && gid == 0)
}

// IsValid checks required fields of an action request.
// It must have a positive course ID and
// a positive user ID or group ID but not both.
func (req *SubmissionRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return req.GetCourseID() > 0 &&
		(uid == 0 && gid > 0) ||
		(uid > 0 && gid == 0)
}

// IsValid ensures that both submission and course IDs are set
func (req *UpdateSubmissionRequest) IsValid() bool {
	return req.GetCourseID() > 0 && req.GetSubmissionID() > 0
}

// IsValid ensures that group ID is provided
func (req *GetGroupRequest) IsValid() bool {
	return req.GetGroupID() > 0
}

// IsValid ensures that course ID and group or user IDs are set
func (req *GroupRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return (uid > 0 || gid > 0) && req.GetCourseID() > 0
}

// IsValid checks that course ID is positive.
func (req *EnrollmentRequest) IsValid() bool {
	return req.GetCourseID() > 0
}

// IsValid ensures that provider string is one of implemented providers
func (req *Provider) IsValid() bool {
	provider := req.GetProvider()
	return provider == "github" ||
		provider == "gitlab" ||
		provider == "fake"
}

// IsValid ensures that course ID is provided
func (req *SubmissionsForCourseRequest) IsValid() bool {
	return req.GetCourseID() != 0
}

// IsValid ensures that both course and submission IDs are set
func (req *RebuildRequest) IsValid() bool {
	aid, sid := req.GetAssignmentID(), req.GetSubmissionID()
	return aid > 0 && sid > 0
}

// IsValid checks that either ID or path field is set
func (org *Organization) IsValid() bool {
	id, path := org.GetID(), org.GetPath()
	return id > 0 || path != ""
}

// IsValidProvider validates provider string coming from front end
func (l *Providers) IsValidProvider(provider string) bool {
	isValid := false
	for _, p := range l.GetProviders() {
		if p == provider {
			isValid = true
		}
	}
	return isValid
}

// IsValid ensures that course ID and submission ID are present.
func (req *SubmissionReviewersRequest) IsValid() bool {
	return req.CourseID > 0 && req.SubmissionID > 0
}

// IsValid ensures that a review always has a reviewer and a submission IDs.
func (r *Review) IsValid() bool {
	return r.ReviewerID > 0 && r.SubmissionID > 0
}

// IsValid ensures that course ID is provided and the review is valid.
func (r *ReviewRequest) IsValid() bool {
	return r.CourseID > 0 && r.Review.IsValid()
}

// IsValid ensures that a grading benchmark always belongs to an assignment
// and is not empty.
func (bm *GradingBenchmark) IsValid() bool {
	return bm.AssignmentID > 0 && bm.Heading != ""
}

// IsValid ensures that a criterion always belongs to a grading benchmark
// and is not empty.
func (c *GradingCriterion) IsValid() bool {
	return c.BenchmarkID > 0 && c.Description != ""
}

// IsValid ensures that course code, year, and student login are set
func (r *CourseUserRequest) IsValid() bool {
	return r.CourseCode != "" && r.UserLogin != "" && r.CourseYear > 2019
}
