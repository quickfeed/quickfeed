package stream

import (
	"errors"
	"sync"
)

// Pool is a pool of channels.
// The channels are used to send data to connected clients.
type Pool[T any] struct {
	mu sync.Mutex
	// Map of channels that are used to send data to the client.
	// The key is the user ID.
	channels map[uint64]chan *T
}

// NewPool creates a new pool.
func NewPool[T any]() *Pool[T] {
	return &Pool[T]{
		channels: make(map[uint64]chan *T),
	}
}

// Add adds a new channel to the pool.
func (p *Pool[T]) Add(id uint64) chan *T {
	// Delete the channel if it already exists.
	p.Remove(id)
	p.mu.Lock()
	defer p.mu.Unlock()
	// Create a new channel.
	p.channels[id] = make(chan *T)
	return p.channels[id]
}

// Remove removes a channel from the pool.
func (p *Pool[T]) Remove(id uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if ch, ok := p.channels[id]; ok {
		close(ch)
	}
	delete(p.channels, id)
}

// Send sends data to all channels in the pool.
func (p *Pool[T]) Send(data *T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ch := range p.channels {
		ch <- data
	}
}

// SendTo attempts to send data to a specific channel in the pool.
func (p *Pool[T]) SendTo(id uint64, data *T) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	ch, ok := p.channels[id]
	if !ok {
		return errors.New("channel not found")
	}
	ch <- data
	return nil
}

// Close closes all channels in the pool.
func (p *Pool[T]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ch := range p.channels {
		close(ch)
	}
}

// CloseBy closes a specific channel in the pool.
func (p *Pool[T]) CloseBy(id uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	ch, ok := p.channels[id]
	if !ok {
		return errors.New("channel not found")
	}
	close(ch)
	delete(p.channels, id)
	return nil
}
