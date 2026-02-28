"""DiscoveryHandler -- application command for the guided discovery flow.

Orchestrates the 10-question DDD discovery session lifecycle. Manages
in-memory session state and delegates all domain logic to DiscoverySession.
"""

from __future__ import annotations

from src.domain.models.bootstrap_session import SessionNotFoundError
from src.domain.models.discovery_session import DiscoverySession


class DiscoveryHandler:
    """Orchestrates the discovery session lifecycle.

    Attributes:
        _sessions: In-memory store of active sessions (keyed by session_id).
    """

    def __init__(self) -> None:
        self._sessions: dict[str, DiscoverySession] = {}

    def _get_session(self, session_id: str) -> DiscoverySession:
        """Look up a session by ID.

        Raises:
            SessionNotFoundError: If no session matches the given ID.
        """
        try:
            return self._sessions[session_id]
        except KeyError:
            raise SessionNotFoundError(
                f"No active discovery session with id '{session_id}'"
            ) from None

    def start_session(self, readme_content: str) -> DiscoverySession:
        """Start a new discovery session from README content.

        Args:
            readme_content: The raw text of the project README.

        Returns:
            The newly created DiscoverySession.
        """
        session = DiscoverySession(readme_content=readme_content)
        self._sessions[session.session_id] = session
        return session

    def detect_persona(self, session_id: str, choice: str) -> DiscoverySession:
        """Detect user persona for the given session.

        Args:
            session_id: The active discovery session identifier.
            choice: The persona choice ("1"-"4").

        Returns:
            The updated DiscoverySession.
        """
        session = self._get_session(session_id)
        session.detect_persona(choice)
        return session

    def answer_question(self, session_id: str, question_id: str, answer: str) -> DiscoverySession:
        """Submit an answer to a discovery question.

        Args:
            session_id: The active discovery session identifier.
            question_id: The question being answered (e.g. "Q1").
            answer: The user's free-text answer.

        Returns:
            The updated DiscoverySession.
        """
        session = self._get_session(session_id)
        session.answer_question(question_id, answer)
        return session

    def confirm_playback(self, session_id: str, *, confirmed: bool) -> DiscoverySession:
        """Confirm or reject a playback summary.

        Args:
            session_id: The active discovery session identifier.
            confirmed: True if the user confirms the playback.

        Returns:
            The updated DiscoverySession.
        """
        session = self._get_session(session_id)
        session.confirm_playback(confirmed=confirmed)
        return session

    def complete(self, session_id: str) -> DiscoverySession:
        """Complete the discovery session.

        Args:
            session_id: The active discovery session identifier.

        Returns:
            The completed DiscoverySession with events.
        """
        session = self._get_session(session_id)
        session.complete()
        return session
