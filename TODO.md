# Pinchtab — TODO

## Completed (P0–P5)
Safety, file split, Go idioms, testability, features — all done.
38 tests passing, 0 lint issues. See git history for details.

Key deliverables: session persistence, stealth mode, ref caching, action registry,
hover/select/scroll actions, smart diff, text format, readability /text,
BridgeAPI interface, handler tests, nil guard, deprecated flag removal.

---

## Bugs & In Progress
- [ ] **Navigate timeout on some SPAs** — `navigatePage` 500ms sleep isn't enough for heavy JS pages. Consider polling `document.readyState` instead.
- [ ] **Restore navigates all tabs at once** — can overwhelm CPU/memory on startup with many tabs. Should queue or limit concurrency.
- [ ] **Screenshot base64 returns raw bytes** — `"base64": <bytes>` in JSON, should be actual base64 string encoding.

## P6: Next Up
- [ ] **Action chaining** — `POST /actions` batch multiple actions in one call (big token saver for agents)
- [ ] **`/cookies` endpoint** — read/set cookies (auth debugging)
- [ ] **LaunchAgent/systemd** — auto-start on boot
- [ ] **Config file** — `~/.pinchtab/config.json` (alternative to env vars)

## P7: Nice to Have
- [ ] **File-based output** — `?output=file` saves snapshot to disk, returns path
- [ ] **Compact format** — YAML or indented text instead of JSON
- [ ] **Docker image** — `docker run pinchtab` with bundled Chromium

## Future: Desktop App Restructure
When a second binary (desktop app via Wails) is needed, restructure to:
```
cmd/pinchtab/main.go        # CLI binary
cmd/pinchtab-app/main.go    # desktop binary
internal/server/             # current Go files move here
internal/config/
app/                         # Wails desktop layer
frontend/                    # dashboard HTML/JS
```
Until then, flat structure is correct. Don't premature-abstract.

## Not Doing
- Plugin system
- Proxy rotation / anti-detection
- Session isolation / multi-tenant
- Selenium compatibility
- React UI
- Cloud anything
- MCP protocol (HTTP is the interface)
