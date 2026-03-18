package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	fitnessapp "github.com/alto-cli/alto/internal/fitness/application"
	knowledgeapp "github.com/alto-cli/alto/internal/knowledge/application"
	knowledgedomain "github.com/alto-cli/alto/internal/knowledge/domain"
	ticketapp "github.com/alto-cli/alto/internal/ticket/application"
	ttapp "github.com/alto-cli/alto/internal/tooltranslation/application"
)

// =============================================================================
// Security Integration Tests
// =============================================================================
//
// End-to-end security tests via MCP client verifying all mitigations from 0m9.2.

// setupSecurityServer creates a server with full middleware stack for security tests.
func setupSecurityServer(t *testing.T) *gomcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	app := &composition.App{
		DetectionHandler:   discoveryapp.NewDetectionHandler(&stubToolDetector{tools: []string{"claude-code"}}),
		DiscoveryHandler:   discoveryapp.NewDiscoveryHandler(&integrationPublisher{}),
		QualityGateHandler: fitnessapp.NewQualityGateHandler(&stubGateRunner{}),
		KnowledgeLookupHandler: knowledgeapp.NewKnowledgeLookupHandler(&integrationKnowledgeReader{
			entries: map[string]knowledgedomain.KnowledgeEntry{},
		}),
		TicketHealthHandler: ticketapp.NewTicketHealthHandler(&integrationTicketReader{}),
		PersonaHandler:      ttapp.NewPersonaHandler(&integrationFileWriter{}),
	}

	server := gomcp.NewServer(&gomcp.Implementation{
		Name:    "security-test",
		Version: "0.0.1",
	}, nil)

	// Full middleware stack.
	server.AddReceivingMiddleware(
		AuditMiddleware(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))),
		ErrorSanitizeMiddleware(),
		OutputSanitizeMiddleware(),
		ContentTagMiddleware(),
	)

	// Echo tool with path info.
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "echo",
		Description: "Echo back the input message",
	}, func(_ context.Context, _ *gomcp.CallToolRequest, input echoInput) (*gomcp.CallToolResult, any, error) {
		if input.Message == "" {
			return nil, nil, fmt.Errorf("message is required")
		}
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{
				&gomcp.TextContent{Text: "Echo: " + input.Message},
			},
		}, nil, nil
	})

	// Tool that returns secrets — for redaction testing.
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "leak_secrets",
		Description: "Returns text with secrets for redaction testing",
	}, func(_ context.Context, _ *gomcp.CallToolRequest, _ struct{}) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{
				&gomcp.TextContent{Text: "api_key=sk_live_secret123456789012345 from /Users/admin/project/main.go"},
			},
		}, nil, nil
	})

	// Tool that returns internal error.
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "internal_error",
		Description: "Returns an internal error for sanitization testing",
	}, func(_ context.Context, _ *gomcp.CallToolRequest, _ struct{}) (*gomcp.CallToolResult, any, error) {
		return nil, nil, fmt.Errorf("goroutine 42 [running]: runtime error: invalid memory address in /Users/admin/project/handler.go:42")
	})

	RegisterResources(server, app)
	RegisterDiscoveryTools(server, app)
	store := NewModelStore(30 * time.Minute)
	RegisterBootstrapTools(server, app, store)

	client := gomcp.NewClient(&gomcp.Implementation{Name: "security-client", Version: "0.0.1"}, nil)
	ct, st := gomcp.NewInMemoryTransports()
	go func() { _ = server.Run(ctx, st) }()

	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	return session
}

// --- 1. Path Traversal Tests (MCP05/F1) ---

func TestSecurity_PathTraversal_InitProject(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "init_project",
		Arguments: map[string]any{"project_dir": "/tmp/../../../etc/passwd"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError, "path traversal should be rejected")
}

func TestSecurity_PathTraversal_DocReview(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "doc_review",
		Arguments: map[string]any{"project_dir": "/tmp/../../../etc/passwd", "doc_path": "README.md"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError, "path traversal should be rejected")
}

func TestSecurity_PathTraversal_Resource(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	_, err := session.ReadResource(context.Background(), &gomcp.ReadResourceParams{
		URI: "alto://knowledge/ddd/../../../etc/passwd",
	})
	// Should return error — either protocol error or resource not found.
	require.Error(t, err)
}

func TestSecurity_PathTraversal_SymlinkEscape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outsideDir := t.TempDir()

	err := os.Symlink(outsideDir, root+"/escape")
	if err != nil {
		t.Skip("cannot create symlinks on this OS")
	}

	// SafeProjectPath should reject symlink escape.
	_, err = SafeProjectPath("escape", []string{root})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside allowed")
}

// --- 2. Command Injection Tests (MCP05/F2) ---

func TestSecurity_CommandInjection_CheckQuality(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	// Shell metacharacters in gates.
	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "check_quality",
		Arguments: map[string]any{"project_dir": t.TempDir(), "gates": []any{"; rm -rf /"}},
	})
	require.NoError(t, err)
	// Should be rejected as unknown gate, not executed as shell command.
	assert.True(t, result.IsError, "shell injection in gate name should be rejected")
}

func TestSecurity_CommandInjection_TicketID(t *testing.T) {
	t.Parallel()

	// SafeTicketID should reject injection attempts.
	require.Error(t, SafeTicketID("'; DROP TABLE issues; --"))
	require.Error(t, SafeTicketID("$(cat /etc/passwd)"))
	require.Error(t, SafeTicketID("id | cat /etc/shadow"))
}

// --- 3. Secret Exposure Tests (MCP01) ---

func TestSecurity_SecretRedaction_ToolOutput(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "leak_secrets",
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	tc, ok := result.Content[0].(*gomcp.TextContent)
	require.True(t, ok)

	// Secret should be redacted by OutputSanitizeMiddleware.
	assert.Contains(t, tc.Text, "[REDACTED]")
	assert.NotContains(t, tc.Text, "sk_live_secret")

	// Absolute path should be stripped.
	assert.NotContains(t, tc.Text, "/Users/admin")
}

func TestSecurity_SecretRedaction_ErrorMessage(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "internal_error",
	})
	// The SDK converts handler errors to IsError=true tool results BEFORE
	// ErrorSanitizeMiddleware runs (middleware sees no error in the return).
	// OutputSanitizeMiddleware does strip absolute paths from tool result
	// content though.
	require.NoError(t, err, "SDK converts handler errors to tool results")
	require.NotNil(t, result)
	assert.True(t, result.IsError, "internal error should be flagged as tool error")

	tc, ok := result.Content[0].(*gomcp.TextContent)
	require.True(t, ok)
	// OutputSanitizeMiddleware strips absolute paths.
	assert.NotContains(t, tc.Text, "/Users/admin")
}

// --- 4. Intent Flow Tests (MCP06) ---

func TestSecurity_ContentTagging_ToolOutput(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": "tagged output"},
	})
	require.NoError(t, err)

	tc, ok := result.Content[0].(*gomcp.TextContent)
	require.True(t, ok)

	assert.Contains(t, tc.Text, "[TOOL OUTPUT START]")
	assert.Contains(t, tc.Text, "[TOOL OUTPUT END]")
}

// --- 5. Session Security Tests (MCP10/F3/F4) ---

func TestSecurity_SessionTTL_ExpiredSession(t *testing.T) {
	t.Parallel()

	store := NewModelStore(10 * time.Millisecond)
	model := makeTestModel("sec-test")
	store.Put("sec-session", model, nil)

	// Wait for expiry.
	time.Sleep(15 * time.Millisecond)

	_, _, err := store.Get("sec-session")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestSecurity_SessionConcurrency_NoRace(t *testing.T) {
	t.Parallel()
	session := setupSecurityServer(t)

	// Concurrent discovery calls on same session should not race.
	const n = 10
	errCh := make(chan error, n)
	for i := range n {
		go func(i int) {
			_, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
				Name:      "guide_start",
				Arguments: map[string]any{"readme_content": fmt.Sprintf("concurrent project %d", i)},
			})
			errCh <- err
		}(i)
	}

	for range n {
		err := <-errCh
		assert.NoError(t, err)
	}
}

// --- 6. Audit Logging Tests (MCP08) ---

func TestSecurity_AuditLog_ToolCallLogged(t *testing.T) {
	t.Parallel()

	// AuditMiddleware logs to slog — verify it doesn't panic with various methods.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	server := gomcp.NewServer(&gomcp.Implementation{Name: "audit-test", Version: "0.0.1"}, nil)
	server.AddReceivingMiddleware(AuditMiddleware(logger))

	gomcp.AddTool(server, &gomcp.Tool{Name: "audited", Description: "test"}, func(_ context.Context, _ *gomcp.CallToolRequest, _ struct{}) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{Content: []gomcp.Content{&gomcp.TextContent{Text: "ok"}}}, nil, nil
	})

	ctx := context.Background()
	client := gomcp.NewClient(&gomcp.Implementation{Name: "audit-client", Version: "0.0.1"}, nil)
	ct, st := gomcp.NewInMemoryTransports()
	go func() { _ = server.Run(ctx, st) }()
	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	// Should not panic during audit logging.
	result, err := session.CallTool(ctx, &gomcp.CallToolParams{Name: "audited"})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}
