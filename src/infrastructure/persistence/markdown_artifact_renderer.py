"""Markdown artifact renderer for the Domain Model bounded context.

Transforms a finalized DomainModel aggregate into PRD.md, DDD.md, and
ARCHITECTURE.md markdown strings following the template structures in
docs/templates/.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.domain_values import (
        AggregateDesign,
        BoundedContext,
        DomainStory,
    )
    from src.domain.models.ubiquitous_language import TermEntry


def _escape_table_cell(text: str) -> str:
    """Escape markdown-special characters inside a table cell."""
    # Backslash first so we don't double-escape our own escapes.
    return text.replace("\\", "\\\\").replace("|", "\\|").replace("*", "\\*")


_FRONTMATTER_PRODUCT = (
    "---",
    "last_reviewed: YYYY-MM-DD",
    "owner: product",
    "status: draft",
    "---",
    "",
)

_FRONTMATTER_ARCH = (
    "---",
    "last_reviewed: YYYY-MM-DD",
    "owner: architecture",
    "status: draft",
    "---",
    "",
)

_APPROACH_MAP: dict[str, str] = {
    "core": "Rich Domain Model (DDD tactical)",
    "supporting": "Simpler architecture (Active Record)",
    "generic": "Buy/use existing (CRUD, library)",
}

_SOURCE_LAYOUT = (
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
    "```",
    "",
)


class MarkdownArtifactRenderer:
    """Renders a finalized DomainModel into PRD, DDD, and Architecture markdown."""

    # ── PRD ──────────────────────────────────────────────────────────

    def render_prd(self, model: DomainModel) -> str:
        """Render the PRD markdown from a domain model."""
        parts: list[str] = list(_FRONTMATTER_PRODUCT)
        parts.append("# Product Requirements Document")
        parts.append("")
        self._render_prd_intro(parts)
        self._render_prd_personas(parts, model)
        self._render_prd_scenarios(parts, model)
        self._render_prd_capabilities(parts, model)
        self._render_prd_tail(parts)
        return "\n".join(parts)

    # ── DDD ──────────────────────────────────────────────────────────

    def render_ddd(self, model: DomainModel) -> str:
        """Render the DDD.md markdown from a domain model."""
        parts: list[str] = list(_FRONTMATTER_ARCH)
        parts.append("# Domain-Driven Design Artifacts")
        parts.append("")

        self._render_ddd_stories(parts, model)
        self._render_ddd_glossary(parts, model)
        self._render_ddd_classifications(parts, model)
        self._render_ddd_contexts(parts, model)
        self._render_ddd_relationships(parts, model)
        self._render_ddd_aggregates(parts, model)

        return "\n".join(parts)

    # ── Architecture ─────────────────────────────────────────────────

    def render_architecture(self, model: DomainModel) -> str:
        """Render the ARCHITECTURE.md markdown from a domain model."""
        parts: list[str] = list(_FRONTMATTER_ARCH)
        parts.append("# Architecture")
        parts.append("")

        self._render_arch_principles(parts)
        self._render_arch_overview(parts, model)
        self._render_arch_layers(parts)
        self._render_arch_classification(parts, model)
        self._render_arch_data_model(parts, model)

        return "\n".join(parts)

    # ── PRD private helpers ──────────────────────────────────────────

    @staticmethod
    def _render_prd_intro(parts: list[str]) -> None:
        parts.extend([
            "## 1. Problem Statement",
            "",
            "> TODO: Describe the core pain point.",
            "",
            "## 2. Vision",
            "",
            "> TODO: Describe the desired end state.",
            "",
        ])

    @staticmethod
    def _render_prd_personas(parts: list[str], model: DomainModel) -> None:
        parts.extend([
            "## 3. Users & Personas",
            "",
            "| Persona | Description | Primary Need |",
            "|---------|-------------|-------------|",
        ])
        actors_seen: set[str] = set()
        for story in model.domain_stories:
            for actor in story.actors:
                if actor not in actors_seen:
                    actors_seen.add(actor)
                    parts.append(
                        f"| {_escape_table_cell(actor)} "
                        f"| Domain actor | TODO |"
                    )
        parts.append("")

    @staticmethod
    def _render_prd_scenarios(parts: list[str], model: DomainModel) -> None:
        parts.extend(["## 4. User Scenarios", ""])
        for story in model.domain_stories:
            parts.extend([
                f"### Scenario: {story.name}",
                "",
                f"**Trigger:** {story.trigger}",
                "",
                "**Flow:**",
            ])
            parts.extend(
                f"{i}. {step}" for i, step in enumerate(story.steps, 1)
            )
            parts.append("")

    @staticmethod
    def _render_prd_capabilities(parts: list[str], model: DomainModel) -> None:
        parts.extend([
            "## 5. Capabilities",
            "",
            "### Must Have (P0)",
            "",
        ])
        parts.extend(
            f"- [ ] {_escape_table_cell(ctx.name)} "
            f"— {_escape_table_cell(ctx.responsibility)}"
            for ctx in model.bounded_contexts
        )
        parts.extend([
            "",
            "### Should Have (P1)",
            "",
            "- [ ] TODO",
            "",
            "### Nice to Have (P2)",
            "",
            "- [ ] TODO",
            "",
        ])

    @staticmethod
    def _render_prd_tail(parts: list[str]) -> None:
        parts.extend([
            "## 6. Constraints",
            "",
            "### Technical Constraints",
            "",
            "| Constraint | Value | Rationale |",
            "|-----------|-------|-----------|",
            "| Language | Python 3.12+ | Team expertise |",
            "| Package manager | uv | Speed, reproducibility |",
            "",
            "## 7. Out of Scope",
            "",
            "- TODO",
            "",
            "## 8. Success Metrics",
            "",
            "| Metric | Target | Measurement Method |",
            "|--------|--------|-------------------|",
            "| TODO | TODO | TODO |",
            "",
            "## 9. Risks & Unknowns",
            "",
            "| Risk | Likelihood | Impact | Mitigation |",
            "|------|-----------|--------|------------|",
            "| TODO | Medium | Medium | TODO |",
            "",
        ])

    # ── DDD private helpers ──────────────────────────────────────────

    @staticmethod
    def _render_story(parts: list[str], story: DomainStory) -> None:
        """Render a single domain story section."""
        parts.extend([
            f"### Story: {story.name}",
            "",
            f"**Actors:** {', '.join(story.actors)}",
            f"**Trigger:** {story.trigger}",
            "",
            "```",
        ])
        parts.extend(
            f"{i}. {step}" for i, step in enumerate(story.steps, 1)
        )
        parts.extend(["```", ""])
        if story.observations:
            parts.append("**Key observations:**")
            parts.extend(f"- {obs}" for obs in story.observations)
            parts.append("")

    @staticmethod
    def _render_term_row(parts: list[str], entry: TermEntry) -> None:
        """Render a single ubiquitous language table row."""
        parts.append(
            f"| {_escape_table_cell(entry.term)} "
            f"| {_escape_table_cell(entry.definition)} "
            f"| {_escape_table_cell(entry.context_name)} |"
        )

    @staticmethod
    def _render_classification_row(parts: list[str], ctx: BoundedContext) -> None:
        """Render a single subdomain classification table row."""
        classification = ctx.classification.value if ctx.classification else "unclassified"
        rationale = ctx.classification_rationale or "—"
        approach = _APPROACH_MAP.get(classification, "TBD")
        parts.append(
            f"| {_escape_table_cell(ctx.name)} "
            f"| **{_escape_table_cell(classification.capitalize())}** "
            f"| {_escape_table_cell(rationale)} "
            f"| {approach} |"
        )

    @staticmethod
    def _render_context_section(parts: list[str], ctx: BoundedContext) -> None:
        """Render a single bounded context section."""
        parts.extend([
            f"### Context: {ctx.name}",
            "",
            f"**Responsibility:** {ctx.responsibility}",
            "",
        ])
        if ctx.key_domain_objects:
            parts.append("**Key domain objects:**")
            parts.extend(f"- {obj}" for obj in ctx.key_domain_objects)
            parts.append("")

    @staticmethod
    def _render_aggregate_section(parts: list[str], agg: AggregateDesign) -> None:
        """Render a single aggregate design section."""
        parts.extend([
            f"### Aggregate: {agg.name} (in {agg.context_name})",
            "",
            f"**Aggregate Root:** {agg.root_entity}",
            "",
        ])
        if agg.contained_objects:
            parts.append("**Contains:**")
            parts.extend(f"- {obj}" for obj in agg.contained_objects)
            parts.append("")
        if agg.invariants:
            parts.append("**Invariants:**")
            parts.extend(
                f"{i}. {inv}" for i, inv in enumerate(agg.invariants, 1)
            )
            parts.append("")
        if agg.commands:
            parts.append("**Commands:**")
            parts.extend(f"- `{cmd}`" for cmd in agg.commands)
            parts.append("")
        if agg.domain_events:
            parts.append("**Domain Events:**")
            parts.extend(f"- `{evt}`" for evt in agg.domain_events)
            parts.append("")

    def _render_ddd_stories(self, parts: list[str], model: DomainModel) -> None:
        parts.extend(["## 1. Domain Stories", ""])
        for story in model.domain_stories:
            self._render_story(parts, story)

    def _render_ddd_glossary(self, parts: list[str], model: DomainModel) -> None:
        parts.extend(["## 2. Ubiquitous Language Glossary", ""])
        terms = model.ubiquitous_language.terms
        if terms:
            parts.extend([
                "| Term | Definition | Context / Bounded Context |",
                "|------|-----------|---------------------------|",
            ])
            for entry in terms:
                self._render_term_row(parts, entry)
            parts.append("")
        else:
            parts.extend(["_No terms defined yet._", ""])

    def _render_ddd_classifications(
        self, parts: list[str], model: DomainModel
    ) -> None:
        parts.extend([
            "## 3. Subdomain Classification",
            "",
            "| Subdomain | Type | Rationale | Architecture Approach |",
            "|-----------|------|-----------|----------------------|",
        ])
        for ctx in model.bounded_contexts:
            self._render_classification_row(parts, ctx)
        parts.append("")

    def _render_ddd_contexts(self, parts: list[str], model: DomainModel) -> None:
        parts.extend(["## 4. Bounded Contexts", ""])
        for ctx in model.bounded_contexts:
            self._render_context_section(parts, ctx)

    @staticmethod
    def _render_ddd_relationships(parts: list[str], model: DomainModel) -> None:
        if not model.context_relationships:
            return
        parts.extend([
            "### Context Map (Relationships)",
            "",
            "| Upstream Context | Downstream Context | Integration Pattern |",
            "|-----------------|-------------------|-------------------|",
        ])
        parts.extend(
            f"| {_escape_table_cell(rel.upstream)} "
            f"| {_escape_table_cell(rel.downstream)} "
            f"| {_escape_table_cell(rel.integration_pattern)} |"
            for rel in model.context_relationships
        )
        parts.append("")

    def _render_ddd_aggregates(self, parts: list[str], model: DomainModel) -> None:
        parts.extend(["## 5. Aggregate Design", ""])
        if model.aggregate_designs:
            for agg in model.aggregate_designs:
                self._render_aggregate_section(parts, agg)
        else:
            parts.extend(["_No aggregate designs yet._", ""])

    # ── Architecture private helpers ─────────────────────────────────

    @staticmethod
    def _render_arch_principles(parts: list[str]) -> None:
        parts.extend([
            "## 1. Design Principles",
            "",
            "1. **Domain purity** — domain layer has zero external dependencies",
            (
                "2. **DDD alignment** — architecture follows bounded context "
                "boundaries from `docs/DDD.md`"
            ),
            (
                "3. **Testability** — every component testable in isolation "
                "with dependency injection"
            ),
            "",
        ])

    @staticmethod
    def _render_arch_overview(parts: list[str], model: DomainModel) -> None:
        parts.extend([
            "## 2. System Overview",
            "",
            "### Component Summary",
            "",
            "| Component | Responsibility | Bounded Context |",
            "| ------------- | -------------- | ------------------------------- |",
        ])
        parts.extend(
            f"| {_escape_table_cell(ctx.name)} "
            f"| {_escape_table_cell(ctx.responsibility)} "
            f"| {_escape_table_cell(ctx.name)} |"
            for ctx in model.bounded_contexts
        )
        parts.append("")

    @staticmethod
    def _render_arch_layers(parts: list[str]) -> None:
        parts.extend([
            "## 3. Layer Architecture",
            "",
            "Following Hexagonal / Clean Architecture aligned with DDD:",
            "",
            "### Layer Rules",
            "",
            "| Layer | Can Depend On | Cannot Depend On |",
            (
                "| -------------- | -------------------------- "
                "| --------------------------------------- |"
            ),
            (
                "| Domain | Nothing (pure Python) "
                "| Application, Infrastructure, frameworks |"
            ),
            (
                "| Application | Domain, Ports (interfaces) "
                "| Infrastructure, frameworks |"
            ),
            (
                "| Infrastructure | Application, Domain "
                "| — (outermost layer) |"
            ),
            "",
            "### Source Layout",
            "",
        ])
        parts.extend(_SOURCE_LAYOUT)

    @staticmethod
    def _render_arch_classification(parts: list[str], model: DomainModel) -> None:
        parts.extend([
            "## 4. Bounded Context Integration",
            "",
            "### Subdomain Classification",
            "",
            "| Subdomain | Type | Rationale |",
            "|-----------|------|-----------|",
        ])
        for ctx in model.bounded_contexts:
            classification = (
                ctx.classification.value if ctx.classification else "unclassified"
            )
            rationale = ctx.classification_rationale or "—"
            parts.append(
                f"| {_escape_table_cell(ctx.name)} "
                f"| **{_escape_table_cell(classification)}** "
                f"| {_escape_table_cell(rationale)} |"
            )
        parts.append("")

    @staticmethod
    def _render_arch_data_model(parts: list[str], model: DomainModel) -> None:
        parts.extend([
            "## 5. Data Model",
            "",
            "### Aggregates and Storage",
            "",
        ])
        if model.aggregate_designs:
            parts.extend([
                "| Aggregate | Context | Root Entity |",
                "| ------------- | -------------------------- | ---------------- |",
            ])
            parts.extend(
                f"| {_escape_table_cell(agg.name)} "
                f"| {_escape_table_cell(agg.context_name)} "
                f"| {_escape_table_cell(agg.root_entity)} |"
                for agg in model.aggregate_designs
            )
            parts.append("")
        else:
            parts.extend(["_No aggregate designs yet._", ""])
