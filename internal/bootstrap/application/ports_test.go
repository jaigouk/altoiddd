package application_test

import (
	"context"

	"github.com/alto-cli/alto/internal/bootstrap/application"
)

// Compile-time interface satisfaction checks.
var _ application.Bootstrap = (*mockBootstrap)(nil)

type mockBootstrap struct{}

func (m *mockBootstrap) Preview(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (m *mockBootstrap) Confirm(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (m *mockBootstrap) Execute(_ context.Context, _ string) (string, error) {
	return "", nil
}
