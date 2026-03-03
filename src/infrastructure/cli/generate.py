"""Generate subcommand group: alty generate <subcommand>.

Subcommands:
    artifacts — Generate PRD, DDD.md, ARCHITECTURE.md
    fitness   — Generate import-linter + pytestarch tests
    tickets   — Generate beads epics + tasks from DDD
    configs   — Generate tool-specific configs (.claude/, .cursor/, etc.)

Reference: ARCHITECTURE.md §6.1
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import TYPE_CHECKING

import typer

if TYPE_CHECKING:
    from src.domain.models.discovery_session import DiscoverySession
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.stack_profile import StackProfile
    from src.infrastructure.composition import AppContext

app = typer.Typer(
    name="generate",
    help="Generate project artifacts, fitness tests, tickets, and configs.",
    no_args_is_help=True,
)


# ── Shared helpers ────────────────────────────────────────────


def _load_session() -> DiscoverySession:
    """Load and validate a completed discovery session from .alty/session.json.

    Returns:
        A DiscoverySession in COMPLETED state.

    Raises:
        typer.Exit: If file missing, invalid, or session not completed.
    """
    from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus

    snapshot_path = Path.cwd() / ".alty" / "session.json"
    if not snapshot_path.exists():
        typer.echo("No discovery session found. Run 'alty guide' first.", err=True)
        raise typer.Exit(code=1)

    try:
        data = json.loads(snapshot_path.read_text())
        session = DiscoverySession.from_snapshot(data)
    except (json.JSONDecodeError, ValueError, KeyError) as exc:
        typer.echo(f"Invalid session file: {exc}", err=True)
        raise typer.Exit(code=1) from None

    if session.status != DiscoveryStatus.COMPLETED:
        typer.echo(
            f"Session not completed (status: {session.status.value}). "
            "Complete the guided discovery first.",
            err=True,
        )
        raise typer.Exit(code=1)

    return session


def _build_domain_model(
    session: DiscoverySession,
) -> tuple[DomainModel, AppContext]:
    """Build a DomainModel from a completed session via ArtifactGenerationHandler."""
    from src.application.commands.artifact_generation_handler import (
        ArtifactGenerationHandler,
    )
    from src.infrastructure.cli.main import _reconstruct_event
    from src.infrastructure.composition import create_app

    ctx = create_app()
    event = _reconstruct_event(session)
    handler = ArtifactGenerationHandler(
        renderer=ctx.artifact_renderer,
        writer=ctx.file_writer,
    )
    preview = handler.build_preview(event)
    return preview.model, ctx


def _resolve_profile(session: DiscoverySession) -> StackProfile:
    """Resolve a StackProfile from the session's tech stack."""
    from src.domain.services.stack_resolver import resolve_profile

    return resolve_profile(session.tech_stack)


# ── Subcommands ───────────────────────────────────────────────


@app.command()
def artifacts() -> None:
    """Generate PRD, DDD.md, and ARCHITECTURE.md from domain model."""
    from src.application.commands.artifact_generation_handler import (
        ArtifactGenerationHandler,
    )
    from src.infrastructure.cli.main import _reconstruct_event
    from src.infrastructure.composition import create_app

    session = _load_session()
    event = _reconstruct_event(session)

    ctx = create_app()
    handler = ArtifactGenerationHandler(
        renderer=ctx.artifact_renderer,
        writer=ctx.file_writer,
    )

    preview = handler.build_preview(event)

    typer.echo("Artifact Generation Preview")
    typer.echo(f"  PRD: {len(preview.prd_content)} chars")
    typer.echo(f"  DDD: {len(preview.ddd_content)} chars")
    typer.echo(f"  Architecture: {len(preview.architecture_content)} chars")

    if not typer.confirm("Write artifacts?"):
        typer.echo("Cancelled.")
        raise typer.Exit(code=0)

    output_dir = Path("docs")
    handler.write_artifacts(preview, output_dir)
    typer.echo("Artifacts written successfully.")


@app.command()
def fitness() -> None:
    """Generate architecture fitness tests (import-linter + pytestarch)."""
    from src.application.commands.fitness_generation_handler import (
        FitnessGenerationHandler,
    )

    session = _load_session()
    model, ctx = _build_domain_model(session)
    profile = _resolve_profile(session)

    # Generate fitness tests.
    handler = FitnessGenerationHandler(writer=ctx.file_writer)
    root_package = profile.to_root_package(Path.cwd().name)
    preview = handler.build_preview(model, root_package, profile=profile)

    if preview is None:
        typer.echo("Fitness tests not available for your stack (requires Python with uv).")
        raise typer.Exit(code=0)

    typer.echo(preview.summary)

    if not typer.confirm("Write fitness tests?"):
        typer.echo("Cancelled.")
        raise typer.Exit(code=0)

    output_dir = Path(".")
    handler.approve_and_write(preview, output_dir)
    typer.echo("Fitness tests written successfully.")


@app.command()
def tickets() -> None:
    """Generate beads epics and tasks from DDD stories."""
    from src.application.commands.ticket_generation_handler import (
        TicketGenerationHandler,
    )

    session = _load_session()
    model, ctx = _build_domain_model(session)
    profile = _resolve_profile(session)

    # Generate tickets.
    handler = TicketGenerationHandler(writer=ctx.file_writer)
    preview = handler.build_preview(model, profile)

    typer.echo(preview.summary)

    if not typer.confirm("Write tickets?"):
        typer.echo("Cancelled.")
        raise typer.Exit(code=0)

    output_dir = Path(".")
    handler.approve_and_write(preview, output_dir)
    typer.echo("Tickets written successfully.")


@app.command()
def configs() -> None:
    """Generate tool-specific configs (.claude/, .cursor/, etc.)."""
    from src.application.commands.config_generation_handler import (
        ConfigGenerationHandler,
    )
    from src.domain.models.tool_config import SupportedTool

    session = _load_session()
    model, ctx = _build_domain_model(session)
    profile = _resolve_profile(session)

    # Generate configs for all supported tools.
    handler = ConfigGenerationHandler(writer=ctx.file_writer)
    tools = tuple(SupportedTool)
    preview = handler.build_preview(model, tools, profile)

    typer.echo(preview.summary)

    if not typer.confirm("Write configs?"):
        typer.echo("Cancelled.")
        raise typer.Exit(code=0)

    output_dir = Path(".")
    handler.approve_and_write(preview, output_dir)
    typer.echo("Configs written successfully.")
