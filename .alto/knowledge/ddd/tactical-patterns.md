# DDD Tactical Patterns

Quick reference for AI agents generating domain model code.

---

## Entities

Objects defined by a unique identity that persists through state changes.

**When to use:** The object has a lifecycle and must be tracked over time (e.g., Order, User, Project).

```python
@dataclass
class Order:
    id: OrderId  # Identity field
    status: OrderStatus = OrderStatus.DRAFT

    def submit(self) -> None:
        if self.status != OrderStatus.DRAFT:
            raise InvalidStateError("Only draft orders can be submitted")
        self.status = OrderStatus.SUBMITTED
```

**Mistakes:** Using DB IDs as domain identity; anemic getters/setters; exposing mutable internals.

---

## Value Objects

Immutable objects defined entirely by their attributes. No identity.

**When to use:** The concept is defined by what it IS, not which one it is (e.g., Money, Email, DateRange).

```python
@dataclass(frozen=True)
class Money:
    amount: Decimal
    currency: str

    def add(self, other: Money) -> Money:
        if self.currency != other.currency:
            raise CurrencyMismatchError(self.currency, other.currency)
        return Money(amount=self.amount + other.amount, currency=self.currency)
```

**Mistakes:** Mutable value objects (always `frozen=True`); adding an `id` field; putting logic in services instead of the object.

---

## Aggregates

Cluster of entities and value objects with a single root entity that enforces consistency.

**When to use:** A group of objects that must change together to maintain invariants.

```python
@dataclass
class Order:  # Aggregate root
    id: OrderId
    items: list[OrderItem] = field(default_factory=list)

    def add_item(self, product_id: ProductId, quantity: int, price: Money) -> None:
        if quantity <= 0:
            raise ValueError("Quantity must be positive")
        self.items.append(OrderItem(product_id=product_id, quantity=quantity, price=price))
```

**Rules:**
- One root entity per aggregate; external references use the root's ID only
- Only the root is accessible from outside the aggregate
- One aggregate per transaction -- do not modify multiple aggregates in one operation

**Mistakes:** Too-large aggregates; referencing internals from outside; loading related aggregates instead of by ID.

---

## Domain Services

Stateless operations that involve multiple entities/aggregates or don't naturally belong to one.

**When to use:** The operation spans aggregates, or forcing it onto an entity feels unnatural.

```python
class PricingService:
    def calculate_discount(self, order: Order, customer_tier: CustomerTier) -> Money:
        base = order.total()
        rate = self._discount_rate(customer_tier, len(order.items))
        return Money(amount=base.amount * rate, currency=base.currency)
```

**Mistakes:** Logic that belongs on an entity (anemic model); stateful services; depending on infrastructure.

---

## Domain Events

Immutable record of something meaningful that happened in the domain.

**When to use:** Other parts of the system need to react to a state change. Named in past tense.

```python
@dataclass(frozen=True)
class OrderSubmitted:
    order_id: OrderId
    submitted_at: datetime
    total: Money
```

**Mistakes:** Events describing intent (use Commands for that); mutable events (always frozen); containing full entity references instead of IDs.

---

## Repositories

Collection-like interface for persisting and retrieving aggregate roots.

**When to use:** Every aggregate root gets one repository. Nothing else does.

```python
class OrderRepository(Protocol):
    def find_by_id(self, order_id: OrderId) -> Order | None: ...
    def save(self, order: Order) -> None: ...
    def next_id(self) -> OrderId: ...
```

**Mistakes:** Repos for non-root entities; leaking persistence details (no SQL/file paths); domain layer depending on infrastructure (define as Protocol).

---

## Factories

Encapsulate complex object creation logic.

**When to use:** Creating an aggregate requires validation, defaults, or coordination beyond a simple constructor.

```python
class OrderFactory:
    def create_from_cart(self, cart: Cart, customer_id: CustomerId) -> Order:
        if not cart.items:
            raise EmptyCartError()
        order = Order(id=OrderId.generate(), customer_id=customer_id)
        for item in cart.items:
            order.add_item(item.product_id, item.quantity, item.price)
        return order
```

**Mistakes:** Using for simple construction; calling repos/external services; returning invalid objects.
