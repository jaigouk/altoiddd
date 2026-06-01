---
name: write-a-workflow-asset
description: Interactive meta-skill for authoring a new workflow asset (command, agent, template, or skill) with enforced canonical frontmatter
kind: command
phase: design
when_to_use: When a contributor wants to author a new workflow asset (command/agent/template/skill) with valid frontmatter
tools_required: Read, Glob, Write
bash_substitution_policy: none
license: Apache-2.0
---

# /write-a-workflow-asset

Walk a contributor through authoring a **new workflow asset** — a command,
agent, template, or skill — with valid canonical frontmatter, then write the
draft into the in-progress lifecycle folder so doc-health can validate it.

This meta-skill is the front door for new scaffold assets. It enforces the
8-field frontmatter schema, rejects unsafe asset names, checks for collisions
across every asset directory, and produces a ready-to-fill skeleton.

## Why This Exists

New scaffold assets fail doc-health when authors guess at the frontmatter
schema, pick an unsafe `name`, or drop a draft into the wrong directory. By
the time the gate flags it, the file already exists with the wrong shape.

This skill prevents that: it gathers every required field with inline
validation, refuses an unsafe name **before** composing any path, and writes
the draft to one canonical location (`alto-scaffold/lifecycle/in-progress/`).
The result passes doc-health on the first run.

## Usage

```
/write-a-workflow-asset
/write-a-workflow-asset my-new-command
```

An optional argument pre-seeds the `name` field. Everything else is gathered
interactively. The skill never writes a file until all 8 fields validate.

## The Canonical Frontmatter Schema

Every workflow asset MUST declare these 8 fields. None has a default — the
author enumerates each one explicitly.

```yaml
---
name: <kebab-case, ^[a-z][a-z0-9-]*$>
description: <one sentence>
kind: command | agent | template | skill
phase: design | groom | implement | review | close
when_to_use: <trigger phrase or example request>
tools_required: Read, Grep, Write
bash_substitution_policy: none | quoted | unrestricted
license: Apache-2.0
---
```

Field reference:

| Field | Rule |
|-------|------|
| `name` | Must match `^[a-z][a-z0-9-]*$` and equal the file basename. Lowercase start, then lowercase letters / digits / dashes. No dots, slashes, spaces, or uppercase. |
| `description` | One sentence, non-empty. Keep it under ~1,500 characters (authoring guidance — not a hard gate). |
| `kind` | One of `command`, `agent`, `template`, `skill`. |
| `phase` | One of `design`, `groom`, `implement`, `review`, `close`. |
| `when_to_use` | Non-empty trigger phrase describing when to reach for the asset. |
| `tools_required` | Non-empty, **inline comma-separated string** (`Read, Grep, Write`). Do NOT use a YAML block list (`- Read`) — the schema check treats a non-string value as a missing field and errors. |
| `bash_substitution_policy` | One of `none`, `quoted`, `unrestricted`. See the safety section below. |
| `license` | Non-empty SPDX identifier. Default to `Apache-2.0`. |

## Process

### Phase 1 — Gather Inputs

Prompt the contributor for each field in order. Validate **inline**, one field
at a time, and re-prompt on any failure instead of accepting a bad value.

1. **`name`** — If an argument was supplied, use it as the proposed name;
   otherwise ask. Validate against `^[a-z][a-z0-9-]*$` immediately (see Phase 2
   — this is the security boundary). Reject, do not sanitise.
2. **`description`** — Ask for one sentence. Reject empty. Note the ~1,500
   character guideline if the answer is very long, but do not block on it.
3. **`kind`** — Offer `command | agent | template | skill`. Reject any other
   value.
4. **`phase`** — Offer `design | groom | implement | review | close`. Reject
   any other value.
5. **`when_to_use`** — Ask for the trigger phrase. Reject empty.
6. **`tools_required`** — Ask for the tool list and record it as an **inline
   comma-separated string** (`Read, Grep, Write`), never a YAML block list — a
   non-string value fails the required-field check. Reject an empty list. For
   each entry outside the common set
   `Read, Grep, Glob, Edit, Write, Bash, Agent`, **warn and ask the contributor
   to confirm** (it may be a typo, an MCP tool, or a less-common native tool).
   Tool names should look like PascalCase (`Read`, `WebFetch`) or the MCP form
   `mcp__<server>__<tool>`; flag anything that does not.
7. **`bash_substitution_policy`** — Offer `none | quoted | unrestricted`.
   Default the suggestion to `none`. Explain the consequences (safety section).
8. **`license`** — Ask for the SPDX identifier; default to `Apache-2.0` if the
   contributor has no preference. Reject empty.

### Phase 2 — Validate

Run these checks in order. Stop and report on the first failure.

1. **Name regex (FIRST — the sole path-traversal defence).** Confirm `name`
   matches `^[a-z][a-z0-9-]*$`. This MUST happen before any path is composed.
   It rejects `.`, `/`, `\`, spaces, uppercase, and leading digits or dashes.
   Examples that MUST be rejected:
   - `My Command` (uppercase + space)
   - `../commands/evil` (slashes + dots — path traversal)
   - `-leading-dash` (leading dash)
   - `1stthing` (leading digit)

   On rejection, print the allowed character class and abort. Never rewrite the
   name to make it pass (no `/` → `-`, no case-folding).

2. **Collision check across all asset directories.** Look for an existing
   `<name>.md` in each of these five sibling directories:
   - `alto-scaffold/lifecycle/in-progress/`
   - `alto-scaffold/commands/`
   - `alto-scaffold/agents/`
   - `alto-scaffold/templates/`
   - `alto-scaffold/skills/`

   Use Glob (one pass per directory, or a combined pattern) to detect any
   match. If `<name>.md` already exists in any of them, show the matching
   path(s) and **confirm-or-abort** — do not silently overwrite.

3. **Path-traversal re-check after composing the path.** Compose the target as
   `alto-scaffold/lifecycle/in-progress/<name>.md`. Verify the cleaned path is a
   **direct child** of `alto-scaffold/lifecycle/in-progress/` — it contains no
   `..` segments and its parent directory is exactly that folder. If not, abort.

4. **Bash + parameters safety review.** If `tools_required` contains `Bash`:
   - With `bash_substitution_policy: none`, the body may contain **no** bash
     fence at all. That covers fenced blocks tagged bash, sh, zsh, shell, or
     console; a fence tagged with an exclamation mark; and the inline
     exclamation-backtick command form. Any such block is a doc-health ERROR.
   - If the asset also declares `parameters:` (user-supplied arguments), warn
     that `disable_model_invocation: true` is normally required so the model
     cannot invoke Bash with attacker-controlled arguments.
   - Every shell-variable substitution inside any bash block MUST be quoted —
     unquoted `$VAR`, `$ARGUMENTS`, `$N`, or `${...}` forms are a hard ERROR
     regardless of policy. Teach the contributor to wrap substitutions in
     double quotes.

### Phase 3 — Assemble and Write

1. Compose the YAML frontmatter from the 8 validated fields, in canonical
   order: `name`, `description`, `kind`, `phase`, `when_to_use`,
   `tools_required`, `bash_substitution_policy`, `license`.
2. Append a skeleton body with these sections (so the draft is editable, not
   empty):
   - `# /<name>` (or the asset title) and a one-line summary.
   - `## Why This Exists` — the problem the asset solves.
   - `## Usage` — how to invoke it (plain fenced block, no shell language tag
     if the policy is `none`).
   - `## Process` — numbered phases the asset performs.
   - `## Rules` — invariants and safety notes.
3. Use the **Write** tool to write the composed content to
   `alto-scaffold/lifecycle/in-progress/<name>.md`. This is the only write the
   skill performs.

A minimal example of the frontmatter the skill assembles:

```yaml
---
name: my-new-command
description: One-sentence description of what this command does
kind: command
phase: implement
when_to_use: When a contributor needs to do the thing this command automates
tools_required: Read, Write
bash_substitution_policy: none
license: Apache-2.0
---
```

### Phase 4 — Post-Write Guidance

After the file is written, print the next steps for the contributor:

1. **Fill the body.** Replace the skeleton sections with real content. Keep the
   body under 500 lines (doc-health warns above that).
2. **Validate.** Run the doc-health check against the in-progress folder:

   ```
   alto doc-health --paths=alto-scaffold/lifecycle/in-progress/
   ```

   Fix every ERROR before promoting. WARNINGs (body size, uncommon tools,
   `unrestricted` policy) do not block but should be reviewed.
3. **Promote** once the draft is stable. Move it from the in-progress folder
   into the matching directory for its `kind` and preserve history:

   ```
   git mv alto-scaffold/lifecycle/in-progress/<name>.md alto-scaffold/commands/<name>.md
   ```

   Use `commands/` for `kind: command`, `agents/` for `kind: agent`,
   `templates/` for `kind: template`, `skills/` for `kind: skill`.
4. **Symlink** (commands and agents only) so the asset is reachable from the
   tool config tree. Create a relative-target symlink, for example:

   ```
   ln -s ../../alto-scaffold/commands/<name>.md .claude/commands/<name>.md
   ```

   Verify the link target resolves with `readlink`.

## Rules

- **Validate the name first, always.** The `^[a-z][a-z0-9-]*$` check is the
  only thing standing between contributor input and a composed filesystem path.
  It runs before any path work and rejects — never sanitises — bad names.
- **One write, one location.** Drafts go to
  `alto-scaffold/lifecycle/in-progress/<name>.md` and nowhere else. Promotion to
  `commands/`, `agents/`, `templates/`, or `skills/` is a separate, explicit
  `git mv` the contributor performs after the draft is stable.
- **All 8 frontmatter fields are required and non-empty.** There are no silent
  defaults; the author enumerates each value.
- **Respect the bash policy.** With `bash_substitution_policy: none` the body
  carries zero executable bash blocks. With `quoted`, every substitution inside
  a bash block must be double-quoted. Reserve `unrestricted` for assets that
  also set `disable_model_invocation: true`.
- **No silent overwrite.** A name that collides with an existing asset in any of
  the five directories requires explicit confirmation before proceeding.
- **Keep GENERIC bodies portable.** Do not reference alto's own source tree,
  module path, or build tooling in a generic asset body — those belong in a
  `.project.md` overlay, not the shared scaffold.
