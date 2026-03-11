---
phase: 01
slug: remote-host-registration-and-canonical-compose-discovery
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — Wave 0 installs |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~20 seconds initially |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-TBD-01 | TBD | TBD | HOST-01 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 01-TBD-02 | TBD | TBD | HOST-02 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 01-TBD-03 | TBD | TBD | HOST-03 | integration | `go test ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — initialize Go module for the CLI
- [ ] `*_test.go` scaffold files — host registration, discovery, and inspect coverage stubs
- [ ] test fixture helpers for SSH/config parsing and Compose discovery scenarios
- [ ] remote-host integration strategy documented in test helpers or fixtures

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Strict host-key mismatch UX is clear and trustworthy | HOST-01 | Message clarity and operator confidence are hard to judge from assertions alone | Run host registration against a known bad host-key scenario and confirm the CLI clearly explains why registration is blocked |
| Summary-first inspect output feels readable and trustworthy | HOST-03 | Human readability and information density need operator judgment | Run inspect against canonical, ambiguous, and blocked fixtures and confirm summary output makes project identity and risk state obvious |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
