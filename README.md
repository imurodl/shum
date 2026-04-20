# Self-Host Upgrade Manager (SHUM)

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/imurodl/shum/.github/workflows/deploy-site.yml?label=site)
![GitHub last commit](https://img.shields.io/github/last-commit/imurodl/shum)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

**Website:** https://shum.imurodl.me/
**Repository:** https://github.com/imurodl/shum
**License:** Apache-2.0

## What it is

`shum` is a CLI for safe, recoverable Docker Compose upgrades on remote SSH hosts. It is built to be driven by AI coding agents — every command speaks `--json`, errors return stable codes with documented exit codes, and the entire surface is discoverable in one shot via `shum agent-help`. Humans drive it just as well from the terminal.

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

## Quickstart for AI agents

Designed for Claude Code, OpenAI Codex CLI, Gemini CLI, and similar tools.

```bash
# 1. Load the full surface once into context.
shum agent-help | jq .

# 2. Discover what to operate on.
shum host list --json
shum project discover prod --json

# 3. Plan before acting.
shum project plan prod web --json

# 4. Dry-run, then real run. Parse .error.code on stderr if either fails.
shum project upgrade prod web --dry-run --json
shum project upgrade prod web --json
```

The full agent contract — error codes, exit codes, and failure-handling rules — lives in [AGENTS.md](./AGENTS.md). Ready-to-use harness configs:

- [`examples/agents/claude-code/`](./examples/agents/claude-code/) — Claude Code skill + slash command
- [`examples/agents/codex/`](./examples/agents/codex/) — OpenAI Codex CLI
- [`examples/agents/gemini/`](./examples/agents/gemini/) — Gemini CLI

## Quickstart for humans

```bash
shum host register prod
shum project discover prod
shum project inspect prod web --project-directory /srv/web
shum project preflight prod web
shum project plan prod web
```

## Command surface

**Discoverability**
- `shum agent-help` — emit the full CLI surface as JSON (commands, flags, error codes, output shapes)

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

- [Agent Contract (AGENTS.md)](./AGENTS.md)
- [Testing Guide](./docs/testing.md)
- [Project Site](https://shum.imurodl.me/)
- [Contributing Guide](./CONTRIBUTING.md)
- [Security](./SECURITY.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)

Open an issue for bugs or feature requests.

## License

Apache 2.0; see [`LICENSE`](./LICENSE).
