"""Value objects for iteration diff display.

DiffType, DiffEntry, ConvergenceMetric, ArtifactVersion, and ArtifactDiff
model the differences between two DomainModel snapshots across iterations.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass
from typing import TYPE_CHECKING

from src.domain.models.errors import InvariantViolationError

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.research import TrustLevel


class DiffType(enum.Enum):
    """Type of change between two artifact versions."""

    ADDED = "added"
    MODIFIED = "modified"
    REMOVED = "removed"
    DISAMBIGUATED = "disambiguated"


@dataclass(frozen=True)
class DiffEntry:
    """A single change between two DomainModel versions.

    Attributes:
        diff_type: Kind of change (added/modified/removed/disambiguated).
        section: Which artifact section changed (e.g. "Ubiquitous Language").
        description: Human-readable description of the change.
        provenance: Trust level of this diff entry.
    """

    diff_type: DiffType
    section: str
    description: str
    provenance: TrustLevel

    def __post_init__(self) -> None:
        if not self.section.strip():
            msg = "DiffEntry section cannot be empty"
            raise InvariantViolationError(msg)
        if not self.description.strip():
            msg = "DiffEntry description cannot be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class ConvergenceMetric:
    """Counts of changes by category to gauge convergence.

    Attributes:
        invariants_delta: Count of invariant changes.
        terms_delta: Count of UL term changes.
        stories_delta: Count of domain story changes.
        canvases_delta: Count of canvas changes.
    """

    invariants_delta: int
    terms_delta: int
    stories_delta: int
    canvases_delta: int

    def __post_init__(self) -> None:
        for field_name in ("invariants_delta", "terms_delta", "stories_delta", "canvases_delta"):
            value = getattr(self, field_name)
            if value < 0:
                msg = f"ConvergenceMetric.{field_name} cannot be negative, got {value}"
                raise InvariantViolationError(msg)


class ArtifactVersion:
    """A versioned snapshot of a DomainModel.

    Uses __slots__ and read-only properties to prevent mutation after
    construction. Not a frozen dataclass because DomainModel is mutable,
    but the reference itself is immutable.

    Attributes:
        version_number: Positive integer (1, 2, 3, ...).
        model: The DomainModel snapshot.
    """

    __slots__ = ("_model", "_version_number")

    def __init__(self, version_number: int, model: DomainModel) -> None:
        if version_number < 1:
            msg = f"ArtifactVersion.version_number must be >= 1, got {version_number}"
            raise InvariantViolationError(msg)
        object.__setattr__(self, "_version_number", version_number)
        object.__setattr__(self, "_model", model)

    @property
    def version_number(self) -> int:
        return object.__getattribute__(self, "_version_number")  # type: ignore[no-any-return]

    @property
    def model(self) -> DomainModel:
        return object.__getattribute__(self, "_model")  # type: ignore[no-any-return]

    def __setattr__(self, _name: str, _value: object) -> None:
        msg = "ArtifactVersion is immutable"
        raise AttributeError(msg)


@dataclass(frozen=True)
class ArtifactDiff:
    """The complete diff between two artifact versions.

    Attributes:
        from_version: Source version number.
        to_version: Target version number (must be > from_version).
        entries: Tuple of individual diff entries.
        convergence: Aggregated change counts.
    """

    from_version: int
    to_version: int
    entries: tuple[DiffEntry, ...]
    convergence: ConvergenceMetric

    def __post_init__(self) -> None:
        if self.from_version >= self.to_version:
            msg = (
                f"ArtifactDiff.from_version ({self.from_version}) "
                f"must be less than to_version ({self.to_version})"
            )
            raise InvariantViolationError(msg)
