// Package ddd provides core DDD building blocks shared across bounded contexts.
package ddd

import (
	"fmt"
	"strings"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// BoundedContext represents a DDD bounded context with its aggregate roots.
// Fields are unexported to enforce invariants through the constructor.
type BoundedContext struct {
	name       string
	aggregates []string
}

// NewBoundedContext creates a validated BoundedContext.
// Returns ErrInvariantViolation if name is empty/whitespace or aggregates contain duplicates.
func NewBoundedContext(name string, aggregates []string) (*BoundedContext, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("bounded context name cannot be empty: %w", domainerrors.ErrInvariantViolation)
	}

	aggs := make([]string, 0, len(aggregates))
	seen := make(map[string]struct{}, len(aggregates))

	for _, agg := range aggregates {
		if _, exists := seen[agg]; exists {
			return nil, fmt.Errorf("duplicate aggregate %q in bounded context %q: %w", agg, name, domainerrors.ErrInvariantViolation)
		}
		seen[agg] = struct{}{}
		aggs = append(aggs, agg)
	}

	return &BoundedContext{
		name:       name,
		aggregates: aggs,
	}, nil
}

// Name returns the bounded context name.
func (bc *BoundedContext) Name() string {
	return bc.name
}

// Aggregates returns a defensive copy of the aggregate root names.
func (bc *BoundedContext) Aggregates() []string {
	out := make([]string, len(bc.aggregates))
	copy(out, bc.aggregates)
	return out
}
