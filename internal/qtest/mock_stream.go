package qtest

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

type MockStream[T any] struct {
	mu         sync.Mutex
	ctx        context.Context
	ch         chan *T
	closed     bool
	counter    *uint32
	Messages   []*T
	MessageMap map[string]int
}

func NewMockStream[T any](t *testing.T, ctx context.Context, counter *uint32) *MockStream[T] {
	t.Helper()
	return &MockStream[T]{
		ctx:        ctx,
		ch:         make(chan *T),
		closed:     false,
		counter:    counter,
		Messages:   make([]*T, 0),
		MessageMap: make(map[string]int),
	}
}

func (m *MockStream[T]) Run() error {
	for {
		select {
		case data, ok := <-m.ch:
			if !ok {
				return fmt.Errorf("stream closed")
			}
			atomic.AddUint32(m.counter, 1)
			m.Messages = append(m.Messages, data)
		case <-m.ctx.Done():
			return m.ctx.Err()
		}
	}
}

func (m *MockStream[T]) GetChannel() chan *T {
	return m.ch
}

func (m *MockStream[T]) Send(data *T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.ch <- data
	}
}

func (m *MockStream[T]) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		close(m.ch)
	}
	m.closed = true
}
