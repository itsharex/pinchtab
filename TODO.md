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

## P1: Daily Driver Quality
- [ ] **Smart diff** — `?diff=true` returns only changes since last snapshot. Massive token savings on multi-step tasks
- [ ] **Scroll actions** — scroll to element, scroll by amount
- [ ] **Wait for navigation** — after click, wait for page load before returning
- [ ] **Better /text** — Readability-style extraction instead of raw innerText

## P2: Nice to Have
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
