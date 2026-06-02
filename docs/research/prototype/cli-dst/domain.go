// Package main contains the CLI Domain Storytelling prototype.
// This is a throwaway research artifact, not production code.
package main

import (
	"fmt"
	"strings"
)

// TrustLevel indicates how a piece of domain knowledge was acquired.
type TrustLevel string

const (
	TrustUserStated    TrustLevel = "user_stated"
	TrustUserConfirmed TrustLevel = "user_confirmed"
	TrustAIResearched  TrustLevel = "ai_researched"
	TrustAIInferred    TrustLevel = "ai_inferred"
)

// StoryType indicates the granularity of a domain story.
type StoryType string

const (
	StoryTypeCoarse StoryType = "coarse-grained"
	StoryTypeFine   StoryType = "fine-grained"
)

// StorySentence represents one numbered activity in a domain story.
type StorySentence struct {
	Number         int
	Subject        string // Actor performing the activity
	Activity       string // Domain verb
	Object         string // Work object acted upon
	Preposition    string // Optional: "with", "using", "for"
	IndirectObject string // Optional: second actor or object
	TrustLevel     TrustLevel
}

// Format returns a human-readable sentence string.
func (s StorySentence) Format() string {
	base := fmt.Sprintf("%d. %s %s %s", s.Number, s.Subject, s.Activity, s.Object)
	if s.Preposition != "" && s.IndirectObject != "" {
		base += fmt.Sprintf(" %s %s", s.Preposition, s.IndirectObject)
	}
	return base
}

// StoryActor represents a named actor in a domain story.
type StoryActor struct {
	Name       string
	Type       string // "person", "system", "group"
	TrustLevel TrustLevel
}

// WorkObject represents a named work object in a domain story.
type WorkObject struct {
	Name       string
	Type       string // "document", "item", "system"
	TrustLevel TrustLevel
}

// Annotation represents a business rule or constraint on a story.
type Annotation struct {
	Text       string
	SentenceNo int    // 0 means story-level, >0 means specific sentence
	Type       string // "constraint", "invariant", "assumption"
	TrustLevel TrustLevel
}

// DomainStory is the core artifact of Domain Storytelling.
type DomainStory struct {
	Title       string
	StoryType   StoryType
	Trigger     string
	Actors      []StoryActor
	WorkObjects []WorkObject
	Sentences   []StorySentence
	Annotations []Annotation
	Variations  []string
}

// FormatText renders the story in alto text format.
func (ds *DomainStory) FormatText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Domain Story: %q\n", ds.Title))
	sb.WriteString(fmt.Sprintf("Type: %s\n", ds.StoryType))
	if ds.Trigger != "" {
		sb.WriteString(fmt.Sprintf("Trigger: %s\n", ds.Trigger))
	}
	sb.WriteString("\n")

	// Actors
	if len(ds.Actors) > 0 {
		names := make([]string, len(ds.Actors))
		for i, a := range ds.Actors {
			names[i] = a.Name
		}
		sb.WriteString(fmt.Sprintf("Actors: %s\n", strings.Join(names, ", ")))
	}

	// Work Objects
	if len(ds.WorkObjects) > 0 {
		names := make([]string, len(ds.WorkObjects))
		for i, w := range ds.WorkObjects {
			names[i] = w.Name
		}
		sb.WriteString(fmt.Sprintf("Work Objects: %s\n", strings.Join(names, ", ")))
	}

	sb.WriteString("\n")

	// Sentences
	for _, s := range ds.Sentences {
		sb.WriteString(s.Format())
		sb.WriteString("\n")
	}

	// Annotations
	if len(ds.Annotations) > 0 {
		sb.WriteString("\nAnnotations:\n")
		for _, a := range ds.Annotations {
			if a.SentenceNo > 0 {
				sb.WriteString(fmt.Sprintf("  [%d] %s\n", a.SentenceNo, a.Text))
			} else {
				sb.WriteString(fmt.Sprintf("  [%s] %s\n", a.Type, a.Text))
			}
		}
	}

	// Variations
	if len(ds.Variations) > 0 {
		sb.WriteString("\nVariations:\n")
		for _, v := range ds.Variations {
			sb.WriteString(fmt.Sprintf("  -> %q (separate story)\n", v))
		}
	}

	return sb.String()
}

// DiscoveryMode controls the depth of the discovery session.
type DiscoveryMode string

const (
	ModeRapid    DiscoveryMode = "RAPID"
	ModeThorough DiscoveryMode = "THOROUGH"
)

// InteractionPattern defines how the moderator engages the user.
type InteractionPattern string

const (
	PatternConsultantProposes InteractionPattern = "consultant-proposes"
	PatternUserNarrates       InteractionPattern = "user-narrates"
)
