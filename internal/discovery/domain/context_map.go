package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ContextMap is a container for bounded context sketches and their strategic relationships.
type ContextMap struct {
	project       string
	contexts      []BoundedContextSketch
	relationships []ContextRelationship
}

// NewContextMap creates a ContextMap, validating invariants.
// Contexts and relationships may be empty.
func NewContextMap(
	project string,
	contexts []BoundedContextSketch,
	relationships []ContextRelationship,
) (*ContextMap, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return nil, fmt.Errorf("context map project must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	// Defensive copies of slices.
	ctxCopy := make([]BoundedContextSketch, len(contexts))
	copy(ctxCopy, contexts)

	relCopy := make([]ContextRelationship, len(relationships))
	copy(relCopy, relationships)

	return &ContextMap{
		project:       project,
		contexts:      ctxCopy,
		relationships: relCopy,
	}, nil
}

// Project returns the project name.
func (m *ContextMap) Project() string { return m.project }

// Contexts returns a defensive copy of the bounded context sketches.
func (m *ContextMap) Contexts() []BoundedContextSketch {
	out := make([]BoundedContextSketch, len(m.contexts))
	copy(out, m.contexts)

	return out
}

// Relationships returns a defensive copy of the context relationships.
func (m *ContextMap) Relationships() []ContextRelationship {
	out := make([]ContextRelationship, len(m.relationships))
	copy(out, m.relationships)

	return out
}
