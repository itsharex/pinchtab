#!/bin/bash
# 09-health.sh — CLI health and status commands

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab health"

pt_ok health
assert_output_json
assert_output_contains "status" "returns status field"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab instances"

pt_ok instances
assert_output_json
# Output is an array like [{id:..., status:...}], check for instance properties
assert_output_contains "id" "returns instance id"
assert_output_contains "status" "returns instance status"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab profiles"

pt_ok profiles
# Output is array, might be truncated in PT_OUT but command succeeds
# Just verify exit code was 0

end_test
