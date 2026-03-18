package application_test

import (
	"context"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
)

// Compile-time interface satisfaction checks.
var (
	_ application.Discovery        = (*mockDiscovery)(nil)
	_ application.ArtifactRenderer = (*mockArtifactRenderer)(nil)
	_ application.ToolDetection    = (*mockToolDetection)(nil)
)

// --- mockDiscovery ---

type mockDiscovery struct{}

func (m *mockDiscovery) StartSession(_ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) DetectPersona(_ string, _ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) AnswerQuestion(_ string, _ string, _ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) SkipQuestion(_ string, _ string, _ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) ConfirmPlayback(_ string, _ bool) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

func (m *mockDiscovery) Complete(_ string) (*discoverydomain.DiscoverySession, error) {
	return nil, nil
}

// --- mockArtifactRenderer ---

type mockArtifactRenderer struct{}

func (m *mockArtifactRenderer) RenderPRD(_ context.Context, _ *ddd.DomainModel) (string, error) {
	return "", nil
}

func (m *mockArtifactRenderer) RenderDDD(_ context.Context, _ *ddd.DomainModel) (string, error) {
	return "", nil
}

func (m *mockArtifactRenderer) RenderArchitecture(_ context.Context, _ *ddd.DomainModel) (string, error) {
	return "", nil
}

// --- mockToolDetection ---

type mockToolDetection struct{}

func (m *mockToolDetection) Detect(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (m *mockToolDetection) ScanConflicts(_ context.Context, _ string) ([]discoverydomain.SettingsConflict, error) {
	return nil, nil
}
