#!/bin/bash
# 38-tab-locking.sh — Tab lock/unlock functionality

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "tab lock: lock and unlock"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\"}"
TAB_ID=$(get_tab_id)

# Lock the tab
pt_post /tab/lock -d "{\"tabId\":\"${TAB_ID}\",\"owner\":\"test-agent\"}"
assert_ok "lock tab"
assert_json_eq "$RESULT" '.locked' 'true' "tab is locked"
assert_json_eq "$RESULT" '.owner' 'test-agent' "owner matches"

# Unlock the tab
pt_post /tab/unlock -d "{\"tabId\":\"${TAB_ID}\",\"owner\":\"test-agent\"}"
assert_ok "unlock tab"
assert_json_eq "$RESULT" '.unlocked' 'true' "tab is unlocked"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tab lock: wrong owner cannot unlock"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\"}"
TAB_ID=$(get_tab_id)

# Lock with one owner
pt_post /tab/lock -d "{\"tabId\":\"${TAB_ID}\",\"owner\":\"agent-a\"}"
assert_ok "lock tab"

# Try unlock with different owner — should fail
pt_post /tab/unlock -d "{\"tabId\":\"${TAB_ID}\",\"owner\":\"agent-b\"}"
assert_not_ok "wrong owner rejected"

# Clean up — unlock with correct owner
pt_post /tab/unlock -d "{\"tabId\":\"${TAB_ID}\",\"owner\":\"agent-a\"}"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tab lock: lock with timeoutSec"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\"}"
TAB_ID=$(get_tab_id)

pt_post /tab/lock -d "{\"tabId\":\"${TAB_ID}\",\"owner\":\"test-ttl\",\"timeoutSec\":60}"
assert_ok "lock with timeout"
assert_json_exists "$RESULT" '.expiresAt' "has expiration time"

# Clean up
pt_post /tab/unlock -d "{\"tabId\":\"${TAB_ID}\",\"owner\":\"test-ttl\"}"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tab lock: path-based lock (POST /tabs/{id}/lock)"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\"}"
TAB_ID=$(get_tab_id)

pt_post "/tabs/${TAB_ID}/lock" -d "{\"owner\":\"path-agent\"}"
assert_ok "path-based lock"
assert_json_eq "$RESULT" '.locked' 'true'

# Unlock via path
pt_post "/tabs/${TAB_ID}/unlock" -d "{\"owner\":\"path-agent\"}"
assert_ok "path-based unlock"

end_test
