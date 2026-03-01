---
last_reviewed: 2026-02-23
owner: architecture
status: draft
---

# Architecture: alty

> **Prerequisites:** This document was written AFTER `docs/PRD.md` (approved 2026-02-22)
> and `docs/DDD.md` (9 bounded contexts, 5 aggregates, 10 subdomains). Architecture
> decisions are informed by domain knowledge, not the other way around.
>
> **Spike inputs:** This document consolidates findings from 7 completed spikes:
> CLI+MCP design (k7m.4), knowledge base structure (k7m.1), multi-tool config formats
> (k7m.3), fitness function design (k7m.10), ticket pipeline design (k7m.11), and
> ripple review design (k7m.12). Every design decision traces to a PRD capability
> or spike ADR.

## 1. Design Principles

1. **Domain purity** -- Domain layer has zero external dependencies. No frameworks, no
   I/O, no file access. Business logic is expressed purely in Python data structures
   and functions. _(Source: DDD.md layer rules; PRD section 6)_

2. **Local-first** -- Everything runs locally. No cloud dependencies, no paid APIs for
   core functionality, no network calls during bootstrap. _(Source: PRD section 6,
   budget/resource constraints)_

3. **Preview before action** -- All file operations, ticket creation, and config
   generation show a preview and require explicit user confirmation before writing
   anything. _(Source: PRD section 5.2, DDD.md Story 1 steps 4/18, Story 4 step 6,
   Story 5 steps 9-10)_

4. **DDD alignment** -- Architecture follows bounded context boundaries from `docs/DDD.md`.
   Each bounded context gets its own module namespace. Cross-context communication uses
   domain events or explicit ports, never direct imports. _(Source: DDD.md section 4
   context map)_

5. **Testability** -- Every component is testable in isolation with dependency injection.
   Application layer depends on port interfaces (Protocols), not concrete implementations.
   _(Source: PRD section 5 P0 quality gates)_

6. **Complexity budget enforcement** -- Architecture treatment level (hexagonal, layered,
   ACL wrapper) is determined by subdomain classification (Core, Supporting, Generic), not
   by developer preference. Core gets the full DDD treatment; Generic gets a thin wrapper.
   _(Source: DDD.md section 3 complexity budget)_

7. **Human-in-the-loop** -- The system flags, suggests, and previews. Humans decide. No
   automatic ticket rewrites, no automatic code generation, no silent file creation.
   _(Source: PRD section 4 scenario 6, DDD.md Story 3 steps 9-10)_

## 2. System Overview

### High-Level Diagram

```
                            User
                         (5 personas)
                             |
                  +----------+----------+
                  |                     |
              CLI (vs)           MCP Server (alty-mcp)
             [Typer 0.24.1]     [FastMCP 1.26.0]
                  |                     |
                  +----------+----------+
                             |
                    Application Layer
                   (13 Ports / Protocols)
                             |
              +--------------+---------------+
              |              |               |
         Domain Layer   Infrastructure   .alty/
         (Pure Python)    Adapters       (Project State)
              |              |               |
     +--------+--------+     |    +-----------+---------+
     | Guided Discovery |    |    | domain-model.yaml   |
     | Domain Model     |    |    | config.toml         |
     | Arch Testing     |    |    | knowledge/          |
     | Ticket Pipeline  |    |    | maintenance/        |
     | Ticket Freshness |    |    +---------------------+
     | Tool Translation |    |
     | Knowledge Base   |    +--- File I/O, Beads CLI,
     | Bootstrap        |         Git, Template Engine,
     | Rescue           |         Tool Detection
     +------------------+
```

### Component Summary

| Component             | Responsibility                                          | Bounded Context      | Classification |
| --------------------- | ------------------------------------------------------- | -------------------- | -------------- |
| `vs` CLI              | Parse commands, format output, delegate to ports        | CLI Framework        | Generic        |
| `alty-mcp` MCP server | Expose tools/resources over stdio, delegate to ports    | MCP Framework        | Generic        |
| 15 Application Ports  | Define interfaces between adapters and domain           | (cross-cutting)      | --             |
| DiscoverySession      | 10-question DDD flow, persona detection, playback       | Guided Discovery     | Core           |
| DomainModel           | Domain stories, ubiquitous language, bounded contexts   | Domain Model         | Core           |
| FitnessTestSuite      | Generate import-linter + pytestarch from context map    | Architecture Testing | Core           |
| TicketPlan            | Dependency-ordered ticket generation with 3-tier detail | Ticket Pipeline      | Core           |
| RippleReview          | Event-driven freshness flagging on ticket close         | Ticket Freshness     | Core           |
| ToolConfig            | Domain model to tool-native config translation          | Tool Translation     | Supporting     |
| KnowledgeEntry        | RLM-addressable docs, TOML-based tool conventions       | Knowledge Base       | Supporting     |
| BootstrapSession      | Orchestrate `alty init` flow                            | Bootstrap            | Supporting     |
| GapAnalysis           | Scan existing projects, generate migration plans        | Rescue               | Supporting     |
| FileScaffoldService   | Render templates, write files with safety rules         | File Generation      | Generic        |
| Composition Root      | Wire ports to implementations at startup                | (infrastructure)     | --             |

## 3. Layer Architecture

Following Hexagonal Architecture (Ports and Adapters) aligned with DDD:

```
+----------------------------------------------------------------------+
|                         Infrastructure                               |
|  +------------------------------------------------------------------+|
|  |  CLI Adapter (Typer)  |  MCP Adapter (FastMCP)  |  File I/O      ||
|  |  Beads CLI Adapter    |  Git Adapter            |  Template Eng. ||
|  +------------------------------------------------------------------+|
|  +------------------------------------------------------------------+|
|  |                      Application Layer                           ||
|  |  +--------------------------------------------------------------+||
|  |  |  Commands (write operations)                                 |||
|  |  |  Queries (read operations)                                   |||
|  |  |  Ports (13 Protocol interfaces)                              |||
|  |  +--------------------------------------------------------------+||
|  |  +--------------------------------------------------------------+||
|  |  |                      Domain Layer                            |||
|  |  |  Models: Entities, Value Objects, Aggregates                 |||
|  |  |  Services: Stateless domain operations                       |||
|  |  |  Events: Domain events (DiscoveryCompleted, etc.)            |||
|  |  +--------------------------------------------------------------+||
|  +------------------------------------------------------------------+|
+----------------------------------------------------------------------+
```

### Layer Rules

| Layer          | Can Depend On                     | Cannot Depend On                           | Enforced By                                   |
| -------------- | --------------------------------- | ------------------------------------------ | --------------------------------------------- |
| Domain         | Nothing (pure Python stdlib only) | Application, Infrastructure, any framework | import-linter `forbidden` + pytestarch `Rule` |
| Application    | Domain, Ports (interfaces only)   | Infrastructure, frameworks                 | import-linter `layers` contract               |
| Infrastructure | Application, Domain               | -- (outermost layer)                       | --                                            |

### Source Layout

```
src/
+-- domain/
|   +-- models/
|   |   +-- discovery_session.py    # DiscoverySession aggregate (Guided Discovery)
|   |   +-- domain_model.py         # DomainModel aggregate (Domain Model)
|   |   +-- fitness_test_suite.py   # FitnessTestSuite aggregate (Architecture Testing)
|   |   +-- ticket_plan.py          # TicketPlan aggregate (Ticket Pipeline)
|   |   +-- ripple_review.py        # RippleReview aggregate (Ticket Freshness)
|   |   +-- persona.py              # Persona, Register value objects
|   |   +-- question.py             # Question, QuestionPhase entities
|   |   +-- contract.py             # Contract, ArchRule entities
|   |   +-- classification.py       # Classification enum, ContractStrictness VO
|   |   +-- bounded_context_spec.py # BoundedContextSpec, ContextRelationship VOs
|   |   +-- generated_ticket.py     # GeneratedEpic, GeneratedTicket entities
|   |   +-- tool_config.py          # ToolConfig aggregate
|   |   +-- knowledge_entry.py      # KnowledgeEntry entity
|   |   +-- bootstrap_session.py    # BootstrapSession aggregate
|   |   +-- gap_analysis.py         # GapAnalysis aggregate
|   +-- services/
|   |   +-- dependency_sorter.py    # Topological sort (Kahn's algorithm)
|   |   +-- complexity_budget.py    # Classification -> treatment level mapping
|   |   +-- contract_generator.py   # BoundedContextMap -> Contract/ArchRule generation
|   |   +-- ticket_generator.py     # DomainModel -> GeneratedEpic/GeneratedTicket
|   +-- events/
|       +-- discovery_events.py     # PersonaDetected, DiscoveryCompleted, etc.
|       +-- model_events.py         # DomainModelGenerated
|       +-- testing_events.py       # FitnessTestsGenerated
|       +-- pipeline_events.py      # TicketPlanApproved
|       +-- freshness_events.py     # TicketFlagged, FlagCleared
|       +-- translation_events.py   # ConfigsGenerated
|       +-- bootstrap_events.py     # BootstrapCompleted
+-- application/
|   +-- commands/
|   |   +-- start_discovery.py      # Begin guided DDD session
|   |   +-- answer_question.py      # Process user answer
|   |   +-- generate_artifacts.py   # Produce DDD.md + domain-model.yaml
|   |   +-- generate_fitness.py     # Produce import-linter TOML + pytestarch tests
|   |   +-- generate_tickets.py     # Produce dependency-ordered beads tickets
|   |   +-- generate_configs.py     # Produce tool-native config files
|   |   +-- init_project.py         # Bootstrap orchestration
|   |   +-- run_ripple.py           # Flag dependents after ticket close
|   +-- queries/
|   |   +-- current_question.py     # What question to ask next
|   |   +-- playback_summary.py     # Summary of answers for confirmation
|   |   +-- ticket_health.py        # Freshness report (flagged count, oldest)
|   |   +-- doc_health.py           # Document freshness report
|   |   +-- knowledge_lookup.py     # RLM knowledge query
|   |   +-- format_output.py        # Persona-aware output formatting
|   +-- ports/
|       +-- bootstrap_port.py       # BootstrapPort
|       +-- rescue_port.py          # RescuePort
|       +-- discovery_port.py       # DiscoveryPort
|       +-- artifact_generation_port.py  # ArtifactGenerationPort
|       +-- fitness_generation_port.py   # FitnessGenerationPort
|       +-- ticket_generation_port.py    # TicketGenerationPort
|       +-- config_generation_port.py    # ConfigGenerationPort
|       +-- tool_detection_port.py  # ToolDetectionPort
|       +-- quality_gate_port.py    # QualityGatePort
|       +-- knowledge_lookup_port.py # KnowledgeLookupPort
|       +-- doc_health_port.py      # DocHealthPort
|       +-- doc_review_port.py     # DocReviewPort
|       +-- ticket_health_port.py   # TicketHealthPort
|       +-- persona_port.py         # PersonaPort
+-- infrastructure/
    +-- cli/
    |   +-- main.py                 # Typer app, subcommand groups
    |   +-- init_cmd.py             # alty init
    |   +-- guide_cmd.py            # alty guide
    |   +-- generate_cmd.py         # alty generate {artifacts,fitness,tickets,configs}
    |   +-- detect_cmd.py           # alty detect
    |   +-- check_cmd.py            # alty check
    |   +-- kb_cmd.py               # alty kb <topic>
    |   +-- doc_health_cmd.py       # alty doc-health
    |   +-- ticket_health_cmd.py    # alty ticket-health
    |   +-- persona_cmd.py          # alty persona {list,generate}
    +-- mcp/
    |   +-- server.py               # FastMCP server, tool/resource registration
    |   +-- tools.py                # MCP tool implementations
    |   +-- resources.py            # MCP resource handlers
    +-- persistence/
    |   +-- file_scaffold.py        # FileScaffoldService (template rendering, file writes)
    |   +-- yaml_parser.py          # BoundedContextMapParser (YAML IR reader)
    |   +-- toml_renderer.py        # ImportLinterContractRenderer (TOML writer)
    |   +-- test_renderer.py        # PyTestArchTestRenderer (Python file writer)
    |   +-- knowledge_store.py      # FileKnowledgeService (TOML/Markdown reader)
    +-- external/
    |   +-- beads_writer.py         # BeadsCliWriter (bd create, bd dep add via subprocess)
    |   +-- git_adapter.py          # Git operations (branch, status, diff)
    |   +-- tool_detector.py        # Detect installed AI tools + global configs
    +-- composition.py              # Composition root: wire all ports to implementations
```

### Architecture Treatment by Classification

The complexity budget (DDD.md section 3) determines the architecture approach per subdomain:

| Classification | Architecture                   | Testing Target                | Fitness Strictness     | Ticket Detail                      | Subdomains                                                                |
| -------------- | ------------------------------ | ----------------------------- | ---------------------- | ---------------------------------- | ------------------------------------------------------------------------- |
| **Core**       | Hexagonal (Ports and Adapters) | >= 90% domain, >= 80% overall | All 4 contract types   | FULL (AC, TDD, SOLID, edge cases)  | Guided Discovery, Architecture Testing, Ticket Pipeline, Ticket Freshness |
| **Supporting** | Simple layered                 | >= 80%                        | layers + forbidden     | STANDARD (AC, basic tests)         | Tool Translation, Knowledge Base, Rescue, Bootstrap                       |
| **Generic**    | ACL wrapper                    | >= 60% boundary               | Single forbidden (ACL) | STUB (integrate + verify boundary) | File Generation, CLI Framework, MCP Framework                             |

_(Source: DDD.md section 3 complexity budget; PRD section 5 P0 complexity budget)_

## 4. Bounded Context Integration

How bounded contexts communicate, from `docs/DDD.md` section 4 context map:

```
[Guided Discovery] --- DiscoveryCompleted event ----> [Domain Model]
[Domain Model] --- DomainModelGenerated event ----> [Architecture Testing]
[Domain Model] --- DomainModelGenerated event ----> [Ticket Pipeline]
[Domain Model] --- DomainModelGenerated event ----> [Tool Translation]
[Knowledge Base] --- ToolConventions (query) ----> [Tool Translation]
[Ticket Pipeline] --- TicketPlanApproved event ----> [Beads (Infrastructure)]
[Beads (Infrastructure)] --- TicketClosed event ----> [Ticket Freshness]
[Architecture Testing] --- FitnessTestsGenerated ----> [File Generation (Infrastructure)]
[Tool Translation] --- ConfigsGenerated ----> [File Generation (Infrastructure)]
[Bootstrap] --- Orchestrates ----> [Guided Discovery, Domain Model, Arch Testing, Ticket Pipeline, Tool Translation]
[Rescue] --- Orchestrates ----> [Bootstrap] (reuses scaffolding flow)
```

| Upstream Context     | Downstream Context   | Integration Pattern                 | Data Format                             |
| -------------------- | -------------------- | ----------------------------------- | --------------------------------------- |
| Guided Discovery     | Domain Model         | Domain Event (DiscoveryCompleted)   | In-memory event object                  |
| Domain Model         | Architecture Testing | Domain Event (DomainModelGenerated) | `.alty/domain-model.yaml`               |
| Domain Model         | Ticket Pipeline      | Domain Event (DomainModelGenerated) | `.alty/domain-model.yaml`               |
| Domain Model         | Tool Translation     | Domain Event (DomainModelGenerated) | `.alty/domain-model.yaml`               |
| Knowledge Base       | Tool Translation     | Query (lookup)                      | TOML entries via KnowledgeLookupPort    |
| Ticket Pipeline      | Beads (external)     | ACL (subprocess)                    | `bd create` + `bd dep add` CLI commands |
| Beads (external)     | Ticket Freshness     | ACL + Domain Event                  | `bd show --json` parsed by ACL adapter  |
| Architecture Testing | File Generation      | ACL                                 | File write via FileScaffoldService      |
| Tool Translation     | File Generation      | ACL                                 | File write via FileScaffoldService      |
| Bootstrap            | All Core/Supporting  | Orchestration                       | Application-layer command dispatch      |

_(Source: DDD.md section 4 context map; CLI+MCP design spike section 4)_

### Event Flow: End-to-End Bootstrap

The complete `alty init` flow crosses all bounded contexts in this order:

```
1. Bootstrap      -> detect installed tools (ToolDetectionPort)
2. Bootstrap      -> show preview, get confirmation
3. Guided Discovery -> 10-question DDD flow (DiscoveryPort)
   emits: DiscoveryCompleted
4. Domain Model   -> generate DDD artifacts (ArtifactGenerationPort)
   writes: docs/DDD.md + .alty/domain-model.yaml
   emits: DomainModelGenerated
5. Architecture Testing -> generate fitness functions (FitnessGenerationPort)
   writes: pyproject.toml [tool.importlinter] + tests/architecture/*.py
   emits: FitnessTestsGenerated
6. Ticket Pipeline -> generate tickets (TicketGenerationPort)
   writes: beads epics + tasks via bd create + bd dep add
   emits: TicketPlanApproved
7. Tool Translation -> generate configs (ConfigGenerationPort)
   reads: Knowledge Base via KnowledgeLookupPort
   writes: .claude/, .cursor/, .roo/, .opencode/ via FileScaffoldService
   emits: ConfigsGenerated
8. Bootstrap      -> emit BootstrapCompleted
```

Each step shows a preview and waits for user approval before proceeding.

## 5. Data Model

### Aggregates and Storage

| Aggregate        | Storage                                   | Rationale                                                   |
| ---------------- | ----------------------------------------- | ----------------------------------------------------------- |
| DiscoverySession | In-memory (session duration)              | Stateful conversation; persisted only when complete         |
| DomainModel      | `.alty/domain-model.yaml` + `docs/DDD.md` | YAML for machine consumption, Markdown for humans           |
| FitnessTestSuite | In-memory during generation               | Output written to `pyproject.toml` + `tests/architecture/`  |
| TicketPlan       | In-memory during generation               | Output written to Beads via `bd create` subprocess          |
| RippleReview     | Beads labels + comments                   | Uses existing beads features; no custom storage needed      |
| ToolConfig       | In-memory during generation               | Output written to `.claude/`, `.cursor/`, etc.              |
| KnowledgeEntry   | `.alty/knowledge/` directory tree         | TOML for tool conventions, Markdown for DDD/convention refs |
| BootstrapSession | In-memory (session duration)              | Orchestration state; no persistence needed                  |
| GapAnalysis      | In-memory during scan                     | Output is a gap report shown to user                        |

### Shared YAML IR: `.alty/domain-model.yaml`

The domain model YAML is the central intermediate representation consumed by multiple
downstream generators. It is produced by `alty generate artifacts` (Domain Model context)
and consumed by:

- **Architecture Testing** -- reads `bounded_contexts` and `subdomains` to generate
  import-linter contracts and pytestarch tests
- **Ticket Pipeline** -- reads the full model to generate dependency-ordered tickets
  with classification-driven detail levels
- **Tool Translation** -- reads `terms`, `bounded_contexts`, and `subdomains` to
  generate domain-aware configs for AI coding tools

_(Source: ticket pipeline spike section 1; fitness function spike section 2)_

#### Schema Summary

```yaml
# .alty/domain-model.yaml
version: "1.0"
project_name: "example-project"
generated_at: "2026-02-23T10:00:00Z"

terms: # Ubiquitous language glossary
  - term: "Order"
    definition: "A customer's request to purchase items"
    context: "Order Management"

subdomains: # Complexity budget
  - name: "Order Management"
    classification: core # core | supporting | generic
    rationale: "..."
    treatment:
      architecture: hexagonal # hexagonal | layered | acl_wrapper
      testing: comprehensive # comprehensive | standard | boundary
      fitness_functions: strict # strict | moderate | minimal
      ticket_detail: full # full | standard | stub

bounded_contexts: # Context map
  - name: "Order Management"
    subdomain: "Order Management"
    responsibility: "..."
    aggregates: # Only required for Core
      - name: "Order"
        root: "Order"
        entities: ["OrderItem"]
        value_objects: ["Money", "OrderStatus"]
        invariants: ["Order total must equal sum..."]
        commands:
          - name: "place_order"
            actor: "Customer"
            produces_event: "OrderPlaced"
        domain_events: ["OrderPlaced", "OrderCancelled"]
    dependencies:
      - context: "Fulfillment"
        type: "domain_event"
        event: "OrderPlaced"

context_map: # Explicit relationships
  - upstream: "Order Management"
    downstream: "Fulfillment"
    pattern: "domain_events"

domain_stories: # For PRD traceability
  - name: "Place Order"
    steps: [...]
    bounded_contexts: ["Order Management"]
    prd_capabilities: ["C1", "C3"]
```

_(Source: ticket pipeline spike section 1 schema; fitness function spike section 2 schema)_

#### Bounded Context Map Schema (for Fitness Functions)

The fitness function generator uses a subset of the same YAML with additional fields:

```yaml
project:
  name: "myproject"
  root_package: "myproject" # Python package name (import-linter root_package)
  src_path: "src" # Relative path to source root (pytestarch scanner)

bounded_contexts:
  - name: "Guided Discovery"
    module_path: "guided_discovery" # Python module under root_package
    classification: core
    layers: [domain, application, infrastructure]
    aggregates: ["discovery_session"]
    relationships:
      - target: "Domain Model"
        direction: downstream
        pattern: domain_event
        via: "infrastructure.events.discovery_completed"
```

_(Source: fitness function spike section 2)_

### Key Entities

| Entity            | Key Attributes                                                       | Aggregate        |
| ----------------- | -------------------------------------------------------------------- | ---------------- |
| DiscoverySession  | persona, register, current_phase, answers, playbacks                 | DiscoverySession |
| Question          | id, phase, technical_text, non_technical_text                        | DiscoverySession |
| DomainStory       | name, steps (actor/action/work_object), bounded_contexts             | DomainModel      |
| BoundedContextMap | contexts, relationships                                              | DomainModel      |
| AggregateDesign   | root, entities, value_objects, invariants, commands, events          | DomainModel      |
| Contract          | name, type (layers/forbidden/independence/acyclic_siblings), modules | FitnessTestSuite |
| ArchRule          | name, type, subject_modules, forbidden_modules, test_class           | FitnessTestSuite |
| GeneratedEpic     | bounded_context, subdomain_classification, tickets                   | TicketPlan       |
| GeneratedTicket   | title, detail_level, aggregate_name, intra_deps, cross_deps          | TicketPlan       |
| ContextDiff       | closed_ticket_id, summary text                                       | RippleReview     |
| FreshnessFlag     | ticket_id, triggering_ticket_id, context                             | RippleReview     |
| KnowledgeEntry    | category, tool, topic, version, content (TOML/Markdown)              | KnowledgeEntry   |
| ToolAdapter       | tool_name, config_format, output_paths                               | ToolConfig       |

## 6. CLI and MCP Interfaces

### 6.1 CLI Command Tree

Each CLI command maps to one bounded context entry point. Commands are thin infrastructure
adapters calling application-layer command/query handlers via ports.

```
vs
+-- init                          # Bootstrap context (orchestrator)
|   +-- --existing                # -> delegates to Rescue context
+-- guide                         # Guided Discovery context
|   +-- --quick                   # 5-question minimum viable mode
|   +-- --resume <session-id>     # Resume interrupted session
|   +-- --persona <type>          # Force persona (developer|po|expert)
+-- generate                      # Group: generation commands
|   +-- artifacts                 # Domain Model -> PRD, DDD.md, ARCHITECTURE.md
|   +-- fitness                   # Architecture Testing -> import-linter + pytestarch
|   +-- tickets                   # Ticket Pipeline -> beads epics + tasks
|   +-- configs                   # Tool Translation -> .claude/, .cursor/, etc.
+-- detect                        # Bootstrap -> global settings detection
+-- check                         # Architecture Testing -> quality gate runner
|   +-- --lint                    # ruff check
|   +-- --types                   # mypy
|   +-- --tests                   # pytest
|   +-- --fitness                 # import-linter + pytestarch
+-- kb                            # Knowledge Base -> RLM lookup
|   +-- <topic>                   # e.g., alty kb ddd/aggregate
+-- doc-health                    # Knowledge Base -> freshness report
+-- doc-review                    # Knowledge Base -> mark docs as reviewed
+-- ticket-health                 # Ticket Freshness -> review_needed report
+-- persona                       # Tool Translation -> agent persona config
|   +-- list                      # Show available personas
|   +-- generate <persona>        # Generate persona config for detected tools
+-- version                       # Show version
+-- help                          # Show help
```

### 6.2 Command to Port Mapping

| Command                   | Bounded Context        | Port (Protocol)          | Aggregate                  |
| ------------------------- | ---------------------- | ------------------------ | -------------------------- |
| `alty init`               | Bootstrap              | `BootstrapPort`          | BootstrapSession           |
| `alty init --existing`    | Rescue (via Bootstrap) | `RescuePort`             | GapAnalysis                |
| `alty guide`              | Guided Discovery       | `DiscoveryPort`          | DiscoverySession           |
| `alty generate artifacts` | Domain Model           | `ArtifactGenerationPort` | DomainModel                |
| `alty generate fitness`   | Architecture Testing   | `FitnessGenerationPort`  | FitnessTestSuite           |
| `alty generate tickets`   | Ticket Pipeline        | `TicketGenerationPort`   | TicketPlan                 |
| `alty generate configs`   | Tool Translation       | `ConfigGenerationPort`   | ToolConfig                 |
| `alty detect`             | Bootstrap              | `ToolDetectionPort`      | (part of BootstrapSession) |
| `alty check`              | Architecture Testing   | `QualityGatePort`        | (orchestration)            |
| `alty kb <topic>`         | Knowledge Base         | `KnowledgeLookupPort`    | KnowledgeEntry             |
| `alty doc-health`         | Knowledge Base         | `DocHealthPort`          | (query)                    |
| `alty doc-review`         | Knowledge Base         | `DocReviewPort`          | (command)                  |
| `alty ticket-health`      | Ticket Freshness       | `TicketHealthPort`       | (query)                    |
| `alty persona`            | Tool Translation       | `PersonaPort`            | ToolConfig                 |

_(Source: CLI+MCP design spike section 2)_

### 6.3 CLI Entry Points

```toml
# pyproject.toml
[project.scripts]
alty = "src.infrastructure.cli.main:app"
alty-mcp = "src.infrastructure.mcp.server:main"
```

### 6.4 MCP Server

The MCP server mirrors CLI commands. Both call the same application-layer ports.
Tools handle write operations; resources handle read-only queries.

**MCP Tools:**

| Tool Name            | CLI Equivalent            | Parameters                                 |
| -------------------- | ------------------------- | ------------------------------------------ |
| `init_project`       | `alty init`               | `project_dir: str, existing: bool = False` |
| `guide_ddd`          | `alty guide`              | `project_dir: str, quick: bool = False`    |
| `generate_artifacts` | `alty generate artifacts` | `project_dir: str, artifact_type: str`     |
| `generate_fitness`   | `alty generate fitness`   | `project_dir: str`                         |
| `generate_tickets`   | `alty generate tickets`   | `project_dir: str, preview: bool = True`   |
| `generate_configs`   | `alty generate configs`   | `project_dir: str, tools: list[str]`       |
| `detect_tools`       | `alty detect`             | `project_dir: str`                         |
| `check_quality`      | `alty check`              | `project_dir: str, gates: list[str]`       |
| `doc_health`         | `alty doc-health`         | `project_dir: str`                         |
| `ticket_health`      | `alty ticket-health`      | `project_dir: str`                         |

**MCP Resources:**

| Resource URI                               | Description                 | Data Source                    |
| ------------------------------------------ | --------------------------- | ------------------------------ |
| `alty://knowledge/tools/{tool}/{subtopic}` | AI tool conventions         | `.alty/knowledge/tools/`       |
| `alty://knowledge/ddd/{topic}`             | DDD patterns/references     | `.alty/knowledge/ddd/`         |
| `alty://knowledge/conventions/{topic}`     | TDD/SOLID/quality gate refs | `.alty/knowledge/conventions/` |
| `alty://knowledge/cross-tool/{topic}`      | Cross-tool mappings         | `.alty/knowledge/cross-tool/`  |
| `alty://project/{dir}/domain-model`        | Current DDD.md              | `docs/DDD.md`                  |
| `alty://project/{dir}/architecture`        | Current ARCHITECTURE.md     | `docs/ARCHITECTURE.md`         |
| `alty://tickets/ready`                     | Tickets in ready state      | beads `bd ready`               |
| `alty://tickets/{id}`                      | Single ticket details       | beads `bd show`                |
| `alty://personas/{name}`                   | Agent persona definition    | Generated persona files        |

_(Source: CLI+MCP design spike sections 3-4; MCP SDK spike)_

### 6.5 Shared Application Core

Both CLI and MCP adapters depend on the same application-layer ports. Neither contains
business logic. The composition root wires ports to implementations at startup.

```
CLI (Typer)  ---+
                +--> Application Ports (Protocols) --> Domain Models
MCP (FastMCP) --+           |
                      Infrastructure Adapters
                      (implement Protocols)
```

**Rules:**

- CLI/MCP adapters ONLY import from `application.ports` and `application.commands/queries`
- Application layer ONLY imports from `domain` and `ports` (interfaces)
- Domain layer has ZERO external dependencies
- Infrastructure implements ports and depends on external libraries

**Composition Root:**

```python
# src/infrastructure/composition.py
def create_app() -> AppContext:
    """Wire all ports to their implementations."""
    knowledge_service = FileKnowledgeService(Path(".alty/knowledge"))
    scaffold_service = FileScaffoldService()
    beads_service = BeadsCliWriter()
    # ... wire all ports
    return AppContext(
        bootstrap_handler=BootstrapHandler(scaffold_service, ...),
        discovery_handler=DiscoveryHandler(knowledge_service, ...),
        fitness_handler=FitnessHandler(scaffold_service, ...),
        ticket_handler=TicketHandler(beads_service, ...),
        # ...
    )
```

Both CLI (`main.py`) and MCP (`server.py`) call `create_app()` at startup to get the
same wired application context.

_(Source: CLI+MCP design spike sections 4, 7)_

### 6.6 Persona-Aware Output

The CLI adapts output based on the detected persona:

| Persona          | Register      | Output Style                                         |
| ---------------- | ------------- | ---------------------------------------------------- |
| Solo Developer   | Technical     | Full DDD terms, aggregate names, code references     |
| Team Lead        | Technical     | DDD terms + team conventions emphasis                |
| AI Tool Switcher | Technical     | Tool-specific output, config differences             |
| Product Owner    | Non-technical | Business language, outcome-focused, no DDD jargon    |
| Domain Expert    | Non-technical | Domain language, story-focused, familiar terminology |

_(Source: CLI+MCP design spike section 5; DDD.md section 2 ubiquitous language)_

## 7. Killer Feature Architectures

### 7.1 Architecture Fitness Function Generation

**PRD reference:** Section 5 P0 "Architecture fitness function generation"
**Spike source:** `docs/research/20260223_fitness_function_design.md`
**ADR:** Hybrid approach accepted -- generate both import-linter TOML and pytestarch tests

#### Pipeline

```
.alty/domain-model.yaml (bounded_contexts section)
        |
        v
BoundedContextMapParser (Infrastructure: YAML reader)
        |
        v
FitnessTestSuite Aggregate (Domain: pure business logic)
  generate_contracts() command
  Validates 5 invariants (from DDD.md section 5)
  Emits Contract + ArchRule entities
        |
   +----+-----+
   |           |
   v           v
ImportLinter    PyTestArch
Contract        Test
Renderer        Renderer
(TOML ->        (Python ->
pyproject.toml) tests/architecture/*.py)
```

#### Contract Generation by Classification

| Classification | import-linter Contracts                                                                                       | pytestarch Rules                                                                       |
| -------------- | ------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| **Core**       | layers + forbidden (cross-context) + forbidden (domain purity) + independence (aggregates) + acyclic_siblings | LayeredArchitecture + cross-context boundary + domain purity + per-aggregate isolation |
| **Supporting** | layers + forbidden (cross-context)                                                                            | LayeredArchitecture + cross-context boundary                                           |
| **Generic**    | forbidden (ACL boundary from all domain layers)                                                               | Single Rule: domain cannot import generic directly                                     |

#### Invariant Enforcement (at generation time, not runtime)

1. Every bounded context must have >= 1 contract
2. Core subdomains must have all 4 contract types
3. Supporting subdomains must have layers + forbidden
4. Generic subdomains must have >= 1 forbidden (ACL boundary)
5. No contract references module outside its BC except via defined relationship

If any invariant is violated, `generate_contracts()` raises a domain error before any
file is written (fail-fast design).

#### Output

- ~28 import-linter contracts in `pyproject.toml` under `[tool.importlinter]`
- ~35 pytestarch tests across `tests/architecture/test_*.py` files
- Shared `tests/architecture/conftest.py` with session-scoped `EvaluableArchitecture` fixture

_(Source: fitness function spike sections 3, 4, 6, 7)_

### 7.2 Domain Story to Ticket Pipeline

**PRD reference:** Section 5 P0 "Domain story to ticket pipeline"
**Spike source:** `docs/research/20260223_ticket_pipeline_design.md`

#### Pipeline

```
.alty/domain-model.yaml
        |
[1. Parse and Validate]     -- Pydantic models in domain layer
        |
[2. Classify Detail Levels] -- subdomain.treatment.ticket_detail
        |
[3. Generate Epics]         -- 1 epic per bounded context
        |
[4. Generate Tickets]       -- FULL/STANDARD/STUB per classification
        |
[5. Topological Sort]       -- Kahn's algorithm on context_map
        |
[6. Intra-Epic Order]       -- VOs -> Entities -> Aggregates -> Commands -> Integration
        |
[7. PRD Traceability]       -- domain_stories[].prd_capabilities coverage check
        |
[8. Preview for Approval]   -- human-in-the-loop
        |
[9. Write to Beads]         -- bd create + bd dep add via subprocess
        |
[10. Verify Graph]          -- bd dep cycles + bd blocked + bd ready
```

#### Ticket Detail Levels

| Level        | Classification | Content                                                                                    | Ticket Count                                              |
| ------------ | -------------- | ------------------------------------------------------------------------------------------ | --------------------------------------------------------- |
| **FULL**     | Core           | AC, TDD phases (RED/GREEN/REFACTOR), SOLID mapping, edge cases, design section, invariants | 1 per aggregate + 1 per command group + 1 per integration |
| **STANDARD** | Supporting     | AC, basic tests, service implementation                                                    | 1 per major responsibility + 1 per integration            |
| **STUB**     | Generic        | One-sentence goal, ACL integration step, boundary test                                     | 1 per BC                                                  |

#### Beads Integration

Decision: Use `bd create` + `bd dep add` via subprocess (not JSONL generation).

Rationale:

1. Beads is an external system with an ACL boundary -- use the official CLI interface
2. `bd create --body-file` avoids shell escaping issues with complex Markdown
3. Formal `bd dep add` is the ONLY reliable way to create traversable dependencies
4. Performance is acceptable: 10-30 tickets at ~100ms/call = 3-6 seconds total

Write sequence: (1) create all epics, (2) create all tickets under epics, (3) set
intra-epic deps, (4) set cross-epic deps, (5) verify with `bd dep cycles` and `bd ready`.

#### BeadsWriterPort Interface

```python
class BeadsWriterPort(Protocol):
    def create_epic(self, title: str, description: str, priority: int = 1) -> str: ...
    def create_ticket(self, title: str, description: str, parent_id: str, ...) -> str: ...
    def set_dependency(self, ticket_id: str, depends_on_id: str) -> None: ...
    def verify_no_cycles(self) -> list[str]: ...
    def get_blocked(self) -> list[str]: ...
    def get_ready(self) -> list[str]: ...
```

_(Source: ticket pipeline spike sections 2, 4, 5, 9)_

### 7.3 Complexity Budget Engine

**PRD reference:** Section 5 P0 "Complexity budget"
**Source:** DDD.md section 3; ticket pipeline spike section 2

The complexity budget flows through the entire system:

```
Guided Discovery
  Q10: "Which parts are truly unique vs commodity?"
        |
        v
classify_subdomain(name, classification, rationale)
  -> SubdomainClassification value object
        |
        v
.alty/domain-model.yaml (subdomains[].treatment)
        |
   +----+----+----+
   |         |         |
   v         v         v
Fitness   Ticket    Tool
Functions Pipeline  Translation
   |         |         |
   v         v         v
Contract   Ticket    Config
Strictness Detail    Detail
Level      Level     Level
```

#### Classification Decision Tree (Khononov)

```
1. Could you buy it? -> YES -> GENERIC
2. Complex rules?    -> NO  -> SUPPORTING
3. Copied by competitor threatens business? -> NO -> SUPPORTING
4. All YES -> CORE
```

#### Treatment Level Mapping

| Aspect            | Core                                  | Supporting                    | Generic               |
| ----------------- | ------------------------------------- | ----------------------------- | --------------------- |
| Architecture      | Hexagonal                             | Simple layered                | ACL wrapper           |
| Testing target    | >= 90% domain, >= 80% overall         | >= 80%                        | >= 60% boundary       |
| Fitness functions | strict (4 contract types)             | moderate (layers + forbidden) | minimal (1 forbidden) |
| Ticket detail     | FULL                                  | STANDARD                      | STUB                  |
| Domain model      | Rich (aggregates, invariants, events) | Service-oriented              | Adapter only          |

_(Source: DDD.md section 3; fitness function spike section 3)_

### 7.4 Tool-Native Context Translation

**PRD reference:** Section 5 P0 "Multi-tool support"
**Source:** Knowledge base spike sections 1-3, 9

#### Supported Tools

| Tool              | Config Dir   | Agent Format                                      | Instructions Format                     | Global Config                      |
| ----------------- | ------------ | ------------------------------------------------- | --------------------------------------- | ---------------------------------- |
| Claude Code 2.1.x | `.claude/`   | Markdown + YAML frontmatter (`.claude/agents/`)   | Markdown (CLAUDE.md, rules/)            | `~/.claude/` (file-based)          |
| Cursor 2.5.x      | `.cursor/`   | N/A (rules only)                                  | MDC (`.cursor/rules/*.mdc`) + AGENTS.md | SQLite DB (detect only)            |
| Roo Code 3.38.x   | `.roo/`      | YAML (`.roomodes`)                                | Markdown (`.roo/rules/`) + AGENTS.md    | `~/.roo/` (file-based)             |
| OpenCode (latest) | `.opencode/` | Markdown + YAML frontmatter (`.opencode/agents/`) | Markdown (AGENTS.md, rules/)            | `~/.config/opencode/` (file-based) |

#### Cross-Tool Bridge: AGENTS.md

AGENTS.md (Agentic AI Foundation, Linux Foundation) is the emerging cross-tool standard.
Supported natively by Cursor, Roo Code, OpenCode. Claude Code uses CLAUDE.md instead.

**Strategy:** Generate both `AGENTS.md` (cross-tool common denominator) and tool-specific
configs (for tools that support richer features like agents, modes, skills).

#### Generation Matrix

From `.alty/knowledge/cross-tool/generation-matrix.toml`:

| Output               | Claude Code                              | Cursor                      | Roo Code                           | OpenCode                        |
| -------------------- | ---------------------------------------- | --------------------------- | ---------------------------------- | ------------------------------- |
| Project instructions | `.claude/CLAUDE.md`                      | `AGENTS.md`                 | `AGENTS.md`                        | `AGENTS.md`                     |
| Agent personas       | `.claude/agents/{persona}.md`            | `.cursor/rules/{topic}.mdc` | `.roomodes` + `.roo/rules-{slug}/` | `.opencode/agents/{persona}.md` |
| Settings             | `.claude/settings.json`                  | --                          | --                                 | `opencode.json`                 |
| Rules                | `.claude/rules/`                         | `.cursor/rules/`            | `.roo/rules/`                      | `.opencode/rules/`              |
| Commands             | `.claude/commands/`                      | --                          | --                                 | `.opencode/commands/`           |
| MCP config           | `.mcp.json`                              | --                          | --                                 | `opencode.json`                 |
| Gitignore entries    | `settings.local.json`, `CLAUDE.local.md` | --                          | --                                 | --                              |

#### Concept Mapping

| Concept              | Claude Code                  | Cursor                | Roo Code        | OpenCode                       |
| -------------------- | ---------------------------- | --------------------- | --------------- | ------------------------------ |
| Persona/Agent        | Subagent                     | Rule file             | Mode            | Agent                          |
| Global instructions  | `~/.claude/CLAUDE.md`        | User Rules (SQLite)   | `~/.roo/rules/` | `~/.config/opencode/AGENTS.md` |
| Project instructions | `.claude/CLAUDE.md` + rules/ | `.cursor/rules/*.mdc` | `.roo/rules/`   | `AGENTS.md` + rules/           |

#### Limitations

- **Cursor global config is SQLite** -- alty cannot generate or compare global
  config files for Cursor. Can detect the DB file exists but cannot read settings without
  SQLite queries. `alty detect` warns users to check manually.
- **No agent/persona concept in Cursor** -- personas encoded as rule files instead.

_(Source: knowledge base spike sections 1-4, 9)_

### 7.5 Rescue Mode (Existing Project Adoption)

**PRD reference:** Section 4 scenario 2; section 5 P0 "Existing project adoption"
**Status:** P0 for basic structural overlay; P1 for smart migration

#### Pipeline

```
alty init --existing
        |
[1. Verify clean git tree]     -- abort if dirty
        |
[2. Create branch]             -- alty/init (abort if exists)
        |
[3. Scan existing project]     -- code, docs, configs, folder structure
        |
[4. Gap analysis]              -- compare against fully-seeded reference
        |
[5. Show gap report]           -- preview: what's missing, what conflicts
        |
[6. Ask DDD questions]         -- adapted for existing domain
        |
[7. Generate missing artifacts] -- PRD, DDD, ARCHITECTURE stubs
        |
[8. Adapt agent profiles]      -- existing domain language
        |
[9. Run existing test suite]   -- HARD GATE: zero regressions
        |
[10. If pass: user reviews branch diff, merges manually]
[    If fail: roll back all changes, report what broke]
```

#### Branch Safety Rules

| Rule                           | Enforcement                                                     |
| ------------------------------ | --------------------------------------------------------------- |
| Never overwrite existing files | Skip if target exists                                           |
| Clean git tree required        | `git status --porcelain` check before any operation             |
| All changes on branch          | `git checkout -b alty/init`                                     |
| Never merge for user           | User reviews diff and merges manually                           |
| Zero test regression           | Run existing test suite after scaffolding; roll back on failure |

_(Source: PRD section 4 scenario 2, section 5.2 behavior, section 6 file safety rules)_

### 7.6 Living Knowledge Base

**PRD reference:** Section 5 P0 "Knowledge base (RLM)"; section 5.1 `.alty/` directory
**Spike source:** `docs/research/20260222_knowledge_base_structure.md`

#### Directory Structure

```
.alty/knowledge/
  _index.toml                     # Master index for RLM O(1) lookup
  tools/
    claude-code/
      _meta.toml                  # Tool metadata (name, versions tracked, changelog URL)
      current/                    # Alias -> latest tracked version
        config-structure.toml     # File tree, formats, paths
        agent-format.toml         # Agent definition schema
        settings-format.toml      # settings.json schema
        rules-format.toml         # CLAUDE.md conventions
        commands-format.toml      # Slash command format
        mcp-config.toml           # MCP server config format
        global-paths.toml         # Global config paths per OS
        gitignore-patterns.toml   # What to .gitignore
      v2.1/                       # Explicit version (current alias target)
      v2.0/                       # Previous major version
    cursor/
      _meta.toml
      current/
        config-structure.toml
        rules-format.toml
        agents-md-support.toml
        global-paths.toml
    roo-code/
      _meta.toml
      current/
        config-structure.toml
        mode-format.toml
        rules-format.toml
        global-paths.toml
    opencode/
      _meta.toml
      current/
        config-structure.toml
        agent-format.toml
        mode-format.toml
        rules-format.toml
        opencode-json-schema.toml
        global-paths.toml
  cross-tool/
    agents-md.toml                # AGENTS.md cross-tool standard
    concept-mapping.toml          # How concepts map across tools
    generation-matrix.toml        # What alty generates per tool
  ddd/
    tactical-patterns.md          # Entities, VOs, Aggregates
    strategic-patterns.md         # Bounded Contexts, Context Maps
    event-storming.md             # Event Storming reference
    domain-storytelling.md        # Domain Storytelling reference
  conventions/
    tdd.md                        # RED/GREEN/REFACTOR reference
    solid.md                      # SOLID principles reference
    quality-gates.md              # ruff + mypy + pytest conventions
```

#### RLM Addressing Scheme

Every knowledge entry is addressable by a deterministic path:

```
alty://knowledge/{category}/{tool_or_topic}/{subtopic}?version={version}
```

Resolution is O(1) -- direct path construction, no search, no index scan:

```python
def _resolve_path(self, category: str, topic: str, version: str) -> Path:
    base = self.knowledge_dir / category
    if category == "tools":
        tool, subtopic = topic.split("/", 1)
        return base / tool / version / f"{subtopic}.toml"
    else:
        return base / f"{topic}.md"
```

#### Versioning

Track major version series (not every patch). Current + 3 previous major versions per tool.

| Tool        | Current | Prev 1 | Prev 2 | Prev 3 |
| ----------- | ------- | ------ | ------ | ------ |
| Claude Code | 2.1.x   | 2.0.x  | 1.x    | --     |
| Cursor      | 2.5.x   | 2.4.x  | 2.0.x  | 1.7.x  |
| Roo Code    | 3.38.x  | 3.22.x | 2.2.x  | --     |
| OpenCode    | latest  | --     | --     | --     |

#### TOML for Tool Knowledge, Markdown for Reference

Tool convention entries are structured data (TOML) consumed by `KnowledgeLookupPort` and
`ConfigGenerationPort`. They need machine-parseable fields, deterministic keys for O(1)
lookup, and easy diffing for drift detection.

Markdown is used for DDD and convention reference material (human consumption).

#### Drift Detection

Every `.toml` entry has a `[_meta]` section with staleness signals:

```toml
[_meta]
last_verified = "2026-02-22"
verified_against = "v2.1.15"
confidence = "high"                # high | medium | low
next_review_date = "2026-05-22"    # 90-day freshness window (PRD NFR)
schema_version = 1
```

`alty doc-health --knowledge` compares these fields against installed tool versions
(from `alty detect`) to report stale entries.

_(Source: knowledge base spike sections 4-7)_

### 7.7 Ticket Freshness and Ripple Review

**PRD reference:** Section 5 P0 "Ticket freshness and ripple review"
**Spike source:** `docs/research/20260223_ripple_review_design.md`

#### Data Model (Labels + Comments, No Custom Fields)

Beads v0.55.4 has a fixed schema. All freshness metadata uses native features:

| Concept            | Beads Mechanism                   | Format                                    |
| ------------------ | --------------------------------- | ----------------------------------------- |
| review_needed flag | `bd label add <id> review_needed` | Label                                     |
| Triggering ticket  | Comment text                      | `**Triggered by:** \`<closed-id>\``       |
| Context diff       | Comment body                      | Structured Markdown with review checklist |
| last_reviewed      | Comment prefix                    | `**Reviewed:** <ISO-date> by <actor>`     |
| Flag stacking      | Multiple comments                 | Each closure adds a new comment           |

#### Ripple Review Traversal

When a ticket is closed, `bd-ripple` traverses:

1. **Siblings** -- children of the same parent epic
2. **Dependents** -- tickets with `blocks` dependency on the closed ticket
3. **Related** -- tickets with `related` dependency (both directions)

Only open/in_progress tickets are flagged. Closed tickets are skipped.

#### After-Close Protocol (4 Steps)

Generated into every bootstrapped project's CLAUDE.md:

```
Step 1: Ripple Review
  bin/bd-ripple <closed-id> "<what this ticket produced>"
  -> flags open dependents/siblings with review_needed + context diff comment

Step 2: Review Flagged Tickets
  bd query label=review_needed
  -> for each: read ripple comments, draft suggested updates, present to user
  -> user approves/edits/dismisses
  -> clear flag, add review comment

Step 3: Follow-Up Tickets
  -> create with templates (NEVER empty descriptions)
  -> set formal deps with bd dep add
  -> far-term tickets use stub format

Step 4: Groom Next
  bd ready -> pick highest-priority -> run 7-step grooming checklist
  -> present results, ask user if ready to start
```

#### `alty ticket-health` Report

Read-only freshness report via `TicketHealthPort`:

- Flagged ticket count and list
- Oldest unreviewed ticket
- Per-epic freshness percentage: `(open - flagged) / open * 100`
- Thresholds: 90-100% healthy, 70-89% acceptable, below 70% action needed

#### Two-Tier Ticket Generation

| Tier      | Criteria                                     | Detail Level     |
| --------- | -------------------------------------------- | ---------------- |
| Near-term | Depth <= 2 hops from root, or Core subdomain | FULL             |
| Far-term  | Depth > 2, Supporting or Generic             | STANDARD or STUB |

Stub tickets are promoted to full detail when their blockers are resolved (detected by
ripple review, executed by agent during grooming).

#### Invariants (from DDD.md, enforced by design)

1. Non-empty context diff required (bd-ripple aborts if empty)
2. Only open tickets flagged (closed tickets skipped)
3. Flag stacking (label is idempotent; comments accumulate)
4. No auto-clear (explicit human review required)

_(Source: ripple review spike sections 1-6)_

## 8. `.alty/` Project Directory

Every project initialized with `alty init` gets this directory:

```
.alty/
+-- config.toml                   # Project-specific alty settings
+-- domain-model.yaml             # Machine-readable DDD IR (generated)
+-- knowledge/                    # RLM-addressable knowledge base (copied from seed)
|   +-- _index.toml               # Master index
|   +-- tools/                    # AI coding tool conventions (versioned TOML)
|   |   +-- claude-code/
|   |   +-- cursor/
|   |   +-- roo-code/
|   |   +-- opencode/
|   +-- cross-tool/               # Cross-tool mappings
|   +-- ddd/                      # DDD pattern references (Markdown)
|   +-- conventions/              # TDD, SOLID, quality gate references (Markdown)
+-- maintenance/                  # Doc health tracking
    +-- doc-registry.toml         # Which docs to track, owners, review intervals
```

_(Source: PRD section 5.1; knowledge base spike section 4)_

## 9. External Integrations

| Integration      | Purpose                                                          | Protocol                    | Auth         | Bounded Context                        |
| ---------------- | ---------------------------------------------------------------- | --------------------------- | ------------ | -------------------------------------- |
| Beads (`bd` CLI) | Issue tracking: create/read/update tickets, dependencies, labels | subprocess CLI calls        | None (local) | Ticket Pipeline, Ticket Freshness      |
| Git              | Branch management for rescue mode; status checks                 | subprocess (`git`)          | None (local) | Rescue, Bootstrap                      |
| ruff             | Python linting quality gate                                      | subprocess                  | None (local) | Architecture Testing (QualityGatePort) |
| mypy             | Type checking quality gate                                       | subprocess                  | None (local) | Architecture Testing (QualityGatePort) |
| pytest           | Test execution quality gate                                      | subprocess                  | None (local) | Architecture Testing (QualityGatePort) |
| import-linter    | Architecture fitness function execution                          | subprocess (`lint-imports`) | None (local) | Architecture Testing                   |
| pytestarch       | Architecture fitness function execution                          | pytest (in-process)         | None (local) | Architecture Testing                   |

All integrations are local. No network calls, no cloud dependencies, no paid APIs.

_(Source: PRD section 6 constraints; CLI+MCP design spike section 2)_

## 10. Security

### Trust Boundaries

```
User Input (README, answers) ---- Validation ----> Domain Models (typed, validated)
                                     |
                                     v
External Tools (beads, git) <---- ACL Layer <---- Application Commands
                                     |
                                     v
File System ---- Safety Rules ----> File writes (preview + confirm + never overwrite)
```

### Security Measures

| Concern                       | Mitigation                                                                                                                                                                                  |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Input validation              | All user input (README content, question answers, persona selection) validated by Pydantic models in the domain layer before processing                                                     |
| File safety                   | Never overwrite existing files. Conflict rename: `filename_alty.md`. Preview all writes. Explicit confirm before any action. _(PRD section 6 file safety rules)_                            |
| Subprocess injection          | All subprocess calls use list-form arguments (not shell=True). Ticket content written to temp files via `--body-file`, never passed as shell arguments. _(ticket pipeline spike section 4)_ |
| Branch safety                 | `alty init --existing` always creates a new branch. Never writes to current branch. Never merges. Requires clean git tree. Zero test regression hard gate. _(PRD section 4 scenario 2)_     |
| No silent installs            | Tool installation (beads, trivy, shannon) is optional and shown separately in preview. _(PRD section 5.2)_                                                                                  |
| Global config detection       | `alty detect` scans for global AI tool configs that override local settings. Reports conflicts. Lets user choose resolution per conflict. _(PRD section 5.2.1)_                             |
| No secrets in generated files | Generated configs contain project structure and domain terms, not API keys, passwords, or personal information.                                                                             |
| No network access             | All operations are local-only. No cloud dependencies, no phone-home, no telemetry. _(PRD section 6 budget constraints)_                                                                     |

## 11. Deployment

| Aspect               | Choice                                                     | Rationale                                                                                          |
| -------------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Runtime              | Python 3.12+                                               | Target audience and our own stack _(PRD section 6)_                                                |
| Package manager      | uv                                                         | Speed, reproducibility, modern standard _(PRD section 6)_                                          |
| CLI framework        | Typer 0.24.1 (MIT)                                         | Type hints = CLI interface; Rich bundled; CliRunner for testing _(CLI framework spike ADR)_        |
| MCP framework        | `mcp` SDK 1.26.0 (MIT), pin `>=1.26,<2.0`                  | Official SDK; FastMCP API; stdio transport _(MCP SDK spike ADR)_                                   |
| Architecture testing | import-linter 2.10 (BSD-2) + pytestarch 4.0.1 (Apache-2.0) | Complementary coverage: TOML config + Python tests _(fitness function spike ADR)_                  |
| TOML editing         | tomlkit (MIT)                                              | Round-trip preservation of existing pyproject.toml formatting _(fitness function spike section 8)_ |
| Issue tracking       | Beads v0.55.4+                                             | Git-native, works offline, embedded Dolt backend                                                   |
| Distribution         | PyPI package                                               | `uv tool install alty` or `pip install alty`                                                       |
| Entry points         | `vs` (CLI) + `alty-mcp` (MCP server)                       | Both defined in `pyproject.toml [project.scripts]`                                                 |

### Dependencies

```toml
# pyproject.toml
[project]
dependencies = [
    "typer>=0.24,<1.0",          # CLI framework (includes click, rich, shellingham)
    "mcp>=1.26,<2.0",            # MCP server SDK
    "pydantic>=2.0",             # Domain model validation
    "tomlkit>=0.12",             # Round-trip TOML editing for pyproject.toml
    "pyyaml>=6.0",               # YAML IR parsing
]

[project.optional-dependencies]
dev = [
    "ruff",
    "mypy",
    "pytest",
    "pytest-cov",
    "import-linter>=2.10",       # Architecture fitness functions
    "pytestarch>=4.0",           # Architecture fitness functions
]
```

## 12. Constraints and Budgets

From `docs/PRD.md` section 6:

| Resource              | Limit                                            | Rationale                                                      |
| --------------------- | ------------------------------------------------ | -------------------------------------------------------------- |
| Bootstrap time        | < 30 minutes                                     | From README to first beads ticket _(PRD section 8)_            |
| Knowledge freshness   | < 90 days stale                                  | Per-doc `next_review_date` in TOML metadata _(PRD NFR)_        |
| Tool coverage         | 4 tools                                          | Claude Code, Cursor, Roo Code, OpenCode _(PRD NFR)_            |
| Cloud dependencies    | Zero                                             | Everything runs locally _(PRD section 6)_                      |
| Paid API dependencies | Zero                                             | Core functionality requires no paid services _(PRD section 6)_ |
| Python version        | 3.12+                                            | Target audience stack _(PRD section 6)_                        |
| File safety           | Never overwrite, preview first, explicit confirm | 9 file safety rules _(PRD section 6)_                          |
| Test regression       | Zero on `alty init --existing`                   | Hard gate, no exceptions _(PRD section 6)_                     |

## 13. Architecture Decision Records

| ADR     | Decision                                                                                       | Status   | Source                                                          |
| ------- | ---------------------------------------------------------------------------------------------- | -------- | --------------------------------------------------------------- |
| ADR-001 | CLI framework: Typer 0.24.1 (MIT) over Click and argparse                                      | Accepted | `docs/research/20260222_cli_framework_comparison.md`            |
| ADR-002 | MCP framework: official `mcp` SDK v1.26.0 (MIT) with FastMCP, pin `>=1.26,<2.0`                | Accepted | `docs/research/20260222_mcp_server_python_sdk.md`               |
| ADR-003 | Architecture testing: hybrid import-linter TOML + pytestarch Python tests                      | Accepted | `docs/research/20260223_fitness_function_design.md` section 5   |
| ADR-004 | Knowledge base: TOML for tool conventions (machine), Markdown for DDD/conventions (human)      | Accepted | `docs/research/20260222_knowledge_base_structure.md` section 10 |
| ADR-005 | Ticket pipeline: `bd create` + `bd dep add` via subprocess (not JSONL generation)              | Accepted | `docs/research/20260223_ticket_pipeline_design.md` section 4    |
| ADR-006 | Ripple review: labels + comments in beads (no custom fields, no beads schema changes)          | Accepted | `docs/research/20260223_ripple_review_design.md` section 1      |
| ADR-007 | Shared YAML IR at `.alty/domain-model.yaml` consumed by fitness, tickets, and tool translation | Accepted | `docs/research/20260223_ticket_pipeline_design.md` section 1    |
| ADR-008 | Cross-tool bridge: generate both AGENTS.md and tool-specific configs                           | Accepted | `docs/research/20260222_knowledge_base_structure.md` section 3  |
| ADR-009 | TOML editing: tomlkit for round-trip pyproject.toml preservation                               | Accepted | `docs/research/20260223_fitness_function_design.md` section 8   |
| ADR-010 | 13 application-layer ports (Protocols) shared between CLI and MCP                              | Accepted | `docs/research/20260222_cli_mcp_design.md` section 4            |
| ADR-011 | Composition root at `src/infrastructure/composition.py`                                        | Accepted | `docs/research/20260222_cli_mcp_design.md` section 4            |
| ADR-012 | MCP server is an infrastructure adapter, not a bounded context                                 | Accepted | `docs/research/20260222_cli_mcp_design.md` section 7            |

## 14. Open Architecture Decisions

Decisions resolved by spikes but requiring validation during implementation:

- [ ] **pytestarch module resolution** -- `pytestarch` resolves modules relative to the scan
      path. If `src_path = "src"`, modules may be `src.myproject.domain` not `myproject.domain`.
      Verify correct path convention during implementation. _(fitness function spike section 9)_

- [ ] **Import-linter `include_external_packages`** -- Should domain purity contracts
      default to `include_external_packages = true` to catch framework imports (django,
      sqlalchemy)? Likely yes, but needs implementation validation. _(fitness function spike
      section 9)_

- [ ] **Regeneration without losing manual edits** -- Users may add custom contracts or
      tests. Regeneration should preserve user-added items. Design a `# alty:generated`
      marker convention. _(fitness function spike section 9)_

- [ ] **Guided DDD flow over MCP (stateful sessions)** -- The 10-question flow is stateful.
      MCP tools are normally stateless request/response. Options: stateful server context,
      context passing, MCP prompts. _(CLI+MCP design spike section 9, risk #5)_

- [ ] **Knowledge base maintenance burden** -- 4 tools x ~7 topics x 4 versions = ~112
      TOML files. Start with `current/` only; add historical versions only on breaking changes.
      _(knowledge base spike section 10 risk)_
