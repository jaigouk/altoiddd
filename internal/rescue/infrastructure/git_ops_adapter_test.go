package infrastructure_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	rescueapp "github.com/alty-cli/alty/internal/rescue/application"
	"github.com/alty-cli/alty/internal/rescue/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitOpsAdapterImplementsPort(t *testing.T) {
	t.Parallel()
	var _ rescueapp.GitOps = (*infrastructure.GitOpsAdapter)(nil)
}

// ---------------------------------------------------------------------------
// Helper: create a temp git repo
// ---------------------------------------------------------------------------

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
	require.NoError(t, cmd.Run())

	// Configure git user for commits
	for _, kv := range [][]string{
		{"user.email", "test@test.com"},
		{"user.name", "Test"},
	} {
		c := exec.Command("git", "-C", dir, "config", kv[0], kv[1])
		require.NoError(t, c.Run())
	}

	// Create an initial commit so branches can be created
	emptyFile := filepath.Join(dir, ".gitkeep")
	require.NoError(t, os.WriteFile(emptyFile, []byte(""), 0o644))
	c := exec.Command("git", "-C", dir, "add", ".")
	require.NoError(t, c.Run())
	c = exec.Command("git", "-C", dir, "commit", "-m", "init")
	c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
	require.NoError(t, c.Run())

	return dir
}

// ---------------------------------------------------------------------------
// HasGit
// ---------------------------------------------------------------------------

func TestHasGitReturnsTrueForGitRepo(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}
	result, err := adapter.HasGit(context.Background(), dir)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestHasGitReturnsFalseForNonGitDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	adapter := &infrastructure.GitOpsAdapter{}
	result, err := adapter.HasGit(context.Background(), dir)
	require.NoError(t, err)
	assert.False(t, result)
}

// ---------------------------------------------------------------------------
// IsClean
// ---------------------------------------------------------------------------

func TestIsCleanReturnsTrueForCleanTree(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}
	result, err := adapter.IsClean(context.Background(), dir)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestIsCleanReturnsFalseForDirtyTree(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0o644))
	adapter := &infrastructure.GitOpsAdapter{}
	result, err := adapter.IsClean(context.Background(), dir)
	require.NoError(t, err)
	assert.False(t, result)
}

// ---------------------------------------------------------------------------
// BranchExists
// ---------------------------------------------------------------------------

func TestBranchExistsReturnsTrueForExistingBranch(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	// The default branch (main/master) exists
	adapter := &infrastructure.GitOpsAdapter{}

	// Create a branch to check
	cmd := exec.Command("git", "-C", dir, "branch", "test-branch")
	require.NoError(t, cmd.Run())

	result, err := adapter.BranchExists(context.Background(), dir, "test-branch")
	require.NoError(t, err)
	assert.True(t, result)
}

func TestBranchExistsReturnsFalseForNonExistentBranch(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}
	result, err := adapter.BranchExists(context.Background(), dir, "nonexistent")
	require.NoError(t, err)
	assert.False(t, result)
}

// ---------------------------------------------------------------------------
// CreateBranch
// ---------------------------------------------------------------------------

func TestCreateBranchCallsGit(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}
	err := adapter.CreateBranch(context.Background(), dir, "alty/init")
	require.NoError(t, err)

	// Verify branch was created and checked out
	result, err := adapter.BranchExists(context.Background(), dir, "alty/init")
	require.NoError(t, err)
	assert.True(t, result)
}

// ---------------------------------------------------------------------------
// Branch name validation
// ---------------------------------------------------------------------------

func TestValidBranchNamePasses(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	validNames := []string{"alty/init", "feature/my-branch", "fix_123"}
	for _, name := range validNames {
		_, err := adapter.BranchExists(context.Background(), dir, name)
		assert.NoError(t, err, "branch name %q should be valid", name)
	}
}

func TestInvalidBranchNameRaises(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	invalidNames := []string{"alty init", "branch;rm -rf /", ""}
	for _, name := range invalidNames {
		_, err := adapter.BranchExists(context.Background(), dir, name)
		assert.Error(t, err, "branch name %q should be invalid", name)
		assert.Contains(t, err.Error(), "invalid branch name")
	}
}

func TestCreateBranchValidatesName(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	err := adapter.CreateBranch(context.Background(), dir, "bad;name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid branch name")
}
