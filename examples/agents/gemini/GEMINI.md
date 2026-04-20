# shum + Gemini CLI

You have access to `shum`, a CLI for safe Docker Compose upgrades on remote SSH hosts. The full agent contract lives at the repo root in [`AGENTS.md`](../../../AGENTS.md). Treat that as authoritative.

## Working rules

- Always pass `--json` to shum commands.
- Errors land on **stderr** as `{"error":{"code":"...","message":"...","hint":"...","details":{...}}}`. Parse `.error.code`, never `.error.message`.
- Exit codes are stable: `64` caller error, `65` host layer, `66` preflight/project, `67` backup, `68` upgrade/rollback, `70` store, `1` unknown.
- If the surface is unfamiliar, run `shum agent-help | jq .` once at the start of the session to load every command, every flag, every error code, every output shape.

## Hard rules — do not skip

- `migration_warning` (exit 68): **STOP.** Do not auto-add `--force`. Surface the warning and ask the user.
- `host_unreachable` (exit 65): **STOP.** Do not retry blindly.
- `rollback_failed` (exit 68): **STOP and page the user.** Half-applied deploy.
- `backup_required` (exit 67): **STOP.** The project policy demands a backup; do not silently fall back to `--skip-backup`.

## Canonical safe-upgrade flow

```bash
shum host inspect <host> --json
shum project inspect <host> <project> --json
shum project policy show <host> <project> --json
shum project plan <host> <project> --json
shum project upgrade <host> <project> --dry-run --json
# pause, summarize the dry-run to the user, get explicit confirmation
shum project upgrade <host> <project> --json
```
