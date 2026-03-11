# Project Research Summary

**Project:** Self-Host Upgrade Manager
**Domain:** Self-hosted Linux and Docker Compose safe-upgrade, backup, rollback, and recovery tooling
**Researched:** 2026-03-11
**Confidence:** MEDIUM

## Executive Summary

This project should be built as a CLI-first, agentless reliability tool for self-hosters running Docker Compose on Linux. The research consistently points toward a local control plane that connects over SSH, discovers the canonical Compose state on the remote host, captures a deterministic pre-change checkpoint, runs backup hooks, performs the upgrade, verifies application health, and rolls back when the result is not healthy. That is a narrower and more defensible wedge than building a broad self-hosted PaaS, a monitoring product, or an auto-updater.

The recommended implementation is a Go core with agentless SSH transport, remote `docker compose` execution as the main orchestration surface, selective Docker Engine API use for inspect/health/events, SQLite for local run state, and `restic` as the default backup integration point. The main risks are not frontend or packaging concerns; they are operational correctness: incomplete Compose context discovery, mutable image tags, weak health criteria, backups that are not actually recoverable, and rollback paths that ignore schema or data migrations. The roadmap should be shaped around those risks from the start.

- What this product is: a safe-change control layer for self-hosted Docker Compose services
- Recommended approach: agentless SSH, deterministic checkpoints, health-gated upgrade saga, explicit backup/restore contracts
- Main risks: Compose context drift, mutable tags, fake health success, non-recoverable backups, migration-blind rollback

## Key Findings

### Recommended Stack

The strongest stack for this product is a single-binary Go application with a static public docs site. Research favored a local CLI/controller instead of a daemon or host agent because it keeps the privilege model legible, simplifies distribution, and matches how self-hosters already operate. The product should drive the official remote `docker compose` CLI for lifecycle actions and use the Docker Engine API only where structured state is materially better than shell output.

For durability and operator trust, local state should live in SQLite with artifact files stored beside it. Backups should use a hook-based contract with `restic` as the default off-host-capable engine rather than trying to invent a new snapshot system. The public site should be static-first with Astro + Starlight so docs quality is high without turning the project website into another full application to operate.

**Core technologies:**
- `Go` — core CLI/runtime and orchestration engine; strongest fit for cross-platform binary distribution and systems work
- `OpenSSH` + remote `docker compose` — trusted host transport and operational control surface; avoids inventing a host agent
- `SQLite` — local durable state for run history, rollback metadata, and host/project inventory
- `restic` — default backup engine and repository contract; strong off-host backup story without custom storage code
- `Astro + Starlight` — landing/docs site; static-first and well suited for OSS product documentation

### Expected Features

The table-stakes set is clear: SSH-based host access, canonical Compose discovery, preflight validation, explicit dry-run planning, image digest capture, health-aware execution, deterministic rollback, and durable run history. Without these, the tool is just an unsafe updater with extra steps.

The differentiators are where the product becomes impressive: transaction-style upgrade runs with checkpoints, evidence bundles for failures, backup adapters that understand storage surfaces, and external health profiles for services that do not define `HEALTHCHECK`. Research also strongly suggests what not to build in v1: unattended auto-apply updates, a general monitoring platform, a mandatory privileged agent, or a full app-deployment/PaaS workflow.

**Must have (table stakes):**
- SSH host access and Compose project discovery — users expect agentless Linux automation
- Preflight validation and dry-run upgrade planning — the tool must show what it will do before mutating anything
- Image digest capture and deterministic rollback metadata — rollback cannot depend on mutable tags
- Health-aware upgrade execution — success must mean healthy, not merely running
- Durable run history and evidence — operators need a post-change audit trail

**Should have (competitive):**
- Transaction-style upgrade checkpoints — makes risky changes inspectable and resumable
- Backup/snapshot adapters with explicit consistency rules — turns rollback into real recovery for stateful apps
- External health profiles and verification packs — improves compatibility with imperfect community Compose stacks
- Failure evidence bundles — useful for debugging and strong OSS product polish

**Defer (v2+):**
- Multi-host orchestration and fleet policies — too much concurrency and blast-radius complexity for first release
- Full web dashboard or remote API — secondary until the CLI workflow is correct and trusted
- SBOM/provenance policy gates — valuable later, but not core to proving the safe-upgrade loop

### Architecture Approach

The recommended architecture has three clear layers: a low-privilege local product surface, a trusted local control plane, and a privileged remote execution plane on the target host. The local side owns CLI commands, state storage, artifacts, policies, and the workflow engine. The remote side owns Compose operations, hook execution, backup/restore interactions, and health verification. The core execution model should be a health-gated saga: discover, preflight, checkpoint, back up, mutate, verify, and either commit success or perform compensating rollback.

**Major components:**
1. Discovery and preflight engine — normalize Compose state, validate prerequisites, and build the upgrade plan
2. Upgrade saga orchestrator — execute ordered upgrade/rollback steps with locks, checkpoints, and compensation
3. Backup and restore adapters — handle volume, bind mount, and optional filesystem/database backup flows
4. State and artifact store — persist history, checkpoints, digests, logs, and evidence bundles
5. Public docs site — explain installation, trust model, backup recipes, and command usage

### Critical Pitfalls

1. **Incomplete Compose context discovery** — always store the rendered `docker compose config`, canonical project identity, active profiles, and relevant env context before any change
2. **Mutable tags causing non-deterministic upgrades** — resolve and persist image digests and config hashes before applying or rolling back
3. **Treating running as healthy** — require layered verification with Compose health plus app-level probes and stabilization windows
4. **Backups that are not actually recoverable** — distinguish crash-consistent vs app-consistent backups, require explicit backup hooks, and prefer off-host artifacts
5. **Rollback that ignores data migrations** — classify rollback safety before execution and require pre-migration backup strategy for stateful upgrades

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Host Connectivity and Canonical Discovery
**Rationale:** The rest of the product is unsafe if the tool cannot identify the real Compose project and host capabilities correctly.
**Delivers:** SSH host management, project discovery, rendered Compose state capture, host/project identity, and local state persistence.
**Addresses:** SSH access, Compose discovery, baseline run state.
**Avoids:** Incomplete Compose context discovery.

### Phase 2: Deterministic Planning and Safety Gates
**Rationale:** Before any mutation, the tool must produce a trustworthy plan with digests, dependency checks, and dry-run visibility.
**Delivers:** Preflight engine, image digest capture, config hashing, external dependency checks, and explicit plan output.
**Uses:** Docker Compose config and dry-run capabilities, local checkpoint storage.
**Implements:** Discovery-to-plan transition in the orchestration layer.

### Phase 3: Backup and Recovery Foundations
**Rationale:** The product promise is not credible without a data-safety story for stateful services.
**Delivers:** Hook contract, backup manifesting, `restic` integration, consistency classification, and artifact recording.
**Addresses:** Backup/snapshot hooks, rollback prerequisites.
**Avoids:** Non-recoverable backups and fake recovery claims.

### Phase 4: Upgrade Execution, Health Verification, and Rollback
**Rationale:** Once plans and backups exist, the core value can be delivered through a safe execution engine.
**Delivers:** Upgrade saga, health-gated rollout, compensating rollback, and failure evidence bundles.
**Uses:** SSH transport, remote `docker compose`, health probes, checkpoint artifacts.
**Implements:** Main orchestration workflow.

### Phase 5: History, Packaging, and Public Docs
**Rationale:** A serious OSS tool needs inspectability, installability, and a public explanation of its trust model.
**Delivers:** History/reporting commands, packaged releases, install docs, landing site, architecture docs, and operator recipes.
**Addresses:** Durable run history, adoption, interview/showcase quality.
**Avoids:** A technically solid core that still looks unfinished to users and recruiters.

### Phase Ordering Rationale

- Discovery and planning come first because wrong target state invalidates every later safety guarantee.
- Backup foundations precede upgrade execution because rollback for stateful services must be designed before mutation logic ships.
- Execution follows only after deterministic planning and recovery prerequisites exist, which keeps the core product promise honest.
- Packaging and docs come after the core workflow is real, but still belong in the first milestone because OSS usability is part of the product.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3:** Backup and Recovery Foundations — app-consistent backup strategies vary by service and need implementation-specific decisions
- **Phase 4:** Upgrade Execution, Health Verification, and Rollback — migration-aware rollback and host-version compatibility need careful test design

Phases with standard patterns (skip research-phase):
- **Phase 1:** Host Connectivity and Canonical Discovery — established SSH and Compose patterns with strong official docs
- **Phase 5:** History, Packaging, and Public Docs — packaging and docs tooling are well understood

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Strongly grounded in official Docker, Go, SQLite, restic, and Astro docs |
| Features | HIGH | Table stakes and anti-features align well with documented platform primitives and adjacent tool behavior |
| Architecture | MEDIUM | Core structure is sound, but some workflow boundaries are still product-specific inference |
| Pitfalls | HIGH | Failure modes are directly supported by Docker, PostgreSQL, restic, and filesystem documentation |

**Overall confidence:** MEDIUM

### Gaps to Address

- Backup adapter depth — decide during planning which backup surfaces are first-class in v1 and which remain hook-only
- Rollback policy for migration-heavy services — make reversibility explicit rather than pretending all upgrades can be rolled back safely
- Discovery on messy real-world hosts — validate assumptions against varied Compose layouts and naming conventions during implementation

## Sources

### Primary (HIGH confidence)
- https://docs.docker.com/reference/cli/docker/compose/ — Compose lifecycle commands, JSON output, dry-run, and health-aware operations
- https://compose-spec.github.io/compose-spec/spec.html — Compose model semantics and configuration boundaries
- https://docs.docker.com/engine/security/protect-access/ — remote Docker/SSH trust model
- https://www.postgresql.org/docs/current/backup-file.html — filesystem-level backup constraints for stateful services
- https://restic.readthedocs.io/en/latest/ — backup, restore, and repository integrity model
- https://go.dev/ — Go runtime baseline and release status

### Secondary (MEDIUM confidence)
- https://containrrr.dev/watchtower/ — adjacent updater feature patterns and tradeoffs
- https://crazymax.dev/diun/ — adjacent notification/change-detection patterns
- https://openzfs.github.io/openzfs-docs/ — optional future snapshot adapter direction

### Tertiary (LOW confidence)
- None used intentionally; research prioritized primary docs and well-documented adjacent tools

---
*Research completed: 2026-03-11*
*Ready for roadmap: yes*
