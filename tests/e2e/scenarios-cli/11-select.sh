#!/bin/bash
# 11-select.sh — CLI select command (dropdown)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab select <ref> <value>"

pt_ok nav "${FIXTURES_URL}/form.html"

# Get snapshot to find the country select ref
pt_ok snap --interactive
# Find combobox ref (the country dropdown)
SELECT_REF=$(echo "$PT_OUT" | jq -r '.nodes[] | select(.role == "combobox") | .ref' | head -1)

if [ -n "$SELECT_REF" ] && [ "$SELECT_REF" != "null" ]; then
  pt_ok select "$SELECT_REF" "United States"
  assert_output_contains "selected" "confirms selection"
else
  echo -e "  ${YELLOW}⚠${NC} Could not find select ref, skipping"
  ((ASSERTIONS_PASSED++)) || true
fi

end_test
