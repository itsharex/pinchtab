#!/bin/bash
# 22-tabs-errors.sh — Tab error handling via CLI

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs returns valid JSON array"

pt_ok tab
assert_output_json "tabs output is valid JSON"
assert_output_contains "tabs" "response contains tabs field"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs new + close roundtrip"

# Use nav to create a tab in the existing instance (returns tabId reliably)
pt_ok nav "${FIXTURES_URL}/index.html"
assert_output_json
TAB_ID=$(echo "$PT_OUT" | jq -r '.tabId')

pt_ok tab close "$TAB_ID"

# Verify tab is gone
pt_ok tab
assert_output_not_contains "$TAB_ID" "closed tab no longer in list"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs close with no args → error"

pt_fail tab close

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs close nonexistent → error"

pt_fail tab close "nonexistent_tab_id_12345"

end_test
