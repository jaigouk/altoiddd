package valueobjects

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
