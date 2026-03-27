package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

const (
	maxRawHTMLBytes       = 4096
	minQueriesWithResults = 2
	duckduckgoLiteURL     = "https://lite.duckduckgo.com/lite/"
)

// ErrSearchUnavailable is returned when web search infrastructure is unavailable.
var ErrSearchUnavailable = errors.New("web search unavailable")

// Compile-time interface check.
var _ discoveryapp.DomainResearcher = (*WebSearchDomainResearcher)(nil)

// WebSearchDomainResearcher implements DomainResearcher using web search + LLM extraction.
type WebSearchDomainResearcher struct {
	client     llm.Client
	httpClient *http.Client
	searchURL  string
}

// NewWebSearchDomainResearcher creates a domain researcher that uses web search for
// external domain knowledge and LLM structured output for extraction.
func NewWebSearchDomainResearcher(client llm.Client, httpClient *http.Client) *WebSearchDomainResearcher {
	return &WebSearchDomainResearcher{
		client:     client,
		httpClient: httpClient,
		searchURL:  duckduckgoLiteURL,
	}
}

// searchResult holds raw HTML from a single search query.
type searchResult struct {
	query string
	body  string
}

// Research orchestrates the three-stage pipeline: generate queries, execute
// searches, extract knowledge. Returns (nil, nil) when infrastructure is
// unavailable per ADR-013 graceful degradation.
func (r *WebSearchDomainResearcher) Research(
	ctx context.Context,
	domainDescription string,
) (*discoverydomain.DomainResearchResult, error) {
	if strings.TrimSpace(domainDescription) == "" {
		return nil, fmt.Errorf("domain description must not be empty")
	}

	start := time.Now()

	queries, err := r.generateQueries(ctx, domainDescription)
	if err != nil {
		return nil, err
	}

	if queries == nil {
		return nil, nil
	}

	results, err := r.executeSearches(ctx, queries)
	if err != nil {
		return nil, err
	}

	if len(results) < minQueriesWithResults {
		return nil, nil
	}

	duration := time.Since(start)

	extracted, err := r.extractKnowledge(ctx, domainDescription, results)
	if err != nil {
		return nil, err
	}

	if extracted == nil {
		return nil, nil
	}

	meta := discoverydomain.NewSearchMetadata(queries, len(queries), len(results), duration)

	domainResult, buildErr := r.buildResult(domainDescription, meta, extracted)
	if buildErr != nil {
		return nil, fmt.Errorf("building domain research result: %w", buildErr)
	}

	return domainResult, nil
}

// generateQueries asks the LLM to produce 4-5 search queries for the domain.
func (r *WebSearchDomainResearcher) generateQueries(
	ctx context.Context,
	domainDescription string,
) ([]string, error) {
	prompt := fmt.Sprintf(`Generate 4-5 web search queries to research the following domain. Use these patterns:
- [domain] workflow steps process
- [domain] management software features
- [domain] regulations requirements compliance
- [domain] roles responsibilities
- [domain] typical failures problems challenges

Domain: %s

Return queries as a JSON array.`, domainDescription)

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"queries": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
	}

	resp, err := r.client.StructuredOutput(ctx, prompt, schema)
	if err != nil {
		if errors.Is(err, llm.ErrLLMUnavailable) {
			return nil, nil
		}

		return nil, fmt.Errorf("generating search queries: %w", err)
	}

	var parsed struct {
		Queries []string `json:"queries"`
	}

	if jsonErr := json.Unmarshal([]byte(resp.Content()), &parsed); jsonErr != nil {
		return nil, fmt.Errorf("parsing query generation response: %w", jsonErr)
	}

	return parsed.Queries, nil
}

// executeSearches performs HTTP searches for each query and returns successful results.
func (r *WebSearchDomainResearcher) executeSearches(
	ctx context.Context,
	queries []string,
) ([]searchResult, error) {
	var results []searchResult

	for _, query := range queries {
		body, err := r.doSearch(ctx, query)
		if err != nil {
			continue
		}

		results = append(results, searchResult{query: query, body: body})
	}

	return results, nil
}

// doSearch executes a single HTTP search and returns the truncated body.
func (r *WebSearchDomainResearcher) doSearch(ctx context.Context, query string) (string, error) {
	searchURL := r.searchURL + "?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating search request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing search request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxRawHTMLBytes)

	bodyBytes, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("reading search response: %w", err)
	}

	return string(bodyBytes), nil
}

// extractionResponse is the expected JSON structure from the LLM extraction call.
type extractionResponse struct {
	Actors       []extractedActor    `json:"actors"`
	Entities     []extractedEntity   `json:"entities"`
	Workflows    []extractedWorkflow `json:"workflows"`
	FailureModes []string            `json:"failure_modes"`
	Regulatory   []extractedRegItem  `json:"regulatory"`
	Software     []extractedSoftware `json:"software"`
}

type extractedActor struct {
	Name       string   `json:"name"`
	Role       string   `json:"role"`
	SourceURLs []string `json:"source_urls"`
}

type extractedEntity struct {
	Name       string   `json:"name"`
	Properties []string `json:"properties"`
	SourceURLs []string `json:"source_urls"`
}

type extractedWorkflow struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Steps      []extractedStep `json:"steps"`
	SourceURLs []string        `json:"source_urls"`
}

type extractedStep struct {
	Sequence   int    `json:"sequence"`
	Actor      string `json:"actor"`
	Activity   string `json:"activity"`
	WorkObject string `json:"work_object"`
}

type extractedRegItem struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	SourceURLs  []string `json:"source_urls"`
}

type extractedSoftware struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SourceURL   string `json:"source_url"`
}

// extractKnowledge sends raw search results to the LLM for structured extraction.
func (r *WebSearchDomainResearcher) extractKnowledge(
	ctx context.Context,
	domainDescription string,
	rawResults []searchResult,
) (*extractionResponse, error) {
	prompt := buildExtractionPrompt(domainDescription, rawResults)
	schema := extractionSchema()

	resp, err := r.client.StructuredOutput(ctx, prompt, schema)
	if err != nil {
		if errors.Is(err, llm.ErrLLMUnavailable) {
			return nil, nil
		}

		return nil, fmt.Errorf("extracting domain knowledge: %w", err)
	}

	var parsed extractionResponse
	if jsonErr := json.Unmarshal([]byte(resp.Content()), &parsed); jsonErr != nil {
		return nil, fmt.Errorf("parsing extraction response: %w", jsonErr)
	}

	return &parsed, nil
}

func buildExtractionPrompt(domainDescription string, results []searchResult) string {
	var b strings.Builder

	b.WriteString("Extract structured domain knowledge from the following web search results.\n\n")
	b.WriteString("Domain: ")
	b.WriteString(domainDescription)
	b.WriteString("\n\n")

	for i, res := range results {
		fmt.Fprintf(&b, "--- Search Result %d (query: %s) ---\n%s\n\n", i+1, res.query, res.body)
	}

	b.WriteString("Extract: actors (name, role, source_urls), entities (name, properties, source_urls), ")
	b.WriteString("workflows (name, type, steps, source_urls), failure_modes, ")
	b.WriteString("regulatory (name, description, source_urls), software (name, description, source_url).")

	return b.String()
}

func extractionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"actors": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":        map[string]any{"type": "string"},
						"role":        map[string]any{"type": "string"},
						"source_urls": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"entities": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":        map[string]any{"type": "string"},
						"properties":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"source_urls": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"workflows": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"type": map[string]any{"type": "string"},
						"steps": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"sequence":    map[string]any{"type": "integer"},
									"actor":       map[string]any{"type": "string"},
									"activity":    map[string]any{"type": "string"},
									"work_object": map[string]any{"type": "string"},
								},
							},
						},
						"source_urls": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"failure_modes": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"regulatory": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":        map[string]any{"type": "string"},
						"description": map[string]any{"type": "string"},
						"source_urls": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"software": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":        map[string]any{"type": "string"},
						"description": map[string]any{"type": "string"},
						"source_url":  map[string]any{"type": "string"},
					},
				},
			},
		},
	}
}

// buildResult maps the extracted response into domain value objects and returns
// a DomainResearchResult.
func (r *WebSearchDomainResearcher) buildResult(
	domainDescription string,
	meta discoverydomain.SearchMetadata,
	extracted *extractionResponse,
) (*discoverydomain.DomainResearchResult, error) {
	actors := make([]discoverydomain.ResearchedActor, 0, len(extracted.Actors))
	for _, a := range extracted.Actors {
		actor, err := discoverydomain.NewResearchedActor(a.Name, a.Role, a.SourceURLs)
		if err != nil {
			continue
		}

		actors = append(actors, actor)
	}

	entities := make([]discoverydomain.ResearchedEntity, 0, len(extracted.Entities))
	for _, e := range extracted.Entities {
		entity, err := discoverydomain.NewResearchedEntity(e.Name, e.Properties, e.SourceURLs)
		if err != nil {
			continue
		}

		entities = append(entities, entity)
	}

	workflows := make([]discoverydomain.ResearchedWorkflow, 0, len(extracted.Workflows))
	for _, w := range extracted.Workflows {
		wfType, err := discoverydomain.NewWorkflowType(w.Type)
		if err != nil {
			continue
		}

		steps := make([]discoverydomain.WorkflowStep, 0, len(w.Steps))
		for _, s := range w.Steps {
			step, stepErr := discoverydomain.NewWorkflowStep(s.Sequence, s.Actor, s.Activity, s.WorkObject)
			if stepErr != nil {
				continue
			}

			steps = append(steps, step)
		}

		workflow, wfErr := discoverydomain.NewResearchedWorkflow(w.Name, wfType, steps, w.SourceURLs)
		if wfErr != nil {
			continue
		}

		workflows = append(workflows, workflow)
	}

	regulatory := make([]discoverydomain.RegulatoryItem, 0, len(extracted.Regulatory))
	for _, reg := range extracted.Regulatory {
		item, err := discoverydomain.NewRegulatoryItem(reg.Name, reg.Description, reg.SourceURLs)
		if err != nil {
			continue
		}

		regulatory = append(regulatory, item)
	}

	software := make([]discoverydomain.ExistingSoftware, 0, len(extracted.Software))
	for _, sw := range extracted.Software {
		item, err := discoverydomain.NewExistingSoftware(sw.Name, sw.Description, sw.SourceURL)
		if err != nil {
			continue
		}

		software = append(software, item)
	}

	result, err := discoverydomain.NewDomainResearchResult(
		domainDescription, meta,
		actors, entities, workflows,
		extracted.FailureModes, regulatory, software,
	)
	if err != nil {
		return nil, fmt.Errorf("creating domain research result: %w", err)
	}

	return &result, nil
}
