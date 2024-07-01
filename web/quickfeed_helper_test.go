package web_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
)

func Cookie(t *testing.T, tm *auth.TokenManager, user *qf.User) string {
	t.Helper()
	cookie, err := tm.NewAuthCookie(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	return cookie.String()
}
