#!/bin/bash
# 26-cookies.sh — Cookie management API tests
# Migrated from: tests/integration/cookies_test.go

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "GET /cookies (read cookies)"

pt_post /navigate "{\"url\":\"${FIXTURES_URL}/index.html\"}"
TAB_ID=$(get_tab_id)

pt_get "/cookies?tabId=${TAB_ID}"
assert_ok "get cookies"
assert_json_exists "$RESULT" '.cookies'

# If cookies exist, verify structure has required fields
COOKIE_COUNT=$(echo "$RESULT" | jq '.cookies | length')
if [ "$COOKIE_COUNT" -gt 0 ]; then
  assert_json_exists "$RESULT" '.cookies[0].name' "cookie has name"
  assert_json_exists "$RESULT" '.cookies[0].value' "cookie has value"
  assert_json_exists "$RESULT" '.cookies[0].domain' "cookie has domain"
  assert_json_exists "$RESULT" '.cookies[0].path' "cookie has path"
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /cookies (set + verify)"

pt_post /cookies "{
  \"tabId\": \"${TAB_ID}\",
  \"url\": \"${FIXTURES_URL}/index.html\",
  \"cookies\": [{\"name\": \"test_e2e\", \"value\": \"hello\", \"path\": \"/\"}]
}"
assert_ok "set cookie"
assert_json_eq "$RESULT" '.set' '1'

# Read back and verify
pt_get "/cookies?tabId=${TAB_ID}&url=${FIXTURES_URL}/index.html"
assert_ok "get cookies after set"
assert_json_exists "$RESULT" '.cookies[] | select(.name == "test_e2e")'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /cookies (non-existent tab → error)"

pt_get "/cookies?tabId=nonexistent_tab_12345"
assert_not_ok "rejects bad tab"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /cookies (bad JSON → error)"

pt_post_raw /cookies "{broken"
assert_http_status "400" "rejects bad JSON"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /cookies (empty array → error)"

pt_post /cookies "{
  \"tabId\": \"${TAB_ID}\",
  \"url\": \"${FIXTURES_URL}/index.html\",
  \"cookies\": []
}"
assert_http_status "400" "rejects empty cookies"

end_test
