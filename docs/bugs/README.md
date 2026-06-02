---
last_reviewed: 2026-06-02
---

# Bug RCA documents

This directory holds **root-cause-analysis documents** for S0 and S1 bugs
(see `alto-scaffold/templates/beads-bug-template.md` for the severity
matrix). One file per bug, written with the blameless Five Whys discipline.

## Naming convention

```
docs/bugs/YYYYMMDD_<bug-id>_<slug>.md
```

- `YYYYMMDD` — date the bug was first investigated, no separators
  (matches the convention used by `docs/research/` spike reports so a
  chronological `ls` is the timeline).
- `<bug-id>` — the beads ticket ID (e.g. `alty-cli-6n4`).
- `<slug>` — lowercase-kebab summary stripped of `fix(...):` prefix,
  punctuation, and articles. Keep ≤ 50 chars.

Example: `docs/bugs/20260602_alty-cli-6n4_bd-close-hook-stale-path.md`.

## Which bugs get a doc?

| Severity | Doc required? |
|----------|---------------|
| S0       | Yes — outage, data loss, security breach |
| S1       | Yes — major feature broken, no workaround |
| S2       | Optional — inline `## Root Cause` + `## Verification` on the bug ticket is sufficient |
| S3       | No — inline only |

A small fraction of S2 bugs deserve a full doc when investigation reveals
unexpected systemic causes. Promote at the user's call, not by default.

## Lifecycle

Generated and walked by `/rca <bug-id>`. The doc moves through
`Investigating → Root cause identified → Fixed → Verified → Closed` as
the fix lands. The final state is a self-contained post-mortem readable
by someone with no prior context.

## Freshness

These documents are **archival**. Once closed, they describe a moment in
time and are not subject to the `last_reviewed` cadence checks that apply
to living docs (PRD, DDD, ARCHITECTURE). The directory is exempted from
the freshness scanner the same way `docs/research/` and `docs/plans/` are
— see `internal/dochealth/application/doc_health_handler.go`.

> If a bug recurs after its RCA closed, **file a new bug + new RCA** and
> cross-reference the original. Never re-open the old doc — the original
> is the historical record of what was true at fix time.

## Related templates

- `alto-scaffold/templates/beads-bug-template.md` — the ticket-side
  template that pairs with these docs.
- `alto-scaffold/templates/bug-rca-template.md` — the body that gets
  copied here.
- `alto-scaffold/commands/rca.md` — the `/rca` driver command.
