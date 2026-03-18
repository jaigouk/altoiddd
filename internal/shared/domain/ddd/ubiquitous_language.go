package ddd

import (
	"fmt"
	"sort"
	"strings"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// UbiquitousLanguage is an entity managing the shared vocabulary of a domain model.
// Terms are added with their definition and bounded context. The entity can detect
// ambiguous terms (same term, different contexts) and verify that all terms appear
// in domain stories.
type UbiquitousLanguage struct {
	terms []vo.TermEntry
}

// NewUbiquitousLanguage creates an empty UbiquitousLanguage entity.
func NewUbiquitousLanguage() *UbiquitousLanguage {
	return &UbiquitousLanguage{}
}

// Terms returns a defensive copy of all term entries.
func (ul *UbiquitousLanguage) Terms() []vo.TermEntry {
	out := make([]vo.TermEntry, len(ul.terms))
	copy(out, ul.terms)
	return out
}

// AddTerm adds a term to the glossary.
func (ul *UbiquitousLanguage) AddTerm(term, definition, contextName string, sourceQuestionIDs []string) error {
	if strings.TrimSpace(term) == "" {
		return fmt.Errorf("term cannot be empty")
	}
	if strings.TrimSpace(definition) == "" {
		return fmt.Errorf("definition cannot be empty")
	}
	ul.terms = append(ul.terms, vo.NewTermEntry(
		strings.TrimSpace(term),
		strings.TrimSpace(definition),
		contextName,
		sourceQuestionIDs,
	))
	return nil
}

// GetTermsForContext returns all terms belonging to a specific bounded context.
func (ul *UbiquitousLanguage) GetTermsForContext(contextName string) []vo.TermEntry {
	var result []vo.TermEntry
	for _, t := range ul.terms {
		if t.ContextName() == contextName {
			result = append(result, t)
		}
	}
	if result == nil {
		return []vo.TermEntry{}
	}
	return result
}

// FindAmbiguousTerms returns terms that appear in multiple bounded contexts.
func (ul *UbiquitousLanguage) FindAmbiguousTerms() []string {
	termContexts := make(map[string]map[string]struct{})
	for _, entry := range ul.terms {
		normalized := strings.ToLower(entry.Term())
		if _, ok := termContexts[normalized]; !ok {
			termContexts[normalized] = make(map[string]struct{})
		}
		termContexts[normalized][entry.ContextName()] = struct{}{}
	}

	var ambiguous []string
	for term, contexts := range termContexts {
		if len(contexts) > 1 {
			ambiguous = append(ambiguous, term)
		}
	}
	sort.Strings(ambiguous)
	return ambiguous
}

// HasPerContextDefinitions checks if an ambiguous term has a definition in each context.
func (ul *UbiquitousLanguage) HasPerContextDefinitions(term string) bool {
	normalized := strings.ToLower(term)
	var entries []vo.TermEntry
	for _, e := range ul.terms {
		if strings.ToLower(e.Term()) == normalized {
			entries = append(entries, e)
		}
	}

	contexts := make(map[string]struct{})
	for _, e := range entries {
		contexts[e.ContextName()] = struct{}{}
	}

	for ctx := range contexts {
		hasDefinition := false
		for _, e := range entries {
			if e.ContextName() == ctx && strings.TrimSpace(e.Definition()) != "" {
				hasDefinition = true
				break
			}
		}
		if !hasDefinition {
			return false
		}
	}
	return true
}

// AllTermNames returns all unique term names (lowercased).
func (ul *UbiquitousLanguage) AllTermNames() map[string]struct{} {
	result := make(map[string]struct{})
	for _, e := range ul.terms {
		result[strings.ToLower(e.Term())] = struct{}{}
	}
	return result
}

// SetTerms replaces the internal terms list (used by DomainModel.ReassignTermsToContext).
func (ul *UbiquitousLanguage) SetTerms(terms []vo.TermEntry) {
	ul.terms = make([]vo.TermEntry, len(terms))
	copy(ul.terms, terms)
}

// AddTermEntry appends a raw TermEntry bypassing validation (used for testing edge cases).
func (ul *UbiquitousLanguage) AddTermEntry(entry vo.TermEntry) {
	ul.terms = append(ul.terms, entry)
}
