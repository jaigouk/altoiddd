package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	"github.com/alto-cli/alto/internal/shared/infrastructure/eventbus"
)

// setupIngestTest creates a temp dir with a persisted session and wired handler.
// Returns the App, session ID, temp dir, and cleanup function.
func setupIngestTest(t *testing.T) (*composition.App, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	altoDir := filepath.Join(tmpDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	sessionRepo := infrastructure.NewFileSystemSessionRepository(altoDir)
	handler := application.NewDiscoveryHandler(publisher, application.WithSessionRepository(sessionRepo))

	// Create and persist a session with persona detected
	session, err := handler.StartSession("test readme content")
	require.NoError(t, err)
	_, err = handler.DetectPersona(session.SessionID(), "1")
	require.NoError(t, err)
	require.NoError(t, sessionRepo.Save(context.Background(), session))

	app := &composition.App{
		DiscoveryHandler: handler,
	}

	return app, session.SessionID(), altoDir
}

// writeJSONL creates a JSONL file with the given envelopes.
func writeJSONL(t *testing.T, dir string, lines []map[string]interface{}) string {
	t.Helper()
	path := filepath.Join(dir, "answers.jsonl")
	var buf bytes.Buffer
	for _, line := range lines {
		data, err := json.Marshal(line)
		require.NoError(t, err)
		buf.Write(data)
		buf.WriteByte('\n')
	}
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))
	return path
}

func makeEnvelope(typ string, data interface{}) map[string]interface{} {
	raw, _ := json.Marshal(data)
	return map[string]interface{}{
		"type": typ,
		"data": json.RawMessage(raw),
	}
}

func allMVPAnswers(sessionID string) []map[string]interface{} {
	// Q1 (actors), Q2 (actors), Q3 (story), Q4 (story), Q5 (story)
	// Q6 (events), Q7 (events), Q8 (events), Q9 (boundaries), Q10 (boundaries)
	// MVP: Q1, Q3, Q4, Q9, Q10
	// Phase order: actors (Q1,Q2) -> story (Q3,Q4,Q5) -> events (Q6,Q7,Q8) -> boundaries (Q9,Q10)
	// Need to answer/skip all earlier phase questions before later phases.
	questions := []struct {
		id     string
		answer string
		skip   bool
		reason string
	}{
		{"Q1", "users and APIs", false, ""},
		{"Q2", "orders, products", false, ""},
		{"Q3", "user places order", false, ""},
		{"Q4", "payment failure", false, ""},
		{"Q5", "", true, "not sure yet"},
		{"Q6", "OrderPlaced, PaymentProcessed", false, ""},
		{"Q7", "", true, "not sure yet"},
		{"Q8", "", true, "not sure yet"},
		{"Q9", "Orders, Payments, Users", false, ""},
		{"Q10", "Orders=core, Payments=supporting, Users=generic", false, ""},
	}

	var envelopes []map[string]interface{}
	for _, q := range questions {
		if q.skip {
			envelopes = append(envelopes, makeEnvelope("answer", domain.AnswerInput{
				SessionID:  sessionID,
				QuestionID: q.id,
				Skipped:    true,
				SkipReason: q.reason,
			}))
		} else {
			envelopes = append(envelopes, makeEnvelope("answer", domain.AnswerInput{
				SessionID:  sessionID,
				QuestionID: q.id,
				Answer:     q.answer,
			}))
		}
	}
	return envelopes
}

func TestGuideAgentIngest_WhenPersonaAndAllAnswers_CompletesSession(t *testing.T) {
	t.Parallel()
	app, sessionID, altoDir := setupIngestTest(t)

	envelopes := allMVPAnswers(sessionID)
	ingestPath := writeJSONL(t, filepath.Dir(altoDir), envelopes)

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.NoError(t, err)

	// Verify final status envelope was emitted
	assert.Contains(t, out.String(), `"type":"session_status"`)
	assert.Contains(t, out.String(), `"completed"`)
}

func TestGuideAgentIngest_WhenSkippedAnswer_CallsSkipQuestion(t *testing.T) {
	t.Parallel()
	app, sessionID, altoDir := setupIngestTest(t)

	envelopes := []map[string]interface{}{
		makeEnvelope("answer", domain.AnswerInput{
			SessionID:  sessionID,
			QuestionID: "Q1",
			Skipped:    true,
			SkipReason: "will answer later",
		}),
	}
	ingestPath := writeJSONL(t, filepath.Dir(altoDir), envelopes)

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.NoError(t, err)

	// Session should show Q1 as skipped
	session, err := app.DiscoveryHandler.LoadOrGetSession(sessionID)
	require.NoError(t, err)
	assert.Equal(t, "will answer later", session.SkipReason("Q1"))
}

func TestGuideAgentIngest_WhenPlaybackPending_AutoConfirms(t *testing.T) {
	t.Parallel()
	app, sessionID, altoDir := setupIngestTest(t)

	// 3 answers trigger playback — auto-confirm should handle it
	envelopes := []map[string]interface{}{
		makeEnvelope("answer", domain.AnswerInput{SessionID: sessionID, QuestionID: "Q1", Answer: "users"}),
		makeEnvelope("answer", domain.AnswerInput{SessionID: sessionID, QuestionID: "Q2", Answer: "orders"}),
		makeEnvelope("answer", domain.AnswerInput{SessionID: sessionID, QuestionID: "Q3", Answer: "user places order"}),
		// After Q3, playback should trigger and be auto-confirmed
		makeEnvelope("answer", domain.AnswerInput{SessionID: sessionID, QuestionID: "Q4", Answer: "payment failure"}),
	}
	ingestPath := writeJSONL(t, filepath.Dir(altoDir), envelopes)

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.NoError(t, err)

	// Verify Q4 was processed (means playback was auto-confirmed after Q3)
	session, err := app.DiscoveryHandler.LoadOrGetSession(sessionID)
	require.NoError(t, err)
	assert.Len(t, session.Answers(), 4)
	assert.Len(t, session.PlaybackConfirmations(), 1)
}

func TestGuideAgentIngest_WhenMismatchedSessionID_ReturnsError(t *testing.T) {
	t.Parallel()
	app, _, altoDir := setupIngestTest(t)

	envelopes := []map[string]interface{}{
		makeEnvelope("answer", domain.AnswerInput{
			SessionID:  "wrong-id",
			QuestionID: "Q1",
			Answer:     "users",
		}),
	}
	ingestPath := writeJSONL(t, filepath.Dir(altoDir), envelopes)

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session ID mismatch")
}

func TestGuideAgentIngest_WhenNoSessionFile_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)

	emptyDir := t.TempDir()
	altoDir := filepath.Join(emptyDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))

	sessionRepo := infrastructure.NewFileSystemSessionRepository(altoDir)
	handler := application.NewDiscoveryHandler(publisher, application.WithSessionRepository(sessionRepo))
	app := &composition.App{DiscoveryHandler: handler}

	ingestPath := writeJSONL(t, emptyDir, []map[string]interface{}{})

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no session found")
}

func TestGuideAgentIngest_WhenStdinDash_ReadsStdin(t *testing.T) {
	t.Parallel()
	app, sessionID, altoDir := setupIngestTest(t)

	// Create JSONL content for stdin
	envelope := makeEnvelope("answer", domain.AnswerInput{
		SessionID:  sessionID,
		QuestionID: "Q1",
		Answer:     "users from stdin",
	})
	data, err := json.Marshal(envelope)
	require.NoError(t, err)

	stdinReader := bytes.NewReader(append(data, '\n'))

	var out bytes.Buffer
	err = runGuideAgentIngestFromReader(context.Background(), app, stdinReader, altoDir, &out)
	require.NoError(t, err)

	session, err := app.DiscoveryHandler.LoadOrGetSession(sessionID)
	require.NoError(t, err)
	assert.Len(t, session.Answers(), 1)
}

func TestGuideAgentIngest_WhenInvalidJSON_ReturnsErrorWithLineNumber(t *testing.T) {
	t.Parallel()
	app, sessionID, altoDir := setupIngestTest(t)

	// Write a file with invalid JSON on line 2
	path := filepath.Join(filepath.Dir(altoDir), "bad.jsonl")
	content := fmt.Sprintf(`{"type":"answer","data":{"session_id":"%s","question_id":"Q1","answer":"ok"}}
not valid json
{"type":"answer","data":{"session_id":"%s","question_id":"Q2","answer":"ok"}}
`, sessionID, sessionID)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, path, altoDir, &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 2")
}

func TestGuideAgentIngest_WhenPartialAnswers_SavesPartialState(t *testing.T) {
	t.Parallel()
	app, sessionID, altoDir := setupIngestTest(t)

	// Only answer Q1 and Q2 — not enough for Complete
	envelopes := []map[string]interface{}{
		makeEnvelope("answer", domain.AnswerInput{SessionID: sessionID, QuestionID: "Q1", Answer: "users"}),
		makeEnvelope("answer", domain.AnswerInput{SessionID: sessionID, QuestionID: "Q2", Answer: "orders"}),
	}
	ingestPath := writeJSONL(t, filepath.Dir(altoDir), envelopes)

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.NoError(t, err)

	// Session should be in answering state with 2 answers
	session, err := app.DiscoveryHandler.LoadOrGetSession(sessionID)
	require.NoError(t, err)
	assert.Len(t, session.Answers(), 2)
	assert.NotEqual(t, domain.StatusCompleted, session.Status())

	// Final status should be emitted
	assert.Contains(t, out.String(), `"type":"session_status"`)
}

func TestGuideAgentIngest_WhenUnknownEnvelopeType_ReturnsError(t *testing.T) {
	t.Parallel()
	app, sessionID, altoDir := setupIngestTest(t)

	envelopes := []map[string]interface{}{
		makeEnvelope("unknown_type", map[string]string{"session_id": sessionID}),
	}
	ingestPath := writeJSONL(t, filepath.Dir(altoDir), envelopes)

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 1")
	assert.Contains(t, err.Error(), "unknown response type")
}

func TestGuideAgentIngest_WhenIngestWithoutAgent_ReturnsError(t *testing.T) {
	t.Parallel()

	app := &composition.App{}
	cmd := NewGuideCmd(app)
	cmd.SetArgs([]string{"--ingest", "answers.jsonl"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--ingest requires --agent")
}

func TestGuideAgentIngest_WhenPersonaResponse_DetectsPersona(t *testing.T) {
	t.Parallel()

	// Create a session that hasn't had persona detected yet
	tmpDir := t.TempDir()
	altoDir := filepath.Join(tmpDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))

	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	sessionRepo := infrastructure.NewFileSystemSessionRepository(altoDir)
	handler := application.NewDiscoveryHandler(publisher, application.WithSessionRepository(sessionRepo))

	// Start session but do NOT detect persona
	session, err := handler.StartSession("test readme")
	require.NoError(t, err)
	require.NoError(t, sessionRepo.Save(context.Background(), session))

	app := &composition.App{DiscoveryHandler: handler}
	sessionID := session.SessionID()

	envelopes := []map[string]interface{}{
		makeEnvelope("persona_response", domain.PersonaResponse{SessionID: sessionID, Choice: "2"}),
		makeEnvelope("answer", domain.AnswerInput{SessionID: sessionID, QuestionID: "Q1", Answer: "users"}),
	}
	ingestPath := writeJSONL(t, tmpDir, envelopes)

	var out bytes.Buffer
	err = runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.NoError(t, err)

	session, err = handler.LoadOrGetSession(sessionID)
	require.NoError(t, err)
	persona, ok := session.Persona()
	require.True(t, ok)
	assert.Equal(t, domain.PersonaProductOwner, persona)
}

func TestGuideAgentIngest_WhenEmptyFile_EmitsCurrentStatus(t *testing.T) {
	t.Parallel()
	app, _, altoDir := setupIngestTest(t)

	ingestPath := writeJSONL(t, filepath.Dir(altoDir), []map[string]interface{}{})

	var out bytes.Buffer
	err := runGuideAgentIngest(context.Background(), app, ingestPath, altoDir, &out)
	require.NoError(t, err)

	// Should still emit final status
	assert.Contains(t, out.String(), `"type":"session_status"`)
}
