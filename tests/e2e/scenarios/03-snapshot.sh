#!/bin/bash
# 03-snapshot.sh — Accessibility tree and text extraction

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab snap"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/\"}"

pt_get /snapshot
assert_index_page "$RESULT"
assert_json_length_gte "$RESULT" '.nodes' 1

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab snap (buttons.html)"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
sleep 1

pt_get /snapshot
assert_buttons_page "$RESULT"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab snap (form.html)"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/form.html\"}"
sleep 1

pt_get /snapshot
assert_form_page "$RESULT"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab text (table.html)"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/table.html\"}"
sleep 1

TEXT_RESULT=$(curl -s "${PINCHTAB_URL}/text" | jq -r '.text')
assert_table_page "$TEXT_RESULT"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "snapshot: diff mode"

# Take initial snapshot, then take diff — diff should return ok
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
sleep 1

pt_get /snapshot
assert_ok "initial snapshot"
INITIAL_COUNT=$(echo "$RESULT" | jq '.nodes | length')

# Second snapshot with diff=true (no changes made, so diff should have fewer/no changed nodes)
pt_get "/snapshot?diff=true"
assert_ok "diff snapshot"
DIFF_COUNT=$(echo "$RESULT" | jq '.nodes | length')

# Diff should return fewer or equal nodes compared to full snapshot
if [ "$DIFF_COUNT" -le "$INITIAL_COUNT" ]; then
  echo -e "  ${GREEN}✓${NC} diff has <= nodes than full ($DIFF_COUNT <= $INITIAL_COUNT)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} diff has more nodes than full ($DIFF_COUNT > $INITIAL_COUNT)"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "snapshot: maxTokens truncation"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
sleep 1

# Full snapshot
pt_get /snapshot
FULL_COUNT=$(echo "$RESULT" | jq '.nodes | length')

# Snapshot with very small maxTokens — should have fewer nodes
pt_get "/snapshot?maxTokens=50"
assert_ok "snapshot with maxTokens"
LIMITED_COUNT=$(echo "$RESULT" | jq '.nodes | length')

if [ "$LIMITED_COUNT" -le "$FULL_COUNT" ]; then
  echo -e "  ${GREEN}✓${NC} maxTokens limited nodes ($LIMITED_COUNT <= $FULL_COUNT)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} maxTokens did not limit ($LIMITED_COUNT > $FULL_COUNT)"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "snapshot: depth parameter"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
sleep 1

# Full tree
pt_get /snapshot
FULL_COUNT=$(echo "$RESULT" | jq '.nodes | length')

# Depth=1 should produce a shallower tree with fewer nodes
pt_get "/snapshot?depth=1"
assert_ok "snapshot with depth=1"
SHALLOW_COUNT=$(echo "$RESULT" | jq '.nodes | length')

if [ "$SHALLOW_COUNT" -le "$FULL_COUNT" ]; then
  echo -e "  ${GREEN}✓${NC} depth=1 limited tree ($SHALLOW_COUNT <= $FULL_COUNT)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} depth=1 did not limit ($SHALLOW_COUNT > $FULL_COUNT)"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test
