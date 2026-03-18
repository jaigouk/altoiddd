package application_test

import (
	"context"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/tooltranslation/application"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// Compile-time interface satisfaction checks.
var (
	_ application.ConfigGeneration = (*mockConfigGeneration)(nil)
	_ application.PersonaManager   = (*mockPersonaManager)(nil)
)

// --- mockConfigGeneration ---

type mockConfigGeneration struct{}

func (m *mockConfigGeneration) Generate(_ context.Context, _ *ddd.DomainModel, _ []ttdomain.SupportedTool, _ string) error {
	return nil
}

// --- mockPersonaManager ---

type mockPersonaManager struct{}

func (m *mockPersonaManager) ListPersonas(_ context.Context) ([]*vo.PersonaDefinition, error) {
	return nil, nil
}

func (m *mockPersonaManager) Generate(_ context.Context, _ string, _ []string, _ string) error {
	return nil
}
