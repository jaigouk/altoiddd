package events_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/events"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// DomainModelGenerated
// ---------------------------------------------------------------------------

func TestDomainModelGeneratedCreate(t *testing.T) {
	t.Parallel()

	core := vo.SubdomainCore
	event := events.NewDomainModelGenerated(
		"test-123",
		[]vo.DomainStory{
			vo.NewDomainStory("Test", []string{"A"}, "T", []string{"S"}, nil),
		},
		[]vo.TermEntry{
			vo.NewTermEntry("Order", "A purchase", "Sales", nil),
		},
		[]vo.DomainBoundedContext{
			vo.NewDomainBoundedContext("Sales", "Orders", nil, &core, ""),
		},
		nil,
		[]vo.AggregateDesign{
			vo.NewAggregateDesign("OrderAgg", "Sales", "Order", nil, nil, nil, nil),
		},
	)
	assert.Equal(t, "test-123", event.ModelID())
	assert.Len(t, event.DomainStories(), 1)
	assert.Len(t, event.BoundedContexts(), 1)
}

func TestDomainModelGeneratedEmpty(t *testing.T) {
	t.Parallel()

	event := events.NewDomainModelGenerated("empty", nil, nil, nil, nil, nil)
	assert.Empty(t, event.DomainStories())
	assert.Empty(t, event.AggregateDesigns())
}

// ---------------------------------------------------------------------------
// ConfigsGenerated
// ---------------------------------------------------------------------------

func TestConfigsGeneratedCreate(t *testing.T) {
	t.Parallel()

	event := events.NewConfigsGenerated(
		[]string{"claude-code", "cursor"},
		[]string{".claude/CLAUDE.md", "AGENTS.md"},
	)
	assert.Equal(t, []string{"claude-code", "cursor"}, event.ToolNames())
	assert.Equal(t, []string{".claude/CLAUDE.md", "AGENTS.md"}, event.OutputPaths())
}

func TestConfigsGeneratedStoresTupleFields(t *testing.T) {
	t.Parallel()

	event := events.NewConfigsGenerated(
		[]string{"claude-code", "cursor", "roo-code"},
		[]string{"a.md", "b.md", "c.md"},
	)
	assert.Len(t, event.ToolNames(), 3)
	assert.Len(t, event.OutputPaths(), 3)
}

func TestConfigsGeneratedEmptySlicesAllowed(t *testing.T) {
	t.Parallel()

	event := events.NewConfigsGenerated(nil, nil)
	assert.Empty(t, event.ToolNames())
	assert.Empty(t, event.OutputPaths())
}

func TestConfigsGeneratedDefensiveCopy(t *testing.T) {
	t.Parallel()

	event := events.NewConfigsGenerated(
		[]string{"claude-code"},
		[]string{".claude/CLAUDE.md"},
	)
	names := event.ToolNames()
	names[0] = "MODIFIED"
	assert.Equal(t, "claude-code", event.ToolNames()[0])
}

// ---------------------------------------------------------------------------
// JSON roundtrip (event bus serialization)
// ---------------------------------------------------------------------------

func TestDomainModelGenerated_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	core := vo.SubdomainCore
	original := events.NewDomainModelGenerated(
		"model-roundtrip",
		[]vo.DomainStory{
			vo.NewDomainStory("Story1", []string{"A"}, "T", []string{"S"}, nil),
		},
		[]vo.TermEntry{
			vo.NewTermEntry("Order", "A purchase", "Sales", nil),
		},
		[]vo.DomainBoundedContext{
			vo.NewDomainBoundedContext("Sales", "Handles sales", nil, &core, ""),
		},
		nil,
		[]vo.AggregateDesign{
			vo.NewAggregateDesign("OrderAgg", "Sales", "Order", nil, nil, nil, nil),
		},
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"model_id"`)
	assert.Contains(t, string(data), `"model-roundtrip"`)

	var restored events.DomainModelGenerated
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.ModelID(), restored.ModelID())
	require.Len(t, restored.DomainStories(), 1)
	assert.Equal(t, "Story1", restored.DomainStories()[0].Name())
	assert.Equal(t, "T", restored.DomainStories()[0].Trigger())
	assert.Equal(t, []string{"A"}, restored.DomainStories()[0].Actors())
	require.Len(t, restored.UbiquitousLanguage(), 1)
	assert.Equal(t, "Order", restored.UbiquitousLanguage()[0].Term())
	assert.Equal(t, "A purchase", restored.UbiquitousLanguage()[0].Definition())
	require.Len(t, restored.BoundedContexts(), 1)
	assert.Equal(t, "Sales", restored.BoundedContexts()[0].Name())
	assert.Equal(t, "Handles sales", restored.BoundedContexts()[0].Responsibility())
	require.NotNil(t, restored.BoundedContexts()[0].Classification())
	assert.Equal(t, vo.SubdomainCore, *restored.BoundedContexts()[0].Classification())
	require.Len(t, restored.AggregateDesigns(), 1)
	assert.Equal(t, "OrderAgg", restored.AggregateDesigns()[0].Name())
	assert.Equal(t, "Sales", restored.AggregateDesigns()[0].ContextName())
	assert.Equal(t, "Order", restored.AggregateDesigns()[0].RootEntity())
}

func TestGapAnalysisCompleted_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := events.NewGapAnalysisCompleted("gap-1", "/tmp/proj", 7, 3)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"analysis_id"`)
	assert.Contains(t, string(data), `"gaps_found"`)

	var restored events.GapAnalysisCompleted
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "gap-1", restored.AnalysisID())
	assert.Equal(t, "/tmp/proj", restored.ProjectDir())
	assert.Equal(t, 7, restored.GapsFound())
	assert.Equal(t, 3, restored.GapsResolved())
}

func TestConfigsGenerated_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := events.NewConfigsGenerated(
		[]string{"claude-code", "cursor"},
		[]string{".claude/CLAUDE.md", "AGENTS.md"},
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"tool_names"`)
	assert.Contains(t, string(data), `"output_paths"`)

	var restored events.ConfigsGenerated
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.ToolNames(), restored.ToolNames())
	assert.Equal(t, original.OutputPaths(), restored.OutputPaths())
}
