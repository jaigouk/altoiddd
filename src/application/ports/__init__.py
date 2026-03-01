"""Application layer ports (interfaces).

Defines Protocol classes that infrastructure adapters must implement.
These ports ensure the application layer depends only on abstractions,
not on concrete infrastructure implementations (Dependency Inversion).
"""

from __future__ import annotations

from src.application.ports.artifact_generation_port import ArtifactRendererPort
from src.application.ports.bootstrap_port import BootstrapPort
from src.application.ports.config_generation_port import ConfigGenerationPort
from src.application.ports.discovery_port import DiscoveryPort
from src.application.ports.doc_health_port import DocHealthPort
from src.application.ports.doc_review_port import DocReviewPort
from src.application.ports.file_writer_port import FileWriterPort
from src.application.ports.fitness_generation_port import FitnessGenerationPort
from src.application.ports.knowledge_lookup_port import KnowledgeLookupPort
from src.application.ports.persona_port import PersonaPort
from src.application.ports.quality_gate_port import QualityGatePort
from src.application.ports.rescue_port import RescuePort
from src.application.ports.ticket_generation_port import TicketGenerationPort
from src.application.ports.ticket_health_port import TicketHealthPort
from src.application.ports.tool_detection_port import ToolDetectionPort

__all__ = [
    "ArtifactRendererPort",
    "BootstrapPort",
    "ConfigGenerationPort",
    "DiscoveryPort",
    "DocHealthPort",
    "DocReviewPort",
    "FileWriterPort",
    "FitnessGenerationPort",
    "KnowledgeLookupPort",
    "PersonaPort",
    "QualityGatePort",
    "RescuePort",
    "TicketGenerationPort",
    "TicketHealthPort",
    "ToolDetectionPort",
]
