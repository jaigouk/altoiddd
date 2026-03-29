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
	researcher               application.DomainResearcher
	writer                   io.Writer
	projectDir               string
}

// NewAgentStorytellingAdapter creates a new AgentStorytellingAdapter.
func NewAgentStorytellingAdapter(
	handler *application.DiscoveryHandler,
	storytellingHandler *application.StorytellingHandler,
	boundaryDetectionHandler *application.BoundaryDetectionHandler,
	researcher application.DomainResearcher,
	writer io.Writer,
	projectDir string,
) *AgentStorytellingAdapter {
	return &AgentStorytellingAdapter{
		handler:                  handler,
		storytellingHandler:      storytellingHandler,
		boundaryDetectionHandler: boundaryDetectionHandler,
		researcher:               researcher,
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
	if _, err = a.handler.DetectPersona(sessionID, "1"); err != nil {
		return fmt.Errorf("detecting persona: %w", err)
	}

	// 5. Create storytelling flow
	flow, err := domain.NewStorytellingFlow(domain.ModeRapid)
	if err != nil {
		return fmt.Errorf("creating storytelling flow: %w", err)
	}

	// 6. Research domain and generate stories
	if a.researcher == nil {
		return fmt.Errorf("agent mode requires a domain researcher")
	}

	result, err := a.researcher.Research(ctx, string(readme))
	if err != nil {
		return fmt.Errorf("researching domain: %w", err)
	}

	if result == nil {
		return fmt.Errorf("agent mode requires LLM credentials for automated story generation. Set ANTHROPIC_API_KEY or use alto guide --no-tui for interactive mode")
	}

	stories, err := a.storytellingHandler.ProposeResearchStories(ctx, result)
	if err != nil {
		return fmt.Errorf("proposing research stories: %w", err)
	}

	if len(stories) == 0 {
		return fmt.Errorf("research produced no stories for this domain; the README may lack sufficient detail")
	}

	// 7. Emit story envelopes
	for i, story := range stories {
		storyOutput := NewStoryOutput(sessionID, i+1, story)

		data, marshalErr := json.Marshal(storyOutput)
		if marshalErr != nil {
			return fmt.Errorf("marshalling story output: %w", marshalErr)
		}

		if writeErr := writeEnvelope(a.writer, "story", data); writeErr != nil {
			return fmt.Errorf("writing story envelope: %w", writeErr)
		}
	}

	// 8. Boundary detection (if handler available and enough stories)
	sketchCount := 0

	if a.boundaryDetectionHandler != nil && flow.CanRunBoundaryDetection(len(stories)) {
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

	// 9. Emit discovery_complete
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
