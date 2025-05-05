package web

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/scm"
)

func TestUserSCMError(t *testing.T) {
	u1 := scm.M("u1: %w", scm.ErrNotMember)
	u2 := scm.M("u2: %w", scm.ErrNotOwner)
	u3 := scm.M("u3: %w", scm.ErrAlreadyExists)
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "nil", err: nil, want: nil},
		{name: "OtherError", err: errors.New("other"), want: nil},
		{name: "ErrNotMember", err: scm.ErrNotMember, want: nil},
		{name: "ErrNotOwner", err: scm.ErrNotOwner, want: nil},
		{name: "ErrAlreadyExists", err: scm.ErrAlreadyExists, want: nil},
		{name: "SCMError{UserError{ErrNotMember}}", err: scm.E(scm.Op("a"), u1), want: scm.ErrNotMember},
		{name: "SCMError{UserError{ErrNotOwner}}", err: scm.E(scm.Op("a"), u2), want: scm.ErrNotOwner},
		{name: "SCMError{UserError{ErrAlreadyExists}}", err: scm.E(scm.Op("a"), u3), want: scm.ErrAlreadyExists},

		{name: "SCMError{UserError{ErrNotMember}}Code", err: scm.E(scm.Op("a"), u1), want: connect.NewError(connect.CodeNotFound, u1)},
		{name: "SCMError{UserError{ErrNotOwner}}Code", err: scm.E(scm.Op("a"), u2), want: connect.NewError(connect.CodePermissionDenied, u2)},
		{name: "SCMError{UserError{ErrAlreadyExists}}Code", err: scm.E(scm.Op("a"), u3), want: connect.NewError(connect.CodeAlreadyExists, u3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := userSCMError(tt.err)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("userScmError() = %v, want %v", got, tt.want)
			}
			if got == nil {
				return
			}

			var gotConnErr *connect.Error
			if !errors.As(got, &gotConnErr) {
				t.Fatalf("Expected a connect.Error, got %T", got)
			}
			var wantConnErr *connect.Error
			if errors.As(tt.want, &wantConnErr) {
				gotCode := connect.CodeOf(got)
				wantCode := connect.CodeOf(wantConnErr)
				if gotCode != wantCode {
					t.Errorf("userScmError() code = %v, want %v", gotCode, wantCode)
				}
				// The target wantConnErr is a different connect.Error instance from the gotConnErr.
				// Therefore, we need to unwrap the target wantConnErr to get the underlying error for comparison.
				if !errors.Is(got, wantConnErr.Unwrap()) {
					t.Errorf("userScmError() err = %v, want %v", got, wantConnErr.Unwrap())
				}
			} else {
				if !errors.Is(got, tt.want) {
					t.Errorf("userScmError() err = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
