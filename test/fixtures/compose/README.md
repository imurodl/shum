## Compose Fixtures

This folder holds local compose scenarios used by integration tests.

Suggested minimum layouts:

- `single-compose/compose.yaml`
- `profiles/compose.yaml`
- `env-required/compose.yaml`
- `multi-file/base.yaml`, `multi-file/web.yaml`

Discovery tests should use explicit `project-directory` inputs only.
Ambiguous or hidden context should be intentionally represented by fixture layout.

If Docker is unavailable in your environment, these tests should skip cleanly.
