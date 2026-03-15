package infrastructure_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
	"github.com/alty-cli/alty/internal/discovery/infrastructure"
)

// testPublisher is a no-op event publisher for agent adapter tests.
type testPublisher struct{}

func (p *testPublisher) Publish(_ context.Context, _ any) error { return nil }

// setupAgentAdapter creates a temp dir with README.md and returns a configured adapter + buffer.
func setupAgentAdapter(t *testing.T) (*infrastructure.AgentDiscoveryAdapter, *bytes.Buffer) {
	t.Helper()

	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test Project\nA test project."), 0o644))

	handler := application.NewDiscoveryHandler(&testPublisher{})
	renderer := infrastructure.NewJSONSessionRenderer()
	var buf bytes.Buffer
	adapter := infrastructure.NewAgentDiscoveryAdapter(handler, renderer, &buf, tmpDir)

	return adapter, &buf
}

// parseOutputLines splits the buffer into non-empty lines.
func parseOutputLines(buf *bytes.Buffer) []string {
	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

// envelope mirrors the JSONL envelope structure for test parsing.
type testEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func TestAgentDiscoveryAdapter_Run_WhenReadmeExists_EmitsSessionStatus(t *testing.T) {
	t.Parallel()
	adapter, buf := setupAgentAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	require.NotEmpty(t, lines)

	var env testEnvelope
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &env))
	assert.Equal(t, "session_status", env.Type)

	var status domain.SessionStatusOutput
	require.NoError(t, json.Unmarshal(env.Data, &status))
	assert.Equal(t, "created", status.Status)
	assert.Equal(t, 10, status.QuestionsTotal)
	assert.Equal(t, "persona_detection", status.NextAction)
}

func TestAgentDiscoveryAdapter_Run_WhenReadmeExists_EmitsPersonaPrompt(t *testing.T) {
	t.Parallel()
	adapter, buf := setupAgentAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	require.GreaterOrEqual(t, len(lines), 2)

	var env testEnvelope
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &env))
	assert.Equal(t, "persona_prompt", env.Type)

	var prompt domain.PersonaPrompt
	require.NoError(t, json.Unmarshal(env.Data, &prompt))
	assert.Equal(t, "Which best describes you?", prompt.Prompt)
	assert.Len(t, prompt.Choices, 4)
}

func TestAgentDiscoveryAdapter_Run_WhenReadmeExists_EmitsAllQuestions(t *testing.T) {
	t.Parallel()
	adapter, buf := setupAgentAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	catalog := domain.QuestionCatalog()

	// Questions are at lines[2] through lines[11] (10 questions)
	for i, q := range catalog {
		lineIdx := i + 2 // offset past session_status and persona_prompt
		require.Less(t, lineIdx, len(lines), "expected line %d for question %s", lineIdx, q.ID())

		var env testEnvelope
		require.NoError(t, json.Unmarshal([]byte(lines[lineIdx]), &env))
		assert.Equal(t, "question", env.Type)

		var qo domain.QuestionOutput
		require.NoError(t, json.Unmarshal(env.Data, &qo))
		assert.Equal(t, q.ID(), qo.QuestionID)
		assert.NotEmpty(t, qo.QuestionText)
	}
}

func TestAgentDiscoveryAdapter_Run_WhenReadmeExists_EmitsFinalStatus(t *testing.T) {
	t.Parallel()
	adapter, buf := setupAgentAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	lastLine := lines[len(lines)-1]

	var env testEnvelope
	require.NoError(t, json.Unmarshal([]byte(lastLine), &env))
	assert.Equal(t, "session_status", env.Type)

	var status domain.SessionStatusOutput
	require.NoError(t, json.Unmarshal(env.Data, &status))
	assert.Equal(t, "created", status.Status)
}

func TestAgentDiscoveryAdapter_Run_WhenReadmeExists_EachLineIsValidJSON(t *testing.T) {
	t.Parallel()
	adapter, buf := setupAgentAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	for i, line := range lines {
		assert.True(t, json.Valid([]byte(line)), "line %d is not valid JSON: %s", i, line)
	}
}

func TestAgentDiscoveryAdapter_Run_WhenNoReadme_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir() // no README.md
	handler := application.NewDiscoveryHandler(&testPublisher{})
	renderer := infrastructure.NewJSONSessionRenderer()
	var buf bytes.Buffer
	adapter := infrastructure.NewAgentDiscoveryAdapter(handler, renderer, &buf, tmpDir)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "README.md")
}

func TestAgentDiscoveryAdapter_Run_WhenReadmeExists_TotalLineCount(t *testing.T) {
	t.Parallel()
	adapter, buf := setupAgentAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	// 1 status + 1 persona + 10 questions + 1 final status = 13
	assert.Len(t, lines, 13)
}
