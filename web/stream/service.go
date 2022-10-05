package stream

import (
	"sync"

	"github.com/quickfeed/quickfeed/qf"
)

// StreamServices contain all available stream services.
// Each service is unique to a specific type.
// The services may be used to send data to connected clients.
// To add a new service, add a new field to this struct and
// initialize the service in the NewStreamServices function.
type StreamServices struct {
	Submission *Service[qf.Submission]
}

// NewStreamServices creates a new StreamServices.
func NewStreamServices() *StreamServices {
	return &StreamServices{
		Submission: NewService[qf.Submission](),
	}
}

// Service[T] is a type specific stream service.
// It also contains a map of streams that are currently connected.
type Service[T any] struct {
	mu sync.Mutex
	// The map of streams.
	streams map[uint64]StreamInterface[T]
}

// NewService creates a new service.
func NewService[T any]() *Service[T] {
	return &Service[T]{
		streams: make(map[uint64]StreamInterface[T]),
	}
}

// SendTo sends data to connected clients with the given IDs.
// If no ID is given, data is sent to all connected clients.
// Unconnected clients are ignored and will not receive the data.
func (s *Service[T]) SendTo(data *T, userIDs ...uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
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

// Add adds a new stream for the given user to the service.
func (s *Service[T]) Add(stream StreamInterface[T], userID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Delete the stream if it already exists.
	s.internalRemove(userID)
	// Add the stream to the map.
	s.streams[userID] = stream
}

// internalRemove removes a stream from the service.
// This closes the stream and removes it from the map.
// This function must only be called when holding the mutex.
func (s *Service[T]) internalRemove(userID uint64) {
	if stream, ok := s.streams[userID]; ok {
		stream.Close()
		delete(s.streams, userID)
	}
}

// Close closes all streams in the service.
func (s *Service[T]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stream := range s.streams {
		stream.Close()
	}
}

// CloseBy closes the given user's stream.
func (s *Service[T]) CloseBy(userID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stream, ok := s.streams[userID]; ok {
		stream.Close()
	}
}
