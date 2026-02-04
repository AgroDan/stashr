package store

import (
	"sync"
	"time"
)

type entry struct {
	value     string
	expiresAt time.Time // zero value means no expiry
}

func (e *entry) expired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

// Store is a thread-safe in-memory key/value store with optional TTL support.
type Store struct {
	mu      sync.RWMutex
	data    map[string]*entry
	stopGC  chan struct{}
}

// New creates a new Store and starts a background goroutine that periodically
// sweeps expired keys. Call Stop to release resources.
func New() *Store {
	s := &Store{
		data:   make(map[string]*entry),
		stopGC: make(chan struct{}),
	}
	go s.gcLoop()
	return s
}

func (s *Store) gcLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.sweep()
		case <-s.stopGC:
			return
		}
	}
}

func (s *Store) sweep() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, e := range s.data {
		if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
			delete(s.data, k)
		}
	}
}

// Stop halts the background GC goroutine.
func (s *Store) Stop() {
	close(s.stopGC)
}

// Get retrieves a value by key. Returns the value and whether the key was found.
// Lazily deletes expired keys on access.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	e, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return "", false
	}
	if e.expired() {
		s.mu.RUnlock()
		// Upgrade to write lock to delete
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return "", false
	}
	val := e.value
	s.mu.RUnlock()
	return val, true
}

// Set stores a key/value pair. If ttl > 0 the key will expire after that duration.
func (s *Store) Set(key, value string, ttl time.Duration) {
	e := &entry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	s.mu.Lock()
	s.data[key] = e
	s.mu.Unlock()
}

// Delete removes a key. Returns true if the key existed (and was not expired).
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[key]
	if !ok || e.expired() {
		delete(s.data, key) // clean up if expired
		return false
	}
	delete(s.data, key)
	return true
}

// List returns all non-expired keys.
func (s *Store) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k, e := range s.data {
		if !e.expired() {
			keys = append(keys, k)
		}
	}
	return keys
}
