"""Tests for TechStack value object."""

from __future__ import annotations

import pytest

from src.domain.models.tech_stack import TechStack


class TestTechStackCreation:
    """TechStack can be created with language and package_manager."""

    def test_create_python_uv(self) -> None:
        ts = TechStack(language="python", package_manager="uv")
        assert ts.language == "python"
        assert ts.package_manager == "uv"

    def test_create_unknown_empty(self) -> None:
        ts = TechStack(language="unknown", package_manager="")
        assert ts.language == "unknown"
        assert ts.package_manager == ""


class TestTechStackFrozen:
    """TechStack fields cannot be mutated (frozen dataclass)."""

    def test_cannot_mutate_language(self) -> None:
        ts = TechStack(language="python", package_manager="uv")
        with pytest.raises(AttributeError):
            ts.language = "rust"  # type: ignore[misc]

    def test_cannot_mutate_package_manager(self) -> None:
        ts = TechStack(language="python", package_manager="uv")
        with pytest.raises(AttributeError):
            ts.package_manager = "pip"  # type: ignore[misc]


class TestTechStackEquality:
    """Two TechStacks with same values are equal (dataclass equality)."""

    def test_equal_values(self) -> None:
        ts1 = TechStack(language="python", package_manager="uv")
        ts2 = TechStack(language="python", package_manager="uv")
        assert ts1 == ts2

    def test_different_language(self) -> None:
        ts1 = TechStack(language="python", package_manager="uv")
        ts2 = TechStack(language="rust", package_manager="uv")
        assert ts1 != ts2

    def test_different_package_manager(self) -> None:
        ts1 = TechStack(language="python", package_manager="uv")
        ts2 = TechStack(language="python", package_manager="pip")
        assert ts1 != ts2
