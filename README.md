# Self-Host Upgrade Manager (SHUM)

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/imurodl/shum/.github/workflows/deploy-site.yml?label=site)
![GitHub last commit](https://img.shields.io/github/last-commit/imurodl/shum)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

**Website:** https://shum.imurodl.me/
**Repository:** https://github.com/imurodl/shum
**License:** Apache-2.0

`shum` is a CLI tool for safe Docker Compose upgrades on self-hosted servers. Back up first, dry-run before applying, and keep a full audit trail — without wrapping your stack in a platform.

## What it solves

SSH and Compose get you to production. They don't give you a repeatable path back when something breaks.

- Upgrades happen from shell history, not a documented flow.
- Backups are optional until the rollout is already failing.
- There's no standard preflight — readiness gets eyeballed every time.
- After a bad run, reconstructing what changed is guesswork.

## What it gives you

- **Register once, target always** — SSH aliases are the identity. No extra credentials or config files.
- **See before you change** — inspect, preflight, and dry-run are part of the normal path, not optional steps.
- **Policy travels with the project** — backup commands and health checks are stored per-project, not per-operator.
- **Recovery is built in** — backup artifacts are created and stored before every upgrade that requires them.
- **Every run leaves a record** — status, changed services, and health outcomes are queryable after the fact.

## Command surface

**Host management**
- `shum host register <alias>` — register an SSH alias and verify connectivity
- `shum host list` / `shum host inspect <alias>` — list and inspect registered hosts

**Project discovery and inspection**
- `shum project discover <alias>` — find running Compose projects on the host
- `shum project inspect <alias> <project>` — read effective config and risk surfaces
- `shum project preflight <alias> <project>` — check Docker, Compose, disk, and project readiness

**Planning**
- `shum project plan <alias> <project>` — preview image and config changes before execution

**Policy**
- `shum project policy set/show <alias> <project>` — store backup requirements and health probes with the project

**Backup**
- `shum project backup take/list/restore <alias> <project>` — create and manage backup artifacts

**Upgrade**
- `shum project upgrade <alias> <project>` — run the upgrade with optional dry-run and health probes

**History**
- `shum project run list/show` — inspect upgrade run history and outcomes

## Install

Requires Go 1.22 or later.

```bash
go install github.com/imurodl/shum/cmd/shum@latest
```

Source build:

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
- [Project Site](https://shum.imurodl.me/)
- [Contributing Guide](./CONTRIBUTING.md)
- [Security](./SECURITY.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)

Open an issue for bugs or feature requests.

## License

Apache 2.0; see [`LICENSE`](./LICENSE).
