# QA Automation

Automated QA scripts for validating acceptance criteria from epics.

## Quick Start

```bash
# Run all QA tests
./scripts/qa/run-all.sh

# Run specific scenario
./scripts/qa/run-all.sh --scenario drift

# Run with stress tests
./scripts/qa/run-all.sh --stress

# Run with coverage report
./scripts/qa/run-all.sh --coverage
```

## Directory Structure

```
scripts/qa/
  run-all.sh              # Main entry point
  README.md               # This file
  scenarios/              # AC-based scenario tests
    drift.sh              # Epic: Knowledge Drift Detection (y79)
    rescue.sh             # Epic: Rescue Mode Hardening (ojl)
    ...
  stress/                 # Performance & stress tests
    drift-stress.sh       # Drift detection under load
    ...
```

## Writing Scenarios

Each scenario maps to an epic and validates its acceptance criteria.

### Template

```bash
#!/usr/bin/env bash
# QA Scenario: <Epic Title> (<epic-id>)
#
# Acceptance Criteria:
# 1. <AC1 description>
# 2. <AC2 description>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$PROJECT_ROOT"

ERRORS=0

check() {
    local description="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        echo "[OK] $description"
    else
        echo "[FAIL] $description"
        ((ERRORS++))
    fi
}

echo "AC1: <Description>"
check "Test 1" <command>
check "Test 2" <command>

# Exit with error count
exit $ERRORS
```

### Helper Functions

| Function | Usage | Description |
|----------|-------|-------------|
| `check` | `check "desc" command args` | Run command, pass if exit 0 |
| `check_output_contains` | `check_output_contains "desc" "expected" cmd` | Pass if output contains string |
| `check_exit_code` | `check_exit_code "desc" 1 cmd` | Pass if exit code matches |

## Writing Stress Tests

Stress tests validate performance under load.

### Guidelines

1. **Time limits**: Use `check_time "desc" 5.0 command` to fail if over 5 seconds
2. **Temp dirs**: Always clean up with `trap cleanup EXIT`
3. **Scale**: Test with 100+ items to find O(n^2) issues
4. **Race detection**: Run with `-race` flag
5. **Memory**: Watch for leaks with repeated runs

### Example Metrics

| Test | Threshold | Why |
|------|-----------|-----|
| Scan 100 tools | < 2s | Linear scan, should be fast |
| Empty KB | < 0.5s | Fast path, no I/O |
| 10 consecutive runs | completes | Memory leak detection |

## CI Integration

Add to `.github/workflows/qa.yml`:

```yaml
qa:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.25'
    - run: ./scripts/qa/run-all.sh --coverage
```

## Epic → Scenario Mapping

| Epic ID | Title | Scenario |
|---------|-------|----------|
| y79 | Knowledge Drift Detection | `drift.sh` |
| ojl | Rescue Mode Hardening | `rescue.sh` (todo) |
| 1ql | AI Challenger Integration | `challenger.sh` (todo) |

## Output Format

```
═══════════════════════════════════════════════════════════════
  QA Automation Suite
═══════════════════════════════════════════════════════════════
Project: alto-cli
Date: 2026-03-12 20:15:00

═══════════════════════════════════════════════════════════════
  Quality Gates
═══════════════════════════════════════════════════════════════

[INFO] Running go build...
[PASS] go build
[INFO] Running go vet...
[PASS] go vet
...

═══════════════════════════════════════════════════════════════
  Summary
═══════════════════════════════════════════════════════════════
  Passed:  15
  Failed:  0
  Skipped: 2

QA PASSED
```
