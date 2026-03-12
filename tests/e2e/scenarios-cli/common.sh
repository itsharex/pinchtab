#!/bin/bash
# Common utilities for CLI E2E tests

set -uo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Environment
PINCHTAB_URL="${PINCHTAB_URL:-http://localhost:9999}"
FIXTURES_URL="${FIXTURES_URL:-http://localhost:8080}"
RESULTS_DIR="${RESULTS_DIR:-/results}"

# Test tracking (preserve across sourced scripts)
TESTS_PASSED="${TESTS_PASSED:-0}"
TESTS_FAILED="${TESTS_FAILED:-0}"
ASSERTIONS_PASSED="${ASSERTIONS_PASSED:-0}"
ASSERTIONS_FAILED="${ASSERTIONS_FAILED:-0}"
CURRENT_TEST="${CURRENT_TEST:-}"

# ─────────────────────────────────────────────────────────────────
# Wait for instance ready (same as curl-based tests)
# ─────────────────────────────────────────────────────────────────

wait_for_instance_ready() {
  local base_url="$1"
  local timeout_sec="${2:-60}"
  local started_at
  started_at=$(date +%s)

  while true; do
    local now
    now=$(date +%s)
    if [ $((now - started_at)) -ge "$timeout_sec" ]; then
      echo -e "  ${RED}✗${NC} instance at ${base_url} did not reach running within ${timeout_sec}s"
      return 1
    fi

    local inst_status
    inst_status=$(curl -sf "${base_url}/health" 2>/dev/null | jq -r '.defaultInstance.status // empty' 2>/dev/null || true)
    if [ "$inst_status" = "running" ]; then
      echo -e "  ${GREEN}✓${NC} instance ready at ${base_url}"
      return 0
    fi

    sleep 1
  done
}

# ─────────────────────────────────────────────────────────────────
# Test lifecycle
# ─────────────────────────────────────────────────────────────────

start_test() {
  CURRENT_TEST="$1"
  echo -e "${BLUE}▶ ${CURRENT_TEST}${NC}"
}

end_test() {
  if [ "$ASSERTIONS_FAILED" -gt 0 ]; then
    echo -e "${RED}✗ ${CURRENT_TEST} failed${NC}"
    ((TESTS_FAILED++)) || true
  else
    echo -e "${GREEN}✓ ${CURRENT_TEST}${NC}"
    ((TESTS_PASSED++)) || true
  fi
  ASSERTIONS_FAILED=0
  ASSERTIONS_PASSED=0
  echo ""
}

# ─────────────────────────────────────────────────────────────────
# CLI execution helpers
# ─────────────────────────────────────────────────────────────────

# Run pinchtab CLI command
# Usage: pt <command> [args...]
# Sets $PT_OUT (stdout), $PT_ERR (stderr), $PT_CODE (exit code)
pt() {
  local tmpout=$(mktemp)
  local tmperr=$(mktemp)
  
  echo -e "  ${BLUE}→ PINCHTAB_URL=$PINCHTAB_URL pinchtab $@${NC}"
  
  set +e
  PINCHTAB_URL="$PINCHTAB_URL" pinchtab "$@" > "$tmpout" 2> "$tmperr"
  PT_CODE=$?
  set -e
  
  PT_OUT=$(cat "$tmpout")
  PT_ERR=$(cat "$tmperr")
  rm -f "$tmpout" "$tmperr"
  
  if [ -n "$PT_OUT" ]; then
    echo "$PT_OUT" | head -5
  fi
}

# Run pinchtab and expect success (exit 0)
# Usage: pt_ok <command> [args...]
pt_ok() {
  pt "$@"
  if [ "$PT_CODE" -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} exit 0"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} expected exit 0, got $PT_CODE"
    echo -e "  ${RED}stderr: $PT_ERR${NC}"
    ((ASSERTIONS_FAILED++)) || true
  fi
}

# Run pinchtab and expect failure (non-zero exit)
# Usage: pt_fail <command> [args...]
pt_fail() {
  pt "$@"
  if [ "$PT_CODE" -ne 0 ]; then
    echo -e "  ${GREEN}✓${NC} exit $PT_CODE (expected failure)"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} expected non-zero exit, got 0"
    ((ASSERTIONS_FAILED++)) || true
  fi
}

# ─────────────────────────────────────────────────────────────────
# Assertions
# ─────────────────────────────────────────────────────────────────

# Assert PT_OUT contains string
assert_output_contains() {
  local expected="$1"
  local desc="${2:-output contains '$expected'}"
  
  if echo "$PT_OUT" | grep -q "$expected"; then
    echo -e "  ${GREEN}✓${NC} $desc"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} $desc"
    echo -e "  ${RED}  output was: $PT_OUT${NC}"
    ((ASSERTIONS_FAILED++)) || true
  fi
}

# Assert PT_OUT does not contain string
assert_output_not_contains() {
  local forbidden="$1"
  local desc="${2:-output does not contain '$forbidden'}"
  
  if echo "$PT_OUT" | grep -q "$forbidden"; then
    echo -e "  ${RED}✗${NC} $desc"
    echo -e "  ${RED}  output was: $PT_OUT${NC}"
    ((ASSERTIONS_FAILED++)) || true
  else
    echo -e "  ${GREEN}✓${NC} $desc"
    ((ASSERTIONS_PASSED++)) || true
  fi
}

# Assert PT_OUT is valid JSON
assert_output_json() {
  local desc="${1:-output is valid JSON}"
  
  if echo "$PT_OUT" | jq . > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} $desc"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} $desc"
    echo -e "  ${RED}  output was: $PT_OUT${NC}"
    ((ASSERTIONS_FAILED++)) || true
  fi
}

# Assert PT_OUT JSON field equals value
assert_json_field() {
  local path="$1"
  local expected="$2"
  local desc="${3:-$path equals '$expected'}"
  
  local actual
  actual=$(echo "$PT_OUT" | jq -r "$path" 2>/dev/null)
  
  if [ "$actual" = "$expected" ]; then
    echo -e "  ${GREEN}✓${NC} $desc"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} $desc (got '$actual')"
    ((ASSERTIONS_FAILED++)) || true
  fi
}

# ─────────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────────

print_summary() {
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  local total=$((TESTS_PASSED + TESTS_FAILED))
  if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "  ${GREEN}All $total CLI tests passed!${NC}"
  else
    echo -e "  ${GREEN}Passed:${NC} $TESTS_PASSED/$total"
    echo -e "  ${RED}Failed:${NC} $TESTS_FAILED/$total"
  fi
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  
  # Save results
  if [ -d "${RESULTS_DIR:-}" ]; then
    echo "passed=$TESTS_PASSED" > "${RESULTS_DIR}/summary.txt"
    echo "failed=$TESTS_FAILED" >> "${RESULTS_DIR}/summary.txt"
  fi
  
  # Exit with failure if any tests failed
  [ "$TESTS_FAILED" -eq 0 ]
}
