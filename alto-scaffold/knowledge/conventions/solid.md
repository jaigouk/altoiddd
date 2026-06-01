# SOLID Principles

Quick reference for AI agents writing domain-driven Python code.

---

## Single Responsibility Principle (SRP)

A class should have only one reason to change.

**Correct:** Separate `OrderValidator` (validation rules) and `OrderNotifier` (notification logic) -- each has one reason to change.

```python
class OrderValidator:
    def validate(self, order: Order) -> list[str]:
        errors = []
        if not order.items:
            errors.append("Order must have at least one item")
        return errors
```

**Violation:** `OrderService` that validates, persists, AND sends email -- three reasons to change.

---

## Open/Closed Principle (OCP)

Open for extension, closed for modification. Use composition, not `if/elif` chains.

**Correct:**

```python
class DiscountStrategy(Protocol):
    def calculate(self, order: Order) -> Money: ...

class PercentageDiscount:
    def __init__(self, rate: Decimal) -> None:
        self._rate = rate
    def calculate(self, order: Order) -> Money:
        return Money(order.total().amount * self._rate, order.total().currency)

class PricingService:
    def __init__(self, discounts: list[DiscountStrategy]) -> None:
        self._discounts = discounts  # New discount types = new class, no modification
```

**Violation:** `if discount_type == "percentage" ... elif "fixed" ... elif "bogo"` -- every new type modifies existing code.

---

## Liskov Substitution Principle (LSP)

Subtypes must be substitutable for their base types without altering correctness.

**Correct:**

```python
class Repository(Protocol):
    def find_by_id(self, id: str) -> Order | None: ...
    def save(self, order: Order) -> None: ...

class InMemoryRepository:
    """Substitutable anywhere Repository is expected."""
    def __init__(self) -> None:
        self._store: dict[str, Order] = {}

    def find_by_id(self, id: str) -> Order | None:
        return self._store.get(id)

    def save(self, order: Order) -> None:
        self._store[str(order.id)] = order
```

**Violation:** `ReadOnlyRepository.save()` raising `NotImplementedError` -- breaks callers expecting save to work.

---

## Interface Segregation Principle (ISP)

Prefer focused Protocol classes over fat interfaces. Clients should not depend on methods they do not use.

**Correct:**

```python
class OrderReader(Protocol):
    def find_by_id(self, id: OrderId) -> Order | None: ...
    def list_by_customer(self, customer_id: CustomerId) -> list[Order]: ...

class OrderWriter(Protocol):
    def save(self, order: Order) -> None: ...
    def delete(self, id: OrderId) -> None: ...

# Query handlers only need OrderReader
class GetOrderHandler:
    def __init__(self, reader: OrderReader) -> None:
        self._reader = reader
```

**Violation:** Single `OrderRepository` Protocol with `find`, `save`, `delete`, `export_to_csv`, AND `run_migration` -- forces all clients to depend on methods they never use.

---

## Dependency Inversion Principle (DIP)

High-level modules should not depend on low-level modules. Both should depend on abstractions.

**Correct:**

```python
class ProjectRepository(Protocol):  # Port (abstraction)
    def save(self, project: Project) -> None: ...
    def find_by_id(self, project_id: ProjectId) -> Project | None: ...

class CreateProjectHandler:  # Depends on abstraction
    def __init__(self, repo: ProjectRepository) -> None:
        self._repo = repo

class FileSystemProjectRepository:  # Infrastructure implements abstraction
    def save(self, project: Project) -> None:
        self._base_path.joinpath(f"{project.id}.json").write_text(project.to_json())
```

**Violation:** `CreateProjectHandler.__init__` creating `FileSystemProjectRepository("/data/projects")` directly.

---

## Quick Decision Guide

| Smell | Likely Violation | Fix |
|-------|-----------------|-----|
| Class does many unrelated things | SRP | Split into focused classes |
| Adding feature requires modifying existing code | OCP | Extract strategy/plugin pattern |
| Subtype raises `NotImplementedError` | LSP | Redesign the hierarchy |
| Class implements methods it does not need | ISP | Split the Protocol |
| Constructor creates its own dependencies | DIP | Inject via constructor parameter |
| `if isinstance(...)` or type checking | OCP + LSP | Use polymorphism |
| God class with 10+ methods | SRP + ISP | Decompose by responsibility |
