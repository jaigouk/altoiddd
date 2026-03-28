package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// Compile-time interface check.
var _ application.StorytellingPrompter = (*AgentStorytellingPrompter)(nil)

// AgentStorytellingPrompter implements StorytellingPrompter for non-interactive agent mode.
// It returns preconfigured responses: mode, sequential narration lines, and auto-accepts.
type AgentStorytellingPrompter struct {
	mode  discoverydomain.DiscoveryMode
	lines []string
	index int
}

// NewAgentStorytellingPrompter creates a new AgentStorytellingPrompter.
func NewAgentStorytellingPrompter(mode discoverydomain.DiscoveryMode, lines []string) *AgentStorytellingPrompter {
	return &AgentStorytellingPrompter{
		mode:  mode,
		lines: lines,
	}
}

// SelectMode returns the preconfigured discovery mode.
func (p *AgentStorytellingPrompter) SelectMode(_ context.Context) (discoverydomain.DiscoveryMode, error) {
	return p.mode, nil
}

// AskNarration returns the next line from the preconfigured lines.
// Returns empty string when all lines are exhausted.
func (p *AgentStorytellingPrompter) AskNarration(_ context.Context, _ string, _ string) (string, error) {
	if p.index >= len(p.lines) {
		return "", nil
	}

	line := p.lines[p.index]
	p.index++

	return line, nil
}

// ConfirmSentence auto-accepts the sentence unchanged.
func (p *AgentStorytellingPrompter) ConfirmSentence(_ context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	return sentence, true, nil
}

// AskChoice returns "1" (first option).
func (p *AgentStorytellingPrompter) AskChoice(_ context.Context, _ string, _ []application.Choice, _ string) (string, error) {
	return "1", nil
}

// AskAnnotation returns empty text, signaling no annotations in agent mode.
func (p *AgentStorytellingPrompter) AskAnnotation(_ context.Context) (string, int, error) {
	return "", 0, nil
}

// SynthesisCheckpoint auto-confirms.
func (p *AgentStorytellingPrompter) SynthesisCheckpoint(_ context.Context, _ application.SynthesisSummary) (bool, error) {
	return true, nil
}

// DisplayStory is a no-op for agent mode.
func (p *AgentStorytellingPrompter) DisplayStory(_ context.Context, _ *discoverydomain.DomainStory) error {
	return nil
}

// ProposeStory returns the proposed story unchanged.
func (p *AgentStorytellingPrompter) ProposeStory(_ context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	if proposed == nil {
		return nil, fmt.Errorf("proposed story must not be nil: %w", errors.ErrUnsupported)
	}

	return proposed, nil
}
