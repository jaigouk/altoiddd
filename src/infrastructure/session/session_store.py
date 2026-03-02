"""In-memory session store with TTL-based expiration.

Stores DiscoverySessions (or any object) keyed by session_id with
automatic time-to-live expiration. Expired entries are lazily removed
on get() and can be bulk-pruned via cleanup_expired().

This is an infrastructure service — it has no domain dependencies.
"""

from __future__ import annotations

import time

from src.domain.models.errors import SessionNotFoundError


class SessionStore:
    """In-memory session store with TTL.

    Attributes:
        ttl_seconds: Time-to-live for stored sessions (default 30 minutes).
    """

    def __init__(self, ttl_seconds: float = 1800) -> None:
        self.ttl_seconds = ttl_seconds
        self._store: dict[str, tuple[object, float]] = {}

    def put(self, session_id: str, session: object) -> None:
        """Store or overwrite a session with a fresh TTL timestamp.

        Args:
            session_id: Unique identifier for the session.
            session: The session object to store.
        """
        self._store[session_id] = (session, time.monotonic())

    def get(self, session_id: str) -> object:
        """Retrieve a session by ID.

        Lazily removes expired entries on access.

        Args:
            session_id: The session identifier to look up.

        Returns:
            The stored session object.

        Raises:
            SessionNotFoundError: If the session_id is not found or has expired.
        """
        entry = self._store.get(session_id)
        if entry is None:
            raise SessionNotFoundError(
                f"No active session with id '{session_id}'"
            )
        session, timestamp = entry
        if time.monotonic() - timestamp > self.ttl_seconds:
            del self._store[session_id]
            raise SessionNotFoundError(
                f"Session '{session_id}' has expired"
            )
        return session

    def cleanup_expired(self) -> int:
        """Remove all expired sessions from the store.

        Returns:
            The number of sessions removed.
        """
        now = time.monotonic()
        expired_keys = [
            key
            for key, (_, timestamp) in self._store.items()
            if now - timestamp > self.ttl_seconds
        ]
        for key in expired_keys:
            del self._store[key]
        return len(expired_keys)

    def active_count(self) -> int:
        """Return the number of sessions in the store (including expired).

        To get an accurate count of non-expired sessions, call
        cleanup_expired() first.
        """
        return len(self._store)
