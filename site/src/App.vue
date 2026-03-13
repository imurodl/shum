<script setup>
import { onMounted } from 'vue'

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
    title: 'Preflight checks are skipped or inconsistent.',
    copy: 'There\'s no standard — image deltas, disk space, and health checks get eyeballed differently on every run.',
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
    copy: 'Your SSH config is the source of truth. SHUM uses your existing aliases — no new credentials or config files needed.',
    command: 'shum host register prod\nshum project discover prod',
  },
  {
    step: '02',
    title: 'Inspect and preflight the Compose project',
    copy: 'See the live compose config, running containers, and any readiness warnings before you change anything.',
    command:
      'shum project inspect prod web --project-directory /srv/web --json\nshum project preflight prod web --json',
  },
  {
    step: '03',
    title: 'Attach policy to the project itself',
    copy: 'Backup commands and health probes travel with the project — not in someone\'s head or a shared doc.',
    command:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \\"$SHUM_BACKUP_ARTIFACT\\"" \\\n  --health-check "http://127.0.0.1:8080/health"',
  },
  {
    step: '04',
    title: 'Plan and dry-run before mutation',
    copy: 'Preview exactly which images will change. Dry-run the full upgrade to confirm the plan before any container restarts.',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
  },
  {
    step: '05',
    title: 'Execute and inspect the resulting run',
    copy: 'Run the upgrade. Every outcome — services changed, backup taken, health check result — is recorded and queryable afterward.',
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
    title: 'What actually happened',
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

const sectionRefs = []
function registerSection(el) {
  if (el) sectionRefs.push(el)
}

onMounted(() => {
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('visible')
          observer.unobserve(entry.target)
        }
      })
    },
    { threshold: 0.08 }
  )

  sectionRefs.forEach((el, i) => {
    if (i === 0) {
      el.classList.add('visible')
    } else {
      observer.observe(el)
    }
  })
})

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
      <section class="hero" id="top" :ref="registerSection">
        <div class="hero-copy">
          <p class="section-kicker">For developers who self-host on Linux</p>
          <div class="hero-badge">
            <span class="board-status">v0.1.0</span>
            <span class="hero-badge-sep">·</span>
            <span class="hero-badge-license">Apache 2.0</span>
          </div>
          <p class="hero-index">01</p>
          <h1>Remote Docker Compose upgrades need procedure, not courage.</h1>
          <p class="hero-lead">
            SHUM is a CLI for Docker Compose stacks on self-hosted servers. Register your server once, attach a backup policy to each project, and run upgrades with dry-runs, health checks, and a full audit trail — no platform required.
          </p>

          <div class="hero-actions">
            <a
              class="button button-primary"
              href="https://github.com/imurodl/shum#quick-start"
              target="_blank"
              rel="noopener noreferrer"
            >
              Quick start
            </a>
            <a
              class="button button-secondary"
              href="https://github.com/imurodl/shum"
              target="_blank"
              rel="noopener noreferrer"
            >
              View on GitHub
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
              <p>Every run is logged. Summaries and backup artifacts stay queryable after the fact.</p>
            </article>
          </div>
        </aside>
      </section>

      <section class="ticker" aria-label="Capabilities">
        <div class="ticker-track" aria-live="off">
          <span v-for="item in tickerItems" :key="'a-' + item">{{ item }}</span>
          <span v-for="item in tickerItems" :key="'b-' + item" aria-hidden="true">{{ item }}</span>
        </div>
      </section>

      <section class="why" id="why" :ref="registerSection">
        <div class="why-intro">
          <p class="section-kicker">Why this exists</p>
          <h2>The dangerous part is the workflow around the upgrade.</h2>
          <p>
            SSH and Compose get you to production. They don't get you back when something breaks.
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

      <section class="flow" id="flow" :ref="registerSection">
        <div class="flow-intro">
          <p class="section-kicker">Operating flow</p>
          <h2>Five moves from trust to audit.</h2>
          <p>
            SHUM is not a platform. It adds order to the moments where Compose upgrades become improvised.
          </p>
        </div>

        <div class="steps">
          <article v-for="(item, index) in flowSteps" :key="item.step" class="step-card" :class="{ 'step-card--last': index === flowSteps.length - 1 }">
            <div class="step-head">
              <span class="step-number">{{ item.step }}</span>
              <h3>{{ item.title }}</h3>
            </div>
            <p>{{ item.copy }}</p>
            <pre><code>{{ item.command }}</code></pre>
          </article>
        </div>
      </section>

      <section class="proof" id="proof" :ref="registerSection">
        <div class="proof-summary">
          <p class="section-kicker">Operational proof</p>
          <h2>Confidence comes from seeing the path before you take it.</h2>
          <p>
            SHUM puts inspection, policy, and recovery into the normal workflow — not as afterthoughts you reach for when something breaks.
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

      <section class="start" id="start" :ref="registerSection">
        <div class="start-copy">
          <p class="section-kicker">Start here</p>
          <h2>Up and running in five minutes.</h2>
          <p>
            Point SHUM at a server you already SSH into. If the upgrade path feels clearer after the first run, it's doing its job.
          </p>
          <div class="start-meta">
            <p><strong>Config</strong> <code>~/.config/shum</code></p>
            <p><strong>State</strong> <code>~/.cache/shum</code></p>
          </div>
        </div>

        <div class="install-card">
          <pre><code>{{ installBlock }}</code></pre>
          <p class="install-note">Source build:</p>
          <pre><code>{{ sourceInstallBlock }}</code></pre>
        </div>
      </section>
    </main>

    <footer class="footer" :ref="registerSection">
      <div class="footer-intro">
        <p class="section-kicker">Resources</p>
        <h2>Open-source. Works with any Docker Compose stack on any Linux server.</h2>
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
