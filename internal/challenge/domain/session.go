// Package domain provides the Challenge bounded context's core domain model.
package domain

import "strconv"

// SessionStatus represents the current state of a challenge session.
type SessionStatus string

const (
	// SessionStatusActive means the session is accepting responses.
	SessionStatusActive SessionStatus = "active"
	// SessionStatusCompleted means all challenges have been responded to.
	SessionStatusCompleted SessionStatus = "completed"
)

// ChallengeSession tracks an in-progress challenge round.
// It is an aggregate root that maintains the invariant: each challenge
// can only be responded to once.
type ChallengeSession struct {
	sessionID     string
	challenges    []Challenge
	responses     map[string]ChallengeResponse // keyed by challengeID
	status        SessionStatus
	domainModelID string
}

// NewChallengeSession creates a new session with the given ID and challenges.
func NewChallengeSession(sessionID, domainModelID string, challenges []Challenge) *ChallengeSession {
	c := make([]Challenge, len(challenges))
	copy(c, challenges)
	return &ChallengeSession{
		sessionID:     sessionID,
		domainModelID: domainModelID,
		challenges:    c,
		responses:     make(map[string]ChallengeResponse),
		status:        SessionStatusActive,
	}
}

// SessionID returns the unique identifier for this session.
func (s *ChallengeSession) SessionID() string { return s.sessionID }

// DomainModelID returns the ID of the domain model being challenged.
func (s *ChallengeSession) DomainModelID() string { return s.domainModelID }

// Status returns the current session status.
func (s *ChallengeSession) Status() SessionStatus { return s.status }

// Challenges returns a defensive copy of all challenges.
func (s *ChallengeSession) Challenges() []Challenge {
	out := make([]Challenge, len(s.challenges))
	copy(out, s.challenges)
	return out
}

// Responses returns a defensive copy of all responses.
func (s *ChallengeSession) Responses() []ChallengeResponse {
	out := make([]ChallengeResponse, 0, len(s.responses))
	for _, r := range s.responses {
		out = append(out, r)
	}
	return out
}

// ChallengeByID returns the challenge with the given ID, or false if not found.
func (s *ChallengeSession) ChallengeByID(challengeID string) (Challenge, bool) {
	for i, c := range s.challenges {
		if challengeID == s.challengeIDAt(i) {
			return c, true
		}
	}
	return Challenge{}, false
}

// HasResponse returns true if the challenge has already been responded to.
func (s *ChallengeSession) HasResponse(challengeID string) bool {
	_, exists := s.responses[challengeID]
	return exists
}

// RecordResponse records a response to a challenge.
// Returns error if the challenge doesn't exist or was already answered.
func (s *ChallengeSession) RecordResponse(response ChallengeResponse) error {
	challengeID := response.ChallengeID()

	// Check challenge exists
	if _, found := s.ChallengeByID(challengeID); !found {
		return ErrChallengeNotFound
	}

	// Check not already answered
	if s.HasResponse(challengeID) {
		return ErrChallengeAlreadyAnswered
	}

	s.responses[challengeID] = response

	// Auto-complete when all challenges answered
	if len(s.responses) == len(s.challenges) {
		s.status = SessionStatusCompleted
	}

	return nil
}

// challengeIDAt returns a deterministic ID for the challenge at index i.
// Format: "c{index}" (e.g., "c0", "c1", "c10").
func (s *ChallengeSession) challengeIDAt(i int) string {
	return "c" + strconv.Itoa(i)
}

// ChallengeIDs returns the IDs for all challenges in this session.
func (s *ChallengeSession) ChallengeIDs() []string {
	ids := make([]string, len(s.challenges))
	for i := range s.challenges {
		ids[i] = s.challengeIDAt(i)
	}
	return ids
}

// ToIteration converts the completed session to a ChallengeIteration.
func (s *ChallengeSession) ToIteration() ChallengeIteration {
	responses := make([]ChallengeResponse, 0, len(s.responses))
	for _, r := range s.responses {
		responses = append(responses, r)
	}

	delta := 0
	for _, r := range responses {
		if r.Accepted() {
			delta += len(r.ArtifactUpdates())
		}
	}

	return NewChallengeIteration(s.challenges, responses, delta)
}
