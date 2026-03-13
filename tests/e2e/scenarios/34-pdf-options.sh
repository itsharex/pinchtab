#!/bin/bash
# 34-pdf-options.sh — PDF generation with various options
# Migrated from: tests/integration/pdf_test.go (PD1-PD12)

source "$(dirname "$0")/common.sh"

pt_post /navigate "{\"url\":\"${FIXTURES_URL}/table.html\"}"
assert_ok "navigate"
TAB_ID=$(get_tab_id)

# ─────────────────────────────────────────────────────────────────
start_test "pdf: base64 output"

pt_get "/tabs/${TAB_ID}/pdf"
assert_ok "pdf base64"
assert_json_exists "$RESULT" '.base64' "has base64 field"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: raw output"

# Raw returns binary — just check status
pinchtab GET "/tabs/${TAB_ID}/pdf?raw=true"
assert_ok "pdf raw"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: landscape"

pt_get "/tabs/${TAB_ID}/pdf?landscape=true"
assert_ok "pdf landscape"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: custom scale"

pt_get "/tabs/${TAB_ID}/pdf?scale=0.5"
assert_ok "pdf scale 0.5"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: custom paper size"

pt_get "/tabs/${TAB_ID}/pdf?paperWidth=7&paperHeight=9"
assert_ok "pdf custom paper"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: custom margins"

pt_get "/tabs/${TAB_ID}/pdf?marginTop=0.75&marginLeft=0.75&marginRight=0.75&marginBottom=0.75"
assert_ok "pdf custom margins"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: page ranges"

pt_get "/tabs/${TAB_ID}/pdf?pageRanges=1"
assert_ok "pdf page range"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: header/footer"

pt_get "/tabs/${TAB_ID}/pdf?displayHeaderFooter=true"
assert_ok "pdf header/footer"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: accessible (tagged + outline)"

pt_get "/tabs/${TAB_ID}/pdf?generateTaggedPDF=true&generateDocumentOutline=true"
assert_ok "pdf accessible"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pdf: prefer CSS page size"

pt_get "/tabs/${TAB_ID}/pdf?preferCSSPageSize=true"
assert_ok "pdf CSS page size"

end_test
