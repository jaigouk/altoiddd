"""Tests for CodebasePortScanner (infrastructure adapter).

Scans src/application/ports/ using Python AST to extract Protocol
definitions and method signatures — used by regression tests and
future ripple automation.
"""

from __future__ import annotations

from pathlib import Path

from src.infrastructure.external.codebase_port_scanner import (
    CodebasePortScanner,
    MethodSignature,
    PortDefinition,
)


class TestCodebasePortScanner:
    """Tests for AST-based port scanning."""

    def test_finds_ports_in_directory(self) -> None:
        """Scanning real ports/ returns known Protocol classes."""
        ports_dir = Path("src/application/ports")
        result = CodebasePortScanner.scan(ports_dir)

        assert isinstance(result, dict)
        assert "ChallengerPort" in result
        assert "WebSearchPort" in result
        assert isinstance(result["ChallengerPort"], PortDefinition)

    def test_extracts_method_signatures(self) -> None:
        """ChallengerPort has generate_challenges with correct params."""
        ports_dir = Path("src/application/ports")
        result = CodebasePortScanner.scan(ports_dir)

        challenger = result["ChallengerPort"]
        method_names = [m.name for m in challenger.methods]
        assert "generate_challenges" in method_names

        gen = next(m for m in challenger.methods if m.name == "generate_challenges")
        assert isinstance(gen, MethodSignature)
        # Should have at least 'model' and 'max_per_type' params (excluding self)
        assert "model" in gen.parameters
        assert "max_per_type" in gen.parameters

    def test_empty_directory_returns_empty(self, tmp_path: Path) -> None:
        """Directory with no .py files returns empty dict."""
        result = CodebasePortScanner.scan(tmp_path)
        assert result == {}

    def test_malformed_file_skipped(self, tmp_path: Path) -> None:
        """Broken Python file is skipped without crashing."""
        bad_file = tmp_path / "broken_port.py"
        bad_file.write_text("class Foo(\n  # missing closing paren")

        result = CodebasePortScanner.scan(tmp_path)
        assert result == {}


class TestPortScannerRegression:
    """Regression: verify scanner finds real ports referenced in 20c.5."""

    def test_finds_web_search_port(self) -> None:
        """Scanner finds WebSearchPort in real ports directory."""
        ports_dir = Path("src/application/ports")
        result = CodebasePortScanner.scan(ports_dir)

        assert "WebSearchPort" in result
        web_search = result["WebSearchPort"]
        method_names = [m.name for m in web_search.methods]
        assert "search" in method_names
