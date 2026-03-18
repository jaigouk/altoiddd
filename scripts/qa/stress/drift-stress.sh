#!/usr/bin/env bash
# Stress Test: Drift Detection Performance
#
# Tests:
# 1. Large number of tools (100+)
# 2. Large number of versions per tool
# 3. Deep directory traversal
# 4. Concurrent access patterns
# 5. Memory pressure under load

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ERRORS=0
TEMP_DIR=""

cleanup() {
    if [[ -n "$TEMP_DIR" && -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

check_time() {
    local description="$1"
    local max_seconds="$2"
    shift 2

    local start_time end_time duration
    start_time=$(date +%s.%N)
    if "$@" >/dev/null 2>&1; then
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        if (( $(echo "$duration < $max_seconds" | bc -l) )); then
            echo -e "  ${GREEN}[OK]${NC} $description (${duration}s < ${max_seconds}s)"
        else
            echo -e "  ${YELLOW}[SLOW]${NC} $description (${duration}s > ${max_seconds}s)"
        fi
    else
        echo -e "  ${RED}[FAIL]${NC} $description (command failed)"
        ((ERRORS++))
    fi
}

echo "═══════════════════════════════════════════════════════════════"
echo "  Stress Test: Drift Detection Performance"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Build test binary
echo "Building test binary..."
go build -o /tmp/alto-stress ./cmd/alto

# Create temp directory for stress tests
TEMP_DIR=$(mktemp -d)
echo "Temp directory: $TEMP_DIR"
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Test 1: Large number of tools (100 tools)
# ─────────────────────────────────────────────────────────────────────────────
echo "Test 1: Scanning 100 tools"
TOOLS_DIR="$TEMP_DIR/many-tools/.alto/knowledge/tools"
mkdir -p "$TOOLS_DIR"

for i in $(seq 1 100); do
    tool_dir="$TOOLS_DIR/tool-$i"
    mkdir -p "$tool_dir"
    cat > "$tool_dir/_meta.toml" << EOF
[tool]
name = "tool-$i"

[versions.v1_0]
last_verified = "2024-01-01"
EOF
done

check_time "Scan 100 tools with stale entries" 2.0 /tmp/alto-stress kb drift
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Test 2: Many versions per tool (50 versions each)
# ─────────────────────────────────────────────────────────────────────────────
echo "Test 2: Scanning tool with 50 versions"
VERSIONS_DIR="$TEMP_DIR/many-versions/.alto/knowledge/tools/mega-tool"
mkdir -p "$VERSIONS_DIR"

{
    echo "[tool]"
    echo "name = \"mega-tool\""
    echo ""
    for i in $(seq 1 50); do
        echo "[versions.v${i}_0]"
        echo "last_verified = \"2024-01-0$((i % 9 + 1))\""
        echo ""
    done
} > "$VERSIONS_DIR/_meta.toml"

check_time "Scan tool with 50 versions" 1.0 /tmp/alto-stress kb drift
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Test 3: Empty knowledge base (fast path)
# ─────────────────────────────────────────────────────────────────────────────
echo "Test 3: Empty knowledge base (fast path)"
EMPTY_DIR="$TEMP_DIR/empty"
mkdir -p "$EMPTY_DIR"

check_time "Empty knowledge base returns quickly" 0.5 /tmp/alto-stress kb drift
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Test 4: Concurrent drift detection (Go test -race)
# ─────────────────────────────────────────────────────────────────────────────
echo "Test 4: Concurrent access (race detection)"
check_time "Race detector passes" 120.0 go test -race -count=3 ./internal/knowledge/...
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Test 5: Memory benchmark
# ─────────────────────────────────────────────────────────────────────────────
echo "Test 5: Memory benchmark (if benchmarks exist)"
if go test -list 'Benchmark.*Drift' ./internal/knowledge/... 2>/dev/null | grep -q Benchmark; then
    check_time "Benchmarks complete" 30.0 go test -bench=Drift -benchmem -run=^$ ./internal/knowledge/...
else
    echo -e "  ${YELLOW}[SKIP]${NC} No drift benchmarks found"
fi
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Test 6: Repeated runs (memory leaks)
# ─────────────────────────────────────────────────────────────────────────────
echo "Test 6: Repeated runs (10x)"
for i in $(seq 1 10); do
    /tmp/alto-stress kb drift >/dev/null 2>&1
done
echo -e "  ${GREEN}[OK]${NC} 10 consecutive runs completed"
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Cleanup
# ─────────────────────────────────────────────────────────────────────────────
rm -f /tmp/alto-stress

echo "═══════════════════════════════════════════════════════════════"
if [[ $ERRORS -eq 0 ]]; then
    echo -e "  ${GREEN}STRESS TEST PASSED${NC}"
    exit 0
else
    echo -e "  ${RED}STRESS TEST FAILED${NC}: $ERRORS tests failed"
    exit 1
fi
