"""Tests for the CLI mode selection prompt (Express vs Deep).

RED phase: these tests define the contract for the _guide_prompt_mode()
function added to the CLI flow (alty-20c.3).
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

import typer
from typer.testing import CliRunner

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import Persona, Register

runner = CliRunner()


# -- Helpers ------------------------------------------------------------------


def _make_mock_session(
    *,
    status: DiscoveryStatus = DiscoveryStatus.CREATED,
) -> MagicMock:
    """Create a minimal mock DiscoverySession."""
    session = MagicMock(spec=DiscoverySession)
    session.session_id = "test-session-id"
    session.status = status
    session.register = Register.TECHNICAL
    session.persona = Persona.DEVELOPER
    session.readme_content = "# Test"
    session.answers = ()
    session.playback_confirmations = ()
    session.tech_stack = None
    session.mode = None  # Will be overridden by tests
    return session


# -- Unit tests for _guide_prompt_mode ----------------------------------------


class TestGuidePromptMode:
    """Test the _guide_prompt_mode function directly."""

    def test_choice_1_returns_express_no_set_mode_call(self) -> None:
        """Choosing '1' (Express) should not call set_mode — EXPRESS is default."""
        from src.infrastructure.cli.main import _guide_prompt_mode

        mock_discovery = MagicMock()
        session = _make_mock_session()
        mock_discovery.get_session.return_value = session

        with patch.object(typer, "prompt", return_value="1"):
            result = _guide_prompt_mode(mock_discovery, "test-session-id")

        mock_discovery.get_session.assert_called_once_with("test-session-id")
        mock_discovery.set_mode.assert_not_called()
        assert result is session

    def test_choice_2_calls_set_mode_deep(self) -> None:
        """Choosing '2' (Deep) should call set_mode with DEEP."""
        from src.domain.models.discovery_values import DiscoveryMode
        from src.infrastructure.cli.main import _guide_prompt_mode

        mock_discovery = MagicMock()
        deep_session = _make_mock_session()
        mock_discovery.set_mode.return_value = deep_session

        with patch.object(typer, "prompt", return_value="2"):
            result = _guide_prompt_mode(mock_discovery, "test-session-id")

        mock_discovery.set_mode.assert_called_once_with(
            "test-session-id", DiscoveryMode.DEEP
        )
        assert result is deep_session

    def test_default_choice_is_express(self) -> None:
        """Default (empty input → '1') should return Express without set_mode."""
        from src.infrastructure.cli.main import _guide_prompt_mode

        mock_discovery = MagicMock()
        session = _make_mock_session()
        mock_discovery.get_session.return_value = session

        with patch.object(typer, "prompt", return_value="1"):
            _guide_prompt_mode(mock_discovery, "test-session-id")

        mock_discovery.set_mode.assert_not_called()

    def test_prompt_displays_mode_options(self) -> None:
        """Verify typer.echo is called with mode option text."""
        from src.infrastructure.cli.main import _guide_prompt_mode

        mock_discovery = MagicMock()
        session = _make_mock_session()
        mock_discovery.get_session.return_value = session

        with (
            patch.object(typer, "prompt", return_value="1"),
            patch.object(typer, "echo") as mock_echo,
        ):
            _guide_prompt_mode(mock_discovery, "test-session-id")

        # Check that mode options were displayed
        echo_calls = [str(c) for c in mock_echo.call_args_list]
        echo_text = " ".join(echo_calls)
        assert "Express" in echo_text
        assert "Deep" in echo_text
