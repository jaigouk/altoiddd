"""Tests for shared domain error hierarchy (M2 + N3)."""

from __future__ import annotations

import pytest

from src.domain.models.errors import (
    DomainError,
    DuplicateStoryError,
    InvariantViolationError,
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
