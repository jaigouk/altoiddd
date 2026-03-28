package domain

import (
	"fmt"
	"sort"
	"strings"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// GlossaryExtractor extracts ubiquitous language entries from domain stories.
type GlossaryExtractor struct{}

// Extract walks actors, work objects, and sentence activities across all stories,
// deduplicates by case-insensitive term name, assigns bounded context from contextMap,
// and returns entries sorted lexicographically by term.
func (e GlossaryExtractor) Extract(
	stories []*DomainStory,
	contextMap *ContextMap,
) ([]vo.UbiquitousLanguageEntry, error) {
	entries := make(map[string]vo.UbiquitousLanguageEntry)

	for _, story := range stories {
		for _, actor := range story.Actors() {
			term := actor.Name()
			definition := fmt.Sprintf("%s in %s", actor.Type(), story.Title())
			e.addOrMerge(entries, term, definition, story.Title(), actor.Trust(), actor.Source(), contextMap)
		}

		for _, wo := range story.WorkObjects() {
			term := wo.Name()
			definition := fmt.Sprintf("%s in %s", wo.Type(), story.Title())
			e.addOrMerge(entries, term, definition, story.Title(), wo.Trust(), wo.Source(), contextMap)
		}

		for _, sentence := range story.Sentences() {
			activity := strings.TrimSpace(sentence.Activity())
			if activity == "" {
				continue
			}

			term := activity
			definition := fmt.Sprintf("Activity: %s %s %s", sentence.Subject(), sentence.Activity(), sentence.Object())
			e.addOrMerge(entries, term, definition, story.Title(), sentence.Trust(), sentence.Source(), contextMap)
		}
	}

	result := make([]vo.UbiquitousLanguageEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Term() < result[j].Term()
	})

	return result, nil
}

// addOrMerge deduplicates entries by case-insensitive term key. When a duplicate is
// found, it upgrades trust if the new trust is higher and merges story references.
func (e GlossaryExtractor) addOrMerge(
	entries map[string]vo.UbiquitousLanguageEntry,
	term, definition, storyTitle string,
	trust vo.TrustLevel,
	source string,
	contextMap *ContextMap,
) {
	key := strings.ToLower(term)

	existing, exists := entries[key]
	if exists {
		if trust.IsHigherTrust(existing.Trust()) {
			existing = existing.WithTrust(trust)
		}

		merged := appendStoryRef(existing.Stories(), storyTitle)
		entries[key] = existing.WithStories(merged)

		return
	}

	ctx := resolveContext(term, contextMap)

	entry, err := vo.NewUbiquitousLanguageEntry(term, definition, ctx, trust, source)
	if err != nil {
		return
	}

	entries[key] = entry.WithStories([]string{storyTitle})
}

// resolveContext determines the bounded context name for a term by searching
// the contextMap's sketches for matching actors or work objects.
func resolveContext(term string, contextMap *ContextMap) string {
	if contextMap == nil {
		return "General"
	}

	for _, sketch := range contextMap.Contexts() {
		for _, actor := range sketch.Actors() {
			if strings.EqualFold(term, actor) {
				return sketch.Name()
			}
		}

		for _, wo := range sketch.WorkObjects() {
			if strings.EqualFold(term, wo) {
				return sketch.Name()
			}
		}
	}

	return "General"
}

// appendStoryRef appends storyTitle to refs if not already present.
func appendStoryRef(refs []string, storyTitle string) []string {
	for _, r := range refs {
		if r == storyTitle {
			return refs
		}
	}

	return append(refs, storyTitle)
}
