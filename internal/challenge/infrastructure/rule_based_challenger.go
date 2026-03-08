// Package infrastructure provides adapters for the Challenge bounded context.
package infrastructure

import (
	"context"

	challengeapp "github.com/alty-cli/alty/internal/challenge/application"
	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// RuleBasedChallengerAdapter implements Challenger using rule-based heuristics.
// No LLM required -- delegates to the ChallengerService domain service.
type RuleBasedChallengerAdapter struct{}

// Compile-time interface check.
var _ challengeapp.Challenger = (*RuleBasedChallengerAdapter)(nil)

// GenerateChallenges generates challenges using the domain service.
func (r *RuleBasedChallengerAdapter) GenerateChallenges(
	_ context.Context,
	model *ddd.DomainModel,
	maxPerType int,
) ([]challengedomain.Challenge, error) {
	return challengedomain.Generate(model, maxPerType), nil
}
