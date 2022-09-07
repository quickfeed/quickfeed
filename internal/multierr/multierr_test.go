package multierr_test

import (
	"errors"
	"testing"

	"github.com/quickfeed/quickfeed/internal/multierr"
)

func TestMultiErr(t *testing.T) {
	nilErr := multierr.Join(nil)
	if nilErr != nil {
		t.Errorf("Join(nil) = %v, want <nil>", nilErr)
	}
	nilErr = multierr.Join(nil, nil, nil)
	if nilErr != nil {
		t.Errorf("Join(nil) = %v, want <nil>", nilErr)
	}

	wantErr1 := errors.New("a")
	wantErr2 := errors.New("b")
	err1 := multierr.Join(wantErr1)
	if err1 == nil {
		t.Errorf("Join(a) = %v, want %v", err1, wantErr1)
	}
	err2 := multierr.Join(wantErr1, wantErr2)
	if err2 == nil {
		t.Errorf("Join(a,b) = %v, want %v, %v", err2, wantErr1, wantErr2)
	}
}
