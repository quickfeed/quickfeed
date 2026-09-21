package interceptor

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"connectrpc.com/connect"
)

// fakeConn is a StreamingHandlerConn that returns a fixed sequence of Receive
// outcomes; only Receive and Spec are exercised.
type fakeConn struct {
	connect.StreamingHandlerConn
	receives []error
	calls    int
}

func (f *fakeConn) Receive(any) error {
	err := f.receives[f.calls]
	f.calls++
	return err
}

func (*fakeConn) Spec() connect.Spec {
	return connect.Spec{Procedure: "/qf.QuickFeedService/SomeStream"}
}

func (*fakeConn) RequestHeader() http.Header { return http.Header{} }

func TestCheckedConnChecksOnlyTheFirstMessage(t *testing.T) {
	checks := 0
	conn := &checkedConn{
		StreamingHandlerConn: &fakeConn{receives: []error{nil, nil, nil}},
		check: func(any) error {
			checks++
			return nil
		},
	}
	for range 3 {
		if err := conn.Receive(nil); err != nil {
			t.Fatalf("Receive() = %v, want nil", err)
		}
	}
	if checks != 1 {
		t.Errorf("check called %d times, want 1", checks)
	}
}

func TestCheckedConnRejectsOnFailedCheck(t *testing.T) {
	want := connect.NewError(connect.CodePermissionDenied, errors.New("access denied"))
	conn := &checkedConn{
		StreamingHandlerConn: &fakeConn{receives: []error{nil}},
		check:                func(any) error { return want },
	}
	if err := conn.Receive(nil); !errors.Is(err, want) {
		t.Errorf("Receive() = %v, want %v", err, want)
	}
}

// A stream the client closes before sending anything must surface the receive
// error rather than run the check against an unpopulated message.
func TestCheckedConnSkipsCheckWhenReceiveFails(t *testing.T) {
	checks := 0
	conn := &checkedConn{
		StreamingHandlerConn: &fakeConn{receives: []error{io.EOF, nil}},
		check: func(any) error {
			checks++
			return nil
		},
	}
	if err := conn.Receive(nil); !errors.Is(err, io.EOF) {
		t.Errorf("Receive() = %v, want %v", err, io.EOF)
	}
	if checks != 0 {
		t.Errorf("check called %d times, want 0", checks)
	}
	// The next successful receive is still the first message to check.
	if err := conn.Receive(nil); err != nil {
		t.Fatalf("Receive() = %v, want nil", err)
	}
	if checks != 1 {
		t.Errorf("check called %d times, want 1", checks)
	}
}
