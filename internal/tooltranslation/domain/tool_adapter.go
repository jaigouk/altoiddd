package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Shared constants
// ---------------------------------------------------------------------------

const afterCloseProtocolText = `## After-Close Protocol

After every ` + "`bd close <id>`" + `, run these steps:

1. **Ripple review** -- ` + "`bin/bd-ripple <id> \"<what this ticket produced>\"`" + `
2. **Review flagged tickets** -- ` + "`bd query label=review_needed`" + `, read ripple comments,
   draft updates, present to user for approval
3. **Follow-up tickets** -- create using beads templates, set dependencies
4. **Groom next ticket** -- ` + "`bd ready`" + `, run grooming checklist on top pick
`

// agentPersonaTemplate is the YAML frontmatter + body template for Claude Code agents.
const agentPersonaTemplate = `---
name: %s
description: >
  %s
tools: %s
model: opus
permissionMode: %s
memory: project
---

%s
`

type agentPersonaDef struct {
	slug        string
	description string
	tools       string
	permission  string
	bodyBuilder func(model *ddd.DomainModel, profile vo.StackProfile) string
}

var defaultAgentPersonas = []agentPersonaDef{
	{
		slug:        "developer",
		description: "Implementation-focused developer agent for writing code, fixing bugs, and implementing features following Red/Green/Refactor.",
		tools:       "Read, Edit, Write, Grep, Glob, Bash",
		permission:  "acceptEdits",
		bodyBuilder: buildDeveloperBody,
	},
	{
		slug:        "tech-lead",
		description: "Technical lead for architecture review, DDD/SOLID compliance, and code quality gate enforcement.",
		tools:       "Read, Grep, Glob, Bash, Write, Edit",
		permission:  "default",
		bodyBuilder: buildTechLeadBody,
	},
	{
		slug:        "qa-engineer",
		description: "QA engineer for writing tests, validating coverage, and investigating failures.",
		tools:       "Read, Edit, Write, Grep, Glob, Bash",
		permission:  "acceptEdits",
		bodyBuilder: buildQAEngineerBody,
	},
	{
		slug:        "researcher",
		description: "Research agent for spike tickets, ADRs, and library evaluation.",
		tools:       "Read, Write, Edit, Grep, Glob, Bash, WebSearch, WebFetch",
		permission:  "default",
		bodyBuilder: buildResearcherBody,
	},
}

// ---------------------------------------------------------------------------
// Shared content builders
// ---------------------------------------------------------------------------

func buildUbiquitousLanguageSection(model *ddd.DomainModel) string {
	var b strings.Builder
	b.WriteString("## Ubiquitous Language\n\n")
	b.WriteString("| Term | Definition | Context |\n")
	b.WriteString("|------|-----------|---------|")
	for _, entry := range model.UbiquitousLanguage().Terms() {
		fmt.Fprintf(&b, "\n| %s | %s | %s |", entry.Term(), entry.Definition(), entry.ContextName())
	}
	b.WriteString("\n")
	return b.String()
}

func buildBoundedContextSection(model *ddd.DomainModel) string {
	var b strings.Builder
	b.WriteString("## Bounded Contexts\n\n")
	for _, ctx := range model.BoundedContexts() {
		classification := "unclassified"
		if ctx.Classification() != nil {
			classification = string(*ctx.Classification())
		}
		fmt.Fprintf(&b, "- **%s** (%s): %s\n", ctx.Name(), classification, ctx.Responsibility())
	}
	return b.String()
}

func buildDDDLayerRules() string {
	return `## DDD Layer Rules

- ` + "`domain/`" + ` has ZERO external dependencies (no frameworks, no DB, no HTTP)
- ` + "`application/`" + ` depends on ` + "`domain/`" + ` and ` + "`ports/`" + ` (interfaces only)
- ` + "`infrastructure/`" + ` implements ` + "`ports/`" + ` and depends on external libraries
- Dependencies flow inward: infrastructure -> application -> domain
`
}

func buildQualityGates(profile vo.StackProfile) string {
	return profile.QualityGateDisplay()
}

func buildAgentsMD(model *ddd.DomainModel, profile vo.StackProfile) string {
	var b strings.Builder
	b.WriteString("# Project Conventions\n\n")
	b.WriteString(buildUbiquitousLanguageSection(model))
	b.WriteString("\n")
	b.WriteString(buildBoundedContextSection(model))
	b.WriteString("\n")
	b.WriteString(buildDDDLayerRules())
	b.WriteString("\n")
	b.WriteString(afterCloseProtocolText)
	gates := buildQualityGates(profile)
	if gates != "" {
		b.WriteString("\n")
		b.WriteString(gates)
	}
	return b.String()
}

func buildMemoryMD(model *ddd.DomainModel, profile vo.StackProfile) string {
	var b strings.Builder
	b.WriteString("# Project Memory\n\n")
	b.WriteString(buildMemoryBeadsWorkflow())
	b.WriteString("\n")
	b.WriteString(afterCloseProtocolText)
	b.WriteString("\n")
	b.WriteString(buildMemoryGroomingChecklist())
	b.WriteString("\n")
	b.WriteString(buildMemoryBoundedContexts(model))
	b.WriteString("\n")
	b.WriteString(buildMemoryUbiquitousLanguage(model))
	gates := buildQualityGates(profile)
	if gates != "" {
		b.WriteString("\n")
		b.WriteString(gates)
	}
	return b.String()
}

func buildMemoryBeadsWorkflow() string {
	return "## Beads Workflow\n\n```bash\n" +
		"bd ready                         # Find available work\n" +
		"bd show <id>                     # View ticket details\n" +
		"bd update <id> --status in_progress  # Claim a ticket\n" +
		"bd close <id>                    # Close completed ticket\n" +
		"bin/bd-ripple <id> \"<summary>\"   # Flag dependents (ripple review)\n" +
		"bd query label=review_needed     # See tickets needing review\n" +
		"bd label remove <id> review_needed   # Clear flag after review\n" +
		"```\n"
}

func buildMemoryGroomingChecklist() string {
	return `## Grooming Checklist

Before claiming a ticket:

1. **Template compliance** -- description follows beads template
2. **Freshness check** -- ` + "`bd label list <id>`" + ` for ` + "`review_needed`" + `
3. **PRD traceability** -- ` + "`/prd-traceability <id>`" + ` to verify capability coverage
4. **DDD alignment** -- bounded context boundaries respected
5. **Ubiquitous language** -- terms match DDD.md glossary
6. **TDD & SOLID** -- RED/GREEN/REFACTOR phases documented
7. **Acceptance criteria** -- testable checkboxes, edge cases, coverage >= 80%
`
}

func buildMemoryBoundedContexts(model *ddd.DomainModel) string {
	var b strings.Builder
	b.WriteString("## Bounded Contexts\n\n")
	for _, ctx := range model.BoundedContexts() {
		classification := "unclassified"
		if ctx.Classification() != nil {
			classification = string(*ctx.Classification())
		}
		fmt.Fprintf(&b, "- **%s** (%s): %s\n", ctx.Name(), classification, ctx.Responsibility())
	}
	return b.String()
}

func buildMemoryUbiquitousLanguage(model *ddd.DomainModel) string {
	var b strings.Builder
	b.WriteString("## Ubiquitous Language\n\n")
	b.WriteString("| Term | Definition | Context |\n")
	b.WriteString("|------|-----------|---------|")
	for _, entry := range model.UbiquitousLanguage().Terms() {
		fmt.Fprintf(&b, "\n| %s | %s | %s |", entry.Term(), entry.Definition(), entry.ContextName())
	}
	b.WriteString("\n")
	return b.String()
}

// ---------------------------------------------------------------------------
// Agent persona body builders
// ---------------------------------------------------------------------------

func buildDeveloperBody(model *ddd.DomainModel, profile vo.StackProfile) string {
	var b strings.Builder
	b.WriteString("You are a **Developer** on this project.\n\n")
	b.WriteString("## Primary Responsibilities\n\n")
	b.WriteString("1. Implement features and fix bugs via beads tickets\n")
	b.WriteString("2. Follow Red/Green/Refactor strictly\n")
	b.WriteString("3. Respect bounded context boundaries\n\n")
	b.WriteString(buildBoundedContextSection(model))
	b.WriteString("\n")
	b.WriteString(buildUbiquitousLanguageSection(model))
	b.WriteString("\n")
	b.WriteString(buildDDDLayerRules())
	gates := buildQualityGates(profile)
	if gates != "" {
		b.WriteString("\n")
		b.WriteString(gates)
	}
	return b.String()
}

func buildTechLeadBody(model *ddd.DomainModel, profile vo.StackProfile) string {
	var b strings.Builder
	b.WriteString("You are a **Tech Lead** on this project.\n\n")
	b.WriteString("## Primary Responsibilities\n\n")
	b.WriteString("1. Review architecture for DDD/SOLID compliance\n")
	b.WriteString("2. Enforce bounded context boundaries\n")
	b.WriteString("3. Run quality gates before approving changes\n\n")
	b.WriteString(buildBoundedContextSection(model))
	b.WriteString("\n")
	b.WriteString(buildUbiquitousLanguageSection(model))
	b.WriteString("\n")
	b.WriteString(buildDDDLayerRules())
	gates := buildQualityGates(profile)
	if gates != "" {
		b.WriteString("\n")
		b.WriteString(gates)
	}
	return b.String()
}

func buildQAEngineerBody(model *ddd.DomainModel, profile vo.StackProfile) string {
	var b strings.Builder
	b.WriteString("You are a **QA Engineer** on this project.\n\n")
	b.WriteString("## Primary Responsibilities\n\n")
	b.WriteString("1. Write and run tests, validate coverage\n")
	b.WriteString("2. Investigate failures and produce root cause analysis\n")
	b.WriteString("3. Verify edge cases from multiple angles\n\n")
	b.WriteString(buildBoundedContextSection(model))
	b.WriteString("\n")
	b.WriteString(buildUbiquitousLanguageSection(model))
	b.WriteString("\n")
	b.WriteString(buildDDDLayerRules())
	gates := buildQualityGates(profile)
	if gates != "" {
		b.WriteString("\n")
		b.WriteString(gates)
	}
	return b.String()
}

func buildResearcherBody(model *ddd.DomainModel, profile vo.StackProfile) string {
	var b strings.Builder
	b.WriteString("You are a **Researcher** on this project.\n\n")
	b.WriteString("## Primary Responsibilities\n\n")
	b.WriteString("1. Evaluate libraries and tools for spike tickets\n")
	b.WriteString("2. Write ADRs and research reports\n")
	b.WriteString("3. Provide concrete facts before architectural decisions\n\n")
	b.WriteString(buildBoundedContextSection(model))
	b.WriteString("\n")
	b.WriteString(buildUbiquitousLanguageSection(model))
	return b.String()
}

func buildAgentPersonas(model *ddd.DomainModel, profile vo.StackProfile) []ConfigSection {
	var sections []ConfigSection
	for _, def := range defaultAgentPersonas {
		body := def.bodyBuilder(model, profile)
		content := fmt.Sprintf(agentPersonaTemplate,
			def.slug,
			def.description,
			def.tools,
			def.permission,
			body,
		)
		path := fmt.Sprintf(".claude/agents/%s.md", def.slug)
		sections = append(sections, NewConfigSection(path, content, "Agent persona"))
	}
	return sections
}

// ---------------------------------------------------------------------------
// ClaudeCodeAdapter
// ---------------------------------------------------------------------------

// ClaudeCodeAdapter generates .claude/CLAUDE.md and .claude/memory/MEMORY.md.
type ClaudeCodeAdapter struct{}

// NewClaudeCodeAdapter creates a ClaudeCodeAdapter.
func NewClaudeCodeAdapter() *ClaudeCodeAdapter { return &ClaudeCodeAdapter{} }

// Translate implements ToolAdapter.
func (a *ClaudeCodeAdapter) Translate(model *ddd.DomainModel, profile vo.StackProfile) []ConfigSection {
	var b strings.Builder
	b.WriteString("# CLAUDE.md\n\n")
	b.WriteString(buildUbiquitousLanguageSection(model))
	b.WriteString("\n")
	b.WriteString(buildBoundedContextSection(model))
	b.WriteString("\n")
	b.WriteString(buildDDDLayerRules())
	b.WriteString("\n")
	b.WriteString(afterCloseProtocolText)
	gates := buildQualityGates(profile)
	if gates != "" {
		b.WriteString("\n")
		b.WriteString(gates)
	}

	memoryContent := buildMemoryMD(model, profile)

	base := []ConfigSection{
		NewConfigSection(".claude/CLAUDE.md", b.String(), "Claude Code config"),
		NewConfigSection(".claude/memory/MEMORY.md", memoryContent, "Claude Code memory"),
	}
	return append(base, buildAgentPersonas(model, profile)...)
}

// ---------------------------------------------------------------------------
// CursorAdapter
// ---------------------------------------------------------------------------

func buildCursorDomainExpertRule(model *ddd.DomainModel) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("description: Domain terminology and bounded context reference\n")
	b.WriteString("globs: **/*\n")
	b.WriteString("alwaysApply: true\n")
	b.WriteString("---\n\n")
	b.WriteString("# Domain Expert Reference\n\n")
	b.WriteString("Use these terms consistently throughout the codebase.\n\n")
	b.WriteString(buildUbiquitousLanguageSection(model))
	b.WriteString("\n")
	b.WriteString(buildBoundedContextSection(model))
	return b.String()
}

// CursorAdapter generates AGENTS.md and .cursor/rules/project-conventions.mdc.
type CursorAdapter struct{}

// NewCursorAdapter creates a CursorAdapter.
func NewCursorAdapter() *CursorAdapter { return &CursorAdapter{} }

// Translate implements ToolAdapter.
func (a *CursorAdapter) Translate(model *ddd.DomainModel, profile vo.StackProfile) []ConfigSection {
	agentsContent := buildAgentsMD(model, profile)

	var mdc strings.Builder
	fmt.Fprintf(&mdc, "---\ndescription: Project conventions derived from domain model\nglobs: %s\n---\n\n", profile.FileGlob())
	mdc.WriteString(buildDDDLayerRules())
	gates := buildQualityGates(profile)
	if gates != "" {
		mdc.WriteString("\n")
		mdc.WriteString(gates)
	}

	sections := []ConfigSection{
		NewConfigSection("AGENTS.md", agentsContent, "Cursor agents"),
		NewConfigSection(".cursor/rules/project-conventions.mdc", mdc.String(), "Cursor rules"),
	}
	if len(model.UbiquitousLanguage().Terms()) > 0 {
		sections = append(sections, NewConfigSection(
			".cursor/rules/domain-expert.mdc",
			buildCursorDomainExpertRule(model),
			"Domain terminology reference",
		))
	}
	return sections
}

// ---------------------------------------------------------------------------
// RooCodeAdapter
// ---------------------------------------------------------------------------

type rooMode struct {
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	RoleDefinition string `json:"roleDefinition,omitempty"`
}

func buildRoleDefinition(model *ddd.DomainModel) string {
	var b strings.Builder
	b.WriteString("You work within these bounded contexts:\n\n")
	for _, ctx := range model.BoundedContexts() {
		classification := "unclassified"
		if ctx.Classification() != nil {
			classification = string(*ctx.Classification())
		}
		fmt.Fprintf(&b, "- %s (%s): %s\n", ctx.Name(), classification, ctx.Responsibility())
	}
	b.WriteString("\nUse domain terminology consistently.")
	return b.String()
}

func buildEnhancedRoomodes(model *ddd.DomainModel) string {
	modes := []rooMode{
		{
			Slug:           "ddd-developer",
			Name:           "DDD Developer",
			Description:    "Follows domain-driven design conventions",
			RoleDefinition: buildRoleDefinition(model),
		},
	}

	for _, ctx := range model.BoundedContexts() {
		if ctx.Classification() == nil {
			continue
		}
		classification := string(*ctx.Classification())
		modes = append(modes, rooMode{
			Slug:           fmt.Sprintf("%s-context", strings.ToLower(ctx.Name())),
			Name:           fmt.Sprintf("%s Context", ctx.Name()),
			Description:    fmt.Sprintf("Work within %s bounded context (%s)", ctx.Name(), classification),
			RoleDefinition: fmt.Sprintf("Focus on %s. %s", ctx.Name(), ctx.Responsibility()),
		})
	}

	data := map[string]interface{}{"customModes": modes}
	roomodesJSON, _ := json.MarshalIndent(data, "", "  ")
	return string(roomodesJSON)
}

// RooCodeAdapter generates AGENTS.md, .roomodes, and .roo/rules/project-conventions.md.
type RooCodeAdapter struct{}

// NewRooCodeAdapter creates a RooCodeAdapter.
func NewRooCodeAdapter() *RooCodeAdapter { return &RooCodeAdapter{} }

// Translate implements ToolAdapter.
func (a *RooCodeAdapter) Translate(model *ddd.DomainModel, profile vo.StackProfile) []ConfigSection {
	agentsContent := buildAgentsMD(model, profile)

	roomodesJSON := buildEnhancedRoomodes(model)

	var rules strings.Builder
	rules.WriteString("# Project Conventions\n\n")
	rules.WriteString(buildDDDLayerRules())
	gates := buildQualityGates(profile)
	if gates != "" {
		rules.WriteString("\n")
		rules.WriteString(gates)
	}

	return []ConfigSection{
		NewConfigSection("AGENTS.md", agentsContent, "Roo Code agents"),
		NewConfigSection(".roomodes", roomodesJSON, "Roo Code modes"),
		NewConfigSection(".roo/rules/project-conventions.md", rules.String(), "Roo Code rules"),
	}
}

// ---------------------------------------------------------------------------
// OpenCodeAdapter
// ---------------------------------------------------------------------------

// OpenCodeAdapter generates AGENTS.md, .opencode/rules/project-conventions.md, and opencode.json.
type OpenCodeAdapter struct{}

// NewOpenCodeAdapter creates an OpenCodeAdapter.
func NewOpenCodeAdapter() *OpenCodeAdapter { return &OpenCodeAdapter{} }

// Translate implements ToolAdapter.
func (a *OpenCodeAdapter) Translate(model *ddd.DomainModel, profile vo.StackProfile) []ConfigSection {
	agentsContent := buildAgentsMD(model, profile)

	var rules strings.Builder
	rules.WriteString("# Project Conventions\n\n")
	rules.WriteString(buildDDDLayerRules())
	gates := buildQualityGates(profile)
	if gates != "" {
		rules.WriteString("\n")
		rules.WriteString(gates)
	}

	opencodeData := map[string]interface{}{
		"context": map[string]interface{}{
			"include": []string{"AGENTS.md", ".opencode/rules/*.md"},
		},
	}
	opencodeJSON, _ := json.MarshalIndent(opencodeData, "", "  ")

	return []ConfigSection{
		NewConfigSection("AGENTS.md", agentsContent, "OpenCode agents"),
		NewConfigSection(".opencode/rules/project-conventions.md", rules.String(), "OpenCode rules"),
		NewConfigSection("opencode.json", string(opencodeJSON), "OpenCode config"),
	}
}

// Compile-time interface checks.
var (
	_ ToolAdapter = (*ClaudeCodeAdapter)(nil)
	_ ToolAdapter = (*CursorAdapter)(nil)
	_ ToolAdapter = (*RooCodeAdapter)(nil)
	_ ToolAdapter = (*OpenCodeAdapter)(nil)
)
