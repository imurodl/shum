# Architecture

shum is a single Go binary that drives safe, recoverable Docker Compose
upgrades on remote SSH hosts. It is designed to be run by both humans and AI
agents: every command speaks `--json`, failures return a typed envelope with a
stable error code, and the entire surface is discoverable in one call via
`shum agent-help`.

## High-level flow

```
        CLI (cobra)                remote host (via ssh)
  ┌───────────────────┐        ┌────────────────────────┐
  │ internal/cli      │        │ docker / docker compose │
  │  host / project   │        └────────────────────────┘
  └─────────┬─────────┘                    ▲
            │ services                      │ ssh (BatchMode, StrictHostKey)
  ┌─────────▼─────────┐        ┌────────────┴───────────┐
  │ hosts / projects  │───────▶│ internal/remote.Runner │
  │ ops (engine)      │        └────────────────────────┘
  └─────────┬─────────┘
            │ persistence
  ┌─────────▼─────────┐
  │ internal/store    │  SQLite (modernc.org/sqlite), embedded migration
  └───────────────────┘
```

Nothing mutates a remote host until an `upgrade` runs. Read commands
(`discover`, `inspect`, `plan`, `preflight`) only observe, and `--dry-run`
walks the full upgrade path without changing anything.

## Packages

| Package | Responsibility |
|---|---|
| `cmd/shum` | Entry point; builds the root command via `cli.NewRootCommand`. |
| `internal/cli` | Cobra command tree; JSON and human output formatting. |
| `internal/config` | Resolves XDG config/cache paths and the SQLite database location. |
| `internal/store` | Opens SQLite and applies the embedded schema migration. |
| `internal/hosts` | Host registry keyed by SSH alias (register / list / inspect). |
| `internal/projects` | Persisted project records and status. |
| `internal/projects/discovery` | Discovers Compose projects already running on a host. |
| `internal/projects/inspect` | Inspects a single project's services and artifacts. |
| `internal/ops` | Upgrade engine and the upgrade-run / event / backup records. |
| `internal/remote` | Runs commands on a host over `ssh` with hardened flags. |
| `internal/ssh` | Parses SSH config and `known_hosts`; probes host OS/arch/Docker. |
| `internal/shumerr` | Typed error envelopes, stable error codes, and exit codes. |

## Command surface

The authoritative surface is `shum agent-help` (emits every command, flag,
error code, and JSON shape as one document). The command tree:

```
shum
├── agent-help
├── host
│   ├── register <alias>
│   ├── list
│   └── inspect <alias>
└── project
    ├── discover <alias>
    ├── inspect <alias> <project-ref>
    ├── preflight <alias> <project-ref>
    ├── plan <alias> <project-ref>
    ├── upgrade <alias> <project-ref>        # --dry-run, --force
    ├── policy
    │   ├── show <alias> <project-ref>
    │   └── set <alias> <project-ref>
    ├── backup
    │   ├── take <alias> <project-ref>
    │   ├── list <alias> <project-ref>
    │   └── restore <alias> <project-ref> <artifact-path>
    └── run
        ├── list
        └── show <run-id>
```

## Errors and exit codes

Failures are written to stderr as a JSON envelope and the process exits with a
documented, stable code:

```json
{
  "error": {
    "code": "project_not_found",
    "message": "project not found",
    "hint": "run `shum project discover <alias>` to populate the project list",
    "details": { "host_alias": "prod", "project_ref": "web" }
  }
}
```

Codes are part of the public surface and are never renamed within a patch
release. Agents should branch on `.error.code`, never on the message text. The
full code-to-exit-code table lives in `internal/shumerr` and is emitted by
`shum agent-help`.

## State

- **Config**: `~/.config/shum` (or the platform config dir).
- **State / cache**: `~/.cache/shum`, including the SQLite database
  (`state.db`) and upgrade artifacts under `artifacts/`.
- `known_hosts` defaults to `~/.ssh/known_hosts` and can be overridden with the
  `SHUM_KNOWN_HOSTS` environment variable.
