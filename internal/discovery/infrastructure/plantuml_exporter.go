package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// Compile-time interface check.
var _ discoveryapp.PlantUMLRenderer = (*PlantUMLExporter)(nil)

// actorMacros maps domain ActorType to PlantUML DomainStory macro names.
var actorMacros = map[discoverydomain.ActorType]string{
	discoverydomain.ActorTypePerson: "Person",
	discoverydomain.ActorTypeSystem: "System",
	discoverydomain.ActorTypeGroup:  "Group",
}

// workObjectMacros maps domain WorkObjectType to PlantUML DomainStory macro names.
var workObjectMacros = map[discoverydomain.WorkObjectType]string{
	discoverydomain.WorkObjectTypeDocument:     "Document",
	discoverydomain.WorkObjectTypeFolder:       "Folder",
	discoverydomain.WorkObjectTypeCall:         "Call",
	discoverydomain.WorkObjectTypeEmail:        "Email",
	discoverydomain.WorkObjectTypeConversation: "Conversation",
	discoverydomain.WorkObjectTypeInfo:         "Info",
	discoverydomain.WorkObjectTypeData:         "Document",
}

// PlantUMLExporter renders a DomainStory as a PlantUML DSL string.
type PlantUMLExporter struct{}

// Render converts a DomainStory into PlantUML DSL format.
func (e *PlantUMLExporter) Render(_ context.Context, story *discoverydomain.DomainStory) (string, error) {
	var b strings.Builder

	// Header.
	b.WriteString("@startuml\n")
	fmt.Fprintf(&b, "' %s\n", story.Title())
	b.WriteString("' DomainStory-PlantUML v0.3.1 paper validation\n")
	b.WriteString("' Reference: github.com/johthor/DomainStory-PlantUML\n")
	b.WriteString("\n")
	b.WriteString("!include <DomainStory/domainStory>\n")

	// Actors section.
	b.WriteString("\n' --- Actors ---\n")

	for _, actor := range story.Actors() {
		macro, ok := actorMacros[actor.Type()]
		if !ok {
			return "", fmt.Errorf("unknown actor type %q: %w", actor.Type(), domainerrors.ErrInvariantViolation)
		}

		identifier := toIdentifier(actor.Name())

		if needsLabel(actor.Name()) {
			fmt.Fprintf(&b, "%s(%s, \"%s\")\n", macro, identifier, addSpaces(actor.Name()))
		} else {
			fmt.Fprintf(&b, "%s(%s)\n", macro, identifier)
		}
	}

	// Work Objects section.
	b.WriteString("\n' --- Work Objects ---\n")

	for _, wo := range story.WorkObjects() {
		macro, ok := workObjectMacros[wo.Type()]
		if !ok {
			return "", fmt.Errorf("unknown work object type %q: %w", wo.Type(), domainerrors.ErrInvariantViolation)
		}

		identifier := toIdentifier(wo.Name())

		if needsLabel(wo.Name()) {
			fmt.Fprintf(&b, "%s(%s, \"%s\")\n", macro, identifier, addSpaces(wo.Name()))
		} else {
			fmt.Fprintf(&b, "%s(%s)\n", macro, identifier)
		}

		if wo.Type() == discoverydomain.WorkObjectTypeData {
			b.WriteString("' mapped from data\n")
		}
	}

	// Sentences section.
	b.WriteString("\n' --- Sentences ---\n")

	for _, s := range story.Sentences() {
		subjectVar := toIdentifier(s.Subject())
		objectVar := toIdentifier(s.Object())

		if s.HasIndirectObject() {
			indirectRef := formatIndirectObject(story, s.IndirectObject())
			fmt.Fprintf(&b, "activity(%d, %s, %s, %s, %s, %s)\n",
				s.Step(), subjectVar, s.Activity(), objectVar, s.Preposition(), indirectRef)
		} else {
			fmt.Fprintf(&b, "activity(%d, %s, %s, %s)\n",
				s.Step(), subjectVar, s.Activity(), objectVar)
		}
	}

	b.WriteString("\n@enduml\n")

	return b.String(), nil
}

// toIdentifier removes spaces from a name to create a PascalCase identifier.
func toIdentifier(name string) string {
	return strings.ReplaceAll(name, " ", "")
}

// needsLabel returns true when a name should have a quoted display label.
// Names with spaces always need labels. PascalCase names need labels only when
// all segments are 3+ characters (real compound words like "PetOwner" → "Pet Owner"),
// not prefix-style identifiers like "MyDocument".
func needsLabel(name string) bool {
	if strings.Contains(name, " ") {
		return true
	}

	words := splitPascalCase(name)
	if len(words) < 2 {
		return false
	}

	for _, w := range words {
		if len([]rune(w)) < 3 {
			return false
		}
	}

	return true
}

// addSpaces converts PascalCase or space-containing names to display form.
// "PetOwner" → "Pet Owner", "Shopping Cart" → "Shopping Cart".
func addSpaces(name string) string {
	if strings.Contains(name, " ") {
		return name
	}

	return strings.Join(splitPascalCase(name), " ")
}

// splitPascalCase splits a PascalCase name into its constituent words.
// "PaymentGateway" → ["Payment", "Gateway"], "Customer" → ["Customer"].
func splitPascalCase(name string) []string {
	var words []string

	runes := []rune(name)
	start := 0

	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}

	words = append(words, string(runes[start:]))

	return words
}

// formatIndirectObject returns the identifier (unquoted) if the name matches a known
// actor or work object, otherwise returns it quoted.
func formatIndirectObject(story *discoverydomain.DomainStory, name string) string {
	if isKnownName(story, name) {
		return toIdentifier(name)
	}

	return fmt.Sprintf("%q", name)
}

// isKnownName checks if a name matches any actor or work object in the story (case-insensitive).
func isKnownName(story *discoverydomain.DomainStory, name string) bool {
	for _, a := range story.Actors() {
		if strings.EqualFold(a.Name(), name) {
			return true
		}
	}

	for _, wo := range story.WorkObjects() {
		if strings.EqualFold(wo.Name(), name) {
			return true
		}
	}

	return false
}
