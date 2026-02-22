# Domain-Driven Design (DDD): 2026 Pragmatic Working Reference

This guide is a practical DDD operating model for this repository, calibrated to current guidance as of 2026-02-22. It is intentionally opinionated: use heavy DDD where complexity is real, and stay simple elsewhere.

## 1. Use DDD Where It Pays Off

Use full tactical DDD only when business rules are complex, high-change, and strategically important.

- Core domain: rich model, strict invariants, explicit aggregates.
- Supporting domain: lighter model, selective DDD patterns.
- Generic domain: buy/build simple CRUD and move on.

If a feature is mostly forms + persistence with low rule complexity, do not force aggregate-heavy design.

## 2. Strategic Design First

Before implementation, produce these artifacts:

1. Ubiquitous language glossary.
2. Subdomain map: core/supporting/generic.
3. Bounded context map with integration relationships.
4. Top domain stories (happy path + key edge scenarios).

### 2.1 Ubiquitous Language Rules

- Code names must mirror business terms exactly.
- One meaning per term per bounded context.
- If two teams use one term differently, split contexts instead of forcing a shared definition.

### 2.2 Bounded Context Rules

- Each bounded context owns its model and invariants.
- Cross-context communication happens through explicit contracts (events/APIs), not shared entities.
- Keep translation/mapping at context boundaries.

## 3. Tactical Modeling Rules (Inside a Bounded Context)

### 3.1 Value Objects

- Prefer value objects first.
- Make them immutable.
- Put validation and normalization in constructors/factories.
- Equality is structural (by value), not identity-based.

### 3.2 Entities

- Entity identity is stable over time.
- Entities enforce behavior; avoid setter-only anemic models.
- Expose intent methods (`approve()`, `schedule()`) rather than state mutation verbs.

### 3.3 Aggregates

Treat aggregate design as consistency-boundary design.

- Model true invariants inside one aggregate boundary.
- Keep aggregates small unless a real invariant requires growth.
- Reference other aggregates by identity only.
- Modify one aggregate per transaction in normal flow.
- Use eventual consistency for cross-aggregate or cross-context rules.

Heuristic: ask "whose job is it to make this consistent now?"  
If it must be immediate for this user action, keep it in one transactional boundary. If not, publish an event and complete asynchronously.

### 3.4 Domain Services

Use domain services only for domain operations that do not fit naturally on one entity/value object. Keep them domain-focused, not orchestration-heavy.

### 3.5 Repositories

- One repository per aggregate root.
- Return aggregate roots, not random internal objects.
- Keep query/read models separate when needed (CQRS style is acceptable where read complexity/perf demands it).

## 4. Integration and Context Mapping

Define the relationship between contexts explicitly:

- Partnership
- Shared Kernel (use sparingly)
- Upstream/Downstream
- Customer/Supplier
- Published Language / ACL (Anti-Corruption Layer)

Default stance: prefer explicit contracts plus ACL over deep shared models.

## 5. 2026 Integration Non-Negotiables

For cross-aggregate and cross-context workflows:

- Publish integration events from a transactional outbox (avoid dual-write bugs).
- Treat consumers as idempotent by default (duplicate delivery is expected).
- Use compensating transaction workflows for multi-step distributed business processes.
- Version event contracts explicitly and evolve with backward compatibility windows.
- Prefer CloudEvents envelope fields for interoperable event metadata where practical.

## 6. Domain Storytelling Workflow (Recommended Discovery Path)

Use `docs/templates/DDD_STORY_TEMPLATE.md` to run workshops.

1. Capture as-is stories with domain experts.
2. Visualize actor -> activity -> work object sequences.
3. Extract terms for ubiquitous language.
4. Identify handoffs/conflicts to propose bounded contexts.
5. Convert stories to candidate aggregates, commands, events, and invariants.

Avoid premature abstraction. Model concrete scenarios first; generalize later.

## 7. Quality Gates for DDD Changes

For each non-trivial domain change, require:

- Tests for aggregate invariants.
- Tests for domain events emitted by key behaviors.
- Tests for cross-context contract compatibility (schema/contract tests where relevant).
- Architecture check: domain layer has no framework/ORM dependencies.
- Eventing checks: outbox publishing path tested and consumers proven idempotent.
- Operational checks: key domain flows emit traceable telemetry (event/correlation IDs).

## 8. Common Failure Modes

- "Large graph aggregate" driven by object navigation instead of invariants.
- Shared entity model across bounded contexts.
- Anemic domain model with all rules in services/controllers.
- Transaction spanning multiple aggregates by default.
- Letting DB schema names define domain language.
- Fire-and-forget events without delivery guarantees or replay strategy.
- Breaking event contracts without versioning/migration windows.

## 9. Decision Cheat Sheet

- Complex rule with strict consistency: aggregate behavior.
- Workflow spanning contexts: domain events + eventual consistency.
- Simple admin CRUD: keep it simple, do not force tactical DDD.
- Terminology conflict between teams: separate bounded contexts.
- Cross-service state change with reliability requirements: outbox + idempotent consumer.

## References

- Microsoft Learn: Using tactical DDD to design microservices (2025-06-17)  
  https://learn.microsoft.com/en-us/azure/architecture/microservices/model/tactical-ddd
- Microsoft Learn: Transactional Outbox pattern (cloud design pattern)  
  https://learn.microsoft.com/en-us/azure/architecture/patterns/transactional-outbox
- Microsoft Learn: Compensating Transaction pattern  
  https://learn.microsoft.com/en-us/azure/architecture/patterns/compensating-transaction
- AWS Prescriptive Guidance: Domain-driven design in software architecture  
  https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/domain-driven-design.html
- AWS Prescriptive Guidance: Event storming  
  https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/event-storming.html
- AWS Prescriptive Guidance: Saga orchestration pattern  
  https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/saga-orchestration.html
- AWS Prescriptive Guidance: Retry with backoff pattern  
  https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/retry-backoff.html
- Domain Storytelling (official)  
  https://domainstorytelling.org/
- Domain Storytelling with DDD  
  https://domainstorytelling.org/domain-driven-design
- CNCF CloudEvents specification (latest release line)  
  https://github.com/cloudevents/spec
