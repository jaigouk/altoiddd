package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// Compile-time interface satisfaction check.
var _ application.SessionRepository = (*FileSystemSessionRepository)(nil)

const sessionFileName = "discovery_session.json"

// FileSystemSessionRepository persists discovery sessions as JSON files.
type FileSystemSessionRepository struct {
	baseDir string
}

// NewFileSystemSessionRepository creates a repository that stores sessions under baseDir.
func NewFileSystemSessionRepository(baseDir string) *FileSystemSessionRepository {
	return &FileSystemSessionRepository{baseDir: baseDir}
}

// Save persists a discovery session to a JSON file.
func (r *FileSystemSessionRepository) Save(_ context.Context, session *domain.DiscoverySession) error {
	if err := os.MkdirAll(r.baseDir, 0o755); err != nil {
		return fmt.Errorf("creating session directory %s: %w", r.baseDir, err)
	}

	snapshot := session.ToSnapshot()

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling session: %w", err)
	}

	filePath := filepath.Join(r.baseDir, sessionFileName)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("writing session file %s: %w", filePath, err)
	}

	return nil
}

// Load retrieves a discovery session from a JSON file.
func (r *FileSystemSessionRepository) Load(_ context.Context, _ string) (*domain.DiscoverySession, error) {
	filePath := filepath.Join(r.baseDir, sessionFileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session file not found: %w", domainerrors.ErrNotFound)
		}
		return nil, fmt.Errorf("reading session file %s: %w", filePath, err)
	}

	var snapshot map[string]interface{}
	if unmarshalErr := json.Unmarshal(data, &snapshot); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshalling session: %w", unmarshalErr)
	}

	session, err := domain.FromSnapshot(snapshot)
	if err != nil {
		return nil, fmt.Errorf("restoring session from snapshot: %w", err)
	}

	return session, nil
}

// Exists checks whether a persisted session file exists.
func (r *FileSystemSessionRepository) Exists(_ context.Context, _ string) (bool, error) {
	filePath := filepath.Join(r.baseDir, sessionFileName)

	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("checking session file %s: %w", filePath, err)
	}

	return true, nil
}
