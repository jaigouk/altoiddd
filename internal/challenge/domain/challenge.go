// Package domain provides the Challenge bounded context's core domain model.
// It contains value objects for the AI Challenger (Round 2): challenge types,
// challenges, responses, and iterations.
package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ChallengeType classifies what a challenge probes in the domain model.
type ChallengeType string

const (
	// ChallengeLanguage probes ambiguous terms used across contexts.
	ChallengeLanguage ChallengeType = "language"
	// ChallengeInvariant probes missing business rules on aggregates.
	ChallengeInvariant ChallengeType = "invariant"
	// ChallengeFailureMode probes unexamined failure paths in domain stories.
	ChallengeFailureMode ChallengeType = "failure_mode"
	// ChallengeBoundary probes questionable bounded context responsibilities.
	ChallengeBoundary ChallengeType = "boundary"
	// ChallengeAggregate probes aggregate design gaps.
	ChallengeAggregate ChallengeType = "aggregate"
	// ChallengeCommunication probes unclear inter-context integration patterns.
	ChallengeCommunication ChallengeType = "communication"
)

// AllChallengeTypes returns all valid ChallengeType values.
func AllChallengeTypes() []ChallengeType {
	return []ChallengeType{
		ChallengeLanguage,
		ChallengeInvariant,
		ChallengeFailureMode,
		ChallengeBoundary,
		ChallengeAggregate,
		ChallengeCommunication,
	}
}

// Challenge is a typed question that probes the domain model for gaps.
// Follows CHALLENGE-AS-QUESTION pattern: always a question, never a fact.
type Challenge struct {
	challengeType   ChallengeType
	questionText    string
	contextName     string
	sourceReference string
	evidence        string
}

// NewChallenge creates a validated Challenge value object.
func NewChallenge(
	challengeType ChallengeType,
	questionText, contextName, sourceReference, evidence string,
) (Challenge, error) {
	if strings.TrimSpace(questionText) == "" {
		return Challenge{}, fmt.Errorf("challenge question_text cannot be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(contextName) == "" {
		return Challenge{}, fmt.Errorf("challenge context_name cannot be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(sourceReference) == "" {
		return Challenge{}, fmt.Errorf("challenge source_reference cannot be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	return Challenge{
		challengeType:   challengeType,
		questionText:    questionText,
		contextName:     contextName,
		sourceReference: sourceReference,
		evidence:        evidence,
	}, nil
}

// ChallengeType returns the challenge classification.
func (c Challenge) ChallengeType() ChallengeType { return c.challengeType }

// QuestionText returns the challenge question.
func (c Challenge) QuestionText() string { return c.questionText }

// ContextName returns which bounded context this targets.
func (c Challenge) ContextName() string { return c.contextName }

// SourceReference returns evidence or citation backing this challenge.
func (c Challenge) SourceReference() string { return c.sourceReference }

// Evidence returns optional supporting detail.
func (c Challenge) Evidence() string { return c.evidence }

// ChallengeResponse is a user's response to a single challenge.
type ChallengeResponse struct {
	challengeID     string
	userResponse    string
	artifactUpdates []string
	accepted        bool
}

// NewChallengeResponse creates a ChallengeResponse value object.
func NewChallengeResponse(challengeID, userResponse string, accepted bool, artifactUpdates []string) ChallengeResponse {
	updates := make([]string, len(artifactUpdates))
	copy(updates, artifactUpdates)
	return ChallengeResponse{
		challengeID:     challengeID,
		userResponse:    userResponse,
		accepted:        accepted,
		artifactUpdates: updates,
	}
}

// ChallengeID returns the identifier linking back to the challenge.
func (r ChallengeResponse) ChallengeID() string { return r.challengeID }

// UserResponse returns what the user said.
func (r ChallengeResponse) UserResponse() string { return r.userResponse }

// Accepted returns whether the user accepted the challenge's premise.
func (r ChallengeResponse) Accepted() bool { return r.accepted }

// ArtifactUpdates returns a defensive copy of DDD.md changes prompted by this response.
func (r ChallengeResponse) ArtifactUpdates() []string {
	out := make([]string, len(r.artifactUpdates))
	copy(out, r.artifactUpdates)
	return out
}

// ChallengeIteration is a complete challenge round: all challenges posed and responses received.
type ChallengeIteration struct {
	challenges       []Challenge
	responses        []ChallengeResponse
	convergenceDelta int
}

// NewChallengeIteration creates a ChallengeIteration value object.
func NewChallengeIteration(challenges []Challenge, responses []ChallengeResponse, convergenceDelta int) ChallengeIteration {
	c := make([]Challenge, len(challenges))
	copy(c, challenges)
	r := make([]ChallengeResponse, len(responses))
	copy(r, responses)
	return ChallengeIteration{
		challenges:       c,
		responses:        r,
		convergenceDelta: convergenceDelta,
	}
}

// Challenges returns a defensive copy of all challenges in this iteration.
func (ci ChallengeIteration) Challenges() []Challenge {
	out := make([]Challenge, len(ci.challenges))
	copy(out, ci.challenges)
	return out
}

// Responses returns a defensive copy of all responses in this iteration.
func (ci ChallengeIteration) Responses() []ChallengeResponse {
	out := make([]ChallengeResponse, len(ci.responses))
	copy(out, ci.responses)
	return out
}

// ConvergenceDelta returns the count of model changes in this iteration.
func (ci ChallengeIteration) ConvergenceDelta() int { return ci.convergenceDelta }
