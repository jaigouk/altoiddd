package mcp_test

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	mcptools "github.com/alto-cli/alto/internal/mcp"
)

// stubDiscoveryPublisher implements sharedapp.EventPublisher for testing.
type stubDiscoveryPublisher struct{}

func (s *stubDiscoveryPublisher) Publish(_ context.Context, _ any) error { return nil }

// --- Test helpers ---

// setupDiscoveryServer creates a test MCP server with discovery tools registered.
func setupDiscoveryServer(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	handler := discoveryapp.NewDiscoveryHandler(&stubDiscoveryPublisher{})
	app := &composition.App{DiscoveryHandler: handler}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	mcptools.RegisterDiscoveryTools(server, app)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	go func() { _ = server.Run(ctx, st) }()

	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

// callTool is a helper that calls a tool and returns the text content.
func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) (string, bool) {
	t.Helper()
	ctx := context.Background()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	require.NoError(t, err)
	if result.IsError {
		text, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		return text.Text, true
	}
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	return text.Text, false
}

// startSession is a helper that starts a discovery session and returns the session_id.
func startSession(t *testing.T, session *mcp.ClientSession) string {
	t.Helper()
	text, isErr := callTool(t, session, "guide_start", map[string]any{
		"readme_content": "Test project README",
	})
	require.False(t, isErr, "guide_start should not error: %s", text)
	// Extract session_id from response
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "session_id: ") {
			return strings.TrimPrefix(line, "session_id: ")
		}
	}
	t.Fatal("session_id not found in guide_start response")
	return ""
}

// startWithPersona starts a session and detects persona, returning session_id.
func startWithPersona(t *testing.T, session *mcp.ClientSession, choice string) string {
	t.Helper()
	sid := startSession(t, session)
	_, isErr := callTool(t, session, "guide_detect_persona", map[string]any{
		"session_id": sid,
		"choice":     choice,
	})
	require.False(t, isErr, "guide_detect_persona should not error")
	return sid
}

// --- guide_start tests ---

func TestGuideStart_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_start", map[string]any{
		"readme_content": "My cool project",
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Discovery session started")
	assert.Contains(t, text, "session_id:")
	assert.Contains(t, text, "guide_detect_persona")
}

func TestGuideStart_EmptyReadme(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_start", map[string]any{
		"readme_content": "",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "readme_content is required")
}

// --- guide_detect_persona tests ---

func TestGuideDetectPersona_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startSession(t, session)

	text, isErr := callTool(t, session, "guide_detect_persona", map[string]any{
		"session_id": sid,
		"choice":     "1",
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Persona detected: developer")
	assert.Contains(t, text, "register: technical")
}

func TestGuideDetectPersona_InvalidSession(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_detect_persona", map[string]any{
		"session_id": "nonexistent",
		"choice":     "1",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "no active discovery session")
}

func TestGuideDetectPersona_InvalidChoice(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startSession(t, session)

	text, isErr := callTool(t, session, "guide_detect_persona", map[string]any{
		"session_id": sid,
		"choice":     "invalid",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "invalid persona choice")
}

// --- guide_answer tests ---

func TestGuideAnswer_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1") // developer → technical register

	text, isErr := callTool(t, session, "guide_answer", map[string]any{
		"session_id":  sid,
		"question_id": "Q1",
		"answer":      "Users and admin systems",
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Recorded answer for Q1")
	// Should show next question (Q2) with technical text
	assert.Contains(t, text, "Q2")
}

func TestGuideAnswer_InvalidSession(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_answer", map[string]any{
		"session_id":  "nonexistent",
		"question_id": "Q1",
		"answer":      "test",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "no active discovery session")
}

func TestGuideAnswer_InvalidQuestionID(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	text, isErr := callTool(t, session, "guide_answer", map[string]any{
		"session_id":  sid,
		"question_id": "Q99",
		"answer":      "test",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "unknown question")
}

// --- guide_skip_question tests ---

func TestGuideSkipQuestion_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "2") // product owner → non-technical

	text, isErr := callTool(t, session, "guide_skip_question", map[string]any{
		"session_id":  sid,
		"question_id": "Q1",
		"reason":      "Not relevant to our domain",
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Skipped Q1")
	assert.Contains(t, text, "Not relevant to our domain")
}

func TestGuideSkipQuestion_EmptyReason(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	text, isErr := callTool(t, session, "guide_skip_question", map[string]any{
		"session_id":  sid,
		"question_id": "Q1",
		"reason":      "",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "reason")
}

// --- guide_confirm_playback tests ---

func TestGuideConfirmPlayback_Confirmed(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer 3 questions to trigger playback (actors phase: Q1, Q2; story phase: Q3)
	_, isErr := callTool(t, session, "guide_answer", map[string]any{
		"session_id": sid, "question_id": "Q1", "answer": "Users",
	})
	require.False(t, isErr)
	_, isErr = callTool(t, session, "guide_answer", map[string]any{
		"session_id": sid, "question_id": "Q2", "answer": "Orders and Products",
	})
	require.False(t, isErr)
	_, isErr = callTool(t, session, "guide_answer", map[string]any{
		"session_id": sid, "question_id": "Q3", "answer": "User places order",
	})
	require.False(t, isErr)

	// Now confirm playback
	text, isErr := callTool(t, session, "guide_confirm_playback", map[string]any{
		"session_id": sid,
		"confirmed":  true,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Playback confirmed")
}

func TestGuideConfirmPlayback_Rejected(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer 3 questions to trigger playback
	for _, q := range []struct{ id, answer string }{
		{"Q1", "Users"}, {"Q2", "Orders"}, {"Q3", "Place order"},
	} {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)
	}

	text, isErr := callTool(t, session, "guide_confirm_playback", map[string]any{
		"session_id": sid,
		"confirmed":  false,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Playback rejected")
}

func TestGuideConfirmPlayback_WrongState(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// No answers yet, should fail (not in PLAYBACK_PENDING)
	text, isErr := callTool(t, session, "guide_confirm_playback", map[string]any{
		"session_id": sid,
		"confirmed":  true,
	})
	require.True(t, isErr)
	assert.Contains(t, text, "PLAYBACK_PENDING")
}

// --- guide_complete tests ---

func TestGuideComplete_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer all MVP questions (Q1, Q3, Q4, Q9, Q10) + enough to satisfy phase ordering
	// Phase order: actors (Q1, Q2) → story (Q3, Q4, Q5) → events (Q6, Q7, Q8) → boundaries (Q9, Q10)
	// Must complete earlier phases before later ones
	answers := []struct{ id, answer string }{
		{"Q1", "Users and admins"},
		{"Q2", "Orders, Products, Users"},
		{"Q3", "User places an order"},
		{"Q4", "Payment fails, stock depleted"},
		{"Q5", "User views history"},
	}
	for _, q := range answers {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)

		// Check if we hit playback
		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, isErr = callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, isErr)
		}
	}

	// Skip non-MVP questions in events phase
	for _, qid := range []string{"Q6", "Q7", "Q8"} {
		_, isErr := callTool(t, session, "guide_skip_question", map[string]any{
			"session_id": sid, "question_id": qid, "reason": "not needed for MVP",
		})
		require.False(t, isErr)
	}

	// Answer remaining MVP questions in boundaries phase
	for _, q := range []struct{ id, answer string }{
		{"Q9", "Orders context, Users context"},
		{"Q10", "Orders=core, Users=supporting"},
	} {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, isErr = callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, isErr)
		}
	}

	text, isErr := callTool(t, session, "guide_complete", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Discovery session completed")
	assert.Contains(t, text, "generate_artifacts")
}

func TestGuideComplete_AlreadyCompleted(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer all questions with playback handling
	allQs := []struct{ id, answer string }{
		{"Q1", "a"}, {"Q2", "b"}, {"Q3", "c"}, {"Q4", "d"}, {"Q5", "e"},
	}
	for _, q := range allQs {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, isErr = callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, isErr)
		}
	}

	for _, qid := range []string{"Q6", "Q7", "Q8"} {
		_, isErr := callTool(t, session, "guide_skip_question", map[string]any{
			"session_id": sid, "question_id": qid, "reason": "skip",
		})
		require.False(t, isErr)
	}

	for _, q := range []struct{ id, answer string }{
		{"Q9", "ctx"}, {"Q10", "core"},
	} {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, isErr = callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, isErr)
		}
	}

	// First complete should succeed
	_, isErr := callTool(t, session, "guide_complete", map[string]any{"session_id": sid})
	require.False(t, isErr)

	// Second complete should fail
	text, isErr := callTool(t, session, "guide_complete", map[string]any{"session_id": sid})
	require.True(t, isErr)
	assert.Contains(t, text, "complete")
}

// --- guide_status tests ---

func TestGuideStatus_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer one question
	_, isErr := callTool(t, session, "guide_answer", map[string]any{
		"session_id": sid, "question_id": "Q1", "answer": "Users",
	})
	require.False(t, isErr)

	text, isErr := callTool(t, session, "guide_status", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Session:")
	assert.Contains(t, text, "Status: answering")
	assert.Contains(t, text, "Persona: developer")
	assert.Contains(t, text, "Register: technical")
	assert.Contains(t, text, "Answered: 1")
	assert.Contains(t, text, "Q1")
}

func TestGuideStatus_SessionNotFound(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_status", map[string]any{
		"session_id": "nonexistent",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "no active discovery session")
}

// --- formatNextQuestion tests ---

func TestFormatNextQuestion_Answering(t *testing.T) {
	t.Parallel()
	// After persona detection, should show first question
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1") // developer → technical

	text, isErr := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
	require.False(t, isErr)
	// Should contain next question info - Q1 is the first unanswered
	assert.Contains(t, text, "Q1")
	assert.Contains(t, text, "actors")
}

func TestFormatNextQuestion_PlaybackPending(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer 3 to trigger playback
	for _, q := range []struct{ id, answer string }{
		{"Q1", "Users"}, {"Q2", "Orders"}, {"Q3", "Place order"},
	} {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)
	}

	text, isErr := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
	require.False(t, isErr)
	assert.Contains(t, text, "PLAYBACK_PENDING")
	assert.Contains(t, text, "guide_confirm_playback")
}

func TestFormatNextQuestion_Completed(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer all, handle playbacks, complete
	allQs := []struct{ id, answer string }{
		{"Q1", "a"}, {"Q2", "b"}, {"Q3", "c"}, {"Q4", "d"}, {"Q5", "e"},
	}
	for _, q := range allQs {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, isErr = callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, isErr)
		}
	}
	for _, qid := range []string{"Q6", "Q7", "Q8"} {
		_, _ = callTool(t, session, "guide_skip_question", map[string]any{
			"session_id": sid, "question_id": qid, "reason": "skip",
		})
	}
	for _, q := range []struct{ id, answer string }{
		{"Q9", "ctx"}, {"Q10", "core"},
	} {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)
		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, isErr = callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, isErr)
		}
	}
	_, _ = callTool(t, session, "guide_complete", map[string]any{"session_id": sid})

	text, isErr := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
	require.False(t, isErr)
	assert.Contains(t, text, "COMPLETED")
}

func TestFormatNextQuestion_AllAnswered(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := startWithPersona(t, session, "1")

	// Answer all 10 questions with playback handling
	allQs := []struct{ id, answer string }{
		{"Q1", "a"},
		{"Q2", "b"},
		{"Q3", "c"},
		{"Q4", "d"},
		{"Q5", "e"},
		{"Q6", "f"},
		{"Q7", "g"},
		{"Q8", "h"},
		{"Q9", "i"},
		{"Q10", "j"},
	}
	for _, q := range allQs {
		_, isErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, isErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, isErr = callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, isErr)
		}
	}

	// Status should say all answered, use guide_complete
	text, isErr := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
	require.False(t, isErr)
	assert.Contains(t, text, "all questions answered")
	assert.Contains(t, text, "guide_complete")
}

func TestFormatNextQuestion_TechnicalVsNonTechnical(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	// Developer (choice "1") → technical register
	sidTech := startWithPersona(t, session, "1")
	techText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sidTech})

	// Product Owner (choice "2") → non-technical register
	sidNonTech := startWithPersona(t, session, "2")
	nonTechText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sidNonTech})

	// Q1 technical: "Who are the actors..."
	// Q1 non-technical: "Who will use this product..."
	catalog := domain.QuestionCatalog()
	assert.Contains(t, techText, catalog[0].TechnicalText())
	assert.Contains(t, nonTechText, catalog[0].NonTechnicalText())
}

// --- Empty session_id tests ---

func TestGuideDetectPersona_EmptySessionID(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_detect_persona", map[string]any{
		"session_id": "",
		"choice":     "1",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session_id is required")
}

func TestGuideAnswer_EmptySessionID(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_answer", map[string]any{
		"session_id":  "",
		"question_id": "Q1",
		"answer":      "test",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session_id is required")
}

// --- End-to-end flow test ---

func TestFullDiscoveryFlow_EndToEnd(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	ctx := context.Background()

	// 1. List tools — should have 7 guide_* tools
	toolList, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	guideTools := 0
	for _, tool := range toolList.Tools {
		if strings.HasPrefix(tool.Name, "guide_") {
			guideTools++
		}
	}
	assert.Equal(t, 8, guideTools)

	// 2. Start session
	sid := startSession(t, session)
	require.NotEmpty(t, sid)

	// 3. Detect persona
	text, isErr := callTool(t, session, "guide_detect_persona", map[string]any{
		"session_id": sid, "choice": "2", // product owner
	})
	require.False(t, isErr)
	assert.Contains(t, text, "product_owner")

	// 4. Answer questions with playback handling
	answers := []struct{ id, answer string }{
		{"Q1", "End users and admin team"},
		{"Q2", "Invoices and customers"},
		{"Q3", "Customer submits invoice"},
		{"Q4", "Payment processing failure"},
		{"Q5", "Customer views invoice history"},
	}
	for _, q := range answers {
		_, ansErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, ansErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, pbErr := callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, pbErr)
		}
	}

	// 5. Skip events phase questions
	for _, qid := range []string{"Q6", "Q7", "Q8"} {
		_, skipErr := callTool(t, session, "guide_skip_question", map[string]any{
			"session_id": sid, "question_id": qid, "reason": "Deferring to implementation",
		})
		require.False(t, skipErr)
	}

	// 6. Answer boundaries phase
	for _, q := range []struct{ id, answer string }{
		{"Q9", "Billing and Customer Management"},
		{"Q10", "Billing=core, Customer=supporting"},
	} {
		_, ansErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, ansErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, pbErr := callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, pbErr)
		}
	}

	// 7. Complete
	text, isErr = callTool(t, session, "guide_complete", map[string]any{"session_id": sid})
	require.False(t, isErr)
	assert.Contains(t, text, "Discovery session completed")
	assert.Contains(t, text, "generate_artifacts")

	// 8. Verify final status
	text, isErr = callTool(t, session, "guide_status", map[string]any{"session_id": sid})
	require.False(t, isErr)
	assert.Contains(t, text, "COMPLETED")
	assert.Contains(t, text, "Answered: 7")
}

// --- guide_classify_subdomain tests ---

func TestGuideClassifySubdomain_Core(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := completeSession(t, session)

	// Core: buy=no, complex=yes, competitor=yes
	text, isErr := callTool(t, session, "guide_classify_subdomain", map[string]any{
		"session_id":        sid,
		"context_name":      "Billing",
		"buy_yes":           false,
		"complex_rules":     true,
		"competitor_threat": true,
	})
	require.False(t, isErr, "should not error: %s", text)
	assert.Contains(t, text, "Classified 'Billing' as core")
	assert.Contains(t, text, "Rationale:")
}

func TestGuideClassifySubdomain_Supporting(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := completeSession(t, session)

	// Supporting: buy=no, complex=yes, competitor=no
	text, isErr := callTool(t, session, "guide_classify_subdomain", map[string]any{
		"session_id":        sid,
		"context_name":      "Reporting",
		"buy_yes":           false,
		"complex_rules":     true,
		"competitor_threat": false,
	})
	require.False(t, isErr, "should not error: %s", text)
	assert.Contains(t, text, "Classified 'Reporting' as supporting")
}

func TestGuideClassifySubdomain_Generic(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := completeSession(t, session)

	// Generic: buy=yes
	text, isErr := callTool(t, session, "guide_classify_subdomain", map[string]any{
		"session_id":        sid,
		"context_name":      "Notifications",
		"buy_yes":           true,
		"complex_rules":     false,
		"competitor_threat": false,
	})
	require.False(t, isErr, "should not error: %s", text)
	assert.Contains(t, text, "Classified 'Notifications' as generic")
}

func TestGuideClassifySubdomain_EmptySessionID(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_classify_subdomain", map[string]any{
		"session_id":        "",
		"context_name":      "Test",
		"buy_yes":           false,
		"complex_rules":     false,
		"competitor_threat": false,
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session_id is required")
}

func TestGuideClassifySubdomain_EmptyContextName(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)
	sid := completeSession(t, session)

	text, isErr := callTool(t, session, "guide_classify_subdomain", map[string]any{
		"session_id":        sid,
		"context_name":      "",
		"buy_yes":           false,
		"complex_rules":     false,
		"competitor_threat": false,
	})
	require.True(t, isErr)
	assert.Contains(t, text, "context_name is required")
}

func TestGuideClassifySubdomain_InvalidSession(t *testing.T) {
	t.Parallel()
	session := setupDiscoveryServer(t)

	text, isErr := callTool(t, session, "guide_classify_subdomain", map[string]any{
		"session_id":        "nonexistent-session-id",
		"context_name":      "Test",
		"buy_yes":           false,
		"complex_rules":     false,
		"competitor_threat": false,
	})
	require.True(t, isErr)
	assert.Contains(t, text, "no active discovery session")
}

// completeSession is a helper that runs through a full discovery session and returns session_id.
func completeSession(t *testing.T, session *mcp.ClientSession) string {
	t.Helper()
	sid := startWithPersona(t, session, "2") // product owner

	// Answer round 1 questions (core phase)
	for _, q := range []struct{ id, answer string }{
		{"Q1", "End users and admin team"},
		{"Q2", "Invoices and customers"},
		{"Q3", "Customer submits invoice"},
		{"Q4", "Payment processing failure"},
		{"Q5", "Customer views invoice history"},
	} {
		_, ansErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, ansErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, pbErr := callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, pbErr)
		}
	}

	// Skip events phase questions (Q6-Q8)
	for _, qid := range []string{"Q6", "Q7", "Q8"} {
		_, skipErr := callTool(t, session, "guide_skip_question", map[string]any{
			"session_id": sid, "question_id": qid, "reason": "Deferring to implementation",
		})
		require.False(t, skipErr)
	}

	// Answer boundaries phase (Q9-Q10)
	for _, q := range []struct{ id, answer string }{
		{"Q9", "Billing and Customer Management"},
		{"Q10", "Billing=core, Customer=supporting"},
	} {
		_, ansErr := callTool(t, session, "guide_answer", map[string]any{
			"session_id": sid, "question_id": q.id, "answer": q.answer,
		})
		require.False(t, ansErr)

		statusText, _ := callTool(t, session, "guide_status", map[string]any{"session_id": sid})
		if strings.Contains(statusText, "playback_pending") {
			_, pbErr := callTool(t, session, "guide_confirm_playback", map[string]any{
				"session_id": sid, "confirmed": true,
			})
			require.False(t, pbErr)
		}
	}

	// Complete the session
	_, completeErr := callTool(t, session, "guide_complete", map[string]any{"session_id": sid})
	require.False(t, completeErr)

	return sid
}
