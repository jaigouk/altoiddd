// Package infrastructure provides adapters for the Rescue bounded context.
package infrastructure

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	sharedapp "github.com/alto-cli/alto/internal/shared/application"
)

var branchNamePattern = regexp.MustCompile(`^[a-zA-Z0-9/_.\-]+$`)

// GitOpsAdapter implements GitOps using the git command-line tool via subprocess.
type GitOpsAdapter struct{}

// Compile-time interface check.
var _ sharedapp.GitOps = (*GitOpsAdapter)(nil)

// HasGit checks whether the directory is inside a git repository.
func (g *GitOpsAdapter) HasGit(ctx context.Context, projectDir string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = projectDir
	err := cmd.Run()
	return err == nil, nil
}

// IsClean checks whether the git working tree is clean.
func (g *GitOpsAdapter) IsClean(ctx context.Context, projectDir string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "", nil
}

// BranchExists checks whether a git branch already exists locally.
func (g *GitOpsAdapter) BranchExists(ctx context.Context, projectDir, branchName string) (bool, error) {
	if err := validateBranchName(branchName); err != nil {
		return false, err
	}
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "refs/heads/"+branchName)
	cmd.Dir = projectDir
	err := cmd.Run()
	return err == nil, nil
}

// CreateBranch creates and checks out a new git branch.
func (g *GitOpsAdapter) CreateBranch(ctx context.Context, projectDir, branchName string) error {
	if err := validateBranchName(branchName); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating branch %s: %w", branchName, err)
	}
	return nil
}

// CheckoutPrevious checks out the previous branch (git checkout -).
func (g *GitOpsAdapter) CheckoutPrevious(ctx context.Context, projectDir string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", "-")
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("checkout previous: %w", err)
	}
	return nil
}

// DeleteBranch deletes a local branch (git branch -D).
func (g *GitOpsAdapter) DeleteBranch(ctx context.Context, projectDir, branchName string) error {
	if err := validateBranchName(branchName); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "branch", "-D", branchName)
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deleting branch %s: %w", branchName, err)
	}
	return nil
}

func validateBranchName(branchName string) error {
	if !branchNamePattern.MatchString(branchName) {
		return fmt.Errorf(
			"invalid branch name: %q. Only alphanumeric characters, '/', '-', '_', and '.' are allowed",
			branchName)
	}
	return nil
}
