---
last_reviewed: 2026-02-22
owner: architecture
status: complete
type: spike
ticket: alto-k7m.4
---

# CLI & MCP Server Design: alto

> **Spike:** k7m.4 — Design CLI (`vs`) and MCP server interfaces
> **Timebox:** 4 hours
> **Decisions:** Typer for CLI, official `mcp` SDK (FastMCP) for MCP server, hexagonal shared core

## 1. ADR: CLI Framework Choice

**Status:** Accepted
**Decision:** Use **Typer 0.24.1** (MIT) as the CLI framework for `vs`.
**Context:** See `docs/research/20260222_cli_framework_comparison.md` for full comparison.

### Why Typer over Click or argparse

| Factor | Typer | Click | argparse |
|--------|-------|-------|----------|
| Type hints → CLI | Function signature IS the interface | Decorators only | None |
| Rich output | Built-in (Rich is a dependency) | Needs rich-click | Manual |
| mypy strict | Full compatibility | Partial (decorators obscure types) | No |
| Boilerplate (9 cmds) | ~25 lines | ~45 lines | ~80+ lines |
| Testing | CliRunner (inherited from Click) | CliRunner (mature) | Manual mock |

### What Typer does NOT solve

The 10-question guided DDD discovery flow (persona detection, branching, playback loops)
is too complex for any CLI framework's built-in prompting. This must be a custom
**application-layer service** (`GuidedDiscoveryService`). The `alto guide` command is a
thin Typer adapter calling that service.

### Dependencies added

```
typer 0.24.1
  click >=8.0.0       (CLI framework Typer wraps)
  rich >=10.11.0       (terminal formatting — needed anyway)
  shellingham >=1.3.2  (shell detection — tiny)
```

---

## 2. CLI Command Tree

### Design Principle

Each CLI command maps to **one bounded context entry point** (from `docs/DDD.md` § 4).
Commands are thin infrastructure adapters calling application-layer command/query handlers.

### Command Tree

```
vs
├── init                          # Bootstrap context (orchestrator)
│   └── --existing                # → delegates to Rescue context
├── guide                         # Guided Discovery context
│   ├── --quick                   # 5-question minimum viable mode
│   └── --resume <session-id>     # Resume interrupted session
├── generate                      # Group: generation commands
│   ├── artifacts                 # Domain Model context → PRD, DDD.md, ARCHITECTURE.md
│   ├── fitness                   # Architecture Testing context → import-linter + pytestarch
│   ├── tickets                   # Ticket Pipeline context → beads epics + tasks
│   └── configs                   # Tool Translation context → .claude/, .cursor/, etc.
├── detect                        # Bootstrap context — global settings detection (C5)
├── check                         # Architecture Testing context — quality gate runner (C13)
│   ├── --lint                    # ruff check
│   ├── --types                   # mypy
│   ├── --tests                   # pytest
│   └── --fitness                 # import-linter + pytestarch
├── kb                            # Knowledge Base context — RLM lookup (C18)
│   └── <topic>                   # e.g., alto kb ddd/aggregate
├── doc-health                    # Knowledge Base context — freshness report (C19)
├── doc-review                    # Knowledge Base context — mark docs as reviewed (C19)
├── ticket-health                 # Ticket Freshness context — review_needed report (C20)
├── persona                       # Tool Translation context — agent persona config (C11)
│   ├── list                      # Show available personas
│   └── generate <persona>        # Generate persona config for detected tools
├── version                       # Show version
└── help                          # Show help
```

### Command → Bounded Context → Application Port Mapping

| Command | Bounded Context | Application Port (Protocol) | DDD Aggregate |
|---------|----------------|---------------------------|---------------|
| `alto init` | Bootstrap | `BootstrapPort` | BootstrapSession |
| `alto init --existing` | Rescue (via Bootstrap) | `RescuePort` | GapAnalysis |
| `alto guide` | Guided Discovery | `DiscoveryPort` | DiscoverySession |
| `alto generate artifacts` | Domain Model | `ArtifactGenerationPort` | DomainModel |
| `alto generate fitness` | Architecture Testing | `FitnessGenerationPort` | FitnessTestSuite |
| `alto generate tickets` | Ticket Pipeline | `TicketGenerationPort` | TicketPlan |
| `alto generate configs` | Tool Translation | `ConfigGenerationPort` | ToolConfig |
| `alto detect` | Bootstrap | `ToolDetectionPort` | (part of BootstrapSession) |
| `alto check` | Architecture Testing | `QualityGatePort` | (orchestration) |
| `alto kb <topic>` | Knowledge Base | `KnowledgeLookupPort` | KnowledgeEntry |
| `alto doc-health` | Knowledge Base | `DocHealthPort` | (query) |
| `alto doc-review` | Knowledge Base | `DocReviewPort` | (command) |
| `alto ticket-health` | Ticket Freshness | `TicketHealthPort` | RippleReview |
| `alto persona list` | Tool Translation | `PersonaQueryPort` | (query) |
| `alto persona generate` | Tool Translation | `PersonaGenerationPort` | ToolConfig |

### CLI Entry Point

```toml
# pyproject.toml
[project.scripts]
alto = "src.infrastructure.cli.main:app"
alto-mcp = "src.infrastructure.mcp.server:main"
```

```python
# src/infrastructure/cli/main.py
import typer

app = typer.Typer(
    name="vs",
    help="alto: guided project bootstrapper (DDD + TDD + SOLID)",
    no_args_is_help=True,
)

# Subcommand groups
generate_app = typer.Typer(help="Generate artifacts from DDD model")
app.add_typer(generate_app, name="generate")

persona_app = typer.Typer(help="Manage agent persona configurations")
app.add_typer(persona_app, name="persona")
```

---

## 3. MCP Tool & Resource Schemas

### Design Principle

MCP tools mirror CLI commands. Both call the **same application-layer ports**.
Tools handle write operations; resources handle read-only queries.

### MCP Tools

| Tool Name | CLI Equivalent | Parameters | Returns |
|-----------|---------------|------------|---------|
| `init_project` | `alto init` | `project_dir: str, existing: bool = False` | `InitResult` |
| `guide_ddd` | `alto guide` | `project_dir: str, quick: bool = False` | `DiscoveryResult` |
| `generate_artifacts` | `alto generate artifacts` | `project_dir: str, artifact_type: str` | `GenerationResult` |
| `generate_fitness` | `alto generate fitness` | `project_dir: str` | `FitnessResult` |
| `generate_tickets` | `alto generate tickets` | `project_dir: str, preview: bool = True` | `TicketPlanResult` |
| `generate_configs` | `alto generate configs` | `project_dir: str, tools: list[str]` | `ConfigResult` |
| `detect_tools` | `alto detect` | `project_dir: str` | `DetectionResult` |
| `check_quality` | `alto check` | `project_dir: str, gates: list[str]` | `QualityResult` |
| `doc_health` | `alto doc-health` | `project_dir: str` | `DocHealthResult` |
| `doc_review` | `alto doc-review` | `project_dir: str, docs: list[str]` | `ReviewResult` |
| `ticket_health` | `alto ticket-health` | `project_dir: str` | `TicketHealthResult` |

### MCP Resources

| Resource URI | Description | Data Source |
|-------------|-------------|-------------|
| `alto://knowledge/ddd/{topic}` | DDD patterns/references | `.alto/knowledge/ddd/` |
| `alto://knowledge/tools/{tool}` | AI tool conventions | `.alto/knowledge/tools/` |
| `alto://knowledge/conventions/{topic}` | TDD/SOLID/quality gate refs | `.alto/knowledge/conventions/` |
| `alto://project/{dir}/domain-model` | Current DDD.md | `docs/DDD.md` |
| `alto://project/{dir}/architecture` | Current ARCHITECTURE.md | `docs/ARCHITECTURE.md` |
| `alto://project/{dir}/prd` | Current PRD.md | `docs/PRD.md` |
| `alto://tickets/ready` | Tickets in ready state | beads `bd ready` |
| `alto://tickets/{id}` | Single ticket details | beads `bd show` |

### MCP Server Entry Point

```python
# src/infrastructure/mcp/server.py
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("alto", lifespan=app_lifespan)

@mcp.tool()
def init_project(project_dir: str, existing: bool = False) -> dict[str, str]:
    """Initialize a alto project."""
    handler = ctx.app.bootstrap_handler
    result = handler.execute(InitProjectCommand(project_dir, existing))
    return {"status": result.status, "summary": result.summary}

def main() -> None:
    mcp.run()  # stdio transport
```

---

## 4. Shared Core Architecture

### Application Layer Ports (Protocols)

Both CLI and MCP adapters depend on these ports. Infrastructure adapters implement them.

```
src/application/ports/
├── bootstrap_port.py        # BootstrapPort — alto init orchestration
├── rescue_port.py           # RescuePort — alto init --existing
├── discovery_port.py        # DiscoveryPort — alto guide
├── artifact_generation_port.py  # ArtifactGenerationPort
├── fitness_generation_port.py   # FitnessGenerationPort
├── ticket_generation_port.py    # TicketGenerationPort
├── config_generation_port.py    # ConfigGenerationPort
├── tool_detection_port.py   # ToolDetectionPort — alto detect
├── quality_gate_port.py     # QualityGatePort — alto check
├── knowledge_lookup_port.py # KnowledgeLookupPort — alto kb
├── doc_health_port.py       # DocHealthPort — alto doc-health
├── ticket_health_port.py    # TicketHealthPort — alto ticket-health
└── persona_port.py          # PersonaPort — alto persona
```

### Port Example

```python
# src/application/ports/discovery_port.py
from __future__ import annotations
from typing import Protocol

class DiscoveryPort(Protocol):
    def start_session(self, readme_content: str) -> DiscoverySession: ...
    def detect_persona(self, session_id: str, choice: str) -> Persona: ...
    def answer_question(self, session_id: str, answer: str) -> QuestionResult: ...
    def confirm_playback(self, session_id: str, confirmed: bool) -> PlaybackResult: ...
    def complete(self, session_id: str) -> DiscoveryCompleted: ...
```

### Dependency Flow

```
CLI (Typer)  ──┐
               ├──> Application Ports (Protocols) ──> Domain Models
MCP (FastMCP) ─┘           │
                     Infrastructure Adapters
                     (implement Protocols)
```

**Rules:**
- CLI/MCP adapters ONLY import from `application.ports` and `application.commands/queries`
- Application layer ONLY imports from `domain` and `ports` (interfaces)
- Domain layer has ZERO external dependencies
- Infrastructure implements ports and depends on external libraries

### Wiring (Composition Root)

```python
# src/infrastructure/composition.py
def create_app() -> AppContext:
    """Wire all ports to their implementations."""
    knowledge_service = FileKnowledgeService(Path(".alto/knowledge"))
    scaffold_service = FileScaffoldService()
    beads_service = BeadsService()
    # ... wire all ports
    return AppContext(
        bootstrap_handler=BootstrapHandler(scaffold_service, ...),
        discovery_handler=DiscoveryHandler(knowledge_service, ...),
        # ...
    )
```

Both CLI and MCP call `create_app()` at startup to get the same wired application context.

---

## 5. Persona-Aware UX Design

### PersonaType → Output Adaptation

The CLI adapts output based on the detected persona (from `docs/DDD.md` § 2: Ubiquitous Language).

| PersonaType | Register | Output Style | Example |
|-------------|----------|-------------|---------|
| Solo Developer | Technical | Full DDD terms, code references, aggregate names | "DiscoverySession aggregate with 5 invariants" |
| Team Lead | Technical | DDD terms + team context, emphasis on conventions | "9 bounded contexts, each gets fitness functions" |
| AI Tool Switcher | Technical | Focus on tool-specific output, config differences | "Generated .claude/ and .cursor/ configs" |
| Product Owner | Non-technical | Business language, no DDD jargon, outcome-focused | "We've mapped out 6 business processes and 30 key terms" |
| Domain Expert | Non-technical | Domain language, familiar terminology, story-focused | "Your HR workflow has 4 main steps and 3 business rules" |

### Output Examples by Persona

**Technical register (Solo Developer):**
```
alto guide — DDD Discovery (question 3/10)

  Phase: Primary Story
  Context: Guided Discovery → DiscoverySession

  Q3: Walk me through the primary business process step by step.
      Who does what, with which work objects, and in what order?

  Tip: Think of this as a domain story — "[Actor] [verb] [work object]"
```

**Non-technical register (Product Owner):**
```
alto guide — Project Discovery (question 3/10)

  Let's map out how your product works in practice.

  Q3: Walk me through the main thing your product does, step by step.
      Who's involved, what do they do, and what do they work with?

  Example: "A customer browses products, adds items to cart,
           and checks out with payment"
```

### Implementation

```python
# src/application/queries/format_output.py
class OutputFormatter:
    def __init__(self, persona: PersonaType) -> None:
        self.register = Register.TECHNICAL if persona.is_technical else Register.NON_TECHNICAL

    def format_question(self, question: Question) -> str:
        if self.register == Register.TECHNICAL:
            return question.technical_text
        return question.non_technical_text

    def format_summary(self, model: DomainModel) -> str:
        if self.register == Register.TECHNICAL:
            return f"{len(model.bounded_contexts)} bounded contexts, {len(model.aggregates)} aggregates"
        return f"We mapped {len(model.domain_stories)} business processes and {len(model.terms)} key terms"
```

### CLI Flags

```
alto guide --persona developer     # Force technical register
alto guide --persona po            # Force non-technical register
alto guide                         # Auto-detect via persona question
```

---

## 6. Agent Persona Configuration Surface (C11)

### What This Is

alto generates agent persona configs for AI coding tools. Each persona (developer,
researcher, tech-lead, PM, QA) has domain-aware instructions that reference the project's
ubiquitous language and bounded contexts.

### CLI Surface

```
alto persona list                                    # Show available personas
alto persona generate developer                      # Generate developer persona for detected tools
alto persona generate --all                           # Generate all personas
alto persona generate developer --tool claude-code    # Generate for specific tool
```

### MCP Surface

```python
@mcp.tool()
def generate_persona(persona_name: str, tool: str | None = None) -> dict:
    """Generate agent persona configuration for an AI coding tool."""
    ...

@mcp.resource("alto://personas/{name}")
def get_persona(name: str) -> str:
    """Get persona definition."""
    ...
```

### Generated Output

For Claude Code, `alto persona generate developer` produces `.claude/agents/developer.md`
with the project's ubiquitous language terms, bounded context boundaries, and DDD/TDD/SOLID
rules embedded.

---

## 7. Open Questions from DDD.md

Addressing the 4 open questions from `docs/DDD.md` § 8:

| Question | Recommendation |
|----------|---------------|
| Should MCP server be its own bounded context? | **No.** MCP is an infrastructure adapter (port), not a bounded context. Same as CLI — both are entry points to the same application core. |
| How does `alto doc-health` relate to Ticket Freshness? | **Separate.** Doc freshness is time-based (Knowledge Base context). Ticket freshness is event-based (Ticket Freshness context). They share a health dashboard but have different domain logic. |
| Should Knowledge Base support user contributions? | **Not in P0.** Start curated-only. P2 could add community patterns via PR workflow. |
| How does complexity budget interact with rescue mode? | **Ask the user.** During `alto init --existing`, after gap analysis, prompt for subdomain classification. Don't auto-classify — domain knowledge requires human input. |

---

## 8. Follow-Up Implementation Tickets

| # | Title | Type | Bounded Context | Depends On |
|---|-------|------|----------------|------------|
| 1 | Set up Typer CLI entry point with subcommand stubs | Task | CLI Framework (Generic) | — |
| 2 | Define application layer port Protocols | Task | Application Layer | — |
| 3 | Implement `alto init` (new project flow) | Feature | Bootstrap | 1, 2 |
| 4 | Implement `alto detect` (global settings scan) | Task | Bootstrap | 1, 2 |
| 5 | Implement `alto guide` (10-question DDD flow) | Feature | Guided Discovery | 1, 2 |
| 6 | Implement `alto generate artifacts` | Task | Domain Model | 5 |
| 7 | Implement `alto generate fitness` | Task | Architecture Testing | 6 |
| 8 | Implement `alto generate tickets` | Task | Ticket Pipeline | 6 |
| 9 | Implement `alto generate configs` | Task | Tool Translation | 6 |
| 10 | Implement `alto check` (quality gate runner) | Task | Architecture Testing | 1, 2 |
| 11 | Implement `alto kb` (knowledge lookup) | Task | Knowledge Base | 1, 2 |
| 12 | Migrate `alto doc-health` from bash to Python | Task | Knowledge Base | 1, 2 |
| 13 | Implement `alto ticket-health` | Task | Ticket Freshness | 1, 2 |
| 14 | Implement `alto persona` commands | Task | Tool Translation | 1, 2 |
| 15 | Implement MCP server adapter | Task | MCP Framework (Generic) | 2 |
| 16 | Implement `alto init --existing` (rescue flow) | Feature | Rescue | 3, 5 |
| 17 | Spike: Guided DDD flow over MCP (multi-turn) | Spike | MCP + Guided Discovery | 15 |

---

## 9. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Typer pre-1.0 (v0.24.1) — API could change | Low | Pin in lockfile; FastAPI ecosystem maintains backward compat |
| MCP SDK v2 breaking changes | Medium | Pin `mcp>=1.26,<2.0`; monitor migration guide |
| Guide flow too complex for CLI prompting | Medium | Application-layer service; CLI/MCP are thin adapters |
| Bash `bin/alto` must coexist with Python `vs` | Low | Transitional: bash handles init/doc-health now; Python takes over progressively |
| 10-question flow over MCP (stateful sessions) | Medium | Spike ticket #17; options: stateful server, context passing, MCP prompts |

---

## Sources

- CLI framework comparison: `docs/research/20260222_cli_framework_comparison.md`
- MCP SDK research: `docs/research/20260222_mcp_server_python_sdk.md`
- DDD artifacts: `docs/DDD.md`
- PRD: `docs/PRD.md`
- Existing CLI: `bin/alto`
