// Package infrastructure provides adapters for the Rescue bounded context.
package infrastructure

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	rescueapp "github.com/alty-cli/alty/internal/rescue/application"
)

var branchNamePattern = regexp.MustCompile(`^[a-zA-Z0-9/_-]+$`)

// GitOpsAdapter implements GitOps using the git command-line tool via subprocess.
type GitOpsAdapter struct{}

// Compile-time interface check.
var _ rescueapp.GitOps = (*GitOpsAdapter)(nil)

// HasGit checks whether the directory is inside a git repository.
func (g *GitOpsAdapter) HasGit(_ context.Context, projectDir string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = projectDir
	err := cmd.Run()
	return err == nil, nil
}

// IsClean checks whether the git working tree is clean.
func (g *GitOpsAdapter) IsClean(_ context.Context, projectDir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "", nil
}

// BranchExists checks whether a git branch already exists locally.
func (g *GitOpsAdapter) BranchExists(_ context.Context, projectDir, branchName string) (bool, error) {
	if err := validateBranchName(branchName); err != nil {
		return false, err
	}
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branchName)
	cmd.Dir = projectDir
	err := cmd.Run()
	return err == nil, nil
}

// CreateBranch creates and checks out a new git branch.
func (g *GitOpsAdapter) CreateBranch(_ context.Context, projectDir, branchName string) error {
	if err := validateBranchName(branchName); err != nil {
		return err
	}
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = projectDir
	return cmd.Run()
}

func validateBranchName(branchName string) error {
	if !branchNamePattern.MatchString(branchName) {
		return fmt.Errorf(
			"invalid branch name: %q. Only alphanumeric characters, '/', '-', and '_' are allowed",
			branchName)
	}
	return nil
}
