---
last_reviewed: 2026-02-22
owner: architecture
status: complete
type: spike
ticket: vibe-seed-k7m.1
---

# Knowledge Base Structure for Tool Conventions

> **Spike:** k7m.1 -- Knowledge base structure for tool conventions
> **Timebox:** 4 hours
> **Research question:** How should we organize the knowledge base for AI coding tool
> conventions so docs are RLM-addressable, version-tracked (current + 3 previous major
> versions), and maintainable -- including a foundation for drift detection?

## 1. Per-Tool Config Documentation

### 1.1 Claude Code

**Current version series:** 2.1.x (Feb 2026)
**License:** Proprietary (Anthropic)
**Config format:** Markdown (CLAUDE.md, agents, rules, skills, commands), JSON (settings)
**Changelog URL:** https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md

#### Project-level directory (`.claude/`)

```
.claude/
  CLAUDE.md                    # Project memory/instructions (committed)
  CLAUDE.local.md              # Personal memory (gitignored)
  settings.json                # Team-shared settings (committed)
  settings.local.json          # Personal settings (gitignored)
  .mcp.json                    # Project-scoped MCP server config
  agents/                      # Custom subagent definitions (*.md)
    developer.md
    researcher.md
  agent-memory/                # Persistent subagent memory (project scope)
    <agent-name>/
      MEMORY.md
  agent-memory-local/          # Persistent subagent memory (local scope, gitignored)
  rules/                       # Additional instruction files (*.md, auto-loaded)
  commands/                    # Project slash commands (*.md)
  skills/                      # Project skills (SKILL.md in dirs)
```

#### Global-level directory (`~/.claude/`)

```
~/.claude/
  CLAUDE.md                    # Global user instructions
  settings.json                # Global user settings
  settings.local.json          # Personal overrides (gitignored)
  .claude.json                 # OAuth, MCP servers, preferences
  agents/                      # Global subagent definitions
  agent-memory/                # Global persistent subagent memory
  commands/                    # Global slash commands
  skills/                      # Global skills
  projects/                    # Session history per project
  statsig/                     # Analytics cache
```

#### System-managed config paths

| OS | Path | Purpose |
|----|------|---------|
| macOS | `/Library/Application Support/ClaudeCode/managed-settings.json` | Admin-managed (highest precedence) |
| Linux/WSL | `/etc/claude-code/managed-settings.json` | Admin-managed |
| Windows | `C:\Program Files\ClaudeCode\managed-settings.json` | Admin-managed |

#### Config precedence (highest to lowest)

1. Managed settings (`managed-settings.json`) -- system admin, cannot be overridden
2. Command line arguments -- session-only
3. `.claude/settings.local.json` -- project personal
4. `.claude/settings.json` -- project team
5. `~/.claude/settings.json` -- user global

#### Agent definition format (Markdown + YAML frontmatter)

```yaml
---
name: developer
description: TDD developer for DDD-structured Python projects
tools: Read, Edit, Write, Bash, Grep, Glob
model: inherit
permissionMode: default
skills:
  - api-conventions
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "./scripts/validate.sh"
memory: project
---

You are a developer agent. Follow TDD: RED -> GREEN -> REFACTOR.
```

**Key fields:** name, description, tools, disallowedTools, model (sonnet/opus/haiku/inherit),
permissionMode (default/acceptEdits/dontAsk/bypassPermissions/plan), maxTurns, skills,
mcpServers, hooks, memory (user/project/local), background, isolation.

#### Settings schema

JSON with schema at `https://json.schemastore.org/claude-code-settings.json`.
Key sections: permissions (allow/ask/deny), model, env, outputStyle, sandbox, hooks.

**Sources:**
- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/settings
- https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md

---

### 1.2 Cursor

**Current version series:** 2.5.x (Feb 2026)
**License:** Proprietary (Anysphere)
**Config format:** MDC (Modular Markdown with YAML frontmatter), AGENTS.md (plain Markdown)
**Changelog URL:** https://cursor-changelog.com/

#### Project-level directory (`.cursor/`)

```
.cursor/
  rules/                       # Rule files (*.mdc)
    00-project-context.mdc
    01-architecture.mdc
    99-agent-behavior.mdc
```

Plus project root files:
- `.cursorrules` (legacy, deprecated -- still supported)
- `AGENTS.md` (standard, supported since 2026)

#### Global-level configuration

Cursor stores global settings in a **SQLite database**, not JSON files:

| OS | Path | Purpose |
|----|------|---------|
| macOS | `~/Library/Application Support/Cursor/User/globalStorage/state.vscdb` | All global settings |
| Linux | `~/.config/Cursor/User/globalStorage/state.vscdb` | All global settings |
| Windows | `%APPDATA%\Cursor\User\globalStorage\state.vscdb` | All global settings |

User Rules (global) are stored in SQLite under key `aicontext.personalContext`.
Team Rules are set in the Cursor team dashboard (cloud-based).

#### Config precedence (highest to lowest)

1. Team Rules (cloud dashboard)
2. Project Rules (`.cursor/rules/*.mdc`)
3. AGENTS.md (project root)
4. `.cursorrules` (legacy fallback)
5. User Rules (global, in SQLite)

#### Rule file format (.mdc)

```yaml
---
description: Architecture enforcement rules for DDD layers
globs: "src/**/*.py"
alwaysApply: false
---

# Architecture Rules

- Domain layer has ZERO external dependencies
- Dependencies flow inward: infrastructure -> application -> domain
```

**Key fields:** description, globs (file pattern for activation), alwaysApply (bool).
When `alwaysApply: false`, the AI decides whether to activate based on description.

#### Notable limitations for vibe-seed

- No subagent/agent concept (uses rules, not personas)
- Global config is SQLite, not file-based -- vibe-seed cannot generate global config
- AGENTS.md support provides a cross-tool bridge

**Sources:**
- https://design.dev/guides/cursor-rules/
- https://www.jackyoustra.com/blog/cursor-settings-location
- https://github.com/sanjeed5/awesome-cursor-rules-mdc

---

### 1.3 Roo Code (formerly Cline / Antigravity)

**Current version series:** 3.38.x (Jan 2026)
**License:** Apache-2.0 (open source)
**Config format:** YAML/JSON (.roomodes), Markdown (.roo/rules/)
**Changelog URL:** https://github.com/RooCodeInc/Roo-Code/blob/main/CHANGELOG.md
**Note:** "Antigravity" was the original project name; now primarily known as "Roo Code".

#### Project-level directory (`.roo/`)

```
.roo/
  rules/                       # Shared rules across all modes (*.md, *.txt)
    01-architecture.md
    02-coding-standards.md
  rules-code/                  # Mode-specific rules (rules-{modeSlug}/)
  rules-architect/
  rules-debug/
  rules-docs-writer/
  modes/                       # Custom mode definitions (*.yaml)
    custom-mode.yaml
  skills/                      # Project-level skills
    skill-name/
```

Plus project root files:
- `.roomodes` (YAML or JSON -- custom mode definitions)
- `.roorules` (global rules -- fallback if .roo/rules/ absent)
- `.roorules-{modeSlug}` (per-mode rules -- fallback if .roo/rules-{slug}/ absent)
- `AGENTS.md` (supported as cross-tool standard)

#### Global-level directory (`~/.roo/`)

```
~/.roo/
  rules/                       # Global shared rules
  rules-{modeSlug}/            # Global mode-specific rules
  skills/                      # Global skills
    skill-name/
```

Global custom modes: Stored in VS Code global state (not file-based).

#### Config precedence

1. Project `.roo/rules-{slug}/` (directory takes precedence)
2. Project `.roorules-{slug}` (file fallback)
3. Project `.roomodes` mode definitions
4. Global `~/.roo/rules-{slug}/`
5. Global custom modes (VS Code state)
6. Built-in modes (code, architect, debug, ask)

#### Mode definition format (.roomodes -- YAML)

```yaml
customModes:
  - slug: ddd-developer
    name: DDD Developer
    description: TDD developer for DDD-structured Python projects
    roleDefinition: |
      You are a developer agent specializing in Domain-Driven Design.
      Follow TDD: RED -> GREEN -> REFACTOR.
    whenToUse: Use for all implementation tasks in the project.
    customInstructions: |
      Use ubiquitous language from docs/DDD.md.
    groups:
      - read
      - edit
      - command
      - mcp
```

**Key fields:** slug, name, description, roleDefinition, whenToUse, customInstructions,
groups (tool groups: read/edit/command/mcp), file restrictions (glob patterns).

**Sources:**
- https://docs.roocode.com/features/custom-modes
- https://docs.roocode.com/features/custom-instructions
- https://github.com/RooCodeInc/Roo-Code

---

### 1.4 OpenCode

**Current version:** Actively maintained (sst/opencode on GitHub)
**License:** MIT
**Config format:** JSON/JSONC (opencode.json), Markdown (agents, modes, rules)
**Changelog URL:** https://opencode.ai/changelog
**Note:** Written in Go; terminal-based TUI. Original opencode-ai/opencode repo archived Sep 2025;
active development at sst/opencode.

#### Project-level directory (`.opencode/`)

```
.opencode/
  agents/                      # Custom agent definitions (*.md)
    review.md
    developer.md
  modes/                       # Custom mode definitions (*.md)
    build.md
    plan.md
  commands/                    # Custom commands
  rules/                       # Project rules (*.md)
  plugins/                     # Plugin loading
  skills/                      # Skills
  tools/                       # Custom tools
  themes/                      # Visual themes
```

Plus project root files:
- `opencode.json` or `opencode.jsonc` (main config)
- `AGENTS.md` (project instructions)
- `CLAUDE.md` (fallback if no AGENTS.md -- backward compatible)

#### Global-level directory (`~/.config/opencode/`)

```
~/.config/opencode/
  opencode.json                # Global config
  AGENTS.md                    # Global rules
  agents/                      # Global agent definitions
  modes/                       # Global mode definitions
```

Fallback: `~/.claude/CLAUDE.md` (backward compatible with Claude Code).

#### Config precedence (6 layers, later overrides earlier)

1. Remote config (`.well-known/opencode`)
2. Global config (`~/.config/opencode/opencode.json`)
3. Custom config (`OPENCODE_CONFIG` env var)
4. Project config (`opencode.json` in root)
5. `.opencode` directories
6. Inline config (`OPENCODE_CONFIG_CONTENT` env var)

#### Agent definition format (Markdown + YAML frontmatter)

```yaml
---
description: Reviews code for quality and best practices
mode: subagent
model: anthropic/claude-sonnet-4-20250514
temperature: 0.1
tools:
  write: false
  edit: false
---

You are in code review mode. Focus on code quality, security, and best practices.
```

**Key fields:** description, mode (subagent), model, temperature, tools (write/edit/bash/read/etc as bools).

#### Mode definition format (Markdown + YAML frontmatter)

```yaml
---
model: anthropic/claude-sonnet-4-20250514
temperature: 0.3
tools:
  bash: true
  edit: true
  write: true
---

Custom build mode instructions here.
```

#### opencode.json schema

Schema at `https://opencode.ai/config.json`. Key sections: provider, model, small_model,
tui, server, tools, agent, default_agent, share, command, keybinds, formatter, permission,
compaction, watcher, mcp, plugin, instructions.

The `instructions` field supports glob patterns and URLs:
```json
{
  "instructions": ["CONTRIBUTING.md", "docs/guidelines.md"]
}
```

**Sources:**
- https://opencode.ai/docs/config/
- https://opencode.ai/docs/agents/
- https://opencode.ai/docs/modes/
- https://opencode.ai/docs/rules/

---

## 2. Global Config Path Registry

### Per-tool, per-OS global config paths

| Tool | OS | Global Path | What It Controls | Override Behavior |
|------|----|-------------|-----------------|-------------------|
| Claude Code | macOS/Linux | `~/.claude/` | Settings, agents, commands, skills, CLAUDE.md | Project `.claude/` overrides global for same keys |
| Claude Code | macOS | `/Library/Application Support/ClaudeCode/managed-settings.json` | Admin lockdown | Highest precedence, cannot be overridden |
| Claude Code | Linux | `/etc/claude-code/managed-settings.json` | Admin lockdown | Highest precedence |
| Claude Code | Windows | `%USERPROFILE%\.claude\` + `C:\Program Files\ClaudeCode\managed-settings.json` | Same as above | Same |
| Cursor | macOS | `~/Library/Application Support/Cursor/User/globalStorage/state.vscdb` | User rules, model prefs | SQLite DB; project `.cursor/rules/` overrides |
| Cursor | Linux | `~/.config/Cursor/User/globalStorage/state.vscdb` | Same | Same |
| Cursor | Windows | `%APPDATA%\Cursor\User\globalStorage\state.vscdb` | Same | Same |
| Roo Code | macOS/Linux | `~/.roo/` | Global rules, skills | Project `.roo/` overrides; VS Code state for modes |
| Roo Code | Windows | `~/.roo/` (in user home) | Same | Same |
| OpenCode | macOS/Linux | `~/.config/opencode/` | Config, agents, modes, AGENTS.md | Project `opencode.json` / `.opencode/` overrides |
| OpenCode | Windows | `%APPDATA%/opencode/` (assumed, follows XDG) | Same | Same |

### Implications for `vs detect`

1. Claude Code and Roo Code have **file-based** global configs that vibe-seed can detect and compare.
2. Cursor has **SQLite-based** global config -- vibe-seed can detect the DB file exists but
   cannot easily read/compare settings without SQLite queries. Best approach: detect presence only,
   warn user to check manually.
3. OpenCode has **file-based** global configs, similar to Claude Code.
4. All four tools support `AGENTS.md` as a cross-tool project instruction file.

---

## 3. Commonalities and Differences

### Feature Comparison

| Feature | Claude Code | Cursor | Roo Code | OpenCode |
|---------|------------|--------|----------|----------|
| **Project config dir** | `.claude/` | `.cursor/` | `.roo/` | `.opencode/` |
| **Global config dir** | `~/.claude/` | SQLite DB | `~/.roo/` | `~/.config/opencode/` |
| **Agent/persona support** | Yes (subagents, `.md`) | No (rules only) | Yes (modes, `.yaml`) | Yes (agents, `.md`) |
| **Agent definition format** | Markdown + YAML FM | N/A | YAML (.roomodes) or Markdown | Markdown + YAML FM or JSON |
| **Rule/instruction format** | Markdown (CLAUDE.md, rules/) | MDC (.mdc) + AGENTS.md | Markdown (rules/) | Markdown (AGENTS.md, rules/) |
| **Settings format** | JSON | SQLite | VS Code settings + YAML | JSON/JSONC |
| **MCP config** | JSON (`.mcp.json`) | JSON (VS Code settings) | VS Code settings | JSON (opencode.json) |
| **AGENTS.md support** | No (uses CLAUDE.md) | Yes | Yes | Yes (primary) |
| **CLAUDE.md support** | Yes (native) | No | No | Yes (fallback) |
| **Version-controlled config** | settings.json, agents, rules | .cursor/rules/*.mdc | .roomodes, .roo/ | opencode.json, .opencode/ |
| **Gitignored config** | settings.local.json, CLAUDE.local.md | N/A | N/A | N/A |
| **Slash commands** | Yes (.claude/commands/) | Yes (built-in + custom) | Yes (built-in) | Yes (.opencode/commands/) |
| **Skills** | Yes (.claude/skills/) | No | Yes (.roo/skills/) | Yes (.opencode/skills/) |
| **Hooks/lifecycle** | Yes (PreToolUse, PostToolUse, etc.) | Yes (beta, hooks) | No (modes only) | No |

### Cross-Tool Bridge: AGENTS.md

AGENTS.md is an emerging standard (stewarded by Agentic AI Foundation under Linux Foundation).
Supported by: Cursor, Roo Code, OpenCode (native), Claude Code (not natively -- uses CLAUDE.md).

**Implication for vibe-seed:** Generate both `AGENTS.md` and tool-specific configs. AGENTS.md
serves as the common denominator. Tool-specific configs add persona/agent definitions that
AGENTS.md cannot express.

### Concept Mapping Across Tools

| Concept | Claude Code | Cursor | Roo Code | OpenCode |
|---------|------------|--------|----------|----------|
| Persona/Agent | Subagent (`.claude/agents/*.md`) | Rule file (`.cursor/rules/*.mdc`) | Mode (`.roomodes` + `.roo/rules-{slug}/`) | Agent (`.opencode/agents/*.md`) |
| Global instructions | `~/.claude/CLAUDE.md` | User Rules (SQLite) | `~/.roo/rules/` | `~/.config/opencode/AGENTS.md` |
| Project instructions | `.claude/CLAUDE.md` + `.claude/rules/` | `.cursor/rules/*.mdc` | `.roo/rules/` + `.roorules` | `AGENTS.md` + `.opencode/rules/` |
| Tool permissions | `settings.json` permissions | Built-in per mode | Mode groups | `opencode.json` tools |
| Model selection | `settings.json` model | UI setting | Per-mode | `opencode.json` model |

---

## 4. Proposed `.vibe-seed/knowledge/` Directory Structure

### Design Principles

1. **RLM-addressable** -- every knowledge entry has a deterministic path: `tool/version/topic`
2. **Version-tracked** -- current + 3 previous major versions per tool
3. **Drift-detection-ready** -- each entry has metadata for staleness checking
4. **Tool-translation-ready** -- structured data (TOML) for machine consumption, not just docs

### Directory Layout

```
.vibe-seed/knowledge/
  _index.toml                          # Master index for RLM O(1) lookup
  tools/
    claude-code/
      _meta.toml                       # Tool metadata (name, URL, license, versions tracked)
      current/                         # Alias -> latest tracked version (e.g., 2.1)
        config-structure.toml          # File tree, formats, paths
        agent-format.toml              # Agent/subagent definition schema
        settings-format.toml           # settings.json schema reference
        rules-format.toml              # Rules and CLAUDE.md conventions
        commands-format.toml           # Slash command format
        mcp-config.toml                # MCP server configuration format
        global-paths.toml              # Global config paths per OS
        gitignore-patterns.toml        # What to .gitignore
      v2.1/                            # Explicit version (current alias target)
        (same files as current/)
      v2.0/                            # Previous major version
        (same files)
      v1.0/                            # Older major version
        (same files)
    cursor/
      _meta.toml
      current/                         # Alias -> e.g., 2.5
        config-structure.toml
        rules-format.toml              # .mdc format details
        agents-md-support.toml         # AGENTS.md support details
        global-paths.toml
        gitignore-patterns.toml
      v2.5/
      v2.4/
      v2.0/
    roo-code/
      _meta.toml
      current/                         # Alias -> e.g., 3.38
        config-structure.toml
        mode-format.toml               # .roomodes schema
        rules-format.toml              # .roo/rules/ conventions
        global-paths.toml
        gitignore-patterns.toml
      v3.38/
      v3.22/
      v2.2/
    opencode/
      _meta.toml
      current/                         # Alias -> latest
        config-structure.toml
        agent-format.toml
        mode-format.toml
        rules-format.toml
        opencode-json-schema.toml      # opencode.json reference
        global-paths.toml
        gitignore-patterns.toml
      (version dirs)
  cross-tool/
    agents-md.toml                     # AGENTS.md cross-tool standard
    concept-mapping.toml               # How concepts map across tools (agent = mode = subagent)
    generation-matrix.toml             # What vibe-seed generates per tool
  ddd/
    tactical-patterns.md               # Entities, VOs, Aggregates, etc.
    strategic-patterns.md              # Bounded Contexts, Context Maps
    event-storming.md                  # Event Storming reference
    domain-storytelling.md             # Domain Storytelling reference
  conventions/
    tdd.md                             # TDD reference (RED/GREEN/REFACTOR)
    solid.md                           # SOLID principles reference
    quality-gates.md                   # ruff + mypy + pytest conventions
```

### Why TOML for Tool Knowledge (not Markdown)

Tool convention entries are **structured data** consumed by `KnowledgeLookupPort` and
`ConfigGenerationPort`. They need:
- Machine-parseable fields (file paths, formats, version ranges)
- Deterministic keys for O(1) lookup
- Easy diffing for drift detection

Markdown is used for DDD and convention reference material (human consumption).
TOML is used for tool conventions (machine + human consumption).

### `current/` Alias Strategy

The `current/` directory is a **symlink or copy** of the latest tracked version.
- On `vs init`, files are copied from `current/` into the project.
- `KnowledgeLookupPort` resolves `tool=claude-code, version=current` to the actual version.
- When a new version is tracked, `current/` is updated and drift detection compares old vs new.

---

## 5. RLM Lookup Table Design

### Addressing Scheme

Every knowledge entry is addressable by a triple:

```
(category, tool_or_topic, subtopic, version?)
```

Mapped to the URI scheme from k7m.4:

```
vibeseed://knowledge/tools/{tool}/{subtopic}?version={version}
vibeseed://knowledge/ddd/{topic}
vibeseed://knowledge/conventions/{topic}
vibeseed://knowledge/cross-tool/{topic}
```

### KnowledgeLookupPort Resolution

```python
class KnowledgeLookupPort(Protocol):
    def lookup(
        self,
        category: str,          # "tools", "ddd", "conventions", "cross-tool"
        topic: str,             # "claude-code/agent-format", "ddd/aggregates"
        version: str = "current",  # "current", "v2.1", "v2.0"
    ) -> KnowledgeEntry: ...

    def list_tools(self) -> list[ToolMeta]: ...
    def list_versions(self, tool: str) -> list[str]: ...
    def list_topics(self, category: str, tool: str | None = None) -> list[str]: ...
```

### File-Based Resolution (O(1))

```python
def _resolve_path(self, category: str, topic: str, version: str) -> Path:
    """Deterministic path resolution -- no search needed."""
    base = self.knowledge_dir / category
    if category == "tools":
        # topic = "claude-code/agent-format"
        tool, subtopic = topic.split("/", 1)
        return base / tool / version / f"{subtopic}.toml"
    else:
        # topic = "aggregates"
        return base / f"{topic}.md"
```

This is O(1) -- a direct path construction, no glob, no search, no index scan needed.

### CLI Usage

```bash
vs kb tools/claude-code/agent-format              # Current version
vs kb tools/claude-code/agent-format --version v2.0  # Specific version
vs kb ddd/aggregates                               # DDD reference
vs kb cross-tool/concept-mapping                   # Cross-tool mappings
```

### MCP Resource Resolution

```python
@mcp.resource("vibeseed://knowledge/tools/{tool}/{subtopic}")
def get_tool_knowledge(tool: str, subtopic: str) -> str:
    entry = knowledge_port.lookup("tools", f"{tool}/{subtopic}")
    return entry.to_json()
```

---

## 6. Versioning Scheme

### Version Granularity

Track **major version series**, not every patch release. The config format changes
at major version boundaries, not between 2.1.3 and 2.1.4.

| Tool | Current | Prev 1 | Prev 2 | Prev 3 | Granularity |
|------|---------|--------|--------|--------|-------------|
| Claude Code | 2.1.x | 2.0.x | 1.x | -- | Major.Minor series |
| Cursor | 2.5.x | 2.4.x | 2.0.x | 1.7.x | Major.Minor series |
| Roo Code | 3.38.x | 3.22.x | 2.2.x | -- | Major series (3.x config stable) |
| OpenCode | latest | -- | -- | -- | Early stage, fewer versions to track |

### Version Metadata (`_meta.toml`)

```toml
[tool]
name = "claude-code"
display_name = "Claude Code"
vendor = "Anthropic"
license = "Proprietary"
homepage = "https://code.claude.com"
changelog_url = "https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md"
schema_url = "https://json.schemastore.org/claude-code-settings.json"

[versions]
current = "v2.1"
tracked = ["v2.1", "v2.0", "v1.0"]
deprecated = []

[versions.v2_1]
version_range = ">=2.1.0"
release_date = "2026-01"
end_of_support = ""  # Still current
breaking_changes = "Added .claude/rules/ directory, skills, subagent memory"
last_verified = "2026-02-22"

[versions.v2_0]
version_range = ">=2.0.0,<2.1.0"
release_date = "2025-11"
end_of_support = ""
breaking_changes = "Subagents, agent teams, .claude/agents/"
last_verified = "2026-02-22"
```

### Adding a New Version

1. Copy `current/` to `v{new}/` (snapshot the outgoing current)
2. Update files in `current/` with new version conventions
3. Update `_meta.toml` versions section
4. Run drift detection to flag what changed

### Deprecation Flow

When a tool version falls out of the "current + 3" window:
1. Mark as `deprecated` in `_meta.toml`
2. Keep files for 1 additional cycle (total 5 versions on disk)
3. Remove after grace period
4. `KnowledgeLookupPort.lookup()` for a deprecated version returns the entry plus a warning

---

## 7. Drift Detection Schema

### Per-Entry Metadata

Every `.toml` knowledge entry has a `[_meta]` section:

```toml
[_meta]
last_verified = "2026-02-22"
verified_against = "v2.1.15"  # Specific version tested
changelog_url = "https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md"
source_urls = [
    "https://code.claude.com/docs/en/settings",
    "https://code.claude.com/docs/en/sub-agents",
]
confidence = "high"  # high | medium | low
deprecated = false
deprecated_since = ""
superseded_by = ""  # Path to replacement entry if deprecated
next_review_date = "2026-05-22"  # 90-day freshness window (from PRD NFR)

# Schema version for this entry format -- enables migration
schema_version = 1
```

### Drift Detection Signals

The `DriftReport` value object (from DDD.md) uses these fields to detect staleness:

| Signal | Source | Detection Method |
|--------|--------|-----------------|
| **Time-based staleness** | `last_verified` + `next_review_date` | `now > next_review_date` |
| **Version mismatch** | `verified_against` vs installed tool version | Compare with `vs detect` output |
| **Changelog delta** | `changelog_url` | Future: fetch and diff (P2 auto-update) |
| **Confidence decay** | `confidence` | `low` entries flagged sooner |
| **Deprecation** | `deprecated` flag | Immediate alert if tool uses deprecated convention |
| **Schema migration** | `schema_version` | Flag entries with outdated schema format |

### `vs doc-health` Integration

```bash
vs doc-health --knowledge

Knowledge Base Health Report:
  claude-code/current:
    agent-format.toml       OK    verified 2026-02-22, next review 2026-05-22
    settings-format.toml    WARN  verified 2025-12-01, overdue by 53 days
    commands-format.toml    OK    verified 2026-01-15, next review 2026-04-15

  cursor/current:
    rules-format.toml       WARN  verified against v2.4, installed v2.5 detected
    agents-md-support.toml  OK    verified 2026-02-10

  Summary: 2 entries need review, 0 deprecated, 12 current
```

### Drift Detection Data Flow

```
vs detect (installed tools + versions)
    |
    v
KnowledgeLookupPort.list_tools() -> tracked tools + verified_against
    |
    v
DriftDetector.compare(installed_versions, tracked_versions)
    |
    v
DriftReport (list of stale/mismatched/deprecated entries)
    |
    v
vs doc-health --knowledge (human-readable report)
```

---

## 8. Example Knowledge Entry

### `tools/claude-code/current/agent-format.toml`

```toml
[_meta]
last_verified = "2026-02-22"
verified_against = "v2.1.15"
changelog_url = "https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md"
source_urls = [
    "https://code.claude.com/docs/en/sub-agents",
]
confidence = "high"
deprecated = false
next_review_date = "2026-05-22"
schema_version = 1

[format]
file_extension = ".md"
uses_frontmatter = true
frontmatter_format = "yaml"
location_project = ".claude/agents/"
location_global = "~/.claude/agents/"

[required_fields]
name = { type = "string", pattern = "^[a-z0-9-]+$", description = "Unique identifier" }
description = { type = "string", description = "When to delegate to this agent" }

[optional_fields]
tools = { type = "list[string]", description = "Allowed tools (Read, Edit, Write, Bash, Grep, Glob, etc.)" }
disallowedTools = { type = "list[string]", description = "Denied tools" }
model = { type = "enum", values = ["sonnet", "opus", "haiku", "inherit"], default = "inherit" }
permissionMode = { type = "enum", values = ["default", "acceptEdits", "dontAsk", "bypassPermissions", "plan"], default = "default" }
maxTurns = { type = "integer", description = "Max agentic turns" }
skills = { type = "list[string]", description = "Skills to preload" }
mcpServers = { type = "list[string|object]", description = "MCP servers available" }
hooks = { type = "object", description = "Lifecycle hooks (PreToolUse, PostToolUse, Stop)" }
memory = { type = "enum", values = ["user", "project", "local"], description = "Persistent memory scope" }
background = { type = "boolean", default = false }
isolation = { type = "enum", values = ["worktree"], description = "Run in git worktree" }

[body]
description = "Markdown content after frontmatter becomes the system prompt"
format = "markdown"

[example]
content = """
---
name: developer
description: TDD developer for DDD-structured Python projects. Use proactively for implementation tasks.
tools: Read, Edit, Write, Bash, Grep, Glob
model: inherit
memory: project
---

You are a developer agent. Follow TDD: RED -> GREEN -> REFACTOR.
Follow DDD layer rules. Domain has ZERO external dependencies.
"""
```

---

## 9. Cross-Tool Generation Matrix

### `cross-tool/generation-matrix.toml`

```toml
[_meta]
last_verified = "2026-02-22"
schema_version = 1

# What vibe-seed generates per tool
[claude_code]
project_instructions = ".claude/CLAUDE.md"
agents = ".claude/agents/{persona}.md"
settings = ".claude/settings.json"
rules = ".claude/rules/"
commands = ".claude/commands/"
mcp_config = ".mcp.json"
gitignore_entries = [".claude/settings.local.json", ".claude/CLAUDE.local.md", ".claude/agent-memory-local/"]

[cursor]
project_instructions = "AGENTS.md"
rules = ".cursor/rules/{topic}.mdc"
gitignore_entries = []
# Note: No agent/persona generation -- personas encoded as rule files

[roo_code]
project_instructions = "AGENTS.md"
modes = ".roomodes"
mode_rules = ".roo/rules-{mode-slug}/"
shared_rules = ".roo/rules/"
gitignore_entries = []

[opencode]
project_instructions = "AGENTS.md"
agents = ".opencode/agents/{persona}.md"
modes = ".opencode/modes/{mode}.md"
config = "opencode.json"
rules = ".opencode/rules/"
gitignore_entries = []
```

---

## 10. Recommendation

### Structure Decision

Use the **TOML-based versioned directory structure** described in Section 4 with:

1. **TOML for tool conventions** -- machine-parseable, supports O(1) RLM lookup,
   enables drift detection via structured metadata fields
2. **Markdown for DDD/convention reference** -- human-readable reference material
3. **`current/` alias pattern** -- simple version resolution without database
4. **`_meta.toml` per tool** -- centralized version tracking and changelog URLs
5. **`[_meta]` per entry** -- drift detection metadata (last_verified, verified_against,
   confidence, next_review_date)

### Rationale

- **RLM O(1) lookup**: Path construction from `(category, tool, topic, version)` is deterministic.
  No search, no index scan. Aligns with KnowledgeLookupPort from k7m.4.
- **Drift detection ready**: Every entry carries verification metadata. `vs doc-health --knowledge`
  can report staleness without any external service.
- **Version tracking**: `current/` + versioned dirs supports the PRD requirement of
  current + 3 previous major versions per tool.
- **Tool translation support**: TOML entries contain structured format schemas that
  `ConfigGenerationPort` can read to generate correct output per tool per version.
- **Cross-tool bridge**: AGENTS.md as the common denominator file generated for all tools;
  tool-specific configs generated additionally for tools that support richer features.

### Key Risk

**Maintenance burden.** 4 tools x ~7 topics x 4 versions = ~112 TOML files to maintain.
Mitigation: Start with `current/` only for all tools. Add historical versions only when
a breaking change occurs. Most version dirs will be identical or have minimal diffs.

### Alternative Considered: Single JSON/TOML Index

A single large index file with all tool conventions was considered but rejected because:
- Harder to version individual entries independently
- Harder to diff (one monolithic file vs per-topic changes)
- Breaks the RLM principle of addressable, independent documents

---

## 11. Follow-Up Tickets

| # | Title | Type | Bounded Context | Priority | Description |
|---|-------|------|----------------|----------|-------------|
| 1 | Implement `.vibe-seed/knowledge/` directory scaffolding | Task | Knowledge Base | P0 | Create the directory structure, `_index.toml`, and `_meta.toml` files for all 4 tools. Initial content for `current/` version of each tool. |
| 2 | Implement `KnowledgeLookupPort` and `FileKnowledgeService` | Task | Knowledge Base | P0 | Port protocol + file-based implementation. O(1) path resolution. TOML parsing. |
| 3 | Populate Claude Code knowledge entries | Task | Knowledge Base | P0 | All 7 topics for claude-code `current/` version based on this research. |
| 4 | Populate Cursor knowledge entries | Task | Knowledge Base | P0 | All topics for cursor `current/` version. |
| 5 | Populate Roo Code knowledge entries | Task | Knowledge Base | P0 | All topics for roo-code `current/` version. |
| 6 | Populate OpenCode knowledge entries | Task | Knowledge Base | P0 | All topics for opencode `current/` version. |
| 7 | Populate cross-tool knowledge entries | Task | Knowledge Base | P0 | concept-mapping.toml, generation-matrix.toml, agents-md.toml. |
| 8 | Implement `vs kb` CLI command | Task | CLI Framework | P0 | Wire `vs kb <topic>` to KnowledgeLookupPort. Rich output for terminal. |
| 9 | Implement drift detection metadata validation | Task | Knowledge Base | P1 | `vs doc-health --knowledge` reads `[_meta]` sections, compares with `vs detect` output. |
| 10 | Add version history tracking for tools | Task | Knowledge Base | P1 | Add `v2.0/` etc. dirs for Claude Code and Cursor where breaking changes are documented. |
| 11 | Implement MCP resource for knowledge lookup | Task | MCP Framework | P1 | Wire `vibeseed://knowledge/*` MCP resources to KnowledgeLookupPort. |

---

## Sources

### Claude Code
- Official docs -- subagents: https://code.claude.com/docs/en/sub-agents
- Official docs -- settings: https://code.claude.com/docs/en/settings
- Changelog: https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md
- Config guide: https://claudelog.com/configuration/
- Settings guide (2026): https://www.thecaio.ai/blog/claude-code-settings-guide
- Customization guide: https://alexop.dev/posts/claude-code-customization-guide-claudemd-skills-subagents/

### Cursor
- Rules guide: https://design.dev/guides/cursor-rules/
- Settings location: https://www.jackyoustra.com/blog/cursor-settings-location
- MDC rules reference: https://github.com/sanjeed5/awesome-cursor-rules-mdc
- Changelog: https://cursor-changelog.com/
- Best practices: https://github.com/digitalchild/cursor-best-practices

### Roo Code
- Custom modes docs: https://docs.roocode.com/features/custom-modes
- Custom instructions: https://docs.roocode.com/features/custom-instructions
- GitHub: https://github.com/RooCodeInc/Roo-Code
- Changelog: https://github.com/RooCodeInc/Roo-Code/blob/main/CHANGELOG.md
- Global config paths: https://github.com/RooCodeInc/Roo-Code/issues/10750

### OpenCode
- Config docs: https://opencode.ai/docs/config/
- Agents docs: https://opencode.ai/docs/agents/
- Modes docs: https://opencode.ai/docs/modes/
- Rules docs: https://opencode.ai/docs/rules/
- Changelog: https://opencode.ai/changelog
- GitHub (active): https://github.com/sst/opencode

### Cross-Tool
- AGENTS.md standard: https://agents.md/
- AGENTS.md overview: https://www.devshorts.in/p/agentsmd-one-file-for-all-agents
- Antigravity setup: https://github.com/irahardianto/antigravity-setup

### Internal References
- PRD: `docs/PRD.md` (capabilities C17, C18, C22, C23)
- DDD: `docs/DDD.md` (Knowledge Base bounded context, section 4)
- CLI/MCP design: `docs/research/20260222_cli_mcp_design.md` (KnowledgeLookupPort, MCP resources)
