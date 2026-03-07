package events_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alty-cli/alty/internal/shared/domain/events"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
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
