// Package infrastructure provides adapters for the Bootstrap bounded context.
package infrastructure

import (
	"context"
	"fmt"
	"os/exec"

	bootstrapapp "github.com/alty-cli/alty/internal/bootstrap/application"
)

// GitCommitterAdapter implements GitCommitter using the git CLI via subprocess.
type GitCommitterAdapter struct{}

// Compile-time interface check.
var _ bootstrapapp.GitCommitter = (*GitCommitterAdapter)(nil)

// HasGit checks whether the directory is inside a git repository.
func (g *GitCommitterAdapter) HasGit(ctx context.Context, projectDir string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = projectDir
	err := cmd.Run()
	return err == nil, nil
}

// StageFiles stages specific file paths for commit.
func (g *GitCommitterAdapter) StageFiles(ctx context.Context, projectDir string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, paths...)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("staging files: %w", err)
	}
	return nil
}

// Commit creates a commit with the given message.
func (g *GitCommitterAdapter) Commit(ctx context.Context, projectDir string, message string) error {
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}
	return nil
}
