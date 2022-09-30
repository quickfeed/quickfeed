package stream_test

import (
	"context"
	"fmt"
	"sync/atomic"
)

type mockStream[T any] struct {
	ctx        context.Context
	ch         chan *T
	id         uint64
	closed     bool
	counter    *uint32
	Messages   []T
	MessageMap map[string]int
}

func NewMockStream[T any](ctx context.Context, id uint64, counter *uint32) *mockStream[T] {
	return &mockStream[T]{
		ctx:        ctx,
		ch:         make(chan *T),
		id:         id,
		closed:     false,
		counter:    counter,
		Messages:   make([]T, 0),
		MessageMap: make(map[string]int),
	}
}

func (m *mockStream[T]) Run() error {
	for {
		select {
		case data, ok := <-m.ch:
			if !ok {
				return fmt.Errorf("stream closed")
			}
			atomic.AddUint32(m.counter, 1)
			m.Messages = append(m.Messages, *data)
			fmt.Println(m.id, data, &m.ch)
		case <-m.ctx.Done():
			return m.ctx.Err()
		}
	}
}

func (m *mockStream[T]) Closed() bool {
	return m.closed
}

func (m *mockStream[T]) GetChannel() chan *T {
	return m.ch
}

func (m *mockStream[T]) Send(data *T) {
	m.ch <- data
}

func (m *mockStream[T]) Close() {
	m.closed = true
	close(m.ch)
}

func (m *mockStream[T]) GetID() uint64 {
	return m.id
}
