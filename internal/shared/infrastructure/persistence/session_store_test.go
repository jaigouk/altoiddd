package persistence_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/shared/infrastructure/persistence"
)

// -- Put / Get --

func TestSessionStore_PutAndGetReturnsSession(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(60 * time.Second)
	obj := "test-session"
	store.Put("s1", obj)
	got, err := store.Get("s1")
	require.NoError(t, err)
	assert.Equal(t, "test-session", got)
}

func TestSessionStore_GetMissingRaisesNotFound(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(60 * time.Second)
	_, err := store.Get("nonexistent-id")
	require.ErrorIs(t, err, domainerrors.ErrNotFound)
}

func TestSessionStore_PutOverwritesExisting(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(60 * time.Second)
	store.Put("key", "first")
	store.Put("key", "second")
	got, err := store.Get("key")
	require.NoError(t, err)
	assert.Equal(t, "second", got)
}

func TestSessionStore_PutRefreshesTTL(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(100 * time.Millisecond)
	store.Put("s1", "val")
	store.Put("s1", "val") // refresh
	got, err := store.Get("s1")
	require.NoError(t, err)
	assert.Equal(t, "val", got)
}

// -- TTL / Expiration --

func TestSessionStore_GetExpiredRaisesNotFound(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(50 * time.Millisecond)
	store.Put("s1", "val")
	time.Sleep(100 * time.Millisecond)
	_, err := store.Get("s1")
	require.ErrorIs(t, err, domainerrors.ErrNotFound)
}

func TestSessionStore_ExpiredSessionLazilyRemoved(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(50 * time.Millisecond)
	store.Put("s1", "val")
	assert.Equal(t, 1, store.ActiveCount())
	time.Sleep(100 * time.Millisecond)
	_, err := store.Get("s1")
	require.ErrorIs(t, err, domainerrors.ErrNotFound)
	assert.Equal(t, 0, store.ActiveCount())
}

// -- Cleanup --

func TestSessionStore_CleanupRemovesExpired(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(50 * time.Millisecond)
	store.Put("s1", "v1")
	store.Put("s2", "v2")
	time.Sleep(100 * time.Millisecond)
	removed := store.CleanupExpired()
	assert.Equal(t, 2, removed)
	assert.Equal(t, 0, store.ActiveCount())
}

func TestSessionStore_CleanupOnEmptyReturnsZero(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(60 * time.Second)
	assert.Equal(t, 0, store.CleanupExpired())
}

func TestSessionStore_CleanupKeepsLiveSessions(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(60 * time.Second)
	store.Put("s1", "val")
	removed := store.CleanupExpired()
	assert.Equal(t, 0, removed)
	assert.Equal(t, 1, store.ActiveCount())
}

func TestSessionStore_CleanupMixedExpiredAndLive(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(50 * time.Millisecond)
	store.Put("old", "v1")
	time.Sleep(100 * time.Millisecond)
	store.Put("fresh", "v2")
	removed := store.CleanupExpired()
	assert.Equal(t, 1, removed)
	assert.Equal(t, 1, store.ActiveCount())
	got, err := store.Get("fresh")
	require.NoError(t, err)
	assert.Equal(t, "v2", got)
}

// -- Active count --

func TestSessionStore_ActiveCountEmpty(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(60 * time.Second)
	assert.Equal(t, 0, store.ActiveCount())
}

func TestSessionStore_ActiveCountAfterPuts(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(60 * time.Second)
	for i := range 3 {
		store.Put(string(rune('a'+i)), "val")
	}
	assert.Equal(t, 3, store.ActiveCount())
}

func TestSessionStore_ActiveCountIncludesExpiredBeforeCleanup(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStore(50 * time.Millisecond)
	store.Put("s1", "val")
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, store.ActiveCount())
	store.CleanupExpired()
	assert.Equal(t, 0, store.ActiveCount())
}

// -- Default TTL --

func TestSessionStore_DefaultTTLIs30Minutes(t *testing.T) {
	t.Parallel()
	store := persistence.NewSessionStoreDefault()
	assert.Equal(t, 30*time.Minute, store.TTL())
}
