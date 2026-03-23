package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ConsultantFlow implements the consultant-proposes interaction pattern.
// AI generates a full story from the README, then the user refines sentence by sentence.
type ConsultantFlow struct {
	scanner *bufio.Scanner
	story   *DomainStory
}

// NewConsultantFlow creates a new consultant flow.
func NewConsultantFlow() *ConsultantFlow {
	return &ConsultantFlow{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// readLine reads a line of user input.
func (f *ConsultantFlow) readLine() string {
	f.scanner.Scan()
	return strings.TrimSpace(f.scanner.Text())
}

// Run executes the consultant-proposes flow.
func (f *ConsultantFlow) Run(readme string) (*DomainStory, error) {
	fmt.Println()
	fmt.Println("=== CONSULTANT-PROPOSES MODE ===")
	fmt.Println()
	fmt.Println("I'll propose a complete story based on your README, then you can")
	fmt.Println("refine it sentence by sentence.")
	fmt.Println()

	// Step 1: Parse README
	parsed := ParseREADME(readme)

	fmt.Println("From your README, I extracted:")
	fmt.Printf("  Probable actors: %s\n", strings.Join(parsed.Actors, ", "))
	fmt.Printf("  Probable work objects: %s\n", strings.Join(parsed.WorkObjects, ", "))
	fmt.Printf("  Domain verbs: %s\n", strings.Join(parsed.Verbs, ", "))
	fmt.Println()

	// Step 2: Generate proposed story from template matching
	proposed := ProposeSentencesFromParsed(parsed)

	// Build story
	f.story = &DomainStory{
		Title:     "Primary Workflow",
		StoryType: StoryTypeCoarse,
		Trigger:   "(inferred from README)",
	}

	// Add actors and work objects
	for _, a := range parsed.Actors {
		f.story.Actors = append(f.story.Actors, StoryActor{
			Name:       a,
			Type:       "person",
			TrustLevel: TrustAIInferred,
		})
	}
	for _, w := range parsed.WorkObjects {
		f.story.WorkObjects = append(f.story.WorkObjects, WorkObject{
			Name:       w,
			Type:       "document",
			TrustLevel: TrustAIInferred,
		})
	}

	f.story.Sentences = proposed

	// Step 3: Present proposed story
	fmt.Println("Here's my proposed story based on your README:")
	fmt.Println()
	fmt.Println(f.story.FormatText())

	// Step 4: Refine sentence by sentence
	fmt.Println("Let's refine this story. For each sentence, you can:")
	fmt.Println("  [Y]     - Accept as-is")
	fmt.Println("  [n]     - Remove this sentence")
	fmt.Println("  [e]dit  - Edit this sentence")
	fmt.Println("  [i]nsert - Insert a sentence before this one")
	fmt.Println()

	f.refineSentences()

	// Step 5: Handle missing elements
	f.askForMissing()

	// Step 6: Ask for title
	fmt.Println()
	fmt.Print("Give this story a title (or press Enter to keep current): ")
	title := f.readLine()
	if title != "" {
		f.story.Title = title
	}

	// Step 7: Ask for trigger
	fmt.Print("What triggers this workflow? (or press Enter to skip): ")
	trigger := f.readLine()
	if trigger != "" {
		f.story.Trigger = trigger
	}

	// Step 8: Detect branching
	f.detectBranching()

	// Step 9: Synthesis checkpoint
	f.synthesisCheckpoint()

	return f.story, nil
}

// refineSentences walks through each sentence for user refinement.
func (f *ConsultantFlow) refineSentences() {
	i := 0
	for i < len(f.story.Sentences) {
		s := f.story.Sentences[i]
		fmt.Printf("\n  Sentence %d: %s\n", s.Number, s.Format())
		fmt.Print("  [Y/n/e/i]? ")

		input := strings.ToLower(f.readLine())

		switch {
		case input == "" || input == "y":
			// Accept - upgrade trust level
			f.story.Sentences[i].TrustLevel = TrustUserConfirmed
			i++

		case input == "n":
			// Remove sentence
			f.story.Sentences = append(f.story.Sentences[:i], f.story.Sentences[i+1:]...)
			// Renumber remaining
			f.renumberSentences()
			fmt.Println("  (removed)")

		case input == "e" || input == "edit":
			// Edit sentence
			edited := f.editSentence(s)
			f.story.Sentences[i] = edited
			i++

		case input == "i" || input == "insert":
			// Insert before
			newSentence := f.composeSentence(s.Number)
			// Shift all sentences from i onward
			f.story.Sentences = append(f.story.Sentences[:i+1], f.story.Sentences[i:]...)
			f.story.Sentences[i] = newSentence
			f.renumberSentences()
			i++ // Move past the inserted sentence

		default:
			fmt.Println("  (unrecognized, treating as accept)")
			f.story.Sentences[i].TrustLevel = TrustUserConfirmed
			i++
		}
	}
}

// editSentence prompts user to edit a sentence's components.
func (f *ConsultantFlow) editSentence(original StorySentence) StorySentence {
	fmt.Printf("  Subject [%s]: ", original.Subject)
	subject := f.readLine()
	if subject == "" {
		subject = original.Subject
	}

	fmt.Printf("  Activity/verb [%s]: ", original.Activity)
	activity := f.readLine()
	if activity == "" {
		activity = original.Activity
	}

	fmt.Printf("  Object [%s]: ", original.Object)
	object := f.readLine()
	if object == "" {
		object = original.Object
	}

	fmt.Printf("  Preposition (e.g., 'using', 'for', 'with') [%s]: ", original.Preposition)
	prep := f.readLine()
	if prep == "" {
		prep = original.Preposition
	}

	indirectObj := original.IndirectObject
	if prep != "" {
		fmt.Printf("  Indirect object [%s]: ", original.IndirectObject)
		indirectObj = f.readLine()
		if indirectObj == "" {
			indirectObj = original.IndirectObject
		}
	}

	// Check if the actor is new
	f.ensureActor(subject)
	f.ensureWorkObject(object)
	if indirectObj != "" {
		// Could be actor or work object; treat as work object by default
		f.ensureWorkObject(indirectObj)
	}

	return StorySentence{
		Number:         original.Number,
		Subject:        subject,
		Activity:       activity,
		Object:         object,
		Preposition:    prep,
		IndirectObject: indirectObj,
		TrustLevel:     TrustUserStated,
	}
}

// composeSentence prompts the user to create a new sentence.
func (f *ConsultantFlow) composeSentence(atNumber int) StorySentence {
	fmt.Println("  Compose a new sentence:")
	fmt.Print("  Subject (who does this): ")
	subject := f.readLine()
	fmt.Print("  Activity (does what): ")
	activity := f.readLine()
	fmt.Print("  Object (to/with what): ")
	object := f.readLine()
	fmt.Print("  Preposition (optional, e.g., 'using', 'for'): ")
	prep := f.readLine()
	indirectObj := ""
	if prep != "" {
		fmt.Print("  Indirect object: ")
		indirectObj = f.readLine()
	}

	f.ensureActor(subject)
	f.ensureWorkObject(object)

	return StorySentence{
		Number:         atNumber,
		Subject:        subject,
		Activity:       activity,
		Object:         object,
		Preposition:    prep,
		IndirectObject: indirectObj,
		TrustLevel:     TrustUserStated,
	}
}

// askForMissing asks if there are actors or work objects not yet in the story.
func (f *ConsultantFlow) askForMissing() {
	fmt.Println()
	fmt.Print("Any actors I missed? (comma-separated, or Enter to skip): ")
	actors := f.readLine()
	if actors != "" {
		for _, a := range strings.Split(actors, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				f.story.Actors = append(f.story.Actors, StoryActor{
					Name:       a,
					Type:       "person",
					TrustLevel: TrustUserStated,
				})
			}
		}
	}

	fmt.Print("Any work objects I missed? (comma-separated, or Enter to skip): ")
	objects := f.readLine()
	if objects != "" {
		for _, w := range strings.Split(objects, ",") {
			w = strings.TrimSpace(w)
			if w != "" {
				f.story.WorkObjects = append(f.story.WorkObjects, WorkObject{
					Name:       w,
					Type:       "document",
					TrustLevel: TrustUserStated,
				})
			}
		}
	}

	// Ask if user wants to add more sentences at the end
	fmt.Print("Add more sentences at the end? [y/N]: ")
	if strings.ToLower(f.readLine()) == "y" {
		for {
			nextNum := len(f.story.Sentences) + 1
			s := f.composeSentence(nextNum)
			f.story.Sentences = append(f.story.Sentences, s)
			f.renumberSentences()
			fmt.Print("Add another? [y/N]: ")
			if strings.ToLower(f.readLine()) != "y" {
				break
			}
		}
	}
}

// detectBranching checks for branching language in the story.
func (f *ConsultantFlow) detectBranching() {
	branchingWords := []string{"sometimes", "or", "alternatively", "if", "optionally", "either"}

	for _, s := range f.story.Sentences {
		combined := strings.ToLower(s.Activity + " " + s.Object)
		for _, bw := range branchingWords {
			if strings.Contains(combined, bw) {
				fmt.Printf("\n  I noticed branching language in sentence %d: %q\n", s.Number, s.Format())
				fmt.Println("  Domain Storytelling uses ONE scenario per story.")
				fmt.Println("  Should this alternative become a separate story?")
				fmt.Print("  [Y/n]: ")
				if strings.ToLower(f.readLine()) != "n" {
					fmt.Print("  Name for the variation story: ")
					name := f.readLine()
					if name != "" {
						f.story.Variations = append(f.story.Variations, name)
					}
				}
				break // Only flag once per sentence
			}
		}
	}
}

// synthesisCheckpoint replays the complete story and asks for confirmation.
func (f *ConsultantFlow) synthesisCheckpoint() {
	fmt.Println()
	fmt.Println("=== SYNTHESIS CHECKPOINT ===")
	fmt.Println()
	fmt.Println("Here's the complete story as I understand it:")
	fmt.Println()
	fmt.Println(f.story.FormatText())

	fmt.Print("Does this capture your primary workflow? [Y/n/edit]: ")
	input := strings.ToLower(f.readLine())

	switch {
	case input == "" || input == "y":
		fmt.Println("Story confirmed.")
	case input == "n":
		fmt.Println("What needs to change?")
		fmt.Print("Enter sentence numbers to edit (comma-separated): ")
		nums := f.readLine()
		for _, numStr := range strings.Split(nums, ",") {
			numStr = strings.TrimSpace(numStr)
			num, err := strconv.Atoi(numStr)
			if err != nil {
				continue
			}
			for i, s := range f.story.Sentences {
				if s.Number == num {
					fmt.Printf("  Editing sentence %d:\n", num)
					f.story.Sentences[i] = f.editSentence(s)
					break
				}
			}
		}
		// Show again after edits
		fmt.Println()
		fmt.Println("Updated story:")
		fmt.Println(f.story.FormatText())
	case input == "edit":
		// Re-run the refinement loop
		f.refineSentences()
	}
}

// ensureActor adds an actor if not already present.
func (f *ConsultantFlow) ensureActor(name string) {
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
func (f *ConsultantFlow) ensureWorkObject(name string) {
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

// renumberSentences reassigns sequential numbers to all sentences.
func (f *ConsultantFlow) renumberSentences() {
	for i := range f.story.Sentences {
		f.story.Sentences[i].Number = i + 1
	}
}
