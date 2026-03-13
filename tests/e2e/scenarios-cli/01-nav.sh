#!/bin/bash
# 01-open.sh — CLI open commands

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab open <url>"

pt_ok open "${FIXTURES_URL}/index.html"
assert_output_json
assert_output_contains "tabId" "returns tab ID"
assert_output_contains "title" "returns page title"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab open (invalid URL)"

pt_fail open "not-a-valid-url"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab open --tab <tabId> <url>"

# First navigate to get a tab
pt_ok open "${FIXTURES_URL}/index.html"
TAB_ID=$(echo "$PT_OUT" | jq -r '.tabId')

# Navigate same tab using --tab flag
pt_ok open "${FIXTURES_URL}/form.html" --tab "$TAB_ID"
assert_output_contains "form.html" "navigated to form.html"

end_test
