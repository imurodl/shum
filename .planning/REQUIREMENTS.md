# Requirements: Self-Host Upgrade Manager

**Defined:** 2026-03-11
**Core Value:** Self-hosters can upgrade Docker Compose apps with confidence because every change is checked, backed up, verified, and reversible.

## v1 Requirements

### Hosts & Discovery

- [ ] **HOST-01**: User can register a Linux host over SSH with verified host key handling
- [ ] **HOST-02**: User can discover Docker Compose projects and services on a registered host
- [ ] **HOST-03**: User can inspect the canonical rendered Compose configuration, project identity, and storage surfaces before an upgrade

### Planning & Safety

- [ ] **PLAN-01**: User can run a preflight check that validates Docker/Compose availability, permissions, disk headroom, and external dependency readiness
- [ ] **PLAN-02**: User can preview an upgrade plan showing current and target image digests before any change is applied
- [ ] **PLAN-03**: User can see which backups, hooks, probes, and rollback boundaries apply to an upgrade run
- [ ] **PLAN-04**: Tool blocks upgrade execution when required prerequisites, backup coverage, or verification policy are missing

### Backups & Recovery

- [ ] **BKUP-01**: User can configure project-specific backup hooks for stateful services before upgrade
- [ ] **BKUP-02**: User can create and record a backup artifact for a project before a mutating upgrade step
- [ ] **BKUP-03**: User can restore a recorded backup artifact during recovery when image rollback alone is insufficient

### Upgrades & Verification

- [ ] **UPGD-01**: User can execute an upgrade for a Docker Compose project on a target host
- [ ] **UPGD-02**: Tool verifies upgrade success using Compose health status plus optional HTTP, TCP, or command probes
- [ ] **UPGD-03**: Tool can roll back to the previous Compose config and image digests when verification fails
- [ ] **UPGD-04**: Tool flags migration-bearing or non-reversible upgrades before execution so the user can make an explicit decision

### History

- [ ] **HIST-01**: User can inspect run history with timestamps, actions, outcomes, and artifact references
- [ ] **HIST-02**: User can inspect failure evidence including logs, config snapshot, event output, and health results for a run

### Docs Site

- [ ] **DOCS-01**: User can learn what the tool does, how it works, and how to install it from a public landing/docs site
- [ ] **DOCS-02**: User can follow documented setup, backup, upgrade, rollback, and recovery recipes from the public docs site

## v2 Requirements

### Automation & Scale

- **AUTO-01**: User can run scheduled monitor-only scans for update availability and safety status
- **AUTO-02**: User can receive notifications or generated upgrade plans without auto-applying changes
- **FLEET-01**: User can manage multiple hosts with fleet-level policies and reporting

### Platform Extensions

- **PLAT-01**: User can use filesystem-specific snapshot adapters such as Btrfs or ZFS
- **PLAT-02**: User can use a lightweight web UI or remote API for viewing history and triggering runs
- **PLAT-03**: User can enforce provenance or SBOM policy gates before upgrades
- **PLAT-04**: User can manage non-Compose targets such as systemd services or Kubernetes workloads

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Unattended auto-apply updates by default | Conflicts with the core promise of deliberate, safe, reversible upgrades |
| General-purpose monitoring or observability suite | Expands the product into a crowded category and dilutes the wedge |
| Full PaaS or deployment platform | The product is a safety layer for existing self-hosted services, not a replacement hosting stack |
| Mandatory privileged agent on every host | Increases install and trust complexity before the core workflow is proven |
| Kubernetes support in v1 | Broadens scope beyond the Compose-first target audience |
| systemd-only service management in v1 | Adds a second operational model too early |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| HOST-01 | Phase 1 | Planned |
| HOST-02 | Phase 1 | Planned |
| HOST-03 | Phase 1 | Planned |
| PLAN-01 | Phase 2 | Planned |
| PLAN-02 | Phase 2 | Planned |
| PLAN-03 | Phase 2 | Planned |
| PLAN-04 | Phase 2 | Planned |
| BKUP-01 | Phase 2 | Planned |
| BKUP-02 | Phase 2 | Planned |
| BKUP-03 | Phase 2 | Planned |
| UPGD-01 | Phase 3 | Planned |
| UPGD-02 | Phase 3 | Planned |
| UPGD-03 | Phase 3 | Planned |
| UPGD-04 | Phase 3 | Planned |
| HIST-01 | Phase 4 | Planned |
| HIST-02 | Phase 4 | Planned |
| DOCS-01 | Phase 4 | Planned |
| DOCS-02 | Phase 4 | Planned |

**Coverage:**
- v1 requirements: 18 total
- Mapped to phases: 18
- Unmapped: 0
- Duplicate mappings: 0
- Coverage: 100%

---
*Requirements defined: 2026-03-11*
*Last updated: 2026-03-11 after roadmap creation*
