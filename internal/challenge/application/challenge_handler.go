// Package application provides command handlers for the Challenge bounded context.
package application

import (
	"context"
	"fmt"

	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// ChallengeHandler orchestrates the research -> generate -> respond -> complete
// cycle for Round 2 challenges.
type ChallengeHandler struct {
	challenger Challenger
	challenges []challengedomain.Challenge
	responses  []challengedomain.ChallengeResponse
}

// NewChallengeHandler creates a new ChallengeHandler with injected dependencies.
func NewChallengeHandler(challenger Challenger) *ChallengeHandler {
	return &ChallengeHandler{
		challenger: challenger,
	}
}

// GenerateChallenges delegates challenge generation to the port.
func (h *ChallengeHandler) GenerateChallenges(
	ctx context.Context,
	model *ddd.DomainModel,
	maxPerType int,
) ([]challengedomain.Challenge, error) {
	challenges, err := h.challenger.GenerateChallenges(ctx, model, maxPerType)
	if err != nil {
		return nil, fmt.Errorf("generate challenges: %w", err)
	}
	h.challenges = challenges
	return challenges, nil
}

// RecordResponse records a user response to a challenge.
func (h *ChallengeHandler) RecordResponse(response challengedomain.ChallengeResponse) {
	h.responses = append(h.responses, response)
}

// Complete finalizes the challenge round and produces a summary.
func (h *ChallengeHandler) Complete() challengedomain.ChallengeIteration {
	delta := 0
	for _, r := range h.responses {
		if r.Accepted() {
			delta += len(r.ArtifactUpdates())
		}
	}
	return challengedomain.NewChallengeIteration(h.challenges, h.responses, delta)
}
