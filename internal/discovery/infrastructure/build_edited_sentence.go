package infrastructure

import (
	"fmt"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// buildEditedSentence creates a new StorySentence from edited field values,
// preserving the original's step number and source.
// If preposition is non-empty, it attaches the indirect object via WithPreposition.
func buildEditedSentence(
	original discoverydomain.StorySentence,
	subject, activity, object, preposition, indirectObject string,
) (discoverydomain.StorySentence, error) {
	sentence, err := discoverydomain.NewStorySentence(
		original.Step(), subject, activity, object, vo.UserStated, original.Source(),
	)
	if err != nil {
		return discoverydomain.StorySentence{}, fmt.Errorf("building edited sentence: %w", err)
	}

	if preposition != "" {
		sentence, err = sentence.WithPreposition(preposition, indirectObject)
		if err != nil {
			return discoverydomain.StorySentence{}, fmt.Errorf("setting preposition on edited sentence: %w", err)
		}
	}

	return sentence, nil
}
