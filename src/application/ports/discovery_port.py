"""Port for the Guided Discovery bounded context.

Defines the interface for the conversational DDD question flow that
extracts domain knowledge using dual-register persona detection.
"""

from __future__ import annotations

from typing import Protocol, runtime_checkable


@runtime_checkable
class DiscoveryPort(Protocol):
    """Interface for guided discovery session operations.

    Adapters implement this to manage the conversational flow of the
    10-question DDD framework with persona detection, register selection,
    and playback confirmation loops.
    """

    def start_session(self, readme_content: str) -> str:
        """Start a new guided discovery session from README content.

        Args:
            readme_content: The raw text of the project README (4-5 sentence idea).

        Returns:
            The session identifier for the new discovery session.
        """
        ...

    def detect_persona(self, session_id: str, choice: str) -> str:
        """Detect the user persona based on their self-identification choice.

        Args:
            session_id: The active discovery session identifier.
            choice: The user's persona selection (e.g., Solo Developer, Team Lead).

        Returns:
            Confirmation of the detected persona and selected register.
        """
        ...

    def answer_question(self, session_id: str, answer: str) -> str:
        """Submit an answer to the current discovery question.

        Args:
            session_id: The active discovery session identifier.
            answer: The user's free-text answer to the current question.

        Returns:
            The next question, or a playback summary if a playback checkpoint is reached.
        """
        ...

    def confirm_playback(self, session_id: str, confirmed: bool) -> str:
        """Confirm or reject the playback summary.

        Args:
            session_id: The active discovery session identifier.
            confirmed: True if the user confirms the playback, False to correct.

        Returns:
            The next question or correction prompt.
        """
        ...

    def complete(self, session_id: str) -> str:
        """Complete the discovery session and produce domain artifacts.

        Args:
            session_id: The active discovery session identifier.

        Returns:
            Summary of the generated domain model artifacts.
        """
        ...
