# Scripts

Development and CI scripts for PinchTab.

## Build & Run

| Script | Purpose |
|--------|---------|
| `build-dashboard.sh` | Generate TS types (tygo) + build React dashboard + copy to Go embed |
| `dev.sh` | Full build (dashboard + Go) and run |

## CI Scripts

Used by GitHub Actions workflows:

| Script | Workflow | Purpose |
|--------|----------|---------|
| `build-dashboard.sh` | `go-verify.yml` | Build dashboard before lint |

## Quality

| Script | Purpose |
|--------|---------|
| `check-docs-json.sh` | `docs-verify.yml` | Validate docs/index.json |
| `check.sh` | Local pre-push checks (mirrors CI: gofmt, vet, build, test, lint, integration) |
| `check-gosec.sh` | Security scan with gosec just to reproduce CI|
| `pre-commit` | Pre-commit hook (format, lint) |
| `doctor.sh` | Check requirements | 

## Hooks

| Script | Purpose |
|--------|---------|
| `install-hooks.sh` | Install git hooks |

## Scripts / Testing

| Script | Purpose |
|--------|---------|
| `simulate-memory-load.sh` | Memory load testing |
| `simulate-ratelimit-leak.sh` | Rate limit leak testing |
