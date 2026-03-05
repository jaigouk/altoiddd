"""MCP server adapter for alty.

Exposes alty capabilities as MCP tools and resources via FastMCP.
This is the infrastructure adapter for the MCP transport -- it imports
only from application.ports and infrastructure.composition.

NOTE: This module intentionally does NOT use ``from __future__ import annotations``
because FastMCP inspects function signatures with ``eval_str=True`` at decoration
time, and deferred string annotations break that introspection.

Entry point: alty-mcp = "src.infrastructure.mcp.server:main"
"""

import asyncio
import re
import shutil
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Any

from mcp.server.fastmcp import Context, FastMCP

from src.domain.models.discovery_session import DiscoveryStatus
from src.domain.models.errors import InvariantViolationError, SessionNotFoundError
from src.infrastructure.composition import AppContext, create_app
from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter
from src.infrastructure.session.session_store import SessionStore

# ── Input validation ─────────────────────────────────────────────────

_SAFE_NAME_RE = re.compile(r"^[a-zA-Z0-9_\-]{1,64}$")
_TICKET_ID_RE = re.compile(r"^[a-zA-Z0-9][a-zA-Z0-9._\-]{0,63}$")


def _safe_component(value: str, label: str) -> str:
    """Validate a URI component is a safe filename (no path traversal)."""
    if not _SAFE_NAME_RE.fullmatch(value):
        msg = f"Invalid {label}: {value!r}"
        raise ValueError(msg)
    return value


def _safe_project_path(raw: str, label: str = "path") -> Path:
    """Resolve a path and assert it is non-empty and absolute.

    Prevents path traversal by resolving to an absolute path and
    rejecting empty inputs.
    """
    if not raw or not raw.strip():
        msg = f"{label} must not be empty"
        raise ValueError(msg)
    return Path(raw).resolve()


# Type alias for the MCP request context parametrized with our AppContext.
# FastMCP introspects tool function signatures at decoration time with
# eval_str=True, so we must keep real (not stringified) annotations.
McpContext = Context[Any, AppContext, Any]

# ── Lifespan ─────────────────────────────────────────────────────────


@asynccontextmanager
async def app_lifespan(server: FastMCP) -> AsyncIterator[AppContext]:
    """Create the application context for the MCP server lifetime.

    Wires the SessionStore and DiscoveryAdapter into AppContext so that
    discovery tools share session state across stateless MCP calls.
    """
    store = SessionStore()
    adapter = DiscoveryAdapter(store=store)
    ctx = create_app()
    # Replace the stub discovery with our real adapter
    object.__setattr__(ctx, "discovery", adapter)
    yield ctx


# ── Server instance ──────────────────────────────────────────────────

mcp = FastMCP("alty", lifespan=app_lifespan)


def _get_app(ctx: McpContext) -> AppContext:
    """Extract AppContext from the MCP request context."""
    app: AppContext = ctx.request_context.lifespan_context
    return app


def _bd_path() -> str:
    """Resolve the full path to the bd executable."""
    return shutil.which("bd") or "bd"


async def _run_bd(*args: str) -> str:
    """Run a bd command asynchronously and return its stdout.

    On non-zero exit, includes stderr in the returned message.
    """
    bd = _bd_path()
    proc = await asyncio.create_subprocess_exec(
        bd,
        *args,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE,
    )
    stdout, stderr = await asyncio.wait_for(proc.communicate(), timeout=10)
    if proc.returncode != 0:
        err_msg = stderr.decode().strip() if stderr else "unknown error"
        return f"bd command failed: {err_msg}"
    return stdout.decode() if stdout else ""


# ── Tools: Bootstrap & Generation (11) ──────────────────────────────


@mcp.tool()
async def init_project(project_dir: str, ctx: McpContext) -> str:
    """Bootstrap a new project from a README idea."""
    app = _get_app(ctx)
    return app.bootstrap.preview(_safe_project_path(project_dir, "project_dir"))


@mcp.tool()
async def generate_artifacts(project_dir: str, ctx: McpContext) -> str:
    """Generate DDD artifacts from a completed discovery session."""
    _get_app(ctx)  # validate context is available
    _ = project_dir  # reserved for future port
    raise NotImplementedError("generate_artifacts: port not yet defined")


@mcp.tool()
async def generate_fitness(
    root_package: str,
    output_dir: str,
    ctx: McpContext,
) -> str:
    """Generate architecture fitness functions from a domain model.

    Requires a completed discovery session with generated domain model.
    """
    app = _get_app(ctx)
    safe_dir = _safe_project_path(output_dir, "output_dir")
    try:
        app.fitness_generation.generate(
            model=None,  # type: ignore[arg-type]
            root_package=root_package,
            output_dir=safe_dir,
        )
    except NotImplementedError:
        return (
            "Error: fitness generation requires a completed domain model. "
            "Run guide_start first."
        )
    return f"Fitness functions generated in {safe_dir}"


@mcp.tool()
async def generate_tickets(output_dir: str, ctx: McpContext) -> str:
    """Generate dependency-ordered beads tickets from a domain model.

    Requires a completed discovery session with generated domain model.
    """
    app = _get_app(ctx)
    safe_dir = _safe_project_path(output_dir, "output_dir")
    try:
        app.ticket_generation.generate(
            model=None,  # type: ignore[arg-type]
            output_dir=safe_dir,
        )
    except NotImplementedError:
        return "Error: ticket generation requires a completed domain model. Run guide_start first."
    return f"Tickets generated in {safe_dir}"


@mcp.tool()
async def generate_configs(output_dir: str, ctx: McpContext) -> str:
    """Generate tool-native configurations for detected AI coding tools.

    Requires a completed discovery session with generated domain model.
    """
    app = _get_app(ctx)
    safe_dir = _safe_project_path(output_dir, "output_dir")
    try:
        app.config_generation.generate(
            model=None,  # type: ignore[arg-type]
            tools=(),
            output_dir=safe_dir,
        )
    except NotImplementedError:
        return "Error: config generation requires a completed domain model. Run guide_start first."
    return f"Configs generated in {safe_dir}"


@mcp.tool()
async def detect_tools(project_dir: str, ctx: McpContext) -> str:
    """Detect installed AI coding tools in a project directory."""
    app = _get_app(ctx)
    detected = app.tool_detection.detect(_safe_project_path(project_dir, "project_dir"))
    if not detected:
        return "No AI coding tools detected."
    return "Detected tools: " + ", ".join(detected)


@mcp.tool()
async def check_quality(ctx: McpContext) -> str:
    """Run quality gates (lint, types, tests, fitness)."""
    app = _get_app(ctx)
    return str(app.quality_gate.check())


@mcp.tool()
async def doc_health(project_dir: str, ctx: McpContext) -> str:
    """Check documentation freshness and health."""
    app = _get_app(ctx)
    return str(app.doc_health.check(_safe_project_path(project_dir, "project_dir")))


@mcp.tool()
async def doc_review(doc_path: str, project_dir: str, ctx: McpContext) -> str:
    """Mark a document as reviewed."""
    app = _get_app(ctx)
    result = app.doc_review.mark_reviewed(
        doc_path=doc_path,
        project_dir=_safe_project_path(project_dir, "project_dir"),
    )
    return f"Reviewed: {result.path} (last_reviewed: {result.new_date})"


@mcp.tool()
async def ticket_health(project_dir: str, ctx: McpContext) -> str:
    """Show ripple review report for tickets needing attention."""
    app = _get_app(ctx)
    return str(app.ticket_health.report(_safe_project_path(project_dir, "project_dir")))


@mcp.tool()
async def spike_follow_up_audit(spike_id: str, project_dir: str, ctx: McpContext) -> str:
    """Audit whether a spike's follow-up tickets were actually created.

    Scans the spike's research report for follow-up intent sections and
    compares them against existing beads tickets using fuzzy matching.

    Args:
        spike_id: The spike ticket identifier.
        project_dir: Project directory containing docs/research/ and .beads/.
    """
    app = _get_app(ctx)
    result = app.spike_follow_up.audit(
        spike_id, _safe_project_path(project_dir, "project_dir")
    )
    if result.defined_count == 0:
        return f"Spike '{spike_id}': no follow-up intents found in research reports."
    if not result.has_orphans:
        return (
            f"Spike '{spike_id}': all {result.defined_count} follow-up intents "
            f"matched to existing tickets."
        )
    orphan_titles = "\n".join(f"  - {o.title}" for o in result.orphaned_intents)
    return (
        f"Spike '{spike_id}': {result.orphaned_count} of {result.defined_count} "
        f"follow-up intents have no matching ticket:\n{orphan_titles}\n\n"
        f"Report: {result.report_path}"
    )


# ── Tools: Guided Discovery (7) ─────────────────────────────────────


def _format_next_question(session: Any) -> str:
    """Build a response describing what the MCP client should do next."""
    from src.domain.models.question import Question

    status = session.status
    answered_ids = {a.question_id for a in session.answers}

    if status == DiscoveryStatus.PLAYBACK_PENDING:
        return (
            f"Session {session.session_id}: PLAYBACK_PENDING.\n"
            f"Please confirm the playback summary before continuing.\n"
            f"Use guide_confirm_playback(session_id, confirmed=True) to proceed."
        )

    if status == DiscoveryStatus.COMPLETED:
        return f"Session {session.session_id}: COMPLETED with {len(session.answers)} answers."

    # Find the next unanswered question
    for q in Question.CATALOG:
        if q.id not in answered_ids:
            register = session.register
            is_technical = register and register.value == "technical"
            text = q.technical_text if is_technical else q.non_technical_text
            return (
                f"Session {session.session_id}: next question {q.id} ({q.phase.value} phase).\n"
                f"{text}"
            )

    return f"Session {session.session_id}: all questions answered. Use guide_complete to finish."


@mcp.tool()
async def guide_start(readme_content: str, ctx: McpContext) -> str:
    """Start a guided DDD discovery session from README content.

    Returns a session_id for use in subsequent guide_* tool calls.
    The session persists server-side for 30 minutes (TTL).
    """
    app = _get_app(ctx)
    session = app.discovery.start_session(readme_content)
    return (
        f"Discovery session started.\n"
        f"session_id: {session.session_id}\n"
        f"Next step: detect persona with guide_detect_persona(session_id, choice)\n"
        f"Choices: 1=Developer, 2=Product Owner, 3=Domain Expert, 4=Mixed"
    )


@mcp.tool()
async def guide_detect_persona(session_id: str, choice: str, ctx: McpContext) -> str:
    """Detect the user persona for a discovery session.

    Args:
        session_id: The session ID from guide_start.
        choice: '1'=Developer, '2'=Product Owner, '3'=Domain Expert, '4'=Mixed.
    """
    app = _get_app(ctx)
    try:
        session = app.discovery.detect_persona(session_id, choice)
    except SessionNotFoundError:
        return f"Error: session '{session_id}' not found or expired."
    except (ValueError, InvariantViolationError) as e:
        return f"Error: {e}"
    persona_val = session.persona.value if session.persona else "unknown"
    register_val = session.register.value if session.register else "unknown"
    return (
        f"Persona detected: {persona_val}, register: {register_val}.\n"
        f"{_format_next_question(session)}"
    )


@mcp.tool()
async def guide_answer(
    session_id: str, question_id: str, answer: str, ctx: McpContext
) -> str:
    """Answer a discovery question.

    Args:
        session_id: The session ID from guide_start.
        question_id: The question to answer (Q1-Q10).
        answer: The user's free-text answer.
    """
    app = _get_app(ctx)
    try:
        session = app.discovery.answer_question(session_id, question_id, answer)
    except SessionNotFoundError:
        return f"Error: session '{session_id}' not found or expired."
    except (ValueError, InvariantViolationError) as e:
        return f"Error: {e}"
    return (
        f"Recorded answer for {question_id}.\n"
        f"{_format_next_question(session)}"
    )


@mcp.tool()
async def guide_skip_question(
    session_id: str, question_id: str, reason: str, ctx: McpContext
) -> str:
    """Skip a discovery question with an explicit reason.

    Args:
        session_id: The session ID from guide_start.
        question_id: The question to skip (Q1-Q10).
        reason: Why it was skipped (must be non-empty).
    """
    app = _get_app(ctx)
    try:
        session = app.discovery.skip_question(session_id, question_id, reason)
    except SessionNotFoundError:
        return f"Error: session '{session_id}' not found or expired."
    except (ValueError, InvariantViolationError) as e:
        return f"Error: {e}"
    return (
        f"Skipped {question_id} (reason: {reason}).\n"
        f"{_format_next_question(session)}"
    )


@mcp.tool()
async def guide_confirm_playback(
    session_id: str, confirmed: bool, ctx: McpContext
) -> str:
    """Confirm or reject the playback summary.

    Args:
        session_id: The session ID from guide_start.
        confirmed: True to accept the summary, False to revise.
    """
    app = _get_app(ctx)
    try:
        session = app.discovery.confirm_playback(session_id, confirmed)
    except SessionNotFoundError:
        return f"Error: session '{session_id}' not found or expired."
    except (ValueError, InvariantViolationError) as e:
        return f"Error: {e}"
    action = "confirmed" if confirmed else "rejected"
    return (
        f"Playback {action}.\n"
        f"{_format_next_question(session)}"
    )


@mcp.tool()
async def guide_complete(session_id: str, ctx: McpContext) -> str:
    """Complete a discovery session and produce domain events.

    Validates that minimum MVP questions (Q1, Q3, Q4, Q9, Q10) have been
    answered before marking complete.
    """
    app = _get_app(ctx)
    try:
        session = app.discovery.complete(session_id)
    except SessionNotFoundError:
        return f"Error: session '{session_id}' not found or expired."
    except (ValueError, InvariantViolationError) as e:
        return f"Error: {e}"
    return (
        f"Discovery session completed.\n"
        f"session_id: {session.session_id}\n"
        f"Answers: {len(session.answers)}\n"
        f"Events: {len(session.events)}\n"
        f"Next step: use generate_artifacts to produce DDD documents."
    )


@mcp.tool()
async def guide_status(session_id: str, ctx: McpContext) -> str:
    """Get the current status of a discovery session.

    Returns phase, persona, answered questions, and next steps.
    """
    app = _get_app(ctx)
    try:
        session = app.discovery.get_session(session_id)
    except SessionNotFoundError:
        return f"Error: session '{session_id}' not found or expired."

    answered_ids = [a.question_id for a in session.answers]
    persona_str = session.persona.value if session.persona else "not detected"
    register_str = session.register.value if session.register else "not set"

    return (
        f"Session: {session_id}\n"
        f"Status: {session.status.value}\n"
        f"Persona: {persona_str}\n"
        f"Register: {register_str}\n"
        f"Phase: {session.current_phase.value}\n"
        f"Answered: {len(answered_ids)} ({', '.join(answered_ids) or 'none'})\n"
        f"Playbacks: {len(session.playback_confirmations)}"
    )


# ── Resources (10) ───────────────────────────────────────────────────


def _kb_root() -> Path:
    """Return the knowledge base root resolved from the current working directory."""
    return Path.cwd() / ".alty" / "knowledge"


@mcp.resource("alty://knowledge/ddd/{topic}")
async def knowledge_ddd(topic: str) -> str:
    """Read a DDD knowledge base entry."""
    try:
        safe_topic = _safe_component(topic, "topic")
    except ValueError:
        return "Invalid topic name."
    kb_path = _kb_root() / "ddd" / f"{safe_topic}.md"
    if kb_path.exists():
        return kb_path.read_text()
    return f"DDD topic '{safe_topic}' not found."


@mcp.resource("alty://knowledge/tools/{tool}")
async def knowledge_tool(tool: str) -> str:
    """Read a tool knowledge base directory listing."""
    try:
        safe_tool = _safe_component(tool, "tool")
    except ValueError:
        return "Invalid tool name."
    kb_path = _kb_root() / "tools" / safe_tool
    if kb_path.exists() and kb_path.is_dir():
        entries = sorted(p.name for p in kb_path.iterdir() if p.is_file())
        return "\n".join(entries) or f"No entries for tool '{safe_tool}'."
    return f"Tool '{safe_tool}' not found in knowledge base."


@mcp.resource("alty://knowledge/tools/{tool}/{subtopic}")
async def knowledge_tool_subtopic(tool: str, subtopic: str) -> str:
    """Read a specific tool knowledge base entry."""
    try:
        safe_tool = _safe_component(tool, "tool")
        safe_subtopic = _safe_component(subtopic, "subtopic")
    except ValueError:
        return "Invalid tool or subtopic name."
    kb_path = _kb_root() / "tools" / safe_tool / f"{safe_subtopic}.toml"
    if kb_path.exists():
        return kb_path.read_text()
    return f"Tool '{safe_tool}' subtopic '{safe_subtopic}' not found."


@mcp.resource("alty://knowledge/conventions/{topic}")
async def knowledge_conventions(topic: str) -> str:
    """Read a conventions knowledge base entry."""
    try:
        safe_topic = _safe_component(topic, "topic")
    except ValueError:
        return "Invalid topic name."
    kb_path = _kb_root() / "conventions" / f"{safe_topic}.md"
    if kb_path.exists():
        return kb_path.read_text()
    return f"Convention topic '{safe_topic}' not found."


@mcp.resource("alty://knowledge/cross-tool/{topic}")
async def knowledge_cross_tool(topic: str) -> str:
    """Read a cross-tool knowledge base entry."""
    try:
        safe_topic = _safe_component(topic, "topic")
    except ValueError:
        return "Invalid topic name."
    kb_path = _kb_root() / "cross-tool" / f"{safe_topic}.toml"
    if kb_path.exists():
        return kb_path.read_text()
    return f"Cross-tool topic '{safe_topic}' not found."


@mcp.resource("alty://project/{dir}/domain-model")
async def project_domain_model(dir: str) -> str:
    """Read a project's DDD domain model document."""
    project_path = _safe_project_path(dir, "project dir")
    doc_path = project_path / "docs" / "DDD.md"
    if doc_path.exists():
        return doc_path.read_text()
    return f"Domain model not found at {doc_path}"


@mcp.resource("alty://project/{dir}/architecture")
async def project_architecture(dir: str) -> str:
    """Read a project's architecture document."""
    project_path = _safe_project_path(dir, "project dir")
    doc_path = project_path / "docs" / "ARCHITECTURE.md"
    if doc_path.exists():
        return doc_path.read_text()
    return f"Architecture document not found at {doc_path}"


@mcp.resource("alty://project/{dir}/prd")
async def project_prd(dir: str) -> str:
    """Read a project's PRD document."""
    project_path = _safe_project_path(dir, "project dir")
    doc_path = project_path / "docs" / "PRD.md"
    if doc_path.exists():
        return doc_path.read_text()
    return f"PRD document not found at {doc_path}"


@mcp.resource("alty://tickets/ready")
async def tickets_ready() -> str:
    """List tickets that are ready to work on."""
    try:
        output = await _run_bd("ready")
        return output or "No ready tickets."
    except (FileNotFoundError, TimeoutError):
        return "Unable to fetch ready tickets (bd command not available)."


@mcp.resource("alty://tickets/{ticket_id}")
async def tickets_by_id(ticket_id: str) -> str:
    """Show details for a specific ticket."""
    if not _TICKET_ID_RE.fullmatch(ticket_id):
        return "Invalid ticket ID format."
    try:
        output = await _run_bd("show", ticket_id)
        return output or "Ticket not found."
    except (FileNotFoundError, TimeoutError):
        return "Unable to fetch ticket (bd command not available)."


# ── Entry point ──────────────────────────────────────────────────────


def main() -> None:
    """Entry point for alty-mcp."""
    mcp.run()
