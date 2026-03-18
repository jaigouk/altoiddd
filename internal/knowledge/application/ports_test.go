package application_test

import (
	"context"

	"github.com/alto-cli/alto/internal/knowledge/application"
	knowledgedomain "github.com/alto-cli/alto/internal/knowledge/domain"
)

// Compile-time interface satisfaction checks.
var (
	_ application.KnowledgeLookup = (*mockKnowledgeLookup)(nil)
	_ application.DriftDetection  = (*mockDriftDetection)(nil)
)

// --- mockKnowledgeLookup ---

type mockKnowledgeLookup struct{}

func (m *mockKnowledgeLookup) Lookup(_ context.Context, _ string, _ string, _ string) (string, error) {
	return "", nil
}

func (m *mockKnowledgeLookup) ListTools(_ context.Context) ([]string, error) {
	return nil, nil
}

func (m *mockKnowledgeLookup) ListVersions(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (m *mockKnowledgeLookup) ListTopics(_ context.Context, _ string, _ *string) ([]string, error) {
	return nil, nil
}

// --- mockDriftDetection ---

type mockDriftDetection struct{}

func (m *mockDriftDetection) Detect(_ context.Context) (knowledgedomain.DriftReport, error) {
	return knowledgedomain.DriftReport{}, nil
}
