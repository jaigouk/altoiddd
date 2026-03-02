"""Tests for code review fixes B1-B3, M1-M5, s1-s2.

These tests verify:
- B1: guide() reuses _run_discovery() and _save_session() (no duplication)
- B2: _reconstruct_event() exists in one place only (generate imports from main)
- B3: _reconstruct_event() uses if/raise, not assert
- M1: Discovery helpers catch specific domain errors, not bare Exception
- M2+M3: generate subcommands use _build_domain_model() helper (no dead code)
- M5: _load_session() uses explicit Path.cwd() for session path
- s1: Consistent error messages between guide() and init()
"""

from __future__ import annotations

import ast
import inspect
import textwrap
from unittest.mock import patch

from typer.testing import CliRunner

from src.infrastructure.cli.main import app

runner = CliRunner()


# ---------------------------------------------------------------------------
# B1: guide() reuses _run_discovery() — no copy-paste
# ---------------------------------------------------------------------------


class TestB1GuideReusesRunDiscovery:
    """guide() should call _run_discovery() instead of duplicating its logic."""

    def test_guide_function_body_does_not_contain_question_loop(self) -> None:
        """guide() should NOT contain a for-loop over Question.CATALOG.

        If it does, it's duplicating _run_discovery() logic.
        """
        from src.infrastructure.cli import main

        source = inspect.getsource(main.guide)
        tree = ast.parse(textwrap.dedent(source))

        # Walk AST looking for a for-loop with "CATALOG" in it
        has_catalog_loop = False
        for node in ast.walk(tree):
            if isinstance(node, ast.For):
                for_source = ast.dump(node)
                if "CATALOG" in for_source:
                    has_catalog_loop = True
                    break

        assert not has_catalog_loop, (
            "guide() contains a Question.CATALOG loop — "
            "it should call _run_discovery() instead"
        )

    def test_guide_function_calls_run_discovery(self) -> None:
        """guide() should contain a call to _run_discovery()."""
        from src.infrastructure.cli import main

        source = inspect.getsource(main.guide)
        assert "_run_discovery" in source, (
            "guide() does not call _run_discovery() — it should delegate to it"
        )

    def test_guide_function_calls_save_session(self) -> None:
        """guide() should call _save_session() instead of inline json.dumps."""
        from src.infrastructure.cli import main

        source = inspect.getsource(main.guide)
        assert "_save_session" in source, (
            "guide() does not call _save_session() — it should delegate to it"
        )


# ---------------------------------------------------------------------------
# B2: _reconstruct_event() consolidated — not duplicated across modules
# ---------------------------------------------------------------------------


class TestB2ReconstructEventConsolidated:
    """_reconstruct_event() should exist in only one place."""

    def test_generate_imports_reconstruct_event_from_main(self) -> None:
        """generate.py should import _reconstruct_event from main, not define its own."""
        from src.infrastructure.cli import generate

        source = inspect.getsource(generate)

        # Should NOT define its own _reconstruct_event function
        tree = ast.parse(source)
        local_defs = [
            node.name
            for node in ast.walk(tree)
            if isinstance(node, ast.FunctionDef) and node.name == "_reconstruct_event"
        ]
        assert len(local_defs) == 0, (
            "generate.py defines its own _reconstruct_event() — "
            "it should import from main"
        )


# ---------------------------------------------------------------------------
# B3: _reconstruct_event() uses if/raise, not assert
# ---------------------------------------------------------------------------


class TestB3NoAssertInReconstructEvent:
    """_reconstruct_event() must NOT use assert for runtime validation."""

    def test_reconstruct_event_has_no_assert_statements(self) -> None:
        """assert is stripped with -O; use if/raise instead."""
        from src.infrastructure.cli import main

        source = inspect.getsource(main._reconstruct_event)
        tree = ast.parse(textwrap.dedent(source))

        assert_count = sum(
            1 for node in ast.walk(tree) if isinstance(node, ast.Assert)
        )
        assert assert_count == 0, (
            f"_reconstruct_event() has {assert_count} assert statement(s) — "
            "replace with if/raise InvariantViolationError"
        )


# ---------------------------------------------------------------------------
# M1: Discovery helpers catch domain errors, not bare Exception
# ---------------------------------------------------------------------------


class TestM1NarrowExceptionHandling:
    """CLI error handling should catch specific domain errors, not bare Exception."""

    def test_guide_prompt_persona_catches_domain_errors(self) -> None:
        from src.infrastructure.cli import main

        source = inspect.getsource(main._guide_prompt_persona)
        assert "except Exception" not in source, (
            "_guide_prompt_persona catches bare Exception — "
            "use specific domain errors"
        )

    def test_guide_handle_question_catches_domain_errors(self) -> None:
        from src.infrastructure.cli import main

        source = inspect.getsource(main._guide_handle_question)
        assert "except Exception" not in source, (
            "_guide_handle_question catches bare Exception — "
            "use specific domain errors"
        )

    def test_guide_handle_playback_catches_domain_errors(self) -> None:
        from src.infrastructure.cli import main

        source = inspect.getsource(main._guide_handle_playback)
        assert "except Exception" not in source, (
            "_guide_handle_playback catches bare Exception — "
            "use specific domain errors"
        )

    def test_run_discovery_complete_catches_domain_errors(self) -> None:
        from src.infrastructure.cli import main

        source = inspect.getsource(main._run_discovery)
        assert "except Exception" not in source, (
            "_run_discovery catches bare Exception — "
            "use specific domain errors"
        )


# ---------------------------------------------------------------------------
# M2+M3: generate subcommands use _build_domain_model helper
# ---------------------------------------------------------------------------


class TestM2M3GenerateUsesHelper:
    """Generate subcommands use _build_domain_model() — no inline duplication."""

    def test_build_domain_model_is_called_by_fitness(self) -> None:
        from src.infrastructure.cli import generate

        source = inspect.getsource(generate.fitness)
        assert "_build_domain_model" in source, (
            "fitness() should call _build_domain_model() instead of "
            "inlining ArtifactGenerationHandler construction"
        )

    def test_build_domain_model_is_called_by_tickets(self) -> None:
        from src.infrastructure.cli import generate

        source = inspect.getsource(generate.tickets)
        assert "_build_domain_model" in source, (
            "tickets() should call _build_domain_model() instead of "
            "inlining ArtifactGenerationHandler construction"
        )

    def test_build_domain_model_is_called_by_configs(self) -> None:
        from src.infrastructure.cli import generate

        source = inspect.getsource(generate.configs)
        assert "_build_domain_model" in source, (
            "configs() should call _build_domain_model() instead of "
            "inlining ArtifactGenerationHandler construction"
        )


# ---------------------------------------------------------------------------
# M5: _load_session() uses Path.cwd() explicitly
# ---------------------------------------------------------------------------


class TestM5ExplicitPathCwd:
    """_load_session() should use Path.cwd() for session path resolution."""

    def test_load_session_uses_cwd(self) -> None:
        from src.infrastructure.cli import generate

        source = inspect.getsource(generate._load_session)
        assert "Path.cwd()" in source or "cwd()" in source, (
            "_load_session() uses relative Path('.alty/...') — "
            "should use Path.cwd() / '.alty' / 'session.json'"
        )


# ---------------------------------------------------------------------------
# s1: Consistent error messages
# ---------------------------------------------------------------------------


class TestS1ConsistentErrorMessages:
    """Error messages should be consistent between guide() and init()."""

    @patch("src.infrastructure.composition.create_app")
    def test_guide_missing_readme_message_format(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        """guide() should say 'No README.md found' (matching init)."""
        monkeypatch.chdir(tmp_path)
        result = runner.invoke(app, ["guide"])
        # Should NOT start with "Error:" — just the message
        output_lower = result.output.lower()
        assert "no readme.md found" in output_lower, (
            f"guide() error message doesn't match expected format. Got: {result.output!r}"
        )
