package infrastructure_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/bootstrap/infrastructure"
)

// initGitRepo creates a git repo in the given directory with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "initial"},
	}
	for _, args := range cmds {
		cmd := exec.CommandContext(context.Background(), args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "cmd %v failed: %s", args, out)
	}
}

func TestGitCommitterAdapter_HasGit_WhenInRepo_ExpectTrue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initGitRepo(t, dir)

	adapter := &infrastructure.GitCommitterAdapter{}
	hasGit, err := adapter.HasGit(context.Background(), dir)
	require.NoError(t, err)
	assert.True(t, hasGit)
}

func TestGitCommitterAdapter_HasGit_WhenNotInRepo_ExpectFalse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	adapter := &infrastructure.GitCommitterAdapter{}
	hasGit, err := adapter.HasGit(context.Background(), dir)
	require.NoError(t, err)
	assert.False(t, hasGit)
}

func TestGitCommitterAdapter_StageFiles_ExpectFilesStaged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Create files to stage.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("prd"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), 0o644))

	adapter := &infrastructure.GitCommitterAdapter{}
	err := adapter.StageFiles(context.Background(), dir, []string{"docs/PRD.md", "README.md"})
	require.NoError(t, err)

	// Verify files are staged.
	cmd := exec.CommandContext(context.Background(), "git", "diff", "--cached", "--name-only")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "docs/PRD.md")
	assert.Contains(t, string(out), "README.md")
}

func TestGitCommitterAdapter_StageFiles_WhenEmpty_ExpectNoOp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	adapter := &infrastructure.GitCommitterAdapter{}
	err := adapter.StageFiles(context.Background(), dir, []string{})
	require.NoError(t, err)
}

func TestGitCommitterAdapter_Commit_ExpectCommitCreated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Create and stage a file.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0o644))
	stageCmd := exec.CommandContext(context.Background(), "git", "add", "test.txt")
	stageCmd.Dir = dir
	require.NoError(t, stageCmd.Run())

	adapter := &infrastructure.GitCommitterAdapter{}
	err := adapter.Commit(context.Background(), dir, "chore: test commit")
	require.NoError(t, err)

	// Verify commit was created.
	logCmd := exec.CommandContext(context.Background(), "git", "log", "--oneline", "-1")
	logCmd.Dir = dir
	out, err := logCmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "chore: test commit")
}
