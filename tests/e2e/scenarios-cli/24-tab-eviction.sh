#!/bin/bash
# 24-tab-eviction.sh — LRU tab eviction via CLI
#
# The default test instance has maxTabs=10.

source "$(dirname "$0")/common.sh"

MAX_TABS=10

# ─────────────────────────────────────────────────────────────────
start_test "tab eviction: open tabs up to limit"

TAB_IDS=()
for i in $(seq 1 $MAX_TABS); do
  pt_ok nav "${FIXTURES_URL}/index.html?t=$i"
  TAB_IDS+=($(echo "$PT_OUT" | jq -r '.tabId'))
done

pt_ok tab
TAB_COUNT=$(echo "$PT_OUT" | jq '.tabs | length')
if [ "$TAB_COUNT" -ge "$MAX_TABS" ]; then
  echo -e "  ${GREEN}✓${NC} $TAB_COUNT tabs open (>= $MAX_TABS)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected >= $MAX_TABS tabs, got $TAB_COUNT"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tab eviction: new tab evicts oldest"

FIRST_TAB="${TAB_IDS[0]}"
sleep 1
pt_ok nav "${FIXTURES_URL}/index.html?t=overflow"

pt_ok tab
assert_output_not_contains "$FIRST_TAB" "oldest tab evicted (LRU)"

end_test
