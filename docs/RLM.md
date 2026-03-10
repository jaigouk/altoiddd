# RLM (Recursive Language Model) Reference Guide

> A practical guide for developers who want to understand and implement RLM patterns for complex reasoning tasks.

**What you'll learn:**

- Why traditional approaches fail on complex, multi-step questions
- How RLM treats context as a variable to explore programmatically
- The REPL loop pattern that powers iterative reasoning
- Security considerations for executing LLM-generated code
- Patterns for building your own RLM-powered applications

**Prerequisites:**

- Basic Go knowledge
- Familiarity with LLM APIs (OpenAI, Anthropic)
- No prior RLM experience required

---

## Table of Contents

1. [The Problem: Why Traditional RAG Falls Short](#1-the-problem-why-traditional-rag-falls-short)
2. [The Solution: Context as a Variable](#2-the-solution-context-as-a-variable)
3. [Core Concepts](#3-core-concepts)
4. [Architecture Deep Dive](#4-architecture-deep-dive)
5. [Security: Sandboxing Code Execution](#5-security-sandboxing-code-execution)
6. [Building Your Own RLM Application](#6-building-your-own-rlm-application)
7. [Implementation Patterns](#7-implementation-patterns)
8. [Common Pitfalls and Solutions](#8-common-pitfalls-and-solutions)
9. [References](#9-references)

---

## 1. The Problem: Why Traditional RAG Falls Short

### 1.1 The Context Window Problem

Consider any question that requires reasoning across multiple knowledge sources:

- "Can I run Model X on my hardware?" (requires: model specs, hardware limits, optimization options)
- "Is this contract compliant with regulation Y?" (requires: contract text, regulation details, cross-references)
- "Why is my application slow?" (requires: logs, metrics, documentation, config files)

**Option A: Stuff everything into the prompt**

```go
response, err := llm.Complete(ctx, fmt.Sprintf(`
Here's all the documentation:
%s  // 50KB
%s  // 30KB
%s  // 20KB

Question: %s
`, doc1, doc2, doc3, userQuestion))
```

Problems:

- Context windows have limits (even 128K tokens isn't infinite)
- Cost scales linearly with input size
- The LLM may "lose focus" in long contexts
- You're paying for everything even if only parts are relevant

**Option B: Traditional RAG (Retrieval-Augmented Generation)**

```go
// 1. Chunk the documentation
chunks := splitIntoChunks(allDocs, 500)

// 2. Embed and store
for _, chunk := range chunks {
    vectorDB.Add(embed(chunk))
}

// 3. Retrieve relevant chunks
relevant := vectorDB.Search(userQuestion, 5)

// 4. Generate answer from fragments
response, _ := llm.Complete(ctx, fmt.Sprintf("Based on these excerpts: %v\n\nAnswer: ...", relevant))
```

This is better, but has fundamental limitations.

### 1.2 The Dependency Chain Problem

Complex questions have **dependency chains** — answering one part reveals what you need to look up next:

```
"Can I run Model X on 24GB VRAM?"
    │
    ├─→ What's the model architecture? ─→ It's MoE (Mixture of Experts)
    │                                        │
    │                                        └─→ Can I offload parts to CPU?
    │                                               │
    │                                               └─→ Check driver documentation
    │
    ├─→ What sizes are available?
    │       │
    │       └─→ List available files → Smallest is 26GB
    │                                    │
    │                                    └─→ With offloading? Maybe fits
    │
    └─→ What's my actual available memory?
            │
            └─→ Check system status → 22GB free
```

Traditional RAG retrieves fragments in isolation. It might get the model specs but miss:

- The architecture detail that enables optimization
- The driver flag that makes it possible
- The calculation formula

**The fundamental issue:** RAG treats knowledge as passive data to be searched. But complex analysis requires _active exploration_ — following leads, checking cross-references, and building cumulative understanding.

### 1.3 The Needle in a Haystack

Technical documentation has "needles" — critical details buried in dense text:

- A configuration flag that enables a key optimization
- An edge case mentioned in a footnote
- A version-specific feature or limitation

RAG's similarity search might not surface these if the query doesn't match the embedding well enough.

---

## 2. The Solution: Context as a Variable

### 2.1 The Key Insight

> Instead of reading information linearly, we treat it as a **variable in a runtime environment**. The LLM acts as a reasoning engine that writes code to explore that variable.

This is the core insight from the RLM paper (Zhang et al., 2025).

**Traditional approach:**

```go
// Context goes INTO the prompt
llm.Complete(ctx, prompt + context)
```

**RLM approach:**

```go
// Context is a VARIABLE the LLM can explore via code
context := map[string]any{"query": userQuestion, "data": initialData}

// LLM writes code to explore it
rlm.Complete("Answer the user's query")
// -> LLM generates: result := searchDatabase(query)
// -> LLM generates: details := fetchDetails(result["id"])
// -> LLM generates: related := findRelated(details)
// -> etc.
```

### 2.2 The Detective Analogy

Think of the LLM as a detective investigating a case:

**Traditional RAG** = Detective receives a folder with 5 random pages from a 500-page case file. "Here's what seemed relevant. Good luck!"

**RLM** = Detective has access to tools and databases. They can:

- Query databases and APIs
- Check system status
- Fetch documentation
- Cross-reference findings
- Ask a colleague (sub-query) to analyze a specific piece

The detective decides what to investigate next based on what they've found so far.

### 2.3 What Makes RLM Different

| Aspect           | Traditional RAG             | RLM                            |
| ---------------- | --------------------------- | ------------------------------ |
| Context handling | Fragments retrieved upfront | Explored on-demand via code    |
| Navigation       | Static (similarity search)  | Dynamic (code-driven)          |
| Dependencies     | Often missed                | Can follow chains              |
| Iteration        | Single retrieval step       | Multiple exploration steps     |
| LLM role         | Answer generator            | Reasoning engine + code writer |

---

## 3. Core Concepts

### 3.1 The REPL Loop

REPL stands for **Read-Eval-Print Loop** — the interactive programming pattern where you:

1. **Read** input (in RLM: the LLM's generated code)
2. **Eval**uate (execute) that code
3. **Print** the result
4. **Loop** back for more input

In RLM, the loop looks like this:

````
┌─────────────────────────────────────────────────────────────┐
│                        RLM REPL Loop                        │
└─────────────────────────────────────────────────────────────┘

     ┌──────────────┐
     │  User Query  │
     └──────┬───────┘
            │
            ▼
┌───────────────────────┐
│   LLM generates code  │◄─────────────────────────┐
│   in ```repl``` block │                          │
└───────────┬───────────┘                          │
            │                                      │
            ▼                                      │
┌───────────────────────┐                          │
│   Execute code in     │                          │
│   sandboxed REPL      │                          │
└───────────┬───────────┘                          │
            │                                      │
            ▼                                      │
┌───────────────────────┐                          │
│   Capture stdout/     │──────────────────────────┤
│   stderr output       │                          │
└───────────┬───────────┘                          │
            │                                      │
            ▼                                      │
┌───────────────────────┐                          │
│   Feed output back    │                          │
│   to LLM as context   │──────────────────────────┘
└───────────┬───────────┘
            │
            ▼ (when LLM outputs FINAL(...))
┌───────────────────────┐
│   Return structured   │
│   answer to user      │
└───────────────────────┘
````

### 3.2 Code Blocks

The LLM writes code in specially marked blocks:

````markdown
I'll start by examining the data.

```repl
// Check what we're working with
console.log("Query:", context.query);
let result = searchDatabase(context.query);
console.log("Found", result.length, "matches");
console.log(result.slice(0, 3));  // Preview first 3
```
````

The system:

1. Extracts code between ` ```repl ` and ` ``` `
2. Executes it in a sandboxed environment
3. Captures `console.log()` output
4. Sends output back to the LLM

### 3.3 The FINAL Answer Pattern

When the LLM has gathered enough information, it signals completion:

```go
// LLM outputs this text (not Go code - just structured output)
FINAL({
    "answer": "Yes, this is possible with configuration X",
    "recommendation": "Use option A with setting B for best results",
    "reasoning_steps": [
        "First, I checked the requirements",
        "Then, I verified the constraints",
        "Finally, I found a compatible configuration"
    ],
    "sources": ["search_database", "fetch_config", "check_status"]
})
```

The `FINAL(...)` marker tells the system to stop the loop and parse the answer.

### 3.4 Recursive Sub-Queries (llm_query)

Sometimes the LLM needs focused analysis on a specific piece:

````markdown
```repl
// This section is complex. Let me analyze it separately.
detailed_analysis := llm_query(
    "What are the key constraints in this configuration?",
    config_text,
)
fmt.Println(detailed_analysis)
```
````

The `llm_query()` function spawns a sub-RLM call with a focused prompt. This is the "recursive" in Recursive Language Model — the system can call itself to analyze sub-problems.

**Why use sub-queries?**
- Keeps the main conversation focused
- Allows deeper analysis without polluting the main context
- Can use different parameters (e.g., a smaller/cheaper model for simple checks)

### 3.5 Available Tools

Beyond raw code execution, RLM environments expose domain-specific tool functions:

```go
// Example tools in the REPL namespace
searchDatabase(query)       // Search your data
fetchDetails(id)            // Get detailed information
checkStatus()               // Check system/resource state
fetchDocumentation(topic)   // Get relevant docs
llmQuery(prompt, context)   // Recursive sub-call
```

**Key principle:** Only expose **read-only tools**. The RLM provides recommendations but never executes mutations (modifying data, starting processes, etc.).

---

## 4. Architecture Deep Dive

### 4.1 System Prompt Anatomy

The system prompt teaches the LLM how to use the REPL:

````go
const systemPrompt = `
You are an assistant operating inside a REPL environment.

## Available Functions

Call these directly in '''repl''' code blocks:

| Function               | Description                          |
|------------------------|--------------------------------------|
| search(query)          | Search the knowledge base            |
| fetch(id)              | Get detailed information by ID       |
| status()               | Check current system state           |
| docs(topic)            | Fetch documentation on a topic       |

You also have json and math functions available.

## Rules

1. **Explore first.** Always call at least one function before answering.
2. **Cite sources.** Reference which function call provided each fact.
3. **No imports.** Do not use import, exec, eval, or open.
4. **No mutations.** You only provide recommendations; the user decides.
5. **Be concise.** Keep code blocks short and focused.
6. **Use fmt.Println().** Print intermediate results so you can reason over them.

## Workflow

1. Examine the context variable (your query + any pre-loaded data).
2. Call tool functions to gather data.
3. Reason over results step-by-step.
4. When ready, provide your final answer using:

FINAL({
  "answer": "...",
  "reasoning_steps": ["step 1", "step 2"],
  "sources": ["source 1", "source 2"]
})
`
````

**Key elements:**

- **Available functions table** — what the LLM can call
- **Rules** — constraints that guide safe behavior
- **Workflow** — the expected exploration pattern
- **FINAL format** — how to signal completion

### 4.2 The Orchestrator

The orchestrator manages the RLM lifecycle:

```go
type RLMOrchestrator struct {
    tools []ToolWrapper
    llm   LLMClient
}

func (o *RLMOrchestrator) Consult(ctx context.Context, query string, maxTurns int) (Result, error) {
    // 1. Create sandboxed environment
    env := NewSafeREPLEnv(o.tools)

    // 2. Initialize message history
    messages := []Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: fmt.Sprintf("Answer: %s", query)},
    }

    // 3. Run the REPL loop
    for turn := 0; turn < maxTurns; turn++ {
        response, err := o.llm.Complete(ctx, messages)
        if err != nil {
            return Result{}, fmt.Errorf("llm completion: %w", err)
        }

        // Check for final answer
        if strings.Contains(response, "FINAL(") {
            return parseFinalAnswer(response)
        }

        // Extract and execute code blocks
        codeBlocks := findCodeBlocks(response)
        for _, code := range codeBlocks {
            output := env.ExecuteCode(code)
            messages = append(messages, Message{
                Role:    "user",
                Content: fmt.Sprintf("Output: %s", output),
            })
        }
    }

    // 4. Force final answer if max turns reached
    return forceFinalAnswer(messages)
}
```

### 4.3 The Environment

The environment provides:

1. **Namespace isolation** — code runs in a restricted scope
2. **Tool injection** — whitelisted functions are available
3. **Safe builtins** — only safe functions allowed
4. **Output capture** — stdout/stderr are captured and returned

```go
type SafeREPLEnv struct {
    namespace map[string]any
    tools     map[string]ToolFunc
}

var whitelistedTools = map[string]bool{
    "search": true,
    "fetch":  true,
    "status": true,
    "docs":   true,
}

var blockedTools = map[string]bool{
    "delete":  true, // Mutating!
    "update":  true, // Mutating!
    "execute": true, // Mutating!
}

func (e *SafeREPLEnv) ExecuteCode(code string) REPLResult {
    // 1. Validate code via AST analysis
    if err := validateCode(code); err != nil {
        return REPLResult{Stderr: err.Error()}
    }

    // 2. Execute in restricted namespace (using embedded JS runtime like goja)
    output, err := e.runtime.RunString(code)
    if err != nil {
        return REPLResult{Stderr: err.Error()}
    }

    // 3. Capture and return output
    return REPLResult{Stdout: output}
}
```

### 4.4 Message Flow Diagram

````
┌──────────────────────────────────────────────────────────────────────────┐
│                           Message History                                 │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  [System] You are an assistant operating inside a REPL...                │
│                                                                          │
│  [User] Answer: What's the best configuration for my use case?           │
│                                                                          │
│  [Assistant] I'll start by understanding the requirements.               │
│              ```repl                                                     │
│              let info = fetch("requirements");                           │
│              console.log(JSON.stringify(info, null, 2));                 │
│              ```                                                         │
│                                                                          │
│  [User] Code executed:                                                   │
│         REPL output:                                                     │
│         {                                                                │
│           "min_memory": 16000,                                           │
│           "features": ["A", "B", "C"]                                    │
│         }                                                                │
│                                                                          │
│  [Assistant] I see the requirements. Let me check available options.     │
│              ```repl                                                     │
│              let options = search("compatible configurations");          │
│              options.slice(0, 5).forEach(opt => {                        │
│                  console.log(opt.name + ": " + opt.memory + "MB");       │
│              });                                                         │
│              ```                                                         │
│                                                                          │
│  [User] Code executed:                                                   │
│         REPL output:                                                     │
│         Option A: 12000MB                                                │
│         Option B: 18000MB                                                │
│         ...                                                              │
│                                                                          │
│  ... (continues until FINAL)                                             │
│                                                                          │
│  [Assistant] FINAL({                                                     │
│      "answer": "Option B is the best fit",                               │
│      "reasoning_steps": [...],                                           │
│      "sources": ["fetch('requirements')", "search(...)"]                 │
│  })                                                                      │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
````

---

## 5. Security: Sandboxing Code Execution

### 5.1 The Risk

RLM executes LLM-generated code. This is inherently risky:

- **Arbitrary code execution** — LLM could generate malicious code
- **Resource exhaustion** — infinite loops, memory bombs
- **Data exfiltration** — accessing files, network requests
- **Sandbox escape** — exploiting runtime internals

### 5.2 Defense in Depth

Use multiple layers of protection:

#### Layer 1: Safe Builtins

Only expose a subset of safe functions. In Go, this is typically done via an embedded scripting runtime (like goja for JavaScript or Yaegi for Go):

```go
// SafeBuiltins defines functions available in the sandbox
var safeBuiltins = map[string]any{
    // Allowed - output
    "print":   sandboxPrint,
    "println": sandboxPrintln,

    // Allowed - type conversions
    "parseInt":   parseInt,
    "parseFloat": parseFloat,
    "toString":   toString,

    // Allowed - collections
    "len":    safeLen,
    "keys":   safeKeys,
    "values": safeValues,

    // Allowed - math
    "min":   safeMin,
    "max":   safeMax,
    "abs":   safeAbs,
    "round": safeRound,

    // Allowed - iteration helpers
    "range":   safeRange,
    "forEach": safeForEach,
    "map":     safeMap,
    "filter":  safeFilter,
}

// blockedFunctions are explicitly denied
var blockedFunctions = map[string]bool{
    "require": true, // No module loading
    "import":  true, // No dynamic imports
    "eval":    true, // No dynamic eval
    "exec":    true, // No command execution
    "open":    true, // No file access
    "fetch":   true, // No network access (unless explicitly provided as tool)
}
```

**Subtle security decisions (important lessons):**

```go
// In Go with embedded JS (goja), block prototype access:
// obj.__proto__ or Object.getPrototypeOf() can escape sandbox
// Block these via runtime.SetFieldNameMapper or property interceptors

// For Go interpreters (Yaegi), restrict package imports:
// Only allow specific packages, never "os", "syscall", "unsafe"
```

#### Layer 2: AST Validation

Before execution, analyze the code's Abstract Syntax Tree. For embedded JS in Go, use a parser like esprima or the goja runtime's built-in capabilities:

```go
import "github.com/dop251/goja/parser"

func validateCode(code string) error {
    // Parse the code into an AST
    program, err := parser.ParseFile(nil, "", code, 0)
    if err != nil {
        return fmt.Errorf("parse error: %w", err)
    }

    // Walk the AST looking for dangerous patterns
    for _, stmt := range program.Body {
        if err := validateStatement(stmt); err != nil {
            return err
        }
    }
    return nil
}

func validateStatement(stmt ast.Statement) error {
    switch s := stmt.(type) {
    case *ast.CallExpression:
        // Block dangerous function calls
        if ident, ok := s.Callee.(*ast.Identifier); ok {
            if blockedFunctions[ident.Name] {
                return fmt.Errorf("blocked function: %s", ident.Name)
            }
        }
    case *ast.MemberExpression:
        // Block dunder/prototype access
        if prop, ok := s.Property.(*ast.Identifier); ok {
            if strings.HasPrefix(prop.Name, "_") || prop.Name == "__proto__" {
                return fmt.Errorf("private access blocked: %s", prop.Name)
            }
        }
    }
    return nil
}
```

#### Layer 3: Function Whitelisting

Only specific tool functions are injected into the namespace:

```go
var whitelistedTools = map[string]bool{
    "search": true, // Read-only
    "fetch":  true, // Read-only
    "status": true, // Read-only
}

var blockedTools = map[string]bool{
    "delete":  true, // Mutating!
    "update":  true, // Mutating!
    "execute": true, // Mutating!
}

func (e *SafeREPLEnv) injectTools(runtime *goja.Runtime) {
    for name, fn := range e.tools {
        if whitelistedTools[name] {
            runtime.Set(name, fn)
        }
    }
    // Blocked tools are simply not injected - calling them returns undefined
}
```

If the LLM tries to call a blocked function, it gets an "undefined" error because the function simply doesn't exist in the namespace.

#### Layer 4: Execution Timeout

```go
const codeTimeoutSeconds = 5

func (e *SafeREPLEnv) ExecuteCode(code string) REPLResult {
    ctx, cancel := context.WithTimeout(context.Background(), codeTimeoutSeconds*time.Second)
    defer cancel()

    resultCh := make(chan REPLResult, 1)

    go func() {
        // Execute in goroutine
        output, err := e.runtime.RunString(code)
        if err != nil {
            resultCh <- REPLResult{Stderr: err.Error()}
            return
        }
        resultCh <- REPLResult{Stdout: output}
    }()

    select {
    case result := <-resultCh:
        return result
    case <-ctx.Done():
        // For goja, use runtime.Interrupt() to stop execution
        e.runtime.Interrupt("execution timeout")
        return REPLResult{Stderr: "TimeoutError: execution exceeded 5s limit"}
    }
}
```

Note: Go's goroutines can be interrupted via runtime-specific mechanisms (e.g., goja.Runtime.Interrupt()). For true isolation, consider OS-level sandboxing.

#### Layer 5: Output Truncation

Long outputs can overwhelm the context window:

```go
const maxOutputChars = 10000

func truncateOutput(output string) string {
    if len(output) > maxOutputChars {
        return output[:maxOutputChars] + "\n... (truncated)"
    }
    return output
}
```

### 5.3 Advanced: OS-Level Sandboxing

For production systems, consider [sandbox-runtime (srt)](https://github.com/anthropic-experimental/sandbox-runtime):

| Platform | Backend      | Description              |
| -------- | ------------ | ------------------------ |
| Linux    | bubblewrap   | User namespace isolation |
| macOS    | sandbox-exec | Seatbelt profiles        |

```bash
# Run process in sandbox with no network, limited filesystem
srt --network=deny --write=/tmp ./myapp
```

This provides true isolation at the OS level, preventing escapes that language-level sandboxing might miss.

---

## 6. Building Your Own RLM Application

### 6.1 Step 1: Identify Your Use Case

RLM works best for questions that require:

- **Multi-step reasoning** — Answer depends on intermediate findings
- **Dynamic exploration** — What to look up next depends on what you find
- **Cross-referencing** — Information from multiple sources must be combined
- **Live data** — Current state matters (not just static documents)

**Good RLM use cases:**

- Technical support: "Can X work with Y given Z constraints?"
- Compliance checking: "Does this document meet requirements A, B, C?"
- Debugging: "Why is this system behaving unexpectedly?"
- Configuration: "What's the optimal setup for my situation?"

**Poor RLM use cases:**

- Simple lookups: "What is the capital of France?"
- Single-source answers: Questions answered by one document
- No dependency chains: Answer doesn't require exploration

### 6.2 Step 2: Define Your Tools

List the read-only operations your RLM needs:

```go
// Example tool definitions

// Search searches the knowledge base. Returns list of matches.
func (t *Tools) Search(query string) ([]map[string]any, error) {
    // Implementation here
    return nil, nil
}

// Fetch gets detailed information by ID.
func (t *Tools) Fetch(id string) (map[string]any, error) {
    // Implementation here
    return nil, nil
}

// Status checks current system/resource state.
func (t *Tools) Status() (map[string]any, error) {
    // Implementation here
    return nil, nil
}

// Docs fetches documentation on a topic.
func (t *Tools) Docs(topic string) (string, error) {
    // Implementation here
    return "", nil
}
```

**Design principles:**

- **Read-only** — Never modify state
- **Deterministic** — Same input → same output (when possible)
- **Bounded output** — Limit response size to avoid context overflow
- **Error handling** — Return useful errors, don't crash

### 6.3 Step 3: Write Your System Prompt

Customize the template for your domain:

```go
func buildSystemPrompt(tools []ToolDefinition) string {
    functionTable := generateFunctionTable(tools)

    return fmt.Sprintf(`
You are a [DOMAIN] assistant operating inside a REPL environment.
Your job is to [PRIMARY TASK] by exploring available data and tools.

## Available Functions

| Function | Description |
|----------|-------------|
%s

Available utilities: JSON parsing, math functions

## Rules

1. **Explore first.** Always call at least one function before answering.
2. **Cite sources.** Reference which function call provided each fact.
3. **No imports.** Do not use import, exec, eval, or open.
4. **No mutations.** You only provide recommendations; the user decides.
5. **Be concise.** Keep code blocks short and focused.
6. **Use console.log().** Print intermediate results so you can reason over them.

## Domain-Specific Guidance

[Add guidance specific to your use case]

## Workflow

1. Examine the context variable (query + pre-loaded data).
2. Call tool functions to gather data.
3. Reason over results step-by-step.
4. When ready, provide your final answer as:

FINAL({
  "answer": "...",
  "reasoning_steps": ["step 1", "step 2"],
  "sources": ["source 1", "source 2"]
})
`, functionTable)
}
```

### 6.4 Step 4: Example Exploration Trajectory

Design an example showing how your RLM should reason:

````
Query: "[Your typical user question]"

Turn 1: Understand the request
─────────────────────────────────
```repl
console.log("Query:", context.query);
let initial = search(context.query);
console.log("Found", initial.length, "relevant items");
```

Output: Found 5 relevant items

Turn 2: Gather details
─────────────────────────
```repl
initial.slice(0, 3).forEach(item => {
    let details = fetch(item.id);
    console.log(item.name + ": " + details.key_property);
});
```

Output:
Item A: value_a
Item B: value_b
Item C: value_c

Turn 3: Check constraints
─────────────────────────
```repl
let current = status();
console.log("Available:", current.available);
console.log("Required:", details.requirement);
```

Output:
Available: 100
Required: 80

Turn 4: Final answer
────────────────────
FINAL({
    "answer": "Yes, Item B is the best choice",
    "reasoning_steps": [
        "Found 5 relevant items matching the query",
        "Item B has the best key_property value",
        "Current availability (100) exceeds requirement (80)"
    ],
    "sources": ["search(query)", "fetch(item_b)", "status()"]
})
````

### 6.5 Step 5: Implement and Test

1. **Build the orchestrator** (see Section 7.1)
2. **Implement tool wrappers** (see Section 7.3)
3. **Add security layers** (see Section 5.2)
4. **Test with adversarial inputs** (see Section 8)

---

## 7. Implementation Patterns

### 7.1 Basic RLM Loop

The minimal implementation:

```go
import (
    "context"
    "fmt"
    "regexp"
    "strings"
)

// RunRLM executes the RLM loop.
func RunRLM(
    ctx context.Context,
    query string,
    tools map[string]ToolFunc,
    llmClient LLMClient,
    maxTurns int,
) (map[string]any, error) {
    // Build REPL environment
    env := NewSafeREPLEnv(tools)
    env.Set("context", map[string]any{"query": query})

    messages := []Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: fmt.Sprintf("Query: %s", query)},
    }

    codeBlockRe := regexp.MustCompile("(?s)```repl(.*?)```")

    for turn := 0; turn < maxTurns; turn++ {
        // Get LLM response
        response, err := llmClient.Complete(ctx, messages)
        if err != nil {
            return nil, fmt.Errorf("llm completion: %w", err)
        }

        // Check for final answer
        if strings.Contains(response, "FINAL(") {
            return parseFinalAnswer(response)
        }

        // Find and execute code blocks
        matches := codeBlockRe.FindAllStringSubmatch(response, -1)

        if len(matches) > 0 {
            for _, match := range matches {
                code := strings.TrimSpace(match[1])
                output := env.ExecuteCode(code)
                messages = append(messages, Message{Role: "assistant", Content: response})
                messages = append(messages, Message{
                    Role:    "user",
                    Content: fmt.Sprintf("Output:\n%s", output),
                })
            }
        } else {
            // No code - nudge the LLM
            messages = append(messages, Message{Role: "assistant", Content: response})
            messages = append(messages, Message{
                Role:    "user",
                Content: "Use ```repl``` blocks to explore, or provide FINAL(...)",
            })
        }
    }

    // Max turns reached
    return map[string]any{
        "error":   "Max turns reached",
        "partial": messages,
    }, nil
}
```

### 7.2 Structured Result Parsing

Parse FINAL answers robustly:

```go
import (
    "encoding/json"
    "regexp"
)

// parseFinalAnswer parses FINAL(...) from response with fallbacks.
func parseFinalAnswer(response string) (map[string]any, error) {
    // Extract content between FINAL( and )
    finalRe := regexp.MustCompile(`(?s)FINAL\((.*)\)`)
    match := finalRe.FindStringSubmatch(response)
    if match == nil {
        return map[string]any{"answer": response, "raw": true}, nil
    }

    content := strings.TrimSpace(match[1])

    // Try JSON parse
    var result map[string]any
    if err := json.Unmarshal([]byte(content), &result); err == nil {
        return result, nil
    }

    // Try finding JSON object within the content
    jsonRe := regexp.MustCompile(`(?s)\{.*\}`)
    jsonMatch := jsonRe.FindString(content)
    if jsonMatch != "" {
        if err := json.Unmarshal([]byte(jsonMatch), &result); err == nil {
            return result, nil
        }
    }

    // Fallback: treat as plain text
    return map[string]any{"answer": content, "raw": true}, nil
}
```

### 7.3 Tool Wrapper Pattern

Wrap your functions for safe injection:

```go
// ToolWrappers creates tool functions with consistent interface.
type ToolWrappers struct {
    backend Backend
}

// CreateToolWrappers builds the tool function map.
func (w *ToolWrappers) CreateToolWrappers() map[string]ToolFunc {
    return map[string]ToolFunc{
        "search": w.search,
        "fetch":  w.fetch,
        "status": w.status,
    }
}

// search searches the knowledge base (read-only).
func (w *ToolWrappers) search(args ...any) (any, error) {
    query, _ := args[0].(string)
    limit := 10
    if len(args) > 1 {
        limit, _ = args[1].(int)
    }

    results, err := w.backend.Search(query, limit)
    if err != nil {
        return nil, err
    }
    // Truncate to prevent context overflow
    if len(results) > limit {
        results = results[:limit]
    }
    return results, nil
}

// fetch gets details by ID (read-only).
func (w *ToolWrappers) fetch(args ...any) (any, error) {
    id, _ := args[0].(string)
    return w.backend.GetByID(id)
}

// status gets current status (read-only).
func (w *ToolWrappers) status(args ...any) (any, error) {
    return w.backend.GetStatus()
}
```

### 7.4 Sub-Query Implementation

Implement recursive `llm_query`:

```go
// LLMQueryFactory creates an llm_query function with depth limiting.
type LLMQueryFactory struct {
    llmClient    LLMClient
    depthLimit   int
    currentDepth int
    mu           sync.Mutex
}

// NewLLMQueryFactory creates a factory with the given depth limit.
func NewLLMQueryFactory(llmClient LLMClient, depthLimit int) *LLMQueryFactory {
    return &LLMQueryFactory{
        llmClient:  llmClient,
        depthLimit: depthLimit,
    }
}

// LLMQuery performs a sub-query with depth tracking.
func (f *LLMQueryFactory) LLMQuery(ctx context.Context, prompt, queryContext string) (string, error) {
    f.mu.Lock()
    if f.currentDepth >= f.depthLimit {
        f.mu.Unlock()
        return "Error: Maximum recursion depth reached", nil
    }
    f.currentDepth++
    f.mu.Unlock()

    defer func() {
        f.mu.Lock()
        f.currentDepth--
        f.mu.Unlock()
    }()

    fullPrompt := fmt.Sprintf("Sub-task: %s", prompt)
    if queryContext != "" {
        // Truncate context to 2000 chars
        if len(queryContext) > 2000 {
            queryContext = queryContext[:2000]
        }
        fullPrompt += fmt.Sprintf("\n\nContext:\n%s", queryContext)
    }

    return f.llmClient.Complete(ctx, []Message{
        {Role: "user", Content: fullPrompt},
    })
}
```

### 7.5 Sandboxed Execution

Execute code safely using an embedded runtime like goja (JavaScript) or similar:

```go
import (
    "context"
    "strings"
    "time"

    "github.com/dop251/goja"
)

// ExecuteSandboxed executes code with output capture and safety checks.
func ExecuteSandboxed(code string, namespace map[string]any, timeout time.Duration) string {
    // 1. AST validation
    if err := validateCode(code); err != nil {
        return fmt.Sprintf("SecurityError: %v", err)
    }

    // 2. Create isolated runtime
    runtime := goja.New()

    // 3. Capture output
    var outputBuf strings.Builder
    runtime.Set("console", map[string]any{
        "log": func(call goja.FunctionCall) goja.Value {
            for _, arg := range call.Arguments {
                outputBuf.WriteString(arg.String())
                outputBuf.WriteString(" ")
            }
            outputBuf.WriteString("\n")
            return goja.Undefined()
        },
    })

    // 4. Inject safe namespace
    for name, value := range namespace {
        runtime.Set(name, value)
    }

    // 5. Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    done := make(chan string, 1)
    go func() {
        _, err := runtime.RunString(code)
        if err != nil {
            done <- fmt.Sprintf("Error: %v", err)
            return
        }
        done <- outputBuf.String()
    }()

    select {
    case output := <-done:
        // 6. Truncate if too long
        if len(output) > 10000 {
            output = output[:10000] + "\n... (truncated)"
        }
        return output
    case <-ctx.Done():
        runtime.Interrupt("timeout")
        return "Error: execution timeout exceeded"
    }
}
```

---

## 8. Common Pitfalls and Solutions

### 8.1 Output Overflow

**Problem:** LLM prints entire datasets, overwhelming context window.

**Solution:**

```go
// Truncate in execution
const maxOutput = 5000

func truncateOutput(output string) string {
    if len(output) > maxOutput {
        return fmt.Sprintf("%s\n... (truncated, %d total chars)", output[:maxOutput], len(output))
    }
    return output
}
```

Also add guidance in the system prompt:

```
- Never print entire results. Use slicing: results[:10]
- Limit loop iterations: for item in items[:10]:
- Use len() to check size before printing
```

### 8.2 Infinite Loops

**Problem:** LLM generates `while(true)` or similar.

**Solution:**

- AST validation to detect unbounded loops
- Execution timeout
- OS-level sandboxing for hard limit

```go
// AST check for while(true) - example using goja parser
func checkForInfiniteLoops(stmt ast.Statement) error {
    switch s := stmt.(type) {
    case *ast.WhileStatement:
        // Check for while(true) pattern
        if lit, ok := s.Test.(*ast.BooleanLiteral); ok && lit.Value {
            return fmt.Errorf("unbounded while loop detected")
        }
    case *ast.ForStatement:
        // Check for for(;;) pattern (no test condition)
        if s.Test == nil {
            return fmt.Errorf("unbounded for loop detected")
        }
    }
    return nil
}
```

### 8.3 LLM Doesn't Write Code

**Problem:** LLM answers directly instead of exploring.

**Solution:** Strengthen system prompt:

````
## CRITICAL RULES
1. You MUST write at least one ```repl``` code block before answering.
2. NEVER answer based on assumptions. Always verify with code.
3. If you're tempted to answer without code, STOP and write code first.
````

And nudge in the loop:

```go
if len(codeBlocks) == 0 {
    messages = append(messages, Message{
        Role:    "user",
        Content: "You must use ```repl``` blocks to explore. Please write code.",
    })
}
```

### 8.4 Stuck in a Loop

**Problem:** LLM keeps exploring without reaching FINAL.

**Solution:**

- Hard limit on turns (`maxTurns`)
- Force final answer when limit reached:

```go
if turn == maxTurns-1 {
    messages = append(messages, Message{
        Role:    "user",
        Content: "IMPORTANT: This is your last turn. Provide your FINAL(...) answer now based on what you've gathered.",
    })
}
```

### 8.5 Import Attempts

**Problem:** LLM tries to `import os` or similar.

**Solution:**

- AST validation blocks all imports
- Import functions not exposed in namespace
- Clear system prompt guidance

```go
// For embedded JS (goja), imports are blocked by not exposing require()
// For Go interpreters, validate AST for import statements:
func checkForImports(stmt ast.Statement) error {
    switch s := stmt.(type) {
    case *ast.ImportDeclaration:
        return fmt.Errorf("imports blocked: %s", s.Source.Value)
    case *ast.CallExpression:
        if ident, ok := s.Callee.(*ast.Identifier); ok {
            if ident.Name == "require" || ident.Name == "import" {
                return fmt.Errorf("dynamic imports blocked: %s", ident.Name)
            }
        }
    }
    return nil
}
```

### 8.6 Sandbox Escape Attempts

**Problem:** LLM tries `obj.__proto__` or `Object.getPrototypeOf()` to access dangerous internals.

**Solution:**

- Block prototype/dunder attribute access in AST
- Restrict property access in the runtime
- Use field name mappers to hide internal properties

```go
// Block private/prototype access in AST validation
func checkForEscapeAttempts(expr ast.Expression) error {
    switch e := expr.(type) {
    case *ast.MemberExpression:
        if prop, ok := e.Property.(*ast.Identifier); ok {
            name := prop.Name
            if strings.HasPrefix(name, "_") || name == "__proto__" || name == "constructor" {
                return fmt.Errorf("private/prototype access blocked: %s", name)
            }
        }
    case *ast.CallExpression:
        // Block Object.getPrototypeOf, Object.setPrototypeOf, etc.
        if member, ok := e.Callee.(*ast.MemberExpression); ok {
            if obj, ok := member.Object.(*ast.Identifier); ok && obj.Name == "Object" {
                if prop, ok := member.Property.(*ast.Identifier); ok {
                    if strings.Contains(prop.Name, "Prototype") {
                        return fmt.Errorf("prototype manipulation blocked: %s", prop.Name)
                    }
                }
            }
        }
    }
    return nil
}
```

---

## 9. References

### Papers and Repositories

- **RLM Paper**: Zhang et al. (2025) - [github.com/alexzhang13/rlm](https://github.com/alexzhang13/rlm)
- **sandbox-runtime**: Anthropic - [github.com/anthropic-experimental/sandbox-runtime](https://github.com/anthropic-experimental/sandbox-runtime)

### Related Concepts

- **ReAct Pattern**: Reasoning + Acting — LLMs that interleave reasoning with tool use
- **Code Interpreters**: Systems like ChatGPT Code Interpreter that execute generated code
- **Agentic AI**: AI systems that take actions in environments

---

## Appendix A: Quick Start Checklist

When implementing RLM for a new project:

- [ ] **Define use case** — What questions require multi-step reasoning?
- [ ] **Design tools** — What read-only operations does the LLM need?
- [ ] **Write system prompt** with:
  - [ ] Available functions table
  - [ ] Domain-specific guidance
  - [ ] Rules (no imports, use print, etc.)
  - [ ] FINAL answer format
- [ ] **Implement safe builtins** (start restrictive, add as needed)
- [ ] **Add AST validation** for imports, dunders, dangerous calls
- [ ] **Create tool wrappers** (read-only only!)
- [ ] **Implement the REPL loop** with:
  - [ ] Code block extraction
  - [ ] Sandboxed execution
  - [ ] Output capture and feedback
  - [ ] FINAL detection
- [ ] **Add safeguards**:
  - [ ] Max turns limit
  - [ ] Output truncation
  - [ ] Execution timeout
- [ ] **Test adversarial inputs** (import attempts, loops, escapes)

---

## Appendix B: System Prompt Template

Copy and customize for your project:

```markdown
You are a [DOMAIN] assistant operating inside a REPL environment.
Your task is to answer questions by exploring available tools and data.

## Available Functions

| Function                  | Description                       |
| ------------------------- | --------------------------------- |
| `context`                 | The query and any pre-loaded data |
| `toolA(param)`            | Description of tool A             |
| `toolB(param)`            | Description of tool B             |
| `llmQuery(prompt, data)`  | Ask a focused sub-question        |

Available utilities: `JSON.stringify()`, `JSON.parse()`, math operations

## Rules

1. **Explore first.** Always write code before answering.
2. **Use console.log().** Results must be logged to be visible.
3. **No require/import.** Do not use `require`, `import`, `eval`, or `Function`.
4. **Be concise.** Keep code blocks short and focused.
5. **Cite sources.** Reference which tool calls provided your data.

## Workflow

1. Examine `context` to understand the query.
2. Call tool functions to gather data.
3. Use `llmQuery()` for detailed sub-analysis if needed.
4. When ready, output your answer as:

FINAL({
"answer": "...",
"reasoning": ["step 1", "step 2"],
"sources": ["toolA()", "toolB()"]
})

Begin exploring now.
```

---

## Appendix C: Complete Minimal Implementation

A complete, copy-paste ready implementation using Go with goja (embedded JavaScript runtime):

````go
// Package rlm provides a minimal RLM (Recursive Language Model) implementation.
// Copy this file to start your own RLM project.
package rlm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// =============================================================================
// TYPES
// =============================================================================

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Result represents the parsed FINAL answer.
type Result struct {
	Answer    string   `json:"answer,omitempty"`
	Reasoning []string `json:"reasoning,omitempty"`
	Sources   []string `json:"sources,omitempty"`
	TurnsUsed int      `json:"turns_used,omitempty"`
	Error     string   `json:"error,omitempty"`
	Raw       bool     `json:"raw,omitempty"`
}

// ToolFunc is a function that can be called from the sandbox.
type ToolFunc func(args ...any) (any, error)

// LLMClient defines the interface for LLM completion.
type LLMClient interface {
	Complete(ctx context.Context, messages []Message) (string, error)
}

// =============================================================================
// SECURITY VALIDATION
// =============================================================================

// SecurityError indicates code failed security validation.
type SecurityError struct {
	Message string
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf("security error: %s", e.Message)
}

// blockedPatterns contains patterns that indicate dangerous code.
var blockedPatterns = []string{
	"__proto__",
	"constructor",
	"prototype",
	"require(",
	"import(",
	"eval(",
	"Function(",
}

// validateCode checks code for security issues.
func validateCode(code string) error {
	for _, pattern := range blockedPatterns {
		if strings.Contains(code, pattern) {
			return &SecurityError{Message: fmt.Sprintf("blocked pattern: %s", pattern)}
		}
	}
	return nil
}

// =============================================================================
// SANDBOXED EXECUTION
// =============================================================================

// SafeREPLEnv provides a sandboxed execution environment.
type SafeREPLEnv struct {
	runtime   *goja.Runtime
	namespace map[string]any
	outputBuf strings.Builder
	mu        sync.Mutex
}

// NewSafeREPLEnv creates a new sandboxed environment.
func NewSafeREPLEnv(tools map[string]ToolFunc) *SafeREPLEnv {
	env := &SafeREPLEnv{
		runtime:   goja.New(),
		namespace: make(map[string]any),
	}

	// Inject console.log for output capture
	env.runtime.Set("console", map[string]any{
		"log": func(call goja.FunctionCall) goja.Value {
			env.mu.Lock()
			defer env.mu.Unlock()
			for i, arg := range call.Arguments {
				if i > 0 {
					env.outputBuf.WriteString(" ")
				}
				env.outputBuf.WriteString(arg.String())
			}
			env.outputBuf.WriteString("\n")
			return goja.Undefined()
		},
	})

	// Inject JSON utilities
	env.runtime.Set("JSON", map[string]any{
		"stringify": func(v any) string {
			b, _ := json.MarshalIndent(v, "", "  ")
			return string(b)
		},
		"parse": func(s string) any {
			var v any
			json.Unmarshal([]byte(s), &v)
			return v
		},
	})

	// Inject tools
	for name, fn := range tools {
		env.runtime.Set(name, fn)
	}

	return env
}

// Set adds a value to the namespace.
func (e *SafeREPLEnv) Set(name string, value any) {
	e.namespace[name] = value
	e.runtime.Set(name, value)
}

// ExecuteCode runs code in the sandbox with timeout.
func (e *SafeREPLEnv) ExecuteCode(code string, timeout time.Duration) string {
	// 1. Validate
	if err := validateCode(code); err != nil {
		return fmt.Sprintf("SecurityError: %v", err)
	}

	// 2. Clear output buffer
	e.mu.Lock()
	e.outputBuf.Reset()
	e.mu.Unlock()

	// 3. Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := e.runtime.RunString(code)
		done <- err
	}()

	select {
	case err := <-done:
		e.mu.Lock()
		output := e.outputBuf.String()
		e.mu.Unlock()

		if err != nil {
			return fmt.Sprintf("%sError: %v", output, err)
		}

		// Truncate if too long
		const maxOutput = 10000
		if len(output) > maxOutput {
			output = output[:maxOutput] + "\n... (truncated)"
		}
		return output

	case <-ctx.Done():
		e.runtime.Interrupt("timeout")
		return "Error: execution timeout exceeded"
	}
}

// =============================================================================
// RESULT PARSING
// =============================================================================

var (
	finalRe = regexp.MustCompile(`(?s)FINAL\((.*)\)`)
	jsonRe  = regexp.MustCompile(`(?s)\{.*\}`)
)

// parseFinalAnswer extracts and parses FINAL(...) from response.
func parseFinalAnswer(response string) Result {
	match := finalRe.FindStringSubmatch(response)
	if match == nil {
		return Result{Answer: response, Raw: true}
	}

	content := strings.TrimSpace(match[1])

	// Try JSON parse
	var result Result
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return result
	}

	// Try extracting JSON object
	jsonMatch := jsonRe.FindString(content)
	if jsonMatch != "" {
		if err := json.Unmarshal([]byte(jsonMatch), &result); err == nil {
			return result
		}
	}

	return Result{Answer: content, Raw: true}
}

// =============================================================================
// MAIN RLM LOOP
// =============================================================================

var codeBlockRe = regexp.MustCompile("(?s)```repl(.*?)```")

// RunRLM executes the RLM loop.
//
// Parameters:
//   - ctx: Context for cancellation
//   - query: User's question
//   - tools: Map of tool_name -> callable (read-only functions)
//   - llmClient: LLM client for completions
//   - systemPrompt: System prompt teaching the LLM how to use the REPL
//   - maxTurns: Maximum iterations before forcing final answer
//
// Returns the parsed FINAL answer or error result.
func RunRLM(
	ctx context.Context,
	query string,
	tools map[string]ToolFunc,
	llmClient LLMClient,
	systemPrompt string,
	maxTurns int,
) Result {
	// Build environment
	env := NewSafeREPLEnv(tools)
	env.Set("context", map[string]any{"query": query})

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Query: %s", query)},
	}

	const execTimeout = 5 * time.Second

	for turn := 0; turn < maxTurns; turn++ {
		// Get LLM response
		response, err := llmClient.Complete(ctx, messages)
		if err != nil {
			return Result{Error: fmt.Sprintf("LLM error: %v", err)}
		}

		// Check for final answer
		if strings.Contains(response, "FINAL(") {
			result := parseFinalAnswer(response)
			result.TurnsUsed = turn + 1
			return result
		}

		// Extract code blocks
		matches := codeBlockRe.FindAllStringSubmatch(response, -1)

		messages = append(messages, Message{Role: "assistant", Content: response})

		if len(matches) > 0 {
			for _, match := range matches {
				code := strings.TrimSpace(match[1])
				output := env.ExecuteCode(code, execTimeout)
				messages = append(messages, Message{
					Role:    "user",
					Content: fmt.Sprintf("Code executed:\n```javascript\n%s\n```\n\nOutput:\n%s", code, output),
				})
			}
		} else {
			// Nudge to write code
			messages = append(messages, Message{
				Role:    "user",
				Content: "Use ```repl``` code blocks to explore, or provide FINAL(...)",
			})
		}

		// Force final on last turn
		if turn == maxTurns-1 {
			messages = append(messages, Message{
				Role:    "user",
				Content: "IMPORTANT: This is your last turn. Provide FINAL(...) now.",
			})
		}
	}

	// Exhausted turns - try one more time for final
	response, err := llmClient.Complete(ctx, messages)
	if err != nil {
		return Result{Error: fmt.Sprintf("LLM error: %v", err), TurnsUsed: maxTurns + 1}
	}

	if strings.Contains(response, "FINAL(") {
		result := parseFinalAnswer(response)
		result.TurnsUsed = maxTurns + 1
		return result
	}

	return Result{
		Error:     "Max turns reached without FINAL answer",
		TurnsUsed: maxTurns + 1,
	}
}

// =============================================================================
// EXAMPLE USAGE
// =============================================================================

// Example tools - replace with your actual implementations
func exampleSearch(args ...any) (any, error) {
	query, _ := args[0].(string)
	_ = query // Use query in real implementation
	return []map[string]any{
		{"id": "1", "title": "Example", "score": 0.95},
	}, nil
}

func exampleFetch(args ...any) (any, error) {
	id, _ := args[0].(string)
	return map[string]any{
		"id":       id,
		"content":  "Example content",
		"metadata": map[string]any{},
	}, nil
}

func exampleStatus(args ...any) (any, error) {
	return map[string]any{
		"available": true,
		"resources": map[string]any{"memory": 1000},
	}, nil
}

const exampleSystemPrompt = `
You are an assistant operating inside a REPL environment.

## Available Functions

| Function | Description |
|----------|-------------|
| search(query, limit) | Search the knowledge base |
| fetch(id) | Get details by ID |
| status() | Check system status |

## Rules

1. **Explore first.** Call functions before answering.
2. **Use console.log().** Results must be printed to be visible.
3. **No require/import.** Don't use require, import, eval.
4. **Cite sources.** Reference which function calls provided data.

## Workflow

1. Examine context.query
2. Call tool functions
3. Output FINAL({...}) when ready

FINAL({
    "answer": "...",
    "reasoning": ["step 1", "step 2"],
    "sources": ["search(...)", "fetch(...)"]
})
`

// Example demonstrates how to use RunRLM.
func Example() {
	tools := map[string]ToolFunc{
		"search": exampleSearch,
		"fetch":  exampleFetch,
		"status": exampleStatus,
	}

	// Replace with actual LLM client
	var llmClient LLMClient // = NewOpenAIClient(apiKey)

	result := RunRLM(
		context.Background(),
		"What information do you have?",
		tools,
		llmClient,
		exampleSystemPrompt,
		5, // maxTurns
	)

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}
````

---

*This guide provides everything needed to implement RLM in any project.*
