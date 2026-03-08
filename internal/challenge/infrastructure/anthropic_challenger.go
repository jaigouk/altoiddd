package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	challengeapp "github.com/alty-cli/alty/internal/challenge/application"
	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	"github.com/alty-cli/alty/internal/shared/infrastructure/llm"
)

// challengeSchema is the JSON Schema for structured LLM output.
var challengeSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"challenges": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"challenge_type":   map[string]any{"type": "string"},
					"question_text":    map[string]any{"type": "string"},
					"context_name":     map[string]any{"type": "string"},
					"source_reference": map[string]any{"type": "string"},
					"evidence":         map[string]any{"type": "string"},
				},
				"required": []string{"challenge_type", "question_text", "context_name", "source_reference"},
			},
		},
	},
	"required": []string{"challenges"},
}

// AnthropicChallengerAdapter implements Challenger using an LLM.
// Falls back to rule-based generation on any failure.
type AnthropicChallengerAdapter struct {
	llm llm.Client
}

// Compile-time interface check.
var _ challengeapp.Challenger = (*AnthropicChallengerAdapter)(nil)

// NewAnthropicChallengerAdapter creates an AnthropicChallengerAdapter.
func NewAnthropicChallengerAdapter(llmClient llm.Client) *AnthropicChallengerAdapter {
	return &AnthropicChallengerAdapter{llm: llmClient}
}

// GenerateChallenges generates challenges via LLM, falling back to rule-based on failure.
func (a *AnthropicChallengerAdapter) GenerateChallenges(
	ctx context.Context,
	model *ddd.DomainModel,
	maxPerType int,
) ([]challengedomain.Challenge, error) {
	challenges, err := a.llmGenerate(ctx, model, maxPerType)
	if err != nil {
		slog.Info("LLM challenge generation failed, falling back to rule-based", "error", err)
		return challengedomain.Generate(model, maxPerType), nil
	}
	return challenges, nil
}

type llmChallengesResponse struct {
	Challenges []struct {
		ChallengeType   string `json:"challenge_type"`
		QuestionText    string `json:"question_text"`
		ContextName     string `json:"context_name"`
		SourceReference string `json:"source_reference"`
		Evidence        string `json:"evidence"`
	} `json:"challenges"`
}

func (a *AnthropicChallengerAdapter) llmGenerate(
	ctx context.Context,
	model *ddd.DomainModel,
	maxPerType int,
) ([]challengedomain.Challenge, error) {
	prompt := buildPrompt(model, maxPerType)
	response, err := a.llm.StructuredOutput(ctx, prompt, challengeSchema)
	if err != nil {
		return nil, fmt.Errorf("LLM structured output: %w", err)
	}

	var data llmChallengesResponse
	if err := json.Unmarshal([]byte(response.Content()), &data); err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", err)
	}

	challenges := make([]challengedomain.Challenge, 0, len(data.Challenges))
	for _, item := range data.Challenges {
		c, err := challengedomain.NewChallenge(
			challengedomain.ChallengeType(item.ChallengeType),
			item.QuestionText,
			item.ContextName,
			item.SourceReference,
			item.Evidence,
		)
		if err != nil {
			return nil, fmt.Errorf("creating challenge %q: %w", item.ChallengeType, err)
		}
		challenges = append(challenges, c)
	}
	return challenges, nil
}

func buildPrompt(model *ddd.DomainModel, maxPerType int) string {
	var parts []string
	parts = append(parts, "Analyze this domain model and generate challenges:\n")

	parts = append(parts, "## Bounded Contexts")
	for _, ctx := range model.BoundedContexts() {
		classification := "unclassified"
		if cl := ctx.Classification(); cl != nil {
			classification = string(*cl)
		}
		parts = append(parts, fmt.Sprintf("- %s (%s): %s", ctx.Name(), classification, ctx.Responsibility()))
	}

	parts = append(parts, "\n## Aggregates")
	for _, agg := range model.AggregateDesigns() {
		invCount := len(agg.Invariants())
		parts = append(parts, fmt.Sprintf("- %s in %s: root=%s, invariants=%d",
			agg.Name(), agg.ContextName(), agg.RootEntity(), invCount))
	}

	parts = append(parts, "\n## Domain Stories")
	for _, story := range model.DomainStories() {
		parts = append(parts, fmt.Sprintf("- %s: %s", story.Name(), strings.Join(story.Steps(), " -> ")))
	}

	parts = append(parts, "\n## Ubiquitous Language")
	for _, entry := range model.UbiquitousLanguage().Terms() {
		parts = append(parts, fmt.Sprintf("- %s (%s): %s", entry.Term(), entry.ContextName(), entry.Definition()))
	}

	types := challengedomain.AllChallengeTypes()
	typeNames := make([]string, len(types))
	for i, ct := range types {
		typeNames[i] = string(ct)
	}
	parts = append(parts, fmt.Sprintf(
		"\nGenerate up to %d challenges per type. Types: %s. "+
			"Each challenge must be a QUESTION (never state facts). "+
			"Cite source_reference for every challenge.",
		maxPerType, strings.Join(typeNames, ", ")))

	return strings.Join(parts, "\n")
}
