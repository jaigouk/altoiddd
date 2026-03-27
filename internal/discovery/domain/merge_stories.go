package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// MergeStories merges a narrated (user-driven) story with a proposed (AI-generated)
// story. Narrated metadata, sentences, annotations, and variations are preserved.
// Actors and work objects are reconciled: higher trust wins on name collision
// (case-insensitive), and novel proposed entries are appended.
func MergeStories(narrated, proposed *DomainStory) (*DomainStory, error) {
	if narrated == nil {
		return nil, fmt.Errorf("narrated story must not be nil: %w", domainerrors.ErrInvariantViolation)
	}

	if proposed == nil {
		return nil, fmt.Errorf("proposed story must not be nil: %w", domainerrors.ErrInvariantViolation)
	}

	merged, err := NewDomainStory(
		narrated.Title(),
		narrated.Type(),
		narrated.Time(),
		narrated.Purity(),
		narrated.Trigger(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating merged story: %w", err)
	}

	// Resolve actors: higher trust wins on collision, novel proposed actors appended.
	for _, actor := range resolveActorList(narrated.Actors(), proposed.Actors()) {
		if err := merged.AddActor(actor); err != nil {
			return nil, fmt.Errorf("adding actor %q: %w", actor.Name(), err)
		}
	}

	// Resolve work objects: same strategy.
	for _, wo := range resolveWorkObjectList(narrated.WorkObjects(), proposed.WorkObjects()) {
		if err := merged.AddWorkObject(wo); err != nil {
			return nil, fmt.Errorf("adding work object %q: %w", wo.Name(), err)
		}
	}

	// Sentences, annotations, variations from narrated only.
	for _, s := range narrated.Sentences() {
		if err := merged.AddSentence(s); err != nil {
			return nil, fmt.Errorf("adding sentence %d: %w", s.Step(), err)
		}
	}

	for _, a := range narrated.Annotations() {
		if err := merged.AddAnnotation(a); err != nil {
			return nil, fmt.Errorf("adding annotation: %w", err)
		}
	}

	for _, v := range narrated.Variations() {
		if err := merged.AddVariation(v); err != nil {
			return nil, fmt.Errorf("adding variation: %w", err)
		}
	}

	if err := merged.Validate(); err != nil {
		return nil, fmt.Errorf("validating merged story: %w", err)
	}

	return merged, nil
}

// resolveActorList builds a merged actor list. For each proposed actor, if a
// narrated actor matches by name (case-insensitive), the version with higher
// trust wins. Novel proposed actors are appended after all narrated actors.
func resolveActorList(narrated, proposed []StoryActor) []StoryActor {
	resolved := make([]StoryActor, len(narrated))
	copy(resolved, narrated)

	for _, pa := range proposed {
		na, found := findActorByName(resolved, pa.Name())
		if !found {
			resolved = append(resolved, pa)

			continue
		}

		// Higher trust wins: replace narrated with proposed if proposed has higher trust.
		if pa.Trust().IsHigherTrust(na.Trust()) {
			for i, a := range resolved {
				if strings.EqualFold(a.Name(), pa.Name()) {
					resolved[i] = pa

					break
				}
			}
		}
	}

	return resolved
}

// resolveWorkObjectList builds a merged work object list using the same
// higher-trust-wins strategy as resolveActorList.
func resolveWorkObjectList(narrated, proposed []WorkObject) []WorkObject {
	resolved := make([]WorkObject, len(narrated))
	copy(resolved, narrated)

	for _, pw := range proposed {
		nw, found := findWorkObjectByName(resolved, pw.Name())
		if !found {
			resolved = append(resolved, pw)

			continue
		}

		if pw.Trust().IsHigherTrust(nw.Trust()) {
			for i, w := range resolved {
				if strings.EqualFold(w.Name(), pw.Name()) {
					resolved[i] = pw

					break
				}
			}
		}
	}

	return resolved
}

func findActorByName(actors []StoryActor, name string) (StoryActor, bool) {
	for _, a := range actors {
		if strings.EqualFold(a.Name(), name) {
			return a, true
		}
	}

	return StoryActor{}, false
}

func findWorkObjectByName(objects []WorkObject, name string) (WorkObject, bool) {
	for _, wo := range objects {
		if strings.EqualFold(wo.Name(), name) {
			return wo, true
		}
	}

	return WorkObject{}, false
}
