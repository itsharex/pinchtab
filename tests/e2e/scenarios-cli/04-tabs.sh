#!/bin/bash
# 04-tabs.sh — CLI tab management commands

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs (list)"

# Create a tab first
pt_ok nav "${FIXTURES_URL}/index.html"

pt_ok tab
assert_output_json
assert_output_contains "tabs" "returns tabs array"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs close <id>"

pt_ok nav "${FIXTURES_URL}/form.html"
TAB_ID=$(echo "$PT_OUT" | jq -r '.tabId')

pt_ok tab close "$TAB_ID"

# Verify tab is gone
pt_ok tab
assert_output_not_contains "$TAB_ID" "tab was closed"

end_test
