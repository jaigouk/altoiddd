"""Domain events for the Architecture Testing bounded context."""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.fitness_values import ArchRule, Contract


@dataclass(frozen=True)
class FitnessTestsGenerated:
    """Emitted when a FitnessTestSuite is approved and ready for output.

    Attributes:
        suite_id: Unique ID of the approved suite.
        root_package: The root Python package these tests target.
        contracts: All generated import-linter contracts.
        arch_rules: All generated pytestarch rules.
    """

    suite_id: str
    root_package: str
    contracts: tuple[Contract, ...]
    arch_rules: tuple[ArchRule, ...]
