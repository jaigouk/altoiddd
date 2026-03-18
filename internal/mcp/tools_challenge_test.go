package mcp_test

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	challengeapp "github.com/alto-cli/alto/internal/challenge/application"
	challengedomain "github.com/alto-cli/alto/internal/challenge/domain"
	challengeinfra "github.com/alto-cli/alto/internal/challenge/infrastructure"
	"github.com/alto-cli/alto/internal/composition"
	mcptools "github.com/alto-cli/alto/internal/mcp"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
)

// --- Test mock ---

// mockChallenger always generates 2 challenges for testing.
type mockChallenger struct{}

func (m *mockChallenger) GenerateChallenges(_ context.Context, _ *ddd.DomainModel, _ int) ([]challengedomain.Challenge, error) {
	c1, _ := challengedomain.NewChallenge(
		challengedomain.ChallengeLanguage,
		"What does 'Order' mean in different contexts?",
		"Sales",
		"UL glossary: Order",
		"",
	)
	c2, _ := challengedomain.NewChallenge(
		challengedomain.ChallengeInvariant,
		"What invariants protect the Order aggregate?",
		"Sales",
		"Aggregate design: Order",
		"",
	)
	return []challengedomain.Challenge{c1, c2}, nil
}

// mockFileReader for testing versioning.
type mockFileReader struct {
	content map[string]string
}

func (m *mockFileReader) ReadFile(_ context.Context, path string) (string, error) {
	if content, ok := m.content[path]; ok {
		return content, nil
	}
	return "", nil
}

// mockFileWriter for testing versioning.
type mockFileWriter struct {
	written map[string]string
}

func (m *mockFileWriter) WriteFile(_ context.Context, path, content string) error {
	if m.written == nil {
		m.written = make(map[string]string)
	}
	m.written[path] = content
	return nil
}

// --- Test helpers ---

// setupChallengeServer creates a test MCP server with challenge tools registered.
func setupChallengeServer(t *testing.T) *mcp.ClientSession {
	t.Helper()
	return setupChallengeServerWithMocks(t, nil, nil)
}

// setupChallengeServerWithMocks creates a test MCP server with optional mock file IO.
func setupChallengeServerWithMocks(t *testing.T, reader *mockFileReader, writer *mockFileWriter) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	challenger := &mockChallenger{}
	challengeHandler := challengeapp.NewChallengeHandler(challenger)

	var versionHandler *challengeapp.VersionHandler
	if reader != nil && writer != nil {
		parser := challengeinfra.NewYAMLFrontmatterParser()
		versionHandler = challengeapp.NewVersionHandler(reader, writer, parser)
	}

	app := &composition.App{
		ChallengeHandler: challengeHandler,
		VersionHandler:   versionHandler,
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	mcptools.RegisterChallengeTools(server, app)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	go func() { _ = server.Run(ctx, st) }()

	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

// callChallengeTool is a helper that calls a tool and returns the text content.
func callChallengeTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) (string, bool) {
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

// startChallengeSession is a helper that starts a challenge session and returns the session_id.
func startChallengeSession(t *testing.T, session *mcp.ClientSession) string {
	t.Helper()
	text, isErr := callChallengeTool(t, session, "challenge_start", map[string]any{
		"max_per_type": 2,
	})
	require.False(t, isErr, "challenge_start should not error: %s", text)
	// Extract session_id from response
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "session_id: ") {
			return strings.TrimPrefix(line, "session_id: ")
		}
	}
	t.Fatal("session_id not found in challenge_start response")
	return ""
}

// --- challenge_start tests ---

func TestChallengeStart_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	text, isErr := callChallengeTool(t, session, "challenge_start", map[string]any{
		"max_per_type": 3,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Challenge session started")
	assert.Contains(t, text, "session_id:")
	assert.Contains(t, text, "challenges:")
}

func TestChallengeStart_DefaultMaxPerType(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	// Should work without max_per_type (defaults to sensible value)
	text, isErr := callChallengeTool(t, session, "challenge_start", map[string]any{})
	require.False(t, isErr)
	assert.Contains(t, text, "Challenge session started")
}

func TestChallengeStart_ZeroMaxPerType(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	// Zero should still work (may produce no challenges)
	text, isErr := callChallengeTool(t, session, "challenge_start", map[string]any{
		"max_per_type": 0,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "session_id:")
}

// --- challenge_respond tests ---

func TestChallengeRespond_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "c0",
		"accepted":      true,
		"user_response": "Good point, we should add this invariant",
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Response recorded")
	assert.Contains(t, text, "c0")
}

func TestChallengeRespond_WithArtifactUpdates(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":       sid,
		"challenge_id":     "c0",
		"accepted":         true,
		"user_response":    "Adding invariant to Order aggregate",
		"artifact_updates": []string{"Add invariant: order total > 0"},
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Response recorded")
}

func TestChallengeRespond_Rejected(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "c0",
		"accepted":      false,
		"user_response": "This is already handled by validation",
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Response recorded")
}

func TestChallengeRespond_EmptySessionID(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    "",
		"challenge_id":  "c0",
		"accepted":      true,
		"user_response": "test",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session_id is required")
}

func TestChallengeRespond_UnknownSession(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    "nonexistent-session",
		"challenge_id":  "c0",
		"accepted":      true,
		"user_response": "test",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session not found")
}

func TestChallengeRespond_UnknownChallenge(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "c999",
		"accepted":      true,
		"user_response": "test",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "challenge not found")
}

func TestChallengeRespond_DuplicateResponse(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	// First response should succeed
	_, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "c0",
		"accepted":      true,
		"user_response": "First response",
	})
	require.False(t, isErr)

	// Second response to same challenge should fail
	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "c0",
		"accepted":      false,
		"user_response": "Second response",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "already answered")
}

func TestChallengeRespond_EmptyChallengeID(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	text, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "",
		"accepted":      true,
		"user_response": "test",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "challenge_id is required")
}

// --- challenge_status tests ---

func TestChallengeStatus_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	text, isErr := callChallengeTool(t, session, "challenge_status", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Session:")
	assert.Contains(t, text, "Status:")
	assert.Contains(t, text, "Challenges:")
	assert.Contains(t, text, "Responses:")
}

func TestChallengeStatus_AfterResponse(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	// Respond to first challenge
	_, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "c0",
		"accepted":      true,
		"user_response": "Good point",
	})
	require.False(t, isErr)

	text, isErr := callChallengeTool(t, session, "challenge_status", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Responses: 1")
}

func TestChallengeStatus_EmptySessionID(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	text, isErr := callChallengeTool(t, session, "challenge_status", map[string]any{
		"session_id": "",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session_id is required")
}

func TestChallengeStatus_UnknownSession(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	text, isErr := callChallengeTool(t, session, "challenge_status", map[string]any{
		"session_id": "nonexistent",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session not found")
}

// --- challenge_complete tests ---

func TestChallengeComplete_HappyPath(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	// Respond to at least one challenge
	_, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":       sid,
		"challenge_id":     "c0",
		"accepted":         true,
		"user_response":    "Adding invariant",
		"artifact_updates": []string{"Add invariant"},
	})
	require.False(t, isErr)

	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Challenge session completed")
	assert.Contains(t, text, "convergence_delta:")
}

func TestChallengeComplete_NoResponses(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	// Complete without any responses - should work, delta = 0
	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "convergence_delta: 0")
}

func TestChallengeComplete_AllRejected(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	// Respond rejecting the challenge
	_, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid,
		"challenge_id":  "c0",
		"accepted":      false,
		"user_response": "Not applicable",
	})
	require.False(t, isErr)

	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	// Delta should be 0 since rejection doesn't add updates
	assert.Contains(t, text, "convergence_delta: 0")
}

func TestChallengeComplete_EmptySessionID(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": "",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session_id is required")
}

func TestChallengeComplete_UnknownSession(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": "nonexistent",
	})
	require.True(t, isErr)
	assert.Contains(t, text, "session not found")
}

// --- End-to-end flow test ---

func TestFullChallengeFlow_EndToEnd(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)
	ctx := context.Background()

	// 1. List tools — should have 4 challenge_* tools
	toolList, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	challengeTools := 0
	for _, tool := range toolList.Tools {
		if strings.HasPrefix(tool.Name, "challenge_") {
			challengeTools++
		}
	}
	assert.Equal(t, 4, challengeTools)

	// 2. Start session
	sid := startChallengeSession(t, session)
	require.NotEmpty(t, sid)

	// 3. Check status
	statusText, isErr := callChallengeTool(t, session, "challenge_status", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, statusText, "active")

	// 4. Respond to challenges
	_, isErr = callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":       sid,
		"challenge_id":     "c0",
		"accepted":         true,
		"user_response":    "Adding missing invariant",
		"artifact_updates": []string{"Add invariant: OrderTotal > 0"},
	})
	require.False(t, isErr)

	// 5. Complete session
	completeText, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, completeText, "Challenge session completed")
	assert.Contains(t, completeText, "convergence_delta: 1")
}

// --- Multiple sessions test ---

func TestChallenge_MultipleSessions(t *testing.T) {
	t.Parallel()
	session := setupChallengeServer(t)

	// Start two sessions
	sid1 := startChallengeSession(t, session)
	sid2 := startChallengeSession(t, session)

	// They should have different IDs
	assert.NotEqual(t, sid1, sid2)

	// Operations on one should not affect the other
	_, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":    sid1,
		"challenge_id":  "c0",
		"accepted":      true,
		"user_response": "Session 1 response",
	})
	require.False(t, isErr)

	// Session 2 should still have no responses
	status2, isErr := callChallengeTool(t, session, "challenge_status", map[string]any{
		"session_id": sid2,
	})
	require.False(t, isErr)
	assert.Contains(t, status2, "Responses: 0")
}

// --- DDD.md versioning tests ---

func TestChallengeComplete_WithVersioning(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": `---
version: 1
round: express
updated: "2026-03-01"
convergence_delta: 0
---

# Domain Model

Content here.
`,
		},
	}
	writer := &mockFileWriter{}

	session := setupChallengeServerWithMocks(t, reader, writer)
	sid := startChallengeSession(t, session)

	// Respond with artifact updates to create convergence delta
	_, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":       sid,
		"challenge_id":     "c0",
		"accepted":         true,
		"user_response":    "Adding invariant",
		"artifact_updates": []string{"Add invariant: OrderTotal > 0"},
	})
	require.False(t, isErr)

	// Complete with ddd_path to trigger versioning
	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": sid,
		"ddd_path":   "docs/DDD.md",
	})
	require.False(t, isErr)
	assert.Contains(t, text, "DDD.md versioned successfully")

	// Verify the file was written with updated version
	require.Contains(t, writer.written, "docs/DDD.md")
	written := writer.written["docs/DDD.md"]
	assert.Contains(t, written, "version: 2")
	assert.Contains(t, written, "round: challenge")
	assert.Contains(t, written, "convergence_delta: 1")
}

func TestChallengeComplete_WithoutVersioning(t *testing.T) {
	t.Parallel()

	// No reader/writer means no VersionHandler
	session := setupChallengeServer(t)
	sid := startChallengeSession(t, session)

	// Respond to a challenge
	_, isErr := callChallengeTool(t, session, "challenge_respond", map[string]any{
		"session_id":       sid,
		"challenge_id":     "c0",
		"accepted":         true,
		"user_response":    "Adding invariant",
		"artifact_updates": []string{"Add invariant"},
	})
	require.False(t, isErr)

	// Complete without ddd_path
	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": sid,
	})
	require.False(t, isErr)
	assert.Contains(t, text, "Challenge session completed")
	assert.Contains(t, text, "DDD model improvements suggested")
	assert.NotContains(t, text, "versioned successfully")
}

func TestChallengeComplete_VersioningWithEmptyPath(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{content: map[string]string{}}
	writer := &mockFileWriter{}

	session := setupChallengeServerWithMocks(t, reader, writer)
	sid := startChallengeSession(t, session)

	// Complete with empty ddd_path (should not attempt versioning)
	text, isErr := callChallengeTool(t, session, "challenge_complete", map[string]any{
		"session_id": sid,
		"ddd_path":   "",
	})
	require.False(t, isErr)
	assert.NotContains(t, text, "versioned")
	assert.Empty(t, writer.written)
}
