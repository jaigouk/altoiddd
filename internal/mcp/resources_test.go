package mcp_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	knowledgeapp "github.com/alto-cli/alto/internal/knowledge/application"
	knowledgedomain "github.com/alto-cli/alto/internal/knowledge/domain"
	mcptools "github.com/alto-cli/alto/internal/mcp"
	shareddomain "github.com/alto-cli/alto/internal/shared/domain"
	ticketapp "github.com/alto-cli/alto/internal/ticket/application"
	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
	ttapp "github.com/alto-cli/alto/internal/tooltranslation/application"
)

// --- Test stubs ---

// stubKnowledgeReader implements knowledgeapp.KnowledgeReader for testing.
type stubKnowledgeReader struct {
	entries map[string]knowledgedomain.KnowledgeEntry
}

func (s *stubKnowledgeReader) ReadEntry(_ context.Context, path knowledgedomain.KnowledgePath, _ string) (knowledgedomain.KnowledgeEntry, error) {
	entry, ok := s.entries[path.Raw()]
	if !ok {
		return knowledgedomain.KnowledgeEntry{}, fmt.Errorf("entry not found: %s", path.Raw())
	}
	return entry, nil
}

func (s *stubKnowledgeReader) ListTopics(_ context.Context, _ knowledgedomain.KnowledgeCategory, _ *string) ([]string, error) {
	return nil, nil
}

// stubTicketReader implements ticketapp.TicketReader for testing.
type stubTicketReader struct {
	tickets []ticketdomain.OpenTicketData
}

func (s *stubTicketReader) ReadOpenTickets(_ context.Context) ([]ticketdomain.OpenTicketData, error) {
	return s.tickets, nil
}

func (s *stubTicketReader) ReadFlags(_ context.Context, _ string) ([]ticketdomain.FreshnessFlag, error) {
	return nil, nil
}

// stubFileWriter implements sharedapp.FileWriter for testing.
type stubFileWriter struct{}

func (s *stubFileWriter) WriteFile(_ context.Context, _ string, _ string) error {
	return nil
}

// --- Setup helpers ---

func makeKnowledgeEntry(path, title, content string) knowledgedomain.KnowledgeEntry {
	p, err := knowledgedomain.NewKnowledgePath(path)
	if err != nil {
		panic(fmt.Sprintf("invalid test path %q: %v", path, err))
	}
	return knowledgedomain.NewKnowledgeEntry(p, title, content, nil, "markdown")
}

func setupResourceServer(t *testing.T, app *composition.App) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	mcptools.RegisterResources(server, app)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	go func() { _ = server.Run(ctx, st) }()

	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

func testAppWithKnowledge(entries map[string]knowledgedomain.KnowledgeEntry) *composition.App {
	reader := &stubKnowledgeReader{entries: entries}
	return &composition.App{
		KnowledgeLookupHandler: knowledgeapp.NewKnowledgeLookupHandler(reader),
		TicketHealthHandler:    ticketapp.NewTicketHealthHandler(&stubTicketReader{}),
		PersonaHandler:         ttapp.NewPersonaHandler(&stubFileWriter{}),
	}
}

func testAppWithTickets(tickets []ticketdomain.OpenTicketData) *composition.App {
	reader := &stubKnowledgeReader{entries: map[string]knowledgedomain.KnowledgeEntry{}}
	return &composition.App{
		KnowledgeLookupHandler: knowledgeapp.NewKnowledgeLookupHandler(reader),
		TicketHealthHandler:    ticketapp.NewTicketHealthHandler(&stubTicketReader{tickets: tickets}),
		PersonaHandler:         ttapp.NewPersonaHandler(&stubFileWriter{}),
	}
}

func testAppMinimal() *composition.App {
	reader := &stubKnowledgeReader{entries: map[string]knowledgedomain.KnowledgeEntry{}}
	return &composition.App{
		KnowledgeLookupHandler: knowledgeapp.NewKnowledgeLookupHandler(reader),
		TicketHealthHandler:    ticketapp.NewTicketHealthHandler(&stubTicketReader{}),
		PersonaHandler:         ttapp.NewPersonaHandler(&stubFileWriter{}),
	}
}

// --- Knowledge DDD resource tests ---

func TestKnowledgeDDDResource_HappyPath(t *testing.T) {
	t.Parallel()
	app := testAppWithKnowledge(map[string]knowledgedomain.KnowledgeEntry{
		"ddd/aggregates": makeKnowledgeEntry("ddd/aggregates", "Aggregates", "# Aggregates\nAn aggregate is..."),
	})
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://knowledge/ddd/aggregates",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Aggregates")
	assert.Equal(t, "text/markdown", result.Contents[0].MIMEType)
}

func TestKnowledgeDDDResource_NotFound(t *testing.T) {
	t.Parallel()
	app := testAppWithKnowledge(map[string]knowledgedomain.KnowledgeEntry{})
	session := setupResourceServer(t, app)

	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://knowledge/ddd/nonexistent",
	})
	require.Error(t, err)
}

// --- Knowledge Tools resource tests ---

func TestKnowledgeToolsResource_HappyPath(t *testing.T) {
	t.Parallel()
	app := testAppWithKnowledge(map[string]knowledgedomain.KnowledgeEntry{
		"tools/claude-code/setup": makeKnowledgeEntry("tools/claude-code/setup", "Claude Setup", "# Setup Claude Code"),
	})
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://knowledge/tools/claude-code/setup",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Setup Claude Code")
}

// --- Knowledge Conventions resource tests ---

func TestKnowledgeConventionsResource_HappyPath(t *testing.T) {
	t.Parallel()
	app := testAppWithKnowledge(map[string]knowledgedomain.KnowledgeEntry{
		"conventions/testing": makeKnowledgeEntry("conventions/testing", "Testing", "# Testing conventions"),
	})
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://knowledge/conventions/testing",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Testing conventions")
}

// --- Knowledge Cross-Tool resource tests ---

func TestKnowledgeCrossToolResource_HappyPath(t *testing.T) {
	t.Parallel()
	app := testAppWithKnowledge(map[string]knowledgedomain.KnowledgeEntry{
		"cross-tool/migration": makeKnowledgeEntry("cross-tool/migration", "Migration", "# Migration guide"),
	})
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://knowledge/cross-tool/migration",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Migration guide")
}

// --- Project document resource tests ---

func TestProjectDomainModelResource_HappyPath(t *testing.T) {
	// Create a temp project dir with docs/DDD.md
	root := t.TempDir()
	projectDir := filepath.Join(root, "myproject")
	docsDir := filepath.Join(projectDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "DDD.md"), []byte("# Domain Model\nBounded contexts..."), 0o644))

	// We need to chdir so SafeProjectPath uses CWD as root
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	app := testAppMinimal()
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://project/myproject/domain-model",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Domain Model")
	assert.Equal(t, "text/markdown", result.Contents[0].MIMEType)
}

func TestProjectDomainModelResource_FileNotFound(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "emptyproject")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	app := testAppMinimal()
	session := setupResourceServer(t, app)

	_, err = session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://project/emptyproject/domain-model",
	})
	require.Error(t, err)
}

func TestProjectArchitectureResource_HappyPath(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "myproject")
	docsDir := filepath.Join(projectDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "ARCHITECTURE.md"), []byte("# Architecture"), 0o644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	app := testAppMinimal()
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://project/myproject/architecture",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Architecture")
}

func TestProjectPRDResource_HappyPath(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "myproject")
	docsDir := filepath.Join(projectDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte("# Product Requirements"), 0o644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	app := testAppMinimal()
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://project/myproject/prd",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Product Requirements")
}

func TestProjectResource_PathTraversal(t *testing.T) {
	root := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	app := testAppMinimal()
	session := setupResourceServer(t, app)

	_, err = session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://project/../../../etc/domain-model",
	})
	require.Error(t, err)
}

// --- Ticket resource tests ---

func TestTicketsReadyResource_HappyPath(t *testing.T) {
	t.Parallel()
	reviewed := "2025-01-15"
	tickets := []ticketdomain.OpenTicketData{
		ticketdomain.NewOpenTicketData("k7m.5", "MCP resources", []string{"review_needed"}, &reviewed),
		ticketdomain.NewOpenTicketData("k7m.6", "PRD review", nil, nil),
	}
	app := testAppWithTickets(tickets)
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://tickets/ready",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Ticket Health Report")
	assert.Contains(t, result.Contents[0].Text, "Total open: 2")
}

func TestTicketsReadyResource_NoTickets(t *testing.T) {
	t.Parallel()
	app := testAppWithTickets(nil)
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://tickets/ready",
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Total open: 0")
	assert.Contains(t, result.Contents[0].Text, "Review needed: 0")
}

// --- Ticket by ID resource tests ---

func TestTicketsByIDResource_ValidID_PlaceholderError(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	// Valid ID format, but handler returns placeholder error (bd binary not available)
	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://tickets/k7m.12",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires bd binary")
}

func TestTicketsByIDResource_InvalidID(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://tickets/'; cat /etc/passwd",
	})
	require.Error(t, err)
}

// --- Persona resource tests ---

func TestPersonasResource_HappyPath(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	// Get the list of actual persona types from the handler
	personas := app.PersonaHandler.ListPersonas()
	require.NotEmpty(t, personas)

	// Use the first persona's type (URL-safe slug, e.g. "solo_developer")
	pType := string(personas[0].PersonaType())
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: fmt.Sprintf("alto://personas/%s", pType),
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "Name: "+personas[0].Name())
}

func TestPersonasResource_NotFound(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://personas/nonexistent",
	})
	require.Error(t, err)
}

// --- List resources/templates tests ---

func TestListResourceTemplates(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	result, err := session.ListResourceTemplates(context.Background(), nil)
	require.NoError(t, err)
	// 4 knowledge + 3 project + 1 ticket by ID + 1 persona = 9 templates
	assert.Len(t, result.ResourceTemplates, 9)

	// Verify some specific templates exist
	uris := make([]string, len(result.ResourceTemplates))
	for i, rt := range result.ResourceTemplates {
		uris[i] = rt.URITemplate
	}
	assert.Contains(t, uris, "alto://knowledge/ddd/{topic}")
	assert.Contains(t, uris, "alto://knowledge/tools/{tool}/{subtopic}")
	assert.Contains(t, uris, "alto://project/{dir}/domain-model")
	assert.Contains(t, uris, "alto://tickets/{id}")
	assert.Contains(t, uris, "alto://personas/{type}")
}

func TestListResources(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	result, err := session.ListResources(context.Background(), nil)
	require.NoError(t, err)
	// 1 static: alto://tickets/ready
	assert.Len(t, result.Resources, 1)
	assert.Equal(t, "alto://tickets/ready", result.Resources[0].URI)
}

// --- Edge case tests ---

func TestKnowledgeResource_EmptyTopic(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	// URI with empty topic segment — should fail SafeComponent
	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://knowledge/ddd/",
	})
	require.Error(t, err)
}

func TestKnowledgeToolsResource_ValidHyphenatedName(t *testing.T) {
	t.Parallel()
	app := testAppWithKnowledge(map[string]knowledgedomain.KnowledgeEntry{
		"tools/claude-code/overview": makeKnowledgeEntry("tools/claude-code/overview", "Overview", "# Claude Code overview"),
	})
	session := setupResourceServer(t, app)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "alto://knowledge/tools/claude-code/overview",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "Claude Code overview")
}

func TestPersonasResource_CaseInsensitive(t *testing.T) {
	t.Parallel()
	app := testAppMinimal()
	session := setupResourceServer(t, app)

	personas := app.PersonaHandler.ListPersonas()
	require.NotEmpty(t, personas)
	pType := string(personas[0].PersonaType())

	// Try uppercase version — should match case-insensitively
	upperType := strings.ToUpper(pType)
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: fmt.Sprintf("alto://personas/%s", upperType),
	})
	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "Name: "+personas[0].Name())
}

// ---------------------------------------------------------------------------
// Session Status Resource Tests (TDD RED phase)
// ---------------------------------------------------------------------------

func TestSessionStatusResource_ReturnsAvailableActions(t *testing.T) {
	t.Parallel()

	coord := shareddomain.NewWorkflowCoordinator()
	sessionID := "status-test-session"

	// Mark some steps as ready
	coord.MarkReady(sessionID, shareddomain.StepFitness, shareddomain.StepTickets)

	app := testAppMinimal()
	session := setupResourceServerWithCoordinator(t, app, coord)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: fmt.Sprintf("alto://session/%s/status", sessionID),
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)

	// Should contain the session_id and available_actions
	text := result.Contents[0].Text
	assert.Contains(t, text, sessionID)
	assert.Contains(t, text, "fitness")
	assert.Contains(t, text, "tickets")
}

func setupResourceServerWithCoordinator(t *testing.T, app *composition.App, coord *shareddomain.WorkflowCoordinator) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	mcptools.RegisterResourcesWithCoordinator(server, app, coord)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	go func() { _ = server.Run(ctx, st) }()

	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}
