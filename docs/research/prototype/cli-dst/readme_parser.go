package main

import (
	"regexp"
	"strings"
)

// ParsedREADME holds extracted nouns and verbs from a README.
type ParsedREADME struct {
	RawText     string
	Nouns       []string // Candidate actors and work objects
	Verbs       []string // Candidate activities
	Actors      []string // Subset of nouns that look like actors
	WorkObjects []string // Subset of nouns that look like work objects
}

// Common stop words and non-domain terms to filter out.
var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true,
	"did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "shall": true, "can": true, "need": true,
	"dare": true, "to": true, "of": true, "in": true, "for": true,
	"on": true, "with": true, "at": true, "by": true, "from": true,
	"as": true, "into": true, "through": true, "during": true,
	"before": true, "after": true, "above": true, "below": true,
	"between": true, "out": true, "off": true, "over": true, "under": true,
	"again": true, "further": true, "then": true, "once": true,
	"here": true, "there": true, "when": true, "where": true, "why": true,
	"how": true, "all": true, "each": true, "every": true, "both": true,
	"few": true, "more": true, "most": true, "other": true, "some": true,
	"such": true, "no": true, "nor": true, "not": true, "only": true,
	"own": true, "same": true, "so": true, "than": true, "too": true,
	"very": true, "just": true, "also": true, "this": true, "that": true,
	"these": true, "those": true, "it": true, "its": true, "they": true,
	"their": true, "them": true, "we": true, "our": true, "you": true,
	"your": true, "he": true, "she": true, "his": true, "her": true,
}

// Common actor-indicating patterns (capitalized nouns that look like roles).
var actorPatterns = []string{
	"customer", "user", "seller", "buyer", "admin", "administrator",
	"manager", "owner", "developer", "operator", "agent", "system",
	"service", "provider", "client", "patient", "doctor", "nurse",
	"receptionist", "veterinarian", "vet", "staff", "employee",
	"platform", "tool", "cli",
}

// Common work object patterns.
var workObjectPatterns = []string{
	"order", "product", "item", "cart", "invoice", "payment",
	"appointment", "schedule", "record", "report", "ticket",
	"document", "config", "configuration", "plan", "project",
	"account", "profile", "catalog", "inventory", "delivery",
	"commission", "treatment", "diagnosis", "prescription",
	"blueprint", "artifact", "template",
}

// ParseREADME extracts domain-relevant nouns and verbs from README text.
func ParseREADME(text string) ParsedREADME {
	result := ParsedREADME{RawText: text}

	// Extract capitalized words (potential proper nouns / domain terms)
	capitalizedRe := regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\b`)
	matches := capitalizedRe.FindAllString(text, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		lower := strings.ToLower(m)
		if stopWords[lower] || seen[lower] {
			continue
		}
		seen[lower] = true
		result.Nouns = append(result.Nouns, m)
	}

	// Extract verbs (words following common patterns)
	verbRe := regexp.MustCompile(`\b(browse|create|manage|book|pay|deliver|check|generate|build|order|track|bill|schedule|examine|record|prescribe|process|confirm|cancel|update|add|remove|select|review|approve|submit|send|receive|assign|notify)\w*\b`)
	verbMatches := verbRe.FindAllString(strings.ToLower(text), -1)

	verbSeen := make(map[string]bool)
	for _, v := range verbMatches {
		base := strings.TrimSuffix(strings.TrimSuffix(v, "s"), "es")
		base = strings.TrimSuffix(base, "ed")
		base = strings.TrimSuffix(base, "ing")
		if !verbSeen[base] {
			verbSeen[base] = true
			result.Verbs = append(result.Verbs, base)
		}
	}

	// Classify nouns as actors vs work objects
	for _, noun := range result.Nouns {
		lower := strings.ToLower(noun)
		isActor := false
		for _, p := range actorPatterns {
			if strings.Contains(lower, p) {
				isActor = true
				break
			}
		}
		if isActor {
			result.Actors = append(result.Actors, noun)
			continue
		}
		isWorkObject := false
		for _, p := range workObjectPatterns {
			if strings.Contains(lower, p) {
				isWorkObject = true
				break
			}
		}
		if isWorkObject {
			result.WorkObjects = append(result.WorkObjects, noun)
		}
	}

	return result
}

// ProposeSentencesFromParsed generates candidate story sentences from parsed README.
// This is the template-matching approach (Decision 2, Option A).
func ProposeSentencesFromParsed(parsed ParsedREADME) []StorySentence {
	var sentences []StorySentence
	seq := 1

	// If we have actors and verbs, try to construct sentences
	actors := parsed.Actors
	if len(actors) == 0 {
		// Fallback: use first few nouns as actors
		for i, n := range parsed.Nouns {
			if i >= 3 {
				break
			}
			actors = append(actors, n)
		}
	}

	objects := parsed.WorkObjects
	if len(objects) == 0 {
		// Fallback: use remaining nouns as objects
		actorSet := make(map[string]bool)
		for _, a := range actors {
			actorSet[strings.ToLower(a)] = true
		}
		for _, n := range parsed.Nouns {
			if !actorSet[strings.ToLower(n)] {
				objects = append(objects, n)
			}
		}
	}

	verbs := parsed.Verbs
	if len(verbs) == 0 {
		verbs = []string{"uses", "creates", "manages"}
	}

	// Generate sentences: pair actors with verbs and objects
	for i, actor := range actors {
		if i >= len(verbs) || i >= len(objects) {
			break
		}
		sentences = append(sentences, StorySentence{
			Number:     seq,
			Subject:    actor,
			Activity:   verbs[i%len(verbs)],
			Object:     objects[i%len(objects)],
			TrustLevel: TrustAIInferred,
		})
		seq++
	}

	// Try to add more by cycling through remaining verbs/objects
	for i := len(actors); i < len(verbs) && i < len(objects)+len(actors); i++ {
		actorIdx := i % len(actors)
		verbIdx := i % len(verbs)
		objIdx := i % len(objects)
		if actorIdx < len(actors) && verbIdx < len(verbs) && objIdx < len(objects) {
			sentences = append(sentences, StorySentence{
				Number:     seq,
				Subject:    actors[actorIdx],
				Activity:   verbs[verbIdx],
				Object:     objects[objIdx],
				TrustLevel: TrustAIInferred,
			})
			seq++
		}
	}

	return sentences
}
