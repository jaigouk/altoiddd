package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

const boundaryDetectionSystemPrompt = "You are a DDD expert analyzing domain stories for bounded context boundary signals."

const boundaryDetectionUserPrompt = `Analyze these domain stories for two types of boundary signals:

1. Language differences: Same term carries different properties/responsibilities across stories
2. Trigger differences: Stories triggered by fundamentally different event types

Stories:
%s`

// validTriggerTypes enumerates the recognized trigger classification types.
var validTriggerTypes = map[string]struct{}{
	"user_initiated":   {},
	"system_initiated": {},
	"time_based":       {},
	"event_driven":     {},
}

// LLMBoundaryDetector detects boundary signals using LLM semantic analysis.
// Infrastructure-internal — NOT an application port. Composed by HybridBoundaryDetector (P3-3).
type LLMBoundaryDetector struct {
	client llm.Client
}

// NewLLMBoundaryDetector creates an LLMBoundaryDetector with the given LLM client.
func NewLLMBoundaryDetector(client llm.Client) *LLMBoundaryDetector {
	return &LLMBoundaryDetector{client: client}
}

// boundaryDetectionResponse is the expected JSON structure from the LLM.
type boundaryDetectionResponse struct {
	LanguageDifferences    []languageDifference    `json:"language_differences"`
	TriggerClassifications []triggerClassification `json:"trigger_classifications"`
}

type languageDifference struct {
	Term        string `json:"term"`
	StoryA      string `json:"story_a"`
	StoryB      string `json:"story_b"`
	Description string `json:"description"`
}

type triggerClassification struct {
	StoryTitle  string `json:"story_title"`
	TriggerType string `json:"trigger_type"`
	TriggerText string `json:"trigger_text"`
}

// DetectBoundarySignals sends story summaries to the LLM and returns
// language_difference and different_trigger BoundarySignal values.
// Returns (nil, nil) when LLM is unavailable (ADR-013 graceful degradation).
// Returns (nil, err) for non-availability errors.
func (d *LLMBoundaryDetector) DetectBoundarySignals(
	ctx context.Context,
	stories []*domain.DomainStory,
) ([]domain.BoundarySignal, error) {
	if len(stories) == 0 {
		return nil, nil
	}

	prompt := buildBoundaryDetectionPrompt(stories)
	schema := boundaryDetectionSchema()

	resp, err := d.client.StructuredOutput(ctx, prompt, schema)
	if err != nil {
		if errors.Is(err, llm.ErrLLMUnavailable) {
			return nil, nil
		}

		return nil, fmt.Errorf("detecting boundary signals: %w", err)
	}

	var parsed boundaryDetectionResponse
	if jsonErr := json.Unmarshal([]byte(resp.Content()), &parsed); jsonErr != nil {
		return nil, fmt.Errorf("parsing boundary detection response: %w", jsonErr)
	}

	signals := produceSignals(parsed, len(stories))

	signals = deduplicateSignals(signals)

	if len(signals) == 0 {
		return nil, nil
	}

	return signals, nil
}

func buildBoundaryDetectionPrompt(stories []*domain.DomainStory) string {
	var storiesText strings.Builder

	for i, story := range stories {
		if i > 0 {
			storiesText.WriteString("\n---\n")
		}

		storiesText.WriteString(story.FormatText())
	}

	userPrompt := fmt.Sprintf(boundaryDetectionUserPrompt, storiesText.String())

	return boundaryDetectionSystemPrompt + "\n\n" + userPrompt
}

func boundaryDetectionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"language_differences": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"term":        map[string]any{"type": "string"},
						"story_a":     map[string]any{"type": "string"},
						"story_b":     map[string]any{"type": "string"},
						"description": map[string]any{"type": "string"},
					},
				},
			},
			"trigger_classifications": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"story_title":  map[string]any{"type": "string"},
						"trigger_type": map[string]any{"type": "string"},
						"trigger_text": map[string]any{"type": "string"},
					},
				},
			},
		},
	}
}

func produceSignals(parsed boundaryDetectionResponse, storyCount int) []domain.BoundarySignal {
	var signals []domain.BoundarySignal

	// Language difference signals: one per unique entry.
	for _, ld := range parsed.LanguageDifferences {
		desc := fmt.Sprintf("Term '%s' differs between story '%s' and story '%s': %s",
			ld.Term, ld.StoryA, ld.StoryB, ld.Description)

		signal, err := domain.NewBoundarySignal(domain.SignalTypeLanguageDifference, desc)
		if err != nil {
			continue
		}

		signals = append(signals, signal)
	}

	// Trigger difference signals: pairs of stories with different trigger types.
	// Skip if fewer than 2 stories — nothing to compare.
	if storyCount >= 2 {
		signals = append(signals, produceTriggerSignals(parsed.TriggerClassifications)...)
	}

	return signals
}

func produceTriggerSignals(classifications []triggerClassification) []domain.BoundarySignal {
	// Filter to known trigger types only.
	var valid []triggerClassification

	for _, tc := range classifications {
		if _, ok := validTriggerTypes[tc.TriggerType]; ok {
			valid = append(valid, tc)
		}
	}

	var signals []domain.BoundarySignal

	// Generate a signal for each pair with different trigger types.
	for i := 0; i < len(valid); i++ {
		for j := i + 1; j < len(valid); j++ {
			if valid[i].TriggerType == valid[j].TriggerType {
				continue
			}

			desc := fmt.Sprintf("Story '%s' (%s) vs story '%s' (%s): different trigger types",
				valid[i].StoryTitle, valid[i].TriggerType,
				valid[j].StoryTitle, valid[j].TriggerType)

			signal, err := domain.NewBoundarySignal(domain.SignalTypeDifferentTrigger, desc)
			if err != nil {
				continue
			}

			signals = append(signals, signal)
		}
	}

	return signals
}

func deduplicateSignals(signals []domain.BoundarySignal) []domain.BoundarySignal {
	type key struct {
		signalType domain.SignalType
		desc       string
	}

	seen := make(map[key]struct{}, len(signals))

	var unique []domain.BoundarySignal

	for _, s := range signals {
		k := key{signalType: s.Type(), desc: s.Description()}
		if _, exists := seen[k]; exists {
			continue
		}

		seen[k] = struct{}{}

		unique = append(unique, s)
	}

	return unique
}
