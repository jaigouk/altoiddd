// Package application defines ports for the Challenge bounded context.
package application

import (
	"context"

	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// Challenger generates typed challenges that probe a DomainModel for gaps:
// ambiguous language, missing invariants, unexamined failure modes, and
// questionable boundaries.
type Challenger interface {
	// GenerateChallenges generates typed challenges by analyzing the domain model.
	GenerateChallenges(ctx context.Context, model *ddd.DomainModel, maxPerType int) ([]challengedomain.Challenge, error)
}
