package stream

import (
	"context"
	"sync"

	"github.com/bufbuild/connect-go"
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
		Submission: NewService[qf.Submission](),
	}
}

// Service[T] is a type specific stream service..
// It also contains a map of streams that are currently connected.
type Service[T any] struct {
	mu sync.RWMutex
	// The map of streams.
	streams map[uint64]Stream[T]
}

// NewService creates a new service.
func NewService[T any]() *Service[T] {
	return &Service[T]{
		streams: make(map[uint64]Stream[T]),
	}
}

// SendTo sends data to client(s) with the given ID(s). If no ID is given, data is sent to all clients.
func (s *Service[T]) SendTo(data *T, userIDs ...uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(userIDs) == 0 {
		// Broadcast to all clients.
		for _, stream := range s.streams {
			stream.Send(data)
		}
		return
	}

	for _, userID := range userIDs {
		stream, ok := s.streams[userID]
		if !ok {
			continue
		}
		stream.Send(data)
	}
}

// Add adds a new stream to the service.
// It returns the stream which must be run by the caller.
func (s *Service[T]) Add(ctx context.Context, userID uint64, st *connect.ServerStream[T]) *stream[T] {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Delete the stream if it already exists.
	s.internalRemove(userID)
	// Add the stream to the map.
	stream := NewStream(ctx, st, userID)
	s.streams[stream.GetID()] = stream
	return stream
}

func (s *Service[T]) AddStream(userID uint64, st Stream[T]) Stream[T] {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.internalRemove(userID)
	// Add the stream to the map.
	s.streams[st.GetID()] = st
	return st
}

// Remove removes a stream from the service.
// This closes the stream and removes it from the map.
func (s *Service[T]) internalRemove(id uint64) {
	if stream, ok := s.streams[id]; ok {
		if !stream.Closed() {
		stream.Close()
		}
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if stream, ok := s.streams[id]; ok {
		if !stream.Closed() {
		stream.Close()
		}
	}
	return nil
}
