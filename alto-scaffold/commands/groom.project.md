# /groom — alto-specific addenda

## Ticket ID format

alto tickets use the `alto-` / `alty-cli-` prefix (e.g. `alto-0m9.2`, `alty-cli-766.1`).
Example usage:

```
/groom alto-0m9.2
/groom alty-cli-766.2
```

## Project-specific verification

```bash
alto ticket-verify <ticket-id>      # detects quantitative claims, verifies against command output
```

| Result | Action |
|--------|--------|
| All claims verified | Proceed to Phase 4 |
| Mismatch detected | Update ticket with correct values before proceeding |
| No claims found | Proceed to Phase 4 (no quantitative claims to verify) |
| Command not in allowlist | Note as UNVERIFIED in report |

## Composition root paths (Go layout)

- Port interfaces: `internal/{context}/application/ports.go`
- Adapters: `internal/{context}/infrastructure/{adapter}.go`
- Composition: `internal/composition/app.go`, `internal/composition/adapters.go`

Example constructor chain trace (replaces Phase 4b chain in the GENERIC parent):

```
NewXxxHandler(port)
  -> port type: XxxPort interface at internal/xxx/application/ports.go:NN
  -> methods: Foo(ctx, string) (Result, error) — confirmed line NN
  -> adapter: XxxAdapter at internal/xxx/infrastructure/xxx_adapter.go
  -> adapter constructor: NewXxxAdapter(dep1, dep2)
  -> dep1 comes from: ...
  -> wired in NewApp() at internal/composition/app.go:NN
  -> imports needed: xxxapp, xxxinfra
```

## Template reference

Generic template path becomes: `alto-scaffold/templates/beads-ticket-template.md`.
