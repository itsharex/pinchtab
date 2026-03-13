#!/bin/bash
# 37-idpi.sh — IDPI (Indirect Prompt Injection Detection) on /find and /pdf
#
# Tests content-based IDPI scanning:
#   - PINCHTAB_URL (main): IDPI enabled, scanContent=true (default), warn mode
#   - PINCHTAB_SECURE_URL (secure): IDPI enabled, strictMode=true

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
# Helper: curl with headers capture
# ─────────────────────────────────────────────────────────────────

# POST and capture response body, HTTP status, and specific header
# Usage: pt_post_hdr <base_url> <path> <body> <header_name>
# Sets: RESULT, HTTP_STATUS, HDR_VALUE
pt_post_hdr() {
  local base_url="$1"
  local path="$2"
  local body="$3"
  local header_name="$4"

  echo -e "${BLUE}→ curl -X POST ${base_url}${path}${NC}" >&2
  local tmpheaders=$(mktemp)
  local response
  response=$(curl -s -w "\n%{http_code}" \
    -X POST \
    "${base_url}${path}" \
    -H "Content-Type: application/json" \
    -D "$tmpheaders" \
    -d "$body")
  RESULT=$(echo "$response" | head -n -1)
  HTTP_STATUS=$(echo "$response" | tail -n 1)
  HDR_VALUE=$(grep -i "^${header_name}:" "$tmpheaders" | sed 's/^[^:]*: *//' | tr -d '\r' | head -1)
  rm -f "$tmpheaders"
}

# GET with headers capture
pt_get_hdr() {
  local base_url="$1"
  local path="$2"
  local header_name="$3"

  echo -e "${BLUE}→ curl -X GET ${base_url}${path}${NC}" >&2
  local tmpheaders=$(mktemp)
  local response
  response=$(curl -s -w "\n%{http_code}" \
    -X GET \
    "${base_url}${path}" \
    -H "Content-Type: application/json" \
    -D "$tmpheaders")
  RESULT=$(echo "$response" | head -n -1)
  HTTP_STATUS=$(echo "$response" | tail -n 1)
  HDR_VALUE=$(grep -i "^${header_name}:" "$tmpheaders" | sed 's/^[^:]*: *//' | tr -d '\r' | head -1)
  rm -f "$tmpheaders"
}

# Helper: navigate and get tab ID
idpi_nav() {
  local base_url="$1"
  local page_url="$2"
  local old_url="$PINCHTAB_URL"
  PINCHTAB_URL="$base_url"
  pt_post /navigate "{\"url\":\"$page_url\"}" >/dev/null
  PINCHTAB_URL="$old_url"
  echo "$RESULT" | jq -r '.tabId'
}

# Helper: close tab
idpi_close() {
  local base_url="$1"
  local tab_id="$2"
  curl -sf -X POST "${base_url}/tab" \
    -H "Content-Type: application/json" \
    -d "{\"tabId\":\"$tab_id\",\"action\":\"close\"}" >/dev/null 2>&1 || true
}

# ═══════════════════════════════════════════════════════════════════
# /find WARN MODE (main instance)
# ═══════════════════════════════════════════════════════════════════

start_test "idpi: /find clean page — no warning (warn mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_URL" "${FIXTURES_URL}/idpi-clean.html")
sleep 1
pt_post_hdr "$PINCHTAB_URL" "/tabs/${TAB_ID}/find" '{"query":"safe action button","threshold":0.1,"topK":5}' "X-IDPI-Warning"
assert_ok "/find clean page"
if [ -z "$HDR_VALUE" ]; then
  echo -e "  ${GREEN}✓${NC} no X-IDPI-Warning header (expected)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} unexpected X-IDPI-Warning: $HDR_VALUE"
  ((ASSERTIONS_FAILED++)) || true
fi
idpi_close "$PINCHTAB_URL" "$TAB_ID"

end_test

# ─────────────────────────────────────────────────────────────────

start_test "idpi: /find injection page — warns (warn mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_URL" "${FIXTURES_URL}/idpi-inject.html")
sleep 1
pt_post_hdr "$PINCHTAB_URL" "/tabs/${TAB_ID}/find" '{"query":"continue button","threshold":0.1,"topK":5}' "X-IDPI-Warning"
assert_ok "/find injection page (warn mode returns 200)"
if [ -n "$HDR_VALUE" ]; then
  echo -e "  ${GREEN}✓${NC} X-IDPI-Warning header present: $HDR_VALUE"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} missing X-IDPI-Warning header"
  ((ASSERTIONS_FAILED++)) || true
fi
# Check idpiWarning in JSON body
IDPI_BODY=$(echo "$RESULT" | jq -r '.idpiWarning // empty')
if [ -n "$IDPI_BODY" ]; then
  echo -e "  ${GREEN}✓${NC} idpiWarning field in response body"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} missing idpiWarning in response body"
  ((ASSERTIONS_FAILED++)) || true
fi
idpi_close "$PINCHTAB_URL" "$TAB_ID"

end_test

# ─────────────────────────────────────────────────────────────────

start_test "idpi: POST /find injection page — warns (warn mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_URL" "${FIXTURES_URL}/idpi-inject.html")
sleep 1
pt_post_hdr "$PINCHTAB_URL" "/find" "{\"query\":\"malicious paragraph\",\"tabId\":\"$TAB_ID\",\"threshold\":0.1}" "X-IDPI-Warning"
assert_ok "POST /find (warn mode)"
if [ -n "$HDR_VALUE" ]; then
  echo -e "  ${GREEN}✓${NC} X-IDPI-Warning on POST /find"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} missing X-IDPI-Warning on POST /find"
  ((ASSERTIONS_FAILED++)) || true
fi
idpi_close "$PINCHTAB_URL" "$TAB_ID"

end_test

# ═══════════════════════════════════════════════════════════════════
# /find STRICT MODE (secure instance)
# ═══════════════════════════════════════════════════════════════════

start_test "idpi: /find clean page — allowed (strict mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_SECURE_URL" "${FIXTURES_URL}/idpi-clean.html")
sleep 1
pt_post_hdr "$PINCHTAB_SECURE_URL" "/tabs/${TAB_ID}/find" '{"query":"safe action button","threshold":0.1,"topK":5}' "X-IDPI-Warning"
assert_ok "/find clean page (strict mode)"
idpi_close "$PINCHTAB_SECURE_URL" "$TAB_ID"

end_test

# ─────────────────────────────────────────────────────────────────

start_test "idpi: /find injection page — blocked (strict mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_SECURE_URL" "${FIXTURES_URL}/idpi-inject.html")
sleep 1
pt_post_hdr "$PINCHTAB_SECURE_URL" "/tabs/${TAB_ID}/find" '{"query":"continue button","threshold":0.1,"topK":5}' "X-IDPI-Warning"
assert_http_status 403 "/find blocked in strict mode"
assert_contains "$RESULT" "idpi" "403 body mentions IDPI"
idpi_close "$PINCHTAB_SECURE_URL" "$TAB_ID"

end_test

# ═══════════════════════════════════════════════════════════════════
# /pdf WARN MODE (main instance)
# ═══════════════════════════════════════════════════════════════════

start_test "idpi: /pdf clean page — no warning (warn mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_URL" "${FIXTURES_URL}/idpi-clean.html")
sleep 1
pt_get_hdr "$PINCHTAB_URL" "/tabs/${TAB_ID}/pdf" "X-IDPI-Warning"
assert_ok "/pdf clean page"
if [ -z "$HDR_VALUE" ]; then
  echo -e "  ${GREEN}✓${NC} no X-IDPI-Warning on clean PDF"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} unexpected X-IDPI-Warning: $HDR_VALUE"
  ((ASSERTIONS_FAILED++)) || true
fi
idpi_close "$PINCHTAB_URL" "$TAB_ID"

end_test

# ─────────────────────────────────────────────────────────────────

start_test "idpi: /pdf injection page — warns (warn mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_URL" "${FIXTURES_URL}/idpi-inject.html")
sleep 1
pt_get_hdr "$PINCHTAB_URL" "/tabs/${TAB_ID}/pdf" "X-IDPI-Warning"
assert_ok "/pdf injection page (warn mode returns 200)"
if [ -n "$HDR_VALUE" ]; then
  echo -e "  ${GREEN}✓${NC} X-IDPI-Warning header on injection PDF"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} missing X-IDPI-Warning on injection PDF"
  ((ASSERTIONS_FAILED++)) || true
fi
idpi_close "$PINCHTAB_URL" "$TAB_ID"

end_test

# ═══════════════════════════════════════════════════════════════════
# /pdf STRICT MODE (secure instance)
# ═══════════════════════════════════════════════════════════════════

start_test "idpi: /pdf clean page — allowed (strict mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_SECURE_URL" "${FIXTURES_URL}/idpi-clean.html")
sleep 1
pt_get_hdr "$PINCHTAB_SECURE_URL" "/tabs/${TAB_ID}/pdf" "X-IDPI-Warning"
assert_ok "/pdf clean page (strict mode)"
idpi_close "$PINCHTAB_SECURE_URL" "$TAB_ID"

end_test

# ─────────────────────────────────────────────────────────────────

start_test "idpi: /pdf injection page — blocked (strict mode)"

TAB_ID=$(idpi_nav "$PINCHTAB_SECURE_URL" "${FIXTURES_URL}/idpi-inject.html")
sleep 1
pt_get_hdr "$PINCHTAB_SECURE_URL" "/tabs/${TAB_ID}/pdf" "X-IDPI-Warning"
assert_http_status 403 "/pdf blocked in strict mode"
assert_contains "$RESULT" "idpi" "403 body mentions IDPI"
idpi_close "$PINCHTAB_SECURE_URL" "$TAB_ID"

end_test

# ═══════════════════════════════════════════════════════════════════
# EDGE CASES
# ═══════════════════════════════════════════════════════════════════

start_test "idpi: multiple injection phrases — single warning header"

TAB_ID=$(idpi_nav "$PINCHTAB_URL" "${FIXTURES_URL}/idpi-inject.html")
sleep 1
# Capture all X-IDPI-Warning headers
tmpheaders=$(mktemp)
curl -s -X POST "${PINCHTAB_URL}/tabs/${TAB_ID}/find" \
  -H "Content-Type: application/json" \
  -D "$tmpheaders" \
  -d '{"query":"continue button","threshold":0.1,"topK":5}' >/dev/null

HDR_COUNT=$(grep -ci "^X-IDPI-Warning:" "$tmpheaders" 2>/dev/null || true)
HDR_COUNT=${HDR_COUNT:-0}
rm -f "$tmpheaders"

if [ "$HDR_COUNT" -eq 1 ]; then
  echo -e "  ${GREEN}✓${NC} exactly one X-IDPI-Warning header"
  ((ASSERTIONS_PASSED++)) || true
elif [ "$HDR_COUNT" -eq 0 ]; then
  echo -e "  ${RED}✗${NC} no X-IDPI-Warning header found"
  ((ASSERTIONS_FAILED++)) || true
else
  echo -e "  ${RED}✗${NC} expected 1 X-IDPI-Warning, got $HDR_COUNT"
  ((ASSERTIONS_FAILED++)) || true
fi
idpi_close "$PINCHTAB_URL" "$TAB_ID"

end_test

# ─────────────────────────────────────────────────────────────────

start_test "idpi: /pdf?raw=true blocked in strict mode"

TAB_ID=$(idpi_nav "$PINCHTAB_SECURE_URL" "${FIXTURES_URL}/idpi-inject.html")
sleep 1
pt_get_hdr "$PINCHTAB_SECURE_URL" "/tabs/${TAB_ID}/pdf?raw=true" "X-IDPI-Warning"
assert_http_status 403 "raw PDF blocked in strict mode"
idpi_close "$PINCHTAB_SECURE_URL" "$TAB_ID"

end_test
