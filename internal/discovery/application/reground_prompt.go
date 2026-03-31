package application

import (
	"fmt"
	"strings"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// RegroundingContext captures the session state needed to re-ground a prompt.
// It is an application-layer value object — no I/O, no domain dependencies beyond
// reading session properties.
type RegroundingContext struct {
	Mode    string // lowercase mode string from DiscoveryMode (e.g. "rapid", "thorough")
	Persona string // display-ready persona label (e.g. "Developer", "Domain Expert"); empty when unset
}

// personaDisplayNames maps domain DiscoveryPersona constants to user-facing labels.
// Labels match the Choice.Label values in cli_discovery_adapter.go personaChoices.
var personaDisplayNames = map[discoverydomain.DiscoveryPersona]string{
	discoverydomain.PersonaDeveloper:    "Developer",
	discoverydomain.PersonaProductOwner: "Product Owner",
	discoverydomain.PersonaDomainExpert: "Domain Expert",
	discoverydomain.PersonaMixed:        "Mixed Team",
}

// NewRegroundingContext builds a RegroundingContext from a live DiscoverySession.
// Returns zero-value context for nil session (safe for agent mode / tests).
func NewRegroundingContext(session *discoverydomain.DiscoverySession) RegroundingContext {
	if session == nil {
		return RegroundingContext{}
	}

	ctx := RegroundingContext{
		Mode: string(session.Mode()),
	}

	if persona, ok := session.Persona(); ok {
		if displayName, found := personaDisplayNames[persona]; found {
			ctx.Persona = displayName
		} else {
			// Fallback: capitalise the raw persona string.
			ctx.Persona = capitalise(string(persona))
		}
	}

	return ctx
}

// BuildRegroundingPrompt prepends a context header to basePrompt.
// Returns basePrompt unchanged when ctx has no Mode set (zero-value or persona-only).
//
// Output format examples:
//
//	[Mode: Rapid]
//	[Mode: Thorough | Persona: Developer]
func BuildRegroundingPrompt(ctx RegroundingContext, basePrompt string) string {
	if ctx.Mode == "" {
		return basePrompt
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Mode: %s", capitalise(ctx.Mode)))

	if ctx.Persona != "" {
		parts = append(parts, fmt.Sprintf("Persona: %s", ctx.Persona))
	}

	header := "[" + strings.Join(parts, " | ") + "]"

	return header + "\n" + basePrompt
}

// capitalise uppercases the first rune of s, leaving the rest unchanged.
func capitalise(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
