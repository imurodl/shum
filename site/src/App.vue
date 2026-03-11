<script setup>
const metrics = [
  { value: 'Remote compose upgrades', label: 'Execution domain', tone: 'blue' },
  { value: 'Preflight + plan', label: 'Safety checkpoints', tone: 'teal' },
  { value: 'artifact + restore', label: 'Recovery model', tone: 'violet' },
  { value: 'run history + events', label: 'Observability', tone: 'blue' },
]

const pains = [
  'Operators run upgrades with partial visibility into what changed.',
  'Backups are often informal and hard to reproduce.',
  'Policy can be tribal knowledge instead of enforced policy.',
  'Incident retrospectives become archaeology instead of engineering.',
]

const outcomes = [
  'Deterministic, documented flow from discovery to execution.',
  'Explicit policy checks for backup, restore, and health verification.',
  'Every operation goes through test-backed services with status transitions.',
  'Complete run records for fast incident reconstruction.',
]

const workflow = [
  {
    phase: '01',
    title: 'Discover host and project topology',
    cmd: 'shum host register prod\nshum project discover prod',
    note: 'Host trust is captured once and reused through aliases.',
  },
  {
    phase: '02',
    title: 'Inspect and evaluate readiness',
    cmd: 'shum project inspect prod web --project-directory /srv/web --json\nshum project preflight prod web --json',
    note: 'Preflight data is a hard stop for unsafe execution.',
  },
  {
    phase: '03',
    title: 'Set policy, then stage upgrades',
    cmd:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\""\nshum project backup take prod web --json',
    note: 'The policy file becomes an explicit safety contract.',
  },
  {
    phase: '04',
    title: 'Dry-run, then execute',
    cmd: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json\nshum project upgrade prod web --json',
    note: 'No surprises: validation first, execution second.',
  },
  {
    phase: '05',
    title: 'Audit and recover',
    cmd:
      'shum project run list --host prod --project web --json\nshum project run show <run-id> --json\nshum project backup list prod web --json',
    note: 'Every run has artifact traces, summaries, and failure context.',
  },
]

const cliSurface = [
  ['shum host register', 'Create a trusted host entry from SSH alias'],
  ['shum project discover', 'Scan for canonical compose projects'],
  ['shum project inspect', 'Inspect mounts, metadata, and risk surfaces'],
  ['shum project preflight', 'Verify environment preconditions'],
  ['shum project plan', 'Resolve target vs current image changes'],
  ['shum project policy', 'Configure backup/probe/migration gating'],
  ['shum project backup', 'Take/list/restore artifacts'],
  ['shum project upgrade', 'Execute upgrades with --dry-run support'],
  ['shum project run', 'Inspect all upgrade history'],
]

const resources = [
  ['Website', 'https://imurodl.me/shum/', 'Live landing page with architecture and workflows'],
  ['Testing Guide', 'https://github.com/imurodl/shum/blob/main/docs/testing.md', 'Verification matrix and optional remote checks'],
  ['GitHub', 'https://github.com/imurodl/shum', 'Repository, issues, releases'],
  ['Contributing', 'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md', 'Review and contribution standards'],
  ['Security', 'https://github.com/imurodl/shum/blob/main/SECURITY.md', 'Security model and disclosure policy'],
  ['Changelog', 'https://github.com/imurodl/shum/blob/main/CHANGELOG.md', 'Project updates and history'],
]
</script>

<template>
  <div class="page">
    <div class="ambient ambient-a" aria-hidden="true"></div>
    <div class="ambient ambient-b" aria-hidden="true"></div>

    <main class="canvas">
      <header class="panel hero">
        <p class="eyebrow">open-source + production operations</p>
        <h1>SHUM</h1>
        <p class="subtitle">Safe, Host-Aware Upgrade Management for self-hosted Linux fleets.</p>
        <p class="lead">Safe, recoverable Docker Compose upgrades with trust-first SSH state and auditable operations.</p>
        <div class="actions">
          <a class="btn btn-primary" href="https://github.com/imurodl/shum" target="_blank" rel="noopener noreferrer">View Source</a>
          <a class="btn btn-outline" href="#workflow">View Upgrade Flow</a>
          <a class="btn btn-outline" href="#cli">CLI Reference</a>
          <a class="btn btn-outline" href="https://github.com/imurodl/shum/blob/main/docs/testing.md" target="_blank" rel="noopener noreferrer">Testing Guide</a>
        </div>
      </header>

      <section class="panel badge-row">
        <article v-for="item in metrics" :key="item.label" :class="['badge', `tone-${item.tone}`]">
          <p class="badge-value">{{ item.value }}</p>
          <p class="badge-label">{{ item.label }}</p>
        </article>
      </section>

      <section class="panel split-2">
        <article>
          <h2>Why teams use this</h2>
          <ul class="list">
            <li v-for="item in pains" :key="item">{{ item }}</li>
          </ul>
        </article>
        <article>
          <h2>What you gain</h2>
          <ul class="outcomes">
            <li v-for="item in outcomes" :key="item">
              <span class="check" aria-hidden="true">✓</span>
              <span>{{ item }}</span>
            </li>
          </ul>
        </article>
      </section>

      <section class="panel" id="workflow">
        <h2>Upgrade flow</h2>
        <p class="section-copy">A strict flow for safe change, with explicit state at each step.</p>
        <div class="roadmap">
          <article v-for="step in workflow" :key="step.phase" class="road-step">
            <p class="phase">{{ step.phase }}</p>
            <div class="road-body">
              <h3>{{ step.title }}</h3>
              <p>{{ step.note }}</p>
              <pre><code>{{ step.cmd }}</code></pre>
            </div>
          </article>
        </div>
      </section>

      <section class="panel split-2" id="cli">
        <article>
          <h2>CLI reference</h2>
          <div class="cli-list">
            <article v-for="item in cliSurface" :key="item[0]" class="cli-row">
              <code>{{ item[0] }}</code>
              <p>{{ item[1] }}</p>
            </article>
          </div>
        </article>
        <article>
          <h2>Install path</h2>
          <pre><code>git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum
shum --help</code></pre>
          <p class="muted">Storage defaults:
            <code>~/.config/shum</code> and <code>~/.cache/shum</code>.
          </p>
          <p class="muted">Tests:
            <code>go test ./...</code> and optional
            <code>go test ./test/e2e</code>.
          </p>
        </article>
      </section>

      <section class="panel">
        <h2>Architecture</h2>
        <p class="section-copy">Minimal layers, explicit ownership, no hidden control plane.</p>
        <div class="architecture">
          <div class="arch-box">CLI layer (command parsing, JSON/human output)</div>
          <div class="arch-arrow">▼</div>
          <div class="arch-box">Ops engine (preflight, planning, upgrade, verification)</div>
          <div class="arch-arrow">▼</div>
          <div class="arch-box">Domain services (hosts, projects, runs, policies)</div>
          <div class="arch-arrow">▼</div>
          <div class="arch-box">Storage (SQLite + artifact store)</div>
        </div>
      </section>

      <footer class="panel footer">
        <h2>Project resources</h2>
        <div class="resource-grid">
          <a
            v-for="item in resources"
            :key="item[0]"
            :href="item[1]"
            target="_blank"
            rel="noopener noreferrer"
            class="resource-card"
          >
            <p class="resource-title">{{ item[0] }}</p>
            <p>{{ item[2] }}</p>
          </a>
        </div>
      </footer>
    </main>
  </div>
</template>
