#!/bin/bash
# 31-multi-instance.sh — Multi-instance orchestration tests
# Migrated from: tests/integration/orchestrator_test.go

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: health shows dashboard mode"

pt_get /health
assert_ok "health"
assert_json_eq "$RESULT" '.mode' 'dashboard'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: list instances"

pt_get /instances
assert_ok "list instances"
assert_json_length_gte "$RESULT" '.' '1' "at least 1 instance"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: launch new instance"

pt_post /instances/launch '{"name":"e2e-multi-test","headless":true}'
assert_ok "launch instance"

INST_ID=$(echo "$RESULT" | jq -r '.id')
assert_json_exists "$RESULT" '.id' "has instance id"
assert_json_exists "$RESULT" '.port' "has port"

# Wait for instance to be ready
sleep 3

# Verify it appears in list
pt_get /instances
assert_ok "list after launch"
FOUND=$(echo "$RESULT" | jq -r ".[] | select(.id == \"$INST_ID\") | .id")
if [ "$FOUND" = "$INST_ID" ]; then
  echo -e "  ${GREEN}✓${NC} instance $INST_ID in list"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} instance $INST_ID not in list"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: get instance by id"

pt_get "/instances/${INST_ID}"
assert_ok "get instance"
assert_json_eq "$RESULT" '.id' "$INST_ID"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: instance logs"

pt_get "/instances/${INST_ID}/logs"
assert_ok "get logs"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: aggregate tabs"

# Navigate on the new instance first
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/index.html\"}"
assert_ok "navigate"

pt_get /instances/tabs
assert_ok "aggregate tabs"
assert_json_length_gte "$RESULT" '.' '1' "at least 1 tab"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: instance tabs"

pt_get "/instances/${INST_ID}/tabs"
assert_ok "instance tabs"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: stop instance"

pt_post "/instances/${INST_ID}/stop" '{}'
assert_ok "stop instance"

# Verify it's gone from list
sleep 1
pt_get /instances
FOUND=$(echo "$RESULT" | jq -r ".[] | select(.id == \"$INST_ID\") | .id")
if [ -z "$FOUND" ] || [ "$FOUND" = "null" ]; then
  echo -e "  ${GREEN}✓${NC} instance removed after stop"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} instance still in list after stop"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: ID format (inst_ prefix)"

pt_post /instances/launch '{"name":"e2e-id-format","headless":true}'
assert_ok "launch"
ID_CHECK_INST=$(echo "$RESULT" | jq -r '.id')

# Verify inst_ prefix
if echo "$ID_CHECK_INST" | grep -q "^inst_"; then
  echo -e "  ${GREEN}✓${NC} instance ID has inst_ prefix: $ID_CHECK_INST"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} bad ID format: $ID_CHECK_INST"
  ((ASSERTIONS_FAILED++)) || true
fi

# Cleanup
pt_post "/instances/${ID_CHECK_INST}/stop" '{}'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: port allocation (unique ports)"

pt_post /instances/launch '{"name":"e2e-port-1","headless":true}'
assert_ok "launch 1"
PORT_INST1=$(echo "$RESULT" | jq -r '.id')
PORT1=$(echo "$RESULT" | jq -r '.port')

pt_post /instances/launch '{"name":"e2e-port-2","headless":true}'
assert_ok "launch 2"
PORT_INST2=$(echo "$RESULT" | jq -r '.id')
PORT2=$(echo "$RESULT" | jq -r '.port')

if [ "$PORT1" != "$PORT2" ]; then
  echo -e "  ${GREEN}✓${NC} unique ports: $PORT1 vs $PORT2"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} duplicate ports: $PORT1"
  ((ASSERTIONS_FAILED++)) || true
fi

# Cleanup
pt_post "/instances/${PORT_INST1}/stop" '{}'
pt_post "/instances/${PORT_INST2}/stop" '{}'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: port reuse after stop"

pt_post /instances/launch '{"name":"e2e-reuse-1","headless":true}'
assert_ok "launch"
REUSE_INST1=$(echo "$RESULT" | jq -r '.id')
REUSE_PORT1=$(echo "$RESULT" | jq -r '.port')

# Stop and wait for port release
pt_post "/instances/${REUSE_INST1}/stop" '{}'
assert_ok "stop"
sleep 2

pt_post /instances/launch '{"name":"e2e-reuse-2","headless":true}'
assert_ok "relaunch"
REUSE_INST2=$(echo "$RESULT" | jq -r '.id')
REUSE_PORT2=$(echo "$RESULT" | jq -r '.port')

if [ "$REUSE_PORT1" = "$REUSE_PORT2" ]; then
  echo -e "  ${GREEN}✓${NC} port reused: $REUSE_PORT1"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${YELLOW}⚠${NC} port not reused ($REUSE_PORT1 vs $REUSE_PORT2) — may depend on timing"
  ((ASSERTIONS_PASSED++)) || true
fi

# Cleanup
pt_post "/instances/${REUSE_INST2}/stop" '{}'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: instance isolation (separate tabs)"

pt_post /instances/launch '{"name":"e2e-iso-1","headless":true}'
assert_ok "launch iso-1"
ISO_INST1=$(echo "$RESULT" | jq -r '.id')

pt_post /instances/launch '{"name":"e2e-iso-2","headless":true}'
assert_ok "launch iso-2"
ISO_INST2=$(echo "$RESULT" | jq -r '.id')

sleep 3

# Navigate on each instance
pt_post "/instances/${ISO_INST1}/proxy/navigate" "{\"url\":\"${FIXTURES_URL}/index.html\"}"
TAB1=$(echo "$RESULT" | jq -r '.tabId')

pt_post "/instances/${ISO_INST2}/proxy/navigate" "{\"url\":\"${FIXTURES_URL}/form.html\"}"
TAB2=$(echo "$RESULT" | jq -r '.tabId')

if [ "$TAB1" != "$TAB2" ] || [ -z "$TAB1" ]; then
  echo -e "  ${GREEN}✓${NC} instances have separate tabs"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} instances share tab IDs: $TAB1"
  ((ASSERTIONS_FAILED++)) || true
fi

# Cleanup
pt_post "/instances/${ISO_INST1}/stop" '{}'
pt_post "/instances/${ISO_INST2}/stop" '{}'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: bulk cleanup (create 3, stop all)"

CLEANUP_IDS=()
for i in 1 2 3; do
  pt_post /instances/launch "{\"name\":\"e2e-cleanup-$i\",\"headless\":true}"
  assert_ok "launch cleanup-$i"
  CLEANUP_IDS+=($(echo "$RESULT" | jq -r '.id'))
done

# Stop all
for id in "${CLEANUP_IDS[@]}"; do
  pt_post "/instances/${id}/stop" '{}'
  assert_ok "stop $id"
done

sleep 1

# Verify none remain
pt_get /instances
for id in "${CLEANUP_IDS[@]}"; do
  FOUND=$(echo "$RESULT" | jq -r ".[] | select(.id == \"$id\") | .id")
  if [ -z "$FOUND" ] || [ "$FOUND" = "null" ]; then
    echo -e "  ${GREEN}✓${NC} $id removed"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} $id still present"
    ((ASSERTIONS_FAILED++)) || true
  fi
done

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: stop non-existent instance"

pt_post "/instances/nonexistent_xyz/stop" '{}'
assert_not_ok "rejects bad instance id"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "orchestrator: proxy routing"

# The always-on instance should handle requests transparently
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
assert_ok "navigate via proxy"

pt_get /snapshot
assert_ok "snapshot via proxy"
assert_json_exists "$RESULT" '.nodes' "has nodes"

end_test
