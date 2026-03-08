// Package application defines ports for the Research bounded context.
package application

import (
	"context"

	researchdomain "github.com/alty-cli/alty/internal/research/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// DomainResearch searches for domain-specific knowledge from external sources.
// Findings carry source attribution and trust levels; the user confirms
// before facts enter the model.
type DomainResearch interface {
	// Research researches domain areas using external sources.
	Research(ctx context.Context, model *ddd.DomainModel, maxAreas int) (researchdomain.ResearchBriefing, error)
}

// SpikeFollowUp audits whether spike-defined follow-up intents have been
// created as beads tickets.
type SpikeFollowUp interface {
	// Audit audits a spike's follow-up intents against created tickets.
	Audit(ctx context.Context, spikeID string, projectDir string) (researchdomain.FollowUpAuditResult, error)
}

// SpikeReportParser extracts FollowUpIntent value objects from spike research
// reports. Infrastructure adapters implement this for specific formats.
type SpikeReportParser interface {
	// Parse extracts follow-up intents from a spike research report.
	Parse(ctx context.Context, reportPath string) ([]researchdomain.FollowUpIntent, error)
}

// WebSearch executes web search queries and returns raw results.
// Internal port consumed by research adapters, not exposed in AppContext.
type WebSearch interface {
	// Search executes a web search query.
	Search(ctx context.Context, query string, maxResults int) ([]researchdomain.WebSearchResult, error)
}
