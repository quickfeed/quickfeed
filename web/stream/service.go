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
	Submission *Service[uint64, qf.Submission]
}

// NewStreamServices creates a new StreamServices.
func NewStreamServices() *StreamServices {
	return &StreamServices{
		Submission: NewService[uint64, qf.Submission](),
	}
}

// ID is the allowed type for stream IDs.
type ID interface {
	uint64 | string
}

// Service[K ID, V any] is a type specific stream service.
// It also contains a map of streams that are currently connected.
type Service[K ID, V any] struct {
	mu sync.Mutex
	// The map of streams.
	streams map[K]StreamInterface[V]
}

// NewService creates a new service.
func NewService[K ID, V any]() *Service[K, V] {
	return &Service[K, V]{
		streams: make(map[K]StreamInterface[V]),
	}
}

// SendTo sends data to connected clients with the given IDs.
// If no ID is given, data is sent to all connected clients.
// Unconnected clients are ignored and will not receive the data.
func (s *Service[K, V]) SendTo(data *V, ids ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		stream, ok := s.streams[id]
		if !ok {
			continue
		}
		stream.Send(data)
	}
}

// Broadcast sends data to all connected clients.
func (s *Service[K, V]) Broadcast(data *V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stream := range s.streams {
		stream.Send(data)
	}
}

// Add adds a new stream for the given identifier.
// The identifier may be a user ID or an external application ID.
func (s *Service[K, V]) Add(stream StreamInterface[V], id K) {
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
func (s *Service[K, V]) internalRemove(id K) {
	if stream, ok := s.streams[id]; ok {
		stream.Close()
		delete(s.streams, id)
	}
}

// Close closes all streams in the service.
func (s *Service[K, V]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.streams {
		s.internalRemove(id)
	}
}

// CloseBy closes a stream for the given ID, if any exists.
func (s *Service[K, V]) CloseBy(id K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.internalRemove(id)
}
