package application

import (
	"context"
	"fmt"

	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
)

// TicketVerifyHandler verifies quantitative claims in tickets.
type TicketVerifyHandler struct {
	reader   TicketContentReader
	runner   CommandRunner
	verifier *ticketdomain.ClaimVerifier
}

// NewTicketVerifyHandler creates a handler for ticket claim verification.
func NewTicketVerifyHandler(reader TicketContentReader, runner CommandRunner) *TicketVerifyHandler {
	return &TicketVerifyHandler{
		reader:   reader,
		runner:   runner,
		verifier: ticketdomain.NewClaimVerifier(),
	}
}

// Verify parses claims from a ticket and verifies them.
// Returns verification results for all claims found.
func (h *TicketVerifyHandler) Verify(ctx context.Context, ticketID string) ([]ticketdomain.VerificationResult, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("ticket ID cannot be empty")
	}

	// Read ticket content
	content, err := h.reader.ReadTicketContent(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("reading ticket content: %w", err)
	}

	// Parse claims
	claims := h.verifier.ParseClaims(ticketID, content)
	if len(claims) == 0 {
		return nil, nil
	}

	// Verify each claim
	results := make([]ticketdomain.VerificationResult, 0, len(claims))
	for _, claim := range claims {
		var actualValue string
		var runErr error

		if claim.IsVerifiable() {
			actualValue, runErr = h.runner.Run(ctx, claim.Command())
		}

		result := ticketdomain.NewVerificationResult(claim, actualValue, runErr)
		results = append(results, result)
	}

	return results, nil
}
