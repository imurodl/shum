# Phase 1 Validation

## Purpose

Capture concrete acceptance checks for `HOST-01`, `HOST-02`, and `HOST-03` implementation in one place so Phase 2 can assume durable host/project identity.

## Wave 1: Host Registration (01-01)

| Check | Command / Verification | Expected |
|---|---|---|
| Module bootstrapping | `go test ./...` (or targeted) and `go build ./cmd/shum` | Build and test pass in CI/laptop with Go toolchain. |
| Alias resolution | `shum host register <alias>` against trusted Linux host | Command succeeds only when SSH alias resolves and host key is present in configured known_hosts. |
| Non-interactive auth | host requiring password / prompts | command fails loudly (interactive auth and TOFU flow rejected). |
| Linux check | register against non-Linux host | command fails with non-Linux message. |
| Docker checks | register host missing docker / compose | command fails with clear diagnostic. |
| Storage integrity | `shum host register`, `shum host list`, `shum host inspect` | Host metadata and trust fingerprint visible and persisted. |
| Remote E2E gate | `SHUM_E2E_SSH_ALIAS=... go test ./test/e2e` | Skips with explicit message unless env var is set. |

## Wave 2: Project Discovery (01-02)

| Check | Command / Verification | Expected |
|---|---|---|
| Runtime-first discovery | `shum project discover <alias>` | Returns discovered runtime projects with `runtime_only` when no canonical context is known. |
| No-guess behavior | inspect discovered ambiguous directories without `--path` / `--file` | Does not implicitly guess canonical context. |
| Explicit path discovery | `shum project discover <alias> --path /path/to/dir` | Applies `compose.yaml` / `compose.yml` / `docker-compose.*` only by explicit directory. |
| Multi-file ambiguity | path contains multiple compose files | Returns `ambiguous` with actionable reason. |
| Registry | repeated discovers for same alias | project rows persist/refresh in local store via `projects` table. |
| Discovery tests | `go test ./...` | Unit tests for status + output pass; compose integration test remains conditional skip when Docker unavailable. |

## Wave 3: Canonical Inspect (01-03)

| Check | Command / Verification | Expected |
|---|---|---|
| Canonical summary | `shum project inspect <alias> <project-ref>` | Summary shows host trust, canonical status, and rendered-context surface fields. |
| Blocked/ambiguous explanation | inspect on unresolved/mismatched context | Returns `blocked` or `ambiguous` with reasons instead of hiding failures. |
| Artifact persistence | successful inspect after remote success | `artifacts/<alias>/<project-ref>/config.json` and `runtime.json` exist. |
| Drill-down behavior | `shum project inspect <alias> <project-ref> --show-config` | Renders full config intentionally. |
| Default redaction | `shum project inspect ...` (no flags) | Config output stays summary-first, not raw config dump. |
| CLI formats | `shum project inspect ... --json` and `shum project discover ... --json` | Machine-readable JSON accepted by operators/tools. |
| End-to-end checks | `go test ./...` and `go build ./cmd/shum` | All phase checks pass including unit + integration suites. |
| Remote E2E gate | `SHUM_E2E_SSH_ALIAS=... go test ./test/e2e` | Skips by default; runs when env var is supplied. |

## Exit Criteria

- Host registration and discovery commands behave fail-closed without explicit trust, auth, Linux, and docker checks.
- Discovery and inspect outputs always surface uncertainty (`runtime_only`, `ambiguous`, `blocked`) instead of hidden assumptions.
- Phase 1 artifacts are persisted and traceable (`state.db`, artifacts under XDG cache path).
- Documentation (`README`, `docs/testing.md`) matches current behavior and state layout.
