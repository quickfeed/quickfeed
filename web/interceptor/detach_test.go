package interceptor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

// cancellationProbe stubs out QuickFeedService methods to observe whether the
// handler's context is canceled when the client disconnects mid-request.
type cancellationProbe struct {
	qfconnect.UnimplementedQuickFeedServiceHandler
	started  chan struct{} // closed when the handler starts executing
	canceled chan bool     // reports whether the handler's context was canceled
}

func newCancellationProbe() *cancellationProbe {
	return &cancellationProbe{
		started:  make(chan struct{}),
		canceled: make(chan bool, 1),
	}
}

// observe reports true if ctx is canceled within the grace period after the
// client disconnects, and false if the handler outlives the disconnect.
func (c *cancellationProbe) observe(ctx context.Context) {
	close(c.started)
	select {
	case <-ctx.Done():
		c.canceled <- true
	case <-time.After(2 * time.Second):
		c.canceled <- false
	}
}

// UpdateGroup is in detachedMethods; the interceptor must shield it from the client's cancellation.
func (c *cancellationProbe) UpdateGroup(ctx context.Context, _ *qf.Group) (*qf.Group, error) {
	c.observe(ctx)
	return &qf.Group{}, nil
}

// GetUser is not in detachedMethods; it must be canceled when the client disconnects.
func (c *cancellationProbe) GetUser(ctx context.Context, _ *qf.Void) (*qf.User, error) {
	c.observe(ctx)
	return &qf.User{}, nil
}

func TestDetachInterceptor(t *testing.T) {
	tests := []struct {
		name         string
		call         func(ctx context.Context, client qfconnect.QuickFeedServiceClient) error
		wantCanceled bool
	}{
		{
			name: "detached method survives client disconnect",
			call: func(ctx context.Context, client qfconnect.QuickFeedServiceClient) error {
				_, err := client.UpdateGroup(ctx, &qf.Group{})
				return err
			},
			wantCanceled: false,
		},
		{
			name: "non-detached method is canceled on client disconnect",
			call: func(ctx context.Context, client qfconnect.QuickFeedServiceClient) error {
				_, err := client.GetUser(ctx, &qf.Void{})
				return err
			},
			wantCanceled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := newCancellationProbe()
			router := http.NewServeMux()
			router.Handle(qfconnect.NewQuickFeedServiceHandler(probe, connect.WithInterceptors(interceptor.NewDetachInterceptor())))
			server := httptest.NewUnstartedServer(router)
			server.EnableHTTP2 = true
			server.StartTLS()
			t.Cleanup(server.Close)
			client := qfconnect.NewQuickFeedServiceClient(server.Client(), server.URL)

			ctx, cancel := context.WithCancel(t.Context())
			done := make(chan error, 1)
			go func() {
				done <- tt.call(ctx, client)
			}()
			<-probe.started
			cancel() // simulate the client closing the connection mid-request
			if canceled := <-probe.canceled; canceled != tt.wantCanceled {
				t.Errorf("handler context canceled = %t, want %t", canceled, tt.wantCanceled)
			}
			<-done
		})
	}
}
