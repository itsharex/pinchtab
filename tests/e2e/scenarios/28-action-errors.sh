#!/bin/bash
# 28-action-errors.sh — Action error handling and batch operations
# Migrated from: tests/integration/actions_test.go (error cases)

source "$(dirname "$0")/common.sh"

# Navigate first
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
assert_ok "navigate"

# ─────────────────────────────────────────────────────────────────
start_test "action: unknown kind → error"

pt_post /action '{"kind":"explode","ref":"e0"}'
assert_not_ok "rejects unknown kind"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "action: missing kind → error"

pt_post /action '{"ref":"e0"}'
assert_http_status "400" "rejects missing kind"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "action: ref not found → error"

pt_post /action '{"kind":"click","ref":"e999"}'
assert_not_ok "rejects missing ref"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "action: batch operations"

pt_post /actions '{"actions":[{"kind":"click","ref":"e4"},{"kind":"click","ref":"e5"}]}'
assert_ok "batch actions"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "action: empty batch → error"

pt_post /actions '{"actions":[]}'
assert_not_ok "rejects empty batch"

end_test
