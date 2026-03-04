"""CLI entry point for alty.

Root Typer application with 9 top-level commands and 2 subcommand groups
(generate, persona). Each command is a stub that will be replaced by
application-layer handler calls in downstream tickets.

Entry point: alty = "src.infrastructure.cli.main:app"
Reference: ARCHITECTURE.md §6.1 CLI Command Tree
"""

from __future__ import annotations

from importlib.metadata import version
from typing import TYPE_CHECKING

import typer

from src.infrastructure.cli import generate, persona

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.discovery_port import DiscoveryPort
    from src.domain.events.discovery_events import DiscoveryCompleted
    from src.domain.models.discovery_session import DiscoverySession
    from src.domain.models.discovery_values import Answer
    from src.domain.models.stack_profile import StackProfile
    from src.domain.models.tech_stack import TechStack
    from src.infrastructure.composition import AppContext

app = typer.Typer(
    name="alty",
    help="The AI architect for vibe coding — guided project bootstrapper.",
    no_args_is_help=True,
)

app.add_typer(generate.app, name="generate")
app.add_typer(persona.app, name="persona")


# ── Top-level commands ───────────────────────────────────────


@app.command()
def init(
    existing: bool = typer.Option(False, "--existing", help="Rescue an existing project."),
) -> None:
    """Bootstrap a new project from a README idea."""
    if existing:
        from pathlib import Path

        from src.application.commands.rescue_handler import RescueHandler
        from src.domain.models.errors import InvariantViolationError
        from src.domain.services.stack_resolver import resolve_profile
        from src.infrastructure.git.git_ops_adapter import GitOpsAdapter
        from src.infrastructure.scanner.project_scanner import ProjectScanner

        scanner = ProjectScanner()
        git_ops = GitOpsAdapter()
        handler = RescueHandler(project_scan=scanner, git_ops=git_ops)
        project_dir = Path.cwd()
        try:
            # Validate preconditions before asking tech stack
            handler.validate_preconditions(project_dir)

            # Ask tech stack after validation so user doesn't waste effort
            tech_stack = _ask_tech_stack()
            profile = resolve_profile(tech_stack)

            analysis = handler.rescue(project_dir, profile=profile, validated=True)
            if not analysis.gaps:
                typer.echo("No gaps found -- project already follows alty conventions.")
                return
            typer.echo(f"Found {len(analysis.gaps)} gap(s):")
            for gap in analysis.gaps:
                typer.echo(f"  [{gap.gap_type.value}] {gap.path}: {gap.description}")
        except InvariantViolationError as e:
            typer.echo(f"Rescue failed: {e}", err=True)
            raise typer.Exit(code=1) from None
    else:
        _init_new_project()


def _init_new_project() -> None:
    """Run the full bootstrap pipeline: discovery → artifacts → fitness → tickets → configs."""
    from pathlib import Path

    from src.infrastructure.composition import create_app

    ctx = create_app()

    # 1. Load README.md
    readme_path = Path.cwd() / "README.md"
    if not readme_path.exists():
        typer.echo("No README.md found. Create one first.", err=True)
        raise typer.Exit(code=1)

    # 2. Run interactive discovery
    session = _run_discovery(ctx, readme_path.read_text())

    # 3. Reconstruct event from completed session
    event = _reconstruct_event(session)

    # 4. Resolve stack profile
    from src.domain.services.stack_resolver import resolve_profile

    profile = resolve_profile(session.tech_stack)

    # 5. Generate all artifacts with preview-approve pattern
    _run_generation_pipeline(ctx, event, Path.cwd(), profile)

    # 6. Save session snapshot
    _save_session(session)

    typer.echo("\nBootstrap complete!")


def _run_discovery(
    ctx: AppContext,
    readme_content: str,
) -> DiscoverySession:
    """Run the 10-question interactive discovery flow and return completed session."""
    from src.domain.models.discovery_session import DiscoveryStatus
    from src.domain.models.discovery_values import Register
    from src.domain.models.errors import DomainError
    from src.domain.models.question import Question

    session = ctx.discovery.start_session(readme_content)
    typer.echo(f"Discovery session started ({session.session_id})\n")

    session = _guide_prompt_tech_stack(ctx.discovery, session.session_id)
    session = _guide_prompt_persona(ctx.discovery, session.session_id)
    register = session.register

    for question in Question.CATALOG:
        q_text = (
            question.technical_text
            if register == Register.TECHNICAL
            else question.non_technical_text
        )
        session = _guide_handle_question(
            ctx.discovery,
            session.session_id,
            question.id,
            q_text,
            question.phase.value,
        )
        if session.status == DiscoveryStatus.PLAYBACK_PENDING:
            session = _guide_handle_playback(
                ctx.discovery, session.session_id, session.answers
            )

    try:
        session = ctx.discovery.complete(session.session_id)
    except (DomainError, ValueError, KeyError) as e:
        _guide_error(str(e))

    return session


def _reconstruct_event(session: DiscoverySession) -> DiscoveryCompleted:
    """Reconstruct a DiscoveryCompleted event from a completed session."""
    from src.domain.events.discovery_events import DiscoveryCompleted
    from src.domain.models.errors import InvariantViolationError

    if session.persona is None:
        raise InvariantViolationError("session.persona must not be None")
    if session.register is None:
        raise InvariantViolationError("session.register must not be None")

    return DiscoveryCompleted(
        session_id=session.session_id,
        persona=session.persona,
        register=session.register,
        answers=session.answers,
        playback_confirmations=session.playback_confirmations,
        tech_stack=session.tech_stack,
    )


def _run_generation_pipeline(
    ctx: AppContext,
    event: DiscoveryCompleted,
    output_dir: Path,
    profile: StackProfile | None = None,
) -> None:
    """Run the four-stage generation pipeline with preview-approve at each stage."""
    from src.application.commands.artifact_generation_handler import (
        ArtifactGenerationHandler,
    )
    from src.application.commands.config_generation_handler import (
        ConfigGenerationHandler,
    )
    from src.application.commands.fitness_generation_handler import (
        FitnessGenerationHandler,
    )
    from src.application.commands.ticket_generation_handler import (
        TicketGenerationHandler,
    )
    from src.domain.models.tool_config import SupportedTool

    # a. Artifacts
    artifact_handler = ArtifactGenerationHandler(
        renderer=ctx.artifact_renderer,
        writer=ctx.file_writer,
    )
    artifact_preview = artifact_handler.build_preview(event)

    typer.echo("\nArtifact Generation Preview")
    typer.echo(f"  PRD: {len(artifact_preview.prd_content)} chars")
    typer.echo(f"  DDD: {len(artifact_preview.ddd_content)} chars")
    typer.echo(f"  Architecture: {len(artifact_preview.architecture_content)} chars")

    if not typer.confirm("Write artifacts?"):
        typer.echo("Cancelled.")
        raise typer.Exit(code=0)
    artifact_handler.write_artifacts(artifact_preview, output_dir / "docs")

    model = artifact_preview.model

    # b. Fitness
    fitness_handler = FitnessGenerationHandler(writer=ctx.file_writer)
    if profile is not None:
        root_package = profile.to_root_package(output_dir.name)
    else:
        root_package = output_dir.name.replace("-", "_")
    fitness_preview = fitness_handler.build_preview(model, root_package, profile=profile)

    if fitness_preview is None:
        typer.echo("\nFitness tests not available for your stack (requires Python with uv).")
    else:
        typer.echo(f"\n{fitness_preview.summary}")

        if not typer.confirm("Write fitness tests?"):
            typer.echo("Cancelled.")
            raise typer.Exit(code=0)
        fitness_handler.approve_and_write(fitness_preview, output_dir)

    # c. Tickets
    ticket_handler = TicketGenerationHandler(writer=ctx.file_writer)
    ticket_preview = ticket_handler.build_preview(model, profile)

    typer.echo(f"\n{ticket_preview.summary}")

    if not typer.confirm("Write tickets?"):
        typer.echo("Cancelled.")
        raise typer.Exit(code=0)
    ticket_handler.approve_and_write(ticket_preview, output_dir)

    # d. Configs
    config_handler = ConfigGenerationHandler(writer=ctx.file_writer)
    tools = tuple(SupportedTool)
    config_preview = config_handler.build_preview(model, tools, profile)

    typer.echo(f"\n{config_preview.summary}")

    if not typer.confirm("Write configs?"):
        typer.echo("Cancelled.")
        raise typer.Exit(code=0)
    config_handler.approve_and_write(config_preview, output_dir)


def _save_session(session: DiscoverySession) -> None:
    """Save session snapshot to .alty/session.json."""
    import json
    from pathlib import Path

    alty_dir = Path.cwd() / ".alty"
    alty_dir.mkdir(parents=True, exist_ok=True)
    session_file = alty_dir / "session.json"
    session_file.write_text(json.dumps(session.to_snapshot(), indent=2))
    typer.echo(f"Session saved to {session_file}")


def _guide_error(msg: str) -> None:
    """Print error and raise SystemExit via typer."""
    typer.echo(f"Error: {msg}", err=True)
    raise typer.Exit(code=1)


def _ask_tech_stack() -> TechStack:
    """Prompt for tech stack selection and return a TechStack value object."""
    from src.domain.models.tech_stack import TechStack

    is_python = typer.confirm("Are you using Python with uv and pyproject.toml?")
    if is_python:
        return TechStack(language="python", package_manager="uv")
    return TechStack(language="unknown", package_manager="")


def _guide_prompt_tech_stack(
    discovery: DiscoveryPort, session_id: str
) -> DiscoverySession:
    """Prompt for tech stack selection and call set_tech_stack."""
    from src.domain.models.errors import DomainError

    tech_stack = _ask_tech_stack()
    try:
        return discovery.set_tech_stack(session_id, tech_stack)
    except (DomainError, ValueError, KeyError) as e:
        _guide_error(str(e))
        raise  # unreachable, but satisfies type checker


def _guide_prompt_persona(
    discovery: DiscoveryPort, session_id: str
) -> DiscoverySession:
    """Prompt for persona selection and call detect_persona."""
    from src.domain.models.errors import DomainError

    typer.echo("Select your persona:")
    typer.echo("  1) Developer (technical register)")
    typer.echo("  2) Product Owner")
    typer.echo("  3) Domain Expert")
    typer.echo("  4) Mixed")
    choice = typer.prompt("Persona [1-4]")
    try:
        return discovery.detect_persona(session_id, choice)
    except (DomainError, ValueError, KeyError) as e:
        _guide_error(str(e))
        raise  # unreachable, but satisfies type checker


def _guide_handle_question(
    discovery: DiscoveryPort,
    session_id: str,
    question_id: str,
    q_text: str,
    phase: str,
) -> DiscoverySession:
    """Present a question, collect answer or skip, return updated session."""
    from src.domain.models.errors import DomainError

    typer.echo(f"\n[{question_id}] ({phase}) {q_text}")
    answer = typer.prompt("Answer (or 'skip' to skip)")
    try:
        if answer.strip().lower() == "skip":
            reason = typer.prompt("Skip reason")
            return discovery.skip_question(session_id, question_id, reason)
        return discovery.answer_question(session_id, question_id, answer)
    except (DomainError, ValueError, KeyError) as e:
        _guide_error(str(e))
        raise  # unreachable


def _guide_handle_playback(
    discovery: DiscoveryPort,
    session_id: str,
    answers: tuple[Answer, ...],
) -> DiscoverySession:
    """Show playback summary and prompt for confirmation."""
    from src.domain.models.errors import DomainError

    typer.echo("\n--- Playback Summary ---")
    recent = answers[-3:] if len(answers) >= 3 else answers
    for a in recent:
        typer.echo(f"  {a.question_id}: {a.response_text}")
    typer.echo("---")

    confirmed = typer.confirm("Confirm playback?")
    corrections = ""
    if not confirmed:
        corrections = typer.prompt("Corrections")
    try:
        return discovery.confirm_playback(session_id, confirmed, corrections)
    except (DomainError, ValueError, KeyError) as e:
        _guide_error(str(e))
        raise  # unreachable


@app.command()
def guide() -> None:
    """Run the 10-question guided DDD discovery flow."""
    from pathlib import Path

    from src.infrastructure.composition import create_app

    ctx = create_app()

    # Load README from current directory
    readme_path = Path.cwd() / "README.md"
    if not readme_path.exists():
        typer.echo("No README.md found. Create one first.", err=True)
        raise typer.Exit(code=1)

    session = _run_discovery(ctx, readme_path.read_text())
    _save_session(session)

    typer.echo("\nDiscovery session complete!")


@app.command()
def detect(
    project_dir: str = typer.Argument(
        ".",
        help="Project directory to scan (defaults to current directory).",
    ),
) -> None:
    """Scan for installed AI coding tools and global settings."""
    from pathlib import Path

    from src.application.commands.detection_handler import DetectionHandler
    from src.domain.models.detection_result import ConflictSeverity
    from src.infrastructure.external.filesystem_tool_scanner import FilesystemToolScanner

    resolved_dir = Path(project_dir).resolve()
    scanner = FilesystemToolScanner()
    handler = DetectionHandler(tool_detection=scanner)
    result = handler.detect(resolved_dir)

    if not result.detected_tools:
        typer.echo("No AI coding tools detected.")
        return

    typer.echo("Detected AI coding tools:")
    for tool in result.detected_tools:
        config_info = f" ({tool.config_path})" if tool.config_path else ""
        typer.echo(f"  - {tool.name}{config_info}")

    if result.conflicts:
        typer.echo("\nConfiguration conflicts:")
        for conflict in result.conflicts:
            severity = result.severity_map.get(conflict, ConflictSeverity.WARNING)
            typer.echo(f"  [{severity.value.upper()}] {conflict}")


@app.command()
def check(
    gate: str = typer.Option(
        "", "--gate", help="Run a specific gate: lint, types, tests, fitness."
    ),
) -> None:
    """Run quality gates (lint, types, tests, fitness)."""
    from src.domain.models.quality_gate import QualityGate
    from src.infrastructure.composition import create_app

    ctx = create_app()

    gates_arg: tuple[QualityGate, ...] | None = None
    if gate:
        try:
            gates_arg = (QualityGate(gate),)
        except ValueError:
            valid = ", ".join(g.value for g in QualityGate)
            typer.echo(f"Invalid gate: {gate!r}. Valid gates: {valid}", err=True)
            raise typer.Exit(code=1) from None

    report = ctx.quality_gate.check(gates=gates_arg)

    for result in report.results:
        status = "PASS" if result.passed else "FAIL"
        typer.echo(f"  [{status}] {result.gate.value} ({result.duration_ms}ms)")
        if not result.passed:
            for line in result.output.strip().splitlines():
                typer.echo(f"         {line}")

    if report.passed:
        typer.echo(f"\nAll {len(report.results)} quality gate(s) passed.")
    else:
        failed = sum(1 for r in report.results if not r.passed)
        typer.echo(f"\n{failed} quality gate(s) failed.", err=True)
        raise typer.Exit(code=1)


@app.command()
def kb(topic: str = typer.Argument("", help="Knowledge base topic to look up.")) -> None:
    """Look up a topic in the RLM knowledge base."""
    typer.echo(f"alty kb: not yet implemented (topic={topic!r})")


@app.command(name="doc-health")
def doc_health() -> None:
    """Check documentation freshness and health."""
    typer.echo("alty doc-health: not yet implemented")


@app.command(name="doc-review")
def doc_review() -> None:
    """Mark documentation as reviewed."""
    typer.echo("alty doc-review: not yet implemented")


@app.command(name="ticket-health")
def ticket_health() -> None:
    """Show ripple review report for tickets needing attention."""
    typer.echo("alty ticket-health: not yet implemented")


@app.command(name="version")
def version_cmd() -> None:
    """Show the alty version."""
    pkg_version = version("alty")
    typer.echo(f"alty {pkg_version}")
