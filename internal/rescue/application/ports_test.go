package application_test

import (
	"context"

	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	"github.com/alty-cli/alty/internal/rescue/application"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// Compile-time interface satisfaction checks.
var (
	_ application.ProjectScan = (*mockProjectScan)(nil)
	_ application.GitOps      = (*mockGitOps)(nil)
	_ application.Rescue      = (*mockRescue)(nil)
)

// --- mockProjectScan ---

type mockProjectScan struct{}

func (m *mockProjectScan) Scan(_ context.Context, _ string, _ vo.StackProfile) (rescuedomain.ProjectScan, error) {
	return rescuedomain.ProjectScan{}, nil
}

// --- mockGitOps ---

type mockGitOps struct{}

func (m *mockGitOps) HasGit(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockGitOps) IsClean(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockGitOps) BranchExists(_ context.Context, _ string, _ string) (bool, error) {
	return false, nil
}

func (m *mockGitOps) CreateBranch(_ context.Context, _ string, _ string) error {
	return nil
}

// --- mockRescue ---

type mockRescue struct{}

func (m *mockRescue) Analyze(_ context.Context, _ string) (*rescuedomain.GapAnalysis, error) {
	return nil, nil
}

func (m *mockRescue) Plan(_ context.Context, _ *rescuedomain.GapAnalysis) (rescuedomain.MigrationPlan, error) {
	return rescuedomain.MigrationPlan{}, nil
}

func (m *mockRescue) Execute(_ context.Context, _ rescuedomain.MigrationPlan) error {
	return nil
}
