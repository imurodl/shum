<script setup>
const pulseStats = [
  ['99%', 'replayable run summaries'],
  ['5', 'release checkpoints'],
  ['1', 'source of truth for upgrades'],
  ['~', 'zero AI dependency'],
]

const painMap = [
  'Upgrade work is too operational and too risky to be tribal knowledge.',
  'Teams lose an upgrade window when backups, prechecks, and rollback steps are implied but undocumented.',
  'Incident review becomes log archaeology because state transitions are incomplete.',
  'Self-hosted deployments lack a consistent and auditable upgrade protocol.',
]

const strengths = [
  {
    badge: 'Policy',
    title: 'Explicit Safety',
    text: 'Backup, migration check, and health probe gates are required configuration in project policy.',
  },
  {
    badge: 'Execution',
    title: 'Deterministic Flow',
    text: 'Host discovery, preflight, plan, upgrade, verify, and audit happen in a strict sequence.',
  },
  {
    badge: 'Recovery',
    title: 'Artifact First',
    text: 'Artifact capture and restore pathways are part of normal execution, not an afterthought.',
  },
]

const flow = [
  {
    index: '01',
    title: 'Map',
    command: 'shum host register prod\nshum project discover prod',
    detail: 'Discover compose projects and lock trusted host intent before upgrades begin.',
  },
  {
    index: '02',
    title: 'Inspect',
    command: 'shum project inspect prod web --project-directory /srv/web --json\nshum project preflight prod web --json',
    detail: 'Validate the environment and capture risk context before any mutation.',
  },
  {
    index: '03',
    title: 'Govern',
    command:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\""\nshum project backup take prod web --json',
    detail: 'Turn safety expectations into explicit, versioned project policy.',
  },
  {
    index: '04',
    title: 'Stage',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
    detail: 'Produce exact planned change set and validate the upgrade path before execution.',
  },
  {
    index: '05',
    title: 'Execute',
    command: 'shum project upgrade prod web --json\nshum project run list --host prod --project web --json',
    detail: 'Execute with traceable outcomes and auditable run artifacts after the fact.',
  },
]

const cliList = [
  ['host', 'register • list • inspect • remove'],
  ['project discover', 'Scan for canonical compose projects by SSH alias'],
  ['project inspect', 'Expose mounts, labels, env, and image baselines'],
  ['project preflight', 'Readiness verification gates'],
  ['project plan', 'Compute planned upgrade delta and image changes'],
  ['project policy', 'Persist policy for backup/probe/migration safety'],
  ['project backup', 'Create, list, and restore backup artifacts'],
  ['project upgrade', 'Dry-run first, then execute'],
  ['project run', 'Inspect execution history and statuses'],
]

const installSteps = `git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum
shum --help`

const resources = [
  ['GitHub', 'https://github.com/imurodl/shum', 'Source repository and issues'],
  ['Testing Guide', 'https://github.com/imurodl/shum/blob/main/docs/testing.md', 'Verification and checks'],
  ['Contributing', 'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md', 'Contribution standards'],
  ['Security', 'https://github.com/imurodl/shum/blob/main/SECURITY.md', 'Security disclosure policy'],
  ['Changelog', 'https://github.com/imurodl/shum/blob/main/CHANGELOG.md', 'Project progress and history'],
]
</script>

<template>
  <div class="app-shell">
    <div class="bg-glow bg-glow-left" aria-hidden="true"></div>
    <div class="bg-glow bg-glow-right" aria-hidden="true"></div>
    <div class="bg-grid" aria-hidden="true"></div>

    <header class="top-bar">
      <a class="brand" href="#top">
        <span class="brand-logo">SHUM</span>
        <span class="brand-desc">Self-Host Upgrade Manager</span>
      </a>
      <nav class="top-nav" aria-label="Main sections">
        <a href="#flow">Flow</a>
        <a href="#commands">CLI</a>
        <a href="#architecture">Architecture</a>
        <a href="#resources">Resources</a>
      </nav>
    </header>

    <main class="container">
      <section class="panel hero" id="top">
        <div class="hero-left">
          <p class="kicker">Operate with confidence</p>
          <h1>Upgrade self-hosted stacks with explicit gates, history, and recovery.</h1>
          <p class="hero-copy">
            SHUM is a practical production tool for remote Compose upgrades. It enforces
            policy-first execution, keeps audit trails first-class, and treats rollback as part of the flow.
          </p>
          <div class="hero-actions">
            <a class="btn btn-primary" href="https://github.com/imurodl/shum" target="_blank" rel="noopener noreferrer">Source</a>
            <a class="btn btn-outline" href="https://github.com/imurodl/shum/blob/main/docs/testing.md" target="_blank" rel="noopener noreferrer">Testing Guide</a>
            <a class="btn btn-outline" href="#flow">See Flow</a>
          </div>

          <ul class="pulse-stats">
            <li v-for="item in pulseStats" :key="item[0]">
              <strong>{{ item[0] }}</strong>
              <span>{{ item[1] }}</span>
            </li>
          </ul>
        </div>

        <aside class="panel terminal">
          <p class="terminal-head">shum terminal</p>
          <pre><code>{{ installSteps }}</code></pre>
        </aside>
      </section>

      <section class="panel pain-grid">
        <article class="half">
          <h2>What teams struggle with</h2>
          <ul>
            <li v-for="item in painMap" :key="item">{{ item }}</li>
          </ul>
        </article>
        <article class="half">
          <h2>What SHUM fixes</h2>
          <div class="strength-grid">
            <article v-for="item in strengths" :key="item.badge" class="strength">
              <p class="strength-tag">{{ item.badge }}</p>
              <h3>{{ item.title }}</h3>
              <p>{{ item.text }}</p>
            </article>
          </div>
        </article>
      </section>

      <section class="panel" id="flow">
        <h2>Flow in practice</h2>
        <p class="section-copy">A strict sequence designed to reduce uncertainty before execution.</p>
        <div class="timeline">
          <article v-for="step in flow" :key="step.index" class="step">
            <p class="step-index">{{ step.index }}</p>
            <div class="step-body">
              <h3>{{ step.title }}</h3>
              <p>{{ step.detail }}</p>
              <pre><code>{{ step.command }}</code></pre>
            </div>
          </article>
        </div>
      </section>

      <section class="panel split" id="commands">
        <article>
          <h2>Command surface</h2>
          <div class="cli-list">
            <article v-for="item in cliList" :key="item[0]" class="cli-row">
              <code>shum {{ item[0] }}</code>
              <p>{{ item[1] }}</p>
            </article>
          </div>
        </article>

        <article>
          <h2>Install</h2>
          <pre><code>{{ installSteps }}</code></pre>
          <p class="muted">Config: <code>~/.config/shum</code></p>
          <p class="muted">State and artifacts: <code>~/.cache/shum</code></p>
          <p class="muted">Tests: <code>go test ./...</code>, optional <code>go test ./test/e2e</code></p>
        </article>
      </section>

      <section class="panel" id="architecture">
        <h2>Architecture</h2>
        <div class="arch-grid">
          <article>CLI Layer<br /><span>JSON + human command output</span></article>
          <article>Ops Engine<br /><span>preflight, plan, upgrade, verify</span></article>
          <article>Domain Layer<br /><span>hosts, projects, policies, runs</span></article>
          <article>Storage<br /><span>SQLite + artifact store</span></article>
        </div>
      </section>

      <footer class="panel resource-panel" id="resources">
        <h2>Resources</h2>
        <div class="resource-links">
          <a v-for="item in resources" :key="item[0]" :href="item[1]" target="_blank" rel="noopener noreferrer">
            <p class="resource-title">{{ item[0] }}</p>
            <span>{{ item[2] }}</span>
          </a>
        </div>
      </footer>
    </main>
  </div>
</template>
