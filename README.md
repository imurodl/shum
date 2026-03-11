# shum — Safe, Recoverable Compose Upgrades for Self-Hosted Infra

`shum` is a CLI-first operations tool for remote Docker Compose fleets.

It was built to solve a concrete gap: upgrades are usually either unsafe or over-engineered.
This tool gives you a practical middle path: deterministic planning, policy gates, rollback paths, and auditable history—without wrapping the operations in abstraction layers that hide behavior.

## What the tool does

- Registers and verifies remote hosts through SSH aliases.
- Discovers and tracks compose projects with canonical naming.
- Performs preflight checks before any mutation (availability, permissions, policy readiness).
- Generates explicit upgrade plans before execution.
- Enforces per-project policy: backups, restore commands, migration warnings, and health probes.
- Executes upgrades with dry-run mode, then real run mode.
- Stores run history, artifact checksums, and failure context for post-incident review.
- Exposes structured output for scripting (`--json`) and human-friendly summary output.

## Quickstart

```bash
git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum

shum host register my-host
shum project discover my-host
shum project inspect my-host web --project-directory /srv/web --json
shum project preflight my-host web
shum project plan my-host web --json
```

## Standard Upgrade Flow

```bash
shum project policy set my-host web \
  --require-backup=true \
  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\"" \
  --restore-command "cat \"$SHUM_BACKUP_ARTIFACT\" | docker exec -i db psql -U app" \
  --health-check "http://127.0.0.1:8080/health"

shum project backup take my-host web --json
shum project upgrade my-host web --dry-run --json
shum project upgrade my-host web --json

shum project run list --host my-host --project web --json
shum project run show <run-id> --json
shum project backup list my-host web --json
```

## CLI Surfaces

- `shum host register|list|inspect`
- `shum project discover|inspect|preflight|plan|policy|backup|upgrade|run`

Run `shum --help` after installation for full command docs.

## Guarantees

`shum` aims to make upgrades predictable:

- No command run should change state without an explicit execution step.
- Plan and preflight data are surfaced before mutation.
- Upgrade artifacts are persisted for recoverability and audit.
- Every run has status transitions and summary output.
- Failures carry context through run history.

## Storage Layout

- Config: `~/.config/shum/`
- State and artifacts: `~/.cache/shum/`
  - `state.db`
  - `artifacts/`

Artifacts are intentionally local to the operator machine by default and can be moved/rotated as part of ops policy.

## Development

```bash
go test ./...
go build ./cmd/shum
go test ./test/e2e # optional; requires SHUM_E2E_SSH_ALIAS env
```

Remote integration tests are opt-in and skip automatically when SSH context is missing.

## Documentation

- [Testing Guide](./docs/testing.md)
- [Project Site](https://imurodl.me/shum/)

## Website

The project documentation site is a Vue + Vite frontend built with Bun and deployed via GitHub Pages.

## License

Apache-2.0.
