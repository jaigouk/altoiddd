package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// WorkflowType enum
// ---------------------------------------------------------------------------

func TestNewWorkflowType_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  WorkflowType
	}{
		{"happy_path", "happy_path", WorkflowTypeHappyPath},
		{"failure_case", "failure_case", WorkflowTypeFailureCase},
		{"secondary", "secondary", WorkflowTypeSecondary},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewWorkflowType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewWorkflowType_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"unknown value", "unknown"},
		{"uppercase variant", "HAPPY_PATH"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewWorkflowType(tt.input)
			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestWorkflowType_MarshalText(t *testing.T) {
	t.Parallel()

	wt := WorkflowTypeHappyPath
	data, err := wt.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "happy_path", string(data))
}

func TestWorkflowType_MarshalText_InvalidReturnsError(t *testing.T) {
	t.Parallel()

	wt := WorkflowType("bogus")
	_, err := wt.MarshalText()
	require.Error(t, err)
}

func TestWorkflowType_UnmarshalText(t *testing.T) {
	t.Parallel()

	var wt WorkflowType
	err := wt.UnmarshalText([]byte("failure_case"))
	require.NoError(t, err)
	assert.Equal(t, WorkflowTypeFailureCase, wt)
}

func TestWorkflowType_UnmarshalText_InvalidReturnsError(t *testing.T) {
	t.Parallel()

	var wt WorkflowType
	err := wt.UnmarshalText([]byte("invalid"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllWorkflowTypes(t *testing.T) {
	t.Parallel()

	all := AllWorkflowTypes()
	assert.Len(t, all, 3)
	assert.Contains(t, all, WorkflowTypeHappyPath)
	assert.Contains(t, all, WorkflowTypeFailureCase)
	assert.Contains(t, all, WorkflowTypeSecondary)
}

func TestWorkflowType_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "happy_path", WorkflowTypeHappyPath.String())
}

// ---------------------------------------------------------------------------
// SearchMetadata
// ---------------------------------------------------------------------------

func TestNewSearchMetadata(t *testing.T) {
	t.Parallel()

	queries := []string{"q1", "q2"}
	meta := NewSearchMetadata(queries, 10, 5, 3*time.Second)

	assert.Equal(t, queries, meta.QueriesUsed())
	assert.Equal(t, 10, meta.TotalSources())
	assert.Equal(t, 5, meta.UsefulSources())
	assert.Equal(t, 3*time.Second, meta.Duration())
}

func TestNewSearchMetadata_ZeroValue(t *testing.T) {
	t.Parallel()

	// No error return — zero-value is valid.
	meta := NewSearchMetadata(nil, 0, 0, 0)

	assert.Empty(t, meta.QueriesUsed())
	assert.Equal(t, 0, meta.TotalSources())
	assert.Equal(t, 0, meta.UsefulSources())
	assert.Equal(t, time.Duration(0), meta.Duration())
}

func TestSearchMetadata_DefensiveCopyQueries(t *testing.T) {
	t.Parallel()

	queries := []string{"q1", "q2"}
	meta := NewSearchMetadata(queries, 10, 5, time.Second)

	// Mutate original — VO must not be affected.
	queries[0] = "mutated"
	assert.Equal(t, "q1", meta.QueriesUsed()[0])

	// Mutate returned slice — VO must not be affected.
	returned := meta.QueriesUsed()
	returned[0] = "mutated_again"
	assert.Equal(t, "q1", meta.QueriesUsed()[0])
}

// ---------------------------------------------------------------------------
// ResearchedActor
// ---------------------------------------------------------------------------

func TestNewResearchedActor_Valid(t *testing.T) {
	t.Parallel()

	actor, err := NewResearchedActor("Veterinarian", "Primary care provider", []string{"https://example.com"})
	require.NoError(t, err)
	assert.Equal(t, "Veterinarian", actor.Name())
	assert.Equal(t, "Primary care provider", actor.Role())
	assert.Len(t, actor.SourceURLs(), 1)
}

func TestNewResearchedActor_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := NewResearchedActor("", "some role", nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewResearchedActor_WhitespaceName(t *testing.T) {
	t.Parallel()

	_, err := NewResearchedActor("   ", "some role", nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestResearchedActor_Getters(t *testing.T) {
	t.Parallel()

	urls := []string{"https://a.com", "https://b.com"}
	actor, err := NewResearchedActor("  Vet  ", "Provider", urls)
	require.NoError(t, err)

	// Name is trimmed.
	assert.Equal(t, "Vet", actor.Name())
	assert.Equal(t, "Provider", actor.Role())
	assert.Equal(t, urls, actor.SourceURLs())
}

func TestResearchedActor_DefensiveCopySourceURLs(t *testing.T) {
	t.Parallel()

	urls := []string{"https://a.com"}
	actor, err := NewResearchedActor("Vet", "Provider", urls)
	require.NoError(t, err)

	// Mutate original.
	urls[0] = "mutated"
	assert.Equal(t, "https://a.com", actor.SourceURLs()[0])

	// Mutate returned slice.
	returned := actor.SourceURLs()
	returned[0] = "mutated_again"
	assert.Equal(t, "https://a.com", actor.SourceURLs()[0])
}

func TestNewResearchedActor_NilSourceURLs(t *testing.T) {
	t.Parallel()

	actor, err := NewResearchedActor("Vet", "Provider", nil)
	require.NoError(t, err)
	assert.Empty(t, actor.SourceURLs())
}

// ---------------------------------------------------------------------------
// ResearchedEntity
// ---------------------------------------------------------------------------

func TestNewResearchedEntity_Valid(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Patient Record", []string{"name", "dob"}, []string{"https://example.com"})
	require.NoError(t, err)
	assert.Equal(t, "Patient Record", entity.Name())
	assert.Equal(t, []string{"name", "dob"}, entity.Properties())
	assert.Len(t, entity.SourceURLs(), 1)
}

func TestNewResearchedEntity_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := NewResearchedEntity("", nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewResearchedEntity_WhitespaceName(t *testing.T) {
	t.Parallel()

	_, err := NewResearchedEntity("  \t  ", nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestResearchedEntity_DefensiveCopies(t *testing.T) {
	t.Parallel()

	props := []string{"name"}
	urls := []string{"https://a.com"}
	entity, err := NewResearchedEntity("Record", props, urls)
	require.NoError(t, err)

	props[0] = "mutated"
	urls[0] = "mutated"

	assert.Equal(t, "name", entity.Properties()[0])
	assert.Equal(t, "https://a.com", entity.SourceURLs()[0])
}

// ---------------------------------------------------------------------------
// WorkflowStep
// ---------------------------------------------------------------------------

func TestNewWorkflowStep_Valid(t *testing.T) {
	t.Parallel()

	step, err := NewWorkflowStep(1, "Vet", "examines", "Patient")
	require.NoError(t, err)
	assert.Equal(t, 1, step.Sequence())
	assert.Equal(t, "Vet", step.Actor())
	assert.Equal(t, "examines", step.Activity())
	assert.Equal(t, "Patient", step.WorkObject())
}

func TestNewWorkflowStep_EmptyActor(t *testing.T) {
	t.Parallel()

	_, err := NewWorkflowStep(1, "", "examines", "Patient")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkflowStep_EmptyActivity(t *testing.T) {
	t.Parallel()

	_, err := NewWorkflowStep(1, "Vet", "", "Patient")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkflowStep_EmptyWorkObject(t *testing.T) {
	t.Parallel()

	_, err := NewWorkflowStep(1, "Vet", "examines", "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkflowStep_NegativeSequence(t *testing.T) {
	t.Parallel()

	// Negative sequence is allowed — it is informational.
	step, err := NewWorkflowStep(-1, "Vet", "examines", "Patient")
	require.NoError(t, err)
	assert.Equal(t, -1, step.Sequence())
}

func TestNewWorkflowStep_WhitespaceFields(t *testing.T) {
	t.Parallel()

	_, err := NewWorkflowStep(1, "  ", "examines", "Patient")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// ---------------------------------------------------------------------------
// ResearchedWorkflow
// ---------------------------------------------------------------------------

func TestNewResearchedWorkflow_Valid(t *testing.T) {
	t.Parallel()

	step, err := NewWorkflowStep(1, "Vet", "examines", "Patient")
	require.NoError(t, err)

	wf, err := NewResearchedWorkflow("Appointment Flow", WorkflowTypeHappyPath, []WorkflowStep{step}, []string{"https://a.com"})
	require.NoError(t, err)
	assert.Equal(t, "Appointment Flow", wf.Name())
	assert.Equal(t, WorkflowTypeHappyPath, wf.WorkflowType())
	assert.Len(t, wf.Steps(), 1)
	assert.Len(t, wf.SourceURLs(), 1)
}

func TestNewResearchedWorkflow_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := NewResearchedWorkflow("", WorkflowTypeHappyPath, nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewResearchedWorkflow_InvalidType(t *testing.T) {
	t.Parallel()

	_, err := NewResearchedWorkflow("Flow", WorkflowType("invalid"), nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewResearchedWorkflow_EmptySteps(t *testing.T) {
	t.Parallel()

	// Zero-length steps is valid.
	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeSecondary, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, wf.Steps())
}

func TestResearchedWorkflow_DefensiveCopySteps(t *testing.T) {
	t.Parallel()

	step, err := NewWorkflowStep(1, "Vet", "examines", "Patient")
	require.NoError(t, err)

	steps := []WorkflowStep{step}
	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, steps, nil)
	require.NoError(t, err)

	// Mutate original slice — VO must not be affected.
	step2, err := NewWorkflowStep(2, "Owner", "pays", "Invoice")
	require.NoError(t, err)

	steps[0] = step2
	assert.Equal(t, "Vet", wf.Steps()[0].Actor())
}

func TestResearchedWorkflow_DefensiveCopySourceURLs(t *testing.T) {
	t.Parallel()

	urls := []string{"https://a.com"}
	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, nil, urls)
	require.NoError(t, err)

	urls[0] = "mutated"
	assert.Equal(t, "https://a.com", wf.SourceURLs()[0])
}

// ---------------------------------------------------------------------------
// RegulatoryItem
// ---------------------------------------------------------------------------

func TestNewRegulatoryItem_Valid(t *testing.T) {
	t.Parallel()

	item, err := NewRegulatoryItem("HIPAA", "Health data privacy", []string{"https://hhs.gov"})
	require.NoError(t, err)
	assert.Equal(t, "HIPAA", item.Name())
	assert.Equal(t, "Health data privacy", item.Description())
	assert.Len(t, item.SourceURLs(), 1)
}

func TestNewRegulatoryItem_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := NewRegulatoryItem("", "desc", nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewRegulatoryItem_WhitespaceName(t *testing.T) {
	t.Parallel()

	_, err := NewRegulatoryItem("  \n  ", "desc", nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestRegulatoryItem_DefensiveCopySourceURLs(t *testing.T) {
	t.Parallel()

	urls := []string{"https://a.com"}
	item, err := NewRegulatoryItem("HIPAA", "desc", urls)
	require.NoError(t, err)

	urls[0] = "mutated"
	assert.Equal(t, "https://a.com", item.SourceURLs()[0])
}

// ---------------------------------------------------------------------------
// ExistingSoftware
// ---------------------------------------------------------------------------

func TestNewExistingSoftware_Valid(t *testing.T) {
	t.Parallel()

	sw, err := NewExistingSoftware("Jira", "Project tracker", "https://jira.com")
	require.NoError(t, err)
	assert.Equal(t, "Jira", sw.Name())
	assert.Equal(t, "Project tracker", sw.Description())
	assert.Equal(t, "https://jira.com", sw.SourceURL())
}

func TestNewExistingSoftware_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := NewExistingSoftware("", "desc", "https://example.com")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewExistingSoftware_WhitespaceName(t *testing.T) {
	t.Parallel()

	_, err := NewExistingSoftware("  ", "desc", "url")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestExistingSoftware_EmptySourceURL(t *testing.T) {
	t.Parallel()

	// Empty source URL is valid — not all software has a URL.
	sw, err := NewExistingSoftware("Jira", "tracker", "")
	require.NoError(t, err)
	assert.Empty(t, sw.SourceURL())
}

// ---------------------------------------------------------------------------
// ComputeResearchQuality
// ---------------------------------------------------------------------------

func TestComputeResearchQuality_MeetsFloor(t *testing.T) {
	t.Parallel()

	// Given: exactly at thresholds — 3 actors, 3 entities, 5+ workflow steps, 5 sources.
	actors := makeActors(3)
	entities := makeEntities(3)
	workflows := makeWorkflowsWithTotalSteps(5)

	quality := ComputeResearchQuality(actors, entities, workflows, 5)

	assert.Equal(t, 3, quality.ActorCount())
	assert.Equal(t, 3, quality.EntityCount())
	assert.Equal(t, 5, quality.WorkflowStepCount())
	assert.Equal(t, 5, quality.UsefulSourceCount())
	assert.True(t, quality.MeetsFloor())
}

func TestComputeResearchQuality_BelowFloor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		actorCount     int
		entityCount    int
		totalSteps     int
		usefulSources  int
		expectMeetsFlr bool
	}{
		{"insufficient actors", 2, 3, 5, 5, false},
		{"insufficient entities", 3, 2, 5, 5, false},
		{"insufficient steps", 3, 3, 4, 5, false},
		{"insufficient sources", 3, 3, 5, 4, false},
		{"all below", 0, 0, 0, 0, false},
		{"above thresholds", 5, 5, 10, 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actors := makeActors(tt.actorCount)
			entities := makeEntities(tt.entityCount)
			workflows := makeWorkflowsWithTotalSteps(tt.totalSteps)

			quality := ComputeResearchQuality(actors, entities, workflows, tt.usefulSources)
			assert.Equal(t, tt.expectMeetsFlr, quality.MeetsFloor())
		})
	}
}

func TestComputeResearchQuality_NilSlices(t *testing.T) {
	t.Parallel()

	quality := ComputeResearchQuality(nil, nil, nil, 0)
	assert.Equal(t, 0, quality.ActorCount())
	assert.Equal(t, 0, quality.EntityCount())
	assert.Equal(t, 0, quality.WorkflowStepCount())
	assert.Equal(t, 0, quality.UsefulSourceCount())
	assert.False(t, quality.MeetsFloor())
}

// ---------------------------------------------------------------------------
// DomainResearchResult
// ---------------------------------------------------------------------------

func TestNewDomainResearchResult_Valid(t *testing.T) {
	t.Parallel()

	actors := makeActors(3)
	entities := makeEntities(3)
	workflows := makeWorkflowsWithTotalSteps(5)
	meta := NewSearchMetadata([]string{"q1"}, 10, 5, time.Second)
	regulatory := makeRegulatoryItems(1)
	software := makeSoftwareItems(1)

	result, err := NewDomainResearchResult(
		"Veterinary",
		meta,
		actors,
		entities,
		workflows,
		[]string{"power outage"},
		regulatory,
		software,
	)
	require.NoError(t, err)
	assert.Equal(t, "Veterinary", result.Domain())
	assert.Len(t, result.Actors(), 3)
	assert.Len(t, result.Entities(), 3)
	assert.NotEmpty(t, result.Workflows())
	assert.Equal(t, []string{"power outage"}, result.FailureModes())
	assert.Len(t, result.Regulatory(), 1)
	assert.Len(t, result.Software(), 1)
}

func TestNewDomainResearchResult_EmptyDomain(t *testing.T) {
	t.Parallel()

	_, err := NewDomainResearchResult("", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewDomainResearchResult_WhitespaceDomain(t *testing.T) {
	t.Parallel()

	_, err := NewDomainResearchResult("   ", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewDomainResearchResult_TrimsDomain(t *testing.T) {
	t.Parallel()

	result, err := NewDomainResearchResult("  Vet  ", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Vet", result.Domain())
}

func TestDomainResearchResult_Quality(t *testing.T) {
	t.Parallel()

	// Given: meets floor (3 actors, 3 entities, 5 steps, 5 useful sources).
	actors := makeActors(3)
	entities := makeEntities(3)
	workflows := makeWorkflowsWithTotalSteps(5)
	meta := NewSearchMetadata(nil, 10, 5, 0)

	result, err := NewDomainResearchResult("Vet", meta, actors, entities, workflows, nil, nil, nil)
	require.NoError(t, err)

	q := result.Quality()
	assert.True(t, q.MeetsFloor())
	assert.Equal(t, 3, q.ActorCount())
	assert.Equal(t, 3, q.EntityCount())
	assert.Equal(t, 5, q.WorkflowStepCount())
	assert.Equal(t, 5, q.UsefulSourceCount())
}

func TestDomainResearchResult_DefensiveCopyActors(t *testing.T) {
	t.Parallel()

	actors := makeActors(2)
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), actors, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	// Mutate original.
	actors[0], _ = NewResearchedActor("Mutated", "role", nil)
	assert.NotEqual(t, "Mutated", result.Actors()[0].Name())

	// Mutate returned slice.
	returned := result.Actors()
	returned[0], _ = NewResearchedActor("Mutated2", "role", nil)
	assert.NotEqual(t, "Mutated2", result.Actors()[0].Name())
}

func TestDomainResearchResult_DefensiveCopyWorkflows(t *testing.T) {
	t.Parallel()

	workflows := makeWorkflowsWithTotalSteps(3)
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, workflows, nil, nil, nil)
	require.NoError(t, err)

	original := result.Workflows()
	originalName := original[0].Name()

	// Mutate returned slice element.
	wf, _ := NewResearchedWorkflow("Mutated", WorkflowTypeSecondary, nil, nil)
	returned := result.Workflows()
	returned[0] = wf
	assert.Equal(t, originalName, result.Workflows()[0].Name())
}

func TestDomainResearchResult_DefensiveCopyFailureModes(t *testing.T) {
	t.Parallel()

	modes := []string{"power outage", "data loss"}
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, modes, nil, nil)
	require.NoError(t, err)

	modes[0] = "mutated"
	assert.Equal(t, "power outage", result.FailureModes()[0])
}

func TestDomainResearchResult_SearchMetadata(t *testing.T) {
	t.Parallel()

	meta := NewSearchMetadata([]string{"q1"}, 10, 5, 2*time.Second)
	result, err := NewDomainResearchResult("Vet", meta, nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	assert.Equal(t, meta.TotalSources(), result.SearchMetadata().TotalSources())
	assert.Equal(t, meta.UsefulSources(), result.SearchMetadata().UsefulSources())
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func makeActors(n int) []ResearchedActor {
	actors := make([]ResearchedActor, n)
	names := []string{"Vet", "Owner", "Technician", "Receptionist", "Manager"}

	for i := range n {
		name := names[i%len(names)]
		a, _ := NewResearchedActor(name, "role", []string{"https://example.com"})
		actors[i] = a
	}

	return actors
}

func makeEntities(n int) []ResearchedEntity {
	entities := make([]ResearchedEntity, n)
	names := []string{"Patient", "Appointment", "Invoice", "Medication", "Record"}

	for i := range n {
		name := names[i%len(names)]
		e, _ := NewResearchedEntity(name, []string{"prop"}, []string{"https://example.com"})
		entities[i] = e
	}

	return entities
}

func makeWorkflowsWithTotalSteps(totalSteps int) []ResearchedWorkflow {
	if totalSteps == 0 {
		return nil
	}

	steps := make([]WorkflowStep, totalSteps)
	for i := range totalSteps {
		s, _ := NewWorkflowStep(i+1, "Actor", "does", "Thing")
		steps[i] = s
	}

	wf, _ := NewResearchedWorkflow("Main Flow", WorkflowTypeHappyPath, steps, nil)

	return []ResearchedWorkflow{wf}
}

func makeRegulatoryItems(n int) []RegulatoryItem {
	items := make([]RegulatoryItem, n)

	for i := range n {
		item, _ := NewRegulatoryItem("Regulation", "desc", nil)
		items[i] = item
	}

	return items
}

func makeSoftwareItems(n int) []ExistingSoftware {
	items := make([]ExistingSoftware, n)

	for i := range n {
		item, _ := NewExistingSoftware("Software", "desc", "https://example.com")
		items[i] = item
	}

	return items
}

// ===========================================================================
// Invariant edge-case tests (QA — alty-cli-1wu.16)
// ===========================================================================

// ---------------------------------------------------------------------------
// 1. Whitespace-only name rejection across all constructors with name fields
// ---------------------------------------------------------------------------

func TestResearchedActor_Invariant_WhitespaceOnlyNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"tab only", "\t"},
		{"newline only", "\n"},
		{"mixed whitespace", "  \t\n  "},
		{"carriage return", "\r"},
		{"tab and spaces", "\t   \t"},
		{"multiple newlines", "\n\n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewResearchedActor(tt.input, "role", nil)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestResearchedEntity_Invariant_WhitespaceOnlyNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"tab only", "\t"},
		{"newline only", "\n"},
		{"mixed whitespace", "  \t\n  "},
		{"carriage return", "\r"},
		{"tab and spaces", "\t   \t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewResearchedEntity(tt.input, nil, nil)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestResearchedWorkflow_Invariant_WhitespaceOnlyNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"tab only", "\t"},
		{"newline only", "\n"},
		{"mixed whitespace", "  \t\n  "},
		{"carriage return", "\r"},
		{"tab and spaces", "\t   \t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewResearchedWorkflow(tt.input, WorkflowTypeHappyPath, nil, nil)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestRegulatoryItem_Invariant_WhitespaceOnlyNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"tab only", "\t"},
		{"newline only", "\n"},
		{"mixed whitespace", "  \t\n  "},
		{"carriage return", "\r"},
		{"tab and spaces", "\t   \t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewRegulatoryItem(tt.input, "desc", nil)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestExistingSoftware_Invariant_WhitespaceOnlyNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"tab only", "\t"},
		{"newline only", "\n"},
		{"mixed whitespace", "  \t\n  "},
		{"carriage return", "\r"},
		{"tab and spaces", "\t   \t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewExistingSoftware(tt.input, "desc", "")
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

// ---------------------------------------------------------------------------
// 2. WorkflowType edge cases
// ---------------------------------------------------------------------------

func TestWorkflowType_Invariant_LeadingTrailingSpaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"leading space", " happy_path"},
		{"trailing space", "happy_path "},
		{"both sides", " happy_path "},
		{"leading tab", "\thappy_path"},
		{"trailing newline", "happy_path\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewWorkflowType(tt.input)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation,
				"WorkflowType must NOT trim — %q should be rejected as-is", tt.input)
		})
	}
}

func TestWorkflowType_Invariant_CaseSensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"all caps", "HAPPY_PATH"},
		{"title case", "Happy_Path"},
		{"mixed case failure_case", "Failure_Case"},
		{"upper secondary", "SECONDARY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewWorkflowType(tt.input)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestWorkflowType_Invariant_MarshalUnmarshalRoundTrip_AllValues(t *testing.T) {
	t.Parallel()

	for _, wt := range AllWorkflowTypes() {
		t.Run(wt.String(), func(t *testing.T) {
			t.Parallel()

			data, err := wt.MarshalText()
			require.NoError(t, err)

			var unmarshaled WorkflowType
			err = unmarshaled.UnmarshalText(data)
			require.NoError(t, err)

			assert.Equal(t, wt, unmarshaled)
		})
	}
}

func TestWorkflowType_Invariant_AllWorkflowTypesDistinct(t *testing.T) {
	t.Parallel()

	all := AllWorkflowTypes()
	assert.Len(t, all, 3)

	seen := make(map[WorkflowType]struct{}, len(all))
	for _, wt := range all {
		_, duplicate := seen[wt]
		assert.False(t, duplicate, "duplicate WorkflowType: %s", wt)
		seen[wt] = struct{}{}
	}
}

// ---------------------------------------------------------------------------
// 3. WorkflowStep independent field rejection
// ---------------------------------------------------------------------------

func TestWorkflowStep_Invariant_IndependentFieldRejection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		actor      string
		activity   string
		workObject string
	}{
		{"empty actor only", "", "examines", "Patient"},
		{"empty activity only", "Vet", "", "Patient"},
		{"empty workObject only", "Vet", "examines", ""},
		{"whitespace tab actor", "\t", "examines", "Patient"},
		{"whitespace spaces activity", "Vet", "  ", "Patient"},
		{"whitespace newline workObject", "Vet", "examines", "\n"},
		{"whitespace mixed actor", "  \t\n  ", "examines", "Patient"},
		{"whitespace mixed activity", "Vet", "\r\n", "Patient"},
		{"whitespace mixed workObject", "Vet", "examines", "\t \n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewWorkflowStep(1, tt.actor, tt.activity, tt.workObject)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestWorkflowStep_Invariant_AllThreeEmptyProducesSingleError(t *testing.T) {
	t.Parallel()

	// When all three fields are empty, the constructor should still return
	// exactly one error (it short-circuits on the first invalid field).
	_, err := NewWorkflowStep(1, "", "", "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)

	// Verify the error message references "actor" (the first checked field),
	// confirming short-circuit evaluation rather than accumulation.
	assert.Contains(t, err.Error(), "actor")
}

// ---------------------------------------------------------------------------
// 4. DomainResearchResult domain validation
// ---------------------------------------------------------------------------

func TestDomainResearchResult_Invariant_WhitespaceDomainVariants(t *testing.T) {
	t.Parallel()

	meta := NewSearchMetadata(nil, 0, 0, 0)

	tests := []struct {
		name   string
		domain string
	}{
		{"tab only", "\t"},
		{"newline only", "\n"},
		{"mixed whitespace", "  \t\n  "},
		{"carriage return", "\r"},
		{"carriage return + newline", "\r\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewDomainResearchResult(tt.domain, meta, nil, nil, nil, nil, nil, nil)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestDomainResearchResult_Invariant_TrimsLeadingTrailingSpaces(t *testing.T) {
	t.Parallel()

	meta := NewSearchMetadata(nil, 0, 0, 0)

	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		{"leading and trailing spaces", "  healthcare  ", "healthcare"},
		{"leading tab", "\thealthcare", "healthcare"},
		{"trailing newline", "healthcare\n", "healthcare"},
		{"mixed surrounding whitespace", "\t  healthcare  \n", "healthcare"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := NewDomainResearchResult(tt.domain, meta, nil, nil, nil, nil, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Domain())
		})
	}
}

// ---------------------------------------------------------------------------
// 5. SearchMetadata zero-value acceptance
// ---------------------------------------------------------------------------

func TestSearchMetadata_Invariant_ZeroValueNoError(t *testing.T) {
	t.Parallel()

	// NewSearchMetadata has no error return. Zero-value must not panic.
	meta := NewSearchMetadata(nil, 0, 0, 0)

	// All getters must work on zero-value instance.
	assert.Empty(t, meta.QueriesUsed())
	assert.Equal(t, 0, meta.TotalSources())
	assert.Equal(t, 0, meta.UsefulSources())
	assert.Equal(t, time.Duration(0), meta.Duration())
}

func TestSearchMetadata_Invariant_NilQueriesReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	meta := NewSearchMetadata(nil, 0, 0, 0)

	// QueriesUsed must return a non-nil empty slice (not nil).
	queries := meta.QueriesUsed()
	assert.NotNil(t, queries)
	assert.Empty(t, queries)
}

// ---------------------------------------------------------------------------
// 6. ExistingSoftware single sourceURL
// ---------------------------------------------------------------------------

func TestExistingSoftware_Invariant_SingleStringSourceURL(t *testing.T) {
	t.Parallel()

	// Verify constructor takes string (not []string) — compiler enforces this,
	// but we verify the getter returns the exact value passed.
	sw, err := NewExistingSoftware("Jira", "tracker", "https://jira.com")
	require.NoError(t, err)
	assert.Equal(t, "https://jira.com", sw.SourceURL())
}

func TestExistingSoftware_Invariant_EmptySourceURLValid(t *testing.T) {
	t.Parallel()

	sw, err := NewExistingSoftware("Jira", "tracker", "")
	require.NoError(t, err)
	assert.Empty(t, sw.SourceURL())
}

func TestExistingSoftware_Invariant_NonEmptySourceURL(t *testing.T) {
	t.Parallel()

	sw, err := NewExistingSoftware("Jira", "tracker", "https://atlassian.com/jira")
	require.NoError(t, err)
	assert.Equal(t, "https://atlassian.com/jira", sw.SourceURL())
}

// ---------------------------------------------------------------------------
// 7. Error wrapping verification — every constructor that returns error
// ---------------------------------------------------------------------------

func TestErrorWrapping_Invariant_AllConstructorsWrapErrInvariantViolation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"NewResearchedActor empty name", func() error {
			_, err := NewResearchedActor("", "role", nil)
			return err
		}},
		{"NewResearchedEntity empty name", func() error {
			_, err := NewResearchedEntity("", nil, nil)
			return err
		}},
		{"NewResearchedWorkflow empty name", func() error {
			_, err := NewResearchedWorkflow("", WorkflowTypeHappyPath, nil, nil)
			return err
		}},
		{"NewResearchedWorkflow invalid type", func() error {
			_, err := NewResearchedWorkflow("Flow", WorkflowType("bogus"), nil, nil)
			return err
		}},
		{"NewRegulatoryItem empty name", func() error {
			_, err := NewRegulatoryItem("", "desc", nil)
			return err
		}},
		{"NewExistingSoftware empty name", func() error {
			_, err := NewExistingSoftware("", "desc", "")
			return err
		}},
		{"NewWorkflowStep empty actor", func() error {
			_, err := NewWorkflowStep(1, "", "act", "obj")
			return err
		}},
		{"NewWorkflowStep empty activity", func() error {
			_, err := NewWorkflowStep(1, "actor", "", "obj")
			return err
		}},
		{"NewWorkflowStep empty workObject", func() error {
			_, err := NewWorkflowStep(1, "actor", "act", "")
			return err
		}},
		{"NewDomainResearchResult empty domain", func() error {
			_, err := NewDomainResearchResult("", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, nil)
			return err
		}},
		{"NewWorkflowType invalid", func() error {
			_, err := NewWorkflowType("invalid")
			return err
		}},
		{"WorkflowType.Validate invalid", func() error {
			wt := WorkflowType("bogus")
			return wt.Validate()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fn()
			require.Error(t, err)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation,
				"error from %s must wrap ErrInvariantViolation", tt.name)
		})
	}
}

// ===========================================================================
// Quality threshold tests (QA — alty-cli-1wu.16)
// ===========================================================================

// ---------------------------------------------------------------------------
// Quality threshold test helpers
// ---------------------------------------------------------------------------

// makeQualityActors builds n ResearchedActors with unique names for quality threshold tests.
func makeQualityActors(t *testing.T, count int) []ResearchedActor {
	t.Helper()

	names := []string{
		"Admin", "Buyer", "Clerk", "Driver", "Editor",
		"Foreman", "Guard", "Host", "Inspector", "Judge",
	}
	actors := make([]ResearchedActor, count)

	for i := range count {
		a, err := NewResearchedActor(names[i%len(names)], "role", nil)
		require.NoError(t, err)

		actors[i] = a
	}

	return actors
}

// makeQualityEntities builds n ResearchedEntities with unique names for quality threshold tests.
func makeQualityEntities(t *testing.T, count int) []ResearchedEntity {
	t.Helper()

	names := []string{
		"Order", "Product", "Customer", "Invoice", "Shipment",
		"Warehouse", "Ticket", "Report", "Account", "Contract",
	}
	entities := make([]ResearchedEntity, count)

	for i := range count {
		e, err := NewResearchedEntity(names[i%len(names)], []string{"prop"}, nil)
		require.NoError(t, err)

		entities[i] = e
	}

	return entities
}

// makeQualityWorkflow builds a single ResearchedWorkflow with the given number of steps.
func makeQualityWorkflow(t *testing.T, name string, stepCount int) ResearchedWorkflow {
	t.Helper()

	steps := make([]WorkflowStep, stepCount)

	for i := range stepCount {
		s, err := NewWorkflowStep(i+1, "Actor", "does", "Thing")
		require.NoError(t, err)

		steps[i] = s
	}

	wf, err := NewResearchedWorkflow(name, WorkflowTypeHappyPath, steps, nil)
	require.NoError(t, err)

	return wf
}

// ---------------------------------------------------------------------------
// 1. Exact threshold boundary
// ---------------------------------------------------------------------------

func TestQuality_ExactThresholdBoundary(t *testing.T) {
	t.Parallel()

	actors := makeQualityActors(t, 3)
	entities := makeQualityEntities(t, 3)
	wf := makeQualityWorkflow(t, "MainFlow", 5)

	quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wf}, 5)

	assert.Equal(t, 3, quality.ActorCount())
	assert.Equal(t, 3, quality.EntityCount())
	assert.Equal(t, 5, quality.WorkflowStepCount())
	assert.Equal(t, 5, quality.UsefulSourceCount())
	assert.True(t, quality.MeetsFloor())
}

// ---------------------------------------------------------------------------
// 2. One-below-each threshold (4 separate subtests)
// ---------------------------------------------------------------------------

func TestQuality_OneBelowEachThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		actors        int
		entities      int
		steps         int
		usefulSources int
	}{
		{"actors one below", 2, 3, 5, 5},
		{"entities one below", 3, 2, 5, 5},
		{"steps one below", 3, 3, 4, 5},
		{"sources one below", 3, 3, 5, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actors := makeQualityActors(t, tt.actors)
			entities := makeQualityEntities(t, tt.entities)

			var workflows []ResearchedWorkflow
			if tt.steps > 0 {
				wf := makeQualityWorkflow(t, "Flow", tt.steps)
				workflows = []ResearchedWorkflow{wf}
			}

			quality := ComputeResearchQuality(actors, entities, workflows, tt.usefulSources)
			assert.False(t, quality.MeetsFloor())
		})
	}
}

// ---------------------------------------------------------------------------
// 3. All zeros
// ---------------------------------------------------------------------------

func TestQuality_AllZeros(t *testing.T) {
	t.Parallel()

	quality := ComputeResearchQuality(nil, nil, nil, 0)

	assert.Equal(t, 0, quality.ActorCount())
	assert.Equal(t, 0, quality.EntityCount())
	assert.Equal(t, 0, quality.WorkflowStepCount())
	assert.Equal(t, 0, quality.UsefulSourceCount())
	assert.False(t, quality.MeetsFloor())
}

// ---------------------------------------------------------------------------
// 4. Well above all thresholds
// ---------------------------------------------------------------------------

func TestQuality_WellAboveAllThresholds(t *testing.T) {
	t.Parallel()

	actors := makeQualityActors(t, 10)
	entities := makeQualityEntities(t, 10)

	// 4 workflows x 5 steps = 20 total steps.
	workflows := make([]ResearchedWorkflow, 4)

	for i := range 4 {
		workflows[i] = makeQualityWorkflow(t, "Flow"+string(rune('A'+i)), 5)
	}

	quality := ComputeResearchQuality(actors, entities, workflows, 30)

	assert.Equal(t, 10, quality.ActorCount())
	assert.Equal(t, 10, quality.EntityCount())
	assert.Equal(t, 20, quality.WorkflowStepCount())
	assert.Equal(t, 30, quality.UsefulSourceCount())
	assert.True(t, quality.MeetsFloor())
}

// ---------------------------------------------------------------------------
// 5. Steps across multiple workflows
// ---------------------------------------------------------------------------

func TestQuality_StepsAcrossMultipleWorkflows(t *testing.T) {
	t.Parallel()

	t.Run("two workflows summing above threshold", func(t *testing.T) {
		t.Parallel()

		actors := makeQualityActors(t, 3)
		entities := makeQualityEntities(t, 3)
		wfA := makeQualityWorkflow(t, "FlowA", 3)
		wfB := makeQualityWorkflow(t, "FlowB", 3)

		quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wfA, wfB}, 5)

		assert.Equal(t, 6, quality.WorkflowStepCount())
		assert.True(t, quality.MeetsFloor())
	})

	t.Run("two workflows summing below threshold", func(t *testing.T) {
		t.Parallel()

		actors := makeQualityActors(t, 3)
		entities := makeQualityEntities(t, 3)
		wfA := makeQualityWorkflow(t, "FlowA", 2)
		wfB := makeQualityWorkflow(t, "FlowB", 2)

		quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wfA, wfB}, 5)

		assert.Equal(t, 4, quality.WorkflowStepCount())
		assert.False(t, quality.MeetsFloor())
	})
}

// ---------------------------------------------------------------------------
// 6. Empty/nil workflows
// ---------------------------------------------------------------------------

func TestQuality_EmptyAndNilWorkflows(t *testing.T) {
	t.Parallel()

	t.Run("nil workflows slice", func(t *testing.T) {
		t.Parallel()

		actors := makeQualityActors(t, 3)
		entities := makeQualityEntities(t, 3)

		quality := ComputeResearchQuality(actors, entities, nil, 5)

		assert.Equal(t, 0, quality.WorkflowStepCount())
		assert.False(t, quality.MeetsFloor())
	})

	t.Run("empty workflows slice", func(t *testing.T) {
		t.Parallel()

		actors := makeQualityActors(t, 3)
		entities := makeQualityEntities(t, 3)

		quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{}, 5)

		assert.Equal(t, 0, quality.WorkflowStepCount())
		assert.False(t, quality.MeetsFloor())
	})
}

// ---------------------------------------------------------------------------
// 7. ResearchQuality VO getters
// ---------------------------------------------------------------------------

func TestQuality_VOGettersReturnCorrectValues(t *testing.T) {
	t.Parallel()

	actors := makeQualityActors(t, 7)
	entities := makeQualityEntities(t, 4)
	wf := makeQualityWorkflow(t, "Flow", 9)

	quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wf}, 12)

	assert.Equal(t, 7, quality.ActorCount())
	assert.Equal(t, 4, quality.EntityCount())
	assert.Equal(t, 9, quality.WorkflowStepCount())
	assert.Equal(t, 12, quality.UsefulSourceCount())
	assert.True(t, quality.MeetsFloor())
}

// ---------------------------------------------------------------------------
// 8. DomainResearchResult.Quality() auto-computation
// ---------------------------------------------------------------------------

func TestQuality_DomainResearchResultAutoComputation(t *testing.T) {
	t.Parallel()

	actors := makeQualityActors(t, 3)
	entities := makeQualityEntities(t, 3)
	wf := makeQualityWorkflow(t, "MainFlow", 5)
	meta := NewSearchMetadata([]string{"query1"}, 20, 5, time.Second)

	result, err := NewDomainResearchResult(
		"TestDomain",
		meta,
		actors,
		entities,
		[]ResearchedWorkflow{wf},
		nil, nil, nil,
	)
	require.NoError(t, err)

	// Quality() should auto-compute from the result's own data.
	quality := result.Quality()

	// Verify it matches a direct ComputeResearchQuality call with the same data.
	expected := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wf}, meta.UsefulSources())

	assert.Equal(t, expected.ActorCount(), quality.ActorCount())
	assert.Equal(t, expected.EntityCount(), quality.EntityCount())
	assert.Equal(t, expected.WorkflowStepCount(), quality.WorkflowStepCount())
	assert.Equal(t, expected.UsefulSourceCount(), quality.UsefulSourceCount())
	assert.Equal(t, expected.MeetsFloor(), quality.MeetsFloor())

	// Verify Quality() uses SearchMetadata.UsefulSources() for the usefulSources param.
	assert.Equal(t, 5, quality.UsefulSourceCount())
	assert.True(t, quality.MeetsFloor())
}

func TestQuality_DomainResearchResultBelowFloor(t *testing.T) {
	t.Parallel()

	actors := makeQualityActors(t, 2)
	entities := makeQualityEntities(t, 3)
	wf := makeQualityWorkflow(t, "Flow", 5)
	meta := NewSearchMetadata(nil, 10, 5, 0)

	result, err := NewDomainResearchResult("Domain", meta, actors, entities, []ResearchedWorkflow{wf}, nil, nil, nil)
	require.NoError(t, err)

	quality := result.Quality()

	assert.Equal(t, 2, quality.ActorCount())
	assert.False(t, quality.MeetsFloor())
}

func TestQuality_DomainResearchResultQualityUsesUsefulSourcesNotTotal(t *testing.T) {
	t.Parallel()

	// Verify that Quality() derives usefulSources from meta.usefulSources, not meta.totalSources.
	actors := makeQualityActors(t, 3)
	entities := makeQualityEntities(t, 3)
	wf := makeQualityWorkflow(t, "Flow", 5)

	// totalSources=100, usefulSources=3. If Quality() mistakenly used totalSources, it would pass.
	meta := NewSearchMetadata(nil, 100, 3, 0)

	result, err := NewDomainResearchResult("Domain", meta, actors, entities, []ResearchedWorkflow{wf}, nil, nil, nil)
	require.NoError(t, err)

	quality := result.Quality()

	assert.Equal(t, 3, quality.UsefulSourceCount())
	assert.False(t, quality.MeetsFloor())
}

// ---------------------------------------------------------------------------
// 9. Zero-length workflows in the mix
// ---------------------------------------------------------------------------

func TestQuality_ZeroLengthWorkflowsInMix(t *testing.T) {
	t.Parallel()

	t.Run("workflow with 5 steps plus workflow with 0 steps", func(t *testing.T) {
		t.Parallel()

		actors := makeQualityActors(t, 3)
		entities := makeQualityEntities(t, 3)
		wfFull := makeQualityWorkflow(t, "FullFlow", 5)

		// A workflow with zero steps is valid per the constructor.
		wfEmpty, err := NewResearchedWorkflow("EmptyFlow", WorkflowTypeSecondary, nil, nil)
		require.NoError(t, err)

		quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wfFull, wfEmpty}, 5)

		assert.Equal(t, 5, quality.WorkflowStepCount())
		assert.True(t, quality.MeetsFloor())
	})

	t.Run("two workflows with 0 steps each", func(t *testing.T) {
		t.Parallel()

		actors := makeQualityActors(t, 3)
		entities := makeQualityEntities(t, 3)

		wfA, err := NewResearchedWorkflow("EmptyA", WorkflowTypeHappyPath, nil, nil)
		require.NoError(t, err)

		wfB, err := NewResearchedWorkflow("EmptyB", WorkflowTypeFailureCase, nil, nil)
		require.NoError(t, err)

		quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wfA, wfB}, 5)

		assert.Equal(t, 0, quality.WorkflowStepCount())
		assert.False(t, quality.MeetsFloor())
	})
}

// ---------------------------------------------------------------------------
// Additional quality edge cases
// ---------------------------------------------------------------------------

func TestQuality_ThresholdConstants(t *testing.T) {
	t.Parallel()

	// Verify the threshold constants match the documented values from spike Section 9.
	assert.Equal(t, 3, QualityFloorActors)
	assert.Equal(t, 3, QualityFloorEntities)
	assert.Equal(t, 5, QualityFloorWorkflowSteps)
	assert.Equal(t, 5, QualityFloorUsefulSources)
}

func TestQuality_NegativeUsefulSources(t *testing.T) {
	t.Parallel()

	// Negative usefulSources should not meet floor (no validation on the int).
	actors := makeQualityActors(t, 3)
	entities := makeQualityEntities(t, 3)
	wf := makeQualityWorkflow(t, "Flow", 5)

	quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wf}, -1)

	assert.Equal(t, -1, quality.UsefulSourceCount())
	assert.False(t, quality.MeetsFloor())
}

func TestQuality_ManyWorkflowsOneStepEach(t *testing.T) {
	t.Parallel()

	// 5 workflows x 1 step each = 5 total -> meets threshold.
	actors := makeQualityActors(t, 3)
	entities := makeQualityEntities(t, 3)

	workflows := make([]ResearchedWorkflow, 5)

	for i := range 5 {
		workflows[i] = makeQualityWorkflow(t, "WF"+string(rune('A'+i)), 1)
	}

	quality := ComputeResearchQuality(actors, entities, workflows, 5)

	assert.Equal(t, 5, quality.WorkflowStepCount())
	assert.True(t, quality.MeetsFloor())
}

func TestQuality_SingleWorkflowExactlyAtStepThreshold(t *testing.T) {
	t.Parallel()

	// Exactly 5 steps in one workflow -- meets threshold.
	actors := makeQualityActors(t, 3)
	entities := makeQualityEntities(t, 3)
	wf := makeQualityWorkflow(t, "Exact5", 5)

	quality := ComputeResearchQuality(actors, entities, []ResearchedWorkflow{wf}, 5)

	assert.Equal(t, 5, quality.WorkflowStepCount())
	assert.True(t, quality.MeetsFloor())
}

// ===========================================================================
// Defensive copy tests (QA — alty-cli-1wu.16, immutability verification)
// ===========================================================================

// ---------------------------------------------------------------------------
// SearchMetadata — queriesUsed []string
// ---------------------------------------------------------------------------

func TestSearchMetadata_Defensive_NilInput(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		meta := NewSearchMetadata(nil, 5, 3, time.Second)
		assert.Empty(t, meta.QueriesUsed())
	})
}

func TestSearchMetadata_Defensive_EmptyInput(t *testing.T) {
	t.Parallel()

	meta := NewSearchMetadata([]string{}, 5, 3, time.Second)
	assert.Empty(t, meta.QueriesUsed())
}

func TestSearchMetadata_Defensive_GetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	meta := NewSearchMetadata([]string{"q1", "q2"}, 10, 5, time.Second)

	first := meta.QueriesUsed()
	second := meta.QueriesUsed()
	first[0] = "MUTATED"
	assert.Equal(t, "q1", second[0], "mutating one getter result must not affect another")
	assert.Equal(t, "q1", meta.QueriesUsed()[0], "internal state must be unchanged")
}

// ---------------------------------------------------------------------------
// ResearchedActor — sourceURLs []string
// ---------------------------------------------------------------------------

func TestResearchedActor_Defensive_InputSliceMutation(t *testing.T) {
	t.Parallel()

	urls := []string{"http://a.com", "http://b.com"}
	actor, err := NewResearchedActor("Actor", "role", urls)
	require.NoError(t, err)

	urls[0] = "MUTATED"
	assert.Equal(t, "http://a.com", actor.SourceURLs()[0])
}

func TestResearchedActor_Defensive_GetterSliceMutation(t *testing.T) {
	t.Parallel()

	actor, err := NewResearchedActor("Actor", "role", []string{"http://a.com"})
	require.NoError(t, err)

	got := actor.SourceURLs()
	got[0] = "MUTATED"
	assert.Equal(t, "http://a.com", actor.SourceURLs()[0])
}

func TestResearchedActor_Defensive_NilSourceURLsNoPanic(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		actor, err := NewResearchedActor("Actor", "role", nil)
		require.NoError(t, err)
		assert.Empty(t, actor.SourceURLs())
	})
}

func TestResearchedActor_Defensive_EmptySourceURLs(t *testing.T) {
	t.Parallel()

	actor, err := NewResearchedActor("Actor", "role", []string{})
	require.NoError(t, err)
	assert.Empty(t, actor.SourceURLs())
}

func TestResearchedActor_Defensive_GetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	actor, err := NewResearchedActor("Actor", "role", []string{"http://a.com", "http://b.com"})
	require.NoError(t, err)

	first := actor.SourceURLs()
	second := actor.SourceURLs()
	first[0] = "MUTATED"
	assert.Equal(t, "http://a.com", second[0], "mutating one getter result must not affect another")
}

// ---------------------------------------------------------------------------
// ResearchedEntity — properties []string AND sourceURLs []string
// ---------------------------------------------------------------------------

func TestResearchedEntity_Defensive_PropertiesInputMutation(t *testing.T) {
	t.Parallel()

	props := []string{"name", "dob"}
	entity, err := NewResearchedEntity("Record", props, nil)
	require.NoError(t, err)

	props[0] = "MUTATED"
	assert.Equal(t, "name", entity.Properties()[0])
}

func TestResearchedEntity_Defensive_PropertiesGetterMutation(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", []string{"name", "dob"}, nil)
	require.NoError(t, err)

	got := entity.Properties()
	got[0] = "MUTATED"
	assert.Equal(t, "name", entity.Properties()[0])
}

func TestResearchedEntity_Defensive_SourceURLsInputMutation(t *testing.T) {
	t.Parallel()

	urls := []string{"http://a.com", "http://b.com"}
	entity, err := NewResearchedEntity("Record", nil, urls)
	require.NoError(t, err)

	urls[0] = "MUTATED"
	assert.Equal(t, "http://a.com", entity.SourceURLs()[0])
}

func TestResearchedEntity_Defensive_SourceURLsGetterMutation(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", nil, []string{"http://a.com", "http://b.com"})
	require.NoError(t, err)

	got := entity.SourceURLs()
	got[0] = "MUTATED"
	assert.Equal(t, "http://a.com", entity.SourceURLs()[0])
}

func TestResearchedEntity_Defensive_NilProperties(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", nil, []string{"http://a.com"})
	require.NoError(t, err)
	assert.Empty(t, entity.Properties())
}

func TestResearchedEntity_Defensive_EmptyProperties(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", []string{}, []string{"http://a.com"})
	require.NoError(t, err)
	assert.Empty(t, entity.Properties())
}

func TestResearchedEntity_Defensive_NilSourceURLs(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", []string{"prop"}, nil)
	require.NoError(t, err)
	assert.Empty(t, entity.SourceURLs())
}

func TestResearchedEntity_Defensive_EmptySourceURLs(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", []string{"prop"}, []string{})
	require.NoError(t, err)
	assert.Empty(t, entity.SourceURLs())
}

func TestResearchedEntity_Defensive_BothNil(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", nil, nil)
	require.NoError(t, err)
	assert.Empty(t, entity.Properties())
	assert.Empty(t, entity.SourceURLs())
}

func TestResearchedEntity_Defensive_BothEmpty(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", []string{}, []string{})
	require.NoError(t, err)
	assert.Empty(t, entity.Properties())
	assert.Empty(t, entity.SourceURLs())
}

func TestResearchedEntity_Defensive_PropertiesGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", []string{"name", "dob"}, nil)
	require.NoError(t, err)

	first := entity.Properties()
	second := entity.Properties()
	first[0] = "MUTATED"
	assert.Equal(t, "name", second[0], "mutating one getter result must not affect another")
}

func TestResearchedEntity_Defensive_SourceURLsGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	entity, err := NewResearchedEntity("Record", nil, []string{"http://a.com", "http://b.com"})
	require.NoError(t, err)

	first := entity.SourceURLs()
	second := entity.SourceURLs()
	first[0] = "MUTATED"
	assert.Equal(t, "http://a.com", second[0], "mutating one getter result must not affect another")
}

// ---------------------------------------------------------------------------
// ResearchedWorkflow — steps []WorkflowStep AND sourceURLs []string
// ---------------------------------------------------------------------------

func TestResearchedWorkflow_Defensive_StepsInputMutation(t *testing.T) {
	t.Parallel()

	step, err := NewWorkflowStep(1, "Vet", "examines", "Patient")
	require.NoError(t, err)

	steps := []WorkflowStep{step}
	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, steps, nil)
	require.NoError(t, err)

	replacement, err := NewWorkflowStep(99, "Impostor", "replaces", "Nothing")
	require.NoError(t, err)
	steps[0] = replacement

	assert.Equal(t, "Vet", wf.Steps()[0].Actor())
}

func TestResearchedWorkflow_Defensive_StepsGetterMutation(t *testing.T) {
	t.Parallel()

	step, err := NewWorkflowStep(1, "Vet", "examines", "Patient")
	require.NoError(t, err)

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, []WorkflowStep{step}, nil)
	require.NoError(t, err)

	got := wf.Steps()
	replacement, err := NewWorkflowStep(99, "Impostor", "replaces", "Nothing")
	require.NoError(t, err)
	got[0] = replacement

	assert.Equal(t, "Vet", wf.Steps()[0].Actor())
}

func TestResearchedWorkflow_Defensive_SourceURLsInputMutation(t *testing.T) {
	t.Parallel()

	urls := []string{"http://a.com"}
	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, nil, urls)
	require.NoError(t, err)

	urls[0] = "MUTATED"
	assert.Equal(t, "http://a.com", wf.SourceURLs()[0])
}

func TestResearchedWorkflow_Defensive_SourceURLsGetterMutation(t *testing.T) {
	t.Parallel()

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, nil, []string{"http://a.com"})
	require.NoError(t, err)

	got := wf.SourceURLs()
	got[0] = "MUTATED"
	assert.Equal(t, "http://a.com", wf.SourceURLs()[0])
}

func TestResearchedWorkflow_Defensive_NilSteps(t *testing.T) {
	t.Parallel()

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, nil, []string{"http://a.com"})
	require.NoError(t, err)
	assert.Empty(t, wf.Steps())
}

func TestResearchedWorkflow_Defensive_EmptySteps(t *testing.T) {
	t.Parallel()

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, []WorkflowStep{}, nil)
	require.NoError(t, err)
	assert.Empty(t, wf.Steps())
}

func TestResearchedWorkflow_Defensive_NilSourceURLs(t *testing.T) {
	t.Parallel()

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, wf.SourceURLs())
}

func TestResearchedWorkflow_Defensive_EmptySourceURLs(t *testing.T) {
	t.Parallel()

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, nil, []string{})
	require.NoError(t, err)
	assert.Empty(t, wf.SourceURLs())
}

func TestResearchedWorkflow_Defensive_StepsGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	step, err := NewWorkflowStep(1, "Vet", "examines", "Patient")
	require.NoError(t, err)

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, []WorkflowStep{step}, nil)
	require.NoError(t, err)

	first := wf.Steps()
	second := wf.Steps()
	replacement, err := NewWorkflowStep(99, "Impostor", "replaces", "Nothing")
	require.NoError(t, err)
	first[0] = replacement

	assert.Equal(t, "Vet", second[0].Actor(), "mutating one getter result must not affect another")
}

func TestResearchedWorkflow_Defensive_SourceURLsGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	wf, err := NewResearchedWorkflow("Flow", WorkflowTypeHappyPath, nil, []string{"http://a.com", "http://b.com"})
	require.NoError(t, err)

	first := wf.SourceURLs()
	second := wf.SourceURLs()
	first[0] = "MUTATED"
	assert.Equal(t, "http://a.com", second[0], "mutating one getter result must not affect another")
}

// ---------------------------------------------------------------------------
// RegulatoryItem — sourceURLs []string
// ---------------------------------------------------------------------------

func TestRegulatoryItem_Defensive_InputMutation(t *testing.T) {
	t.Parallel()

	urls := []string{"http://a.com", "http://b.com"}
	item, err := NewRegulatoryItem("HIPAA", "desc", urls)
	require.NoError(t, err)

	urls[0] = "MUTATED"
	assert.Equal(t, "http://a.com", item.SourceURLs()[0])
}

func TestRegulatoryItem_Defensive_GetterMutation(t *testing.T) {
	t.Parallel()

	item, err := NewRegulatoryItem("HIPAA", "desc", []string{"http://a.com", "http://b.com"})
	require.NoError(t, err)

	got := item.SourceURLs()
	got[0] = "MUTATED"
	assert.Equal(t, "http://a.com", item.SourceURLs()[0])
}

func TestRegulatoryItem_Defensive_NilSourceURLs(t *testing.T) {
	t.Parallel()

	item, err := NewRegulatoryItem("HIPAA", "desc", nil)
	require.NoError(t, err)
	assert.Empty(t, item.SourceURLs())
}

func TestRegulatoryItem_Defensive_EmptySourceURLs(t *testing.T) {
	t.Parallel()

	item, err := NewRegulatoryItem("HIPAA", "desc", []string{})
	require.NoError(t, err)
	assert.Empty(t, item.SourceURLs())
}

func TestRegulatoryItem_Defensive_GetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	item, err := NewRegulatoryItem("HIPAA", "desc", []string{"http://a.com", "http://b.com"})
	require.NoError(t, err)

	first := item.SourceURLs()
	second := item.SourceURLs()
	first[0] = "MUTATED"
	assert.Equal(t, "http://a.com", second[0], "mutating one getter result must not affect another")
}

// ---------------------------------------------------------------------------
// DomainResearchResult — all 6 slice fields
// ---------------------------------------------------------------------------

// --- actors ---

func TestDomainResearchResult_Defensive_ActorsInputMutation(t *testing.T) {
	t.Parallel()

	actors := makeActors(2)
	originalName := actors[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), actors, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	replacement, err := NewResearchedActor("Mutated", "role", nil)
	require.NoError(t, err)
	actors[0] = replacement

	assert.Equal(t, originalName, result.Actors()[0].Name())
}

func TestDomainResearchResult_Defensive_ActorsGetterMutation(t *testing.T) {
	t.Parallel()

	actors := makeActors(2)
	originalName := actors[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), actors, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	got := result.Actors()
	replacement, err := NewResearchedActor("Mutated", "role", nil)
	require.NoError(t, err)
	got[0] = replacement

	assert.Equal(t, originalName, result.Actors()[0].Name())
}

func TestDomainResearchResult_Defensive_ActorsGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	actors := makeActors(2)
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), actors, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	first := result.Actors()
	second := result.Actors()
	replacement, err := NewResearchedActor("Impostor", "role", nil)
	require.NoError(t, err)
	first[0] = replacement

	assert.NotEqual(t, "Impostor", second[0].Name(), "mutating one getter result must not affect another")
}

// --- entities ---

func TestDomainResearchResult_Defensive_EntitiesInputMutation(t *testing.T) {
	t.Parallel()

	entities := makeEntities(2)
	originalName := entities[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, entities, nil, nil, nil, nil)
	require.NoError(t, err)

	replacement, err := NewResearchedEntity("Mutated", nil, nil)
	require.NoError(t, err)
	entities[0] = replacement

	assert.Equal(t, originalName, result.Entities()[0].Name())
}

func TestDomainResearchResult_Defensive_EntitiesGetterMutation(t *testing.T) {
	t.Parallel()

	entities := makeEntities(2)
	originalName := entities[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, entities, nil, nil, nil, nil)
	require.NoError(t, err)

	got := result.Entities()
	replacement, err := NewResearchedEntity("Mutated", nil, nil)
	require.NoError(t, err)
	got[0] = replacement

	assert.Equal(t, originalName, result.Entities()[0].Name())
}

func TestDomainResearchResult_Defensive_EntitiesGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	entities := makeEntities(2)
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, entities, nil, nil, nil, nil)
	require.NoError(t, err)

	first := result.Entities()
	second := result.Entities()
	replacement, err := NewResearchedEntity("Impostor", nil, nil)
	require.NoError(t, err)
	first[0] = replacement

	assert.NotEqual(t, "Impostor", second[0].Name(), "mutating one getter result must not affect another")
}

// --- workflows ---

func TestDomainResearchResult_Defensive_WorkflowsInputMutation(t *testing.T) {
	t.Parallel()

	workflows := makeWorkflowsWithTotalSteps(3)
	originalName := workflows[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, workflows, nil, nil, nil)
	require.NoError(t, err)

	replacement, err := NewResearchedWorkflow("Mutated", WorkflowTypeSecondary, nil, nil)
	require.NoError(t, err)
	workflows[0] = replacement

	assert.Equal(t, originalName, result.Workflows()[0].Name())
}

func TestDomainResearchResult_Defensive_WorkflowsGetterMutation(t *testing.T) {
	t.Parallel()

	workflows := makeWorkflowsWithTotalSteps(3)
	originalName := workflows[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, workflows, nil, nil, nil)
	require.NoError(t, err)

	got := result.Workflows()
	replacement, err := NewResearchedWorkflow("Mutated", WorkflowTypeSecondary, nil, nil)
	require.NoError(t, err)
	got[0] = replacement

	assert.Equal(t, originalName, result.Workflows()[0].Name())
}

func TestDomainResearchResult_Defensive_WorkflowsGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	workflows := makeWorkflowsWithTotalSteps(3)
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, workflows, nil, nil, nil)
	require.NoError(t, err)

	first := result.Workflows()
	second := result.Workflows()
	replacement, err := NewResearchedWorkflow("Impostor", WorkflowTypeSecondary, nil, nil)
	require.NoError(t, err)
	first[0] = replacement

	assert.NotEqual(t, "Impostor", second[0].Name(), "mutating one getter result must not affect another")
}

// --- failureModes ---

func TestDomainResearchResult_Defensive_FailureModesInputMutation(t *testing.T) {
	t.Parallel()

	modes := []string{"power outage", "data loss"}
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, modes, nil, nil)
	require.NoError(t, err)

	modes[0] = "MUTATED"
	assert.Equal(t, "power outage", result.FailureModes()[0])
}

func TestDomainResearchResult_Defensive_FailureModesGetterMutation(t *testing.T) {
	t.Parallel()

	modes := []string{"power outage", "data loss"}
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, modes, nil, nil)
	require.NoError(t, err)

	got := result.FailureModes()
	got[0] = "MUTATED"
	assert.Equal(t, "power outage", result.FailureModes()[0])
}

func TestDomainResearchResult_Defensive_FailureModesGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	modes := []string{"power outage", "data loss"}
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, modes, nil, nil)
	require.NoError(t, err)

	first := result.FailureModes()
	second := result.FailureModes()
	first[0] = "MUTATED"
	assert.Equal(t, "power outage", second[0], "mutating one getter result must not affect another")
}

// --- regulatory ---

func TestDomainResearchResult_Defensive_RegulatoryInputMutation(t *testing.T) {
	t.Parallel()

	regulatory := makeRegulatoryItems(2)
	originalName := regulatory[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, regulatory, nil)
	require.NoError(t, err)

	replacement, err := NewRegulatoryItem("Mutated", "desc", nil)
	require.NoError(t, err)
	regulatory[0] = replacement

	assert.Equal(t, originalName, result.Regulatory()[0].Name())
}

func TestDomainResearchResult_Defensive_RegulatoryGetterMutation(t *testing.T) {
	t.Parallel()

	regulatory := makeRegulatoryItems(2)
	originalName := regulatory[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, regulatory, nil)
	require.NoError(t, err)

	got := result.Regulatory()
	replacement, err := NewRegulatoryItem("Mutated", "desc", nil)
	require.NoError(t, err)
	got[0] = replacement

	assert.Equal(t, originalName, result.Regulatory()[0].Name())
}

func TestDomainResearchResult_Defensive_RegulatoryGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	regulatory := makeRegulatoryItems(2)
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, regulatory, nil)
	require.NoError(t, err)

	first := result.Regulatory()
	second := result.Regulatory()
	replacement, err := NewRegulatoryItem("Impostor", "desc", nil)
	require.NoError(t, err)
	first[0] = replacement

	assert.NotEqual(t, "Impostor", second[0].Name(), "mutating one getter result must not affect another")
}

// --- software ---

func TestDomainResearchResult_Defensive_SoftwareInputMutation(t *testing.T) {
	t.Parallel()

	software := makeSoftwareItems(2)
	originalName := software[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, software)
	require.NoError(t, err)

	replacement, err := NewExistingSoftware("Mutated", "desc", "http://x.com")
	require.NoError(t, err)
	software[0] = replacement

	assert.Equal(t, originalName, result.Software()[0].Name())
}

func TestDomainResearchResult_Defensive_SoftwareGetterMutation(t *testing.T) {
	t.Parallel()

	software := makeSoftwareItems(2)
	originalName := software[0].Name()
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, software)
	require.NoError(t, err)

	got := result.Software()
	replacement, err := NewExistingSoftware("Mutated", "desc", "http://x.com")
	require.NoError(t, err)
	got[0] = replacement

	assert.Equal(t, originalName, result.Software()[0].Name())
}

func TestDomainResearchResult_Defensive_SoftwareGetterReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	software := makeSoftwareItems(2)
	result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, software)
	require.NoError(t, err)

	first := result.Software()
	second := result.Software()
	replacement, err := NewExistingSoftware("Impostor", "desc", "http://x.com")
	require.NoError(t, err)
	first[0] = replacement

	assert.NotEqual(t, "Impostor", second[0].Name(), "mutating one getter result must not affect another")
}

// --- nil/empty for all 6 fields ---

func TestDomainResearchResult_Defensive_AllNilSlicesNoPanic(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		result, err := NewDomainResearchResult("Vet", NewSearchMetadata(nil, 0, 0, 0), nil, nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, result.Actors())
		assert.Empty(t, result.Entities())
		assert.Empty(t, result.Workflows())
		assert.Empty(t, result.FailureModes())
		assert.Empty(t, result.Regulatory())
		assert.Empty(t, result.Software())
	})
}

func TestDomainResearchResult_Defensive_AllEmptySlices(t *testing.T) {
	t.Parallel()

	result, err := NewDomainResearchResult(
		"Vet",
		NewSearchMetadata([]string{}, 0, 0, 0),
		[]ResearchedActor{},
		[]ResearchedEntity{},
		[]ResearchedWorkflow{},
		[]string{},
		[]RegulatoryItem{},
		[]ExistingSoftware{},
	)
	require.NoError(t, err)
	assert.Empty(t, result.Actors())
	assert.Empty(t, result.Entities())
	assert.Empty(t, result.Workflows())
	assert.Empty(t, result.FailureModes())
	assert.Empty(t, result.Regulatory())
	assert.Empty(t, result.Software())
}
