// Package domain provides the Knowledge bounded context's core domain model.
// It contains value objects for the knowledge base: categories, paths, metadata,
// entries, and drift detection.
package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// KnowledgeCategory enumerates knowledge entry categories.
type KnowledgeCategory string

// Knowledge category constants.
const (
	CategoryDDD         KnowledgeCategory = "ddd"
	CategoryTools       KnowledgeCategory = "tools"
	CategoryConventions KnowledgeCategory = "conventions"
	CategoryCrossTool   KnowledgeCategory = "cross-tool"
)

var validCategories = map[string]KnowledgeCategory{
	"ddd":         CategoryDDD,
	"tools":       CategoryTools,
	"conventions": CategoryConventions,
	"cross-tool":  CategoryCrossTool,
}

// AllCategoryValues returns all valid category string values.
func AllCategoryValues() []string {
	return []string{
		string(CategoryDDD),
		string(CategoryTools),
		string(CategoryConventions),
		string(CategoryCrossTool),
	}
}

// ParseCategory parses a string into a KnowledgeCategory.
func ParseCategory(s string) (KnowledgeCategory, error) {
	cat, ok := validCategories[s]
	if !ok {
		return "", fmt.Errorf("invalid knowledge category %q: %w", s, domainerrors.ErrInvariantViolation)
	}
	return cat, nil
}

// KnowledgePath is an RLM-addressable path to a knowledge entry.
type KnowledgePath struct {
	raw string
}

// NewKnowledgePath creates a validated KnowledgePath.
func NewKnowledgePath(raw string) (KnowledgePath, error) {
	if raw == "" {
		return KnowledgePath{}, fmt.Errorf("knowledge path must not be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	if strings.Contains(raw, "..") {
		return KnowledgePath{}, fmt.Errorf("knowledge path must not contain path traversal (..): %w",
			domainerrors.ErrInvariantViolation)
	}
	segments := strings.Split(raw, "/")
	if _, ok := validCategories[segments[0]]; !ok {
		sorted := []string{"conventions", "cross-tool", "ddd", "tools"}
		return KnowledgePath{}, fmt.Errorf(
			"knowledge path must start with a valid category (%s), got '%s': %w",
			strings.Join(sorted, ", "), segments[0], domainerrors.ErrInvariantViolation)
	}
	if len(segments) < 2 || segments[1] == "" {
		return KnowledgePath{}, fmt.Errorf(
			"knowledge path requires category/topic format, got %q: %w",
			raw, domainerrors.ErrInvariantViolation)
	}
	return KnowledgePath{raw: raw}, nil
}

// Raw returns the raw path string.
func (p KnowledgePath) Raw() string { return p.raw }

// Category extracts the category from the first path segment.
func (p KnowledgePath) Category() KnowledgeCategory {
	first := strings.Split(p.raw, "/")[0]
	return validCategories[first]
}

// Topic extracts the topic portion of the path.
func (p KnowledgePath) Topic() string {
	segments := strings.Split(p.raw, "/")
	if p.Category() == CategoryTools {
		return strings.Join(segments[1:], "/")
	}
	return segments[1]
}

// Tool extracts the tool name for TOOLS category paths, nil otherwise.
func (p KnowledgePath) Tool() *string {
	if p.Category() != CategoryTools {
		return nil
	}
	segments := strings.Split(p.raw, "/")
	return &segments[1]
}

// Subtopic extracts the subtopic for TOOLS category paths, nil otherwise.
func (p KnowledgePath) Subtopic() *string {
	if p.Category() != CategoryTools {
		return nil
	}
	segments := strings.Split(p.raw, "/")
	if len(segments) < 3 {
		return nil
	}
	return &segments[2]
}

// EntryMetadata holds verification and confidence metadata for a knowledge entry.
type EntryMetadata struct {
	lastVerified    string
	verifiedAgainst string
	confidence      string
	nextReviewDate  string
	schemaVersion   string
	sourceURLs      []string
	deprecated      bool
}

// NewEntryMetadata creates an EntryMetadata value object.
func NewEntryMetadata(
	lastVerified, verifiedAgainst, confidence string,
	deprecated bool,
	nextReviewDate, schemaVersion string,
	sourceURLs []string,
) EntryMetadata {
	urls := make([]string, len(sourceURLs))
	copy(urls, sourceURLs)
	return EntryMetadata{
		lastVerified:    lastVerified,
		verifiedAgainst: verifiedAgainst,
		confidence:      confidence,
		deprecated:      deprecated,
		nextReviewDate:  nextReviewDate,
		schemaVersion:   schemaVersion,
		sourceURLs:      urls,
	}
}

// LastVerified returns when the entry was last verified.
func (m EntryMetadata) LastVerified() string { return m.lastVerified }

// VerifiedAgainst returns what the entry was verified against.
func (m EntryMetadata) VerifiedAgainst() string { return m.verifiedAgainst }

// Confidence returns the confidence level.
func (m EntryMetadata) Confidence() string { return m.confidence }

// Deprecated returns whether the entry is deprecated.
func (m EntryMetadata) Deprecated() bool { return m.deprecated }

// NextReviewDate returns when the entry should next be reviewed.
func (m EntryMetadata) NextReviewDate() string { return m.nextReviewDate }

// SchemaVersion returns the schema version.
func (m EntryMetadata) SchemaVersion() string { return m.schemaVersion }

// SourceURLs returns a defensive copy.
func (m EntryMetadata) SourceURLs() []string {
	out := make([]string, len(m.sourceURLs))
	copy(out, m.sourceURLs)
	return out
}

// KnowledgeEntry is an entity in the Knowledge Base, identified by its KnowledgePath.
type KnowledgeEntry struct {
	path     KnowledgePath
	title    string
	content  string
	metadata *EntryMetadata
	format   string
}

// NewKnowledgeEntry creates a KnowledgeEntry entity.
func NewKnowledgeEntry(path KnowledgePath, title, content string, metadata *EntryMetadata, format string) KnowledgeEntry {
	return KnowledgeEntry{
		path:     path,
		title:    title,
		content:  content,
		metadata: metadata,
		format:   format,
	}
}

// Path returns the knowledge entry path.
func (e KnowledgeEntry) Path() KnowledgePath { return e.path }

// Title returns the knowledge entry title.
func (e KnowledgeEntry) Title() string { return e.title }

// Content returns the knowledge entry content.
func (e KnowledgeEntry) Content() string { return e.content }

// Metadata returns the entry metadata.
func (e KnowledgeEntry) Metadata() *EntryMetadata { return e.metadata }

// Format returns the entry format.
func (e KnowledgeEntry) Format() string { return e.format }

// EqualByPath compares two entries by path identity.
func (e KnowledgeEntry) EqualByPath(other KnowledgeEntry) bool {
	return e.path.raw == other.path.raw
}
