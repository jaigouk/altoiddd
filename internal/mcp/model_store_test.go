package mcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func makeTestModel(id string) *ddd.DomainModel {
	model := ddd.NewDomainModel(id)
	// Minimal valid model: 1 story with actor, 1 term in story text,
	// 1 bounded context with classification, 1 aggregate for core.
	story := vo.NewDomainStory("Primary Flow", []string{"User"}, "starts app", []string{"User starts app"}, nil)
	_ = model.AddDomainStory(story)
	_ = model.AddTerm("User", "A person using the system", "Auth", []string{"Q2"})

	bc := vo.NewDomainBoundedContext("Auth", "Manages authentication", nil, nil, "")
	_ = model.AddBoundedContext(bc)

	core := vo.SubdomainCore
	_ = model.ClassifySubdomain("Auth", core, "core domain")

	agg := vo.NewAggregateDesign("AuthRoot", "Auth", "AuthRoot", nil, []string{"valid credentials"}, nil, []string{"UserLoggedIn"})
	_ = model.DesignAggregate(agg)

	_ = model.Finalize()
	return model
}

func TestModelStore_PutAndGet(t *testing.T) {
	t.Parallel()
	store := NewModelStore(30 * time.Minute)

	model := makeTestModel("test-1")
	profile := vo.PythonUvProfile{}

	store.Put("session-abc", model, profile)

	got, gotProfile, err := store.Get("session-abc")
	require.NoError(t, err)
	assert.Equal(t, "test-1", got.ModelID())
	assert.IsType(t, vo.PythonUvProfile{}, gotProfile)
}

func TestModelStore_GetNotFound(t *testing.T) {
	t.Parallel()
	store := NewModelStore(30 * time.Minute)

	_, _, err := store.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no domain model found for session")
}

func TestModelStore_TTLExpiry(t *testing.T) {
	t.Parallel()
	store := NewModelStore(10 * time.Millisecond)

	model := makeTestModel("test-2")
	store.Put("session-ttl", model, vo.GenericProfile{})

	// Should work immediately.
	_, _, err := store.Get("session-ttl")
	require.NoError(t, err)

	// Wait for expiry.
	time.Sleep(15 * time.Millisecond)

	_, _, err = store.Get("session-ttl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestModelStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	store := NewModelStore(30 * time.Minute)

	// Write from multiple goroutines.
	done := make(chan struct{})
	for range 10 {
		go func() {
			defer func() { done <- struct{}{} }()
			model := ddd.NewDomainModel("concurrent")
			store.Put("session-concurrent", model, vo.GenericProfile{})
			_, _, _ = store.Get("session-concurrent")
		}()
	}

	for range 10 {
		<-done
	}

	// Should still be readable.
	_, _, err := store.Get("session-concurrent")
	require.NoError(t, err)
}

func TestModelStore_LazyEviction(t *testing.T) {
	t.Parallel()
	store := NewModelStore(10 * time.Millisecond)

	model := makeTestModel("evict-me")
	store.Put("session-evict", model, vo.GenericProfile{})

	time.Sleep(15 * time.Millisecond)

	// First Get should return error AND evict.
	_, _, err := store.Get("session-evict")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")

	// Second Get should return "not found" (evicted), not "expired".
	_, _, err = store.Get("session-evict")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no domain model found for session")
}

func TestModelStore_OverwriteExisting(t *testing.T) {
	t.Parallel()
	store := NewModelStore(30 * time.Minute)

	model1 := makeTestModel("first")
	model2 := makeTestModel("second")

	store.Put("session-x", model1, vo.GenericProfile{})
	store.Put("session-x", model2, vo.PythonUvProfile{})

	got, gotProfile, err := store.Get("session-x")
	require.NoError(t, err)
	assert.Equal(t, "second", got.ModelID())
	assert.IsType(t, vo.PythonUvProfile{}, gotProfile)
}
