#!/usr/bin/env bash
# QA Automation Runner
# Usage: ./scripts/qa/run-all.sh [--scenario <name>] [--stress] [--coverage]
#
# Examples:
#   ./scripts/qa/run-all.sh                    # Run all scenarios
#   ./scripts/qa/run-all.sh --scenario drift   # Run specific scenario
#   ./scripts/qa/run-all.sh --stress           # Run stress tests
#   ./scripts/qa/run-all.sh --coverage         # Run with coverage report

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
SKIPPED=0

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; ((PASSED++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; ((FAILED++)); }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; ((SKIPPED++)); }
log_header() { echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"; echo -e "${BLUE}  $1${NC}"; echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"; }

# Parse arguments
SCENARIO=""
RUN_STRESS=false
RUN_COVERAGE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --scenario) SCENARIO="$2"; shift 2 ;;
        --stress) RUN_STRESS=true; shift ;;
        --coverage) RUN_COVERAGE=true; shift ;;
        --help)
            echo "Usage: $0 [--scenario <name>] [--stress] [--coverage]"
            echo ""
            echo "Options:"
            echo "  --scenario <name>  Run specific scenario (e.g., drift, rescue)"
            echo "  --stress           Run stress tests"
            echo "  --coverage         Generate coverage report"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# ─────────────────────────────────────────────────────────────────────────────
# Quality Gates
# ─────────────────────────────────────────────────────────────────────────────

run_quality_gates() {
    log_header "Quality Gates"

    log_info "Running go build..."
    if go build ./... 2>&1; then
        log_pass "go build"
    else
        log_fail "go build"
        return 1
    fi

    log_info "Running go vet..."
    if go vet ./... 2>&1; then
        log_pass "go vet"
    else
        log_fail "go vet"
        return 1
    fi

    log_info "Running golangci-lint..."
    if golangci-lint run ./... 2>&1; then
        log_pass "golangci-lint"
    else
        log_fail "golangci-lint"
        return 1
    fi

    log_info "Running tests with race detector..."
    if $RUN_COVERAGE; then
        if go test ./... -race -coverprofile=coverage.out 2>&1; then
            log_pass "go test -race -cover"
            go tool cover -func=coverage.out | tail -1
        else
            log_fail "go test -race -cover"
            return 1
        fi
    else
        if go test ./... -race 2>&1; then
            log_pass "go test -race"
        else
            log_fail "go test -race"
            return 1
        fi
    fi
}

# ─────────────────────────────────────────────────────────────────────────────
# Scenario Runner
# ─────────────────────────────────────────────────────────────────────────────

run_scenarios() {
    log_header "Scenario Tests"

    if [[ -n "$SCENARIO" ]]; then
        # Run specific scenario
        scenario_file="$SCRIPT_DIR/scenarios/${SCENARIO}.sh"
        if [[ -f "$scenario_file" ]]; then
            log_info "Running scenario: $SCENARIO"
            if bash "$scenario_file"; then
                log_pass "Scenario: $SCENARIO"
            else
                log_fail "Scenario: $SCENARIO"
            fi
        else
            log_fail "Scenario not found: $scenario_file"
        fi
    else
        # Run all scenarios
        for scenario_file in "$SCRIPT_DIR"/scenarios/*.sh; do
            if [[ -f "$scenario_file" ]]; then
                scenario_name=$(basename "$scenario_file" .sh)
                log_info "Running scenario: $scenario_name"
                if bash "$scenario_file"; then
                    log_pass "Scenario: $scenario_name"
                else
                    log_fail "Scenario: $scenario_name"
                fi
            fi
        done
    fi
}

# ─────────────────────────────────────────────────────────────────────────────
# Stress Tests
# ─────────────────────────────────────────────────────────────────────────────

run_stress_tests() {
    log_header "Stress Tests"

    for stress_file in "$SCRIPT_DIR"/stress/*.sh; do
        if [[ -f "$stress_file" ]]; then
            stress_name=$(basename "$stress_file" .sh)
            log_info "Running stress test: $stress_name"
            if bash "$stress_file"; then
                log_pass "Stress: $stress_name"
            else
                log_fail "Stress: $stress_name"
            fi
        fi
    done
}

# ─────────────────────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────────────────────

main() {
    log_header "QA Automation Suite"
    echo "Project: $(basename "$PROJECT_ROOT")"
    echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""

    run_quality_gates
    run_scenarios

    if $RUN_STRESS; then
        run_stress_tests
    fi

    # Summary
    log_header "Summary"
    echo -e "  ${GREEN}Passed:${NC}  $PASSED"
    echo -e "  ${RED}Failed:${NC}  $FAILED"
    echo -e "  ${YELLOW}Skipped:${NC} $SKIPPED"
    echo ""

    if [[ $FAILED -gt 0 ]]; then
        echo -e "${RED}QA FAILED${NC}"
        exit 1
    else
        echo -e "${GREEN}QA PASSED${NC}"
        exit 0
    fi
}

main
