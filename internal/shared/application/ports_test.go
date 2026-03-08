package application_test

import (
	"context"

	"github.com/alty-cli/alty/internal/shared/application"
)

// Compile-time interface satisfaction check.
var _ application.FileWriter = (*mockFileWriter)(nil)

type mockFileWriter struct{}

func (m *mockFileWriter) WriteFile(_ context.Context, _ string, _ string) error {
	return nil
}
