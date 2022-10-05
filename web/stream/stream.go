package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/bufbuild/connect-go"
)

type StreamInterface[T any] interface {
	Close()
	Run() error
	Send(data *T)
}

// stream wraps a connect.ServerStream.
type stream[T any] struct {
	mu sync.Mutex
	// stream is the underlying connect stream
	// that does the actual transfer of data
	// between the server and a client
	stream *connect.ServerStream[T]
	// context is the context of the stream
	ctx context.Context
	// The channel that we listen to for any
	// new data that we need to send to the client.
	ch chan *T
	// The client ID. This should be the same as
	// the user ID of the user that is connected,
	// retrieved from claims.
	id uint64
	// closed is a flag that indicates whether
	// the stream has been closed.
	closed bool
}

// newStream creates a new stream.
func newStream[T any](ctx context.Context, st *connect.ServerStream[T], id uint64) *stream[T] {
	return &stream[T]{
		stream: st,
		ctx:    ctx,
		ch:     make(chan *T),
		id:     id,
	}
}

// Close closes the stream.
func (s *stream[T]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		close(s.ch)
	}
	s.closed = true
}

// Run runs the stream.
// Run will block until the stream is closed.
func (s *stream[T]) Run() error {
	defer s.Close()
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case data, ok := <-s.ch:
			if !ok {
				return fmt.Errorf("stream closed")
			}
			if err := s.stream.Send(data); err != nil {
				return err
			}
		}
	}
}

// Send sends data to this stream's connected client.
func (s *stream[T]) Send(data *T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.ch <- data
	}
}
