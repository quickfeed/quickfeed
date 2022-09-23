package stream

import (
	"errors"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/internal/multierr"
	"github.com/quickfeed/quickfeed/qf"
)

// StreamServices contain all available stream services.
// Each service is unique to a specific type.
// The services may be used to send data to connected clients.
// To add a new service, add a new field to this struct.
type StreamServices struct {
	Submission *Service[qf.Submission]
}

// NewStreamServices creates a new StreamServices.
func NewStreamServices() *StreamServices {
	return &StreamServices{
		Submission: newService[qf.Submission](),
	}
}

// Service[T] is a type specific stream service.
// It contains a pool that is used to send data to connected clients.
// It also contains a map of streams that are currently connected.
type Service[T any] struct {
	mu sync.RWMutex
	// The pool of channels that are used to send data to clients.
	pool *Pool[T]
	// The map of streams.
	streams map[uint64]*Stream[T]
}

// NewService creates a new service.
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

// Add adds a new stream to the service.
// It returns the stream which must be run by the caller.
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

// Remove removes a stream from the service.
// This closes the stream and removes it from the pool and map.
func (s *Service[T]) Remove(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.streams[id]; ok {
		s.streams[id].Close()
		delete(s.streams, id)
	}
}

// Close closes all streams in the service.
func (s *Service[T]) Close() {
	for _, stream := range s.streams {
		stream.Close()
	}
}

// CloseBy closes a single stream by ID.
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

// GetPool returns the pool of the service.
func (s *Service[T]) GetPool() *Pool[T] {
	return s.pool
}
