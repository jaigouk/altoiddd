package application

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	"github.com/alto-cli/alto/internal/shared/domain/stringutil"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// classificationKeywords maps keywords to subdomain classifications.
var classificationKeywords = map[string]vo.SubdomainClassification{
	"core":          vo.SubdomainCore,
	"secret sauce":  vo.SubdomainCore,
	"competitive":   vo.SubdomainCore,
	"supporting":    vo.SubdomainSupporting,
	"plumbing":      vo.SubdomainSupporting,
	"necessary":     vo.SubdomainSupporting,
	"generic":       vo.SubdomainGeneric,
	"off-the-shelf": vo.SubdomainGeneric,
	"commodity":     vo.SubdomainGeneric,
	"buy":           vo.SubdomainGeneric,
}

// ArtifactPreview holds rendered artifact content ready for user review.
type ArtifactPreview struct {
	Model                 *ddd.DomainModel
	PRDContent            string
	DDDContent            string
	ArchitectureContent   string
	BoundedContextMapYAML string
}

// ArtifactGenerationHandler transforms DiscoveryCompleted into DDD artifacts.
type ArtifactGenerationHandler struct {
	renderer  ArtifactRenderer
	writer    sharedapp.FileWriter
	publisher sharedapp.EventPublisher
}

// NewArtifactGenerationHandler creates a new ArtifactGenerationHandler.
func NewArtifactGenerationHandler(
	renderer ArtifactRenderer,
	writer sharedapp.FileWriter,
	publisher sharedapp.EventPublisher,
) *ArtifactGenerationHandler {
	return &ArtifactGenerationHandler{
		renderer:  renderer,
		writer:    writer,
		publisher: publisher,
	}
}

// BuildPreview builds a domain model from discovery answers and renders artifacts.
func (h *ArtifactGenerationHandler) BuildPreview(
	ctx context.Context,
	event discoverydomain.DiscoveryCompletedEvent,
) (*ArtifactPreview, error) {
	answers := event.Answers()
	if len(answers) == 0 {
		return nil, fmt.Errorf("no substantive answers to generate artifacts from: %w",
			domainerrors.ErrInvariantViolation)
	}

	model, err := buildModel(event)
	if err != nil {
		return nil, fmt.Errorf("build domain model: %w", err)
	}

	prd, err := h.renderer.RenderPRD(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("render PRD: %w", err)
	}
	dddContent, err := h.renderer.RenderDDD(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("render DDD: %w", err)
	}
	arch, err := h.renderer.RenderArchitecture(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("render architecture: %w", err)
	}

	for _, event := range model.Events() {
		_ = h.publisher.Publish(ctx, event)
	}

	bcMapYAML, err := renderBoundedContextMapYAML(model)
	if err != nil {
		return nil, fmt.Errorf("render bounded context map: %w", err)
	}

	return &ArtifactPreview{
		Model:                 model,
		PRDContent:            prd,
		DDDContent:            dddContent,
		ArchitectureContent:   arch,
		BoundedContextMapYAML: bcMapYAML,
	}, nil
}

// WriteArtifacts writes previously previewed artifacts to disk.
// docsDir is where PRD.md, DDD.md, ARCHITECTURE.md go (typically docs/).
// projectDir is the project root where .alto/bounded_context_map.yaml goes.
func (h *ArtifactGenerationHandler) WriteArtifacts(
	ctx context.Context,
	preview *ArtifactPreview,
	docsDir string,
	projectDir string,
) error {
	if err := h.writer.WriteFile(ctx, filepath.Join(docsDir, "PRD.md"), preview.PRDContent); err != nil {
		return fmt.Errorf("write PRD: %w", err)
	}
	if err := h.writer.WriteFile(ctx, filepath.Join(docsDir, "DDD.md"), preview.DDDContent); err != nil {
		return fmt.Errorf("write DDD: %w", err)
	}
	if err := h.writer.WriteFile(ctx, filepath.Join(docsDir, "ARCHITECTURE.md"), preview.ArchitectureContent); err != nil {
		return fmt.Errorf("write architecture: %w", err)
	}
	if preview.BoundedContextMapYAML != "" {
		bcMapPath := filepath.Join(projectDir, ".alto", "bounded_context_map.yaml")
		if err := h.writer.WriteFile(ctx, bcMapPath, preview.BoundedContextMapYAML); err != nil {
			return fmt.Errorf("write bounded context map: %w", err)
		}
	}
	return nil
}

// Generate is a convenience method that builds preview and writes in one step.
// docsDir is where PRD.md, DDD.md, ARCHITECTURE.md go.
// projectDir is the project root where .alto/bounded_context_map.yaml goes.
func (h *ArtifactGenerationHandler) Generate(
	ctx context.Context,
	event discoverydomain.DiscoveryCompletedEvent,
	docsDir string,
	projectDir string,
) (*ddd.DomainModel, error) {
	preview, err := h.BuildPreview(ctx, event)
	if err != nil {
		return nil, err
	}
	if err := h.WriteArtifacts(ctx, preview, docsDir, projectDir); err != nil {
		return nil, err
	}
	return preview.Model, nil
}

// GenerateFromStories builds a domain model from story files and renders artifacts.
// Unlike Generate, this does not require Q&A answers — it reads stories directly.
func (h *ArtifactGenerationHandler) GenerateFromStories(
	ctx context.Context,
	storyReader StoryReader,
	storyPaths []string,
	contextMap *discoverydomain.ContextMap,
	docsDir string,
	projectDir string,
) error {
	if len(storyPaths) == 0 {
		return fmt.Errorf("no story paths provided (empty): %w", domainerrors.ErrInvariantViolation)
	}

	model, err := h.buildModelFromStories(ctx, storyReader, storyPaths, contextMap, filepath.Base(projectDir))
	if err != nil {
		return fmt.Errorf("build model from stories: %w", err)
	}

	prd, err := h.renderer.RenderPRD(ctx, model)
	if err != nil {
		return fmt.Errorf("render PRD: %w", err)
	}

	dddContent, err := h.renderer.RenderDDD(ctx, model)
	if err != nil {
		return fmt.Errorf("render DDD: %w", err)
	}

	arch, err := h.renderer.RenderArchitecture(ctx, model)
	if err != nil {
		return fmt.Errorf("render architecture: %w", err)
	}

	preview := &ArtifactPreview{
		Model:               model,
		PRDContent:          prd,
		DDDContent:          dddContent,
		ArchitectureContent: arch,
	}

	return h.WriteArtifacts(ctx, preview, docsDir, projectDir)
}

// buildModelFromStories reads stories via storyReader and extracts actors, work objects,
// and sentences into DomainModel terms and bounded contexts.
func (h *ArtifactGenerationHandler) buildModelFromStories(
	ctx context.Context,
	storyReader StoryReader,
	storyPaths []string,
	contextMap *discoverydomain.ContextMap,
	projectName string,
) (*ddd.DomainModel, error) {
	model := ddd.NewDomainModel(projectName)

	// Add bounded contexts from context map if available.
	if contextMap != nil {
		for _, sketch := range contextMap.Contexts() {
			cls := sketch.Classification()
			bc := vo.NewDomainBoundedContext(sketch.Name(), "Manages "+sketch.Name()+" domain", nil, &cls, "")
			if err := model.AddBoundedContext(bc); err != nil {
				return nil, fmt.Errorf("add bounded context %q: %w", sketch.Name(), err)
			}
		}
	}

	// Determine the target context name for terms.
	contextName := "General"
	bcs := model.BoundedContexts()
	if len(bcs) > 0 {
		contextName = bcs[0].Name()
	} else {
		// Single-context: add a default General bounded context.
		generic := vo.SubdomainGeneric
		bc := vo.NewDomainBoundedContext("General", "Default bounded context", nil, &generic, "auto-generated")
		if err := model.AddBoundedContext(bc); err != nil {
			return nil, fmt.Errorf("add default bounded context: %w", err)
		}
	}

	seenTerms := make(map[string]struct{})

	for _, path := range storyPaths {
		story, err := storyReader.Read(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("reading story %q: %w", path, err)
		}

		// Add story to model as a value object.
		storyVO := vo.NewDomainStory(
			story.Title(),
			actorNames(story),
			story.Trigger(),
			sentenceTexts(story),
			nil,
		)
		if err := model.AddDomainStory(storyVO); err != nil {
			if errors.Is(err, domainerrors.ErrAlreadyExists) {
				continue // Skip duplicate stories.
			}
			return nil, fmt.Errorf("add domain story %q: %w", story.Title(), err)
		}

		// Extract actors as terms.
		for _, actor := range story.Actors() {
			termKey := strings.ToLower(actor.Name())
			if _, seen := seenTerms[termKey]; seen {
				continue
			}
			seenTerms[termKey] = struct{}{}

			if err := model.AddTerm(actor.Name(), actor.Name()+" actor", contextName, []string{path}); err != nil {
				return nil, fmt.Errorf("add actor term %q: %w", actor.Name(), err)
			}
		}
	}

	if err := model.Finalize(); err != nil {
		return nil, fmt.Errorf("finalize domain model: %w", err)
	}

	return model, nil
}

func actorNames(story *discoverydomain.DomainStory) []string {
	actors := story.Actors()
	names := make([]string, len(actors))
	for i, a := range actors {
		names[i] = a.Name()
	}
	return names
}

func sentenceTexts(story *discoverydomain.DomainStory) []string {
	sentences := story.Sentences()
	texts := make([]string, len(sentences))
	for i, s := range sentences {
		texts[i] = fmt.Sprintf("%s %s %s", s.Subject(), s.Activity(), s.Object())
	}
	return texts
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
	if err := extractContexts(model, answersMap); err != nil {
		return nil, fmt.Errorf("extract contexts: %w", err)
	}

	// Stories from Q1, Q3, Q5.
	if err := extractStories(model, answersMap); err != nil {
		return nil, fmt.Errorf("extract stories: %w", err)
	}

	// Terms from Q2.
	if err := extractTerms(model, answersMap); err != nil {
		return nil, fmt.Errorf("extract terms: %w", err)
	}

	// Classifications from Q10.
	if err := extractClassifications(model, answersMap); err != nil {
		return nil, fmt.Errorf("extract classifications: %w", err)
	}

	// Aggregates from Q4, Q6 for Core subdomains.
	if err := extractAggregates(model, answersMap); err != nil {
		return nil, fmt.Errorf("extract aggregates: %w", err)
	}

	if err := model.Finalize(); err != nil {
		return nil, fmt.Errorf("finalize domain model: %w", err)
	}
	return model, nil
}

func extractContexts(model *ddd.DomainModel, answers map[string]string) error {
	contexts := SplitAnswer(answers["Q9"])
	for _, ctxName := range contexts {
		ctxName = strings.TrimSpace(ctxName)
		if ctxName != "" {
			bc := vo.NewDomainBoundedContext(ctxName, "Manages "+ctxName+" domain", nil, nil, "")
			if err := model.AddBoundedContext(bc); err != nil {
				return fmt.Errorf("add bounded context %q: %w", ctxName, err)
			}
		}
	}
	return nil
}

func extractStories(model *ddd.DomainModel, answers map[string]string) error {
	actors := SplitAnswer(answers["Q1"])
	primarySteps := SplitAnswer(answers["Q3"])

	if len(primarySteps) > 0 {
		actorList := actors
		if len(actorList) == 0 {
			actorList = []string{"User"}
		}
		trigger := primarySteps[0]
		story := vo.NewDomainStory("Primary Flow", actorList, trigger, primarySteps, nil)
		if err := model.AddDomainStory(story); err != nil {
			return fmt.Errorf("add primary flow story: %w", err)
		}
	}

	secondary := SplitAnswer(answers["Q5"])
	if len(secondary) > 0 {
		actorList := actors
		if len(actorList) == 0 {
			actorList = []string{"User"}
		}
		story := vo.NewDomainStory("Secondary Flows", actorList, "Various", secondary, nil)
		if err := model.AddDomainStory(story); err != nil {
			return fmt.Errorf("add secondary flows story: %w", err)
		}
	}
	return nil
}

func extractTerms(model *ddd.DomainModel, answers map[string]string) error {
	entities := SplitAnswer(answers["Q2"])
	bcs := model.BoundedContexts()
	contextName := "General"
	if len(bcs) > 0 {
		contextName = bcs[0].Name()
	}

	for _, entity := range entities {
		entity = strings.TrimSpace(entity)
		if entity != "" {
			if err := model.AddTerm(entity, entity+" entity", contextName, []string{"Q2"}); err != nil {
				return fmt.Errorf("add term %q: %w", entity, err)
			}
		}
	}
	return nil
}

func extractClassifications(model *ddd.DomainModel, answers map[string]string) error {
	q10 := strings.ToLower(answers["Q10"])
	if q10 == "" {
		return nil
	}

	segments := splitIntoSegments(q10)

	for _, ctx := range model.BoundedContexts() {
		if ctx.Classification() != nil {
			continue
		}
		ctxLower := strings.ToLower(ctx.Name())
		for _, seg := range segments {
			if !strings.Contains(seg, ctxLower) {
				continue
			}
			for keyword, cls := range classificationKeywords {
				if strings.Contains(seg, keyword) {
					if err := model.ClassifySubdomain(ctx.Name(), cls, fmt.Sprintf("Classified as %s based on Q10 answer", cls)); err != nil {
						return fmt.Errorf("classify subdomain %q: %w", ctx.Name(), err)
					}
					break
				}
			}
			break
		}
	}
	return nil
}

// splitIntoSegments splits a Q10 answer into segments by comma, period, or semicolon.
func splitIntoSegments(s string) []string {
	// Replace periods and semicolons with commas for uniform splitting.
	s = strings.ReplaceAll(s, ".", ",")
	s = strings.ReplaceAll(s, ";", ",")
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func extractAggregates(model *ddd.DomainModel, answers map[string]string) error {
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
		if err := model.DesignAggregate(agg); err != nil {
			return fmt.Errorf("design aggregate for %q: %w", ctx.Name(), err)
		}
	}
	return nil
}

// -- Bounded Context Map YAML Generation --------------------------------------

// boundedContextMapYAML is the YAML structure for bounded_context_map.yaml.
type boundedContextMapYAML struct {
	Project         projectYAML          `yaml:"project"`
	BoundedContexts []boundedContextYAML `yaml:"bounded_contexts"`
}

type projectYAML struct {
	Name        string `yaml:"name"`
	RootPackage string `yaml:"root_package"`
}

type boundedContextYAML struct {
	Name           string   `yaml:"name"`
	ModulePath     string   `yaml:"module_path"`
	Classification string   `yaml:"classification"`
	Rationale      string   `yaml:"rationale,omitempty"`
	Layers         []string `yaml:"layers"`
}

// renderBoundedContextMapYAML generates YAML from a DomainModel.
// Note: The generated project.root_package is a placeholder that should be updated
// by the CLI based on actual go.mod detection when available.
func renderBoundedContextMapYAML(model *ddd.DomainModel) (string, error) {
	bcMap := boundedContextMapYAML{
		Project: projectYAML{
			Name:        model.ModelID(),
			RootPackage: "github.com/project/" + stringutil.ToSnakeCase(model.ModelID()),
		},
		BoundedContexts: make([]boundedContextYAML, 0, len(model.BoundedContexts())),
	}

	for _, bc := range model.BoundedContexts() {
		classification := ""
		if bc.Classification() != nil {
			classification = string(*bc.Classification())
		}

		bcMap.BoundedContexts = append(bcMap.BoundedContexts, boundedContextYAML{
			Name:           bc.Name(),
			ModulePath:     stringutil.ToSnakeCase(bc.Name()),
			Classification: classification,
			Rationale:      bc.ClassificationRationale(),
			Layers:         []string{"domain", "application", "infrastructure"},
		})
	}

	out, err := yaml.Marshal(&bcMap)
	if err != nil {
		return "", fmt.Errorf("marshal bounded context map: %w", err)
	}

	return string(out), nil
}
