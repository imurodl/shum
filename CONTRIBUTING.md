# Contributing to shum

Thank you for your interest in improving `shum`.

This project is intentionally practical: keep changes focused on operational reliability and explain behavior clearly.

## Code of conduct

By contributing, you agree to follow the behavior in [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md).

## Setup

```bash
git clone https://github.com/imurodl/shum.git
cd shum

go test ./...
go test ./test/e2e # optional, requires SHUM_E2E_SSH_ALIAS
```

## Development flow

- Keep PRs narrow and specific.
- Add or update tests for behavior changes.
- Update docs when behavior or CLI output changes.
- Keep command and package names explicit and descriptive.

## CLI-focused contribution guidelines

This repository is CLI-first. If you change command behavior:

- keep flags orthogonal (`--json` remains available where useful),
- preserve non-breaking defaults,
- include examples in docs for changed behavior,
- and ensure `go test ./...` passes.

## Commit format

- Use short imperative summaries in commit messages.
- Mention test impact in the body if a behavior changed.

## Reporting issues

- Include the exact command and command output.
- Provide OS, Go version, and remote host context when relevant.
- Include repository state snapshots when the failure is stateful.
