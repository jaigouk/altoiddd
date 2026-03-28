package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
)

// AgentStorytellingAdapter orchestrates the agent-mode storytelling discovery flow,
// emitting JSONL envelopes for each story and a discovery_complete at the end.
type AgentStorytellingAdapter struct {
	handler                  *application.DiscoveryHandler
	storytellingHandler      *application.StorytellingHandler
	boundaryDetectionHandler *application.BoundaryDetectionHandler
	writer                   io.Writer
	projectDir               string
}

// NewAgentStorytellingAdapter creates a new AgentStorytellingAdapter.
func NewAgentStorytellingAdapter(
	handler *application.DiscoveryHandler,
	storytellingHandler *application.StorytellingHandler,
	boundaryDetectionHandler *application.BoundaryDetectionHandler,
	writer io.Writer,
	projectDir string,
) *AgentStorytellingAdapter {
	return &AgentStorytellingAdapter{
		handler:                  handler,
		storytellingHandler:      storytellingHandler,
		boundaryDetectionHandler: boundaryDetectionHandler,
		writer:                   writer,
		projectDir:               projectDir,
	}
}

// Run executes the agent storytelling flow, emitting JSONL envelopes.
func (a *AgentStorytellingAdapter) Run(ctx context.Context) error {
	// 1. Read README.md
	readmePath := filepath.Join(a.projectDir, "README.md")

	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("reading README.md: %w", err)
	}

	// 2. Start session
	session, err := a.handler.StartSession(string(readme))
	if err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	sessionID := session.SessionID()

	// 3. Set mode to RAPID for agent mode
	if setModeErr := session.SetMode(domain.ModeRapid); setModeErr != nil {
		return fmt.Errorf("setting mode: %w", setModeErr)
	}

	// 4. Detect persona as Developer (choice "1")
	session, err = a.handler.DetectPersona(sessionID, "1")
	if err != nil {
		return fmt.Errorf("detecting persona: %w", err)
	}

	// 5. Create storytelling flow
	flow, err := domain.NewStorytellingFlow(domain.ModeRapid)
	if err != nil {
		return fmt.Errorf("creating storytelling flow: %w", err)
	}

	// 6. Story loop
	var stories []*domain.DomainStory

	for storyIndex := 1; ; storyIndex++ {
		story, _, storyErr := a.storytellingHandler.RunStory(ctx, session, storyIndex, flow)
		if storyErr != nil {
			return fmt.Errorf("running story %d: %w", storyIndex, storyErr)
		}

		stories = append(stories, story)

		// Emit story envelope
		storyOutput := NewStoryOutput(sessionID, storyIndex, story)

		storyData, marshalErr := json.Marshal(storyOutput)
		if marshalErr != nil {
			return fmt.Errorf("marshaling story %d: %w", storyIndex, marshalErr)
		}

		if writeErr := writeEnvelope(a.writer, "story", storyData); writeErr != nil {
			return writeErr
		}

		// Check completeness
		if flow.CheckStoryCompleteness(session.StoryCount()) == nil {
			break
		}
	}

	// 7. Boundary detection (if handler available and enough stories)
	sketchCount := 0

	if a.boundaryDetectionHandler != nil && flow.CanRunBoundaryDetection(session.StoryCount()) {
		sketches, detectErr := a.boundaryDetectionHandler.Detect(ctx, stories, domain.ModeRapid)
		if detectErr != nil {
			return fmt.Errorf("detecting boundaries: %w", detectErr)
		}

		sketchCount = len(sketches)

		for _, sketch := range sketches {
			proposal := BoundaryProposalOutput{
				Name:           sketch.Name(),
				Classification: string(sketch.Classification()),
				Confidence:     sketch.Confidence(),
				Actors:         sketch.Actors(),
				WorkObjects:    sketch.WorkObjects(),
				Stories:        sketch.Stories(),
			}

			proposalData, marshalErr := json.Marshal(proposal)
			if marshalErr != nil {
				return fmt.Errorf("marshaling boundary proposal: %w", marshalErr)
			}

			if writeErr := writeEnvelope(a.writer, "boundary_proposals", proposalData); writeErr != nil {
				return writeErr
			}
		}
	}

	// 8. Emit discovery_complete
	completeOutput := DiscoveryCompleteOutput{
		SessionID:   sessionID,
		StoryCount:  len(stories),
		SketchCount: sketchCount,
		Mode:        string(domain.ModeRapid),
	}

	completeData, err := json.Marshal(completeOutput)
	if err != nil {
		return fmt.Errorf("marshaling discovery complete: %w", err)
	}

	if writeErr := writeEnvelope(a.writer, "discovery_complete", completeData); writeErr != nil {
		return writeErr
	}

	return nil
}
