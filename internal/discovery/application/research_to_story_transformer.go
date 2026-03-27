package application

import (
	"context"
	"fmt"
	"sort"
	"strings"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

const (
	researchSourceFallback = "domain research"
	researchTrigger        = "AI-proposed from domain research"
)

// workflowTypeOrder defines the processing order for workflow types.
var workflowTypeOrder = []discoverydomain.WorkflowType{
	discoverydomain.WorkflowTypeHappyPath,
	discoverydomain.WorkflowTypeFailureCase,
	discoverydomain.WorkflowTypeSecondary,
}

// ResearchToStoryTransformer transforms a DomainResearchResult into DomainStories.
// Stateless — safe for concurrent use.
type ResearchToStoryTransformer struct{}

// NewResearchToStoryTransformer creates a new ResearchToStoryTransformer.
func NewResearchToStoryTransformer() *ResearchToStoryTransformer {
	return &ResearchToStoryTransformer{}
}

// Transform maps a DomainResearchResult to DomainStories. Returns (nil, nil) if
// the research quality does not meet the floor. Context is accepted for future use.
func (t *ResearchToStoryTransformer) Transform(
	_ context.Context,
	result *discoverydomain.DomainResearchResult,
) ([]*discoverydomain.DomainStory, error) {
	if !result.Quality().MeetsFloor() {
		return nil, nil
	}

	selected := selectWorkflows(result.Workflows())
	if len(selected) == 0 {
		return nil, nil
	}

	stories := make([]*discoverydomain.DomainStory, 0, len(selected))

	for _, wf := range selected {
		story, err := buildStory(wf, result)
		if err != nil {
			return nil, fmt.Errorf("building story %q: %w", wf.Name(), err)
		}

		stories = append(stories, story)
	}

	return stories, nil
}

// selectWorkflows picks at most one workflow per type in the defined order.
// First workflow of each type wins; duplicates are ignored. Max 3 returned.
func selectWorkflows(workflows []discoverydomain.ResearchedWorkflow) []discoverydomain.ResearchedWorkflow {
	byType := make(map[discoverydomain.WorkflowType]discoverydomain.ResearchedWorkflow, len(workflowTypeOrder))

	for _, wf := range workflows {
		if _, exists := byType[wf.WorkflowType()]; !exists {
			byType[wf.WorkflowType()] = wf
		}
	}

	result := make([]discoverydomain.ResearchedWorkflow, 0, len(workflowTypeOrder))

	for _, wfType := range workflowTypeOrder {
		if wf, ok := byType[wfType]; ok {
			result = append(result, wf)
		}
	}

	return result
}

// buildStory creates a single DomainStory from a workflow and the research result.
func buildStory(
	wf discoverydomain.ResearchedWorkflow,
	result *discoverydomain.DomainResearchResult,
) (*discoverydomain.DomainStory, error) {
	story, err := discoverydomain.NewDomainStory(
		wf.Name(),
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypeDigitalized,
		researchTrigger,
	)
	if err != nil {
		return nil, fmt.Errorf("creating domain story: %w", err)
	}

	steps := sortStepsBySequence(wf.Steps())
	wfSource := sourceFallback(wf.SourceURLs())

	actors, err := resolveActors(steps, result.Actors(), wfSource)
	if err != nil {
		return nil, fmt.Errorf("resolving actors: %w", err)
	}

	for _, actor := range actors {
		if addErr := story.AddActor(actor); addErr != nil {
			return nil, fmt.Errorf("adding actor %q: %w", actor.Name(), addErr)
		}
	}

	workObjects, err := resolveWorkObjects(steps, result.Entities(), wfSource)
	if err != nil {
		return nil, fmt.Errorf("resolving work objects: %w", err)
	}

	for _, wo := range workObjects {
		if addErr := story.AddWorkObject(wo); addErr != nil {
			return nil, fmt.Errorf("adding work object %q: %w", wo.Name(), addErr)
		}
	}

	for i, step := range steps {
		sentence, sentenceErr := discoverydomain.NewStorySentence(
			i+1,
			step.Actor(),
			step.Activity(),
			step.WorkObject(),
			vo.AIResearched,
			wfSource,
		)
		if sentenceErr != nil {
			return nil, fmt.Errorf("creating sentence for step %d: %w", i+1, sentenceErr)
		}

		if addErr := story.AddSentence(sentence); addErr != nil {
			return nil, fmt.Errorf("adding sentence for step %d: %w", i+1, addErr)
		}
	}

	return story, nil
}

// sortStepsBySequence returns steps sorted by Sequence() ascending.
func sortStepsBySequence(steps []discoverydomain.WorkflowStep) []discoverydomain.WorkflowStep {
	sorted := make([]discoverydomain.WorkflowStep, len(steps))
	copy(sorted, steps)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Sequence() < sorted[j].Sequence()
	})

	return sorted
}

// resolveActors collects unique actors (case-insensitive) from workflow steps,
// resolving source URLs from research actors when available.
func resolveActors(
	steps []discoverydomain.WorkflowStep,
	researchActors []discoverydomain.ResearchedActor,
	fallbackSource string,
) ([]discoverydomain.StoryActor, error) {
	seen := make(map[string]struct{})
	var actors []discoverydomain.StoryActor

	for _, step := range steps {
		key := strings.ToLower(step.Actor())
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}

		source := resolveActorSource(step.Actor(), researchActors, fallbackSource)

		actor, err := discoverydomain.NewStoryActor(
			step.Actor(),
			discoverydomain.ActorTypePerson,
			vo.AIResearched,
			source,
		)
		if err != nil {
			return nil, fmt.Errorf("creating story actor %q: %w", step.Actor(), err)
		}

		actors = append(actors, actor)
	}

	return actors, nil
}

// resolveActorSource finds the source URL for an actor name in research actors.
// Case-insensitive match. Falls back to fallbackSource.
func resolveActorSource(actorName string, researchActors []discoverydomain.ResearchedActor, fallbackSource string) string {
	for _, ra := range researchActors {
		if strings.EqualFold(ra.Name(), actorName) {
			source := sourceFallback(ra.SourceURLs())
			if source != researchSourceFallback {
				return source
			}

			return fallbackSource
		}
	}

	return researchSourceFallback
}

// resolveWorkObjects collects unique work objects (case-insensitive) from workflow steps,
// resolving source URLs from research entities when available.
func resolveWorkObjects(
	steps []discoverydomain.WorkflowStep,
	researchEntities []discoverydomain.ResearchedEntity,
	fallbackSource string,
) ([]discoverydomain.WorkObject, error) {
	seen := make(map[string]struct{})
	var objects []discoverydomain.WorkObject

	for _, step := range steps {
		key := strings.ToLower(step.WorkObject())
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}

		source := resolveEntitySource(step.WorkObject(), researchEntities, fallbackSource)

		wo, err := discoverydomain.NewWorkObject(
			step.WorkObject(),
			discoverydomain.WorkObjectTypeDocument,
			vo.AIResearched,
			source,
		)
		if err != nil {
			return nil, fmt.Errorf("creating work object %q: %w", step.WorkObject(), err)
		}

		objects = append(objects, wo)
	}

	return objects, nil
}

// resolveEntitySource finds the source URL for an entity name in research entities.
// Case-insensitive match. Falls back to fallbackSource.
func resolveEntitySource(entityName string, researchEntities []discoverydomain.ResearchedEntity, fallbackSource string) string {
	for _, re := range researchEntities {
		if strings.EqualFold(re.Name(), entityName) {
			source := sourceFallback(re.SourceURLs())
			if source != researchSourceFallback {
				return source
			}

			return fallbackSource
		}
	}

	return researchSourceFallback
}

// sourceFallback returns the first non-whitespace URL, or "domain research".
func sourceFallback(sourceURLs []string) string {
	for _, url := range sourceURLs {
		trimmed := strings.TrimSpace(url)
		if trimmed != "" {
			return trimmed
		}
	}

	return researchSourceFallback
}
