package infrastructure

import (
	"context"
	"fmt"
	"sort"
	"strings"

	application "github.com/alto-cli/alto/internal/discovery/application"
	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// AlgorithmicDetector detects bounded context boundaries using structural
// heuristics over domain stories. It implements the BoundaryDetector port
// for the algorithmic (non-LLM) detection path.
type AlgorithmicDetector struct{}

// Compile-time check that AlgorithmicDetector satisfies BoundaryDetector.
var _ application.BoundaryDetector = (*AlgorithmicDetector)(nil)

// NewAlgorithmicDetector creates an AlgorithmicDetector.
func NewAlgorithmicDetector() *AlgorithmicDetector {
	return &AlgorithmicDetector{}
}

// notificationVerbs is the set of verbs that receive a confidence discount
// in one-way flow detection.
var notificationVerbs = map[string]struct{}{
	"displays": {},
	"notifies": {},
	"shows":    {},
	"presents": {},
}

// DetectBoundaries analyzes domain stories for bounded context boundary signals
// using three structural heuristics, clusters the results, scores them, and
// returns bounded context sketches.
//
// The mode parameter is accepted for interface compatibility but ignored;
// mode filtering is handled by HybridBoundaryDetector in P3-3.
func (d *AlgorithmicDetector) DetectBoundaries(
	_ context.Context,
	stories []*domain.DomainStory,
	_ domain.DiscoveryMode,
) ([]domain.BoundedContextSketch, error) {
	if len(stories) < 2 {
		return nil, fmt.Errorf("boundary detection requires at least 2 stories: %w", domainerrors.ErrInvariantViolation)
	}

	for i, s := range stories {
		if s == nil {
			return nil, fmt.Errorf("nil story at index %d: %w", i, domainerrors.ErrInvariantViolation)
		}
	}

	// Collect signals from three detectors.
	var allSignals []signalWithContext
	allSignals = append(allSignals, d.detectSameObjectDiffContext(stories)...)
	allSignals = append(allSignals, d.detectOneWayFlow(stories)...)
	allSignals = append(allSignals, d.detectOrgBoundary(stories)...)

	if len(allSignals) == 0 {
		return d.detectWorkObjectClusters(stories)
	}

	return d.clusterAndScore(allSignals)
}

// signalWithContext wraps a BoundarySignal with the actors, work objects, and
// story titles that contributed to it, enabling clustering.
type signalWithContext struct {
	signal      domain.BoundarySignal
	actors      []string
	workObjects []string
	stories     []string
	confidence  float64
}

// storyActorWO records that an actor acted on a work object in a specific story.
type storyActorWO struct {
	storyTitle string
	actor      string
	workObject string
}

// --- Signal Detector: Same Object, Different Context ---

func (d *AlgorithmicDetector) detectSameObjectDiffContext(stories []*domain.DomainStory) []signalWithContext {
	// objectRecord tracks (workObjectName -> {activity -> set of storyTitles}).
	type activityStories struct {
		activities map[string]map[string]struct{} // activity -> storyTitles
	}

	objectMap := make(map[string]*activityStories)

	for _, story := range stories {
		woNames := buildWorkObjectNameSet(story)

		for _, sentence := range story.Sentences() {
			objName := strings.ToLower(sentence.Object())
			if _, isWO := woNames[objName]; !isWO {
				continue
			}

			activity := strings.ToLower(sentence.Activity())
			storyTitle := story.Title()

			rec, ok := objectMap[objName]
			if !ok {
				rec = &activityStories{activities: make(map[string]map[string]struct{})}
				objectMap[objName] = rec
			}

			if rec.activities[activity] == nil {
				rec.activities[activity] = make(map[string]struct{})
			}

			rec.activities[activity][storyTitle] = struct{}{}
		}
	}

	var signals []signalWithContext

	for objName, rec := range objectMap {
		// Count distinct activities and distinct stories.
		distinctActivities := len(rec.activities)

		storySet := make(map[string]struct{})

		var activityNames []string

		for act, storyMap := range rec.activities {
			activityNames = append(activityNames, act)

			for st := range storyMap {
				storySet[st] = struct{}{}
			}
		}

		distinctStories := len(storySet)

		if distinctStories >= 2 && distinctActivities >= 2 {
			desc := fmt.Sprintf("work object '%s' used with activities [%s] across %d stories",
				objName, strings.Join(activityNames, ", "), distinctStories)

			signal, err := domain.NewBoundarySignal(domain.SignalTypeSameObjectDiffContext, desc)
			if err != nil {
				continue
			}

			storyTitles := setToSlice(storySet)

			signals = append(signals, signalWithContext{
				signal:      signal,
				workObjects: []string{objName},
				stories:     storyTitles,
				confidence:  domain.BaseConfidenceSameObjectDiffContext,
			})
		}
	}

	return signals
}

// --- Signal Detector: One-Way Flow ---

func (d *AlgorithmicDetector) detectOneWayFlow(stories []*domain.DomainStory) []signalWithContext {
	// Build directed edge graph: subject -> object (when object is an actor).
	type edgeInfo struct {
		activities []string
		stories    map[string]struct{}
	}

	type edgeKey struct {
		from string
		to   string
	}

	edges := make(map[edgeKey]*edgeInfo)

	// Also track per-story: which actors act on which work objects (for sequential flow discount).
	var actorWorkObjectPairs []storyActorWO

	for _, story := range stories {
		actorNames := buildActorNameSet(story)
		woNames := buildWorkObjectNameSet(story)

		for _, sentence := range story.Sentences() {
			subj := strings.ToLower(sentence.Subject())
			obj := strings.ToLower(sentence.Object())

			// Record actor-workObject pairs for sequential flow discount.
			if _, isWO := woNames[obj]; isWO {
				actorWorkObjectPairs = append(actorWorkObjectPairs, storyActorWO{
					storyTitle: story.Title(),
					actor:      subj,
					workObject: obj,
				})
			}

			// Check if object is an actor.
			if _, isActor := actorNames[obj]; isActor && subj != obj {
				key := edgeKey{from: subj, to: obj}

				info, ok := edges[key]
				if !ok {
					info = &edgeInfo{stories: make(map[string]struct{})}
					edges[key] = info
				}

				info.activities = append(info.activities, strings.ToLower(sentence.Activity()))
				info.stories[story.Title()] = struct{}{}
			}

			// Check indirect object too.
			if sentence.HasIndirectObject() {
				indObj := strings.ToLower(sentence.IndirectObject())
				if _, isActor := actorNames[indObj]; isActor && subj != indObj {
					key := edgeKey{from: subj, to: indObj}

					info, ok := edges[key]
					if !ok {
						info = &edgeInfo{stories: make(map[string]struct{})}
						edges[key] = info
					}

					info.activities = append(info.activities, strings.ToLower(sentence.Activity()))
					info.stories[story.Title()] = struct{}{}
				}
			}
		}
	}

	// Find asymmetric edges: A->B exists but B->A does not.
	var signals []signalWithContext

	for key, info := range edges {
		reverseKey := edgeKey{from: key.to, to: key.from}
		if _, exists := edges[reverseKey]; exists {
			continue
		}

		confidence := domain.BaseConfidenceOneWayFlow

		// Notification verb discount: if ALL activities are notification verbs.
		allNotification := true

		for _, act := range info.activities {
			if _, isNotif := notificationVerbs[act]; !isNotif {
				allNotification = false

				break
			}
		}

		if allNotification {
			confidence -= domain.NotificationVerbDiscount
		}

		// Sequential flow discount: if from and to co-appear in the SAME story
		// and BOTH act on the same work object.
		if d.hasSequentialFlow(key.from, key.to, actorWorkObjectPairs) {
			confidence -= domain.SequentialFlowDiscount
		}

		if confidence <= 0 {
			continue
		}

		desc := fmt.Sprintf("one-way flow from '%s' to '%s' with activities [%s]",
			key.from, key.to, strings.Join(info.activities, ", "))

		signal, err := domain.NewBoundarySignal(domain.SignalTypeOneWayFlow, desc)
		if err != nil {
			continue
		}

		signals = append(signals, signalWithContext{
			signal:     signal,
			actors:     []string{key.from, key.to},
			stories:    setToSlice(info.stories),
			confidence: confidence,
		})
	}

	return signals
}

// hasSequentialFlow checks if two actors co-appear in the same story and both
// act on the same work object.
func (d *AlgorithmicDetector) hasSequentialFlow(actorA, actorB string, pairs []storyActorWO) bool {
	// Group by story -> actor -> set of work objects.
	type storyActors struct {
		actorWOs map[string]map[string]struct{}
	}

	storyMap := make(map[string]*storyActors)

	for _, p := range pairs {
		sa, ok := storyMap[p.storyTitle]
		if !ok {
			sa = &storyActors{actorWOs: make(map[string]map[string]struct{})}
			storyMap[p.storyTitle] = sa
		}

		if sa.actorWOs[p.actor] == nil {
			sa.actorWOs[p.actor] = make(map[string]struct{})
		}

		sa.actorWOs[p.actor][p.workObject] = struct{}{}
	}

	for _, sa := range storyMap {
		aWOs, hasA := sa.actorWOs[actorA]
		bWOs, hasB := sa.actorWOs[actorB]

		if !hasA || !hasB {
			continue
		}

		for wo := range aWOs {
			if _, ok := bWOs[wo]; ok {
				return true
			}
		}
	}

	return false
}

// --- Signal Detector: Org Boundary ---

func (d *AlgorithmicDetector) detectOrgBoundary(stories []*domain.DomainStory) []signalWithContext {
	// Collect actor sets per story.
	type storyActors struct {
		title  string
		actors map[string]struct{}
	}

	var storyGroups []storyActors

	for _, story := range stories {
		storyGroups = append(storyGroups, storyActors{
			title:  story.Title(),
			actors: buildActorNameSet(story),
		})
	}

	var signals []signalWithContext

	// Check each pair of stories for zero overlap.
	for i := 0; i < len(storyGroups); i++ {
		for j := i + 1; j < len(storyGroups); j++ {
			if actorOverlap(storyGroups[i].actors, storyGroups[j].actors) > 0 {
				continue
			}

			desc := fmt.Sprintf("non-overlapping actors between '%s' and '%s'",
				storyGroups[i].title, storyGroups[j].title)

			signal, err := domain.NewBoundarySignal(domain.SignalTypeOrgBoundary, desc)
			if err != nil {
				continue
			}

			// Merge actors from both stories.
			mergedActors := make(map[string]struct{})

			for a := range storyGroups[i].actors {
				mergedActors[a] = struct{}{}
			}

			for a := range storyGroups[j].actors {
				mergedActors[a] = struct{}{}
			}

			signals = append(signals, signalWithContext{
				signal:     signal,
				actors:     setToSlice(mergedActors),
				stories:    []string{storyGroups[i].title, storyGroups[j].title},
				confidence: domain.BaseConfidenceOrgBoundary,
			})
		}
	}

	return signals
}

// actorOverlap computes Jaccard-style overlap: |intersection| / |either set|.
func actorOverlap(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}

	intersection := 0

	for name := range a {
		if _, ok := b[name]; ok {
			intersection++
		}
	}

	// Union = |a| + |b| - |intersection|.
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// --- Signal Detector: Work Object Cluster (fallback) ---

// detectWorkObjectClusters groups stories by shared work objects when all
// primary heuristics produce zero signals. It uses union-find to merge
// stories that share any work object name.
func (d *AlgorithmicDetector) detectWorkObjectClusters(stories []*domain.DomainStory) ([]domain.BoundedContextSketch, error) {
	n := len(stories)
	parent := make([]int, n)

	for i := range parent {
		parent[i] = i
	}

	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}

		return x
	}

	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}

	// Build index: work object name (lowercase) -> story indices.
	woIndex := make(map[string][]int)

	for i, story := range stories {
		for name := range buildWorkObjectNameSet(story) {
			woIndex[name] = append(woIndex[name], i)
		}
	}

	// Union stories that share any work object.
	for _, indices := range woIndex {
		for k := 1; k < len(indices); k++ {
			union(indices[0], indices[k])
		}
	}

	// Group stories by root.
	groups := make(map[int][]int)

	for i := range stories {
		root := find(i)
		groups[root] = append(groups[root], i)
	}

	// Sort roots for deterministic output.
	roots := make([]int, 0, len(groups))

	for root := range groups {
		roots = append(roots, root)
	}

	sort.Ints(roots)

	var sketches []domain.BoundedContextSketch

	for _, root := range roots {
		indices := groups[root]

		actorSet := make(map[string]struct{})
		woSet := make(map[string]struct{})

		var storyTitles []string

		for _, idx := range indices {
			story := stories[idx]
			storyTitles = append(storyTitles, story.Title())

			for name := range buildActorNameSet(story) {
				actorSet[name] = struct{}{}
			}

			for name := range buildWorkObjectNameSet(story) {
				woSet[name] = struct{}{}
			}
		}

		name := deriveName(actorSet, woSet)

		desc := fmt.Sprintf("stories [%s] clustered by shared work objects", strings.Join(storyTitles, ", "))

		signal, err := domain.NewBoundarySignal(domain.SignalTypeWorkObjectCluster, desc)
		if err != nil {
			return nil, fmt.Errorf("creating work object cluster signal: %w", err)
		}

		score := domain.ComputeBoundaryScore(domain.BaseConfidenceWorkObjectCluster, 1, len(storyTitles))
		if score > 1.0 {
			score = 1.0
		}

		actors := setToSlice(actorSet)
		workObjects := setToSlice(woSet)
		sort.Strings(storyTitles)

		sketch, err := domain.NewBoundedContextSketch(
			name, vo.SubdomainGeneric, score, actors, workObjects, storyTitles, []domain.BoundarySignal{signal}, vo.AIInferred,
		)
		if err != nil {
			return nil, fmt.Errorf("creating work object cluster sketch %q: %w", name, err)
		}

		sketches = append(sketches, sketch)
	}

	return sketches, nil
}

// --- Clustering and Scoring ---

func (d *AlgorithmicDetector) clusterAndScore(
	signals []signalWithContext,
) ([]domain.BoundedContextSketch, error) {
	// Group signals by overlap of actors/work objects.
	clusters := d.clusterSignals(signals)

	var sketches []domain.BoundedContextSketch

	for _, cluster := range clusters {
		if len(cluster) == 0 {
			continue
		}

		// Collect cluster-level data.
		actorSet := make(map[string]struct{})
		woSet := make(map[string]struct{})
		storySet := make(map[string]struct{})
		typeSet := make(map[domain.SignalType]struct{})

		var confSum float64

		var domainSignals []domain.BoundarySignal

		for _, sc := range cluster {
			domainSignals = append(domainSignals, sc.signal)
			confSum += sc.confidence
			typeSet[sc.signal.Type()] = struct{}{}

			for _, a := range sc.actors {
				actorSet[a] = struct{}{}
			}

			for _, wo := range sc.workObjects {
				woSet[wo] = struct{}{}
			}

			for _, st := range sc.stories {
				storySet[st] = struct{}{}
			}
		}

		signalAvg := confSum / float64(len(cluster))
		distinctTypeCount := len(typeSet)
		storyCount := len(storySet)

		score := domain.ComputeBoundaryScore(signalAvg, distinctTypeCount, storyCount)

		// Apply org boundary story ceiling.
		if storyCount < domain.OrgBoundaryStoryCeilingCount {
			hasOnlyOrgSignals := true

			for _, sc := range cluster {
				if sc.signal.Type() != domain.SignalTypeOrgBoundary {
					hasOnlyOrgSignals = false

					break
				}
			}

			if hasOnlyOrgSignals && score > domain.OrgBoundaryStoryCeilingScore {
				score = domain.OrgBoundaryStoryCeilingScore
			}
		}

		// Cap at 1.0.
		if score > 1.0 {
			score = 1.0
		}

		name := deriveName(actorSet, woSet)
		actors := setToSlice(actorSet)
		workObjects := setToSlice(woSet)
		storyTitles := setToSlice(storySet)

		sketch, err := domain.NewBoundedContextSketch(
			name, vo.SubdomainGeneric, score, actors, workObjects, storyTitles, domainSignals, vo.AIInferred,
		)
		if err != nil {
			return nil, fmt.Errorf("creating bounded context sketch %q: %w", name, err)
		}

		sketches = append(sketches, sketch)
	}

	return sketches, nil
}

// clusterSignals groups signals by overlap of actors and work objects.
func (d *AlgorithmicDetector) clusterSignals(signals []signalWithContext) [][]signalWithContext {
	// Simple clustering: merge signals that share any actor or work object.
	// Use union-find for efficient merging.
	n := len(signals)
	parent := make([]int, n)

	for i := range parent {
		parent[i] = i
	}

	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}

		return x
	}

	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}

	// Build index: entity name -> signal indices.
	entityIndex := make(map[string][]int)

	for i, sc := range signals {
		for _, a := range sc.actors {
			entityIndex[a] = append(entityIndex[a], i)
		}

		for _, wo := range sc.workObjects {
			entityIndex[wo] = append(entityIndex[wo], i)
		}

		// Also cluster by story title overlap.
		for _, st := range sc.stories {
			entityIndex["story:"+st] = append(entityIndex["story:"+st], i)
		}
	}

	// Union signals that share entities.
	for _, indices := range entityIndex {
		for k := 1; k < len(indices); k++ {
			union(indices[0], indices[k])
		}
	}

	// Group by root.
	groups := make(map[int][]signalWithContext)

	for i, sc := range signals {
		root := find(i)
		groups[root] = append(groups[root], sc)
	}

	// Collect clusters in deterministic order (by lowest signal index in group).
	roots := make([]int, 0, len(groups))
	for root := range groups {
		roots = append(roots, root)
	}

	sort.Ints(roots)

	clusters := make([][]signalWithContext, 0, len(roots))
	for _, root := range roots {
		clusters = append(clusters, groups[root])
	}

	return clusters
}

// deriveName creates a sketch name from dominant actors and work objects.
// TODO: plural stripping deferred — add in follow-up ticket.
func deriveName(actors, workObjects map[string]struct{}) string {
	var parts []string

	// Prefer work objects for naming (they describe "what" the context is about).
	for wo := range workObjects {
		parts = append(parts, capitalize(wo))
	}

	if len(parts) == 0 {
		for a := range actors {
			parts = append(parts, capitalize(a))
		}
	}

	if len(parts) == 0 {
		return "Unknown"
	}

	sort.Strings(parts)

	// Limit to 2 parts for readability.
	if len(parts) > 2 {
		parts = parts[:2]
	}

	return strings.Join(parts, "-")
}

// capitalize returns s with its first letter uppercased.
func capitalize(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

// --- Helpers ---

// buildWorkObjectNameSet returns a set of lowercased work object names from a story.
func buildWorkObjectNameSet(story *domain.DomainStory) map[string]struct{} {
	wos := story.WorkObjects()
	set := make(map[string]struct{}, len(wos))

	for _, wo := range wos {
		set[strings.ToLower(wo.Name())] = struct{}{}
	}

	return set
}

// buildActorNameSet returns a set of lowercased actor names from a story.
func buildActorNameSet(story *domain.DomainStory) map[string]struct{} {
	actors := story.Actors()
	set := make(map[string]struct{}, len(actors))

	for _, a := range actors {
		set[strings.ToLower(a.Name())] = struct{}{}
	}

	return set
}

// setToSlice converts a string set to a sorted slice.
func setToSlice(set map[string]struct{}) []string {
	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}

	sort.Strings(result)

	return result
}
