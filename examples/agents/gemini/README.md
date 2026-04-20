# shum + Gemini CLI

Use [`shum`](https://github.com/imurodl/shum) from [Gemini CLI](https://github.com/google-gemini/gemini-cli) to drive safe Compose upgrades on remote hosts.

## Install

1. `shum` on PATH: `go install github.com/imurodl/shum/cmd/shum@latest`
2. `jq` on PATH (recommended for inspecting `shum agent-help`)
3. At least one host registered: `shum host register prod`

## Configure Gemini CLI to use the shum contract

Gemini CLI loads `GEMINI.md` files hierarchically — from `~/.gemini/GEMINI.md` (user-scope), from your workspace, and from parent directories of the file currently being touched. The file in this directory scopes Gemini's behavior when invoked from here.

For project scope, copy it to your repo root:

```bash
# from your own repo
cp path/to/shum/examples/agents/gemini/GEMINI.md ./GEMINI.md
```

For user-scope (every Gemini session, every project):

```bash
mkdir -p ~/.gemini
cp path/to/shum/examples/agents/gemini/GEMINI.md ~/.gemini/GEMINI.md
```

If you already have a `GEMINI.md`, merge the **Working rules**, **Hard rules**, and **Canonical safe-upgrade flow** sections in. Gemini supports `@file.md` imports, so you can also reference this file from your existing one:

```markdown
# in your existing GEMINI.md
@path/to/shum/examples/agents/gemini/GEMINI.md
```

## Try it

From a directory that contains the `GEMINI.md` from this example (or that imports it):

```bash
gemini
```

Then paste one of the prompts from [`PROMPT.md`](./PROMPT.md), e.g.:

> Upgrade the web service on the prod host using shum. Dry-run first, summarize the plan to me, then ask before applying.

Gemini will pick up the rules from `GEMINI.md` and follow the canonical safe-upgrade flow.

## Reference

- Repo-root [AGENTS.md](../../../AGENTS.md) — full agent contract.
- `shum agent-help` — same data, machine-readable.
- [Gemini CLI GEMINI.md docs](https://geminicli.com/docs/cli/gemini-md/) — upstream convention.
