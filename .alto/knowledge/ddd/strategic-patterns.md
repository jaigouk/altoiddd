# DDD Strategic Patterns

Quick reference for AI agents generating project structure and bounded context maps.

---

## Bounded Contexts

An explicit boundary within which a particular domain model applies. The same real-world concept may have different representations in different contexts.

**When to use:** Always. Every project has at least one bounded context. Separate contexts when the same term means different things to different parts of the system.

**Practical guidance:**
- Each bounded context gets its own directory under `src/`
- Models do NOT cross context boundaries -- use IDs or translation layers
- A `Customer` in Billing is not the same class as `Customer` in Shipping
- If two teams would argue about what a term means, you have two contexts

**Signals you need a new boundary:**
- Same word, different meaning (e.g., "Account" in Auth vs Finance)
- Different rates of change (e.g., Catalog changes weekly, Orders change per second)
- Different domain experts for different parts of the system

---

## Context Maps

Visual and documented relationships between bounded contexts.

### Relationship Patterns

| Pattern | Direction | Description | When to use |
|---------|-----------|-------------|-------------|
| **Shared Kernel** | Bidirectional | Shared code both contexts depend on | Small teams, tightly coupled concepts |
| **Customer/Supplier** | Upstream/Downstream | Upstream provides, downstream consumes | Clear producer-consumer relationship |
| **Conformist** | Downstream conforms | Downstream adopts upstream model as-is | No leverage to change upstream |
| **Anticorruption Layer** | Downstream protects | Translation layer between contexts | Upstream model is messy or foreign |
| **Open Host Service** | Upstream exposes | Published API/protocol for consumers | Multiple downstream consumers |
| **Separate Ways** | None | No integration, duplicate if needed | Integration cost exceeds value |

**Practical guidance for AI agents:**
- Default to Anticorruption Layer when integrating with external systems
- Use Shared Kernel sparingly -- it creates coupling
- Document the map in `docs/DDD.md` with a diagram or table
- Each integration point should name its pattern explicitly

---

## Ubiquitous Language

A shared vocabulary between domain experts and developers, used consistently in code, docs, and conversation.

**Rules:**
- Class names, method names, and variable names MUST use domain terms
- If the domain expert says "submit an order," the method is `order.submit()`, not `order.process()`
- The glossary lives in `docs/DDD.md` -- every bounded context has its own terms
- When terms conflict between contexts, that confirms a boundary

**Practical guidance:**
- Extract terms from domain expert interviews and event storming sessions
- Reject technical jargon in domain code (`handle`, `process`, `manage` are red flags)
- When renaming, rename everywhere -- code, tests, docs, tickets

**Red flags in code:**
- `Manager`, `Handler`, `Processor`, `Helper` classes (not domain language)
- Method names that describe HOW instead of WHAT (`updateDatabaseRecord` vs `placeOrder`)
- Comments explaining what a domain term means (the name should be self-evident)

---

## Subdomains

Classification of business capabilities by strategic importance.

### Core Domain

The competitive advantage. This is why the business exists.

- **Investment:** Highest effort, best developers, most testing
- **Code quality:** Rich domain model, full DDD tactical patterns
- **Example:** For alto, the project bootstrapping and DDD artifact generation

### Supporting Domain

Necessary for the business but not a differentiator.

- **Investment:** Moderate effort, good-enough solutions
- **Code quality:** Simpler patterns acceptable, still tested
- **Example:** For alto, template file management

### Generic Domain

Commodity functionality. Buy or use a library.

- **Investment:** Minimal custom code, prefer existing solutions
- **Code quality:** Thin wrappers around libraries
- **Example:** For alto, file I/O, TOML/YAML parsing, Git operations

**Practical guidance for AI agents:**
- Ask "would a competitor be embarrassed if they lacked this?" -- if yes, it is Core
- Core subdomain code gets the most tests and strictest architecture enforcement
- Generic subdomain code should wrap libraries, not reimplement them
- Classification drives the complexity budget: Core gets complex patterns, Generic stays simple

---

## Applying Strategic Patterns in alto

When generating project structure:

1. **Identify subdomains** from the PRD -- classify each capability as Core/Supporting/Generic
2. **Draw bounded context boundaries** -- one context may cover multiple subdomains
3. **Map relationships** between contexts using the patterns above
4. **Generate directory structure** -- one top-level package per bounded context
5. **Create glossary** -- extract ubiquitous language per context into `docs/DDD.md`
6. **Assign complexity budget** -- Core contexts get full DDD; Generic gets thin wrappers
