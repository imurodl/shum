# SHUM - Safe, Host-Aware Upgrade Management for self-hosted Linux fleets

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/imurodl/shum/.github/workflows/deploy-site.yml?label=site)
![GitHub last commit](https://img.shields.io/github/last-commit/imurodl/shum)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

**Website:** https://imurodl.me/shum/  
**Repository:** https://github.com/imurodl/shum  
**License:** Apache-2.0

`shum` is an operations CLI that makes remote Docker Compose upgrades explicit, policy-gated, and recoverable.

## What it solves

- Remote upgrades are often risky and inconsistently documented.
- Teams lack a canonical flow from discovery to rollback.
- Backup and verification are frequently informal or forgotten.
- Run history is hard to audit under pressure.

`shum` gives you a single command surface to solve these problems without wrapping your stack in extra abstraction.

## Core guarantees

- **Trust-first host model**: register once, then target stable SSH aliases.
- **Deterministic upgrade flow**: discover → preflight → plan → policy → execute.
- **Policy gates**: mandatory backup and migration warning controls.
- **Recovery**: persisted backup artifacts and restore pathways.
- **Auditability**: structured run history with status transitions and summaries.

## Architecture

- `host`: register/list/inspect remote hosts and identity metadata.
- `project discover`: remote compose project discovery.
- `project inspect`: analyze compose risk surfaces and effective configuration.
- `project preflight`: verify readiness signals before mutation.
- `project plan`: compute planned image/config changes before execution.
- `project policy`: store per-project safety controls.
- `project backup`: create/list/restore backup artifacts.
- `project upgrade`: dry-run then execute upgrades with optional probes.
- `project run`: inspect run history and details.

## Install

```bash
git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum
```

## Quick start

```bash
shum host register prod
shum project discover prod
shum project inspect prod web --project-directory /srv/web --json
shum project preflight prod web
shum project plan prod web --json
```

## Standard flow

```bash
shum project policy set prod web \
  --require-backup=true \
  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\"" \
  --restore-command "cat \"$SHUM_BACKUP_ARTIFACT\" | docker exec -i db psql -U app" \
  --health-check "http://127.0.0.1:8080/health"

shum project backup take prod web --json
shum project upgrade prod web --dry-run --json
shum project upgrade prod web --json

shum project run list --host prod --project web --json
shum project run show <run-id> --json
```

## Development

```bash
go test ./...
go build ./cmd/shum
go test ./test/e2e # optional; requires SHUM_E2E_SSH_ALIAS
```

Remote integration tests are optional and skip automatically if SSH context is unavailable.

## Testing

- Local unit and integration coverage in `internal/*` and `test/integration`.
- Optional remote suite in `test/e2e`.
- CLI outputs are intentionally explicit for observability in scripts.

## Storage

- Config: `~/.config/shum/`
- State and artifacts: `~/.cache/shum/`
  - `state.db`
  - `artifacts/`

## Docs

- [Testing Guide](./docs/testing.md)
- [Project Site](https://imurodl.me/shum/)
- [Contributing Guide](./CONTRIBUTING.md)
- [Security](./SECURITY.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)

## Support and roadmap

- Open issues in the repository for requests and reproducible bugs.
- Track upcoming improvements through commits and changelog entries.

## License

Apache 2.0; see [`LICENSE`](./LICENSE).
