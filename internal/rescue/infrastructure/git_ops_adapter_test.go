package infrastructure_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rescueapp "github.com/alty-cli/alty/internal/rescue/application"
	"github.com/alty-cli/alty/internal/rescue/infrastructure"
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
	cmd := exec.CommandContext(context.Background(), "git", "init", dir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
	require.NoError(t, cmd.Run())

	// Configure git user for commits
	for _, kv := range [][]string{
		{"user.email", "test@test.com"},
		{"user.name", "Test"},
	} {
		c := exec.CommandContext(context.Background(), "git", "-C", dir, "config", kv[0], kv[1])
		require.NoError(t, c.Run())
	}

	// Create an initial commit so branches can be created
	emptyFile := filepath.Join(dir, ".gitkeep")
	require.NoError(t, os.WriteFile(emptyFile, []byte(""), 0o644))
	c := exec.CommandContext(context.Background(), "git", "-C", dir, "add", ".")
	require.NoError(t, c.Run())
	c = exec.CommandContext(context.Background(), "git", "-C", dir, "commit", "-m", "init")
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
	cmd := exec.CommandContext(context.Background(), "git", "-C", dir, "branch", "test-branch")
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

	validNames := []string{
		"alty/init",
		"feature/my-branch",
		"fix_123",
		"release/1.0",   // dots allowed
		"hotfix/v2.3.1", // multiple dots
		"feature/JIRA-123.fix",
	}
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
		require.Error(t, err, "branch name %q should be invalid", name)
		assert.Contains(t, err.Error(), "invalid branch name")
	}
}

func TestCreateBranchValidatesName(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	err := adapter.CreateBranch(context.Background(), dir, "bad;name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid branch name")
}

// ---------------------------------------------------------------------------
// CheckoutPrevious
// ---------------------------------------------------------------------------

func TestCheckoutPreviousReturnsToPreviousBranch(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	// Create and checkout a new branch
	require.NoError(t, adapter.CreateBranch(context.Background(), dir, "feature/test"))

	// Now checkout previous (should go back to main/master)
	err := adapter.CheckoutPrevious(context.Background(), dir)
	require.NoError(t, err)

	// Verify we're back on the original branch (not feature/test)
	cmd := exec.CommandContext(context.Background(), "git", "-C", dir, "branch", "--show-current")
	output, err := cmd.Output()
	require.NoError(t, err)
	currentBranch := strings.TrimSpace(string(output))
	assert.NotEqual(t, "feature/test", currentBranch)
}

func TestCheckoutPreviousReturnsErrorWhenNoPrevious(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	// First checkout in a fresh repo - no previous branch
	err := adapter.CheckoutPrevious(context.Background(), dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checkout previous")
}

// ---------------------------------------------------------------------------
// DeleteBranch
// ---------------------------------------------------------------------------

func TestDeleteBranchRemovesBranch(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	// Create a branch
	cmd := exec.CommandContext(context.Background(), "git", "-C", dir, "branch", "to-delete")
	require.NoError(t, cmd.Run())

	// Verify it exists
	exists, err := adapter.BranchExists(context.Background(), dir, "to-delete")
	require.NoError(t, err)
	require.True(t, exists)

	// Delete it
	err = adapter.DeleteBranch(context.Background(), dir, "to-delete")
	require.NoError(t, err)

	// Verify it's gone
	exists, err = adapter.BranchExists(context.Background(), dir, "to-delete")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestDeleteBranchValidatesName(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	err := adapter.DeleteBranch(context.Background(), dir, "bad;name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid branch name")
}

func TestDeleteBranchReturnsErrorForNonExistent(t *testing.T) {
	t.Parallel()
	dir := initGitRepo(t)
	adapter := &infrastructure.GitOpsAdapter{}

	err := adapter.DeleteBranch(context.Background(), dir, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleting branch")
}
