package domain

import (
	"fmt"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// Story count constants — DDD.md invariant 3.
const (
	rapidRequiredStories    = 3
	thoroughRequiredStories = 5
)

// Checkpoint constants — DDD.md invariants 6 and 8.
const (
	midStoryCheckpointEvery  = 3 // DDD.md invariant 6
	boundaryDetectionMinimum = 2 // DDD.md invariant 8
)

// StorytellingFlow implements DiscoveryFlow for RAPID and THOROUGH
// story-driven discovery modes.
type StorytellingFlow struct {
	mode            DiscoveryMode
	requiredStories int
}

// Compile-time check that StorytellingFlow satisfies DiscoveryFlow.
var _ DiscoveryFlow = (*StorytellingFlow)(nil)

// NewStorytellingFlow creates a StorytellingFlow for the given mode.
// Returns error if mode is not ModeRapid or ModeThorough.
func NewStorytellingFlow(mode DiscoveryMode) (*StorytellingFlow, error) {
	var required int

	switch mode {
	case ModeRapid:
		required = rapidRequiredStories
	case ModeThorough:
		required = thoroughRequiredStories
	case ModeExpress, ModeDeep, ModeConversational:
		return nil, fmt.Errorf("storytelling flow requires rapid or thorough mode, got %q: %w",
			mode, domainerrors.ErrInvariantViolation)
	default:
		return nil, fmt.Errorf("storytelling flow requires rapid or thorough mode, got %q: %w",
			mode, domainerrors.ErrInvariantViolation)
	}

	return &StorytellingFlow{mode: mode, requiredStories: required}, nil
}

// ValidateQuestionOrder is a no-op for storytelling flow.
// Storytelling has no question ordering constraints.
func (f *StorytellingFlow) ValidateQuestionOrder(_ QuestionRef, _ []Answer, _ map[string]bool) error {
	return nil
}

// IsPlaybackDue returns true when at least one story has completed since the last checkpoint.
func (f *StorytellingFlow) IsPlaybackDue(storiesSinceLastCheckpoint int) bool {
	return storiesSinceLastCheckpoint >= 1
}

// PlaybackInterval returns 1 — synthesis after every story.
func (f *StorytellingFlow) PlaybackInterval() int { return 1 }

// CheckCompleteness is a no-op for storytelling flow.
// Story completeness is checked via CheckStoryCompleteness with a type assertion.
func (f *StorytellingFlow) CheckCompleteness(_ []Answer, _ map[string]bool) error {
	return nil
}

// RequiredStoryCount returns the minimum number of stories for this mode.
// RAPID=3, THOROUGH=5 (DDD.md invariant 3).
func (f *StorytellingFlow) RequiredStoryCount() int {
	return f.requiredStories
}

// CheckStoryCompleteness returns nil if completedStoryCount >= required stories.
// Returns ErrInvariantViolation if below the threshold.
func (f *StorytellingFlow) CheckStoryCompleteness(completedStoryCount int) error {
	if completedStoryCount < f.requiredStories {
		return fmt.Errorf("storytelling %s mode requires at least %d stories, got %d: %w",
			f.mode, f.requiredStories, completedStoryCount, domainerrors.ErrInvariantViolation)
	}

	return nil
}

// IsSynthesisCheckpointDue returns true when completedStoryCount >= 1.
func (f *StorytellingFlow) IsSynthesisCheckpointDue(completedStoryCount int) bool {
	return completedStoryCount >= 1
}

// IsMidStoryCheckpointDue returns true when sentencesSinceLastCheckpoint >= 3.
func (f *StorytellingFlow) IsMidStoryCheckpointDue(sentencesSinceLastCheckpoint int) bool {
	return sentencesSinceLastCheckpoint >= midStoryCheckpointEvery
}

// CanRunBoundaryDetection returns true when completedStoryCount >= 2.
func (f *StorytellingFlow) CanRunBoundaryDetection(completedStoryCount int) bool {
	return completedStoryCount >= boundaryDetectionMinimum
}

// ShouldSuggestBranchingSplit delegates to story.HasBranching().
// Returns false for nil story (defensive, no panic).
func (f *StorytellingFlow) ShouldSuggestBranchingSplit(story *DomainStory) bool {
	if story == nil {
		return false
	}

	return story.HasBranching()
}

// Mode returns the active discovery mode.
func (f *StorytellingFlow) Mode() DiscoveryMode {
	return f.mode
}
