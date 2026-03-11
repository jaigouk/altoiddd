package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/infrastructure"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
)

// --- Fake Prompter for Testing ---

type fakePrompter struct {
	personaChoice string
	personaErr    error
}

func (f *fakePrompter) SelectPersona(_ context.Context) (string, error) {
	return f.personaChoice, f.personaErr
}

// Compile-time check.
var _ application.Prompter = (*fakePrompter)(nil)

// --- Fake Event Publisher ---

type fakePublisher struct{}

func (f *fakePublisher) Publish(_ context.Context, _ any) error { return nil }

var _ sharedapp.EventPublisher = (*fakePublisher)(nil)

// --- Tests ---

func TestCLIDiscoveryAdapter_Run_HappyPath(t *testing.T) {
	t.Parallel()

	// Setup: create temp dir with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// Create handler and adapter
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	prompter := &fakePrompter{personaChoice: "1"}
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// Run
	err := adapter.Run(context.Background())
	require.NoError(t, err)
}

func TestCLIDiscoveryAdapter_Run_PersonaCanceled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	prompter := &fakePrompter{personaErr: context.Canceled}
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	err := adapter.Run(context.Background())
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCLIDiscoveryAdapter_Run_MissingREADME(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir() // No README.md

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	prompter := &fakePrompter{personaChoice: "1"}
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "README")
}

func TestCLIDiscoveryAdapter_Run_EmptyREADME(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte(""), 0o644))

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	prompter := &fakePrompter{personaChoice: "2"}
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// Empty README is allowed - handler accepts empty string
	err := adapter.Run(context.Background())
	require.NoError(t, err)
}

func TestCLIDiscoveryAdapter_Run_InvalidPersonaChoice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My idea"), 0o644))

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	prompter := &fakePrompter{personaChoice: "5"} // Invalid choice
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "persona")
}
