package stream

import (
	"errors"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/internal/multierr"
	"github.com/quickfeed/quickfeed/qf"
)

type Pool[T any] struct {
	mu sync.Mutex
	// Map of channels that are used to send data to the client.
	// The key is the user ID.
	channels map[uint64]chan *T
}

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

// Stream wraps a connect.ServerStream to provide a channel that can be used to send data to the client.
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
	//ctx, cancel := context.WithCancel(ctx)
	newStream := &Stream[T]{
		stream: stream,
		pool:   pool,
		ch:     pool.Add(id),
		id:     id,
	}
	return newStream
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
func (s *Stream[T]) Run() error {
	for data := range s.ch {
		if err := s.stream.Send(data); err != nil {
			return err
		}
	}
	return nil
}

// Service[T] is a gRPC streaming server.
type Service[T any] struct {
	mu sync.RWMutex
	// The pool of channels that are used to send data to the client.
	pool *Pool[T]
	// The map of streams.
	streams map[uint64]*Stream[T]
}

// NewServer creates a new server.
func newService[T any]() *Service[T] {
	return &Service[T]{
		pool:    NewPool[T](),
		streams: make(map[uint64]*Stream[T]),
	}
}

// SendTo sends data to client(s) with the given ID(s). If no ID is given, data is sent to all clients.
func (s *Service[T]) SendTo(data *T, userIDs ...uint64) error {
	if len(userIDs) == 0 {
		// Broadcast to all clients.
		s.pool.Send(data)
	}
	var errs []error
	for _, userID := range userIDs {
		err := s.pool.SendTo(userID, data)
		errs = append(errs, err)
	}
	return multierr.Join(errs...)
}

func (s *Service[T]) Add(userID uint64, st *connect.ServerStream[T]) *Stream[T] {
	// Delete the stream if it already exists.
	s.Remove(userID)
	// Add the stream to the map.
	stream := NewStream(st, s.GetPool(), userID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams[stream.GetID()] = stream
	return stream
}

func (s *Service[T]) Remove(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.streams[id]; ok {
		s.streams[id].Close()
		delete(s.streams, id)
	}
}

func (s *Service[T]) Close() {
	s.pool.Close()
	for _, stream := range s.streams {
		stream.Close()
	}
}

func (s *Service[T]) CloseBy(id uint64) error {
	if stream, ok := s.streams[id]; ok {
		stream.Close()
	}
	return nil
}

// GetStream returns a stream by ID.
func (s *Service[T]) GetStream(id uint64) (*Stream[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stream, ok := s.streams[id]
	if !ok {
		return nil, errors.New("stream not found")
	}
	return stream, nil
}

// GetPool returns the pool of channels that are used to send data to the client.
func (s *Service[T]) GetPool() *Pool[T] {
	return s.pool
}

// StreamServices contain all available stream services.
// Each service is unique to a specific type.
type StreamServices struct {
	Submission *Service[qf.Submission]
}

func NewStreamServices() *StreamServices {
	return &StreamServices{
		Submission: newService[qf.Submission](),
	}
}
