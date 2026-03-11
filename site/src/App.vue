<script setup>
const coreValues = [
  {
    icon: '🔐',
    title: 'Trust-first host model',
    text: 'Register remote SSH identities first, then execute only on known-good targets with explicit host key validation.',
  },
  {
    icon: '🧭',
    title: 'Deterministic safety gates',
    text: 'Preflight checks, digest-aware plans, and policy rules are evaluated before any mutating step.',
  },
  {
    icon: '🧯',
    title: 'Built-in rollback path',
    text: 'Backups plus health verification drive explicit rollback decisions to keep state transitions reversible.',
  },
  {
    icon: '📈',
    title: 'Auditable operations',
    text: 'Every run is persisted with events, health signals, and backup references for incident review.',
  },
]

const workflow = [
  {
    title: '1. Trust the host',
    command: 'shum host register prod',
    note: 'Alias the SSH endpoint and bind trust boundaries before anything touches state.',
  },
  {
    title: '2. Discover and inspect',
    command: 'shum project discover prod',
    note: 'Discover compose projects and verify canonical config context before acting.',
  },
  {
    title: '3. Validate and plan',
    command: 'shum project preflight prod web && shum project plan prod web --json',
    note: 'Generate a deterministic plan and explicit blocker summary for each update.',
  },
  {
    title: '4. Execute with guardrails',
    command: 'shum project upgrade prod web --json',
    note: 'Optionally dry-run first, then run with health probes and auto rollback policy.',
  },
]

const commandPlaybook = [
  'shum project policy set prod web --require-backup=true \\\n  --backup-command "docker exec db pg_dumpall -U app > \"$SHUM_BACKUP_ARTIFACT\""',
  'shum project backup take prod web --json',
  'shum project run list --host prod --project web --json',
  'shum project run show run-<id> --json',
]
</script>

<template>
  <div class="landing">
    <div class="halo"></div>
    <header class="hero">
      <p class="eyebrow">Self-hosted operations with confidence</p>
      <h1>shum</h1>
      <p class="subtitle">A trust-first toolkit for deterministic Docker Compose upgrades across remote Linux hosts.</p>
      <div class="actions">
        <a class="btn btn-solid" href="https://github.com/imurodl/shum" target="_blank" rel="noopener">Read source</a>
        <a class="btn btn-ghost" href="#workflow">Inspect workflow</a>
      </div>
      <div class="meta-strip">Safe by default • Evidence by design • Built for operators</div>
    </header>

    <main>
      <section class="panel" :style="{ '--delay': '0.0s' }">
        <h2>What shum solves</h2>
        <div class="grid">
          <article class="tile" v-for="item in coreValues" :key="item.title">
            <p class="tile-icon">{{ item.icon }}</p>
            <h3>{{ item.title }}</h3>
            <p>{{ item.text }}</p>
          </article>
        </div>
      </section>

      <section class="panel" id="workflow" :style="{ '--delay': '0.15s' }">
        <h2>Operational flow</h2>
        <p class="section-subtitle">Four-phase operator path from trust to recovery, with visibility at every step.</p>
        <div class="steps">
          <article v-for="(item, index) in workflow" :key="item.title" class="step">
            <p class="step-index">0{{ index + 1 }}</p>
            <h3>{{ item.title }}</h3>
            <pre><code>{{ item.command }}</code></pre>
            <p>{{ item.note }}</p>
          </article>
        </div>
      </section>

      <section class="panel" :style="{ '--delay': '0.3s' }">
        <h2>Quick playbook</h2>
        <p class="section-subtitle">Composable commands you can run for controlled upgrades.</p>
        <ul class="playbook">
          <li v-for="command in commandPlaybook" :key="command">
            <code>{{ command }}</code>
          </li>
        </ul>
      </section>

      <section class="panel" :style="{ '--delay': '0.45s' }">
        <h2>Roadmap completion</h2>
        <p class="section-subtitle">Phase 1-4 delivery is implemented in this repository.</p>
        <div class="status">
          <span>HOST-01..03</span>
          <span>PLAN-01..04</span>
          <span>BKUP-01..03</span>
          <span>UPGD-01..04</span>
          <span>HIST-01..02</span>
          <span>DOCS-01..02</span>
        </div>
      </section>
    </main>
  </div>
</template>

