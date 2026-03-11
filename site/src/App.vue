<script setup>
const capabilityPills = [
  'Trusted SSH host aliases',
  'Docker Compose project discovery',
  'Dry-run before mutation',
  'Backup and restore workflows',
  'Run history and artifacts',
]

const summaryCards = [
  {
    title: 'SSH trust',
    copy: 'Register a host once, then target it through a stable alias.',
  },
  {
    title: 'Compose-aware',
    copy: 'Discover projects, inspect config, and plan image changes.',
  },
  {
    title: 'Recovery-ready',
    copy: 'Backup and restore commands are part of project policy.',
  },
  {
    title: 'Auditable',
    copy: 'Every upgrade produces readable status and run records.',
  },
]

const operatorRisks = [
  'Release steps drift between engineers and hosts.',
  'Rollback planning starts after the deployment has already gone wrong.',
  'Compose upgrades are often executed with weak preflight validation.',
  'Incident review slows down because the run is only partly documented.',
]

const ledger = [
  ['host', 'trusted SSH alias loaded'],
  ['project', 'compose stack discovered'],
  ['policy', 'backup and health rules active'],
  ['plan', 'target image delta computed'],
  ['run', 'artifact and status history persisted'],
]

const flow = [
  {
    stage: '01',
    title: 'Register trust',
    copy: 'Create a durable host entry from a known SSH alias.',
    command: 'shum host register prod\nshum project discover prod',
  },
  {
    stage: '02',
    title: 'Inspect stack',
    copy: 'Read project state and run preflight checks before touching the host.',
    command: 'shum project inspect prod web --project-directory /srv/web --json\nshum project preflight prod web --json',
  },
  {
    stage: '03',
    title: 'Load policy',
    copy: 'Require backups, probe checks, and explicit migration warnings.',
    command:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\""',
  },
  {
    stage: '04',
    title: 'Preview change',
    copy: 'Generate a plan and run a dry-run before the real upgrade.',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
  },
  {
    stage: '05',
    title: 'Execute and review',
    copy: 'Upgrade deliberately, then inspect run history and artifacts.',
    command: 'shum project upgrade prod web --json\nshum project run list --host prod --project web --json',
  },
]

const featureCards = [
  {
    label: 'policy',
    title: 'Project rules are executable',
    copy: 'Backup requirements, restore commands, probes, and warnings are stored with the project rather than improvised during the rollout.',
    sample: '--require-backup=true\n--health-check http://127.0.0.1:8080/health',
  },
  {
    label: 'artifacts',
    title: 'Recovery paths stay close to execution',
    copy: 'Artifacts and restore actions are captured by the same CLI that performs the upgrade.',
    sample: '~/.cache/shum/artifacts/\nbackup-2026-03-11T14:08:22Z.tar',
  },
  {
    label: 'history',
    title: 'Runs are readable after the fact',
    copy: 'Run summaries and status changes remain available for review, debugging, and post-incident cleanup.',
    sample: 'status: completed\nchanged_services: 3\nhealth_check: passed',
  },
]

const commandAtlas = [
  ['shum host register', 'Create a trusted host entry from an SSH alias'],
  ['shum project discover', 'Scan the remote host for Compose projects'],
  ['shum project inspect', 'Inspect mounts, labels, images, and project metadata'],
  ['shum project preflight', 'Validate the environment before mutation'],
  ['shum project plan', 'Compute image and manifest changes'],
  ['shum project policy', 'Persist backup, restore, and probe rules'],
  ['shum project backup', 'Take, list, and restore artifacts'],
  ['shum project upgrade', 'Dry-run first, then execute the rollout'],
  ['shum project run', 'Read status, run summaries, and history'],
]

const architecture = [
  ['CLI', 'Human-readable and JSON outputs for operators and scripts'],
  ['Ops engine', 'Preflight, planning, upgrade execution, and verification'],
  ['Domain state', 'Hosts, projects, policies, runs, and artifact metadata'],
  ['Storage', 'SQLite plus filesystem artifacts for recovery workflows'],
]

const resources = [
  ['GitHub', 'https://github.com/imurodl/shum', 'Source, issues, releases'],
  ['Testing Guide', 'https://github.com/imurodl/shum/blob/main/docs/testing.md', 'Verification matrix and remote checks'],
  ['Contributing', 'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md', 'Contribution standards'],
  ['Security', 'https://github.com/imurodl/shum/blob/main/SECURITY.md', 'Security disclosure policy'],
  ['Changelog', 'https://github.com/imurodl/shum/blob/main/CHANGELOG.md', 'Project evolution'],
]

const bootstrap = `git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum
shum --help`

const previewBlock = `shum project policy set prod web --require-backup=true
shum project backup take prod web --json
shum project upgrade prod web --dry-run --json
shum project run list --host prod --project web --json`

const artifactBlock = `shum project run show run_01JQ4KQ9DZH8 --json
{
  "status": "completed",
  "changed_services": 3,
  "artifact_count": 1,
  "health_check": "passed"
}`
</script>

<template>
  <div class="app-shell">
    <div class="bg-grid" aria-hidden="true"></div>
    <div class="bg-glow bg-glow-left" aria-hidden="true"></div>
    <div class="bg-glow bg-glow-right" aria-hidden="true"></div>

    <header class="topbar">
      <a class="brand" href="#top">
        <span class="brand-mark">SHUM</span>
        <span class="brand-copy">Self-Host Upgrade Manager</span>
      </a>

      <nav class="topnav" aria-label="Primary">
        <a href="#flow">Flow</a>
        <a href="#atlas">CLI</a>
        <a href="#architecture">Architecture</a>
        <a href="#resources">Resources</a>
      </nav>
    </header>

    <main class="page">
      <section class="hero panel" id="top">
        <div class="hero-copy">
          <p class="eyebrow">docker compose upgrades for self-hosted linux</p>
          <h1>Deterministic upgrades for remote Docker Compose hosts.</h1>
          <p class="lead">
            SHUM is a CLI for operators managing Compose stacks on VPSes and self-hosted Linux fleets.
            It makes upgrades previewable, policy-gated, recoverable, and easy to audit after the run.
          </p>

          <div class="hero-actions">
            <a class="button button-primary" href="https://github.com/imurodl/shum" target="_blank" rel="noopener noreferrer">View source</a>
            <a class="button button-secondary" href="#flow">See upgrade flow</a>
            <a class="button button-secondary" href="https://github.com/imurodl/shum/blob/main/docs/testing.md" target="_blank" rel="noopener noreferrer">Read testing guide</a>
          </div>

          <div class="capability-pills">
            <span v-for="item in capabilityPills" :key="item">{{ item }}</span>
          </div>
        </div>

        <aside class="hero-console">
          <div class="console-head">
            <span>release preview</span>
            <span>prod/web</span>
          </div>

          <div class="console-status">
            <article>
              <strong>policy</strong>
              <span>backup required</span>
            </article>
            <article>
              <strong>probe</strong>
              <span>health check enabled</span>
            </article>
            <article>
              <strong>mode</strong>
              <span>dry-run first</span>
            </article>
          </div>

          <pre><code>{{ previewBlock }}</code></pre>
        </aside>
      </section>

      <section class="summary-grid">
        <article v-for="item in summaryCards" :key="item.title" class="summary-card">
          <p class="summary-title">{{ item.title }}</p>
          <p>{{ item.copy }}</p>
        </article>
      </section>

      <section class="panel operator-section">
        <div class="section-head">
          <p class="section-tag">Why operators use SHUM</p>
          <h2>Upgrade workflows stop living in shell history.</h2>
        </div>

        <div class="operator-grid">
          <div class="risk-panel">
            <ul class="risk-list">
              <li v-for="item in operatorRisks" :key="item">{{ item }}</li>
            </ul>
          </div>

          <div class="ledger-panel">
            <p class="ledger-head">run state</p>
            <div class="ledger-rows">
              <div v-for="item in ledger" :key="item[0]" class="ledger-row">
                <span class="ledger-key">{{ item[0] }}</span>
                <span class="ledger-value">{{ item[1] }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="panel flow-section" id="flow">
        <div class="section-head">
          <p class="section-tag">Upgrade flow</p>
          <h2>Five steps from trust to audit.</h2>
        </div>

        <div class="flow-grid">
          <article v-for="item in flow" :key="item.stage" class="flow-card">
            <div class="flow-top">
              <span class="flow-stage">{{ item.stage }}</span>
              <h3>{{ item.title }}</h3>
            </div>
            <p class="flow-copy">{{ item.copy }}</p>
            <pre><code>{{ item.command }}</code></pre>
          </article>
        </div>
      </section>

      <section class="artifacts-section">
        <div class="artifact-console panel">
          <p class="section-tag">Run artifacts</p>
          <h2>Operators can inspect what just happened.</h2>
          <p class="artifact-copy">
            Policy, backup artifacts, and run summaries stay close to the same CLI surface used to execute the upgrade.
          </p>
          <pre><code>{{ artifactBlock }}</code></pre>
        </div>

        <div class="feature-stack">
          <article v-for="item in featureCards" :key="item.label" class="feature-card panel">
            <p class="feature-tag">{{ item.label }}</p>
            <h3>{{ item.title }}</h3>
            <p class="feature-copy">{{ item.copy }}</p>
            <pre><code>{{ item.sample }}</code></pre>
          </article>
        </div>
      </section>

      <section class="panel atlas" id="atlas">
        <div class="atlas-main">
          <div class="section-head">
            <p class="section-tag">CLI surface</p>
            <h2>Command groups built for real upgrade work.</h2>
          </div>

          <div class="atlas-list">
            <article v-for="item in commandAtlas" :key="item[0]" class="atlas-row">
              <code>{{ item[0] }}</code>
              <p>{{ item[1] }}</p>
            </article>
          </div>
        </div>

        <aside class="install-card">
          <p class="section-tag">Install</p>
          <pre><code>{{ bootstrap }}</code></pre>
          <p class="install-note">Config path: <code>~/.config/shum</code></p>
          <p class="install-note">State path: <code>~/.cache/shum</code></p>
          <p class="install-note">Validation: <code>go test ./...</code> and optional <code>go test ./test/e2e</code></p>
        </aside>
      </section>

      <section class="panel architecture" id="architecture">
        <div class="section-head">
          <p class="section-tag">Architecture</p>
          <h2>Compact layers, explicit responsibilities.</h2>
        </div>

        <div class="architecture-grid">
          <article v-for="item in architecture" :key="item[0]" class="architecture-card">
            <strong>{{ item[0] }}</strong>
            <span>{{ item[1] }}</span>
          </article>
        </div>
      </section>

      <footer class="footer panel" id="resources">
        <div class="footer-copy">
          <p class="section-tag">Resources</p>
          <h2>Open-source infrastructure tooling, without mystery state.</h2>
          <p>
            SHUM is for teams that want a concrete CLI for Docker Compose upgrades on remote Linux hosts,
            not a generic deployment platform or another dashboard layer.
          </p>
        </div>

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
