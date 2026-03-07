package valueobjects

import (
	"fmt"
	"strings"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// PersonaType enumerates agent persona archetypes.
type PersonaType string

// Persona type constants.
const (
	PersonaSoloDeveloper  PersonaType = "solo_developer"
	PersonaTeamLead       PersonaType = "team_lead"
	PersonaAIToolSwitcher PersonaType = "ai_tool_switcher"
	PersonaProductOwner   PersonaType = "product_owner"
	PersonaDomainExpert   PersonaType = "domain_expert"
)

// AllPersonaTypes returns all valid persona type values.
func AllPersonaTypes() []PersonaType {
	return []PersonaType{
		PersonaSoloDeveloper,
		PersonaTeamLead,
		PersonaAIToolSwitcher,
		PersonaProductOwner,
		PersonaDomainExpert,
	}
}

// Register classifies personas as technical or non-technical.
type Register string

// Register constants.
const (
	RegisterTechnical    Register = "technical"
	RegisterNonTechnical Register = "non_technical"
)

// PersonaDefinition is an immutable value object describing one agent persona.
type PersonaDefinition struct {
	name                 string
	personaType          PersonaType
	register             Register
	description          string
	instructionsTemplate string
}

// NewPersonaDefinition creates a validated PersonaDefinition.
func NewPersonaDefinition(
	name string,
	personaType PersonaType,
	register Register,
	description string,
	instructionsTemplate string,
) (*PersonaDefinition, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("persona name cannot be empty: %w", domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(instructionsTemplate) == "" {
		return nil, fmt.Errorf("persona instructions template cannot be empty: %w", domainerrors.ErrInvariantViolation)
	}
	return &PersonaDefinition{
		name:                 name,
		personaType:          personaType,
		register:             register,
		description:          description,
		instructionsTemplate: instructionsTemplate,
	}, nil
}

// Name returns the persona name.
func (p *PersonaDefinition) Name() string { return p.name }

// PersonaType returns the persona archetype.
func (p *PersonaDefinition) PersonaType() PersonaType { return p.personaType }

// Register returns the communication register.
func (p *PersonaDefinition) Register() Register { return p.register }

// Description returns the persona description.
func (p *PersonaDefinition) Description() string { return p.description }

// InstructionsTemplate returns the agent instructions template.
func (p *PersonaDefinition) InstructionsTemplate() string { return p.instructionsTemplate }

// PersonaRegistry returns the canonical registry of all persona definitions.
func PersonaRegistry() map[PersonaType]*PersonaDefinition {
	solo, _ := NewPersonaDefinition(
		"Solo Developer", PersonaSoloDeveloper, RegisterTechnical,
		"Full-stack developer following DDD + TDD + SOLID principles",
		"# Solo Developer Agent\n\nYou are a solo developer agent responsible for implementing features\nusing DDD, TDD, and SOLID principles. Follow Red/Green/Refactor strictly.\nWrite failing tests first, then minimal code to pass, then refactor.",
	)
	lead, _ := NewPersonaDefinition(
		"Team Lead", PersonaTeamLead, RegisterTechnical,
		"Technical lead responsible for architecture and code quality",
		"# Team Lead Agent\n\nYou are a team lead agent responsible for architecture review\nand DDD compliance. Ensure bounded context boundaries are respected\nand code quality meets project standards.",
	)
	switcher, _ := NewPersonaDefinition(
		"AI Tool Switcher", PersonaAIToolSwitcher, RegisterTechnical,
		"Agent specialized in cross-tool configuration and migration",
		"# AI Tool Switcher Agent\n\nYou are an AI tool switcher agent responsible for cross-tool\nconfiguration mapping and config translation. Ensure consistent\nbehavior across supported AI coding tools.",
	)
	po, _ := NewPersonaDefinition(
		"Product Owner", PersonaProductOwner, RegisterNonTechnical,
		"Business-focused agent for requirements and prioritization",
		"# Product Owner Agent\n\nYou are a product owner agent focused on business requirements\nand prioritization. Use business language and ensure requirements\nare captured clearly for the development team.",
	)
	de, _ := NewPersonaDefinition(
		"Domain Expert", PersonaDomainExpert, RegisterNonTechnical,
		"Domain knowledge specialist for business rules and terminology",
		"# Domain Expert Agent\n\nYou are a domain expert agent responsible for business rules\nand domain terminology. Ensure ubiquitous language is used correctly\nand domain invariants are properly captured.",
	)

	return map[PersonaType]*PersonaDefinition{
		PersonaSoloDeveloper:  solo,
		PersonaTeamLead:       lead,
		PersonaAIToolSwitcher: switcher,
		PersonaProductOwner:   po,
		PersonaDomainExpert:   de,
	}
}

// SupportedTools returns the tuple of supported tool identifiers.
func SupportedTools() []string {
	return []string{"claude-code", "cursor", "roo-code", "opencode"}
}

// ToolTargetPaths returns the mapping of tool identifiers to file path templates.
func ToolTargetPaths() map[string]string {
	return map[string]string{
		"claude-code": ".claude/agents/{name}.md",
		"cursor":      ".cursor/rules/{name}.mdc",
		"roo-code":    ".roo-code/modes/{name}.md",
		"opencode":    ".opencode/agents/{name}.md",
	}
}
