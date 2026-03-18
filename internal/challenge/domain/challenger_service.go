package domain

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Generate inspects a DomainModel and produces typed Challenge value objects
// that probe for gaps: ambiguous language, missing invariants, unexamined
// failure modes, and questionable boundaries.
//
// Rule-based covers 4 of 6 types: LANGUAGE, INVARIANT, FAILURE_MODE, BOUNDARY.
// AGGREGATE and COMMUNICATION require deeper analysis (LLM-only).
func Generate(model *ddd.DomainModel, maxPerType int) []Challenge {
	var challenges []Challenge
	challenges = append(challenges, languageChallenges(model, maxPerType)...)
	challenges = append(challenges, invariantChallenges(model, maxPerType)...)
	challenges = append(challenges, failureModeChallenges(model, maxPerType)...)
	challenges = append(challenges, boundaryChallenges(model, maxPerType)...)
	return challenges
}

func languageChallenges(model *ddd.DomainModel, maxCount int) []Challenge {
	ambiguous := model.UbiquitousLanguage().FindAmbiguousTerms()
	terms := model.UbiquitousLanguage().Terms()

	var challenges []Challenge
	for _, term := range ambiguous {
		if len(challenges) >= maxCount {
			break
		}
		// Find which contexts use this term
		contextSet := make(map[string]struct{})
		for _, e := range terms {
			if strings.EqualFold(e.Term(), term) {
				contextSet[e.ContextName()] = struct{}{}
			}
		}
		contextNames := sortedKeys(contextSet)
		if len(contextNames) < 2 {
			continue
		}
		targetContext := contextNames[0]
		c, err := NewChallenge(
			ChallengeLanguage,
			fmt.Sprintf("The term '%s' appears in %s. Does it mean the same thing in each context, or should each context have its own definition?",
				term, strings.Join(contextNames, ", ")),
			targetContext,
			fmt.Sprintf("UL glossary: '%s' in %s", term, strings.Join(contextNames, ", ")),
			"",
		)
		if err == nil {
			challenges = append(challenges, c)
		}
	}
	return challenges
}

func invariantChallenges(model *ddd.DomainModel, maxCount int) []Challenge {
	coreSupportingNames := make(map[string]struct{})
	for _, ctx := range model.BoundedContexts() {
		cl := ctx.Classification()
		if cl != nil && (*cl == vo.SubdomainCore || *cl == vo.SubdomainSupporting) {
			coreSupportingNames[ctx.Name()] = struct{}{}
		}
	}

	var challenges []Challenge
	for _, agg := range model.AggregateDesigns() {
		if len(challenges) >= maxCount {
			break
		}
		if _, ok := coreSupportingNames[agg.ContextName()]; !ok {
			continue
		}
		if len(agg.Invariants()) == 0 {
			c, err := NewChallenge(
				ChallengeInvariant,
				fmt.Sprintf("Aggregate '%s' in %s has no invariants listed. What business rules must this aggregate protect?",
					agg.Name(), agg.ContextName()),
				agg.ContextName(),
				fmt.Sprintf("Aggregate design: %s", agg.Name()),
				"",
			)
			if err == nil {
				challenges = append(challenges, c)
			}
		}
	}
	return challenges
}

func failureModeChallenges(model *ddd.DomainModel, maxCount int) []Challenge {
	coreContextNames := make(map[string]struct{})
	for _, ctx := range model.BoundedContexts() {
		cl := ctx.Classification()
		if cl != nil && *cl == vo.SubdomainCore {
			coreContextNames[ctx.Name()] = struct{}{}
		}
	}
	if len(coreContextNames) == 0 {
		return nil
	}

	targetContext := sortedKeys(coreContextNames)[0]

	var challenges []Challenge
	for _, story := range model.DomainStories() {
		if len(challenges) >= maxCount {
			break
		}
		for _, step := range story.Steps() {
			if len(challenges) >= maxCount {
				break
			}
			c, err := NewChallenge(
				ChallengeFailureMode,
				fmt.Sprintf("In story '%s', what happens if this step fails: '%s'?",
					story.Name(), step),
				targetContext,
				fmt.Sprintf("Domain story: %s", story.Name()),
				"",
			)
			if err == nil {
				challenges = append(challenges, c)
			}
		}
	}
	return challenges
}

func boundaryChallenges(model *ddd.DomainModel, maxCount int) []Challenge {
	contexts := model.BoundedContexts()
	if len(contexts) < 2 {
		return nil
	}

	contextNames := make(map[string]struct{})
	for _, c := range contexts {
		contextNames[c.Name()] = struct{}{}
	}

	var challenges []Challenge
	for _, rel := range model.ContextRelationships() {
		if len(challenges) >= maxCount {
			break
		}
		if _, ok := contextNames[rel.Downstream()]; !ok {
			continue
		}
		c, err := NewChallenge(
			ChallengeBoundary,
			fmt.Sprintf("Context '%s' depends on '%s' via %s. Could '%s' own this data directly instead?",
				rel.Downstream(), rel.Upstream(), rel.IntegrationPattern(), rel.Downstream()),
			rel.Downstream(),
			fmt.Sprintf("Context map: %s → %s", rel.Upstream(), rel.Downstream()),
			"",
		)
		if err == nil {
			challenges = append(challenges, c)
		}
	}
	return challenges
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
