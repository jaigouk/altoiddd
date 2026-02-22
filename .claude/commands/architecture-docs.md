---
name: architecture-docs
description: Access project architecture documentation using RLM pattern
allowed-tools: Read
---

# /architecture-docs <topic>

Access architecture knowledge using the **RLM pattern**: documents are addressable variables, not search targets.

## Usage

```
/architecture-docs <topic>
```

## RLM Approach

**Do NOT use Grep/search.** Instead:
1. Identify the topic from the query
2. Look up the file path from the knowledge map below
3. Read the file directly using the Read tool
4. Analyze content and cite the source

## Knowledge Map (Documents as Variables)

### Architecture Documents

| Topic | Variable Path |
|-------|---------------|
| Overview / Design Principles | `docs/ARCHITECTURE.md` |
| Domain Model / Bounded Contexts | `docs/DDD.md` |
| Product Requirements | `docs/PRD.md` |
| PRD Capability → Ticket Map | `.claude/commands/prd-traceability.md` |
| DDD Question Framework | `docs/research/20260222_guided_ddd_question_framework_consolidated.md` |
| Subdomain Classification | `docs/DDD.md` → Section 3 (Complexity Budget YAML) |
| Context Map | `docs/DDD.md` → Section 4 (Bounded Contexts + Relationships) |

<!-- CUSTOMIZE: Add your architecture docs as the project grows -->

## Keyword to Document Mapping

| Keywords | Document |
|----------|----------|
| domain, model, entity, value object, aggregate | `docs/DDD.md` |
| bounded context, ubiquitous language, subdomain | `docs/DDD.md` |
| complexity budget, treatment level, core/supporting/generic | `docs/DDD.md` → Section 3 |
| context map, integration pattern, upstream, downstream | `docs/DDD.md` → Section 4 |
| architecture, design, principle, layer | `docs/ARCHITECTURE.md` |
| requirement, constraint, user story, scenario | `docs/PRD.md` |
| capability, coverage, traceability, gap | `.claude/commands/prd-traceability.md` |
| question framework, dual register, persona, discovery | `docs/research/20260222_guided_ddd_question_framework_consolidated.md` |

<!-- CUSTOMIZE: Add project-specific keyword mappings -->

## Implementation

When `/architecture-docs <topic>` is invoked:

### Step 1: Identify Topic Type

Parse the query to determine which document to access using the keyword mapping above.

### Step 2: Direct File Access

Use the Read tool to load the document:

```
Read docs/ARCHITECTURE.md
```

### Step 3: Analyze and Respond

Once content is loaded, extract and cite:

```markdown
## [Topic]

**Source:** `docs/ARCHITECTURE.md` → [section]

> [Quoted text from document]

[Analysis of relevant sections]
```

## Notes

- **Never search** — Use direct file access based on the lookup table
- **Always cite** — Include the file path in your response
- **Quote text** — Use blockquotes for direct citations from sources
