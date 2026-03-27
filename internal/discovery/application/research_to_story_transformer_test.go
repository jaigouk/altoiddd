package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- Test Helpers ---

// mustResearchedActor creates a ResearchedActor or panics.
func mustResearchedActor(t *testing.T, name, role string, sourceURLs []string) discoverydomain.ResearchedActor {
	t.Helper()

	a, err := discoverydomain.NewResearchedActor(name, role, sourceURLs)
	require.NoError(t, err)

	return a
}

// mustResearchedEntity creates a ResearchedEntity or panics.
func mustResearchedEntity(t *testing.T, name string, props, sourceURLs []string) discoverydomain.ResearchedEntity {
	t.Helper()

	e, err := discoverydomain.NewResearchedEntity(name, props, sourceURLs)
	require.NoError(t, err)

	return e
}

// mustWorkflowStep creates a WorkflowStep or panics.
func mustWorkflowStep(t *testing.T, seq int, actor, activity, workObject string) discoverydomain.WorkflowStep {
	t.Helper()

	s, err := discoverydomain.NewWorkflowStep(seq, actor, activity, workObject)
	require.NoError(t, err)

	return s
}

// mustResearchedWorkflow creates a ResearchedWorkflow or panics.
func mustResearchedWorkflow(t *testing.T, name string, wfType discoverydomain.WorkflowType, steps []discoverydomain.WorkflowStep, sourceURLs []string) discoverydomain.ResearchedWorkflow {
	t.Helper()

	wf, err := discoverydomain.NewResearchedWorkflow(name, wfType, steps, sourceURLs)
	require.NoError(t, err)

	return wf
}

// mustDomainResearchResult creates a DomainResearchResult or panics.
func mustDomainResearchResult(
	t *testing.T,
	domain string,
	meta discoverydomain.SearchMetadata,
	actors []discoverydomain.ResearchedActor,
	entities []discoverydomain.ResearchedEntity,
	workflows []discoverydomain.ResearchedWorkflow,
) discoverydomain.DomainResearchResult {
	t.Helper()

	r, err := discoverydomain.NewDomainResearchResult(domain, meta, actors, entities, workflows, nil, nil, nil)
	require.NoError(t, err)

	return r
}

// qualityFloorMeta returns SearchMetadata with enough useful sources to meet quality floor.
func qualityFloorMeta() discoverydomain.SearchMetadata {
	return discoverydomain.NewSearchMetadata([]string{"q1", "q2"}, 10, 5, time.Second)
}

// minActors returns 3 ResearchedActors to meet quality floor.
func minActors(t *testing.T) []discoverydomain.ResearchedActor {
	t.Helper()

	return []discoverydomain.ResearchedActor{
		mustResearchedActor(t, "Customer", "buyer", []string{"https://example.com/a"}),
		mustResearchedActor(t, "Clerk", "employee", []string{"https://example.com/b"}),
		mustResearchedActor(t, "Manager", "supervisor", []string{"https://example.com/c"}),
	}
}

// minEntities returns 3 ResearchedEntities to meet quality floor.
func minEntities(t *testing.T) []discoverydomain.ResearchedEntity {
	t.Helper()

	return []discoverydomain.ResearchedEntity{
		mustResearchedEntity(t, "Order", nil, []string{"https://example.com/e1"}),
		mustResearchedEntity(t, "Invoice", nil, []string{"https://example.com/e2"}),
		mustResearchedEntity(t, "Receipt", nil, []string{"https://example.com/e3"}),
	}
}

// happyPathWorkflow returns a single happy_path workflow with 5 steps (meets quality floor).
func happyPathWorkflow(t *testing.T) discoverydomain.ResearchedWorkflow {
	t.Helper()

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}

	return mustResearchedWorkflow(t, "Place Order", discoverydomain.WorkflowTypeHappyPath, steps, []string{"https://example.com/wf1"})
}

// --- Tests ---

func TestResearchToStoryTransformer_Transform_QualityFloorNotMet(t *testing.T) {
	t.Parallel()

	// Given: a research result that does NOT meet quality floor (only 1 actor, 1 entity, 1 step).
	meta := discoverydomain.NewSearchMetadata(nil, 1, 1, time.Second)
	actors := []discoverydomain.ResearchedActor{
		mustResearchedActor(t, "User", "user", nil),
	}
	entities := []discoverydomain.ResearchedEntity{
		mustResearchedEntity(t, "Form", nil, nil),
	}
	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "User", "fills", "Form"),
	}
	wf := mustResearchedWorkflow(t, "Submit", discoverydomain.WorkflowTypeHappyPath, steps, nil)
	result := mustDomainResearchResult(t, "test", meta, actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: nil, nil — no error, no stories.
	require.NoError(t, err)
	assert.Nil(t, stories)
}

func TestResearchToStoryTransformer_Transform_HappyPathStory(t *testing.T) {
	t.Parallel()

	// Given: a quality-floor-passing result with one happy_path workflow.
	actors := minActors(t)
	entities := minEntities(t)
	wf := happyPathWorkflow(t)
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: 1 story returned.
	require.NoError(t, err)
	require.Len(t, stories, 1)

	story := stories[0]

	// Metadata checks.
	assert.Equal(t, "Place Order", story.Title())
	assert.Equal(t, discoverydomain.StoryTypeCoarseGrained, story.Type())
	assert.Equal(t, discoverydomain.TimeTypeAsIs, story.Time())
	assert.Equal(t, discoverydomain.PurityTypeDigitalized, story.Purity())
	assert.Equal(t, "AI-proposed from domain research", story.Trigger())

	// Sentences: 5 steps, renumbered 1-5.
	sentences := story.Sentences()
	require.Len(t, sentences, 5)

	for i, s := range sentences {
		assert.Equal(t, i+1, s.Step())
	}

	// First sentence maps correctly.
	assert.Equal(t, "Customer", sentences[0].Subject())
	assert.Equal(t, "places", sentences[0].Activity())
	assert.Equal(t, "Order", sentences[0].Object())

	// Validate passes.
	require.NoError(t, story.Validate())
}

func TestResearchToStoryTransformer_Transform_ThreeWorkflows(t *testing.T) {
	t.Parallel()

	// Given: 3 workflows of different types.
	actors := minActors(t)
	entities := minEntities(t)
	hp := happyPathWorkflow(t)

	failSteps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "submits", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "rejects", "Order"),
	}
	fc := mustResearchedWorkflow(t, "Reject Order", discoverydomain.WorkflowTypeFailureCase, failSteps, []string{"https://example.com/wf2"})

	secSteps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Manager", "audits", "Invoice"),
	}
	sec := mustResearchedWorkflow(t, "Audit Invoice", discoverydomain.WorkflowTypeSecondary, secSteps, []string{"https://example.com/wf3"})

	// Provide out of order to verify ordering: secondary, failure_case, happy_path.
	workflows := []discoverydomain.ResearchedWorkflow{sec, fc, hp}
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, workflows)

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: 3 stories in correct order: happy_path, failure_case, secondary.
	require.NoError(t, err)
	require.Len(t, stories, 3)
	assert.Equal(t, "Place Order", stories[0].Title())
	assert.Equal(t, "Reject Order", stories[1].Title())
	assert.Equal(t, "Audit Invoice", stories[2].Title())

	// All validate.
	for _, s := range stories {
		require.NoError(t, s.Validate())
	}
}

func TestResearchToStoryTransformer_Transform_MissingFailureCaseWorkflow(t *testing.T) {
	t.Parallel()

	// Given: only happy_path and secondary workflows (no failure_case).
	actors := minActors(t)
	entities := minEntities(t)
	hp := happyPathWorkflow(t)

	secSteps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Manager", "audits", "Invoice"),
	}
	sec := mustResearchedWorkflow(t, "Audit Invoice", discoverydomain.WorkflowTypeSecondary, secSteps, nil)
	workflows := []discoverydomain.ResearchedWorkflow{hp, sec}
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, workflows)

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: 2 stories — happy_path first, secondary second.
	require.NoError(t, err)
	require.Len(t, stories, 2)
	assert.Equal(t, "Place Order", stories[0].Title())
	assert.Equal(t, "Audit Invoice", stories[1].Title())

	for _, s := range stories {
		require.NoError(t, s.Validate())
	}
}

func TestResearchToStoryTransformer_Transform_ActorWithNoSourceURL(t *testing.T) {
	t.Parallel()

	// Given: actors and entities with NO source URLs → should fallback to "domain research".
	actors := []discoverydomain.ResearchedActor{
		mustResearchedActor(t, "Customer", "buyer", nil),
		mustResearchedActor(t, "Clerk", "employee", nil),
		mustResearchedActor(t, "Manager", "supervisor", nil),
	}
	entities := []discoverydomain.ResearchedEntity{
		mustResearchedEntity(t, "Order", nil, nil),
		mustResearchedEntity(t, "Invoice", nil, nil),
		mustResearchedEntity(t, "Receipt", nil, nil),
	}
	// Workflow with no source URLs.
	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Place Order", discoverydomain.WorkflowTypeHappyPath, steps, nil)
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: stories created with "domain research" as source for actors/objects/sentences.
	require.NoError(t, err)
	require.Len(t, stories, 1)

	story := stories[0]

	// All actors should have "domain research" source.
	for _, a := range story.Actors() {
		assert.Equal(t, "domain research", a.Source(), "actor %s should have fallback source", a.Name())
	}

	// All work objects should have "domain research" source.
	for _, wo := range story.WorkObjects() {
		assert.Equal(t, "domain research", wo.Source(), "work object %s should have fallback source", wo.Name())
	}

	require.NoError(t, story.Validate())
}

func TestResearchToStoryTransformer_Transform_StepActorNotInResearchActors(t *testing.T) {
	t.Parallel()

	// Given: a workflow step references an actor "Auditor" not in research actors.
	actors := minActors(t) // Customer, Clerk, Manager
	entities := minEntities(t)

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "Auditor", "approves", "Invoice"), // Auditor not in actors
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Place Order", discoverydomain.WorkflowTypeHappyPath, steps, []string{"https://example.com/wf"})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: story created with Auditor as minimal actor (source = "domain research"), Validate passes.
	require.NoError(t, err)
	require.Len(t, stories, 1)

	story := stories[0]

	// Find the Auditor actor.
	var foundAuditor bool

	for _, a := range story.Actors() {
		if a.Name() == "Auditor" {
			foundAuditor = true
			assert.Equal(t, "domain research", a.Source())
		}
	}

	assert.True(t, foundAuditor, "Auditor should be present in story actors")
	require.NoError(t, story.Validate())
}

func TestResearchToStoryTransformer_Transform_CaseInsensitiveActorDedup(t *testing.T) {
	t.Parallel()

	// Given: workflow steps reference "Customer" and "customer" (different cases).
	actors := minActors(t)
	entities := minEntities(t)

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "customer", "pays", "Invoice"), // lowercase duplicate
		mustWorkflowStep(t, 4, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Place Order", discoverydomain.WorkflowTypeHappyPath, steps, []string{"https://example.com/wf"})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: only 1 "Customer" actor (not 2).
	require.NoError(t, err)
	require.Len(t, stories, 1)

	story := stories[0]

	customerCount := 0

	for _, a := range story.Actors() {
		if a.Name() == "Customer" {
			customerCount++
		}
	}

	assert.Equal(t, 1, customerCount, "should have exactly 1 Customer actor, not 2")
	require.NoError(t, story.Validate())
}

// --- QA Edge Case Tests ---

func TestResearchToStoryTransformer_QA_QualityFloorFalse_EmptySlice(t *testing.T) {
	t.Parallel()

	// Given: a research result that does NOT meet quality floor.
	meta := discoverydomain.NewSearchMetadata(nil, 1, 0, time.Second)
	actors := []discoverydomain.ResearchedActor{
		mustResearchedActor(t, "User", "user", nil),
	}
	entities := []discoverydomain.ResearchedEntity{
		mustResearchedEntity(t, "Form", nil, nil),
	}
	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "User", "fills", "Form"),
	}
	wf := mustResearchedWorkflow(t, "Submit", discoverydomain.WorkflowTypeHappyPath, steps, nil)
	result := mustDomainResearchResult(t, "test", meta, actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: no error, nil return (quality floor not met → early exit).
	require.NoError(t, err)
	assert.Nil(t, stories, "quality floor not met should return nil")
}

func TestResearchToStoryTransformer_QA_NoWorkflowsFloorMet_EmptySlice(t *testing.T) {
	t.Parallel()

	// Given: quality floor met but zero workflows.
	actors := minActors(t)
	entities := minEntities(t)
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, nil)

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: no error, nil return (no workflows to process).
	require.NoError(t, err)
	assert.Nil(t, stories, "no workflows should return nil")
}

func TestResearchToStoryTransformer_QA_StepRenumbering(t *testing.T) {
	t.Parallel()

	// Given: 5 steps with non-sequential sequence numbers (10, 7, 3, 1, 5).
	actors := minActors(t)
	entities := minEntities(t)

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 10, "Clerk", "issues", "Receipt"),
		mustWorkflowStep(t, 7, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 3, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 5, "Clerk", "generates", "Invoice"),
	}
	wf := mustResearchedWorkflow(t, "Out Of Order", discoverydomain.WorkflowTypeHappyPath, steps, []string{"https://example.com/wf"})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: sentences renumbered sequentially 1..5.
	require.NoError(t, err)
	require.Len(t, stories, 1)

	sentences := stories[0].Sentences()
	require.Len(t, sentences, 5)

	for i, s := range sentences {
		assert.Equal(t, i+1, s.Step(), "sentence %d should have step %d", i, i+1)
	}

	// Verify sorted order by original sequence: 1→"places", 3→"reviews", 5→"generates", 7→"pays", 10→"issues".
	assert.Equal(t, "places", sentences[0].Activity())
	assert.Equal(t, "reviews", sentences[1].Activity())
	assert.Equal(t, "generates", sentences[2].Activity())
	assert.Equal(t, "pays", sentences[3].Activity())
	assert.Equal(t, "issues", sentences[4].Activity())

	require.NoError(t, stories[0].Validate())
}

func TestResearchToStoryTransformer_QA_DuplicateWorkflowTypes_FirstWins(t *testing.T) {
	t.Parallel()

	// Given: two happy_path workflows — first should win.
	actors := minActors(t)
	entities := minEntities(t)

	steps1 := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	first := mustResearchedWorkflow(t, "First HP", discoverydomain.WorkflowTypeHappyPath, steps1, []string{"https://example.com/first"})

	steps2 := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Manager", "audits", "Invoice"),
	}
	second := mustResearchedWorkflow(t, "Second HP", discoverydomain.WorkflowTypeHappyPath, steps2, []string{"https://example.com/second"})

	workflows := []discoverydomain.ResearchedWorkflow{first, second}
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, workflows)

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: only 1 story, title matches first workflow.
	require.NoError(t, err)
	require.Len(t, stories, 1)
	assert.Equal(t, "First HP", stories[0].Title())
	require.NoError(t, stories[0].Validate())
}

func TestResearchToStoryTransformer_QA_SourceURLWhitespaceOnly_Fallback(t *testing.T) {
	t.Parallel()

	// Given: workflow with whitespace-only source URLs.
	actors := minActors(t)
	entities := minEntities(t)

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Whitespace Source", discoverydomain.WorkflowTypeHappyPath, steps, []string{"   ", "  \t  "})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: sentences fall back to "domain research".
	require.NoError(t, err)
	require.Len(t, stories, 1)

	for _, s := range stories[0].Sentences() {
		assert.Equal(t, "domain research", s.Source(), "sentence %d should fallback source", s.Step())
	}

	require.NoError(t, stories[0].Validate())
}

func TestResearchToStoryTransformer_QA_SourceURLEmpty_Fallback(t *testing.T) {
	t.Parallel()

	// Given: workflow with empty string source URLs.
	actors := minActors(t)
	entities := minEntities(t)

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Empty Source", discoverydomain.WorkflowTypeHappyPath, steps, []string{""})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: sentences fall back to "domain research".
	require.NoError(t, err)
	require.Len(t, stories, 1)

	for _, s := range stories[0].Sentences() {
		assert.Equal(t, "domain research", s.Source(), "sentence %d should fallback source", s.Step())
	}

	require.NoError(t, stories[0].Validate())
}

func TestResearchToStoryTransformer_QA_CaseInsensitiveActorDedup(t *testing.T) {
	t.Parallel()

	// Given: steps with "Customer" and "customer" (case variation).
	actors := minActors(t)
	entities := minEntities(t)

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "customer", "reviews", "Order"), // lowercase dup
		mustWorkflowStep(t, 3, "CUSTOMER", "pays", "Invoice"),  // uppercase dup
		mustWorkflowStep(t, 4, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Case Dedup", discoverydomain.WorkflowTypeHappyPath, steps, []string{"https://example.com/wf"})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: exactly 2 actors: first "Customer" variant and "Clerk".
	require.NoError(t, err)
	require.Len(t, stories, 1)

	storyActors := stories[0].Actors()
	assert.Len(t, storyActors, 2, "should have exactly 2 unique actors")

	// First occurrence wins: "Customer" (PascalCase), then "Clerk".
	assert.Equal(t, "Customer", storyActors[0].Name())
	assert.Equal(t, "Clerk", storyActors[1].Name())

	require.NoError(t, stories[0].Validate())
}

func TestResearchToStoryTransformer_QA_StepActorNotInResearch_MinimalActor(t *testing.T) {
	t.Parallel()

	// Given: a step references "Auditor" which is NOT in research actors.
	actors := minActors(t) // Customer, Clerk, Manager
	entities := minEntities(t)

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Auditor", "verifies", "Order"), // unknown actor
		mustWorkflowStep(t, 3, "Clerk", "generates", "Invoice"),
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Unknown Actor", discoverydomain.WorkflowTypeHappyPath, steps, []string{"https://example.com/wf"})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: Auditor exists with "domain research" fallback source.
	require.NoError(t, err)
	require.Len(t, stories, 1)

	var auditor *discoverydomain.StoryActor
	for _, a := range stories[0].Actors() {
		if a.Name() == "Auditor" {
			a := a // copy for pointer
			auditor = &a
		}
	}

	require.NotNil(t, auditor, "Auditor actor must be present")
	assert.Equal(t, "domain research", auditor.Source(), "unknown actor should get fallback source")
	assert.Equal(t, vo.AIResearched, auditor.Trust())

	require.NoError(t, stories[0].Validate())
}

func TestResearchToStoryTransformer_QA_StepWorkObjectNotInResearch_MinimalWorkObject(t *testing.T) {
	t.Parallel()

	// Given: a step references "Report" which is NOT in research entities.
	actors := minActors(t)
	entities := minEntities(t) // Order, Invoice, Receipt

	steps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "places", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "reviews", "Order"),
		mustWorkflowStep(t, 3, "Clerk", "generates", "Report"), // unknown entity
		mustWorkflowStep(t, 4, "Customer", "pays", "Invoice"),
		mustWorkflowStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf := mustResearchedWorkflow(t, "Unknown Object", discoverydomain.WorkflowTypeHappyPath, steps, []string{"https://example.com/wf"})
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: Report exists with "domain research" fallback source.
	require.NoError(t, err)
	require.Len(t, stories, 1)

	var report *discoverydomain.WorkObject
	for _, wo := range stories[0].WorkObjects() {
		if wo.Name() == "Report" {
			wo := wo // copy for pointer
			report = &wo
		}
	}

	require.NotNil(t, report, "Report work object must be present")
	assert.Equal(t, "domain research", report.Source(), "unknown entity should get fallback source")
	assert.Equal(t, vo.AIResearched, report.Trust())

	require.NoError(t, stories[0].Validate())
}

func TestResearchToStoryTransformer_QA_AllTrustAIResearched(t *testing.T) {
	t.Parallel()

	// Given: a standard quality-floor-passing result.
	actors := minActors(t)
	entities := minEntities(t)
	wf := happyPathWorkflow(t)
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, []discoverydomain.ResearchedWorkflow{wf})

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: every actor, work object, and sentence has trust == AIResearched.
	require.NoError(t, err)
	require.Len(t, stories, 1)

	story := stories[0]

	for _, a := range story.Actors() {
		assert.Equal(t, vo.AIResearched, a.Trust(), "actor %q trust should be AIResearched", a.Name())
	}

	for _, wo := range story.WorkObjects() {
		assert.Equal(t, vo.AIResearched, wo.Trust(), "work object %q trust should be AIResearched", wo.Name())
	}

	for _, s := range story.Sentences() {
		assert.Equal(t, vo.AIResearched, s.Trust(), "sentence %d trust should be AIResearched", s.Step())
	}
}

func TestResearchToStoryTransformer_QA_AllStoriesPassValidate(t *testing.T) {
	t.Parallel()

	// Given: all 3 workflow types present.
	actors := minActors(t)
	entities := minEntities(t)
	hp := happyPathWorkflow(t)

	failSteps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Customer", "submits", "Order"),
		mustWorkflowStep(t, 2, "Clerk", "rejects", "Order"),
	}
	fc := mustResearchedWorkflow(t, "Reject Order", discoverydomain.WorkflowTypeFailureCase, failSteps, []string{"https://example.com/wf2"})

	secSteps := []discoverydomain.WorkflowStep{
		mustWorkflowStep(t, 1, "Manager", "audits", "Invoice"),
	}
	sec := mustResearchedWorkflow(t, "Audit Invoice", discoverydomain.WorkflowTypeSecondary, secSteps, []string{"https://example.com/wf3"})

	workflows := []discoverydomain.ResearchedWorkflow{hp, fc, sec}
	result := mustDomainResearchResult(t, "retail", qualityFloorMeta(), actors, entities, workflows)

	// When: transforming.
	transformer := NewResearchToStoryTransformer()
	stories, err := transformer.Transform(context.Background(), &result)

	// Then: all 3 stories pass Validate().
	require.NoError(t, err)
	require.Len(t, stories, 3)

	for _, story := range stories {
		require.NoError(t, story.Validate(), "story %q should pass Validate()", story.Title())

		// Additional structural checks.
		assert.Equal(t, discoverydomain.StoryTypeCoarseGrained, story.Type())
		assert.Equal(t, discoverydomain.TimeTypeAsIs, story.Time())
		assert.Equal(t, discoverydomain.PurityTypeDigitalized, story.Purity())
		assert.Equal(t, "AI-proposed from domain research", story.Trigger())
		assert.NotEmpty(t, story.Actors())
		assert.NotEmpty(t, story.WorkObjects())
		assert.NotEmpty(t, story.Sentences())

		// Verify sequential step numbers on every story.
		for i, s := range story.Sentences() {
			assert.Equal(t, i+1, s.Step(), "story %q sentence %d has wrong step", story.Title(), i)
		}
	}
}
