// Package infrastructure provides adapters for the Discovery bounded context.
package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
)

var approachMap = map[string]string{
	"core":       "Rich Domain Model (DDD tactical)",
	"supporting": "Simpler architecture (Active Record)",
	"generic":    "Buy/use existing (CRUD, library)",
}

// MarkdownArtifactRenderer renders a finalized DomainModel into PRD, DDD, and Architecture markdown.
type MarkdownArtifactRenderer struct{}

// NewMarkdownArtifactRenderer creates a new MarkdownArtifactRenderer.
func NewMarkdownArtifactRenderer() *MarkdownArtifactRenderer {
	return &MarkdownArtifactRenderer{}
}

func escapeTableCell(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "*", "\\*")
	return text
}

// RenderPRD renders the PRD markdown from a domain model.
func (r *MarkdownArtifactRenderer) RenderPRD(_ context.Context, model *ddd.DomainModel) (string, error) {
	var parts []string
	parts = append(parts, "---", "last_reviewed: YYYY-MM-DD", "owner: product", "status: draft", "---", "")
	parts = append(parts, "# Product Requirements Document", "")

	// Intro
	parts = append(parts, "## 1. Problem Statement", "", "> TODO: Describe the core pain point.", "")
	parts = append(parts, "## 2. Vision", "", "> TODO: Describe the desired end state.", "")

	// Personas
	parts = append(parts, "## 3. Users & Personas", "",
		"| Persona | Description | Primary Need |",
		"|---------|-------------|-------------|")
	seen := make(map[string]bool)
	for _, story := range model.DomainStories() {
		for _, actor := range story.Actors() {
			if !seen[actor] {
				seen[actor] = true
				parts = append(parts, fmt.Sprintf("| %s | Domain actor | TODO |", escapeTableCell(actor)))
			}
		}
	}
	parts = append(parts, "")

	// Scenarios
	parts = append(parts, "## 4. User Scenarios", "")
	for _, story := range model.DomainStories() {
		parts = append(parts, fmt.Sprintf("### Scenario: %s", story.Name()), "")
		parts = append(parts, fmt.Sprintf("**Trigger:** %s", story.Trigger()), "")
		parts = append(parts, "**Flow:**")
		for i, step := range story.Steps() {
			parts = append(parts, fmt.Sprintf("%d. %s", i+1, step))
		}
		parts = append(parts, "")
	}

	// Capabilities
	parts = append(parts, "## 5. Capabilities", "", "### Must Have (P0)", "")
	for _, ctx := range model.BoundedContexts() {
		parts = append(parts, fmt.Sprintf("- [ ] %s -- %s",
			escapeTableCell(ctx.Name()), escapeTableCell(ctx.Responsibility())))
	}
	parts = append(parts, "", "### Should Have (P1)", "", "- [ ] TODO", "",
		"### Nice to Have (P2)", "", "- [ ] TODO", "")

	// Constraints & tail
	parts = append(parts,
		"## 6. Constraints", "", "### Technical Constraints", "",
		"| Constraint | Value | Rationale |",
		"|-----------|-------|-----------|",
		"| Language | Python 3.12+ | Team expertise |",
		"| Package manager | uv | Speed, reproducibility |", "",
		"## 7. Out of Scope", "", "- TODO", "",
		"## 8. Success Metrics", "",
		"| Metric | Target | Measurement Method |",
		"|--------|--------|-------------------|",
		"| TODO | TODO | TODO |", "",
		"## 9. Risks & Unknowns", "",
		"| Risk | Likelihood | Impact | Mitigation |",
		"|------|-----------|--------|------------|",
		"| TODO | Medium | Medium | TODO |", "")

	return strings.Join(parts, "\n"), nil
}

// RenderDDD renders the DDD.md markdown from a domain model.
func (r *MarkdownArtifactRenderer) RenderDDD(_ context.Context, model *ddd.DomainModel) (string, error) {
	var parts []string
	parts = append(parts, "---", "last_reviewed: YYYY-MM-DD", "owner: architecture", "status: draft", "---", "")
	parts = append(parts, "# Domain-Driven Design Artifacts", "")

	// Stories
	parts = append(parts, "## 1. Domain Stories", "")
	for _, story := range model.DomainStories() {
		parts = append(parts, fmt.Sprintf("### Story: %s", story.Name()), "")
		parts = append(parts, fmt.Sprintf("**Actors:** %s", strings.Join(story.Actors(), ", ")))
		parts = append(parts, fmt.Sprintf("**Trigger:** %s", story.Trigger()), "")
		parts = append(parts, "```")
		for i, step := range story.Steps() {
			parts = append(parts, fmt.Sprintf("%d. %s", i+1, step))
		}
		parts = append(parts, "```", "")
		if len(story.Observations()) > 0 {
			parts = append(parts, "**Key observations:**")
			for _, obs := range story.Observations() {
				parts = append(parts, fmt.Sprintf("- %s", obs))
			}
			parts = append(parts, "")
		}
	}

	// Glossary
	parts = append(parts, "## 2. Ubiquitous Language Glossary", "")
	terms := model.UbiquitousLanguage().Terms()
	if len(terms) > 0 {
		parts = append(parts,
			"| Term | Definition | Context / Bounded Context |",
			"|------|-----------|---------------------------|")
		for _, entry := range terms {
			parts = append(parts, fmt.Sprintf("| %s | %s | %s |",
				escapeTableCell(entry.Term()),
				escapeTableCell(entry.Definition()),
				escapeTableCell(entry.ContextName())))
		}
		parts = append(parts, "")
	} else {
		parts = append(parts, "_No terms defined yet._", "")
	}

	// Classifications
	parts = append(parts, "## 3. Subdomain Classification", "",
		"| Subdomain | Type | Rationale | Architecture Approach |",
		"|-----------|------|-----------|----------------------|")
	for _, ctx := range model.BoundedContexts() {
		classification := "unclassified"
		if ctx.Classification() != nil {
			classification = string(*ctx.Classification())
		}
		rationale := ctx.ClassificationRationale()
		if rationale == "" {
			rationale = "--"
		}
		approach := approachMap[classification]
		if approach == "" {
			approach = "TBD"
		}
		parts = append(parts, fmt.Sprintf("| %s | **%s** | %s | %s |",
			escapeTableCell(ctx.Name()),
			escapeTableCell(titleCase(classification)),
			escapeTableCell(rationale),
			approach))
	}
	parts = append(parts, "")

	// Bounded Contexts
	parts = append(parts, "## 4. Bounded Contexts", "")
	for _, ctx := range model.BoundedContexts() {
		parts = append(parts, fmt.Sprintf("### Context: %s", ctx.Name()), "")
		parts = append(parts, fmt.Sprintf("**Responsibility:** %s", ctx.Responsibility()), "")
		if len(ctx.KeyDomainObjects()) > 0 {
			parts = append(parts, "**Key domain objects:**")
			for _, obj := range ctx.KeyDomainObjects() {
				parts = append(parts, fmt.Sprintf("- %s", obj))
			}
			parts = append(parts, "")
		}
	}

	// Context relationships
	rels := model.ContextRelationships()
	if len(rels) > 0 {
		parts = append(parts, "### Context Map (Relationships)", "",
			"| Upstream Context | Downstream Context | Integration Pattern |",
			"|-----------------|-------------------|-------------------|")
		for _, rel := range rels {
			parts = append(parts, fmt.Sprintf("| %s | %s | %s |",
				escapeTableCell(rel.Upstream()),
				escapeTableCell(rel.Downstream()),
				escapeTableCell(rel.IntegrationPattern())))
		}
		parts = append(parts, "")
	}

	// Aggregates
	parts = append(parts, "## 5. Aggregate Design", "")
	aggs := model.AggregateDesigns()
	if len(aggs) > 0 {
		for _, agg := range aggs {
			parts = append(parts, fmt.Sprintf("### Aggregate: %s (in %s)", agg.Name(), agg.ContextName()), "")
			parts = append(parts, fmt.Sprintf("**Aggregate Root:** %s", agg.RootEntity()), "")
			if len(agg.ContainedObjects()) > 0 {
				parts = append(parts, "**Contains:**")
				for _, obj := range agg.ContainedObjects() {
					parts = append(parts, fmt.Sprintf("- %s", obj))
				}
				parts = append(parts, "")
			}
			if len(agg.Invariants()) > 0 {
				parts = append(parts, "**Invariants:**")
				for i, inv := range agg.Invariants() {
					parts = append(parts, fmt.Sprintf("%d. %s", i+1, inv))
				}
				parts = append(parts, "")
			}
			if len(agg.Commands()) > 0 {
				parts = append(parts, "**Commands:**")
				for _, cmd := range agg.Commands() {
					parts = append(parts, fmt.Sprintf("- `%s`", cmd))
				}
				parts = append(parts, "")
			}
			if len(agg.DomainEvents()) > 0 {
				parts = append(parts, "**Domain Events:**")
				for _, evt := range agg.DomainEvents() {
					parts = append(parts, fmt.Sprintf("- `%s`", evt))
				}
				parts = append(parts, "")
			}
		}
	} else {
		parts = append(parts, "_No aggregate designs yet._", "")
	}

	return strings.Join(parts, "\n"), nil
}

// RenderArchitecture renders the ARCHITECTURE.md markdown from a domain model.
func (r *MarkdownArtifactRenderer) RenderArchitecture(_ context.Context, model *ddd.DomainModel) (string, error) {
	var parts []string
	parts = append(parts, "---", "last_reviewed: YYYY-MM-DD", "owner: architecture", "status: draft", "---", "")
	parts = append(parts, "# Architecture", "")

	// Design Principles
	parts = append(parts, "## 1. Design Principles", "",
		"1. **Domain purity** -- domain layer has zero external dependencies",
		"2. **DDD alignment** -- architecture follows bounded context boundaries from `docs/DDD.md`",
		"3. **Testability** -- every component testable in isolation with dependency injection", "")

	// System Overview
	parts = append(parts, "## 2. System Overview", "", "### Component Summary", "",
		"| Component | Responsibility | Bounded Context |",
		"| ------------- | -------------- | ------------------------------- |")
	for _, ctx := range model.BoundedContexts() {
		parts = append(parts, fmt.Sprintf("| %s | %s | %s |",
			escapeTableCell(ctx.Name()),
			escapeTableCell(ctx.Responsibility()),
			escapeTableCell(ctx.Name())))
	}
	parts = append(parts, "")

	// Layer Architecture
	parts = append(parts, "## 3. Layer Architecture", "",
		"Following Hexagonal / Clean Architecture aligned with DDD:", "",
		"### Layer Rules", "",
		"| Layer | Can Depend On | Cannot Depend On |",
		"| -------------- | -------------------------- | --------------------------------------- |",
		"| Domain | Nothing (pure Python) | Application, Infrastructure, frameworks |",
		"| Application | Domain, Ports (interfaces) | Infrastructure, frameworks |",
		"| Infrastructure | Application, Domain | -- (outermost layer) |", "",
		"### Source Layout", "",
		"```",
		"src/",
		"├── domain/",
		"│   ├── models/          # Entities, Value Objects, Aggregates",
		"│   ├── services/        # Domain Services",
		"│   └── events/          # Domain Events",
		"├── application/",
		"│   ├── commands/        # Command handlers (write operations)",
		"│   ├── queries/         # Query handlers (read operations)",
		"│   └── ports/           # Interfaces (Protocols) for infrastructure",
		"└── infrastructure/",
		"    ├── persistence/     # Database adapters",
		"    ├── messaging/       # Message bus adapters",
		"    └── external/        # External API clients",
		"```", "")

	// Bounded Context Integration
	parts = append(parts, "## 4. Bounded Context Integration", "",
		"### Subdomain Classification", "",
		"| Subdomain | Type | Rationale |",
		"|-----------|------|-----------|")
	for _, ctx := range model.BoundedContexts() {
		classification := "unclassified"
		if ctx.Classification() != nil {
			classification = string(*ctx.Classification())
		}
		rationale := ctx.ClassificationRationale()
		if rationale == "" {
			rationale = "--"
		}
		parts = append(parts, fmt.Sprintf("| %s | **%s** | %s |",
			escapeTableCell(ctx.Name()),
			escapeTableCell(classification),
			escapeTableCell(rationale)))
	}
	parts = append(parts, "")

	// Data Model
	parts = append(parts, "## 5. Data Model", "", "### Aggregates and Storage", "")
	aggs := model.AggregateDesigns()
	if len(aggs) > 0 {
		parts = append(parts,
			"| Aggregate | Context | Root Entity |",
			"| ------------- | -------------------------- | ---------------- |")
		for _, agg := range aggs {
			parts = append(parts, fmt.Sprintf("| %s | %s | %s |",
				escapeTableCell(agg.Name()),
				escapeTableCell(agg.ContextName()),
				escapeTableCell(agg.RootEntity())))
		}
		parts = append(parts, "")
	} else {
		parts = append(parts, "_No aggregate designs yet._", "")
	}

	return strings.Join(parts, "\n"), nil
}

// titleCase capitalises the first rune of s without using the deprecated strings.Title.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
