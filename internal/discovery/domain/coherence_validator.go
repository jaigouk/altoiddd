package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// CoherenceSeverity classifies the severity of a coherence finding.
type CoherenceSeverity string

// CoherenceSeverity constants.
const (
	CoherenceSeverityWarning CoherenceSeverity = "warning"
)

// CoherenceFinding represents a single cross-story coherence issue.
type CoherenceFinding struct {
	severity    CoherenceSeverity
	location    string
	description string
}

// NewCoherenceFinding creates a CoherenceFinding, enforcing domain invariants.
func NewCoherenceFinding(severity CoherenceSeverity, location, description string) (CoherenceFinding, error) {
	if strings.TrimSpace(description) == "" {
		return CoherenceFinding{}, fmt.Errorf("coherence finding description must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	return CoherenceFinding{
		severity:    severity,
		location:    location,
		description: description,
	}, nil
}

// Severity returns the finding's severity level.
func (f CoherenceFinding) Severity() CoherenceSeverity { return f.severity }

// Location returns the finding's location context.
func (f CoherenceFinding) Location() string { return f.location }

// Description returns the finding's description.
func (f CoherenceFinding) Description() string { return f.description }

// CoherenceReport contains the results of cross-story coherence validation.
type CoherenceReport struct {
	findings []CoherenceFinding
}

// NewCoherenceReport creates a CoherenceReport from the given findings.
func NewCoherenceReport(findings []CoherenceFinding) CoherenceReport {
	stored := make([]CoherenceFinding, len(findings))
	copy(stored, findings)

	return CoherenceReport{findings: stored}
}

// Findings returns a defensive copy of the report's findings.
func (r CoherenceReport) Findings() []CoherenceFinding {
	out := make([]CoherenceFinding, len(r.findings))
	copy(out, r.findings)

	return out
}

// IsCoherent returns true when no findings were detected.
func (r CoherenceReport) IsCoherent() bool { return len(r.findings) == 0 }

// HasFindings returns true when at least one finding exists.
func (r CoherenceReport) HasFindings() bool { return len(r.findings) > 0 }

// mustCoherenceFinding creates a CoherenceFinding, panicking on construction error.
// Only used internally where descriptions are always non-empty string literals.
func mustCoherenceFinding(severity CoherenceSeverity, location, description string) CoherenceFinding {
	f, err := NewCoherenceFinding(severity, location, description)
	if err != nil {
		panic(fmt.Sprintf("BUG: invalid CoherenceFinding: %v", err))
	}

	return f
}

// storyPair is a map key for deduplicating findings between two stories.
type storyPair struct {
	a, b string
}

// termEntry tracks where a term name was seen and what type it had.
type termEntry struct {
	originalName string // preserves original casing
	typeName     string // e.g. "person", "document"
	dimension    string // "actor" or "work object"
	storyTitle   string
}

// CoherenceValidator detects cross-story contradictions in domain stories.
type CoherenceValidator struct{}

// Validate checks a set of domain stories for term type conflicts and
// undeclared cross-story references.
func (v CoherenceValidator) Validate(stories []*DomainStory) CoherenceReport {
	var findings []CoherenceFinding

	findings = append(findings, v.detectTermConflicts(stories)...)
	findings = append(findings, v.detectUndeclaredRefs(stories)...)

	return NewCoherenceReport(findings)
}

// detectTermConflicts finds terms that appear with different types across stories.
func (v CoherenceValidator) detectTermConflicts(stories []*DomainStory) []CoherenceFinding {
	// Map: lowercased name -> list of term entries
	seen := make(map[string][]termEntry)

	for _, story := range stories {
		for _, actor := range story.Actors() {
			key := strings.ToLower(actor.Name())
			seen[key] = append(seen[key], termEntry{
				originalName: actor.Name(),
				typeName:     string(actor.Type()),
				dimension:    "actor",
				storyTitle:   story.Title(),
			})
		}

		for _, wo := range story.WorkObjects() {
			key := strings.ToLower(wo.Name())
			seen[key] = append(seen[key], termEntry{
				originalName: wo.Name(),
				typeName:     string(wo.Type()),
				dimension:    "work object",
				storyTitle:   story.Title(),
			})
		}
	}

	var findings []CoherenceFinding

	for _, entries := range seen {
		if len(entries) < 2 {
			continue
		}

		reported := make(map[storyPair]struct{})

		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				a, b := entries[i], entries[j]

				// Deduplicate: one finding per unique pair of (storyTitle, storyTitle).
				key := storyPair{a.storyTitle, b.storyTitle}
				if _, ok := reported[key]; ok {
					continue
				}

				if a.dimension != b.dimension {
					findings = append(findings, mustCoherenceFinding(
						CoherenceSeverityWarning,
						fmt.Sprintf("stories: %s, %s", a.storyTitle, b.storyTitle),
						fmt.Sprintf("term %q is used as %s in %q and as %s in %q",
							a.originalName, a.dimension, a.storyTitle, b.dimension, b.storyTitle),
					))

					reported[key] = struct{}{}
				} else if a.typeName != b.typeName {
					findings = append(findings, mustCoherenceFinding(
						CoherenceSeverityWarning,
						fmt.Sprintf("stories: %s, %s", a.storyTitle, b.storyTitle),
						fmt.Sprintf("term %q has conflicting types: %s in %q vs %s in %q",
							a.originalName, a.typeName, a.storyTitle, b.typeName, b.storyTitle),
					))

					reported[key] = struct{}{}
				}
			}
		}
	}

	return findings
}

// detectUndeclaredRefs finds sentence references that don't appear in any story's
// actor or work object declarations.
func (v CoherenceValidator) detectUndeclaredRefs(stories []*DomainStory) []CoherenceFinding {
	// Build global sets of all declared actor and work object names.
	globalActors := make(map[string]struct{})
	globalWorkObjects := make(map[string]struct{})

	for _, story := range stories {
		for _, actor := range story.Actors() {
			globalActors[strings.ToLower(actor.Name())] = struct{}{}
		}

		for _, wo := range story.WorkObjects() {
			globalWorkObjects[strings.ToLower(wo.Name())] = struct{}{}
		}
	}

	var findings []CoherenceFinding

	for _, story := range stories {
		for _, sentence := range story.Sentences() {
			location := fmt.Sprintf("story: %s, step: %d", story.Title(), sentence.Step())

			// Check subject is in global actor set.
			subjectKey := strings.ToLower(sentence.Subject())
			if _, ok := globalActors[subjectKey]; !ok {
				findings = append(findings, mustCoherenceFinding(
					CoherenceSeverityWarning,
					location,
					fmt.Sprintf("subject %q is not declared as an actor in any story", sentence.Subject()),
				))
			}

			// Check object is in global actor set OR global work object set.
			objectKey := strings.ToLower(sentence.Object())
			_, inActors := globalActors[objectKey]
			_, inWorkObjects := globalWorkObjects[objectKey]

			if !inActors && !inWorkObjects {
				findings = append(findings, mustCoherenceFinding(
					CoherenceSeverityWarning,
					location,
					fmt.Sprintf("object %q is not declared as an actor or work object in any story", sentence.Object()),
				))
			}

			// Check indirect object if present.
			if sentence.HasIndirectObject() {
				indirectKey := strings.ToLower(sentence.IndirectObject())
				_, inActors := globalActors[indirectKey]
				_, inWorkObjects := globalWorkObjects[indirectKey]

				if !inActors && !inWorkObjects {
					findings = append(findings, mustCoherenceFinding(
						CoherenceSeverityWarning,
						location,
						fmt.Sprintf("indirect object %q is not declared as an actor or work object in any story",
							sentence.IndirectObject()),
					))
				}
			}
		}
	}

	return findings
}
