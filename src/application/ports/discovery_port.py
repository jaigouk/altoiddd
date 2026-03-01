"""Port for the Guided Discovery bounded context.

Defines the interface for the conversational DDD question flow that
extracts domain knowledge using dual-register persona detection.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.discovery_session import DiscoverySession


@runtime_checkable
class DiscoveryPort(Protocol):
    """Interface for guided discovery session operations.

    Adapters implement this to manage the conversational flow of the
    10-question DDD framework with persona detection, register selection,
    and playback confirmation loops.
    """

    def start_session(self, readme_content: str) -> DiscoverySession:
        """Start a new guided discovery session from README content.

        Args:
            readme_content: The raw text of the project README (4-5 sentence idea).

        Returns:
            The newly created DiscoverySession.
        """
        ...

    def detect_persona(self, session_id: str, choice: str) -> DiscoverySession:
        """Detect the user persona based on their self-identification choice.

        Args:
            session_id: The active discovery session identifier.
            choice: The user's persona selection ("1"-"4").

        Returns:
            The updated DiscoverySession.
        """
        ...

    def answer_question(self, session_id: str, question_id: str, answer: str) -> DiscoverySession:
        """Submit an answer to a discovery question.

        Args:
            session_id: The active discovery session identifier.
            question_id: The question being answered (e.g. "Q1").
            answer: The user's free-text answer.

        Returns:
            The updated DiscoverySession.
        """
        ...

    def skip_question(self, session_id: str, question_id: str, reason: str) -> DiscoverySession:
        """Skip a question with an explicit reason.

        Args:
            session_id: The active discovery session identifier.
            question_id: The question to skip (e.g. "Q5").
            reason: Why it was skipped (must be non-empty).

        Returns:
            The updated DiscoverySession.
        """
        ...

    def confirm_playback(self, session_id: str, confirmed: bool) -> DiscoverySession:
        """Confirm or reject the playback summary.

        Args:
            session_id: The active discovery session identifier.
            confirmed: True if the user confirms the playback, False to correct.

        Returns:
            The updated DiscoverySession.
        """
        ...

    def complete(self, session_id: str) -> DiscoverySession:
        """Complete the discovery session and produce domain artifacts.

        Args:
            session_id: The active discovery session identifier.

        Returns:
            The completed DiscoverySession with events.
        """
        ...
