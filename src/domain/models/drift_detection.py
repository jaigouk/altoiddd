"""Domain models for knowledge base drift detection.

DriftSignalType enumerates the kinds of drift that can be detected.
DriftSeverity classifies how urgent a drift signal is.
DriftSignal is a frozen value object describing a single detected drift.
DriftReport aggregates signals with computed summary properties.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


class DriftSignalType(enum.Enum):
    """Types of drift that can be detected in knowledge entries."""

    VERSION_CHANGE = "version_change"
    DOC_CODE_MISMATCH = "doc_code_mismatch"
    STALE = "stale"


class DriftSeverity(enum.Enum):
    """Severity classification for drift signals."""

    INFO = "info"
    WARNING = "warning"
    ERROR = "error"


@dataclass(frozen=True)
class DriftSignal:
    """A single detected drift in a knowledge entry.

    Attributes:
        entry_path: RLM-style path to the affected entry.
        signal_type: What kind of drift was detected.
        description: Human-readable description of the drift.
        severity: How urgent this drift is.
    """

    entry_path: str
    signal_type: DriftSignalType
    description: str
    severity: DriftSeverity

    def __post_init__(self) -> None:
        if not self.entry_path.strip():
            msg = "DriftSignal entry_path must not be empty"
            raise InvariantViolationError(msg)
        if not self.description.strip():
            msg = "DriftSignal description must not be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class DriftReport:
    """Aggregate report of drift signals across the knowledge base.

    Attributes:
        signals: Tuple of individual drift signals.
    """

    signals: tuple[DriftSignal, ...]

    @property
    def total_count(self) -> int:
        """Total number of drift signals."""
        return len(self.signals)

    @property
    def has_drift(self) -> bool:
        """Whether any drift was detected."""
        return self.total_count > 0

    def count_by_severity(self, severity: DriftSeverity) -> int:
        """Count signals with a specific severity."""
        return sum(1 for s in self.signals if s.severity == severity)

    def count_by_type(self, signal_type: DriftSignalType) -> int:
        """Count signals with a specific type."""
        return sum(1 for s in self.signals if s.signal_type == signal_type)
