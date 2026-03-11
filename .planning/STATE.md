# Planning State: Self-Host Upgrade Manager

**Updated:** 2026-03-11
**Roadmap status:** Created
**Mode:** yolo
**Granularity:** coarse
**Current milestone:** v1
**Current phase:** Phase 1 - Remote Host Registration and Canonical Compose Discovery
**Next action:** Create the detailed plan for Phase 1

## Coverage

- Total v1 requirements: 18
- Mapped requirements: 18
- Unmapped requirements: 0
- Duplicate mappings: 0
- Coverage status: 100% complete

## Phase Queue

| Phase | Status | Requirement Count | Goal |
|-------|--------|-------------------|------|
| Phase 1 | Next | 3 | Establish trusted SSH host registration and canonical Compose discovery |
| Phase 2 | Queued | 7 | Build deterministic planning, safety gates, and recoverability prerequisites |
| Phase 3 | Queued | 4 | Deliver health-gated upgrade execution with rollback control |
| Phase 4 | Queued | 4 | Ship operator-facing history and public docs for adoption |

## Notes

- Roadmap derived directly from v1 requirements and research findings on operational correctness risks.
- Phase count kept low to match coarse granularity while preserving explicit safety boundaries.
