# Phase 1 Research: Remote Host Registration and Canonical Compose Discovery

**Phase:** 1 - Remote Host Registration and Canonical Compose Discovery
**Requirements:** HOST-01, HOST-02, HOST-03
**Research date:** 2026-03-11
**Confidence:** HIGH

## Planning Outcome

Phase 1 should establish two authoritative identities:

1. **Host identity**: SSH alias + verified host key + resolved SSH config + remote Linux/Docker capability.
2. **Project identity**: project name + exact Compose file set + project directory + active profiles + interpolation environment fingerprint + canonical rendered Compose model.

The implementation should not hand-roll SSH or Compose semantics. Use the system OpenSSH client for connection and trust behavior, and use remote `docker compose` commands for Compose parsing, interpolation, profile activation, and canonical rendering.

## Requirement Mapping

| Requirement | What Phase 1 must deliver |
| --- | --- |
| HOST-01 | Register a Linux host by SSH alias only after strict host-key verification and non-interactive SSH authentication succeed. |
| HOST-02 | Discover Compose project candidates on a registered host, with runtime-first discovery by default and explicit path discovery as the opt-in fallback. |
| HOST-03 | Inspect a project's canonical rendered Compose configuration, identity, services, and storage surfaces, while surfacing ambiguity and blocked states instead of guessing. |

## Recommended Phase Scope

Implement these as the hard boundaries of the phase:

- Register hosts through existing SSH aliases, not through a custom host/user/key wizard.
- Fail closed on unknown or changed host keys.
- Require key-based SSH auth or agent-based auth that works in batch mode.
- Discover projects from existing Compose/runtime signals first.
- Allow deeper discovery only in operator-specified directories.
- Treat canonical Compose rendering as a context tuple, not as "read a nearby YAML file".
- Persist enough project context that later phases can plan and execute without rediscovering heuristically.

Explicitly leave these out of Phase 1:

- Password-based SSH login.
- Tool-managed SSH keys.
- Whole-host recursive filesystem crawling by default.
- Preflight, backup, upgrade, rollback, or repair actions.

## Standard Stack

- **Runtime**: Go CLI, following the project-wide Go recommendation.
- **SSH transport**: system `ssh` and `scp`/`sftp` if file transfer is later needed.
- **Compose control surface**: remote `docker compose` CLI.
- **Structured state**: local SQLite plus artifact files, or an equivalent local durable store if planning keeps storage minimal in Phase 1.

## Architecture Patterns

### 1. Alias-First Host Adapter

Treat the SSH alias as the operator-facing host ID. Resolve it with `ssh -G <alias>` and persist the effective host metadata the tool actually used.

Why:

- It honors `~/.ssh/config`, `Include`, `Match`, `ProxyJump`, `IdentityFile`, and `UserKnownHostsFile` without reimplementing SSH config semantics.
- It matches the project's explicit "SSH alias first" decision.

### 2. Runtime-First, Context-Second Discovery

Discovery should happen in two layers:

1. **Runtime candidate inventory**: find Compose-managed projects or containers already known to Docker/Compose.
2. **Explicit context resolution**: promote a candidate to canonical only when the exact Compose execution context is known and `docker compose config --format json` succeeds.

This preserves the "trustworthy, not magical" constraint from the phase context.

### 3. Canonical Context Tuple

For planning purposes, define canonical project identity as this tuple:

- host alias
- project name
- project directory
- compose file list in order
- active profiles
- env file inputs / interpolation environment fingerprint
- canonical rendered config JSON

If any part of the tuple is unknown, the project can still be discovered, but it is not canonical yet.

### 4. Summary-First Inspect

Default inspect output should answer these questions immediately:

- What host is this?
- What project is this?
- Does the tool trust the host connection?
- Does the tool trust the Compose render as canonical?
- What services and storage surfaces are in scope?
- What is blocking or ambiguous?

Raw rendered config should require an explicit drill-down flag.

## Operator-Facing Behavior

Recommended operator flow:

1. `host register <ssh-alias>`
2. `project discover <ssh-alias>`
3. `project inspect <ssh-alias> <project>`

Expected behavior:

- `host register` should be fast, batch-safe, and fail clearly when trust or auth is not ready.
- `project discover` should default to a concise summary of discovered candidates, not dump rendered YAML/JSON.
- `project inspect` should show whether the project is:
  - canonical and ready for later planning, or
  - runtime-only, ambiguous, or blocked.

The first successful inspect should also create the persisted project context that later phases reuse.

## CLI Shape Implications

Exact command names are still discretionary, but the CLI must expose explicit flags for Compose context because canonical identity depends on them.

Recommended shape:

```text
tool host register <ssh-alias> [--json]
tool host list [--json]
tool host inspect <ssh-alias> [--json]

tool project discover <ssh-alias> [--path DIR ...] [--json]
tool project inspect <ssh-alias> <project-ref> \
  [--project-directory DIR] \
  [--file FILE ...] \
  [--project-name NAME] \
  [--profile PROFILE ...] \
  [--env-file FILE ...] \
  [--show-config] \
  [--show-mounts] \
  [--json]
```

Key CLI implications:

- `--file`, `--project-directory`, `--project-name`, `--profile`, and `--env-file` are not "advanced extras"; they are part of project identity.
- The tool should persist the exact context used for the first canonical render and reuse it later.
- `--json` must be explicit. Human-readable summary is the default.
- `discover` and `inspect` should not silently infer profile or env choices from filename conventions.

## Trusted SSH Host Registration Expectations

### Prescriptive recommendation

Use the system OpenSSH client for Phase 1 instead of an in-process SSH library.

That gives the implementation the real behavior operators already depend on:

- SSH alias resolution
- host key verification
- agent usage
- per-host config
- jump hosts / proxying
- non-interactive batch behavior

### Registration algorithm

1. Resolve the alias with `ssh -G <alias>`.
2. Extract at least:
   - `hostname`
   - `user`
   - `port`
   - `identityfile`
   - `userknownhostsfile`
   - `globalknownhostsfile`
3. Verify there is a known-hosts match for the effective host name, and for `[host]:port` when a non-default port is used.
   - `ssh-keygen -F <host>`
   - `ssh-keygen -F [<host>]:<port>`
4. Perform a non-interactive connectivity probe:
   - `ssh -o BatchMode=yes -o StrictHostKeyChecking=yes -T <alias> 'uname -s && uname -m && command -v docker && docker compose version'`
5. Register only if:
   - host-key verification succeeds
   - SSH auth succeeds without prompting
   - remote OS is Linux
   - `docker` exists
   - `docker compose` exists and responds

### Fail-closed trust expectations

- Unknown host key: registration must fail.
- Changed host key: registration must fail hard.
- Password prompt required: registration must fail.
- Passphrase prompt required and not satisfied by agent: registration must fail.

This is the correct Phase 1 tradeoff because the product promise depends on deliberate trust, not convenience.

### `ssh-keyscan` guidance

Do not auto-trust `ssh-keyscan` output.

OpenSSH explicitly warns that building `known_hosts` from unverified `ssh-keyscan` output leaves users vulnerable to man-in-the-middle attacks. In Phase 1, `ssh-keyscan` can at most support an operator remediation flow, but it must not be the tool's default trust-establishment path.

### Persisted host metadata

Persist at least:

- host alias
- resolved hostname
- user
- port
- known-hosts file path(s) used
- last successful verification timestamp
- remote OS / arch
- Docker version and Compose version if available

If feasible, also persist the verified host-key fingerprint derived from the known-hosts entry so inspect output can show trust details without another lookup.

## Canonical Docker Compose Discovery Rules

### Hard rule

The tool's canonical Compose source of truth is:

`docker compose config --format json`

Docker documents `docker compose config` as the command that parses, resolves, merges, interpolates, normalizes, and renders the actual Compose model to be applied on the Docker Engine. That is the right canonicalization boundary for Phase 1.

### Discovery order

Use this discovery precedence:

1. **Stored explicit context** for a previously canonicalized project.
2. **Runtime candidate discovery** from the remote host.
3. **Explicit path discovery** in operator-chosen directories.

Do not do these in Phase 1:

- whole-host recursive scan by default
- parsing arbitrary YAML locally and pretending it matches Compose behavior
- guessing override chains from filenames like `compose.prod.yaml`

### Runtime candidate discovery

Default discovery should start with projects the host is already running or has recently managed:

- `docker compose ls --all --format json`

This is the right default because it is explicit, fast, and aligned with the phase context's "known Compose projects first" rule.

For each candidate, gather runtime state with:

- `docker compose ps --format json` when explicit context is already known, or
- `docker container ls --all --filter label=com.docker.compose.project --format json` style runtime inspection as a fallback inventory technique when only runtime resources are available.

**Inference from Docker docs:** Compose-managed resources carry the `com.docker.compose.project` label, and the Docker CLI supports label filtering on container inventory. That makes runtime label inspection acceptable as a discovery aid, but not as the canonical render source.

### Explicit path discovery

Path-based discovery should be opt-in:

- `project discover <host> --path /srv/app`

Recommended path resolution behavior:

1. Look for an explicit stored context first.
2. If absent, honor `COMPOSE_FILE` if the selected directory's `.env` makes it part of the project context.
3. Otherwise, check only Compose default filenames in the chosen directory:
   - `compose.yaml`
   - `compose.yml`
   - `docker-compose.yaml`
   - `docker-compose.yml`
4. If the project requires multiple files or non-default file names and they are not explicit, mark the result ambiguous and require `--file`.

For canonicalization, do not rely on Compose's default upward parent-directory file search. Resolve the exact file list first, then invoke Compose with explicit `-f` paths so the tool's behavior stays explainable.

### Compose context that must be captured

For a project to become canonical, the implementation must capture:

- ordered Compose file list
- project directory
- project name
- project name source, if inferable:
  - CLI flag
  - `COMPOSE_PROJECT_NAME`
  - top-level `name:`
  - directory basename fallback
- active profiles
- declared profiles
- interpolation environment used by Compose
- canonical rendered config JSON

Canonical inspect should run with a controlled environment. If the render depends on ambient shell variables that are not represented by `.env`, `--env-file`, or persisted project metadata, the project should remain non-canonical until that env context is made explicit.

Recommended command set for a canonical inspect:

- `docker compose config --format json`
- `docker compose config --environment`
- `docker compose config --profiles`
- `docker compose config --services`
- `docker compose config --volumes`
- `docker compose config --networks`
- `docker compose config --variables`
- `docker compose ps --format json`

### Project name rules that must not be reimplemented

Docker documents the project name precedence order:

1. `-p`
2. `COMPOSE_PROJECT_NAME`
3. top-level `name:`
4. basename of the project directory or first Compose file
5. basename of current directory if no file is specified

Phase 1 should not implement its own project-name resolution logic beyond invoking Compose with explicit context and recording the result.

### Multi-file, profile, env, and include handling

These are all part of canonical identity:

- Multiple `-f` files are merged in order.
- `--profile` / `COMPOSE_PROFILES` changes the active model.
- `.env`, `--env-file`, shell environment, and `COMPOSE_FILE` affect interpolation and project resolution.
- `include:` can introduce additional files, project directories, and env files.

Therefore:

- The tool must let Compose resolve these.
- The tool must store the resolved context that produced the canonical config.
- The tool must not infer "same project" from directory name alone.

### Storage surface discovery for HOST-03

Canonical inspect should surface at least:

- named volumes
- external volumes
- bind mounts
- read-only vs read-write mount intent
- mount source and destination

Recommended approach:

1. Use rendered Compose config for declared surfaces.
2. Use `docker compose ps --format json` to map live services and containers.
3. Use `docker inspect --type=container` on the relevant containers for `Mounts`.
4. Use `docker volume inspect` for named volumes.

Important Docker behavior to reflect in output:

- bind mounts live on the daemon host, not the client
- named volumes have inspectable metadata and mountpoints
- external volumes are dependencies, not resources the tool owns

## Ambiguity and Blocking Behavior

The tool should use explicit status classes instead of a generic "failed".

Recommended status model:

- `canonical`: full context known and `docker compose config --format json` succeeded
- `runtime_only`: runtime project discovered, but file/context is not yet authoritative
- `ambiguous`: multiple plausible canonical contexts exist and operator input is required
- `blocked`: canonical render or trust validation failed

Recommended rules:

| Situation | Status | Required behavior |
| --- | --- | --- |
| Unknown SSH host key | blocked | Refuse registration; instruct operator to verify and seed trust outside the default path. |
| Changed SSH host key | blocked | Refuse registration and discovery; treat as a security event. |
| SSH password/passphrase prompt required | blocked | Refuse registration in v1; require batch-safe key/agent flow. |
| Docker or Compose missing | blocked | Refuse registration because the host is not usable for this product. |
| Runtime project found, but no Compose file context | runtime_only | Show project summary and runtime surfaces, but do not mark canonical. |
| Multiple possible Compose files or overrides | ambiguous | Require explicit `--file` ordering. |
| Declared profiles exist, but none were explicitly chosen | ambiguous | Show declared profiles and require explicit profile selection before treating the result as canonical. |
| Render depends on ambient shell env that is not explicitly captured | ambiguous | Require explicit env-file or persisted env context before treating the render as canonical. |
| `docker compose config` fails due missing required vars / invalid model / missing files | blocked | Keep the candidate visible but mark canonical render blocked and show the error summary. |
| External resources exist in the model | canonical or ambiguous, depending on render | Do not block canonical render solely because a resource is external; flag it as operator-owned dependency. |

### Profile-specific recommendation

Because the phase context explicitly says profile-dependent canonicalization requires explicit user choice, use this rule:

- If `docker compose config --profiles` returns any profile names and the inspect request did not specify active profiles from stored context or CLI flags, the project should not be considered canonical yet.

The tool may still render a base model with no profiles for preview, but it must label that result as non-canonical until profiles are chosen explicitly.

## Inspect Output Strategy

### Human-readable default

Default inspect output should be a concise summary, for example:

```text
Host: prod-eu1
Trust: verified via ~/.ssh/known_hosts, fingerprint SHA256:...
Project: paperless
Canonical status: ambiguous (profiles require explicit selection)
Compose context: /srv/paperless | files: compose.yaml | project name: paperless
Profiles: declared [ocr, debug] | active []
Services: 4 running, 1 exited
Storage: 2 named volumes, 1 bind mount, 1 external volume
Issues: explicit profile selection required before this can be treated as canonical
```

That matches the phase context's requirement that the operator immediately understand:

- what project this is
- whether the tool trusts it as canonical

### Progressive drill-down

Recommended drill-down behavior:

- default: summary only
- `--show-config`: rendered Compose config
- `--show-mounts`: storage surfaces in detail
- `--show-env`: interpolation inputs and env sources
- `--json`: full structured payload

### Raw config safety

`docker compose config --format json` may expose sensitive interpolated values.

Recommended behavior:

- Default human output should never dump raw config.
- Structured output should make it obvious whether values are redacted.
- If the planner wants full fidelity for later phases, persist the exact rendered config in local artifacts, but avoid printing it automatically.

### Structured output shape

The JSON shape should center on trust, identity, and blockers. Recommended fields:

```json
{
  "host": {
    "alias": "prod-eu1",
    "resolvedHostname": "10.0.0.20",
    "user": "deploy",
    "port": 22,
    "trust": {
      "status": "verified",
      "knownHostsFiles": ["/home/user/.ssh/known_hosts"],
      "fingerprint": "SHA256:..."
    }
  },
  "project": {
    "ref": "paperless",
    "status": "ambiguous",
    "identity": {
      "projectName": "paperless",
      "projectDirectory": "/srv/paperless",
      "composeFiles": ["/srv/paperless/compose.yaml"],
      "profilesDeclared": ["ocr", "debug"],
      "profilesActive": []
    },
    "services": [],
    "storageSurfaces": [],
    "issues": []
  }
}
```

## Implementation Notes for Planning

### Do not hand-roll

- SSH config parsing
- known-hosts trust behavior
- Compose file merging
- Compose interpolation
- profile activation logic
- project-name precedence logic

All of those already exist in the tools self-hosters use today. Phase 1 should reuse them.

### Local persistence recommendation

Persist three classes of records:

1. **Host record**
2. **Project record**
3. **Discovery snapshot / inspect artifact**

Minimum host fields:

- alias
- resolved host metadata
- trust metadata
- last verified timestamp

Minimum project fields:

- host alias / foreign key
- project ref
- canonical status
- compose context tuple
- last inspected timestamp

Minimum artifact fields:

- rendered config JSON path
- runtime service snapshot path
- storage surface snapshot path
- env fingerprint

### Environment handling recommendation

Do not persist raw env values in the primary database by default.

Instead:

- persist env source metadata
- persist a stable fingerprint of `docker compose config --environment`
- optionally store the raw environment snapshot in an artifact file if later phases need it

This gives later phases drift detection without making the default inspect output a secret leak.

## Validation Architecture

Phase 1 needs validation at three main layers: pure logic, transport behavior, and real remote-host behavior.

### 1. Unit tests

Cover deterministic logic without requiring Docker or SSH:

- SSH config parsing of `ssh -G` output
- host registration state transitions
- project status classification (`canonical`, `runtime_only`, `ambiguous`, `blocked`)
- Compose context tuple normalization
- inspect summary rendering
- env fingerprinting and redaction logic

### 2. Integration tests for Compose context resolution

Run these against a local Docker + Compose environment using temporary fixture directories:

- single-file project with default `compose.yaml`
- project with top-level `name:`
- project with `COMPOSE_PROJECT_NAME`
- project with `COMPOSE_FILE` pointing outside the working directory
- multi-file project with ordered `-f` inputs
- project with `include:`
- project with profiles
- project with required `${VAR?error}` interpolation
- project with bind mounts + named volumes + external volumes

These tests should assert:

- discovered project name
- captured compose file list
- project directory
- profiles declared / active
- canonical render success or expected blocked state
- storage surfaces returned by inspect

### 3. SSH transport integration tests

Use a disposable Linux test target with OpenSSH enabled and pre-seeded host keys.

Test cases:

- registration succeeds when known_hosts already trusts the host
- registration fails on unknown host key
- registration fails on changed host key
- registration fails when auth requires a password prompt
- registration succeeds through a normal key/agent flow
- `ssh -G` resolution honors alias-specific config

The important assertion is not only "command ran"; it is "strict host-key verification and batch-safe auth behavior were enforced".

### 4. End-to-end remote host tests

Run a real Linux-host test suite, not only local Docker-on-the-controller tests.

Recommended fixture host characteristics:

- Linux
- Docker Engine installed
- Compose plugin installed
- SSH access via a real alias in a temp SSH config

Recommended end-to-end scenarios:

1. **Happy path canonical**
   - `compose.yaml`
   - one named volume
   - inspect becomes `canonical`
2. **Runtime-only candidate**
   - live project exists
   - file context unavailable or removed
   - inspect becomes `runtime_only`
3. **Profiles ambiguous**
   - project declares optional services behind profiles
   - inspect remains `ambiguous` until profile(s) are chosen
4. **Render blocked by required env**
   - `${VAR?error}` missing
   - candidate remains visible but canonical render is `blocked`
5. **Storage surface mix**
   - named volume + bind mount + external volume
   - inspect reports all surfaces correctly

### 5. Acceptance mapping

Map validation directly to the requirements:

- **HOST-01**
  - passes only when registration requires trusted host keys and non-interactive SSH auth
- **HOST-02**
  - passes only when discovery lists Compose project candidates on a registered host
- **HOST-03**
  - passes only when inspect can show canonical rendered config, project identity, and storage surfaces, or clearly explain why canonicalization is blocked

### 6. CI strategy

Practical split:

- unit tests on every change
- local Compose integration tests on every change if Docker is available in CI
- full remote-host E2E in nightly CI or a dedicated integration job

This phase is trust- and environment-sensitive. A remote-host test lane is not optional if the product wants credible operator behavior.

## Planner-Relevant Decisions

These are the decisions the planner should treat as fixed unless implementation evidence proves otherwise:

- Use OpenSSH CLI, not a custom SSH stack, for Phase 1.
- Use `ssh -G` as the source of truth for alias resolution.
- Fail closed on unknown or changed host keys.
- Require batch-safe SSH auth in v1.
- Use `docker compose config --format json` as the canonical Compose render source.
- Discover runtime candidates first; do path-based discovery only when the operator points the tool at directories.
- Treat profile selection and env/interpolation context as part of project identity.
- Persist the exact Compose context tuple for later phases.
- Keep default inspect output concise and summary-first.

## Primary Sources

- Docker Compose CLI reference: https://docs.docker.com/reference/cli/docker/compose/
- `docker compose config`: https://docs.docker.com/reference/cli/docker/compose/config/
- `docker compose ls`: https://docs.docker.com/reference/cli/docker/compose/ls/
- `docker compose ps`: https://docs.docker.com/reference/cli/docker/compose/ps/
- Compose application model and default file names: https://docs.docker.com/compose/intro/compose-application-model/
- Project name precedence: https://docs.docker.com/compose/how-tos/project-name/
- Compose profiles: https://docs.docker.com/compose/how-tos/profiles/
- Compose interpolation and env precedence: https://docs.docker.com/compose/how-tos/environment-variables/variable-interpolation/
- Pre-defined Compose env vars (`COMPOSE_FILE`, `COMPOSE_PROJECT_NAME`, `COMPOSE_PROFILES`): https://docs.docker.com/compose/how-tos/environment-variables/envvars/
- Compose `include:` semantics: https://docs.docker.com/reference/compose-file/include/
- Compose volumes and project labels: https://docs.docker.com/reference/compose-file/volumes/
- Docker container listing and label filters: https://docs.docker.com/reference/cli/docker/container/ls/
- Docker inspect: https://docs.docker.com/reference/cli/docker/inspect/
- Docker volume inspect: https://docs.docker.com/reference/cli/docker/volume/inspect/
- Docker bind mount behavior: https://docs.docker.com/engine/storage/bind-mounts/
- OpenSSH `ssh(1)`: https://man.openbsd.org/ssh.1
- OpenSSH `ssh_config(5)`: https://man.openbsd.org/ssh_config.5
- OpenSSH `ssh-keygen(1)`: https://man.openbsd.org/ssh-keygen.1
- OpenSSH `ssh-keyscan(1)`: https://man.openbsd.org/ssh-keyscan.1
