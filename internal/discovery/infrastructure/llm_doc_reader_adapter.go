package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

// Compile-time interface check.
var _ discoveryapp.LLMDocReader = (*LLMDocReaderAdapter)(nil)

const inferencePrompt = `Analyze the following project documentation and extract DDD (Domain-Driven Design) elements.
Return a JSON object with:
- bounded_contexts: array of {name, responsibility, classification} where classification is "core", "supporting", or "generic"
- domain_stories: array of {title, actors, trigger, steps}
- domain_events: array of event name strings
- actors: array of actor name strings
- entities: array of {name, type} where type is "aggregate", "entity", or "value_object"

Documentation:
`

// LLMDocReaderAdapter implements LLMDocReader by calling an LLM client.
type LLMDocReaderAdapter struct {
	client llm.Client
}

// NewLLMDocReaderAdapter creates an adapter that uses the given LLM client.
func NewLLMDocReaderAdapter(client llm.Client) *LLMDocReaderAdapter {
	return &LLMDocReaderAdapter{client: client}
}

// inferenceResponse is the expected JSON structure from the LLM.
type inferenceResponse struct {
	BoundedContexts []inferredContext `json:"bounded_contexts"`
	DomainStories   []inferredStory   `json:"domain_stories"`
	DomainEvents    []string          `json:"domain_events"`
	Actors          []string          `json:"actors"`
	Entities        []inferredEntity  `json:"entities"`
}

type inferredContext struct {
	Name           string `json:"name"`
	Responsibility string `json:"responsibility"`
	Classification string `json:"classification"`
}

type inferredStory struct {
	Title   string   `json:"title"`
	Actors  []string `json:"actors"`
	Trigger string   `json:"trigger"`
	Steps   []string `json:"steps"`
}

type inferredEntity struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// InferModel sends doc contents to the LLM and parses the response into a DomainModel.
func (a *LLMDocReaderAdapter) InferModel(ctx context.Context, docs map[string]string) (*discoverydomain.InferenceResult, error) {
	prompt := buildInferencePrompt(docs)
	schema := inferenceSchema()

	resp, err := a.client.StructuredOutput(ctx, prompt, schema)
	if err != nil {
		return nil, fmt.Errorf("calling LLM: %w", err)
	}

	var parsed inferenceResponse
	if jsonErr := json.Unmarshal([]byte(resp.Content()), &parsed); jsonErr != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", jsonErr)
	}

	model, buildErr := buildDomainModel(parsed)
	if buildErr != nil {
		return nil, fmt.Errorf("building domain model: %w", buildErr)
	}

	sourceDocs := sortedKeys(docs)

	result, resultErr := discoverydomain.NewInferenceResult(model, "high", sourceDocs)
	if resultErr != nil {
		return nil, fmt.Errorf("creating inference result: %w", resultErr)
	}
	return result, nil
}

func buildInferencePrompt(docs map[string]string) string {
	var b strings.Builder
	b.WriteString(inferencePrompt)

	keys := sortedKeys(docs)
	for _, name := range keys {
		fmt.Fprintf(&b, "\n--- %s ---\n%s\n", name, docs[name])
	}
	return b.String()
}

func inferenceSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"bounded_contexts": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":           map[string]any{"type": "string"},
						"responsibility": map[string]any{"type": "string"},
						"classification": map[string]any{"type": "string", "enum": []string{"core", "supporting", "generic"}},
					},
				},
			},
			"domain_stories": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title":   map[string]any{"type": "string"},
						"actors":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"trigger": map[string]any{"type": "string"},
						"steps":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"domain_events": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"actors":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"entities": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"type": map[string]any{"type": "string", "enum": []string{"aggregate", "entity", "value_object"}},
					},
				},
			},
		},
	}
}

func buildDomainModel(parsed inferenceResponse) (*ddd.DomainModel, error) {
	model := ddd.NewDomainModel("llm-inferred")

	for _, ctx := range parsed.BoundedContexts {
		classification := parseSubdomainClassification(ctx.Classification)
		bc := vo.NewDomainBoundedContext(ctx.Name, ctx.Responsibility, nil, classification, "")
		if err := model.AddBoundedContext(bc); err != nil {
			return nil, fmt.Errorf("adding context %q: %w", ctx.Name, err)
		}
	}

	for _, story := range parsed.DomainStories {
		ds := vo.NewDomainStory(story.Title, story.Actors, story.Trigger, story.Steps, nil)
		if err := model.AddDomainStory(ds); err != nil {
			return nil, fmt.Errorf("adding story %q: %w", story.Title, err)
		}
	}

	return model, nil
}

func parseSubdomainClassification(s string) *vo.SubdomainClassification {
	switch strings.ToLower(s) {
	case "core":
		c := vo.SubdomainCore
		return &c
	case "supporting":
		c := vo.SubdomainSupporting
		return &c
	case "generic":
		c := vo.SubdomainGeneric
		return &c
	default:
		return nil
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
