"""Application command handler for artifact generation.

ArtifactGenerationHandler transforms a DiscoveryCompleted event into a
DomainModel aggregate by mapping question answers to domain artifacts,
then renders PRD, DDD.md, and ARCHITECTURE.md.

Supports a preview-before-write workflow: build_preview() renders content
without writing, write_artifacts() commits a preview to disk.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.artifact_generation_port import ArtifactRendererPort
    from src.application.ports.file_writer_port import FileWriterPort
    from src.domain.events.discovery_events import DiscoveryCompleted

# Classification keyword mapping for Q10 answers.
_CLASSIFICATION_KEYWORDS: dict[str, SubdomainClassification] = {
    "core": SubdomainClassification.CORE,
    "secret sauce": SubdomainClassification.CORE,
    "competitive": SubdomainClassification.CORE,
    "supporting": SubdomainClassification.SUPPORTING,
    "plumbing": SubdomainClassification.SUPPORTING,
    "necessary": SubdomainClassification.SUPPORTING,
    "generic": SubdomainClassification.GENERIC,
    "off-the-shelf": SubdomainClassification.GENERIC,
    "commodity": SubdomainClassification.GENERIC,
    "buy": SubdomainClassification.GENERIC,
}


@dataclass
class ArtifactPreview:
    """Rendered artifact content ready for user review before writing.

    Attributes:
        model: The finalized DomainModel aggregate.
        prd_content: Rendered PRD markdown.
        ddd_content: Rendered DDD.md markdown.
        architecture_content: Rendered ARCHITECTURE.md markdown.
    """

    model: DomainModel
    prd_content: str
    ddd_content: str
    architecture_content: str


class ArtifactGenerationHandler:
    """Orchestrates the transformation of discovery answers into DDD artifacts.

    Builds a DomainModel from DiscoveryCompleted event data, renders it
    into markdown documents, and writes them via the FileWriterPort.
    """

    def __init__(
        self,
        renderer: ArtifactRendererPort,
        writer: FileWriterPort,
    ) -> None:
        self._renderer = renderer
        self._writer = writer

    def build_preview(self, event: DiscoveryCompleted) -> ArtifactPreview:
        """Build domain model, finalize, and render artifacts for preview.

        Does NOT write any files. Call write_artifacts() after user approval.

        Args:
            event: The DiscoveryCompleted event with all answers.

        Returns:
            ArtifactPreview with rendered content and finalized model.

        Raises:
            ValueError: If no substantive answers are present.
        """
        if not event.answers:
            msg = "No substantive answers to generate artifacts from"
            raise ValueError(msg)

        model = self._build_model(event)
        model.finalize()

        return ArtifactPreview(
            model=model,
            prd_content=self._renderer.render_prd(model),
            ddd_content=self._renderer.render_ddd(model),
            architecture_content=self._renderer.render_architecture(model),
        )

    def write_artifacts(self, preview: ArtifactPreview, output_dir: Path) -> None:
        """Write previously previewed artifacts to disk.

        Args:
            preview: The ArtifactPreview from build_preview().
            output_dir: Directory to write PRD.md, DDD.md, ARCHITECTURE.md.
        """
        self._writer.write_file(output_dir / "PRD.md", preview.prd_content)
        self._writer.write_file(output_dir / "DDD.md", preview.ddd_content)
        self._writer.write_file(
            output_dir / "ARCHITECTURE.md", preview.architecture_content
        )

    def generate(
        self,
        event: DiscoveryCompleted,
        output_dir: Path,
    ) -> DomainModel:
        """Build, preview, and write artifacts in one step (convenience).

        Args:
            event: The DiscoveryCompleted event with all answers.
            output_dir: Directory to write PRD.md, DDD.md, ARCHITECTURE.md.

        Returns:
            The finalized DomainModel.

        Raises:
            ValueError: If no substantive answers are present.
        """
        preview = self.build_preview(event)
        self.write_artifacts(preview, output_dir)
        return preview.model

    def _build_model(self, event: DiscoveryCompleted) -> DomainModel:
        """Build a DomainModel from discovery answers."""
        model = DomainModel()
        answers_by_id = {a.question_id: a.response_text for a in event.answers}

        # Extract contexts FIRST so terms get correct context names (M5).
        self._extract_contexts(model, answers_by_id)

        # Stories from Q1, Q3, Q5.
        self._extract_stories(model, answers_by_id)

        # Terms from Q2 — assigned to real context, never "Default".
        self._extract_terms(model, answers_by_id)

        # Classifications from Q10.
        self._extract_classifications(model, answers_by_id)

        # Aggregate designs from Q4, Q6 for Core subdomains.
        self._extract_aggregates(model, answers_by_id)

        return model

    def _extract_contexts(
        self,
        model: DomainModel,
        answers: dict[str, str],
    ) -> None:
        """Extract bounded contexts from Q9 answers."""
        contexts = self._split_answer(answers.get("Q9", ""))
        for ctx_name in contexts:
            if ctx_name.strip():
                model.add_bounded_context(
                    BoundedContext(
                        name=ctx_name.strip(),
                        responsibility=f"Manages {ctx_name.strip()} domain",
                    )
                )

    def _extract_stories(
        self,
        model: DomainModel,
        answers: dict[str, str],
    ) -> None:
        """Extract domain stories from Q1, Q3, Q5 answers."""
        actors = self._split_answer(answers.get("Q1", ""))
        primary_steps = self._split_answer(answers.get("Q3", ""))

        if primary_steps:
            model.add_domain_story(
                DomainStory(
                    name="Primary Flow",
                    actors=tuple(actors) if actors else ("User",),
                    trigger=primary_steps[0] if primary_steps else "User initiates",
                    steps=tuple(primary_steps),
                )
            )

        secondary = self._split_answer(answers.get("Q5", ""))
        if secondary:
            model.add_domain_story(
                DomainStory(
                    name="Secondary Flows",
                    actors=tuple(actors) if actors else ("User",),
                    trigger="Various",
                    steps=tuple(secondary),
                )
            )

    def _extract_terms(
        self,
        model: DomainModel,
        answers: dict[str, str],
    ) -> None:
        """Extract ubiquitous language terms from Q2 answers.

        Terms are assigned to the first bounded context (from Q9),
        or "General" if no contexts were extracted.
        """
        entities = self._split_answer(answers.get("Q2", ""))
        context_name = (
            model.bounded_contexts[0].name
            if model.bounded_contexts
            else "General"
        )

        for entity in entities:
            if entity.strip():
                model.add_term(
                    term=entity.strip(),
                    definition=f"{entity.strip()} entity",
                    context_name=context_name,
                    source_question_ids=("Q2",),
                )

    def _extract_classifications(
        self,
        model: DomainModel,
        answers: dict[str, str],
    ) -> None:
        """Extract subdomain classifications from Q10 answers."""
        q10 = answers.get("Q10", "").lower()
        if not q10:
            for ctx in model.bounded_contexts:
                if ctx.classification is None:
                    model.classify_subdomain(
                        ctx.name,
                        SubdomainClassification.SUPPORTING,
                        "No classification provided — defaulting to Supporting",
                    )
            return

        for ctx in model.bounded_contexts:
            classification = SubdomainClassification.SUPPORTING
            rationale = "Default classification"

            ctx_lower = ctx.name.lower()
            for keyword, cls in _CLASSIFICATION_KEYWORDS.items():
                if keyword in q10 and ctx_lower in q10:
                    classification = cls
                    rationale = f"Classified as {cls.value} based on Q10 answer"
                    break

            if ctx.classification is None:
                model.classify_subdomain(ctx.name, classification, rationale)

    def _extract_aggregates(
        self,
        model: DomainModel,
        answers: dict[str, str],
    ) -> None:
        """Extract aggregate designs for Core subdomains from Q4, Q6 answers."""
        invariants = self._split_answer(answers.get("Q4", ""))
        events = self._split_answer(answers.get("Q6", ""))

        core_contexts = [
            ctx
            for ctx in model.bounded_contexts
            if ctx.classification == SubdomainClassification.CORE
        ]

        for ctx in core_contexts:
            model.design_aggregate(
                AggregateDesign(
                    name=f"{ctx.name}Root",
                    context_name=ctx.name,
                    root_entity=f"{ctx.name}Root",
                    invariants=tuple(invariants),
                    domain_events=tuple(events),
                )
            )

    @staticmethod
    def _split_answer(answer: str) -> list[str]:
        """Split a free-text answer into meaningful parts.

        Handles comma-separated, newline-separated, and numbered lists.
        """
        if not answer.strip():
            return []

        # Try comma separation first.
        parts = answer.split(",")
        if len(parts) > 1:
            return [p.strip() for p in parts if p.strip()]

        # Try newline separation.
        parts = answer.split("\n")
        if len(parts) > 1:
            cleaned = []
            for p in parts:
                stripped = p.strip().lstrip("0123456789.-) ").strip()
                if stripped:
                    cleaned.append(stripped)
            return cleaned

        # Single item.
        return [answer.strip()] if answer.strip() else []
