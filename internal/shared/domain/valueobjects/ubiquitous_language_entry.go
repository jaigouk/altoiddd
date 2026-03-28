package valueobjects

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// UbiquitousLanguageEntry represents a term in the ubiquitous language glossary
// within a bounded context.
type UbiquitousLanguageEntry struct {
	term       string
	definition string
	context    string
	trust      TrustLevel
	source     string
	aliases    []string
	note       string
	stories    []string
}

// NewUbiquitousLanguageEntry creates a UbiquitousLanguageEntry, enforcing all domain invariants.
func NewUbiquitousLanguageEntry(
	term, definition, context string,
	trust TrustLevel,
	source string,
) (UbiquitousLanguageEntry, error) {
	term = strings.TrimSpace(term)
	if term == "" {
		return UbiquitousLanguageEntry{}, fmt.Errorf("term must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	definition = strings.TrimSpace(definition)
	if definition == "" {
		return UbiquitousLanguageEntry{}, fmt.Errorf("definition must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	context = strings.TrimSpace(context)
	if context == "" {
		return UbiquitousLanguageEntry{}, fmt.Errorf("context must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := trust.Validate(); err != nil {
		return UbiquitousLanguageEntry{}, fmt.Errorf("validating trust level: %w", err)
	}

	source = strings.TrimSpace(source)
	if trust == AIResearched && source == "" {
		return UbiquitousLanguageEntry{}, fmt.Errorf("source required when trust is ai_researched: %w", domainerrors.ErrInvariantViolation)
	}

	return UbiquitousLanguageEntry{
		term:       term,
		definition: definition,
		context:    context,
		trust:      trust,
		source:     source,
	}, nil
}

// Term returns the entry's term.
func (e UbiquitousLanguageEntry) Term() string { return e.term }

// Definition returns the entry's definition.
func (e UbiquitousLanguageEntry) Definition() string { return e.definition }

// Context returns the bounded context this term belongs to.
func (e UbiquitousLanguageEntry) Context() string { return e.context }

// Trust returns the entry's trust level.
func (e UbiquitousLanguageEntry) Trust() TrustLevel { return e.trust }

// Source returns the entry's source reference.
func (e UbiquitousLanguageEntry) Source() string { return e.source }

// Aliases returns a defensive copy of the entry's aliases.
func (e UbiquitousLanguageEntry) Aliases() []string {
	out := make([]string, len(e.aliases))
	copy(out, e.aliases)

	return out
}

// Note returns the entry's additional note.
func (e UbiquitousLanguageEntry) Note() string { return e.note }

// Stories returns a defensive copy of the story references where this term appears.
func (e UbiquitousLanguageEntry) Stories() []string {
	out := make([]string, len(e.stories))
	copy(out, e.stories)

	return out
}

// WithAliases returns a new UbiquitousLanguageEntry with the given aliases.
func (e UbiquitousLanguageEntry) WithAliases(aliases []string) UbiquitousLanguageEntry {
	aliasesCopy := make([]string, len(aliases))
	copy(aliasesCopy, aliases)

	return UbiquitousLanguageEntry{
		term:       e.term,
		definition: e.definition,
		context:    e.context,
		trust:      e.trust,
		source:     e.source,
		aliases:    aliasesCopy,
		note:       e.note,
		stories:    e.stories,
	}
}

// WithNote returns a new UbiquitousLanguageEntry with the given note.
func (e UbiquitousLanguageEntry) WithNote(note string) UbiquitousLanguageEntry {
	return UbiquitousLanguageEntry{
		term:       e.term,
		definition: e.definition,
		context:    e.context,
		trust:      e.trust,
		source:     e.source,
		aliases:    e.aliases,
		note:       note,
		stories:    e.stories,
	}
}

// WithTrust returns a new UbiquitousLanguageEntry with the given trust level.
// Trust can only be upgraded (lower numeric value = higher trust).
// If the requested trust is not higher than the current trust, the original entry is returned.
// Source is cleared when upgrading away from AIResearched.
func (e UbiquitousLanguageEntry) WithTrust(trust TrustLevel) UbiquitousLanguageEntry {
	if !trust.IsHigherTrust(e.trust) {
		return e
	}

	newSource := e.source
	if e.trust == AIResearched {
		newSource = ""
	}

	return UbiquitousLanguageEntry{
		term:       e.term,
		definition: e.definition,
		context:    e.context,
		trust:      trust,
		source:     newSource,
		aliases:    e.aliases,
		note:       e.note,
		stories:    e.stories,
	}
}

// WithStories returns a new UbiquitousLanguageEntry with the given story references.
func (e UbiquitousLanguageEntry) WithStories(stories []string) UbiquitousLanguageEntry {
	storiesCopy := make([]string, len(stories))
	copy(storiesCopy, stories)

	return UbiquitousLanguageEntry{
		term:       e.term,
		definition: e.definition,
		context:    e.context,
		trust:      e.trust,
		source:     e.source,
		aliases:    e.aliases,
		note:       e.note,
		stories:    storiesCopy,
	}
}

// HasAlias returns true if the entry has the given alias (case-insensitive).
func (e UbiquitousLanguageEntry) HasAlias(candidate string) bool {
	for _, a := range e.aliases {
		if strings.EqualFold(a, candidate) {
			return true
		}
	}

	return false
}

// String returns a human-readable representation of the entry.
func (e UbiquitousLanguageEntry) String() string {
	aliasStr := strings.Join(e.aliases, ", ")

	return fmt.Sprintf("UbiquitousLanguageEntry: %s (%s, %s)", e.term, aliasStr, e.trust)
}
