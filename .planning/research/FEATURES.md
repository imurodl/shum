# Feature Research

**Domain:** Self-hosted Linux + Docker Compose reliability and safe-upgrade tooling
**Researched:** 2026-03-11
**Confidence:** HIGH

Competitive framing below is inferred from the documented Docker/Compose/OpenSSH primitives, adjacent tool docs, and the project wedge in `PROJECT.md`.

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Secure remote host access over SSH | Self-hosters expect agentless Linux automation over SSH, not a bespoke control plane | MEDIUM | Build around SSH-hosted Docker access and standard SSH config/known-hosts behavior. Requires host-key policy, connection reuse, and a remote user that can access the Docker socket. |
| Compose project discovery and canonical config parsing | A Compose-first tool must identify projects/services and understand the real merged model before changing anything | HIGH | `docker compose ls` helps for known/running projects, but full discovery usually also needs remote filesystem scanning plus `docker compose config` to resolve files, profiles, paths, and environment interpolation. |
| Preflight validation and dry-run planning | A “safe upgrade” tool is not credible without showing what it will do before it does it | MEDIUM | Needs config validation, feature/version checks, image/build feasibility checks, and a dry-run path using Compose’s global `--dry-run`. |
| Image change detection and digest capture | Users need to know what image actually changed and what the previous deploy was | MEDIUM | Capture candidate updates with `docker compose pull` and pin or snapshot image state with `docker compose config --lock-image-digests` / `--resolve-image-digests`. Built services are harder because rollback may not be reproducible without stored image IDs/artifacts. |
| Health-aware execution and dependency ordering | Operators expect the tool to wait for the app to come back healthy instead of only checking “container started” | HIGH | Compose supports `healthcheck`, `depends_on.condition=service_healthy`, `restart: true`, `docker compose up --wait`, `events`, and `ps --format json`. External probes are still needed because many community Compose files omit healthchecks. |
| Backup/snapshot hooks | Stateful apps need pre-upgrade dumps, quiesce steps, or filesystem snapshots before image replacement | HIGH | Compose has no generic backup primitive, so this must be hook-based. Hook failure should block the upgrade; otherwise the “safe” promise is hollow. |
| Deterministic rollback | Rollback is central to the project brief and is the main reason to choose this tool over auto-updaters | HIGH | Compose has no native rollback operation, so the tool must store pre-upgrade config, image digests/IDs, and optional data snapshot references, then redeploy the prior state. |
| Run history and audit evidence | Self-hosters need to inspect what happened after a risky change | MEDIUM | Persist plans, timestamps, old/new image state, hook output, `docker compose events --json`, `docker compose ps --format json`, and final health/result summaries. |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Transaction-style upgrade runs with checkpoints and resume points | Makes upgrades feel deliberate and reversible instead of “pull and hope” | HIGH | Inference from Compose dry-run/events plus the project wedge. Persist a state machine around discovery, backup, pull, rollout, verify, and rollback so a failed run is inspectable and resumable. |
| Adapter-based data safety layer for Btrfs/ZFS/database hooks | Makes rollback meaningful for stateful services, not just stateless containers | HIGH | Btrfs snapshots are fast but are not backups; ZFS offers atomic snapshots and explicit rollback/clone flows. Adapters should encapsulate stack-specific commands and safety checks. |
| External health profiles for apps that lack container healthchecks | Removes a common blocker in self-hosted Compose stacks, where upstream images often skip `HEALTHCHECK` | MEDIUM | Let users define HTTP/TCP/command probes outside the image and combine them with Compose-native `--wait` where healthchecks already exist. |
| Failure evidence bundles | Gives users something better than “upgrade failed”: config hash/diff, events, logs, inspect data, hook output | MEDIUM | Strong operational UX and a strong OSS showcase feature. Can back both local inspection and issue reports. |
| Approval-centric automation modes | Preserves the safe-upgrade position while still fitting scheduled maintenance workflows | MEDIUM | Offer monitor-only, scheduled discovery, or “generate plan now, apply later” modes instead of default unattended mutation. This captures some Watchtower/Diun convenience without collapsing into an auto-updater. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Unattended auto-apply updates by default | Users want a hands-off maintenance tool | It directly conflicts with the product’s safety wedge and recreates the risk profile of generic auto-updaters | Default to monitor-only or scheduled plan generation, then require explicit approval for apply |
| General-purpose monitoring/dashboard suite | “One pane of glass” sounds attractive | It pushes the product into a crowded, much broader category and dilutes the upgrade-safety wedge | Export metrics, webhooks, and docs for integration with existing monitoring tools |
| Mandatory privileged agent on every host | Seems like it would make discovery and scheduling easier | Adds packaging, upgrade, privilege, and cross-distro support burden before the core workflow is proven | Stay agentless over SSH first; add optional helpers later only if validated |
| Editing user Compose files or becoming an app-deployment platform | Users often ask tools to also manage deployment definitions | It turns the tool into a PaaS/config owner and creates conflict over user-managed files and conventions | Keep discovery read-only; generate lock files, plan artifacts, or explicit overrides instead |

## Feature Dependencies

```text
[Safe upgrade run]
    └──requires──> [SSH host access]
                       └──requires──> [Host key management + docker socket permission]

[Project discovery]
    └──requires──> [Compose config normalization]
                       └──requires──> [Path/profile/env resolution]

[Health-aware rollout]
    └──requires──> [Service health signals]
                       └──enhances──> [Compose healthcheck + depends_on.service_healthy]

[Rollback]
    └──requires──> [Pre-upgrade state capture]
                       └──requires──> [Image digest lock + config snapshot]

[Snapshot adapters] ──enhances──> [Rollback]
[Failure evidence bundle] ──enhances──> [Run history]
[Unattended auto-apply] ──conflicts──> [Approval-centric safe upgrades]
```

### Dependency Notes

- **Safe upgrade run requires SSH host access:** Docker documents SSH as a secure remote access path for the daemon; the product should lean on that instead of inventing its own transport.
- **Project discovery requires Compose config normalization:** the actual model can change based on merged files, `include`, profiles, path resolution, and env interpolation, so discovery cannot stop at raw YAML filenames.
- **Health-aware rollout requires service health signals:** `docker compose up --wait` is useful only when Compose healthchecks exist; external probes are needed for the many stacks that do not define them.
- **Rollback requires pre-upgrade state capture:** Compose does not provide a native “rollback this project” command, so the tool must persist the prior config/image state before mutation.
- **Snapshot adapters enhance rollback:** filesystem/database adapters are what turns rollback from “restart an older image” into a real recovery story for stateful apps.
- **Failure evidence bundle enhances run history:** durable artifacts make failures actionable and keep support/debugging from degenerating into ad hoc log scraping.
- **Unattended auto-apply conflicts with approval-centric safe upgrades:** the stronger the guarantee of explicit checks, backups, and reversibility, the weaker the case for silent autonomous mutation.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [ ] SSH-based host connection with secure host-key handling and Compose project discovery — core entry point for the target user
- [ ] Preflight validation plus explicit dry-run plan and image digest capture — makes change risk understandable before apply
- [ ] Safe upgrade execution with hooks, health-aware verification, and deterministic rollback to prior config/image state — core product promise
- [ ] Durable run history with structured evidence — required so operators can trust and debug the workflow

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] Scheduled monitor-only scans and notifications — add when users want regular awareness without giving up control
- [ ] Btrfs/ZFS/database backup adapters — add once manual/script hooks prove the workflow and the most common host stacks are clear
- [ ] External health profiles and reusable verification packs — add when lack of upstream healthchecks becomes a repeated adoption blocker

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Multi-host orchestration/fleet policies — defer because it multiplies state, concurrency, and blast-radius complexity
- [ ] Optional web UI or trigger API — defer until the CLI workflow is sharp and stable
- [ ] Provenance/SBOM policy gates for built services — valuable, but secondary to proving the safe-upgrade loop first

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| SSH host access + project discovery | HIGH | MEDIUM | P1 |
| Compose config normalization + preflight validation | HIGH | MEDIUM | P1 |
| Image change detection + digest locking | HIGH | MEDIUM | P1 |
| Health-aware upgrade + rollback | HIGH | HIGH | P1 |
| Run history + evidence | HIGH | MEDIUM | P1 |
| Scheduled monitor-only scans + notifications | MEDIUM | MEDIUM | P2 |
| Snapshot adapters for common Linux filesystems/datastores | HIGH | HIGH | P2 |
| External health profiles | MEDIUM | MEDIUM | P2 |
| Multi-host orchestration | MEDIUM | HIGH | P3 |
| Provenance/SBOM policy gates | MEDIUM | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Competitor A | Competitor B | Our Approach |
|---------|--------------|--------------|--------------|
| Change detection and triggering | [Watchtower](https://containrrr.dev/watchtower/) monitors containers, supports schedules, HTTP-triggered updates, and `--monitor-only` | [Diun](https://crazymax.dev/diun/) watches images/providers and sends notifications when tags move | Detect candidate changes per Compose project, capture current digests/config, and generate an explicit upgrade plan before mutation |
| Hook semantics | Watchtower supports lifecycle hooks, but hook failures only log and do not block the update | Diun is a notifier, not an upgrade executor | Treat hooks as first-class checkpoints that can fail the run and trigger rollback logic |
| Safe rollout behavior | Watchtower documents rolling restart and stop timeout controls | Diun stops at notification/awareness | Use Compose dependency data plus health-aware verification and rollback to manage project-level upgrades, not isolated container restarts |
| Visibility after failures | Watchtower exposes experimental metrics and runtime logs | Diun emphasizes notifications, logging, and provider coverage | Store durable run history and export evidence bundles, not only live metrics or one-shot alerts |

## Sources

- [Docker Compose CLI reference](https://docs.docker.com/reference/cli/docker/compose/)
- [docker compose config](https://docs.docker.com/reference/cli/docker/compose/config/)
- [docker compose up](https://docs.docker.com/reference/cli/docker/compose/up/)
- [docker compose pull](https://docs.docker.com/reference/cli/docker/compose/pull/)
- [docker compose ls](https://docs.docker.com/reference/cli/docker/compose/ls/)
- [docker compose ps](https://docs.docker.com/reference/cli/docker/compose/ps/)
- [docker compose events](https://docs.docker.com/reference/cli/docker/compose/events/)
- [Control startup and shutdown order in Compose](https://docs.docker.com/compose/how-tos/startup-order/)
- [Compose Specification](https://compose-spec.github.io/compose-spec/spec.html)
- [Compose profiles specification](https://compose-spec.github.io/compose-spec/15-profiles.html)
- [Protect the Docker daemon socket](https://docs.docker.com/engine/security/protect-access/)
- [docker system events](https://docs.docker.com/reference/cli/docker/system/events/)
- [Dockerfile reference (`HEALTHCHECK`)](https://docs.docker.com/reference/builder)
- [docker compose build (`--provenance`, `--sbom`)](https://docs.docker.com/reference/cli/docker/compose/build/)
- [OpenSSH `ssh_config(5)`](https://man.openbsd.org/ssh_config)
- [Btrfs subvolume and snapshot docs](https://btrfs.readthedocs.io/en/latest/btrfs-subvolume.html)
- [OpenZFS snapshot docs](https://openzfs.github.io/openzfs-docs/man/v2.2/8/zfs-snapshot.8.html)
- [OpenZFS rollback docs](https://openzfs.github.io/openzfs-docs/man/v2.3/8/zfs-rollback.8.html)
- [Watchtower overview](https://containrrr.dev/watchtower/)
- [Watchtower arguments](https://containrrr.dev/watchtower/arguments/)
- [Watchtower lifecycle hooks](https://containrrr.dev/watchtower/lifecycle-hooks/)
- [Watchtower metrics](https://containrrr.dev/watchtower/metrics/)
- [Watchtower HTTP API mode](https://containrrr.dev/watchtower/http-api-mode/)
- [Diun overview](https://crazymax.dev/diun/)
- [Diun Docker provider](https://crazymax.dev/diun/providers/docker/)

---
*Feature research for: self-hosted Linux + Docker Compose reliability and safe-upgrade tooling*
*Researched: 2026-03-11*
