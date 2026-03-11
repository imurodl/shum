# Self-Host Upgrade Manager (`shum`)

`shum` is a CLI-first operations tool for safer Docker Compose upgrades on remote Linux hosts.

Phase 1 is implemented (host trust and discovery), and phase 2-4 now cover planning, backups, upgrade execution, rollback hooks, and history/audit output.

## Quickstart

```bash
go install ./cmd/shum

# 1) trust a host
shum host register myserver

# 2) discover projects
shum project discover myserver

# 3) inspect one project for canonical metadata
shum project inspect myserver web --project-directory /srv/web --json
```

## Operational Flow

```bash
# preflight check
shum project preflight myserver web

# generate plan
shum project plan myserver web --json

# configure policy
shum project policy set myserver web \
  --require-backup=true \
  --backup-command "bash -c 'docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\"'" \
  --restore-command "bash -c 'cat \"$SHUM_BACKUP_ARTIFACT\" | docker exec -i db psql -U app'" \
  --migration-warning=false \
  --health-check "http://localhost:8080/health"

# optional: take a backup
shum project backup take myserver web --json

# run upgrade (dry-run first)
shum project upgrade myserver web --dry-run --json
shum project upgrade myserver web --json

# inspect history
shum project run list --limit 10
shum project run list --host myserver --project web
shum project run show run-<id> --json
shum project backup list myserver web --json
```

## Install and Verification

```bash
go test ./...
go build ./cmd/shum
```

Remote-heavy tests are optional and gated by `SHUM_E2E_SSH_ALIAS`:

```bash
export SHUM_E2E_SSH_ALIAS=your-alias
go test ./test/e2e
```

## Storage

State and artifacts are stored under:

- Config: `~/.config/shum/`
- Data and artifacts: `~/.cache/shum/` (`state.db`, `artifacts/`)

## Roadmap Summary

- `HOST-01..03`: host trust + discovery + inspect
- `PLAN-01..04`: preflight and planning policies
- `BKUP-01..03`: policy-backed backups and restore support
- `UPGD-01..04`: upgrade execution + verification + rollback
- `HIST-01..02`: run and backup history
- `DOCS-01..02`: public landing documentation and command examples
