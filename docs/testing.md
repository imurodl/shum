# Testing Reference

## Automated

Baseline checks:

```bash
go test ./...
go build ./cmd/shum
```

Optional remote suite:

```bash
go test ./test/e2e
```

## Quick CLI checks (requires a registered host)

Replace `your-alias` and `web` with your registered host alias and project name.

```bash
shum host register your-alias
shum project discover your-alias
shum project preflight your-alias web
shum project plan your-alias web
shum project policy show your-alias web
shum project run list
```

## Remote E2E (Opt-In)

Set SSH alias and known-host trust for a Linux target with Docker and Compose:

```bash
export SHUM_E2E_SSH_ALIAS=your-remote-alias
go test ./test/e2e
```

If variables are not set, remote suites skip with an explicit message.

## Local compose integration

The local compose discovery/inspect integration tests (when present) require local Docker.
They run only when Docker is reachable; otherwise they exit with a skip.
