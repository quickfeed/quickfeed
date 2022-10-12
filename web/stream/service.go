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
	Submission *Service[qf.Submission, uint64]
}

// NewStreamServices creates a new StreamServices.
func NewStreamServices() *StreamServices {
	return &StreamServices{
		Submission: NewService[qf.Submission, uint64](),
	}
}

// ID is the allowed type for stream IDs.
type ID interface {
	uint64 | string
}

// Service[T] is a type specific stream service.
// It also contains a map of streams that are currently connected.
type Service[T any, V ID] struct {
	mu sync.Mutex
	// The map of streams.
	streams map[V]StreamInterface[T]
}

// NewService creates a new service.
func NewService[T any, V ID]() *Service[T, V] {
	return &Service[T, V]{
		streams: make(map[V]StreamInterface[T]),
	}
}

// SendTo sends data to connected clients with the given IDs.
// If no ID is given, data is sent to all connected clients.
// Unconnected clients are ignored and will not receive the data.
func (s *Service[T, V]) SendTo(data *T, ids ...V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(ids) == 0 {
		// Broadcast to all clients.
		for _, stream := range s.streams {
			stream.Send(data)
		}
		return
	}

	for _, id := range ids {
		stream, ok := s.streams[id]
		if !ok {
			continue
		}
		stream.Send(data)
	}
}

// Add adds a new stream for the given user to the service.
func (s *Service[T, V]) Add(stream StreamInterface[T], id V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Delete the stream if it already exists.
	s.internalRemove(id)
	// Add the stream to the map.
	s.streams[id] = stream
}

// internalRemove removes a stream from the service.
// This closes the stream and removes it from the map.
// This function must only be called when holding the mutex.
func (s *Service[T, V]) internalRemove(id V) {
	if stream, ok := s.streams[id]; ok {
		stream.Close()
		delete(s.streams, id)
	}
}

// Close closes all streams in the service.
func (s *Service[T, V]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stream := range s.streams {
		stream.Close()
	}
}

// CloseBy closes the given user's stream.
func (s *Service[T, V]) CloseBy(id V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stream, ok := s.streams[id]; ok {
		stream.Close()
	}
}
