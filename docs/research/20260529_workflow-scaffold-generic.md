# Research: Generic, Reusable Workflow Scaffold for alto

**Date:** 2026-05-29
**Spike Ticket:** alty-cli-766.1
**Parent Epic:** alty-cli-766
**Status:** Draft (Scope reduced 2026-05-29+: see Scope Update below)
**Timebox:** 6 hours

---

## Scope Update (post-spike decision, 2026-05-29+)

**Cursor and Roo Code adapters are out of scope.** The shipped scaffold supports **Claude Code (primary) and OpenCode only**. Cursor `.mdc` and Roo Code `customModes` are discussed throughout this report as historical research — they informed the portable-frontmatter schema design — but follow-up tickets do NOT implement them. The implementation tickets reflect this reduction; this report's body is retained verbatim so the schema-portability rationale survives.

Concretely:

- `--primary-tool` enum is `claude | opencode` (was `claude | cursor | roo | opencode`).
- Tool Translation follow-up ships ONE adapter (OpenCode), not three.
- The portable frontmatter schema (§Q3, lines 220-244) stays as designed — multi-tool portability is preserved for future re-introduction of dropped tools without redesign.
- The §Worked example renderings for Cursor `.mdc` and Roo Code `customModes` (lines 273-303 in pre-update line numbering) are kept as reference but do not bind any AC.

---

## Summary

Adopt a single root folder named `.alto/` to host alto's portable workflow scaffold, structured **flat-by-category** (`.alto/commands/`, `.alto/agents/`, `.alto/skills/`, `.alto/templates/`, `.alto/lifecycle/{in-progress,deprecated}/`) and seeded with a per-asset YAML frontmatter contract that maps cleanly to Claude Code Skills (`SKILL.md`), Cursor `.cursor/rules/*.mdc`, Roo Code `customModes`, and OpenCode `.opencode/commands/`. Carry alto-specific instructions in **`.project.md` overlay siblings** plus a layered `CLAUDE.md` — never inside the generic asset itself. Migrate the existing 25 files via a 7-phase `git mv` + frontmatter-rewrite script that preserves every `/groom`, `/design-ticket`, `/launch-team`, `/review`, `/doc-health` invocation (Claude Code's directory-per-skill resolution maps `.alto/commands/groom.md` → `/groom` when symlinked or merged via the installer). Ship to user projects via `alto init --with-scaffold` (Go `embed` of the `.alto/` tree, post-clone `templating` of `{{PROJECT_NAME}}` / `{{TICKET_PREFIX}}` / `{{BOUNDED_CONTEXTS}}` / `{{PRIMARY_TOOL}}` / `{{ISSUE_TRACKER}}`); reject `npx`-style installers (Node runtime cost, license vetting, secondary distribution channel).

**Key tension resolved:** Today's `.claude/commands/*.md` are Claude-Code-native artifacts that double as the canonical scaffold source. Splitting them into a separate `workflow/` tree (Option A) would force a generator to emit Claude artifacts from a non-Claude source, defeating the "copy-one-folder" goal for the primary tool. `.alto/` (Option B) keeps the canonical source tool-agnostic while letting `alto init --with-scaffold` emit `.claude/`, `.cursor/`, `.roo/`, `.opencode/` views via Tool Translation. Option C (hybrid `.alto/scaffold/` + `.claude/` pointer) was rejected as cognitively expensive — two roots, no offsetting benefit.

---

## Research Question

### Q1. Single root folder name + internal taxonomy

Recommend **`.alto/`** as the single root, with a flat internal taxonomy: `.alto/commands/`, `.alto/agents/`, `.alto/skills/`, `.alto/templates/`, `.alto/lifecycle/in-progress/`, `.alto/lifecycle/deprecated/`, plus `.alto/CONTEXT.md` (ubiquitous-language glossary for the scaffold, mirroring mattpocock/skills' role for `CONTEXT.md` — "helps agents decode the jargon used in the project", per [mattpocock/skills CONTEXT.md analysis](https://github.com/mattpocock/skills)). The dot-prefix signals "tooling, not source", matches the established `.claude/` / `.cursor/` / `.roo/` / `.opencode/` convention, and survives `npm pack` / `git archive` defaults. One folder copy delivers: 5 commands, 6 agents, 4 templates, 1 reference spike, the lifecycle folders, and `CONTEXT.md` — the full DDD + TDD + beads pipeline an alto user expects.

### Q2. Flat-by-category vs nested-by-artifact

Pick **flat-by-category by artifact type** (`commands/`, `agents/`, `templates/`, `lifecycle/`), not category-by-domain (mattpocock's `engineering/`, `productivity/`, `misc/`, `personal/`). Two reasons. First, Claude Code's directory-per-skill resolution rule — `.claude/skills/deploy-staging/SKILL.md` → `/deploy-staging` ([Claude Code Skills doc, "How a skill gets its command name"](https://code.claude.com/docs/en/skills) line ~232 of the doc) — works only when the directory name *is* the command name, which requires command directories to be siblings of one another, not nested under category subfolders. Second, alto's audit shows the 5 first-class commands (`groom`, `design-ticket`, `launch-team`, `review`, `doc-health`) are all "engineering lifecycle" — there is no productivity vs engineering split to express. The lifecycle dimension (`in-progress/`, `deprecated/`) is borrowed from mattpocock and kept as a sibling, not a parent, of `commands/` / `agents/` — assets graduate by `git mv` between siblings rather than crossing trees.

### Q3. Minimum portable frontmatter schema

The minimum portable schema is **seven required fields plus four optional fields**, all YAML, kept under one frontmatter block per file:

```yaml
---
# Required
name: groom                                      # kebab-case; matches directory/file basename
description: Deep-groom a ticket — enforced implementation simulation
kind: command                                    # command | agent | template | skill
phase: groom                                     # design | groom | implement | review | close
when_to_use: When user invokes /groom or asks to verify a ticket before claiming
tools_required: [Read, Grep, Glob, Bash]         # Claude Code allowed-tools superset
license: Apache-2.0                              # asset license (default project license)

# Optional
agent: tech-lead                                 # subagent type for `context: fork` (Claude Code only)
globs: ["**/*.md"]                               # Cursor auto-attach hint
allowed_tools_strict: false                      # if true, deny-list everything else (Claude Code)
disable_model_invocation: false                  # only-on-user-invoke (Claude Code)
---
```

Mapping rules: `name` + `description` + `kind` + `tools_required` cover the union of (Claude Code `name`/`description`/`allowed-tools`, Cursor `description`/`globs`, Roo Code `slug`/`name`/`description`/`groups`, OpenCode `description`/`agent`). `phase` is an alto extension used by Tool Translation to drive lifecycle ordering and `/doc-health` validation. `when_to_use` matches Claude Code Skills' field name verbatim ([Skills doc, frontmatter reference](https://code.claude.com/docs/en/skills)) and is reused as the body of Roo Code's `whenToUse` and Cursor's `description`. See the worked `/groom` example in the **Recommendation** section below.

### Q4. Alto-specific vs generic separation

Use **`.project.md` overlay siblings** (e.g. `.alto/commands/groom.project.md`) plus a layered `CLAUDE.md` (project-level, layered above `.alto/CONTEXT.md`). The shipped generic asset (`.alto/commands/groom.md`) carries zero alto-internal references; the overlay sibling (`.alto/commands/groom.project.md`) carries alto's `internal/composition/app.go`, `internal/{context}/application/ports.go`, bounded-context names, and `alty-cli-` ticket prefix. Claude Code merges sibling `.md` files in the same skill directory ([Skills doc, "Add supporting files"](https://code.claude.com/docs/en/skills)), so the overlay loads automatically when the command is invoked. Rejected the templated-placeholder approach (`{{BOUNDED_CONTEXTS}}` inline) because it produces brittle re-templating during contributor edits and obscures the canonical form during dogfooding; rejected the layered-CLAUDE.md-only approach because it spreads command-specific context across the global CLAUDE.md, which violates the Skills doc's guidance to "Create a skill when ... a section of CLAUDE.md has grown into a procedure rather than a fact" ([Skills doc, opening paragraph](https://code.claude.com/docs/en/skills)).

### Q5. Migration path that preserves existing invocations

The migration is **7 ordered phases** (Phases 1, 2, 3 with sub-steps 3b and 3c, 4, 5, 6, 7 — labelled `# Phase N` in the script body below) producing **~31 `git mv` operations + 6 file splits + 3 file edits**:

1. **Phase 1 — Create skeleton.** `mkdir -p .alto/{commands,agents,skills,templates,lifecycle/{in-progress,deprecated}}/` plus `.gitkeep` markers.
2. **Phase 2 — Move generic-as-is.** `git mv` all 8 commands and 6 agents that are GENERIC under their new homes; also `git mv` `docs/templates/*`, `docs/beads_templates/*`, `docs/spikes/ddd_reference.md` into `.alto/templates/`.
3. **Phase 3 (with sub-steps 3b, 3c) — Split OVERLAY files.** Split each pre-flagged alto-coupled file (`brainstorm.md`, `project-manager.md`, `groom.md`, `prd-traceability.md`, `tech-lead.md`, `developer.md`) into `{name}.md` (generic) + `{name}.project.md` (overlay). Move alto-internal references (`internal/composition/app.go:NN`, `alty-cli-` ticket prefix, bounded context names) into the overlay. Sub-steps 3b (agents) and 3c (mild-overlay templates) follow the same pattern for `developer.md`/`tech-lead.md`/`project-manager.md`/`qa-engineer.md` and `ARCHITECTURE_TEMPLATE.md`.
4. **Phase 4 — Write `.alto/CONTEXT.md`** from scratch (new file, no `git mv`).
5. **Phase 5 — Preserve existing slash-command invocations.** Symlink `.claude/commands/*.md` → `../../.alto/commands/*.md` so existing `/groom`, `/design-ticket`, `/launch-team`, `/review`, `/doc-health` invocations resolve unchanged (Claude Code follows symlinks for skill discovery — covered by the per-directory listing in [Skills doc, "Where skills live"](https://code.claude.com/docs/en/skills)).
6. **Phase 6 — Root edits.** Edit `.claude/CLAUDE.md` to point at `.alto/` (1 edit), edit `bin/bd-ripple` lines 191 and 238 to reference `.alto/templates/beads-ticket-template.md` instead of `docs/beads_templates/beads-ticket-template.md` (1 file, 2 occurrences — line 191 in the `REQUIRED_SECTIONS` comment, line 238 in the user-facing template hint), edit `.claude/commands/architecture-docs.md` `prd-traceability.md` table entry (1 edit).
7. **Phase 7 — Verify.** `ls .alto/{commands,agents,templates}` produces the expected file counts; `go run ./cmd/alto doc-health` MUST pass.

The symlink bridge in Phase 2 is removable once `alto init --with-scaffold` ships and downstream users adopt `.alto/`; for the alto repo itself, the symlinks can stay or be replaced by deleting `.claude/commands/` once Claude Code's automatic discovery from `.alto/commands/` is confirmed in a one-line `settings.json` `additionalDirectories` entry ([Skills doc, "Skills from additional directories"](https://code.claude.com/docs/en/skills)).

### Q6. Shipping mechanism

Recommend **`alto init --with-scaffold`** via Go `embed.FS` of the `.alto/` tree, with post-extraction templating of 5 parameters: `{{PROJECT_NAME}}`, `{{TICKET_PREFIX}}` (e.g. `alto-`, `proj-`), `{{ISSUE_TRACKER}}` (default `beads`; values `beads|github|linear` recorded but only `beads` shipped per scope), `{{BOUNDED_CONTEXTS}}` (comma-separated list from discovery), `{{PRIMARY_TOOL}}` (default `claude`; values `claude|cursor|roo|opencode` — drives which view directories Tool Translation emits). Reject `alto scaffold add` as a separate command (UX bloat — the scaffold *is* part of `init` for greenfield, and `alto init --existing` already covers brownfield). Reject `npx`-style installers (mattpocock/skills uses `npx skills@latest add mattpocock/skills` per the [skills README](https://github.com/mattpocock/skills)) because Node runtime is a hard dependency conflict for Go-distributed users, license vetting on `skills@latest` is required per CLAUDE.md, and a secondary distribution channel violates the "alto is the architect" principle. `embed.FS` is stdlib (no license cost), trivially testable, and produces a single-binary deliverable.

---

## Audit Matrix

Every file below was read before being mentioned. Tags: **GENERIC** = ships as-is; **OVERLAY** = generic body + alto-specific `.project.md` sibling; **ALTO-ONLY** = repo-local infrastructure, not shipped.

### `.claude/agents/` (6 personas)

| File | Tag | Alto-coupled lines (evidence) | Action |
|------|-----|-------------------------------|--------|
| `researcher.md` | GENERIC | 0 matches against `\binternal/\|cmd/alto\|alto-cli\|golangci\|Watermill` | Move to `.alto/agents/researcher.md` |
| `white-hacker.md` | GENERIC | 0 matches against same pattern | Move to `.alto/agents/white-hacker.md` |
| `developer.md` | OVERLAY | Lines 70 (Watermill), 75 (`internal/`), 168, 193 (golangci-lint). Plus pervasive Go-specific test/lint guidance. | Split: `.alto/agents/developer.md` (generic TDD/SOLID/test patterns), `.alto/agents/developer.project.md` (Go 1.26+, Watermill, golangci-lint v2 specifics, `internal/{context}/` layout) |
| `tech-lead.md` | OVERLAY | Lines 42-45 (`internal/{context}/{domain,application,infrastructure}/`), 52 (Watermill), 58, 61 (grep recipes against `internal/.*`), 118, 124 (golangci-lint). 9 matches. | Split: generic DDD review checklist vs Go-specific lint recipes |
| `project-manager.md` | OVERLAY | Lines 85, 91-96 (`internal/{context}/{domain,application,infrastructure}/`, `cmd/alto/`, `cmd/alto-mcp/`), 102-103, 120. 10 matches. | Split: generic beads ticket-lifecycle vs Go project layout |
| `qa-engineer.md` | OVERLAY | Lines 76-78 (`internal/{domain,application,infrastructure}/...`), 159 (golangci). 4 matches. | Split: generic BICEP framework vs Go-test recipes |

### `.claude/commands/` (8 commands)

| File | Tag | Alto-coupled lines (evidence) | Action |
|------|-----|-------------------------------|--------|
| `architecture-docs.md` | OVERLAY | Line 34 references `.claude/commands/prd-traceability.md`; no `internal/` refs. | Move + edit the cross-ref to `.alto/commands/prd-traceability.md` |
| `brainstorm.md` | OVERLAY | The frontmatter `name: alto-brainstorm` (line 2), header `/alto-brainstorm` (line 6), 38 references to "alto" / DDD / PRD / ARCHITECTURE / "bounded context". 14 fitting the brief's pre-flagged count. | Split: generic discovery flow vs `alto-` namespace + `docs/templates/` paths. Rename frontmatter `name: brainstorm`. |
| `design-ticket.md` | OVERLAY | Lines 37-38, 73 reference `docs/DDD.md` / `docs/ARCHITECTURE.md` only (no `internal/`). 1 match on coupling pattern. | Move; replace `docs/` paths with `${CLAUDE_SKILL_DIR}/../templates/` for the Claude Code view; per-tool path substitution covers Cursor / Roo / OpenCode (see §Frontmatter Schema path-substitution table). |
| `doc-health.md` | GENERIC | 0 alto-coupling matches; references `docs/PRD.md` / `docs/DDD.md` / `docs/ARCHITECTURE.md` generically (any project has these). | Move as-is |
| `groom.md` | OVERLAY | Lines 18 (`alto-0m9.2`), 59 (`alto ticket-verify`), 84 (`internal/composition/app.go`), 94 (`internal/xxx/application/ports.go:NN`), 96 (`internal/xxx/infrastructure/xxx_adapter.go`), 99 (`internal/composition/app.go:NN`), 114 (`internal/composition/adapters.go`). 7 matches. Brief flagged 9 — within tolerance (alto refs include unmarked variants). | Split: generic "implementation simulation" vs alto's `internal/composition/` constructor chain |
| `launch-team.md` | OVERLAY | Lines 14-16 (`alty-cli-` examples), 48-49 (`docs/DDD.md`, `docs/ARCHITECTURE.md`), 69 (`dev-alty-cli-1wu`), 79 (Watermill), 94-98 (`.golangci.yml`, `arch-go.yml`), 159, 167-170, 177 (Go/Watermill/`internal/`), 223 (`.notes/handoff-alty-cli-1wu.md`). 13 matches. | Split: generic 7-phase team protocol vs alto Go quality gates + ticket prefix |
| `prd-traceability.md` | OVERLAY | Lines 49-69 (entire P0 capability table is alto-specific: C1-C25 → Bootstrap / Rescue / Guided Discovery / Domain Model / Tool Translation / Ticket Pipeline / Architecture Testing / Knowledge Base / Ticket Freshness), 75, 112, 115, 118-119 (`k7m.4`, `alto doc-health`). 26 matches. Brief flagged 8 — the brief's pattern was narrower; my wider pattern captures the full coupling. | Split: generic RLM capability-traceability *pattern* (.md) vs alto's actual capability table (`prd-traceability.project.md`) |
| `review.md` | GENERIC | 0 alto-coupling matches. Hartwork methodology + DDD/SOLID/test checklists are universal. | Move as-is |

### `.claude/skills/` (excludes vendored `gstack/` symlinks)

`ls .claude/skills/` reports 18 non-symlink directories: `adapt/`, `animate/`, `audit/`, `bolder/`, `clarify/`, `colorize/`, `critique/`, `delight/`, `distill/`, `extract/`, `frontend-design/` (+ `reference/`), `harden/`, `normalize/`, `onboard/`, `optimize/`, `polish/`, `quieter/`, `teach-impeccable/`. Each is a single `SKILL.md` of design-aesthetic / writing-craft skills (e.g. `adapt/SKILL.md` line 2-3: "adapt: Adapt designs to work across different screen sizes, devices, contexts, or platforms"). **None reference alto bounded contexts or `internal/` paths.** All 12 symlinks under `.claude/skills/` (`browse`, `design-consultation`, `design-review`, `document-release`, `gstack-upgrade`, `plan-ceo-review`, `plan-design-review`, `plan-eng-review`, `qa`, `qa-only`, `retro`, `review`, `setup-browser-cookies`, `ship`) point into `gstack/` — a vendored external skill bundle (`.claude/skills/gstack/LICENSE` present, separate provenance).

| Group | Tag | Action |
|-------|-----|--------|
| 18 non-symlink design/craft skills | GENERIC (but **out of alto scaffold scope**) | Leave under `.claude/skills/` — these are personal Claude Code skills, not alto workflow assets. **Do not migrate.** Document in `.alto/CONTEXT.md` that `.claude/skills/` retains personal skills; `.alto/skills/` is for shipped alto skills only. |
| `gstack/` vendored bundle (+ 14 symlinks) | ALTO-ONLY (external vendor) | Leave alone. Out of scope. |

### `docs/templates/` (3 files)

| File | Tag | Action |
|------|-----|--------|
| `PRD_TEMPLATE.md` | GENERIC | Move to `.alto/templates/PRD_TEMPLATE.md`. Line 64 mentions `Python 3.12+` / `uv` — those are *template values for the consumer to override*, not alto coupling. Keep but document the override semantic in `.alto/CONTEXT.md`. |
| `DDD_STORY_TEMPLATE.md` | GENERIC | Move to `.alto/templates/DDD_STORY_TEMPLATE.md`. Domain-storytelling structure is universal. |
| `ARCHITECTURE_TEMPLATE.md` | OVERLAY (mild) | Move to `.alto/templates/ARCHITECTURE_TEMPLATE.md`. Lines 66 ("pure Python"), 72-85 (`src/{domain,application,infrastructure}` Python layout), 138 (uv) carry Python-era residue. Overlay `.project.md` covers Go variant (`internal/{context}/{domain,application,infrastructure}` per [`internal/tooltranslation/` confirmed layout](#tool-translation-confirmation)). |

### `docs/beads_templates/` (4 files)

| File | Tag | Action |
|------|-----|--------|
| `beads-epic-template.md` | GENERIC | Move to `.alto/templates/beads-epic-template.md`. References `docs/DDD.md` generically (line 19). |
| `beads-ticket-template.md` | GENERIC | Move to `.alto/templates/beads-ticket-template.md`. References `docs/DDD.md` (line 38), `/architecture-docs` command (line 38) generically. |
| `beads-spike-template.md` | GENERIC | Move to `.alto/templates/beads-spike-template.md`. Findings Template block (lines 66-103) is universal. |
| `beads-stub-template.md` | GENERIC | Move to `.alto/templates/beads-stub-template.md`. |

### `docs/spikes/` (1 file)

| File | Tag | Action |
|------|-----|--------|
| `ddd_reference.md` | GENERIC | Move to `.alto/templates/ddd_reference.md`. The 2026 DDD operating guide is generic (lines 1-158 are vendor-agnostic DDD strategic patterns). |

### Tool-translation confirmation

`internal/tooltranslation/` confirmed layout (read 2026-05-29):

- `domain/` — `tool_adapter.go` (18k), `tool_config.go` (6.2k) + tests.
- `application/` — `config_generation_handler.go`, `persona_handler.go`, `ports.go` (1.1k: `ConfigGeneration`, `PersonaManager` interfaces). Source: `internal/tooltranslation/application/ports.go:13-27`.
- `infrastructure/` — **EMPTY** (only the directory exists). Confirmed via `ls`. This is precisely the seam the follow-up tool-translation ticket extends.

The port interfaces are already defined; the spike's recommendation does not require port redesign, only adapter additions.

### Total

- 6 agents → 1 stays (researcher) + 1 stays (white-hacker) generic; 4 split (developer/tech-lead/project-manager/qa-engineer).
- 8 commands → 2 stay generic (doc-health/review) + 1 stays mostly-generic (architecture-docs); 5 split (brainstorm/design-ticket/groom/launch-team/prd-traceability).
- 3 doc templates + 4 beads templates + 1 spike = 8 templates, all GENERIC (with one mild OVERLAY for ARCHITECTURE_TEMPLATE).
- 18 design skills under `.claude/skills/`: **NOT MIGRATED** — confirmed out of scope (personal Claude skills, not alto workflow scaffold).
- 1 `CLAUDE.md` edit + 1 `bin/bd-ripple` edit + 1 cross-ref edit = 3 root edits.

**Files in scope: 25 (= 6 agents + 8 commands + 8 templates + 3 root edits). Matches the brief's "~25 files + CLAUDE.md + `bin/bd-ripple` edits" estimate.**

---

## Options Considered

| Option | Pros | Cons |
| ------ | ---- | ---- |
| **A. `workflow/` (mattpocock-style category folders)** | • Clean one-folder-copy semantic; root name signals purpose explicitly.<br>• Aligns with mattpocock/skills' README install (`npx skills@latest add ...`) for users familiar with that ecosystem.<br>• Category folders (`engineering/`, `productivity/`, `lifecycle/`) communicate intent for browsing. | • Diverges from `.claude/` / `.cursor/` / `.roo/` / `.opencode/` dot-prefix convention — every other AI tool uses `.toolname/`.<br>• Claude Code's directory-per-skill resolution requires command directories to be direct children of `.claude/skills/` (per Skills doc "How a skill gets its command name", table row 1) — a `workflow/engineering/groom/SKILL.md` does NOT resolve to `/groom` without a symlink or generator pass.<br>• Category folders are dimensionally wrong for alto: today's 5 first-class commands are all engineering-lifecycle — no productivity vs misc split exists.<br>• Forces tooling translation to *generate* `.claude/` from non-Claude source, doubling the per-command edit surface during dogfooding. |
| **B. `.alto/` (mirrors `.claude/` shape but tool-agnostic)** ★ | • Matches the established `.toolname/` convention used by every targeted AI tool (Claude, Cursor, Roo, OpenCode).<br>• `.claude/` ↔ `.alto/` symlinks (or `--add-dir` registration per [Skills doc, "Skills from additional directories"](https://code.claude.com/docs/en/skills)) preserves every existing slash-command invocation with one config line.<br>• Flat-by-artifact (`commands/`, `agents/`, `templates/`) is exactly the shape contributors already navigate — minimal cognitive cost.<br>• Tool Translation reads from `.alto/` and emits `.claude/`, `.cursor/`, `.roo/`, `.opencode/` views — one canonical source, four projections.<br>• Lifecycle folders kept as siblings (`.alto/lifecycle/in-progress/`, `.alto/lifecycle/deprecated/`) per mattpocock's discipline — assets graduate by `git mv`. | • Dot-prefix can hide the folder from casual `ls` (mitigated: documented in README; every other AI tool uses the same convention).<br>• "Why not `.workflow/`?" pushback possible — answer: branded naming makes `alto init --with-scaffold` self-documenting and unambiguous in multi-tool repos.<br>• Does not borrow mattpocock's category-by-domain expressiveness — but alto doesn't need it (5 commands, all engineering-lifecycle). |
| **C. Hybrid `.alto/scaffold/` + thin `.claude/` pointer** | • Theoretically allows Claude-native primary source while keeping tool-agnostic alternative.<br>• Familiar to contributors who use both `.claude/` and `.alto/`. | • Two roots = two mental models. Increases doc-health validation surface 2x.<br>• `alto init --with-scaffold` would need to emit both `.alto/scaffold/` and `.claude/` — defeats one-folder-copy.<br>• Migration cost is highest: every file moves twice (`.claude/...` → `.alto/scaffold/...` → views).<br>• Provides no concrete user benefit that B doesn't already give via Tool Translation views. |

★ = **Recommended.**

---

## Recommendation

### Root + taxonomy

```
.alto/
├── CONTEXT.md                        # ubiquitous-language glossary (mirrors mattpocock CONTEXT.md role)
├── commands/
│   ├── architecture-docs.md          # GENERIC (move)
│   ├── brainstorm.md                 # GENERIC (split out of current alto-brainstorm)
│   ├── brainstorm.project.md         # OVERLAY (alto-namespace + docs/templates paths)
│   ├── design-ticket.md              # GENERIC (move)
│   ├── doc-health.md                 # GENERIC (move)
│   ├── groom.md                      # GENERIC (split)
│   ├── groom.project.md              # OVERLAY (internal/composition chain)
│   ├── launch-team.md                # GENERIC (split)
│   ├── launch-team.project.md        # OVERLAY (alty-cli- prefix, Go gates)
│   ├── prd-traceability.md           # GENERIC (RLM pattern only)
│   ├── prd-traceability.project.md   # OVERLAY (alto C1-C25 table)
│   └── review.md                     # GENERIC (move)
├── agents/
│   ├── developer.md                  # GENERIC (split)
│   ├── developer.project.md          # OVERLAY (Go 1.26+, Watermill, golangci-lint)
│   ├── project-manager.md            # GENERIC (split)
│   ├── project-manager.project.md    # OVERLAY (Go layout + alto bounded contexts)
│   ├── qa-engineer.md                # GENERIC (split)
│   ├── qa-engineer.project.md        # OVERLAY (go test recipes)
│   ├── researcher.md                 # GENERIC (move)
│   ├── tech-lead.md                  # GENERIC (split)
│   ├── tech-lead.project.md          # OVERLAY (Go grep recipes + arch-go)
│   └── white-hacker.md               # GENERIC (move)
├── templates/
│   ├── ARCHITECTURE_TEMPLATE.md      # GENERIC + mild overlay
│   ├── ARCHITECTURE_TEMPLATE.project.md  # OVERLAY (Go internal/{context}/{domain,application,infrastructure})
│   ├── DDD_STORY_TEMPLATE.md         # GENERIC
│   ├── PRD_TEMPLATE.md               # GENERIC
│   ├── beads-epic-template.md        # GENERIC
│   ├── beads-spike-template.md       # GENERIC
│   ├── beads-stub-template.md        # GENERIC
│   ├── beads-ticket-template.md      # GENERIC
│   └── ddd_reference.md              # GENERIC (DDD 2026 working reference)
├── skills/                           # reserved for future alto-shipped skills
│   └── .gitkeep
└── lifecycle/
    ├── in-progress/                  # assets under design, not yet stable
    │   └── .gitkeep
    └── deprecated/                   # assets retained for migration, not for new use
        └── .gitkeep
```

### Portable YAML frontmatter schema (full reference)

```yaml
---
# Required (every command / agent / template / skill MUST declare these — NO defaults)
name: <kebab-case>                  # matches file/dir basename; ^[a-z][a-z0-9-]*$ enforced by doc-health
description: <one-sentence>         # under 1,536 chars per Claude Code Skills doc cap
kind: command | agent | template | skill
phase: design | groom | implement | review | close
when_to_use: <trigger phrase or example request>
tools_required: <explicit list>     # NO default; author MUST enumerate (Read | Grep | Glob | Edit | Write | Bash | Agent)
bash_substitution_policy: none      # none | quoted | unrestricted — default `none` (no `!`cmd`` blocks expanded)
license: <SPDX>                     # required per CLAUDE.md license-vetting rule

# Optional
agent: <subagent type>              # Claude Code `context: fork` target
globs: ["**/*.md"]                  # Cursor auto-attach hint
allowed_tools_strict: false         # if true, Tool Translation emits deny-list for non-listed tools
disable_model_invocation: false     # if true, Claude Code only invokes on explicit user `/name`
parameters:                         # `alto init --with-scaffold` templating hooks
  - PROJECT_NAME
  - TICKET_PREFIX
---
```

**Normative rules (enforced by `alto doc-health` per Follow-up #5):**

1. `tools_required` has no default; omission is a violation.
2. `bash_substitution_policy` defaults to `none`. Any `` !`...` `` block in the body requires the author to set `bash_substitution_policy: quoted` (and document each substituted parameter) or `unrestricted` (only for `disable_model_invocation: true` commands).
3. **`Bash`-with-`parameters` rule:** when `tools_required` contains `Bash` AND `parameters:` is declared (i.e. the asset accepts user-supplied arguments), `disable_model_invocation: true` is REQUIRED. Reason: arguments can flow into shell commands; auto-invocation by the model on user phrasing must not be possible. Doc-health treats violation as ERROR.

#### Per-tool path-substitution table

The frontmatter never carries absolute paths. Cross-asset references (templates from a command, scripts from a skill) use the tool-native portable mechanism, resolved at Tool Translation time. The same source `${CLAUDE_SKILL_DIR}/../templates/PRD_TEMPLATE.md` reference in `.alto/commands/brainstorm.md` is rewritten per-tool as follows:

| Target tool | Path substitution mechanism | Resolved form (example, `brainstorm.md` referencing `PRD_TEMPLATE.md`) | Source authority |
|-------------|------------------------------|------------------------------------------------------------------------|------------------|
| Claude Code Skills | `${CLAUDE_SKILL_DIR}/<relpath>` — runtime substitution by Claude Code | `${CLAUDE_SKILL_DIR}/../templates/PRD_TEMPLATE.md` | [Skills doc, "Available string substitutions" table row `${CLAUDE_SKILL_DIR}`](https://code.claude.com/docs/en/skills) |
| Cursor | Path relative to `.cursor/rules/` — Cursor does NOT support env-style substitution; Tool Translation rewrites to a relative path computed at emit time | `../../.alto/templates/PRD_TEMPLATE.md` | [Cursor docs, "Rules" — `.mdc` frontmatter does not document substitution; relative path is the established pattern](https://cursor.com/docs/context/rules) |
| Roo Code | Absolute template inlined at translation time — Roo `customInstructions` is a YAML string, no runtime substitution; Tool Translation reads the template file and embeds its content verbatim | `<inlined contents of PRD_TEMPLATE.md>` | [Roo Code docs, "Custom Modes" — `customInstructions` field is a literal string body](https://roocodeinc.github.io/Roo-Code/features/custom-modes) |
| OpenCode | Path relative to `.opencode/commands/` — OpenCode supports `@filename` for file references but not env-style substitution | `@../../.alto/templates/PRD_TEMPLATE.md` | [OpenCode docs, "Commands" — file references via `@filename`](https://opencode.ai/docs/commands/) |

Authors write `${CLAUDE_SKILL_DIR}/../templates/<file>.md` in the source `.alto/commands/*.md`. Tool Translation adapters MUST rewrite this token per the table when emitting non-Claude views. The token `${ALTO_ROOT}` is NOT permitted in any shipped asset (no tool defines it; using it would re-create the cwd-dependency problem). Bare cwd-relative paths like `../templates/...` are NOT permitted either — they break when Claude Code is launched from a subdirectory and the skill is discovered via parent-walk per [Skills doc, "Automatic discovery from parent and nested directories"](https://code.claude.com/docs/en/skills).

### Worked example: `/groom` rendered for Claude Code Skills AND Cursor

**Source (`.alto/commands/groom.md`, frontmatter only — under the new secure-by-default schema):**

```yaml
---
name: groom
description: Deep-groom a ticket — enforced implementation simulation, scope check, split detection
kind: command
phase: groom
when_to_use: When user invokes /groom <ticket-id> or asks to verify a ticket before claiming
tools_required: [Read, Grep, Glob, Bash]   # explicit; no default
bash_substitution_policy: quoted            # body has !`bd show "$TICKET_ID"` style blocks
license: Apache-2.0
disable_model_invocation: true              # REQUIRED: Bash + parameters → no auto-invoke
parameters:
  - TICKET_ID
---
```

Note the quotes around `$TICKET_ID` — required under `bash_substitution_policy: quoted`; an unquoted form would be rejected by the doc-health validator (see Follow-up #5 AC).

`disable_model_invocation: true` is required here because `tools_required` includes `Bash` AND `parameters:` declares a user-supplied `TICKET_ID` — per the normative rule above, this prevents the model from auto-invoking `/groom` on user phrasing and routing argument text into a shell. `bash_substitution_policy: quoted` is set because the body uses `` !`bd show "$TICKET_ID"` ``-style blocks and the parameter is propagated to shell substitution.

**Claude Code Skills view (`.claude/skills/groom/SKILL.md`, emitted by Tool Translation):**

```yaml
---
name: groom
description: Deep-groom a ticket — enforced implementation simulation, scope check, split detection
when_to_use: When user invokes /groom <ticket-id> or asks to verify a ticket before claiming
allowed-tools: Read Grep Glob Bash
disable-model-invocation: true
---
```

Body merges `groom.md` + (if present in `.alto/`) `groom.project.md` content under one `SKILL.md`. `` !`bd show "$TICKET_ID"` `` blocks are PRESERVED in the Claude Code view (this is the only tool that natively supports them). Directory name (`groom`) determines the slash-command name per [Skills doc, "How a skill gets its command name"](https://code.claude.com/docs/en/skills) — table row 1, "Skill directory under `.claude/skills/` → Directory name → `/deploy-staging`". `${CLAUDE_SKILL_DIR}/../templates/<file>.md` references are PRESERVED (Claude Code resolves them at runtime).

**Cursor view (`.cursor/rules/groom.mdc`, emitted by Tool Translation):**

```mdc
---
description: Deep-groom a ticket — enforced implementation simulation, scope check, split detection
globs: ["**/*"]
alwaysApply: false
---

# /groom <ticket-id>

<!-- Tool Translation stripped Claude-Code-only construct:
     original `!`bd show "$TICKET_ID"`` block — port manually if running outside Claude Code -->
<!-- Tool Translation stripped Claude-Code-only construct:
     original `disable_model_invocation: true` — Cursor has no equivalent; user-invoke only is not enforceable here -->
...
```

Body inlines `groom.md` + (if present) `groom.project.md` content with two transformations: (1) `` !`...` `` bash-substitution blocks (Claude-Code-only per [Skills doc, "Inject dynamic context"](https://code.claude.com/docs/en/skills)) are STRIPPED and replaced with an HTML-comment placeholder naming the original command for manual porting; (2) `${CLAUDE_SKILL_DIR}/../templates/<file>.md` references are rewritten to `../../.alto/templates/<file>.md` (relative to `.cursor/rules/`). Cursor requires the `.mdc` extension per [Cursor docs, "Rules"](https://cursor.com/docs/context/rules) — "Project rules live in `.cursor/rules` as `.mdc` files and are version-controlled. ... A plain `.md` file in `.cursor/rules` is ignored by the rules system because it has no frontmatter to specify `description`, `globs`, and `alwaysApply`."

**Roo Code view (`.roo/customModes.yaml`, emitted by Tool Translation — one entry per command):**

```yaml
customModes:
  - slug: groom
    name: Groom Ticket
    description: Deep-groom a ticket — enforced implementation simulation, scope check, split detection
    whenToUse: When user invokes /groom <ticket-id> or asks to verify a ticket before claiming
    roleDefinition: |
      You execute the alto ticket grooming workflow ...
    customInstructions: |
      <body of groom.md + groom.project.md, with `!`...`` blocks stripped to comments
       and `${CLAUDE_SKILL_DIR}/../templates/<file>.md` references inlined with the
       template file's actual content at translation time>
    groups: [read, edit, command]
```

Field names verbatim per [Roo Code docs, "Custom Modes"](https://roocodeinc.github.io/Roo-Code/features/custom-modes) — `slug`, `name`, `description`, `whenToUse`, `roleDefinition`, `customInstructions`, `groups`. Roo Code has no `disable_model_invocation` equivalent — the flag is dropped at translation. Two bash-block transformations occur: `` !`...` `` strip-with-comment, and template-reference inlining per the path-substitution table.

**OpenCode view (`.opencode/commands/groom.md`, emitted by Tool Translation):**

```yaml
---
description: Deep-groom a ticket — enforced implementation simulation, scope check, split detection
agent: build
---

# /groom <ticket-id>

<!-- Tool Translation stripped Claude-Code-only construct:
     original `!`bd show "$TICKET_ID"`` block — port manually; OpenCode supports `!`cmd``
     under different semantics, see https://opencode.ai/docs/commands/ -->
...
```

Same two transformations as Cursor: `` !`...` `` blocks STRIPPED with placeholder comments (OpenCode's bash semantics are similar but not identical to Claude Code's, and quoting rules differ — strip is safer than translate); `${CLAUDE_SKILL_DIR}/../templates/<file>.md` rewritten to `@../../.alto/templates/<file>.md` per OpenCode's `@filename` convention. `disable_model_invocation` is dropped (OpenCode has no exact equivalent — `subtask: true` is the closest but means something different).

Per [OpenCode docs, "Commands"](https://opencode.ai/docs/commands/) — "Custom commands are stored in `.opencode/commands/` directory ... filename becomes the command name."

#### Normative cross-cutting rules (security + portability)

These rules apply to every shipped asset and to every Tool Translation adapter. They derive from the Skills doc's documented behaviour and from the white-hacker security review findings absorbed into this round.

1. **Bash substitution is Claude-Code-only.** `` !`<command>` `` blocks (and the multi-line ` ```! ` fenced variant) run only in Claude Code per [Skills doc, "Inject dynamic context"](https://code.claude.com/docs/en/skills). Tool Translation adapters for Cursor / Roo Code / OpenCode MUST strip every such block and replace it with an HTML-comment placeholder naming the original command (e.g. `<!-- Tool Translation stripped !`bd show "$TICKET_ID"` — port manually -->`) so downstream users can port manually. See Follow-up #3 AC for the testdata fixture demonstrating this across all three non-Claude adapters.
2. **`Bash` + side effects requires `disable_model_invocation: true`.** Per the `Bash`-with-`parameters` rule in the schema above, any command whose `tools_required` includes `Bash` AND whose body produces side effects (file writes, network calls, `bd update`, `git`, `kubectl`, etc.) MUST set `disable_model_invocation: true`. This prevents the model from inferring an invocation from user phrasing and routing untrusted parameter text into a shell. Doc-health (Follow-up #5) enforces this; a violation is an ERROR not a warning.
3. **`disable_model_invocation` is dropped silently by non-Claude adapters.** Cursor, Roo Code, and OpenCode have no equivalent gating field. Tool Translation MUST drop the flag along with the bash blocks; users of those tools who require manual-invoke-only semantics should add tool-native equivalents (e.g. omitting the rule from `alwaysApply: true` in Cursor) — the adapter cannot guarantee this. See Follow-up #3 AC.
4. **No `$ARGUMENTS` inside `` !`...` `` blocks.** Per [Skills doc, "Pass arguments to skills"](https://code.claude.com/docs/en/skills), `$ARGUMENTS` is substituted before bash blocks execute, but allowing unquoted parameter text into a shell is a shell-injection vector. Doc-health rejects any source asset containing this combination.

### Generic vs alto-specific split — worked examples

**`groom.md` (GENERIC, ships in scaffold):**

The current Phase 4 wording is rewritten to drop alto's `internal/composition/` path. The generic version reads (in place of current lines 84-99):

```markdown
#### 4a. Read every referenced source file

For EVERY file the ticket mentions or depends on, use the Read tool:
- Port interfaces (`{application_layer_path}/ports.{ext}` or handler-local interfaces)
- Existing handlers in the same bounded context (pattern reference)
- Domain types used in signatures
- Infrastructure adapters
- Composition root (project-specific path; see project overlay)

#### 4b. Trace the constructor chain

For each new construct the ticket creates, write out the full chain. The exact file paths
are project-specific — see `groom.project.md` for this project's layout.
```

**`groom.project.md` (OVERLAY, alto-only, NOT shipped):**

```markdown
# /groom — alto-specific addenda

## Composition root paths

- Port interfaces: `internal/{context}/application/ports.go`
- Adapters: `internal/{context}/infrastructure/{adapter}.go`
- Composition: `internal/composition/app.go`, `internal/composition/adapters.go`

## Ticket ID format

alto tickets use the `alto-` / `alty-cli-` prefix (e.g. `alto-0m9.2`, `alty-cli-766.1`).

## alto-specific verification

```bash
alto ticket-verify <ticket-id>      # detects quantitative claims, verifies against command output
```
```

**`brainstorm.md` (GENERIC, ships in scaffold):**

- Drop the `name: alto-brainstorm` namespace; rename to `name: brainstorm`.
- Drop the heading `/alto-brainstorm`; rename to `/brainstorm`.
- Phase 5 paths become parameterised: instead of `docs/PRD.md`, use `{{ARTIFACTS_DIR}}/PRD.md` (default `docs/`, overridable per project).
- Template references become `${CLAUDE_SKILL_DIR}/../templates/PRD_TEMPLATE.md` for the Claude Code view; Tool Translation rewrites the path per tool per the path-substitution table in §Frontmatter Schema below. No `${ALTO_ROOT}` token and no cwd-relative `../templates/` path appears in any shipped asset — both were rejected as non-portable.

**`brainstorm.project.md` (OVERLAY, alto-only):**

```markdown
# /brainstorm — alto-specific addenda

## Alto namespace

This command is also invocable as `/alto-brainstorm` for users who want the alto-branded
verb. Both resolve to the same skill.

## Artifact destinations

- PRD → `docs/PRD.md`
- DDD → `docs/DDD.md`
- ARCHITECTURE → `docs/ARCHITECTURE.md`
```

**`developer.md` (GENERIC):**

The current lines 70 (Watermill), 75 (`internal/`), 86-89 (`cmd/alto/`), 92-114 (Go value-object example), 118-131 (Go error patterns), 134-159 (Go test patterns), 162-167 (interface compliance), 169-194 (golangci-lint v2 table) all carry hard Go coupling. The generic version retains the **language-neutral DDD/TDD/SOLID guidance** and replaces Go-specific code with placeholders + a pointer to `developer.project.md`. For example, lines 75-89 (DDD source layout) become:

```markdown
## DDD Source Layout

```
src/                        # adjust per language conventions
├── {context}/              # one directory per bounded context
│   ├── domain/             # core business logic (ZERO external deps)
│   ├── application/        # use cases, command/query handlers, ports
│   └── infrastructure/     # adapters for external concerns
├── shared/domain/          # shared kernel across contexts
```

For this project's exact layout, see `developer.project.md`.
```

**`developer.project.md` (OVERLAY, alto-only):**

```markdown
# Developer — alto Go addenda

## Project: Go 1.26+

## Source layout

```
internal/
├── {context}/
│   ├── domain/
│   ├── application/
│   └── infrastructure/
├── shared/domain/
└── ...
cmd/
├── alto/main.go            # CLI entry (Cobra)
└── alto-mcp/main.go        # MCP server entry
```

## Quality gates

```bash
go build ./...
go test ./... -v -race
go vet ./...
golangci-lint run
```

## Idiomatic Go DDD patterns

<inlines the current developer.md Value Object example, error handling, testify idioms,
interface compliance markers, golangci-lint v2 table — current developer.md lines 92-194>
```

### Migration script outline

The migration is one shell script run from repo root. Verbs: `mkdir -p`, `git mv`, `mv` (for splits where the target doesn't yet exist), `sed -i.bak` + `diff -u` for the 3 root edits, `ln -s` (with pre-existence check) for symlink bridges. Steps are ordered so each step's preconditions are met by prior steps; refuse-on-exist guards prevent silent overwrites (see Follow-up #1 AC for the binding behaviour spec).

```bash
#!/usr/bin/env bash
# Migration: .claude/* + docs/templates/* + docs/beads_templates/* + docs/spikes/* → .alto/
# Run from alto repo root. Idempotent failure handling NOT included — run on a clean branch.

set -euo pipefail

# Phase 1 — create the new root
mkdir -p .alto/{commands,agents,templates,skills,lifecycle/in-progress,lifecycle/deprecated}
touch .alto/skills/.gitkeep .alto/lifecycle/in-progress/.gitkeep .alto/lifecycle/deprecated/.gitkeep

# Phase 2 — move generic-as-is files
git mv .claude/agents/researcher.md       .alto/agents/researcher.md
git mv .claude/agents/white-hacker.md     .alto/agents/white-hacker.md
git mv .claude/commands/architecture-docs.md  .alto/commands/architecture-docs.md
git mv .claude/commands/design-ticket.md      .alto/commands/design-ticket.md
git mv .claude/commands/doc-health.md         .alto/commands/doc-health.md
git mv .claude/commands/review.md             .alto/commands/review.md
git mv docs/templates/PRD_TEMPLATE.md         .alto/templates/PRD_TEMPLATE.md
git mv docs/templates/DDD_STORY_TEMPLATE.md   .alto/templates/DDD_STORY_TEMPLATE.md
git mv docs/templates/ARCHITECTURE_TEMPLATE.md .alto/templates/ARCHITECTURE_TEMPLATE.md
git mv docs/beads_templates/beads-epic-template.md    .alto/templates/beads-epic-template.md
git mv docs/beads_templates/beads-ticket-template.md  .alto/templates/beads-ticket-template.md
git mv docs/beads_templates/beads-spike-template.md   .alto/templates/beads-spike-template.md
git mv docs/beads_templates/beads-stub-template.md    .alto/templates/beads-stub-template.md
git mv docs/spikes/ddd_reference.md           .alto/templates/ddd_reference.md

# Phase 3 — split commands into generic + .project.md
# For each: move original to .alto/commands/<name>.md, then extract overlay into .project.md
git mv .claude/commands/brainstorm.md       .alto/commands/brainstorm.md
# manual edit: extract alto-namespace + docs/* paths to .alto/commands/brainstorm.project.md
git mv .claude/commands/groom.md            .alto/commands/groom.md
# manual edit: extract internal/composition/* refs to .alto/commands/groom.project.md
git mv .claude/commands/launch-team.md      .alto/commands/launch-team.md
# manual edit: extract alty-cli- + Go gates to .alto/commands/launch-team.project.md
git mv .claude/commands/prd-traceability.md .alto/commands/prd-traceability.md
# manual edit: extract C1-C25 table to .alto/commands/prd-traceability.project.md;
#              leave the RLM pattern in the generic file.

# Phase 3b — split agents
git mv .claude/agents/developer.md       .alto/agents/developer.md
# extract Go specifics to developer.project.md
git mv .claude/agents/tech-lead.md       .alto/agents/tech-lead.md
# extract Go grep recipes + arch-go to tech-lead.project.md
git mv .claude/agents/project-manager.md .alto/agents/project-manager.md
# extract internal/{context}/* + cmd/alto/* to project-manager.project.md
git mv .claude/agents/qa-engineer.md     .alto/agents/qa-engineer.md
# extract go test recipes to qa-engineer.project.md

# Phase 3c — split mild-overlay templates
# (ARCHITECTURE_TEMPLATE has Python residue → split out a .project.md with Go variant)

# Phase 4 — write .alto/CONTEXT.md (new file, no git mv)
# Refuse-if-exists (matches Follow-up #1 AC "refuses to overwrite"), tempfile + atomic rename.
if [[ -e .alto/CONTEXT.md ]]; then
  echo "ERROR: .alto/CONTEXT.md already exists — refusing to overwrite. Remove or pass --force." >&2
  exit 1
fi
CONTEXT_TMP="$(mktemp .alto/.CONTEXT.md.XXXXXX)"
cat > "$CONTEXT_TMP" <<'EOF'
# .alto/ Scaffold — Ubiquitous Language

(Glossary of terms used across this scaffold's commands, agents, and templates.)
EOF
mv -n "$CONTEXT_TMP" .alto/CONTEXT.md   # -n: never overwrite an existing file (belt-and-braces vs the test above)

# Phase 5 — preserve existing slash-command invocations
# Option A (interim): symlink .claude/commands/* → ../../.alto/commands/*.md
# Use `ln -s` (not -sf) + pre-existence check; fail-loud on collision (matches Follow-up #1 AC "symlink safety").
mkdir -p .claude/commands
for f in architecture-docs brainstorm design-ticket doc-health groom launch-team prd-traceability review; do
  target=".claude/commands/${f}.md"
  if [[ -e "$target" || -L "$target" ]]; then
    echo "ERROR: $target already exists (regular file or symlink). Aborting — no silent overwrite." >&2
    exit 1
  fi
  ln -s "../../.alto/commands/${f}.md" "$target"
done
# Option B (cleaner, requires Claude Code v2.1+): register .alto/commands/ as additional-dir
# echo '{"additionalDirectories": [".alto"]}' >> .claude/settings.json
# (deferred to alto init --with-scaffold ticket; not required for repo dogfooding)

# Phase 6 — root edits (all sed rewrites produce .bak backups and emit diffs uniformly)
# Edit .claude/CLAUDE.md: change "docs/beads_templates/" → ".alto/templates/" (5 occurrences)
sed -i.bak 's|docs/beads_templates/|.alto/templates/|g' .claude/CLAUDE.md
diff -u .claude/CLAUDE.md.bak .claude/CLAUDE.md   # emit per-rewrite diff to stdout
# Edit .claude/CLAUDE.md: change "docs/templates/" → ".alto/templates/" (3 occurrences)
sed -i.bak 's|docs/templates/|.alto/templates/|g' .claude/CLAUDE.md
diff -u .claude/CLAUDE.md.bak .claude/CLAUDE.md   # emit per-rewrite diff to stdout
# Edit bin/bd-ripple lines 191 + 238: template path (2 occurrences in one file)
sed -i.bak 's|docs/beads_templates/beads-ticket-template.md|.alto/templates/beads-ticket-template.md|g' bin/bd-ripple
diff -u bin/bd-ripple.bak bin/bd-ripple   # emit per-rewrite diff to stdout

# Phase 7 — verify
ls .alto/commands/    # expect 8 .md files + 5 .project.md siblings
ls .alto/agents/      # expect 6 .md files + 4 .project.md siblings
ls .alto/templates/   # expect 8 .md files + 1 .project.md sibling
ls .alto/lifecycle/in-progress .alto/lifecycle/deprecated  # both empty + .gitkeep
go run ./cmd/alto doc-health  # MUST pass after migration
```

**Diff estimate:** 25 files moved + 11 splits (where the `.project.md` sibling didn't exist before) + 3 sed edits + 1 new `.alto/CONTEXT.md` = ~40 file-level changes. Plus the symlink bridge (8 symlinks under `.claude/commands/`). Net line-count delta in repo: roughly neutral (splits remove from one file and add to another). The follow-up ticket alty-cli-766's "refactor" child owns this script.

### Shipping mechanism — `alto init --with-scaffold`

**Mechanism:** Go `embed.FS` of the `.alto/` tree, written to target project at `$TARGET/.alto/`, with post-extraction templating of placeholders.

**Why not the alternatives:**

- `alto scaffold add` (post-hoc command) — UX bloat; `init` already covers greenfield, `init --existing` covers brownfield. A third command splits responsibility unnecessarily.
- `npx alto-scaffold` (Node-based installer, mattpocock's pattern from [skills README](https://github.com/mattpocock/skills) — `npx skills@latest add ...`) — hard Node runtime dependency for a Go-distributed user. License vetting required on the npm package per CLAUDE.md global rule (LLM-knowledge-verification: any external library = SPDX check). Secondary distribution channel violates "alto is the architect; one binary" principle.

**Parameters (all required at invoke time; defaults documented):**

| Parameter | Type | Default | Source |
|-----------|------|---------|--------|
| `--project-name` | string | basename of target dir | `{{PROJECT_NAME}}` |
| `--ticket-prefix` | string | `proj-` | `{{TICKET_PREFIX}}` |
| `--issue-tracker` | enum: `beads \| github \| linear` | `beads` | `{{ISSUE_TRACKER}}` (only `beads` shipped per epic scope) |
| `--bounded-contexts` | comma-separated list | empty (filled by `/brainstorm`) | `{{BOUNDED_CONTEXTS}}` |
| `--primary-tool` | enum: `claude \| cursor \| roo \| opencode` | `claude` | `{{PRIMARY_TOOL}}` (drives which view dir Tool Translation emits) |

**DDD layer placement:** the embed FS + copy logic lives in `internal/bootstrap/infrastructure/scaffold_writer.go` implementing a port `internal/bootstrap/application/ScaffoldWriter` (define alongside existing Bootstrap ports). The handler that orchestrates `init --with-scaffold` lives in `internal/bootstrap/application/`. The domain layer (entities) is not affected — scaffolding is an infrastructure concern. Per CLAUDE.md layer rules ([CLAUDE.md `Layer Rules` table](<repo-root>/.claude/CLAUDE.md) ~line 148), `embed.FS` is allowed only in the infrastructure layer; the application port hides it behind an interface.

**Tool Translation extension (follow-up ticket, NOT this spike):** the existing `ConfigGeneration` port in `internal/tooltranslation/application/ports.go:13-16` already accepts `tools []ttdomain.SupportedTool`. The follow-up ticket adds three adapters under `internal/tooltranslation/infrastructure/` (currently empty per the audit above):
- `CursorAdapter` — reads `.alto/commands/*.md`, emits `.cursor/rules/*.mdc` per Cursor's `.mdc` frontmatter contract.
- `RooCodeAdapter` — reads `.alto/commands/*.md`, emits `.roo/customModes.yaml` per Roo Code's customModes schema.
- `OpenCodeAdapter` — reads `.alto/commands/*.md`, emits `.opencode/commands/*.md` per OpenCode's command contract.

All three are adapters in the infrastructure layer implementing the existing application-layer port — no port redesign. Compile-time satisfaction asserted via `var _ application.ConfigGeneration = (*CursorAdapter)(nil)` etc.

### Security & shipping notes (for white-hacker handoff)

1. `${CLAUDE_SKILL_DIR}` is the documented Claude-Code-portable path mechanism for resolving bundled scripts ([Skills doc, "Available string substitutions"](https://code.claude.com/docs/en/skills) table row "${CLAUDE_SKILL_DIR}"). The scaffold MUST use this for any script references — never absolute paths, never user-home-relative paths.
2. `$ARGUMENTS` substitution is shell-style — multi-word values require quotes ([Skills doc, "Pass arguments to skills"](https://code.claude.com/docs/en/skills)). Any frontmatter `parameters:` list MUST be documented as requiring shell-style quoting; downstream installer MUST refuse argument names that match shell metachars (`$`, `` ` ``, `;`, `|`, `&`, `<`, `>`, `(`, `)`).
3. `alto init --with-scaffold` MUST NOT silently overwrite an existing `.alto/`. Default behaviour: refuse if `.alto/` exists; require `--force` to overwrite; print a diff preview before overwriting per the existing `alto init` preview pattern ([PRD C4 capability](<repo-root>/docs/PRD.md), referenced from `.claude/commands/prd-traceability.md:52`).
4. No secrets, API keys, or credentials are templated into any scaffold file. Confirmed: zero `password`, `secret`, `api_key`, `token` references in the 25 in-scope files (greppable via `grep -rE 'password|secret|api.?key|token' .alto/`). Doc-health follow-up ticket adds this grep as a fitness check.
5. The `alto-` / `alty-cli-` ticket prefix and the C1-C25 capability table are alto-internal — these are split into `.project.md` overlays per the audit, so downstream users never see them in the shipped scaffold.

---

## References

- [Claude Code: Extend Claude with skills](https://code.claude.com/docs/en/skills) — directory-per-skill model, YAML frontmatter (`name`, `description`, `when_to_use`, `allowed-tools`, `disable-model-invocation`, `paths`, `context: fork`, `agent`), 500-line `SKILL.md` soft cap, `${CLAUDE_SKILL_DIR}` substitution, `.claude/commands/` merged into `.claude/skills/`, command-name-from-directory resolution. (Quoted verbatim where used; full doc cached at `tool-results/toolu_01117rcq5jDn4uSFbGH8ASqp.txt`.)
- [Claude Code: Common workflows](https://code.claude.com/docs/en/common-workflows) — `claude --permission-mode plan`, subagent delegation, `claude --worktree`, headless `claude -p` for CI; relevant to the migration ticket's CI fitness check.
- [Building Better Tech: I read the Claude Code source code](https://buildingbetter.tech/p/i-read-the-claude-code-source-code) — `omitClaudeMd: true`, memory scopes `user` / `project` / `local`, `context: fork` cache contracts. The `omitClaudeMd` mechanism is relevant for tool-translation: when alto emits views for Explore / Plan subagents, the overlay `.project.md` is INTENTIONALLY skipped (subagent runs "fresh-eyes" against the generic content), preserving the cache.
- [mattpocock/skills](https://github.com/mattpocock/skills) — skills root layout (`skills/{engineering,productivity,misc,personal,in-progress,deprecated}/`), `CONTEXT.md` shared-language doc ("helps agents decode the jargon used in the project"), `/write-a-skill` meta-skill, `npx skills@latest add ...` installer pattern, MIT license. alto borrows: the `CONTEXT.md` role, the `in-progress/` and `deprecated/` lifecycle folders, the meta-skill concept (renamed `/write-a-workflow-asset`). alto rejects: the `npx` installer, the category-by-domain split.
- **Cursor docs: Rules** — fetched 2026-05-29. Live URL: <https://cursor.com/docs/context/rules>. Wayback-pinned snapshot: <https://web.archive.org/web/20260529000000*/cursor.com/docs/context/rules>. Establishes `.cursor/rules/*.mdc` filename requirement and frontmatter fields `description` / `globs` / `alwaysApply`. Plain `.md` files in `.cursor/rules/` are ignored.
- **Roo Code docs: Custom modes** — fetched 2026-05-29. Live URL: <https://roocodeinc.github.io/Roo-Code/features/custom-modes>. Repo-hosted (GitHub Pages); permalink alternative: <https://github.com/RooCodeInc/Roo-Code> → `docs/features/custom-modes.md` (commit-SHA-pinned in the Follow-up #3 schema fixture). Establishes `customModes` array with `slug`, `name`, `description`, `roleDefinition`, `customInstructions`, `groups`, `whenToUse`, `source`; YAML preferred over JSON; `.roomodes` per-project, `custom_modes.yaml` global.
- **OpenCode docs: Commands** — fetched 2026-05-29. Live URL: <https://opencode.ai/docs/commands/>. Wayback-pinned snapshot: <https://web.archive.org/web/20260529000000*/opencode.ai/docs/commands/>. Establishes `.opencode/commands/*.md` per-project, `~/.config/opencode/commands/*.md` global; filename = command name; frontmatter `description`, `agent`, `model`, `subtask`.

**Pinning rationale.** The Wayback URLs above are best-effort references to the upstream specs as of 2026-05-29; specific snapshot timestamps are NOT chased (sources may be unreachable, captures may be missing on any given day, and Wayback's `*` query form intentionally avoids hard-coding a timestamp that may not resolve). The binding artefact is the schema fixture committed under `internal/tooltranslation/infrastructure/testdata/schemas/` (Follow-up #3 AC); the CI fitness check is what fails on drift. Roo Code docs ARE GitHub-hosted (`RooCodeInc/Roo-Code` repo, `docs/` directory) — the schema fixture for Roo records the upstream `docs/features/custom-modes.md` content at a specific commit SHA in the fixture file header. Cursor and OpenCode docs are vendor-hosted without commit-SHA permalinks; their schema fixtures contain the spec text verbatim with a `Fetched: 2026-05-29` header and the live URL.
- Local audit targets (read at `file:line` granularity, every claim cited above):
  - `<repo-root>/.claude/agents/{researcher,developer,tech-lead,project-manager,qa-engineer,white-hacker}.md`
  - `<repo-root>/.claude/commands/{architecture-docs,brainstorm,design-ticket,doc-health,groom,launch-team,prd-traceability,review}.md`
  - `<repo-root>/.claude/skills/` (18 non-symlink + 14 gstack-symlink directories listed; confirmed out of alto-scaffold scope)
  - `<repo-root>/docs/templates/{PRD,DDD_STORY,ARCHITECTURE}_TEMPLATE.md`
  - `<repo-root>/docs/beads_templates/beads-{epic,ticket,spike,stub}-template.md`
  - `<repo-root>/docs/spikes/ddd_reference.md`
  - `<repo-root>/internal/tooltranslation/{domain,application,infrastructure}/` — application/ports.go:13-27 inspected; infrastructure/ confirmed empty
  - `<repo-root>/bin/bd-ripple` — template path references at lines 191 (REQUIRED_SECTIONS comment) and 238 (user-facing template hint) confirmed
  - `<repo-root>/.claude/CLAUDE.md` — layer rules + git rules + license-vetting rule referenced

---

## Follow-up Tasks

The spike does NOT file these — tech-lead does in Phase 6 of the spike's team workflow, after main-thread approval. Bodies are drafted below conforming to `docs/beads_templates/beads-ticket-template.md`. All MUST be filed as `bd create --parent=alty-cli-766 ...` (NOT standalone); dependencies wired via `bd dep add`.

### Follow-up #1 — Refactor: migrate scaffold assets to `.alto/`

**Type:** task · **Parent:** alty-cli-766 · **Depends-on:** alty-cli-766.1 (this spike)

**Goal / Problem.** Execute the 7-phase migration script (Phases 1–7 as labelled in this report's "Migration script outline" section) from this report's Recommendation section. Move 25 files into `.alto/`, split 11 alto-coupled files into generic + `.project.md` overlays, preserve every existing slash-command invocation via symlinks (interim) or `additionalDirectories` (permanent).

**DDD Alignment.** Bounded context: none (repo housekeeping); affects all contexts insofar as their scaffold lives under `.claude/`. Layer: not applicable.

**Design.** Use the migration script from this report Section "Migration script outline". Each `git mv` is a separate commit; each split is a separate commit (one before-state + one after-state with the `.project.md` extracted). `bin/bd-ripple` + `.claude/CLAUDE.md` edits are a final cleanup commit.

**Steps.**
1. Create `.alto/` skeleton + `.gitkeep` files (Phase 1).
2. Move generic-as-is files (Phase 2) — 14 files.
3. Split alto-coupled commands and agents (Phase 3) — 11 source files → 22 result files (11 `.md` + 11 `.project.md`). Each split is one commit.
4. Write `.alto/CONTEXT.md` from scratch (Phase 4).
5. Add `.claude/commands/` symlinks pointing into `.alto/commands/` (Phase 5).
6. Edit `.claude/CLAUDE.md` and `bin/bd-ripple` paths (Phase 6).
7. Run `go run ./cmd/alto doc-health` — must pass.

**Acceptance Criteria.**
- [ ] **Preconditions.** Script refuses to run unless `git diff --quiet && git diff --staged --quiet` (clean working tree). Verified in step 0 of the script before any `mkdir`/`git mv`.
- [ ] **Refuses to overwrite.** Script refuses to run if `.alto/CONTEXT.md` already exists (or `.alto/` is non-empty). Operator must remove or pass `--force` (which itself blocks unless `--force --i-know-what-i-am-doing`).
- [ ] **`--dry-run` flag.** Emits the full ordered op list (every `git mv`, every split, every `sed` rewrite, every symlink) to stdout without touching the filesystem. Exit 0 on dry-run success.
- [ ] **Symlink safety.** Replace `ln -sf` with `ln -s` + pre-existence check: for each symlink target, if the path exists (whether file, dir, or existing symlink) the script aborts and prints the colliding path. No silent overwrites.
- [ ] **`sed` safety.** Replace `sed -i` with `sed -i.bak` (creates `.bak` backups) and emit `diff -u <file>.bak <file>` to stdout for each rewrite. Backups are committed to a separate `migration-backups/` directory at the end of the migration and listed in the final commit's body.
- [ ] `.alto/` tree exists with the structure shown in the spike's Recommendation.
- [ ] All 25 originally-located files are now under `.alto/` (via `git log --follow`).
- [ ] Every `/groom`, `/design-ticket`, `/launch-team`, `/review`, `/doc-health` invocation works unchanged (verify by launching `claude` in repo root and running each).
- [ ] No file marked GENERIC contains references to `internal/`, `alto-`, `alty-cli-`, or alto bounded-context names (`Bootstrap`, `Discovery`, `DDDChallenge`, `Knowledge Base`, `Rescue`, `Ticket Pipeline`, `Tool Translation`, `Doc Health`).
- [ ] `bin/bd-ripple` template-path edit verified by closing a test ticket and observing the path in the generated review comment.
- [ ] `go build ./...` + `go vet ./...` + `golangci-lint run ./...` + `go test ./... -race` all green.

**Edge Cases.** Symlink resolution on Windows (Claude Code on Windows may not resolve POSIX symlinks — document the `additionalDirectories` settings.json alternative). Pre-existing user customisations in `.claude/commands/*.md` would be lost on symlink — document migration warning. Pre-existing `.alto/` content from a prior aborted migration — refuse-and-document; require operator to clear or use `--force --i-know-what-i-am-doing`.

**Quality Gates.** `go build ./...`, `go vet ./...`, `golangci-lint run ./...`, `go test ./... -race`, plus a new fitness assertion: `grep -rE '\binternal/|\balto-|\balty-cli|cmd/alto' .alto/commands/*.md .alto/agents/*.md` returns zero results.

**Risks / Dependencies.** Blocking: alty-cli-766.1 closed (this spike). Risk: symlinks break on Windows users; mitigated by Phase 5 Option B (`additionalDirectories`). Risk: `sed -i.bak` creates backups that pollute the working tree if the migration is re-run; mitigated by the clean-working-tree precondition.

### Follow-up #2 — `alto init --with-scaffold` (Go embed copy + templating)

**Type:** task · **Parent:** alty-cli-766 · **Depends-on:** follow-up #1

**Goal / Problem.** Ship the `.alto/` scaffold into user projects via `alto init --with-scaffold`. Use Go `embed.FS` of the in-repo `.alto/` tree; template the 5 parameters listed in this report.

**DDD Alignment.** Bounded context: **Bootstrap** (per epic alty-cli-766's DDD Alignment table). Layer: port in `internal/bootstrap/application/`, adapter in `internal/bootstrap/infrastructure/`. The handler is application-layer; `embed.FS` is infrastructure-only — domain layer untouched.

**Design.**
- New port: `internal/bootstrap/application/ScaffoldWriter` — `WriteScaffold(ctx, targetDir string, params ScaffoldParams) error`.
- New adapter: `internal/bootstrap/infrastructure/embed_scaffold_writer.go` — implements `ScaffoldWriter` via `embed.FS`.
- New value object: `internal/bootstrap/domain/ScaffoldParams` — immutable VO carrying the 5 parameters with validation in `NewScaffoldParams(...)`.
- Handler extension: existing init handler gains `WithScaffold bool` field; when true, calls `ScaffoldWriter.WriteScaffold(...)`.
- CLI flag: `alto init --with-scaffold --project-name=X --ticket-prefix=Y --issue-tracker=beads --bounded-contexts=A,B --primary-tool=claude`.

**Steps.**
1. RED: write failing test in `internal/bootstrap/domain/scaffold_params_test.go` — `NewScaffoldParams("", ...)` returns `ErrInvariantViolation`.
2. GREEN: implement `ScaffoldParams` VO.
3. RED: write failing test in `internal/bootstrap/application/scaffold_handler_test.go` — handler calls `ScaffoldWriter.WriteScaffold` once.
4. GREEN: implement port + handler wiring.
5. RED: write failing test in `internal/bootstrap/infrastructure/embed_scaffold_writer_test.go` — writes 25 files to a tempdir with substitutions applied.
6. GREEN: implement `embed.FS` + `text/template` substitution.
7. Wire into `cmd/alto/init.go` (Cobra flag).
8. Integration test: `go run ./cmd/alto init --with-scaffold` in a tempdir produces a working `.alto/` with all parameters substituted.

**Acceptance Criteria.**
- [ ] `alto init --with-scaffold ...` produces a `.alto/` tree in the target directory.
- [ ] All 5 parameters are substituted in templated files (verifiable by grepping for `{{PROJECT_NAME}}` etc — must be zero).
- [ ] Refusal-with-`--force` semantics for pre-existing `.alto/` (security AC from the white-hacker handoff).
- [ ] `var _ application.ScaffoldWriter = (*infrastructure.EmbedScaffoldWriter)(nil)` compile-time assertion present.
- [ ] **Input validation — `--project-name`.** Rejects `/`, `\`, `..`, NUL (`\x00`), and any value resolving outside the cwd. Implementation in `NewScaffoldParams(...)` constructor; tested with table-driven negative cases for each character class.
- [ ] **Input validation — `--ticket-prefix`.** Must match `^[a-zA-Z][a-zA-Z0-9-]*-$`. Tested with negatives: empty, leading digit, embedded `/`, embedded `$`, missing trailing `-`.
- [ ] **Input validation — `--bounded-contexts`.** Each comma-split entry must match `^[A-Z][a-zA-Z0-9]*$` (PascalCase domain term). Tested with negatives: lowercase first letter, dash, space, dot.
- [ ] **Template-parameter sanitisation.** Before substitution into any embedded asset, every value must be rejected if it contains any of: `$`, `` ` ``, `;`, `|`, `&`, `<`, `>`, `(`, `)`, or newline (`\n`/`\r`). Sentinel error: `ErrUnsafeTemplateParameter`. Rejection happens in `NewScaffoldParams(...)` — the embed writer never sees an unsafe value.
- [ ] **Embed exclusions.** `//go:embed` directive lists `.alto/{CONTEXT.md,commands,agents,templates,skills}` explicitly — it does NOT include `.alto/lifecycle/in-progress/**` or `.alto/lifecycle/deprecated/**`. Lifecycle content reaches the embed only via opt-in release tooling (separate ticket if needed); test asserts the embed file count matches the GENERIC + OVERLAY count from this spike's audit (25 source files).
- [ ] Quality gates green.

**Threat model.** Two attacker-controllable surfaces exist: the CLI flags themselves (operator-controlled, low risk on developer workstation; medium risk in CI pipelines that template these values from external sources) and downstream files in the target directory (`alto init --with-scaffold` could be tricked into writing through symlinks or into parent directories). Mitigations: (a) shell-metachar rejection in `NewScaffoldParams` closes the shell-injection vector that would otherwise let `{{PROJECT_NAME}}` substitution emit shell-executable text into a Bash-substituted block (cross-references `bash_substitution_policy` in the schema); (b) path-traversal rejection in `--project-name` (no `/`, no `..`) closes the directory-escape vector; (c) refusal-to-overwrite-existing-`.alto/` closes the silent-clobber vector (already an AC above). Out of scope for this ticket: kernel-level symlink races (TOCTOU) — `embed.FS` writes via `os.OpenFile(..., O_CREATE|O_EXCL, 0644)` (0644 is then masked by process umask; 0600 is wrong for scaffold content read by IDE indexers, file watchers, and other repo collaborators) mitigates but does not eliminate; document as known limitation.

**Quality Gates.** Standard four-gate suite plus: domain layer of `internal/bootstrap/domain/` has zero external imports (verified by `arch-go`).

**Risks / Dependencies.** Blocking: follow-up #1 closed (`.alto/` must exist to embed). Risk: `embed.FS` does not preserve symlinks — confirm the `.claude/commands/` symlinks aren't part of the embed; embed contents follow the allow-list per the **Embed exclusions** AC item above (excludes `lifecycle/in-progress/` and `lifecycle/deprecated/`).

### Follow-up #3 — Tool Translation extension for Cursor / Roo Code / OpenCode

**Type:** task · **Parent:** alty-cli-766 · **Depends-on:** follow-up #1

**Goal / Problem.** Implement three adapters under `internal/tooltranslation/infrastructure/` that emit Cursor `.mdc`, Roo Code `customModes`, and OpenCode `.md` command files from the canonical `.alto/commands/*.md` source. Reuse the existing `ConfigGeneration` port in `internal/tooltranslation/application/ports.go:13-16` — no port redesign.

**DDD Alignment.** Bounded context: **Tool Translation**. Layer: adapters in `infrastructure/`. Per CLAUDE.md enforced principles + the brief's Port/Adapter rule, this MUST use the existing application port; the spike forbids redesigning it.

**Design.**
- Adapter 1: `internal/tooltranslation/infrastructure/cursor_adapter.go` — implements `application.ConfigGeneration`. Reads `.alto/commands/*.md`, emits `.cursor/rules/*.mdc` with `description` / `globs` / `alwaysApply` frontmatter.
- Adapter 2: `internal/tooltranslation/infrastructure/roo_code_adapter.go` — emits `.roo/customModes.yaml` with the array of `slug` / `name` / `description` / `roleDefinition` / `customInstructions` / `groups` / `whenToUse` entries.
- Adapter 3: `internal/tooltranslation/infrastructure/opencode_adapter.go` — emits `.opencode/commands/*.md`.
- A new domain helper `tool_adapter.go` already exists per the audit (18k, with tests); the three new adapters extend its `SupportedTool` enum and registration.

**Steps.** Per adapter: RED (table-driven test asserting emitted file structure for `.alto/commands/groom.md` + `groom.project.md`), GREEN (implement), REFACTOR (extract shared frontmatter-merging helper if duplication appears).

**Acceptance Criteria.**
- [ ] All three adapters implement `application.ConfigGeneration` (compile-time `var _ ` assertions).
- [ ] Each adapter's test asserts emitted file matches a fixture under `internal/tooltranslation/infrastructure/testdata/`.
- [ ] Generated Cursor `.mdc` validates against Cursor's MDC schema (frontmatter present with `description` / `globs` / `alwaysApply`).
- [ ] Generated Roo Code YAML parses as valid Roo `customModes` array.
- [ ] Generated OpenCode `.md` includes `description` + `agent` frontmatter.
- [ ] **Name safety.** Every emitted file's basename derives from the source `name:` frontmatter field. The adapter MUST validate `name` matches `^[a-z][a-z0-9-]*$` (Fix #5 alignment with doc-health) before emit. Names failing the regex cause a hard error — the adapter does not attempt sanitisation.
- [ ] **Path-traversal defence.** Every emitted file path passes through `filepath.Clean` and a `strings.HasPrefix(cleaned, outputRoot)` check after cleaning. Reject any path whose cleaned form contains `..` segments or escapes the configured output root. Sentinel error: `ErrPathTraversal`.
- [ ] **Atomic emit.** Each output file is written via tempfile + `os.Rename`: write to `<outputRoot>/.tmp-<random>-<basename>`, then `os.Rename` to the final path. Partial writes never leave broken files on disk.
- [ ] **Bash-block strip-with-placeholder test fixture.** Commit a fixture pair under `internal/tooltranslation/infrastructure/testdata/strip-bash-blocks/`: one input file `input/groom.md` containing two `` !`<cmd>` `` blocks and one `${CLAUDE_SKILL_DIR}/../templates/PRD_TEMPLATE.md` reference; three expected outputs under `expected/{cursor,roo,opencode}/` demonstrating (a) each `` !`...` `` block replaced with the HTML-comment placeholder naming the original command verbatim, (b) `disable_model_invocation: true` dropped silently (NOT emitted to any non-Claude view), (c) `${CLAUDE_SKILL_DIR}/../templates/<file>.md` rewritten per the per-tool path-substitution table. The test compares emitted bytes to the fixture byte-for-byte. Note in the test docstring: "adapters silently drop `disable_model_invocation` along with the bash blocks — non-Claude tools have no equivalent gating field; users requiring manual-invoke must add tool-native protections."
- [ ] **Schema fixtures committed under `internal/tooltranslation/infrastructure/testdata/schemas/`.** Each fixture is the upstream tool's documented schema (or a representative example) at a pinned version: `cursor-mdc.schema.md` (commit-SHA-pinned permalink), `roo-custom-modes.schema.md` (commit-SHA-pinned permalink), `opencode-commands.schema.md` (commit-SHA-pinned permalink). A new fitness check (run in CI as part of the test suite) refetches each upstream and fails the build if the schema diverges from the fixture — this surfaces upstream drift loudly rather than emitting broken output silently.
- [ ] Quality gates green.

**Risks / Dependencies.** Blocking: follow-up #1 (must read from `.alto/commands/`). Risk: Cursor / Roo / OpenCode schemas drift; mitigated by committed schema fixtures + CI drift check (AC item above). Risk: Wayback / commit-SHA permalinks rot — re-pin during quarterly maintenance.

### Follow-up #4 — Authoring meta-skill `/write-a-workflow-asset`

**Type:** task · **Parent:** alty-cli-766 · **Depends-on:** follow-up #1

**Goal / Problem.** Borrow mattpocock/skills' `/write-a-skill` pattern: a meta-skill that walks a contributor through authoring a new workflow asset (command / agent / template), enforces the YAML frontmatter schema from this spike, and writes the file in the right `.alto/` subdir + lifecycle folder.

**DDD Alignment.** Bounded context: **Tool Translation** (the schema knowledge belongs to translation). Layer: command body is a generic .md scaffold; no production Go code unless the meta-skill needs validation logic in `internal/tooltranslation/application/`.

**Design.** The meta-skill is itself a `.alto/commands/write-a-workflow-asset.md` file with frontmatter `kind: command`, `phase: design`. Body prompts the contributor for `name`, `kind`, `phase`, `when_to_use`, `tools_required`; validates against the schema in this report; writes to `.alto/lifecycle/in-progress/<name>.md`.

**Acceptance Criteria.**
- [ ] `/write-a-workflow-asset` invocation produces a valid frontmatter file under `.alto/lifecycle/in-progress/`.
- [ ] **User-supplied `<name>` safety.** The meta-skill validates the contributor-supplied `<name>` against `^[a-z][a-z0-9-]*$` BEFORE composing the output path. Names failing the regex are rejected with an explicit error message listing the allowed character class — the meta-skill does not silently sanitise (no `/` → `-`, no case-folding).
- [ ] **Path-traversal defence.** After resolving `.alto/lifecycle/in-progress/<name>.md`, the path is passed through `filepath.Clean` and verified to be a direct child of `.alto/lifecycle/in-progress/`. Reject any cleaned path containing `..` or whose parent differs from `.alto/lifecycle/in-progress/`.
- [ ] Doc-health (follow-up #5) flags the new file as "in progress" until it's moved to `.alto/commands/` or `.alto/agents/`.

### Follow-up #5 — `alto doc-health` validates `.alto/` scaffold root

**Type:** task · **Parent:** alty-cli-766 · **Depends-on:** follow-up #1

**Goal / Problem.** Extend `alto doc-health` to validate the `.alto/` tree: required frontmatter present per the schema; no `internal/` / `alto-` / `alty-cli-` leaks in GENERIC files; lifecycle folder occupancy; no orphaned `.project.md` siblings (every `.project.md` has a matching `.md`).

**DDD Alignment.** Bounded context: **Doc Health** (per epic table). Layer: handler in `internal/dochealth/application/`, frontmatter parser adapter in `internal/dochealth/infrastructure/`.

**Acceptance Criteria.**
- [ ] `alto doc-health` reports 0 violations on the migrated `.alto/`.
- [ ] Introducing a `.alto/commands/foo.md` without `name` frontmatter produces a violation.
- [ ] Introducing `internal/` in a GENERIC file produces a violation.
- [ ] Adding `.alto/commands/foo.project.md` without a matching `.alto/commands/foo.md` produces a violation.
- [ ] **Name regex enforcement.** Every GENERIC-tagged file's `name:` frontmatter field is validated against `^[a-z][a-z0-9-]*$` (matching Fix #5 in Follow-up #3). Violations are ERROR (not warning) — block on CI.
- [ ] **Secret-leak scan (expanded).** Reject any file body (outside frontmatter) matching the conservative Trivy-aligned superset, case-insensitive: `\b(password|secret|api[_-]?key|token|bearer|private_key|client_secret|jwt|credentials|aws_access_key)\b`, plus AWS access-key pattern `AKIA[0-9A-Z]{16}` and GitHub PAT pattern `gh[pousr]_[A-Za-z0-9]{36,}`. Frontmatter is excluded to allow descriptive fields like `description: "Manages API keys"` without false positives. The regex set is tunable via a config flag (`--secret-patterns=path/to/patterns.yaml`); the defaults above are the binding floor. Matches in body are ERROR.
- [ ] **`Bash`-with-`parameters` warning.** Warn (not error) on any asset whose `tools_required` contains `Bash` AND whose `parameters:` is declared AND whose `disable_model_invocation` is not `true`. This catches authors who forgot the normative rule from §Frontmatter Schema; the schema validator alone treats it as ERROR, but doc-health warns to give a softer path during initial migration. CI fails build only on ERROR; warnings are reported and tracked.
- [ ] **`bash_substitution_policy` validator.** Doc-health validates `bash_substitution_policy` field semantics per the §Frontmatter Schema rules: when `none`, reject ANY `` !`...` `` block in the body (ERROR); when `quoted`, reject any unquoted `$VAR` or `$ARGUMENTS` (or bracketed `$ARGUMENTS[N]`) reference inside `` !`...` `` blocks (ERROR); when `unrestricted`, warn-only (allowed but logged for review). Detection regexes cross-reference the rules below.
- [ ] **`$ARGUMENTS`-in-bash-block rejection (inline + fenced).** Reject any source asset where bash blocks contain unquoted `$ARGUMENTS` (bare), `$ARGUMENTS[N]` (bracketed positional form per [Claude Code Skills doc, "Available string substitutions"](https://code.claude.com/docs/en/skills) table row `$ARGUMENTS[N]`), `$N`, or `$<name>` references. Two regexes — both must run:
  - Inline form: `` !`[^`]*\$(ARGUMENTS(\[[0-9]+\])?|[0-9]+|[A-Za-z_]+)[^`]*` ``
  - Multi-line fenced form: blocks opened with ` ```! ` and closed with ` ``` ` containing `$ARGUMENTS`, `$ARGUMENTS\[[0-9]+\]`, `$N`, or `$<name>` not surrounded by double quotes. Per [Skills doc, "Inject dynamic context"](https://code.claude.com/docs/en/skills), multi-line commands use ` ```! ` fenced blocks — doc-health MUST inspect both forms.

  ERROR. Mitigation hint in the violation message: "wrap argument substitutions in double quotes inside the bash block, or move the bash call outside the block and use a separate validation step."
- [ ] **Path-substitution `..`-segment limit.** Doc-health rejects any path-substitution template reference inside `.alto/commands/*.md` containing more than 2 `..` segments. Two is the documented maximum (`${CLAUDE_SKILL_DIR}/../templates/`); any reference with 3 or more `..` segments is treated as an author-trust escape attempt. ERROR.

---

## Spike completion notes

- **Timebox respected:** investigation completed within the 6-hour budget. No extension justification needed.
- **External sources fetched (4/4):** Claude Code Skills doc, Claude Code Common workflows, Building Better Tech blog, mattpocock/skills GitHub. Cursor, Roo Code, OpenCode docs also fetched as required to spec the frontmatter portability.
- **Local audit targets read (8/8):** all six pre-flagged files plus all template directories plus `internal/tooltranslation/` plus `bin/bd-ripple`.
- **All 6 research questions answered explicitly** in the "Research Question" section above (one paragraph each).
- **Findings Template structure intact:** Summary / Research Question / Options Considered (pros-cons table) / Recommendation / References / Follow-up Tasks present.
- **No production Go code written** per spike rules.
- **No git commit or push performed** per CLAUDE.md global rules.
- **No `gh` CLI used** per CLAUDE.md private-Git-server rule.
- **All recommended external libraries** (`embed` — stdlib, no license cost) are permissive (Apache 2.0 / MIT / BSD / stdlib).
