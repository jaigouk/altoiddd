"""Shared domain error hierarchy.

All domain-level exceptions inherit from DomainError so callers
can catch domain errors generically or by specific subclass.
"""

from __future__ import annotations


class DomainError(Exception):
    """Base exception for all domain-layer errors."""


class InvariantViolationError(DomainError):
    """Raised when a domain invariant is violated."""


class DuplicateStoryError(DomainError):
    """Raised when adding a domain story with a name that already exists."""
