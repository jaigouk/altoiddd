"""Tests for Typer CLI entry point and subcommand stubs.

Verifies that all 15 CLI commands exist, exit cleanly,
and that subcommand groups are properly wired.

Reference: ARCHITECTURE.md §6.1 CLI Command Tree
"""

from __future__ import annotations

import pytest
from typer.testing import CliRunner

from src.infrastructure.cli.main import app

runner = CliRunner()


# ── Help output ──────────────────────────────────────────────


class TestHelpOutput:
    """alty --help lists all top-level commands and subcommand groups."""

    TOP_LEVEL_COMMANDS = [
        "init",
        "guide",
        "detect",
        "check",
        "kb",
        "doc-health",
        "doc-review",
        "ticket-health",
        "version",
    ]

    SUBCOMMAND_GROUPS = [
        "generate",
        "persona",
    ]

    def test_help_exits_cleanly(self):
        result = runner.invoke(app, ["--help"])
        assert result.exit_code == 0

    @pytest.mark.parametrize("command", TOP_LEVEL_COMMANDS + SUBCOMMAND_GROUPS)
    def test_help_lists_command(self, command):
        result = runner.invoke(app, ["--help"])
        assert command in result.output, f"'{command}' not found in help output"

    def test_no_args_shows_help(self):
        """Running alty with no arguments shows help text (exit 0 or 2)."""
        result = runner.invoke(app, [])
        # Typer/Click exits with 2 for no_args_is_help, which is acceptable
        assert result.exit_code in (0, 2)
        assert "init" in result.output


# ── Top-level command stubs ──────────────────────────────────


class TestTopLevelCommandStubs:
    """Each top-level command stub exits cleanly (exit code 0)."""

    STUB_COMMANDS = [
        "init",
        "guide",
        "detect",
        "kb",
        "doc-health",
        "doc-review",
        "ticket-health",
    ]

    @pytest.mark.parametrize("command", STUB_COMMANDS)
    def test_stub_exits_cleanly(self, command):
        result = runner.invoke(app, [command])
        assert result.exit_code == 0

    def test_version_shows_version_string(self):
        """alty version prints the package version."""
        result = runner.invoke(app, ["version"])
        assert result.exit_code == 0
        assert "0.1.0" in result.output


# ── Generate subcommand group ────────────────────────────────


class TestGenerateSubcommands:
    """alty generate subcommand group with 4 stubs."""

    SUBCOMMANDS = ["artifacts", "fitness", "tickets", "configs"]

    def test_generate_help_exits_cleanly(self):
        result = runner.invoke(app, ["generate", "--help"])
        assert result.exit_code == 0

    @pytest.mark.parametrize("subcommand", SUBCOMMANDS)
    def test_generate_help_lists_subcommand(self, subcommand):
        result = runner.invoke(app, ["generate", "--help"])
        assert subcommand in result.output, f"'{subcommand}' not found in generate help output"

    @pytest.mark.parametrize("subcommand", SUBCOMMANDS)
    def test_generate_stub_exits_cleanly(self, subcommand):
        result = runner.invoke(app, ["generate", subcommand])
        assert result.exit_code == 0


# ── Persona subcommand group ─────────────────────────────────


class TestPersonaSubcommands:
    """alty persona subcommand group with 2 stubs."""

    def test_persona_help_exits_cleanly(self):
        result = runner.invoke(app, ["persona", "--help"])
        assert result.exit_code == 0

    def test_persona_help_lists_list(self):
        result = runner.invoke(app, ["persona", "--help"])
        assert "list" in result.output

    def test_persona_help_lists_generate(self):
        result = runner.invoke(app, ["persona", "--help"])
        assert "generate" in result.output

    def test_persona_list_exits_cleanly(self):
        result = runner.invoke(app, ["persona", "list"])
        assert result.exit_code == 0

    def test_persona_generate_exits_cleanly(self):
        result = runner.invoke(app, ["persona", "generate"])
        assert result.exit_code == 0


# ── Edge cases ───────────────────────────────────────────────


class TestEdgeCases:
    """CLI edge cases: unknown commands, deep help."""

    def test_unknown_command_fails(self):
        result = runner.invoke(app, ["nonexistent"])
        assert result.exit_code != 0

    def test_unknown_generate_subcommand_fails(self):
        result = runner.invoke(app, ["generate", "nonexistent"])
        assert result.exit_code != 0

    def test_generate_artifacts_help_exits_cleanly(self):
        """Deep help: alty generate artifacts --help."""
        result = runner.invoke(app, ["generate", "artifacts", "--help"])
        assert result.exit_code == 0
