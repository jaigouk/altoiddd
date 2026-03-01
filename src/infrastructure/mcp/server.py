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

from src.infrastructure.composition import AppContext, create_app

# ── Input validation ─────────────────────────────────────────────────

_SAFE_NAME_RE = re.compile(r"^[a-zA-Z0-9_\-]{1,64}$")
_TICKET_ID_RE = re.compile(r"^[a-zA-Z0-9][a-zA-Z0-9._\-]{0,63}$")
_REVIEWER_RE = re.compile(r"^[a-zA-Z0-9_\-.@]{1,100}$")


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
    """Create the application context for the MCP server lifetime."""
    ctx = create_app()
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
    """Run a bd command asynchronously and return its stdout."""
    bd = _bd_path()
    proc = await asyncio.create_subprocess_exec(
        bd,
        *args,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE,
    )
    stdout, _stderr = await asyncio.wait_for(proc.communicate(), timeout=10)
    return stdout.decode() if stdout else ""


# ── Tools (11) ───────────────────────────────────────────────────────


@mcp.tool()
async def init_project(project_dir: str, ctx: McpContext) -> str:
    """Bootstrap a new project from a README idea."""
    app = _get_app(ctx)
    return app.bootstrap.preview(_safe_project_path(project_dir, "project_dir"))


@mcp.tool()
async def guide_ddd(readme_content: str, ctx: McpContext) -> str:
    """Start a guided DDD discovery session from README content."""
    app = _get_app(ctx)
    return app.discovery.start_session(readme_content)


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
    """Generate architecture fitness functions from a domain model."""
    app = _get_app(ctx)
    safe_dir = _safe_project_path(output_dir, "output_dir")
    app.fitness_generation.generate(
        model=None,  # type: ignore[arg-type]
        root_package=root_package,
        output_dir=safe_dir,
    )
    return f"Fitness functions generated in {safe_dir}"


@mcp.tool()
async def generate_tickets(output_dir: str, ctx: McpContext) -> str:
    """Generate dependency-ordered beads tickets from a domain model."""
    app = _get_app(ctx)
    safe_dir = _safe_project_path(output_dir, "output_dir")
    app.ticket_generation.generate(
        model=None,  # type: ignore[arg-type]
        output_dir=safe_dir,
    )
    return f"Tickets generated in {safe_dir}"


@mcp.tool()
async def generate_configs(output_dir: str, ctx: McpContext) -> str:
    """Generate tool-native configurations for detected AI coding tools."""
    app = _get_app(ctx)
    safe_dir = _safe_project_path(output_dir, "output_dir")
    app.config_generation.generate(
        model=None,  # type: ignore[arg-type]
        tools=(),
        output_dir=safe_dir,
    )
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
async def doc_review(doc_path: str, reviewer: str, ctx: McpContext) -> str:
    """Mark a document as reviewed."""
    if not _REVIEWER_RE.fullmatch(reviewer):
        return "Invalid reviewer identifier."
    app = _get_app(ctx)
    return app.doc_review.mark_reviewed(
        _safe_project_path(doc_path, "doc_path"), reviewer
    )


@mcp.tool()
async def ticket_health(project_dir: str, ctx: McpContext) -> str:
    """Show ripple review report for tickets needing attention."""
    app = _get_app(ctx)
    return str(app.ticket_health.report(_safe_project_path(project_dir, "project_dir")))


# ── Resources (10) ───────────────────────────────────────────────────


@mcp.resource("alty://knowledge/ddd/{topic}")
async def knowledge_ddd(topic: str) -> str:
    """Read a DDD knowledge base entry."""
    try:
        safe_topic = _safe_component(topic, "topic")
    except ValueError:
        return "Invalid topic name."
    kb_path = Path(".alty/knowledge/ddd") / f"{safe_topic}.md"
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
    kb_path = Path(".alty/knowledge/tools") / safe_tool
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
    kb_path = Path(".alty/knowledge/tools") / safe_tool / f"{safe_subtopic}.toml"
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
    kb_path = Path(".alty/knowledge/conventions") / f"{safe_topic}.md"
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
    kb_path = Path(".alty/knowledge/cross-tool") / f"{safe_topic}.toml"
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
