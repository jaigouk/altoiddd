// Package domain provides the Research bounded context's core domain model.
// It contains value objects for research: trust levels, confidence, source attribution,
// web search results, research findings, and research briefings.
package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// TrustLevel classifies how much a piece of knowledge can be trusted.
// Lower numeric value = higher trust. Supports comparison operators via int.
type TrustLevel int

// Trust level constants ordered from most to least trusted.
const (
	TrustUserStated    TrustLevel = 1
	TrustUserConfirmed TrustLevel = 2
	TrustAIResearched  TrustLevel = 3
	TrustAIInferred    TrustLevel = 4
)

// AllTrustLevels returns all valid TrustLevel values.
func AllTrustLevels() []TrustLevel {
	return []TrustLevel{TrustUserStated, TrustUserConfirmed, TrustAIResearched, TrustAIInferred}
}

// Confidence classifies confidence level of a research finding or source.
type Confidence string

// Confidence level constants.
const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// AllConfidenceLevels returns all valid Confidence values.
func AllConfidenceLevels() []Confidence {
	return []Confidence{ConfidenceHigh, ConfidenceMedium, ConfidenceLow}
}

// SourceAttribution is provenance metadata for a research finding.
type SourceAttribution struct {
	url           string
	title         string
	retrievedDate string
	confidence    Confidence
}

// NewSourceAttribution creates a validated SourceAttribution.
func NewSourceAttribution(url, title, retrievedDate string, confidence Confidence) (SourceAttribution, error) {
	if strings.TrimSpace(url) == "" {
		return SourceAttribution{}, fmt.Errorf("sourceAttribution url cannot be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(title) == "" {
		return SourceAttribution{}, fmt.Errorf("sourceAttribution title cannot be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	return SourceAttribution{
		url:           url,
		title:         title,
		retrievedDate: retrievedDate,
		confidence:    confidence,
	}, nil
}

// URL returns the source URL.
func (sa SourceAttribution) URL() string { return sa.url }

// Title returns the source title.
func (sa SourceAttribution) Title() string { return sa.title }

// RetrievedDate returns when the source was retrieved.
func (sa SourceAttribution) RetrievedDate() string { return sa.retrievedDate }

// Confidence returns the confidence level.
func (sa SourceAttribution) Confidence() Confidence { return sa.confidence }

// WebSearchResult is a raw result from a web search query.
type WebSearchResult struct {
	url     string
	title   string
	snippet string
}

// NewWebSearchResult creates a WebSearchResult.
func NewWebSearchResult(url, title, snippet string) WebSearchResult {
	return WebSearchResult{url: url, title: title, snippet: snippet}
}

// URL returns the result URL.
func (w WebSearchResult) URL() string { return w.url }

// Title returns the result title.
func (w WebSearchResult) Title() string { return w.title }

// Snippet returns the result snippet.
func (w WebSearchResult) Snippet() string { return w.snippet }

// ResearchFinding is a single research insight with source attribution and trust level.
type ResearchFinding struct {
	content    string
	source     SourceAttribution
	domainArea string
	trustLevel TrustLevel
	outdated   bool
}

// NewResearchFinding creates a ResearchFinding.
func NewResearchFinding(content string, source SourceAttribution, trustLevel TrustLevel, domainArea string, outdated bool) ResearchFinding {
	return ResearchFinding{
		content:    content,
		source:     source,
		trustLevel: trustLevel,
		domainArea: domainArea,
		outdated:   outdated,
	}
}

// Content returns the finding content.
func (f ResearchFinding) Content() string { return f.content }

// Source returns the source attribution.
func (f ResearchFinding) Source() SourceAttribution { return f.source }

// TrustLevel returns the trust level.
func (f ResearchFinding) TrustLevel() TrustLevel { return f.trustLevel }

// DomainArea returns the domain area.
func (f ResearchFinding) DomainArea() string { return f.domainArea }

// Outdated returns whether the finding is outdated.
func (f ResearchFinding) Outdated() bool { return f.outdated }

// ResearchBriefing is the complete research output for presentation to the user.
type ResearchBriefing struct {
	summary     string
	findings    []ResearchFinding
	noDataAreas []string
}

// NewResearchBriefing creates a ResearchBriefing.
func NewResearchBriefing(findings []ResearchFinding, noDataAreas []string, summary string) ResearchBriefing {
	f := make([]ResearchFinding, len(findings))
	copy(f, findings)
	nda := make([]string, len(noDataAreas))
	copy(nda, noDataAreas)
	return ResearchBriefing{findings: f, noDataAreas: nda, summary: summary}
}

// Findings returns a defensive copy.
func (b ResearchBriefing) Findings() []ResearchFinding {
	out := make([]ResearchFinding, len(b.findings))
	copy(out, b.findings)
	return out
}

// NoDataAreas returns a defensive copy.
func (b ResearchBriefing) NoDataAreas() []string {
	out := make([]string, len(b.noDataAreas))
	copy(out, b.noDataAreas)
	return out
}

// Summary returns the human-readable summary.
func (b ResearchBriefing) Summary() string { return b.summary }
