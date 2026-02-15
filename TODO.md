# Pinchtab — TODO

## Done ✅
- [x] Session persistence — save/restore tabs on shutdown/startup
- [x] Graceful shutdown — save state on SIGTERM/SIGINT
- [x] Launch helper — self-launches Chrome, no manual CDP flags
- [x] Snapshot pruning — `?filter=interactive`, `?depth=N`
- [x] `/text` endpoint — body text extraction
- [x] Ref resolution via DOM.resolveNode + backendDOMNodeId
- [x] Stealth mode — webdriver hidden, UA spoofed, automation flags removed
- [x] Tab registry — contexts survive across requests
- [x] Ref stability — snapshot caches ref→nodeID mapping per tab, actions use cached refs
- [x] Action timeouts — 15s default, prevents hung pages blocking handlers
- [x] Tab cleanup — background goroutine removes stale entries every 30s
- [x] Tab restore on startup — loadState() called, tabs reopened

## Done ✅ (P0)
- [x] `http.MaxBytesReader` (1MB) on all POST handlers
- [x] `r.Context()` propagated — `cancelOnClientDone` cancels CDP on client disconnect
- [x] Graceful shutdown with `context.WithTimeout` (10s)
- [x] `cleanStaleTabs` accepts `context.Context` — no goroutine leak
- [x] `tabContext` lock — RLock fast path, Lock only on miss, double-check pattern
- [x] Errors wrapped with `%w` consistently
- [x] All ignored errors handled (`os.MkdirAll`, `json.Encode`, `os.WriteFile`, `json.MarshalIndent`)

## P1: Code Structure (File Split)
Split `main.go` (1045 lines) into single-package files:
- [ ] `config.go` — env vars, constants, magic strings
- [ ] `bridge.go` — Bridge struct, tabContext, cleanStaleTabs, tab registry
- [ ] `handlers.go` — HTTP handlers
- [ ] `snapshot.go` — a11y types, tree parsing, snapshot logic
- [ ] `cdp.go` — clickByNodeID, typeByNodeID, listTargets, resolveNode
- [ ] `state.go` — save/restore, markCleanExit
- [ ] `middleware.go` — auth, CORS
- [ ] `server.go` — main(), route setup, signal handling

## P2: Go Idioms & Clean Code
- [ ] Eliminate global `bridge` var — pass as receiver or inject into handlers
- [ ] Action registry `map[string]ActionFunc` instead of switch in handleAction
- [ ] Add `scrollIntoViewIfNeeded` before click/focus actions
- [ ] Constants for magic strings (`"page"`, `"interactive"`, action kinds)
- [ ] Use `slog` for structured logging (stdlib since Go 1.21)
- [ ] Add godoc comments on exported types (Bridge, A11yNode, TabState)
- [ ] `//go:embed` for stealth JS script
- [ ] Group Chrome opts by concern (stealth, perf, UI) with comments

## P3: Testability
- [ ] Extract `Browser` interface (navigate, screenshot, evaluate)
- [ ] Extract `TabManager` interface (get, create, close, list)
- [ ] Add handler tests using `httptest` + mock interfaces
- [ ] Add snapshot unit tests — a11y tree filtering/parsing

## P4: Features
- [ ] **Smart diff** — `?diff=true` returns only changes since last snapshot. Massive token savings on multi-step tasks
- [ ] **Scroll actions** — scroll to element, scroll by amount
- [ ] **Wait for navigation** — after click, wait for page load before returning
- [ ] **Better /text** — Readability-style extraction instead of raw innerText

## P5: Nice to Have
- [ ] **File-based output** — `?output=file` saves snapshot to disk, returns path (Playwright CLI approach)
- [ ] **Compact format** — YAML or indented text instead of JSON
- [ ] **Action chaining** — `POST /actions` batch multiple actions in one call
- [ ] **Docker image** — `docker run pinchtab` with bundled Chromium
- [ ] **Config file** — `~/.pinchtab/config.json`
- [ ] **LaunchAgent/systemd** — auto-start on boot

## Not Doing
- Plugin system
- Proxy rotation / anti-detection
- Session isolation / multi-tenant
- Selenium compatibility
- React UI
- Cloud anything
- MCP protocol (HTTP is the interface)
