# shum + Claude Code

Two ways to drive `shum` from [Claude Code](https://claude.com/claude-code):

1. **Skill** (`.claude/skills/shum/SKILL.md`) — ambient capability. Once installed, Claude knows when to use shum and how, without an explicit command. Trigger with natural language: *"upgrade the web service on prod"*.
2. **Slash command** (`.claude/commands/shum-upgrade.md`) — one-shot button. Trigger with `/shum-upgrade prod web`.

Both share the same safe-upgrade flow and the same failure-handling rules. Use the skill when you want shum to be available across many turns and contexts; use the slash command when you want a discoverable, named action.

## Install

### Project-scoped (recommended for one repo)

From the root of a repo where you'll be operating:

```bash
mkdir -p .claude
cp -r path/to/shum/examples/agents/claude-code/.claude/skills/shum  .claude/skills/
cp    path/to/shum/examples/agents/claude-code/.claude/commands/shum-upgrade.md .claude/commands/
```

### User-scoped (available everywhere)

```bash
mkdir -p ~/.claude
cp -r path/to/shum/examples/agents/claude-code/.claude/skills/shum  ~/.claude/skills/
cp    path/to/shum/examples/agents/claude-code/.claude/commands/shum-upgrade.md ~/.claude/commands/
```

## Prerequisites

- `shum` on your PATH: `go install github.com/imurodl/shum/cmd/shum@latest`
- `jq` on your PATH (the skill pipes `agent-help` through `jq`)
- At least one host registered: `shum host register prod`
- At least one project discovered: `shum project discover prod`
- (Recommended) A policy on the project: `shum project policy set prod web --require-backup=true --backup-command "..." --restore-command "..."`

## Try it

### Skill

Start a Claude Code session in the repo where the skill is installed and try:

> upgrade the web service on prod

Claude will pick up the skill and run the canonical safe-upgrade flow: inspect host, inspect project, read policy, plan, dry-run, ask you to confirm, then real run.

### Slash command

```
/shum-upgrade prod web
```

Same flow, but explicitly invoked.

## What to expect

A typical successful run looks roughly like this:

```
> /shum-upgrade prod web

Verifying host prod... OK (Docker 26.1.4, Compose v2.29.1)
Verifying project web exists... OK (/srv/web)
Reading policy... require_backup=true, backup_command set, 1 health check
Planning... 2 services will change:
  - api: ghcr.io/acme/api:latest  sha256:9a1...  ->  sha256:b2f...
  - worker: ghcr.io/acme/worker:latest  sha256:01c...  ->  sha256:e88...
Dry-run... OK, no blocks, 0 warnings.

Apply this upgrade for real?

> yes

Running upgrade... backup taken (artifacts/backups/prod/web/171...txt)
status: success
summary: upgrade completed
run_id: run-1714834290291

Want me to show the full run record?
```

A failure with a stable code looks like:

```
$ shum project upgrade prod web --json
# ... on stderr:
{
  "error": {
    "code": "migration_warning",
    "message": "migration warning is enabled; use --force to continue",
    "hint": "review the plan, then re-run with --force if the upgrade is intentional",
    "details": {"host_alias": "prod", "project_ref": "web"}
  }
}
# exit code: 68
```

The skill knows to stop on this code and ask you before adding `--force`.

## Reference

- Repo-root [AGENTS.md](../../../AGENTS.md) — full agent contract: every command, every error code, every exit code.
- `shum agent-help` — same data, machine-readable.
