---
name: doc-health
description: Check documentation health - freshness, broken links, missing metadata
---

# /doc-health

Run documentation health checks to identify stale or incomplete docs.

## Implementation

Scan all markdown files in `docs/` for:

1. **Frontmatter completeness** — `last_reviewed`, `owner`, `status` fields
2. **Freshness** — Docs older than 30 days flagged as stale
3. **File references** — Validates that referenced source files exist

### Key Documents to Track

Check these docs exist and are current:

| Document | Expected |
|----------|----------|
| `docs/PRD.md` | Product requirements |
| `docs/DDD.md` | Domain model and bounded contexts |
| `docs/ARCHITECTURE.md` | Technical architecture |

### Steps

1. Use `Glob` to find all `docs/**/*.md` files
2. Use `Read` to check each file for frontmatter with `last_reviewed` date
3. Flag any doc older than 30 days as stale
4. Check that `docs/PRD.md`, `docs/DDD.md`, and `docs/ARCHITECTURE.md` exist
5. Report findings

## Response Format

```
============================================================
DOC HEALTH REPORT
============================================================

OK docs/PRD.md
  Last reviewed: 2026-02-20 (2d ago)

WARNING docs/ARCHITECTURE.md
  Last reviewed: 2026-01-10 (43d ago)
  Stale: 43 days since review (threshold: 30)

MISSING docs/DDD.md
  This document should exist but was not found.

============================================================
X issue(s) found
============================================================
```

## Fixing Issues

### Stale Document

Update the frontmatter:

```yaml
---
last_reviewed: 2026-02-22
owner: team-name
status: current
---
```

### Missing Document

Create from the appropriate template in `docs/templates/`.

## When to Run

- **Before starting work** — Check if relevant docs are current
- **After major changes** — Update frontmatter dates
- **Before closing an epic** — Ensure all docs are fresh
