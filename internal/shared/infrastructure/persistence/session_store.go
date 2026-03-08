package persistence

import (
	"fmt"
	"sync"
	"time"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

type sessionEntry struct {
	value    any
	storedAt time.Time
}

// SessionStore is an in-memory session store with TTL-based expiration.
// Expired entries are lazily removed on Get() and can be bulk-pruned
// via CleanupExpired().
type SessionStore struct {
	mu    sync.RWMutex
	store map[string]sessionEntry
	ttl   time.Duration
}

// NewSessionStore creates a SessionStore with the given TTL.
func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{
		store: make(map[string]sessionEntry),
		ttl:   ttl,
	}
}

// NewSessionStoreDefault creates a SessionStore with a 30-minute TTL.
func NewSessionStoreDefault() *SessionStore {
	return NewSessionStore(30 * time.Minute)
}

// TTL returns the configured time-to-live.
func (s *SessionStore) TTL() time.Duration {
	return s.ttl
}

// Put stores or overwrites a session with a fresh TTL timestamp.
func (s *SessionStore) Put(sessionID string, session any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[sessionID] = sessionEntry{value: session, storedAt: time.Now()}
}

// Get retrieves a session by ID. Returns ErrNotFound if the session
// does not exist or has expired. Expired entries are lazily removed.
func (s *SessionStore) Get(sessionID string) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.store[sessionID]
	if !ok {
		return nil, fmt.Errorf("no active session with id '%s': %w",
			sessionID, domainerrors.ErrNotFound)
	}
	if time.Since(entry.storedAt) > s.ttl {
		delete(s.store, sessionID)
		return nil, fmt.Errorf("session '%s' has expired: %w",
			sessionID, domainerrors.ErrNotFound)
	}
	return entry.value, nil
}

// CleanupExpired removes all expired sessions from the store.
// Returns the number of sessions removed.
func (s *SessionStore) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var expired []string
	for key, entry := range s.store {
		if now.Sub(entry.storedAt) > s.ttl {
			expired = append(expired, key)
		}
	}
	for _, key := range expired {
		delete(s.store, key)
	}
	return len(expired)
}

// ActiveCount returns the number of sessions in the store (including expired).
// To get an accurate count of non-expired sessions, call CleanupExpired() first.
func (s *SessionStore) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.store)
}
