# Stack Research

**Domain:** Self-hosted Linux upgrade, backup, rollback, and recovery tooling for Docker Compose hosts
**Researched:** 2026-03-11
**Confidence:** HIGH

## Recommended Stack

### Prescriptive Summary

| Area | Recommendation | Confidence | Current-Version Awareness |
|------|----------------|------------|---------------------------|
| Runtime | Go `1.26.0` | HIGH | `go.dev` listed Go 1.26.0 as current stable on 2026-03-11 |
| Host orchestration | Agentless SSH using OpenSSH `10.x`, executing remote `docker compose` CLI workflows | HIGH | Compose v5.0.1 is current, but many self-hosted boxes still run Compose v2; feature-detect capabilities rather than assuming latest everywhere |
| Compose control plane | Prefer remote `docker compose` CLI; use Docker Engine API v1.52 selectively for inspect/events/health detail | HIGH | Docker Engine docs exposed API v1.52; Compose SDK exists in v5 but is still a sharper dependency surface than the CLI for v1 |
| State storage | SQLite `3.49.1` in WAL mode, local to the CLI/controller | HIGH | SQLite 3.49.1 was current on sqlite.org at research time |
| Backup engine | `restic 0.18.1` via hook-based integration | HIGH | restic 0.18.1 stable docs and release notes were current at research time |
| Packaging | Single static Go binaries released with GoReleaser `v2.14`, plus `deb`, `rpm`, and tarballs | MEDIUM-HIGH | GoReleaser v2.14 docs/blog were current at research time |
| Docs site | Astro `6.0` + Starlight `0.37.0` on Node `24` LTS | MEDIUM-HIGH | Astro 6 docs were current; Node 24 was Active LTS per nodejs.org schedule |

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.0 | Core CLI/runtime and any optional remote helper binary | This is the standard 2026 choice for systems-oriented CLIs: strong concurrency, easy cross-compilation, low memory overhead, and single-binary distribution without a language runtime on the host. |
| OpenSSH client | 10.x | Remote transport and command execution | OpenSSH remains the standard, battle-tested control plane for Linux host automation. It handles host keys, agent auth, ProxyJump, config files, and multiplexing better than most embedded SSH stacks. |
| Docker Compose CLI | 5.0.1 target, support Compose v2 installations via capability detection | Upgrade/pull/up/down/config/ps operations on remote hosts | The product is Compose-first, so the safest v1 is to drive the official control surface already installed on the host instead of re-implementing Compose semantics. |
| Docker Engine API | v1.52 | Low-level inspect, events, image digest, container health, and daemon detail | Use it where typed state is better than parsing CLI output, but do not make it the only orchestration layer for Compose project lifecycle actions. |
| SQLite | 3.49.1 | Local durable run history, host inventory, rollback metadata, and artifact index | Single-file transactional storage fits a CLI-first OSS tool. WAL mode gives good durability and concurrency without introducing a database service dependency. |
| restic | 0.18.1 | Snapshot-based backup and restore backend | Official support for local, SFTP, and S3-compatible repositories makes it a strong default for self-hosters. Deduplication, tags, retention, and restore workflows are already solved. |
| Astro | 6.0 | Static landing/docs site framework | Excellent fit for docs-first OSS sites: content-centric, static by default, fast to ship, and lower maintenance than a full app framework. |
| Starlight | 0.37.0 | Astro docs system | Gives a polished docs IA, search, nav, and MD/MDX authoring without turning the public site into a React app that must be operated like product infrastructure. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/spf13/cobra` | 1.10.2 | CLI command tree, flags, help, completion | Use for the main operator-facing CLI and subcommands such as `discover`, `plan`, `upgrade`, `rollback`, `recover`, and `history`. |
| `modernc.org/sqlite` | 1.44.3 | Pure-Go SQLite driver | Use as the default driver so release builds stay CGO-free and packaging stays simple across Linux/macOS targets. |
| `database/sql` | Go stdlib 1.26.0 | DB access layer | Use with explicit SQL and migrations rather than introducing an ORM. The schema will be operational metadata, not product business data. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| GoReleaser v2.14 | Build, package, checksum, and publish releases | Use to emit signed tarballs plus `deb`/`rpm`. Keep the runtime install story binary-first. |
| `nfpm` via GoReleaser | Native Linux packages | Emit minimal packages that install the binary, shell completions, man pages, and example config. |
| Node 24 LTS | Docs-site toolchain runtime | Use only for the docs site build. Do not make Node a runtime dependency of the core product. |

## Architecture Decisions

### Runtime

Use Go for the core product and keep the product binary-first. The main control loop should be a local CLI that can also act as a controller for scripted automation later. If a remote helper is needed, make it an ephemeral static binary uploaded per run, not a resident agent service.

Why this is the standard fit:
- Static binaries simplify install, upgrades, and rollback of the tool itself.
- Go is well aligned with SSH/process orchestration, timeouts, streaming logs, JSON handling, and local SQLite usage.
- A greenfield reliability tool benefits more from operational simplicity than from language ergonomics alone.

### Orchestration Approach

Use agentless SSH and treat each Compose project upgrade as a transactional state machine:

1. Discover project metadata on the remote host.
2. Capture the pre-change manifest:
   Current Compose config via `docker compose config --format json`.
   Running container state via `docker compose ps --format json`.
   Image digests via Docker Engine inspect.
3. Run preflight checks.
4. Run backup hooks and record the resulting snapshot IDs.
5. Pull and apply the upgrade via remote `docker compose pull` and `docker compose up -d --wait`.
6. Run layered health verification.
7. If health fails, reapply the saved manifest and pinned digests, then restore data only when rollback requires it.

This is more robust than trying to model the whole product around the Docker Engine API alone. Compose project semantics already live in the Compose CLI. Use the Engine API only where it is clearly better: image inspection, health state, container events, and daemon info.

### Data Storage

Use a local SQLite database in WAL mode for:
- host definitions
- project discovery snapshots
- upgrade run history
- backup snapshot references
- rollback manifests
- event/log artifact indexes

Store bulky artifacts beside the database in a structured local artifact directory, for example:

```text
~/.local/state/<tool>/
  state.db
  artifacts/<run-id>/
    compose-config-before.json
    compose-ps-before.json
    image-digests-before.json
    upgrade.log
    health-report.json
```

Do not start with PostgreSQL. There is no multi-user server requirement in v1, and adding a service dependency would make install, local testing, backup, and OSS adoption worse.

### Backup Integration Patterns

Use a hook-based backup contract with restic as the default engine:

- `pre-backup`: app-consistency work such as `pg_dump`, `mysqldump`, or temporarily quiescing writes
- `snapshot`: restic backup of named volume mount paths, bind mounts, and generated dump artifacts
- `post-backup`: cleanup and resume writes
- `pre-restore` / `post-restore`: restore-time quiesce/repair hooks

Prescriptive pattern:
- Make backups explicit per project, not magical discovery-only behavior.
- Capture both filesystem data and an upgrade manifest that pins prior image digests and Compose config.
- Record the backup snapshot ID in the run record before any mutating upgrade step.
- Prefer restore-by-snapshot over ad hoc copy-back logic.

Support external repository types through restic’s official backends:
- local path
- SFTP
- S3-compatible object storage

Leave filesystem-native snapshot systems like ZFS/Btrfs/LVM as optional adapters later. They are valuable, but not portable enough to be the default v1 contract.

### SSH / Remote Execution

Use the system OpenSSH client, not a custom always-on agent and not a default in-process SSH library.

Recommended transport defaults:
- `BatchMode=yes`
- `StrictHostKeyChecking=yes`
- `ControlMaster=auto`
- `ControlPersist=60s`
- `ConnectTimeout` set explicitly
- `ServerAliveInterval` and `ServerAliveCountMax` set explicitly

Recommended execution pattern:
- Open one multiplexed SSH session per host per run.
- Upload a small run bundle to `/tmp/<tool>/<run-id>/`.
- Execute idempotent remote shell steps under `bash -euo pipefail`.
- Transfer structured JSON status back to the controller.
- Remove the remote temp bundle at the end unless `--debug-keep-remote` is set.

This keeps the product agentless while still allowing rich, replayable operations.

### Health-Check Strategy

Use layered health checks and do not trust any single signal:

1. Docker/Compose health:
   Require Compose projects to define `healthcheck` where possible.
   Use `docker compose up -d --wait` as the first readiness gate.
2. Service-level probes:
   Support HTTP, TCP, and command probes defined in the tool config.
3. Stabilization window:
   Re-check after the initial success window so crash loops or migrations that fail late still trip rollback.
4. Rollback gate:
   If any required probe fails within the upgrade budget, rollback automatically.

Do not use only “container is running” or “port is open” as your success criterion. Those signals are too weak for a product whose value proposition is safe upgrades.

### Packaging

Ship the core tool as:
- a statically linked tarball for Linux `amd64` and `arm64`
- `deb` and `rpm` packages for common distributions
- a Homebrew tap for local developer install

Packaging rules:
- no Python virtualenv requirement
- no Node runtime requirement
- no database service dependency
- no long-running daemon install in v1

Add checksums and signed artifacts from the start. This is a reliability/security product; unsigned release artifacts undercut the story.

### Docs Site Stack

Use Astro + Starlight for the public landing/docs site. Keep it static-first and separate from the Go codebase’s runtime concerns.

Why not Next.js here:
- The site is mostly docs, guides, landing pages, and reference content.
- Astro/Starlight gives lower operational overhead and cleaner markdown-first authoring.
- A static docs site is easier to host anywhere, including GitHub Pages, Netlify, Cloudflare Pages, or self-hosted object storage plus CDN.

## Installation

```bash
# Core CLI
go mod init github.com/<org>/<repo>
go get github.com/spf13/cobra@v1.10.2
go get modernc.org/sqlite@v1.44.3

# Packaging
go install github.com/goreleaser/goreleaser/v2@latest

# Docs site
npm create astro@latest docs
cd docs
npm install @astrojs/starlight
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Go 1.26.0 | Rust 1.87+ | Use Rust if the primary goal is maximum compile-time safety and you accept a slower iteration loop and a steeper contributor bar. |
| Agentless OpenSSH + remote Compose CLI | Embedded Compose SDK | Revisit once Compose SDK dependency stability improves and you want deeper in-process Compose control without shelling out. |
| SQLite 3.49.1 | PostgreSQL 17/18 | Use PostgreSQL only when the tool becomes a multi-user service with concurrent controllers and a shared server-side API. |
| restic 0.18.1 | Kopia | Use Kopia if repository UX or specific policy features fit your operators better, but restic is the simpler default for a first OSS release. |
| Astro 6 + Starlight 0.37.0 | Docusaurus 3 | Use Docusaurus if the docs site will become React-heavy and needs deeper plugin ecosystem support. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Python as the core runtime | Virtualenvs, interpreter drift, and native-extension packaging make binary-style OSS distribution weaker for a host automation product. | Go 1.26.0 |
| Kubernetes/operator architecture in v1 | The project’s wedge is Compose-host safety. Kubernetes would expand scope, dilute the value proposition, and move the control plane away from the actual user problem. | Agentless SSH + remote Compose CLI |
| Compose SDK as the only orchestration layer today | Official docs note current dependency sharp edges, including pinning `docker/cli` 28.5.2 because Compose v5 still depends on `github.com/docker/docker`. That is too fragile for a v1 core dependency. | Remote `docker compose` CLI plus selective Engine API use |
| `mattn/go-sqlite3` as the default SQLite driver | CGO complicates static builds, cross-compilation, and release packaging. | `modernc.org/sqlite` |
| Raw `tar`/`rsync` as the only backup story | No snapshot metadata, weak retention/deduplication, and poor restore discipline for app-consistent recovery. | restic with explicit backup/restore hooks |
| `latest` image tags in upgrade plans | They destroy reproducibility and make rollback ambiguous. | Pin and record immutable image digests |
| “Container running” as the only health check | It misses broken migrations, bad readiness, and partial boot failures. | Compose healthchecks plus HTTP/TCP/command probes and stabilization windows |
| Next.js for the docs site by default | It adds app-framework complexity to a mostly static docs problem. | Astro + Starlight |

## Stack Patterns by Variant

**If v1 stays single-user CLI-only:**
- Use local SQLite plus local artifact storage.
- Because it keeps install friction near zero and matches the current product scope.

**If the project later grows into a shared coordinator service:**
- Keep Go, SSH transport, and run-state model, but move state storage to PostgreSQL and expose a server API.
- Because the coordination problem changes, but the host-side execution model does not need to.

**If target hosts commonly use ZFS/Btrfs/LVM snapshots:**
- Add optional snapshot adapters ahead of restic upload.
- Because filesystem-native snapshots can reduce backup windows, but they should be an optimization layer, not the default product contract.

**If remote hosts are highly locked down:**
- Keep agentless SSH, but rely on pre-installed shell tooling and restic binaries rather than uploading a helper.
- Because some operators will allow SSH but not foreign binary execution.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| Go 1.26.0 | Cobra 1.10.2, modernc.org/sqlite 1.44.3 | Good baseline for a static CLI and CGO-free packaging. |
| Docker Compose 5.0.1 | Docker CLI/Engine 28.5.x era | Best target for current hosts, but feature-detect and support Compose v2 installations in the field. |
| Compose SDK 5.0.0 | `docker/cli` 28.5.2 | Official docs explicitly call out this pin because Docker CLI 29.0.0 moved from `github.com/docker/docker` to `github.com/moby/moby`. This is why it should not be the v1 default orchestration core. |
| Astro 6.0 | Node >= 22.12.0 | Use Node 24 LTS for greenfield work in 2026 even though Astro’s minimum is lower. |
| restic 0.18.1 | Local, SFTP, and S3-compatible repos | Fits the likely self-hoster backup targets without adding service-specific code to the product. |

## Sources

- https://go.dev/ — verified current Go stable release (`1.26.0`)
- https://docs.docker.com/compose/releases/release-notes/ — verified current Compose release line and Compose v5 availability
- https://github.com/docker/compose/releases — verified Compose `v5.0.1` current release tag
- https://docs.docker.com/reference/cli/docker/compose/up/ — verified `--wait` support for post-upgrade readiness gating
- https://docs.docker.com/reference/cli/docker/compose/config/ — verified structured Compose config export via `--format json`
- https://docs.docker.com/reference/cli/docker/compose/ps/ — verified structured service state output via `--format json`
- https://docs.docker.com/reference/api/engine/ — verified Docker Engine API current version (`v1.52`)
- https://docs.docker.com/compose/compose-sdk/ — verified current Compose SDK caveats and dependency pin note
- https://www.openssh.com/releasenotes.html — verified OpenSSH current release line (`10.x`)
- https://man.openbsd.org/ssh_config — verified `BatchMode`, `ControlMaster`, `ControlPersist`, and host-key related SSH options
- https://www.sqlite.org/index.html — verified SQLite current release (`3.49.1`)
- https://restic.readthedocs.io/en/stable/ — verified restic stable docs, repository model, backup/restore concepts, and backend support
- https://github.com/restic/restic/releases — verified restic `0.18.1` release
- https://goreleaser.com/blog/v2.14/ — verified current GoReleaser release (`v2.14`)
- https://docs.astro.build/en/getting-started/ — verified Astro 6 current docs and runtime requirements
- https://github.com/withastro/starlight/releases — verified Starlight `0.37.0` release
- https://nodejs.org/en/about/previous-releases — verified Node 24 Active LTS status
- https://pkg.go.dev/github.com/spf13/cobra — verified Cobra current package version (`v1.10.2`)
- https://pkg.go.dev/modernc.org/sqlite — verified modernc SQLite driver current package version (`v1.44.3`)

---
*Stack research for: self-hosted Docker Compose upgrade/backup/rollback/recovery tooling*
*Researched: 2026-03-11*
