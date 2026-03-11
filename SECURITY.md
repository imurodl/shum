# Security

`shum` manages operational workflows for remote hosts. Security is part of the design surface.

## Reporting

If you discover a security vulnerability:

1. Do not post it publicly.
2. Open a private security disclosure using GitHub security reporting.
3. Include reproduction steps, command examples, and impact level.

## Hardening defaults

- SSH access uses host aliasing and stored key fingerprints.
- All remote actions are executed over SSH command channels, not stored credentials.
- Backups, restore commands, and health probes are explicit per-project policy.

## Safe operation

- Use a dedicated operator account on targets.
- Review host trust and known-host checks before running destructive commands.
- Store state directories under user-owned paths and protect them by file system permissions.
