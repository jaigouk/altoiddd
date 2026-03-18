#!/usr/bin/env bash
# QA Scenario: Knowledge Drift Detection (alto-cli-y79)
#
# Acceptance Criteria:
# 1. DriftDetectionHandler filters by tool (case-insensitive)
# 2. DriftDetectionAdapter scans _meta.toml for staleness
# 3. Default threshold is 14 days
# 4. CLI command: alto kb drift [tool]
# 5. Returns exit code 1 when errors present
# 6. Gracefully handles missing/malformed files

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

check() {
    local description="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        echo -e "  ${GREEN}[OK]${NC} $description"
    else
        echo -e "  ${RED}[FAIL]${NC} $description"
        ((ERRORS++))
    fi
}

check_output_contains() {
    local description="$1"
    local expected="$2"
    shift 2
    local output
    output=$("$@" 2>&1) || true
    if echo "$output" | grep -q "$expected"; then
        echo -e "  ${GREEN}[OK]${NC} $description"
    else
        echo -e "  ${RED}[FAIL]${NC} $description (expected: '$expected')"
        ((ERRORS++))
    fi
}

check_exit_code() {
    local description="$1"
    local expected_code="$2"
    shift 2
    local actual_code=0
    "$@" >/dev/null 2>&1 || actual_code=$?
    if [[ $actual_code -eq $expected_code ]]; then
        echo -e "  ${GREEN}[OK]${NC} $description (exit code: $actual_code)"
    else
        echo -e "  ${RED}[FAIL]${NC} $description (expected: $expected_code, got: $actual_code)"
        ((ERRORS++))
    fi
}

echo "═══════════════════════════════════════════════════════════════"
echo "  Scenario: Knowledge Drift Detection (alto-cli-y79)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# AC1: DriftDetectionHandler exists and is wired
# ─────────────────────────────────────────────────────────────────────────────
echo "AC1: DriftDetectionHandler exists and is wired"
check "Handler file exists" test -f internal/knowledge/application/drift_detection_handler.go
check "Handler test file exists" test -f internal/knowledge/application/drift_detection_handler_test.go
check "Handler wired in composition" grep -q "DriftDetectionHandler" internal/composition/app.go

# ─────────────────────────────────────────────────────────────────────────────
# AC2: DriftDetectionAdapter scans _meta.toml
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "AC2: DriftDetectionAdapter scans _meta.toml"
check "Adapter file exists" test -f internal/knowledge/infrastructure/drift_detection_adapter.go
check "Adapter test file exists" test -f internal/knowledge/infrastructure/drift_detection_adapter_test.go
check "Adapter has compile-time interface check" grep -q "var _ knowledgeapp.DriftDetection" internal/knowledge/infrastructure/drift_detection_adapter.go
check "Adapter reads _meta.toml" grep -q "_meta.toml" internal/knowledge/infrastructure/drift_detection_adapter.go

# ─────────────────────────────────────────────────────────────────────────────
# AC3: Default threshold is 14 days
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "AC3: Default threshold is 14 days"
check "DefaultStaleThresholdDays = 14" grep -q "DefaultStaleThresholdDays = 14" internal/knowledge/infrastructure/drift_detection_adapter.go
check "WithStaleThreshold method exists" grep -q "func.*WithStaleThreshold" internal/knowledge/infrastructure/drift_detection_adapter.go

# ─────────────────────────────────────────────────────────────────────────────
# AC4: CLI command alto kb drift [tool]
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "AC4: CLI command alto kb drift [tool]"
check "kb.go has drift subcommand" grep -q "newKBDriftCmd" cmd/alto/commands/kb.go
check "Drift command accepts optional tool arg" grep -q "MaximumNArgs(1)" cmd/alto/commands/kb.go

# Build and test CLI
echo ""
echo "AC4b: CLI integration tests"
go build -o /tmp/alto-test ./cmd/alto
check_output_contains "alto kb drift runs without error" "No drift detected" /tmp/alto-test kb drift
check_output_contains "alto kb shows categories" "Knowledge Base Categories" /tmp/alto-test kb
rm -f /tmp/alto-test

# ─────────────────────────────────────────────────────────────────────────────
# AC5: Graceful handling of edge cases
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "AC5: Edge case handling (test coverage)"
check "Test: no knowledge dir" grep -q "returns empty report when no knowledge dir" internal/knowledge/infrastructure/drift_detection_adapter_test.go
check "Test: no meta files" grep -q "returns empty report when no meta files" internal/knowledge/infrastructure/drift_detection_adapter_test.go
check "Test: malformed TOML" grep -q "skips malformed toml gracefully" internal/knowledge/infrastructure/drift_detection_adapter_test.go
check "Test: missing last_verified" grep -q "treats missing last_verified as stale" internal/knowledge/infrastructure/drift_detection_adapter_test.go
check "Test: context cancellation" grep -q "context cancellation returns error" internal/knowledge/infrastructure/drift_detection_adapter_test.go
check "Test: custom threshold" grep -q "custom threshold changes staleness detection" internal/knowledge/infrastructure/drift_detection_adapter_test.go

# ─────────────────────────────────────────────────────────────────────────────
# AC6: Documentation
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "AC6: Documentation"
check "README documents kb drift" grep -q "alto kb drift" README.md
check "README explains 14-day threshold" grep -q "14 days" README.md

# ─────────────────────────────────────────────────────────────────────────────
# AC7: Run unit tests for drift detection
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "AC7: Unit tests pass"
check "Handler tests pass" go test -v ./internal/knowledge/application/... -run Drift
check "Adapter tests pass" go test -v ./internal/knowledge/infrastructure/... -run Drift

# ─────────────────────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════════"
if [[ $ERRORS -eq 0 ]]; then
    echo -e "  ${GREEN}SCENARIO PASSED${NC}: All acceptance criteria met"
    exit 0
else
    echo -e "  ${RED}SCENARIO FAILED${NC}: $ERRORS acceptance criteria failed"
    exit 1
fi
