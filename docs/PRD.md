---
last_reviewed: 2026-02-22
owner: product
status: draft
---

# Product Requirements Document: vibe-seed

## 1. Problem Statement

Developers starting new "vibe coding" projects with AI assistants face a recurring problem: they jump straight from idea to code, skipping domain discovery, architecture planning, and structured issue tracking. This leads to projects with wrong abstractions, anemic domain models, and ad-hoc workflows. The project conventions, agent profiles, CI configuration, and beads templates need to be recreated from scratch each time. Different AI coding tools (Claude Code, Cursor, Antigravity, OpenCode) have different configuration formats, making it harder to maintain consistency.

## 2. Vision

A developer describes their project idea in 4-5 sentences. vibe-seed guides them through a structured conversation — asking the right DDD and Domain Storytelling questions — and progressively generates PRD, domain model, bounded contexts, architecture, and beads ticket structure. The generated project works immediately with their AI coding tool of choice, includes agent personas that understand DDD/TDD/SOLID, and has a knowledge base that keeps guidance current.

## 3. Users & Personas

| Persona | Description | Primary Need |
|---------|-------------|-------------|
| Solo Developer | Individual building a Python project with AI assistance | Turn an idea into a well-structured project without manual scaffolding |
| Team Lead | Person setting up conventions for a small team | Consistent project structure, enforced quality gates, shared agent profiles |
| AI Tool Switcher | Developer using multiple AI coding tools | Same project structure and conventions regardless of tool |

## 4. User Scenarios

### Scenario 1: New Project Bootstrap

**As a** Solo Developer, **I want to** describe my project idea and have vibe-seed guide me through domain discovery, **so that** I get a properly structured project without skipping DDD.

**Flow:**
1. Developer writes 4-5 sentences describing their idea in README
2. vibe-seed asks guided DDD questions (actors, domain events, ubiquitous language)
3. Answers are used to generate PRD, DDD artifacts, and architecture skeleton
4. Beads epics and spikes are created for unknowns
5. Developer starts implementation with proper tickets

### Scenario 2: Apply to Existing Project

**As a** Team Lead, **I want to** apply vibe-seed conventions to an existing project, **so that** we get structured issue tracking and DDD alignment without starting over.

**Flow:**
1. Team lead runs `vs init --existing` in the project directory
2. vibe-seed **creates a new git branch** (`vibe-seed/init`) — all changes happen there, never on main
3. vibe-seed scans existing code, docs, configs, and folder structure
4. **Gap analysis** — identifies what's missing vs a fully-seeded project:
   - Missing docs (PRD, DDD, ARCHITECTURE)
   - Missing tooling (.claude/, .beads/, quality gates)
   - Missing structure (DDD layers, test mirrors)
   - Existing files that conflict with vibe-seed defaults
5. Shows the gap report and proposed changes (preview, like `vs init`)
6. Asks clarifying questions about existing domain (ubiquitous language, bounded contexts)
7. Generates missing artifacts, adapts agent profiles to existing domain language
8. User reviews the branch diff, merges when satisfied

**Branch safety rules:**
- `vs init --existing` MUST be run in a clean git working tree (no uncommitted changes)
- All changes go to `vibe-seed/init` branch, never directly to current branch
- If branch already exists, abort with message (user must clean up or use `--force-branch`)
- User merges manually — vibe-seed never merges for you
- **Existing tests MUST pass** — after scaffolding, `vs init --existing` runs the project's existing test suite. If any test fails, it rolls back all changes on the branch and reports what broke. Zero test regressions is a hard gate.

### Scenario 3: Tool Adaptation

**As an** AI Tool Switcher, **I want to** generate project configs for Claude Code, Cursor, or OpenCode, **so that** I use the same conventions regardless of which AI tool I open.

**Flow:**
1. Developer selects target tool(s) during bootstrap
2. vibe-seed generates tool-specific config files (`.claude/`, `.cursor/`, etc.)
3. Agent personas and commands adapt to each tool's format
4. Quality gates and beads workflow remain identical

## 5. Capabilities

### Must Have (P0)

- [ ] **CLI tool (`vs`)** — Primary user interface for all vibe-seed operations (`vs init`, `vs guide`, `vs generate`)
- [ ] **MCP server** — Expose guided bootstrap and knowledge base as MCP tools for AI tool integration
- [ ] **`.vibe-seed/` project directory** — Per-project state, knowledge base, and doc maintenance config (see section 5.1)
- [ ] **`vs init` with preview** — Show exactly what will be installed/copied, require user confirmation before any action
- [ ] **Global settings detection** — Detect tool global configs (`~/.claude/`, `~/.cursor/`, etc.), report conflicts with local settings, let user choose resolution per conflict
- [ ] **Existing project adoption (`vs init --existing`)** — Branch-based gap analysis and scaffolding for existing projects (see Scenario 2)
- [ ] **Gap analysis** — Scan existing project, compare against full vibe-seed structure, report what's missing/conflicting
- [ ] **Guided project bootstrap** — Conversational flow from README idea to full project structure
- [ ] **DDD question framework** — Structured questions for domain stories, ubiquitous language, bounded contexts, aggregate design
- [ ] **Artifact generation** — Generate PRD, DDD.md, ARCHITECTURE.md from guided answers
- [ ] **Agent personas** — Developer, researcher, tech-lead, PM, QA, security agents with DDD awareness
- [ ] **Beads integration** — Epic/spike/ticket templates enforcing DDD+TDD+SOLID
- [ ] **Quality gates** — ruff + mypy + pytest enforced before ticket closure
- [ ] **Knowledge base (RLM)** — Addressable docs for DDD patterns, coding tool conventions
- [ ] **Doc maintenance commands** — Slash commands for doc health, architecture lookup, knowledge refresh (like doc-health, architecture-docs, owasp-docs in Tachikoma)

### Should Have (P1)

- [ ] **Multi-tool support** — Generate configs for Claude Code, Cursor, Antigravity, OpenCode
- [ ] **Tool knowledge versioning** — Maintain current + 3 previous major versions per tool
- [ ] **Spike workflow** — Guided spike creation with clear output goals → ADR docs

### Nice to Have (P2)

- [ ] **Template library** — Domain-specific templates (web API, CLI tool, data pipeline, etc.)
- [ ] **Knowledge auto-update** — Fetch latest tool docs and update knowledge base

### 5.1 `.vibe-seed/` Directory (per-project)

Every project initialized with `vs init` gets a `.vibe-seed/` directory:

```
.vibe-seed/
├── config.toml              # Project-specific vibe-seed settings
├── knowledge/               # RLM-addressable knowledge base (copied from seed)
│   ├── ddd/                 # DDD patterns, tactical/strategic references
│   ├── tools/               # AI coding tool conventions (versioned)
│   │   ├── claude-code/     # .claude/ format, agents, commands
│   │   ├── cursor/          # .cursor/ format, rules
│   │   ├── antigravity/     # Antigravity config format
│   │   └── opencode/        # OpenCode config format
│   └── conventions/         # TDD, SOLID, quality gate references
└── maintenance/             # Doc health tracking, review schedules
    └── doc-registry.toml    # Which docs to track, owners, review intervals
```

### 5.2 `vs init` Behavior

**Safety-first approach:**

1. **Preview** — Show everything that will be created/installed (dry-run by default)
2. **Confirm** — User must explicitly agree before any file operations
3. **Never overwrite** — If a file already exists, skip it
4. **Conflict resolution** — If vibe-seed wants to create a file that exists, rename ours: `filename_vibe_seed.md`
5. **Tool installation** — Optionally install beads, trivy, shannon (show what + ask first)
6. **Global settings detection** — Scan for global configs that override local project settings (see 5.2.1)

#### 5.2.1 Global Settings Detection

AI coding tools have global configs that **always win** over local project settings:

| Tool | Global Location | Overrides |
|------|----------------|-----------|
| Claude Code | `~/.claude/CLAUDE.md`, `~/.claude/settings.json` | Project `.claude/CLAUDE.md` |
| Cursor | `~/.cursor/`, global rules | Project `.cursor/` rules |
| Antigravity | TBD (spike needed) | TBD |
| OpenCode | TBD (spike needed) | TBD |

`vs init` must:

1. **Detect** — Scan known global config paths for each detected tool
2. **Compare** — Check for conflicts between global settings and what vibe-seed wants to set locally
3. **Report** — Show conflicts clearly with what the global setting does vs what we want
4. **Ask** — Let user choose per conflict:
   - **Keep global** — skip the local setting (global wins anyway)
   - **Update global** — add/merge into the global config (user must confirm)
   - **Set local anyway** — create the local setting knowing global overrides it (with a warning comment in the file)

**Example with global conflict:**

```
$ vs init

Detecting tools...
  Found: Claude Code (global config at ~/.claude/)
  Found: Beads (already installed)

Global settings scan:
  ⚠ CONFLICT  ~/.claude/CLAUDE.md defines git rules that differ from vibe-seed defaults
              Global: "NEVER add Co-Authored-By lines"
              Local:  (vibe-seed would set the same — no conflict)
              → OK, compatible

  ⚠ CONFLICT  ~/.claude/settings.json has custom model preferences
              Global sets default model → sonnet
              Local:  vibe-seed has no model preference
              → OK, no conflict

  ⚠ CONFLICT  ~/.claude/CLAUDE.md has project-specific paths (src/tachikoma/)
              These reference another project and won't apply here
              → OK, global is scoped to other project

  ⚠ CONFLICT  ~/.claude/settings.json has allowedTools restrictions
              Global restricts: Edit, Write require approval
              Local:  vibe-seed agents expect Edit, Write available
              → WARNING: agents may hit permission prompts

              Options:
                [1] Keep global (agents will prompt for permissions)
                [2] Update global to allow Edit, Write
                [3] Note in local CLAUDE.md that agents need these tools

Project files:
  CREATE  .vibe-seed/config.toml
  CREATE  .vibe-seed/knowledge/ddd/...          (12 files)
  CREATE  .vibe-seed/knowledge/tools/...        (8 files)
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

- `vs doc-health` — Check freshness, broken references, missing metadata (like `/doc-health`)
- `vs doc-review` — Mark docs as reviewed, update `last_reviewed` dates
- `.vibe-seed/maintenance/doc-registry.toml` — Tracks which docs to monitor, owners, review cadence

This mirrors the pattern from Tachikoma's `/doc-health`, `/architecture-docs`, and `/owasp-docs` commands but is generalized and project-independent.

## 6. Constraints

### Technical Constraints

| Constraint | Value | Rationale |
|-----------|-------|-----------|
| Language | Python 3.12+ | Target audience and our own stack |
| Package manager | uv | Speed, reproducibility, modern standard |
| CLI name | `vs` (vibe-seed) | Short, memorable, unix-style |
| Issue tracking | Beads | Git-native, works offline, no external service |
| Interfaces | CLI (`vs`) + MCP server | CLI for humans, MCP for AI tools |
| AI interaction | Conversational | Must work in terminal (Claude Code) and IDE (Cursor) |

### Non-Functional Requirements

| Requirement | Target | Measurement |
|------------|--------|-------------|
| Bootstrap time | < 30 minutes | From README to first beads ticket |
| Knowledge freshness | < 90 days stale | Per-doc last_reviewed dates |
| Tool coverage | 4 tools | Claude Code, Cursor, Antigravity, OpenCode |

### File Safety Rules (HARD CONSTRAINTS)

| Rule | Behavior |
|------|----------|
| Never overwrite | If target file exists, skip it |
| Conflict rename | If vibe-seed needs to create a conflicting file, suffix ours: `filename_vibe_seed.md` |
| Preview first | All file operations shown before execution (`vs init` dry-run) |
| Explicit confirm | User must agree before any write/copy/install |
| No silent installs | Tool installation (beads, trivy, shannon) is optional and shown separately |
| Branch for existing | `vs init --existing` always creates a new branch, never writes to current branch |
| Clean tree required | `vs init --existing` refuses to run with uncommitted changes |
| Never merge | vibe-seed never merges branches — user merges manually after review |
| **Zero test regression** | `vs init --existing` runs existing test suite after scaffolding — if ANY test fails, roll back all changes and report. This is a **hard gate**, no exceptions. |

### Budget / Resource Constraints

- No cloud dependencies — everything runs locally
- No paid APIs required for core functionality
- Knowledge base maintained manually (RLM pattern, not auto-scraping)

## 7. Out of Scope

- Language support beyond Python (future consideration)
- Package manager support beyond uv (future consideration)
- Automated code generation from DDD artifacts (we generate structure, not business logic)
- IDE plugin development (we generate config files, not plugins)
- Hosting or deployment automation

## 8. Success Metrics

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Time from idea to first ticket | < 30 min | Manual timing |
| Projects using correct lifecycle | 100% | All have PRD + DDD + ARCH before code |
| Knowledge base coverage | 4 tools | Claude Code, Cursor, Antigravity, OpenCode docs present |
| Quality gate enforcement | 100% | No ticket closed without passing gates |

## 9. Risks & Unknowns

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Tool config formats change frequently | High | Medium | Version knowledge base (current + 3 prev) |
| DDD questions too abstract for beginners | Medium | High | Provide concrete examples per question |
| Guided flow feels too rigid | Medium | Medium | Allow skipping with explicit acknowledgment |
| MCP server adds complexity | Low | Medium | Spike first, implement only if justified |

### Open Questions (need spikes)

- [x] ~~**Spike: MCP vs CLI** — Decided: both. CLI (`vs`) for humans, MCP server for AI tools.~~
- [ ] **Spike: CLI + MCP design** — Command tree for `vs`, MCP tool schemas, shared application core
- [ ] **Spike: Knowledge base structure** — How to organize `.vibe-seed/knowledge/` with RLM addressability and version tracking?
- [ ] **Spike: Multi-tool config generation** — What are the config formats for Cursor, Antigravity, OpenCode? How similar/different?
- [ ] **Spike: Guided question framework** — What's the minimal effective set of DDD questions to go from idea to bounded contexts?

## 10. Timeline

| Phase | Deliverable |
|-------|-------------|
| Phase 1: Foundation | PRD + DDD + Architecture for vibe-seed itself |
| Phase 2: Core CLI | `vs init` + `vs guide` + artifact generation |
| Phase 3: Knowledge Base | `.vibe-seed/knowledge/` with RLM docs for DDD + tool conventions |
| Phase 4: MCP Server | MCP tools exposing same core as CLI |
| Phase 5: Multi-Tool | Config generation for Claude Code, Cursor, Antigravity, OpenCode |
| Phase 6: Doc Maintenance | `vs doc-health`, `vs doc-review`, maintenance registry |
