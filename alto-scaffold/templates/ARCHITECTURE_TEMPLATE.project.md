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
