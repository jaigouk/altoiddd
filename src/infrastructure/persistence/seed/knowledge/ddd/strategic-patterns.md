---
last_verified: "2026-03-04"
confidence: high
next_review_date: "2026-06-04"
---

# DDD Strategic Patterns

Quick reference for AI agents generating project structure and bounded context maps.

## Bounded Contexts

An explicit boundary within which a particular domain model applies.

**Practical guidance:**
- Each bounded context gets its own directory under `src/`
- Models do NOT cross context boundaries -- use IDs or translation layers
- If two teams would argue about what a term means, you have two contexts

## Context Map Relationships

| Pattern | Description | When to use |
|---------|-------------|-------------|
| **Shared Kernel** | Shared code both contexts depend on | Small teams, tightly coupled |
| **Customer/Supplier** | Upstream provides, downstream consumes | Producer-consumer |
| **Anticorruption Layer** | Translation layer between contexts | Upstream is messy or foreign |
| **Open Host Service** | Published API for consumers | Multiple downstream consumers |
| **Separate Ways** | No integration | Integration cost exceeds value |

**Default:** Use Anticorruption Layer when integrating with external systems.

## Ubiquitous Language

Shared vocabulary between domain experts and developers, used in code, docs, and conversation.

**Rules:**
- Class/method names MUST use domain terms
- Domain expert says "submit an order" -> method is `order.submit()`, not `order.process()`
- Glossary lives in `docs/DDD.md`

**Red flags:** `Manager`, `Handler`, `Processor` classes (not domain language).

## Subdomains

| Type | Investment | Code quality |
|------|-----------|-------------|
| **Core** | Highest effort, best developers | Rich domain model, full DDD |
| **Supporting** | Moderate effort | Simpler patterns, still tested |
| **Generic** | Minimal custom code | Thin wrappers around libraries |

Core subdomain = competitive advantage. Generic = commodity (buy or use a library).
