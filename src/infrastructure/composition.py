"""Composition root for the alty application.

Wires all port implementations into a single AppContext dataclass.
This is the entry point for dependency injection -- infrastructure
adapters are constructed here and injected into the application layer.

Real adapters: discovery, quality_gate, file_writer, artifact_renderer.
Stub implementations (Phase 3+) raise NotImplementedError.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from datetime import date
    from pathlib import Path

    from src.application.ports.artifact_generation_port import ArtifactRendererPort
    from src.application.ports.bootstrap_port import BootstrapPort
    from src.application.ports.challenger_port import ChallengerPort
    from src.application.ports.config_generation_port import ConfigGenerationPort
    from src.application.ports.discovery_port import DiscoveryPort
    from src.application.ports.doc_health_port import DocHealthPort
    from src.application.ports.doc_review_port import DocReviewPort
    from src.application.ports.domain_research_port import DomainResearchPort
    from src.application.ports.file_writer_port import FileWriterPort
    from src.application.ports.fitness_generation_port import FitnessGenerationPort
    from src.application.ports.quality_gate_port import QualityGatePort
    from src.application.ports.spike_follow_up_port import SpikeFollowUpPort
    from src.application.ports.ticket_generation_port import TicketGenerationPort
    from src.application.ports.ticket_health_port import TicketHealthPort
    from src.application.ports.tool_detection_port import ToolDetectionPort
    from src.domain.models.doc_health import (
        DocHealthReport,
        DocReviewResult,
        DocStatus,
    )
    from src.domain.models.follow_up_intent import FollowUpAuditResult
    from src.domain.models.ticket_freshness import TicketHealthReport
    from src.infrastructure.external.llm_client import LLMClient


@dataclass
class AppContext:
    """Holds all port implementations for the application.

    Each attribute corresponds to one application port Protocol.
    Downstream adapters (CLI, MCP) access ports through this context.
    """

    bootstrap: BootstrapPort
    discovery: DiscoveryPort
    tool_detection: ToolDetectionPort
    fitness_generation: FitnessGenerationPort
    ticket_generation: TicketGenerationPort
    config_generation: ConfigGenerationPort
    quality_gate: QualityGatePort
    doc_health: DocHealthPort
    doc_review: DocReviewPort
    ticket_health: TicketHealthPort
    spike_follow_up: SpikeFollowUpPort
    file_writer: FileWriterPort
    artifact_renderer: ArtifactRendererPort
    llm_client: LLMClient | None = None
    challenger: ChallengerPort | None = None
    domain_research: DomainResearchPort | None = None


# ── Stub implementations ────────────────────────────────────────────


class _StubBootstrap:
    """Stub BootstrapPort -- raises NotImplementedError on all methods."""

    def preview(self, project_dir: Path) -> str:
        raise NotImplementedError

    def confirm(self, session_id: str) -> str:
        raise NotImplementedError

    def execute(self, session_id: str) -> str:
        raise NotImplementedError


class _StubToolDetection:
    """Stub ToolDetectionPort -- raises NotImplementedError on all methods."""

    def detect(self, project_dir: Path) -> list[str]:
        raise NotImplementedError

    def scan_conflicts(self, project_dir: Path) -> list[str]:
        raise NotImplementedError


class _StubFitnessGeneration:
    """Stub FitnessGenerationPort -- raises NotImplementedError."""

    def generate(
        self,
        model: object,
        root_package: str,
        output_dir: Path,
    ) -> None:
        raise NotImplementedError


class _StubTicketGeneration:
    """Stub TicketGenerationPort -- raises NotImplementedError."""

    def generate(self, model: object, output_dir: Path) -> None:
        raise NotImplementedError


class _StubConfigGeneration:
    """Stub ConfigGenerationPort -- raises NotImplementedError."""

    def generate(
        self,
        model: object,
        tools: tuple[object, ...],
        output_dir: Path,
    ) -> None:
        raise NotImplementedError


class _StubDocHealth:
    """Stub DocHealthPort -- raises NotImplementedError."""

    def check(self, project_dir: Path) -> DocHealthReport:
        raise NotImplementedError

    def check_knowledge(self, knowledge_dir: Path) -> DocHealthReport:
        raise NotImplementedError


class _StubDocReview:
    """Stub DocReviewPort -- raises NotImplementedError.

    Deprecated: replaced by DocReviewHandler in create_app().
    Kept for reference only.
    """

    def reviewable_docs(self, project_dir: Path) -> tuple[DocStatus, ...]:
        raise NotImplementedError

    def mark_reviewed(
        self,
        doc_path: str,
        project_dir: Path,
        review_date: date | None = None,
    ) -> DocReviewResult:
        raise NotImplementedError

    def mark_all_reviewed(
        self,
        project_dir: Path,
        review_date: date | None = None,
    ) -> tuple[DocReviewResult, ...]:
        raise NotImplementedError


class _StubTicketHealth:
    """Stub TicketHealthPort -- raises NotImplementedError.

    Deprecated: replaced by BeadsTicketHealthAdapter in create_app().
    Kept for reference only.
    """

    def report(self, project_dir: Path) -> TicketHealthReport:
        raise NotImplementedError


class _StubSpikeFollowUp:
    """Stub SpikeFollowUpPort -- raises NotImplementedError."""

    def audit(self, spike_id: str, project_dir: Path) -> FollowUpAuditResult:
        raise NotImplementedError


def create_app() -> AppContext:
    """Wire all port implementations and return the application context.

    Real adapters are used for ports with concrete implementations.
    Stubs remain for ports not yet implemented (Phase 3+).
    """
    from src.application.commands.doc_review_handler import DocReviewHandler
    from src.application.commands.quality_gate_handler import QualityGateHandler
    from src.infrastructure.external.beads_ticket_health_adapter import (
        BeadsTicketHealthAdapter,
    )
    from src.infrastructure.external.subprocess_gate_runner import SubprocessGateRunner
    from src.infrastructure.persistence.filesystem_doc_scanner import (
        FilesystemDocScanner,
    )
    from src.infrastructure.persistence.filesystem_file_writer import (
        FilesystemFileWriter,
    )
    from src.infrastructure.persistence.markdown_artifact_renderer import (
        MarkdownArtifactRenderer,
    )
    from src.infrastructure.session.in_memory_discovery_adapter import (
        InMemoryDiscoveryAdapter,
    )
    from src.infrastructure.session.session_store import SessionStore

    store = SessionStore()

    return AppContext(
        bootstrap=_StubBootstrap(),
        discovery=InMemoryDiscoveryAdapter(store=store),
        tool_detection=_StubToolDetection(),
        fitness_generation=_StubFitnessGeneration(),
        ticket_generation=_StubTicketGeneration(),
        config_generation=_StubConfigGeneration(),
        quality_gate=QualityGateHandler(runner=SubprocessGateRunner()),
        doc_health=_StubDocHealth(),
        doc_review=DocReviewHandler(scanner=FilesystemDocScanner()),
        ticket_health=BeadsTicketHealthAdapter(),
        spike_follow_up=_StubSpikeFollowUp(),
        file_writer=FilesystemFileWriter(),
        artifact_renderer=MarkdownArtifactRenderer(),
    )
