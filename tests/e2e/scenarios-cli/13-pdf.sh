#!/bin/bash
# 13-pdf.sh — CLI PDF export command

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab pdf -o <file>"

pt_ok nav "${FIXTURES_URL}/form.html"

TMPFILE="/tmp/test-export-$$.pdf"
pt_ok pdf -o "$TMPFILE"

if [ -f "$TMPFILE" ] && [ -s "$TMPFILE" ]; then
  echo -e "  ${GREEN}✓${NC} pdf saved to file"
  ((ASSERTIONS_PASSED++)) || true
  rm -f "$TMPFILE"
else
  echo -e "  ${RED}✗${NC} pdf file not created or empty"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# Note: pdf without -o outputs binary to stdout which doesn't work
# well with our text-based pt() function, so we only test file output
