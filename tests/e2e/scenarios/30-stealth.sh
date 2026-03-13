#!/bin/bash
# 30-stealth.sh — Stealth and fingerprint tests
# Migrated from: tests/integration/stealth_test.go

source "$(dirname "$0")/common.sh"

# Navigate first
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/index.html\"}"
assert_ok "navigate"
sleep 1

# ─────────────────────────────────────────────────────────────────
start_test "stealth: webdriver is undefined"

# Poll for stealth injection (up to 2s)
STEALTH_OK=false
for i in $(seq 1 5); do
  pt_post /evaluate '{"expression":"navigator.webdriver === undefined"}'
  if echo "$RESULT" | jq -r '.result' 2>/dev/null | grep -q "true"; then
    STEALTH_OK=true
    break
  fi
  sleep 0.4
done
if [ "$STEALTH_OK" = "true" ]; then
  echo -e "  ${GREEN}✓${NC} navigator.webdriver is undefined"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} navigator.webdriver still defined"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "stealth: plugins present"

STEALTH_OK=false
for i in $(seq 1 5); do
  pt_post /evaluate '{"expression":"navigator.plugins.length > 0"}'
  if echo "$RESULT" | jq -r '.result' 2>/dev/null | grep -q "true"; then
    STEALTH_OK=true
    break
  fi
  sleep 0.4
done
if [ "$STEALTH_OK" = "true" ]; then
  echo -e "  ${GREEN}✓${NC} navigator.plugins spoofed"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} navigator.plugins empty"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "stealth: chrome.runtime present"

STEALTH_OK=false
for i in $(seq 1 5); do
  pt_post /evaluate '{"expression":"!!window.chrome && !!window.chrome.runtime"}'
  if echo "$RESULT" | jq -r '.result' 2>/dev/null | grep -q "true"; then
    STEALTH_OK=true
    break
  fi
  sleep 0.4
done
if [ "$STEALTH_OK" = "true" ]; then
  echo -e "  ${GREEN}✓${NC} window.chrome.runtime present"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} window.chrome.runtime missing"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "stealth: fingerprint rotate"

pt_post /fingerprint/rotate '{"os":"windows"}'
assert_ok "fingerprint rotate (windows)"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "stealth: fingerprint rotate (random)"

pt_post /fingerprint/rotate '{}'
assert_ok "fingerprint rotate (random)"

end_test

# ─────────────────────────────────────────────────────────────────
# ─────────────────────────────────────────────────────────────────
start_test "stealth: fingerprint rotate (specific tab)"

pt_post /navigate "{\"url\":\"${FIXTURES_URL}/index.html\"}"
assert_ok "navigate"
TAB_ID=$(get_tab_id)

pt_post /fingerprint/rotate "{\"tabId\":\"${TAB_ID}\",\"os\":\"mac\"}"
assert_ok "fingerprint rotate on tab"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "stealth: status endpoint"

pt_get /stealth/status
assert_ok "stealth status"
assert_json_exists "$RESULT" '.score'
assert_json_exists "$RESULT" '.level'

end_test
