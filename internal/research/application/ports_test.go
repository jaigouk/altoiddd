package application_test

import (
	"context"

	"github.com/alty-cli/alty/internal/research/application"
	researchdomain "github.com/alty-cli/alty/internal/research/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// Compile-time interface satisfaction checks.
var (
	_ application.DomainResearch    = (*mockDomainResearch)(nil)
	_ application.SpikeFollowUp     = (*mockSpikeFollowUp)(nil)
	_ application.SpikeReportParser = (*mockSpikeReportParser)(nil)
	_ application.WebSearch         = (*mockWebSearch)(nil)
)

// --- mockDomainResearch ---

type mockDomainResearch struct{}

func (m *mockDomainResearch) Research(_ context.Context, _ *ddd.DomainModel, _ int) (researchdomain.ResearchBriefing, error) {
	return researchdomain.ResearchBriefing{}, nil
}

// --- mockSpikeFollowUp ---

type mockSpikeFollowUp struct{}

func (m *mockSpikeFollowUp) Audit(_ context.Context, _ string, _ string) (researchdomain.FollowUpAuditResult, error) {
	return researchdomain.FollowUpAuditResult{}, nil
}

// --- mockSpikeReportParser ---

type mockSpikeReportParser struct{}

func (m *mockSpikeReportParser) Parse(_ context.Context, _ string) ([]researchdomain.FollowUpIntent, error) {
	return nil, nil
}

// --- mockWebSearch ---

type mockWebSearch struct{}

func (m *mockWebSearch) Search(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
	return nil, nil
}
