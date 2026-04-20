---
name: shum
description: Safely upgrade Docker Compose services on remote SSH hosts using the shum CLI. Use this skill whenever the user wants to update, redeploy, pull new images for, or roll back a service running under docker compose on a remote host. Also handles inspection ("what's deployed on prod?") and run history ("what changed yesterday?").
allowed-tools: Bash(shum *), Bash(jq *), Read
---

# shum — safe remote Docker Compose upgrades

`shum` is a CLI for safe, recoverable Docker Compose upgrades on remote SSH hosts. It exposes a stable `--json` contract designed to be driven by AI agents.

Use this skill when the user wants to update a service running under docker compose on a remote host, roll back a failed upgrade, inspect remote state, or audit upgrade history.

## At session start: load the surface

If you have not yet seen `shum agent-help` output in this session, call it once before doing anything else. It returns a single JSON document with every command, every flag, every error code, the JSON output shape per command, and canonical examples.

```bash
shum agent-help | jq .
```

The full agent contract is in the repo's `AGENTS.md`. Trust the binary over the docs if they disagree.

## Working rules

1. **Always use `--json`.** Every command supports it. Parse the output, don't string-match on human text.
2. **Parse `.error.code`, never `.error.message`.** Codes are stable across patch releases; messages may change. Errors are written to **stderr** as `{"error":{"code":"...","message":"...","hint":"...","details":{...}}}`.
3. **Read exit codes.** `64` = caller error, `65` = host layer, `66` = preflight/project, `67` = backup, `68` = upgrade/rollback, `70` = local store, `1` = unknown.
4. **Always dry-run first.** Run `shum project upgrade <host> <project> --dry-run --json` before any real upgrade and show the plan summary to the user.
5. **Confirm destructive runs.** Unless the user has explicitly said "go ahead, upgrade it for real," ask for confirmation between the dry-run and the real run.

## Failure-handling rules (do not skip)

- **`migration_warning` (exit 68): STOP.** Do not auto-add `--force`. Surface the warning text to the user and ask whether to proceed.
- **`host_unreachable` (exit 65): STOP.** Do not retry blindly. Show the hint, ask the user to check network/auth.
- **`rollback_failed` (exit 68): STOP and page the user.** The deploy is half-applied and the rollback also failed. Do not loop.
- **`backup_required` (exit 67): STOP.** The project's policy requires a backup but none is configured. Either help the user run `shum project policy set --backup-command "..."` or fall back to `--skip-backup` only if the user explicitly approves it.
- **`host_not_found` (exit 65) on inspect/discover:** offer to register it with `shum host register <alias>`.
- **`project_not_found` (exit 66):** offer to discover with `shum project discover <alias> --json` first.

## Canonical safe-upgrade flow

When the user says "upgrade <service> on <host>":

```bash
# 1. Verify the host. On exit 65, stop and ask.
shum host inspect <host> --json

# 2. Verify the project exists. On project_not_found, run discover first.
shum project inspect <host> <project> --json

# 3. Read the policy and surface backup/restore/health settings to the user.
shum project policy show <host> <project> --json

# 4. Plan and dry-run. Show the plan summary to the user.
shum project plan <host> <project> --json
shum project upgrade <host> <project> --dry-run --json

# 5. Confirm with the user, then real upgrade.
shum project upgrade <host> <project> --json
# Read .status: success | failed | rolled_back. Surface .summary either way.

# 6. If the user wants details, show the run.
shum project run show <run-id> --json
```

## Inspection-only flow

When the user just wants to see what's there:

```bash
shum host list --json
shum project discover <host> --json
shum project inspect <host> <project> --json
shum project run list --host <host> --project <project> --limit 5 --json
```

No mutations, safe to run without confirmation.

## What this skill does NOT do

- Does not install new services from scratch (shum is for upgrading existing ones).
- Does not run arbitrary remote commands (use Bash with explicit ssh).
- Does not manage non-Compose workloads (systemd, k8s, bare binaries).
- Does not auto-add `--force` or `--skip-backup` — those require explicit user approval.
