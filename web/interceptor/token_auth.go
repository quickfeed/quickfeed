package interceptor

import (
	"context"
	"errors"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const tokenHeader = "Authorization"

type TokenAuthInterceptor struct {
	tm       *auth.TokenManager
	logger   *zap.SugaredLogger
	db       database.Database
	tokenMap map[string]string
}

func NewTokenAuthInterceptor(logger *zap.SugaredLogger, tm *auth.TokenManager, db database.Database) *TokenAuthInterceptor {
	return &TokenAuthInterceptor{
		tm:       tm,
		logger:   logger,
		db:       db,
		tokenMap: make(map[string]string),
	}
}

func (t *TokenAuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (*TokenAuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

func (t *TokenAuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		token := request.Header().Get(tokenHeader)
		if len(token) == 0 {
			return next(ctx, request)
		}

		if cookie, exists := t.tokenMap[token]; exists {
			request.Header().Set(auth.Cookie, cookie)
			response, err := next(ctx, request)

			if response != nil {
				updatedCookie := response.Header().Get(auth.SetCookie)
				if len(updatedCookie) != 0 && updatedCookie != cookie {
					t.tokenMap[token] = updatedCookie
				}
			}

			return response, err
		}

		// Verify that token has correct prefixes before continuing
		if !(strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "github_pat_")) {
			// could also pass through for next interceptor to determine if the request
			// has a valid cookie
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid token"))
		}

		// Attempt to fetch user from GitHub using provided token
		externalUser, err := auth.FetchExternalUser(&oauth2.Token{
			AccessToken: token,
		})
		if err != nil {
			// Abort if any error occurs
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}
		t.logger.Debug("Retrieved user", externalUser)
		// Fetch user from database using the remote identity received
		// from GitHub.
		user, err := t.db.GetUserByRemoteIdentity(&qf.RemoteIdentity{
			ID: externalUser.ID,
			// Unsure if required
			Provider: env.ScmProvider(),
		})
		if err != nil {
			// Abort if any error occurs
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		// Create a new authentication cookie, which contains
		// claims for the user associated with the token
		// received in the request
		cookie, err := t.tm.NewAuthCookie(user.ID)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		// Store the generated cookie in our token map
		t.tokenMap[token] = cookie.String()

		// Set the cookie to the request header for consumption
		// in subsequent interceptors in the chain
		request.Header().Set(auth.Cookie, cookie.String())
		return next(ctx, request)
	})
}
