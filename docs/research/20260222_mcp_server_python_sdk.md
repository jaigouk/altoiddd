---
last_reviewed: 2026-02-22
type: spike
status: complete
topic: MCP Server Python SDK for alto
---

# MCP Server Implementation in Python: Research Findings

## 1. Decision Context

alto exposes two interfaces (PRD Section 6, Constraints):
- **CLI (`vs`)** -- primary interface for humans
- **MCP server** -- exposes the same capabilities to AI coding tools (Claude Code, Cursor, etc.)

Both interfaces must share the same application core. The MCP server is Phase 5 in the
PRD timeline, but understanding its design constraints now informs the ports/adapters
architecture being built in earlier phases.

### Research Questions

1. What is the official Python SDK for MCP servers?
2. How do you define MCP tools with input schemas?
3. How do you define MCP resources?
4. How does server transport work (stdio vs SSE vs HTTP)?
5. Can an MCP server and CLI share the same application layer?
6. What is the minimal boilerplate to expose a tool?

---

## 2. The MCP Python SDK Landscape

### Two Packages, One Lineage

| Package | Version | License | PyPI | Relationship |
|---------|---------|---------|------|-------------|
| `mcp` | 1.26.0 (Jan 2026) | MIT | [pypi.org/project/mcp](https://pypi.org/project/mcp/) | Official Anthropic SDK. Includes FastMCP 1.x as `mcp.server.fastmcp` |
| `fastmcp` | 3.0.2 (Feb 2026) | Apache-2.0 | [pypi.org/project/fastmcp](https://pypi.org/project/fastmcp/) | Standalone project by jlowin/Prefect. FastMCP 1.0 was merged into official SDK in 2024; standalone continued evolving to v2/v3 with additional features |

Source: [github.com/modelcontextprotocol/python-sdk](https://github.com/modelcontextprotocol/python-sdk), [github.com/PrefectHQ/fastmcp](https://github.com/jlowin/fastmcp)

**Key distinction:**
- `mcp` (official) -- the reference implementation. Includes `FastMCP` class for high-level
  server building and `mcp.server.lowlevel.Server` for full control. Stable, well-documented.
- `fastmcp` (standalone) -- more features (OpenAPI import, UI builder, Provider abstraction,
  advanced composition). But adds another dependency layer and diverges from the official API.

### Recommendation: Use `mcp` (official SDK)

Rationale:
- MIT license (compatible with any project license)
- Python 3.10+ (alto targets 3.12+ -- fully compatible)
- Maintained by Anthropic / MCP project -- guaranteed protocol compliance
- FastMCP high-level API is included -- no need for the standalone package
- Fewer dependencies = simpler supply chain
- All AI tools (Claude Code, Cursor) implement the official MCP protocol

---

## 3. Answer: How to Define MCP Tools

### High-Level API (FastMCP) -- Recommended

Tools are defined with the `@mcp.tool()` decorator. Input schemas are **auto-generated
from Python type annotations**. Docstrings become the tool description.

```python
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("alto")

@mcp.tool()
def init_project(project_dir: str, existing: bool = False) -> dict[str, str]:
    """Initialize a alto project.

    Args:
        project_dir: Path to the project directory
        existing: If True, adopt an existing project (creates branch)
    """
    # Delegates to application layer
    result = init_use_case.execute(project_dir, existing=existing)
    return {"status": "ok", "message": result.summary}
```

The SDK automatically generates this JSON Schema for the tool:
```json
{
  "type": "object",
  "properties": {
    "project_dir": {"type": "string", "description": "Path to the project directory"},
    "existing": {"type": "boolean", "default": false, "description": "If True, adopt..."}
  },
  "required": ["project_dir"]
}
```

Source: [Context7 - MCP Python SDK docs](https://context7.com/modelcontextprotocol/python-sdk/llms.txt)

### Tool Features

| Feature | How | Source |
|---------|-----|--------|
| **Async tools** | `async def` -- fully supported | SDK README |
| **Context injection** | Add `ctx: Context` parameter for logging, progress | SDK README |
| **Structured output** | Return Pydantic `BaseModel`, `TypedDict`, or `dict` | SDK README |
| **Progress reporting** | `await ctx.report_progress(progress=3, total=10)` | SDK README |
| **Logging** | `await ctx.info("message")`, `ctx.debug()`, `ctx.warning()` | SDK README |
| **Input validation** | Pydantic `Field(description=...)` on model fields | SDK README |

### Context Object

The `Context` object provides access to MCP capabilities inside tool handlers:

```python
from mcp.server.fastmcp import Context

@mcp.tool()
async def guide_ddd(readme_content: str, ctx: Context) -> str:
    """Start guided DDD discovery from a README."""
    await ctx.info("Starting DDD discovery flow...")
    await ctx.report_progress(progress=1, total=10, message="Analyzing README")
    # ... business logic ...
    return result
```

### Low-Level API (full schema control)

For cases where you need explicit JSON Schema control (not auto-generated):

```python
from mcp.server.lowlevel import Server
import mcp.types as types

server = Server("alto")

@server.list_tools()
async def list_tools() -> list[types.Tool]:
    return [
        types.Tool(
            name="init_project",
            description="Initialize a alto project",
            inputSchema={
                "type": "object",
                "properties": {
                    "project_dir": {"type": "string"},
                    "existing": {"type": "boolean", "default": False}
                },
                "required": ["project_dir"]
            }
        )
    ]

@server.call_tool()
async def call_tool(name: str, arguments: dict) -> dict:
    if name == "init_project":
        return init_use_case.execute(**arguments)
    raise ValueError(f"Unknown tool: {name}")
```

Source: [Context7 - MCP Python SDK low-level example](https://context7.com/modelcontextprotocol/python-sdk/llms.txt)

**Recommendation for alto:** Use the high-level FastMCP API. Auto-generated schemas
from type annotations keep tool definitions DRY and in sync with the application layer.
Reserve the low-level API only if a specific tool needs a custom schema.

---

## 4. Answer: How to Define MCP Resources

Resources expose read-only data to LLMs. They are analogous to GET endpoints in REST.

### Static Resources

```python
@mcp.resource("alto://knowledge/ddd/{topic}")
def get_ddd_knowledge(topic: str) -> str:
    """Get DDD knowledge base content."""
    return knowledge_service.get_topic("ddd", topic)
```

### Resource Templates (URI parameters)

URI templates use `{param}` syntax. Parameters are extracted and passed to the function:

```python
@mcp.resource("alto://project/{project_dir}/health")
def get_project_health(project_dir: str) -> str:
    """Get doc-health and ticket-health for a project."""
    return health_service.check(project_dir)

@mcp.resource("alto://project/{project_dir}/domain-model")
def get_domain_model(project_dir: str) -> str:
    """Get the current DDD model for a project."""
    return read_file(f"{project_dir}/docs/DDD.md")
```

### Async Resources with Context

```python
@mcp.resource("alto://tickets/ready")
async def get_ready_tickets(ctx: Context) -> str:
    """List tickets ready for work."""
    await ctx.info("Fetching ready tickets")
    return ticket_query_service.list_ready()
```

Source: [Context7 - MCP Python SDK resource examples](https://context7.com/modelcontextprotocol/python-sdk/llms.txt)

### Resource vs Tool Decision

| Use Case | MCP Primitive | Rationale |
|----------|--------------|-----------|
| `alto init`, `alto guide`, `alto generate` | **Tool** | Write operations, side effects |
| `alto doc-health`, `alto ticket-health` | **Tool** (with structured output) | Analysis with potential side effects (flags) |
| Knowledge base lookup | **Resource** | Read-only data retrieval |
| Current domain model | **Resource** | Read-only project state |
| Ticket list / details | **Resource** | Read-only query |

---

## 5. Answer: Transport Options

### Three Transport Mechanisms

| Transport | Protocol | Use Case | Claude Code | Cursor |
|-----------|----------|----------|-------------|--------|
| **stdio** | stdin/stdout | Local process, spawned by AI tool | Yes (primary) | Yes |
| **Streamable HTTP** | HTTP + streaming | Remote/networked, production | Possible | Possible |
| **SSE** | HTTP + Server-Sent Events | Legacy, being superseded | Limited | Limited |

Source: [MCP SDK README](https://github.com/modelcontextprotocol/python-sdk), [MCP specification](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports)

### stdio Transport (recommended for alto)

This is the standard for local MCP servers. The AI tool spawns the server as a subprocess
and communicates via stdin/stdout. This is how Claude Code and Cursor invoke MCP servers.

```python
# Default -- runs over stdio
if __name__ == "__main__":
    mcp.run()  # transport="stdio" is the default
```

Configuration in `.claude/settings.json`:
```json
{
  "mcpServers": {
    "alto": {
      "command": "uv",
      "args": ["run", "alto-mcp"],
      "cwd": "/path/to/project"
    }
  }
}
```

### Streamable HTTP Transport

For remote access or when multiple clients need to connect:

```python
mcp.run(transport="streamable-http", host="0.0.0.0", port=8000)
```

### ASGI Mounting (for embedding in existing web app)

```python
from starlette.applications import Starlette
from starlette.routing import Mount

app = Starlette(routes=[
    Mount("/mcp", app=mcp.streamable_http_app()),
])
```

**Recommendation for alto:** stdio as primary transport (matches how AI tools
invoke local servers). Streamable HTTP as optional for future remote/team use cases.

---

## 6. Answer: CLI + MCP Shared Application Layer

### Architecture: Ports and Adapters

The key insight is that **both the CLI and MCP server are infrastructure adapters** that
delegate to the same application layer (use cases / command handlers). This is exactly
the ports/adapters (hexagonal) architecture already specified in CLAUDE.md.

```
                    +------------------+
                    |   Domain Layer   |
                    |  (models, rules) |
                    +--------+---------+
                             |
                    +--------+---------+
                    | Application Layer|
                    | (use cases/ports)|
                    +--+------------+--+
                       |            |
              +--------+--+    +----+--------+
              | CLI Adapter|    | MCP Adapter |
              | (click/    |    | (FastMCP    |
              |  typer)    |    |  server)    |
              +------------+    +-------------+
```

### Concrete Pattern

**Application layer (shared):**

```python
# src/application/commands/init_project.py
class InitProjectCommand:
    def __init__(self, project_dir: str, existing: bool = False):
        self.project_dir = project_dir
        self.existing = existing

class InitProjectHandler:
    def __init__(self, scaffold_port: ScaffoldPort, git_port: GitPort):
        self.scaffold_port = scaffold_port
        self.git_port = git_port

    def execute(self, cmd: InitProjectCommand) -> InitProjectResult:
        # Business logic here -- same regardless of CLI or MCP
        ...
```

**CLI adapter:**

```python
# src/infrastructure/cli/commands.py
import click

@click.command()
@click.argument("project_dir")
@click.option("--existing", is_flag=True)
def init(project_dir: str, existing: bool):
    handler = container.get(InitProjectHandler)
    result = handler.execute(InitProjectCommand(project_dir, existing))
    click.echo(result.summary)
```

**MCP adapter:**

```python
# src/infrastructure/mcp/server.py
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("alto")

@mcp.tool()
def init_project(project_dir: str, existing: bool = False) -> dict[str, str]:
    """Initialize a alto project."""
    handler = container.get(InitProjectHandler)
    result = handler.execute(InitProjectCommand(project_dir, existing))
    return {"status": result.status, "message": result.summary}
```

### Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| CLI framework | click or typer | Well-known, supports subcommands |
| MCP SDK | `mcp` (official, FastMCP) | MIT, protocol compliance |
| Shared logic | Application command handlers | Both adapters call same handlers |
| DI container | Simple factory or `dependency-injector` | Wire ports at startup |
| Return types | Domain result objects | CLI formats as text, MCP as JSON |
| Entry points | Two `pyproject.toml` scripts | `vs` for CLI, `alto-mcp` for server |

### pyproject.toml Entry Points

```toml
[project.scripts]
alto = "src.infrastructure.cli.main:app"
alto-mcp = "src.infrastructure.mcp.server:main"
```

Both entry points wire the same application layer with the same port implementations,
but expose different interfaces (terminal text vs MCP JSON).

---

## 7. Answer: Minimal Tool Boilerplate

The absolute minimum to expose a tool via MCP:

```python
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("alto")

@mcp.tool()
def doc_health(project_dir: str) -> dict[str, list[str]]:
    """Check document freshness and broken references."""
    return {"stale": ["docs/PRD.md"], "broken_refs": []}

if __name__ == "__main__":
    mcp.run()  # stdio transport by default
```

That is 10 lines of code for a fully functional MCP tool with auto-generated JSON Schema,
stdio transport, and protocol compliance.

---

## 8. Lifespan Pattern (Dependency Injection)

The SDK supports async context managers for managing server lifecycle resources:

```python
from contextlib import asynccontextmanager
from collections.abc import AsyncIterator
from dataclasses import dataclass

@dataclass
class AppContext:
    knowledge_service: KnowledgeService
    scaffold_service: ScaffoldService

@asynccontextmanager
async def app_lifespan(server: FastMCP) -> AsyncIterator[AppContext]:
    """Initialize shared services once at server startup."""
    knowledge = KnowledgeService(Path(".alto/knowledge"))
    scaffold = ScaffoldService()
    try:
        yield AppContext(knowledge=knowledge, scaffold=scaffold)
    finally:
        pass  # cleanup if needed

mcp = FastMCP("alto", lifespan=app_lifespan)

@mcp.tool()
async def lookup_knowledge(topic: str, ctx: Context) -> str:
    """Look up DDD knowledge base."""
    app = ctx.request_context.lifespan_context
    return app.knowledge_service.lookup(topic)
```

Source: [MCP SDK README - Lifespan pattern](https://github.com/modelcontextprotocol/python-sdk)

This is the recommended pattern for alto: initialize the application layer services
in the lifespan, access them via `ctx.request_context.lifespan_context` in tool handlers.

---

## 9. alto MCP Tool Inventory (Proposed)

Based on PRD P0 capabilities, here are the MCP tools and resources the server should expose:

### Tools (write operations / analysis)

| Tool Name | CLI Equivalent | Description | Key Parameters |
|-----------|---------------|-------------|----------------|
| `init_project` | `alto init` | Initialize new project | `project_dir`, `existing`, `tools[]` |
| `guide_ddd` | `alto guide` | Start/continue guided DDD discovery | `project_dir`, `readme_content` |
| `generate_artifacts` | `alto generate` | Generate PRD/DDD/ARCH from answers | `project_dir`, `artifact_type` |
| `generate_fitness` | `alto generate fitness` | Generate architecture fitness tests | `project_dir` |
| `generate_tickets` | `alto generate tickets` | Generate beads tickets from DDD | `project_dir`, `preview` |
| `doc_health` | `alto doc-health` | Check document freshness | `project_dir` |
| `ticket_health` | `alto ticket-health` | Check ticket staleness/flags | `project_dir` |
| `classify_subdomain` | (within guide) | Classify Core/Supporting/Generic | `subdomain_name`, `description` |

### Resources (read-only data)

| Resource URI | Description |
|-------------|-------------|
| `alto://knowledge/ddd/{topic}` | DDD patterns and references |
| `alto://knowledge/tools/{tool_name}` | AI tool convention docs |
| `alto://knowledge/conventions/{topic}` | TDD/SOLID/quality gate references |
| `alto://project/{dir}/domain-model` | Current DDD.md content |
| `alto://project/{dir}/architecture` | Current ARCHITECTURE.md |
| `alto://project/{dir}/prd` | Current PRD.md |
| `alto://tickets/ready` | Tickets in ready state |
| `alto://tickets/{id}` | Single ticket details |

### Prompts (optional, for guided flows)

| Prompt Name | Description |
|------------|-------------|
| `ddd_discovery` | Guided DDD question flow with persona adaptation |
| `subdomain_classification` | Interactive Core/Supporting/Generic classification |

---

## 10. Server Composition (Future)

The official SDK supports mounting multiple servers via Starlette ASGI routing:

```python
app = Starlette(routes=[
    Mount("/alto", app=mcp.streamable_http_app()),
])
```

The standalone `fastmcp` v3 package has more advanced composition (Providers, Transforms,
sub-server mounting with prefix namespacing), but this adds a dependency we do not need
for Phase 5. If composition becomes needed later, it can be added without changing the
tool/resource definitions.

---

## 11. Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| SDK v2 breaking changes (v2 pre-alpha on main branch) | Medium | Pin to `mcp>=1.26,<2.0` in pyproject.toml |
| stdio transport limitations for long-running guide flows | Low | Guide flow is stateless Q&A -- fits stdio well |
| FastMCP decorator API differs from low-level API | Low | Use FastMCP exclusively; low-level only if needed |
| MCP protocol version changes | Low | Official SDK tracks protocol versions automatically |
| Standalone `fastmcp` v3 diverges from official API | Medium | Use only `mcp` package; avoid standalone `fastmcp` |

---

## 12. Summary and Recommendation

### Recommendation

Use the **official `mcp` package (v1.26.0, MIT)** with its built-in **FastMCP** high-level
API. Define tools via `@mcp.tool()` decorators with Python type annotations for automatic
schema generation. Use **stdio transport** as the primary mechanism. Share business logic
between CLI and MCP through the **application layer command/query handlers** (hexagonal
architecture already specified in CLAUDE.md).

### Key Finding

The MCP Python SDK's FastMCP API generates JSON Schemas automatically from Python type
annotations and Pydantic models. This means tool definitions stay DRY -- the same type
annotations that drive CLI argument parsing also drive MCP input schemas. The application
layer command handlers are completely transport-agnostic.

### Biggest Risk

The SDK has a v2 pre-alpha on the main branch. Pin to `mcp>=1.26,<2.0` to avoid breaking
changes. Monitor the v2 migration guide when it stabilizes.

### Follow-up Tickets Needed

1. **Task: Define application layer ports** -- Create Protocol interfaces for all P0
   capabilities (init, guide, generate, health checks, knowledge lookup) that both CLI
   and MCP adapters will use.
2. **Task: Implement MCP server adapter** -- Wire FastMCP tool/resource decorators to
   application command handlers. Entry point `alto-mcp` in pyproject.toml.
3. **Task: MCP server configuration generation** -- `alto init` should generate
   `.claude/settings.json` and equivalent Cursor config with the MCP server definition.
4. **Spike: Guided DDD flow over MCP** -- Research how to implement multi-turn
   conversational flows (the guide questions) within MCP's request/response model.
   Options: stateful server with session, stateless with context passing, or MCP prompts.

---

## Sources

- [MCP Python SDK (official) -- GitHub](https://github.com/modelcontextprotocol/python-sdk)
- [MCP Python SDK -- PyPI v1.26.0](https://pypi.org/project/mcp/)
- [FastMCP standalone -- PyPI v3.0.2](https://pypi.org/project/fastmcp/)
- [FastMCP standalone -- GitHub](https://github.com/jlowin/fastmcp)
- [Context7 -- MCP Python SDK documentation](https://context7.com/modelcontextprotocol/python-sdk/llms.txt)
- [MCP Specification -- Transports](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports)
