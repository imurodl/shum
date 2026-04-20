# shum + OpenAI Codex CLI

Use [`shum`](https://github.com/imurodl/shum) from [Codex CLI](https://developers.openai.com/codex/cli) to drive safe Compose upgrades on remote hosts.

## Install

1. `shum` on PATH: `go install github.com/imurodl/shum/cmd/shum@latest`
2. `jq` on PATH (recommended for inspecting `shum agent-help`)
3. At least one host registered: `shum host register prod`

## Configure Codex to use the shum contract

Codex automatically reads `AGENTS.md` from your project root (and walks up parents). The file in this directory scopes Codex's behavior when invoked from here.

To use it project-wide, copy it to your repo root, or merge its contents into your existing `AGENTS.md`:

```bash
# from your own repo
cp path/to/shum/examples/agents/codex/AGENTS.md ./AGENTS.md
# or merge the "Working rules", "Hard rules", and "Canonical safe-upgrade flow"
# sections into your existing AGENTS.md
```

For user-scope (every Codex session, every project), copy to `~/.codex/AGENTS.md`.

## Try it

From a directory that contains the AGENTS.md from this example:

```bash
codex
```

Then paste one of the prompts from [`PROMPT.md`](./PROMPT.md), e.g.:

> Upgrade the web service on the prod host using shum. Dry-run first, summarize the plan to me, then ask before applying.

Codex will pick up the rules from AGENTS.md and follow the canonical safe-upgrade flow.

## Reference

- Repo-root [AGENTS.md](../../../AGENTS.md) — full agent contract.
- `shum agent-help` — same data, machine-readable.
- [Codex AGENTS.md docs](https://developers.openai.com/codex/guides/agents-md) — upstream convention.
