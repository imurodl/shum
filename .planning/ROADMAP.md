# Roadmap: Self-Host Upgrade Manager

**Created:** 2026-03-11
**Config:** coarse granularity, yolo mode
**Phase count:** 4

## Roadmap Principles

- Group work by operational risk boundaries, not by generic app layers
- Finish recoverability prerequisites before shipping mutating upgrade execution
- Keep the first milestone CLI-first; public docs support adoption but do not drive architecture

## Phase 1: Remote Host Registration and Canonical Compose Discovery

**Why this phase exists:** Every later safety guarantee depends on identifying the real target host, Compose project, rendered config, and storage surfaces correctly.

**Requirements mapped here:** `HOST-01`, `HOST-02`, `HOST-03`

Status: implementation complete.

**Success criteria:**
- User registers a Linux host over SSH and later connections fail loudly on host key mismatch.
- User discovers Docker Compose projects and services on a registered host without manually reconstructing project metadata.
- User inspects the canonical rendered Compose configuration, project identity, and storage surfaces before starting any upgrade workflow.

## Phase 2: Deterministic Planning and Recovery Preconditions

**Why this phase exists:** The tool only becomes trustworthy when it can explain the exact change, prove prerequisites, and require recoverability before any mutation happens.

**Requirements mapped here:** `PLAN-01`, `PLAN-02`, `PLAN-03`, `PLAN-04`, `BKUP-01`, `BKUP-02`, `BKUP-03`

Status: implementation complete.

**Success criteria:**
- User runs a preflight check and sees pass/fail results for Docker or Compose availability, permissions, disk headroom, and external dependency readiness.
- User previews an upgrade plan showing current and target image digests before any change is applied.
- User configures project-specific backup hooks and can see which backups, probes, and rollback boundaries apply to a run.
- Tool blocks execution when prerequisites, verification policy, or required backup coverage are missing, and reports the blocking reason clearly.
- User can create a recorded backup artifact before a mutating step and restore that exact artifact when recovery requires more than image rollback.

## Phase 3: Health-Gated Upgrade Execution and Rollback Control

**Why this phase exists:** The core product promise is a safe upgrade saga that treats unhealthy or non-reversible changes as explicit operator risk, not as normal success cases.

**Requirements mapped here:** `UPGD-01`, `UPGD-02`, `UPGD-03`, `UPGD-04`

Status: implementation complete.

**Success criteria:**
- User executes an upgrade for a selected Docker Compose project on a target host from an approved plan.
- Tool declares success only after Compose health plus configured HTTP, TCP, or command probes pass within the expected stabilization window.
- When verification fails, tool rolls back to the previous Compose configuration and image digests instead of leaving the project partially upgraded.
- Tool flags migration-bearing or otherwise non-reversible upgrades before execution so the user must make an explicit decision to continue.

## Phase 4: Run Audit Trail and Public Operator Docs

**Why this phase exists:** Operators need a durable audit trail and clear public documentation before the tool is credible as a real OSS product.

**Requirements mapped here:** `HIST-01`, `HIST-02`, `DOCS-01`, `DOCS-02`

Status: implementation complete.

**Success criteria:**
- User lists past upgrade runs with timestamps, actions, outcomes, and artifact references.
- User opens a failed run and inspects logs, config snapshots, event output, and health results without reproducing the incident.
- A public landing or docs site explains what the tool does, how it works, and how to install it.
- User follows published setup, backup, upgrade, rollback, and recovery recipes end to end from the public docs site.

## Coverage Validation

| Phase | Requirement Count | Requirement IDs |
|-------|-------------------|-----------------|
| Phase 1 | 3 | `HOST-01`, `HOST-02`, `HOST-03` |
| Phase 2 | 7 | `PLAN-01`, `PLAN-02`, `PLAN-03`, `PLAN-04`, `BKUP-01`, `BKUP-02`, `BKUP-03` |
| Phase 3 | 4 | `UPGD-01`, `UPGD-02`, `UPGD-03`, `UPGD-04` |
| Phase 4 | 4 | `HIST-01`, `HIST-02`, `DOCS-01`, `DOCS-02` |

**Validation result (current):**
- Total v1 requirements: 18
- Mapped to exactly one phase: 18
- Unmapped requirements: 0
- Duplicate mappings: 0
- Coverage: 100%

## Ordering Rationale

- Phase 1 comes first because incorrect host or Compose discovery invalidates every downstream safety guarantee.
- Phase 2 precedes execution because backup and restore contracts must exist before mutating stateful services.
- Phase 3 delivers the product's core upgrade and rollback loop only after deterministic planning and recovery gates exist.
- Phase 4 completes the v1 trust surface with inspectable history and public documentation for adoption.
