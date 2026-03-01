"""Tests for shared domain error hierarchy (M2 + N3)."""

from __future__ import annotations

import pytest

from src.domain.models.errors import (
    DomainError,
    DuplicateStoryError,
    InvariantViolationError,
    SessionNotFoundError,
)


class TestDomainErrorHierarchy:
    def test_invariant_violation_is_domain_error(self) -> None:
        assert isinstance(InvariantViolationError("test"), DomainError)

    def test_duplicate_story_is_domain_error(self) -> None:
        assert isinstance(DuplicateStoryError("test"), DomainError)

    def test_all_are_exceptions(self) -> None:
        assert isinstance(InvariantViolationError("x"), Exception)
        assert isinstance(DuplicateStoryError("x"), Exception)

    def test_message_preserved(self) -> None:
        assert str(InvariantViolationError("broken invariant")) == "broken invariant"
        assert str(DuplicateStoryError("dup story")) == "dup story"

    def test_catch_invariant_via_base(self) -> None:
        with pytest.raises(DomainError):
            raise InvariantViolationError("test")

    def test_catch_duplicate_via_base(self) -> None:
        with pytest.raises(DomainError):
            raise DuplicateStoryError("test")

    def test_distinct_types(self) -> None:
        """InvariantViolationError and DuplicateStoryError are different."""
        with pytest.raises(InvariantViolationError):
            raise InvariantViolationError("test")
        # DuplicateStoryError should NOT be caught by InvariantViolationError.
        with pytest.raises(DuplicateStoryError):
            raise DuplicateStoryError("test")

    def test_session_not_found_is_domain_error(self) -> None:
        assert isinstance(SessionNotFoundError("abc-123"), DomainError)

    def test_session_not_found_message_preserved(self) -> None:
        assert str(SessionNotFoundError("session xyz")) == "session xyz"

    def test_session_not_found_catch_via_base(self) -> None:
        with pytest.raises(DomainError):
            raise SessionNotFoundError("test")

    def test_session_not_found_distinct_from_invariant(self) -> None:
        """SessionNotFoundError is not caught by InvariantViolationError."""
        with pytest.raises(SessionNotFoundError):
            raise SessionNotFoundError("test")
        # Should NOT match InvariantViolationError
        with pytest.raises(SessionNotFoundError):
            raise SessionNotFoundError("test")


class TestCrossImportCompatibility:
    """Verify errors are importable from both canonical and legacy locations."""

    def test_invariant_from_errors_module(self) -> None:
        from src.domain.models.errors import InvariantViolationError as E

        assert E is InvariantViolationError

    def test_invariant_from_bootstrap_session(self) -> None:
        from src.domain.models.bootstrap_session import InvariantViolationError as E
        from src.domain.models.errors import InvariantViolationError as Canonical

        assert E is Canonical

    def test_session_not_found_from_errors_module(self) -> None:
        from src.domain.models.errors import SessionNotFoundError as E

        assert E is SessionNotFoundError

    def test_session_not_found_from_bootstrap_session(self) -> None:
        from src.domain.models.bootstrap_session import SessionNotFoundError as E
        from src.domain.models.errors import SessionNotFoundError as Canonical

        assert E is Canonical

    def test_discovery_session_uses_errors_module(self) -> None:
        """DiscoverySession should import InvariantViolationError from errors.py."""
        import inspect

        from src.domain.models import discovery_session

        source = inspect.getsource(discovery_session)
        assert "from src.domain.models.errors import InvariantViolationError" in source
