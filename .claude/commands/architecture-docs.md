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

<!-- CUSTOMIZE: Add your architecture docs as the project grows -->
<!-- Example entries:
| Components / Services | `docs/architecture/components.md` |
| Messaging / Events | `docs/architecture/messaging.md` |
| API Design | `docs/architecture/api.md` |
| Data Model | `docs/architecture/data-model.md` |
| Deployment | `docs/architecture/deployment.md` |
-->

## Keyword to Document Mapping

| Keywords | Document |
|----------|----------|
| domain, model, entity, value object, aggregate | `docs/DDD.md` |
| bounded context, ubiquitous language, subdomain | `docs/DDD.md` |
| architecture, design, principle, layer | `docs/ARCHITECTURE.md` |
| requirement, constraint, user story, scenario | `docs/PRD.md` |

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
