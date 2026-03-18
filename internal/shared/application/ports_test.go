package application_test

import (
	"context"

	"github.com/alto-cli/alto/internal/shared/application"
)

// Compile-time interface satisfaction checks.
var (
	_ application.FileWriter = (*mockFileWriter)(nil)
	_ application.GitOps     = (*mockGitOps)(nil)
)

type mockFileWriter struct{}

func (m *mockFileWriter) WriteFile(_ context.Context, _ string, _ string) error {
	return nil
}

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

func (m *mockGitOps) CheckoutPrevious(_ context.Context, _ string) error {
	return nil
}

func (m *mockGitOps) DeleteBranch(_ context.Context, _ string, _ string) error {
	return nil
}
