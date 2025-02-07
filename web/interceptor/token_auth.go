package interceptor

import (
	"context"
	"errors"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
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
	mu       sync.Mutex
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
		token := conn.RequestHeader().Get(tokenHeader)
		if len(token) == 0 {
			return next(ctx, conn)
		}

		cookie, err := t.lookupToken(token)
		if err != nil {
			return err
		}

		conn.RequestHeader().Set(auth.Cookie, cookie)
		if err = next(ctx, conn); err != nil {
			return err
		}
		updatedCookie := conn.ResponseHeader().Get(auth.SetCookie)
		if len(updatedCookie) != 0 && updatedCookie != cookie {
			t.update(token, updatedCookie)
		}
		return nil
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

		cookie, err := t.lookupToken(token)
		if err != nil {
			return nil, err
		}

		request.Header().Set(auth.Cookie, cookie)
		response, err := next(ctx, request)
		if response != nil {
			updatedCookie := response.Header().Get(auth.SetCookie)
			if len(updatedCookie) != 0 && updatedCookie != cookie {
				t.update(token, updatedCookie)
			}
		}
		return response, err
	})
}

func (t *TokenAuthInterceptor) lookup(token string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cookie, exists := t.tokenMap[token]
	return cookie, exists
}

func (t *TokenAuthInterceptor) update(token, cookie string) {
	t.mu.Lock()
	t.tokenMap[token] = cookie
	t.mu.Unlock()
}

// lookupToken checks if a given token exists in the tokenMap. If it does
// not, it will attempt to query GitHub for user information associated
// with the token. If a user exists for the token, we verify that the user
// exists in our database, and create a cookie with claims for the user.
func (t *TokenAuthInterceptor) lookupToken(token string) (string, error) {
	if cookie, exists := t.lookup(token); exists {
		return cookie, nil
	}

	// Verify that token has correct prefixes before continuing
	if !(strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "github_pat_")) {
		// could also pass through for next interceptor to determine if the request
		// has a valid cookie
		return "", connect.NewError(connect.CodeInvalidArgument, errors.New("invalid token"))
	}

	// Attempt to fetch user from GitHub using provided token
	externalUser, err := auth.FetchExternalUser(&oauth2.Token{
		AccessToken: token,
	})
	if err != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, err)
	}
	t.logger.Debug("Retrieved user", externalUser)
	// Fetch user from database using the remote identity received from GitHub.
	user, err := t.db.GetUserByRemoteIdentity(externalUser.ID)
	if err != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Create a new authentication cookie, which contains
	// claims for the user associated with the token
	// received in the request
	cookie, err := t.tm.NewAuthCookie(user.ID)
	if err != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Store the generated cookie in our token map
	cookieString := cookie.String()
	t.update(token, cookieString)
	return cookieString, nil
}
