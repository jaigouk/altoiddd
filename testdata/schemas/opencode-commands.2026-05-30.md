# OpenCode Custom Commands — Frontmatter Schema Snapshot

Fetched: 2026-05-30
Source: https://opencode.ai/docs/commands
Retrieved via: context7 (`/websites/opencode_ai`)

This snapshot is committed verbatim to detect schema drift between alto's
`OpenCodeCommandAdapter` and the upstream OpenCode docs. If a schema change
is detected, update the adapter transformation table and re-pin to a new
dated file under `testdata/schemas/`.

## Storage location

- Per-project: `.opencode/commands/<name>.md`
- Global: `~/.config/opencode/commands/<name>.md`

The filename (without `.md`) becomes the command's invocation name.

## Frontmatter schema

```md
---
description: <one-line summary>      # required for OpenCode UI
agent: <agent-name>                  # optional — pins this command to a named agent
model: <provider/model>              # optional — overrides default model for this command
---

<body>
```

**Observed fields (only these are honoured by OpenCode commands):**

| Field         | Required | Notes                                                                 |
|---------------|----------|-----------------------------------------------------------------------|
| `description` | yes      | shown in the OpenCode command palette                                 |
| `agent`       | no       | pin invocation to a specific agent (e.g. `build`, `plan`)             |
| `model`       | no       | `provider/model` form; overrides project default                      |

**Fields NOT recognised on individual command files** (verified 2026-05-30):

- `tools` — tool restriction is configured per-agent in `opencode.json`, not
  per-command. The agent named in `agent:` controls tool access. The portable
  `tools` field in `alto-scaffold/commands/*.md` therefore has no direct
  OpenCode-command equivalent and is dropped at translation time. If a future
  schema revision adds `tools:` to commands, update the transformation table.
- `disable_model_invocation` — OpenCode has no equivalent gating field. Per
  ticket alty-cli-766.5 AC FIX-1, the adapter SKIPS rendering for any source
  asset with `disable_model_invocation: true` AND returns
  `ErrInvocationProtectionNotSupported` (aggregated, non-fatal) so the handler
  can log the skip list.
- `globs` — OpenCode commands are invoked explicitly (`/<name>`) or by the
  model when bound to an agent; there is no glob auto-attach. Dropped at
  translation time.

## Body / prompt syntax

OpenCode commands support the following in the body, but the adapter
TRANSFORMS them per alto's portability rules:

| Construct                  | OpenCode behaviour                                  | Adapter behaviour (per ticket alty-cli-766.5)                                                                  |
|----------------------------|------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `$ARGUMENTS`, `$1`, `$2`   | Argument substitution at invocation                  | Passed through verbatim                                                                                         |
| `` !`<cmd>` ``             | Inline bash output injection                         | Replaced with HTML comment naming the original command verbatim (quoting semantics differ from Claude Code)     |
| ` ```! ` fenced block       | Multi-line bash output injection                     | Replaced with HTML comment naming the original block verbatim                                                   |
| `@filename`                | File content include at runtime                      | Passed through verbatim                                                                                         |
| `${CLAUDE_SKILL_DIR}/...`  | Not recognised by OpenCode                           | Template reference is INLINED at translation time (per ticket FIX-4); missing template → `ErrMissingTemplate`   |

## Source citations

- "Custom commands are stored in `.opencode/commands/` directory ... filename
  becomes the command name." — https://opencode.ai/docs/commands
- "Markdown command files can be placed in global
  (`~/.config/opencode/commands/`) or per-project (`.opencode/commands/`)
  directories." — https://opencode.ai/docs/commands
- "Prompts for custom commands support special placeholders and syntax:
  `$ARGUMENTS` for all arguments, `$1`, `$2`, etc., for positional arguments,
  `` !`command` `` to inject bash command output, and `@filename` to include
  file content." — https://opencode.ai/docs/commands

## Example (canonical)

```md
---
description: Run tests with coverage
agent: build
model: anthropic/claude-3-5-sonnet-20241022
---

Run the full test suite with coverage report and show any failures.
Focus on the failing tests and suggest fixes.
```
