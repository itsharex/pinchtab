# Pinchtab — TODO

**Philosophy**: 12MB binary. HTTP API. Minimal deps. Internal tool, not a product.

---

## DONE

Core HTTP API (16 endpoints), session persistence, ref caching, action registry,
smart diff, readability `/text`, config file, Dockerfile, YAML/file output,
stealth suite, human interaction (bezier mouse, typing sim), fingerprint rotation,
image/media blocking, stealth injection on all tabs, K1-K11 all fixed,
multi-agent concurrency verified (MA1-MA8), token optimization (`maxTokens`,
`selector`, `format=compact`), Dockerfile env vars (`CHROME_BINARY`/`CHROME_FLAGS`
now consumed by Go). **92 unit tests + ~100 integration, 28.9% coverage.**

---

## Open

### ~~P0-P2~~ — DONE
P0 (K10 profile hang), P1 (token optimization: maxTokens/selector/compact),
P2 (K11 file path, blockImages on CreateTab) — all resolved.

### P3: Multi-Agent
- [ ] **Tab locking** — `POST /tab/lock`, `POST /tab/unlock` with timeout-based deadlock prevention.
- [ ] **Tab ownership tracking** — Show owner in `/tabs` response.

### P4: Quality of Life
- [ ] **Headed mode testing** — Run Section 2 tests to validate non-headless.
- [ ] **Ad blocking** — Basic tracker blocking for cleaner snapshots.
- [ ] **CSS animation disabling** — Faster page loads, more consistent snapshots.
- [ ] **Randomized window sizes** — Avoid automation fingerprint.

### Minor
- [ ] **humanType global rand** — Accept `*rand.Rand` for reproducible tests.

---

## Not Doing
Desktop app, plugin system, proxy rotation, SaaS, Selenium compat, MCP protocol,
cloud anything, distributed clusters, workflow orchestration.
