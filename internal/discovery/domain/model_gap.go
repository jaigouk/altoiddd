package domain

import (
	"fmt"
	"strings"
)

// ModelGap identifies a missing element in the domain model.
// Used by ConversationalFlow to track which phases still need answers.
type ModelGap struct {
	phase       QuestionPhase
	description string
}

// NewModelGap creates a ModelGap value object.
func NewModelGap(phase QuestionPhase, description string) (ModelGap, error) {
	if strings.TrimSpace(description) == "" {
		return ModelGap{}, fmt.Errorf("model gap description cannot be empty")
	}
	return ModelGap{phase: phase, description: description}, nil
}

// Phase returns the question phase this gap relates to.
func (g ModelGap) Phase() QuestionPhase { return g.phase }

// Description returns a human-readable description of what is missing.
func (g ModelGap) Description() string { return g.description }

// Equal returns true if two ModelGaps have the same values.
func (g ModelGap) Equal(other ModelGap) bool {
	return g.phase == other.phase && g.description == other.description
}
