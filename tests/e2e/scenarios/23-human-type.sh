#!/bin/bash
# 23-human-type.sh — humanType and humanClick actions

source "$(dirname "$0")/common.sh"
source "$(dirname "$0")/helpers/snapshot.sh"
source "$(dirname "$0")/helpers/actions.sh"

# ─────────────────────────────────────────────────────────────────
start_test "humanClick: click input by ref"

navigate_fixture "human-type.html"
fresh_snapshot

require_ref "textbox" "Email" EMAIL_REF && \
  action_human_click "$EMAIL_REF"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "humanType: type into input by ref"

fresh_snapshot
require_ref "textbox" "Email" EMAIL_REF && {
  action_human_type "$EMAIL_REF" "test@example.com"

  fresh_snapshot
  assert_value "textbox" "Email" "test@example.com"
}

end_test

# ─────────────────────────────────────────────────────────────────
start_test "humanType: type into second input by ref"

fresh_snapshot
require_ref "textbox" "Name" NAME_REF && {
  action_human_type "$NAME_REF" "John Doe"

  fresh_snapshot
  assert_value "textbox" "Name" "John Doe"
}

end_test

# ─────────────────────────────────────────────────────────────────
start_test "humanType: type with CSS selector"

action_human_type_selector "#name" " Jr."

end_test
