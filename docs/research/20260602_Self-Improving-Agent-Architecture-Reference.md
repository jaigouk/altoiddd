# A Reference Architecture for Self-Improving AI Agents

*A framework-agnostic, vendor-neutral design reference for systems where an AI agent gets measurably better at its job over time — without retraining the underlying model.*

---

## How to use this document

This is a design reference, not a tutorial for any single product. It distills the architectural patterns that recur across current self-improving agent systems and states them generically so you can map them onto your own design.

Four reading conventions:

- **Patterns are vendor-neutral.** Each subsystem is described as a *pattern* first. Concrete tools appear only as dated, non-exhaustive examples — never as requirements.
- **Depend on interfaces, not vendors.** The agent tooling market churns fast; any specific tool named here may be eclipsed within months. The durable decision is which *open interface* a component speaks, so you can swap the component without rewriting your system. See §3, and the landscape table in §12.
- **Numbers are calibration points, not constants.** Specific figures (token budgets, character caps, cost-per-run) come from particular implementations and are flagged as such. Treat them as starting estimates to tune.
- **Source confidence is marked.** §13 separates what is peer-reviewed or standardized from what is vendor-reported or anecdotal.

A practical adoption checklist is in §11.

---

## 1. The core idea: learning is a property of the system, not only the weights

A common assumption is that an agent only improves when its model is retrained or fine-tuned. In practice, most of the reliability gap between "a model that can do the task" and "a system that does the task dependably" is closed *outside* the weights — in the orchestration code, the prompts, the tool definitions, and the memory the system carries between sessions.

This matters because each of those surfaces can be changed cheaply and reversibly. Rewriting an instruction or tightening a tool description is a text edit you can test, ship, and roll back in minutes. Adjusting billions of parameters to fix one deterministic failure is expensive, slow, and risks **catastrophic forgetting**, where new training overwrites previously reliable behavior.

The design consequence: **treat model retraining as the last lever, not the first.** Build the capacity for continuous improvement into the surfaces you can edit and verify on every run. A modest open-weights model wrapped in a well-engineered, self-correcting harness will often outperform a larger raw model on a narrow, well-specified workflow — not because it reasons better, but because its failure modes have been systematically engineered out.

This insight is reflected directly in the GEPA result (Agrawal et al., 2025): reflecting on execution traces in natural language to rewrite prompts beat reinforcement learning (GRPO) by roughly 10% while using up to ~35x fewer rollouts. Improvement came from editing text, not from gradient updates.

---

## 2. The three learning surfaces: Model, Harness, Context

A useful lens for any agent system is to separate where adaptation can happen into three layers. Each has a different cost, speed, and risk profile.

| Surface | What it is | How you change it | Cost & risk |
| :--- | :--- | :--- | :--- |
| **Model** | The base reasoning engine (the LLM and its weights). | Fine-tuning, RL, LoRA adapters. | Expensive, slow, coarse-grained; risks catastrophic forgetting. Last resort. |
| **Harness** | The execution code around the model: planning loops, tool orchestration, retries, error handling, state. | Editing logic and adding structural guardrails informed by failure traces. | Cheap, fast, deterministic. Where most reliability is actually engineered. |
| **Context** | Everything assembled into the prompt at runtime: system prompt, loaded skills, retrieved memory, user model. | Rewriting instructions, swapping which skills load, updating memory. | Cheapest and fastest. Changes behavior on the very next inference, at zero training cost. |

The practical takeaway is an ordering. When the agent fails, ask in sequence: *Can I fix this in the context (an instruction, a description)? If not, in the harness (a guardrail, a retry rule)? Only if neither suffices, is this a model-level limitation?* Most failures are resolved in the top two rows.

> The Model/Harness/Context decomposition is a synthesis framing from practitioner writing on context engineering, not a formal standard. It is useful precisely because it maps cleanly onto cost and reversibility — use it as a triage tool.

---

## 3. Design for substitution: depend on interfaces, not vendors

This is the most important principle in the document, and the one that keeps the rest of it from going stale. The agent-tooling ecosystem changes on a scale of weeks: memory libraries, observability backends, orchestration frameworks, and sandbox providers appear, merge, get acquired, and change their licenses constantly. (For instance, between late 2025 and mid 2026 the leading open-source tracing tool changed corporate ownership, and several "open-source" memory libraries moved their most useful features behind paid tiers.) If your architecture hard-codes a specific product, you inherit its roadmap, its pricing, and its lifespan.

The defense is to treat every component as **replaceable behind a thin internal interface** (the ports-and-adapters / hexagonal pattern). Your core logic should depend on a capability ("store an episodic memory," "trace this call," "run this code safely"), not on a brand. Concretely, standardize on the open interface for each layer and let the implementation behind it be swappable:

| Layer | Standardize on this interface | So you can swap the implementation freely |
| :--- | :--- | :--- |
| Model access | OpenAI-compatible chat/completions API; a model router (e.g. LiteLLM, OpenRouter) | Any frontier or open-weights model, hosted or local. |
| Tool / connector access | **MCP** (Model Context Protocol) | Any MCP-compatible tool server. |
| Agent-to-agent calls | **A2A** (Agent-to-Agent) | Agents built in different frameworks. |
| Telemetry / tracing | **OpenTelemetry GenAI** semantic conventions (OpenLLMetry / OpenInference) | Any compatible observability backend. |
| Procedural memory | **Agent Skills** (`SKILL.md`) open standard | Any skills-compatible runtime. |
| Episodic / user memory | Your own storage schema over a standard store | Any vendor memory layer, or raw vector/graph DB. |
| Code execution | OCI containers, SSH, or an RPC boundary | Any sandbox provider or self-hosted runtime. |

Three traps to watch for when evaluating any component, surfaced repeatedly in 2026 tooling reviews:

- **"Self-hosted" is not the same as "open-source."** Some products let you self-host only under an enterprise agreement, or ship an open core whose useful features (often graph reasoning or advanced retrieval) are gated behind a paid cloud tier. Verify the license for the exact features you need.
- **Framework lock-in.** Memory and eval layers that are tied to one orchestration framework lose most of their value if you switch frameworks. Prefer standalone components if there is any chance you will re-platform.
- **Managed-only or API-key-only.** A required hosted API key usually signals a managed service. That can be fine — but confirm whether a self-hostable build exists if data residency or cost control matters to you.

The rest of this document names tools as examples, but every one of them sits behind one of the interfaces above. Swap freely.

---

## 4. System orchestration and the decoupled runtime

The backbone of a self-improving agent is an orchestration loop that is independent of both the client interface and the physical execution environment. You can hand-roll this loop or adopt a framework; either way the invariants below matter more than the choice of library.

### 4.1 Operational invariants of the agent loop

A robust loop — handling prompt assembly, model calls, tool routing, retries, and persistence — should hold to a few invariants across long autonomous runs:

1. **Prompt stability for prefix caching.** Keep the system prompt structurally immutable mid-conversation. A stable prefix lets the model reuse cached attention over the foundational prompt instead of recomputing it every turn, which cuts latency and token cost. Avoid mutating the prompt mid-session unless an explicit user action requires it.
2. **Observable execution.** Emit every tool call, reasoning step, and memory access to a telemetry layer. This serves two purposes: live progress for the user, and — critically — the structured traces that the learning loop (§5) consumes.
3. **Interruptibility.** Any in-flight model call or tool execution must be cancellable via an async signal or user input, so a runaway loop can be stopped without killing the process.
4. **Loose coupling.** The core loop should serve a CLI, an API server, batch jobs, and messaging gateways without modification. Optional subsystems (memory, execution backends) should attach through registries or plugin interfaces, not hardcoded dependencies.

> **Options (orchestration, June 2026 snapshot).** Build on a framework such as LangGraph (graph/checkpointing), CrewAI (role-based), the OpenAI Agents SDK or Claude Agent SDK (vendor-native, MCP-first), Google ADK, Pydantic AI (typed), LlamaIndex (retrieval-heavy), Microsoft Semantic Kernel / Agent Framework, smolagents, AutoGen/AG2, or Mastra (TypeScript) — or hand-roll the loop for full control. These are all open-source (MIT/Apache) as of mid 2026; the framework cost is usually negligible next to model spend. Anchor on **MCP** for tools and **A2A** for cross-framework calls so the orchestrator stays swappable. Choose by coordination model (graph vs. role vs. handoff) and by control-vs-velocity.

### 4.2 Orchestration flow

```
                 +-----------------------------------------+
                 |            Client Interfaces            |
                 | CLI  .  API server  .  chat / messaging |
                 +-----------------------------------------+
                                      |  standardized input
                                      v
   +---------------------------------------------------------------------+
   |                      Core Orchestration Engine                      |
   |                                                                     |
   |   +--------------+      +--------------+      +-------------------+ |
   |   |   Context    |  ->  |    Model     |  ->  |   Tool / Action   | |
   |   |  Assembler   |      |  Inference   |      |      Router       | |
   |   | frozen snap  |      | reasoning    |      | RPC . subagents   | |
   |   +------^-------+      +--------------+      +---------+---------+ |
   |          |                                             |            |
   |          |                                             v            |
   |   +------+-------+      +--------------+      +-------------------+ |
   |   |   Context    |  <-  |   Trace &    |  <-  |     Sandboxed     | |
   |   |  Compressor  |      |  Telemetry   |      |     Execution     | |
   |   +--------------+      +--------------+      +-------------------+ |
   +---------------------------------------------------------------------+
                                      |  structured traces + state
                                      v
   +---------------------------------------------------------------------+
   |                   State & Memory Persistence Layer                  |
   |       skills . episodic memory . user model . session database      |
   +---------------------------------------------------------------------+
```

The right-hand return path is the important part: execution produces telemetry, telemetry feeds trace capture, and traces feed both context compression (to stay within the window) and the offline learning loop described next.

---

## 5. The closed learning loop

The defining feature of a self-improving agent is a loop that lets the system analyze its own history, find recurring failures, and rewrite its textual directives to prevent them. The loop runs *on* the agent's configuration (prompts, skill files, tool descriptions, and sometimes harness code) and operates entirely through API calls and text edits — no weight updates.

### 5.1 Trace capture: knowing *why*, not just *that*

Improvement is impossible without granular observability. Knowing a task failed is not enough; you need the data to see *why*. A common technique is to intercept model calls at the client-library level (wrapping the SDK) so every call is captured automatically, then wrap the agent loop in a context manager that records the full input prompt, the reasoning trajectory, each tool request and its output, and any errors.

Serialize these into structured traces. Over hundreds of runs they become an empirical dataset mapping the agent's successes, recurring errors, and missed opportunities — the raw material for everything below.

> **Options (observability, June 2026 snapshot).** Instrument once against the **OpenTelemetry GenAI** conventions (via OpenLLMetry or OpenInference), then point traces at any backend: Langfuse (MIT, self-hostable), MLflow (Apache 2.0, full stack), Arize Phoenix, or Laminar for self-hosted/open options; Helicone or Portkey as drop-in gateway proxies; LangSmith, W&B Weave, or Braintrust as managed. Standardizing on the OTEL conventions means your instrumentation code does not change when you switch backend. Choose by self-host vs. managed, evaluation depth, and existing framework.

### 5.2 The optimization pipeline

Rather than treating execution as a black box, an evaluator reads the captured traces and proposes targeted text changes. A generic version of the pipeline:

1. **Select a target and baseline.** Map the agent's architecture and pick one asset to optimize (a tool description, a skill file, a prompt section). Snapshot its current version as the baseline.
2. **Build the evaluation set.** Mine real examples from the trace history, plus synthetic and hand-curated cases (see §6.2), and split into train / validation / holdout.
3. **Measure baseline performance.** Run detectors for failure signatures — infinite loops, premature give-ups, failed recoveries — to establish quantitative metrics on the current version.
4. **Propose candidates.** A reflective optimizer diagnoses the failure patterns in natural language and proposes text mutations aimed at them, prioritized by expected impact.
5. **Evaluate and compare.** Run candidates in parallel on the holdout set; compare accuracy, latency, and token cost against the baseline with a significance check. Optionally route the plan to a human for approval.
6. **Ship the winner.** Compile the best variant that passes all gates into a version-control branch and open a pull request with diffs, metrics, and before/after traces — so changes are reviewable and reversible.

> **Options (prompt/skill optimization, June 2026 snapshot).** DSPy with GEPA or MIPROv2 is the most cited reflective approach; alternatives include TextGrad, OPRO, APE, and AdalFlow. Because all of them operate on plain-text artifacts (your prompts and skill files), the optimizer itself is swappable and never needs to touch your model. One published reference pipeline reports a full optimization run costing roughly $2–10 — implementation-specific, but the structural point is general: evolving text is orders of magnitude cheaper than retraining.

### 5.3 The autonomous ratchet

When the loop runs unattended — typically during idle periods — it becomes a **ratchet**: it benchmarks the agent, generates trace insights, mutates the configuration, and re-evaluates. A change is kept *only* if it produces a statistically significant improvement with no guardrail regressions; otherwise it is reverted. This keep-or-revert rule is what makes improvement monotonic rather than a random walk — the system can only get better or stay the same on the measured metrics.

```
   +-------------------------------------------------------------------+
   |                    The Autonomous Ratchet Loop                    |
   +-------------------------------------------------------------------+

      +---------------------+
      |  1. Capture traces  |  <-----------  mined usage history
      |     from telemetry  |
      +----------+----------+
                 |  identify recurring failure patterns
                 v
      +---------------------+
      |  2. Propose         |  ----------->  candidate prompts / code
      |     candidates      |               (reflective mutation)
      +----------+----------+
                 |
                 v
      +---------------------+        +--------------------------------+
      |  3. Validate        |  --->  |  constraint gates:             |
      |     against gates   |        |  tests . size caps . benchmarks|
      +----------+----------+        +--------------------------------+
                 |  pass?
        +--------+--------+
        | yes             | no
        v                 v
   +-------------+   +-------------+
   |  4. Keep:   |   |   Revert:   |
   |  open PR /  |   |   discard   |
   |  deploy     |   |   variant   |
   +-------------+   +-------------+
```

---

## 6. Guardrails and evaluation: improving without drifting

Unconstrained self-modification is dangerous. The main failure mode is **drift**: the agent over-optimizes for a narrow set of measured tasks and quietly degrades elsewhere, or its prompts bloat until they break caching and inflate cost. Every candidate change should pass a fixed sequence of gates before it can ship.

### 6.1 Constraint gates

- **Size limits.** Cap the length of evolved assets to protect prefix-caching boundaries and control cost. Block any change that exceeds its baseline size by more than a set margin. *(Example caps from one implementation: tool descriptions ≤ 500 chars, parameter descriptions ≤ 200 chars, skill files ≤ ~15 KB, prompt sections ≤ 120% of baseline. Tune to your own token budget.)*
- **Semantic integrity / cross-asset evaluation.** Evaluate all tool descriptions together, not in isolation. A common pathology is a "parasitic" description that becomes so broad the model selects that tool inappropriately. Penalize any candidate that regresses *another* tool's selection accuracy.
- **Identity preservation.** For system-prompt changes, verify the agent's core traits survive (e.g. it stays direct, admits uncertainty) and that it doesn't hallucinate capabilities it lacks on the current interface (e.g. emitting terminal color codes over a plain-text chat).
- **Functional validation.** When the loop edits actual harness code, the change must pass the full test suite — no exceptions — before it is eligible to ship.

### 6.2 Evaluation data strategy

Mutations are only as good as the data they're judged against. Use three complementary sources:

- **Synthetic generation.** For new capabilities with no usage history, have a strong model read the target asset and bootstrap a small set (order of dozens) of realistic cases. Score against a rubric describing correct *behavior*, not exact text.
- **Mined sessions.** Query the session store for real uses of the target asset, score them with an LLM-as-judge; high-scoring sessions become positive examples, low-scoring ones become diagnostic material for reflection.
- **Hand-curated golden sets.** For mission-critical paths, keep human-authored cases with expected outputs as ground truth.

A caution on LLM-as-judge: judge models carry documented biases (verbosity bias, position bias). Calibrate judges against human-labeled examples, and don't let a single judge model silently define "correct."

### 6.3 Regression benchmarking

The final gate runs the modified agent against holistic benchmark suites to confirm global stability before any change merges. Reject a candidate that improves a targeted metric but regresses the overall suite beyond a small threshold.

> **Options (evaluation, June 2026 snapshot).** Public agent benchmarks worth integrating include **τ²-bench** (Sierra Research; dual-control tool-agent-user tasks) and **Terminal-Bench** (terminal/sysadmin/software tasks). For your own suites, code-first eval frameworks include DeepEval (pytest-style, agent metrics), Ragas (RAG-specific), promptfoo (YAML, cross-model and red-teaming/security), Inspect, EleutherAI's lm-evaluation-harness, and OpenAI Evals; managed lifecycle platforms include Braintrust, Arize Phoenix, and Opik. Keep evaluation datasets versioned and scoring expressed as code so the same checks run in CI and in production. Choose by offline-vs-production, the workload type (agent / RAG / security), and CI integration.

---

## 7. Procedural memory: skills and progressive disclosure

An agent cannot hold instructions for every workflow in a single static prompt — doing so blows the context budget, dilutes attention, and inflates cost. The modern pattern is **on-demand procedural memory**, standardized as the **Agent Skills** open standard (published December 2025, maintained at agentskills.io, adopted across many tools). Because it is an open standard rather than a product, it is a safe portability anchor — a skill authored once works across compatible runtimes unchanged.

### 7.1 The skill package

A skill is a folder with a strict, minimal shape:

- **`SKILL.md` (required)** — YAML frontmatter (`name`, `description`) plus Markdown instructions. The frontmatter is the only mandatory surface area (`name` ≤ 64 chars, `description` ≤ 1,024 chars).
- **`scripts/` (optional)** — executable code the agent may run during the skill.
- **`references/` (optional)** — documentation or domain knowledge loaded only when needed.
- **`assets/` (optional)** — templates and supporting files.

### 7.2 Progressive disclosure

The reason skills scale is **progressive disclosure** — loading detail in three stages so the context window holds only what the current task needs:

```
   +-------------------------------------------------------------------+
   |                  Progressive Disclosure (3 levels)                |
   +-------------------------------------------------------------------+

   Level 0 - DISCOVERY   (always loaded)
   +---------------------------------------------------------------+
   |  Index of name + description for every skill                  |
   |  ~30-50 tokens each  ->  broad awareness, tiny footprint      |
   +-------------------------------+-------------------------------+
                                   |  task semantically matches a skill
                                   v
   Level 1 - ACTIVATION  (loaded on match)
   +---------------------------------------------------------------+
   |  Agent reads the full SKILL.md into context                   |
   |  e.g.  skill_view("deploy-k8s")                               |
   +-------------------------------+-------------------------------+
                                   |  execution needs extra data
                                   v
   Level 2 - EXECUTION   (loaded on demand)
   +---------------------------------------------------------------+
   |  Agent pulls a specific reference or runs a bundled script    |
   |  e.g.  skill_view("deploy-k8s", "references/schema.json")     |
   +---------------------------------------------------------------+
```

At discovery, only the lightweight index is in the prompt; full instructions and reference files are pulled in only when a task actually calls for them. This is the same short-term/long-term split human memory uses, applied to context. The same idea applies to live tools exposed over **MCP** and to a framework's native tool registry — keep the always-loaded surface tiny, and pull depth on demand.

### 7.3 Autonomous skill generation

A self-improving system should also *write* its own procedural memory. Rather than creating skills randomly, trigger documentation on meaningful signals: completing a genuinely complex task (e.g. one needing many distinct tool calls), discovering a working path after several dead ends, or receiving a user correction. On a trigger, the agent synthesizes the successful trajectory into a new `SKILL.md`.

A minimal management API gives the agent control over its own skill directory:

| Action | Purpose |
| :--- | :--- |
| **create** | Build a new skill folder and populate `SKILL.md`. |
| **patch** | Targeted string replacement for small fixes. |
| **edit** | Replace the full instruction file for a major rewrite. |
| **write_file** | Add or update a supporting file under `references/` or `scripts/`. |

When the agent later finds an inefficiency, it patches the skill in place — so capabilities compound across sessions instead of being rediscovered each time.

---

## 8. Episodic and user memory

Beyond procedural skills, an agent needs to track facts, situational state, and a model of the people it works with. Without structure, long-term memory degenerates into an unstructured log that *hurts* retrieval. This is also the layer with the most crowded, fastest-moving tool market — so it is exactly where you most want to depend on your own schema rather than a vendor.

### 8.1 Bounded local memory and the frozen-snapshot pattern

For fast-access facts, keep small, hard-capped local files — for example a short ledger of environment facts and lessons learned, and a condensed user-profile file. The caps are deliberate: a tight budget forces the agent to *distill* knowledge rather than append indefinitely. *(One implementation uses ~2,200 chars for the agent ledger and ~1,375 chars for the user profile — calibrate to your budget.)*

The **frozen-snapshot pattern** preserves prefix caching: at session start, read these files once and render them into the highest-priority part of the system prompt as a single immutable block. If the agent edits memory mid-session, persist it to disk immediately but **do not** inject it into the live prompt until the next session boundary. This keeps the cached prefix intact for the whole conversation.

When data outgrows the caps, fall back to a searchable store. This gives effectively unlimited history the agent can query on demand without bloating the core prompt.

> **Options (storage backend, June 2026 snapshot).** Self-hostable building blocks: SQLite full-text search for small/local cases; pgvector (Postgres), Qdrant, Weaviate, Chroma, Milvus, LanceDB, or Redis for vectors; Neo4j (Community) for graphs. Managed: Pinecone and others. Many higher-level memory layers (below) embed a store, so you may not need a separate one. Choose by ops burden, scale, and whether you need hybrid (keyword + vector) or graph search.

### 8.2 Reasoning-first vs. retrieval-first memory

There are two broad philosophies, and the distinction matters more than the brand:

- **Retrieval-first** memory stores raw statements (and embeddings) and matches them later by similarity. Simple and transparent; can miss anything not stated verbatim.
- **Reasoning-first** memory runs a background process that *derives* conclusions from the message stream (preferences, beliefs, contradictions) that were never explicitly stated, and synthesizes longer-term patterns during idle periods. Instead of searching an embedding store directly, the agent asks the memory layer a natural-language question ("How technical is this person?") and gets a synthesized answer.

Two structural ideas from the reasoning-first camp are worth borrowing regardless of which tool you pick:

- **Multi-entity ("peer") modeling.** Drop the rigid "user vs. assistant" split. Treat every participant — human, subagent, external bot — as a first-class entity with its own evolving representation, and control which entities model which others. This cleanly prevents context cross-contamination when multiple agents share an environment.
- **Query by question, not by vector.** Expose memory through a natural-language query interface so retrieval logic can change behind it.

> **Options (memory layer, June 2026 snapshot — verify license/tier before adopting).** Fully open-source and self-hostable: Letta (formerly MemGPT; OS-style tiered memory), Cognee (graph + vector, on-prem/air-gapped), Zep's Graphiti engine (temporal knowledge graph). Open core with paid tiers or cloud-gated features: Mem0 (graph features behind Pro), Honcho (AGPL; reasoning-first/peer model self-hostable, with a managed cloud that uses an API key). Framework-coupled: LangMem (LangGraph), LlamaIndex Memory. Backend-style: Redis Agent Memory Server. The diagram below illustrates one concrete reasoning-first design (a peer-centric, two-layer injection); the *pattern* is portable even if you implement it yourself over a plain store. Choose by retrieval strategy, self-host/OSS requirement, graph need, and lock-in risk.

### 8.3 Two-layer context injection (one concrete reasoning-first design)

```
   +-------------------------------------------------------------------+
   |               Reasoning-First (Peer-Centric) Memory               |
   +-------------------------------------------------------------------+

      message stream  ->  background reasoning
                          (derive facts . synthesize patterns)
                                     |
                                     v
                  +-----------------------------------+
                  |     Peer-Centric Knowledge Store  |
                  | (humans & agents modeled as peers)|
                  +-----------------+-----------------+
                       +------------+------------+
                       v                         v
        +----------------------------+  +----------------------------+
        | LAYER 1 - Base Context     |  | LAYER 2 - Dialectic        |
        | (slow cadence)             |  | Supplement (fast cadence)  |
        |  . session summary         |  |  . real-time intent        |
        |  . stable user profile     |  |  . current cognitive load  |
        |  . relational "peer card"  |  |  . immediate need          |
        | "Who am I talking to?"     |  | "What matters right now?"  |
        +-------------+--------------+  +-------------+--------------+
                      +----------+-----------+--------+
                                 v
                  +-----------------------------------+
                  |     Agent System-Prompt Assembly  |
                  +-----------------------------------+
```

When you integrate a reasoning-first memory provider, assemble its output into the prompt in two layers refreshed at different cadences: a slow-moving **base context** (session summary, stable profile, relational card) answering "who am I talking to," and a fast-moving **dialectic supplement** (real-time intent, current need) answering "what matters right now." To avoid cold-start latency, **prewarm** at session start: fire an asynchronous evaluation of recent history in the background so the agent already holds synthesized context before the user's first message.

---

## 9. Execution safety and decoupled runtimes

Once an agent can write code, generate its own scripts, and call external APIs autonomously, the execution environment becomes an attack surface — agent-generated code can read environment variables, open outbound connections, write to disk, and, with weak isolation, escape to the host. Two design choices contain the risk.

### 9.1 Sandboxing and the isolation spectrum

Abstract the runtime away from the reasoning loop and drive remote execution backends over RPC. Isolation strength runs along a spectrum; pick the weakest point that meets your threat model, since stronger isolation usually costs cold-start time:

- **Plain containers (e.g. Docker)** — share the host kernel; weakest boundary. Fine for trusted code, risky for arbitrary agent output.
- **User-space kernel (gVisor)** — intercepts syscalls in userspace; stronger, with some performance and GPU-passthrough limits.
- **microVMs (Firecracker, Kata Containers)** — a dedicated kernel per workload via hardware virtualization; strongest boundary, and snapshot/restore enables fast state persistence across turns.

> **Options (sandboxing, June 2026 snapshot).** Self-host the isolation layer directly with Docker, gVisor, or Firecracker/Kata microVMs. Managed sandbox services: E2B (Firecracker), Daytona (fast cold start), Modal (gVisor, GPU), Vercel Sandbox (Firecracker), Northflank (Kata/gVisor/Firecracker, bring-your-own-cloud), Together, and Cloudflare. Choose by isolation strength, cold-start latency, session-length and GPU needs, and whether you require bring-your-own-cloud for data residency.

A serverless/hibernating runtime is especially useful for autonomy: the environment instantiates on demand, runs the work, and suspends when idle, so an agent can hold a multi-week asynchronous workflow without paying for an always-on server.

### 9.2 Subagent delegation

For parallelizable work, the orchestrator spawns short-lived **subagents**, each in its own conversational thread and isolated execution namespace to prevent collisions, communicating over structured RPC (or **A2A** across frameworks). From the orchestrator's view, delegation collapses a long multi-step chain into a single near-zero-context turn — expanding effective throughput while keeping each subagent's context clean.

---

## 10. Strategic implications

A few second-order effects follow from moving improvement out of training and into the execution layer.

- **Reliability decouples from parameter count.** Turning historical failures into structural constraints and refined prompts compounds reliability without frontier-scale models. A well-harnessed small model can beat a raw large one on a specific workflow.
- **The data moat shifts to interaction history.** As the system builds detailed models of its users, tools, and workflows, the durable asset becomes the accumulated procedural skills and longitudinal interaction history — not the base weights, which are increasingly commoditized.
- **Compute moves from training to inference-time optimization.** Spend shifts from centralized pre-training toward continuous, cheap optimization runs. Where a text-evolution run costs single-digit dollars, recursive self-improvement is viable for individuals and small teams, not just hyperscalers.

These are directional arguments — well-supported in their mechanics but extrapolated in their economic reach. Hold them as hypotheses to test against your own deployment.

---

## 11. Applying this to your own system — a checklist

A pragmatic order of adoption. You do not need all of it; each item stands alone and compounds with the others.

**Foundations (do these first)**
- [ ] Decide your portability anchors (§3) before picking tools: model API, MCP, A2A, OpenTelemetry, Agent Skills, your own memory schema.
- [ ] Put each external component behind a thin internal interface so it can be swapped.
- [ ] Instrument the agent loop for full trace capture (inputs, reasoning, tool calls, outputs, errors) — nothing else works without this.
- [ ] Freeze the system prompt mid-session and assemble memory as an immutable snapshot to preserve prefix caching.
- [ ] Make every in-flight call interruptible.

**Procedural memory**
- [ ] Adopt the Agent Skills format (`SKILL.md` + optional `scripts/`, `references/`, `assets/`).
- [ ] Implement progressive disclosure: index at startup, full skill on match, references on demand.
- [ ] Add triggers and an API for the agent to write and patch its own skills.

**The learning loop**
- [ ] Build the offline pipeline: select target → build eval set → measure baseline → propose (a reflective optimizer) → evaluate on holdout → PR.
- [ ] Make every change ship as a reviewable, revertible diff.
- [ ] Run the ratchet on idle time with a strict keep-or-revert rule.

**Guardrails (before you let it run unattended)**
- [ ] Enforce size caps and a max-growth margin on evolved assets.
- [ ] Evaluate tool descriptions jointly to catch parasitic selection.
- [ ] Add identity-preservation and capability-hallucination checks for prompt edits.
- [ ] Require the full test suite to pass for any harness-code change.
- [ ] Gate every merge on a regression benchmark; calibrate any LLM-as-judge against human labels.

**Memory & people**
- [ ] Keep bounded local memory with hard caps that force distillation.
- [ ] For multi-session or multi-user products, choose reasoning-first vs. retrieval-first deliberately; expose memory by natural-language query so the backend stays swappable.

**Safety**
- [ ] Match sandbox isolation strength to your threat model; never run agent-generated code with a shared kernel if you can't trust it.
- [ ] Use isolated subagents for parallel work; never share execution namespaces.

---

## 12. Component landscape at a glance (June 2026 snapshot)

This table is a swap-reference, not a recommendation. Tool lists age quickly; the **interface** column is the part meant to last. All option lists are non-exhaustive and current as of June 2026 — re-verify licenses, tiers, and availability before adopting.

| Component | Depend on this interface | Representative options (illustrative) | Choose by |
| :--- | :--- | :--- | :--- |
| Model | OpenAI-compatible API; router (LiteLLM, OpenRouter) | any frontier or open-weights model, hosted or local | capability, cost, latency, self-host need |
| Orchestration | MCP (tools), A2A (agents) | LangGraph, CrewAI, OpenAI Agents SDK, Claude Agent SDK, Google ADK, Pydantic AI, LlamaIndex, Semantic Kernel, smolagents, AutoGen/AG2, Mastra; or hand-rolled | coordination model, control vs. velocity, language |
| Observability | OpenTelemetry GenAI (OpenLLMetry / OpenInference) | Langfuse, MLflow, Arize Phoenix, Laminar, Helicone/Portkey, LangSmith, W&B Weave, Braintrust | self-host vs. managed, eval depth, framework |
| Prompt/skill optimization | plain-text artifacts (prompts & skills as files) | DSPy (+GEPA, MIPROv2), TextGrad, OPRO, APE, AdalFlow | reflective vs. search, integration |
| Procedural memory | Agent Skills (`SKILL.md`) | Agent Skills runtimes; MCP for live tools; framework tool registries | portability, ecosystem |
| Episodic / user memory | your own schema; reasoning- vs. retrieval-first | Letta, Cognee, Zep/Graphiti, Mem0, Honcho, LangMem, Redis Agent Memory Server | retrieval strategy, OSS/self-host, graph need, lock-in |
| Vector / graph store | SQL/standard query interfaces | pgvector, Qdrant, Weaviate, Chroma, Milvus, LanceDB, Redis, Neo4j; Pinecone (managed) | ops burden, scale, hybrid/graph search |
| Sandboxed execution | OCI containers / SSH / RPC | self-host: Docker, gVisor, Firecracker/Kata; managed: E2B, Daytona, Modal, Vercel Sandbox, Northflank, Cloudflare | isolation strength, cold start, session/GPU, BYOC |
| Evaluation | versioned datasets + scoring as code | benchmarks: τ²-bench, Terminal-Bench; frameworks: DeepEval, Ragas, promptfoo, Inspect, lm-eval-harness, OpenAI Evals, Braintrust, Phoenix, Opik | offline vs. prod, agent/RAG/security, CI |

---

## 13. Source confidence

Weight the claims in this document according to their grounding:

- **Peer-reviewed / standardized (high confidence).** GEPA's results (ICLR 2026); the Agent Skills open standard and progressive-disclosure model; MCP, A2A, and OpenTelemetry GenAI as interoperability standards; τ²-bench and Terminal-Bench as benchmarks; catastrophic forgetting and prefix caching as established phenomena.
- **Documented in primary project sources (solid, but specific implementations).** Reasoning-first/peer-centric memory designs; reflective optimization pipelines built on DSPy + GEPA with PR-based review and constraint gates. Real and well-documented, but reflecting particular teams' design choices.
- **Market snapshot, will age fast (verify before relying).** Every named tool and its licensing/tier/self-hosting status in §3–§12. Accurate to the best available information as of June 2026; this is exactly the category the substitution principle (§3) exists to protect you from.
- **Implementation-specific calibration numbers (treat as examples).** All exact character caps, token budgets, and the ~$2–10 cost-per-run figure. Useful as starting points; tune to your system.
- **Anecdotal (treat with caution).** Single-run improvement figures from forum posts (e.g. a "~25% gain after one cycle" claim). Plausible and directionally consistent with the literature, but not independently verified — validate on your own benchmarks.

---

## References

**Standards & interfaces**
1. Agent Skills open standard. https://agentskills.io/home
2. Model Context Protocol (MCP). https://modelcontextprotocol.io
3. Agent-to-Agent (A2A) protocol. https://a2a-protocol.org
4. OpenTelemetry GenAI semantic conventions. https://opentelemetry.io/docs/specs/semconv/gen-ai/

**Research & benchmarks**
5. Agrawal, L. A., et al. (2025). *GEPA: Reflective Prompt Evolution Can Outperform Reinforcement Learning.* arXiv:2507.19457. https://arxiv.org/abs/2507.19457
6. DSPy — framework for programming (not prompting) LLMs. https://dspy.ai
7. Sierra Research. *τ²-bench.* https://github.com/sierra-research/tau2-bench · arXiv:2506.07982
8. *Terminal-Bench.* https://www.tbench.ai

**Component categories (June 2026 landscape; verify before adopting)**
9. Memory layers — Letta/MemGPT (https://github.com/letta-ai/letta), Cognee (https://www.cognee.ai), Zep/Graphiti (https://github.com/getzep/graphiti), Mem0 (https://github.com/mem0ai/mem0), Honcho (https://github.com/plastic-labs/honcho).
10. Observability — Langfuse (https://langfuse.com), MLflow (https://mlflow.org), Arize Phoenix (https://github.com/Arize-ai/phoenix), Laminar (https://laminar.sh).
11. Sandboxes — E2B (https://e2b.dev), Daytona (https://www.daytona.io), Modal (https://modal.com), Northflank (https://northflank.com); Firecracker (https://firecracker-microvm.github.io), gVisor (https://gvisor.dev).
12. Evaluation — DeepEval (https://github.com/confident-ai/deepeval), Ragas (https://github.com/explodinggradients/ragas), promptfoo (https://www.promptfoo.dev), Inspect (https://inspect.aisi.org.uk).
13. Orchestration — LangGraph (https://github.com/langchain-ai/langgraph), CrewAI (https://www.crewai.com), OpenAI Agents SDK, Pydantic AI (https://github.com/pydantic/pydantic-ai).

*Last verified: June 2, 2026. The component landscape (§3, §8, §9, §12 and references 9–13) is a fast-moving market snapshot — confirm current licenses, tiers, ownership, and self-hosting options against upstream sources before building.*
