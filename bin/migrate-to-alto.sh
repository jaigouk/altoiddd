#!/usr/bin/env bash
# migrate-to-alto.sh — Scaffold migration script for ticket alty-cli-766.2
#
# Migrates scaffold assets from `.claude/` + `docs/templates/` + `docs/beads_templates/`
# + `docs/spikes/` into the canonical `.alto/` root (flat-by-category) with alto-specific
# overlay siblings `.project.md`. Produces 17 commits in execution order.
#
# Binding plan: bd show alty-cli-766.2
# Binding spike: docs/research/20260529_workflow-scaffold-generic.md
#
# Usage:
#   bin/migrate-to-alto.sh --dry-run                              # Plan only, no mutation
#   bin/migrate-to-alto.sh --force --i-know-what-i-am-doing        # Mutating mode
#
# Safety:
#   * Refuses to run on a dirty working tree (git diff --quiet check)
#   * Refuses to overwrite a populated .alto/ (CONTEXT.md or commands/ present)
#   * Two-flag override required (BOTH --force AND --i-know-what-i-am-doing)
#   * Uses `sed -i.bak` for all in-place edits with backups committed in final commit
#   * Uses `ln -s` (not `ln -sf`) with pre-existence collision abort
#   * Stages specific files per commit; never `git add -A`

set -euo pipefail

# ---------- Flag parsing ----------
DRY_RUN=0
FORCE=0
CONFIRM=0

for arg in "$@"; do
    case "$arg" in
        --dry-run) DRY_RUN=1 ;;
        --force) FORCE=1 ;;
        --i-know-what-i-am-doing) CONFIRM=1 ;;
        -h|--help)
            cat <<'USAGE'
migrate-to-alto.sh — Scaffold migration to .alto/ root (alty-cli-766.2)

Usage:
  bin/migrate-to-alto.sh --dry-run
      Print the planned operations and exit. No filesystem changes.

  bin/migrate-to-alto.sh --force --i-know-what-i-am-doing
      Execute the migration. BOTH flags required to override safety gates.

Exit codes:
  0   success (dry-run or completed migration)
  1   refusal (missing override flags, dirty tree, populated .alto/)
  2   verification failure (post-migration checks failed; backups not committed)
USAGE
            exit 0
            ;;
        *)
            echo "ERROR: unknown argument: $arg" >&2
            echo "Run 'bin/migrate-to-alto.sh --help' for usage." >&2
            exit 1
            ;;
    esac
done

# Override semantics: BOTH --force AND --i-know-what-i-am-doing required for mutation.
# Single-flag invocations refuse (even when .alto/ is empty) so the operator confirms intent.
# --dry-run bypasses ALL preflight refusals and prints the plan without touching the filesystem.

# ---------- Console helpers ----------
say() {
    if [ "$DRY_RUN" -eq 1 ]; then
        printf '[DRY] %s\n' "$*"
    else
        printf '[ALTO] %s\n' "$*"
    fi
}

fail() {
    printf 'ERROR: %s\n' "$*" >&2
    exit 1
}

# ---------- Phase 0 — Preflight ----------
phase0_preflight() {
    say "Phase 0 — Preflight"

    if [ "$DRY_RUN" -eq 1 ]; then
        say "  skip clean-tree refusal (dry-run)"
        say "  skip .alto/{CONTEXT.md,commands/} refusal (dry-run)"
        say "  skip two-flag override refusal (dry-run)"
        return 0
    fi

    if [ "$FORCE" -ne 1 ] || [ "$CONFIRM" -ne 1 ]; then
        fail "override requires --force AND --i-know-what-i-am-doing (both flags, in any order)"
    fi

    if ! git diff --quiet; then
        fail "dirty working tree (unstaged changes present) — stash or commit first"
    fi
    if ! git diff --staged --quiet; then
        fail "staged changes present — commit or unstage first"
    fi

    if [ -e .alto/CONTEXT.md ] || [ -d .alto/commands ]; then
        fail ".alto/ already populated (CONTEXT.md or commands/ exists) — remove before re-running"
    fi

    say "  clean tree, .alto/ unpopulated, override flags set — proceed"
}

# ---------- commit_phase helper ----------
# Usage: commit_phase "<commit subject>" <file> [<file> ...]
# Stages explicit files (never -A / -.) and commits with the supplied subject.
commit_phase() {
    local subject="$1"
    shift
    if [ "$DRY_RUN" -eq 1 ]; then
        say "  git add ${*}"
        say "  git commit -m \"${subject}\""
        return 0
    fi
    git add -- "$@"
    git commit -m "$subject"
}

# ---------- Phase 1 — Create .alto/ skeleton ----------
# Commit #1: chore(scaffold): create .alto/ skeleton with .gitkeep files
phase1_skeleton() {
    say "Phase 1 — Create .alto/ skeleton + migration-backups/"

    if [ "$DRY_RUN" -eq 1 ]; then
        say "  mkdir -p .alto/{commands,agents,templates,skills,lifecycle/{in-progress,deprecated}}"
        say "  touch 3 .gitkeep files (skills, lifecycle/in-progress, lifecycle/deprecated)"
        say "  mkdir -p migration-backups/"
        say "  COMMIT #1: chore(scaffold): create .alto/ skeleton with .gitkeep files"
        return 0
    fi

    mkdir -p .alto/commands .alto/agents .alto/templates .alto/skills \
             .alto/lifecycle/in-progress .alto/lifecycle/deprecated
    touch .alto/skills/.gitkeep \
          .alto/lifecycle/in-progress/.gitkeep \
          .alto/lifecycle/deprecated/.gitkeep
    mkdir -p migration-backups
    commit_phase "chore(scaffold): create .alto/ skeleton with .gitkeep files" \
        .alto/skills/.gitkeep \
        .alto/lifecycle/in-progress/.gitkeep \
        .alto/lifecycle/deprecated/.gitkeep
}

# ---------- Phase 2 — git mv 14 + 4 OVERLAY-source files (18 total) ----------
# Commit #2: chore(scaffold): git-mv 14 generic assets into .alto/{agents,commands,templates}
#
# 14 GENERIC moves listed in the charter, PLUS the 4 OVERLAY command sources
# (brainstorm/groom/launch-team/prd-traceability) — both batches in one commit because
# the spike & ticket body both place them in Phase 2 (they need to live at .alto/commands/
# before being split in commits #3-6). The 5 OVERLAY agent sources move in the same commit
# for symmetry. ARCHITECTURE_TEMPLATE.md is already in the 14 GENERIC list (it splits in #13).
#
# Total git mv calls: 14 charter list + 4 OVERLAY commands + 5 OVERLAY agents = 23 moves.
# We override the charter subject's "14" wording because the binding ticket body Phase 2 says
# move generic AND pre-split (Phase 3 splits operate on .alto/* paths). The commit message
# keeps the charter wording verbatim per the orchestrator instruction.
phase2_moves() {
    say "Phase 2 — git mv 14 GENERIC files + pre-position 9 OVERLAY sources"

    # 14 GENERIC moves (charter list)
    local generic_pairs=(
        ".claude/agents/researcher.md|.alto/agents/researcher.md"
        ".claude/agents/white-hacker.md|.alto/agents/white-hacker.md"
        ".claude/commands/architecture-docs.md|.alto/commands/architecture-docs.md"
        ".claude/commands/design-ticket.md|.alto/commands/design-ticket.md"
        ".claude/commands/doc-health.md|.alto/commands/doc-health.md"
        ".claude/commands/review.md|.alto/commands/review.md"
        "docs/templates/PRD_TEMPLATE.md|.alto/templates/PRD_TEMPLATE.md"
        "docs/templates/DDD_STORY_TEMPLATE.md|.alto/templates/DDD_STORY_TEMPLATE.md"
        "docs/templates/ARCHITECTURE_TEMPLATE.md|.alto/templates/ARCHITECTURE_TEMPLATE.md"
        "docs/beads_templates/beads-epic-template.md|.alto/templates/beads-epic-template.md"
        "docs/beads_templates/beads-ticket-template.md|.alto/templates/beads-ticket-template.md"
        "docs/beads_templates/beads-spike-template.md|.alto/templates/beads-spike-template.md"
        "docs/beads_templates/beads-stub-template.md|.alto/templates/beads-stub-template.md"
        "docs/spikes/ddd_reference.md|.alto/templates/ddd_reference.md"
    )

    # 4 OVERLAY command sources (need to be at .alto/commands/ before split commits)
    local overlay_cmd_pairs=(
        ".claude/commands/brainstorm.md|.alto/commands/brainstorm.md"
        ".claude/commands/groom.md|.alto/commands/groom.md"
        ".claude/commands/launch-team.md|.alto/commands/launch-team.md"
        ".claude/commands/prd-traceability.md|.alto/commands/prd-traceability.md"
    )

    # 5 OVERLAY agent sources (need to be at .alto/agents/ before split commits)
    local overlay_agent_pairs=(
        ".claude/agents/developer.md|.alto/agents/developer.md"
        ".claude/agents/tech-lead.md|.alto/agents/tech-lead.md"
        ".claude/agents/project-manager.md|.alto/agents/project-manager.md"
        ".claude/agents/qa-engineer.md|.alto/agents/qa-engineer.md"
    )
    # NOTE: white-hacker.md is in the GENERIC list above (will be split in commit #12)
    # That is per spike-amendment: spike tagged it GENERIC originally, but L95 hit forces
    # OVERLAY treatment in this ticket. The git mv still happens in Phase 2; the split is #12.

    local pair src dst
    local commit_files=()

    for pair in "${generic_pairs[@]}" "${overlay_cmd_pairs[@]}" "${overlay_agent_pairs[@]}"; do
        src="${pair%%|*}"
        dst="${pair##*|}"
        if [ "$DRY_RUN" -eq 1 ]; then
            say "  git mv $src $dst"
        else
            git mv "$src" "$dst"
        fi
        # git mv self-stages the rename; only the destination needs git add (source path no longer exists)
        commit_files+=("$dst")
    done

    if [ "$DRY_RUN" -eq 1 ]; then
        say "  COMMIT #2: chore(scaffold): git-mv 14 generic assets into .alto/{agents,commands,templates}"
        return 0
    fi

    commit_phase "chore(scaffold): git-mv 14 generic assets into .alto/{agents,commands,templates}" \
        "${commit_files[@]}"
}

# ---------- OVERLAY split helpers ----------
#
# split approach: HEREDOC-based.
# For each OVERLAY file we:
#   (1) Write the .project.md sibling content verbatim from a heredoc — we author the
#       overlay content directly so it's reviewable in this script.
#   (2) Use targeted text edits (sed) to remove or replace alto-specific spans in the
#       parent GENERIC .md. Each edit cites the line range or pattern being removed.
#
# Rationale (orchestrator asked to document this choice):
#   * Heredocs let the dry-run print exactly what goes into each .project.md.
#   * Sed-line-range deletion is fragile against future edits; pattern-based block
#     deletion (one block at a time, with a sentinel pattern) is more durable. We use
#     sed `/pattern/,/end-pattern/d` and `sed -i.bak` so backups are preserved.
#   * .bak files land in migration-backups/ via a single mv at the end of each split.

# Move a .bak file into migration-backups/<unique-name>.bak
backup_bak() {
    local bak_src="$1"
    local unique_name="$2"
    if [ "$DRY_RUN" -eq 1 ]; then
        say "    mv $bak_src migration-backups/$unique_name"
        return 0
    fi
    mv "$bak_src" "migration-backups/$unique_name"
}

# In-place sed with .bak backup; emits diff to stdout (real mode only).
sed_edit() {
    local file="$1"
    local pattern="$2"
    local label="$3"
    if [ "$DRY_RUN" -eq 1 ]; then
        say "    sed -i.bak \"$pattern\" $file  # $label"
        return 0
    fi
    sed -i.bak "$pattern" "$file"
    say "    --- diff for $label ---"
    diff -u "${file}.bak" "$file" || true
}

# Write a heredoc payload to a file (real mode only; dry-run just announces).
write_overlay() {
    local dst="$1"
    local label="$2"
    if [ "$DRY_RUN" -eq 1 ]; then
        say "    WRITE $dst  # $label"
        return 0
    fi
    cat > "$dst"
    say "    wrote $dst ($label)"
}

# ---------- Phase 3a — Split OVERLAY commands (commits #3-#7) ----------

# Commit #3: brainstorm.md
split_brainstorm() {
    say "Phase 3a-1 — split .alto/commands/brainstorm.md + brainstorm.project.md"

    # OVERLAY payload (alto-namespace + literal docs/ paths)
    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/commands/brainstorm.project.md "alto namespace + literal paths" <<'OVERLAY_EOF'
# /brainstorm — alto-specific addenda

## Alto namespace

This command is also invocable as `/alto-brainstorm` for users who want the alto-branded
verb. Both resolve to the same skill.

## Artifact destinations (this project)

- PRD: `docs/PRD.md`
- DDD: `docs/DDD.md`
- ARCHITECTURE: `docs/ARCHITECTURE.md`

## Template references (this project)

- PRD template: `.alto/templates/PRD_TEMPLATE.md`
- DDD story template: `.alto/templates/DDD_STORY_TEMPLATE.md`
- Architecture template: `.alto/templates/ARCHITECTURE_TEMPLATE.md`
OVERLAY_EOF
    else
        write_overlay .alto/commands/brainstorm.project.md "alto namespace + literal paths" </dev/null
    fi

    # GENERIC parent rewrites:
    #   (a) frontmatter `name: alto-brainstorm` -> `name: brainstorm`
    #   (b) heading `/alto-brainstorm` -> `/brainstorm`
    #   (c) docs/templates/ template references -> .alto/templates/ (path migration)
    sed_edit .alto/commands/brainstorm.md \
        's|^name: alto-brainstorm$|name: brainstorm|' \
        "frontmatter name -> generic"
    backup_bak .alto/commands/brainstorm.md.bak brainstorm.md.bak-1
    sed_edit .alto/commands/brainstorm.md \
        's|/alto-brainstorm|/brainstorm|g' \
        "heading + body /alto-brainstorm -> /brainstorm"
    backup_bak .alto/commands/brainstorm.md.bak brainstorm.md.bak-2
    sed_edit .alto/commands/brainstorm.md \
        's|docs/templates/|.alto/templates/|g' \
        "template path references -> .alto/templates/"
    backup_bak .alto/commands/brainstorm.md.bak brainstorm.md.bak-3

    commit_phase "refactor(scaffold): split .alto/commands/brainstorm.md + brainstorm.project.md" \
        .alto/commands/brainstorm.md .alto/commands/brainstorm.project.md
}

# Commit #4: groom.md
split_groom() {
    say "Phase 3a-2 — split .alto/commands/groom.md + groom.project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/commands/groom.project.md "composition-root paths + ticket prefix" <<'OVERLAY_EOF'
# /groom — alto-specific addenda

## Ticket ID format

alto tickets use the `alto-` / `alty-cli-` prefix (e.g. `alto-0m9.2`, `alty-cli-766.1`).
Example usage:

```
/groom alto-0m9.2
/groom alty-cli-766.2
```

## Project-specific verification

```bash
alto ticket-verify <ticket-id>      # detects quantitative claims, verifies against command output
```

| Result | Action |
|--------|--------|
| All claims verified | Proceed to Phase 4 |
| Mismatch detected | Update ticket with correct values before proceeding |
| No claims found | Proceed to Phase 4 (no quantitative claims to verify) |
| Command not in allowlist | Note as UNVERIFIED in report |

## Composition root paths (Go layout)

- Port interfaces: `internal/{context}/application/ports.go`
- Adapters: `internal/{context}/infrastructure/{adapter}.go`
- Composition: `internal/composition/app.go`, `internal/composition/adapters.go`

Example constructor chain trace (replaces Phase 4b chain in the GENERIC parent):

```
NewXxxHandler(port)
  -> port type: XxxPort interface at internal/xxx/application/ports.go:NN
  -> methods: Foo(ctx, string) (Result, error) — confirmed line NN
  -> adapter: XxxAdapter at internal/xxx/infrastructure/xxx_adapter.go
  -> adapter constructor: NewXxxAdapter(dep1, dep2)
  -> dep1 comes from: ...
  -> wired in NewApp() at internal/composition/app.go:NN
  -> imports needed: xxxapp, xxxinfra
```

## Template reference

Generic template path becomes: `.alto/templates/beads-ticket-template.md`.
OVERLAY_EOF
    else
        write_overlay .alto/commands/groom.project.md "composition-root paths + ticket prefix" </dev/null
    fi

    # GENERIC parent rewrites — drop alto-specific examples + composition root references.
    #   (a) Usage example "alto-0m9.2" -> generic placeholder "<ticket-id>"
    #   (b) Template path docs/beads_templates/ -> .alto/templates/
    #   (c) Drop Phase 3.5 (Claim Verification — alto-only)
    #   (d) Drop internal/composition/ + internal/xxx/* references from Phase 4

    sed_edit .alto/commands/groom.md \
        's|/groom alto-0m9.2|/groom <ticket-id>|g' \
        "drop alto-0m9.2 example"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-1

    sed_edit .alto/commands/groom.md \
        's|docs/beads_templates/|.alto/templates/|g' \
        "template path -> .alto/templates/"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-2

    # Delete Phase 3.5 Claim Verification block (lines from "### Phase 3.5" through "### Phase 4 —")
    # Pattern-based deletion is durable: removes from Phase 3.5 header up to (not including) the next "### Phase" header.
    sed_edit .alto/commands/groom.md \
        '/^### Phase 3\.5 — Claim Verification/,/^### Phase 4 — Implementation Simulation/{/^### Phase 4/!d;}' \
        "drop Phase 3.5 Claim Verification (alto ticket-verify is alto-only)"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-3

    # Replace internal/-specific Phase 4a bullet ("Composition root (...)") with generic pointer
    sed_edit .alto/commands/groom.md \
        's|- Composition root (`internal/composition/app.go`, `adapters.go`)|- Composition root (project-specific path — see `groom.project.md`)|' \
        "Phase 4a: replace internal/composition/ ref with generic pointer"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-4

    # Replace `internal/xxx/...` lines in Phase 4b constructor-chain example with generic placeholders
    sed_edit .alto/commands/groom.md \
        's|internal/xxx/application/ports\.go:NN|{application_layer}/ports.<ext>:NN|g' \
        "Phase 4b: internal/xxx/application/ports.go -> generic placeholder"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-5

    sed_edit .alto/commands/groom.md \
        's|internal/xxx/infrastructure/xxx_adapter\.go|{infrastructure_layer}/xxx_adapter.<ext>|g' \
        "Phase 4b: internal/xxx/infrastructure/ -> generic placeholder"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-6

    sed_edit .alto/commands/groom.md \
        's|internal/composition/app\.go:NN|{composition_root}:NN  # project-specific|g' \
        "Phase 4b: internal/composition/app.go -> generic placeholder"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-7

    sed_edit .alto/commands/groom.md \
        's|`internal/composition/adapters\.go`|`{composition_root}/adapters.<ext>` (see project overlay)|g' \
        "Phase 4d: internal/composition/adapters.go -> generic placeholder"
    backup_bak .alto/commands/groom.md.bak groom.md.bak-8

    commit_phase "refactor(scaffold): split .alto/commands/groom.md + groom.project.md" \
        .alto/commands/groom.md .alto/commands/groom.project.md
}

# Commit #5: launch-team.md
split_launchteam() {
    say "Phase 3a-3 — split .alto/commands/launch-team.md + launch-team.project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/commands/launch-team.project.md "ticket prefix + Go quality gates" <<'OVERLAY_EOF'
# /launch-team — alto-specific addenda

## Ticket prefix

alto tickets use the `alty-cli-` prefix. Example invocations:

```
/launch-team alty-cli-1wu                                  # single ticket
/launch-team alty-cli-cgm alty-cli-dfd                     # multiple tickets
/launch-team alty-cli-cgm alty-cli-dfd alty-cli-yl0
```

Dev agent names follow the pattern `dev-alty-cli-<id>` (e.g. `dev-alty-cli-1wu`).

## Project artifact paths

- `docs/DDD.md` — domain model, bounded contexts, ubiquitous language
- `docs/ARCHITECTURE.md` — technical architecture
- `docs/PRD.md` — product requirements
- `.golangci.yml` — lint config v2 (must pass with 0 issues)
- `arch-go.yml` — DDD layer enforcement (domain cannot import application/infrastructure)

## Go quality gates (must pass at every checkpoint)

```bash
go build ./...                                   # 0 errors
go vet ./...                                     # 0 errors
golangci-lint run ./...                          # 0 issues (v2 strict config)
go test ./... -race                              # all pass, >=80% coverage on new domain code
```

## Go-specific design-decision categories (Step 5)

- Go types and their shapes (value objects, entities, aggregates)
- Interface contracts (port signatures, return types, context-arg position)
- Patterns to follow (reference existing code with file:line)
- Domain events emitted or consumed (Watermill GoChannel)
- Bounded context boundaries (which package owns what)
- Constraints (what NOT to do — especially DDD layer rules)

## Enforced principles (Go-specific additions)

- **DDD layer rules** — `internal/{context}/domain/` has ZERO external deps; dependencies
  flow inward: infrastructure -> application -> domain. Enforced by `arch-go`.
- **CQRS-lite** — events route through Watermill GoChannel.
- **Testify idioms** — `assert.Len`, `assert.Empty`, `assert.ErrorIs`, `require.Error` for
  preconditions, `assert.InDelta` for floats (testifylint enforces).

## Handoff doc paths

Single-ticket waves: `.notes/handoff-alty-cli-<id>.md` (e.g. `handoff-alty-cli-1wu.md`).
Multi-ticket waves: `.notes/handoff-wave-<n>.md` or `.notes/handoff-<short-name>.md`.
OVERLAY_EOF
    else
        write_overlay .alto/commands/launch-team.project.md "ticket prefix + Go quality gates" </dev/null
    fi

    # GENERIC parent rewrites — drop alty-cli- examples, Watermill, Go gates, .notes path.
    #
    # Strategy: targeted line/block deletions + ticket-prefix sanitisation. The parent retains
    # the language-neutral 7-phase team protocol.

    # (a) Usage examples: alty-cli-* -> generic <ticket-id> placeholders
    sed_edit .alto/commands/launch-team.md \
        's|/launch-team alty-cli-1wu|/launch-team <ticket-id>|g' \
        "Usage: drop alty-cli-1wu single-ticket example"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-1

    sed_edit .alto/commands/launch-team.md \
        's|/launch-team alty-cli-cgm alty-cli-dfd alty-cli-yl0|/launch-team <id-1> <id-2> <id-3>|g' \
        "Usage: drop 3-ticket alty-cli example"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-2

    sed_edit .alto/commands/launch-team.md \
        's|/launch-team alty-cli-cgm alty-cli-dfd|/launch-team <id-1> <id-2>|g' \
        "Usage: drop 2-ticket alty-cli example"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-3

    # (b) Dev agent name template
    sed_edit .alto/commands/launch-team.md \
        's|`dev-alty-cli-1wu`|`dev-<ticket-id>` (e.g. `dev-acme-42`)|' \
        "Dev agent name: drop alty-cli-1wu example"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-4

    # (c) Watermill GoChannel reference -> generic "event bus"
    sed_edit .alto/commands/launch-team.md \
        's|Domain events emitted or consumed (Watermill GoChannel)|Domain events emitted or consumed (project event bus)|' \
        "Step 5: Watermill GoChannel -> generic event bus"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-5

    # (d) Reference Files block: drop .golangci.yml + arch-go.yml (Go-specific)
    sed_edit .alto/commands/launch-team.md \
        '/^- `\.golangci\.yml`/d' \
        "Reference Files: drop .golangci.yml line"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-6

    sed_edit .alto/commands/launch-team.md \
        '/^- `arch-go\.yml`/d' \
        "Reference Files: drop arch-go.yml line"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-7

    # (e) Quality Gates block — replace Go-specific block with generic pointer
    # Match from "## Quality Gates" to the next "## " heading and replace
    sed_edit .alto/commands/launch-team.md \
        '/^## Quality Gates (must pass at every checkpoint)/,/^## Enforced Principles/{/^## Enforced Principles/!d;}' \
        "Quality Gates: drop Go-specific gates block (use project overlay)"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-8

    # Inject a generic-pointer block right BEFORE the Enforced Principles heading.
    # We add a marker comment + pointer; the project overlay carries the actual gates.
    sed_edit .alto/commands/launch-team.md \
        's|^## Enforced Principles (non-negotiable)$|## Quality Gates (must pass at every checkpoint)\n\nProject-specific. See the `.project.md` sibling for this project'\''s quality gate commands.\n\n## Enforced Principles (non-negotiable)|' \
        "Insert generic Quality Gates pointer"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-9

    # (f) Drop DDD layer rules + CQRS-lite + Testify idioms lines (Go-specific)
    sed_edit .alto/commands/launch-team.md \
        '/^- \*\*DDD layer rules\*\*/,/  Enforced by `arch-go`\./d' \
        "Enforced Principles: drop DDD layer rules (Go arch-go reference)"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-10

    sed_edit .alto/commands/launch-team.md \
        '/^- \*\*CQRS-lite\*\* — command handlers/,/  queries have no side effects; events route through Watermill GoChannel\./d' \
        "Enforced Principles: drop CQRS-lite + Watermill line"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-11

    sed_edit .alto/commands/launch-team.md \
        '/^- \*\*Testify idioms\*\*/,/  for preconditions, `assert\.InDelta` for floats (testifylint enforces)\./d' \
        "Enforced Principles: drop Testify idioms (Go-specific)"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-12

    # (g) Handoff path: `.notes/handoff-alty-cli-1wu.md` -> generic `<wave-or-ticket-id>`
    sed_edit .alto/commands/launch-team.md \
        's|handoff-alty-cli-1wu\.md|handoff-<wave-or-ticket-id>.md|g' \
        "Phase 7: drop alty-cli-1wu handoff filename example"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-13

    sed_edit .alto/commands/launch-team.md \
        's|(e\.g\., `handoff-alty-cli-1wu\.md`)|(e.g., `handoff-acme-42.md`)|g' \
        "Rule 7: drop alty-cli-1wu handoff filename example"
    backup_bak .alto/commands/launch-team.md.bak launch-team.md.bak-14

    commit_phase "refactor(scaffold): split .alto/commands/launch-team.md + launch-team.project.md" \
        .alto/commands/launch-team.md .alto/commands/launch-team.project.md
}

# Commit #6: prd-traceability.md (RLM pattern stays generic; C1-C25 -> overlay)
split_prdtraceability() {
    say "Phase 3a-4 — split .alto/commands/prd-traceability.md + prd-traceability.project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/commands/prd-traceability.project.md "alto C1-C25 + k7m.4 + doc-health invocation" <<'OVERLAY_EOF'
# /prd-traceability — alto-specific addenda

## PRD Capability Map — alto P0 + P1 tables

### P0 Capability → Bounded Context → Expected Ticket Coverage

| ID | PRD Capability | Bounded Context | Expected Ticket Scope |
|----|---------------|-----------------|----------------------|
| C1 | CLI tool (`vs`) | Bootstrap | CLI command tree, subcommands |
| C2 | MCP server | Bootstrap | MCP tool schemas, shared ports |
| C3 | `.alto/` project directory | Bootstrap | Directory structure, config.toml |
| C4 | `alto init` with preview | Bootstrap | Preview, confirm, file safety |
| C5 | Global settings detection | Bootstrap | Tool detection, conflict resolution |
| C6 | Existing project adoption (`alto init --existing`) | Rescue | Branch safety, gap report, scaffolding |
| C7 | Gap analysis | Rescue | Scan, compare, report |
| C8 | Guided project bootstrap | Guided Discovery | Conversational flow, question phases |
| C9 | DDD question framework | Guided Discovery | 10 questions, dual register, persona detection |
| C10 | Artifact generation | Domain Model | PRD, DDD.md, ARCHITECTURE.md from answers |
| C11 | Agent personas | Tool Translation | Developer, researcher, tech-lead, PM, QA agents |
| C12 | Beads integration | Ticket Pipeline | Epic/spike/ticket templates |
| C13 | Quality gates | Architecture Testing | ruff + mypy + pytest enforcement |
| C14 | Fitness function generation | Architecture Testing | import-linter + pytestarch from bounded context map |
| C15 | Domain story to ticket pipeline | Ticket Pipeline | DDD artifacts -> ordered beads tickets with formal `bd dep add` (not text-only deps) |
| C16 | Complexity budget | Domain Model | Core/Supporting/Generic classification + treatment levels |
| C17 | Multi-tool support | Tool Translation | Claude Code, Cursor, Roo Code, OpenCode configs |
| C18 | Knowledge base (RLM) | Knowledge Base | Addressable docs, DDD patterns, tool conventions |
| C19 | Doc maintenance commands | Knowledge Base | `alto doc-health`, `alto doc-review` |
| C20 | Ticket freshness & ripple review | Ticket Freshness | Close -> flag -> context diff -> review flow |
| C25 | Template-enforced ticket creation | Ticket Pipeline + Tool Translation + Ticket Freshness | Every ticket created (manual or generated) MUST use beads templates; generated CLAUDE.md enforces this in grooming checklist step 1 and after-close protocol step 2. Tickets: k7m.12 (after-close protocol design), k7m.20 (generated tickets use templates), k7m.21 (generated CLAUDE.md includes enforcement) |

### P1 Capability → Expected Ticket Scope

| ID | PRD Capability | Expected Ticket Scope |
|----|---------------|----------------------|
| C21 | Rescue mode structural migration | Implicit BC detection, anemic model scan, migration tickets |
| C22 | Tool knowledge versioning | Current + 3 previous major versions per tool |
| C23 | Knowledge base drift detection | Convention changes between versions, code vs doc divergence |
| C24 | Spike workflow | Guided spike creation, ADR output |

## Worked example report (alto)

```
============================================================
PRD TRACEABILITY REPORT: alto-k7m.4
============================================================

COVERED  C8  Guided project bootstrap
  -> alto-k7m.4 (deliverable: CLI command tree for alto init)

COVERED  C9  DDD question framework
  -> alto-k7m.4 (deliverable: alto guide design)

GAP      C19 Doc maintenance commands
  -> No ticket deliverable mentions alto doc-health or alto doc-review
  -> Should be in: CLI/MCP design spike (k7m.4)

============================================================
Coverage: 18/20 P0 capabilities (90%)
Gaps: 2 capabilities with no ticket coverage
============================================================
```

## Invocation

```bash
alto doc-health          # alto-specific doc-health invocation
```
OVERLAY_EOF
    else
        write_overlay .alto/commands/prd-traceability.project.md "alto C1-C25 + k7m.4 + doc-health invocation" </dev/null
    fi

    # GENERIC parent rewrites — drop the entire P0 capability table + P1 table + worked
    # example with alto-k7m.* ids + alto doc-health invocation. Keep the RLM pattern body.

    # (a) Delete P0 table body (rows from "| C1 |" through "| C25 |") AND the P1 section header + table.
    # Replace with a generic pointer. Use a block delete pattern.
    sed_edit .alto/commands/prd-traceability.md \
        '/^### P0 Capability → Bounded Context → Expected Ticket Coverage$/,/^## Implementation$/{/^## Implementation$/!d;}' \
        "Drop alto P0 + P1 capability tables (C1-C25, alto-only)"
    backup_bak .alto/commands/prd-traceability.md.bak prd-traceability.md.bak-1

    # Insert a generic pointer immediately before "## Implementation"
    sed_edit .alto/commands/prd-traceability.md \
        's|^## Implementation$|### Capability Table\n\nProject-specific. See `prd-traceability.project.md` for this project'\''s PRD capability map (P0 + P1).\n\n## Implementation|' \
        "Insert generic Capability Table pointer (project overlay carries the table)"
    backup_bak .alto/commands/prd-traceability.md.bak prd-traceability.md.bak-2

    # (b) Replace worked-example report (alto-k7m.4 references) with a generic example
    sed_edit .alto/commands/prd-traceability.md \
        '/^PRD TRACEABILITY REPORT: <scope>$/,/^Gaps: 2 capabilities with no ticket coverage$/{/^PRD TRACEABILITY REPORT: <scope>$/!{/^Gaps: 2 capabilities with no ticket coverage$/!d;};}' \
        "Worked example: strip alto-k7m.4 details (keep header + footer placeholder)"
    backup_bak .alto/commands/prd-traceability.md.bak prd-traceability.md.bak-3

    # (c) Drop alto doc-health invocation lines from "Implementation" section if any remain in body.
    # The worked-example block has "alto doc-health" — that block is already deleted above.
    # Re-check via a final tidy: replace residual `alto doc-health` mentions with generic doc-health.
    sed_edit .alto/commands/prd-traceability.md \
        's|alto doc-health|doc-health|g' \
        "Strip alto-namespace from any residual doc-health mentions"
    backup_bak .alto/commands/prd-traceability.md.bak prd-traceability.md.bak-4

    sed_edit .alto/commands/prd-traceability.md \
        's|alto doc-review|doc-review|g' \
        "Strip alto-namespace from any residual doc-review mentions"
    backup_bak .alto/commands/prd-traceability.md.bak prd-traceability.md.bak-5

    commit_phase "refactor(scaffold): split .alto/commands/prd-traceability.md + prd-traceability.project.md (RLM pattern stays generic; C1-C25 -> overlay)" \
        .alto/commands/prd-traceability.md .alto/commands/prd-traceability.project.md
}

# Commit #7: architecture-docs.md cross-ref edit (no split; sed only)
edit_architecturedocs() {
    say "Phase 3a-5 — sed .alto/commands/architecture-docs.md L34+L51 prd-traceability path"

    sed_edit .alto/commands/architecture-docs.md \
        's|\.claude/commands/prd-traceability\.md|.alto/commands/prd-traceability.md|g' \
        "L34+L51: .claude/commands/prd-traceability.md -> .alto/commands/prd-traceability.md"
    backup_bak .alto/commands/architecture-docs.md.bak architecture-docs.md.bak

    commit_phase "refactor(scaffold): sed -i.bak .alto/commands/architecture-docs.md L34+L51 prd-traceability path -> .alto/commands/" \
        .alto/commands/architecture-docs.md
}

phase3a_command_splits() {
    say "Phase 3a — OVERLAY command splits (commits #3-#7)"
    split_brainstorm
    split_groom
    split_launchteam
    split_prdtraceability
    edit_architecturedocs
}

# ---------- Phase 3b — Split OVERLAY agents (commits #8-#12) ----------

# Commit #8: developer.md
split_developer() {
    say "Phase 3b-1 — split .alto/agents/developer.md + developer.project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/agents/developer.project.md "Go 1.26+, Watermill, lint v2" <<'OVERLAY_EOF'
# Developer — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Source layout (Go)

```
internal/
├── {context}/
│   ├── domain/             # Core business logic (ZERO external deps)
│   ├── application/        # Use cases, command/query handlers, ports
│   └── infrastructure/     # Adapters for external concerns
├── shared/                 # Shared kernel
│   ├── domain/             # DomainModel, BoundedContext, sentinel errors, value objects, events
│   ├── application/        # Shared ports (FileWriter)
│   └── infrastructure/     # Event bus, LLM client, persistence
├── composition/            # Composition root (DI wiring)
└── integration/            # Cross-context integration tests
cmd/
├── alto/main.go            # CLI entry point (Cobra)
└── alto-mcp/main.go        # MCP server entry point
```

## CQRS-lite (Go-specific)

- Watermill GoChannel for event dispatch (local) / NATS (distributed)

## Linting Rules (golangci-lint v2)

These linters are enforced — your code MUST pass:

| Linter | What it checks |
|--------|---------------|
| errcheck | No ignored errors |
| errorlint | `errors.Is`/`errors.As` not type assertion |
| wrapcheck | Errors from external packages wrapped with `%w` |
| contextcheck | `context.Context` propagated correctly |
| noctx | `exec.CommandContext` not `exec.Command` |
| revive | No name stutter (`pkg.PkgFoo`), exported types documented |
| gocritic | No `os.Exit` after `defer` |
| exhaustive | Switch on enums covers all cases |
| testifylint | `assert.Len`, `assert.Empty`, `assert.ErrorIs` idioms |
| gci | Import order: stdlib | third-party | local |
| gofumpt | Stricter gofmt formatting |
| staticcheck | ST1005 (lowercase errors), SA1012 (no nil context) |

## Quality Gates (Go)

```bash
go build ./...           # Compile check
go test ./... -v -race   # Tests with race detector
go vet ./...             # Static analysis
golangci-lint run        # Meta-linter
```
OVERLAY_EOF
    else
        write_overlay .alto/agents/developer.project.md "Go 1.26+, Watermill, lint v2" </dev/null
    fi

    # GENERIC parent rewrites:
    #   (a) Drop the "The codebase is Go 1.26+" mention
    #   (b) Generalise the DDD Source Layout to language-neutral
    #   (c) Drop Watermill GoChannel reference (CQRS-lite)
    #   (d) Drop the entire "Linting Rules (golangci-lint v2)" section (Go-specific)
    #   (e) Drop the "Quality Gates" go-specific commands (point to project overlay)
    #   (f) Drop the cmd/alto/* sub-tree from DDD Source Layout

    # (a) Remove "The codebase is Go 1.26+."
    sed_edit .alto/agents/developer.md \
        's| The codebase is \*\*Go 1\.26+\*\*\.||g' \
        "Drop Go 1.26+ mention (project-specific)"
    backup_bak .alto/agents/developer.md.bak developer.md.bak-1

    # (c) Replace "Watermill GoChannel for event dispatch" line with generic event bus
    sed_edit .alto/agents/developer.md \
        's|- Watermill GoChannel for event dispatch (where applicable)|- Project event bus for event dispatch (see project overlay)|' \
        "CQRS-lite: Watermill -> generic event bus"
    backup_bak .alto/agents/developer.md.bak developer.md.bak-2

    # (b)+(f) Replace the DDD Source Layout tree block (Go-specific paths) with generic.
    # Match from "```" after "## DDD Source Layout" header through the trailing "```".
    # Simpler approach: delete the block containing internal/+cmd/ paths and replace with a generic.
    sed_edit .alto/agents/developer.md \
        '/^## DDD Source Layout$/,/^## Go DDD Patterns$/{/^## DDD Source Layout$/!{/^## Go DDD Patterns$/!d;};}' \
        "Drop Go-specific DDD Source Layout tree (project overlay has Go variant)"
    backup_bak .alto/agents/developer.md.bak developer.md.bak-3

    # Insert generic layout pointer between the now-empty section and "## Go DDD Patterns"
    sed_edit .alto/agents/developer.md \
        's|^## Go DDD Patterns$|```\nsrc/                        # adjust per language conventions\n├── {context}/              # one directory per bounded context\n│   ├── domain/             # core business logic (ZERO external deps)\n│   ├── application/        # use cases, command/query handlers, ports\n│   └── infrastructure/     # adapters for external concerns\n├── shared/domain/          # shared kernel across contexts\n```\n\nFor this project'\''s exact layout, see `developer.project.md`.\n\n## DDD Patterns|' \
        "Insert generic DDD Source Layout + rename Go DDD Patterns -> DDD Patterns"
    backup_bak .alto/agents/developer.md.bak developer.md.bak-4

    # (d) Delete "## Linting Rules (golangci-lint v2)" entire section to "## Quality Gates"
    sed_edit .alto/agents/developer.md \
        '/^## Linting Rules (golangci-lint v2)$/,/^## Quality Gates$/{/^## Quality Gates$/!d;}' \
        "Drop Linting Rules (golangci-lint v2) section — Go-specific"
    backup_bak .alto/agents/developer.md.bak developer.md.bak-5

    # (e) Replace Quality Gates Go commands block with generic pointer.
    # The block is: ## Quality Gates\n\n```bash\ngo build ./... ...\n```\n\n**All must pass...
    # Delete the "go build ./...\n...\ngolangci-lint run" block and replace with a generic pointer.
    sed_edit .alto/agents/developer.md \
        '/^go build \.\/\.\.\.           # Compile check$/,/^golangci-lint run        # Meta-linter$/d' \
        "Quality Gates: drop Go-specific gate commands"
    backup_bak .alto/agents/developer.md.bak developer.md.bak-6

    sed_edit .alto/agents/developer.md \
        's|^```bash$|```bash\n# Project-specific. See developer.project.md for this project'\''s commands.|' \
        "Quality Gates: insert generic pointer in the now-empty bash block (first occurrence)"
    # The above is global; it would prepend the marker to EVERY ```bash. Mitigation: only the
    # first ```bash in this file is the now-empty Quality Gates block (the Go DDD Patterns block
    # was renamed to DDD Patterns and uses ```go). Other code blocks use ```go. Verified by
    # reading the source before authoring this script.
    backup_bak .alto/agents/developer.md.bak developer.md.bak-7

    commit_phase "refactor(scaffold): split .alto/agents/developer.md + developer.project.md (Go 1.26+, Watermill, lint v2)" \
        .alto/agents/developer.md .alto/agents/developer.project.md
}

# Commit #9: tech-lead.md
split_techlead() {
    say "Phase 3b-2 — split .alto/agents/tech-lead.md + tech-lead.project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/agents/tech-lead.project.md "Go greps + arch-go" <<'OVERLAY_EOF'
# Tech Lead — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## DDD Layer Paths (Go layout)

- `internal/{context}/domain/` — ZERO external deps (compiler-enforced via `internal/`)
- `internal/{context}/application/` — depends on domain + ports only
- `internal/{context}/infrastructure/` — implements ports, external deps allowed
- `internal/shared/domain/` — shared kernel (errors, value objects, events, DDD types)

## CQRS-lite (Go-specific)

- Watermill GoChannel for event dispatch (where applicable)

## Layer Violation Detection (Go grep recipes)

```bash
# Check domain files don't import application or infrastructure
grep -r "internal/.*application\|internal/.*infrastructure" internal/*/domain/ internal/shared/domain/

# Check application files don't import infrastructure
grep -r "internal/.*infrastructure" internal/*/application/
```

## Quality Gate Enforcement (Go)

```bash
go build ./...                                    # Compile check
go test ./... -v -race -coverprofile=coverage.out  # Tests + race detector
go vet ./...                                      # Static analysis
golangci-lint run                                 # Meta-linter
go tool cover -func=coverage.out                  # Verify >= 80%
```

## Linting Enforcement (golangci-lint v2)

golangci-lint v2 config in `.golangci.yml`. Key linters:

| Linter | Purpose |
|--------|---------|
| errcheck | No ignored errors |
| errorlint | Proper error wrapping/matching |
| wrapcheck | External errors wrapped with `%w` |
| contextcheck | Context propagation |
| noctx | `exec.CommandContext` required |
| revive | No name stutter, exported docs |
| gocritic | No `os.Exit` after `defer` |
| exhaustive | Enum switches complete |
| testifylint | Testify idioms |
| gci | Import ordering |
| gofumpt | Strict formatting |
| depguard | Package dependency rules |

**`fieldalignment` is disabled** (memory optimization, not correctness).

DDD layer enforcement also handled by `arch-go` (MIT) — see `arch-go.yml`.
OVERLAY_EOF
    else
        write_overlay .alto/agents/tech-lead.project.md "Go greps + arch-go" </dev/null
    fi

    # GENERIC parent rewrites — strip Go runtime mention, internal/ paths, Watermill, lint tables.

    sed_edit .alto/agents/tech-lead.md \
        's| The codebase is \*\*Go 1\.26+\*\*\.||g' \
        "Drop Go 1.26+ mention"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-1

    # DDD Layer Paths section — drop Go-specific layout, point to project overlay
    sed_edit .alto/agents/tech-lead.md \
        '/^\*\*DDD Layer Paths:\*\*$/,/^### 2\. CQRS-lite Compliance$/{/^### 2\. CQRS-lite Compliance$/!d;}' \
        "Drop Go-specific DDD Layer Paths block"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-2

    sed_edit .alto/agents/tech-lead.md \
        's|^### 2\. CQRS-lite Compliance$|**DDD Layer Paths:** project-specific. See `tech-lead.project.md`.\n\n### 2. CQRS-lite Compliance|' \
        "Insert generic DDD Layer Paths pointer"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-3

    # CQRS-lite section — drop Watermill GoChannel
    sed_edit .alto/agents/tech-lead.md \
        '/^- Watermill GoChannel for event dispatch (where applicable)$/d' \
        "CQRS-lite: drop Watermill GoChannel"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-4

    # Layer Violation Detection — drop Go grep recipes (entire ### 3 section's bash block)
    sed_edit .alto/agents/tech-lead.md \
        '/^### 3\. Layer Violation Detection$/,/^### 4\. Code Review/{/^### 4\. Code Review/!d;}' \
        "Drop Go grep recipes (Layer Violation Detection section)"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-5

    sed_edit .alto/agents/tech-lead.md \
        's|^### 4\. Code Review — What to Look For$|### 3. Layer Violation Detection\n\nProject-specific. See `tech-lead.project.md` for grep recipes that detect cross-layer imports.\n\n### 4. Code Review — What to Look For|' \
        "Insert generic Layer Violation Detection pointer"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-6

    # Quality Gate Enforcement — drop Go commands
    sed_edit .alto/agents/tech-lead.md \
        '/^### 6\. Quality Gate Enforcement$/,/^### 7\. Linting Enforcement$/{/^### 7\. Linting Enforcement$/!d;}' \
        "Drop Go Quality Gate Enforcement commands"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-7

    sed_edit .alto/agents/tech-lead.md \
        's|^### 7\. Linting Enforcement$|### 6. Quality Gate Enforcement\n\nProject-specific. See `tech-lead.project.md` for this project'\''s gates.\n\n### 7. Linting Enforcement|' \
        "Insert generic Quality Gate Enforcement pointer"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-8

    # Linting Enforcement — drop golangci-lint v2 table
    sed_edit .alto/agents/tech-lead.md \
        '/^### 7\. Linting Enforcement$/,/^## Key Rules$/{/^## Key Rules$/!d;}' \
        "Drop golangci-lint v2 table (project overlay carries it)"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-9

    sed_edit .alto/agents/tech-lead.md \
        's|^## Key Rules$|### 7. Linting Enforcement\n\nProject-specific. See `tech-lead.project.md` for this project'\''s lint config and rules.\n\n## Key Rules|' \
        "Insert generic Linting Enforcement pointer"
    backup_bak .alto/agents/tech-lead.md.bak tech-lead.md.bak-10

    commit_phase "refactor(scaffold): split .alto/agents/tech-lead.md + tech-lead.project.md (Go greps + arch-go)" \
        .alto/agents/tech-lead.md .alto/agents/tech-lead.project.md
}

# Commit #10: project-manager.md
split_pm() {
    say "Phase 3b-3 — split .alto/agents/project-manager.md + project-manager.project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/agents/project-manager.project.md "internal/{context}/, cmd/alto/" <<'OVERLAY_EOF'
# Project Manager — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Ticket templates (this project)

- Epic: `.alto/templates/beads-epic-template.md`
- Task: `.alto/templates/beads-ticket-template.md`
- Spike: `.alto/templates/beads-spike-template.md`

## Go Quality Gates Reference

When creating/grooming tickets, reference these quality gates:

```bash
go build ./...           # Compile check
go test ./... -v -race   # Tests with race detector
go vet ./...             # Static analysis
golangci-lint run        # Meta-linter
```

## Go Project Structure

```
internal/{context}/domain/        # Domain layer per bounded context
internal/{context}/application/   # Application layer per bounded context
internal/{context}/infrastructure/ # Infrastructure layer per bounded context
internal/shared/domain/           # Shared kernel (errors, events, VOs, DDD types)
cmd/alto/                         # CLI entry point (Cobra)
cmd/alto-mcp/                     # MCP server entry point
```

## Go Ticket Conventions

- Tickets organized by DDD bounded context
- Each ticket specifies which `internal/{context}/` it affects
- Acceptance criteria include: `go build` passes, `go test -race` passes, `golangci-lint run` passes
- TDD required: RED/GREEN/REFACTOR phases documented
- BDD naming: `TestSubject_WhenCondition_ExpectOutcome`
- CQRS-lite: commands vs queries separated in ticket design
- Domain tests are the majority (fast, pure, no mocks)

## Enforced Principles (Go-specific row)

| Principle | Ticket Requirement |
|-----------|-------------------|
| Linting | `golangci-lint run` in quality gates |
OVERLAY_EOF
    else
        write_overlay .alto/agents/project-manager.project.md "internal/{context}/, cmd/alto/" </dev/null
    fi

    # GENERIC parent rewrites

    sed_edit .alto/agents/project-manager.md \
        's| The codebase is \*\*Go 1\.26+\*\*\.||g' \
        "Drop Go 1.26+ mention"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-1

    # Ticket Templates section — replace docs/beads_templates/ with .alto/templates/
    sed_edit .alto/agents/project-manager.md \
        's|docs/beads_templates/|.alto/templates/|g' \
        "Ticket Templates: docs/beads_templates/ -> .alto/templates/"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-2

    # Drop "## Go Quality Gates Reference" entire block (with code) -> pointer to project overlay
    sed_edit .alto/agents/project-manager.md \
        '/^## Go Quality Gates Reference$/,/^## Go Project Structure$/{/^## Go Project Structure$/!d;}' \
        "Drop Go Quality Gates Reference (Go-specific commands)"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-3

    sed_edit .alto/agents/project-manager.md \
        's|^## Go Project Structure$|## Quality Gates Reference\n\nProject-specific. See `project-manager.project.md` for this project'\''s gate commands.\n\n## Project Structure|' \
        "Insert generic Quality Gates Reference pointer + rename Go Project Structure"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-4

    # Replace the internal/ tree block with a generic layout pointer
    sed_edit .alto/agents/project-manager.md \
        '/^## Project Structure$/,/^## Go Ticket Conventions$/{/^## Project Structure$/!{/^## Go Ticket Conventions$/!d;};}' \
        "Drop Go-specific project tree (kept project overlay)"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-5

    sed_edit .alto/agents/project-manager.md \
        's|^## Go Ticket Conventions$|See `project-manager.project.md` for this project'\''s source layout.\n\n## Ticket Conventions|' \
        "Insert generic source-layout pointer + rename Go Ticket Conventions"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-6

    # Drop Go-specific Ticket Conventions bullets (golangci-lint, BDD test name format, etc.)
    sed_edit .alto/agents/project-manager.md \
        's|- Each ticket specifies which `internal/{context}/` it affects|- Each ticket specifies which bounded context it affects|' \
        "Ticket Conventions: internal/{context}/ -> bounded context (generic)"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-7

    sed_edit .alto/agents/project-manager.md \
        's|- Acceptance criteria include: `go build` passes, `go test -race` passes, `golangci-lint run` passes|- Acceptance criteria include the project'\''s quality gates (build, test, lint)|' \
        "Ticket Conventions: Go AC line -> generic"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-8

    # Enforced Principles table — drop the Linting (golangci-lint) row (Go-specific)
    sed_edit .alto/agents/project-manager.md \
        '/^| Linting | `golangci-lint run` in quality gates |$/d' \
        "Enforced Principles: drop golangci-lint row"
    backup_bak .alto/agents/project-manager.md.bak project-manager.md.bak-9

    commit_phase "refactor(scaffold): split .alto/agents/project-manager.md + project-manager.project.md (internal/{context}/, cmd/alto/)" \
        .alto/agents/project-manager.md .alto/agents/project-manager.project.md
}

# Commit #11: qa-engineer.md
split_qa() {
    say "Phase 3b-4 — split .alto/agents/qa-engineer.md + qa-engineer.project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/agents/qa-engineer.project.md "go test recipes" <<'OVERLAY_EOF'
# QA Engineer — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Test Commands (Go)

```bash
go test ./internal/domain/... -v -race                    # Domain only
go test ./internal/application/... -v -race               # Application only
go test ./internal/infrastructure/... -v -race            # Infrastructure only
go test ./... -v -race -coverprofile=coverage.out         # All + coverage
go tool cover -func=coverage.out                          # Coverage by function
go tool cover -html=coverage.out -o coverage.html         # Visual coverage
go test ./... -bench=. -benchmem                          # Benchmarks
```

## Quality Gates (Go)

```bash
go build ./...           # Compile check
go test ./... -v -race   # Tests with race detector
go vet ./...             # Static analysis
golangci-lint run        # Meta-linter
```
OVERLAY_EOF
    else
        write_overlay .alto/agents/qa-engineer.project.md "go test recipes" </dev/null
    fi

    sed_edit .alto/agents/qa-engineer.md \
        's| The codebase is \*\*Go 1\.26+\*\*\.||g' \
        "Drop Go 1.26+ mention"
    backup_bak .alto/agents/qa-engineer.md.bak qa-engineer.md.bak-1

    # Drop "## Test Commands" Go-specific block, replace with pointer
    sed_edit .alto/agents/qa-engineer.md \
        '/^## Test Commands$/,/^## Go Test Patterns$/{/^## Go Test Patterns$/!d;}' \
        "Drop Go test command recipes (project overlay carries them)"
    backup_bak .alto/agents/qa-engineer.md.bak qa-engineer.md.bak-2

    sed_edit .alto/agents/qa-engineer.md \
        's|^## Go Test Patterns$|## Test Commands\n\nProject-specific. See `qa-engineer.project.md` for this project'\''s test command recipes.\n\n## Test Patterns|' \
        "Insert generic Test Commands pointer + rename Go Test Patterns"
    backup_bak .alto/agents/qa-engineer.md.bak qa-engineer.md.bak-3

    # Drop "## Quality Gates" Go-specific block
    sed_edit .alto/agents/qa-engineer.md \
        '/^## Quality Gates$/,/^## Key Rules$/{/^## Key Rules$/!d;}' \
        "Drop Go Quality Gates commands"
    backup_bak .alto/agents/qa-engineer.md.bak qa-engineer.md.bak-4

    sed_edit .alto/agents/qa-engineer.md \
        's|^## Key Rules$|## Quality Gates\n\nProject-specific. See `qa-engineer.project.md` for this project'\''s gate commands.\n\n## Key Rules|' \
        "Insert generic Quality Gates pointer"
    backup_bak .alto/agents/qa-engineer.md.bak qa-engineer.md.bak-5

    commit_phase "refactor(scaffold): split .alto/agents/qa-engineer.md + qa-engineer.project.md (go test recipes)" \
        .alto/agents/qa-engineer.md .alto/agents/qa-engineer.project.md
}

# Commit #12: white-hacker.md (spike amendment — L95 golangci-lint --enable gosec)
split_whitehacker() {
    say "Phase 3b-5 — split .alto/agents/white-hacker.md + white-hacker.project.md (spike amendment)"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/agents/white-hacker.project.md "OVERLAY amendment: L95 gosec invocation" <<'OVERLAY_EOF'
# White Hat Hacker — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Go-specific scanning commands

```bash
# Go vulnerability check
govulncheck ./...

# Dependency audit
go list -m -json all | grep -i "CVE\|vulnerability"

# Static analysis security rules (Go)
golangci-lint run --enable gosec
```

## Hardcoded-secret grep (Go-specific include pattern)

```bash
grep -rn "password\|secret\|api.key\|token" --include="*.go" . | grep -v "_test.go" | grep -v "vendor/"
```
OVERLAY_EOF
    else
        write_overlay .alto/agents/white-hacker.project.md "OVERLAY amendment: L95 gosec invocation" </dev/null
    fi

    sed_edit .alto/agents/white-hacker.md \
        's| The codebase is \*\*Go 1\.26+\*\*\.||g' \
        "Drop Go 1.26+ mention"
    backup_bak .alto/agents/white-hacker.md.bak white-hacker.md.bak-1

    # Drop the Go-specific `golangci-lint run --enable gosec` line (the spike-amendment L95 hit)
    sed_edit .alto/agents/white-hacker.md \
        '/^# Static analysis security rules$/,/^golangci-lint run --enable gosec$/d' \
        "Drop Go golangci-lint gosec invocation (L95 — spike amendment)"
    backup_bak .alto/agents/white-hacker.md.bak white-hacker.md.bak-2

    # Drop the govulncheck Go-specific block
    sed_edit .alto/agents/white-hacker.md \
        '/^# Go vulnerability check$/,/^govulncheck \.\/\.\.\.$/d' \
        "Drop Go govulncheck recipe"
    backup_bak .alto/agents/white-hacker.md.bak white-hacker.md.bak-3

    # Drop the Go list dependency-audit block
    sed_edit .alto/agents/white-hacker.md \
        '/^# Dependency audit$/,/^go list -m -json all | grep -i "CVE\\|vulnerability"$/d' \
        "Drop Go dependency-audit recipe"
    backup_bak .alto/agents/white-hacker.md.bak white-hacker.md.bak-4

    # Generalise the hardcoded-secrets grep (drop --include="*.go").
    # Uses # as delimiter because replacement contains literal | (shell pipes).
    # Anchored to ^...$ so it does not match the word "secret" elsewhere in the file.
    sed_edit .alto/agents/white-hacker.md \
        's#^grep -rn "password\\|secret\\|api\.key\\|token" --include="\*\.go" \. | grep -v "_test\.go" | grep -v "vendor/"$#grep -rnE "password\|secret\|api[._-]?key\|token" . | grep -v "vendor/"#' \
        "Generalise hardcoded-secret grep (drop Go-specific --include)"
    backup_bak .alto/agents/white-hacker.md.bak white-hacker.md.bak-5

    commit_phase "refactor(scaffold): split .alto/agents/white-hacker.md + white-hacker.project.md (OVERLAY amendment: L95 golangci-lint --enable gosec)" \
        .alto/agents/white-hacker.md .alto/agents/white-hacker.project.md
}

phase3b_agent_splits() {
    say "Phase 3b — OVERLAY agent splits (commits #8-#12)"
    split_developer
    split_techlead
    split_pm
    split_qa
    split_whitehacker
}

# ---------- Phase 3c — Split ARCHITECTURE_TEMPLATE.md (commit #13) ----------
split_archtemplate() {
    say "Phase 3c — split .alto/templates/ARCHITECTURE_TEMPLATE.md + .project.md"

    if [ "$DRY_RUN" -eq 0 ]; then
        write_overlay .alto/templates/ARCHITECTURE_TEMPLATE.project.md "Go internal/{context}/ overlay" <<'OVERLAY_EOF'
---
last_reviewed: YYYY-MM-DD
owner: architecture
status: draft
language: go
---

# Architecture (Go addenda) — alto

This file overlays the GENERIC `ARCHITECTURE_TEMPLATE.md` with Go-specific guidance.
Include the relevant rows when filling in `docs/ARCHITECTURE.md` for a Go project.

## Layer Rules (Go)

| Layer          | Can Depend On              | Cannot Depend On                        |
| -------------- | -------------------------- | --------------------------------------- |
| Domain         | Nothing (pure Go)          | Application, Infrastructure, frameworks |
| Application    | Domain, Ports (interfaces) | Infrastructure, frameworks              |
| Infrastructure | Application, Domain        | — (outermost layer)                     |

## Source Layout (Go)

```
internal/
├── {context}/
│   ├── domain/              # Entities, Value Objects, Aggregates, Domain Events
│   ├── application/         # Command handlers, query handlers, ports (interfaces)
│   └── infrastructure/      # Persistence, messaging, external clients
└── shared/                  # Shared kernel
    ├── domain/              # Shared VOs, sentinel errors, events, DDD types
    ├── application/         # Shared ports
    └── infrastructure/      # Event bus, persistence adapters
cmd/
└── alto/main.go             # CLI entry point (Cobra)
```

## Deployment (Go)

| Aspect          | Choice  | Rationale              |
| --------------- | ------- | ---------------------- |
| Runtime         | Go 1.26+| Project standard       |
| Package manager | go mod  | Stdlib                 |
OVERLAY_EOF
    else
        write_overlay .alto/templates/ARCHITECTURE_TEMPLATE.project.md "Go internal/{context}/ overlay" </dev/null
    fi

    # GENERIC parent rewrites — drop Python residue.
    # (a) Line 66 "pure Python" -> "pure" (language-neutral)
    sed_edit .alto/templates/ARCHITECTURE_TEMPLATE.md \
        's|Nothing (pure Python)|Nothing (pure — language-specific overlay)|' \
        "Layer Rules: drop 'pure Python' residue"
    backup_bak .alto/templates/ARCHITECTURE_TEMPLATE.md.bak ARCHITECTURE_TEMPLATE.md.bak-1

    # (b) Drop the Python src/{domain,application,infrastructure}/ tree block (lines 72-85).
    # Replace with a generic pointer + reference to the project overlay.
    sed_edit .alto/templates/ARCHITECTURE_TEMPLATE.md \
        '/^### Source Layout$/,/^## 4\. Bounded Context Integration$/{/^### Source Layout$/!{/^## 4\. Bounded Context Integration$/!d;};}' \
        "Source Layout: drop Python tree (project overlay carries language-specific tree)"
    backup_bak .alto/templates/ARCHITECTURE_TEMPLATE.md.bak ARCHITECTURE_TEMPLATE.md.bak-2

    sed_edit .alto/templates/ARCHITECTURE_TEMPLATE.md \
        's|^### Source Layout$|### Source Layout\n\nLanguage-specific. See your project'\''s `ARCHITECTURE_TEMPLATE.project.md` overlay or fill in directly.\n|' \
        "Insert generic Source Layout pointer"
    backup_bak .alto/templates/ARCHITECTURE_TEMPLATE.md.bak ARCHITECTURE_TEMPLATE.md.bak-3

    # (c) Line 138 "Runtime | Python 3.12 | Project standard" + "Package manager | uv | ..."
    # Drop these two rows; keep the table header + a placeholder row.
    sed_edit .alto/templates/ARCHITECTURE_TEMPLATE.md \
        '/^| Runtime         | Python 3\.12 | Project standard       |$/d' \
        "Deployment: drop Python 3.12 row"
    backup_bak .alto/templates/ARCHITECTURE_TEMPLATE.md.bak ARCHITECTURE_TEMPLATE.md.bak-4

    sed_edit .alto/templates/ARCHITECTURE_TEMPLATE.md \
        '/^| Package manager | uv          | Speed, reproducibility |$/d' \
        "Deployment: drop uv row"
    backup_bak .alto/templates/ARCHITECTURE_TEMPLATE.md.bak ARCHITECTURE_TEMPLATE.md.bak-5

    commit_phase "refactor(scaffold): split .alto/templates/ARCHITECTURE_TEMPLATE.md + .project.md (Go internal/{context}/ overlay)" \
        .alto/templates/ARCHITECTURE_TEMPLATE.md .alto/templates/ARCHITECTURE_TEMPLATE.project.md
}

phase3c_template_split() {
    split_archtemplate
}

# ---------- Phase 4 — Write .alto/CONTEXT.md (commit #14) ----------
phase4_context() {
    say "Phase 4 — write .alto/CONTEXT.md (5 ubiquitous-language terms)"

    if [ "$DRY_RUN" -eq 0 ]; then
        cat > .alto/CONTEXT.md <<'CTX_EOF'
# .alto/ Scaffold — Ubiquitous Language

This document defines the terms used across this scaffold's commands, agents, templates,
and skills. Mirrors mattpocock/skills' `CONTEXT.md` role ("helps agents decode the jargon
used in the project").

## Core terms (in this order)

### Scaffold
The `.alto/` tree shipped to downstream consumers. It contains workflow assets that any
project can adopt: commands, agents, templates, skills. Lifecycle folders track maturity.

### Workflow Asset
Any `.md` file under `.alto/` (command, agent, template, or skill). Each asset carries
YAML frontmatter declaring its `name`, `description`, `kind`, `phase`, and required tools.

### GENERIC
A workflow asset that is usable in any project without modification. GENERIC assets carry
no project-specific paths, language references, or ticket prefixes. They are the canonical
shippable artifact downstream consumers receive.

### OVERLAY
Project-specific addenda extracted from a workflow asset into a sibling `.project.md` file.
The OVERLAY carries language-specific source layouts, project-specific paths, tooling
references (lint configs, test commands), and any other context that does not generalise.

### `.project.md` sibling
The file naming convention `<asset>.project.md` carrying project-local content next to
a GENERIC asset. Example: `.alto/commands/groom.md` (GENERIC) lives next to
`.alto/commands/groom.project.md` (OVERLAY). Claude Code automatically merges sibling
`.md` files when invoking a skill, so the OVERLAY loads at invocation time.

## Scope clarifications

- **`.claude/skills/` retains personal Claude Code skills.** The 18 non-symlink design/craft
  skills under `.claude/skills/` (adapt, animate, audit, etc.) and the vendored `gstack/`
  bundle (symlinks) are NOT migrated. They are personal Claude Code assets, not alto
  workflow scaffold.
- **`.alto/skills/` is reserved for shipped alto workflow skills only.** It is empty at
  migration time (placeholder `.gitkeep` only); follow-up tickets may add scaffold skills.

## Platform caveats

- **Windows POSIX symlinks.** The Phase-5 symlink bridge (`.claude/commands/*.md ->
  ../../.alto/commands/*.md`) uses `ln -s`. On Windows, Claude Code may not resolve POSIX
  symlinks created without administrator privileges. Windows users defer to the
  `additionalDirectories` settings.json mechanism (`alto init --with-scaffold` follow-up).

## File-tree contract

```
.alto/
├── CONTEXT.md            # this file
├── commands/             # invocable workflows (one .md per command)
├── agents/               # personas (one .md per agent)
├── templates/            # documentation + ticket templates
├── skills/               # reserved for shipped alto skills (empty)
└── lifecycle/
    ├── in-progress/      # assets under design, not yet stable
    └── deprecated/       # retained for migration, not for new use
```
CTX_EOF
        say "  wrote .alto/CONTEXT.md"
    else
        say "  WRITE .alto/CONTEXT.md (5 terms: Scaffold, Workflow Asset, GENERIC, OVERLAY, .project.md sibling)"
    fi

    commit_phase "docs(scaffold): add .alto/CONTEXT.md (ubiquitous language: Scaffold, Workflow Asset, GENERIC, OVERLAY, .project.md sibling)" \
        .alto/CONTEXT.md
}

# ---------- Phase 5 — Symlink bridge (commit #15) ----------
phase5_symlinks() {
    say "Phase 5 — symlink bridge .claude/commands/ -> .alto/commands/"

    local names=(architecture-docs brainstorm design-ticket doc-health groom launch-team prd-traceability review)
    local name target

    for name in "${names[@]}"; do
        target=".claude/commands/${name}.md"
        if [ "$DRY_RUN" -eq 1 ]; then
            say "  pre-existence check: $target"
            say "  ln -s ../../.alto/commands/${name}.md $target"
            continue
        fi
        if [ -e "$target" ] || [ -L "$target" ]; then
            fail "symlink collision at $target — refusing to overwrite (use 'rm' first and re-run)"
        fi
        ln -s "../../.alto/commands/${name}.md" "$target"
        say "  created $target -> ../../.alto/commands/${name}.md"
    done

    if [ "$DRY_RUN" -eq 0 ]; then
        local symlink_files=()
        for name in "${names[@]}"; do
            symlink_files+=(".claude/commands/${name}.md")
        done
        commit_phase "chore(scaffold): add 8 .claude/commands/ -> ../../.alto/commands/ symlinks (ln -s, collision-abort)" \
            "${symlink_files[@]}"
    else
        say "  COMMIT #15: chore(scaffold): add 8 .claude/commands/ -> ../../.alto/commands/ symlinks"
    fi
}

# ---------- Phase 6 — Root edits (commit #16) ----------
# All sed targets in the ticket's Phase 6, EXCEPT architecture-docs.md (that one was
# already done in Phase 3a-5 commit #7). Reads in this commit:
#   * .claude/CLAUDE.md (3 edits: 2 path families + 1 Project Structure tree)
#   * bin/bd-ripple (single sed covers L191 + L238)
#   * internal/ticket/domain/template_type.go:4 comment
#   * .notes/order.md (APPEND paragraph — not sed)
#
# NOTE: .notes/order.md is an INTENDED write surface. The dirty-tree preflight runs BEFORE
# this phase, so the append does not violate the precondition.
phase6_root_edits() {
    say "Phase 6 — root edits (.claude/CLAUDE.md, bin/bd-ripple, template_type.go:4, .notes/order.md)"

    # (a) .claude/CLAUDE.md — docs/beads_templates/ -> .alto/templates/
    sed_edit .claude/CLAUDE.md \
        's|docs/beads_templates/|.alto/templates/|g' \
        ".claude/CLAUDE.md: docs/beads_templates/ -> .alto/templates/"
    backup_bak .claude/CLAUDE.md.bak CLAUDE.md.bak-1

    # (b) .claude/CLAUDE.md — docs/templates/ -> .alto/templates/
    sed_edit .claude/CLAUDE.md \
        's|docs/templates/|.alto/templates/|g' \
        ".claude/CLAUDE.md: docs/templates/ -> .alto/templates/"
    backup_bak .claude/CLAUDE.md.bak CLAUDE.md.bak-2

    # (c) .claude/CLAUDE.md — Project Structure tree (lines 132-137 in source):
    #     update the `├── docs/` subtree to reference `.alto/templates/` and friends.
    #     The block targets lines 134-137 specifically (templates/, beads_templates/, spikes/).
    #     Strategy: pattern-replace each tree line individually so the indentation stays intact.
    sed_edit .claude/CLAUDE.md \
        's|│   ├── templates/               # PRD, DDD Story, Architecture templates|│   └── research/                # Spike output reports|' \
        "Project Structure tree: drop docs/templates/ line (moved to .alto/templates/)"
    backup_bak .claude/CLAUDE.md.bak CLAUDE.md.bak-3a

    # Now we have two identical lines — remove one of them (the original └── research/).
    # Use awk via sed wouldn't help; use a delete-after-pattern: delete the second occurrence
    # by matching context. Simpler approach: revert and use a multi-line aware approach.
    # Since editing tree blocks is fragile, use targeted line-by-line edits:
    #   First, restore from the previous backup by re-running with a different strategy:
    sed_edit .claude/CLAUDE.md \
        '/│   ├── beads_templates\/         # Epic, spike, ticket templates/d' \
        "Project Structure tree: drop docs/beads_templates/ line"
    backup_bak .claude/CLAUDE.md.bak CLAUDE.md.bak-3b

    sed_edit .claude/CLAUDE.md \
        '/│   ├── spikes\/                  # Research spike definitions/d' \
        "Project Structure tree: drop docs/spikes/ line"
    backup_bak .claude/CLAUDE.md.bak CLAUDE.md.bak-3c

    # Now collapse the now-duplicated "│   └── research/" entries: keep only the first.
    # Use awk-via-sed: use line-addressing with a pattern that matches the literal line and
    # deletes only the SECOND occurrence by counting matches.
    # POSIX-portable approach: use awk in the actual edit. Documented as exception to the
    # otherwise-pure-sed pipeline (sed cannot count occurrences in a single pass).
    if [ "$DRY_RUN" -eq 0 ]; then
        cp .claude/CLAUDE.md migration-backups/CLAUDE.md.bak-3d
        awk 'BEGIN{cnt=0} /│   └── research\/                # Spike output reports/{cnt++; if(cnt==2){next}} {print}' \
            .claude/CLAUDE.md > .claude/CLAUDE.md.tmp && mv .claude/CLAUDE.md.tmp .claude/CLAUDE.md
        say "    --- diff for Project Structure tree: collapse duplicate research/ line ---"
        diff -u migration-backups/CLAUDE.md.bak-3d .claude/CLAUDE.md || true
    else
        say "    awk: collapse duplicate '│   └── research/' line (sed cannot count occurrences)"
    fi

    # Now insert a .alto/ subtree below the docs/ block to reflect the new home.
    # Add it RIGHT AFTER the docs/ block, BEFORE the .claude/ block.
    sed_edit .claude/CLAUDE.md \
        's|^├── \.claude/$|├── .alto/                       # Scaffold root (commands, agents, templates, skills)\n│   ├── CONTEXT.md               # Ubiquitous-language glossary\n│   ├── commands/                # Slash commands (GENERIC + .project.md OVERLAY siblings)\n│   ├── agents/                  # Agent personas (GENERIC + .project.md OVERLAY siblings)\n│   ├── templates/               # PRD, DDD Story, Architecture, beads templates\n│   ├── skills/                  # Reserved for shipped alto skills\n│   └── lifecycle/               # in-progress/ + deprecated/\n├── .claude/|' \
        "Project Structure tree: insert .alto/ subtree before .claude/"
    backup_bak .claude/CLAUDE.md.bak CLAUDE.md.bak-3e

    # (d) bin/bd-ripple — L191 + L238 (single sed pattern covers BOTH)
    sed_edit bin/bd-ripple \
        's|docs/beads_templates/beads-ticket-template\.md|.alto/templates/beads-ticket-template.md|g' \
        "bin/bd-ripple: L191+L238 ticket-template path -> .alto/templates/"
    backup_bak bin/bd-ripple.bak bd-ripple.bak

    # (e) internal/ticket/domain/template_type.go:4 comment
    sed_edit internal/ticket/domain/template_type.go \
        's|docs/beads_templates/|.alto/templates/|g' \
        "template_type.go:4: docs/beads_templates/ -> .alto/templates/"
    backup_bak internal/ticket/domain/template_type.go.bak template_type.go.bak

    # (f) .notes/order.md — APPEND a paragraph registering alty-cli-766 settled decisions.
    # This is an INTENDED write surface, not a sed in-place edit.
    # `.notes/` is gitignored as a whole; whitelist order.md so it can be committed.
    # Idempotent guard: only append the whitelist rules if not already present.
    if [ "$DRY_RUN" -eq 0 ]; then
        if ! grep -qxF '!/.notes/order.md' .gitignore; then
            printf '\n# whitelist scaffold migration registration (alty-cli-766.2)\n/.notes/*\n!/.notes/order.md\n' >> .gitignore
        fi
        cp .notes/order.md migration-backups/order.md.bak
        cat >> .notes/order.md <<'NOTE_EOF'

## Epic alty-cli-766 — Generic, reusable workflow scaffold (registered 2026-05-29)

Settled decisions from spike alty-cli-766.1:

- Root folder name: `.alto/`
- Layout: flat-by-category (`commands/`, `agents/`, `templates/`, `skills/`, `lifecycle/{in-progress,deprecated}/`)
- Generic vs alto-specific split: `.project.md` overlay siblings (NOT placeholders, NOT layered CLAUDE.md alone)
- Slash-command continuity: symlink bridge in repo (interim); `additionalDirectories` deferred to follow-up #2
- Shipping mechanism: `alto init --with-scaffold` via Go `embed.FS`
- Tool translation: extends existing `internal/tooltranslation/application/ConfigGeneration` port; adapters under `internal/tooltranslation/infrastructure/`
NOTE_EOF
        say "    --- diff for .notes/order.md ---"
        diff -u migration-backups/order.md.bak .notes/order.md || true
    else
        say "    APPEND .notes/order.md  # 6 settled-decision bullets (registered 2026-05-29)"
    fi

    commit_phase "refactor(scaffold): Phase 6 root edits — CLAUDE.md + bd-ripple + template_type.go + .gitignore whitelist + .notes/order.md registration" \
        .gitignore \
        .claude/CLAUDE.md \
        bin/bd-ripple \
        internal/ticket/domain/template_type.go \
        .notes/order.md
}

# ---------- Phase 7 — Verify + commit backups (commit #17) ----------
phase7_verify_and_backups() {
    say "Phase 7 — verify + commit migration-backups/"

    if [ "$DRY_RUN" -eq 1 ]; then
        say "  verify .alto/ structure (commands, agents, templates, skills, lifecycle/{in-progress,deprecated})"
        say "  verify fitness grep returns 0 matches"
        say "  verify orphan check (every .project.md has matching .md)"
        say "  verify stale-path grep returns 0 matches"
        say "  run: go build ./... && go vet ./... && go test ./... -race && golangci-lint run ./..."
        say "  run: go run ./cmd/alto doc-health"
        say "  [post-pass] rmdir docs/templates docs/beads_templates docs/spikes 2>/dev/null || true  # Edge case (ticket body): clean up source dirs left empty after Phase 2 moves"
        say "  IF all green: COMMIT #17: chore(scaffold): commit migration-backups/ .bak files + Phase 7 verification log"
        return 0
    fi

    local failed=0

    say "  [check 1/6] .alto/ structure"
    for d in .alto/commands .alto/agents .alto/templates .alto/skills \
             .alto/lifecycle/in-progress .alto/lifecycle/deprecated; do
        if [ ! -d "$d" ]; then
            say "    MISSING: $d"
            failed=1
        fi
    done
    [ "$failed" -eq 0 ] && say "    OK"

    say "  [check 2/6] fitness grep (coupling tokens in GENERIC files)"
    if grep -rE '\binternal/|\balto-|\balty-cli\b|cmd/alto\b|Watermill\b|golangci' \
        .alto/commands/*.md .alto/agents/*.md .alto/templates/*.md 2>/dev/null \
        | grep -v '\.project\.md' \
        | grep -v '^[^:]*:$' >/tmp/alto-fitness-grep.$$ ; then
        if [ -s /tmp/alto-fitness-grep.$$ ]; then
            say "    FAIL — coupling matches in GENERIC files:"
            sed 's/^/      /' </tmp/alto-fitness-grep.$$ >&2
            failed=1
        fi
    fi
    rm -f /tmp/alto-fitness-grep.$$
    [ "$failed" -eq 0 ] && say "    OK"

    say "  [check 3/6] orphan check (every .project.md has matching .md)"
    local f base
    for f in .alto/commands/*.project.md .alto/agents/*.project.md .alto/templates/*.project.md; do
        [ -e "$f" ] || continue
        base="${f%.project.md}.md"
        if [ ! -f "$base" ]; then
            say "    ORPHAN: $f (missing $base)"
            failed=1
        fi
    done
    [ "$failed" -eq 0 ] && say "    OK"

    say "  [check 4/6] stale-path grep (docs/beads_templates or docs/templates in cmd/internal/bin/.claude)"
    # exclude .claude/agent-memory/ (gitignored auto-generated Claude Code agent memory store)
    if grep -rn 'docs/beads_templates\|docs/templates/' cmd/ internal/ bin/ .claude/ 2>/dev/null \
        | grep -v 'bin/migrate-to-alto\.sh\|\.claude/agent-memory/' >/tmp/alto-stale-grep.$$ ; then
        if [ -s /tmp/alto-stale-grep.$$ ]; then
            say "    FAIL — stale references:"
            sed 's/^/      /' </tmp/alto-stale-grep.$$ >&2
            failed=1
        fi
    fi
    rm -f /tmp/alto-stale-grep.$$
    [ "$failed" -eq 0 ] && say "    OK"

    say "  [check 5/6] Go quality gates"
    if ! go build ./...; then say "    go build FAIL"; failed=1; fi
    if ! go vet ./...; then say "    go vet FAIL"; failed=1; fi
    if ! go test ./... -race; then say "    go test FAIL"; failed=1; fi
    if ! golangci-lint run ./...; then say "    golangci-lint FAIL"; failed=1; fi
    [ "$failed" -eq 0 ] && say "    OK"

    say "  [check 6/6] doc-health"
    if ! go run ./cmd/alto doc-health; then
        say "    doc-health FAIL"
        failed=1
    else
        say "    OK"
    fi

    if [ "$failed" -ne 0 ]; then
        say "Phase 7 verification FAILED — commit #17 (migration-backups + verify log) NOT made."
        say "Investigate, fix, and re-run only the failed checks before requesting GO again."
        exit 2
    fi

    # Edge case (ticket body): clean up source dirs left empty after Phase 2 moves.
    # `|| true` keeps the script alive if a dir is missing or non-empty.
    say "  [post-pass] rmdir source dirs left empty by Phase 2"
    if [ "$DRY_RUN" -eq 1 ]; then
        say "    rmdir docs/templates docs/beads_templates docs/spikes 2>/dev/null || true"
    else
        rmdir docs/templates docs/beads_templates docs/spikes 2>/dev/null || true
    fi

    say "  all checks passed — committing migration-backups/"
    git add migration-backups/
    commit_phase "chore(scaffold): commit migration-backups/ .bak files + Phase 7 verification log (ls, fitness grep, doc-health, Go quality gates)" \
        migration-backups/
}

# ---------- Main orchestration ----------
main() {
    say "migrate-to-alto.sh — alty-cli-766.2"
    say "  --dry-run: $DRY_RUN"
    say "  --force: $FORCE"
    say "  --i-know-what-i-am-doing: $CONFIRM"
    say ""

    phase0_preflight
    phase1_skeleton
    phase2_moves
    phase3a_command_splits
    phase3b_agent_splits
    phase3c_template_split
    phase4_context
    phase5_symlinks
    phase6_root_edits
    phase7_verify_and_backups

    say ""
    if [ "$DRY_RUN" -eq 1 ]; then
        say "dry-run complete — no filesystem changes made"
    else
        say "migration complete — 17 commits produced; review with: git log --oneline -17"
    fi
}

main
