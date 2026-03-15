package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
	"github.com/alty-cli/alty/internal/discovery/infrastructure"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// Compile-time interface check
// ---------------------------------------------------------------------------

func TestFileSystemSessionRepositoryImplementsPort(t *testing.T) {
	t.Parallel()
	var _ discoveryapp.SessionRepository = (*infrastructure.FileSystemSessionRepository)(nil)
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestFileSystemSessionRepository_Save_WhenDirNotExists_ExpectCreatedAndWritten(t *testing.T) {
	t.Parallel()
	baseDir := filepath.Join(t.TempDir(), "nested", ".alty")
	repo := infrastructure.NewFileSystemSessionRepository(baseDir)

	session := domain.NewDiscoverySession("Test README content")

	err := repo.Save(context.Background(), session)
	require.NoError(t, err)

	// Verify the file was created
	filePath := filepath.Join(baseDir, "discovery_session.json")
	_, err = os.Stat(filePath)
	require.NoError(t, err, "session file should exist after save")
}

// ---------------------------------------------------------------------------
// Load
// ---------------------------------------------------------------------------

func TestFileSystemSessionRepository_Load_WhenFileExists_ExpectSessionRestored(t *testing.T) {
	t.Parallel()
	baseDir := filepath.Join(t.TempDir(), ".alty")
	repo := infrastructure.NewFileSystemSessionRepository(baseDir)

	// Save first
	session := domain.NewDiscoverySession("Test README content")
	require.NoError(t, repo.Save(context.Background(), session))

	// Load it back
	loaded, err := repo.Load(context.Background(), session.SessionID())
	require.NoError(t, err)
	assert.Equal(t, session.SessionID(), loaded.SessionID())
	assert.Equal(t, "Test README content", loaded.ReadmeContent())
	assert.Equal(t, domain.StatusCreated, loaded.Status())
}

func TestFileSystemSessionRepository_Load_WhenFileNotExists_ExpectNotFoundError(t *testing.T) {
	t.Parallel()
	baseDir := filepath.Join(t.TempDir(), ".alty")
	repo := infrastructure.NewFileSystemSessionRepository(baseDir)

	_, err := repo.Load(context.Background(), "nonexistent-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrNotFound)
}

func TestFileSystemSessionRepository_Load_WhenCorruptedJSON_ExpectWrappedError(t *testing.T) {
	t.Parallel()
	baseDir := filepath.Join(t.TempDir(), ".alty")
	require.NoError(t, os.MkdirAll(baseDir, 0o755))

	// Write corrupted JSON
	filePath := filepath.Join(baseDir, "discovery_session.json")
	require.NoError(t, os.WriteFile(filePath, []byte("{not valid json!!!"), 0o644))

	repo := infrastructure.NewFileSystemSessionRepository(baseDir)
	_, err := repo.Load(context.Background(), "any-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling session")
}

// ---------------------------------------------------------------------------
// Exists
// ---------------------------------------------------------------------------

func TestFileSystemSessionRepository_Exists_WhenFileExists_ExpectTrue(t *testing.T) {
	t.Parallel()
	baseDir := filepath.Join(t.TempDir(), ".alty")
	repo := infrastructure.NewFileSystemSessionRepository(baseDir)

	session := domain.NewDiscoverySession("Test README")
	require.NoError(t, repo.Save(context.Background(), session))

	exists, err := repo.Exists(context.Background(), session.SessionID())
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFileSystemSessionRepository_Exists_WhenFileNotExists_ExpectFalse(t *testing.T) {
	t.Parallel()
	baseDir := filepath.Join(t.TempDir(), ".alty")
	repo := infrastructure.NewFileSystemSessionRepository(baseDir)

	exists, err := repo.Exists(context.Background(), "nonexistent-id")
	require.NoError(t, err)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// Round-trip
// ---------------------------------------------------------------------------

func TestFileSystemSessionRepository_SaveLoad_RoundTrip_ExpectAllFieldsPreserved(t *testing.T) {
	t.Parallel()
	baseDir := filepath.Join(t.TempDir(), ".alty")
	repo := infrastructure.NewFileSystemSessionRepository(baseDir)

	// Build a session with persona, answers, skipped, and playback
	session := domain.NewDiscoverySession("Round-trip README")
	require.NoError(t, session.DetectPersona("1")) // developer, technical

	// Answer Q1-Q3 to trigger playback
	require.NoError(t, session.AnswerQuestion("Q1", "Users and admins"))
	require.NoError(t, session.AnswerQuestion("Q2", "The core entities"))
	require.NoError(t, session.AnswerQuestion("Q3", "Main use case"))

	// Confirm playback
	require.NoError(t, session.ConfirmPlayback(true, ""))

	// Skip Q4
	require.NoError(t, session.SkipQuestion("Q4", "Not relevant now"))

	// Save
	require.NoError(t, repo.Save(context.Background(), session))

	// Load
	loaded, err := repo.Load(context.Background(), session.SessionID())
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, session.SessionID(), loaded.SessionID())
	assert.Equal(t, "Round-trip README", loaded.ReadmeContent())
	assert.Equal(t, domain.StatusAnswering, loaded.Status())

	persona, ok := loaded.Persona()
	assert.True(t, ok)
	assert.Equal(t, domain.PersonaDeveloper, persona)

	register, ok := loaded.Register()
	assert.True(t, ok)
	assert.Equal(t, domain.RegisterTechnical, register)

	assert.Len(t, loaded.Answers(), 3)
	assert.Equal(t, "Q1", loaded.Answers()[0].QuestionID())
	assert.Equal(t, "Users and admins", loaded.Answers()[0].ResponseText())

	assert.Len(t, loaded.PlaybackConfirmations(), 1)
	assert.True(t, loaded.PlaybackConfirmations()[0].Confirmed())

	assert.Equal(t, "Not relevant now", loaded.SkipReason("Q4"))
}
