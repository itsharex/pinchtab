#!/bin/bash
# 33-upload-errors.sh — Upload error cases
# Migrated from: tests/integration/upload_test.go (UP6-UP9, UP11)

source "$(dirname "$0")/common.sh"

pt_post /navigate "{\"url\":\"${FIXTURES_URL}/upload.html\"}"
assert_ok "navigate"
sleep 1

# ─────────────────────────────────────────────────────────────────
start_test "upload: default selector"

FILE_CONTENT="data:text/plain;base64,SGVsbG8="
pt_post /upload "{\"files\":[\"${FILE_CONTENT}\"]}"
assert_ok "upload with default selector"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "upload: invalid selector → error"

pt_post /upload '{"selector":"#nonexistent","files":["data:text/plain;base64,SGVsbG8="]}'
assert_not_ok "rejects invalid selector"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "upload: missing files → error"

pt_post /upload '{"selector":"#single-file"}'
assert_not_ok "rejects missing files"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "upload: bad JSON → error"

pt_post_raw /upload "{broken"
assert_http_status "400" "rejects bad JSON"

end_test
