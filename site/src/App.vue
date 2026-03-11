<script setup>
const signals = [
  {
    title: 'Policy-first',
    copy: 'Backups, probes, and migration warnings are part of the upgrade contract.',
  },
  {
    title: 'Host-aware',
    copy: 'Trusted SSH aliases and project discovery keep execution tied to real infrastructure.',
  },
  {
    title: 'Recoverable',
    copy: 'Artifacts and restore paths are treated as normal operations, not crisis work.',
  },
]

const painPoints = [
  'Release knowledge lives in shell history, chat messages, and memory.',
  'Self-hosted Compose stacks often have no strict upgrade path.',
  'Rollback planning happens too late, usually after something breaks.',
  'Incident review gets blocked by incomplete run history.',
]

const ledger = [
  ['discover', 'trusted host alias registered'],
  ['preflight', 'compose environment and target validated'],
  ['policy', 'backup and health gates loaded'],
  ['dry-run', 'upgrade delta previewed before execution'],
  ['audit', 'run record stored with artifacts'],
]

const flow = [
  {
    label: '01 / map',
    title: 'Map the host and stack',
    copy: 'Register a trusted SSH target and discover the Compose project before making any changes.',
    command: 'shum host register prod\nshum project discover prod',
  },
  {
    label: '02 / inspect',
    title: 'Inspect reality',
    copy: 'Read project metadata, compare current state, and run preflight checks before mutation.',
    command: 'shum project inspect prod web --project-directory /srv/web --json\nshum project preflight prod web --json',
  },
  {
    label: '03 / govern',
    title: 'Load policy',
    copy: 'Backups and restore commands become executable policy, not oral tradition.',
    command:
      'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\""\nshum project backup take prod web --json',
  },
  {
    label: '04 / stage',
    title: 'Preview the change',
    copy: 'Generate the upgrade plan and dry-run first so the action is visible before it is real.',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
  },
  {
    label: '05 / execute',
    title: 'Execute and audit',
    copy: 'Run the upgrade intentionally and preserve the history needed for review or rollback.',
    command: 'shum project upgrade prod web --json\nshum project run list --host prod --project web --json',
  },
]

const commandAtlas = [
  ['shum host register', 'Create a stable trusted host reference from an SSH alias'],
  ['shum project discover', 'Locate canonical Compose projects on that host'],
  ['shum project inspect', 'Expose stack metadata, mounts, labels, and image baselines'],
  ['shum project preflight', 'Run readiness checks before mutation'],
  ['shum project plan', 'Compute the planned upgrade delta'],
  ['shum project policy', 'Persist backup, probe, and migration rules'],
  ['shum project backup', 'Take, list, and restore recovery artifacts'],
  ['shum project upgrade', 'Dry-run first, then execute the rollout'],
  ['shum project run', 'Inspect run history, summaries, and status changes'],
]

const installBlock = `git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum
shum --help`

const resources = [
  ['GitHub', 'https://github.com/imurodl/shum', 'Source, issues, releases'],
  ['Testing Guide', 'https://github.com/imurodl/shum/blob/main/docs/testing.md', 'Verification matrix and remote checks'],
  ['Contributing', 'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md', 'Contribution standards'],
  ['Security', 'https://github.com/imurodl/shum/blob/main/SECURITY.md', 'Security disclosure policy'],
  ['Changelog', 'https://github.com/imurodl/shum/blob/main/CHANGELOG.md', 'Project evolution'],
]
</script>

<template>
  <div class="site-shell">
    <div class="paper-noise" aria-hidden="true"></div>
    <div class="hero-word" aria-hidden="true">SHUM</div>

    <header class="site-header">
      <a class="brand" href="#top">
        <span class="brand-mark">SHUM</span>
        <span class="brand-copy">Self-Host Upgrade Manager</span>
      </a>

      <nav class="site-nav" aria-label="Primary">
        <a href="#flow">Flow</a>
        <a href="#atlas">CLI</a>
        <a href="#architecture">Architecture</a>
        <a href="#resources">Resources</a>
      </nav>
    </header>

    <main class="page">
      <section class="hero" id="top">
        <div class="hero-copy">
          <p class="eyebrow">self-hosted release operations</p>
          <h1>Make upgrades feel engineered, not improvised.</h1>
          <p class="lead">
            SHUM gives remote Docker Compose upgrades a defined operating model:
            discover, preflight, plan, back up, execute, and audit with explicit state.
          </p>

          <div class="hero-actions">
            <a class="button button-primary" href="https://github.com/imurodl/shum" target="_blank" rel="noopener noreferrer">View source</a>
            <a class="button button-secondary" href="#flow">See the flow</a>
            <a class="button button-secondary" href="https://github.com/imurodl/shum/blob/main/docs/testing.md" target="_blank" rel="noopener noreferrer">Read testing guide</a>
          </div>

          <div class="signal-grid">
            <article v-for="item in signals" :key="item.title" class="signal-card">
              <p class="signal-title">{{ item.title }}</p>
              <p>{{ item.copy }}</p>
            </article>
          </div>
        </div>

        <aside class="runbook-card">
          <div class="runbook-header">
            <span>release.runbook</span>
            <span>prod/web</span>
          </div>

          <div class="runbook-layout">
            <div>
              <p class="runbook-label">Protocol</p>
              <ol class="protocol-list">
                <li>trust the host</li>
                <li>inspect the stack</li>
                <li>load policy</li>
                <li>preview the change</li>
                <li>execute with history</li>
              </ol>
            </div>

            <div class="terminal-block">
              <p class="runbook-label">Bootstrap</p>
              <pre><code>{{ installBlock }}</code></pre>
            </div>
          </div>
        </aside>
      </section>

      <section class="manifesto panel">
        <div class="manifesto-copy">
          <p class="section-tag">Operational premise</p>
          <h2>If the release process only lives in someone’s head, it is already unreliable.</h2>
          <ul class="pain-list">
            <li v-for="item in painPoints" :key="item">{{ item }}</li>
          </ul>
        </div>

        <div class="ledger-card">
          <p class="section-tag">Run ledger</p>
          <div class="ledger-rows">
            <div v-for="item in ledger" :key="item[0]" class="ledger-row">
              <span class="ledger-phase">{{ item[0] }}</span>
              <span class="ledger-copy">{{ item[1] }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="workflow panel" id="flow">
        <div class="section-header">
          <p class="section-tag">Upgrade flow</p>
          <h2>Five deliberate stages before and after rollout.</h2>
        </div>

        <div class="workflow-grid">
          <article v-for="step in flow" :key="step.label" class="step-card">
            <p class="step-label">{{ step.label }}</p>
            <h3>{{ step.title }}</h3>
            <p class="step-copy">{{ step.copy }}</p>
            <pre><code>{{ step.command }}</code></pre>
          </article>
        </div>
      </section>

      <section class="atlas panel" id="atlas">
        <div class="atlas-copy">
          <p class="section-tag">Command atlas</p>
          <h2>A CLI surface built around upgrade discipline.</h2>
          <div class="atlas-list">
            <article v-for="item in commandAtlas" :key="item[0]" class="atlas-row">
              <code>{{ item[0] }}</code>
              <p>{{ item[1] }}</p>
            </article>
          </div>
        </div>

        <div class="install-card">
          <p class="section-tag">Install</p>
          <pre><code>{{ installBlock }}</code></pre>
          <p class="install-note">Config lives in <code>~/.config/shum</code>.</p>
          <p class="install-note">State and artifacts live in <code>~/.cache/shum</code>.</p>
          <p class="install-note">Validation: <code>go test ./...</code> and optional <code>go test ./test/e2e</code>.</p>
        </div>
      </section>

      <section class="architecture panel" id="architecture">
        <div class="section-header">
          <p class="section-tag">Architecture</p>
          <h2>Small surface area. Explicit layers.</h2>
        </div>

        <div class="architecture-rail">
          <article>
            <strong>CLI layer</strong>
            <span>Command parsing plus JSON and human-readable output modes.</span>
          </article>
          <article>
            <strong>Ops engine</strong>
            <span>Preflight, planning, upgrade execution, verification, and recovery.</span>
          </article>
          <article>
            <strong>Domain services</strong>
            <span>Hosts, projects, policies, run history, and artifact metadata.</span>
          </article>
          <article>
            <strong>Storage</strong>
            <span>SQLite state plus artifact directories for recovery workflows.</span>
          </article>
        </div>
      </section>

      <footer class="resource-strip panel" id="resources">
        <div class="section-header">
          <p class="section-tag">Resources</p>
          <h2>Everything needed to evaluate or contribute.</h2>
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
