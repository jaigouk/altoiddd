package domain

import (
	"fmt"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ResumeCheckpoint captures where a storytelling session was interrupted.
type ResumeCheckpoint struct {
	StoryIndex     int // next story to narrate, 1-based
	BoundariesDone bool
}

// ComputeResumeCheckpoint determines the resume point for a storytelling session.
// Returns ErrInvariantViolation if the session mode is not ModeRapid or ModeThorough,
// or if the mode is nil (pre-mode legacy sessions).
func (s *DiscoverySession) ComputeResumeCheckpoint() (ResumeCheckpoint, error) {
	if s.mode == nil {
		return ResumeCheckpoint{}, fmt.Errorf("session has no mode set (pre-mode legacy session): %w",
			domainerrors.ErrInvariantViolation)
	}

	switch *s.mode {
	case ModeRapid, ModeThorough:
		// valid — these are resumable storytelling modes
	case ModeExpress, ModeDeep, ModeConversational:
		return ResumeCheckpoint{}, fmt.Errorf("resume requires rapid or thorough mode, got %q: %w",
			*s.mode, domainerrors.ErrInvariantViolation)
	}

	return ResumeCheckpoint{
		StoryIndex:     s.StoryCount() + 1,
		BoundariesDone: s.boundariesConfirmed,
	}, nil
}
