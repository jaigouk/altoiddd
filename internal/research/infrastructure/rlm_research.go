package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	researchapp "github.com/alty-cli/alty/internal/research/application"
	researchdomain "github.com/alty-cli/alty/internal/research/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	"github.com/alty-cli/alty/internal/shared/infrastructure/llm"
)

// RlmResearchAdapter implements DomainResearch using the Recursive Language Model pattern:
// search -> read -> reason -> search again -> build findings.
type RlmResearchAdapter struct {
	llm llm.Client
	web researchapp.WebSearch
}

// Compile-time interface check.
var _ researchapp.DomainResearch = (*RlmResearchAdapter)(nil)

// NewRlmResearchAdapter creates an RlmResearchAdapter.
func NewRlmResearchAdapter(llmClient llm.Client, webSearch researchapp.WebSearch) *RlmResearchAdapter {
	return &RlmResearchAdapter{llm: llmClient, web: webSearch}
}

// Research researches domain areas using iterative web search + LLM synthesis.
func (r *RlmResearchAdapter) Research(
	ctx context.Context,
	model *ddd.DomainModel,
	maxAreas int,
) (researchdomain.ResearchBriefing, error) {
	areas := r.extractAreas(model, maxAreas)
	var findings []researchdomain.ResearchFinding
	var noData []string

	for _, area := range areas {
		areaFindings := r.researchArea(ctx, area)
		if len(areaFindings) > 0 {
			findings = append(findings, areaFindings...)
		} else {
			noData = append(noData, area)
		}
	}

	summary := ""
	if len(findings) > 0 {
		summary = r.buildSummary(ctx, findings)
	}
	return researchdomain.NewResearchBriefing(findings, noData, summary), nil
}

func (r *RlmResearchAdapter) extractAreas(model *ddd.DomainModel, maxAreas int) []string {
	contexts := model.BoundedContexts()
	limit := len(contexts)
	if maxAreas < limit {
		limit = maxAreas
	}
	areas := make([]string, limit)
	for i := 0; i < limit; i++ {
		areas[i] = contexts[i].Name()
	}
	return areas
}

func (r *RlmResearchAdapter) researchArea(ctx context.Context, area string) []researchdomain.ResearchFinding {
	// Round 1: broad search
	results, err := r.web.Search(ctx, area+" domain patterns best practices", 10)
	if err != nil || len(results) == 0 {
		return nil
	}

	// Try LLM synthesis + refined search
	synthesis, err := r.llm.TextCompletion(ctx, r.synthesisPrompt(area, results))
	if err != nil {
		slog.Info("LLM unavailable for area, using raw results", "area", area, "error", err)
		return r.buildFindings(area, results, researchdomain.ConfidenceLow)
	}

	refinedQuery := r.extractRefinedQuery(synthesis.Content(), area)

	// Round 2: refined search
	results2, err := r.web.Search(ctx, refinedQuery, 10)
	if err != nil {
		results2 = nil
	}
	results = append(results, results2...)
	return r.buildFindings(area, results, researchdomain.ConfidenceMedium)
}

func (r *RlmResearchAdapter) synthesisPrompt(area string, results []researchdomain.WebSearchResult) string {
	limit := 5
	if len(results) < limit {
		limit = len(results)
	}
	var snippets []string
	for _, res := range results[:limit] {
		snippets = append(snippets, fmt.Sprintf("- [%s](%s): %s", res.Title(), res.URL(), res.Snippet()))
	}
	return fmt.Sprintf(
		"Analyze these search results about '%s' domain patterns:\n\n%s\n\n"+
			"Provide:\n"+
			"1. A refined search query to find deeper information (prefix with 'Refined query: ')\n"+
			"2. A brief synthesis of key patterns found (prefix with 'Synthesis: ')",
		area, strings.Join(snippets, "\n"))
}

func (r *RlmResearchAdapter) extractRefinedQuery(content, area string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "refined query:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return area + " industry patterns competitive analysis"
}

func (r *RlmResearchAdapter) buildFindings(
	area string,
	results []researchdomain.WebSearchResult,
	confidence researchdomain.Confidence,
) []researchdomain.ResearchFinding {
	today := time.Now().Format("2006-01-02")
	trustLevel := researchdomain.TrustAIResearched
	if confidence == researchdomain.ConfidenceLow {
		trustLevel = researchdomain.TrustAIInferred
	}

	var findings []researchdomain.ResearchFinding
	for _, res := range results {
		if res.URL() == "" || res.Snippet() == "" {
			continue
		}
		source, err := researchdomain.NewSourceAttribution(res.URL(), res.Title(), today, confidence)
		if err != nil {
			continue
		}
		finding := researchdomain.NewResearchFinding(res.Snippet(), source, trustLevel, area, false)
		findings = append(findings, finding)
	}
	return findings
}

func (r *RlmResearchAdapter) buildSummary(ctx context.Context, findings []researchdomain.ResearchFinding) string {
	areas := make(map[string]struct{})
	for _, f := range findings {
		areas[f.DomainArea()] = struct{}{}
	}

	limit := 10
	if len(findings) < limit {
		limit = len(findings)
	}
	var lines []string
	for _, f := range findings[:limit] {
		lines = append(lines, fmt.Sprintf("- [%s] %s", f.DomainArea(), f.Content()))
	}

	sortedAreas := make([]string, 0, len(areas))
	for a := range areas {
		sortedAreas = append(sortedAreas, a)
	}

	prompt := fmt.Sprintf(
		"Summarize these research findings across %d domain area(s) (%s):\n\n%s\n\nProvide a 2-3 sentence summary.",
		len(areas), strings.Join(sortedAreas, ", "), strings.Join(lines, "\n"))

	resp, err := r.llm.TextCompletion(ctx, prompt)
	if err != nil {
		return fmt.Sprintf("Research found %d finding(s) across %d area(s).", len(findings), len(areas))
	}
	return resp.Content()
}
