---
last_verified: "2026-03-04"
confidence: high
next_review_date: "2026-06-04"
---

# DDD Tactical Patterns

Quick reference for AI agents generating domain model code.

## Entities

Objects defined by a unique identity that persists through state changes.

**When to use:** The object has a lifecycle and must be tracked over time.

Pattern: `@dataclass` with an `id` field and behavior methods that enforce state transitions.

**Mistakes:** Using DB IDs as domain identity; anemic getters/setters; exposing mutable internals.

## Value Objects

Immutable objects defined entirely by their attributes. No identity.

**When to use:** The concept is defined by what it IS, not which one it is (e.g., Money, Email).

Pattern: `@dataclass(frozen=True)` with domain logic methods that return new instances.

**Mistakes:** Mutable value objects; adding an `id` field; putting logic in services.

## Aggregates

Cluster of entities and value objects with a single root entity enforcing consistency.

**Rules:**
- One root entity per aggregate; external references use the root's ID only
- Only the root is accessible from outside the aggregate
- One aggregate per transaction

## Domain Services

Stateless operations spanning multiple aggregates or not naturally belonging to one.

**When to use:** The operation spans aggregates, or forcing it onto an entity feels unnatural.

**Mistakes:** Logic that belongs on an entity (anemic model); stateful services; depending on infrastructure.

## Repositories

Collection-like interface for persisting and retrieving aggregate roots.

```python
class OrderRepository(Protocol):
    def find_by_id(self, order_id: OrderId) -> Order | None: ...
    def save(self, order: Order) -> None: ...
```

**Mistakes:** Repos for non-root entities; leaking persistence details; domain depending on infrastructure.
