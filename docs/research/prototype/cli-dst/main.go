package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Sample READMEs for testing.
var sampleREADMEs = map[string]string{
	"ecommerce": `An online marketplace where customers browse products from multiple sellers, add items to their cart, pay via credit card or PayPal, and receive home delivery. Sellers manage their own inventory and pricing. The platform takes a commission on each sale.`,

	"vet-clinic": `A management system for a veterinary clinic. Pet owners book appointments by phone or online. The receptionist manages the schedule. Veterinarians examine animals, record diagnoses, prescribe treatments. The clinic tracks medical history per animal and bills owners after visits.`,
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	readLine := func() string {
		scanner.Scan()
		return strings.TrimSpace(scanner.Text())
	}

	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println("  CLI Domain Storytelling Prototype")
	fmt.Println("  Research artifact for alto discovery redesign")
	fmt.Println("==========================================================")
	fmt.Println()

	// Step 1: Mode selection
	fmt.Println("How deep should we go?")
	fmt.Println()
	fmt.Println("  A) RAPID   — 3 stories, ~15 min")
	fmt.Println("     Enough for MVP project setup. Covers primary workflow,")
	fmt.Println("     main failure case, and one secondary workflow.")
	fmt.Println()
	fmt.Println("  B) THOROUGH — 5+ stories, ~30 min")
	fmt.Println("     Comprehensive domain model. Adds fine-grained stories for")
	fmt.Println("     core subdomains and explicit business rules extraction.")
	fmt.Println()
	fmt.Print("Choose mode [A/b]: ")
	modeChoice := strings.ToLower(readLine())
	mode := ModeRapid
	if modeChoice == "b" {
		mode = ModeThorough
	}
	fmt.Printf("\nMode: %s\n", mode)

	// Step 2: Select domain
	fmt.Println()
	fmt.Println("Select a sample domain (or provide your own README):")
	fmt.Println()
	fmt.Println("  1) E-commerce marketplace")
	fmt.Println("  2) Veterinary clinic")
	fmt.Println("  3) Read from file (provide path)")
	fmt.Println()
	fmt.Print("Choice [1/2/3]: ")
	domainChoice := readLine()

	var readme string
	switch domainChoice {
	case "1":
		readme = sampleREADMEs["ecommerce"]
	case "2":
		readme = sampleREADMEs["vet-clinic"]
	case "3":
		fmt.Print("Path to README.md: ")
		path := readLine()
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		readme = string(data)
	default:
		readme = sampleREADMEs["vet-clinic"]
	}

	fmt.Println()
	fmt.Println("README content:")
	fmt.Println("---")
	fmt.Println(readme)
	fmt.Println("---")

	// Step 3: Select interaction pattern
	fmt.Println()
	fmt.Println("Choose interaction pattern to test:")
	fmt.Println()
	fmt.Println("  A) Consultant-proposes")
	fmt.Println("     I generate a full story from your README, then you")
	fmt.Println("     refine it sentence by sentence.")
	fmt.Println()
	fmt.Println("  B) User-narrates")
	fmt.Println("     I ask moderator questions ('What happens first?',")
	fmt.Println("     'Who does that?'), you tell the story, I structure it.")
	fmt.Println()
	fmt.Println("  C) Both (test one after the other)")
	fmt.Println()
	fmt.Print("Choice [A/b/c]: ")
	patternChoice := strings.ToLower(readLine())

	var stories []*DomainStory

	switch patternChoice {
	case "b":
		flow := NewNarratorFlow()
		story, err := flow.Run(readme)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		stories = append(stories, story)

	case "c":
		// Test both
		fmt.Println("\n--- Testing Consultant-Proposes first ---")
		cFlow := NewConsultantFlow()
		cStory, err := cFlow.Run(readme)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in consultant flow: %v\n", err)
			os.Exit(1)
		}
		stories = append(stories, cStory)

		fmt.Println("\n--- Now testing User-Narrates ---")
		nFlow := NewNarratorFlow()
		nStory, err := nFlow.Run(readme)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in narrator flow: %v\n", err)
			os.Exit(1)
		}
		stories = append(stories, nStory)

	default: // "a" or empty
		flow := NewConsultantFlow()
		story, err := flow.Run(readme)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		stories = append(stories, story)
	}

	// Step 4: Show comparison (if both were tested)
	if len(stories) > 1 {
		fmt.Println()
		fmt.Println("==========================================================")
		fmt.Println("  COMPARISON: Consultant-Proposes vs User-Narrates")
		fmt.Println("==========================================================")
		fmt.Println()

		for i, s := range stories {
			pattern := "Consultant-Proposes"
			if i == 1 {
				pattern = "User-Narrates"
			}
			fmt.Printf("--- %s ---\n", pattern)
			fmt.Println(s.FormatText())
			fmt.Println()
		}

		fmt.Println("Which felt more natural?")
		fmt.Println("  A) Consultant-proposes (I propose, you refine)")
		fmt.Println("  B) User-narrates (you tell, I structure)")
		fmt.Println("  C) Hybrid (I propose first story, you narrate subsequent)")
		fmt.Print("Your preference: ")
		readLine()
	}

	// Step 5: Final output
	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println("  FINAL OUTPUT")
	fmt.Println("==========================================================")
	for _, s := range stories {
		fmt.Println()
		fmt.Println(s.FormatText())
	}

	// Trust level summary
	fmt.Println()
	fmt.Println("Trust Distribution:")
	for _, s := range stories {
		counts := map[TrustLevel]int{}
		for _, sent := range s.Sentences {
			counts[sent.TrustLevel]++
		}
		total := len(s.Sentences)
		if total == 0 {
			continue
		}
		fmt.Printf("  Story: %q\n", s.Title)
		for _, tl := range []TrustLevel{TrustUserStated, TrustUserConfirmed, TrustAIResearched, TrustAIInferred} {
			count := counts[tl]
			pct := float64(count) / float64(total) * 100
			fmt.Printf("    %s: %d (%.0f%%)\n", tl, count, pct)
		}
	}

	fmt.Println()
	fmt.Printf("Mode: %s\n", mode)
	fmt.Println("Session complete.")
}
