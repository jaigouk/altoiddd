"""Port for the Architecture Testing bounded context (fitness function generation).

Defines the interface for generating architecture fitness functions
(import-linter contracts, pytestarch tests) from a domain model.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.domain_model import DomainModel


@runtime_checkable
class FitnessGenerationPort(Protocol):
    """Interface for generating architecture fitness functions.

    Adapters implement this to produce import-linter TOML contracts and
    pytestarch test files driven by bounded context maps and subdomain
    classification (complexity budget).

    Handlers using this port implement the preview-before-action pattern:
    build_preview() renders content, approve_and_write() commits it.
    """

    def generate(
        self,
        model: DomainModel,
        root_package: str,
        output_dir: Path,
    ) -> None:
        """Generate fitness function tests from a domain model.

        Args:
            model: DomainModel with classified bounded contexts.
            root_package: Root Python package name.
            output_dir: Directory where generated test files will be written.
        """
        ...
