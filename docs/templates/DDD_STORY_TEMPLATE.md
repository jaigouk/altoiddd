---
last_reviewed: YYYY-MM-DD
owner: architecture
status: draft
---

# Domain-Driven Design Artifacts: [Project Name]

> **Purpose:** This document captures domain knowledge BEFORE writing architecture
> or code. It ensures we build the right thing by understanding the business domain
> first, using Domain Storytelling and DDD strategic patterns.

## 1. Domain Stories

Domain Stories are visual narratives of how domain experts describe their work.
Write them as step-by-step flows using domain language (not technical terms).

### Story 1: [Name of business process]

**Actors:** [Who is involved — use domain terms, not "User"]
**Trigger:** [What starts this process]

```
1. [Actor] [verb] [work object]
2. [Actor] [verb] [work object] using [tool/system]
3. [System] [verb] [work object] to [Actor]
4. [Actor] [verb] [work object]
```

**Key observations:**
- What vocabulary did the domain expert use?
- What surprised you?
- What was ambiguous?

### Story 2: [Name of business process]

_Repeat the same structure for each core business flow._

## 2. Ubiquitous Language Glossary

> **Rule:** These terms MUST be used identically in code (class names, method names,
> variable names). If the business says "Approve Policy", the code says `approve_policy()`,
> NOT `update_status()`.

| Term | Definition | Context / Bounded Context |
|------|-----------|---------------------------|
| [Term 1] | What it means in this domain | Where it applies |
| [Term 2] | What it means in this domain | Where it applies |
| [Term 3] | What it means in this domain | Where it applies |

**Ambiguous terms** (same word, different meaning in different contexts):

| Term | Context A Meaning | Context B Meaning |
|------|------------------|------------------|
| [Term] | Meaning in Context A | Meaning in Context B |

## 3. Subdomain Classification

Classify each area of the domain to determine investment level:

| Subdomain | Type | Rationale | Architecture Approach |
|-----------|------|-----------|----------------------|
| [Subdomain 1] | **Core** | This is our competitive advantage | Rich Domain Model (DDD tactical) |
| [Subdomain 2] | **Supporting** | Necessary but not differentiating | Simpler architecture (Active Record) |
| [Subdomain 3] | **Generic** | Standard industry problem | Buy/use existing (CRUD, library) |

**Core Domain** gets the most investment:
- Strict DDD tactical patterns (Aggregates, Value Objects, Domain Events)
- Highest test coverage
- Best developer talent
- Most careful design

**Supporting/Generic** can be simpler:
- Active Record or Transaction Script patterns
- Standard CRUD where appropriate
- Third-party solutions preferred for Generic

## 4. Bounded Contexts

> A Bounded Context is an explicit boundary around a domain model where terms
> have specific, unambiguous meanings.

### Context: [Name]

**Responsibility:** What this context owns and manages.

**Key domain objects:**
- [Entity/Aggregate 1] — description
- [Value Object 1] — description
- [Domain Event 1] — description

**External dependencies:** What other contexts does this one need?

### Context: [Name]

_Repeat for each bounded context._

### Context Map (Relationships)

Describe how bounded contexts communicate:

```
[Context A] ──── Published Language ────> [Context B]
[Context C] ──── Anticorruption Layer ──> [External System]
[Context D] ──── Shared Kernel ─────────> [Context E]
```

| Upstream Context | Downstream Context | Integration Pattern |
|-----------------|-------------------|-------------------|
| [Context A] | [Context B] | Published Language / Domain Events |
| [External System] | [Context C] | Anticorruption Layer |

## 5. Aggregate Design

For each **Core Domain** bounded context, identify aggregates:

### Aggregate: [Name] (in [Context Name])

**Aggregate Root:** [Entity name]

**Contains:**
- [Entity/Value Object 1]
- [Entity/Value Object 2]

**Invariants (business rules this aggregate protects):**
1. [Rule 1 — e.g., "An order cannot have negative quantity"]
2. [Rule 2 — e.g., "Total cannot exceed credit limit"]

**Commands (things you can ask this aggregate to do):**
- `[command_name]` — description

**Domain Events (things this aggregate announces):**
- `[EventName]` — when [trigger]

**Design rules:**
- Reference other aggregates by ID only
- One aggregate per transaction
- The aggregate root is the only entry point

## 6. Event Storming Summary (Optional)

If you ran an Event Storming session, capture the key artifacts:

### Domain Events (orange sticky notes)

| Event | Triggered By | Bounded Context |
|-------|-------------|-----------------|
| [EventName] | [Command or external trigger] | [Context] |

### Commands (blue sticky notes)

| Command | Actor | Triggers Event |
|---------|-------|---------------|
| [CommandName] | [Who initiates] | [EventName] |

### Read Models / Queries (green sticky notes)

| Query | Purpose | Data Source |
|-------|---------|-------------|
| [QueryName] | What information is needed | [Aggregate/Context] |

## 7. DDD Checklist

Before proceeding to architecture:

- [ ] All domain stories written with domain expert vocabulary
- [ ] Ubiquitous language glossary complete — no ambiguous terms
- [ ] Subdomains classified (Core / Supporting / Generic)
- [ ] Bounded contexts identified with clear boundaries
- [ ] Context map shows all relationships and integration patterns
- [ ] Aggregates designed for Core domain with invariants documented
- [ ] No technical terms leaked into domain language
- [ ] Domain experts would recognize and agree with this document

## 8. Open Questions for Domain Experts

- [ ] Question 1 — What happens when [edge case]?
- [ ] Question 2 — Is [term] the same as [other term]?
- [ ] Question 3 — Who is responsible for [process]?
