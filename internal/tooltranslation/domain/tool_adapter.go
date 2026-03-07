package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
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

	return []ConfigSection{
		NewConfigSection(".claude/CLAUDE.md", b.String(), "Claude Code config"),
		NewConfigSection(".claude/memory/MEMORY.md", memoryContent, "Claude Code memory"),
	}
}

// ---------------------------------------------------------------------------
// CursorAdapter
// ---------------------------------------------------------------------------

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

	return []ConfigSection{
		NewConfigSection("AGENTS.md", agentsContent, "Cursor agents"),
		NewConfigSection(".cursor/rules/project-conventions.mdc", mdc.String(), "Cursor rules"),
	}
}

// ---------------------------------------------------------------------------
// RooCodeAdapter
// ---------------------------------------------------------------------------

// RooCodeAdapter generates AGENTS.md, .roomodes, and .roo/rules/project-conventions.md.
type RooCodeAdapter struct{}

// NewRooCodeAdapter creates a RooCodeAdapter.
func NewRooCodeAdapter() *RooCodeAdapter { return &RooCodeAdapter{} }

// Translate implements ToolAdapter.
func (a *RooCodeAdapter) Translate(model *ddd.DomainModel, profile vo.StackProfile) []ConfigSection {
	agentsContent := buildAgentsMD(model, profile)

	roomodesData := map[string]interface{}{
		"customModes": []map[string]string{
			{
				"slug":        "ddd-developer",
				"name":        "DDD Developer",
				"description": "Follows domain-driven design conventions",
			},
		},
	}
	roomodesJSON, _ := json.MarshalIndent(roomodesData, "", "  ")

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
		NewConfigSection(".roomodes", string(roomodesJSON), "Roo Code modes"),
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
