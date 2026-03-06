"""Domain service that assembles Bounded Context Canvases from a DomainModel.

CanvasAssembler is stateless: it reads from a DomainModel aggregate and
produces immutable BoundedContextCanvas value objects.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.bounded_context_canvas import (
    BoundedContextCanvas,
    CommunicationMessage,
    DomainRole,
    StrategicClassification,
)
from src.domain.models.domain_values import SubdomainClassification

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel

_CLASSIFICATION_TO_ROLE: dict[SubdomainClassification, DomainRole] = {
    SubdomainClassification.CORE: DomainRole.EXECUTION,
    SubdomainClassification.SUPPORTING: DomainRole.SPECIFICATION,
    SubdomainClassification.GENERIC: DomainRole.GATEWAY,
}


class CanvasAssembler:
    """Stateless domain service that assembles BC Canvases from a DomainModel."""

    @staticmethod
    def assemble(model: DomainModel) -> tuple[BoundedContextCanvas, ...]:
        """Build one canvas per bounded context in the model."""
        canvases: list[BoundedContextCanvas] = []

        for ctx in model.bounded_contexts:
            domain_cls = ctx.classification or SubdomainClassification.GENERIC

            classification = StrategicClassification(
                domain=domain_cls,
                business_model="unclassified",
                evolution="unclassified",
            )

            role = _CLASSIFICATION_TO_ROLE.get(domain_cls, DomainRole.DRAFT)

            inbound = tuple(
                CommunicationMessage(
                    message=rel.integration_pattern,
                    message_type="Event",
                    counterpart=rel.upstream,
                )
                for rel in model.context_relationships
                if rel.downstream == ctx.name
            )
            outbound = tuple(
                CommunicationMessage(
                    message=rel.integration_pattern,
                    message_type="Event",
                    counterpart=rel.downstream,
                )
                for rel in model.context_relationships
                if rel.upstream == ctx.name
            )

            ul_terms = tuple(
                (entry.term, entry.definition)
                for entry in model.ubiquitous_language.terms
                if entry.context_name == ctx.name
            )

            decisions: list[str] = []
            for agg in model.aggregate_designs:
                if agg.context_name == ctx.name:
                    decisions.extend(agg.invariants)

            canvases.append(
                BoundedContextCanvas(
                    context_name=ctx.name,
                    purpose=ctx.responsibility,
                    classification=classification,
                    domain_roles=(role,),
                    inbound_communication=inbound,
                    outbound_communication=outbound,
                    ubiquitous_language=ul_terms,
                    business_decisions=tuple(decisions),
                    assumptions=(),
                    open_questions=(),
                )
            )

        return tuple(canvases)

    @staticmethod
    def render_markdown(canvases: tuple[BoundedContextCanvas, ...]) -> str:
        """Render canvases to markdown following ddd-crew v5 format."""
        if not canvases:
            return ""

        sections = [_render_single_canvas(canvas) for canvas in canvases]
        return "\n---\n\n".join(sections)


def _render_single_canvas(canvas: BoundedContextCanvas) -> str:
    """Render a single canvas to markdown."""
    lines: list[str] = [
        f"# Bounded Context Canvas: {canvas.context_name}",
        "",
    ]
    _add_purpose(lines, canvas)
    _add_classification(lines, canvas)
    _add_domain_roles(lines, canvas)
    _add_communication(lines, "Inbound Communication", "Sender", canvas.inbound_communication)
    _add_communication(lines, "Outbound Communication", "Receiver", canvas.outbound_communication)
    _add_ul(lines, canvas)
    _add_list_section(lines, "Business Decisions & Rules", canvas.business_decisions)
    _add_list_section(lines, "Assumptions", canvas.assumptions)
    _add_list_section(lines, "Open Questions", canvas.open_questions)
    return "\n".join(lines)


def _add_purpose(lines: list[str], canvas: BoundedContextCanvas) -> None:
    lines.extend(["## Purpose", "", canvas.purpose, ""])


def _add_classification(lines: list[str], canvas: BoundedContextCanvas) -> None:
    c = canvas.classification
    lines.extend([
        "## Strategic Classification",
        "",
        "| Aspect | Value |",
        "| --- | --- |",
        f"| Domain | {c.domain.value} |",
        f"| Business Model | {c.business_model} |",
        f"| Evolution | {c.evolution} |",
        "",
    ])


def _add_domain_roles(lines: list[str], canvas: BoundedContextCanvas) -> None:
    lines.extend(["## Domain Roles", ""])
    lines.extend(f"- [x] {role.value}" for role in canvas.domain_roles)
    lines.append("")


def _add_communication(
    lines: list[str],
    heading: str,
    third_col: str,
    messages: tuple[CommunicationMessage, ...],
) -> None:
    lines.extend([f"## {heading}", ""])
    if messages:
        lines.extend([
            f"| Message | Type | {third_col} |",
            "| --- | --- | --- |",
        ])
        lines.extend(
            f"| {m.message} | {m.message_type} | {m.counterpart} |" for m in messages
        )
    else:
        lines.append("*None*")
    lines.append("")


def _add_ul(lines: list[str], canvas: BoundedContextCanvas) -> None:
    lines.extend(["## Ubiquitous Language", ""])
    if canvas.ubiquitous_language:
        lines.extend(["| Term | Definition |", "| --- | --- |"])
        lines.extend(f"| {term} | {defn} |" for term, defn in canvas.ubiquitous_language)
    else:
        lines.append("*None*")
    lines.append("")


def _add_list_section(lines: list[str], heading: str, items: tuple[str, ...]) -> None:
    lines.extend([f"## {heading}", ""])
    if items:
        lines.extend(f"- {item}" for item in items)
    else:
        lines.append("*None*")
    lines.append("")
