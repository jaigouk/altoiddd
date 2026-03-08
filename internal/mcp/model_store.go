package mcp

import (
	"fmt"
	"sync"
	"time"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ModelStore caches DomainModels by SessionID so that generate_artifacts
// can build the model once and generate_fitness/tickets/configs can reuse it.
type ModelStore struct {
	ttl     time.Duration
	mu      sync.RWMutex
	entries map[string]*modelEntry
}

type modelEntry struct {
	model   *ddd.DomainModel
	profile vo.StackProfile
	created time.Time
}

// NewModelStore creates a ModelStore with the given TTL for entries.
func NewModelStore(ttl time.Duration) *ModelStore {
	return &ModelStore{
		ttl:     ttl,
		entries: make(map[string]*modelEntry),
	}
}

// Put stores a DomainModel and StackProfile under the given sessionID.
// Overwrites any existing entry for that sessionID.
func (s *ModelStore) Put(sessionID string, model *ddd.DomainModel, profile vo.StackProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[sessionID] = &modelEntry{
		model:   model,
		profile: profile,
		created: time.Now(),
	}
}

// Get retrieves a cached DomainModel and StackProfile by sessionID.
// Returns an error if not found or expired. Expired entries are lazily evicted.
func (s *ModelStore) Get(sessionID string) (*ddd.DomainModel, vo.StackProfile, error) {
	s.mu.RLock()
	entry, ok := s.entries[sessionID]
	if !ok {
		s.mu.RUnlock()
		return nil, nil, fmt.Errorf(
			"no domain model found for session %q — run generate_artifacts first", sessionID)
	}

	if time.Since(entry.created) > s.ttl {
		s.mu.RUnlock()
		// Lazy eviction: upgrade to write lock and delete.
		s.mu.Lock()
		delete(s.entries, sessionID)
		s.mu.Unlock()
		return nil, nil, fmt.Errorf(
			"session %q expired — run generate_artifacts again", sessionID)
	}

	model, profile := entry.model, entry.profile
	s.mu.RUnlock()
	return model, profile, nil
}
