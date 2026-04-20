<script setup>
import { onMounted, ref } from 'vue'

const navLinks = [
  { label: 'Agents', href: '#agents' },
  { label: 'Harness', href: '#harness' },
  { label: 'Flow', href: '#flow' },
  { label: 'Start', href: '#start' },
]

const agentPillars = [
  {
    title: 'Every command speaks --json',
    copy: 'Read commands, plans, and runs all return parseable JSON. Agents never scrape human text.',
    lang: 'json',
    code: `{
  "host_alias": "prod",
  "project_ref": "web",
  "preflight": { "passed": true, "docker_version": "26.1.4" },
  "services": [
    {
      "service": "api",
      "current_digest": "sha256:9a1...",
      "target_digest": "sha256:b2f..."
    }
  ],
  "warnings": [],
  "blocks": []
}`,
  },
  {
    title: 'Errors carry stable codes',
    copy: 'On failure shum writes a typed envelope to stderr and exits with a documented code. Codes are part of the public surface — never renamed in a patch release.',
    lang: 'json',
    code: `{
  "error": {
    "code": "migration_warning",
    "message": "migration warning is enabled; use --force to continue",
    "hint": "review the plan, then re-run with --force if intentional",
    "details": { "host_alias": "prod", "project_ref": "web" }
  }
}`,
  },
  {
    title: 'Surface loads in one shot',
    copy: '`shum agent-help` emits the entire CLI surface — every command, flag, error code, and JSON shape — as a single JSON document. One call at session start.',
    lang: 'bash',
    code: `$ shum agent-help | jq '{
    commands: (.commands | length),
    errors: (.errors | length)
  }'
{
  "commands": 16,
  "errors": 22
}`,
  },
]

const harnessCards = [
  {
    name: 'Claude Code',
    tag: 'Skill + slash command',
    href: 'https://github.com/imurodl/shum/tree/main/examples/agents/claude-code',
    blurb: 'Trigger by natural language ("upgrade web on prod") or by /shum-upgrade. Includes the canonical safe-upgrade flow and hard failure-handling rules.',
    install: 'cp -r .claude/skills/shum ~/.claude/skills/',
  },
  {
    name: 'OpenAI Codex',
    tag: 'AGENTS.md drop-in',
    href: 'https://github.com/imurodl/shum/tree/main/examples/agents/codex',
    blurb: 'Codex reads AGENTS.md hierarchically. Drop the file in your repo or merge the rules into your existing one.',
    install: 'cp AGENTS.md ./AGENTS.md',
  },
  {
    name: 'Gemini CLI',
    tag: 'GEMINI.md drop-in',
    href: 'https://github.com/imurodl/shum/tree/main/examples/agents/gemini',
    blurb: 'Gemini CLI loads GEMINI.md hierarchically and supports @file imports. Reference shum\'s rules from your existing GEMINI.md.',
    install: 'cp GEMINI.md ~/.gemini/GEMINI.md',
  },
]

const flowSteps = [
  {
    step: '01',
    title: 'Load the surface',
    copy: 'Once per session. Every command, every flag, every error code, every output shape — one JSON document.',
    command: 'shum agent-help | jq .',
    lang: 'bash',
  },
  {
    step: '02',
    title: 'Discover hosts and projects',
    copy: 'SSH aliases are the identity. Discover compose projects already running on a host.',
    command: 'shum host list --json\nshum project discover prod --json',
    lang: 'bash',
  },
  {
    step: '03',
    title: 'Read the policy',
    copy: 'Backup commands, restore commands, and health checks travel with the project — not with the operator.',
    command: 'shum project policy show prod web --json',
    lang: 'bash',
  },
  {
    step: '04',
    title: 'Plan, then dry-run',
    copy: 'Preview which images will change and run the full upgrade flow without mutating anything.',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
    lang: 'bash',
  },
  {
    step: '05',
    title: 'Execute, then audit',
    copy: 'Real upgrade. One JSON record per run: status, services changed, backup taken, health probe outcomes.',
    command: 'shum project upgrade prod web --json\nshum project run show <run-id> --json',
    lang: 'bash',
  },
]

const proofSession = `$ shum project upgrade prod web --json
{
  "run_id": "run-1714834290291",
  "status": "rolled_back",
  "summary": "compose pull failed: connection reset"
}
$ echo $?
68

$ shum project upgrade prod web --dry-run --json 2>/dev/null | \\
    jq '{services: .services | length, blocks, warnings}'
{
  "services": 3,
  "blocks": [],
  "warnings": []
}`

const installBlock = `go install github.com/imurodl/shum/cmd/shum@latest
shum agent-help`

const sourceInstallBlock = `git clone https://github.com/imurodl/shum.git
cd shum
go install ./cmd/shum`

const skillInstallBlock = `# Drop the Claude Code skill into ~/.claude/
cp -r examples/agents/claude-code/.claude/skills/shum \\
  ~/.claude/skills/`

const resourceCards = [
  {
    title: 'Agent contract',
    href: 'https://github.com/imurodl/shum/blob/main/AGENTS.md',
    copy: 'Full agent contract: command surface, error codes, exit codes, failure-handling rules.',
  },
  {
    title: 'Repository',
    href: 'https://github.com/imurodl/shum',
    copy: 'Source, issues, releases, and implementation detail.',
  },
  {
    title: 'Quickstart',
    href: 'https://github.com/imurodl/shum#quickstart-for-ai-agents',
    copy: 'Five-step quickstart for AI agents and humans alike.',
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

const highlighter = ref(null)

function escapeHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function renderCode(code, lang = 'bash') {
  if (highlighter.value) {
    return highlighter.value.codeToHtml(code, {
      lang,
      theme: 'github-dark-default',
    })
  }
  return `<pre><code>${escapeHtml(code)}</code></pre>`
}

onMounted(async () => {
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

  // Explicit imports keep the bundle small: only 3 langs + 1 theme + JS regex engine.
  const [
    { createHighlighterCore, createJavaScriptRegexEngine },
    json,
    bash,
    yaml,
    githubDarkDefault,
  ] = await Promise.all([
    import('shiki/core'),
    import('shiki/langs/json.mjs').then((m) => m.default),
    import('shiki/langs/bash.mjs').then((m) => m.default),
    import('shiki/langs/yaml.mjs').then((m) => m.default),
    import('shiki/themes/github-dark-default.mjs').then((m) => m.default),
  ])
  highlighter.value = await createHighlighterCore({
    themes: [githubDarkDefault],
    langs: [json, bash, yaml],
    engine: createJavaScriptRegexEngine(),
  })
})
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
        <span class="brand-copy">agent-driveable Compose CLI</span>
      </a>

      <nav class="nav" aria-label="Primary">
        <a v-for="item in navLinks" :key="item.href" :href="item.href">{{ item.label }}</a>
      </nav>

      <a
        class="masthead-link"
        href="https://github.com/imurodl/shum"
        target="_blank"
        rel="noopener noreferrer"
      >
        View on GitHub
      </a>
    </header>

    <main class="page">
      <section class="hero" id="top" :ref="registerSection">
        <div class="hero-copy">
          <p class="section-kicker">For AI coding agents and the humans who run them</p>
          <div class="hero-badge">
            <span class="board-status">v0.1.0</span>
            <span class="hero-badge-sep">·</span>
            <span class="hero-badge-license">Apache 2.0</span>
          </div>
          <h1>The Compose upgrade CLI your AI agent can drive.</h1>
          <p class="hero-lead">
            shum is a CLI for safe, recoverable Docker Compose upgrades on remote SSH hosts. Every command speaks <code>--json</code>, errors return stable codes, and the entire surface loads in one shot via <code>shum agent-help</code>. Use it from Claude Code, Codex, Gemini CLI, or your terminal.
          </p>

          <div class="hero-actions">
            <a
              class="button button-primary"
              href="https://github.com/imurodl/shum/blob/main/AGENTS.md"
              target="_blank"
              rel="noopener noreferrer"
            >
              Read the agent contract
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
              <p class="section-kicker">Error envelope</p>
              <h2>stable, parseable</h2>
            </div>
            <span class="board-status">stderr · exit 68</span>
          </div>

          <div class="code" v-html="renderCode(agentPillars[1].code, 'json')" />

          <div class="board-notes">
            <article>
              <span>code</span>
              <p>Stable across patch releases. Parse <code>.error.code</code>, never the message.</p>
            </article>
            <article>
              <span>hint</span>
              <p>Operator guidance the agent can surface verbatim to the user.</p>
            </article>
            <article>
              <span>details</span>
              <p>Structured context: which alias, which project, which artifact.</p>
            </article>
          </div>
        </aside>
      </section>

      <section class="agents" id="agents" :ref="registerSection">
        <div class="agents-intro">
          <p class="section-kicker">Designed for agents</p>
          <h2>Three contracts. One CLI.</h2>
          <p>
            LLMs already know SSH and Docker from training. shum gives them a typed surface on top — so the agent makes the right call, not a clever guess.
          </p>
        </div>

        <div class="pillars">
          <article v-for="pillar in agentPillars" :key="pillar.title" class="pillar-card">
            <h3>{{ pillar.title }}</h3>
            <p>{{ pillar.copy }}</p>
            <div class="code" v-html="renderCode(pillar.code, pillar.lang)" />
          </article>
        </div>
      </section>

      <section class="harness" id="harness" :ref="registerSection">
        <div class="harness-intro">
          <p class="section-kicker">Works with your agent</p>
          <h2>Ready-to-use configs for the major harnesses.</h2>
          <p>
            Each example is a small drop-in: an instructions file, a sample prompt, and a README. Use them as-is or merge the rules into your existing setup.
          </p>
        </div>

        <div class="harness-grid">
          <a
            v-for="card in harnessCards"
            :key="card.name"
            class="harness-card"
            :href="card.href"
            target="_blank"
            rel="noopener noreferrer"
          >
            <div class="harness-head">
              <h3>{{ card.name }}</h3>
              <p class="harness-tag">{{ card.tag }}</p>
            </div>
            <p>{{ card.blurb }}</p>
            <code class="harness-install">{{ card.install }}</code>
          </a>
        </div>
      </section>

      <section class="flow" id="flow" :ref="registerSection">
        <div class="flow-intro">
          <p class="section-kicker">Operating flow</p>
          <h2>Five moves from cold start to audited upgrade.</h2>
          <p>
            Same flow whether an agent or a human is driving. The CLI returns the same JSON either way.
          </p>
        </div>

        <div class="steps">
          <article
            v-for="(item, index) in flowSteps"
            :key="item.step"
            class="step-card"
            :class="{ 'step-card--last': index === flowSteps.length - 1 }"
          >
            <div class="step-head">
              <span class="step-number">{{ item.step }}</span>
              <h3>{{ item.title }}</h3>
            </div>
            <p>{{ item.copy }}</p>
            <div class="code" v-html="renderCode(item.command, item.lang)" />
          </article>
        </div>
      </section>

      <section class="proof" id="proof" :ref="registerSection">
        <div class="proof-summary">
          <p class="section-kicker">A real session</p>
          <h2>Read the run, not the terminal.</h2>
          <p>
            Failures route through the same envelope as successes. Exit codes follow a documented table — agents branch on <code>.error.code</code>, never on regex.
          </p>
        </div>

        <div class="proof-board">
          <div class="code" v-html="renderCode(proofSession, 'bash')" />
        </div>
      </section>

      <section class="start" id="start" :ref="registerSection">
        <div class="start-copy">
          <p class="section-kicker">Start here</p>
          <h2>Install. Register. Hand it to your agent.</h2>
          <p>
            Point shum at a server you already SSH into. The first <code>shum agent-help</code> call gives any agent everything it needs.
          </p>
          <div class="start-meta">
            <p><strong>Config</strong> <code>~/.config/shum</code></p>
            <p><strong>State</strong> <code>~/.cache/shum</code></p>
          </div>
        </div>

        <div class="install-stack">
          <div class="install-card">
            <p class="install-note">Install the CLI</p>
            <div class="code" v-html="renderCode(installBlock, 'bash')" />
            <p class="install-note">From source</p>
            <div class="code" v-html="renderCode(sourceInstallBlock, 'bash')" />
          </div>
          <div class="install-card">
            <p class="install-note">Install the Claude Code skill</p>
            <div class="code" v-html="renderCode(skillInstallBlock, 'bash')" />
          </div>
        </div>
      </section>
    </main>

    <footer class="footer" :ref="registerSection">
      <div class="footer-intro">
        <p class="section-kicker">Resources</p>
        <h2>Open-source. Apache 2.0. Built for self-hosted Linux.</h2>
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
