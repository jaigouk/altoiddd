"""Tests for the alty guide CLI command.

Verifies that `alty guide` wires to DiscoveryPort via composition root,
runs the interactive discovery flow, and handles errors correctly.
"""

from __future__ import annotations

import json
from unittest.mock import MagicMock, patch

from typer.testing import CliRunner

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import Answer, Persona, Register
from src.domain.models.question import Question
from src.infrastructure.cli.main import app

runner = CliRunner()


# ── Helpers ─────────────────────────────────────────────────────


def _make_session(
    *,
    status: DiscoveryStatus = DiscoveryStatus.CREATED,
    register: Register = Register.TECHNICAL,
    persona: Persona = Persona.DEVELOPER,
    answers: tuple[Answer, ...] = (),
) -> MagicMock:
    """Create a mock DiscoverySession with sensible defaults."""
    session = MagicMock(spec=DiscoverySession)
    session.session_id = "test-session-id"
    session.status = status
    session.register = register
    session.persona = persona
    session.readme_content = "# Test Project\nA test project."
    session.answers = answers
    session.playback_confirmations = ()
    session.to_snapshot.return_value = {
        "session_id": "test-session-id",
        "status": status.value,
        "persona": persona.value,
        "register": register.value,
        "answers": [
            {"question_id": a.question_id, "response_text": a.response_text}
            for a in answers
        ],
        "skipped": [],
        "playback_confirmations": [],
        "answers_since_last_playback": 0,
        "readme_content": "# Test Project",
    }
    return session


def _setup_happy_path(mock_create_app: MagicMock) -> MagicMock:
    """Configure mock for a full happy-path run through all 10 questions.

    Returns the mock AppContext.
    """
    mock_ctx = MagicMock()
    mock_create_app.return_value = mock_ctx

    mock_ctx.discovery.start_session.return_value = _make_session()
    mock_ctx.discovery.set_tech_stack.return_value = _make_session()
    mock_ctx.discovery.detect_persona.return_value = _make_session(
        status=DiscoveryStatus.PERSONA_DETECTED
    )

    answering = _make_session(status=DiscoveryStatus.ANSWERING)
    playback = _make_session(status=DiscoveryStatus.PLAYBACK_PENDING)
    completed = _make_session(status=DiscoveryStatus.COMPLETED)

    # 10 questions: answers trigger playback every 3 answers
    mock_ctx.discovery.answer_question.side_effect = [
        answering, answering, playback,     # Q1, Q2, Q3 -> playback
        answering, answering, playback,     # Q4, Q5, Q6 -> playback
        answering, answering, playback,     # Q7, Q8, Q9 -> playback
        answering,                          # Q10
    ]
    mock_ctx.discovery.confirm_playback.return_value = answering
    mock_ctx.discovery.complete.return_value = completed
    return mock_ctx


def _happy_path_input() -> str:
    """Build stdin: tech stack + persona + 10 answers + 3 playbacks."""
    return (
        "y\n"                   # tech stack: Python
        "1\n"                   # persona
        "answer1\n"             # Q1
        "answer2\n"             # Q2
        "answer3\n"             # Q3
        "y\n"                   # playback 1 confirm
        "answer4\n"             # Q4
        "answer5\n"             # Q5
        "answer6\n"             # Q6
        "y\n"                   # playback 2 confirm
        "answer7\n"             # Q7
        "answer8\n"             # Q8
        "answer9\n"             # Q9
        "y\n"                   # playback 3 confirm
        "answer10\n"            # Q10
    )


# ── Wiring ──────────────────────────────────────────────────────


class TestGuideWiring:
    """alty guide uses composition root, not standalone construction."""

    @patch("src.infrastructure.composition.create_app")
    def test_calls_create_app(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test Project")
        _setup_happy_path(mock_create_app)
        runner.invoke(app, ["guide"], input=_happy_path_input())
        mock_create_app.assert_called_once()

    @patch("src.infrastructure.composition.create_app")
    def test_uses_discovery_port(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test Project")
        mock_ctx = _setup_happy_path(mock_create_app)
        runner.invoke(app, ["guide"], input=_happy_path_input())
        mock_ctx.discovery.start_session.assert_called_once_with("# Test Project")


# ── Persona Selection ───────────────────────────────────────────


class TestGuidePersonaSelection:
    """guide prompts for persona and calls detect_persona."""

    @patch("src.infrastructure.composition.create_app")
    def test_prompts_persona_and_calls_detect(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = _setup_happy_path(mock_create_app)
        input_text = (
            "y\n" + "2\n" + "answer\n" * 3 + "y\n" + "answer\n" * 3
            + "y\n" + "answer\n" * 3 + "y\n" + "answer\n"
        )
        runner.invoke(app, ["guide"], input=input_text)
        mock_ctx.discovery.detect_persona.assert_called_once_with(
            "test-session-id", "2"
        )

    @patch("src.infrastructure.composition.create_app")
    def test_shows_persona_options(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        _setup_happy_path(mock_create_app)
        result = runner.invoke(app, ["guide"], input=_happy_path_input())
        assert "Developer" in result.output
        assert "Product Owner" in result.output
        assert "Domain Expert" in result.output
        assert "Mixed" in result.output


# ── Question Flow ───────────────────────────────────────────────


class TestGuideQuestionFlow:
    """guide presents questions and collects answers."""

    @patch("src.infrastructure.composition.create_app")
    def test_presents_questions_in_technical_register(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        _setup_happy_path(mock_create_app)
        result = runner.invoke(app, ["guide"], input=_happy_path_input())
        assert Question.CATALOG[0].technical_text in result.output

    @patch("src.infrastructure.composition.create_app")
    def test_presents_questions_in_non_technical_register(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        mock_ctx.discovery.start_session.return_value = _make_session()
        mock_ctx.discovery.set_tech_stack.return_value = _make_session()
        mock_ctx.discovery.detect_persona.return_value = _make_session(
            status=DiscoveryStatus.PERSONA_DETECTED,
            register=Register.NON_TECHNICAL,
            persona=Persona.PRODUCT_OWNER,
        )
        answering = _make_session(
            status=DiscoveryStatus.ANSWERING,
            register=Register.NON_TECHNICAL,
        )
        playback = _make_session(
            status=DiscoveryStatus.PLAYBACK_PENDING,
            register=Register.NON_TECHNICAL,
        )
        completed = _make_session(
            status=DiscoveryStatus.COMPLETED,
            register=Register.NON_TECHNICAL,
        )
        mock_ctx.discovery.answer_question.side_effect = [
            answering, answering, playback,
            answering, answering, playback,
            answering, answering, playback,
            answering,
        ]
        mock_ctx.discovery.confirm_playback.return_value = answering
        mock_ctx.discovery.complete.return_value = completed

        result = runner.invoke(app, ["guide"], input=_happy_path_input())
        assert Question.CATALOG[0].non_technical_text in result.output

    @patch("src.infrastructure.composition.create_app")
    def test_skip_question(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        mock_ctx.discovery.start_session.return_value = _make_session()
        mock_ctx.discovery.set_tech_stack.return_value = _make_session()
        mock_ctx.discovery.detect_persona.return_value = _make_session(
            status=DiscoveryStatus.PERSONA_DETECTED,
        )
        answering = _make_session(status=DiscoveryStatus.ANSWERING)
        playback = _make_session(status=DiscoveryStatus.PLAYBACK_PENDING)
        completed = _make_session(status=DiscoveryStatus.COMPLETED)
        mock_ctx.discovery.skip_question.return_value = answering
        mock_ctx.discovery.answer_question.side_effect = [
            answering, playback,            # Q2, Q3 -> playback
            answering, answering, playback,  # Q4, Q5, Q6 -> playback
            answering, answering, playback,  # Q7, Q8, Q9 -> playback
            answering,                       # Q10
        ]
        mock_ctx.discovery.confirm_playback.return_value = answering
        mock_ctx.discovery.complete.return_value = completed

        input_text = (
            "y\n"                   # tech stack
            "1\n"                   # persona
            "skip\ntoo early\n"     # Q1 skip + reason
            "answer\nanswer\n"      # Q2, Q3
            "y\n"                   # playback 1
            "answer\nanswer\nanswer\n"  # Q4, Q5, Q6
            "y\n"                   # playback 2
            "answer\nanswer\nanswer\n"  # Q7, Q8, Q9
            "y\n"                   # playback 3
            "answer\n"              # Q10
        )
        runner.invoke(app, ["guide"], input=input_text)
        mock_ctx.discovery.skip_question.assert_called_once_with(
            "test-session-id", "Q1", "too early"
        )


# ── Playback ────────────────────────────────────────────────────


class TestGuidePlayback:
    """guide handles playback confirmation loop."""

    @patch("src.infrastructure.composition.create_app")
    def test_playback_confirm_count(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = _setup_happy_path(mock_create_app)
        runner.invoke(app, ["guide"], input=_happy_path_input())
        assert mock_ctx.discovery.confirm_playback.call_count == 3

    @patch("src.infrastructure.composition.create_app")
    def test_playback_with_corrections(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = _setup_happy_path(mock_create_app)

        # First playback: reject with corrections
        input_text = (
            "y\n"                        # tech stack
            "1\n"
            "answer\nanswer\nanswer\n"
            "n\nfix Q1 wording\n"       # reject + corrections
            "answer\nanswer\nanswer\n"
            "y\n"                        # confirm
            "answer\nanswer\nanswer\n"
            "y\n"                        # confirm
            "answer\n"
        )
        runner.invoke(app, ["guide"], input=input_text)
        first_call = mock_ctx.discovery.confirm_playback.call_args_list[0]
        assert first_call[0] == ("test-session-id", False, "fix Q1 wording")


# ── Completion ──────────────────────────────────────────────────


class TestGuideCompletion:
    """guide completes session and shows summary."""

    @patch("src.infrastructure.composition.create_app")
    def test_calls_complete(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = _setup_happy_path(mock_create_app)
        result = runner.invoke(app, ["guide"], input=_happy_path_input())
        assert result.exit_code == 0
        mock_ctx.discovery.complete.assert_called_once_with("test-session-id")

    @patch("src.infrastructure.composition.create_app")
    def test_shows_completion_summary(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        _setup_happy_path(mock_create_app)
        result = runner.invoke(app, ["guide"], input=_happy_path_input())
        assert "complete" in result.output.lower()


# ── Error Handling ──────────────────────────────────────────────


class TestGuideErrorHandling:
    """guide handles errors: missing README, domain errors."""

    @patch("src.infrastructure.composition.create_app")
    def test_missing_readme_exits_with_error(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        # No README.md created
        result = runner.invoke(app, ["guide"])
        assert result.exit_code == 1

    @patch("src.infrastructure.composition.create_app")
    def test_missing_readme_shows_error_message(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        result = runner.invoke(app, ["guide"])
        assert "readme" in result.output.lower()

    @patch("src.infrastructure.composition.create_app")
    def test_invalid_persona_exits_with_error(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        mock_ctx.discovery.start_session.return_value = _make_session()
        mock_ctx.discovery.set_tech_stack.return_value = _make_session()
        mock_ctx.discovery.detect_persona.side_effect = ValueError("Invalid choice '9'")
        result = runner.invoke(app, ["guide"], input="y\n9\n")
        assert result.exit_code == 1


# ── Session Save ────────────────────────────────────────────────


class TestGuideSessionSave:
    """guide saves session snapshot to .alty/session.json."""

    @patch("src.infrastructure.composition.create_app")
    def test_saves_session_snapshot(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        _setup_happy_path(mock_create_app)
        result = runner.invoke(app, ["guide"], input=_happy_path_input())
        assert result.exit_code == 0
        session_file = tmp_path / ".alty" / "session.json"
        assert session_file.exists()
        data = json.loads(session_file.read_text())
        assert data["session_id"] == "test-session-id"

    @patch("src.infrastructure.composition.create_app")
    def test_creates_alty_dir_if_missing(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        _setup_happy_path(mock_create_app)
        assert not (tmp_path / ".alty").exists()
        runner.invoke(app, ["guide"], input=_happy_path_input())
        assert (tmp_path / ".alty").is_dir()
