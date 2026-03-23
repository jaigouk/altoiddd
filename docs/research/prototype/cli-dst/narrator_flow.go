package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// NarratorFlow implements the user-narrates interaction pattern.
// AI asks moderator questions, user tells the story, AI structures into sentences.
type NarratorFlow struct {
	scanner *bufio.Scanner
	story   *DomainStory
}

// NewNarratorFlow creates a new narrator flow.
func NewNarratorFlow() *NarratorFlow {
	return &NarratorFlow{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (f *NarratorFlow) readLine() string {
	f.scanner.Scan()
	return strings.TrimSpace(f.scanner.Text())
}

// Run executes the user-narrates flow.
func (f *NarratorFlow) Run(readme string) (*DomainStory, error) {
	fmt.Println()
	fmt.Println("=== USER-NARRATES MODE ===")
	fmt.Println()
	fmt.Println("I'll guide you through telling a domain story with moderator questions.")
	fmt.Println("You tell the story in your own words. I'll structure it into sentences.")
	fmt.Println()

	f.story = &DomainStory{
		StoryType: StoryTypeCoarse,
	}

	// Step 1: Context from README
	fmt.Println("From your README, I can see the general idea of your system.")
	fmt.Println("Let's capture the primary workflow as a domain story.")
	fmt.Println()

	// Step 2: Title
	fmt.Print("What is the main thing users do in your system?\n> ")
	mainAction := f.readLine()
	f.story.Title = mainAction

	// Step 3: Trigger
	fmt.Print("\nWhat triggers this workflow? (What makes it start?)\n> ")
	trigger := f.readLine()
	f.story.Trigger = trigger

	// Step 4: First actor
	fmt.Print("\nWho starts this process? (Name the role, not a person)\n> ")
	firstActor := f.readLine()
	f.story.Actors = append(f.story.Actors, StoryActor{
		Name:       firstActor,
		Type:       "person",
		TrustLevel: TrustUserStated,
	})

	// Step 5: First action
	fmt.Printf("\nWhat does %s do first?\n> ", firstActor)
	firstAction := f.readLine()

	// Parse the action into a sentence
	sentence := f.parseNaturalSentence(firstActor, firstAction, 1)
	f.story.Sentences = append(f.story.Sentences, sentence)

	fmt.Printf("\n  I captured: %s\n", sentence.Format())
	fmt.Print("  Is that right? [Y/n/edit]: ")
	f.handleSentenceConfirmation(0)

	// Step 6: Continue the story
	seq := 2
	for {
		fmt.Printf("\nWhat happens next? (or type 'done' to finish)\n> ")
		nextAction := f.readLine()
		if strings.ToLower(nextAction) == "done" {
			break
		}

		// Check for branching
		if f.containsBranching(nextAction) {
			fmt.Println()
			fmt.Println("  I noticed you mentioned an alternative (\"sometimes\", \"or\", etc.).")
			fmt.Println("  In Domain Storytelling, each story covers ONE scenario.")
			fmt.Print("  Should the alternative become a separate story? [Y/n]: ")
			if strings.ToLower(f.readLine()) != "n" {
				fmt.Print("  Name for the alternative story: ")
				varName := f.readLine()
				if varName != "" {
					f.story.Variations = append(f.story.Variations, varName)
				}
				fmt.Println("  Noted. Let's continue with the main scenario.")
				fmt.Printf("\nSo in the main scenario, what happens next?\n> ")
				nextAction = f.readLine()
				if strings.ToLower(nextAction) == "done" {
					break
				}
			}
		}

		// Ask who does it
		fmt.Print("Who does this? (actor name): ")
		actor := f.readLine()
		if actor == "" {
			// Default to last actor
			if len(f.story.Sentences) > 0 {
				actor = f.story.Sentences[len(f.story.Sentences)-1].Subject
			}
		}

		sentence := f.parseNaturalSentence(actor, nextAction, seq)
		f.story.Sentences = append(f.story.Sentences, sentence)

		fmt.Printf("\n  I captured: %s\n", sentence.Format())
		fmt.Print("  Is that right? [Y/n/edit]: ")
		f.handleSentenceConfirmation(len(f.story.Sentences) - 1)

		// Ensure actor is tracked
		f.ensureActor(actor)

		seq++

		// Periodic synthesis (every 3 sentences)
		if len(f.story.Sentences) > 0 && len(f.story.Sentences)%3 == 0 {
			f.midStoryCheckpoint()
		}
	}

	// Step 7: Ask about annotations
	f.askAnnotations()

	// Step 8: Final synthesis
	f.finalSynthesis()

	return f.story, nil
}

// parseNaturalSentence attempts to structure a natural language description into a StorySentence.
func (f *NarratorFlow) parseNaturalSentence(actor string, description string, seq int) StorySentence {
	// Simple parsing: try to extract verb and object from the description.
	words := strings.Fields(description)

	sentence := StorySentence{
		Number:     seq,
		Subject:    actor,
		TrustLevel: TrustUserStated,
	}

	if len(words) == 0 {
		sentence.Activity = "does"
		sentence.Object = "something"
		return sentence
	}

	// First word is likely the verb
	sentence.Activity = words[0]

	// Look for preposition patterns
	prepWords := map[string]bool{
		"using": true, "with": true, "for": true, "to": true,
		"from": true, "into": true, "via": true,
	}

	objectParts := []string{}
	indirectParts := []string{}
	foundPrep := false
	prepWord := ""

	for _, w := range words[1:] {
		if prepWords[strings.ToLower(w)] && !foundPrep && len(objectParts) > 0 {
			foundPrep = true
			prepWord = strings.ToLower(w)
			continue
		}
		if foundPrep {
			indirectParts = append(indirectParts, w)
		} else {
			objectParts = append(objectParts, w)
		}
	}

	if len(objectParts) > 0 {
		sentence.Object = strings.Join(objectParts, " ")
	} else {
		sentence.Object = "(unspecified)"
	}

	if foundPrep {
		sentence.Preposition = prepWord
		sentence.IndirectObject = strings.Join(indirectParts, " ")
	}

	// Track any work objects mentioned
	if sentence.Object != "(unspecified)" {
		f.ensureWorkObject(sentence.Object)
	}
	if sentence.IndirectObject != "" {
		f.ensureWorkObject(sentence.IndirectObject)
	}

	return sentence
}

// handleSentenceConfirmation handles Y/n/edit response for a sentence.
func (f *NarratorFlow) handleSentenceConfirmation(idx int) {
	input := strings.ToLower(f.readLine())

	switch {
	case input == "" || input == "y":
		f.story.Sentences[idx].TrustLevel = TrustUserStated
	case input == "n":
		fmt.Print("  What should it say instead?\n  > ")
		replacement := f.readLine()
		actor := f.story.Sentences[idx].Subject
		num := f.story.Sentences[idx].Number
		f.story.Sentences[idx] = f.parseNaturalSentence(actor, replacement, num)
	case input == "edit" || input == "e":
		s := f.story.Sentences[idx]
		fmt.Printf("  Subject [%s]: ", s.Subject)
		subj := f.readLine()
		if subj != "" {
			f.story.Sentences[idx].Subject = subj
			f.ensureActor(subj)
		}
		fmt.Printf("  Activity [%s]: ", s.Activity)
		act := f.readLine()
		if act != "" {
			f.story.Sentences[idx].Activity = act
		}
		fmt.Printf("  Object [%s]: ", s.Object)
		obj := f.readLine()
		if obj != "" {
			f.story.Sentences[idx].Object = obj
			f.ensureWorkObject(obj)
		}
	}
}

// midStoryCheckpoint replays the story so far for confirmation.
func (f *NarratorFlow) midStoryCheckpoint() {
	fmt.Println()
	fmt.Println("  --- Quick checkpoint ---")
	fmt.Println("  Here's what we have so far:")
	fmt.Println()
	for _, s := range f.story.Sentences {
		fmt.Printf("    %s\n", s.Format())
	}
	fmt.Println()
	fmt.Print("  Does this look right so far? [Y/n]: ")
	if strings.ToLower(f.readLine()) == "n" {
		fmt.Print("  Which sentence number needs fixing? ")
		numStr := f.readLine()
		for i, s := range f.story.Sentences {
			if fmt.Sprintf("%d", s.Number) == numStr {
				fmt.Printf("  Editing sentence %d:\n", s.Number)
				f.handleSentenceConfirmation(i)
				break
			}
		}
	}
}

// askAnnotations prompts for business rules and constraints.
func (f *NarratorFlow) askAnnotations() {
	fmt.Println()
	fmt.Println("Now let's capture any business rules or constraints.")
	fmt.Println("These are things like 'must', 'only if', 'cannot', 'always'.")
	fmt.Println()

	for {
		fmt.Print("Any business rules or constraints? (or Enter to skip): ")
		rule := f.readLine()
		if rule == "" {
			break
		}

		ann := Annotation{
			Text:       rule,
			Type:       "invariant",
			TrustLevel: TrustUserStated,
		}

		fmt.Print("Does this apply to a specific sentence number? (or Enter for story-level): ")
		numStr := f.readLine()
		if numStr != "" {
			num := 0
			fmt.Sscanf(numStr, "%d", &num)
			ann.SentenceNo = num
		}

		f.story.Annotations = append(f.story.Annotations, ann)
	}
}

// finalSynthesis replays the complete story for final confirmation.
func (f *NarratorFlow) finalSynthesis() {
	fmt.Println()
	fmt.Println("=== FINAL SYNTHESIS ===")
	fmt.Println()
	fmt.Println("Here's the complete story:")
	fmt.Println()
	fmt.Println(f.story.FormatText())

	fmt.Print("Does this capture your primary workflow? [Y/n]: ")
	input := strings.ToLower(f.readLine())
	if input == "n" {
		fmt.Println("What needs to change? (You can re-run the refinement.)")
		// In a real implementation, we'd loop back. For prototype purposes, just note it.
		fmt.Print("Notes for revision: ")
		notes := f.readLine()
		if notes != "" {
			f.story.Annotations = append(f.story.Annotations, Annotation{
				Text:       "Revision needed: " + notes,
				Type:       "assumption",
				TrustLevel: TrustUserStated,
			})
		}
	} else {
		fmt.Println("Story confirmed.")
	}
}

// containsBranching checks if text contains branching language.
func (f *NarratorFlow) containsBranching(text string) bool {
	lower := strings.ToLower(text)
	branchingWords := []string{"sometimes", "or ", "alternatively", "either", "optionally"}
	for _, bw := range branchingWords {
		if strings.Contains(lower, bw) {
			return true
		}
	}
	return false
}

// ensureActor adds an actor if not already present.
func (f *NarratorFlow) ensureActor(name string) {
	for _, a := range f.story.Actors {
		if strings.EqualFold(a.Name, name) {
			return
		}
	}
	f.story.Actors = append(f.story.Actors, StoryActor{
		Name:       name,
		Type:       "person",
		TrustLevel: TrustUserStated,
	})
}

// ensureWorkObject adds a work object if not already present.
func (f *NarratorFlow) ensureWorkObject(name string) {
	for _, w := range f.story.WorkObjects {
		if strings.EqualFold(w.Name, name) {
			return
		}
	}
	f.story.WorkObjects = append(f.story.WorkObjects, WorkObject{
		Name:       name,
		Type:       "document",
		TrustLevel: TrustUserStated,
	})
}
