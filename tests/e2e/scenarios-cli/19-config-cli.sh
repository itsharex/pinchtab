#!/bin/bash
# 19-config-cli.sh — Config CLI subcommands (init, show, path, set, patch, validate, get)

source "$(dirname "$0")/common.sh"

# ═══════════════════════════════════════════════════════════════════
# CONFIG INIT
# ═══════════════════════════════════════════════════════════════════

start_test "config init creates valid config"

TMPDIR=$(mktemp -d)
PINCHTAB_CONFIG="$TMPDIR/config.json" HOME="$TMPDIR" pt_ok config init
# Verify file was created and is valid JSON
if [ -f "$TMPDIR/.pinchtab/config.json" ] || [ -f "$TMPDIR/config.json" ]; then
  echo -e "  ${GREEN}✓${NC} config file created"
  ((ASSERTIONS_PASSED++)) || true
  # Check it has expected sections (read from whichever file exists)
  CFG_FILE="$TMPDIR/config.json"
  [ -f "$CFG_FILE" ] || CFG_FILE="$TMPDIR/.pinchtab/config.json"
  if jq -e '.server' "$CFG_FILE" >/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} has server section"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} missing server section"
    ((ASSERTIONS_FAILED++)) || true
  fi
  if jq -e '.browser' "$CFG_FILE" >/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} has browser section"
    ((ASSERTIONS_PASSED++)) || true
  else
    echo -e "  ${RED}✗${NC} missing browser section"
    ((ASSERTIONS_FAILED++)) || true
  fi
else
  echo -e "  ${RED}✗${NC} config file not created"
  ((ASSERTIONS_FAILED++)) || true
fi
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG SHOW
# ═══════════════════════════════════════════════════════════════════

start_test "config show displays config with env override"

TMPDIR=$(mktemp -d)
PINCHTAB_CONFIG="$TMPDIR/nonexistent.json" PINCHTAB_PORT=9999 pt_ok config show
assert_output_contains "9999" "shows port from env"
assert_output_contains "Server" "has Server section header"
assert_output_contains "Browser" "has Browser section header"
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG PATH
# ═══════════════════════════════════════════════════════════════════

start_test "config path outputs config file path"

TMPDIR=$(mktemp -d)
EXPECTED_PATH="$TMPDIR/custom-config.json"
PINCHTAB_CONFIG="$EXPECTED_PATH" pt_ok config path

if echo "$PT_OUT" | grep -qF "$EXPECTED_PATH"; then
  echo -e "  ${GREEN}✓${NC} path matches expected"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected $EXPECTED_PATH, got: $PT_OUT"
  ((ASSERTIONS_FAILED++)) || true
fi
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG SET
# ═══════════════════════════════════════════════════════════════════

start_test "config set updates a value"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
PINCHTAB_CONFIG="$CFG" HOME="$TMPDIR" pt_ok config init
PINCHTAB_CONFIG="$CFG" pt_ok config set server.port 8080
assert_output_contains "Set server.port = 8080" "success message"

# Verify file was actually updated
ACTUAL=$(jq -r '.server.port' "$CFG")
if [ "$ACTUAL" = "8080" ]; then
  echo -e "  ${GREEN}✓${NC} file contains port 8080"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected port 8080 in file, got $ACTUAL"
  ((ASSERTIONS_FAILED++)) || true
fi
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG PATCH
# ═══════════════════════════════════════════════════════════════════

start_test "config patch merges JSON"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
PINCHTAB_CONFIG="$CFG" HOME="$TMPDIR" pt_ok config init
PINCHTAB_CONFIG="$CFG" pt_ok config patch '{"server":{"port":"7777"},"instanceDefaults":{"maxTabs":100}}'

# Verify values
PORT_VAL=$(jq -r '.server.port' "$CFG")
TABS_VAL=$(jq -r '.instanceDefaults.maxTabs' "$CFG")
if [ "$PORT_VAL" = "7777" ]; then
  echo -e "  ${GREEN}✓${NC} port set to 7777"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected port 7777, got $PORT_VAL"
  ((ASSERTIONS_FAILED++)) || true
fi
if [ "$TABS_VAL" = "100" ]; then
  echo -e "  ${GREEN}✓${NC} maxTabs set to 100"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected maxTabs 100, got $TABS_VAL"
  ((ASSERTIONS_FAILED++)) || true
fi
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG VALIDATE (valid)
# ═══════════════════════════════════════════════════════════════════

start_test "config validate accepts valid config"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
cat > "$CFG" <<'EOF'
{
  "server": {"port": "9867"},
  "instanceDefaults": {"stealthLevel": "light", "tabEvictionPolicy": "reject"},
  "multiInstance": {"strategy": "simple", "allocationPolicy": "fcfs"}
}
EOF
PINCHTAB_CONFIG="$CFG" pt_ok config validate
assert_output_contains "valid" "reports valid"
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG VALIDATE (invalid)
# ═══════════════════════════════════════════════════════════════════

start_test "config validate rejects invalid config"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
cat > "$CFG" <<'EOF'
{
  "server": {"port": "99999"},
  "instanceDefaults": {"stealthLevel": "superstealth"},
  "multiInstance": {"strategy": "magical"}
}
EOF
PINCHTAB_CONFIG="$CFG" pt_fail config validate
assert_output_contains "error" "reports error"
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG GET
# ═══════════════════════════════════════════════════════════════════

start_test "config get retrieves a value"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
PINCHTAB_CONFIG="$CFG" HOME="$TMPDIR" pt_ok config init
PINCHTAB_CONFIG="$CFG" pt_ok config set server.port 7654
PINCHTAB_CONFIG="$CFG" pt_ok config get server.port

if echo "$PT_OUT" | grep -q "7654"; then
  echo -e "  ${GREEN}✓${NC} got value 7654"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected 7654, got: $PT_OUT"
  ((ASSERTIONS_FAILED++)) || true
fi
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG GET (unknown path)
# ═══════════════════════════════════════════════════════════════════

start_test "config get fails for unknown path"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
PINCHTAB_CONFIG="$CFG" pt_fail config get unknown.field
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# CONFIG GET (slice field)
# ═══════════════════════════════════════════════════════════════════

start_test "config get returns slice as comma-separated"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
PINCHTAB_CONFIG="$CFG" HOME="$TMPDIR" pt_ok config init
PINCHTAB_CONFIG="$CFG" pt_ok config set security.attach.allowHosts "127.0.0.1,localhost"
PINCHTAB_CONFIG="$CFG" pt_ok config get security.attach.allowHosts

if echo "$PT_OUT" | grep -q "127.0.0.1,localhost"; then
  echo -e "  ${GREEN}✓${NC} got comma-separated value"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected '127.0.0.1,localhost', got: $PT_OUT"
  ((ASSERTIONS_FAILED++)) || true
fi
rm -rf "$TMPDIR"

end_test

# ═══════════════════════════════════════════════════════════════════
# LEGACY CONFIG DETECTION
# ═══════════════════════════════════════════════════════════════════

start_test "config show loads legacy flat config"

TMPDIR=$(mktemp -d)
CFG="$TMPDIR/config.json"
cat > "$CFG" <<'EOF'
{
  "port": "8765",
  "headless": true,
  "maxTabs": 30
}
EOF
PINCHTAB_CONFIG="$CFG" pt_ok config show
assert_output_contains "8765" "shows port from legacy config"
rm -rf "$TMPDIR"

end_test
