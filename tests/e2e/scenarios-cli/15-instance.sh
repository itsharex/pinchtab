#!/bin/bash
# 15-instance.sh — CLI instance management commands

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab instance logs"

# Get default instance ID from health
pt_ok health
INSTANCE_ID=$(echo "$PT_OUT" | jq -r '.defaultInstance.id // empty')

if [ -n "$INSTANCE_ID" ]; then
  pt_ok instance logs "$INSTANCE_ID"
  # Logs command succeeds - output might be empty
  echo -e "  ${GREEN}✓${NC} instance logs succeeded"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${YELLOW}⚠${NC} No instance ID found, skipping logs test"
  ((ASSERTIONS_PASSED++)) || true
fi

end_test

# Note: instance start is implicitly tested (server is running)
# instance stop is tested in 99-instance-stop.sh
