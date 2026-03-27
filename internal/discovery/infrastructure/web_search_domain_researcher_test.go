package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

// multiCallLLMStub records prompts and returns sequential responses per call.
// DO NOT modify the existing stubLLMClient in llm_boundary_detector_test.go.
type multiCallLLMStub struct {
	responses []llm.Response
	errors    []error
	callIndex int
	prompts   []string
	mu        sync.Mutex
}

func (s *multiCallLLMStub) StructuredOutput(_ context.Context, prompt string, _ map[string]any) (llm.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prompts = append(s.prompts, prompt)

	idx := s.callIndex
	s.callIndex++

	var resp llm.Response
	if idx < len(s.responses) {
		resp = s.responses[idx]
	}

	var err error
	if idx < len(s.errors) {
		err = s.errors[idx]
	}

	return resp, err
}

func (s *multiCallLLMStub) TextCompletion(_ context.Context, _ string) (llm.Response, error) {
	return llm.NewResponse("", "", 0), nil
}

// queryGenJSON builds a valid query-generation LLM response.
func queryGenJSON(queries ...string) string {
	quoted := make([]string, len(queries))
	for i, q := range queries {
		quoted[i] = fmt.Sprintf("%q", q)
	}

	return fmt.Sprintf(`{"queries": [%s]}`, strings.Join(quoted, ", "))
}

// happyExtractionJSON returns a valid extraction response for domain research.
func happyExtractionJSON() string {
	return `{
	"actors": [
		{"name": "Warehouse Manager", "role": "oversees inventory", "source_urls": ["https://example.com/a"]}
	],
	"entities": [
		{"name": "Inventory Item", "properties": ["sku", "quantity"], "source_urls": ["https://example.com/b"]}
	],
	"workflows": [
		{
			"name": "Receive Shipment",
			"type": "happy_path",
			"steps": [
				{"sequence": 1, "actor": "Warehouse Manager", "activity": "inspects", "work_object": "Shipment"}
			],
			"source_urls": ["https://example.com/c"]
		}
	],
	"failure_modes": ["stock discrepancy"],
	"regulatory": [
		{"name": "OSHA Safety", "description": "workplace safety requirements", "source_urls": ["https://example.com/d"]}
	],
	"software": [
		{"name": "SAP WM", "description": "warehouse management module", "source_url": "https://example.com/e"}
	]
}`
}

// newTestDDGServer returns an httptest.Server that responds to search queries.
// queryResponses maps query substrings to HTML bodies. Unmatched queries get 500.
func newTestDDGServer(queryResponses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		for substr, body := range queryResponses {
			if strings.Contains(query, substr) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, body)

				return
			}
		}
		// No match: simulate failure.
		w.WriteHeader(http.StatusInternalServerError)
	}))
}

func TestWebSearchDomainResearcher_HappyPath(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations", "warehouse roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	server := newTestDDGServer(map[string]string{
		"warehouse workflow":    "<html>workflow info</html>",
		"warehouse software":    "<html>software info</html>",
		"warehouse regulations": "<html>regulations info</html>",
		"warehouse roles":       "<html>roles info</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "warehouse management", result.Domain())
	assert.NotEmpty(t, result.Actors())
	assert.NotEmpty(t, result.Entities())
	assert.NotEmpty(t, result.Workflows())
	assert.NotEmpty(t, result.FailureModes())
	assert.NotEmpty(t, result.Regulatory())
	assert.NotEmpty(t, result.Software())
}

func TestWebSearchDomainResearcher_LLMUnavailableAtQueryGen(t *testing.T) {
	t.Parallel()

	stub := &multiCallLLMStub{
		errors: []error{llm.ErrLLMUnavailable},
	}

	researcher := NewWebSearchDomainResearcher(stub, &http.Client{})

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestWebSearchDomainResearcher_LLMUnavailableAtExtraction(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			{},
		},
		errors: []error{nil, llm.ErrLLMUnavailable},
	}

	server := newTestDDGServer(map[string]string{
		"warehouse workflow":    "<html>workflow</html>",
		"warehouse software":    "<html>software</html>",
		"warehouse regulations": "<html>regulations</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestWebSearchDomainResearcher_AllQueriesFail(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
		},
		errors: []error{nil},
	}

	// Server returns 500 for all queries (no matches in the map).
	server := newTestDDGServer(map[string]string{})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestWebSearchDomainResearcher_FewerThanTwoResults(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations", "warehouse roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
		},
		errors: []error{nil},
	}

	// Only 1 out of 4 queries succeeds.
	server := newTestDDGServer(map[string]string{
		"warehouse workflow": "<html>some data</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestWebSearchDomainResearcher_PartialHTTPFailures(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations", "warehouse roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// 2 succeed, 2 fail — exactly at minQueriesWithResults threshold.
	server := newTestDDGServer(map[string]string{
		"warehouse workflow": "<html>workflow data</html>",
		"warehouse software": "<html>software data</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestWebSearchDomainResearcher_RawHTMLTruncated(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// Return a body much larger than maxRawHTMLBytes (4096).
	largeBody := strings.Repeat("X", 10000)
	// Need at least 2 results to pass minQueriesWithResults; add a second query.
	queries = []string{"warehouse workflow", "warehouse software"}
	stub.responses[0] = llm.NewResponse(queryGenJSON(queries...), "test-model", 100)

	server := newTestDDGServer(map[string]string{
		"warehouse workflow": largeBody,
		"warehouse software": largeBody,
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	require.NotNil(t, result)

	// The extraction prompt (second call) should not contain the full 10000-byte body.
	stub.mu.Lock()
	extractionPrompt := stub.prompts[1]
	stub.mu.Unlock()

	// Each query result is truncated to 4096 bytes, so the prompt should not
	// contain the full 10000-byte string.
	assert.NotContains(t, extractionPrompt, largeBody)
	// But it should contain a truncated portion.
	assert.Contains(t, extractionPrompt, strings.Repeat("X", 100))
}

func TestWebSearchDomainResearcher_NonAvailabilityLLMError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		errors    []error
		responses []llm.Response
		wantCtx   string
	}{
		{
			name:    "error at query generation",
			errors:  []error{errors.New("connection timeout")},
			wantCtx: "generating search queries",
		},
		{
			name:   "error at extraction",
			errors: []error{nil, errors.New("rate limit exceeded")},
			responses: []llm.Response{
				llm.NewResponse(queryGenJSON("warehouse workflow", "warehouse software"), "test-model", 100),
				{},
			},
			wantCtx: "extracting domain knowledge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stub := &multiCallLLMStub{
				responses: tt.responses,
				errors:    tt.errors,
			}

			// For extraction error case, need an HTTP server that returns results.
			server := newTestDDGServer(map[string]string{
				"warehouse workflow": "<html>data</html>",
				"warehouse software": "<html>data</html>",
			})
			defer server.Close()

			researcher := NewWebSearchDomainResearcher(stub, server.Client())
			researcher.searchURL = server.URL + "/"

			result, err := researcher.Research(context.Background(), "warehouse management")

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.wantCtx)
		})
	}
}

func TestWebSearchDomainResearcher_EmptyDomain(t *testing.T) {
	t.Parallel()

	stub := &multiCallLLMStub{}

	researcher := NewWebSearchDomainResearcher(stub, &http.Client{})

	result, err := researcher.Research(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, result)

	// No LLM calls should have been made.
	stub.mu.Lock()
	callCount := stub.callIndex
	stub.mu.Unlock()
	assert.Equal(t, 0, callCount)
}

func TestWebSearchDomainResearcher_MalformedExtractionJSON(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse("not json at all", "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	server := newTestDDGServer(map[string]string{
		"warehouse workflow": "<html>data</html>",
		"warehouse software": "<html>data</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "parsing extraction response")
}

func TestWebSearchDomainResearcher_SearchMetadata(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// 2 out of 3 succeed.
	server := newTestDDGServer(map[string]string{
		"warehouse workflow": "<html>workflow</html>",
		"warehouse software": "<html>software</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	require.NotNil(t, result)

	meta := result.SearchMetadata()
	assert.Equal(t, queries, meta.QueriesUsed())
	assert.Positive(t, meta.Duration())
	assert.Equal(t, 3, meta.TotalSources())
	assert.Equal(t, 2, meta.UsefulSources())
}

// ---------------------------------------------------------------------------
// Degradation edge-case tests (QA: alty-cli-1wu.18)
// ---------------------------------------------------------------------------

func TestWebSearchDomainResearcher_Degradation_ExactlyTwoQueriesSucceed(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations", "warehouse roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// Exactly 2 of 4 queries match — the minimum threshold.
	server := newTestDDGServer(map[string]string{
		"warehouse workflow": "<html>workflow data</html>",
		"warehouse software": "<html>software data</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	require.NotNil(t, result, "result must not be nil when exactly 2 queries succeed")

	// Verify extraction was called: callIndex == 2 means queryGen (0) + extraction (1).
	stub.mu.Lock()
	callCount := stub.callIndex
	stub.mu.Unlock()
	assert.Equal(t, 2, callCount, "expected exactly 2 LLM calls: query generation + extraction")
}

func TestWebSearchDomainResearcher_Degradation_ContextCancelledBeforeAnyCall(t *testing.T) {
	t.Parallel()

	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON("warehouse workflow"), "test-model", 100),
		},
		errors: []error{nil},
	}

	researcher := NewWebSearchDomainResearcher(stub, &http.Client{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately before any call.

	result, err := researcher.Research(ctx, "warehouse management")

	// The cancelled context propagates through the LLM StructuredOutput call.
	// The stub ignores context, so we check if the error propagates from the
	// adapter's perspective. Since the stub ignores ctx, queryGen succeeds
	// and searches fail due to cancelled ctx. With 0 results < 2, we get (nil, nil).
	// Either an error or (nil, nil) is acceptable — the key invariant is no panic
	// and no LLM calls beyond what the cancelled context permits.
	if err != nil {
		assert.Nil(t, result)
	} else {
		// If no error, result should be nil since searches can't succeed.
		assert.Nil(t, result)
	}

	// The stub ignores context, so queryGen may still fire. But with a real
	// HTTP client and a cancelled context, no HTTP requests will succeed.
	// Verify we didn't panic and the method returned cleanly.
}

func TestWebSearchDomainResearcher_Degradation_ContextCancelledMidSearch(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations", "warehouse roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// Channel to gate when the server should block.
	gate := make(chan struct{})
	requestCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		count := requestCount
		mu.Unlock()

		if count <= 2 {
			// First 2 requests succeed immediately.
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "<html>data</html>")

			return
		}
		// Remaining requests block until gate is closed (simulating slow response).
		<-gate
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html>late data</html>")
	}))
	defer server.Close()
	defer close(gate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	// The adapter iterates queries sequentially, so the first 2 succeed,
	// then the 3rd blocks. Cancel context while the 3rd is blocked.
	go func() {
		// Give time for the first 2 requests to complete and 3rd to block.
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := researcher.Research(ctx, "warehouse management")
	// The adapter should not panic. With 2 successful results >= minQueriesWithResults,
	// extraction may proceed if context cancellation races favorably.
	// Or the 3rd/4th search fails and we get 2 results -> extraction proceeds.
	// Or context is cancelled before extraction -> error.
	// All of these are acceptable; the key is no panic.
	if err != nil {
		assert.Nil(t, result, "on error, result must be nil")
	}
	// If no error and result is non-nil, verify it's structurally valid.
	if err == nil && result != nil {
		assert.Equal(t, "warehouse management", result.Domain())
	}
}

func TestWebSearchDomainResearcher_Degradation_HTTPTimeoutOnOneQuery(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if strings.Contains(query, "warehouse+regulations") || strings.Contains(query, "warehouse regulations") {
			// This query blocks long enough to trigger HTTP client timeout.
			time.Sleep(500 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "<html>regulations</html>")

			return
		}
		// Other queries succeed immediately.
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html>data for "+query+"</html>")
	}))
	defer server.Close()

	// Short timeout HTTP client: 100ms will timeout the slow query.
	httpClient := &http.Client{Timeout: 100 * time.Millisecond}

	researcher := NewWebSearchDomainResearcher(stub, httpClient)
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	require.NotNil(t, result, "should succeed: 2 of 3 queries returned within timeout")

	// Verify extraction was called (2 LLM calls: queryGen + extraction).
	stub.mu.Lock()
	callCount := stub.callIndex
	stub.mu.Unlock()
	assert.Equal(t, 2, callCount)
}

func TestWebSearchDomainResearcher_Degradation_DDGReturnsEmptyBody(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// One query returns empty body with 200 OK, two return real content.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if strings.Contains(query, "regulations") {
			// 200 OK with empty body.
			w.WriteHeader(http.StatusOK)

			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html>data for "+query+"</html>")
	}))
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err)
	// FINDING: The adapter treats 200-OK-empty-body as a successful result.
	// doSearch returns ("", nil) for empty bodies. executeSearches counts this
	// as a success and includes searchResult{query: q, body: ""} in results.
	// This means an empty-body query inflates the success count.
	// All 3 queries "succeed" (2 real + 1 empty), so extraction proceeds.
	require.NotNil(t, result, "adapter treats empty 200 as success — extraction proceeds")

	// Verify the extraction prompt includes the empty body as a search result.
	stub.mu.Lock()
	extractionPrompt := stub.prompts[1]
	stub.mu.Unlock()
	assert.Contains(t, extractionPrompt, "regulations",
		"empty-body query should still appear in extraction prompt")
}

func TestWebSearchDomainResearcher_Degradation_AllQueriesReturn500(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software", "warehouse regulations", "warehouse roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
		},
		errors: []error{nil},
	}

	// All queries return explicit 500 status codes.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err, "all-500 is graceful degradation, not an error")
	assert.Nil(t, result, "should return (nil, nil) when 0 < minQueriesWithResults")

	// Verify extraction was NOT called (only 1 LLM call: queryGen).
	stub.mu.Lock()
	callCount := stub.callIndex
	stub.mu.Unlock()
	assert.Equal(t, 1, callCount, "only query generation should have been called")
}

func TestWebSearchDomainResearcher_Degradation_NonLLMErrorAtQueryGenPropagated(t *testing.T) {
	t.Parallel()

	customErr := errors.New("API key expired")
	stub := &multiCallLLMStub{
		errors: []error{customErr},
	}

	researcher := NewWebSearchDomainResearcher(stub, &http.Client{})

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "generating search queries",
		"error must include query generation context")
	assert.ErrorIs(t, err, customErr,
		"original error must be in the chain via wrapping")
}

func TestWebSearchDomainResearcher_Degradation_NonLLMErrorAtExtractionPropagated(t *testing.T) {
	t.Parallel()

	queries := []string{"warehouse workflow", "warehouse software"}
	customErr := errors.New("rate limit exceeded")
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			{},
		},
		errors: []error{nil, customErr},
	}

	server := newTestDDGServer(map[string]string{
		"warehouse workflow": "<html>data</html>",
		"warehouse software": "<html>data</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "extracting domain knowledge",
		"error must include extraction context")
	assert.ErrorIs(t, err, customErr,
		"original error must be in the chain via wrapping")
}

func TestWebSearchDomainResearcher_Degradation_WhitespaceOnlyDomain(t *testing.T) {
	t.Parallel()

	stub := &multiCallLLMStub{}

	researcher := NewWebSearchDomainResearcher(stub, &http.Client{})

	result, err := researcher.Research(context.Background(), "   ")

	require.Error(t, err, "whitespace-only domain should be rejected like empty string")
	assert.Nil(t, result)

	// Verify no LLM calls were made (validation happens before any I/O).
	stub.mu.Lock()
	callCount := stub.callIndex
	stub.mu.Unlock()
	assert.Equal(t, 0, callCount, "no LLM calls should be made for whitespace-only domain")
}

func TestWebSearchDomainResearcher_Degradation_LLMReturnsEmptyQueryList(t *testing.T) {
	t.Parallel()

	// LLM returns valid JSON but with an empty queries array.
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(`{"queries": []}`, "test-model", 100),
		},
		errors: []error{nil},
	}

	server := newTestDDGServer(map[string]string{
		"warehouse workflow": "<html>data</html>",
	})
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")

	require.NoError(t, err, "empty query list is graceful degradation, not an error")
	assert.Nil(t, result, "no queries means 0 results < minQueriesWithResults -> (nil, nil)")

	// Verify only 1 LLM call (queryGen) and no extraction call.
	stub.mu.Lock()
	callCount := stub.callIndex
	stub.mu.Unlock()
	assert.Equal(t, 1, callCount, "only query generation should have been called")
}

// ---------------------------------------------------------------------------
// Pipeline correctness tests (QA: alty-cli-1wu.18)
// ---------------------------------------------------------------------------

// richExtractionJSON returns an extraction response with multiple items in every
// VO category, including multi-step workflows.
func richExtractionJSON() string {
	return `{
	"actors": [
		{"name": "Warehouse Manager", "role": "oversees inventory", "source_urls": ["https://example.com/a1", "https://example.com/a2"]},
		{"name": "Picker", "role": "picks items from shelves", "source_urls": ["https://example.com/a3"]}
	],
	"entities": [
		{"name": "Inventory Item", "properties": ["sku", "quantity", "location"], "source_urls": ["https://example.com/e1"]},
		{"name": "Purchase Order", "properties": ["order_id", "supplier", "status"], "source_urls": ["https://example.com/e2", "https://example.com/e3"]}
	],
	"workflows": [
		{
			"name": "Receive Shipment",
			"type": "happy_path",
			"steps": [
				{"sequence": 1, "actor": "Warehouse Manager", "activity": "inspects", "work_object": "Shipment"},
				{"sequence": 2, "actor": "Picker", "activity": "shelves", "work_object": "Inventory Item"}
			],
			"source_urls": ["https://example.com/w1"]
		},
		{
			"name": "Handle Damage",
			"type": "failure_case",
			"steps": [
				{"sequence": 1, "actor": "Warehouse Manager", "activity": "reports", "work_object": "Damaged Goods"}
			],
			"source_urls": ["https://example.com/w2", "https://example.com/w3"]
		}
	],
	"failure_modes": ["stock discrepancy", "shipment delay", "mislabeled item"],
	"regulatory": [
		{"name": "OSHA Safety", "description": "workplace safety requirements", "source_urls": ["https://example.com/r1"]},
		{"name": "FDA Cold Chain", "description": "temperature-controlled storage rules", "source_urls": ["https://example.com/r2"]}
	],
	"software": [
		{"name": "SAP WM", "description": "warehouse management module", "source_url": "https://example.com/s1"},
		{"name": "Manhattan WMS", "description": "cloud warehouse system", "source_url": "https://example.com/s2"}
	]
}`
}

// pipelineSetup creates a standard researcher wired to the given LLM stub and
// a DDG server that succeeds for the provided query substrings.
func pipelineSetup(stub *multiCallLLMStub, queryMap map[string]string) (*WebSearchDomainResearcher, *httptest.Server) {
	server := newTestDDGServer(queryMap)
	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	return researcher, server
}

func TestWebSearchDomainResearcher_Pipeline_AllDomainVOsPopulated(t *testing.T) {
	t.Parallel()

	queries := []string{"wh workflow", "wh software", "wh regulations", "wh roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(richExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"wh workflow":    "<html>workflow info</html>",
		"wh software":    "<html>software info</html>",
		"wh regulations": "<html>regulations info</html>",
		"wh roles":       "<html>roles info</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "warehouse management")
	require.NoError(t, err)
	require.NotNil(t, result)

	// --- Actors ---
	actors := result.Actors()
	assert.Len(t, actors, 2)
	assert.Equal(t, "Warehouse Manager", actors[0].Name())
	assert.Equal(t, "oversees inventory", actors[0].Role())
	assert.Equal(t, []string{"https://example.com/a1", "https://example.com/a2"}, actors[0].SourceURLs())
	assert.Equal(t, "Picker", actors[1].Name())
	assert.Equal(t, "picks items from shelves", actors[1].Role())
	assert.Equal(t, []string{"https://example.com/a3"}, actors[1].SourceURLs())

	// --- Entities ---
	entities := result.Entities()
	assert.Len(t, entities, 2)
	assert.Equal(t, "Inventory Item", entities[0].Name())
	assert.Equal(t, []string{"sku", "quantity", "location"}, entities[0].Properties())
	assert.Equal(t, []string{"https://example.com/e1"}, entities[0].SourceURLs())
	assert.Equal(t, "Purchase Order", entities[1].Name())
	assert.Equal(t, []string{"order_id", "supplier", "status"}, entities[1].Properties())
	assert.Equal(t, []string{"https://example.com/e2", "https://example.com/e3"}, entities[1].SourceURLs())

	// --- Workflows ---
	workflows := result.Workflows()
	assert.Len(t, workflows, 2)

	wf0 := workflows[0]
	assert.Equal(t, "Receive Shipment", wf0.Name())
	assert.Equal(t, "happy_path", wf0.WorkflowType().String())
	assert.Equal(t, []string{"https://example.com/w1"}, wf0.SourceURLs())
	steps0 := wf0.Steps()
	assert.Len(t, steps0, 2)
	assert.Equal(t, 1, steps0[0].Sequence())
	assert.Equal(t, "Warehouse Manager", steps0[0].Actor())
	assert.Equal(t, "inspects", steps0[0].Activity())
	assert.Equal(t, "Shipment", steps0[0].WorkObject())
	assert.Equal(t, 2, steps0[1].Sequence())
	assert.Equal(t, "Picker", steps0[1].Actor())
	assert.Equal(t, "shelves", steps0[1].Activity())
	assert.Equal(t, "Inventory Item", steps0[1].WorkObject())

	wf1 := workflows[1]
	assert.Equal(t, "Handle Damage", wf1.Name())
	assert.Equal(t, "failure_case", wf1.WorkflowType().String())
	assert.Equal(t, []string{"https://example.com/w2", "https://example.com/w3"}, wf1.SourceURLs())
	steps1 := wf1.Steps()
	assert.Len(t, steps1, 1)
	assert.Equal(t, "Warehouse Manager", steps1[0].Actor())
	assert.Equal(t, "reports", steps1[0].Activity())
	assert.Equal(t, "Damaged Goods", steps1[0].WorkObject())

	// --- FailureModes ---
	assert.Equal(t, []string{"stock discrepancy", "shipment delay", "mislabeled item"}, result.FailureModes())

	// --- Regulatory ---
	regs := result.Regulatory()
	assert.Len(t, regs, 2)
	assert.Equal(t, "OSHA Safety", regs[0].Name())
	assert.Equal(t, "workplace safety requirements", regs[0].Description())
	assert.Equal(t, []string{"https://example.com/r1"}, regs[0].SourceURLs())
	assert.Equal(t, "FDA Cold Chain", regs[1].Name())
	assert.Equal(t, "temperature-controlled storage rules", regs[1].Description())
	assert.Equal(t, []string{"https://example.com/r2"}, regs[1].SourceURLs())

	// --- Software ---
	sw := result.Software()
	assert.Len(t, sw, 2)
	assert.Equal(t, "SAP WM", sw[0].Name())
	assert.Equal(t, "warehouse management module", sw[0].Description())
	assert.Equal(t, "https://example.com/s1", sw[0].SourceURL())
	assert.Equal(t, "Manhattan WMS", sw[1].Name())
	assert.Equal(t, "cloud warehouse system", sw[1].Description())
	assert.Equal(t, "https://example.com/s2", sw[1].SourceURL())
}

func TestWebSearchDomainResearcher_Pipeline_SearchMetadataAccurate(t *testing.T) {
	t.Parallel()

	queries := []string{"wh workflow", "wh software", "wh regulations", "wh roles", "wh challenges"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// 3 out of 5 queries succeed.
	researcher, server := pipelineSetup(stub, map[string]string{
		"wh workflow":    "<html>data</html>",
		"wh software":    "<html>data</html>",
		"wh regulations": "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "warehouse management")
	require.NoError(t, err)
	require.NotNil(t, result)

	meta := result.SearchMetadata()
	assert.Equal(t, queries, meta.QueriesUsed())
	assert.Positive(t, meta.Duration())
	// TotalSources = number of queries generated.
	assert.Equal(t, 5, meta.TotalSources())
	// UsefulSources = number of successful HTTP responses.
	assert.Equal(t, 3, meta.UsefulSources())
}

func TestWebSearchDomainResearcher_Pipeline_HTMLTruncationVerified(t *testing.T) {
	t.Parallel()

	queries := []string{"wh workflow", "wh software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// Each server response is 10000 bytes -- larger than maxRawHTMLBytes (4096).
	largeBody := strings.Repeat("Z", 10000)
	researcher, server := pipelineSetup(stub, map[string]string{
		"wh workflow": largeBody,
		"wh software": largeBody,
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "warehouse management")
	require.NoError(t, err)
	require.NotNil(t, result)

	stub.mu.Lock()
	extractionPrompt := stub.prompts[1]
	stub.mu.Unlock()

	// Count occurrences of "Z" in the extraction prompt. Each query result
	// should be truncated to maxRawHTMLBytes (4096), so total Z count must be
	// at most 2 * 4096 = 8192.
	zCount := strings.Count(extractionPrompt, "Z")
	assert.LessOrEqual(t, zCount, 2*4096, "HTML should be truncated to 4096 bytes per query result")
	// Verify truncation actually happened -- we should have exactly 2*4096
	// because io.LimitReader reads exactly maxRawHTMLBytes.
	assert.Equal(t, 2*4096, zCount, "each query result should be truncated to exactly 4096 bytes")
}

func TestWebSearchDomainResearcher_Pipeline_QueryURLEncodingCorrect(t *testing.T) {
	t.Parallel()

	queryWithSpecialChars := "warehouse & logistics \"cold chain\""
	queries := []string{queryWithSpecialChars, "warehouse software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	// Custom server that records received query parameter values.
	var receivedQueries []string
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		mu.Lock()
		receivedQueries = append(receivedQueries, q)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html>data</html>")
	}))
	defer server.Close()

	researcher := NewWebSearchDomainResearcher(stub, server.Client())
	researcher.searchURL = server.URL + "/"

	result, err := researcher.Research(context.Background(), "warehouse management")
	require.NoError(t, err)
	require.NotNil(t, result)

	mu.Lock()
	defer mu.Unlock()

	// The server's r.URL.Query().Get("q") automatically URL-decodes, so we
	// should see the original query strings -- proving they were properly encoded.
	assert.Len(t, receivedQueries, 2)
	assert.Equal(t, queryWithSpecialChars, receivedQueries[0])
	assert.Equal(t, "warehouse software", receivedQueries[1])
}

func TestWebSearchDomainResearcher_Pipeline_ExtractionMapsToCorrectVOConstructors(t *testing.T) {
	t.Parallel()

	// Extraction JSON with known exact values -- verify every field survives the
	// pipeline mapping without mutation.
	extractionJSON := `{
	"actors": [
		{"name": "QA Lead", "role": "validates releases", "source_urls": ["https://qa.example.com"]}
	],
	"entities": [
		{"name": "Test Suite", "properties": ["suite_id", "coverage_pct"], "source_urls": ["https://ts.example.com"]}
	],
	"workflows": [
		{
			"name": "Release Validation",
			"type": "secondary",
			"steps": [
				{"sequence": 42, "actor": "QA Lead", "activity": "executes", "work_object": "Test Suite"}
			],
			"source_urls": ["https://rv.example.com"]
		}
	],
	"failure_modes": ["flaky test"],
	"regulatory": [
		{"name": "SOC 2", "description": "audit compliance", "source_urls": ["https://soc2.example.com"]}
	],
	"software": [
		{"name": "Selenium", "description": "browser automation", "source_url": "https://selenium.example.com"}
	]
}`

	queries := []string{"qa workflow", "qa software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(extractionJSON, "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"qa workflow": "<html>data</html>",
		"qa software": "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "quality assurance")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Exact field matching -- not just non-empty checks.
	require.Len(t, result.Actors(), 1)
	assert.Equal(t, "QA Lead", result.Actors()[0].Name())
	assert.Equal(t, "validates releases", result.Actors()[0].Role())
	assert.Equal(t, []string{"https://qa.example.com"}, result.Actors()[0].SourceURLs())

	require.Len(t, result.Entities(), 1)
	assert.Equal(t, "Test Suite", result.Entities()[0].Name())
	assert.Equal(t, []string{"suite_id", "coverage_pct"}, result.Entities()[0].Properties())
	assert.Equal(t, []string{"https://ts.example.com"}, result.Entities()[0].SourceURLs())

	require.Len(t, result.Workflows(), 1)
	wf := result.Workflows()[0]
	assert.Equal(t, "Release Validation", wf.Name())
	assert.Equal(t, "secondary", wf.WorkflowType().String())
	assert.Equal(t, []string{"https://rv.example.com"}, wf.SourceURLs())
	require.Len(t, wf.Steps(), 1)
	assert.Equal(t, 42, wf.Steps()[0].Sequence())
	assert.Equal(t, "QA Lead", wf.Steps()[0].Actor())
	assert.Equal(t, "executes", wf.Steps()[0].Activity())
	assert.Equal(t, "Test Suite", wf.Steps()[0].WorkObject())

	assert.Equal(t, []string{"flaky test"}, result.FailureModes())

	require.Len(t, result.Regulatory(), 1)
	assert.Equal(t, "SOC 2", result.Regulatory()[0].Name())
	assert.Equal(t, "audit compliance", result.Regulatory()[0].Description())
	assert.Equal(t, []string{"https://soc2.example.com"}, result.Regulatory()[0].SourceURLs())

	require.Len(t, result.Software(), 1)
	assert.Equal(t, "Selenium", result.Software()[0].Name())
	assert.Equal(t, "browser automation", result.Software()[0].Description())
	assert.Equal(t, "https://selenium.example.com", result.Software()[0].SourceURL())
}

func TestWebSearchDomainResearcher_Pipeline_QualityAutoComputed(t *testing.T) {
	t.Parallel()

	// richExtractionJSON has 2 actors, 2 entities, 2 workflows with 3 total steps.
	// With 4 useful sources: actors < 3, entities < 3, steps < 5 => meetsFloor = false.
	queries := []string{"wh workflow", "wh software", "wh regulations", "wh roles"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(richExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"wh workflow":    "<html>data</html>",
		"wh software":    "<html>data</html>",
		"wh regulations": "<html>data</html>",
		"wh roles":       "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "warehouse management")
	require.NoError(t, err)
	require.NotNil(t, result)

	quality := result.Quality()

	assert.Equal(t, 2, quality.ActorCount())
	assert.Equal(t, 2, quality.EntityCount())
	assert.Equal(t, 3, quality.WorkflowStepCount())
	assert.Equal(t, 4, quality.UsefulSourceCount())
	assert.False(t, quality.MeetsFloor(),
		"2 actors < 3 floor, 2 entities < 3 floor, 3 steps < 5 floor => should not meet floor")
}

func TestWebSearchDomainResearcher_Pipeline_CompileTimeInterfaceCheck(t *testing.T) {
	t.Parallel()

	// Both are non-nil error sentinels.
	require.Error(t, ErrSearchUnavailable)
	require.Error(t, llm.ErrLLMUnavailable)

	// ErrSearchUnavailable and llm.ErrLLMUnavailable must be distinct sentinels.
	require.NotErrorIs(t, ErrSearchUnavailable, llm.ErrLLMUnavailable,
		"ErrSearchUnavailable must not be the same as llm.ErrLLMUnavailable")
	require.NotErrorIs(t, llm.ErrLLMUnavailable, ErrSearchUnavailable,
		"llm.ErrLLMUnavailable must not be the same as ErrSearchUnavailable")

	// Their error strings differ.
	assert.NotEqual(t, ErrSearchUnavailable.Error(), llm.ErrLLMUnavailable.Error())
}

func TestWebSearchDomainResearcher_Pipeline_QueryGenerationPromptContainsDomain(t *testing.T) {
	t.Parallel()

	domainDesc := "veterinary practice management for exotic animals"
	queries := []string{"vet workflow", "vet software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"vet workflow": "<html>data</html>",
		"vet software": "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), domainDesc)
	require.NoError(t, err)
	require.NotNil(t, result)

	stub.mu.Lock()
	queryGenPrompt := stub.prompts[0]
	stub.mu.Unlock()

	assert.Contains(t, queryGenPrompt, domainDesc,
		"query generation prompt must contain the full domain description")
}

func TestWebSearchDomainResearcher_Pipeline_InvalidActorSkipped(t *testing.T) {
	t.Parallel()

	// One actor with empty name (should be skipped by NewResearchedActor),
	// one valid actor.
	extractionJSON := `{
	"actors": [
		{"name": "", "role": "invalid actor", "source_urls": []},
		{"name": "Valid Actor", "role": "does things", "source_urls": ["https://example.com/valid"]}
	],
	"entities": [],
	"workflows": [],
	"failure_modes": [],
	"regulatory": [],
	"software": []
}`

	queries := []string{"test workflow", "test software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(extractionJSON, "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"test workflow": "<html>data</html>",
		"test software": "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "test domain")
	require.NoError(t, err)
	require.NotNil(t, result)

	actors := result.Actors()
	assert.Len(t, actors, 1, "invalid actor with empty name should be skipped")
	assert.Equal(t, "Valid Actor", actors[0].Name())
	assert.Equal(t, "does things", actors[0].Role())
}

func TestWebSearchDomainResearcher_Pipeline_InvalidWorkflowTypeSkipped(t *testing.T) {
	t.Parallel()

	// One workflow with invalid type (should be skipped), one with valid type.
	extractionJSON := `{
	"actors": [],
	"entities": [],
	"workflows": [
		{
			"name": "Bad Workflow",
			"type": "invalid_type",
			"steps": [{"sequence": 1, "actor": "Someone", "activity": "does", "work_object": "Thing"}],
			"source_urls": []
		},
		{
			"name": "Good Workflow",
			"type": "happy_path",
			"steps": [{"sequence": 1, "actor": "Manager", "activity": "approves", "work_object": "Request"}],
			"source_urls": ["https://example.com/good"]
		}
	],
	"failure_modes": [],
	"regulatory": [],
	"software": []
}`

	queries := []string{"test workflow", "test software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(extractionJSON, "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"test workflow": "<html>data</html>",
		"test software": "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "test domain")
	require.NoError(t, err)
	require.NotNil(t, result)

	workflows := result.Workflows()
	assert.Len(t, workflows, 1, "workflow with invalid type should be skipped")
	assert.Equal(t, "Good Workflow", workflows[0].Name())
	assert.Equal(t, "happy_path", workflows[0].WorkflowType().String())
}

func TestWebSearchDomainResearcher_Pipeline_InvalidWorkflowStepSkipped(t *testing.T) {
	t.Parallel()

	// Workflow with one invalid step (empty actor) and one valid step.
	// The workflow itself should still be included with only the valid step.
	extractionJSON := `{
	"actors": [],
	"entities": [],
	"workflows": [
		{
			"name": "Mixed Steps Workflow",
			"type": "secondary",
			"steps": [
				{"sequence": 1, "actor": "", "activity": "fails", "work_object": "Nothing"},
				{"sequence": 2, "actor": "Operator", "activity": "processes", "work_object": "Order"}
			],
			"source_urls": ["https://example.com/mixed"]
		}
	],
	"failure_modes": [],
	"regulatory": [],
	"software": []
}`

	queries := []string{"test workflow", "test software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(extractionJSON, "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"test workflow": "<html>data</html>",
		"test software": "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), "test domain")
	require.NoError(t, err)
	require.NotNil(t, result)

	workflows := result.Workflows()
	require.Len(t, workflows, 1, "workflow should still be included despite one invalid step")
	assert.Equal(t, "Mixed Steps Workflow", workflows[0].Name())

	steps := workflows[0].Steps()
	assert.Len(t, steps, 1, "only valid step should be present")
	assert.Equal(t, 2, steps[0].Sequence())
	assert.Equal(t, "Operator", steps[0].Actor())
	assert.Equal(t, "processes", steps[0].Activity())
	assert.Equal(t, "Order", steps[0].WorkObject())
}

func TestWebSearchDomainResearcher_Pipeline_DomainPassedThroughToResult(t *testing.T) {
	t.Parallel()

	// Use a domain description with special characters to verify exact pass-through.
	domainDesc := "multi-tenant SaaS for B2B e-commerce (2024 edition)"
	queries := []string{"saas workflow", "saas software"}
	stub := &multiCallLLMStub{
		responses: []llm.Response{
			llm.NewResponse(queryGenJSON(queries...), "test-model", 100),
			llm.NewResponse(happyExtractionJSON(), "test-model", 200),
		},
		errors: []error{nil, nil},
	}

	researcher, server := pipelineSetup(stub, map[string]string{
		"saas workflow": "<html>data</html>",
		"saas software": "<html>data</html>",
	})
	defer server.Close()

	result, err := researcher.Research(context.Background(), domainDesc)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, domainDesc, result.Domain(),
		"result.Domain() must exactly match the input domainDescription")
}
