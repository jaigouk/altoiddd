---
last_verified: "2026-03-04"
confidence: high
next_review_date: "2026-06-04"
---

# SOLID Principles

Quick reference for AI agents writing domain-driven Python code.

## Single Responsibility (SRP)

A class should have only one reason to change. Separate validation, persistence, and notification into distinct classes.

**Violation:** `OrderService` that validates, persists, AND sends email.

## Open/Closed (OCP)

Open for extension, closed for modification. Use composition, not `if/elif` chains.

```python
class DiscountStrategy(Protocol):
    def calculate(self, order: Order) -> Money: ...

class PricingService:
    def __init__(self, discounts: list[DiscountStrategy]) -> None:
        self._discounts = discounts  # New type = new class, no modification
```

**Violation:** `if discount_type == "percentage" ... elif "fixed"` -- every new type modifies existing code.

## Liskov Substitution (LSP)

Subtypes must be substitutable for their base types without altering correctness.

**Violation:** `ReadOnlyRepository.save()` raising `NotImplementedError`.

## Interface Segregation (ISP)

Prefer focused Protocol classes. Clients should not depend on methods they do not use.

```python
class OrderReader(Protocol):
    def find_by_id(self, id: OrderId) -> Order | None: ...

class OrderWriter(Protocol):
    def save(self, order: Order) -> None: ...
```

## Dependency Inversion (DIP)

High-level modules depend on abstractions, not low-level modules.

```python
class ProjectRepository(Protocol):
    def save(self, project: Project) -> None: ...

class CreateProjectHandler:
    def __init__(self, repo: ProjectRepository) -> None:
        self._repo = repo  # Depends on abstraction, not FileSystemRepo
```
