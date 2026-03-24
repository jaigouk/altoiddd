package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ConversationTurn captures a single exchange between the discovery consultant and the user.
type ConversationTurn struct {
	consultantAction string
	userResponse     string
	synthesis        string
	confirmed        bool
}

// NewConversationTurn creates a ConversationTurn, validating that both fields are non-empty.
func NewConversationTurn(consultantAction, userResponse string) (ConversationTurn, error) {
	consultantAction = strings.TrimSpace(consultantAction)
	if consultantAction == "" {
		return ConversationTurn{}, fmt.Errorf("consultant action must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	userResponse = strings.TrimSpace(userResponse)
	if userResponse == "" {
		return ConversationTurn{}, fmt.Errorf("user response must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	return ConversationTurn{
		consultantAction: consultantAction,
		userResponse:     userResponse,
	}, nil
}

// ConsultantAction returns the moderator's question or proposal.
func (c ConversationTurn) ConsultantAction() string { return c.consultantAction }

// UserResponse returns the user's answer.
func (c ConversationTurn) UserResponse() string { return c.userResponse }

// Synthesis returns the optional synthesis text for this turn.
func (c ConversationTurn) Synthesis() string { return c.synthesis }

// IsConfirmed returns whether the user confirmed the synthesis.
func (c ConversationTurn) IsConfirmed() bool { return c.confirmed }

// WithSynthesis returns a new ConversationTurn with the given synthesis text.
func (c ConversationTurn) WithSynthesis(text string) ConversationTurn {
	c.synthesis = text
	return c
}

// Confirm returns a new ConversationTurn marked as confirmed.
func (c ConversationTurn) Confirm() ConversationTurn {
	c.confirmed = true
	return c
}

// String returns a human-readable representation of the turn.
func (c ConversationTurn) String() string {
	status := "unconfirmed"
	if c.confirmed {
		status = "confirmed"
	}

	return fmt.Sprintf("Turn: [%s] %s → %s", status, c.consultantAction, c.userResponse)
}
