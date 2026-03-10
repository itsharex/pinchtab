# E2E Test Suite

End-to-end tests for PinchTab that exercise the full stack including browser automation.

## Quick Start

### With Docker (recommended)

```bash
./pdev e2e
```

Or directly:
```bash
docker compose -f tests/e2e/docker-compose.yml up --build
```

## Architecture

```
tests/e2e/
├── docker-compose.yml      # Orchestrates all services
├── config/                 # E2E-specific PinchTab configs
│   ├── pinchtab.json
│   └── pinchtab-secure.json
├── fixtures/               # Static HTML test pages
│   ├── index.html
│   ├── form.html
│   ├── table.html
│   └── buttons.html
├── scenarios/              # Test scripts
│   ├── common.sh           # Shared utilities
│   ├── run-all.sh          # Orchestrator
│   ├── 01-health.sh
│   ├── 02-navigate.sh
│   ├── 03-snapshot.sh
│   ├── 04-tabs-api.sh      # Regression test for #207
│   ├── 05-actions.sh
│   └── 06-screenshot-pdf.sh
├── runner/                 # Test runner container
│   └── Dockerfile
└── results/                # Test output (gitignored)
```

The Docker stack reuses the repository root `Dockerfile` and mounts explicit config files with `PINCHTAB_CONFIG` instead of maintaining separate e2e-only images.

## Test Scenarios

| Script | Tests |
|--------|-------|
| 01-health | Basic connectivity, health endpoint |
| 02-navigate | Navigation, tab creation, tab listing |
| 03-snapshot | A11y tree extraction, text content |
| 04-tabs-api | Tab-scoped APIs (regression #207) |
| 05-actions | Click, type, press actions |
| 06-screenshot-pdf | Screenshot and PDF export |

## Adding Tests

1. Create a new script in `scenarios/` following the naming pattern `NN-name.sh`
2. Source `common.sh` for utilities
3. Use the assertion helpers:

```bash
#!/bin/bash
source "$(dirname "$0")/common.sh"

start_test "My test name"

# Assert HTTP status
assert_status 200 "${PINCHTAB_URL}/health"

# Assert JSON field equals value
RESULT=$(pt_get "/some/endpoint")
assert_json_eq "$RESULT" '.field' 'expected'

# Assert JSON contains substring
assert_json_contains "$RESULT" '.message' 'success'

# Assert array length
assert_json_length "$RESULT" '.items' 5

end_test
```

## Adding Fixtures

Add HTML files to `fixtures/` for testing specific scenarios:

- Forms and inputs
- Tables and data
- Dynamic content
- iframes
- File upload/download

## CI Integration

The E2E tests run automatically:
- On release tags (`v*`)
- On PRs that modify `tests/e2e/`
- Manually via workflow dispatch

## Debugging

### View container logs
```bash
docker compose -f tests/e2e/docker-compose.yml logs pinchtab
```

### Interactive shell in runner
```bash
docker compose -f tests/e2e/docker-compose.yml run runner bash
```

### Run specific scenario
```bash
docker compose -f tests/e2e/docker-compose.yml run runner /scenarios/04-tabs-api.sh
```
