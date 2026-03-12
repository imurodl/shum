<script setup>
const navLinks = [
  { label: 'Why', href: '#why' },
  { label: 'Flow', href: '#flow' },
  { label: 'Proof', href: '#proof' },
  { label: 'Start', href: '#start' },
]

const tickerItems = [
  'trusted SSH aliases',
  'compose-aware inspection',
  'dry-run before mutation',
  'backup policy gates',
  'artifact and run history',
]

const failureModes = [
  {
    index: 'A1',
    title: 'Deployment knowledge drifts between operators.',
    copy: 'The real plan often lives in shell history, notes, and one engineer’s memory.',
  },
  {
    index: 'A2',
    title: 'Rollback preparation starts after the damage.',
    copy: 'Backups and restore commands stay optional until the rollout is already unstable.',
  },
  {
    index: 'A3',
    title: 'Preflight checks are uneven and fragile.',
    copy: 'Image deltas, health checks, and warnings get validated differently every time.',
  },
  {
    index: 'A4',
    title: 'After the run, evidence is incomplete.',
    copy: 'Teams struggle to reconstruct what changed, what was protected, and what actually passed.',
  },
]

const flowSteps = [
  {
    step: '01',
    title: 'Register the host you already trust',
    copy: 'Start from a known SSH alias and keep that identity stable during live work.',
    command: 'shum host register prod\nshum project discover prod',
  },
  {
    step: '02',
    title: 'Inspect and preflight the Compose project',
    copy: 'Read effective project state and validate readiness before a real change.',
    command:
      'shum project inspect prod web --project-directory /srv/web --json\nshum project preflight prod web --json',
  },
  {
    step: '03',
    title: 'Attach policy to the project itself',
    copy: 'Store backup requirements and health probes with the project instead of operator memory.',
    command:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \\"$SHUM_BACKUP_ARTIFACT\\"" \\\n  --health-check "http://127.0.0.1:8080/health"',
  },
  {
    step: '04',
    title: 'Plan and dry-run before mutation',
    copy: 'Compute the intended change and make preview output part of the normal path.',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
  },
  {
    step: '05',
    title: 'Execute and inspect the resulting run',
    copy: 'The same CLI that performs the rollout also preserves summaries and artifacts afterward.',
    command: 'shum project upgrade prod web --json\nshum project run list --host prod --project web --json',
  },
]

const proofCards = [
  {
    label: 'Preview',
    title: 'You do not learn the blast radius at execution time.',
    copy: 'Inspection, planning, and dry-runs make the path readable before the host changes.',
  },
  {
    label: 'Policy',
    title: 'Safety controls stop being optional.',
    copy: 'Backup requirements and health checks become project rules instead of team habits.',
  },
  {
    label: 'Recovery',
    title: 'Artifacts stay close to the upgrade workflow.',
    copy: 'Recovery paths stay near the run instead of being reconstructed after something breaks.',
  },
  {
    label: 'Audit',
    title: 'A rollout leaves evidence behind.',
    copy: 'Operators can inspect status, changed services, and health outcomes after the run.',
  },
]

const evidencePanels = [
  {
    label: 'Field test',
    title: 'Fast path to evaluation',
    copy: 'Judge SHUM by running the workflow against a real host and seeing whether the path gets clearer.',
    code: `shum host register prod
shum project discover prod
shum project inspect prod web --project-directory /srv/web --json
shum project preflight prod web --json
shum project upgrade prod web --dry-run --json`,
  },
  {
    label: 'Run evidence',
    title: 'Post-run facts',
    copy: 'Review the actual result instead of hunting through old terminal buffers.',
    code: `shum project run show run_01JQ4KQ9DZH8 --json
{
  "status": "completed",
  "changed_services": 3,
  "artifact_count": 1,
  "health_check": "passed"
}`,
  },
]

const resourceCards = [
  {
    title: 'Quick start',
    href: 'https://github.com/imurodl/shum#quick-start',
    copy: 'Install SHUM and walk through the standard operator flow.',
  },
  {
    title: 'Repository',
    href: 'https://github.com/imurodl/shum',
    copy: 'Source, issues, releases, and implementation detail.',
  },
  {
    title: 'Testing guide',
    href: 'https://github.com/imurodl/shum/blob/main/docs/testing.md',
    copy: 'Coverage strategy, integration checks, and optional remote tests.',
  },
  {
    title: 'Contributing',
    href: 'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md',
    copy: 'Contribution standards and workflow expectations.',
  },
]

const installBlock = `go install github.com/imurodl/shum/cmd/shum@latest
shum --help`

const sourceInstallBlock = `git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum`

const statusBlock = `release_brief:
  host: prod
  project: web
  mode: dry-run
  backup_required: true
  planned_service_changes: 3
  health_probe: ready`
</script>

<template>
  <div class="app-shell">
    <div class="paper-noise" aria-hidden="true"></div>
    <div class="ink-grid" aria-hidden="true"></div>
    <div class="shape shape-one" aria-hidden="true"></div>
    <div class="shape shape-two" aria-hidden="true"></div>

    <header class="masthead">
      <a class="brand" href="#top">
        <span class="brand-stamp">SHUM</span>
        <span class="brand-copy">Self-Host Upgrade Manager</span>
      </a>

      <nav class="nav" aria-label="Primary">
        <a v-for="item in navLinks" :key="item.href" :href="item.href">{{ item.label }}</a>
      </nav>

      <a
        class="masthead-link"
        href="https://github.com/imurodl/shum#quick-start"
        target="_blank"
        rel="noopener noreferrer"
      >
        Open quick start
      </a>
    </header>

    <main class="page">
      <section class="hero" id="top">
        <div class="hero-copy">
          <p class="section-kicker">Field manual for Compose operators</p>
          <p class="hero-index">01</p>
          <h1>Remote Docker Compose upgrades need procedure, not courage.</h1>
          <p class="hero-lead">
            SHUM is a CLI for operators who want a readable path from inspection to review. It keeps upgrades
            explicit, policy-gated, dry-runnable, recoverable, and auditable on normal self-hosted Linux hosts.
          </p>

          <div class="hero-actions">
            <a
              class="button button-primary"
              href="https://github.com/imurodl/shum#quick-start"
              target="_blank"
              rel="noopener noreferrer"
            >
              Read docs / quick start
            </a>
            <a
              class="button button-secondary"
              href="https://github.com/imurodl/shum"
              target="_blank"
              rel="noopener noreferrer"
            >
              View repository
            </a>
          </div>
        </div>

        <aside class="hero-board">
          <div class="board-header">
            <div>
              <p class="section-kicker">Release brief</p>
              <h2>Prod / web</h2>
            </div>
            <span class="board-status">safe path loaded</span>
          </div>

          <pre><code>{{ statusBlock }}</code></pre>

          <div class="board-notes">
            <article>
              <span>Known target</span>
              <p>Trusted SSH alias remains the operator entrypoint.</p>
            </article>
            <article>
              <span>Policy active</span>
              <p>Backup and health constraints are attached before execution.</p>
            </article>
            <article>
              <span>Evidence retained</span>
              <p>Run summaries and artifacts remain reviewable after the rollout.</p>
            </article>
          </div>
        </aside>
      </section>

      <section class="ticker" aria-label="Capabilities">
        <div class="ticker-track">
          <span v-for="item in tickerItems" :key="item">{{ item }}</span>
        </div>
      </section>

      <section class="why" id="why">
        <div class="why-intro">
          <p class="section-kicker">Why this exists</p>
          <h2>The dangerous part is usually the workflow around the upgrade.</h2>
          <p>
            Teams usually have SSH and Compose. What they lack is a repeatable release path when the consequences are real.
          </p>
        </div>

        <div class="failure-grid">
          <article v-for="item in failureModes" :key="item.index" class="failure-card">
            <p class="failure-index">{{ item.index }}</p>
            <h3>{{ item.title }}</h3>
            <p>{{ item.copy }}</p>
          </article>
        </div>
      </section>

      <section class="flow" id="flow">
        <div class="flow-intro">
          <p class="section-kicker">Operating flow</p>
          <h2>Five moves from trust to audit.</h2>
          <p>
            SHUM is not a platform. It adds order to the moments where Compose upgrades usually become improvised.
          </p>
        </div>

        <div class="steps">
          <article v-for="item in flowSteps" :key="item.step" class="step-card">
            <div class="step-head">
              <span class="step-number">{{ item.step }}</span>
              <h3>{{ item.title }}</h3>
            </div>
            <p>{{ item.copy }}</p>
            <pre><code>{{ item.command }}</code></pre>
          </article>
        </div>
      </section>

      <section class="proof" id="proof">
        <div class="proof-summary">
          <p class="section-kicker">Operational proof</p>
          <h2>Trust comes from changing how upgrades are executed.</h2>
          <p>
            Preview before mutation, policy before execution, recovery before panic, and readable evidence after the run.
          </p>
        </div>

        <div class="proof-grid">
          <article v-for="item in proofCards" :key="item.label" class="proof-card">
            <p class="proof-label">{{ item.label }}</p>
            <h3>{{ item.title }}</h3>
            <p>{{ item.copy }}</p>
          </article>
        </div>

        <div class="evidence-grid">
          <article v-for="item in evidencePanels" :key="item.label" class="evidence-card">
            <p class="proof-label">{{ item.label }}</p>
            <h3>{{ item.title }}</h3>
            <p>{{ item.copy }}</p>
            <pre><code>{{ item.code }}</code></pre>
          </article>
        </div>
      </section>

      <section class="start" id="start">
        <div class="start-copy">
          <p class="section-kicker">Start here</p>
          <h2>Install the CLI and test the flow against a real host.</h2>
          <p>
            If the upgrade path becomes clearer, SHUM is doing its job.
          </p>
          <div class="start-meta">
            <p><strong>Config</strong> <code>~/.config/shum</code></p>
            <p><strong>State</strong> <code>~/.cache/shum</code></p>
            <p><strong>Validation</strong> <code>go test ./...</code> and optional <code>go test ./test/e2e</code></p>
          </div>
        </div>

        <div class="install-card">
          <pre><code>{{ installBlock }}</code></pre>
          <p class="install-note">Source build:</p>
          <pre><code>{{ sourceInstallBlock }}</code></pre>
        </div>
      </section>
    </main>

    <footer class="footer">
      <div class="footer-intro">
        <p class="section-kicker">Resources</p>
        <h2>Open-source tooling for teams that want explicit operations.</h2>
      </div>

      <div class="resource-grid">
        <a
          v-for="item in resourceCards"
          :key="item.title"
          :href="item.href"
          class="resource-card"
          target="_blank"
          rel="noopener noreferrer"
        >
          <p class="resource-title">{{ item.title }}</p>
          <p>{{ item.copy }}</p>
        </a>
      </div>
    </footer>
  </div>
</template>
