#!/bin/bash
# 02-navigate.sh — Navigation and tab management

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab nav <url>"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/\"}"
assert_json_contains "$RESULT" '.title' 'E2E Test'
assert_json_contains "$RESULT" '.url' 'fixtures'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab nav (multiple pages)"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/form.html\"}"
assert_json_contains "$RESULT" '.title' 'Form'

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/table.html\"}"
assert_json_contains "$RESULT" '.title' 'Table'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs"

assert_tab_count_gte 2

end_test

# ─────────────────────────────────────────────────────────────────
start_test "navigate: blockImages flag"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\",\"blockImages\":true}"
assert_ok "navigate with blockImages"
assert_json_contains "$RESULT" '.url' 'index.html'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "navigate: blockAds flag"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\",\"blockAds\":true}"
assert_ok "navigate with blockAds"
assert_json_contains "$RESULT" '.url' 'index.html'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "navigate: timeout parameter"

# Navigate with a generous timeout (in seconds) — should succeed
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\",\"timeout\":10}"
assert_ok "navigate with custom timeout"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "navigate: waitSelector parameter"

# waitSelector waits for a CSS selector after navigation
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/buttons.html\",\"waitSelector\":\"button\"}"
assert_ok "navigate with waitSelector"
assert_json_contains "$RESULT" '.title' 'Button'

end_test
