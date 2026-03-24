package domain

import (
	"fmt"
	"sort"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// DomainStory is the aggregate root for a domain story telling session.
// It is mutable — Add* methods mutate in place.
//
//nolint:revive // "DomainStory" is the ubiquitous language term from Domain Storytelling methodology.
type DomainStory struct {
	title       string
	storyType   StoryType
	timeType    TimeType
	purityType  PurityType
	trigger     string
	actors      []StoryActor
	workObjects []WorkObject
	sentences   []StorySentence
	annotations []Annotation
	variations  []string
}

// NewDomainStory creates a DomainStory aggregate, validating all invariants.
func NewDomainStory(
	title string,
	storyType StoryType,
	timeType TimeType,
	purityType PurityType,
	trigger string,
) (*DomainStory, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("domain story title must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := storyType.Validate(); err != nil {
		return nil, fmt.Errorf("validating story type: %w", err)
	}

	if err := timeType.Validate(); err != nil {
		return nil, fmt.Errorf("validating time type: %w", err)
	}

	if err := purityType.Validate(); err != nil {
		return nil, fmt.Errorf("validating purity type: %w", err)
	}

	trigger = strings.TrimSpace(trigger)
	if trigger == "" {
		return nil, fmt.Errorf("domain story trigger must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	return &DomainStory{
		title:      title,
		storyType:  storyType,
		timeType:   timeType,
		purityType: purityType,
		trigger:    trigger,
	}, nil
}

// AddActor appends an actor, rejecting case-insensitive duplicate names.
func (d *DomainStory) AddActor(actor StoryActor) error {
	for _, existing := range d.actors {
		if strings.EqualFold(existing.Name(), actor.Name()) {
			return fmt.Errorf("duplicate actor %q: %w", actor.Name(), domainerrors.ErrInvariantViolation)
		}
	}

	d.actors = append(d.actors, actor)

	return nil
}

// AddWorkObject appends a work object, rejecting case-insensitive duplicate names.
func (d *DomainStory) AddWorkObject(wo WorkObject) error {
	for _, existing := range d.workObjects {
		if strings.EqualFold(existing.Name(), wo.Name()) {
			return fmt.Errorf("duplicate work object %q: %w", wo.Name(), domainerrors.ErrInvariantViolation)
		}
	}

	d.workObjects = append(d.workObjects, wo)

	return nil
}

// AddSentence appends a sentence. Step sequence validation is deferred to Validate().
func (d *DomainStory) AddSentence(sentence StorySentence) error {
	d.sentences = append(d.sentences, sentence)

	return nil
}

// AddAnnotation appends an annotation. Sentence ref validation is deferred to Validate().
func (d *DomainStory) AddAnnotation(annotation Annotation) error {
	d.annotations = append(d.annotations, annotation)

	return nil
}

// AddVariation appends a variation description, rejecting empty strings.
func (d *DomainStory) AddVariation(description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return fmt.Errorf("variation description must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	d.variations = append(d.variations, description)

	return nil
}

// Validate checks all referential integrity rules for the domain story.
func (d *DomainStory) Validate() error {
	// Rule 1: at least 1 actor, 1 work object, 1 sentence.
	if len(d.actors) == 0 {
		return fmt.Errorf("domain story must have at least one actor: %w", domainerrors.ErrInvariantViolation)
	}

	if len(d.workObjects) == 0 {
		return fmt.Errorf("domain story must have at least one work object: %w", domainerrors.ErrInvariantViolation)
	}

	if len(d.sentences) == 0 {
		return fmt.Errorf("domain story must have at least one sentence: %w", domainerrors.ErrInvariantViolation)
	}

	// Rule 2: sequential step numbers 1..N.
	steps := make([]int, len(d.sentences))
	for i, s := range d.sentences {
		steps[i] = s.Step()
	}

	sort.Ints(steps)

	for i, step := range steps {
		if step != i+1 {
			return fmt.Errorf(
				"sentence steps must be sequential 1..%d, got %d at position %d: %w",
				len(d.sentences), step, i+1, domainerrors.ErrInvariantViolation,
			)
		}
	}

	// Rule 3: every sentence subject references an actor name.
	for _, s := range d.sentences {
		if !d.hasActorNamed(s.Subject()) {
			return fmt.Errorf(
				"sentence %d subject %q does not reference a known actor: %w",
				s.Step(), s.Subject(), domainerrors.ErrInvariantViolation,
			)
		}
	}

	// Rule 4: every sentence object references an actor OR work object name.
	for _, s := range d.sentences {
		if !d.hasActorNamed(s.Object()) && !d.hasWorkObjectNamed(s.Object()) {
			return fmt.Errorf(
				"sentence %d object %q does not reference a known actor or work object: %w",
				s.Step(), s.Object(), domainerrors.ErrInvariantViolation,
			)
		}
	}

	// Rule 5: every indirect object (if set) references an actor OR work object.
	for _, s := range d.sentences {
		if s.HasIndirectObject() {
			if !d.hasActorNamed(s.IndirectObject()) && !d.hasWorkObjectNamed(s.IndirectObject()) {
				return fmt.Errorf(
					"sentence %d indirect object %q does not reference a known actor or work object: %w",
					s.Step(), s.IndirectObject(), domainerrors.ErrInvariantViolation,
				)
			}
		}
	}

	// Rule 6: annotation sentence refs are valid step numbers.
	for _, a := range d.annotations {
		ref := a.SentenceRef()
		if ref != nil && (*ref < 1 || *ref > len(d.sentences)) {
			return fmt.Errorf(
				"annotation references sentence %d, but story has %d sentences: %w",
				*ref, len(d.sentences), domainerrors.ErrInvariantViolation,
			)
		}
	}

	return nil
}

// Title returns the domain story title.
func (d *DomainStory) Title() string { return d.title }

// Type returns the story type.
func (d *DomainStory) Type() StoryType { return d.storyType }

// Time returns the time type.
func (d *DomainStory) Time() TimeType { return d.timeType }

// Purity returns the purity type.
func (d *DomainStory) Purity() PurityType { return d.purityType }

// Trigger returns the story trigger.
func (d *DomainStory) Trigger() string { return d.trigger }

// Actors returns a defensive copy of the actors slice.
func (d *DomainStory) Actors() []StoryActor {
	return append([]StoryActor(nil), d.actors...)
}

// WorkObjects returns a defensive copy of the work objects slice.
func (d *DomainStory) WorkObjects() []WorkObject {
	return append([]WorkObject(nil), d.workObjects...)
}

// Sentences returns a defensive copy of the sentences slice.
func (d *DomainStory) Sentences() []StorySentence {
	return append([]StorySentence(nil), d.sentences...)
}

// Annotations returns a defensive copy of the annotations slice.
func (d *DomainStory) Annotations() []Annotation {
	return append([]Annotation(nil), d.annotations...)
}

// Variations returns a defensive copy of the variations slice.
func (d *DomainStory) Variations() []string {
	return append([]string(nil), d.variations...)
}

// SentenceCount returns the number of sentences in the story.
func (d *DomainStory) SentenceCount() int { return len(d.sentences) }

// HasBranching returns true if any sentence contains branching keywords.
func (d *DomainStory) HasBranching() bool {
	for _, s := range d.sentences {
		if s.ContainsBranching() {
			return true
		}
	}

	return false
}

// FormatText renders the domain story in alto text format.
func (d *DomainStory) FormatText() string {
	var b strings.Builder

	// Header.
	fmt.Fprintf(&b, "# %s (%s, %s, %s)\n", d.title, d.storyType, d.timeType, d.purityType)
	fmt.Fprintf(&b, "Trigger: %s", d.trigger)

	// Actors.
	if len(d.actors) > 0 {
		b.WriteString("\n\n## Actors\n")

		for _, a := range d.actors {
			fmt.Fprintf(&b, "- %s (%s)\n", a.Name(), a.Type())
		}
	}

	// Work Objects.
	if len(d.workObjects) > 0 {
		b.WriteString("\n## Work Objects\n")

		for _, wo := range d.workObjects {
			fmt.Fprintf(&b, "- %s (%s)\n", wo.Name(), wo.Type())
		}
	}

	// Story sentences.
	if len(d.sentences) > 0 {
		b.WriteString("\n## Story\n")

		for _, s := range d.sentences {
			fmt.Fprintf(&b, "%s\n", s.FormatText())
		}
	}

	// Annotations.
	if len(d.annotations) > 0 {
		b.WriteString("\n## Annotations\n")

		for _, a := range d.annotations {
			scope := "story-wide"

			ref := a.SentenceRef()
			if ref != nil {
				scope = fmt.Sprintf("sentence %d", *ref)
			}

			fmt.Fprintf(&b, "- [%s] (%s) %s\n", a.Type(), scope, a.Text())
		}
	}

	// Variations.
	if len(d.variations) > 0 {
		b.WriteString("\n## Variations\n")

		for _, v := range d.variations {
			fmt.Fprintf(&b, "- %s\n", v)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func (d *DomainStory) hasActorNamed(name string) bool {
	for _, a := range d.actors {
		if strings.EqualFold(a.Name(), name) {
			return true
		}
	}

	return false
}

func (d *DomainStory) hasWorkObjectNamed(name string) bool {
	for _, wo := range d.workObjects {
		if strings.EqualFold(wo.Name(), name) {
			return true
		}
	}

	return false
}
