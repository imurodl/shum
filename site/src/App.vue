<script setup>
const principles = [
  {
    name: 'Deterministic Safety',
    body: 'Every mutation starts from preflight checks and an explicit plan. Surprise change is treated as a bug, not a feature.',
  },
  {
    name: 'Trust-First Operations',
    body: 'Remote hosts are registered once, verified through SSH identity and capability probing, and then reused across runs.',
  },
  {
    name: 'Policy-Gated Mutations',
    body: 'Backups, health checks, and migration warnings are policy controls, not tribal knowledge in your head.',
  },
  {
    name: 'Auditability',
    body: 'Each run stores state transitions, outcomes, and artifacts so incidents can be replayed reliably.',
  },
]

const pipeline = [
  {
    title: 'Host Onboarding',
    command: 'shum host register <alias>',
    detail: 'Trust and fingerprint a production host through SSH, then gather capabilities once.',
  },
  {
    title: 'Project Discovery',
    command: 'shum project discover <alias>',
    detail: 'Discover compose projects and normalize metadata, so future steps target canonical project refs.',
  },
  {
    title: 'Readiness + Plan',
    command: 'shum project preflight <alias> <project-ref>\nshum project plan <alias> <project-ref> --json',
    detail: 'Run reproducibility checks and inspect exactly what will change before mutating.',
  },
  {
    title: 'Policy + Execute',
    command: 'shum project policy set <alias> <project-ref> --require-backup=true --health-check "http://127.0.0.1:8080/health"\nshum project upgrade <alias> <project-ref> --dry-run --json',
    detail: 'Attach backup/restore and post-upgrade verification policy, then dry-run before commit.',
  },
  {
    title: 'Recover + Record',
    command: 'shum project upgrade <alias> <project-ref>\nshum project run list --host <alias> --project <project-ref>',
    detail: 'Execute upgrade with rollback-capable artifact tracking and complete history for team review.',
  },
]

const cliRef = [
  {
    command: 'shum host register <alias>',
    description: 'Capture SSH host identity and environment fingerprints.',
  },
  {
    command: 'shum project inspect <alias> <project-ref> --project-directory ...',
    description: 'Render risk surfaces from compose files, mounts, and project metadata.',
  },
  {
    command: 'shum project backup take <alias> <project-ref> [--command "..."]',
    description: 'Create explicit backup artifacts, optionally overriding policy command.',
  },
  {
    command: 'shum project upgrade <alias> <project-ref> --dry-run',
    description: 'Safe preview mode to review pre-change plan, checks, and constraints.',
  },
  {
    command: 'shum project run show <run-id>',
    description: 'Inspect run summaries, failure reasons, artifacts, and event history.',
  },
]

const installCommands = [
  'git clone https://github.com/imurodl/shum.git',
  'cd shum',
  'go test ./...',
  'go install ./cmd/shum',
  'shum --help',
]

const checks = [
  'No implicit remote mutation without preflight and plan execution',
  'Backup policy controls are explicit and versioned by project',
  'Upgrade runs are persisted with failure context and exit summaries',
  'History can be filtered by host/project for incident archaeology',
  'Tooling is test-covered and runnable with `go test ./...`',
]

const architecture = [
  {
    name: 'CLI Surface',
    text: 'A focused Cobra command surface (host, project, run, backup, policy, upgrade) designed for operators.',
  },
  {
    name: 'Ops Service',
    text: 'Preflight, planning, execution, backup capture, restore, and verification are orchestrated with explicit state transitions.',
  },
  {
    name: 'Repository Layer',
    text: 'SQLite-backed records for hosts, projects, runs, and artifacts; no hidden state.',
  },
  {
    name: 'Remote Runner',
    text: 'SSH-based interaction layer with conservative timeouts and structured output parsing.',
  },
]
</script>

<template>
  <div class="page">
    <div class="halo" aria-hidden="true"></div>
    <main class="stage">
      <header class="hero">
        <p class="eyebrow">Self-hosted Operations Infrastructure</p>
        <h1>shum</h1>
        <p class="subtitle">
          A reliability-first CLI for safer Docker Compose upgrades on remote Linux hosts.
        </p>
        <p class="description">
          It is not an AI wrapper. It is a concrete operations engine with trust-first host management,
          policy-gated mutation, and auditable upgrade history.
        </p>
        <div class="actions">
          <a class="btn btn-primary" href="https://github.com/imurodl/shum" target="_blank" rel="noopener noreferrer">GitHub</a>
          <a class="btn btn-ghost" href="#start">Install</a>
          <a class="btn btn-ghost" href="#commands">Command Reference</a>
          <a class="btn btn-ghost" href="https://github.com/imurodl/shum/blob/main/docs/testing.md" target="_blank" rel="noopener noreferrer">Testing Guide</a>
        </div>
      </header>

      <section class="panel">
        <h2>Why this exists</h2>
        <div class="grid">
          <article v-for="item in principles" :key="item.name" class="card">
            <h3>{{ item.name }}</h3>
            <p>{{ item.body }}</p>
          </article>
        </div>
      </section>

      <section class="panel">
        <h2>Pipeline</h2>
        <p class="section-copy">A practical upgrade path you can automate and review.</p>
        <ol class="steps">
          <li v-for="(step, index) in pipeline" :key="step.title" class="step">
            <span class="step-index">0{{ index + 1 }}</span>
            <h3>{{ step.title }}</h3>
            <p>{{ step.detail }}</p>
            <pre><code>{{ step.command }}</code></pre>
          </li>
        </ol>
      </section>

      <section class="panel" id="commands">
        <h2>Core CLI Surface</h2>
        <div class="command-list">
          <div v-for="item in cliRef" :key="item.command" class="command-item">
            <pre><code>{{ item.command }}</code></pre>
            <p>{{ item.description }}</p>
          </div>
        </div>
      </section>

      <section class="split">
        <article id="start" class="panel">
          <h2>Install</h2>
          <p class="section-copy">
            Build from source and run against a repository clone. This keeps your deployment path explicit.
          </p>
          <pre><code>{{ installCommands.join('\n') }}</code></pre>
          <p class="muted">State lives by default in <code>~/.config/shum</code> and artifacts under <code>~/.cache/shum</code>.</p>
        </article>

        <article class="panel">
          <h2>Reliability Guarantees</h2>
          <ul>
            <li v-for="item in checks" :key="item">{{ item }}</li>
          </ul>
        </article>
      </section>

      <section class="panel">
        <h2>Architecture</h2>
        <div class="grid architecture">
          <article v-for="item in architecture" :key="item.name" class="card">
            <h3>{{ item.name }}</h3>
            <p>{{ item.text }}</p>
          </article>
        </div>
      </section>
    </main>
  </div>
</template>
