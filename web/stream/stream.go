package stream

import (
	"github.com/bufbuild/connect-go"
)

// Stream wraps a connect.ServerStream.
type Stream[T any] struct {
	// stream is the underlying connect stream
	// that does the actual transfer of data
	// between the server and a client
	stream *connect.ServerStream[T]
	// pool is stored to allow the stream to
	// removed itself from the pool when the
	// stream is closed
	pool *Pool[T]
	// The channel that we listen to for any
	// new data that we need to send to the client.
	ch chan *T
	// The client ID. This should be the same as
	// the user ID of the user that is connected,
	// retrieved from claims.
	id uint64
}

// NewStream creates a new stream.
func NewStream[T any](stream *connect.ServerStream[T], pool *Pool[T], id uint64) *Stream[T] {
	return &Stream[T]{
		stream: stream,
		pool:   pool,
		ch:     pool.Add(id),
		id:     id,
	}
}

// Close closes the stream.
func (s *Stream[T]) Close() {
	s.pool.Remove(s.id)
}

// GetID returns the user ID of the stream.
func (s *Stream[T]) GetID() uint64 {
	return s.id
}

// Run runs the stream.
// Run will block until the stream is closed.
func (s *Stream[T]) Run() error {
	for data := range s.ch {
		if err := s.stream.Send(data); err != nil {
			return err
		}
	}
	return nil
}
