package infrastructure_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
)

// ---------------------------------------------------------------------------
// Mock: DomainResearcher
// ---------------------------------------------------------------------------

// mockDomainResearcher implements application.DomainResearcher for test doubles.
type mockDomainResearcher struct {
	result *discoverydomain.DomainResearchResult
	err    error
}

func (m *mockDomainResearcher) Research(_ context.Context, _ string) (*discoverydomain.DomainResearchResult, error) {
	return m.result, m.err
}

// Compile-time check.
var _ application.DomainResearcher = (*mockDomainResearcher)(nil)

// ---------------------------------------------------------------------------
// Test helper: buildTestResearchResult
// ---------------------------------------------------------------------------

// buildTestResearchResult creates a DomainResearchResult that meets the quality
// floor (>=3 actors, >=3 entities, >=5 workflow steps, >=5 useful sources).
func buildTestResearchResult(t *testing.T) *discoverydomain.DomainResearchResult {
	t.Helper()

	// 3 actors: Customer, Seller, Admin.
	customer, err := discoverydomain.NewResearchedActor("Customer", "end user who buys products", []string{"https://example.com/customer"})
	require.NoError(t, err)

	seller, err := discoverydomain.NewResearchedActor("Seller", "merchant who lists products", []string{"https://example.com/seller"})
	require.NoError(t, err)

	admin, err := discoverydomain.NewResearchedActor("Admin", "platform administrator", []string{"https://example.com/admin"})
	require.NoError(t, err)

	actors := []discoverydomain.ResearchedActor{customer, seller, admin}

	// 3 entities: Order, Product, Payment.
	order, err := discoverydomain.NewResearchedEntity("Order", []string{"id", "status", "total"}, []string{"https://example.com/order"})
	require.NoError(t, err)

	product, err := discoverydomain.NewResearchedEntity("Product", []string{"id", "name", "price"}, []string{"https://example.com/product"})
	require.NoError(t, err)

	payment, err := discoverydomain.NewResearchedEntity("Payment", []string{"id", "amount", "method"}, []string{"https://example.com/payment"})
	require.NoError(t, err)

	entities := []discoverydomain.ResearchedEntity{order, product, payment}

	// 1 happy_path workflow with 5 steps.
	step1, err := discoverydomain.NewWorkflowStep(1, "Customer", "browses", "Product")
	require.NoError(t, err)

	step2, err := discoverydomain.NewWorkflowStep(2, "Customer", "adds to cart", "Product")
	require.NoError(t, err)

	step3, err := discoverydomain.NewWorkflowStep(3, "Customer", "places", "Order")
	require.NoError(t, err)

	step4, err := discoverydomain.NewWorkflowStep(4, "Customer", "submits", "Payment")
	require.NoError(t, err)

	step5, err := discoverydomain.NewWorkflowStep(5, "Seller", "fulfills", "Order")
	require.NoError(t, err)

	workflow, err := discoverydomain.NewResearchedWorkflow(
		"Purchase Flow",
		discoverydomain.WorkflowTypeHappyPath,
		[]discoverydomain.WorkflowStep{step1, step2, step3, step4, step5},
		[]string{"https://example.com/purchase-flow"},
	)
	require.NoError(t, err)

	workflows := []discoverydomain.ResearchedWorkflow{workflow}

	// SearchMetadata with usefulSources=5.
	meta := discoverydomain.NewSearchMetadata(
		[]string{"e-commerce domain", "order processing"},
		10,
		5,
		2*time.Second,
	)

	result, err := discoverydomain.NewDomainResearchResult(
		"E-Commerce Platform",
		meta,
		actors,
		entities,
		workflows,
		[]string{"payment declined", "out of stock"},
		nil, // no regulatory items
		nil, // no existing software
	)
	require.NoError(t, err)

	return &result
}

// buildBelowFloorResearchResult creates a DomainResearchResult that does NOT meet
// the quality floor (1 actor, 1 entity, 1 step, 1 useful source).
func buildBelowFloorResearchResult(t *testing.T) *discoverydomain.DomainResearchResult {
	t.Helper()

	actor, err := discoverydomain.NewResearchedActor("User", "a user", nil)
	require.NoError(t, err)

	entity, err := discoverydomain.NewResearchedEntity("Thing", nil, nil)
	require.NoError(t, err)

	step, err := discoverydomain.NewWorkflowStep(1, "User", "does", "Thing")
	require.NoError(t, err)

	workflow, err := discoverydomain.NewResearchedWorkflow(
		"Simple Flow",
		discoverydomain.WorkflowTypeHappyPath,
		[]discoverydomain.WorkflowStep{step},
		nil,
	)
	require.NoError(t, err)

	meta := discoverydomain.NewSearchMetadata([]string{"query"}, 2, 1, time.Second)

	result, err := discoverydomain.NewDomainResearchResult(
		"Minimal Domain",
		meta,
		[]discoverydomain.ResearchedActor{actor},
		[]discoverydomain.ResearchedEntity{entity},
		[]discoverydomain.ResearchedWorkflow{workflow},
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	return &result
}

// ---------------------------------------------------------------------------
// noopStoryWriter — shared no-op writer for adapter tests.
// ---------------------------------------------------------------------------

// noopStoryWriter is a no-op StoryWriter for adapter tests.
type noopStoryWriter struct{}

func (w *noopStoryWriter) Write(_ context.Context, _ string, _ *discoverydomain.DomainStory) error {
	return nil
}

// Compile-time check.
var _ application.StoryWriter = (*noopStoryWriter)(nil)

// ---------------------------------------------------------------------------
// Setup helpers
// ---------------------------------------------------------------------------

// setupAgentStorytellingAdapterWithResearcher creates an adapter using the NEW
// constructor signature that accepts a DomainResearcher. Uses nil narration lines
// (agent mode delegates to research pipeline, not sequential narration).
func setupAgentStorytellingAdapterWithResearcher(
	t *testing.T,
	researcher application.DomainResearcher,
) (*infrastructure.AgentStorytellingAdapter, *bytes.Buffer) {
	t.Helper()

	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	readmeContent := `# E-Commerce Platform
An online marketplace for buying and selling products.

## Overview
Customers browse product listings and add items to their shopping cart.
When ready, they proceed to checkout and place an order.
The system processes payments through a payment gateway.
Sellers receive notifications about new orders.
Inventory is updated automatically after each purchase.
The platform supports multiple shipping providers.
Returns and refunds are handled through a dedicated workflow.
Customer support agents can view order history and resolve disputes.
`
	require.NoError(t, os.WriteFile(readmePath, []byte(readmeContent), 0o644))

	handler := application.NewDiscoveryHandler(&testPublisher{})
	// nil narration lines — research pipeline replaces sequential narration.
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)
	transformer := application.NewResearchToStoryTransformer()
	storytellingHandler := application.NewStorytellingHandler(&noopStoryWriter{}, prompter, transformer)

	var buf bytes.Buffer

	// NEW constructor signature: researcher parameter between boundaryDetectionHandler and writer.
	adapter := infrastructure.NewAgentStorytellingAdapter(
		handler,
		storytellingHandler,
		nil, // no boundary detection handler
		researcher,
		&buf,
		tmpDir,
	)

	return adapter, &buf
}

// setupAgentStorytellingAdapter creates a temp dir with a rich README and returns
// a configured AgentStorytellingAdapter + output buffer. Updated to use the new
// constructor signature with a valid mock researcher.
func setupAgentStorytellingAdapter(t *testing.T) (*infrastructure.AgentStorytellingAdapter, *bytes.Buffer) {
	t.Helper()

	researcher := &mockDomainResearcher{
		result: buildTestResearchResult(t),
		err:    nil,
	}

	return setupAgentStorytellingAdapterWithResearcher(t, researcher)
}

// ---------------------------------------------------------------------------
// Existing tests — updated to use new setup (research pipeline path).
// ---------------------------------------------------------------------------

func TestAgentStorytellingAdapter_Run_EmitsStoryEnvelopes(t *testing.T) {
	t.Parallel()

	adapter, buf := setupAgentStorytellingAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	require.NotEmpty(t, lines)

	hasStory := false
	for _, line := range lines {
		var env testEnvelope
		require.NoError(t, json.Unmarshal([]byte(line), &env))

		if env.Type == "story" {
			hasStory = true

			break
		}
	}

	assert.True(t, hasStory, "expected at least one 'story' envelope in output")
}

func TestAgentStorytellingAdapter_Run_EachLineIsValidJSON(t *testing.T) {
	t.Parallel()

	adapter, buf := setupAgentStorytellingAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	require.NotEmpty(t, lines)

	for i, line := range lines {
		var raw json.RawMessage
		assert.NoError(t, json.Unmarshal([]byte(line), &raw), "line %d is not valid JSON: %s", i, line)
	}
}

func TestAgentStorytellingAdapter_Run_EmitsDiscoveryComplete(t *testing.T) {
	t.Parallel()

	adapter, buf := setupAgentStorytellingAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	require.NotEmpty(t, lines)

	lastLine := lines[len(lines)-1]
	var env testEnvelope
	require.NoError(t, json.Unmarshal([]byte(lastLine), &env))
	assert.Equal(t, "discovery_complete", env.Type)
}

func TestAgentStorytellingAdapter_Run_StoryEnvelopeHasExpectedFields(t *testing.T) {
	t.Parallel()

	adapter, buf := setupAgentStorytellingAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)
	require.NotEmpty(t, lines)

	for _, line := range lines {
		var env testEnvelope
		require.NoError(t, json.Unmarshal([]byte(line), &env))

		if env.Type == "story" {
			var storyOut infrastructure.StoryOutput
			require.NoError(t, json.Unmarshal(env.Data, &storyOut))

			assert.NotEmpty(t, storyOut.SessionID, "story envelope should have non-empty session_id")
			assert.Positive(t, storyOut.StoryIndex, "story envelope should have story_index > 0")
			assert.NotEmpty(t, storyOut.Title, "story envelope should have non-empty title")

			return // verified first story envelope
		}
	}

	t.Fatal("no 'story' envelope found in output")
}

func TestAgentStorytellingAdapter_Run_NoReadme_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir() // empty — no README.md

	handler := application.NewDiscoveryHandler(&testPublisher{})
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)
	transformer := application.NewResearchToStoryTransformer()
	storytellingHandler := application.NewStorytellingHandler(&noopStoryWriter{}, prompter, transformer)
	researcher := &mockDomainResearcher{result: buildTestResearchResult(t)}

	var buf bytes.Buffer

	adapter := infrastructure.NewAgentStorytellingAdapter(
		handler, storytellingHandler, nil, researcher, &buf, tmpDir,
	)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "README")
}

func TestAgentStorytellingAdapter_Run_ReadmeTooShort_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Short\n"), 0o644))

	handler := application.NewDiscoveryHandler(&testPublisher{})
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)
	transformer := application.NewResearchToStoryTransformer()
	storytellingHandler := application.NewStorytellingHandler(&noopStoryWriter{}, prompter, transformer)

	// Researcher returns nil result for short README — simulates research
	// infrastructure declining to research insufficient content.
	researcher := &mockDomainResearcher{result: nil, err: nil}

	var buf bytes.Buffer

	adapter := infrastructure.NewAgentStorytellingAdapter(
		handler, storytellingHandler, nil, researcher, &buf, tmpDir,
	)

	err := adapter.Run(context.Background())
	require.Error(t, err)
}

func TestAgentStorytellingAdapter_Run_NilBoundaryHandler_NoBoundaryProposals(t *testing.T) {
	t.Parallel()

	adapter, buf := setupAgentStorytellingAdapter(t)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	lines := parseOutputLines(buf)

	for _, line := range lines {
		var env testEnvelope
		require.NoError(t, json.Unmarshal([]byte(line), &env))
		assert.NotEqual(t, "boundary_proposals", env.Type, "should not emit boundary_proposals when handler is nil")
	}
}

// ---------------------------------------------------------------------------
// NEW tests — research pipeline error paths.
// ---------------------------------------------------------------------------

func TestAgentStorytellingAdapter_Run_NilResearcher_ReturnsError(t *testing.T) {
	t.Parallel()

	// Create adapter with nil researcher via the new setup helper.
	adapter, _ := setupAgentStorytellingAdapterWithResearcher(t, nil)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain researcher")
}

func TestAgentStorytellingAdapter_Run_ResearchReturnsNil_ReturnsError(t *testing.T) {
	t.Parallel()

	// Mock returns (nil, nil) — research infrastructure unavailable (no LLM credentials).
	researcher := &mockDomainResearcher{result: nil, err: nil}

	adapter, _ := setupAgentStorytellingAdapterWithResearcher(t, researcher)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "LLM credentials")
}

func TestAgentStorytellingAdapter_Run_ResearchReturnsError_PropagatesError(t *testing.T) {
	t.Parallel()

	// Mock returns an error — unexpected failure in research infrastructure.
	researchErr := fmt.Errorf("network timeout")
	researcher := &mockDomainResearcher{result: nil, err: researchErr}

	adapter, _ := setupAgentStorytellingAdapterWithResearcher(t, researcher)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, researchErr)
}

func TestAgentStorytellingAdapter_Run_ResearchProducesNoStories_ReturnsError(t *testing.T) {
	t.Parallel()

	// Research result that does NOT meet quality floor — transformer returns (nil, nil).
	researcher := &mockDomainResearcher{
		result: buildBelowFloorResearchResult(t),
		err:    nil,
	}

	adapter, _ := setupAgentStorytellingAdapterWithResearcher(t, researcher)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no stories")
}
