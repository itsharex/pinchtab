#!/bin/bash
# 12-type-ref.sh — CLI type and click with element refs

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab type <ref> <text>"

pt_ok nav "${FIXTURES_URL}/form.html"
pt_ok snap

USERNAME_REF=$(find_ref_by_name "Username:" "$PT_OUT")
if assert_ref_found "$USERNAME_REF" "username input ref"; then
  pt_ok type "$USERNAME_REF" "typed-via-ref"
  assert_output_contains "typed" "confirms text was typed"
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab click <ref>"

pt_ok nav "${FIXTURES_URL}/buttons.html"
pt_ok snap

BUTTON_REF=$(find_ref_by_role "button" "$PT_OUT")
if assert_ref_found "$BUTTON_REF" "button ref"; then
  pt_ok click "$BUTTON_REF"
  assert_output_contains "clicked" "confirms click by ref"
fi

end_test
