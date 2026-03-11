# Architecture Research

**Domain:** Agentless upgrade, backup, rollback, and recovery orchestration for Docker Compose on Linux
**Researched:** 2026-03-11
**Confidence:** MEDIUM

## Standard Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────────────────┐
│              Product Surface / Low-Privilege Plane                          │
├──────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐  │
│  │ CLI / future TUI │  │ Local config +       │  │ Public landing/docs  │  │
│  │ command surface  │  │ run history store    │  │ site (read-only)     │  │
│  └────────┬─────────┘  └──────────┬───────────┘  └──────────────────────┘  │
├───────────┴───────────────────────┴──────────────────────────────────────────┤
│                 Local Control Plane / Trusted Coordinator                   │
├──────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────────────────┐ │
│  │ Host/project │ │ Preflight +  │ │ Upgrade saga │ │ Artifact + journal │ │
│  │ discovery    │ │ policy engine│ │ orchestrator │ │ store              │ │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘ └─────────┬──────────┘ │
│         │                │                │                   │            │
│  ┌──────┴────────────────┴────────────────┴───────────────────┴──────────┐ │
│  │ SSH transport, host locks, remote command protocol, log/event stream  │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
├──────────────────────────────────────────────────────────────────────────────┤
│              Remote Execution Plane / Privileged Target Host               │
├──────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────────────────┐ │
│  │ Compose      │ │ Hook runner  │ │ Backup /     │ │ Health verifier    │ │
│  │ adapter      │ │ exec + run   │ │ restore      │ │ ps + inspect + app │ │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘ └─────────┬──────────┘ │
│         │                │                │                   │            │
│  ┌──────┴─────────────┐  │   ┌────────────┴─────────────┐  ┌──┴─────────┐ │
│  │ Docker Engine +    │  │   │ Named volumes, bind      │  │ Images +   │ │
│  │ Compose plugin     │  │   │ mounts, snapshot tools   │  │ registries │ │
│  └────────────────────┘  │   └──────────────────────────┘  └────────────┘ │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Inference from sources:** the tool should be agentless in v1. Docker documents SSH-protected remote daemon access, rootless Docker uses a user-scoped daemon, and bind mounts exist on the daemon host rather than the client. That combination favors a local coordinator that runs remote `docker compose` and filesystem operations over SSH, instead of a long-lived privileged agent on every host.

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| CLI surface | Parse commands, show plans, stream logs, print recovery guidance | Go CLI with Cobra or similar |
| Local state store | Persist hosts, policies, run summaries, artifact index, host key policy | SQLite + filesystem artifact directory |
| Discovery service | Identify Compose projects, resolved config, services, images, mounts, health state | Remote `docker compose ls/config/ps` + `docker inspect` |
| Preflight engine | Validate Docker/Compose availability, permissions, storage coverage, disk headroom, backup policy, probe config | Pure domain rules over discovered state |
| Upgrade saga orchestrator | Execute ordered steps, enforce locks/timeouts, trigger compensating rollback | Finite-state workflow with append-only events |
| SSH transport | Reuse connections, copy small manifests, run commands, stream stdout/stderr, capture exit codes | OpenSSH subprocesses or native SSH client |
| Compose adapter | Encapsulate all Compose-specific commands and output parsing | Typed wrapper around `docker compose` |
| Hook runner | Run pre-backup, pre-upgrade, post-upgrade, and post-restore commands in running or one-off containers | `docker compose exec` and `docker compose run --rm` |
| Backup/restore adapters | Handle named volumes, bind mounts, image archives, and optional host snapshot backends | Strategy interfaces with per-surface implementations |
| Health verifier | Decide success or failure after upgrade or restore | `docker compose up --wait`, `ps --format json`, custom HTTP/TCP/command probes |
| Public docs site | Explain installation, trust model, backup recipes, and command reference | Static site generated from repo content |

### Trust Boundaries

| Boundary | What crosses it | Why it matters |
|----------|------------------|----------------|
| Operator machine ↔ target host | SSH commands, streamed logs, staged artifacts | This is the main control boundary; host authenticity and operator identity must be explicit |
| Target user ↔ Docker daemon | Access to Docker socket or rootless daemon | Docker documents that membership in the `docker` group grants root-level privileges, so this boundary is effectively privileged |
| Target host ↔ registries / backup destinations | Pulled images, pushed artifacts, remote snapshots | Network failures here must not corrupt local rollback state |
| CLI/docs site boundary | Generated docs and examples only | Public docs must never access operator secrets, host inventory, or run artifacts |

**Inference from sources:** treat any remote account that can use Docker as privileged, even if it is not `root`. Rootless Docker reduces attack surface, but the tool still needs an explicit privilege model because both rootful `docker` group access and rootless user daemons can mutate production services.

## Recommended Project Structure

```text
cmd/
├── shum/                       # CLI entrypoint
internal/
├── app/
│   ├── command/                # Cobra/TUI handlers mapped to use cases
│   └── usecase/                # upgrade, rollback, recover, inspect flows
├── domain/
│   ├── host/                   # host inventory and capabilities
│   ├── project/                # Compose project model
│   ├── release/                # immutable release/checkpoint model
│   ├── run/                    # run state machine and audit events
│   └── policy/                 # preflight and safety policy rules
├── orchestrator/
│   ├── saga/                   # ordered steps + compensating actions
│   └── lock/                   # per-host and per-project locking
├── transport/
│   ├── ssh/                    # command execution, file copy, session reuse
│   └── stream/                 # stdout/stderr/event framing
├── remote/
│   ├── compose/                # docker compose wrappers and parsers
│   ├── hooks/                  # exec/run helpers
│   ├── probe/                  # health and readiness probes
│   └── host/                   # disk, kernel, daemon capability checks
├── backup/
│   ├── manifest/               # storage surface manifests
│   ├── volume/                 # named volume backup/restore
│   ├── bindmount/              # bind mount backup/restore
│   ├── imagearchive/           # docker image save/load handling
│   └── external/               # optional host snapshot/backups
├── store/
│   ├── state/                  # SQLite repositories
│   ├── artifact/               # local artifact layout and checksums
│   └── report/                 # history and summary read models
├── docspec/                    # generated CLI/reference data for docs site
└── testutil/                   # integration helpers and fixtures
site/
├── src/                        # landing/docs site source
└── public/                     # static assets
docs/
├── architecture/               # generated diagrams and ADRs
└── recipes/                    # host-specific backup recipes
examples/
├── compose/                    # sample apps and failure scenarios
└── policies/                   # policy and hook examples
```

### Structure Rationale

- **`cmd/` + `internal/`:** matches normal Go packaging and keeps the operational core in one deployable CLI binary.
- **`domain/`:** isolates policy and workflow rules from SSH and shell details.
- **`orchestrator/`:** makes upgrade, rollback, and recovery a first-class workflow engine rather than scattered command handlers.
- **`remote/`:** centralizes all Docker/Compose/Linux interactions so CLI code never shells out directly.
- **`backup/`:** separates storage-surface logic because named volumes, bind mounts, and optional host snapshots have different guarantees.
- **`store/`:** keeps audit history and artifact retention explicit; rollback must depend on persisted state, not in-memory state.
- **`site/`:** keeps the public docs surface read-only and isolated from operator credentials.

## Architectural Patterns

### Pattern 1: Agentless Control Plane Over SSH

**What:** Run the orchestrator locally and execute privileged operations on the remote host via SSH.
**When to use:** v1 CLI-first OSS product, especially when the target environment is Linux + Docker Compose and the operator already has SSH access.
**Trade-offs:** Lowest host footprint and clearest trust model, but large backups move over SSH and long-running workflows need careful reconnect and timeout handling.

**Example:**
```go
type RemoteSession interface {
    Run(ctx context.Context, req Command) (Result, error)
    Stream(ctx context.Context, req Command) (<-chan StreamEvent, error)
    CopyFrom(ctx context.Context, remotePath, localPath string) error
    CopyTo(ctx context.Context, localPath, remotePath string) error
}

type Command struct {
    Dir     string
    Env     map[string]string
    Args    []string
    Timeout time.Duration
}
```

**Inference from sources:** prefer invoking the remote `docker` CLI under the remote user instead of talking directly to a hard-coded daemon socket. Docker documents SSH-protected daemon access and rootless daemons; remote CLI execution naturally works in both rootful and rootless setups.

### Pattern 2: Immutable Release Bundle + Run Journal

**What:** Before mutation, capture a checkpoint containing the resolved Compose model, project identity, env inputs, image digests, live mount inventory, health baseline, and artifact references.
**When to use:** Every upgrade, rollback, or explicit recovery operation.
**Trade-offs:** More disk usage and more up-front work, but deterministic rollback and useful audit history.

**Example:**
```go
type ReleaseBundle struct {
    RunID             string
    HostID            string
    ProjectName       string
    ComposeConfigJSON []byte
    DigestLockFile    []byte
    EnvSnapshot       map[string]string
    ServiceHashes     map[string]string
    StorageManifest   []StorageSurface
    ImageRefs         []ImageRef
    ArtifactRefs      []ArtifactRef
    CapturedAt        time.Time
}
```

**Inference from sources:** `docker compose config` can validate the model, emit JSON, hash services, and pin tags to digests. That makes it the right primitive for a canonical rollback checkpoint instead of storing only raw YAML and mutable image tags.

### Pattern 3: Storage-Surface Inventory + Backup Adapter Strategy

**What:** Build a manifest of all persisted data surfaces, then dispatch each surface to the correct backup/restore adapter.
**When to use:** Any workflow that promises rollback or disaster recovery.
**Trade-offs:** More implementation work than “tar everything,” but it prevents false safety claims and makes unsupported surfaces explicit.

**Example:**
```go
type StorageSurface struct {
    Kind        string // volume, bind, image-archive, external-snapshot
    Service     string
    Source      string
    Target      string
    ReadOnly    bool
    Consistency string // crash-consistent or app-consistent
}

type BackupAdapter interface {
    Supports(StorageSurface) bool
    Backup(context.Context, StorageSurface) (ArtifactRef, error)
    Restore(context.Context, StorageSurface, ArtifactRef) error
}
```

**Inference from sources:** Docker documents that volumes are easier to back up or migrate than bind mounts, and bind mounts are created on the daemon host rather than the client. The tool should therefore treat volumes and bind mounts as different classes of risk and implementation.

### Pattern 4: Health-Gated Saga With Compensating Rollback

**What:** Model upgrade as ordered steps with explicit compensating actions: discover, preflight, checkpoint, backup, mutate, verify, then either commit success or restore the prior checkpoint.
**When to use:** All writes to a running service stack.
**Trade-offs:** Slower than blind `pull && up`, but it is auditable, interrupt-safe, and compatible with future TUI or API surfaces.

**Example:**
```go
type Step interface {
    Name() string
    Run(ctx context.Context, s *State) error
    Compensate(ctx context.Context, s *State) error
}
```

## Data Flow

### Upgrade Action Flow

```text
[Operator runs upgrade]
    ↓
[CLI command handler]
    ↓
[Host/project lock]
    ↓
[Discovery]
    ├→ docker compose ls
    ├→ docker compose config --format json --hash --images --volumes
    ├→ docker compose ps --format json
    └→ docker inspect (mounts, labels, current images)
    ↓
[Preflight + plan]
    ├→ permission checks
    ├→ disk/headroom checks
    ├→ backup coverage checks
    └→ probe/hook validation
    ↓
[Checkpoint + backup]
    ├→ resolve image digests / write lock file
    ├→ run quiesce hooks
    ├→ back up volumes and bind mounts
    └→ optionally docker image save critical images
    ↓
[Upgrade execution]
    ├→ docker compose pull
    └→ docker compose up --wait
    ↓
[Verification]
    ├→ docker compose ps --format json
    ├→ docker compose events --json
    └→ custom HTTP/TCP/command probes
    ↓
 success ───────────────→ [Persist success summary + retain checkpoint]
    │
 failure
    ↓
[Rollback executor]
    ├→ restore prior compose config + digest lock
    ├→ restore images/data artifacts
    └→ docker compose up --wait
    ↓
[Persist failure + rollback result]
```

### Recovery Flow

```text
[Operator selects historical checkpoint]
    ↓
[CLI loads release bundle + artifacts]
    ↓
[SSH session to target host]
    ↓
[Optional pre-restore hooks]
    ↓
[Restore storage surfaces + images]
    ↓
[Re-apply pinned compose config]
    ↓
[docker compose up --wait]
    ↓
[Health verification + recovery report]
```

### State Management

```text
[SQLite run store]
    ↓ append
[Command handlers] → [Saga steps] → [Run events + artifact refs] → [SQLite run store]
    ↑                                                         ↓
[history / inspect / recover commands] ← [materialized summaries + artifact index]
```

### Key Data Flows

1. **Discovery baseline:** remote Compose metadata and live container state are normalized into a project snapshot before any change is allowed.
2. **Release checkpoint:** canonical config, env resolution, digests, mount inventory, and backup artifact refs become the rollback source of truth.
3. **Execution telemetry:** stdout/stderr, Compose events, probe results, and step timings stream back to the local journal in real time.
4. **Recovery artifacts:** data archives and image archives move through a local artifact store, not ephemeral in-memory state.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1-10 hosts / 1-50 projects | Single local binary, SQLite, local artifact directory, one workflow per project lock is enough |
| 10-100 hosts / 50-500 projects | Add bounded worker pools, SSH session reuse, artifact retention policies, optional remote artifact store |
| 100+ hosts / fleet operations | Keep CLI as client but move orchestration into a separate controller service only if a real multi-user UI appears; do not start here |

### Scaling Priorities

1. **First bottleneck:** backup transfer time and artifact storage size. Fix with per-surface adapters, compression, retention, and optional host-native snapshots for large datasets.
2. **Second bottleneck:** connection churn and serialized host work. Fix with per-host worker pools and SSH connection reuse or multiplexing.

## Anti-Patterns

### Anti-Pattern 1: Blind `pull && up` Against Mutable Tags

**What people do:** pull the newest image tags and recreate containers without storing the prior resolved release.
**Why it's wrong:** rollback becomes non-deterministic because the previous tag may now point to a different digest.
**Do this instead:** persist a release bundle with pinned digests before mutation, and optionally save images needed for offline recovery.

### Anti-Pattern 2: Treating the Compose File as the Only Source of Truth for Data

**What people do:** inspect only YAML and assume that covers all persistent state.
**Why it's wrong:** live containers may have anonymous volumes, renamed projects, or bind mounts that are only visible at runtime; Docker documents that anonymous volumes are not reattached predictably by a later `up`.
**Do this instead:** merge resolved Compose config with live `docker inspect` mount data and fail closed on unsupported or unprotected storage surfaces.

### Anti-Pattern 3: Long-Lived Privileged Agents on Every Host

**What people do:** install a resident root daemon to manage upgrades.
**Why it's wrong:** it expands the attack surface, complicates upgrades of the tool itself, and weakens the OSS “easy to trust” story.
**Do this instead:** keep the control plane local and use SSH plus ephemeral working directories on the host.

### Anti-Pattern 4: Declaring Success When Containers Are Merely Running

**What people do:** treat “container started” as equivalent to “service healthy.”
**Why it's wrong:** Compose startup order and health checks exist because process start does not guarantee readiness.
**Do this instead:** gate success on Compose health/readiness plus app-specific probes and only then finalize the run.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| OpenSSH server | Command execution and optional file copy over SSH | Reuse host keys and existing SSH agent/config where possible |
| Docker Engine + Compose plugin | Remote CLI operations | Keep upgrade semantics at the Compose layer, not raw container mutations |
| Container registry | Pull by tag, record digest, optional fallback to saved image archive | Network failure must trigger rollback, not partial success |
| Backup target | Local artifact dir first, optional external destination via adapter | Every artifact should be checksummed and referenced from the run journal |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| CLI ↔ use cases | Typed command DTOs | Command parsing should not know shell details |
| Use cases ↔ orchestrator | Typed plan and step interfaces | Keeps rollback logic reusable across CLI/TUI/API |
| Orchestrator ↔ transport | Command/result/stream abstractions | Domain code should not build raw SSH strings |
| Orchestrator ↔ Compose adapter | Typed operations like `Discover`, `Pull`, `UpWait`, `Ps`, `Events` | Centralizes Docker version quirks and output parsing |
| Orchestrator ↔ backup adapters | Storage manifest + artifact refs | Adapters must report their consistency guarantees |
| Store ↔ reporting | Append-only events and materialized summaries | Preserve raw audit data while keeping reads fast |
| Core repo ↔ docs site | Generated command/reference JSON and markdown | Docs remain static and disconnected from production credentials |

## Build Order

1. **Core domain + journaling:** define host, project, release, run-event, and policy models; add SQLite-backed run history and artifact index first.
2. **SSH transport + locking:** implement reliable remote command execution, host/project locks, timeout handling, and streamed logs.
3. **Discovery + preflight:** ship `discover` and `plan` commands using `docker compose ls/config/ps` and `docker inspect`.
4. **Immutable checkpointing:** add config canonicalization, digest pinning, service hashes, and release bundle persistence before any write path.
5. **Backup/restore baseline:** support named volume backups, bind mount backups, hook execution, and optional image save/load for offline rollback.
6. **Upgrade saga:** add pull, recreate, wait, probe, and compensating rollback logic with clear step-level audit output.
7. **Recovery commands:** let operators list checkpoints, inspect artifacts, dry-run a restore, and execute full recovery from a chosen checkpoint.
8. **Docs site + examples:** publish install docs, trust model, recovery recipes, and failure drills after the CLI contracts stabilize.

## Sources

- Docker Docs: Protect the Docker daemon socket. https://docs.docker.com/engine/security/protect-access/
- Docker Docs: Post-installation steps for Linux. https://docs.docker.com/engine/install/linux-postinstall/
- Docker Docs: Rootless mode. https://docs.docker.com/engine/security/rootless/
- Docker Docs: Live restore. https://docs.docker.com/engine/daemon/live-restore/
- Docker Docs: Specify a project name. https://docs.docker.com/compose/how-tos/project-name/
- Docker Docs: Version and name top-level elements. https://docs.docker.com/reference/compose-file/version-and-name/
- Docker Docs: Control startup order. https://docs.docker.com/compose/how-tos/startup-order/
- Docker Docs: `docker compose ls`. https://docs.docker.com/reference/cli/docker/compose/ls/
- Docker Docs: `docker compose config`. https://docs.docker.com/reference/cli/docker/compose/config/
- Docker Docs: `docker compose up`. https://docs.docker.com/reference/cli/docker/compose/up/
- Docker Docs: `docker compose ps`. https://docs.docker.com/reference/cli/docker/compose/ps/
- Docker Docs: `docker compose events`. https://docs.docker.com/reference/cli/docker/compose/events/
- Docker Docs: `docker compose down`. https://docs.docker.com/reference/cli/docker/compose/down/
- Docker Docs: `docker compose exec`. https://docs.docker.com/reference/cli/docker/compose/exec/
- Docker Docs: `docker compose run`. https://docs.docker.com/reference/cli/docker/compose/run/
- Docker Docs: Volumes. https://docs.docker.com/engine/storage/volumes/
- Docker Docs: Bind mounts. https://docs.docker.com/engine/storage/bind-mounts/
- Docker Docs: `docker inspect`. https://docs.docker.com/reference/cli/docker/inspect/
- Docker Docs: `docker image save`. https://docs.docker.com/reference/cli/docker/image/save/
- Docker Docs: `docker image load`. https://docs.docker.com/reference/cli/docker/image/load/
- OpenBSD manual: `ssh(1)`. https://man.openbsd.org/ssh
- OpenBSD manual: `ssh_config(5)`. https://man.openbsd.org/ssh_config

---
*Architecture research for: self-hosted Docker Compose upgrade manager*
*Researched: 2026-03-11*
