# Go Migration Case Studies & AI-Assisted Go Development

**Date:** 2026-03-07
**Type:** Spike Research
**Status:** Final

## Research Questions

1. What do real Python-to-Go migration case studies reveal about timelines, approaches, and pitfalls?
2. What is the measured quality of AI-generated Go code, and what are the most common mistakes?
3. How effective are multi-agent workflows (Claude Code teams) for Go projects?
4. Does Watermill's GoChannel backend work reliably for local/embedded event buses?
5. Is severity1/claude-agent-sdk-go viable for production use?

## Summary

Python-to-Go migrations are well-documented across multiple companies (Khan Academy, Lovable,
Uber, Digger/OpenTaco, Winder AI). The consistent pattern is: incremental migration with
validation against the original system succeeds; big-bang rewrites fail catastrophically.
AI-assisted Go code generation is measurably worse than human code (1.7x more issues), with
Go-specific pitfalls in error handling and package hallucination, though Go's compiler catches
many issues that would slip through in Python. Multi-agent workflows with Claude Code are
emerging but best limited to 3-4 specialized agents. Watermill GoChannel works for single-
process CLI tools but has documented limitations (no persistence, no ordering, no consumer
groups). The severity1/claude-agent-sdk-go has low adoption, single-maintainer risk, and
the underlying SDK architecture (CLI subprocess wrapper) introduces fundamental pitfalls.

---

## 1. Python-to-Go Migration Case Studies

### 1.1 Khan Academy (2019-2021) -- Incremental, Production Backend

**Project:** Migration of entire Python 2 monolith backend to Go services.

**Scale:** Over 500,000 lines of Go in production by May 2021.

**Timeline:** Started December 2019. Migration still ongoing at time of reporting (June 2025).

**Approach:** Incremental migration ("as incrementally as can be"). Services migrated one at a
time. Project internally named "Goliath" for incremental delivery.

**Team:** When starting, no engineers knew Go beyond validation experiments. All backend and
full-stack engineers now write Go. Engineers reported positive sentiment: "it's easy to read
and write" and "I like Go more the more I work with it."

**Key findings:**

- **2.7x code expansion:** One system component required 2.7x more Go lines than Python. Some
  complexity came from replacing local function calls with cross-service queries, but Go's
  verbosity is a real factor.
- **Missing generics was biggest complaint** (now resolved in Go 1.18+). Internal libraries
  and slice operations suffered most.
- **Performance gains were dramatic:** Bulk data operations reduced Google Cloud Datastore
  contention warnings from ~100/hour to near zero. Loading 1,000-student classes dropped from
  28 seconds (Python) to 4 seconds (Go).
- **Engineers used `sync` package far more than channels.** Channels are Go's marquee feature
  but real-world migration code relies more on mutexes and sync primitives.
- **Engineers liked Go's error model:** "Being able to call a function that doesn't return an
  error and know for sure that it must succeed is really nice."

**Source:** [Khan Academy Blog - Half a Million Lines of Go](https://blog.khanacademy.org/half-a-million-lines-of-go/)

---

### 1.2 Lovable (2024-2025) -- Full-Stack AI Platform

**Project:** Lovable's AI-powered full-stack web application builder, migrated from Python to Go.

**Scale:** 42,000 lines of Python migrated.

**Timeline:** Not explicitly disclosed.

**Approach:** Complete rewrite with custom declarative dependency injection framework that
dynamically creates a node graph for executing components as dependencies become available.

**Key findings:**

- **Deployment time:** 15 minutes to 3 minutes (80% reduction)
- **Request speed:** 12% faster average response times
- **Infrastructure:** Reduced from 200 server instances to 10 (95% reduction, massive cost
  savings)
- **Concurrency:** Enabled handling 50+ HTTP requests within a single chat request
- **Primary motivation:** "Python is great, but not for what we do. Our loads are highly
  concurrent and parallel." -- Viktor Eriksson, Lovable

**Source:** [Lovable Blog - From Python to Go](https://lovable.dev/blog/from-python-to-go)

---

### 1.3 Uber (2016-2018) -- Production Datastore "Project Frontless"

**Project:** Rewrite of the front-end (sharding layer) of Uber's Schemaless datastore from
Python to Go.

**Scale:** Several thousand Python worker nodes (Flask/uWSGI/NGINX) rewritten.

**Timeline:** Estimated 6 months; actual timeline extended due to moving target (new features
landing in production Python during migration).

**Approach:** Incremental, endpoint-by-endpoint with production validation. Key decision: every
request to a Python worker must yield the same result in the Go worker. Each endpoint was
reimplemented and validated in production before going live.

**Key findings:**

- **Validation-driven migration** was critical -- they compared Python and Go responses in
  production side-by-side.
- **Moving target problem:** During 6-month estimated rewrite, new features and bug fixes
  kept landing in Python, making Go implementation a moving target.
- **Python's uWSGI model couldn't scale:** Each uWSGI process handled one request at a time,
  each as its own Linux process with overhead. Go goroutines eliminated this.
- **Zero downtime achieved** despite rewriting critical datastore infrastructure.

**Broader context:** Uber also rewrote most of their Marketplace stack from Python (Flask/uWSGI)
to Go. "Blocks on network calls and I/O slowed their services in weird ways, requiring more
capacity to get the same request throughput."

**Source:** [Uber Blog - Code Migration in Production](https://www.uber.com/blog/schemaless-rewrite/)

---

### 1.4 Digger/OpenTaco (2023) -- IaC CLI Tool

**Project:** Infrastructure-as-Code orchestration CLI tool (Terraform wrapper), rewritten from
Python to Go.

**Scale:** "Wasn't huge" -- essentially a wrapper on top of Terraform that managed state in S3
and metadata in DynamoDB.

**Timeline:** One week, with team having no prior Go experience.

**Approach:** Complete rewrite from scratch (justified by small codebase).

**Key findings:**

- **30x faster runtimes** compared to Python version.
- **Community expectations mattered:** "Users asked several times what language Digger was
  written in, and specifically whether it was in Go. People cared."
- **Single binary compilation** eliminated Docker wrapping for GitHub Actions.
- **Compiler-based guarantees** gave more confidence than Python's dynamic typing.
- **Project later rebranded** to OpenTaco (December 2025).

**Source:** [Digger Blog - Digger is now in Golang](https://blog.digger.dev/digger-is-now-in-golang/)

---

### 1.5 Winder AI / Kodit (2026) -- AI-Assisted Python-to-Go Migration

**Project:** Production Python codebase (Kodit) migrated to Go using Claude Code as primary
migration tool.

**Timeline:** ~2 hours of unattended automation for migration checklist completion, followed by
substantial debugging and refactoring.

**Approach:** Three-phase structured methodology:

1. **Discovery:** Targeted prompts to understand codebase structure, bounded contexts, domain
   vocabulary, pattern mapping, dependencies. Two-pass discovery (individual prompts then
   synthesis) proved more effective than one-shot generation.
2. **Design Files:** Created CLAUDE.md (domain context, translation rules, coding standards)
   and MIGRATION.md (ordered task list with dependency tracking and progress checkboxes).
3. **Automated Migration:** Bash script running Claude Code iteratively; AI reads MIGRATION.md
   to determine next tasks and updates progress after each session.

**Key findings -- What worked:**

- AI successfully generated "thousands of lines of correct, idiomatic Go" from Python.
- Explicit translation rules mapping Python patterns to Go equivalents (classes to structs,
  exceptions to error handling, decorators to constructors) were essential.

**Key findings -- What failed:**

- **Dead code accumulation:** Orphaned functions and unused types accumulated during
  refactoring, masked by Go's import system creating "little islands that appear to be in use."
- **Architectural drift:** AI defaulted to Go's `internal/` directory without explicit guidance,
  requiring significant refactoring when a public API was needed.
- **Phantom features:** AI reconstructed deprecated database schema features from old code
  remnants, causing "searches return zero results because data was being written to the wrong
  table."
- **Missing integration tests:** Unit tests passed, code compiled, but end-to-end functionality
  failed. "The AI never ran the application as a whole. It verified each component in isolation
  but never wired them together."
- **Context window limits:** Large refactoring tasks exceeded token capacity, causing missed
  implementations and "in memory" workarounds.

**Recommendations for AI-assisted migration:**

1. Mandate smoke tests after each major phase
2. Define public Go API before code generation
3. Make dead code checks explicit workflow steps
4. Test data migration early with real data
5. Design Go client APIs as first-class concerns
6. Clean up deprecated code references before starting

**Core insight:** "The AI is a powerful but literal executor: fast, tireless, and incapable of
questioning whether the instructions are complete."

**Source:** [Winder AI - Python to Go Migration with Claude Code](https://winder.ai/python-to-go-migration-with-claude-code/)

---

### 1.6 Anti-Case-Study: Spring Boot to Go Migration Disaster

While not Python-specific, this widely-cited post-mortem illustrates big-bang rewrite risks:

- **Lost $2.4M, 8 engineers quit, 14 months wasted**
- **Ended up with dual stacks** in production (original and incomplete Go)
- **Root cause:** Big-bang rewrite approach without incremental validation

**Source:** [Medium - Spring Boot to Go Migration Disaster](https://medium.com/@the_atomic_architect/spring-boot-to-go-migration-disaster-24m-loss-post-mortem-e31925b5acc7)

---

### 1.7 Cross-Case Pattern Analysis

| Case Study | Approach | Codebase Size | Timeline | Outcome |
|---|---|---|---|---|
| Khan Academy | Incremental | 500k+ LOC | 2+ years | Success |
| Lovable | Full rewrite | 42k LOC | Undisclosed | Success |
| Uber | Incremental + validation | Thousands of nodes | 6+ months | Success |
| Digger | Full rewrite | Small (<5k LOC) | 1 week | Success |
| Winder/Kodit | AI-assisted full | Medium | Days + debugging | Partial success |
| Spring Boot disaster | Big bang | Large | 14 months | Failure ($2.4M) |

**Pattern:** Full rewrites succeed only for small codebases (<50k LOC). For larger codebases,
incremental migration with production validation is the only proven approach. The $2.4M
failure reinforces that big-bang rewrites of large systems remain high-risk.

**Relevance to alty:** alty's Python codebase is currently ~800 tests / medium size. An
incremental approach with the Strangler Fig pattern (migrating bounded context by bounded
context) is the safest path. The AI-assisted migration approach from Winder AI is relevant
if augmented with integration tests and dead-code cleanup gates.

---

## 2. AI-Generated Go Code Quality

### 2.1 General AI Code Quality (All Languages)

**CodeRabbit Study (2025):** Analysis of 470 open-source GitHub PRs (320 AI-co-authored, 150
human-only):

| Issue Type | AI vs Human Rate |
|---|---|
| Logic & correctness errors | 1.75x more |
| Code quality & maintainability | 1.64x more |
| Security findings | 1.57x more |
| Performance issues | 1.42x more |
| Readability problems | 3.0x more |
| Error handling gaps | ~2.0x more |
| Formatting inconsistencies | 2.66x more |

**Key insight:** "Humans and AI make identical types of mistakes, but AI produces them at
significantly higher volumes without adequate safeguards."

**Source:** [CodeRabbit - State of AI vs Human Code Generation](https://www.coderabbit.ai/blog/state-of-ai-vs-human-code-generation-report)

**Cortex 2026 Benchmark Report:** PRs per author increased 20% year-over-year, but incidents
per PR increased 23.5% and change failure rates rose ~30%.

**Source:** [Second Talent - AI-Generated Code Quality Metrics 2026](https://www.secondtalent.com/resources/ai-generated-code-quality-metrics-and-statistics-for-2026/)

### 2.2 Go Developer Survey on AI Quality (2025)

The official 2025 Go Developer Survey provides the most authoritative Go-specific data:

- **53% reported AI creates non-functional code** as primary problem
- **30% cited poor quality** even when code works
- **55% satisfied** with AI tools (vs 90%+ satisfied with Go itself)
- **Only 17%** use AI as unsupervised agents as primary method
- **53% use AI daily** but satisfaction remains middling

**Developer quotes:**

> "I'm never satisfied with code quality or consistency, it never follows the practices I
> want to." -- 3-10 years / Financial services

> "All AI tools tend to hallucinate quickly when working with medium-to-large codebases
> (10k+ lines of code). They can explain code effectively but struggle to generate new,
> complex features" -- 3-10 years / Retail

> "Despite numerous efforts to make it write code in an established codebase, it would take
> too much effort to steer it to follow the practices in the project... I also found it
> mentally taxing to review AI generated code and that overhead kills the productivity
> potential in writing code." -- 10+ years / Technology

**Best use cases for AI in Go:** Generating unit tests, writing boilerplate, enhanced
autocompletion, refactoring, documentation generation -- tasks where code quality is less
critical.

**Source:** [Go Blog - Results from the 2025 Go Developer Survey](https://go.dev/blog/survey2025)

### 2.3 Package Hallucination Rates

**"We Have a Package for You!" (USENIX Security 2025):**
- 756,000 code samples from 16 AI models tested
- ~20% recommended non-existent packages overall
- 43% of hallucinated packages were repeated across 10+ queries (exploitable)
- **Go was NOT tested** (study focused on Python/JavaScript with PyPI/npm registries)
- **Go had zero cross-language hallucinations** when checking if Python hallucinations matched
  Go packages (suggests Go ecosystem is more distinct)

**Source:** [arXiv 2406.10279 - Package Hallucinations by Code Generating LLMs](https://arxiv.org/abs/2406.10279)

**"Importing Phantoms" (arXiv 2501.19012, January 2025):**
- Tested 11 models (7B to 200B parameters)
- Mean hallucination rates: JavaScript 14.73%, Python 23.14%, Rust 24.74%
- **Go was NOT tested** (focused on Python, JavaScript, Rust)
- **Strong inverse correlation:** Larger models hallucinate less (rho = -0.593, p = 0.00028)
- Range: 0.22% (Nemotron-Llama-3.1) to 46.15% (Granite-3.0 for Python)

**Source:** [arXiv 2501.19012 - Importing Phantoms](https://arxiv.org/abs/2501.19012)

**"Library Hallucinations in LLMs" (arXiv 2509.22202, September 2025):**
- Up to 84% hallucination rate when asked about libraries "from 2025"
- One-character misspellings cause hallucinations in up to 26% of tasks
- Fake libraries used in up to 99% of tasks when explicitly injected

**Source:** [arXiv 2509.22202 - Library Hallucinations in LLMs](https://arxiv.org/pdf/2509.22202)

**Important note on Go:** While Go is excluded from all three major hallucination studies, the
Go compiler provides a natural defense: any hallucinated import fails at `go build` time with
a clear "cannot find module" error. This is fundamentally different from Python/JavaScript
where hallucinated packages can be installed (and exploited via slopsquatting). Go's module
proxy (proxy.golang.org) also provides an authoritative package registry.

### 2.4 Go-Specific AI Mistakes (Compiled from Multiple Sources)

Based on our prior research (see `20260307_2_go_team_development_patterns.md`), the 2025 Go
Developer Survey, and the Winder AI migration case study, the top AI mistakes in Go are:

| Mistake | Frequency | Detection |
|---|---|---|
| **Unchecked errors** (ignoring returned `error`) | 2x rate vs humans | `errcheck` linter |
| **Hallucinated packages** (non-existent imports) | 5-21% of suggestions | `go build` (fails immediately) |
| **Wrong context.Context usage** | Common | `noctx`, `contextcheck` linters |
| **Deprecated packages** (e.g., `io/ioutil`) | Common (training data bias) | `staticcheck` SA1019 |
| **Dead code accumulation** during refactoring | High (Winder AI case) | `deadcode` tool |
| **Phantom features** (reconstructing deprecated code) | Observed (Winder AI) | Integration tests |
| **Not following project conventions** | 53% report (Go survey) | `.golangci.yml` + CLAUDE.md rules |
| **Architectural drift** (wrong directory structure) | Observed (Winder AI) | `go-arch-lint` |

**JetBrains GoLand Plugin for Claude Code:** JetBrains released `go-modern-guidelines` plugin
to address AI training data bias. AI models tend to generate outdated Go code because older
patterns appear more frequently in training data. The plugin detects Go version from `go.mod`
and automatically applies modern recommendations (e.g., `slices.Contains()` instead of manual
loops for Go 1.21+).

**Source:** [JetBrains GoLand Blog - Write Modern Go Code with Junie and Claude Code](https://blog.jetbrains.com/go/2026/02/20/write-modern-go-code-with-junie-and-claude-code/)

### 2.5 Go's Natural Defense Against AI Mistakes

Go's toolchain provides built-in defenses that Python lacks:

| Defense | What It Catches | Python Equivalent |
|---|---|---|
| `go build` | Hallucinated imports, wrong signatures, unused imports, type errors | mypy (opt-in, partial) |
| `go vet` | Suspicious constructs, printf format mismatches | ruff (subset) |
| `gofmt`/`gofumpt` | All formatting issues | ruff format (opt-in) |
| Module proxy | Package existence verification at install time | None (pip installs anything) |
| `internal/` | Unauthorized cross-boundary imports | import-linter (third-party) |

**This is a key advantage for AI-assisted Go development:** Many AI mistakes that would silently
pass in Python are caught at compile time in Go.

---

## 3. Claude Code Multi-Agent Go Development

### 3.1 Current State of Multi-Agent Development (2026)

Claude Code has evolved into a multi-agent orchestration platform with three approaches:

1. **Official Subagents:** Specialized AI assistants with custom system prompts and independent
   context windows. Each handles specific task types.
2. **Agent Teams (Experimental):** One session acts as team lead, assigns tasks to teammates
   working in isolated branches.
3. **Third-party frameworks:** Ruflo, Agentrooms, etc.

**Source:** [eesel.ai - Claude Code Multiple Agent Systems Guide 2026](https://www.eesel.ai/blog/claude-code-multiple-agent-systems-complete-2026-guide)

### 3.2 Go-Specific Multi-Agent Usage

**GoLand Plugin (`go-modern-guidelines`):** JetBrains created guidelines specifically for AI
agents writing Go, acknowledging that AI-assisted Go development requires version-aware
guardrails.

**Source:** [JetBrains GoLand Blog](https://blog.jetbrains.com/go/2026/02/20/write-modern-go-code-with-junie-and-claude-code/)

### 3.3 Production Experiences with Agentic Go Development

**iximiuz.com case study** (grounded production experience):

**What worked:**

- High-velocity implementation of well-defined tasks (10x-100x productivity gains)
- Effective bug investigation with reproducible test cases
- Component-level work broken into specific, sequential tasks

**What failed:**

- Vague high-level product prompts produced unusable results
- Complex architectural decisions defeated agents until developers specified the approach
- Domain-specific edge cases (e.g., Go S3-compatible API client with HMAC-SHA256 signing and
  GCS header customization) failed after hours of agent work
- **Copy-paste over reusability:** Agents replicate patterns without consolidating duplicated code
- **Success-oriented shortcuts:** AI skipped tests, removed features, and ignored requirements
  to declare tasks complete
- Processing ~2,000 lines addressing XSS vulnerabilities, schema design issues, and logical
  gaps that a fully autonomous system missed

**Practical advice:**

- Max out at 3-4 specialized agents; more than that decreases productivity
- Decompose large features into agent-sized chunks
- Maintain domain expertise -- 15+ years of experience was essential for recognizing
  architectural pitfalls and API incompatibilities
- "Agents are force multipliers for developers who know what they want to build and how, not
  replacements for judgment, architecture, or domain knowledge"

**Source:** [iximiuz.com - A Grounded Take on Agentic Coding](https://iximiuz.com/en/posts/grounded-take-on-agentic-coding/)

### 3.4 Industry Trends

- Claude Code's plugin system has 9,000+ plugins as of February 2026
- Multi-agent system inquiries surged 1,445% from Q1 2024 to Q2 2025
- Zapier: 89% AI adoption, 800+ agents deployed internally
- By end of 2026, 40% of enterprise apps expected to include task-specific AI agents

**Source:** [Anthropic - 2026 Agentic Coding Trends Report](https://resources.anthropic.com/hubfs/2026%20Agentic%20Coding%20Trends%20Report.pdf)

### 3.5 Relevance to alty

alty's agent persona model (researcher, developer, qa-engineer, tech-lead, project-manager)
maps directly to Claude Code's subagent architecture. For a Go migration:

- Use **researcher** subagent for spike research
- Use **developer** subagent for TDD implementation (bounded to single package)
- Use **qa-engineer** subagent for test generation
- Use **tech-lead** subagent for architecture reviews and `go-arch-lint` validation
- Limit to 3-4 active agents maximum per the production experience data

---

## 4. Watermill GoChannel Production Usage

### 4.1 GoChannel Design and Intended Use

GoChannel is Watermill's simplest Pub/Sub implementation, based on Go channels for in-process
message delivery. It is explicitly designed for development, testing, and single-process
applications.

### 4.2 Documented Limitations

| Limitation | Impact | Mitigation |
|---|---|---|
| **No global state** | Must use same instance for publish and subscribe | Pass instance via DI |
| **Not persistent** (default) | Messages lost if no subscriber present | Enable `persistent: true` or accept loss |
| **No message ordering** (persistent mode) | Event handlers may process out of order | Use causality markers or sequence numbers |
| **Memory-only persistence** | Large volumes cause OOM | Not suitable for high-volume scenarios |
| **No consumer groups** | Every consumer gets every message | Not suitable for load balancing |
| **Non-blocking publish** | Messages sent in background; no delivery guarantee | Acceptable for CLI |
| **Context preservation** | Context travels in-process (differs from network-based Pub/Subs) | Will break if switching to network transport later |

**Source:** [Watermill Docs - GoChannel](https://watermill.io/pubsubs/gochannel/)

### 4.3 Wild Workouts Reference Project

The [Wild Workouts Go DDD example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
by ThreeDotsLabs uses Watermill with CQRS components. While it demonstrates the Watermill API
with GoChannel for development, the production deployment uses Google Cloud Pub/Sub. This
confirms that GoChannel is intended as a development/testing backend with production
deployments using a real message broker.

**Source:** [GitHub - ThreeDotsLabs/wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)

### 4.4 SQLite Pub/Sub (Watermill 1.5)

As of Watermill 1.5 (September 2024), there is now a **SQLite Pub/Sub** backend. This provides
an interesting middle ground: persistence + single-binary deployment without an external
message broker. For CLI tools that need event durability, SQLite may be preferable to
GoChannel.

**Source:** [ThreeDotsLabs Blog - Watermill 1.5 Released](https://threedots.tech/post/watermill-1-5/)

### 4.5 Assessment for alty CLI

GoChannel is **suitable** for alty's CLI use case because:

- alty CLI is a single-process, short-lived application
- Events are used for intra-process communication (not cross-service)
- Message loss is acceptable (events drive side-effects, not core logic)
- The CQRS abstraction layer means swapping to NATS/SQLite requires zero app code changes

GoChannel is **not suitable** if:

- alty MCP server needs durable event delivery
- Events must survive process crashes
- Consumer groups are needed for parallel processing

**Recommendation:** Start with GoChannel for CLI. Use SQLite Pub/Sub for MCP server (embedded,
persistent, single-binary). Scale to embedded NATS only if concurrent MCP sessions require it.

---

## 5. severity1/claude-agent-sdk-go Production Viability

### 5.1 Current State (March 2026)

| Metric | Value |
|---|---|
| Stars | 101 |
| Total commits | 83 |
| Open issues | 5 |
| Open PRs | 6 |
| License | MIT |
| Latest version | v0.6.12 |
| Last notable release | January 24, 2026 |
| Contributors | Single primary maintainer |

**Source:** [GitHub - severity1/claude-agent-sdk-go](https://github.com/severity1/claude-agent-sdk-go)

### 5.2 Community Ecosystem

Multiple forks exist, suggesting community interest but also fragmentation:

- `schlunsen/claude-agent-sdk-go` -- port of official Python SDK
- `connerohnesorge/claude-agent-sdk-go` -- separate implementation
- `M1n9X/claude-agent-sdk-go` -- claims "complete feature parity with Python SDK, all 204
  features including all 12 hook events and complete MCP server support"
- `clsx524/claude-agent-sdk-go` -- another fork
- `dotcommander/agent-sdk-go` -- yet another implementation

**Source:** [GitHub Topics - claude-agent-sdk](https://github.com/topics/claude-agent-sdk)

### 5.3 Known Issues and Pitfalls

**Architecture-level problems** (apply to ALL Claude Agent SDK implementations, not just Go):

1. **CLAUDECODE=1 environment variable inheritance:** Subprocess inherits this env var,
   preventing SDK usage from within Claude Code hooks/plugins/subagents. This is a Python
   SDK issue (anthropics/claude-agent-sdk-python#573) that affects Go wrappers too since they
   spawn the same CLI.

2. **Permission system complexity:** `permissionMode`, `settings.json`, hooks, and other
   controls have overlapping and unclear interactions.

3. **Node.js runtime dependency:** The SDK spawns Claude Code CLI, which requires Node.js.
   This defeats Go's single-binary advantage.

4. **API configuration conflicts:** Custom environment variables can be overridden by user-level
   settings in `~/.claude/settings.json`.

5. **SDK is a CLI wrapper, not a library:** "The SDK functions as a wrapper around the CLI
   rather than a standalone library, creating tight coupling that limits functionality and
   developer experience."

**Source:** [liruifengv - Common Pitfalls with Claude Agent SDK](https://liruifengv.com/posts/claude-agent-sdk-pitfalls-en/)
**Source:** [GitHub Issue #573](https://github.com/anthropics/claude-agent-sdk-python/issues/573)

### 5.4 Community Demand for Official Go SDK

GitHub Issue #498 on anthropics/claude-agent-sdk-python requests official Go SDK support:

- **41+ positive reactions** (20 thumbs up, 18 hearts, 3 rockets)
- **Zero official Anthropic response** as of March 2026
- Proposed use cases: high-performance agent services, Go microservices integration, CLI tools,
  cloud-native applications
- Current workarounds described as "problematic": manual HTTP API calls (complex, error-prone),
  spawning separate Node.js/Python processes (deployment overhead), third-party libraries
  (security/maintenance concerns)

**Source:** [GitHub Issue #498 - Go SDK Support Request](https://github.com/anthropics/claude-agent-sdk-python/issues/498)

### 5.5 Assessment for alty

**severity1/claude-agent-sdk-go is NOT recommended for production use** because:

1. **Single maintainer risk:** One person maintains the project; no Anthropic backing
2. **Version lag:** Tracks Python v0.1.22 while current is v0.1.47+ (25+ releases behind)
3. **Stale signals:** Last release January 24, 2026 (10+ weeks ago)
4. **Node.js dependency:** Defeats Go's single-binary advantage
5. **Fragmented ecosystem:** 5+ competing forks, no community consensus
6. **Fundamental architecture flaw:** CLI subprocess wrapper introduces 1-3s latency per query

**Alternative strategy for alty Go port:**

- Use `anthropic-sdk-go` (MIT, official, v1.26.0+) for direct API calls -- this covers the
  port adapter pattern (LLM interaction via Anthropic Messages API)
- Build agent orchestration natively in Go using alty's own domain model + Watermill CQRS
- The Claude Agent SDK's value (agent loop, built-in tools) can be replicated since alty
  already defines its own agent personas and tool access patterns

---

## 6. Synthesis: Migration Risk Matrix for alty

### 6.1 Risk Assessment

| Risk | Severity | Likelihood | Mitigation |
|---|---|---|---|
| Code expansion (2-3x LOC) | Medium | High | Accept; Go verbosity is offset by compiler safety |
| Loss of Python AI/ML ecosystem | Low | N/A | alty does not use AI/ML libraries; LLM calls use HTTP API |
| AI agent generates poor Go | High | High | Go compiler catches most issues; golangci-lint covers rest |
| Package hallucination | Medium | Medium | `go build` fails immediately; Go module proxy validates |
| Multi-agent coordination overhead | Medium | Medium | Limit to 3-4 agents; decompose tasks to package level |
| No official Agent SDK for Go | High | Certain | Use anthropic-sdk-go for API; build agent loop natively |
| Watermill GoChannel limitations | Low | Low | Acceptable for CLI; SQLite or NATS for MCP server |
| Big-bang rewrite failure | Critical | Variable | Use incremental approach per case study evidence |
| Dead code accumulation (AI-assisted) | Medium | High | Add dead-code linting to CI; clean before migrating |

### 6.2 Recommended Approach

Based on the five case studies:

1. **Use incremental (Strangler Fig) approach** -- migrate one bounded context at a time
2. **Validate against Python** -- run both implementations side-by-side during migration
   (Uber pattern)
3. **Go compiler is your safety net** -- most AI mistakes caught at compile time
4. **CLAUDE.md + go-modern-guidelines** -- essential context for AI agents writing Go
5. **Integration tests after each phase** -- the Winder AI case study's biggest lesson
6. **Dead code cleanup gates** -- explicit workflow step, not afterthought

---

## 7. Follow-Up Investigation Needed

| Topic | Why | Suggested Ticket Type |
|---|---|---|
| SQLite Pub/Sub for Watermill MCP server | Better than GoChannel for durable events | Spike |
| go-modern-guidelines plugin integration | Ensure alty generates correct CLAUDE.md for Go projects | Task |
| Strangler Fig migration plan for alty | Bounded context migration order and validation strategy | Spike |
| Dead code detection workflow for AI-assisted Go | Automated gate in CI/CD pipeline | Task |
| anthropic-sdk-go agent loop design | Build alty's agent orchestration natively | Spike |
