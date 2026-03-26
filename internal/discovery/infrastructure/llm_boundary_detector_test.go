package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

// stubLLMClient implements llm.Client with configurable responses for testing.
type stubLLMClient struct {
	response llm.Response
	err      error
	called   bool
}

func (s *stubLLMClient) StructuredOutput(_ context.Context, _ string, _ map[string]any) (llm.Response, error) {
	s.called = true
	return s.response, s.err
}

func (s *stubLLMClient) TextCompletion(_ context.Context, _ string) (llm.Response, error) {
	return llm.NewResponse("", "", 0), nil
}

// makeTestStory creates a minimal valid DomainStory for testing.
func makeTestStory(t *testing.T, title, trigger string) *domain.DomainStory {
	t.Helper()

	story, err := domain.NewDomainStory(
		title,
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		trigger,
	)
	require.NoError(t, err)

	return story
}

func TestLLMBoundaryDetector_LanguageDifference_Detected(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		response: llm.NewResponse(`{
			"language_differences": [
				{
					"term": "Order",
					"story_a": "Place Order",
					"story_b": "Fulfill Order",
					"description": "Order has different properties in each context"
				}
			],
			"trigger_classifications": []
		}`, "test-model", 100),
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
		makeTestStory(t, "Fulfill Order", "Warehouse receives order"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.NoError(t, err)
	require.Len(t, signals, 1)
	assert.Equal(t, domain.SignalTypeLanguageDifference, signals[0].Type())
	assert.Contains(t, signals[0].Description(), "Order")
}

func TestLLMBoundaryDetector_DifferentTrigger_Detected(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		response: llm.NewResponse(`{
			"language_differences": [],
			"trigger_classifications": [
				{"story_title": "Place Order", "trigger_type": "user_initiated", "trigger_text": "Customer clicks buy"},
				{"story_title": "Process Refund", "trigger_type": "system_initiated", "trigger_text": "System detects chargeback"}
			]
		}`, "test-model", 100),
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
		makeTestStory(t, "Process Refund", "System detects chargeback"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.NoError(t, err)
	require.Len(t, signals, 1)
	assert.Equal(t, domain.SignalTypeDifferentTrigger, signals[0].Type())
	assert.Contains(t, signals[0].Description(), "Place Order")
	assert.Contains(t, signals[0].Description(), "Process Refund")
}

func TestLLMBoundaryDetector_SameTriggerType_NoSignal(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		response: llm.NewResponse(`{
			"language_differences": [],
			"trigger_classifications": [
				{"story_title": "Place Order", "trigger_type": "user_initiated", "trigger_text": "Customer clicks buy"},
				{"story_title": "Cancel Order", "trigger_type": "user_initiated", "trigger_text": "Customer clicks cancel"}
			]
		}`, "test-model", 100),
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
		makeTestStory(t, "Cancel Order", "Customer clicks cancel"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.NoError(t, err)
	assert.Empty(t, signals)
}

func TestLLMBoundaryDetector_LLMUnavailable_ReturnsNilNil(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		err: llm.ErrLLMUnavailable,
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.NoError(t, err)
	assert.Nil(t, signals)
}

func TestLLMBoundaryDetector_LLMError_Propagated(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		err: errors.New("connection timeout"),
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.Error(t, err)
	assert.Nil(t, signals)
	assert.Contains(t, err.Error(), "detecting boundary signals")
}

func TestLLMBoundaryDetector_EmptyStories_NoLLMCall(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{}
	detector := NewLLMBoundaryDetector(stub)

	signals, err := detector.DetectBoundarySignals(context.Background(), nil)

	require.NoError(t, err)
	assert.Nil(t, signals)
	assert.False(t, stub.called)
}

func TestLLMBoundaryDetector_MalformedJSON_Error(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		response: llm.NewResponse("not json at all", "test-model", 100),
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.Error(t, err)
	assert.Nil(t, signals)
	assert.Contains(t, err.Error(), "parsing boundary detection response")
}

func TestLLMBoundaryDetector_DeduplicateSignals(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		response: llm.NewResponse(`{
			"language_differences": [
				{"term": "Order", "story_a": "Place Order", "story_b": "Fulfill Order", "description": "Order has different properties"},
				{"term": "Order", "story_a": "Place Order", "story_b": "Fulfill Order", "description": "Order has different properties"}
			],
			"trigger_classifications": []
		}`, "test-model", 100),
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
		makeTestStory(t, "Fulfill Order", "Warehouse receives order"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.NoError(t, err)
	require.Len(t, signals, 1)
}

func TestLLMBoundaryDetector_UnknownTriggerType_Skipped(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		response: llm.NewResponse(`{
			"language_differences": [],
			"trigger_classifications": [
				{"story_title": "Place Order", "trigger_type": "user_initiated", "trigger_text": "Customer clicks buy"},
				{"story_title": "Process Emotion", "trigger_type": "emotional_decision", "trigger_text": "User feels bad"}
			]
		}`, "test-model", 100),
	}
	detector := NewLLMBoundaryDetector(stub)
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
		makeTestStory(t, "Process Emotion", "User feels bad"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.NoError(t, err)
	// Only 1 valid trigger type remains after filtering → no pair → no signal.
	assert.Empty(t, signals)
}

func TestLLMBoundaryDetector_SingleStory_NoTriggerSignal(t *testing.T) {
	t.Parallel()

	stub := &stubLLMClient{
		response: llm.NewResponse(`{
			"language_differences": [],
			"trigger_classifications": [
				{"story_title": "Place Order", "trigger_type": "user_initiated", "trigger_text": "Customer clicks buy"},
				{"story_title": "Cancel Order", "trigger_type": "system_initiated", "trigger_text": "System cancels"}
			]
		}`, "test-model", 100),
	}
	detector := NewLLMBoundaryDetector(stub)
	// Only 1 story — nothing to compare for trigger differences.
	stories := []*domain.DomainStory{
		makeTestStory(t, "Place Order", "Customer clicks buy"),
	}

	signals, err := detector.DetectBoundarySignals(context.Background(), stories)

	require.NoError(t, err)
	assert.Empty(t, signals)
}
