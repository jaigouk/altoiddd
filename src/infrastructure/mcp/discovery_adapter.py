"""DiscoveryAdapter -- bridges SessionStore and DiscoverySession.

Implements the DiscoveryPort protocol by managing session lifecycle
through the SessionStore. Each port method retrieves the session from
the store, delegates to the DiscoverySession aggregate, and returns
the updated session.

This adapter replaces the DiscoveryHandler's internal dict with the
shared SessionStore so that MCP tools can access sessions across
stateless request/response cycles.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.discovery_session import DiscoverySession
from src.infrastructure.session.session_store import SessionStore

if TYPE_CHECKING:
    from src.domain.models.discovery_values import DiscoveryMode
    from src.domain.models.tech_stack import TechStack


class DiscoveryAdapter:
    """Implements DiscoveryPort backed by a SessionStore.

    Attributes:
        _store: The session store for persisting sessions across calls.
    """

    def __init__(self, store: SessionStore | None = None) -> None:
        self._store = store or SessionStore()

    def get_session(self, session_id: str) -> DiscoverySession:
        """Retrieve a session from the store.

        Raises:
            SessionNotFoundError: If the session_id is not found or expired.
        """
        session = self._store.get(session_id)
        assert isinstance(session, DiscoverySession)  # type narrowing
        return session

    def start_session(self, readme_content: str) -> DiscoverySession:
        """Start a new discovery session and store it."""
        session = DiscoverySession(readme_content=readme_content)
        self._store.put(session.session_id, session)
        return session

    def set_mode(self, session_id: str, mode: DiscoveryMode) -> DiscoverySession:
        """Set the discovery mode on a session."""
        session = self.get_session(session_id)
        session.set_mode(mode)
        self._store.put(session.session_id, session)
        return session

    def set_tech_stack(self, session_id: str, tech_stack: TechStack) -> DiscoverySession:
        """Set the tech stack on a discovery session."""
        session = self.get_session(session_id)
        session.set_tech_stack(tech_stack)
        return session

    def detect_persona(self, session_id: str, choice: str) -> DiscoverySession:
        """Detect user persona for the given session."""
        session = self.get_session(session_id)
        session.detect_persona(choice)
        return session

    def answer_question(
        self, session_id: str, question_id: str, answer: str
    ) -> DiscoverySession:
        """Submit an answer to a discovery question."""
        session = self.get_session(session_id)
        session.answer_question(question_id, answer)
        return session

    def skip_question(
        self, session_id: str, question_id: str, reason: str
    ) -> DiscoverySession:
        """Skip a question with an explicit reason."""
        session = self.get_session(session_id)
        session.skip_question(question_id, reason)
        return session

    def confirm_playback(
        self, session_id: str, confirmed: bool
    ) -> DiscoverySession:
        """Confirm or reject a playback summary."""
        session = self.get_session(session_id)
        session.confirm_playback(confirmed=confirmed)
        return session

    def complete(self, session_id: str) -> DiscoverySession:
        """Complete the discovery session."""
        session = self.get_session(session_id)
        session.complete()
        return session
