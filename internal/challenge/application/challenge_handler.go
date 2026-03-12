// Package application provides command handlers for the Challenge bounded context.
package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// ChallengeHandler orchestrates the research -> generate -> respond -> complete
// cycle for Round 2 challenges.
type ChallengeHandler struct {
	challenger Challenger
	challenges []challengedomain.Challenge
	responses  []challengedomain.ChallengeResponse

	// Session management
	sessions map[string]*challengedomain.ChallengeSession
	mu       sync.RWMutex
}

// NewChallengeHandler creates a new ChallengeHandler with injected dependencies.
func NewChallengeHandler(challenger Challenger) *ChallengeHandler {
	return &ChallengeHandler{
		challenger: challenger,
		sessions:   make(map[string]*challengedomain.ChallengeSession),
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

// ---------------------------------------------------------------------------
// Session Management
// ---------------------------------------------------------------------------

// StartChallenge creates a new challenge session with generated challenges.
func (h *ChallengeHandler) StartChallenge(
	ctx context.Context,
	model *ddd.DomainModel,
	maxPerType int,
) (*challengedomain.ChallengeSession, error) {
	challenges, err := h.challenger.GenerateChallenges(ctx, model, maxPerType)
	if err != nil {
		return nil, fmt.Errorf("generate challenges: %w", err)
	}

	sessionID := uuid.New().String()
	session := challengedomain.NewChallengeSession(sessionID, model.ModelID(), challenges)

	h.mu.Lock()
	h.sessions[sessionID] = session
	h.mu.Unlock()

	return session, nil
}

// GetSession retrieves a session by ID.
// NOTE: The returned pointer is for read-only inspection. To modify the session
// (e.g., record responses), use RespondToChallenge which holds the write lock.
func (h *ChallengeHandler) GetSession(sessionID string) (*challengedomain.ChallengeSession, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	session, exists := h.sessions[sessionID]
	if !exists {
		return nil, challengedomain.ErrSessionNotFound
	}
	return session, nil
}

// RespondToChallenge records a response to a challenge in a session.
func (h *ChallengeHandler) RespondToChallenge(
	sessionID, challengeID, userResponse string,
	accepted bool,
	artifactUpdates []string,
) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, exists := h.sessions[sessionID]
	if !exists {
		return challengedomain.ErrSessionNotFound
	}

	response := challengedomain.NewChallengeResponse(challengeID, userResponse, accepted, artifactUpdates)
	if err := session.RecordResponse(response); err != nil {
		return fmt.Errorf("recording response: %w", err)
	}
	return nil
}

// CompleteSession finalizes a session and returns its iteration summary.
func (h *ChallengeHandler) CompleteSession(sessionID string) (challengedomain.ChallengeIteration, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	session, exists := h.sessions[sessionID]
	if !exists {
		return challengedomain.ChallengeIteration{}, challengedomain.ErrSessionNotFound
	}

	return session.ToIteration(), nil
}
