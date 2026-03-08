package valueobjects

import (
	"encoding/json"
	"fmt"
)

// TermEntry is a single term in the ubiquitous language glossary.
type TermEntry struct {
	term              string
	definition        string
	contextName       string
	sourceQuestionIDs []string
}

// NewTermEntry creates a TermEntry value object.
func NewTermEntry(term, definition, contextName string, sourceQuestionIDs []string) TermEntry {
	ids := make([]string, len(sourceQuestionIDs))
	copy(ids, sourceQuestionIDs)
	return TermEntry{
		term:              term,
		definition:        definition,
		contextName:       contextName,
		sourceQuestionIDs: ids,
	}
}

// Term returns the domain term.
func (te TermEntry) Term() string { return te.term }

// Definition returns the term definition.
func (te TermEntry) Definition() string { return te.definition }

// ContextName returns the bounded context this term belongs to.
func (te TermEntry) ContextName() string { return te.contextName }

// SourceQuestionIDs returns a defensive copy of the source question IDs.
func (te TermEntry) SourceQuestionIDs() []string {
	out := make([]string, len(te.sourceQuestionIDs))
	copy(out, te.sourceQuestionIDs)
	return out
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (te TermEntry) MarshalJSON() ([]byte, error) {
	type proxy struct {
		Term              string   `json:"term"`
		Definition        string   `json:"definition"`
		ContextName       string   `json:"context_name"`
		SourceQuestionIDs []string `json:"source_question_ids"`
	}
	data, err := json.Marshal(proxy{
		Term:              te.term,
		Definition:        te.definition,
		ContextName:       te.contextName,
		SourceQuestionIDs: te.sourceQuestionIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling TermEntry: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (te *TermEntry) UnmarshalJSON(data []byte) error {
	type proxy struct {
		Term              string   `json:"term"`
		Definition        string   `json:"definition"`
		ContextName       string   `json:"context_name"`
		SourceQuestionIDs []string `json:"source_question_ids"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling TermEntry: %w", err)
	}
	te.term = p.Term
	te.definition = p.Definition
	te.contextName = p.ContextName
	te.sourceQuestionIDs = p.SourceQuestionIDs
	return nil
}

// WithContextName returns a new TermEntry with a different context name.
func (te TermEntry) WithContextName(contextName string) TermEntry {
	ids := make([]string, len(te.sourceQuestionIDs))
	copy(ids, te.sourceQuestionIDs)
	return TermEntry{
		term:              te.term,
		definition:        te.definition,
		contextName:       contextName,
		sourceQuestionIDs: ids,
	}
}
