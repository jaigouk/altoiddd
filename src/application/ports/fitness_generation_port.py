"""Port for the Architecture Testing bounded context (fitness function generation).

Defines the interface for generating architecture fitness functions
(import-linter contracts, pytestarch tests) from a domain model.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class FitnessGenerationPort(Protocol):
    """Interface for generating architecture fitness functions.

    Adapters implement this to produce import-linter TOML contracts and
    pytestarch test files driven by bounded context maps and subdomain
    classification (complexity budget).
    """

    def generate(self, domain_model: str, output_dir: Path) -> str:
        """Generate fitness function tests from a domain model.

        Args:
            domain_model: Serialized domain model with bounded context map
                and subdomain classification.
            output_dir: Directory where generated test files will be written.

        Returns:
            Summary of the generated fitness functions.
        """
        ...
