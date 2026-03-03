"""Domain events for the Guided Discovery bounded context.

DiscoveryCompleted is emitted when a discovery session finishes with
sufficient domain knowledge extracted from the user.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.discovery_values import Answer, Persona, Playback, Register
    from src.domain.models.tech_stack import TechStack


@dataclass(frozen=True)
class DiscoveryCompleted:
    """Emitted when a discovery session completes successfully.

    Attributes:
        session_id: Unique identifier of the completed session.
        persona: The detected user persona.
        register: The language register used during discovery.
        answers: All answers collected during the session.
        playback_confirmations: All playback confirmations from the session.
    """

    session_id: str
    persona: Persona
    register: Register
    answers: tuple[Answer, ...]
    playback_confirmations: tuple[Playback, ...]
    tech_stack: TechStack | None = None
