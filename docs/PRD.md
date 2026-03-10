---
last_reviewed: 2026-02-22
owner: product
status: approved
---

# Product Requirements Document: alty

## 1. Problem Statement

People with software ideas — developers, product owners, domain experts — face a recurring problem: the jump from idea to code skips domain discovery, architecture planning, and structured issue tracking. Developers using AI assistants build fast but with wrong abstractions, anemic domain models, and ad-hoc workflows. Non-coders with domain knowledge have no way to capture it in a form that developers or AI tools can act on. The project conventions, agent profiles, CI configuration, and beads templates need to be recreated from scratch each time. Different AI coding tools (Claude Code, Cursor, Roo Code, OpenCode) have different configuration formats, making it harder to maintain consistency.

## 2. Vision

Someone with an idea — a developer, product owner, or domain expert — describes it in 4-5 sentences. alty guides them through a structured conversation — asking the right DDD and Domain Storytelling questions in plain language — and progressively generates PRD, domain model, bounded contexts, architecture, and beads ticket structure. The generated project works immediately with an AI coding tool of choice, includes agent personas that understand DDD/TDD/SOLID, and has a knowledge base that keeps guidance current. Non-coders get structured handoff artifacts; developers get actionable tickets.

## 3. Users & Personas

| Persona | Description | Primary Need |
|---------|-------------|-------------|
| Solo Developer | Individual building a project with AI assistance | Turn an idea into a well-structured project without manual scaffolding |
| Team Lead | Person setting up conventions for a small team | Consistent project structure, enforced quality gates, shared agent profiles |
| AI Tool Switcher | Developer using multiple AI coding tools | Same project structure and conventions regardless of tool |
| Product Owner | Non-coder who defines what to build but not how | Turn product vision into structured requirements and tickets that developers can execute, without needing to understand DDD or architecture |
| Domain Expert | HR, sales, ops person who spotted a problem worth solving | Describe the problem in business language and get a properly structured project handed off to a developer or AI coding tool |

## 4. User Scenarios

### Scenario 1: New Project Bootstrap

**As a** Solo Developer, **I want to** describe my project idea and have alty guide me through domain discovery, **so that** I get a properly structured project without skipping DDD.

**Flow:**
1. Developer writes 4-5 sentences describing their idea in README
2. alty asks guided DDD questions (actors, domain events, ubiquitous language)
3. Answers are used to generate PRD, DDD artifacts, and architecture skeleton
4. Beads epics and spikes are created for unknowns
5. Developer starts implementation with proper tickets

### Scenario 2: Apply to Existing Project

**As a** Team Lead, **I want to** apply alty conventions to an existing project, **so that** we get structured issue tracking and DDD alignment without starting over.

**Flow:**
1. Team lead runs `alty init --existing` in the project directory
2. alty **creates a new git branch** (`alty/init`) — all changes happen there, never on main
3. alty scans existing code, docs, configs, and folder structure
4. **Gap analysis** — identifies what's missing vs a fully-seeded project:
   - Missing docs (PRD, DDD, ARCHITECTURE)
   - Missing tooling (.claude/, .beads/, quality gates)
   - Missing structure (DDD layers, test mirrors)
   - Existing files that conflict with alty defaults
5. Shows the gap report and proposed changes (preview, like `alty init`)
6. Asks clarifying questions about existing domain (ubiquitous language, bounded contexts)
7. Generates missing artifacts, adapts agent profiles to existing domain language
8. User reviews the branch diff, merges when satisfied

**Branch safety rules:**
- `alty init --existing` MUST be run in a clean git working tree (no uncommitted changes)
- All changes go to `alty/init` branch, never directly to current branch
- If branch already exists, abort with message (user must clean up or use `--force-branch`)
- User merges manually — alty never merges for you
- **Existing tests MUST pass** — after scaffolding, `alty init --existing` runs the project's existing test suite. If any test fails, it rolls back all changes on the branch and reports what broke. Zero test regressions is a hard gate.

### Scenario 3: Tool Adaptation

**As an** AI Tool Switcher, **I want to** generate project configs for Claude Code, Cursor, or OpenCode, **so that** I use the same conventions regardless of which AI tool I open.

**Flow:**
1. Developer selects target tool(s) during bootstrap
2. alty generates tool-specific config files (`.claude/`, `.cursor/`, etc.)
3. Agent personas and commands adapt to each tool's format
4. Quality gates and beads workflow remain identical

### Scenario 4: Product Owner Handoff

**As a** Product Owner, **I want to** describe my product vision and have alty produce structured requirements and tickets, **so that** I can hand off to developers with clear scope and priorities.

**Flow:**
1. PO writes a 4-5 sentence product vision in README
2. alty asks guided questions in business language (no DDD jargon)
3. Answers are translated into PRD, domain model, and beads tickets
4. PO reviews generated tickets — acceptance criteria are in business terms
5. Developer picks up tickets with architecture already decided

### Scenario 5: Domain Expert Idea Capture

**As a** Domain Expert (HR, sales, ops), **I want to** describe a problem I see in my work and get a project started, **so that** my domain knowledge is captured before a developer or AI tool starts building.

**Flow:**
1. Domain expert describes the problem in plain business language
2. alty asks clarifying questions using the expert's own terminology
3. Ubiquitous language glossary is built from the expert's words — not invented by developers
4. Domain stories capture the real workflow before any code is written
5. Output is handed to a developer or AI tool with the domain model already defined

### Scenario 6: Ticket Freshness & Ripple Review

**As a** Solo Developer or Team Lead, **I want** open tickets to be flagged when a completed spike or task changes their context, **so that** I never start work based on stale assumptions.

**Flow:**
1. Developer closes a spike (e.g., k7m.9 competitive research)
2. alty identifies all open tickets that depend on or are siblings of the closed ticket
3. Flagged tickets are marked `review_needed` with a context summary of what changed
4. When a human or AI agent picks up a flagged ticket, it sees the flag and context diff
5. Agent presents suggested updates to the user for approval before starting work
6. User approves, modifies, or dismisses the suggestions
7. Flag is cleared and `last_reviewed` is updated

**Key principle:** The system flags and suggests; the human decides. No automatic ticket rewrites.

## 5. Capabilities

### Must Have (P0)

- [ ] **CLI tool (`vs`)** — Primary user interface for all alty operations (`alty init`, `alty guide`, `alty generate`)
- [ ] **MCP server** — Expose guided bootstrap and knowledge base as MCP tools for AI tool integration
- [ ] **`.alty/` project directory** — Per-project state, knowledge base, and doc maintenance config (see section 5.1)
- [ ] **`alty init` with preview** — Show exactly what will be installed/copied, require user confirmation before any action
- [ ] **Global settings detection** — Detect tool global configs (`~/.claude/`, `~/.cursor/`, etc.), report conflicts with local settings, let user choose resolution per conflict
- [ ] **Existing project adoption (`alty init --existing`)** — Branch-based scaffolding for existing projects: gap report, missing artifact generation, agent profile adaptation (see Scenario 2). Basic structural overlay only; smart migration is P1.
- [ ] **Gap analysis** — Scan existing project, compare against full alty structure, report what's missing/conflicting
- [ ] **Guided project bootstrap** — Conversational flow from README idea to full project structure
- [ ] **DDD question framework** — Structured questions for domain stories, ubiquitous language, bounded contexts, aggregate design
- [ ] **Artifact generation** — Generate PRD, DDD.md, ARCHITECTURE.md from guided answers
- [ ] **Agent personas** — Developer, researcher, tech-lead, PM, QA, security agents with DDD awareness
- [ ] **Beads integration** — Epic/spike/ticket templates enforcing DDD+TDD+SOLID. Every ticket created — whether manually or via `alty generate tickets` — MUST use the appropriate beads template (ticket template for tasks/features, spike template for research). Generated CLAUDE.md must enforce template compliance as step 1 of the grooming checklist. After-close protocol must require follow-up tickets to include template-formatted descriptions (never empty).
- [ ] **Quality gates** — go vet + golangci-lint + go test -race enforced before ticket closure
- [ ] **Architecture fitness function generation** — Generate executable architecture tests (depguard rules, architecture tests) from bounded context map. Enforce layer boundaries, dependency direction, aggregate isolation. Tests run as part of quality gates.
- [ ] **Domain story to ticket pipeline** — Auto-generate dependency-ordered beads epics from DDD artifacts. Tickets include TDD phases, SOLID mapping, acceptance criteria from domain invariants. Preview before creation (human-in-the-loop). Dependencies MUST be formally set via `bd dep add` (not just described in text) so that `bd blocked`/`bd ready`/ripple review can traverse the graph.
- [ ] **Complexity budget** — Classify subdomains as Core/Supporting/Generic during DDD discovery. Core gets full DDD treatment, Supporting gets simple services, Generic gets CRUD/library recommendations. Budget enforced in tickets and fitness functions.
- [ ] **Multi-tool support** — Generate domain-aware configs for Claude Code, Cursor, Roo Code, OpenCode from a single domain model. Configs contain ubiquitous language, bounded context rules, and agent personas tuned to the project.
- [ ] **Knowledge base (RLM)** — Addressable docs for DDD patterns, coding tool conventions
- [ ] **Doc maintenance commands** — Slash commands for doc health, architecture lookup, knowledge refresh (like doc-health, architecture-docs, owasp-docs in Tachikoma)
- [ ] **Ticket freshness & ripple review** — When a ticket closes, traverse the dependency graph and flag open dependents/siblings with `review_needed`. Record a context summary of what the closed ticket produced. `alty ticket-health` reports flagged tickets. Agents picking up flagged tickets must present suggested updates to the user before starting work. Two-tier ticket generation: near-term tickets get full AC, far-term tickets are stubs until promoted. **Implementability validation:** ripple review must check compatibility (do the downstream ticket's assumed interfaces match what was actually delivered?), not just freshness (did something change?). Ticket pipeline must validate implementability before presenting tickets to users — detect unresolved dependencies, signature mismatches between ticket sections, and breaking constructor changes. (See Scenario 6)

### Should Have (P1)

- [ ] **Iterative DDD discovery protocol** — Extend guided discovery from a single Express pass (~15 min) to an optional Deep mode (~60-90 min) with three rounds: (1) Discovery — existing 10-question flow producing DDD.md v1, (2) Challenge — AI Challenger probes for invariant gaps, language ambiguities, boundary questions, and failure modes producing DDD.md v2, (3) Simulate — Customer Simulator traces concrete scenarios through bounded contexts flagging where the model breaks or is silent producing DDD.md v3. Anti-hallucination design: AI never states facts — it asks questions (CHALLENGE-AS-QUESTION) and generates scenarios only from existing entities (SCENARIO-FROM-MODEL). Knowledge Trust Hierarchy: USER_STATED > USER_CONFIRMED > AI_RESEARCHED > AI_INFERRED. Domain research via RLM adapter with source attribution. Convergence metrics track model stabilization. Iteration diffs shown at pomodoro breaks for user approval. Bounded Context Canvas (ddd-crew v5) format in artifact output.
- [ ] **Rescue mode (`alty init --existing`) structural migration** — Beyond scaffolding overlay: scan for implicit bounded contexts, identify anemic models, generate migration tickets with before/after test verification
- [ ] **Tool knowledge versioning** — Maintain current + 3 previous major versions per tool
- [ ] **Knowledge base drift detection** — Detect tool convention changes between versions, flag stale architecture docs vs actual code structure, suggest updates
- [ ] **Spike workflow** — Guided spike creation with clear output goals → ADR docs

### Nice to Have (P2)

- [ ] **Template library** — Domain-specific templates (web API, CLI tool, data pipeline, etc.)
- [ ] **Knowledge auto-update** — Fetch latest tool docs and update knowledge base

### 5.1 `.alty/` Directory (per-project)

Every project initialized with `alty init` gets a `.alty/` directory:

```
.alty/
├── config.toml              # Project-specific alty settings
├── knowledge/               # RLM-addressable knowledge base (copied from seed)
│   ├── ddd/                 # DDD patterns, tactical/strategic references
│   ├── tools/               # AI coding tool conventions (versioned)
│   │   ├── claude-code/     # .claude/ format, agents, commands
│   │   ├── cursor/          # .cursor/ format, rules
│   │   ├── roo-code/        # Roo Code config format
│   │   └── opencode/        # OpenCode config format
│   └── conventions/         # TDD, SOLID, quality gate references
└── maintenance/             # Doc health tracking, review schedules
    └── doc-registry.toml    # Which docs to track, owners, review intervals
```

### 5.2 `alty init` Behavior

**Safety-first approach:**

1. **Preview** — Show everything that will be created/installed (dry-run by default)
2. **Confirm** — User must explicitly agree before any file operations
3. **Never overwrite** — If a file already exists, skip it
4. **Conflict resolution** — If alty wants to create a file that exists, rename ours: `filename_alty.md`
5. **Tool installation** — Optionally install beads, trivy, shannon (show what + ask first)
6. **Global settings detection** — Scan for global configs that override local project settings (see 5.2.1)

#### 5.2.1 Global Settings Detection

AI coding tools have global configs that **always win** over local project settings:

| Tool | Global Location | Overrides |
|------|----------------|-----------|
| Claude Code | `~/.claude/CLAUDE.md`, `~/.claude/settings.json` | Project `.claude/CLAUDE.md` |
| Cursor | `~/.cursor/`, global rules | Project `.cursor/` rules |
| Roo Code | TBD (spike needed) | TBD |
| OpenCode | TBD (spike needed) | TBD |

`alty init` must:

1. **Detect** — Scan known global config paths for each detected tool
2. **Compare** — Check for conflicts between global settings and what alty wants to set locally
3. **Report** — Show conflicts clearly with what the global setting does vs what we want
4. **Ask** — Let user choose per conflict:
   - **Keep global** — skip the local setting (global wins anyway)
   - **Update global** — add/merge into the global config (user must confirm)
   - **Set local anyway** — create the local setting knowing global overrides it (with a warning comment in the file)

**Example with global conflict:**

```
$ alty init

Detecting tools...
  Found: Claude Code (global config at ~/.claude/)
  Found: Beads (already installed)

Global settings scan:
  ⚠ CONFLICT  ~/.claude/CLAUDE.md defines git rules that differ from alty defaults
              Global: "NEVER add Co-Authored-By lines"
              Local:  (alty would set the same — no conflict)
              → OK, compatible

  ⚠ CONFLICT  ~/.claude/settings.json has custom model preferences
              Global sets default model → sonnet
              Local:  alty has no model preference
              → OK, no conflict

  ⚠ CONFLICT  ~/.claude/CLAUDE.md has project-specific paths (src/tachikoma/)
              These reference another project and won't apply here
              → OK, global is scoped to other project

  ⚠ CONFLICT  ~/.claude/settings.json has allowedTools restrictions
              Global restricts: Edit, Write require approval
              Local:  alty agents expect Edit, Write available
              → WARNING: agents may hit permission prompts

              Options:
                [1] Keep global (agents will prompt for permissions)
                [2] Update global to allow Edit, Write
                [3] Note in local CLAUDE.md that agents need these tools

Project files:
  CREATE  .alty/config.toml
  CREATE  .alty/knowledge/ddd/...          (12 files)
  CREATE  .alty/knowledge/tools/...        (8 files)
  ...
  SKIP    .claude/CLAUDE.md                     (already exists)
  CREATE  .claude/agents/developer.md
  ...

  INSTALL trivy                                 [optional]

Proceed? [y/N]
```

**Key principle:** We never silently create local settings that will be overridden by global ones. The user always knows when global wins.

### 5.3 Doc Maintenance Philosophy

Markdown files are living documents. After each epic or at regular intervals:

- `alty doc-health` — Check freshness, broken references, missing metadata (like `/doc-health`)
- `alty doc-review` — Mark docs as reviewed, update `last_reviewed` dates
- `.alty/maintenance/doc-registry.toml` — Tracks which docs to monitor, owners, review cadence

This mirrors the pattern from Tachikoma's `/doc-health`, `/architecture-docs`, and `/owasp-docs` commands but is generalized and project-independent.

## 6. Constraints

### Technical Constraints

| Constraint | Value | Rationale |
|-----------|-------|-----------|
| Language | Go 1.26+ | Target audience and our own stack |
| Package manager | Go modules | Standard Go dependency management |
| CLI name | `vs` (alty) | Short, memorable, unix-style |
| Issue tracking | Beads | Git-native, works offline, no external service |
| Interfaces | CLI (`vs`) + MCP server | CLI for humans, MCP for AI tools |
| AI interaction | Conversational | Must work in terminal (Claude Code) and IDE (Cursor) |

### Non-Functional Requirements

| Requirement | Target | Measurement |
|------------|--------|-------------|
| Bootstrap time | < 30 minutes | From README to first beads ticket |
| Knowledge freshness | < 90 days stale | Per-doc last_reviewed dates |
| Tool coverage | 4 tools | Claude Code, Cursor, Roo Code, OpenCode |

### File Safety Rules (HARD CONSTRAINTS)

| Rule | Behavior |
|------|----------|
| Never overwrite | If target file exists, skip it |
| Conflict rename | If alty needs to create a conflicting file, suffix ours: `filename_alty.md` |
| Preview first | All file operations shown before execution (`alty init` dry-run) |
| Explicit confirm | User must agree before any write/copy/install |
| No silent installs | Tool installation (beads, trivy, shannon) is optional and shown separately |
| Branch for existing | `alty init --existing` always creates a new branch, never writes to current branch |
| Clean tree required | `alty init --existing` refuses to run with uncommitted changes |
| Never merge | alty never merges branches — user merges manually after review |
| **Zero test regression** | `alty init --existing` runs existing test suite after scaffolding — if ANY test fails, roll back all changes and report. This is a **hard gate**, no exceptions. |

### Budget / Resource Constraints

- No cloud dependencies — everything runs locally
- No paid APIs required for core functionality
- Knowledge base maintained manually (RLM pattern, not auto-scraping)

## 7. Out of Scope

- Language support beyond Go (future consideration)
- Package manager support beyond Go modules (future consideration)
- Automated code generation from DDD artifacts (we generate structure, not business logic)
- IDE plugin development (we generate config files, not plugins)
- Hosting or deployment automation

## 8. Success Metrics

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Time from idea to first ticket | < 30 min | Manual timing |
| Projects using correct lifecycle | 100% | All have PRD + DDD + ARCH before code |
| Knowledge base coverage | 4 tools | Claude Code, Cursor, Roo Code, OpenCode docs present |
| Quality gate enforcement | 100% | No ticket closed without passing gates |
| Architecture test generation | 100% of bounded contexts | Every context has at least one fitness function test |
| Ticket pipeline accuracy | Zero manual reordering | Generated dependency order matches actual build order |
| Ticket freshness | Zero stale starts | No ticket claimed as in_progress while `review_needed` is set without reviewing first |

## 9. Risks & Unknowns

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Tool config formats change frequently | High | Medium | Version knowledge base (current + 3 prev) |
| DDD questions too abstract for beginners | Medium | High | Provide concrete examples per question |
| Guided flow feels too rigid | Medium | Medium | Allow skipping with explicit acknowledgment |
| MCP server adds complexity | Low | Medium | Spike first, implement only if justified |
| depguard config too limited for generation | Medium | Medium | Spike k7m.10; fallback to architecture test generation |
| Kiro (AWS) adds DDD support | Low | High | Ship first, establish community, local-first advantage |
| Ticket context decay (AI implements stale specs) | High | High | Ripple review flags dependents on close; freshness check before claiming |

### Open Questions (need spikes)

- [x] ~~**Spike: MCP vs CLI** — Decided: both. CLI (`vs`) for humans, MCP server for AI tools.~~
- [ ] **Spike: CLI + MCP design** — Command tree for `vs`, MCP tool schemas, shared application core
- [ ] **Spike: Knowledge base structure** — How to organize `.alty/knowledge/` with RLM addressability and version tracking?
- [ ] **Spike: Multi-tool config generation** — What are the config formats for Cursor, Roo Code, OpenCode? How similar/different?
- [ ] **Spike: Guided question framework** — What's the minimal effective set of DDD questions to go from idea to bounded contexts? Includes complexity budget classification (Core/Supporting/Generic).
- [ ] **Spike: Fitness function generation** — How to map bounded context map to depguard rules / architecture tests? (k7m.10)
- [ ] **Spike: Ticket pipeline** — How to auto-generate ordered beads tickets from DDD artifacts? (k7m.11)

## 10. Timeline

| Phase | Deliverable |
|-------|-------------|
| Phase 1: Foundation | PRD + DDD + Architecture for alty itself |
| Phase 2: Core CLI | `alty init` + guided DDD questions + artifact generation |
| Phase 3: Fitness & Tickets | Architecture fitness function generation + ticket pipeline + complexity budget |
| Phase 4: Multi-Tool | Config generation for Claude Code, Cursor, Roo Code, OpenCode |
| Phase 5: MCP Server | MCP tools exposing same core as CLI |
| Phase 6: Rescue Mode | `alty init --existing` with structural migration |
| Phase 7: Ticket & Doc Health | `alty ticket-health`, `alty doc-health`, drift detection, ripple review, maintenance registry |
