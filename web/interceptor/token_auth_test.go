package interceptor_test

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

func TestTokenPrefixValidation(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		tokenPrefix string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid ghp_ token prefix",
			tokenPrefix: "ghp_",
			wantErr:     true,              // Will fail because token is invalid, but prefix is accepted
			errContains: "unauthenticated", // GitHub will reject the fake token
		},
		{
			name:        "valid github_pat_ token prefix",
			tokenPrefix: "github_pat_",
			wantErr:     true,
			errContains: "unauthenticated",
		},
		{
			name:        "valid gho_ token prefix",
			tokenPrefix: "gho_",
			wantErr:     true,
			errContains: "unauthenticated",
		},
		{
			name:        "invalid token prefix",
			tokenPrefix: "invalid_",
			wantErr:     true,
			errContains: "invalid token",
		},
		{
			name:        "no token prefix",
			tokenPrefix: "",
			wantErr:     true,
			errContains: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fake token with the test prefix
			fakeToken := tt.tokenPrefix + "fakeTokenValueForTesting1234567890"

			// Create a client with both server-side and client-side token auth interceptors
			client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"),
				web.WithInterceptors(web.TokenAuthInterceptorFunc),
				web.WithClientOptions(connect.WithInterceptors(
					interceptor.NewTokenAuthClientInterceptor(fakeToken),
				)),
			)

			_, err := client.GetUser(context.Background(), connect.NewRequest(&qf.Void{}))

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				errMsg := strings.ToLower(err.Error())
				expectedMsg := strings.ToLower(tt.errContains)
				if !strings.Contains(errMsg, expectedMsg) {
					t.Errorf("GetUser() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}
