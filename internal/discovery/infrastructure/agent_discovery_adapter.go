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

// AgentDiscoveryAdapter emits the full discovery session as JSONL for AI agent consumption.
// Each line is a JSON envelope with a "type" and "data" field.
type AgentDiscoveryAdapter struct {
	handler    *application.DiscoveryHandler
	renderer   *JSONSessionRenderer
	writer     io.Writer
	projectDir string
}

// NewAgentDiscoveryAdapter creates a new AgentDiscoveryAdapter.
func NewAgentDiscoveryAdapter(
	handler *application.DiscoveryHandler,
	renderer *JSONSessionRenderer,
	writer io.Writer,
	projectDir string,
) *AgentDiscoveryAdapter {
	return &AgentDiscoveryAdapter{
		handler:    handler,
		renderer:   renderer,
		writer:     writer,
		projectDir: projectDir,
	}
}

// envelope is the JSONL wrapper emitted for each line.
type envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Run executes the agent discovery flow, emitting session state as JSONL.
func (a *AgentDiscoveryAdapter) Run(_ context.Context) error {
	// Step 1: Read README.md
	readmePath := filepath.Join(a.projectDir, "README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("reading README.md: %w", err)
	}

	// Step 2: Start session
	session, err := a.handler.StartSession(string(readme))
	if err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	// Step 3: Emit initial session status
	statusData, err := a.renderer.RenderSessionStatus(session)
	if err != nil {
		return fmt.Errorf("rendering session status: %w", err)
	}
	if writeErr := writeEnvelope(a.writer, "session_status", statusData); writeErr != nil {
		return writeErr
	}

	// Step 4: Emit persona prompt
	personaData, err := a.renderer.RenderPersonaPrompt(session)
	if err != nil {
		return fmt.Errorf("rendering persona prompt: %w", err)
	}
	if writeErr := writeEnvelope(a.writer, "persona_prompt", personaData); writeErr != nil {
		return writeErr
	}

	// Step 5: Emit all questions
	questions := domain.QuestionCatalog()
	for _, question := range questions {
		qData, qErr := a.renderer.RenderQuestion(session, question)
		if qErr != nil {
			return fmt.Errorf("rendering question %s: %w", question.ID(), qErr)
		}
		if writeErr := writeEnvelope(a.writer, "question", qData); writeErr != nil {
			return writeErr
		}
	}

	// Step 6: Emit final session status
	finalData, err := a.renderer.RenderSessionStatus(session)
	if err != nil {
		return fmt.Errorf("rendering final session status: %w", err)
	}
	if writeErr := writeEnvelope(a.writer, "session_status", finalData); writeErr != nil {
		return writeErr
	}

	return nil
}

// writeEnvelope marshals an envelope and writes it as a single line.
func writeEnvelope(w io.Writer, typ string, data []byte) error {
	env := envelope{
		Type: typ,
		Data: json.RawMessage(data),
	}
	line, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshaling envelope: %w", err)
	}
	if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
		return fmt.Errorf("writing envelope: %w", err)
	}
	return nil
}
