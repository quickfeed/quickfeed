package auth_test

import (
	"testing"

	"github.com/autograde/quickfeed/web/auth"
)

func TestAddGetCookie(t *testing.T) {
	tests := []struct {
		cookie string
		id     uint64
		wantID uint64
	}{
		{cookie: "cookie1", id: 1, wantID: 1},
		{cookie: "cookie2", id: 1, wantID: 1},
	}
	for _, test := range tests {
		auth.Add(test.cookie, test.id)
		got := auth.Get(test.cookie)
		if got != test.wantID {
			t.Errorf("Get(%s) = %d, want %d", test.cookie, got, test.wantID)
		}
	}
}
