package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// classificationKeywords maps keywords to subdomain classifications.
var classificationKeywords = map[string]vo.SubdomainClassification{
	"core":           vo.SubdomainCore,
	"secret sauce":   vo.SubdomainCore,
	"competitive":    vo.SubdomainCore,
	"supporting":     vo.SubdomainSupporting,
	"plumbing":       vo.SubdomainSupporting,
	"necessary":      vo.SubdomainSupporting,
	"generic":        vo.SubdomainGeneric,
	"off-the-shelf":  vo.SubdomainGeneric,
	"commodity":      vo.SubdomainGeneric,
	"buy":            vo.SubdomainGeneric,
}

// ArtifactPreview holds rendered artifact content ready for user review.
type ArtifactPreview struct {
	Model               *ddd.DomainModel
	PRDContent          string
	DDDContent          string
	ArchitectureContent string
}

// ArtifactGenerationHandler transforms DiscoveryCompleted into DDD artifacts.
type ArtifactGenerationHandler struct {
	renderer ArtifactRenderer
	writer   sharedapp.FileWriter
}

// NewArtifactGenerationHandler creates a new ArtifactGenerationHandler.
func NewArtifactGenerationHandler(
	renderer ArtifactRenderer,
	writer sharedapp.FileWriter,
) *ArtifactGenerationHandler {
	return &ArtifactGenerationHandler{
		renderer: renderer,
		writer:   writer,
	}
}

// BuildPreview builds a domain model from discovery answers and renders artifacts.
func (h *ArtifactGenerationHandler) BuildPreview(
	ctx context.Context,
	event discoverydomain.DiscoveryCompletedEvent,
) (*ArtifactPreview, error) {
	answers := event.Answers()
	if len(answers) == 0 {
		return nil, fmt.Errorf("No substantive answers to generate artifacts from: %w",
			domainerrors.ErrInvariantViolation)
	}

	model, err := buildModel(event)
	if err != nil {
		return nil, err
	}

	prd, err := h.renderer.RenderPRD(ctx, model)
	if err != nil {
		return nil, err
	}
	dddContent, err := h.renderer.RenderDDD(ctx, model)
	if err != nil {
		return nil, err
	}
	arch, err := h.renderer.RenderArchitecture(ctx, model)
	if err != nil {
		return nil, err
	}

	return &ArtifactPreview{
		Model:               model,
		PRDContent:          prd,
		DDDContent:          dddContent,
		ArchitectureContent: arch,
	}, nil
}

// WriteArtifacts writes previously previewed artifacts to disk.
func (h *ArtifactGenerationHandler) WriteArtifacts(
	ctx context.Context,
	preview *ArtifactPreview,
	outputDir string,
) error {
	if err := h.writer.WriteFile(ctx, filepath.Join(outputDir, "PRD.md"), preview.PRDContent); err != nil {
		return err
	}
	if err := h.writer.WriteFile(ctx, filepath.Join(outputDir, "DDD.md"), preview.DDDContent); err != nil {
		return err
	}
	return h.writer.WriteFile(ctx, filepath.Join(outputDir, "ARCHITECTURE.md"), preview.ArchitectureContent)
}

// Generate is a convenience method that builds preview and writes in one step.
func (h *ArtifactGenerationHandler) Generate(
	ctx context.Context,
	event discoverydomain.DiscoveryCompletedEvent,
	outputDir string,
) (*ddd.DomainModel, error) {
	preview, err := h.BuildPreview(ctx, event)
	if err != nil {
		return nil, err
	}
	if err := h.WriteArtifacts(ctx, preview, outputDir); err != nil {
		return nil, err
	}
	return preview.Model, nil
}

// SplitAnswer splits a free-text answer into meaningful parts.
// Exported for testing.
func SplitAnswer(answer string) []string {
	trimmed := strings.TrimSpace(answer)
	if trimmed == "" {
		return nil
	}

	// Try comma separation first.
	parts := strings.Split(trimmed, ",")
	if len(parts) > 1 {
		var result []string
		for _, p := range parts {
			s := strings.TrimSpace(p)
			if s != "" {
				result = append(result, s)
			}
		}
		return result
	}

	// Try newline separation.
	parts = strings.Split(trimmed, "\n")
	if len(parts) > 1 {
		var result []string
		for _, p := range parts {
			s := strings.TrimSpace(p)
			s = strings.TrimLeft(s, "0123456789.-) ")
			s = strings.TrimSpace(s)
			if s != "" {
				result = append(result, s)
			}
		}
		return result
	}

	// Single item.
	return []string{trimmed}
}

// -- Private model building ---------------------------------------------------

func buildModel(event discoverydomain.DiscoveryCompletedEvent) (*ddd.DomainModel, error) {
	model := ddd.NewDomainModel("discovery-" + event.SessionID())
	answersMap := make(map[string]string)
	for _, a := range event.Answers() {
		answersMap[a.QuestionID()] = a.ResponseText()
	}

	// Extract contexts FIRST so terms get correct context names.
	extractContexts(model, answersMap)

	// Stories from Q1, Q3, Q5.
	extractStories(model, answersMap)

	// Terms from Q2.
	extractTerms(model, answersMap)

	// Classifications from Q10.
	extractClassifications(model, answersMap)

	// Aggregates from Q4, Q6 for Core subdomains.
	extractAggregates(model, answersMap)

	if err := model.Finalize(); err != nil {
		return nil, err
	}
	return model, nil
}

func extractContexts(model *ddd.DomainModel, answers map[string]string) {
	contexts := SplitAnswer(answers["Q9"])
	for _, ctxName := range contexts {
		ctxName = strings.TrimSpace(ctxName)
		if ctxName != "" {
			bc := vo.NewDomainBoundedContext(ctxName, "Manages "+ctxName+" domain", nil, nil, "")
			model.AddBoundedContext(bc)
		}
	}
}

func extractStories(model *ddd.DomainModel, answers map[string]string) {
	actors := SplitAnswer(answers["Q1"])
	primarySteps := SplitAnswer(answers["Q3"])

	if len(primarySteps) > 0 {
		actorList := actors
		if len(actorList) == 0 {
			actorList = []string{"User"}
		}
		trigger := primarySteps[0]
		story := vo.NewDomainStory("Primary Flow", actorList, trigger, primarySteps, nil)
		model.AddDomainStory(story)
	}

	secondary := SplitAnswer(answers["Q5"])
	if len(secondary) > 0 {
		actorList := actors
		if len(actorList) == 0 {
			actorList = []string{"User"}
		}
		story := vo.NewDomainStory("Secondary Flows", actorList, "Various", secondary, nil)
		model.AddDomainStory(story)
	}
}

func extractTerms(model *ddd.DomainModel, answers map[string]string) {
	entities := SplitAnswer(answers["Q2"])
	bcs := model.BoundedContexts()
	contextName := "General"
	if len(bcs) > 0 {
		contextName = bcs[0].Name()
	}

	for _, entity := range entities {
		entity = strings.TrimSpace(entity)
		if entity != "" {
			model.AddTerm(entity, entity+" entity", contextName, []string{"Q2"})
		}
	}
}

func extractClassifications(model *ddd.DomainModel, answers map[string]string) {
	q10 := strings.ToLower(answers["Q10"])
	if q10 == "" {
		return
	}

	for _, ctx := range model.BoundedContexts() {
		if ctx.Classification() != nil {
			continue
		}
		ctxLower := strings.ToLower(ctx.Name())
		for keyword, cls := range classificationKeywords {
			if strings.Contains(q10, keyword) && strings.Contains(q10, ctxLower) {
				model.ClassifySubdomain(ctx.Name(), cls, fmt.Sprintf("Classified as %s based on Q10 answer", cls))
				break
			}
		}
	}
}

func extractAggregates(model *ddd.DomainModel, answers map[string]string) {
	invariants := SplitAnswer(answers["Q4"])
	domainEvents := SplitAnswer(answers["Q6"])

	for _, ctx := range model.BoundedContexts() {
		if ctx.Classification() == nil || *ctx.Classification() != vo.SubdomainCore {
			continue
		}
		agg := vo.NewAggregateDesign(
			ctx.Name()+"Root",
			ctx.Name(),
			ctx.Name()+"Root",
			nil,
			invariants,
			nil,
			domainEvents,
		)
		model.DesignAggregate(agg)
	}
}
