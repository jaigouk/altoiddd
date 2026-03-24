package domain

// ConversationNarrative is an ordered collection of conversation turns in a discovery session.
type ConversationNarrative struct {
	turns []ConversationTurn
}

// NewConversationNarrative creates an empty ConversationNarrative.
func NewConversationNarrative() ConversationNarrative {
	return ConversationNarrative{}
}

// AddTurn returns a new ConversationNarrative with the given turn appended.
func (n ConversationNarrative) AddTurn(turn ConversationTurn) ConversationNarrative {
	newTurns := make([]ConversationTurn, len(n.turns), len(n.turns)+1)
	copy(newTurns, n.turns)

	return ConversationNarrative{
		turns: append(newTurns, turn),
	}
}

// TurnCount returns the number of turns in the narrative.
func (n ConversationNarrative) TurnCount() int {
	return len(n.turns)
}

// LastTurn returns a pointer to a copy of the last turn, or nil if the narrative is empty.
func (n ConversationNarrative) LastTurn() *ConversationTurn {
	if len(n.turns) == 0 {
		return nil
	}

	last := n.turns[len(n.turns)-1]

	return &last
}

// Turns returns a defensive copy of all turns.
func (n ConversationNarrative) Turns() []ConversationTurn {
	out := make([]ConversationTurn, len(n.turns))
	copy(out, n.turns)

	return out
}

// SynthesisCheckpoints returns only turns that have a non-empty synthesis.
func (n ConversationNarrative) SynthesisCheckpoints() []ConversationTurn {
	var checkpoints []ConversationTurn

	for _, turn := range n.turns {
		if turn.Synthesis() != "" {
			checkpoints = append(checkpoints, turn)
		}
	}

	return checkpoints
}
