package domain

import (
	"fmt"
	"maps"
	"strings"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// TrustDistribution counts actors, work objects, and sentences by trust level.
// Annotations are excluded — they are metadata attached to sentences, not
// independent domain elements.
type TrustDistribution struct {
	counts map[vo.TrustLevel]int
}

// NewTrustDistribution creates an empty TrustDistribution with an initialized map.
func NewTrustDistribution() TrustDistribution {
	return TrustDistribution{counts: make(map[vo.TrustLevel]int)}
}

// AddStory returns a new TrustDistribution with counts from the story's actors,
// work objects, and sentences accumulated. Copy-on-write: the receiver is unchanged.
func (d TrustDistribution) AddStory(story *DomainStory) TrustDistribution {
	newCounts := make(map[vo.TrustLevel]int, len(d.counts))
	maps.Copy(newCounts, d.counts)

	for _, a := range story.Actors() {
		newCounts[a.Trust()]++
	}

	for _, wo := range story.WorkObjects() {
		newCounts[wo.Trust()]++
	}

	for _, s := range story.Sentences() {
		newCounts[s.Trust()]++
	}

	return TrustDistribution{counts: newCounts}
}

// AddSketch returns a new TrustDistribution with the sketch's trust level counted.
// Copy-on-write: the receiver is unchanged.
func (d TrustDistribution) AddSketch(sketch BoundedContextSketch) TrustDistribution {
	newCounts := make(map[vo.TrustLevel]int, len(d.counts))
	maps.Copy(newCounts, d.counts)

	newCounts[sketch.Trust()]++

	return TrustDistribution{counts: newCounts}
}

// Count returns the count for the given trust level.
func (d TrustDistribution) Count(level vo.TrustLevel) int {
	return d.counts[level]
}

// Total returns the sum of all trust level counts.
func (d TrustDistribution) Total() int {
	total := 0
	for _, v := range d.counts {
		total += v
	}

	return total
}

// String returns a human-readable representation using AllTrustLevels() for consistent ordering.
func (d TrustDistribution) String() string {
	var parts []string
	for _, level := range vo.AllTrustLevels() {
		parts = append(parts, fmt.Sprintf("%s:%d", level, d.counts[level]))
	}

	return strings.Join(parts, " ")
}

// StorySummary holds a summary of a single domain story including its trust distribution.
type StorySummary struct {
	Title         string
	ActorCount    int
	SentenceCount int
	Distribution  TrustDistribution
}

// SummarizeStory computes a StorySummary from a DomainStory.
func SummarizeStory(story *DomainStory) StorySummary {
	dist := NewTrustDistribution().AddStory(story)

	return StorySummary{
		Title:         story.Title(),
		ActorCount:    len(story.Actors()),
		SentenceCount: len(story.Sentences()),
		Distribution:  dist,
	}
}
