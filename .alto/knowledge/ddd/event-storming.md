# Event Storming

Quick reference for AI agents facilitating domain discovery and generating domain models.

---

## What It Is

A workshop technique where domain experts and developers collaboratively discover business processes by focusing on domain events -- things that happen in the system.

---

## Key Elements

| Element | Color | Description | Example |
|---------|-------|-------------|---------|
| **Domain Event** | Orange | Something that happened (past tense) | `OrderPlaced`, `PaymentReceived` |
| **Command** | Blue | An intention to do something | `PlaceOrder`, `ProcessPayment` |
| **Aggregate** | Yellow | The entity that handles the command and emits the event | `Order`, `Payment` |
| **Read Model** | Green | Data needed to make a decision | `OrderSummaryView`, `InventoryLevel` |
| **Policy** | Lilac | Automated reaction: "When X happens, do Y" | "When OrderPlaced, reserve inventory" |
| **External System** | Pink | System outside our boundary | Payment Gateway, Email Service |
| **Actor** | Small sticky | Person or role that issues a command | Customer, Admin |
| **Hotspot** | Red/pink | Unresolved question, conflict, or confusion | "What happens if payment fails?" |

---

## Process

### Phase 1: Big Picture

Discover the full event flow with no constraints.

1. Everyone writes domain events on orange stickies (past tense, domain language)
2. Arrange events on a timeline left to right
3. Mark hotspots for conflicts, unknowns, or edge cases
4. Identify pivotal events that mark phase transitions

**Output:** Timeline of events, list of hotspots

### Phase 2: Process Modeling

Add structure around the events.

1. For each event, identify the command that triggered it (blue)
2. Identify who or what issues each command (actor or policy)
3. Group commands and events around aggregates (yellow)
4. Add read models needed for decisions (green)
5. Identify external systems (pink)

**Output:** Command-event flows grouped by aggregate

### Phase 3: Software Design

Map the model to code structure.

1. Aggregates become domain entities/aggregate roots
2. Commands become application layer command handlers
3. Policies become domain event handlers or sagas
4. Read models become query handlers
5. External systems become infrastructure adapters
6. Aggregate clusters suggest bounded context boundaries

**Output:** Domain model skeleton ready for implementation

---

## Mapping Event Storm to Code

| Event Storm Element | Code Artifact | Layer |
|---------------------|---------------|-------|
| Domain Event | `@dataclass(frozen=True)` event class | `domain/events/` |
| Command | Command dataclass + handler | `application/commands/` |
| Aggregate | Entity class with invariants | `domain/models/` |
| Read Model | Query handler | `application/queries/` |
| Policy | Event handler / subscriber | `application/commands/` or `domain/services/` |
| External System | Adapter behind a port (Protocol) | `infrastructure/external/` |

---

## How alto Uses Event Storming

1. **During `alto init`:** Guide the user through a simplified event storming flow
   - Ask for key domain events ("What are the important things that happen?")
   - Identify commands ("What triggers each event?")
   - Discover aggregates ("What entity is responsible?")

2. **Generating DDD artifacts:** Map event storm output to:
   - Bounded context definitions in `docs/DDD.md`
   - Domain model files in `src/domain/models/`
   - Event definitions in `src/domain/events/`

3. **Generating tickets:** Each aggregate cluster becomes a set of beads tickets:
   - One ticket per aggregate (entity + value objects + repository port)
   - One ticket per policy (event handler)
   - One ticket per external system integration (adapter)

---

## Tips for AI Agents

- Events are always **past tense**: `OrderPlaced`, not `PlaceOrder`
- Commands are always **imperative**: `PlaceOrder`, not `OrderPlaced`
- If an event has no clear aggregate, it may belong to a different bounded context
- Hotspots are valuable -- they become spike tickets for unknowns
- Policies that cross aggregate boundaries often signal context boundaries
- Keep aggregates small; if an aggregate handles too many commands, consider splitting
