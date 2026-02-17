# Pinchtab Test Report — autorun

**Date:** 2026-02-17 00:33 GMT  
**Branch:** autorun  
**Run type:** Even hour (test execution)

---

## Unit Tests

**Result: ✅ ALL PASS — 54/54**

| Package | Tests | Pass | Fail | Skip | Duration |
|---------|-------|------|------|------|----------|
| github.com/pinchtab/pinchtab | 54 | 54 | 0 | 0 | 0.190s |

All 54 unit tests passed covering: ref cache, config, handlers (navigate, action, evaluate, tab, snapshot, text, screenshot, cookies, stealth, fingerprint, batch actions), human simulation (mouse, typing, click), HTTP middleware (auth, CORS, logging), snapshot building/filtering/diff, stealth JS, fingerprint generation, and stealth recommendations.

---

## Integration Tests

**Result: ✅ ALL PASS — 7 pass, 1 skip**

| # | Test | Status | Duration |
|---|------|--------|----------|
| SI1 | TestStealthScriptInjected | ✅ Pass | 0.40s |
| SI2 | TestCanvasNoiseApplied | ✅ Pass | 0.39s |
| SI3 | TestFontMetricsNoise | ✅ Pass | 0.37s |
| SI4 | TestWebGLVendorSpoofed | ⏭️ Skip (headless, no GPU) | 0.36s |
| SI5 | TestPluginsPresent | ✅ Pass | 0.38s |
| SI6 | TestFingerprintRotation | ✅ Pass | 0.38s |
| SI7 | TestCDPTimezoneOverride | ✅ Pass | 0.40s |
| SI8 | TestStealthStatusEndpoint | ✅ Pass | 0.39s |

**Note:** TestFingerprintRotation logs a WARN about `hardwareConcurrency` redefinition — non-blocking, fingerprint rotation still works via CDP override.

Total integration suite duration: 3.251s

---

## TEST-PLAN.md Scenario Coverage

### Covered by automated tests

| Section | Scenarios Covered | Status |
|---------|-------------------|--------|
| 1.1 Health & Startup | H5 (auth required), H6 (auth accepted) — via TestAuthMiddleware | ✅ |
| 1.2 Navigation | N5 (invalid URL), N6 (missing URL), N7 (bad JSON) — via TestHandleNavigate_* | ✅ |
| 1.3 Snapshot | S2 (interactive filter), S3 (depth filter), S4 (text format), S6/S7 (diff), S10 (no tab) — via TestBuildSnapshot*, TestDiffSnapshot*, TestHandleSnapshot_NoTab | ✅ |
| 1.4 Text | T4 (no tab) — via TestHandleText_NoTab | ✅ |
| 1.5 Actions | A9 (unknown kind), A10 (missing kind), A11 (ref not found), A13 (no tab), A14/A15 (batch) — via TestHandleAction_*, TestHandleActions_* | ✅ |
| 1.6 Tabs | TB4 (close no tabId), TB5 (bad action) — via TestHandleTab_* | ✅ |
| 1.7 Screenshots | SS3 (no tab) — via TestHandleScreenshot_NoTab | ✅ |
| 1.8 Evaluate | E3 (missing expr), E4 (bad JSON), E5 (no tab) — via TestHandleEvaluate_* | ✅ |
| 1.9 Cookies | C3 (no tab), C4 (bad JSON), C5 (empty) — via TestHandleCookies_* | ✅ |
| 1.10 Stealth | ST1 (status), ST2 (webdriver), ST4 (plugins), ST5/ST6 (fingerprint rotate), ST7 (no tab) — via integration + unit tests | ✅ |
| 4. Integration | SI1–SI8 — via integration_test.go | ✅ (7 pass, 1 skip) |

### Not covered by automated tests (require live browser / manual)

| Section | Scenarios | Notes |
|---------|-----------|-------|
| 1.1 | H1-H4, H7 | Require running server |
| 1.2 | N1-N4, N8 | Require live navigation |
| 1.3 | S1, S5, S8, S9, S11, S12 | Require live browser |
| 1.4 | T1-T3, T5 | Require live pages |
| 1.5 | A1-A8, A12, A16, A17 | Require live DOM |
| 1.6 | TB1-TB3, TB6 | Require running browser |
| 1.7 | SS1-SS2 | Require running browser |
| 1.8 | E1-E2 | Require live page |
| 1.9 | C1-C2 | Require live browser |
| 1.10 | ST3, ST8 | Require headed browser |
| 1.11-1.12 | CF1-CF5, SP1-SP3 | Config & persistence (manual) |
| 2 | HM1-HM3 | Headed mode (manual) |
| 3 | MA1-MA8 | Multi-agent (manual) |
| 4 | SI9-SI11 | Quarterly manual checks |
| 5 | D1-D7 | Docker (manual) |
| 6 | CF6-CF8 | Extended config (manual) |
| 7 | ER1-ER8 | Edge cases (manual) |

---

## Known Issues (Section 8) Status

| # | Issue | Status | Notes |
|---|-------|--------|-------|
| K1 | Active tab tracking unreliable | 🔴 OPEN | Still requires explicit tabId workaround |
| K2 | Tab close hangs | 🟡 OPEN | Not testable in unit tests |
| K3 | x.com title always empty | 🟢 OPEN | SPA limitation |
| K4 | Chrome flag warning banner | 🟢 OPEN | Chrome 144+ deprecation |
| K5 | Stealth PRNG weak | ✅ FIXED | Confirmed via TestCanvasNoiseApplied |
| K6 | Chrome UA hardcoded | ✅ FIXED | Confirmed via TestFingerprintRotation |
| K7 | Fingerprint rotation JS-only | ✅ FIXED | CDP override confirmed in SI6 |
| K8 | Timezone hardcoded EST | ✅ FIXED | Confirmed via TestCDPTimezoneOverride |
| K9 | Stealth status hardcoded | ✅ FIXED | Confirmed via TestStealthStatusEndpoint |

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Build time | 0.139s (0.11s user, 0.42s system) |
| Binary size | 12 MB |
| Unit test duration | 0.190s |
| Integration test duration | 3.251s |
| Benchmarks | None defined (no Benchmark* functions found) |

---

## Summary

- **Unit tests:** 54/54 pass ✅
- **Integration tests:** 7/8 pass, 1 skip (WebGL headless) ✅
- **Zero failures, zero crashes**
- **Known issues K5–K9 confirmed fixed; K1–K4 remain open**
- **Release criteria (Section 9):** Unit + integration gates met. P0 blockers K1 and K2 still open.
