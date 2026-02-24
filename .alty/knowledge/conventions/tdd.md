# Test-Driven Development (TDD)

Quick reference for AI agents writing tests and production code.

---

## The Cycle

### RED: Write a Failing Test

Write a test that describes the desired behavior. Run it. It must fail.

```python
def test_order_submit_changes_status_to_submitted():
    order = Order(id=OrderId("ord-1"), status=OrderStatus.DRAFT)
    order.submit()
    assert order.status == OrderStatus.SUBMITTED
```

**Rules:**
- The test defines WHAT, not HOW
- Test must fail for the right reason (not a syntax error or import error)
- Test name describes the behavior: `test_<unit>_<scenario>_<expected_outcome>`

### GREEN: Write Minimum Code to Pass

Write the simplest implementation that makes the test pass. Nothing more.

```python
@dataclass
class Order:
    id: OrderId
    status: OrderStatus = OrderStatus.DRAFT

    def submit(self) -> None:
        self.status = OrderStatus.SUBMITTED
```

**Rules:**
- Do not add code that no test requires
- Do not handle edge cases until a test demands it
- "Minimum" means minimum -- even hardcoded returns are valid if they pass

### REFACTOR: Clean Up

Improve the code while keeping all tests green.

```python
def submit(self) -> None:
    if self.status != OrderStatus.DRAFT:
        raise InvalidStateError(f"Cannot submit order in {self.status} state")
    self.status = OrderStatus.SUBMITTED
```

Wait -- that guard clause is new behavior. It needs its own RED phase first:

```python
def test_order_submit_raises_when_not_draft():
    order = Order(id=OrderId("ord-1"), status=OrderStatus.SUBMITTED)
    with pytest.raises(InvalidStateError):
        order.submit()
```

**Rules:** All tests still pass; refactor production AND test code; extract duplication.

---

## Test Structure (Arrange-Act-Assert)

```python
def test_money_add_same_currency():
    # Arrange
    a = Money(amount=Decimal("10.00"), currency="USD")
    b = Money(amount=Decimal("5.00"), currency="USD")

    # Act
    result = a.add(b)

    # Assert
    assert result == Money(amount=Decimal("15.00"), currency="USD")
```

---

## Rules

1. **Never write production code without a failing test**
2. **One concept per test** -- one behavior, one assertion (related asserts on same object OK)
3. **Test behavior, not implementation** -- test what it does, not how
4. **Tests are documentation** -- readable by a new developer
5. **Fast tests** -- no filesystem, network, or database in domain tests
6. **Isolated tests** -- no test depends on another test's state

---

## Pytest Patterns

### Fixtures for shared setup

```python
@pytest.fixture
def draft_order() -> Order:
    return Order(id=OrderId("ord-1"), status=OrderStatus.DRAFT)

def test_submit_draft_order(draft_order: Order):
    draft_order.submit()
    assert draft_order.status == OrderStatus.SUBMITTED
```

### Parametrize for multiple inputs

```python
@pytest.mark.parametrize("quantity", [0, -1])
def test_order_add_item_rejects_invalid_quantity(quantity: int):
    order = Order(id=OrderId("ord-1"))
    with pytest.raises(ValueError, match="Quantity must be positive"):
        order.add_item(ProductId("p-1"), quantity, Money(Decimal("10"), "USD"))
```

### Testing exceptions

```python
def test_money_add_rejects_currency_mismatch():
    usd, eur = Money(Decimal("10"), "USD"), Money(Decimal("5"), "EUR")
    with pytest.raises(CurrencyMismatchError):
        usd.add(eur)
```

---

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Testing implementation | Breaks on refactor | Test observable behavior only |
| Skipping RED | No proof test catches failures | Always see it fail first |
| Multiple behaviors per test | Hard to diagnose failures | One test, one concept |
| Test depends on other tests | Order-dependent failures | Each test sets up its own state |
| Testing private methods | Coupling to internals | Test through the public API |
| No assertion | Test always passes | Every test must assert something |
| Mocking the thing under test | Tests nothing real | Only mock dependencies, not the subject |
| Writing tests after code | Tests conform to implementation | Tests should drive design |

---

## Test Organization

- `tests/domain/` -- Pure unit tests, no mocks, no I/O
- `tests/application/` -- Unit tests with mocked ports (repositories, external services)
- `tests/infrastructure/` -- Integration tests with real adapters
