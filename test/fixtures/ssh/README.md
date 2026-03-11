## SSH Fixtures

SSH E2E scenarios are opt-in and should use a disposable Linux host with:

- alias in your local `~/.ssh/config`
- trust already established in `~/.ssh/known_hosts`
- key/agent authentication only (no password prompts)
- Docker + Docker Compose plugin available remotely

Set:

```bash
export SHUM_E2E_SSH_ALIAS=your-alias
```

Then run remote-facing tests from the project root:

```bash
go test ./test/e2e
```

If the variable is missing, host tests should skip.
