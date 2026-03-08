package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	researchapp "github.com/alty-cli/alty/internal/research/application"
	researchdomain "github.com/alty-cli/alty/internal/research/domain"
	"github.com/alty-cli/alty/internal/research/infrastructure"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/shared/infrastructure/llm"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeModel(t *testing.T, contextNames ...string) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("test-model")
	for _, name := range contextNames {
		ctx := vo.NewDomainBoundedContext(name, name+" responsibility", nil, nil, "")
		require.NoError(t, model.AddBoundedContext(ctx))
	}
	return model
}

func makeSearchResults(n int) []researchdomain.WebSearchResult {
	results := make([]researchdomain.WebSearchResult, n)
	for i := 0; i < n; i++ {
		results[i] = researchdomain.NewWebSearchResult(
			fmt.Sprintf("https://example.com/%d", i),
			fmt.Sprintf("Result %d", i),
			fmt.Sprintf("Snippet %d", i),
		)
	}
	return results
}

// ---------------------------------------------------------------------------
// Mock LLM Client
// ---------------------------------------------------------------------------

type mockLLMClient struct {
	textCompletionCalls int
	textCompletionFn    func(ctx context.Context, prompt string) (llm.Response, error)
}

func (m *mockLLMClient) StructuredOutput(_ context.Context, _ string, _ map[string]any) (llm.Response, error) {
	return llm.Response{}, nil
}

func (m *mockLLMClient) TextCompletion(ctx context.Context, prompt string) (llm.Response, error) {
	m.textCompletionCalls++
	if m.textCompletionFn != nil {
		return m.textCompletionFn(ctx, prompt)
	}
	return llm.Response{}, nil
}

// ---------------------------------------------------------------------------
// Mock Web Search
// ---------------------------------------------------------------------------

type mockWebSearch struct {
	calls    int
	searchFn func(ctx context.Context, query string, maxResults int) ([]researchdomain.WebSearchResult, error)
}

func (m *mockWebSearch) Search(ctx context.Context, query string, maxResults int) ([]researchdomain.WebSearchResult, error) {
	m.calls++
	if m.searchFn != nil {
		return m.searchFn(ctx, query, maxResults)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// NoopResearch
// ---------------------------------------------------------------------------

func TestNoopResearchSatisfiesPort(t *testing.T) {
	t.Parallel()
	var _ researchapp.DomainResearch = (*infrastructure.NoopResearchAdapter)(nil)
}

func TestNoopResearchReturnsEmptyBriefing(t *testing.T) {
	t.Parallel()
	adapter := &infrastructure.NoopResearchAdapter{}
	model := ddd.NewDomainModel("empty")
	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.Empty(t, briefing.Findings())
	assert.Empty(t, briefing.Summary())
}

func TestNoopResearchContextNamesAsNoData(t *testing.T) {
	t.Parallel()
	adapter := &infrastructure.NoopResearchAdapter{}
	model := makeModel(t, "Sales", "Billing")
	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	noData := briefing.NoDataAreas()
	assert.Contains(t, noData, "Sales")
	assert.Contains(t, noData, "Billing")
}

func TestNoopResearchEmptyModelReturnsEmptyNoData(t *testing.T) {
	t.Parallel()
	adapter := &infrastructure.NoopResearchAdapter{}
	model := ddd.NewDomainModel("empty")
	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.Empty(t, briefing.NoDataAreas())
}

// ---------------------------------------------------------------------------
// DuckDuckGoSearch
// ---------------------------------------------------------------------------

func TestDDGSSearchSatisfiesPort(t *testing.T) {
	t.Parallel()
	var _ researchapp.WebSearch = (*infrastructure.DuckDuckGoSearchAdapter)(nil)
}

// ---------------------------------------------------------------------------
// RlmResearchAdapter
// ---------------------------------------------------------------------------

func TestRlmResearchSatisfiesPort(t *testing.T) {
	t.Parallel()
	var _ researchapp.DomainResearch = (*infrastructure.RlmResearchAdapter)(nil)
}

func TestRlmResearchReturnsBriefingWithFindings(t *testing.T) {
	t.Parallel()
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			return makeSearchResults(2), nil
		},
	}
	llmClient := &mockLLMClient{
		textCompletionFn: func(_ context.Context, _ string) (llm.Response, error) {
			return llm.NewResponse("Refined query: Sales competition analysis\nSynthesis: Key patterns found.", "test-model", 100), nil
		},
	}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales")

	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, briefing.Findings())
	assert.Empty(t, briefing.NoDataAreas())
}

func TestRlmResearchSummaryFromLLM(t *testing.T) {
	t.Parallel()
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			return makeSearchResults(2), nil
		},
	}
	callIdx := 0
	llmClient := &mockLLMClient{
		textCompletionFn: func(_ context.Context, _ string) (llm.Response, error) {
			callIdx++
			if callIdx == 1 {
				return llm.NewResponse("Refined query: test\nSynthesis: done", "test-model", 50), nil
			}
			return llm.NewResponse("Key patterns across Sales domain.", "test-model", 30), nil
		},
	}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales")

	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.Equal(t, "Key patterns across Sales domain.", briefing.Summary())
}

func TestRlmResearchFiltersEmptyURLOrSnippet(t *testing.T) {
	t.Parallel()
	resultsWithGaps := []researchdomain.WebSearchResult{
		researchdomain.NewWebSearchResult("https://example.com/1", "Good", "Content"),
		researchdomain.NewWebSearchResult("", "No URL", "Has snippet"),
		researchdomain.NewWebSearchResult("https://example.com/3", "No Snippet", ""),
	}
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			return resultsWithGaps, nil
		},
	}
	llmClient := &mockLLMClient{
		textCompletionFn: func(_ context.Context, _ string) (llm.Response, error) {
			return llm.NewResponse("Refined query: test\nSynthesis: done", "test-model", 50), nil
		},
	}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales")

	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	for _, finding := range briefing.Findings() {
		assert.NotEmpty(t, finding.Source().URL())
		assert.NotEmpty(t, finding.Content())
	}
}

func TestRlmResearchEveryFindingHasSourceURL(t *testing.T) {
	t.Parallel()
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			return makeSearchResults(3), nil
		},
	}
	llmClient := &mockLLMClient{
		textCompletionFn: func(_ context.Context, _ string) (llm.Response, error) {
			return llm.NewResponse("Refined query: test\nSynthesis: patterns", "test-model", 50), nil
		},
	}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales")

	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	for _, finding := range briefing.Findings() {
		assert.NotEmpty(t, finding.Source().URL(), "finding should have source URL")
	}
}

func TestRlmResearchIterativeSearchAtLeastTwoCalls(t *testing.T) {
	t.Parallel()
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			return makeSearchResults(2), nil
		},
	}
	llmClient := &mockLLMClient{
		textCompletionFn: func(_ context.Context, _ string) (llm.Response, error) {
			return llm.NewResponse("Refined query: deeper Sales analysis\nSynthesis: done", "test-model", 50), nil
		},
	}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales")

	_, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, webSearch.calls, 2, "RLM pattern requires at least 2 search rounds")
}

// ---------------------------------------------------------------------------
// Degradation
// ---------------------------------------------------------------------------

func TestRlmResearchLLMUnavailableReturnsRawResults(t *testing.T) {
	t.Parallel()
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			return makeSearchResults(2), nil
		},
	}
	llmClient := &mockLLMClient{
		textCompletionFn: func(_ context.Context, _ string) (llm.Response, error) {
			return llm.Response{}, fmt.Errorf("No API key: %w", llm.ErrLLMUnavailable)
		},
	}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales")

	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, briefing.Findings())
	for _, finding := range briefing.Findings() {
		assert.Equal(t, researchdomain.ConfidenceLow, finding.Source().Confidence())
		assert.Equal(t, researchdomain.TrustAIInferred, finding.TrustLevel())
	}
}

func TestRlmResearchCompleteSearchFailureAllNoData(t *testing.T) {
	t.Parallel()
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			return nil, nil
		},
	}
	llmClient := &mockLLMClient{}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales", "Billing")

	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.Empty(t, briefing.Findings())
	assert.Contains(t, briefing.NoDataAreas(), "Sales")
	assert.Contains(t, briefing.NoDataAreas(), "Billing")
}

func TestRlmResearchPartialFailureMixedResults(t *testing.T) {
	t.Parallel()
	callIdx := 0
	webSearch := &mockWebSearch{
		searchFn: func(_ context.Context, _ string, _ int) ([]researchdomain.WebSearchResult, error) {
			callIdx++
			switch callIdx {
			case 1: // Sales round 1
				return makeSearchResults(2), nil
			case 2: // Sales round 2
				return makeSearchResults(1), nil
			default: // Billing round 1
				return nil, nil
			}
		},
	}
	llmClient := &mockLLMClient{
		textCompletionFn: func(_ context.Context, _ string) (llm.Response, error) {
			return llm.NewResponse("Refined query: test\nSynthesis: done", "test-model", 50), nil
		},
	}
	adapter := infrastructure.NewRlmResearchAdapter(llmClient, webSearch)
	model := makeModel(t, "Sales", "Billing")

	briefing, err := adapter.Research(context.Background(), model, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, briefing.Findings())
	assert.Contains(t, briefing.NoDataAreas(), "Billing")
	findingAreas := make(map[string]bool)
	for _, f := range briefing.Findings() {
		findingAreas[f.DomainArea()] = true
	}
	assert.True(t, findingAreas["Sales"])
}
