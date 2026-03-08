package application

import (
	"context"
	"fmt"

	researchdomain "github.com/alty-cli/alty/internal/research/domain"
)

// SpikeFollowUpHandler orchestrates spike follow-up auditing.
// It delegates to the SpikeFollowUp port for scanning reports and matching tickets.
type SpikeFollowUpHandler struct {
	followUp SpikeFollowUp
}

// NewSpikeFollowUpHandler creates a new SpikeFollowUpHandler.
func NewSpikeFollowUpHandler(followUp SpikeFollowUp) *SpikeFollowUpHandler {
	return &SpikeFollowUpHandler{followUp: followUp}
}

// Audit audits a spike's follow-up intents against created tickets.
func (h *SpikeFollowUpHandler) Audit(ctx context.Context, spikeID, projectDir string) (researchdomain.FollowUpAuditResult, error) {
	result, err := h.followUp.Audit(ctx, spikeID, projectDir)
	if err != nil {
		return researchdomain.FollowUpAuditResult{}, fmt.Errorf("spike follow-up audit: %w", err)
	}
	return result, nil
}
