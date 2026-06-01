// Package integration provides BDD-style integration tests.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
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
// Helpers
// ---------------------------------------------------------------------------

// newThreeStoryPrompterWithDisjointActors creates a prompter for 3 rapid-mode
// stories where stories 1 and 2 have completely different actors, triggering
// SignalTypeOrgBoundary in the AlgorithmicDetector.
//
// Story 1: actor "Customer", activity "places order", object "Order Form"
// Story 2: actor "Warehouse Manager", activity "ships package", object "Shipping Label"
// Story 3: actor "Customer", activity "tracks delivery", object "Tracking Page"
func newThreeStoryPrompterWithDisjointActors() *fakeStorytellingPrompter {
	// Each RunStory call consumes:
	// 1. trigger (opening MQ-O2)
	// 2. actor (opening MQ-O1)
	// 3. activity (narration MQ-N2) — one sentence
	// 4. subject (narration MQ-N3)
	// 5. object (narration MQ-N4)
	// 6. "" (narration MQ-N2 = end narration loop)
	narrations := []string{
		// Story 1: Customer places order
		"Customer places order",
		"Customer",
		"places order",
		"Customer",
		"Order Form",
		"",
		// Story 2: Warehouse Manager ships package
		"Warehouse Manager ships package",
		"Warehouse Manager",
		"ships package",
		"Warehouse Manager",
		"Shipping Label",
		"",
		// Story 3: Customer tracks delivery
		"Customer tracks delivery",
		"Customer",
		"tracks delivery",
		"Customer",
		"Tracking Page",
		"",
	}

	// AskChoice: first call is persona ("1"), then "yes" to continue for non-final stories
	choices := []string{"1", "yes", "yes"}

	return &fakeStorytellingPrompter{
		modeChoice:              discoverydomain.ModeRapid,
		narrationResponses:      narrations,
		confirmSentenceAccepted: true,
		synthesisResult:         true,
		choiceResponses:         choices,
	}
}

// jsonlEnvelope is a minimal representation of a JSONL envelope for parsing.
type jsonlEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// parseEnvelopes reads all JSONL lines from a buffer and returns typed envelopes.
func parseEnvelopes(t *testing.T, buf *bytes.Buffer) []jsonlEnvelope {
	t.Helper()

	var envelopes []jsonlEnvelope

	decoder := json.NewDecoder(buf)

	for decoder.More() {
		var env jsonlEnvelope
		require.NoError(t, decoder.Decode(&env))
		envelopes = append(envelopes, env)
	}

	return envelopes
}

// ---------------------------------------------------------------------------
// Scenario 1 — RAPID E2E: full pipeline completes with 3 stories
// ---------------------------------------------------------------------------

// TestStorytellingPipeline_RAPID_E2E_CompletesWithStatusCompleted verifies the
// full cross-layer wiring through CLIDiscoveryAdapter with 3 rapid-mode stories.
func TestStorytellingPipeline_RAPID_E2E_CompletesWithStatusCompleted(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea for pipeline test"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "alto-scaffold"), 0o755))

	// And: a prompter configured for 3 rapid-mode stories
	prompter := newIntegrationStoryPrompter(3)
	writer := &fakeStoryWriter{}

	// And: fully wired handler + storytelling handler + boundary fakes + adapter
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector, bPrompter, cmWriter := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error — full pipeline completed
	require.NoError(t, err)

	// And: exactly 3 stories were persisted via the writer
	assert.Len(t, writer.written, 3)

	// And: each story has a non-empty title
	for i, story := range writer.written {
		assert.NotEmpty(t, story.Title(), "story %d should have a non-empty title", i+1)
	}

	// And: each story has at least one sentence
	for i, story := range writer.written {
		assert.NotEmpty(t, story.Sentences(), "story %d should have at least one sentence", i+1)
	}
}

// ---------------------------------------------------------------------------
// Scenario 2 — Boundary detection with real AlgorithmicDetector
// ---------------------------------------------------------------------------

// TestStorytellingPipeline_BoundaryDetection_AlgorithmicDetector_ProducesSketches
// verifies that the real AlgorithmicDetector produces boundary sketches when
// stories have disjoint actors, using direct handler calls.
func TestStorytellingPipeline_BoundaryDetection_AlgorithmicDetector_ProducesSketches(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("Boundary detection test project"), 0o644))

	// And: a prompter configured for 3 stories with disjoint actors
	prompter := newThreeStoryPrompterWithDisjointActors()
	writer := &fakeStoryWriter{}

	// And: wired handler with real AlgorithmicDetector
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	realDetector := infrastructure.NewAlgorithmicDetector()
	bdHandler := application.NewBoundaryDetectionHandler(realDetector)

	// And: a session with mode and persona
	session, err := handler.StartSession("Boundary detection test project")
	require.NoError(t, err)
	require.NoError(t, session.SetMode(discoverydomain.ModeRapid))
	session, err = handler.DetectPersona(session.SessionID(), "1")
	require.NoError(t, err)

	flow, err := discoverydomain.NewStorytellingFlow(discoverydomain.ModeRapid)
	require.NoError(t, err)

	// When: running 3 stories via direct handler calls
	var accumulatedStories []*discoverydomain.DomainStory

	for i := 1; i <= 3; i++ {
		story, _, storyErr := storytellingHandler.RunStory(context.Background(), session, i, flow)
		require.NoError(t, storyErr, "RunStory %d should not fail", i)
		accumulatedStories = append(accumulatedStories, story)
	}

	// And: running boundary detection with the accumulated stories
	sketches, err := bdHandler.Detect(context.Background(), accumulatedStories, discoverydomain.ModeRapid)

	// Then: no error
	require.NoError(t, err)

	// And: at least 1 sketch was produced (disjoint actors trigger OrgBoundary signal)
	require.NotEmpty(t, sketches)

	// And: at least one sketch has a non-empty name
	foundNamedSketch := false
	for _, sketch := range sketches {
		if sketch.Name() != "" {
			foundNamedSketch = true
			break
		}
	}
	assert.True(t, foundNamedSketch, "at least one sketch should have a non-empty name")
}

// ---------------------------------------------------------------------------
// Scenario 3 — Session resume: snapshot round-trip preserves story refs
// ---------------------------------------------------------------------------

// TestStorytellingPipeline_SessionResume_SnapshotRoundTrip_PreservesStoryRefs
// verifies that session state (including StoryRefs) survives a JSON
// save/load round-trip through FileSystemSessionRepository.
func TestStorytellingPipeline_SessionResume_SnapshotRoundTrip_PreservesStoryRefs(t *testing.T) {
	t.Parallel()

	// Given: a real FileSystemSessionRepository with a temp dir
	tmpDir := t.TempDir()
	repo := infrastructure.NewFileSystemSessionRepository(tmpDir)

	// And: a session with mode and persona
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	session, err := handler.StartSession("Snapshot round-trip test")
	require.NoError(t, err)
	require.NoError(t, session.SetMode(discoverydomain.ModeRapid))
	session, err = handler.DetectPersona(session.SessionID(), "1")
	require.NoError(t, err)

	// And: a storytelling handler with fakes
	prompter := newIntegrationStoryPrompter(3)
	writer := &fakeStoryWriter{}
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	flow, err := discoverydomain.NewStorytellingFlow(discoverydomain.ModeRapid)
	require.NoError(t, err)

	// When: running story 1
	_, _, err = storytellingHandler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)

	// And: saving the session
	require.NoError(t, repo.Save(context.Background(), session))

	// And: loading the session back
	loaded, err := repo.Load(context.Background(), session.SessionID())
	require.NoError(t, err)

	// Then: loaded session has 1 story ref
	assert.Len(t, loaded.StoryRefs(), 1)

	// When: running story 2 on the loaded session
	_, _, err = storytellingHandler.RunStory(context.Background(), loaded, 2, flow)
	require.NoError(t, err)

	// And: saving and loading again
	require.NoError(t, repo.Save(context.Background(), loaded))
	loaded2, err := repo.Load(context.Background(), loaded.SessionID())
	require.NoError(t, err)

	// Then: loaded session has 2 story refs
	assert.Len(t, loaded2.StoryRefs(), 2)
}

// ---------------------------------------------------------------------------
// Fake DomainResearcher for agent mode integration tests
// ---------------------------------------------------------------------------

type fakeDomainResearcher struct {
	result *discoverydomain.DomainResearchResult
	err    error
}

func (f *fakeDomainResearcher) Research(_ context.Context, _ string) (*discoverydomain.DomainResearchResult, error) {
	return f.result, f.err
}

var _ application.DomainResearcher = (*fakeDomainResearcher)(nil)

// buildIntegrationResearchResult creates a DomainResearchResult that meets the
// quality floor for the ResearchToStoryTransformer (>=3 actors, >=3 entities,
// >=5 workflow steps, >=5 useful sources).
func buildIntegrationResearchResult(t *testing.T) *discoverydomain.DomainResearchResult {
	t.Helper()

	customer, err := discoverydomain.NewResearchedActor("Customer", "end user who places orders", []string{"https://example.com/customer"})
	require.NoError(t, err)

	warehouse, err := discoverydomain.NewResearchedActor("Warehouse Manager", "fulfills orders", []string{"https://example.com/warehouse"})
	require.NoError(t, err)

	admin, err := discoverydomain.NewResearchedActor("Admin", "platform administrator", []string{"https://example.com/admin"})
	require.NoError(t, err)

	order, err := discoverydomain.NewResearchedEntity("Order", []string{"id", "status", "total"}, []string{"https://example.com/order"})
	require.NoError(t, err)

	product, err := discoverydomain.NewResearchedEntity("Product", []string{"id", "name", "price"}, []string{"https://example.com/product"})
	require.NoError(t, err)

	payment, err := discoverydomain.NewResearchedEntity("Payment", []string{"id", "amount", "method"}, []string{"https://example.com/payment"})
	require.NoError(t, err)

	step1, err := discoverydomain.NewWorkflowStep(1, "Customer", "browses", "Product")
	require.NoError(t, err)

	step2, err := discoverydomain.NewWorkflowStep(2, "Customer", "adds to cart", "Product")
	require.NoError(t, err)

	step3, err := discoverydomain.NewWorkflowStep(3, "Customer", "places", "Order")
	require.NoError(t, err)

	step4, err := discoverydomain.NewWorkflowStep(4, "Customer", "submits", "Payment")
	require.NoError(t, err)

	step5, err := discoverydomain.NewWorkflowStep(5, "Warehouse Manager", "fulfills", "Order")
	require.NoError(t, err)

	workflow, err := discoverydomain.NewResearchedWorkflow(
		"Purchase Flow",
		discoverydomain.WorkflowTypeHappyPath,
		[]discoverydomain.WorkflowStep{step1, step2, step3, step4, step5},
		[]string{"https://example.com/purchase-flow"},
	)
	require.NoError(t, err)

	meta := discoverydomain.NewSearchMetadata(
		[]string{"e-commerce domain", "order processing"},
		10, 5, 2*time.Second,
	)

	result, err := discoverydomain.NewDomainResearchResult(
		"E-Commerce Platform",
		meta,
		[]discoverydomain.ResearchedActor{customer, warehouse, admin},
		[]discoverydomain.ResearchedEntity{order, product, payment},
		[]discoverydomain.ResearchedWorkflow{workflow},
		[]string{"payment declined", "out of stock"},
		nil,
		nil,
	)
	require.NoError(t, err)

	return &result
}

// ---------------------------------------------------------------------------
// Scenario 4 — Agent mode: JSONL output structure (research pipeline)
// ---------------------------------------------------------------------------

// TestStorytellingPipeline_AgentMode_JSONL_OutputStructure verifies that the
// AgentStorytellingAdapter emits well-structured JSONL envelopes with expected types.
func TestStorytellingPipeline_AgentMode_JSONL_OutputStructure(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("Agent JSONL test project"), 0o644))

	// And: a mock researcher returning a valid research result
	researcher := &fakeDomainResearcher{
		result: buildIntegrationResearchResult(t),
	}

	// And: a prompter (nil narration lines — research pipeline) and transformer
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)
	writer := &fakeStoryWriter{}
	transformer := application.NewResearchToStoryTransformer()

	// And: wired handlers
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, transformer)
	detector, _, _ := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	// And: an AgentStorytellingAdapter writing to a buffer
	var buf bytes.Buffer
	adapter := infrastructure.NewAgentStorytellingAdapter(handler, storytellingHandler, bdHandler, researcher, &buf, tmpDir)

	// When: running the agent storytelling flow
	err := adapter.Run(context.Background())

	// Then: no error
	require.NoError(t, err)

	// And: JSONL output is non-empty
	require.NotEmpty(t, buf.Bytes())

	// And: all lines parse as valid envelopes
	envelopes := parseEnvelopes(t, &buf)
	require.NotEmpty(t, envelopes)

	// And: each envelope has a non-empty type
	for i, env := range envelopes {
		assert.NotEmpty(t, env.Type, "envelope %d should have a non-empty type", i)
		assert.NotEmpty(t, env.Data, "envelope %d should have non-empty data", i)
	}

	// And: at least one "story" type envelope exists
	foundStory := false
	for _, env := range envelopes {
		if env.Type == "story" {
			foundStory = true
			break
		}
	}
	assert.True(t, foundStory, "should have at least one 'story' envelope")

	// And: last envelope is "discovery_complete"
	assert.Equal(t, "discovery_complete", envelopes[len(envelopes)-1].Type)
}
