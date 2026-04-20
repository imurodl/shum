# AGENTS.md — shum for AI agents

`shum` is a CLI for safe, recoverable Docker Compose upgrades on remote SSH hosts.
It is designed to be driven by AI coding agents. Every command speaks `--json`,
errors return stable codes, and the entire surface is discoverable in one shot
via `shum agent-help`.

This file is the contract. If you are an agent that landed in this repo: read
this file, then run `shum agent-help` once at the start of the session.

## When to use shum

| User intent                                                                | Use shum?                            |
| -------------------------------------------------------------------------- | ------------------------------------ |
| Update / redeploy / pull new images for a service running under docker compose on a remote host | Yes                       |
| Roll back a failed upgrade                                                 | Yes (`project run show`, `backup restore`) |
| Inspect what's deployed on a remote host                                   | Yes (`project discover`, `project inspect`) |
| Install or start a brand-new service from scratch                          | No (out of scope)                    |
| Run an arbitrary remote command                                            | No (use ssh directly)                |
| Manage non-Compose workloads (systemd, k8s, bare binaries)                 | No                                   |

## The contract

| Element        | Convention                                                                                       |
| -------------- | ------------------------------------------------------------------------------------------------ |
| Output flag    | `--json` on every read/plan/upgrade command (success path: parseable JSON on stdout)             |
| Error channel  | `stderr`, only when `--json` was set on the failing command                                      |
| Error envelope | `{"error":{"code":"<stable_code>","message":"...","hint":"...","details":{...}}}`                |
| Exit codes     | See table below; treat exit code as the source of truth for severity                             |
| Idempotency    | Read commands are idempotent. `project upgrade --dry-run` is safe to call any number of times    |
| Confirmation   | shum never prompts. Anything destructive happens only when explicitly invoked (no `--yes` needed)|

### Exit code table

| Code | Meaning                                                       | Example error codes                                         |
| ---- | ------------------------------------------------------------- | ----------------------------------------------------------- |
| 0    | Success                                                       | —                                                           |
| 1    | Internal / unclassified failure                               | `internal_error`                                            |
| 64   | Caller error (bad args, invalid flag values)                  | `usage_error`, `probe_invalid`                              |
| 65   | Host-layer problem (SSH, trust, registration)                 | `host_unreachable`, `host_not_found`, `known_hosts_missing` |
| 66   | Project / preflight problem                                   | `project_not_found`, `preflight_blocked`, `policy_missing`  |
| 67   | Backup / artifact problem                                     | `backup_required`, `backup_failed`, `artifact_not_found`    |
| 68   | Upgrade / rollback / health problem                           | `upgrade_failed`, `migration_warning`, `rollback_failed`    |
| 70   | Local store (SQLite) failure                                  | `store_failure`                                             |

## First call: `shum agent-help`

Run once at the start of any session. Returns a single JSON document with the
authoritative surface — version, every command, every flag, every error code,
the JSON shape returned by each command, and a few canonical examples.

```bash
shum agent-help | jq .
```

Top-level keys:

- `tool` — `{name, version, repo}`
- `contract` — `{revision, json_flag, error_channel, error_shape}`
- `commands[]` — every command with `path`, `short`, `long`, `args`, `flags[]`, `returns` (JSON shape)
- `errors{code: {description, exit_code}}` — every error code shum can emit
- `output_shapes{path: shape}` — quick map from command path to JSON shape
- `examples[]` — canonical agent invocations

If something in this file disagrees with `shum agent-help` output, the binary
wins. This file is hand-curated and may drift; the binary is generated.

## Canonical safe-upgrade flow

This is the flow agents should default to when the user says "upgrade X on
host Y". Every step uses `--json`.

```bash
# 1. Verify the host is registered and reachable.
shum host inspect prod --json
#    On exit 65 (host_*): stop and ask the user. Do not retry blindly.

# 2. Confirm the project is known and inspect its current state.
shum project inspect prod web --json
#    On exit 66 / project_not_found: run `shum project discover prod --json` first.

# 3. Read the policy. If require_backup=true and backup_command is empty,
#    that's exit 67 / backup_required — surface it to the user, do not paper over.
shum project policy show prod web --json

# 4. Plan, then dry-run. Show the plan to the user before any non-dry-run.
shum project plan prod web --json
shum project upgrade prod web --dry-run --json

# 5. Real upgrade. Confirm with the user first if this isn't an
#    explicitly-approved automation context.
shum project upgrade prod web --json
#    Read .status from the result: success | failed | rolled_back.
#    On exit 68 / migration_warning: STOP, do not auto-add --force.
#    On exit 68 / rollback_failed: STOP, page the user, manual intervention.
```

## Command surface (16 commands)

Run `shum agent-help` for full flag/return-shape detail. Quick reference:

| Path                     | What it does                                              |
| ------------------------ | --------------------------------------------------------- |
| `host register`          | Register an SSH alias and probe Docker/Compose version    |
| `host list`              | List registered hosts                                     |
| `host inspect`           | Show details for one registered host                      |
| `project discover`       | Find compose projects running on a host                   |
| `project inspect`        | Show effective config, files, mounts for a project        |
| `project preflight`      | Check Docker/Compose/disk/permissions on the host         |
| `project plan`           | Preview image and service changes before upgrade          |
| `project policy show`    | Show backup + health policy for a project                 |
| `project policy set`     | Update backup + health policy                             |
| `project backup take`    | Take a project backup using the configured command        |
| `project backup list`    | List existing backup artifacts                            |
| `project backup restore` | Restore a backup artifact                                 |
| `project upgrade`        | Run the safe-upgrade flow (preflight, plan, backup, apply)|
| `project run list`       | List historical upgrade runs                              |
| `project run show`       | Show details of one upgrade run, including events         |
| `agent-help`             | Emit this contract as JSON for context loading            |

## Error codes (22 codes)

Stable across patch releases. Renaming a code is a breaking change.

| Code                    | Exit | Meaning                                                                           |
| ----------------------- | ---- | --------------------------------------------------------------------------------- |
| `internal_error`        | 1    | Unclassified failure inside shum. File a bug if reproducible.                     |
| `usage_error`           | 64   | Caller invoked shum with invalid or missing arguments.                            |
| `probe_invalid`         | 64   | Health probe specification could not be parsed.                                   |
| `host_unreachable`      | 65   | SSH could not reach the host (network, auth, or sshd down).                       |
| `host_unverified`       | 65   | Host key fingerprint did not match a trusted known_hosts entry.                   |
| `host_not_found`        | 65   | No host registered under the given alias.                                         |
| `host_not_linux`        | 65   | Target host is not running Linux; shum only supports Linux hosts.                 |
| `ssh_config_invalid`    | 65   | SSH config could not resolve the alias to a host.                                 |
| `known_hosts_missing`   | 65   | No known_hosts file is configured for the SSH alias.                              |
| `project_not_found`     | 66   | No compose project registered with the given reference on this host.              |
| `compose_unavailable`   | 66   | Docker or docker compose is not installed on the remote host.                     |
| `policy_missing`        | 66   | Project has no safety policy configured. Run `project policy set` first.          |
| `preflight_blocked`     | 66   | Preflight checks reported one or more blocking issues.                            |
| `backup_required`       | 67   | Project policy requires a backup command but none is configured.                  |
| `backup_failed`         | 67   | Backup command exited non-zero on the remote host.                                |
| `restore_failed`        | 67   | Restore command exited non-zero during rollback.                                  |
| `artifact_not_found`    | 67   | Backup artifact path does not exist.                                              |
| `migration_warning`     | 68   | Policy flags this upgrade as risky. Re-run with `--force` to proceed.             |
| `upgrade_failed`        | 68   | Compose pull/up step failed on the remote host.                                   |
| `rollback_failed`       | 68   | Rollback after a failed upgrade also failed; manual intervention required.        |
| `health_check_failed`   | 68   | Post-upgrade health probes did not pass within the timeout.                       |
| `store_failure`         | 70   | Local SQLite store returned an error.                                             |

## Failure handling rules

A few rules that matter more than the rest:

- **Never `--force` automatically.** `migration_warning` (exit 68) means the policy author marked this upgrade as risky. Stop, surface the warning to the user, ask before re-running with `--force`.
- **Never retry `host_unreachable` blindly.** Exit 65 from the host layer usually means a real problem (down, key changed, network). Surface the hint to the user.
- **`rollback_failed` (exit 68) is a page-the-human signal.** The deploy is half-applied and the rollback also failed. Stop the agent loop and surface to the user immediately.
- **Parse `.error.code`, not `.error.message`.** Messages are human-readable and may change across patch releases. Codes are stable.

## Examples directory

Ready-to-use harness setups:

- [`examples/agents/claude-code/`](./examples/agents/claude-code/) — Claude Code skill + slash command
- [`examples/agents/codex/`](./examples/agents/codex/) — OpenAI Codex CLI configuration
- [`examples/agents/gemini/`](./examples/agents/gemini/) — Google Gemini CLI configuration
