package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"charm.land/huh/v2"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// Compile-time interface satisfaction check.
var _ application.StorytellingPrompter = (*HuhStorytellingPrompter)(nil)

// HuhStorytellingPrompter implements StorytellingPrompter using charmbracelet/huh v2
// for interactive TUI prompts during the Domain Storytelling discovery flow.
type HuhStorytellingPrompter struct{}

// NewHuhStorytellingPrompter creates a new HuhStorytellingPrompter.
func NewHuhStorytellingPrompter() *HuhStorytellingPrompter {
	return &HuhStorytellingPrompter{}
}

// modeOptions maps display text to DiscoveryMode string values.
var modeOptions = []huh.Option[string]{
	huh.NewOption("Rapid — fast discovery, minimal questions", string(discoverydomain.ModeRapid)),
	huh.NewOption("Thorough — deep discovery, comprehensive questions", string(discoverydomain.ModeThorough)),
}

// SelectMode prompts the user to choose between RAPID and THOROUGH discovery modes.
func (p *HuhStorytellingPrompter) SelectMode(ctx context.Context) (discoverydomain.DiscoveryMode, error) {
	if ctx.Err() != nil {
		return "", context.Canceled
	}

	var choice string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select discovery mode").
				Options(modeOptions...).
				Value(&choice),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", context.Canceled
		}

		return "", fmt.Errorf("running select mode form: %w", err)
	}

	mode, err := discoverydomain.ParseDiscoveryMode(choice)
	if err != nil {
		return "", fmt.Errorf("parsing discovery mode: %w", err)
	}

	return mode, nil
}

// ProposeStory is a stub — story proposal UI is not yet implemented.
func (p *HuhStorytellingPrompter) ProposeStory(_ context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	if proposed == nil {
		return nil, fmt.Errorf("proposed story must not be nil: %w", errors.ErrUnsupported)
	}

	return nil, fmt.Errorf("propose story not yet implemented: %w", errors.ErrUnsupported)
}

// AskNarration asks a moderator question and returns the user's narration response.
func (p *HuhStorytellingPrompter) AskNarration(ctx context.Context, question string, promptContext string) (string, error) {
	if ctx.Err() != nil {
		return "", context.Canceled
	}

	var answer string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title(question).
				Description(promptContext).
				Lines(6).
				Value(&answer),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", context.Canceled
		}

		return "", fmt.Errorf("running narration form: %w", err)
	}

	return answer, nil
}

// ConfirmSentence presents a structured sentence for confirmation.
// Returns the (possibly edited) sentence, whether it was accepted, and any error.
func (p *HuhStorytellingPrompter) ConfirmSentence(ctx context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	if ctx.Err() != nil {
		return discoverydomain.StorySentence{}, false, context.Canceled
	}

	var choice string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(sentence.FormatText()).
				Options(
					huh.NewOption("Accept", "accept"),
					huh.NewOption("Edit", "edit"),
					huh.NewOption("Reject", "reject"),
				).
				Value(&choice),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return discoverydomain.StorySentence{}, false, context.Canceled
		}

		return discoverydomain.StorySentence{}, false, fmt.Errorf("running confirm sentence form: %w", err)
	}

	switch choice {
	case "accept":
		return sentence, true, nil
	case "reject":
		return sentence, false, nil
	case "edit":
		return p.runEditFlow(ctx, sentence)
	default:
		return discoverydomain.StorySentence{}, false, fmt.Errorf("unexpected choice %q: %w", choice, errors.ErrUnsupported)
	}
}

// runEditFlow presents edit forms for sentence fields and builds an edited sentence.
func (p *HuhStorytellingPrompter) runEditFlow(ctx context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	editedSubject := sentence.Subject()
	editedActivity := sentence.Activity()
	editedObject := sentence.Object()

	editForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Subject (actor)").
				Value(&editedSubject),
			huh.NewInput().
				Title("Activity (verb)").
				Value(&editedActivity),
			huh.NewInput().
				Title("Object (work object)").
				Value(&editedObject),
		),
	)

	if err := editForm.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return discoverydomain.StorySentence{}, false, context.Canceled
		}

		return discoverydomain.StorySentence{}, false, fmt.Errorf("running edit form: %w", err)
	}

	var editedPreposition, editedIndirectObject string

	if sentence.HasIndirectObject() {
		editedPreposition = sentence.Preposition()
		editedIndirectObject = sentence.IndirectObject()

		prepForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Preposition").
					Value(&editedPreposition),
				huh.NewInput().
					Title("Indirect object").
					Value(&editedIndirectObject),
			),
		)

		if err := prepForm.RunWithContext(ctx); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return discoverydomain.StorySentence{}, false, context.Canceled
			}

			return discoverydomain.StorySentence{}, false, fmt.Errorf("running preposition edit form: %w", err)
		}
	}

	edited, err := buildEditedSentence(sentence, editedSubject, editedActivity, editedObject, editedPreposition, editedIndirectObject)
	if err != nil {
		return discoverydomain.StorySentence{}, false, fmt.Errorf("building edited sentence: %w", err)
	}

	return edited, true, nil
}

// AskChoice presents lettered options with an optional recommendation.
// Returns the selected option key.
func (p *HuhStorytellingPrompter) AskChoice(ctx context.Context, prompt string, options []application.Choice, recommended string) (string, error) {
	if ctx.Err() != nil {
		return "", context.Canceled
	}

	huhOptions := make([]huh.Option[string], 0, len(options))
	for _, opt := range options {
		huhOptions = append(huhOptions, huh.NewOption(opt.Label, opt.Key))
	}

	sel := huh.NewSelect[string]().
		Title(prompt).
		Options(huhOptions...)

	if recommended != "" {
		sel = sel.Description(fmt.Sprintf("Recommended: %s", recommended))
	}

	var choice string

	form := huh.NewForm(
		huh.NewGroup(
			sel.Value(&choice),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", context.Canceled
		}

		return "", fmt.Errorf("running choice form: %w", err)
	}

	return choice, nil
}

// DisplayStory renders a complete domain story for read-only display.
func (p *HuhStorytellingPrompter) DisplayStory(ctx context.Context, story *discoverydomain.DomainStory) error {
	if ctx.Err() != nil {
		return context.Canceled
	}

	if story == nil {
		return fmt.Errorf("story must not be nil")
	}

	fmt.Println(story.FormatText()) //nolint:forbidigo // intentional user-facing output

	return nil
}

// SynthesisCheckpoint presents a synthesis summary for user confirmation.
func (p *HuhStorytellingPrompter) SynthesisCheckpoint(ctx context.Context, synthesis application.SynthesisSummary) (bool, error) {
	if ctx.Err() != nil {
		return false, context.Canceled
	}

	summary := fmt.Sprintf(
		"Stories: %d | Actors: %d | Objects: %d | Signals: %d | Glossary terms: %d",
		len(synthesis.StoriesSoFar),
		len(synthesis.ActorInventory),
		len(synthesis.ObjectInventory),
		len(synthesis.BoundarySignals),
		len(synthesis.GlossaryTerms),
	)

	fmt.Println(summary) //nolint:forbidigo // intentional user-facing output

	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Does this synthesis look correct?").
				Affirmative("Yes, continue").
				Negative("No, let me revise").
				Value(&confirmed),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, context.Canceled
		}

		return false, fmt.Errorf("running synthesis checkpoint form: %w", err)
	}

	return confirmed, nil
}

// AskAnnotation prompts for a business rule annotation.
// Returns (annotation text, sentence number, error). Empty text = done. Sentence 0 = story-wide.
func (p *HuhStorytellingPrompter) AskAnnotation(ctx context.Context) (string, int, error) {
	if ctx.Err() != nil {
		return "", 0, context.Canceled
	}

	var wantsAnnotation bool

	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add a business rule annotation?").
				Affirmative("Yes").
				Negative("No").
				Value(&wantsAnnotation),
		),
	)

	if err := confirmForm.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", 0, context.Canceled
		}

		return "", 0, fmt.Errorf("running annotation confirm form: %w", err)
	}

	if !wantsAnnotation {
		return "", 0, nil
	}

	var annotationText string

	textForm := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Annotation text").
				Description("Describe the business rule").
				Lines(4).
				Value(&annotationText),
		),
	)

	if err := textForm.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", 0, context.Canceled
		}

		return "", 0, fmt.Errorf("running annotation text form: %w", err)
	}

	var sentenceNumStr string

	numForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Sentence number (0 = story-wide)").
				Placeholder("0").
				Value(&sentenceNumStr),
		),
	)

	if err := numForm.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", 0, context.Canceled
		}

		return "", 0, fmt.Errorf("running annotation sentence number form: %w", err)
	}

	if sentenceNumStr == "" {
		sentenceNumStr = "0"
	}

	sentenceNum, err := strconv.Atoi(sentenceNumStr)
	if err != nil {
		return "", 0, fmt.Errorf("parsing sentence number %q: %w", sentenceNumStr, err)
	}

	return annotationText, sentenceNum, nil
}
