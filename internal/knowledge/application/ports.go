// Package application defines ports for the Knowledge bounded context.
package application

import (
	"context"

	knowledgedomain "github.com/alty-cli/alty/internal/knowledge/domain"
)

// KnowledgeLookup provides versioned, RLM-addressable access to DDD patterns,
// tool conventions, and coding standards with drift detection support.
type KnowledgeLookup interface {
	// Lookup looks up a specific knowledge entry.
	Lookup(ctx context.Context, category string, topic string, version string) (string, error)

	// ListTools lists all known AI coding tools.
	ListTools(ctx context.Context) ([]string, error)

	// ListVersions lists available versions for a specific tool.
	ListVersions(ctx context.Context, tool string) ([]string, error)

	// ListTopics lists available topics within a category.
	ListTopics(ctx context.Context, category string, tool *string) ([]string, error)
}

// DriftDetection scans knowledge entries for staleness, version-to-version
// changes, and doc-vs-code mismatches. Separate from KnowledgeLookup (ISP).
type DriftDetection interface {
	// Detect detects drift across all knowledge entries.
	Detect(ctx context.Context) (knowledgedomain.DriftReport, error)
}
