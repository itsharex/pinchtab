#!/bin/bash
# 32-tabs-errors.sh — Tab management error cases and edge cases
# Migrated from: tests/integration/tabs_test.go (TB4-TB8)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "tabs: list returns array"

pt_get /tabs
assert_ok "list tabs"
assert_json_exists "$RESULT" '.tabs'
assert_json_length_gte "$RESULT" '.tabs' '1' "at least 1 tab"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tabs: new + close roundtrip"

pt_post /tab "{\"action\":\"new\",\"url\":\"${FIXTURES_URL}/index.html\"}"
assert_ok "new tab"
NEW_TAB=$(echo "$RESULT" | jq -r '.tabId')

pt_post /tab "{\"action\":\"close\",\"tabId\":\"${NEW_TAB}\"}"
assert_ok "close tab"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tabs: close without tabId → 400"

pt_post /tab '{"action":"close"}'
assert_http_status "400" "rejects close without tabId"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tabs: bad action → 400"

pt_post /tab '{"action":"explode"}'
assert_http_status "400" "rejects bad action"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tabs: new tab returns tabId"

pt_post /tab '{"action":"new","url":"about:blank"}'
assert_ok "new tab"
TAB_ID=$(echo "$RESULT" | jq -r '.tabId')

if [ -n "$TAB_ID" ] && [ "$TAB_ID" != "null" ]; then
  echo -e "  ${GREEN}✓${NC} tabId returned: ${TAB_ID:0:12}..."
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} no tabId in response"
  ((ASSERTIONS_FAILED++)) || true
fi

# Cleanup
pt_post /tab "{\"action\":\"close\",\"tabId\":\"${TAB_ID}\"}"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tabs: nonexistent tab → 404"

FAKE_TAB="A25658CE1BA82659EBE9C93C46CEE63A"

pt_post "/tabs/${FAKE_TAB}/navigate" "{\"url\":\"${FIXTURES_URL}/index.html\"}"
assert_http_status "404" "navigate on fake tab"

pt_get "/tabs/${FAKE_TAB}/snapshot"
assert_http_status "404" "snapshot on fake tab"

end_test
