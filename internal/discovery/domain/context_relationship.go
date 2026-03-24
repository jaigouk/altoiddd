package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ContextRelationship represents a strategic relationship between two bounded contexts.
type ContextRelationship struct {
	upstream    string
	downstream  string
	relType     RelationshipType
	shared      []string
	description string
}

// NewContextRelationship creates a ContextRelationship, enforcing all domain invariants.
func NewContextRelationship(
	upstream, downstream string,
	relType RelationshipType,
	shared []string,
	description string,
) (ContextRelationship, error) {
	upstream = strings.TrimSpace(upstream)
	if upstream == "" {
		return ContextRelationship{}, fmt.Errorf("context relationship upstream must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	downstream = strings.TrimSpace(downstream)
	if downstream == "" {
		return ContextRelationship{}, fmt.Errorf("context relationship downstream must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := relType.Validate(); err != nil {
		return ContextRelationship{}, fmt.Errorf("validating relationship type: %w", err)
	}

	// Defensive copy of shared slice.
	sharedCopy := make([]string, len(shared))
	copy(sharedCopy, shared)

	return ContextRelationship{
		upstream:    upstream,
		downstream:  downstream,
		relType:     relType,
		shared:      sharedCopy,
		description: description,
	}, nil
}

// Upstream returns the upstream context name.
func (r ContextRelationship) Upstream() string { return r.upstream }

// Downstream returns the downstream context name.
func (r ContextRelationship) Downstream() string { return r.downstream }

// Type returns the relationship type.
func (r ContextRelationship) Type() RelationshipType { return r.relType }

// Shared returns a defensive copy of the shared concepts.
func (r ContextRelationship) Shared() []string {
	out := make([]string, len(r.shared))
	copy(out, r.shared)

	return out
}

// Description returns the relationship description.
func (r ContextRelationship) Description() string { return r.description }

// String returns a human-readable representation of the context relationship.
func (r ContextRelationship) String() string {
	return fmt.Sprintf("ContextRelationship: %s -> %s (%s)", r.upstream, r.downstream, r.relType)
}
