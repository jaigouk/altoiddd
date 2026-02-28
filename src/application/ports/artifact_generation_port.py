"""Port for the Domain Model bounded context (artifact generation).

Defines the interface for generating DDD artifacts (domain stories,
bounded context maps, ubiquitous language) from a domain model.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class ArtifactGenerationPort(Protocol):
    """Interface for generating DDD artifacts.

    Adapters implement this to produce domain stories, bounded context
    maps, and ubiquitous language glossaries from a domain model.
    """

    def generate(self, domain_model: str, output_dir: Path) -> str:
        """Generate DDD artifacts from a domain model.

        Args:
            domain_model: Serialized domain model from the discovery session.
            output_dir: Directory where generated artifacts will be written.

        Returns:
            Summary of the generated artifacts.
        """
        ...
