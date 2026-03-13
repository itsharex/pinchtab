#!/bin/bash
# 27-metrics.sh — Metrics API tests
# Migrated from: tests/integration/metrics_test.go

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "GET /instances/metrics"

pt_post /navigate "{\"url\":\"${FIXTURES_URL}/index.html\"}"
assert_ok "navigate"

pt_get "/instances/metrics"
assert_ok "get instance metrics"
assert_json_exists "$RESULT" '.[0].instanceId'
assert_json_exists "$RESULT" '.[0].jsHeapUsedMB'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /tabs/{id}/metrics"

pt_post /navigate "{\"url\":\"${FIXTURES_URL}/form.html\"}"
assert_ok "navigate"
TAB_ID=$(get_tab_id)

pt_get "/tabs/${TAB_ID}/metrics"
assert_ok "get tab metrics"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /tabs/{invalid}/metrics → error"

pt_get "/tabs/invalid_tab_id/metrics"
assert_not_ok "rejects invalid tab"

end_test
