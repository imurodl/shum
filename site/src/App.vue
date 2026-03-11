<script setup>
const principles = [
  {
    title: 'Risk Isolation',
    text: 'Compose upgrades run behind explicit policy gates, so every change has a recovery story before rollout.',
    metric: 'Policy-first',
  },
  {
    title: 'Operational Memory',
    text: 'Runs are preserved as ordered events with status transitions, so incidents are diagnosable under pressure.',
    metric: 'Traceability',
  },
  {
    title: 'Recovery by Design',
    text: 'Backups and restore commands are treated as first-class actions, not tribal knowledge.',
    metric: 'Safe rollback',
  },
]

const pains = [
  'Teams lose confidence during upgrade nights.',
  'Policies differ by engineer and drift over time.',
  'Backups are created late or not consistently at all.',
  'Debugging a failed rollout takes longer than the rollout itself.',
]

const workflow = [
  {
    phase: '01',
    title: 'Discover',
    command: 'shum host register prod\nshum project discover prod',
    details: 'Register trusted SSH targets once, then discover canonical compose projects.',
  },
  {
    phase: '02',
    title: 'Inspect',
    command: 'shum project inspect prod web --project-directory /srv/web --json\nshum project preflight prod web --json',
    details: 'Inspect stacks and run preflight checks before any mutating action.',
  },
  {
    phase: '03',
    title: 'Govern',
    command:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\""\nshum project backup take prod web --json',
    details: 'Capture restore and probe policy in versioned project state.',
  },
  {
    phase: '04',
    title: 'Dry-Run',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
    details: 'Show exactly what will change, then execute intentionally.',
  },
  {
    phase: '05',
    title: 'Audit',
    command:
      'shum project upgrade prod web --json\nshum project run list --host prod --project web --json\nshum project backup list prod web --json',
    details: 'Keep every action reproducible with run metadata and artifact evidence.',
  },
]

const commandCenter = `shum host register prod
shum project discover prod
shum project preflight prod web
shum project plan prod web --json
shum project backup take prod web --json
shum project upgrade prod web --dry-run --json
shum project run list --host prod --project web --json`

const cliSurface = [
  ['shum host register', 'Create a trusted host entry from SSH alias'],
  ['shum project discover', 'Locate canonical compose projects on a target'],
  ['shum project inspect', 'Inspect stack risk surfaces and metadata'],
  ['shum project preflight', 'Verify environment readiness'],
  ['shum project plan', 'Resolve image and manifest changes'],
  ['shum project policy', 'Enforce backup/probe/migration gates'],
  ['shum project backup', 'Take/list/restore artifacts'],
  ['shum project upgrade', 'Dry-run first, then execute'],
  ['shum project run', 'Read status, summaries, and histories'],
]

const resources = [
  ['Website', 'https://imurodl.me/shum/', 'Overview and architecture'],
  ['GitHub', 'https://github.com/imurodl/shum', 'Repository, issues, releases'],
  ['Testing Guide', 'https://github.com/imurodl/shum/blob/main/docs/testing.md', 'Verification matrix'],
  ['Contributing', 'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md', 'Contribution workflow'],
  ['Security', 'https://github.com/imurodl/shum/blob/main/SECURITY.md', 'Vulnerability policy'],
  ['Changelog', 'https://github.com/imurodl/shum/blob/main/CHANGELOG.md', 'Project evolution'],
]
</script>

<template>
  <div class="site-shell">
    <div class="grain" aria-hidden="true"></div>
    <div class="glow glow-left" aria-hidden="true"></div>
    <div class="glow glow-right" aria-hidden="true"></div>

    <main class="layout">
      <header class="topbar panel">
        <a class="brand" href="#top">
          <span class="brand-mark">SHUM</span>
          <span class="brand-copy">Self-Host Upgrade Manager</span>
        </a>
        <nav class="top-links" aria-label="section links">
          <a href="#flow">Flow</a>
          <a href="#commands">CLI</a>
          <a href="#architecture">Architecture</a>
          <a href="#resources">Resources</a>
        </nav>
      </header>

      <section class="panel hero" id="top">
        <div class="hero-copy">
          <p class="eyebrow">for production teams running self-hosted fleets</p>
          <h1>Safe, Host-Aware upgrades without guesswork.</h1>
          <p class="lead">
            SHUM provides a deterministic path from discovery to execution: policy gates, backups,
            auditability, and recoverability built into every step.
          </p>
          <div class="actions">
            <a class="btn btn-primary" href="https://github.com/imurodl/shum" target="_blank" rel="noopener noreferrer">View Source</a>
            <a class="btn btn-ghost" href="#flow">View Upgrade Flow</a>
            <a class="btn btn-ghost" href="https://github.com/imurodl/shum/blob/main/docs/testing.md" target="_blank" rel="noopener noreferrer">Read Testing Guide</a>
          </div>
        </div>
        <aside class="panel terminal">
          <p class="terminal-head">
            <span>shum shell</span>
            <span>operational mode: strict</span>
          </p>
          <pre><code>{{ commandCenter }}</code></pre>
        </aside>
      </section>

      <section class="panel grid-principles">
        <article class="hero-metric">
          <p class="metric-value">5</p>
          <p class="metric-label">Core phases before rollout</p>
        </article>
        <article class="hero-metric">
          <p class="metric-value">100%</p>
          <p class="metric-label">Audit trail for mutating runs</p>
        </article>
        <article class="hero-metric">
          <p class="metric-value">Policy</p>
          <p class="metric-label">Backups, probes, and migration gates</p>
        </article>
      </section>

      <section class="panel two-up">
        <article>
          <h2>Why teams adopt SHUM</h2>
          <ul class="list">
            <li v-for="item in pains" :key="item">{{ item }}</li>
          </ul>
        </article>
        <article class="principles">
          <h2>How it solves it</h2>
          <div v-for="item in principles" :key="item.title" class="principle-card">
            <p class="pill">{{ item.metric }}</p>
            <h3>{{ item.title }}</h3>
            <p>{{ item.text }}</p>
          </div>
        </article>
      </section>

      <section class="panel" id="flow">
        <h2>Upgrade flow</h2>
        <p class="section-copy">A controlled sequence that favors explicit state over assumptions.</p>
        <div class="roadmap">
          <article v-for="step in workflow" :key="step.phase" class="road-step">
            <p class="phase">{{ step.phase }}</p>
            <div class="road-body">
              <h3>{{ step.title }}</h3>
              <p>{{ step.details }}</p>
              <pre><code>{{ step.command }}</code></pre>
            </div>
          </article>
        </div>
      </section>

      <section class="panel two-up" id="commands">
        <article>
          <h2>CLI Surface</h2>
          <div class="cli-grid">
            <article v-for="item in cliSurface" :key="item[0]" class="cli-row">
              <code>{{ item[0] }}</code>
              <p>{{ item[1] }}</p>
            </article>
          </div>
        </article>
        <article>
          <h2>Install in seconds</h2>
          <pre><code>git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum
shum --help</code></pre>
          <p class="muted">
            Storage:
            <code>~/.config/shum</code> and <code>~/.cache/shum</code>
          </p>
          <p class="muted">
            Validation:
            <code>go test ./...</code> and optional <code>go test ./test/e2e</code>
          </p>
        </article>
      </section>

      <section class="panel" id="architecture">
        <h2>Architecture</h2>
        <p class="section-copy">Minimal layers, explicit ownership, no hidden control plane.</p>
        <div class="architecture">
          <div class="arch-box">CLI entrypoint: structured JSON + human-readable outputs</div>
          <div class="arch-box">Ops engine: preflight, plan, upgrade, verification</div>
          <div class="arch-box">Domain layer: host, project, policy, run state</div>
          <div class="arch-box">Storage layer: SQLite + artifact store</div>
        </div>
      </section>

      <footer class="panel footer" id="resources">
        <h2>Resources</h2>
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
