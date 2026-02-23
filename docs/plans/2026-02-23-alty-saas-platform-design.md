# alty SaaS Platform Design

**Date:** 2026-02-23
**Status:** Approved
**Author:** kusanagi + Claude

---

## 1. Product Overview

**Product:** alty
**Pitch:** "From idea to production-ready project, then keeps it healthy."

alty evolves from a CLI bootstrapper into a full platform. The CLI scaffolds projects with DDD + TDD + SOLID. The web dashboard monitors, visualizes, and keeps projects healthy. The VSCode extension keeps it all accessible from the IDE.

### Three Surfaces, One Product

| Surface | Role | License | Revenue |
|---------|------|---------|---------|
| **alty CLI** | Bootstraps projects (DDD, templates, tickets) | Open-source (MIT) | Free -- drives adoption |
| **alty Web** | Dashboard: DDD canvas, health scores, domain stories, team views, decay heatmaps, ticket grooming | Proprietary SaaS | $9-15/mo per seat |
| **alty for VSCode** | IDE companion: Beads explorer, health badges, canvas preview, quick actions | Free extension | Free -- drives web signups |

### Competitive Landscape

| Competitor | What it does | Gap |
|-----------|-------------|-----|
| [jdillon/vscode-beads](https://github.com/jdillon/vscode-beads) (34 stars) | Kanban board, drag-and-drop, filters | Zero DDD awareness, no health scoring |
| [kris-hansen/beads-vscode](https://github.com/kris-hansen/beads-vscode) (3 stars) | Tree view, dependency graph | No DDD features, no releases |
| [Egon.io](https://egon.io/) | Domain Storytelling canvas (browser) | Manual-only, no IDE integration, no ticket connection |
| [Pencil.dev](https://www.pencil.dev/) | Design canvas in IDE (a16z-backed) | Design-focused, not DDD, VC-funded free model |

**Nobody combines Beads + DDD visualization + domain stories + health scoring.** That is the gap.

---

## 2. Architecture

### System Diagram

```
+-----------------------------------------------------+
|                    Clients                            |
|                                                      |
|  +-----------+  +-----------+  +------------------+  |
|  | CLI (oss) |  | VSCode    |  | Web Dashboard    |  |
|  | Python    |  | TypeScript|  | HTMX + Jinja     |  |
|  | Typer     |  | TreeView  |  | + Three.js 3D    |  |
|  +-----+-----+  +-----+-----+  | + SVG 2D Stories |  |
|        |              |        +--------+---------+  |
+--------+--------------+-----------------+------------+
         |              |                 |
         v              v                 v
+-----------------------------------------------------+
|          alty API (FastAPI, Python)              |
|      HTMX partials + JSON API from same app          |
|                                                      |
|  Auth (Hanko) | Projects | LLM | Health | DDD Graph  |
+-------------------------+----------------------------+
                          |
                +---------+---------+
                v         v         v
          +----------+ +------+ +----------+
          | SQLite   | |Redis | |Hetzner S3|
          | (data)   | |(opt) | | (files)  |
          +----------+ +------+ +----------+
```

### Tech Stack

| Component | Tech | Why |
|-----------|------|-----|
| API + Web | **FastAPI + HTMX + Jinja2** | One codebase, one deploy, Python-native |
| 3D Canvas | **Three.js** (existing Tachikoma navigator engine) | Already built, proven, GitS aesthetic, hyperbolic layout |
| 2D Story Canvas | **SVG renderer** (custom, lightweight) | Domain Storytelling notation: actors, work objects, arrows, sequence numbers |
| Extension | **TypeScript** (TreeView + Webview for canvas preview) | VSCode requires TS |
| Auth | **Hanko** (Germany, open-source) | GDPR-native, passkeys |
| DB | **SQLite** (WAL mode) | Zero ops, file backup, DHH-approved for single-server SaaS |
| Cache | **Redis** (optional, defer until needed) | SQLite WAL mode is fast enough to start |
| Storage | **Hetzner S3** | DDD artifacts, backups, SQLite snapshots |
| CDN | **Bunny.net** (Slovenia) | Static assets, Three.js bundles |
| Email | **Scaleway** (France) | Transactional email |
| LLM | **OpenRouter** (BYOK) | User provides key, cheapest models (DeepSeek V3.2, Qwen, Gemini Flash) |
| Hosting | **Hetzner** VM CX22 (~$6/mo) | Germany, GDPR-compliant |

### Why SQLite

- Single server, single process (FastAPI)
- WAL mode handles concurrent reads
- Backup = copy the file to Hetzner S3
- No Postgres ops, no connection pooling, no Docker
- Scales to millions of requests/day on a single machine
- When outgrown (good problem), migrate to Postgres

Reference: [DHH on SQLite](https://highperformancesqlite.com/watch/dhh-on-sqlite), [Rails 8 all-in on SQLite](https://bhavyansh001.medium.com/sqlite-in-production-rails-8-made-me-a-believer-28a621db864a)

### GDPR Compliance

| Need | Provider | Location |
|------|----------|----------|
| Compute | Hetzner | Germany |
| Auth | Hanko | Germany |
| CDN/DNS | Bunny.net | Slovenia |
| Email | Scaleway | France |
| Storage | Hetzner S3 | Germany |

Only unavoidable US dependency: LLM API calls via OpenRouter (mitigable by self-hosting on Nebius EU).

Reference: [Made in EU -- it was harder than I thought](https://www.coinerella.com/made-in-eu-it-was-harder-than-i-thought/)

### Shared Domain (CLI <-> API)

The CLI and API share the same Python domain models (`src/domain/`):

```
alty/
+-- src/domain/          <-- shared between CLI and API
+-- src/application/     <-- shared ports, CLI + API implement adapters
+-- cli/                 <-- CLI-specific (Typer)
+-- api/                 <-- API-specific (FastAPI)
+-- web/                 <-- HTMX templates + static assets (Three.js, SVG)
+-- extension/           <-- VSCode TypeScript extension
```

---

## 3. Two Canvases

### 3.1 3D Architecture Navigator (Three.js)

Ported from the existing Tachikoma Knowledge Navigator prototype. Hyperbolic-inspired layout, Ghost in the Shell aesthetic, aerospace HUD heuristics.

**Source:** `~/AI/Tachikoma/docs/research/20260222_3D_Knowledge_Visualization.html`
**Research:** `~/AI/Tachikoma/docs/research/20260222_3D_Knowledge_Visualization_Research.md` (67 citations)

#### Data Sources

| Source | Parser | Canvas Nodes |
|--------|--------|-------------|
| `DDD.md` | Headings + YAML | Bounded Contexts (L1), Aggregates (L2), Entities/VOs (L3) |
| `.beads/issues.jsonl` | JSON | Ticket nodes attached to their BC/aggregate |
| `docs/ARCHITECTURE.md` | Heading parse | Layer nodes (domain/application/infra) |
| `.alty/knowledge/` | TOML | Knowledge base nodes (future) |

#### Node Visual Encoding

| Node Type | Color | Shape |
|-----------|-------|-------|
| Core BC | Cyan `#00d4ff` | Large wireframe icosahedron + rings |
| Supporting BC | Green `#00ff88` | Medium icosahedron |
| Generic BC | Orange `#ffaa00` | Medium icosahedron |
| Aggregate | Purple `#8855ff` | Small icosahedron + inner glow |
| Entity/VO | Dim cyan | Tiny sphere (grandchild LOD) |
| Ticket (healthy) | White `#ffffff` | Flat sprite badge |
| Ticket (stale) | Red pulse `#ff3344` | Pulsing sprite + glow |
| Ticket (blocked) | Orange flash `#ffaa00` | Flashing sprite |

#### Navigation Modes

| Mode | Trigger | View |
|------|---------|------|
| **Strategic** (MACRO) | Default | All bounded contexts, context map edges, health heatmap |
| **Tactical** (MESO) | Double-click BC | Aggregates within BC, attached tickets, dependency edges |
| **Implementation** (MICRO) | Double-click aggregate | Entities, VOs, individual tickets with detail panels |
| **Health** (overlay) | `H` key | All nodes colored by health score (green/yellow/red) |

#### HUD Elements (from Tachikoma prototype)

| Element | DDD Application |
|---------|----------------|
| Top-left (Focus Node) | Current BC name + health score (0-100) |
| Top-right (Depth Level) | Strategic / Tactical / Implementation |
| Bottom-left (Log Panel) | Recent decay events, ripple review triggers |
| Bottom-right (Counters) | Total tickets, blocked count, stale count |
| Zoom indicator | Depth into DDD hierarchy |

### 3.2 2D Domain Story Canvas (SVG)

Implements the [Domain Storytelling](https://domainstorytelling.org/) visual notation. Auto-generated from DDD artifacts, then editable by the user.

#### Notation Elements

| Element | Visual | Role |
|---------|--------|------|
| **Actor** | Person/group/system icon (large) | Who performs activities |
| **Work Object** | Document/physical/digital icon (smaller) | What actors create, exchange, manipulate |
| **Activity** | Labeled arrow with verb | What the actor does |
| **Sequence Number** | Number at arrow origin | Order of events in the story |
| **Annotation** | Text note | Edge cases, domain terms, assumptions |
| **Group** | Rectangle/circle outline | Clusters related steps |

#### Reading a Domain Story

```
  (1)                   (2)                       (3)
Customer --creates--> [Order] --submits--> OrderService --validates--> [Payment]
(actor)    (activity)  (work obj) (activity)  (system)     (activity)   (work obj)
```

Subject -> verb -> object, with sequence numbers establishing narrative flow.

#### What Makes This Better Than Egon.io

| Egon.io (manual) | alty (auto + edit) |
|------------------|------------------------|
| Start from blank canvas | Parse DDD.md domain stories -> generate initial layout |
| Manually place actors | Actors extracted from bounded context definitions |
| Manually draw arrows | Activities extracted from use cases / tickets |
| No connection to code | Sequence numbers link to beads tickets |
| Export as SVG/PNG only | Stories versioned in git as YAML |
| No health indicators | Stale steps highlighted when linked ticket is outdated |
| No architecture context | Toggle to 3D navigator to see where the story fits |

#### Story Data Format

```yaml
# .alty/stories/order-fulfillment.story.yaml
name: "Customer Places Order"
bounded_context: OrderFulfillment
actors:
  - id: customer
    type: person
    label: Customer
  - id: order_service
    type: system
    label: OrderService
work_objects:
  - id: order
    type: document
    label: Order
  - id: payment
    type: digital
    label: Payment
activities:
  - seq: 1
    from: customer
    to: order
    verb: "creates"
  - seq: 2
    from: customer
    to: order_service
    verb: "submits"
    via: order
  - seq: 3
    from: order_service
    to: payment
    verb: "validates"
annotations:
  - text: "Order must contain at least one line item"
    attached_to: order
links:
  - ticket: k7m.15
    activity: 2
```

#### Auto-Generation Flow

```
DDD.md (bounded contexts, aggregates, use cases)
    | LLM parses (~$0.01 per story)
    v
.alty/stories/*.story.yaml
    | SVG renderer
    v
2D Story Canvas (interactive, editable)
    | User edits
    v
Story files updated -> git versioned
```

### 3.3 Canvas Integration

```
+--------------------------------------------------+
|  Web Dashboard                                    |
|                                                   |
|  +-------------------+  +---------------------+  |
|  |  3D Navigator     |  |  2D Story Canvas   |  |
|  |  (architecture)   |<>|  (domain stories)  |  |
|  |                   |  |                     |  |
|  |  Click a BC ------+->|  Shows stories for  |  |
|  |                   |  |  that context       |  |
|  |                   |  |                     |  |
|  |  <----------------+--|  Click actor goes   |  |
|  |  to BC in 3D      |  |  back to 3D view   |  |
|  +-------------------+  +---------------------+  |
|                                                   |
|  Toggle: [3D Architecture] [2D Stories] [Health]  |
+--------------------------------------------------+
```

---

## 4. Health Scoring

### Score Calculation

```python
for ticket in project.tickets:
    score = 100
    score -= age_penalty(ticket.updated_at)          # -1 per day stale
    score -= blocked_penalty(ticket.dependencies)     # -10 if blocked
    score -= label_penalty(ticket.labels)             # -20 if review_needed
    score -= llm_semantic_check(ticket, ddd_model)    # -5 to -30 for drift
    ticket.health_score = max(0, score)
```

### LLM Semantic Check (BYOK, ~$0.001/call)

Detects:
- Ticket references a renamed/deleted bounded context
- AC contradicts current DDD model
- Ticket spans multiple BCs (boundary violation)
- Ubiquitous language mismatch

### Visual Encoding

- **Green (80-100):** Healthy, up-to-date
- **Yellow (50-79):** Aging, may need review
- **Red (0-49):** Stale, contradictions detected, needs immediate attention

---

## 5. VSCode Extension

The extension does NOT replicate the full 3D/2D canvases. It provides:

1. **TreeView** -- Rich Beads explorer (labels as colored badges, full description panel, dependency graph). Works offline, no backend needed.
2. **Webview panel** -- Lightweight read-only embed of the canvas (syncs from API when connected).
3. **Health badges** -- File decorations showing health scores on `.beads/` files.
4. **CodeLens** -- Above DDD model classes, shows which BC they belong to and health status.
5. **Commands** -- "Open in alty dashboard" opens the full web canvas for that BC.

---

## 6. Pricing

### Tiers

| Tier | Price | Target | Includes |
|------|-------|--------|----------|
| **Free** | $0 | Solo dev trying it out | CLI (OSS), 1 project, manual 2D story canvas, no health scoring, no LLM |
| **Solo** | $9/mo or $79/yr | Solo dev, serious use | Unlimited projects, 3D navigator, auto-generated stories, health scoring, LLM (BYOK), VSCode sync |
| **Team** | $15/seat/mo | Small teams (2-10) | Everything in Solo + shared dashboards, cross-repo DDD maps, alerts, story collaboration |

### Unit Economics

| Item | Monthly Cost |
|------|-------------|
| Hetzner VM (CX22) | ~$6 |
| Hetzner S3 | ~$1 |
| Bunny.net CDN | ~$1 |
| Hanko auth | Free (up to 10k MAU) |
| Scaleway email | ~$1 |
| LLM | $0 (BYOK) |
| **Total infra** | **~$9/mo** |

Break-even: 1 paying Solo user. Target $2-5k/mo at 150-300 users.

---

## 7. Launch Phases

### Phase 1: CLI + Web MVP (months 1-3)

- alty CLI already exists (open-source, working)
- FastAPI backend with SQLite
- HTMX dashboard: project list, ticket view, basic health badges
- 2D Story Canvas: manual authoring (SVG renderer, drag/drop)
- Auth via Hanko, deploy to Hetzner
- **Goal:** Usable product, beta users from CLI community

### Phase 2: Intelligence Layer (months 3-5)

- LLM health scoring (BYOK OpenRouter)
- Auto-generation of domain stories from DDD.md
- Staleness alerts (webhook + email)
- 3D Navigator (port Tachikoma prototype to DDD data)
- **Goal:** Differentiated product, launch Solo tier

### Phase 3: Team + VSCode (months 5-8)

- Team features (shared dashboards, multi-seat)
- VSCode extension (TreeView + Webview canvas preview)
- Cross-repo DDD maps
- Story collaboration (multiple editors)
- **Goal:** Launch Team tier

### Phase 4: Growth (months 8+)

- Godot-based desktop client (optional, premium feel)
- CI/CD integration (health gates in pipeline)
- Export to Miro/Confluence
- Public API for integrations
- **Goal:** Expand beyond alty users to general DDD market

---

## 8. Marketing

| Channel | Approach | Cost |
|---------|----------|------|
| VS Marketplace | Free extension drives discovery | $0 |
| CLI upsell | Every `alty init` user sees "connect to dashboard" | $0 |
| Domain Storytelling community | Auto-generated stories vs manual Egon.io | $0 |
| DDD community (Slack, Discord, conferences) | Niche, high-intent | $0 |
| Twitter/X | 3D navigator demos are inherently viral (GitS aesthetic) | $0 |
| Hacker News | "Show HN: 3D DDD knowledge navigator with auto-generated domain stories" | $0 |

The 3D navigator is the marketing weapon. A 30-second screen recording of flying through bounded contexts with the GitS aesthetic will get attention that no flat dashboard screenshot ever could.

---

## 9. Key Differentiators

1. **Auto-generated domain stories** -- LLM parses DDD artifacts into visual stories (nobody else does this)
2. **3D DDD architecture navigator** -- Hyperbolic layout, GitS aesthetic, fly through bounded contexts (nobody else has this)
3. **LLM-powered health scoring** -- Semantic decay detection, boundary violation alerts, ubiquitous language checks
4. **Beads-native integration** -- Only DDD-aware issue tracker extension
5. **EU-first GDPR stack** -- All infrastructure in EU, no US data residency
6. **SQLite simplicity** -- Zero-ops database, file-based backups, DHH-approved

---

## 10. Risks

| Risk | Mitigation |
|------|-----------|
| Small market (DDD practitioners) | Free CLI builds broad base; DDD is growing with AI-assisted dev |
| Three.js complexity for web dashboard | Prototype already exists and works (Tachikoma navigator) |
| BYOK friction (users must get API key) | Clear onboarding guide; free tier works without LLM |
| SQLite concurrency limits | WAL mode handles typical SaaS load; migrate to Postgres if needed |
| Competing with free Egon.io | Auto-generation + ticket integration + health scoring = different product |
| Solo founder bandwidth | Phase 1 is minimal (HTMX + SQLite + existing CLI); complexity grows incrementally |

---

## References

- [Domain Storytelling](https://domainstorytelling.org/) -- notation, quick-start guide
- [Egon.io](https://egon.io/) -- existing open-source DST tool (GPLv3)
- [Pencil.dev](https://www.pencil.dev/) -- design canvas in IDE (a16z Speedrun)
- [DHH on SQLite](https://highperformancesqlite.com/watch/dhh-on-sqlite) -- production SQLite advocacy
- [Rails 8 SQLite](https://bhavyansh001.medium.com/sqlite-in-production-rails-8-made-me-a-believer-28a621db864a) -- real-world results
- [Made in EU](https://www.coinerella.com/made-in-eu-it-was-harder-than-i-thought/) -- EU GDPR infrastructure choices
- [OpenRouter models](https://www.teamday.ai/blog/top-ai-models-openrouter-2026) -- cheapest LLMs for 2026
- [VSCode extension monetization](https://markaicode.com/sell-vs-code-extensions-2025/) -- pricing strategies
- Tachikoma Knowledge Navigator -- `~/AI/Tachikoma/docs/research/20260222_3D_Knowledge_Visualization_Research.md`
