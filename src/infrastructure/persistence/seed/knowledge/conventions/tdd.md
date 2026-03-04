---
last_verified: "2026-03-04"
confidence: high
next_review_date: "2026-06-04"
---

# Test-Driven Development (TDD)

Quick reference for AI agents writing tests and production code.

## The Cycle

### RED: Write a Failing Test

Write a test that describes the desired behavior. Run it. It must fail.

```python
def test_order_submit_changes_status():
    order = Order(id=OrderId("ord-1"), status=OrderStatus.DRAFT)
    order.submit()
    assert order.status == OrderStatus.SUBMITTED
```

**Rules:** Test defines WHAT, not HOW. Must fail for the right reason (not syntax error).

### GREEN: Write Minimum Code to Pass

Simplest implementation that makes the test pass. Nothing more.

**Rules:** No code without a failing test. No edge case handling until a test demands it.

### REFACTOR: Clean Up

Improve code while keeping all tests green. New behavior needs its own RED phase first.

## Test Structure (Arrange-Act-Assert)

```python
def test_money_add_same_currency():
    a = Money(amount=Decimal("10.00"), currency="USD")  # Arrange
    result = a.add(Money(amount=Decimal("5.00"), currency="USD"))  # Act
    assert result == Money(amount=Decimal("15.00"), currency="USD")  # Assert
```

## Rules

1. Never write production code without a failing test
2. One concept per test
3. Test behavior, not implementation
4. Tests are documentation
5. Fast tests -- no I/O in domain tests
6. Isolated tests -- no shared state between tests

## Anti-Patterns

| Anti-Pattern | Fix |
|-------------|-----|
| Testing implementation | Test observable behavior only |
| Skipping RED | Always see it fail first |
| Multiple behaviors per test | One test, one concept |
| Testing private methods | Test through the public API |
| Writing tests after code | Tests should drive design |

## Test Organization

- `tests/domain/` -- Pure unit tests, no mocks, no I/O
- `tests/application/` -- Unit tests with mocked ports
- `tests/infrastructure/` -- Integration tests with real adapters
