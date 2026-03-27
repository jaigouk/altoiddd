package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// slugifyPattern is compiled once at package level for slugify.
var slugifyPattern = regexp.MustCompile(`[^a-z0-9]+`)

// StorytellingHandler orchestrates the interactive domain storytelling flow.
type StorytellingHandler struct {
	storyWriter StoryWriter
	prompter    StorytellingPrompter
}

// NewStorytellingHandler creates a new StorytellingHandler.
func NewStorytellingHandler(
	storyWriter StoryWriter,
	prompter StorytellingPrompter,
) *StorytellingHandler {
	return &StorytellingHandler{
		storyWriter: storyWriter,
		prompter:    prompter,
	}
}

// RunStory executes a single domain story narration session, returning the completed
// story, the conversation narrative, and any error.
func (h *StorytellingHandler) RunStory(
	ctx context.Context,
	session *discoverydomain.DiscoverySession,
	storyIndex int,
	flow *discoverydomain.StorytellingFlow,
) (*discoverydomain.DomainStory, discoverydomain.ConversationNarrative, error) {
	narrative := discoverydomain.NewConversationNarrative()

	// 1. OPENING PHASE
	register, ok := session.Register()
	if !ok {
		register = discoverydomain.RegisterNonTechnical
	}

	openingQs := discoverydomain.QuestionsByPhase(discoverydomain.NarrationPhaseOpening)

	// Ask trigger (MQ-O2)
	triggerQuestion := openingQs[1].Text(register)

	triggerResp, err := h.prompter.AskNarration(ctx, triggerQuestion, "")
	if err != nil {
		return nil, narrative, fmt.Errorf("asking trigger question: %w", err)
	}

	turn, err := discoverydomain.NewConversationTurn(triggerQuestion, triggerResp)
	if err != nil {
		return nil, narrative, fmt.Errorf("creating trigger turn: %w", err)
	}

	narrative = narrative.AddTurn(turn)

	// Ask first actor (MQ-O1)
	actorQuestion := openingQs[0].Text(register)

	actorResp, err := h.prompter.AskNarration(ctx, actorQuestion, "")
	if err != nil {
		return nil, narrative, fmt.Errorf("asking actor question: %w", err)
	}

	turn, err = discoverydomain.NewConversationTurn(actorQuestion, actorResp)
	if err != nil {
		return nil, narrative, fmt.Errorf("creating actor turn: %w", err)
	}

	narrative = narrative.AddTurn(turn)

	// Create first actor
	firstActor, err := discoverydomain.NewStoryActor(actorResp, discoverydomain.ActorTypePerson, vo.UserStated, "")
	if err != nil {
		return nil, narrative, fmt.Errorf("creating first actor: %w", err)
	}

	// Create DomainStory
	story, err := discoverydomain.NewDomainStory(
		triggerResp,
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypeDigitalized,
		triggerResp,
	)
	if err != nil {
		return nil, narrative, fmt.Errorf("creating domain story: %w", err)
	}

	if err := story.AddActor(firstActor); err != nil {
		return nil, narrative, fmt.Errorf("adding first actor: %w", err)
	}

	lastActor := actorResp

	// 2. NARRATION PHASE
	narrationQs := discoverydomain.QuestionsByPhase(discoverydomain.NarrationPhaseNarration)
	sentenceCount := 0
	sentencesSinceCheckpoint := 0

	for {
		// Ask "what happens next?" (MQ-N2)
		storyContext := story.FormatText()
		activityQuestion := narrationQs[1].Text(register)

		activityResp, err := h.prompter.AskNarration(ctx, activityQuestion, storyContext)
		if err != nil {
			return nil, narrative, fmt.Errorf("asking narration question: %w", err)
		}

		if activityResp == "" {
			break
		}

		turn, err = discoverydomain.NewConversationTurn(activityQuestion, activityResp)
		if err != nil {
			return nil, narrative, fmt.Errorf("creating narration turn: %w", err)
		}

		narrative = narrative.AddTurn(turn)

		// Ask subject (MQ-N3)
		subjectQuestion := narrationQs[2].Text(register)

		subjectResp, err := h.prompter.AskNarration(ctx, subjectQuestion, "")
		if err != nil {
			return nil, narrative, fmt.Errorf("asking subject question: %w", err)
		}

		if subjectResp == "" {
			subjectResp = lastActor
		}

		turn, err = discoverydomain.NewConversationTurn(subjectQuestion, subjectResp)
		if err != nil {
			return nil, narrative, fmt.Errorf("creating subject turn: %w", err)
		}

		narrative = narrative.AddTurn(turn)

		// Ask work object (MQ-N4)
		objectQuestion := narrationQs[3].Text(register)

		objectResp, err := h.prompter.AskNarration(ctx, objectQuestion, "")
		if err != nil {
			return nil, narrative, fmt.Errorf("asking work object question: %w", err)
		}

		turn, err = discoverydomain.NewConversationTurn(objectQuestion, objectResp)
		if err != nil {
			return nil, narrative, fmt.Errorf("creating work object turn: %w", err)
		}

		narrative = narrative.AddTurn(turn)

		// Build sentence
		sentence, err := discoverydomain.NewStorySentence(
			sentenceCount+1, subjectResp, activityResp, objectResp, vo.UserStated, "",
		)
		if err != nil {
			return nil, narrative, fmt.Errorf("creating sentence: %w", err)
		}

		// Confirm sentence
		proposedSentence := sentence
		sentence, accepted, err := h.prompter.ConfirmSentence(ctx, sentence)
		if err != nil {
			return nil, narrative, fmt.Errorf("confirming sentence: %w", err)
		}

		if accepted {
			// Propagate trust when confirming AI-researched sentences.
			if proposedSentence.Trust() == vo.AIResearched {
				sentence = discoverydomain.PropagateConfirmation(proposedSentence, sentence, accepted, story)
			}

			if err := story.AddSentence(sentence); err != nil {
				return nil, narrative, fmt.Errorf("adding sentence: %w", err)
			}

			// Ensure actor exists (ignore expected duplicate error)
			actor, err := discoverydomain.NewStoryActor(subjectResp, discoverydomain.ActorTypePerson, vo.UserStated, "")
			if err == nil {
				if addErr := story.AddActor(actor); addErr != nil && !errors.Is(addErr, domainerrors.ErrInvariantViolation) {
					return nil, narrative, fmt.Errorf("adding actor: %w", addErr)
				}
			}

			// Ensure work object exists (ignore expected duplicate error)
			wo, err := discoverydomain.NewWorkObject(objectResp, discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
			if err == nil {
				if addErr := story.AddWorkObject(wo); addErr != nil && !errors.Is(addErr, domainerrors.ErrInvariantViolation) {
					return nil, narrative, fmt.Errorf("adding work object: %w", addErr)
				}
			}

			lastActor = subjectResp
			sentenceCount++
			sentencesSinceCheckpoint++

			// Check branching on accepted sentences only
			if sentence.ContainsBranching() && flow.ShouldSuggestBranchingSplit(story) {
				branchChoices := []Choice{
					{Key: "yes", Label: "Yes", Description: "Split into separate story"},
					{Key: "no", Label: "No", Description: "Keep in current story"},
				}

				choice, err := h.prompter.AskChoice(
					ctx, "This sentence contains branching. Split into separate story?", branchChoices, "yes",
				)
				if err != nil {
					return nil, narrative, fmt.Errorf("asking branching choice: %w", err)
				}

				if choice == "yes" {
					variationResp, err := h.prompter.AskNarration(ctx, "Describe the variation:", "")
					if err != nil {
						return nil, narrative, fmt.Errorf("asking variation description: %w", err)
					}

					if err := story.AddVariation(variationResp); err != nil {
						return nil, narrative, fmt.Errorf("adding variation: %w", err)
					}
				}
			}
		}

		// Check mid-story checkpoint
		if flow.IsMidStoryCheckpointDue(sentencesSinceCheckpoint) {
			summary := SynthesisSummary{
				StoriesSoFar:    []*discoverydomain.DomainStory{story},
				ActorInventory:  story.Actors(),
				ObjectInventory: story.WorkObjects(),
			}

			_, err := h.prompter.SynthesisCheckpoint(ctx, summary)
			if err != nil {
				return nil, narrative, fmt.Errorf("synthesis checkpoint: %w", err)
			}

			sentencesSinceCheckpoint = 0
		}
	}

	// 3. DEEPENING PHASE
	for {
		text, sentenceNum, err := h.prompter.AskAnnotation(ctx)
		if err != nil {
			return nil, narrative, fmt.Errorf("asking annotation: %w", err)
		}

		if text == "" {
			break
		}

		var sentenceRef *int
		if sentenceNum > 0 {
			sentenceRef = &sentenceNum
		}

		annotation, err := discoverydomain.NewAnnotation(
			text, discoverydomain.AnnotationTypeInvariant, sentenceRef, vo.UserStated, "",
		)
		if err != nil {
			return nil, narrative, fmt.Errorf("creating annotation: %w", err)
		}

		if err := story.AddAnnotation(annotation); err != nil {
			return nil, narrative, fmt.Errorf("adding annotation: %w", err)
		}
	}

	// 4. CLOSING PHASE
	if err := h.prompter.DisplayStory(ctx, story); err != nil {
		return nil, narrative, fmt.Errorf("displaying story: %w", err)
	}

	finalSummary := SynthesisSummary{
		StoriesSoFar:    []*discoverydomain.DomainStory{story},
		ActorInventory:  story.Actors(),
		ObjectInventory: story.WorkObjects(),
	}

	if _, err := h.prompter.SynthesisCheckpoint(ctx, finalSummary); err != nil {
		return nil, narrative, fmt.Errorf("final synthesis checkpoint: %w", err)
	}

	// Validate before persisting (FAIL FAST)
	if err := story.Validate(); err != nil {
		return nil, narrative, fmt.Errorf("validating story: %w", err)
	}

	// 5. PERSIST
	storyPath := fmt.Sprintf(".alto/stories/%02d-%s.story.yaml", storyIndex, slugify(story.Title()))

	if err := h.storyWriter.Write(ctx, storyPath, story); err != nil {
		return nil, narrative, fmt.Errorf("writing story: %w", err)
	}

	if err := session.AddStoryRef(storyPath); err != nil {
		return nil, narrative, fmt.Errorf("adding story ref to session: %w", err)
	}

	return story, narrative, nil
}

// slugify converts a string to a URL-friendly slug.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugifyPattern.ReplaceAllString(s, "-")

	return strings.Trim(s, "-")
}
