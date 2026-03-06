"""In-memory implementation of DiscoveryPort.

Bridges the DiscoveryPort protocol to the DiscoverySession aggregate
using SessionStore for persistence between calls.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, cast

from src.domain.models.discovery_session import DiscoverySession

if TYPE_CHECKING:
    from src.domain.models.discovery_values import DiscoveryMode
    from src.domain.models.tech_stack import TechStack
    from src.infrastructure.session.session_store import SessionStore


class InMemoryDiscoveryAdapter:
    """In-memory adapter for guided discovery sessions.

    Wraps SessionStore and DiscoverySession aggregate, delegating
    all state transitions to the domain model.
    """

    def __init__(self, store: SessionStore) -> None:
        self._store = store

    def _get(self, session_id: str) -> DiscoverySession:
        """Retrieve and cast a session from the store."""
        return cast("DiscoverySession", self._store.get(session_id))

    def start_session(self, readme_content: str) -> DiscoverySession:
        """Start a new guided discovery session from README content.

        Args:
            readme_content: The raw text of the project README.

        Returns:
            The newly created DiscoverySession.
        """
        session = DiscoverySession(readme_content)
        self._store.put(session.session_id, session)
        return session

    def set_mode(self, session_id: str, mode: DiscoveryMode) -> DiscoverySession:
        """Set the discovery mode on a session.

        Args:
            session_id: The active discovery session identifier.
            mode: The DiscoveryMode to set.

        Returns:
            The updated DiscoverySession.
        """
        session = self._get(session_id)
        session.set_mode(mode)
        self._store.put(session_id, session)
        return session

    def set_tech_stack(self, session_id: str, tech_stack: TechStack) -> DiscoverySession:
        """Set the tech stack on a discovery session.

        Args:
            session_id: The active discovery session identifier.
            tech_stack: The TechStack value object to store.

        Returns:
            The updated DiscoverySession.
        """
        session = self._get(session_id)
        session.set_tech_stack(tech_stack)
        self._store.put(session_id, session)
        return session

    def detect_persona(self, session_id: str, choice: str) -> DiscoverySession:
        """Detect user persona for the given session.

        Args:
            session_id: The active discovery session identifier.
            choice: The persona choice ("1"-"4").

        Returns:
            The updated DiscoverySession.
        """
        session = self._get(session_id)
        session.detect_persona(choice)
        self._store.put(session_id, session)
        return session

    def answer_question(
        self, session_id: str, question_id: str, answer: str
    ) -> DiscoverySession:
        """Submit an answer to a discovery question.

        Args:
            session_id: The active discovery session identifier.
            question_id: The question being answered (e.g. "Q1").
            answer: The user's free-text answer.

        Returns:
            The updated DiscoverySession.
        """
        session = self._get(session_id)
        session.answer_question(question_id, answer)
        self._store.put(session_id, session)
        return session

    def skip_question(
        self, session_id: str, question_id: str, reason: str
    ) -> DiscoverySession:
        """Skip a question with an explicit reason.

        Args:
            session_id: The active discovery session identifier.
            question_id: The question to skip.
            reason: Why it was skipped (must be non-empty).

        Returns:
            The updated DiscoverySession.
        """
        session = self._get(session_id)
        session.skip_question(question_id, reason)
        self._store.put(session_id, session)
        return session

    def confirm_playback(
        self, session_id: str, confirmed: bool, corrections: str = ""
    ) -> DiscoverySession:
        """Confirm or reject a playback summary.

        Args:
            session_id: The active discovery session identifier.
            confirmed: True if the user confirms the playback.
            corrections: Optional correction text when rejecting playback.

        Returns:
            The updated DiscoverySession.
        """
        session = self._get(session_id)
        session.confirm_playback(confirmed, corrections)
        self._store.put(session_id, session)
        return session

    def get_session(self, session_id: str) -> DiscoverySession:
        """Retrieve a session by ID.

        Args:
            session_id: The session identifier to look up.

        Returns:
            The stored DiscoverySession.

        Raises:
            SessionNotFoundError: If no session matches the given ID.
        """
        return self._get(session_id)

    def complete(self, session_id: str) -> DiscoverySession:
        """Complete the discovery session and produce domain artifacts.

        Args:
            session_id: The active discovery session identifier.

        Returns:
            The completed DiscoverySession with events.
        """
        session = self._get(session_id)
        session.complete()
        self._store.put(session_id, session)
        return session
