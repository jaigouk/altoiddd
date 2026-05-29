---
last_reviewed: YYYY-MM-DD
owner: architecture
status: draft
---

# Architecture: [Project Name]

> **Prerequisites:** This document should be written AFTER `docs/PRD.md` and `docs/DDD.md`.
> Architecture decisions must be informed by domain knowledge, not the other way around.

## 1. Design Principles

List the guiding principles for architectural decisions:

1. **[Principle 1]** — Description (e.g., "Domain purity — domain layer has zero external dependencies")
2. **[Principle 2]** — Description (e.g., "Local-first — minimize cloud dependencies")
3. **[Principle 3]** — Description
4. **DDD alignment** — Architecture follows bounded context boundaries from `docs/DDD.md`
5. **Testability** — Every component testable in isolation with dependency injection

## 2. System Overview

### High-Level Diagram

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  [Input]    │────>│  [Process]   │────>│  [Output]       │
│             │     │              │     │                 │
└─────────────┘     └──────────────┘     └─────────────────┘
```

### Component Summary

| Component     | Responsibility | Bounded Context                 |
| ------------- | -------------- | ------------------------------- |
| [Component 1] | What it does   | Which DDD context it belongs to |
| [Component 2] | What it does   | Which DDD context it belongs to |

## 3. Layer Architecture

Following Hexagonal / Clean Architecture aligned with DDD:

```
┌─────────────────────────────────────────────┐
│              Infrastructure                 │
│  ┌───────────────────────────────────────┐  │
│  │           Application                 │  │
│  │  ┌───────────────────────────────┐    │  │
│  │  │         Domain                │    │  │
│  │  │  (Entities, Value Objects,    │    │  │
│  │  │   Aggregates, Domain Services,│    │  │
│  │  │   Domain Events)              │    │  │
│  │  └───────────────────────────────┘    │  │
│  │  (Commands, Queries, Ports)           │  │
│  └───────────────────────────────────────┘  │
│  (Persistence, Messaging, External APIs)    │
└─────────────────────────────────────────────┘
```

### Layer Rules

| Layer          | Can Depend On              | Cannot Depend On                        |
| -------------- | -------------------------- | --------------------------------------- |
| Domain         | Nothing (pure Python)      | Application, Infrastructure, frameworks |
| Application    | Domain, Ports (interfaces) | Infrastructure, frameworks              |
| Infrastructure | Application, Domain        | — (outermost layer)                     |

### Source Layout

```
src/
├── domain/
│   ├── models/          # Entities, Value Objects, Aggregates
│   ├── services/        # Domain Services
│   └── events/          # Domain Events
├── application/
│   ├── commands/        # Command handlers (write operations)
│   ├── queries/         # Query handlers (read operations)
│   └── ports/           # Interfaces (Protocols) for infrastructure
└── infrastructure/
    ├── persistence/     # Database adapters
    ├── messaging/       # Message bus adapters
    └── external/        # External API clients
```

## 4. Bounded Context Integration

How bounded contexts communicate (from `docs/DDD.md` context map):

| From Context | To Context  | Mechanism                  | Data Format |
| ------------ | ----------- | -------------------------- | ----------- |
| [Context A]  | [Context B] | [Events / API / Shared DB] | [Format]    |

## 5. Data Model

### Aggregates and Storage

| Aggregate     | Storage                    | Rationale        |
| ------------- | -------------------------- | ---------------- |
| [Aggregate 1] | [PostgreSQL / SQLite / KV] | Why this storage |

### Key Entities

| Entity     | Attributes | Aggregate       |
| ---------- | ---------- | --------------- |
| [Entity 1] | Key fields | Which aggregate |

## 6. External Integrations

| Integration | Purpose      | Protocol          | Auth            |
| ----------- | ------------ | ----------------- | --------------- |
| [Service 1] | What it does | REST / gRPC / etc | API key / OAuth |

## 7. Security

### Trust Boundaries

```
[Untrusted] ──── Validation ────> [Trusted Internal]
```

### Security Measures

| Concern          | Mitigation |
| ---------------- | ---------- |
| Input validation | [Approach] |
| Authentication   | [Approach] |
| Authorization    | [Approach] |
| Data protection  | [Approach] |

## 8. Deployment

<!-- CUSTOMIZE: Fill in your deployment approach -->

| Aspect          | Choice      | Rationale              |
| --------------- | ----------- | ---------------------- |
| Runtime         | Python 3.12 | Project standard       |
| Package manager | uv          | Speed, reproducibility |
| [Other]         | [Choice]    | [Why]                  |

## 9. Constraints & Budgets

From `docs/PRD.md`:

| Resource     | Limit   | Rationale |
| ------------ | ------- | --------- |
| [Resource 1] | [Limit] | [Why]     |

## 10. Open Architecture Decisions

Decisions that need spikes before committing:

- [ ] ADR-001: [Decision needed] — Spike: [link to spike ticket]
- [ ] ADR-002: [Decision needed] — Spike: [link to spike ticket]

## 11. Architecture Decision Records

| ADR     | Decision | Status                           |
| ------- | -------- | -------------------------------- |
| ADR-001 | [Title]  | Proposed / Accepted / Superseded |
