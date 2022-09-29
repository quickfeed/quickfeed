package stream

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/bufbuild/connect-go"
)

type Stream[T any] interface {
	Close()
	GetID() uint64
	Run() error
	Send(data *T)
}

// Stream wraps a connect.ServerStream.
type stream[T any] struct {
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
}

// NewStream creates a new stream.
func NewStream[T any](ctx context.Context, st *connect.ServerStream[T], id uint64) *stream[T] {
	return &stream[T]{
		stream: st,
		ctx:    ctx,
		ch:     make(chan *T),
		id:     id,
	}
}

// Close closes the stream.
func (s *stream[T]) Close() {
	close(s.ch)
}

// GetID returns the user ID of the stream.
func (s *stream[T]) GetID() uint64 {
	return s.id
}

// Run runs the stream.
// Run will block until the stream is closed.
func (s *stream[T]) Run() error {
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
	return nil
}

func (s *stream[T]) Send(data *T) {
	s.ch <- data
}

type mockStream[T any] struct {
	*stream[T]
	counter *uint32
}

func NewTestStream[T any](ctx context.Context, id uint64, counter *uint32) *mockStream[T] {
	return &mockStream[T]{
		stream:  &stream[T]{ctx: ctx, ch: make(chan *T), id: id},
		counter: counter,
	}
}

func (m *mockStream[T]) Run() error {
	for {
		select {
		case data, ok := <-m.ch:
			if !ok {
				fmt.Println("stream closed")
				return nil
			}
			atomic.AddUint32(m.counter, 1)
			fmt.Println(m.GetID(), data, &m.ch)
		case <-m.ctx.Done():
			fmt.Println("context closed")
			return m.ctx.Err()
		}
	}
}

func (m *mockStream[T]) GetChannel() chan *T {
	return m.ch
}
