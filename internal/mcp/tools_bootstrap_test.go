package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/composition"
	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
	fitnessapp "github.com/alty-cli/alty/internal/fitness/application"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// --- Stub implementations for happy-path tests ---

// stubToolDetector implements discoveryapp.ToolDetector.
type stubToolDetector struct {
	tools     []string
	conflicts []string
}

func (s *stubToolDetector) Detect(_ string) ([]string, error) { return s.tools, nil }
func (s *stubToolDetector) ScanConflicts(_ string) ([]string, error) {
	return s.conflicts, nil
}

// stubGateRunner implements fitnessapp.GateRunner.
type stubGateRunner struct{}

func (s *stubGateRunner) Run(_ context.Context, gate vo.QualityGate) (vo.GateResult, error) {
	return vo.NewGateResult(gate, true, "all good", 42), nil
}

// assertToolText verifies a successful tool result contains expected text.
func assertToolText(t *testing.T, result *mcp.CallToolResult, contains string) {
	t.Helper()
	require.NotNil(t, result)
	require.False(t, result.IsError, "expected success, got tool error")
	require.NotEmpty(t, result.Content)
	tc, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", result.Content[0])
	assert.Contains(t, tc.Text, contains)
}

// setupBootstrapServer creates a test server with bootstrap tools registered.
func setupBootstrapServer(t *testing.T, app *composition.App, store *ModelStore) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	RegisterBootstrapTools(server, app, store)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	ct, st := mcp.NewInMemoryTransports()

	go func() { _ = server.Run(ctx, st) }()

	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	return session
}

func callBootstrapTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	return session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
}

func assertToolError(t *testing.T, result *mcp.CallToolResult, contains string) {
	t.Helper()
	require.NotNil(t, result)
	assert.True(t, result.IsError, "expected tool error")
	if len(result.Content) > 0 {
		tc, ok := result.Content[0].(*mcp.TextContent)
		if ok {
			assert.Contains(t, tc.Text, contains)
		}
	}
}

// --- detect_tools happy path ---

func TestDetectToolsTool_HappyPath(t *testing.T) {
	t.Parallel()
	detector := &stubToolDetector{tools: []string{"claude-code", "cursor"}}
	app := &composition.App{
		DetectionHandler: discoveryapp.NewDetectionHandler(detector),
	}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "detect_tools", map[string]any{
		"project_dir": t.TempDir(),
	})
	require.NoError(t, err)
	assertToolText(t, result, "Detected 2 tool(s)")
	assertToolText(t, result, "claude-code")
	assertToolText(t, result, "cursor")
}

func TestDetectToolsTool_NoToolsFound(t *testing.T) {
	t.Parallel()
	detector := &stubToolDetector{tools: []string{}}
	app := &composition.App{
		DetectionHandler: discoveryapp.NewDetectionHandler(detector),
	}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "detect_tools", map[string]any{
		"project_dir": t.TempDir(),
	})
	require.NoError(t, err)
	assertToolText(t, result, "No AI coding tools detected")
}

// --- check_quality happy path ---

func TestCheckQualityTool_HappyPath(t *testing.T) {
	t.Parallel()
	app := &composition.App{
		QualityGateHandler: fitnessapp.NewQualityGateHandler(&stubGateRunner{}),
	}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "check_quality", map[string]any{
		"project_dir": t.TempDir(),
		"gates":       []any{"lint"},
	})
	require.NoError(t, err)
	assertToolText(t, result, "PASS")
	assertToolText(t, result, "PASSED")
}

// --- detect_tools error tests ---

func TestDetectToolsTool_NoHandler(t *testing.T) {
	t.Parallel()
	app := &composition.App{} // nil DetectionHandler
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "detect_tools", map[string]any{
		"project_dir": t.TempDir(),
	})
	require.NoError(t, err)
	assertToolError(t, result, "detection handler not available")
}

func TestDetectToolsTool_EmptyProjectDir(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "detect_tools", map[string]any{
		"project_dir": "",
	})
	require.NoError(t, err)
	assertToolError(t, result, "project_dir is required")
}

func TestDetectToolsTool_PathTraversal(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "detect_tools", map[string]any{
		"project_dir": "/tmp/../../../etc/passwd",
	})
	require.NoError(t, err)
	assertToolError(t, result, "path traversal")
}

// --- check_quality tests ---

func TestCheckQualityTool_InvalidGate(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "check_quality", map[string]any{
		"project_dir": t.TempDir(),
		"gates":       []any{"nonexistent"},
	})
	require.NoError(t, err)
	assertToolError(t, result, "unknown quality gate")
}

// --- ticket_health tests ---

func TestTicketHealthTool_NoHandler(t *testing.T) {
	t.Parallel()
	// nil TicketHealthHandler should return a tool error, not panic.
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "ticket_health", map[string]any{})
	require.NoError(t, err)
	// Will be a tool error since handler is nil.
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

// --- generate_fitness without model ---

func TestGenerateFitnessTool_NoModelInStore(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "generate_fitness", map[string]any{
		"session_id":  "nonexistent-session",
		"project_dir": t.TempDir(),
	})
	require.NoError(t, err)
	assertToolError(t, result, "no domain model found for session")
}

// --- generate_tickets without model ---

func TestGenerateTicketsTool_NoModelInStore(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "generate_tickets", map[string]any{
		"session_id": "nonexistent-session",
	})
	require.NoError(t, err)
	assertToolError(t, result, "no domain model found for session")
}

// --- generate_configs without model ---

func TestGenerateConfigsTool_NoModelInStore(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "generate_configs", map[string]any{
		"session_id": "nonexistent-session",
		"tools":      []any{"claude-code"},
	})
	require.NoError(t, err)
	assertToolError(t, result, "no domain model found for session")
}

func TestGenerateConfigsTool_UnknownTool(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	// Put a model so we get past the ModelStore check.
	store.Put("sess-1", makeTestModel("m1"), vo.GenericProfile{})
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "generate_configs", map[string]any{
		"session_id": "sess-1",
		"tools":      []any{"invalid-tool"},
	})
	require.NoError(t, err)
	assertToolError(t, result, "unknown tool")
}

// --- empty project_dir tests ---

func TestInitProjectTool_EmptyProjectDir(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "init_project", map[string]any{
		"project_dir": "",
	})
	require.NoError(t, err)
	assertToolError(t, result, "project_dir is required")
}

func TestDocHealthTool_EmptyProjectDir(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "doc_health", map[string]any{
		"project_dir": "",
	})
	require.NoError(t, err)
	assertToolError(t, result, "project_dir is required")
}

// --- spike_follow_up_audit tests ---

func TestSpikeFollowUpAuditTool_EmptySpikeID(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "spike_follow_up_audit", map[string]any{
		"spike_id":    "",
		"project_dir": t.TempDir(),
	})
	require.NoError(t, err)
	assertToolError(t, result, "spike_id is required")
}

// --- generate_artifacts tests ---

func TestGenerateArtifactsTool_EmptySessionID(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "generate_artifacts", map[string]any{
		"session_id":  "",
		"project_dir": t.TempDir(),
	})
	require.NoError(t, err)
	assertToolError(t, result, "session_id is required")
}

// --- doc_review tests ---

func TestDocReviewTool_EmptyDocPath(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "doc_review", map[string]any{
		"project_dir": t.TempDir(),
		"doc_path":    "",
	})
	require.NoError(t, err)
	assertToolError(t, result, "doc_path is required")
}

// --- rescue_project tests ---

func TestRescueProjectTool_EmptyProjectDir(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := callBootstrapTool(t, session, "rescue_project", map[string]any{
		"project_dir": "",
	})
	require.NoError(t, err)
	assertToolError(t, result, "project_dir is required")
}

// --- tool registration count ---

func TestRegisterBootstrapTools_RegistersAll12(t *testing.T) {
	t.Parallel()
	app := &composition.App{}
	store := NewModelStore(30 * time.Minute)
	session := setupBootstrapServer(t, app, store)

	result, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, result.Tools, 12,
		"expected 12 bootstrap tools: init_project, rescue_project, generate_artifacts, "+
			"generate_fitness, generate_tickets, generate_configs, detect_tools, check_quality, "+
			"doc_health, doc_review, ticket_health, spike_follow_up_audit")

	// Verify tool names.
	names := make(map[string]bool)
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	for _, expected := range []string{
		"init_project", "rescue_project",
		"generate_artifacts", "generate_fitness", "generate_tickets", "generate_configs",
		"detect_tools", "check_quality",
		"doc_health", "doc_review",
		"ticket_health", "spike_follow_up_audit",
	} {
		assert.True(t, names[expected], "missing tool: %s", expected)
	}
}
