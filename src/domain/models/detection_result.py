"""DetectionResult value object and ConflictSeverity enum.

Captures the outcome of a tool detection scan: which tools were found,
any configuration conflicts, and their severity classification.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass, field
from types import MappingProxyType
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from collections.abc import Mapping

    from src.domain.models.detected_tool import DetectedTool


class ConflictSeverity(enum.Enum):
    """Severity of a configuration conflict between global and local settings.

    COMPATIBLE: Global and local settings agree; no action needed.
    WARNING:    A potential issue that the user should be aware of.
    CONFLICT:   A direct contradiction that may cause problems.
    """

    COMPATIBLE = "compatible"
    WARNING = "warning"
    CONFLICT = "conflict"


@dataclass(frozen=True)
class DetectionResult:
    """Immutable result of scanning for AI coding tools and conflicts.

    Attributes:
        detected_tools: Tuple of detected tool value objects.
        conflicts: Tuple of human-readable conflict descriptions.
        severity_map: Read-only mapping from conflict description to severity.
    """

    detected_tools: tuple[DetectedTool, ...]
    conflicts: tuple[str, ...]
    severity_map: Mapping[str, ConflictSeverity] = field(default_factory=dict)

    def __post_init__(self) -> None:
        """Wrap severity_map in MappingProxyType to enforce immutability."""
        object.__setattr__(self, "severity_map", MappingProxyType(dict(self.severity_map)))
