// Throwaway prototype: boundary detection heuristics validation.
// This is a RESEARCH artifact, not production code.
// It parses alto text-format domain stories and scans for boundary signals.
package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// --- Domain types ---

// Sentence represents one numbered activity in a domain story.
type Sentence struct {
	Number  int
	Subject string
	Verb    string
	Object  string
	Raw     string
}

// Story represents a parsed domain story.
type Story struct {
	Title       string
	Filename    string
	Actors      []string
	WorkObjects []string
	Sentences   []Sentence
	Annotations []string
}

// BoundarySignal records evidence of a bounded context boundary.
type BoundarySignal struct {
	SignalType  string // one_way_flow, language_difference, different_trigger, org_boundary, same_object_diff_context
	Description string
	StoryRefs   []string
	Confidence  float64 // 0.0 - 1.0
}

// BoundedContextCandidate is a proposed bounded context.
type BoundedContextCandidate struct {
	Name        string
	Actors      []string
	WorkObjects []string
	Signals     []BoundarySignal
	Score       float64
}

// --- Parsing ---

func parseStory(path string) (*Story, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening story file: %w", err)
	}
	defer func() { _ = f.Close() }()

	story := &Story{Filename: filepath.Base(path)}
	scanner := bufio.NewScanner(f)

	sentenceRe := regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
	actorsRe := regexp.MustCompile(`(?i)^Actors:\s*(.+)$`)
	objectsRe := regexp.MustCompile(`(?i)^Work Objects:\s*(.+)$`)
	titleRe := regexp.MustCompile(`^Domain Story:\s*"(.+)"$`)
	annotationRe := regexp.MustCompile(`^\s+\[.+\]\s+(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		if m := titleRe.FindStringSubmatch(line); m != nil {
			story.Title = m[1]
			continue
		}
		if m := actorsRe.FindStringSubmatch(line); m != nil {
			story.Actors = splitAndTrim(m[1])
			continue
		}
		if m := objectsRe.FindStringSubmatch(line); m != nil {
			story.WorkObjects = splitAndTrim(m[1])
			continue
		}
		if m := sentenceRe.FindStringSubmatch(line); m != nil {
			s := parseSentence(m[1], m[2])
			story.Sentences = append(story.Sentences, s)
			continue
		}
		if m := annotationRe.FindStringSubmatch(line); m != nil {
			story.Annotations = append(story.Annotations, m[1])
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning story file: %w", err)
	}
	return story, nil
}

func parseSentence(numStr, text string) Sentence {
	num := 0
	_, _ = fmt.Sscanf(numStr, "%d", &num)

	// Simple heuristic: first word is subject, second is verb, rest is object-ish
	parts := strings.Fields(text)
	s := Sentence{Number: num, Raw: text}

	if len(parts) >= 1 {
		// Subject may be multi-word (e.g., "Pet Owner", "Payment Provider")
		// Heuristic: capitalize-check for compound subjects
		s.Subject = extractSubject(parts)
	}
	if len(parts) >= 2 {
		verbIdx := countSubjectWords(parts)
		if verbIdx < len(parts) {
			s.Verb = strings.ToLower(parts[verbIdx])
		}
		if verbIdx+1 < len(parts) {
			s.Object = strings.Join(parts[verbIdx+1:], " ")
		}
	}
	return s
}

// extractSubject handles multi-word proper-noun subjects.
func extractSubject(parts []string) string {
	n := countSubjectWords(parts)
	return strings.Join(parts[:n], " ")
}

// countSubjectWords counts consecutive capitalized words at the start.
func countSubjectWords(parts []string) int {
	count := 0
	for i, p := range parts {
		if i == 0 {
			count++
			continue
		}
		// If it starts uppercase and is NOT a common verb, it's part of the subject
		if len(p) > 0 && p[0] >= 'A' && p[0] <= 'Z' && !isCommonVerb(strings.ToLower(p)) {
			count++
		} else {
			break
		}
	}
	return count
}

func isCommonVerb(w string) bool {
	verbs := map[string]bool{
		"browses": true, "selects": true, "adds": true, "proceeds": true,
		"chooses": true, "processes": true, "calculates": true, "receives": true,
		"prepares": true, "picks": true, "delivers": true, "calls": true,
		"checks": true, "creates": true, "arrives": true, "examines": true,
		"records": true, "prescribes": true, "generates": true, "pays": true,
		"submits": true, "declines": true, "displays": true, "offers": true,
		"confirms": true, "logs": true, "reviews": true, "updates": true,
		"adjusts": true, "recalculates": true, "publishes": true, "sees": true,
		"registers": true, "performs": true, "assigns": true, "administers": true,
		"runs": true, "reads": true, "detects": true, "asks": true,
		"proposes": true, "refines": true, "identifies": true, "classifies": true,
		"applies": true, "constructs": true, "presents": true, "sets": true,
		"writes": true, "agrees": true, "searches": true, "provides": true,
		"marks": true, "removes": true, "views": true, "sends": true,
		"notifies": true, "hands": true, "renews": true, "dispenses": true,
		"evaluates": true, "schedules": true,
	}
	return verbs[w]
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

// --- Signal Detection ---

// detectOneWayFlow checks if work objects flow in only one direction between
// actor groups. If actor A sends object X to actor B but B never sends X
// (or anything) back to A, that is a one-way flow signal.
func detectOneWayFlow(stories []*Story) []BoundarySignal {
	// Build a flow graph: (from_actor, to_actor, object) -> count
	type flowKey struct {
		from, to, obj string
	}
	flows := make(map[flowKey]int)

	for _, story := range stories {
		for i := 0; i < len(story.Sentences)-1; i++ {
			curr := story.Sentences[i]
			next := story.Sentences[i+1]
			if curr.Subject != next.Subject && curr.Subject != "" && next.Subject != "" {
				// There is a handoff: curr.Subject did something, next.Subject does something
				// The object being handed off is curr.Object (simplified)
				key := flowKey{from: curr.Subject, to: next.Subject, obj: simplifyObject(curr.Object)}
				flows[key]++
			}
		}
	}

	// Check for one-way flows: A->B exists but B->A does not
	var signals []BoundarySignal
	checked := make(map[string]bool)

	for key, count := range flows {
		pairKey := key.from + "|||" + key.to
		reversePairKey := key.to + "|||" + key.from
		if checked[pairKey] || checked[reversePairKey] {
			continue
		}
		checked[pairKey] = true

		// Check if reverse exists
		hasReverse := false
		for rk := range flows {
			if rk.from == key.to && rk.to == key.from {
				hasReverse = true
				break
			}
		}

		if !hasReverse && count > 0 {
			refs := findStoriesWithActors(stories, key.from, key.to)
			signals = append(signals, BoundarySignal{
				SignalType:  "one_way_flow",
				Description: fmt.Sprintf("%s -> %s (never reverse). Object: %s", key.from, key.to, key.obj),
				StoryRefs:   refs,
				Confidence:  math.Min(0.3+float64(count)*0.2, 0.9),
			})
		}
	}
	return signals
}

func simplifyObject(obj string) string {
	// Remove prepositions and indirect objects for comparison
	for _, prep := range []string{" for ", " with ", " using ", " to ", " from ", " in ", " based on ", " into "} {
		if idx := strings.Index(strings.ToLower(obj), prep); idx > 0 {
			return strings.TrimSpace(obj[:idx])
		}
	}
	return strings.TrimSpace(obj)
}

// detectLanguageDifferences checks if the same work object name appears in
// multiple stories but in different verb/property contexts, suggesting
// different meanings.
func detectLanguageDifferences(stories []*Story) []BoundarySignal {
	// Map work object -> set of (verb, story) pairs
	type usage struct {
		verb      string
		storyFile string
	}
	objectUsage := make(map[string][]usage)

	for _, story := range stories {
		for _, s := range story.Sentences {
			obj := simplifyObject(s.Object)
			obj = normalizeObjectName(obj)
			if obj == "" {
				continue
			}
			objectUsage[obj] = append(objectUsage[obj], usage{
				verb:      s.Verb,
				storyFile: story.Filename,
			})
		}
	}

	var signals []BoundarySignal
	for obj, usages := range objectUsage {
		// Count unique stories
		storySet := make(map[string]bool)
		verbSet := make(map[string]bool)
		for _, u := range usages {
			storySet[u.storyFile] = true
			verbSet[u.verb] = true
		}
		if len(storySet) < 2 {
			continue
		}
		// If the object appears in multiple stories with different verbs,
		// that is a signal (different contexts treat it differently)
		if len(verbSet) >= 2 {
			refs := make([]string, 0, len(storySet))
			for s := range storySet {
				refs = append(refs, s)
			}
			sort.Strings(refs)
			verbs := make([]string, 0, len(verbSet))
			for v := range verbSet {
				verbs = append(verbs, v)
			}
			sort.Strings(verbs)

			signals = append(signals, BoundarySignal{
				SignalType:  "same_object_diff_context",
				Description: fmt.Sprintf("'%s' used with verbs [%s] across %d stories", obj, strings.Join(verbs, ", "), len(storySet)),
				StoryRefs:   refs,
				Confidence:  math.Min(0.2+float64(len(verbSet))*0.15+float64(len(storySet))*0.1, 0.85),
			})
		}
	}
	return signals
}

func normalizeObjectName(obj string) string {
	obj = strings.TrimSpace(obj)
	// Remove articles
	for _, prefix := range []string{"the ", "a ", "an "} {
		obj = strings.TrimPrefix(strings.ToLower(obj), prefix)
	}
	// Title case the result back
	if obj != "" {
		return strings.ToUpper(obj[:1]) + obj[1:]
	}
	return ""
}

// detectOrgBoundaries checks for actor groups that never co-appear in the
// same story. Non-overlapping actor groups suggest organizational boundaries.
func detectOrgBoundaries(stories []*Story) []BoundarySignal {
	// Build actor -> stories map
	actorStories := make(map[string]map[string]bool)
	for _, story := range stories {
		for _, actor := range story.Actors {
			if actorStories[actor] == nil {
				actorStories[actor] = make(map[string]bool)
			}
			actorStories[actor][story.Filename] = true
		}
	}

	// Also check subjects in sentences (more precise than declared actors)
	actorSentenceSubjects := make(map[string]map[string]bool) // actor -> set of story filenames where they act
	for _, story := range stories {
		for _, s := range story.Sentences {
			subj := s.Subject
			if subj == "" {
				continue
			}
			if actorSentenceSubjects[subj] == nil {
				actorSentenceSubjects[subj] = make(map[string]bool)
			}
			actorSentenceSubjects[subj][story.Filename] = true
		}
	}

	// Find actor pairs that never co-appear
	var signals []BoundarySignal
	actors := make([]string, 0, len(actorSentenceSubjects))
	for a := range actorSentenceSubjects {
		actors = append(actors, a)
	}
	sort.Strings(actors)

	checked := make(map[string]bool)
	for i, a1 := range actors {
		for _, a2 := range actors[i+1:] {
			pairKey := a1 + "|||" + a2
			if checked[pairKey] {
				continue
			}
			checked[pairKey] = true

			// Check overlap
			overlap := false
			for s := range actorSentenceSubjects[a1] {
				if actorSentenceSubjects[a2][s] {
					overlap = true
					break
				}
			}

			if !overlap && len(actorSentenceSubjects[a1]) > 0 && len(actorSentenceSubjects[a2]) > 0 {
				refs := make([]string, 0)
				for s := range actorSentenceSubjects[a1] {
					refs = append(refs, s)
				}
				for s := range actorSentenceSubjects[a2] {
					refs = append(refs, s)
				}
				sort.Strings(refs)

				signals = append(signals, BoundarySignal{
					SignalType:  "org_boundary",
					Description: fmt.Sprintf("'%s' and '%s' never appear in the same story (non-overlapping actor groups)", a1, a2),
					StoryRefs:   unique(refs),
					Confidence:  0.4, // Low base — could be coincidence with few stories
				})
			}
		}
	}
	return signals
}

// detectDifferentTriggers examines whether different stories have fundamentally
// different trigger patterns (time-based vs event-based vs on-demand).
func detectDifferentTriggers(stories []*Story) []BoundarySignal {
	// This is hard to automate purely from text. We use heuristic keyword matching.
	type triggerClass int
	const (
		triggerEvent  triggerClass = iota // "arrives", "calls", "submits", user-initiated
		triggerTime                       // "scheduled", "daily", "weekly", time-words
		triggerSystem                     // "generates", "calculates", system-initiated
		triggerUnknown
	)

	classifyFirstSentence := func(s Sentence) triggerClass {
		raw := strings.ToLower(s.Raw)
		// Time-based triggers
		for _, w := range []string{"scheduled", "daily", "weekly", "monthly", "every", "batch", "cron", "timer"} {
			if strings.Contains(raw, w) {
				return triggerTime
			}
		}
		// System-initiated
		for _, w := range []string{"system", "platform", "cli", "service", "automatically"} {
			if strings.Contains(raw, w) {
				return triggerSystem
			}
		}
		return triggerEvent
	}

	// Group stories by trigger class
	classStories := make(map[triggerClass][]string)
	for _, story := range stories {
		if len(story.Sentences) > 0 {
			cls := classifyFirstSentence(story.Sentences[0])
			classStories[cls] = append(classStories[cls], story.Filename)
		}
	}

	var signals []BoundarySignal
	classNames := map[triggerClass]string{
		triggerEvent:  "event/user-initiated",
		triggerTime:   "time-based/scheduled",
		triggerSystem: "system-initiated",
	}

	if len(classStories) >= 2 {
		var parts []string
		var allRefs []string
		for cls, files := range classStories {
			if name, ok := classNames[cls]; ok {
				parts = append(parts, fmt.Sprintf("%s: %s", name, strings.Join(files, ", ")))
			}
			allRefs = append(allRefs, files...)
		}
		sort.Strings(parts)
		sort.Strings(allRefs)

		signals = append(signals, BoundarySignal{
			SignalType:  "different_trigger",
			Description: fmt.Sprintf("Stories have %d trigger types: %s", len(classStories), strings.Join(parts, "; ")),
			StoryRefs:   unique(allRefs),
			Confidence:  math.Min(0.3+float64(len(classStories))*0.15, 0.7),
		})
	}
	return signals
}

// --- Clustering into bounded context candidates ---

// clusterContexts proposes bounded context candidates from detected signals.
func clusterContexts(stories []*Story, signals []BoundarySignal) []BoundedContextCandidate {
	// Build actor clusters: group actors that frequently co-appear
	actorCoOccurrence := make(map[string]map[string]int) // actor1 -> actor2 -> count
	for _, story := range stories {
		subjects := extractUniqueSubjects(story)
		for i, a1 := range subjects {
			for _, a2 := range subjects[i+1:] {
				if actorCoOccurrence[a1] == nil {
					actorCoOccurrence[a1] = make(map[string]int)
				}
				if actorCoOccurrence[a2] == nil {
					actorCoOccurrence[a2] = make(map[string]int)
				}
				actorCoOccurrence[a1][a2]++
				actorCoOccurrence[a2][a1]++
			}
		}
	}

	// Build object-to-actor affinity: which actors touch which objects most
	objectActorAffinity := make(map[string]map[string]int)
	for _, story := range stories {
		for _, s := range story.Sentences {
			obj := normalizeObjectName(simplifyObject(s.Object))
			if obj == "" || s.Subject == "" {
				continue
			}
			if objectActorAffinity[obj] == nil {
				objectActorAffinity[obj] = make(map[string]int)
			}
			objectActorAffinity[obj][s.Subject]++
		}
	}

	// Simple greedy clustering: group actors that co-appear most often
	// and assign objects by affinity
	allActors := make(map[string]bool)
	for _, story := range stories {
		for _, s := range story.Sentences {
			if s.Subject != "" {
				allActors[s.Subject] = true
			}
		}
	}

	assigned := make(map[string]bool)
	var candidates []BoundedContextCandidate

	// Sort actors by number of stories they appear in (most active first)
	type actorCount struct {
		name  string
		count int
	}
	var actorList []actorCount
	for a := range allActors {
		count := 0
		for _, story := range stories {
			for _, s := range story.Sentences {
				if s.Subject == a {
					count++
					break
				}
			}
		}
		actorList = append(actorList, actorCount{a, count})
	}
	sort.Slice(actorList, func(i, j int) bool {
		return actorList[i].count > actorList[j].count
	})

	for _, ac := range actorList {
		if assigned[ac.name] {
			continue
		}
		cluster := []string{ac.name}
		assigned[ac.name] = true

		// Add co-occurring actors not yet assigned
		if coMap, ok := actorCoOccurrence[ac.name]; ok {
			for coActor, count := range coMap {
				if !assigned[coActor] && count >= 1 {
					// Check if this co-actor is separated by a boundary signal
					separated := false
					for _, sig := range signals {
						if sig.SignalType == "one_way_flow" || sig.SignalType == "org_boundary" {
							if strings.Contains(sig.Description, coActor) && strings.Contains(sig.Description, ac.name) {
								separated = true
								break
							}
						}
					}
					if !separated {
						cluster = append(cluster, coActor)
						assigned[coActor] = true
					}
				}
			}
		}

		// Assign objects by affinity
		var objects []string
		for obj, actorMap := range objectActorAffinity {
			bestActor := ""
			bestCount := 0
			for a, c := range actorMap {
				if c > bestCount {
					bestActor = a
					bestCount = c
				}
			}
			for _, clusterActor := range cluster {
				if bestActor == clusterActor {
					objects = append(objects, obj)
					break
				}
			}
		}

		// Collect signals relevant to this cluster
		var relevantSignals []BoundarySignal
		for _, sig := range signals {
			for _, actor := range cluster {
				if strings.Contains(sig.Description, actor) {
					relevantSignals = append(relevantSignals, sig)
					break
				}
			}
		}

		// Calculate score
		score := 0.0
		for _, sig := range relevantSignals {
			score += sig.Confidence
		}
		if len(relevantSignals) > 0 {
			score /= float64(len(relevantSignals))
		}

		sort.Strings(cluster)
		sort.Strings(objects)

		candidates = append(candidates, BoundedContextCandidate{
			Name:        suggestContextName(cluster, objects),
			Actors:      cluster,
			WorkObjects: objects,
			Signals:     relevantSignals,
			Score:       score,
		})
	}

	return candidates
}

func extractUniqueSubjects(story *Story) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range story.Sentences {
		if s.Subject != "" && !seen[s.Subject] {
			seen[s.Subject] = true
			result = append(result, s.Subject)
		}
	}
	return result
}

func suggestContextName(actors []string, objects []string) string {
	// Simple heuristic: use primary actor or primary object as context name
	if len(objects) > 0 {
		return objects[0] + " Management"
	}
	if len(actors) > 0 {
		return actors[0] + " Operations"
	}
	return "Unknown Context"
}

// --- Utilities ---

func findStoriesWithActors(stories []*Story, a1, a2 string) []string {
	var result []string
	for _, story := range stories {
		hasA1, hasA2 := false, false
		for _, s := range story.Sentences {
			if s.Subject == a1 {
				hasA1 = true
			}
			if s.Subject == a2 {
				hasA2 = true
			}
		}
		if hasA1 || hasA2 {
			result = append(result, story.Filename)
		}
	}
	return unique(result)
}

func unique(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// --- Output ---

func printStories(stories []*Story) {
	fmt.Println("=== PARSED STORIES ===")
	for _, story := range stories {
		fmt.Printf("\n--- %s ---\n", story.Title)
		fmt.Printf("  File: %s\n", story.Filename)
		fmt.Printf("  Actors: %s\n", strings.Join(story.Actors, ", "))
		fmt.Printf("  Work Objects: %s\n", strings.Join(story.WorkObjects, ", "))
		fmt.Printf("  Sentences: %d\n", len(story.Sentences))
		for _, s := range story.Sentences {
			fmt.Printf("    %d. [%s] --%s--> [%s]\n", s.Number, s.Subject, s.Verb, s.Object)
		}
	}
}

func printSignals(signals []BoundarySignal) {
	fmt.Println("\n=== BOUNDARY SIGNALS ===")
	for _, sig := range signals {
		fmt.Printf("  [%s] (confidence: %.2f)\n", sig.SignalType, sig.Confidence)
		fmt.Printf("    %s\n", sig.Description)
		fmt.Printf("    Stories: %s\n", strings.Join(sig.StoryRefs, ", "))
	}
}

func printContexts(contexts []BoundedContextCandidate) {
	fmt.Println("\n=== BOUNDED CONTEXT CANDIDATES ===")
	for _, ctx := range contexts {
		fmt.Printf("\n  Context: %s (score: %.2f)\n", ctx.Name, ctx.Score)
		fmt.Printf("    Actors: %s\n", strings.Join(ctx.Actors, ", "))
		fmt.Printf("    Work Objects: %s\n", strings.Join(ctx.WorkObjects, ", "))
		fmt.Printf("    Supporting signals: %d\n", len(ctx.Signals))
		for _, sig := range ctx.Signals {
			fmt.Printf("      - [%s] %.2f: %s\n", sig.SignalType, sig.Confidence, sig.Description)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <story-dir-or-files...>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Analyzes domain stories for bounded context boundary signals.\n")
		fmt.Fprintf(os.Stderr, "  Pass a directory to scan all .story.txt files, or individual files.\n")
		os.Exit(1)
	}

	var paths []string
	for _, arg := range os.Args[1:] {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s: %v\n", arg, err)
			os.Exit(1)
		}
		if info.IsDir() {
			matches, _ := filepath.Glob(filepath.Join(arg, "*.story.txt"))
			paths = append(paths, matches...)
		} else {
			paths = append(paths, arg)
		}
	}

	if len(paths) == 0 {
		fmt.Fprintf(os.Stderr, "No story files found.\n")
		os.Exit(1)
	}

	sort.Strings(paths)

	var stories []*Story
	for _, p := range paths {
		story, err := parseStory(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", p, err)
			continue
		}
		stories = append(stories, story)
	}

	if len(stories) == 0 {
		fmt.Fprintf(os.Stderr, "No stories parsed successfully.\n")
		os.Exit(1)
	}

	fmt.Printf("Analyzing %d stories...\n", len(stories))
	printStories(stories)

	// Run all signal detectors
	var allSignals []BoundarySignal
	allSignals = append(allSignals, detectOneWayFlow(stories)...)
	allSignals = append(allSignals, detectLanguageDifferences(stories)...)
	allSignals = append(allSignals, detectOrgBoundaries(stories)...)
	allSignals = append(allSignals, detectDifferentTriggers(stories)...)

	printSignals(allSignals)

	// Cluster into context candidates
	contexts := clusterContexts(stories, allSignals)
	printContexts(contexts)

	// Summary stats
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Stories analyzed: %d\n", len(stories))
	fmt.Printf("Signals detected: %d\n", len(allSignals))
	fmt.Printf("Context candidates: %d\n", len(contexts))
	signalCounts := make(map[string]int)
	for _, s := range allSignals {
		signalCounts[s.SignalType]++
	}
	fmt.Println("Signal type breakdown:")
	for st, count := range signalCounts {
		fmt.Printf("  %s: %d\n", st, count)
	}
}
