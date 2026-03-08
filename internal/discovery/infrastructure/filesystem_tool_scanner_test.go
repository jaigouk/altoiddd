package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemToolScannerImplementsPort(t *testing.T) {
	t.Parallel()
	var _ discoveryapp.ToolDetection = (*infrastructure.FilesystemToolScanner)(nil)
}

// ---------------------------------------------------------------------------
// Detect
// ---------------------------------------------------------------------------

func TestDetectNoToolsInstalled(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestDetectClaudeCode(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".claude"), 0o755))
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Contains(t, result, "claude-code")
}

func TestDetectCursor(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".cursor"), 0o755))
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Contains(t, result, "cursor")
}

func TestDetectRooCode(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".roo"), 0o755))
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Contains(t, result, "roo-code")
}

func TestDetectOpencode(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".config", "opencode"), 0o755))
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Contains(t, result, "opencode")
}

func TestDetectMultipleTools(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".claude"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".roo"), 0o755))
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Contains(t, result, "claude-code")
	assert.Contains(t, result, "roo-code")
	assert.Len(t, result, 2)
}

func TestDetectToolDirExistsButEmpty(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".claude"), 0o755))
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Contains(t, result, "claude-code")
}

// ---------------------------------------------------------------------------
// Conflicts
// ---------------------------------------------------------------------------

func TestNoConflictsWhenNoTools(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), homeDir)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestCursorSQLiteProducesWarning(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".cursor"), 0o755))
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), homeDir)
	require.NoError(t, err)
	found := false
	for _, c := range result {
		if assert.ObjectsAreEqual("cursor: SQLite-based config detected, cannot read", c) {
			found = true
		}
	}
	assert.True(t, found, "expected cursor SQLite warning")
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestPermissionDeniedHandledGracefully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not applicable on Windows")
	}
	t.Parallel()
	homeDir := t.TempDir()
	unreadable := filepath.Join(homeDir, ".claude")
	require.NoError(t, os.Mkdir(unreadable, 0o000))
	defer os.Chmod(unreadable, 0o755) //nolint:errcheck

	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.Detect(context.Background(), homeDir)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestHomeDirNotSet(t *testing.T) {
	t.Parallel()
	nonexistent := filepath.Join(t.TempDir(), "does_not_exist")
	scanner := infrastructure.NewFilesystemToolScanner(nonexistent)
	result, err := scanner.Detect(context.Background(), nonexistent)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestDefaultHomeDirUsesRealHome(t *testing.T) {
	t.Parallel()
	scanner := infrastructure.NewFilesystemToolScanner("")
	assert.NotNil(t, scanner)
}
