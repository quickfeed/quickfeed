package interceptor

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
)

// detachTimeout bounds handler execution for detached methods, replacing the
// cancellation signal from the client connection. It must be generous enough
// to allow several sequential GitHub API calls, e.g. one per group member.
const detachTimeout = 2 * time.Minute

// detachedMethods lists methods that perform multi-step SCM and database
// mutations without rollback support. Aborting such a method midway, e.g.,
// because the client closed the connection, can leave the GitHub organization
// and the database in inconsistent states. Once started, these methods should
// be allowed to complete. A zero timeout means that the handler owns its timeout
// policy instead of using an interceptor-level deadline.
var detachedMethods = map[string]time.Duration{
	"UpdateGroup":        detachTimeout, // may create the group repo and add/remove collaborators on GitHub
	"DeleteGroup":        detachTimeout, // deletes the GitHub repo before removing database records
	"UpdateEnrollments":  detachTimeout, // may create/delete user repos and update org membership before database records
	"RebuildSubmissions": 0,             // each submission uses the assignment's container timeout
}

// DetachInterceptor detaches handlers for methods in detachedMethods from the
// connection's context: context values, such as user claims, are preserved,
// but the client's cancellation is ignored. Methods with a configured timeout
// run under that server-controlled deadline; other methods manage their own
// deadlines. A disconnected client never receives the response, but the server
// completes the operation, keeping GitHub and the database consistent.
type DetachInterceptor struct{}

func NewDetachInterceptor() *DetachInterceptor {
	return &DetachInterceptor{}
}

func (*DetachInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := request.Spec().Procedure
		method := procedure[strings.LastIndex(procedure, "/")+1:]
		if timeout, exists := detachedMethods[method]; exists {
			ctx = context.WithoutCancel(ctx)
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
		}
		return next(ctx, request)
	})
}

// WrapStreamingHandler leaves streams attached to the connection's context;
// SubmissionStream relies on cancellation to terminate when the client disconnects.
func (*DetachInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func (*DetachInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}
