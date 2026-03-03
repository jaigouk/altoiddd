"""Tests for resolve_profile domain service."""

from __future__ import annotations

from src.domain.models.stack_profile import GenericProfile, PythonUvProfile, StackProfile
from src.domain.models.tech_stack import TechStack
from src.domain.services.stack_resolver import resolve_profile


class TestResolveProfilePython:
    """Python TechStack resolves to PythonUvProfile."""

    def test_python_uv(self) -> None:
        ts = TechStack(language="python", package_manager="uv")
        profile = resolve_profile(ts)
        assert isinstance(profile, PythonUvProfile)

    def test_python_pip(self) -> None:
        """Python with non-uv manager still resolves to PythonUvProfile."""
        ts = TechStack(language="python", package_manager="pip")
        profile = resolve_profile(ts)
        assert isinstance(profile, PythonUvProfile)

    def test_python_empty_manager(self) -> None:
        ts = TechStack(language="python", package_manager="")
        profile = resolve_profile(ts)
        assert isinstance(profile, PythonUvProfile)


class TestResolveProfileGeneric:
    """Non-Python TechStack resolves to GenericProfile."""

    def test_unknown_language(self) -> None:
        ts = TechStack(language="unknown", package_manager="")
        profile = resolve_profile(ts)
        assert isinstance(profile, GenericProfile)

    def test_rust_cargo(self) -> None:
        ts = TechStack(language="rust", package_manager="cargo")
        profile = resolve_profile(ts)
        assert isinstance(profile, GenericProfile)

    def test_javascript(self) -> None:
        ts = TechStack(language="javascript", package_manager="npm")
        profile = resolve_profile(ts)
        assert isinstance(profile, GenericProfile)


class TestResolveProfileNone:
    """None tech_stack (old sessions) resolves to GenericProfile."""

    def test_none_returns_generic(self) -> None:
        profile = resolve_profile(None)
        assert isinstance(profile, GenericProfile)


class TestResolveProfileProtocol:
    """All returned profiles satisfy StackProfile protocol."""

    def test_python_satisfies_protocol(self) -> None:
        ts = TechStack(language="python", package_manager="uv")
        profile = resolve_profile(ts)
        assert isinstance(profile, StackProfile)

    def test_generic_satisfies_protocol(self) -> None:
        profile = resolve_profile(None)
        assert isinstance(profile, StackProfile)

    def test_none_satisfies_protocol(self) -> None:
        ts = TechStack(language="rust", package_manager="cargo")
        profile = resolve_profile(ts)
        assert isinstance(profile, StackProfile)
