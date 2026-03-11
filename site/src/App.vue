<script setup>
const metrics = [
  { label: 'Operator-focused scope', value: 'Docker Compose upgrades', tone: 'blue' },
  { label: 'Safety controls', value: 'preflight + policy', tone: 'teal' },
  { label: 'Recovery path', value: 'artifact + restore', tone: 'blue' },
  { label: 'Traceability', value: 'run history + events', tone: 'violet' },
]

const pains = [
  'Remote upgrades happen in fragile, undocumented sequences',
  'Backups are ad-hoc, or they are forgotten under pressure',
  'Teams have no canonical way to review what changed',
  'Incident retrospectives waste hours reconstructing the last run',
]

const outcomes = [
  'Predictable command path: discover, preflight, plan, policy, execute',
  'Explicit safeguards before mutation: backups and migration warning gates',
  'Auditable trails for every run and artifact persisted locally',
  'Single command line surface for ops work and scripts',
]

const commandFlow = [
  {
    title: '1) Register and inspect',
    command: 'shum host register prod\nshum project discover prod',
    note: 'Trust your host once, then operate against stable aliases.',
  },
  {
    title: '2) Evaluate and lock intent',
    command: 'shum project preflight prod web\nshum project plan prod web --json',
    note: 'No state changes happen before these checks and plan output.',
  },
  {
    title: '3) Enforce policy',
    command:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\""\nshum project backup take prod web',
    note: 'Policy keeps operator intent explicit and replayable.',
  },
  {
    title: '4) Execute with confidence',
    command: 'shum project upgrade prod web --dry-run --json\nshum project upgrade prod web --json',
    note: 'Dry-run first, then run with recorded output and run-id.',
  },
  {
    title: '5) Verify and audit',
    command:
      'shum project run list --host prod --project web --json\nshum project run show <run-id> --json\nshum project backup list prod web --json',
    note: 'Review history and failures in one consistent record.',
  },
]

const cliSurface = [
  ['shum host', 'register, list, inspect'],
  ['shum project discover', 'discover canonical compose projects'],
  ['shum project inspect', 'inspect metadata, mounts, rendered config'],
  ['shum project preflight|plan', 'safety gates + explicit upgrade plan'],
  ['shum project policy', 'set/show project policy controls'],
  ['shum project backup', 'take, list, restore artifacts'],
  ['shum project upgrade', 'dry-run then execute with probes'],
  ['shum project run', 'run history and detailed status'],
]

const docs = [
  ['Documentation Site', 'https://imurodl.me/shum/', 'Public landing + architecture + quickstart'],
  ['Testing Guide', 'https://github.com/imurodl/shum/blob/main/docs/testing.md', 'Verification matrix and CLI checks'],
  ['GitHub Repo', 'https://github.com/imurodl/shum', 'Source, issues, history'],
  ['Contributing', 'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md', 'How to propose and review changes'],
  ['Security', 'https://github.com/imurodl/shum/blob/main/SECURITY.md', 'Security model and vulnerability reporting'],
  ['Code of Conduct', 'https://github.com/imurodl/shum/blob/main/CODE_OF_CONDUCT.md', 'Community and conduct expectations'],
  ['Changelog', 'https://github.com/imurodl/shum/blob/main/CHANGELOG.md', 'Release notes and project evolution'],
]

</script>

<template>
  <div class="page">
    <div class="halo" aria-hidden="true"></div>
    <div class="glow glow-a"></div>
    <div class="glow glow-b"></div>
    <main class="sheet">
      <header class="hero panel">
        <p class="eyebrow">Self-Hosted Operations Platform</p>
        <h1>shum</h1>
        <p class="subtitle">Safe, recoverable Docker Compose upgrades for real Linux fleets.</p>
        <p class="lead">You get deterministic operational safety without giving up control. Register hosts, validate state, execute with policy gates, and keep every run auditable.</p>
        <div class="actions">
          <a class="btn btn-primary" href="https://github.com/imurodl/shum" target="_blank" rel="noopener noreferrer">Open Source Repo</a>
          <a class="btn btn-secondary" href="#flow">Open the flow</a>
          <a class="btn btn-ghost" href="#cli">CLI Surface</a>
          <a class="btn btn-ghost" href="https://github.com/imurodl/shum/blob/main/docs/testing.md">Testing</a>
        </div>
      </header>

      <section class="panel metrics">
        <article v-for="item in metrics" :key="item.label" :class="['metric', `tone-${item.tone}`]">
          <p class="metric-value">{{ item.value }}</p>
          <p class="metric-label">{{ item.label }}</p>
        </article>
      </section>

      <section class="panel split-2">
        <article>
          <h2>Why it exists</h2>
          <ul class="list">
            <li v-for="item in pains" :key="item">{{ item }}</li>
          </ul>
        </article>
        <article>
          <h2>What you get</h2>
          <ol class="check-list">
            <li v-for="item in outcomes" :key="item">
              <span class="dot" aria-hidden="true">✓</span>
              <span>{{ item }}</span>
            </li>
          </ol>
        </article>
      </section>

      <section class="panel" id="flow">
        <h2>Execution flow</h2>
        <p class="section-copy">Every upgrade follows a strict sequence that favors explicit verification over surprise.</p>
        <div class="timeline">
          <article v-for="step in commandFlow" :key="step.title" class="timeline-card">
            <h3>{{ step.title }}</h3>
            <p>{{ step.note }}</p>
            <pre><code>{{ step.command }}</code></pre>
          </article>
        </div>
      </section>

      <section class="panel" id="cli">
        <h2>CLI surface map</h2>
        <div class="cli-grid">
          <article v-for="entry in cliSurface" :key="entry[0]" class="cli-row">
            <code>{{ entry[0] }}</code>
            <p>{{ entry[1] }}</p>
          </article>
        </div>
      </section>

      <section class="panel split-2">
        <article>
          <h2>Install path</h2>
          <pre><code>git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum
shum --help</code></pre>
          <p class="muted">Storage defaults: <code>~/.config/shum</code> and <code>~/.cache/shum</code>.</p>
        </article>
        <article>
          <h2>Quality commitments</h2>
          <ul class="list">
            <li>No implicit mutation. Dry-run and checks before run.</li>
            <li>Policy is the switchboard for backup, restore, and probes.</li>
            <li>All operations go through test-backed services.</li>
            <li>Failure context and event history remain queryable and scriptable.</li>
          </ul>
        </article>
      </section>

      <section class="panel">
        <h2>Architecture</h2>
        <p class="section-copy">A minimal stack intentionally avoids magic and keeps state explicit.</p>
        <div class="arch">
          <div class="arch-layer">CLI Layer (Cobra commands, JSON + human output)</div>
          <div class="arch-arrow">↳</div>
          <div class="arch-layer">Ops Engine (preflight, planning, upgrade, verify)</div>
          <div class="arch-arrow">↳</div>
          <div class="arch-layer">Service Layer (hosts, projects, runs, policies)</div>
          <div class="arch-arrow">↳</div>
          <div class="arch-layer">Storage (SQLite + local artifacts)</div>
        </div>
      </section>

      <footer class="panel footer">
        <h2>Resources</h2>
        <div class="resource-grid">
          <a v-for="resource in docs" :key="resource[0]" :href="resource[1]" target="_blank" rel="noopener noreferrer" class="resource-card">
            <p class="resource-title">{{ resource[0] }}</p>
            <p>{{ resource[2] }}</p>
          </a>
        </div>
        <p class="muted tiny">Built to show operational quality, not abstract AI capability.</p>
      </footer>
    </main>
  </div>
</template>
