package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"go.uber.org/zap"
)

// MockClient is a mock QuickFeed client.
type MockClient struct {
	qfconnect.QuickFeedServiceClient
	tm *auth.TokenManager
}

// Cookie returns an authentication cookie for the given user.
func (m *MockClient) Cookie(t *testing.T, user *qf.User) string {
	t.Helper()
	cookie, err := m.tm.NewAuthCookie(user.GetID())
	if err != nil {
		t.Fatal(err)
	}
	return cookie.String()
}

// TokenManager returns the underlying TokenManager for advanced test scenarios.
// This is primarily used by tests that need direct access to the token manager,
// such as the token refresh tests.
func (m *MockClient) TokenManager() *auth.TokenManager {
	return m.tm
}

type MockClientOptions struct {
	clientOptions    []connect.ClientOption
	interceptorFuncs []InterceptorFunc
}

// InterceptorFunc is a function that creates an interceptor.
// It receives logger, token manager, and database, but can ignore unused parameters with _.
type InterceptorFunc func(logger *zap.SugaredLogger, tm *auth.TokenManager, db database.Database) connect.Interceptor

type MockClientOption func(*MockClientOptions)

// WithClientOptions sets the client options for the mock client.
func WithClientOptions(opts ...connect.ClientOption) MockClientOption {
	return func(o *MockClientOptions) {
		o.clientOptions = opts
	}
}

// WithInterceptors configures custom interceptors using the provided constructor functions.
// If no interceptor functions are provided, the standard set of auth interceptors is used.
// The TokenManager is automatically created and shared between interceptors.
//
// Example usage with standard interceptors:
//
//	client := NewMockClient(t, db, scmOpt, WithInterceptors())
//
// Which is the equivalent to these custom interceptors:
//
//	client := NewMockClient(t, db, scmOpt,
//		WithInterceptors(
//			ValidationInterceptorFunc,
//			TokenAuthInterceptorFunc,
//			UserInterceptorFunc,
//			AccessControlInterceptorFunc,
//		),
//	)
//
// Example usage with custom interceptors:
//
//	client := NewMockClient(t, db, scmOpt,
//		WithInterceptors(
//			ValidationInterceptorFunc,
//			UserInterceptorFunc,
//			AccessControlInterceptorFunc,
//		),
//	)
func WithInterceptors(funcs ...InterceptorFunc) MockClientOption {
	return func(o *MockClientOptions) {
		if len(funcs) > 0 {
			o.interceptorFuncs = funcs
		} else {
			o.interceptorFuncs = []InterceptorFunc{
				ValidationInterceptorFunc,
				TokenAuthInterceptorFunc,
				UserInterceptorFunc,
				AccessControlInterceptorFunc,
			}
		}
	}
}

// Individual interceptor functions that match InterceptorFunc signature
func ValidationInterceptorFunc(logger *zap.SugaredLogger, _ *auth.TokenManager, _ database.Database) connect.Interceptor {
	return interceptor.NewValidationInterceptor(logger)
}

func TokenAuthInterceptorFunc(logger *zap.SugaredLogger, tm *auth.TokenManager, db database.Database) connect.Interceptor {
	return interceptor.NewTokenAuthInterceptor(logger, tm, db)
}

func UserInterceptorFunc(logger *zap.SugaredLogger, tm *auth.TokenManager, _ database.Database) connect.Interceptor {
	return interceptor.NewUserInterceptor(logger, tm)
}

func AccessControlInterceptorFunc(_ *zap.SugaredLogger, tm *auth.TokenManager, _ database.Database) connect.Interceptor {
	return interceptor.NewAccessControlInterceptor(tm)
}

func TokenInterceptorFunc(_ *zap.SugaredLogger, tm *auth.TokenManager, _ database.Database) connect.Interceptor {
	return interceptor.NewTokenInterceptor(tm)
}

// NewMockClient returns a QuickFeed client for invoking RPCs.
func NewMockClient(t *testing.T, db database.Database, scmOpt scm.MockOption, opts ...MockClientOption) *MockClient {
	t.Helper()
	options := &MockClientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	mgr := scm.MockManager(t, scmOpt)
	logger := qtest.Logger(t)
	qfService := NewQuickFeedService(logger.Desugar(), db, mgr, BaseHookOptions{}, &ci.Local{})

	// Create token manager when needed
	var tm *auth.TokenManager
	var interceptors []connect.Interceptor
	if len(options.interceptorFuncs) > 0 {
		var err error
		tm, err = auth.NewTokenManager(db)
		if err != nil {
			t.Fatal(err)
		}
		// Build custom interceptors from the provided interceptor constructor functions
		for _, createInterceptor := range options.interceptorFuncs {
			interceptors = append(interceptors, createInterceptor(logger, tm, db))
		}
	}

	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(qfService, connect.WithInterceptors(interceptors...)))
	server := httptest.NewUnstartedServer(router)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	client := qfconnect.NewQuickFeedServiceClient(server.Client(), server.URL, options.clientOptions...)
	return &MockClient{
		QuickFeedServiceClient: client,
		tm:                     tm,
	}
}
