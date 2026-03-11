package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
	"github.com/alty-cli/alty/internal/discovery/infrastructure"
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
	assert.Empty(t, result)
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
// ScanConflicts — global only (no local project config)
// ---------------------------------------------------------------------------

func TestScanConflicts_NoConflictsWhenNoTools(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	projectDir := t.TempDir()
	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), projectDir)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestScanConflicts_GlobalOnly(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	projectDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".claude"), 0o755))

	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), projectDir)
	require.NoError(t, err)

	var found bool
	for _, sc := range result {
		if sc.Tool() == "claude-code" && sc.Type() == "global_only" {
			found = true
			assert.Equal(t, domain.SettingsSeverityInfo, sc.Severity())
			assert.Contains(t, sc.Message(), "global config exists")
		}
	}
	assert.True(t, found, "expected global_only conflict for claude-code")
}

// ---------------------------------------------------------------------------
// ScanConflicts — both exist, same content (no conflict)
// ---------------------------------------------------------------------------

func TestScanConflicts_BothExistSameContent(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	projectDir := t.TempDir()

	globalDir := filepath.Join(homeDir, ".claude")
	localDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.Mkdir(globalDir, 0o755))
	require.NoError(t, os.Mkdir(localDir, 0o755))

	content := []byte(`{"model": "opus"}`)
	require.NoError(t, os.WriteFile(filepath.Join(globalDir, "settings.json"), content, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "settings.json"), content, 0o644))

	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), projectDir)
	require.NoError(t, err)

	// Both exist with same content — no content_mismatch conflict.
	for _, sc := range result {
		if sc.Tool() == "claude-code" {
			assert.NotEqual(t, "content_mismatch", sc.Type(),
				"same content should not produce a mismatch conflict")
		}
	}
}

// ---------------------------------------------------------------------------
// ScanConflicts — both exist, different content
// ---------------------------------------------------------------------------

func TestScanConflicts_BothExistDifferentContent(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	projectDir := t.TempDir()

	globalDir := filepath.Join(homeDir, ".claude")
	localDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.Mkdir(globalDir, 0o755))
	require.NoError(t, os.Mkdir(localDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(globalDir, "settings.json"), []byte(`{"model": "opus"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "settings.json"), []byte(`{"model": "sonnet"}`), 0o644))

	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), projectDir)
	require.NoError(t, err)

	var found bool
	for _, sc := range result {
		if sc.Tool() == "claude-code" && sc.Type() == "content_mismatch" {
			found = true
			assert.Equal(t, domain.SettingsSeverityWarning, sc.Severity())
			assert.Contains(t, sc.Message(), "settings.json")
			assert.NotEmpty(t, sc.Global())
			assert.NotEmpty(t, sc.Local())
		}
	}
	assert.True(t, found, "expected content_mismatch conflict for claude-code")
}

// ---------------------------------------------------------------------------
// ScanConflicts — no global config
// ---------------------------------------------------------------------------

func TestScanConflicts_NoGlobal(t *testing.T) {
	t.Parallel()
	homeDir := t.TempDir()
	projectDir := t.TempDir()

	// Only local exists, no global.
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".claude"), 0o755))

	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), projectDir)
	require.NoError(t, err)

	for _, sc := range result {
		if sc.Tool() == "claude-code" {
			t.Errorf("unexpected conflict for claude-code when no global config: %s", sc.Type())
		}
	}
}

// ---------------------------------------------------------------------------
// ScanConflicts — permission denied
// ---------------------------------------------------------------------------

func TestScanConflicts_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not applicable on Windows")
	}
	t.Parallel()
	homeDir := t.TempDir()
	projectDir := t.TempDir()

	unreadable := filepath.Join(homeDir, ".claude")
	require.NoError(t, os.Mkdir(unreadable, 0o000))
	defer os.Chmod(unreadable, 0o755) //nolint:errcheck

	scanner := infrastructure.NewFilesystemToolScanner(homeDir)
	result, err := scanner.ScanConflicts(context.Background(), projectDir)
	require.NoError(t, err)

	var found bool
	for _, sc := range result {
		if sc.Tool() == "claude-code" {
			found = true
			assert.Equal(t, domain.SettingsSeverityWarning, sc.Severity())
			assert.Contains(t, sc.Message(), "permission denied")
		}
	}
	assert.True(t, found, "expected permission denied conflict for claude-code")
}

// ---------------------------------------------------------------------------
// ScanConflicts — nonexistent home dir
// ---------------------------------------------------------------------------

func TestScanConflicts_NonexistentHomeDir(t *testing.T) {
	t.Parallel()
	nonexistent := filepath.Join(t.TempDir(), "does_not_exist")
	projectDir := t.TempDir()
	scanner := infrastructure.NewFilesystemToolScanner(nonexistent)
	result, err := scanner.ScanConflicts(context.Background(), projectDir)
	require.NoError(t, err)
	assert.Empty(t, result)
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
	assert.Empty(t, result)
}

func TestDefaultHomeDirUsesRealHome(t *testing.T) {
	t.Parallel()
	scanner := infrastructure.NewFilesystemToolScanner("")
	assert.NotNil(t, scanner)
}
