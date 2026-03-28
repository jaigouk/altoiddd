package infrastructure_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
)

// noopStoryWriter is a no-op StoryWriter for adapter tests.
type noopStoryWriter struct{}

func (w *noopStoryWriter) Write(_ context.Context, _ string, _ *discoverydomain.DomainStory) error {
	return nil
}

// Compile-time check.
var _ application.StoryWriter = (*noopStoryWriter)(nil)

// setupAgentStorytellingAdapter creates a temp dir with a rich README and returns
// a configured AgentStorytellingAdapter + output buffer.
func setupAgentStorytellingAdapter(t *testing.T) (*infrastructure.AgentStorytellingAdapter, *bytes.Buffer) {
	t.Helper()

	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	// README with enough content for RAPID mode narration (9+ non-empty lines).
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

	// Narration lines extracted from README for the prompter.
	narrationLines := []string{
		"Customer visits online store",
		"Customer",
		"Customer browses ProductListing",
		"Customer adds ProductListing to ShoppingCart",
		"Customer creates Order",
		"",
		"Customer visits store again",
		"Seller",
		"Seller reviews Order",
		"Seller ships Order to Customer",
		"Seller updates Inventory",
		"",
		"Admin monitors platform",
		"Admin",
		"Admin generates Report",
		"Admin resolves Dispute for Customer",
		"Admin updates Policy",
		"",
	}

	handler := application.NewDiscoveryHandler(&testPublisher{})
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, narrationLines)
	transformer := application.NewResearchToStoryTransformer()
	storytellingHandler := application.NewStorytellingHandler(&noopStoryWriter{}, prompter, transformer)

	var buf bytes.Buffer
	adapter := infrastructure.NewAgentStorytellingAdapter(
		handler,
		storytellingHandler,
		nil, // no boundary detection handler
		&buf,
		tmpDir,
	)

	return adapter, &buf
}

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

	var buf bytes.Buffer
	adapter := infrastructure.NewAgentStorytellingAdapter(handler, storytellingHandler, nil, &buf, tmpDir)

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
	// Only 1 narration line — not enough for a RAPID session (needs 3 stories).
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, []string{"too short"})
	transformer := application.NewResearchToStoryTransformer()
	storytellingHandler := application.NewStorytellingHandler(&noopStoryWriter{}, prompter, transformer)

	var buf bytes.Buffer
	adapter := infrastructure.NewAgentStorytellingAdapter(handler, storytellingHandler, nil, &buf, tmpDir)

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
