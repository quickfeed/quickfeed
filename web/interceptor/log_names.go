package interceptor

import (
	"sync"
	"time"
)

const (
	logNameTTL     = 5 * time.Minute
	maxLogNameKeys = 4096
)

type cachedLogName struct {
	name    string
	expires time.Time
}

// logNames stores only display names, never authorization data or user models.
// Entries expire so that renamed users and courses are eventually refreshed.
type logNames struct {
	mu     sync.Mutex
	values map[uint64]cachedLogName
	lookup func(uint64) (string, error)
}

func (c *logNames) get(id uint64) (string, error) {
	if id == 0 {
		return "", nil
	}
	c.mu.Lock()
	entry, ok := c.values[id]
	c.mu.Unlock()
	if ok && time.Now().Before(entry.expires) {
		return entry.name, nil
	}

	// Database work must not block unrelated cache hits. Concurrent misses
	// may perform the same lookup, but never share mutable database models.
	name, err := c.lookup(id)
	if err != nil || name == "" {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values == nil {
		c.values = make(map[uint64]cachedLogName)
	}
	if len(c.values) >= maxLogNameKeys {
		// Any entry can be reloaded; arbitrary eviction keeps this cache small
		// without maintaining recency information on every request.
		for key := range c.values {
			delete(c.values, key)
			break
		}
	}
	c.values[id] = cachedLogName{name: name, expires: time.Now().Add(logNameTTL)}
	return name, nil
}
