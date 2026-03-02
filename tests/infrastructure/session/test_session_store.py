"""Tests for SessionStore -- in-memory session storage with TTL.

RED phase: these tests define the SessionStore contract.
"""

from __future__ import annotations

import time
from typing import Any

import pytest

from src.domain.models.errors import SessionNotFoundError
from src.infrastructure.session.session_store import SessionStore

# ── Helpers ──────────────────────────────────────────────────────────


def _make_session() -> Any:
    """Create a minimal DiscoverySession for testing."""
    from src.domain.models.discovery_session import DiscoverySession

    return DiscoverySession(readme_content="Test idea")


# ── Put / Get ────────────────────────────────────────────────────────


class TestSessionStorePutGet:
    def test_put_and_get_returns_session(self) -> None:
        store = SessionStore(ttl_seconds=60)
        session = _make_session()
        store.put(session.session_id, session)
        assert store.get(session.session_id) is session

    def test_get_missing_raises_not_found(self) -> None:
        store = SessionStore(ttl_seconds=60)
        with pytest.raises(SessionNotFoundError):
            store.get("nonexistent-id")

    def test_put_overwrites_existing(self) -> None:
        store = SessionStore(ttl_seconds=60)
        session1 = _make_session()
        session2 = _make_session()
        # Use same key for both
        store.put("same-key", session1)
        store.put("same-key", session2)
        assert store.get("same-key") is session2

    def test_put_refreshes_ttl(self) -> None:
        store = SessionStore(ttl_seconds=0.1)
        session = _make_session()
        store.put(session.session_id, session)
        # Re-put to refresh TTL
        store.put(session.session_id, session)
        # Should still be accessible immediately after refresh
        assert store.get(session.session_id) is session


# ── TTL / Expiration ─────────────────────────────────────────────────


class TestSessionStoreTTL:
    def test_get_expired_raises_not_found(self) -> None:
        store = SessionStore(ttl_seconds=0.05)
        session = _make_session()
        store.put(session.session_id, session)
        time.sleep(0.1)
        with pytest.raises(SessionNotFoundError):
            store.get(session.session_id)

    def test_expired_session_is_lazily_removed(self) -> None:
        store = SessionStore(ttl_seconds=0.05)
        session = _make_session()
        store.put(session.session_id, session)
        assert store.active_count() == 1
        time.sleep(0.1)
        with pytest.raises(SessionNotFoundError):
            store.get(session.session_id)
        # After failed get, session should be removed
        assert store.active_count() == 0


# ── Cleanup ──────────────────────────────────────────────────────────


class TestSessionStoreCleanup:
    def test_cleanup_removes_expired(self) -> None:
        store = SessionStore(ttl_seconds=0.05)
        s1 = _make_session()
        s2 = _make_session()
        store.put(s1.session_id, s1)
        store.put(s2.session_id, s2)
        time.sleep(0.1)
        removed = store.cleanup_expired()
        assert removed == 2
        assert store.active_count() == 0

    def test_cleanup_on_empty_store_returns_zero(self) -> None:
        store = SessionStore(ttl_seconds=60)
        assert store.cleanup_expired() == 0

    def test_cleanup_keeps_live_sessions(self) -> None:
        store = SessionStore(ttl_seconds=60)
        session = _make_session()
        store.put(session.session_id, session)
        removed = store.cleanup_expired()
        assert removed == 0
        assert store.active_count() == 1

    def test_cleanup_mixed_expired_and_live(self) -> None:
        store = SessionStore(ttl_seconds=0.05)
        old = _make_session()
        store.put(old.session_id, old)
        time.sleep(0.1)
        # Add a fresh session
        fresh = _make_session()
        store.put(fresh.session_id, fresh)
        removed = store.cleanup_expired()
        assert removed == 1
        assert store.active_count() == 1
        assert store.get(fresh.session_id) is fresh


# ── Active count ─────────────────────────────────────────────────────


class TestSessionStoreActiveCount:
    def test_active_count_empty(self) -> None:
        store = SessionStore(ttl_seconds=60)
        assert store.active_count() == 0

    def test_active_count_after_puts(self) -> None:
        store = SessionStore(ttl_seconds=60)
        for _ in range(3):
            s = _make_session()
            store.put(s.session_id, s)
        assert store.active_count() == 3

    def test_active_count_includes_expired_before_cleanup(self) -> None:
        """active_count is a raw count; cleanup_expired must be called to prune."""
        store = SessionStore(ttl_seconds=0.05)
        session = _make_session()
        store.put(session.session_id, session)
        time.sleep(0.1)
        # Expired but not yet cleaned up
        assert store.active_count() == 1
        store.cleanup_expired()
        assert store.active_count() == 0


# ── Default TTL ──────────────────────────────────────────────────────


class TestSessionStoreDefaults:
    def test_default_ttl_is_30_minutes(self) -> None:
        store = SessionStore()
        assert store.ttl_seconds == 1800
