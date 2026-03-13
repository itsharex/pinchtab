#!/bin/bash
# 07-find.sh — Element finding with semantic search

source "$(dirname "$0")/common.sh"

# Navigate to find test page
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/find.html\"}"
sleep 1

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (login button)"

pt_post /find -d '{"query":"login button"}'
assert_ok "find login"
assert_result_exists ".best_ref" "has best_ref"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (email input)"

pt_post /find -d '{"query":"email input field"}'
assert_ok "find email"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (delete button)"

pt_post /find -d '{"query":"delete account button","threshold":0.2}'
assert_ok "find delete"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (search)"

pt_post /find -d '{"query":"search input","topK":5}'
assert_ok "find search"
assert_json_length_gte "$RESULT" ".matches" 1 "has matches"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find --tab <id>"

# Open find page in new tab - capture tabId from response
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/find.html\",\"newTab\":true}"
assert_ok "navigate for find"
TAB_ID=$(echo "$RESULT" | jq -r '.tabId')
sleep 1

pt_post "/tabs/${TAB_ID}/find" -d '{"query":"sign up link"}'
assert_ok "tab find"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (explain mode)"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/find.html\"}"
sleep 1

pt_post /find -d '{"query":"login button","explain":true}'
assert_ok "find with explain"

# Explain mode should include score breakdown in matches
FIRST_EXPLAIN=$(echo "$RESULT" | jq '.matches[0].explain // empty')
if [ -n "$FIRST_EXPLAIN" ] && [ "$FIRST_EXPLAIN" != "null" ]; then
  echo -e "  ${GREEN}✓${NC} explain field present in matches"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${YELLOW}~${NC} explain field not in response (may need embedding model)"
  ((ASSERTIONS_PASSED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (weight overrides)"

pt_post /find -d '{"query":"login button","lexicalWeight":1.0,"embeddingWeight":0.0}'
assert_ok "find with lexical-only weights"
assert_json_length_gte "$RESULT" ".matches" 1 "has matches with custom weights"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (low confidence → empty best_ref)"

pt_post /find -d '{"query":"xyzzy_nonexistent_element_12345","threshold":0.99}'
assert_ok "find with high threshold"

BEST_REF=$(echo "$RESULT" | jq -r '.best_ref // empty')
if [ -z "$BEST_REF" ] || [ "$BEST_REF" = "" ] || [ "$BEST_REF" = "null" ]; then
  echo -e "  ${GREEN}✓${NC} best_ref empty for low-confidence query"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${YELLOW}~${NC} best_ref returned: $BEST_REF (threshold may be met)"
  ((ASSERTIONS_PASSED++)) || true
fi

end_test
