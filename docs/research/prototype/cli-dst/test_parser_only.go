//go:build ignore

// This is a standalone test to exercise the README parser on all three sample domains.
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	samples := map[string]string{
		"ecommerce": `An online marketplace where customers browse products from multiple sellers, add items to their cart, pay via credit card or PayPal, and receive home delivery. Sellers manage their own inventory and pricing. The platform takes a commission on each sale.`,

		"vet-clinic": `A management system for a veterinary clinic. Pet owners book appointments by phone or online. The receptionist manages the schedule. Veterinarians examine animals, record diagnoses, prescribe treatments. The clinic tracks medical history per animal and bills owners after visits.`,
	}

	// Also load alto README
	readme, err := os.ReadFile("/home/kusanagi/Alty/alty-cli/README.md")
	if err == nil {
		// Take just the first paragraph
		lines := strings.Split(string(readme), "\n")
		var para []string
		for _, l := range lines {
			if len(para) > 0 && strings.TrimSpace(l) == "" {
				break
			}
			if strings.TrimSpace(l) != "" && !strings.HasPrefix(l, "#") && !strings.HasPrefix(l, "<") && !strings.HasPrefix(l, ">") && !strings.HasPrefix(l, "---") && !strings.HasPrefix(l, "```") {
				para = append(para, strings.TrimSpace(l))
			}
		}
		if len(para) > 0 {
			samples["alto"] = strings.Join(para, " ")
		}
	}

	for name, text := range samples {
		fmt.Printf("=== %s ===\n", strings.ToUpper(name))
		fmt.Printf("Input: %s\n\n", text)

		parsed := ParseREADME(text)
		fmt.Printf("All Nouns: %v\n", parsed.Nouns)
		fmt.Printf("Verbs: %v\n", parsed.Verbs)
		fmt.Printf("Actors: %v\n", parsed.Actors)
		fmt.Printf("Work Objects: %v\n", parsed.WorkObjects)
		fmt.Println()

		sentences := ProposeSentencesFromParsed(parsed)
		fmt.Println("Proposed sentences:")
		for _, s := range sentences {
			fmt.Printf("  %s [trust: %s]\n", s.Format(), s.TrustLevel)
		}
		fmt.Println()
	}
}
