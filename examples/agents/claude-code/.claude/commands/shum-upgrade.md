---
description: Run a safe shum upgrade on a remote Compose service. Args: <host-alias> <project-ref>
---

You are running a safe upgrade for a Docker Compose service on a remote SSH host using the `shum` CLI.

Target host: `$1`
Target project: `$2`

Follow this flow exactly. Use `--json` on every step. Errors land on stderr as `{"error":{"code":"...","message":"...","hint":"...","details":{...}}}` — parse `.error.code`, never the message text.

1. **Verify the host.**
   ```bash
   shum host inspect $1 --json
   ```
   On exit 65 (any `host_*` or `ssh_*` or `known_hosts_missing` code): stop, surface the hint to the user, do not retry blindly.

2. **Verify the project exists.**
   ```bash
   shum project inspect $1 $2 --json
   ```
   On `project_not_found` (exit 66): offer to run `shum project discover $1 --json` first.

3. **Read the policy.** Show the user `require_backup`, `backup_command`, `restore_command`, and `health_checks`.
   ```bash
   shum project policy show $1 $2 --json
   ```
   On `backup_required` (exit 67) here or in step 5: stop and ask the user before proceeding.

4. **Plan and dry-run.**
   ```bash
   shum project plan $1 $2 --json
   shum project upgrade $1 $2 --dry-run --json
   ```
   Summarize the plan to the user: services changing, current vs target digests, warnings, blocks. On any `blocks` array being non-empty (`preflight_blocked`, exit 66): stop.

5. **Pause and confirm with the user before the real run.** Quote the dry-run summary, then ask: "Apply this upgrade for real?" Wait for explicit yes.

6. **Real upgrade.**
   ```bash
   shum project upgrade $1 $2 --json
   ```
   Read `.status`: `success`, `failed`, or `rolled_back`. Surface `.summary` regardless.

7. **Failure handling — these are hard rules:**
   - `migration_warning` (exit 68): **STOP.** Do not auto-add `--force`. Show the warning, ask the user.
   - `host_unreachable` (exit 65): **STOP.** Do not retry. Ask the user.
   - `rollback_failed` (exit 68): **STOP and page the user immediately.** Half-applied deploy, rollback failed.
   - `health_check_failed` (exit 68): the upgrade was rolled back automatically. Show the failed probe.

8. **If the user wants details on what happened:**
   ```bash
   shum project run show <run-id-from-step-6> --json
   ```

If you have not yet run `shum agent-help` in this session, start there before step 1 — it returns the full surface (commands, flags, error codes, output shapes) as one JSON document.
