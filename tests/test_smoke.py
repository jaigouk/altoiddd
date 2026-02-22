"""Smoke test to verify the test infrastructure works."""

from __future__ import annotations


def test_imports() -> None:
    """Verify core packages are importable."""
    import src.application
    import src.domain
    import src.infrastructure

    assert src.domain is not None
    assert src.application is not None
    assert src.infrastructure is not None
