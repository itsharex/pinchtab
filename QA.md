# Pinchtab QA Report — 2026-02-15

## Bugs

1. **`newTab:true` broken** — `/navigate` with `newTab:true` silently ignores the flag and navigates the existing tab instead of opening a new one. Tabs get overwritten/lost.

2. **`/action` unhelpful error message** — Missing `kind` field returns `{"error":"unknown action: "}` instead of telling the user the field is missing or listing valid values (`click`, `type`, `hover`, `select`, `scroll`).

3. **`/navigate` returns empty title** — Some sites (BBC, x.com) return `"title":""` in the response. Likely a race condition — title isn't set by the time the response is sent.

4. **`/text` Google language blob** — Readability extraction on google.com includes the full language picker (all locale names) as content. Should be filtered out.

## Improvements

- **Better `/navigate` title** — Wait for `document.title` to be non-empty (with short timeout) before returning response.
- **Actionable error messages** — `/action` should list valid `kind` values in error responses.
- **`/text` content filtering** — Strip known noise patterns (language pickers, cookie banners) from readability output.
- **Tab management** — Fix `newTab:true`, add `/tab/close` endpoint if not present.

## Token Benchmarks

| Site | /snapshot | /text | Savings |
|------|----------|-------|---------|
| Google | ~2K | ~764 | 2.5x |
| GitHub | ~9.8K | ~1.2K | 7.8x |
| x.com | ~2K | ~121 | 17x |
| BBC | ~26.7K | ~3.5K | 7.7x |
| Wikipedia | ~20.5K | ~3.5K | 5.8x |
| LinkedIn | ~7.5K | ~6.1K | 1.2x |

## Sites Tested

Google, GitHub, BBC, Wikipedia, x.com, LinkedIn — all loaded fine, no bot detection, zero crashes.
