# Pinchtab QA Report — 2026-02-15

**Build:** commit 4fc2a3e (latest main)  
**Testers:** Bosch, Mario

---

## Bugs

### 🔴 P0 — Active Tab Not Tracked After Navigate (Bosch)

`/navigate` opens a new tab (or reuses one unpredictably) but `/snapshot` and `/text` don't follow — they return data from a stale tab. This is the main blocker for reliable automation.

**Repro:**
1. Fresh start — Pinchtab opens with `about:blank`
2. `/navigate` to `https://example.com` → works (reuses initial tab)
3. `/navigate` to `https://github.com/luigi-agosti/pinchtab` → opens new tab
4. `/snapshot` → returns `about:blank` or the previous tab, not GitHub

**Expected:** After `/navigate`, subsequent `/snapshot`/`/text` should return data from the navigated tab.

### 🔴 P0 — Invalid JSON in Snapshot/Text (Bosch)

Both `/snapshot` and `/text` produce invalid JSON on pages with certain content. Unescaped control characters break standard JSON parsers.

**Affected sites:** Yahoo Finance, StackOverflow (confirmed), likely others  
**Error:** `json.JSONDecodeError: Invalid control character at: line 1 column N`  
**Workaround:** `json.loads(raw, strict=False)` in Python  
**Fix:** Escape control characters (U+0000–U+001F) in all string values before JSON serialization.

### 🟡 P1 — `newTab:true` Broken (Mario)

`/navigate` with `newTab:true` silently ignores the flag and navigates the existing tab instead of opening a new one. Tabs get overwritten/lost.

### 🟡 P1 — Tab Close Returns 400 (Bosch)

`POST /tab {"action":"close","tabId":"..."}` returns `{"error":"tabId required"}`.  
Tried both `"id"` and `"tabId"` as field names — neither works.

**Impact:** Can't clean up tabs programmatically, leading to tab accumulation.

### 🟡 P1 — No Tab Switch/Focus API (Bosch)

No way to manually set the active tab. When active tab tracking drifts, there's no recovery path.

**Suggestion:** Add `POST /tab {"action":"switch","tabId":"..."}` or similar.

### 🟡 P1 — `/action` Unhelpful Error Message (Mario)

Missing `kind` field returns `{"error":"unknown action: "}` instead of telling the user the field is missing or listing valid values (`click`, `type`, `hover`, `select`, `scroll`).

### 🟢 P2 — `/navigate` Returns Empty Title (Mario)

Some sites (BBC, x.com) return `"title":""` in the response. Likely a race condition — title isn't set by the time the response is sent.

### 🟢 P2 — `/text` Google Language Blob (Mario)

Readability extraction on google.com includes the full language picker (all locale names) as content. Should be filtered out.

### 🟢 P2 — Chrome Flag Warning (Bosch)

```
You are using an unsupported command-line flag: --disable-blink-features=AutomationControlled.
Stability and security will suffer.
```

Minor — cosmetic warning but could affect stability per Chrome's own messaging.

---

## Improvements

- **Better `/navigate` title** — Wait for `document.title` to be non-empty (with short timeout) before returning response.
- **Actionable error messages** — `/action` should list valid `kind` values in error responses.
- **`/text` content filtering** — Strip known noise patterns (language pickers, cookie banners) from readability output.
- **Tab management** — Fix `newTab:true`, fix tab close, add tab switch endpoint.
- **Compact snapshot format** — Consider an aria-tree-style text output (indented, role + name per line) instead of/alongside verbose JSON. Could cut snapshot tokens by 3–4×.

---

## Performance — Token Usage

### Pinchtab `/snapshot` vs `/text` (Mario)

| Site | /snapshot tokens | /text tokens | Savings |
|------|-----------------|-------------|---------|
| Google | ~2K | ~764 | 2.5× |
| GitHub | ~9.8K | ~1.2K | 7.8× |
| x.com | ~2K | ~121 | 17× |
| BBC | ~26.7K | ~3.5K | 7.7× |
| Wikipedia | ~20.5K | ~3.5K | 5.8× |
| LinkedIn | ~7.5K | ~6.1K | 1.2× |

### Pinchtab vs OpenClaw Browser (Bosch)

| Site | Pinchtab snapshot | Pinchtab /text | OpenClaw aria tree |
|------|-------------------|---------------|-------------------|
| Yahoo Finance | ~16K tokens | ~1.4K tokens | ~3.5K tokens |
| Google Finance | ~12K tokens | ~1.1K tokens | ~3.7K tokens |
| Hacker News | ~24K tokens | ~875 tokens | — |

### Key Findings

- **Pinchtab snapshots are 3–4× larger** than OpenClaw aria trees (verbose JSON with `ref`, `role`, `name`, `depth`, `nodeId` per node)
- **Pinchtab `/text` is the most token-efficient format** (~1K tokens for complex finance pages) — great for content extraction
- **OpenClaw aria tree** is the best balance for interactive browsing (~3.5K tokens, structured + compact)
- **`/text` is a real strength** — already very efficient, should be the primary endpoint for read-only tasks

---

## Retest Results (post-fix, 2026-02-15)

| Bug | Status | Notes |
|-----|--------|-------|
| 🟡 `newTab:true` broken | ✅ FIXED | Creates new CDP tab, returns new tabId |
| 🟡 `/action` unhelpful error | ✅ FIXED | Lists valid `kind` values |
| 🟢 `/navigate` empty title | ✅ PARTIAL | BBC works ("BBC - Home"), x.com still empty (SPA >2s) |
| 🟢 `/text` Google blob | ✅ FIXED | Tokens dropped ~764 → ~143 |
| 🔴 Active tab tracking | ❌ STILL BROKEN | Navigate→read returns stale tab content |

**Active tab tracking remains the critical P0.** After navigating to x.com, `/text` returned Google's content. Sequential navigate→read is unreliable without explicit `tabId` targeting.

---

## Sites Tested

**Mario:** Google, GitHub, BBC, Wikipedia, x.com, LinkedIn  
**Bosch:** HN, Example.com, Yahoo Finance, Google Finance, Bloomberg, StackOverflow

All loaded fine, no bot detection, zero crashes. ✅

---

## What Works Well

- ✅ `/navigate` — fast, returns title+URL correctly
- ✅ `/snapshot` — comprehensive a11y tree when on the right tab
- ✅ `/snapshot?filter=interactive` — properly filters to actionable elements
- ✅ `/text` — clean, compact content extraction
- ✅ `/action` with `click` — reliable (tested cookie consent acceptance)
- ✅ `/tabs` — accurate tab listing
- ✅ Startup is fast (~3 seconds to ready)
- ✅ Headless mode works well
- ✅ No bot detection on any tested sites
