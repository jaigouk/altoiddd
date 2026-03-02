"""Composition root for the alty application.

Wires all port implementations into a single AppContext dataclass.
This is the entry point for dependency injection -- infrastructure
adapters are constructed here and injected into the application layer.

Generic subdomain: stub implementations raise NotImplementedError.
Real adapters will replace stubs in downstream tickets.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.bootstrap_port import BootstrapPort
    from src.application.ports.config_generation_port import ConfigGenerationPort
    from src.application.ports.discovery_port import DiscoveryPort
    from src.application.ports.doc_health_port import DocHealthPort
    from src.application.ports.doc_review_port import DocReviewPort
    from src.application.ports.fitness_generation_port import FitnessGenerationPort
    from src.application.ports.quality_gate_port import QualityGatePort
    from src.application.ports.spike_follow_up_port import SpikeFollowUpPort
    from src.application.ports.ticket_generation_port import TicketGenerationPort
    from src.application.ports.ticket_health_port import TicketHealthPort
    from src.application.ports.tool_detection_port import ToolDetectionPort
    from src.domain.models.discovery_session import DiscoverySession
    from src.domain.models.doc_health import DocHealthReport
    from src.domain.models.follow_up_intent import FollowUpAuditResult
    from src.domain.models.quality_gate import QualityGate, QualityReport
    from src.domain.models.ticket_freshness import TicketHealthReport


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


# ── Stub implementations ────────────────────────────────────────────


class _StubBootstrap:
    """Stub BootstrapPort -- raises NotImplementedError on all methods."""

    def preview(self, project_dir: Path) -> str:
        raise NotImplementedError

    def confirm(self, session_id: str) -> str:
        raise NotImplementedError

    def execute(self, session_id: str) -> str:
        raise NotImplementedError


class _StubDiscovery:
    """Stub DiscoveryPort -- raises NotImplementedError on all methods."""

    def get_session(self, session_id: str) -> DiscoverySession:
        raise NotImplementedError

    def start_session(self, readme_content: str) -> DiscoverySession:
        raise NotImplementedError

    def detect_persona(self, session_id: str, choice: str) -> DiscoverySession:
        raise NotImplementedError

    def answer_question(self, session_id: str, question_id: str, answer: str) -> DiscoverySession:
        raise NotImplementedError

    def skip_question(self, session_id: str, question_id: str, reason: str) -> DiscoverySession:
        raise NotImplementedError

    def confirm_playback(
        self, session_id: str, confirmed: bool, corrections: str = ""
    ) -> DiscoverySession:
        raise NotImplementedError

    def complete(self, session_id: str) -> DiscoverySession:
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


class _StubQualityGate:
    """Stub QualityGatePort -- raises NotImplementedError."""

    def check(
        self,
        gates: tuple[QualityGate, ...] | None = None,
    ) -> QualityReport:
        raise NotImplementedError


class _StubDocHealth:
    """Stub DocHealthPort -- raises NotImplementedError."""

    def check(self, project_dir: Path) -> DocHealthReport:
        raise NotImplementedError

    def check_knowledge(self, knowledge_dir: Path) -> DocHealthReport:
        raise NotImplementedError


class _StubDocReview:
    """Stub DocReviewPort -- raises NotImplementedError."""

    def mark_reviewed(self, doc_path: Path, reviewer: str) -> str:
        raise NotImplementedError

    def review_status(self, project_dir: Path) -> str:
        raise NotImplementedError


class _StubTicketHealth:
    """Stub TicketHealthPort -- raises NotImplementedError."""

    def report(self, project_dir: Path) -> TicketHealthReport:
        raise NotImplementedError


class _StubSpikeFollowUp:
    """Stub SpikeFollowUpPort -- raises NotImplementedError."""

    def audit(self, spike_id: str, project_dir: Path) -> FollowUpAuditResult:
        raise NotImplementedError


def create_app() -> AppContext:
    """Wire all port implementations and return the application context.

    Currently returns stub implementations that raise NotImplementedError.
    Real adapters will be wired in downstream tickets.
    """
    return AppContext(
        bootstrap=_StubBootstrap(),
        discovery=_StubDiscovery(),
        tool_detection=_StubToolDetection(),
        fitness_generation=_StubFitnessGeneration(),
        ticket_generation=_StubTicketGeneration(),
        config_generation=_StubConfigGeneration(),
        quality_gate=_StubQualityGate(),
        doc_health=_StubDocHealth(),
        doc_review=_StubDocReview(),
        ticket_health=_StubTicketHealth(),
        spike_follow_up=_StubSpikeFollowUp(),
    )
