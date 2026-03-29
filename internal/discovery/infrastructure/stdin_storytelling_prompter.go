package infrastructure

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// Compile-time interface satisfaction check.
var _ application.StorytellingPrompter = (*StdinStorytellingPrompter)(nil)

// StdinStorytellingPrompter implements StorytellingPrompter using plain stdin/stdout
// for accessibility and CI.
type StdinStorytellingPrompter struct {
	scanner *bufio.Scanner
	writer  io.Writer
}

// NewStdinStorytellingPrompter creates a new StdinStorytellingPrompter with the given reader and writer.
func NewStdinStorytellingPrompter(r io.Reader, w io.Writer) *StdinStorytellingPrompter {
	return &StdinStorytellingPrompter{
		scanner: bufio.NewScanner(r),
		writer:  w,
	}
}

// scanOrCancel reads the next line, returning context.Canceled on EOF
// or wrapping any scanner error.
func (p *StdinStorytellingPrompter) scanOrCancel() (string, error) {
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		return "", context.Canceled // EOF
	}
	return strings.TrimSpace(p.scanner.Text()), nil
}

// SelectMode prompts the user to choose between RAPID and THOROUGH discovery modes.
func (p *StdinStorytellingPrompter) SelectMode(_ context.Context) (discoverydomain.DiscoveryMode, error) {
	_, _ = fmt.Fprintln(p.writer, "Select discovery mode:")
	_, _ = fmt.Fprintln(p.writer, "1. Rapid — fast discovery, minimal questions")
	_, _ = fmt.Fprintln(p.writer, "2. Thorough — deep discovery, comprehensive questions")
	_, _ = fmt.Fprint(p.writer, "Enter choice (1-2): ")

	choice, err := p.scanOrCancel()
	if err != nil {
		return "", err
	}

	switch choice {
	case "1":
		return discoverydomain.ModeRapid, nil
	case "2":
		return discoverydomain.ModeThorough, nil
	default:
		return "", fmt.Errorf("invalid mode selection: %q", choice)
	}
}

// AskNarration asks a moderator question and returns the user's narration response.
func (p *StdinStorytellingPrompter) AskNarration(_ context.Context, question string, promptContext string) (string, error) {
	_, _ = fmt.Fprintln(p.writer, question)
	if promptContext != "" {
		_, _ = fmt.Fprintln(p.writer, promptContext)
	}
	_, _ = fmt.Fprint(p.writer, "> ")

	return p.scanOrCancel()
}

// ConfirmSentence presents a structured sentence for confirmation.
// Returns the (possibly edited) sentence, whether it was accepted, and any error.
func (p *StdinStorytellingPrompter) ConfirmSentence(_ context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	_, _ = fmt.Fprintln(p.writer, sentence.FormatText())
	_, _ = fmt.Fprint(p.writer, "[a]ccept / [r]eject / [e]dit: ")

	choice, err := p.scanOrCancel()
	if err != nil {
		return discoverydomain.StorySentence{}, false, err
	}

	switch strings.ToLower(choice) {
	case "a":
		return sentence, true, nil
	case "r":
		return sentence, false, nil
	case "e":
		return p.runSentenceEditFlow(sentence)
	default:
		return discoverydomain.StorySentence{}, false, fmt.Errorf("invalid choice: %q", choice)
	}
}

// runSentenceEditFlow prompts for edited sentence fields and builds an edited sentence.
func (p *StdinStorytellingPrompter) runSentenceEditFlow(sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	subject, err := p.promptField("Subject", sentence.Subject())
	if err != nil {
		return discoverydomain.StorySentence{}, false, err
	}

	activity, err := p.promptField("Activity", sentence.Activity())
	if err != nil {
		return discoverydomain.StorySentence{}, false, err
	}

	object, err := p.promptField("Object", sentence.Object())
	if err != nil {
		return discoverydomain.StorySentence{}, false, err
	}

	var preposition, indirectObject string
	if sentence.HasIndirectObject() {
		preposition, err = p.promptField("Preposition", sentence.Preposition())
		if err != nil {
			return discoverydomain.StorySentence{}, false, err
		}

		indirectObject, err = p.promptField("Indirect object", sentence.IndirectObject())
		if err != nil {
			return discoverydomain.StorySentence{}, false, err
		}
	}

	edited, err := buildEditedSentence(sentence, subject, activity, object, preposition, indirectObject)
	if err != nil {
		return discoverydomain.StorySentence{}, false, fmt.Errorf("building edited sentence: %w", err)
	}

	return edited, true, nil
}

// promptField displays a field prompt with its current value and reads the replacement.
// Returns the original value if the user enters an empty line.
func (p *StdinStorytellingPrompter) promptField(label, current string) (string, error) {
	_, _ = fmt.Fprintf(p.writer, "%s [%s]: ", label, current)

	value, err := p.scanOrCancel()
	if err != nil {
		return "", err
	}

	if value == "" {
		return current, nil
	}

	return value, nil
}

// AskChoice presents lettered options with an optional recommendation.
// Returns the selected option key.
func (p *StdinStorytellingPrompter) AskChoice(_ context.Context, prompt string, options []application.Choice, recommended string) (string, error) {
	_, _ = fmt.Fprintln(p.writer, prompt)

	validKeys := make(map[string]bool, len(options))
	for _, opt := range options {
		validKeys[opt.Key] = true
		line := fmt.Sprintf("  [%s] %s", opt.Key, opt.Label)
		if opt.Description != "" {
			line += " — " + opt.Description
		}
		if opt.Key == recommended {
			line += " (recommended)"
		}
		_, _ = fmt.Fprintln(p.writer, line)
	}

	_, _ = fmt.Fprint(p.writer, "Enter choice: ")

	choice, err := p.scanOrCancel()
	if err != nil {
		return "", err
	}

	if !validKeys[choice] {
		return "", fmt.Errorf("invalid choice: %q", choice)
	}

	return choice, nil
}

// DisplayStory renders a complete domain story for read-only display.
func (p *StdinStorytellingPrompter) DisplayStory(_ context.Context, story *discoverydomain.DomainStory) error {
	if story == nil {
		return fmt.Errorf("story must not be nil")
	}

	_, _ = fmt.Fprintln(p.writer, story.FormatText())

	return nil
}

// SynthesisCheckpoint presents a synthesis summary for user confirmation.
func (p *StdinStorytellingPrompter) SynthesisCheckpoint(_ context.Context, synthesis application.SynthesisSummary) (bool, error) {
	summary := fmt.Sprintf(
		"Stories: %d | Actors: %d | Objects: %d | Signals: %d | Glossary terms: %d",
		len(synthesis.StoriesSoFar),
		len(synthesis.ActorInventory),
		len(synthesis.ObjectInventory),
		len(synthesis.BoundarySignals),
		len(synthesis.GlossaryTerms),
	)

	_, _ = fmt.Fprintln(p.writer, summary)
	_, _ = fmt.Fprint(p.writer, "Does this look correct? (y/n): ")

	answer, err := p.scanOrCancel()
	if err != nil {
		return false, err
	}

	answer = strings.ToLower(answer)

	return answer == "y" || answer == "yes", nil
}

// AskAnnotation prompts for a business rule annotation.
// Returns (annotation text, sentence number, error). Empty text = done. Sentence 0 = story-wide.
func (p *StdinStorytellingPrompter) AskAnnotation(_ context.Context) (string, int, error) {
	_, _ = fmt.Fprint(p.writer, "Add a business rule annotation? (y/n): ")

	answer, err := p.scanOrCancel()
	if err != nil {
		return "", 0, err
	}

	if strings.ToLower(answer) != "y" {
		return "", 0, nil
	}

	_, _ = fmt.Fprintln(p.writer, "Enter annotation text (blank line to finish):")

	var lines []string
	for {
		line, readErr := p.scanOrCancel()
		if readErr != nil {
			return "", 0, readErr
		}
		if line == "" {
			break
		}
		lines = append(lines, line)
	}

	text := strings.Join(lines, "\n")

	_, _ = fmt.Fprint(p.writer, "Sentence number (0 = story-wide): ")

	numStr, err := p.scanOrCancel()
	if err != nil {
		return "", 0, err
	}

	if numStr == "" {
		return text, 0, nil
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", 0, fmt.Errorf("parsing sentence number %q: %w", numStr, err)
	}

	return text, num, nil
}

// ProposeStory presents a consultant-proposed story for user review and refinement.
func (p *StdinStorytellingPrompter) ProposeStory(_ context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	if proposed == nil {
		return nil, fmt.Errorf("proposed story must not be nil")
	}

	_, _ = fmt.Fprintln(p.writer, proposed.FormatText())
	_, _ = fmt.Fprint(p.writer, "[y]es / [n]o / [e]dit: ")

	choice, err := p.scanOrCancel()
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(choice) {
	case "y", "yes":
		return proposed, nil
	case "n", "no":
		return nil, nil
	case "e", "edit":
		return p.runStoryEditFlow(proposed)
	default:
		return nil, fmt.Errorf("invalid choice: %q", choice)
	}
}

// runStoryEditFlow prompts for edited story title and trigger, then rebuilds the story.
func (p *StdinStorytellingPrompter) runStoryEditFlow(original *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	title, err := p.promptField("Title", original.Title())
	if err != nil {
		return nil, err
	}

	trigger, err := p.promptField("Trigger", original.Trigger())
	if err != nil {
		return nil, err
	}

	newStory, err := discoverydomain.NewDomainStory(
		title, original.Type(), original.Time(), original.Purity(), trigger,
	)
	if err != nil {
		return nil, fmt.Errorf("creating edited story: %w", err)
	}

	for _, a := range original.Actors() {
		if addErr := newStory.AddActor(a); addErr != nil {
			return nil, fmt.Errorf("replaying actor: %w", addErr)
		}
	}

	for _, wo := range original.WorkObjects() {
		if addErr := newStory.AddWorkObject(wo); addErr != nil {
			return nil, fmt.Errorf("replaying work object: %w", addErr)
		}
	}

	for _, s := range original.Sentences() {
		if addErr := newStory.AddSentence(s); addErr != nil {
			return nil, fmt.Errorf("replaying sentence: %w", addErr)
		}
	}

	for _, a := range original.Annotations() {
		if addErr := newStory.AddAnnotation(a); addErr != nil {
			return nil, fmt.Errorf("replaying annotation: %w", addErr)
		}
	}

	for _, v := range original.Variations() {
		if addErr := newStory.AddVariation(v); addErr != nil {
			return nil, fmt.Errorf("replaying variation: %w", addErr)
		}
	}

	return newStory, nil
}
