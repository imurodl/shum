<script setup>
import { computed, onMounted, ref } from 'vue'

// ---------------------------------------------------------------------------
// i18n
// ---------------------------------------------------------------------------
const LOCALES = ['en', 'ko']
const STORAGE_KEY = 'shum-locale'

function detectLocale() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved && LOCALES.includes(saved)) return saved
  } catch {
    /* localStorage unavailable */
  }
  const nav = (navigator.language || 'en').toLowerCase()
  return nav.startsWith('ko') ? 'ko' : 'en'
}

const locale = ref(detectLocale())

function applyLocale(l) {
  document.documentElement.lang = l
  try {
    localStorage.setItem(STORAGE_KEY, l)
  } catch {
    /* ignore */
  }
}

function toggleLocale() {
  locale.value = locale.value === 'en' ? 'ko' : 'en'
  applyLocale(locale.value)
}

const messages = {
  en: {
    brandCopy: 'agent-driveable Compose CLI',
    githubLink: 'View on GitHub',
    langToggle: '한국어',
    nav: ['Agents', 'Harness', 'Flow', 'Start'],
    hero: {
      kicker: 'For AI coding agents and the humans who run them',
      h1: 'The Compose upgrade CLI your AI agent can drive.',
      lead: 'shum is a CLI for safe, recoverable Docker Compose upgrades on remote SSH hosts. Every command speaks <code>--json</code>, errors return stable codes, and the entire surface loads in one shot via <code>shum agent-help</code>. Use it from Claude Code, Codex, Gemini CLI, or your terminal.',
      primaryCta: 'Read the agent contract',
      secondaryCta: 'View on GitHub',
    },
    board: {
      kicker: 'Error envelope',
      h2: 'stable, parseable',
      status: 'stderr · exit 68',
      notes: {
        code: 'Stable across patch releases. Parse <code>.error.code</code>, never the message.',
        hint: 'Operator guidance the agent can surface verbatim to the user.',
        details: 'Structured context: which alias, which project, which artifact.',
      },
    },
    agents: {
      kicker: 'Designed for agents',
      h2: 'Three contracts. One CLI.',
      copy: 'LLMs already know SSH and Docker from training. shum gives them a typed surface on top — so the agent makes the right call, not a clever guess.',
      pillars: [
        {
          title: 'Every command speaks --json',
          copy: 'Read commands, plans, and runs all return parseable JSON. Agents never scrape human text.',
        },
        {
          title: 'Errors carry stable codes',
          copy: 'On failure shum writes a typed envelope to stderr and exits with a documented code. Codes are part of the public surface — never renamed in a patch release.',
        },
        {
          title: 'Surface loads in one shot',
          copy: '`shum agent-help` emits the entire CLI surface — every command, flag, error code, and JSON shape — as a single JSON document. One call at session start.',
        },
      ],
    },
    harness: {
      kicker: 'Works with your agent',
      h2: 'Ready-to-use configs for the major harnesses.',
      copy: 'Each example is a small drop-in: an instructions file, a sample prompt, and a README. Use them as-is or merge the rules into your existing setup.',
      cards: [
        {
          tag: 'Skill + slash command',
          blurb: 'Trigger by natural language ("upgrade web on prod") or by /shum-upgrade. Includes the canonical safe-upgrade flow and hard failure-handling rules.',
        },
        {
          tag: 'AGENTS.md drop-in',
          blurb: 'Codex reads AGENTS.md hierarchically. Drop the file in your repo or merge the rules into your existing one.',
        },
        {
          tag: 'GEMINI.md drop-in',
          blurb: 'Gemini CLI loads GEMINI.md hierarchically and supports @file imports. Reference shum\'s rules from your existing GEMINI.md.',
        },
      ],
    },
    flow: {
      kicker: 'Operating flow',
      h2: 'Five moves from cold start to audited upgrade.',
      copy: 'Same flow whether an agent or a human is driving. The CLI returns the same JSON either way.',
      steps: [
        {
          title: 'Load the surface',
          copy: 'Once per session. Every command, every flag, every error code, every output shape — one JSON document.',
        },
        {
          title: 'Discover hosts and projects',
          copy: 'SSH aliases are the identity. Discover compose projects already running on a host.',
        },
        {
          title: 'Read the policy',
          copy: 'Backup commands, restore commands, and health checks travel with the project — not with the operator.',
        },
        {
          title: 'Plan, then dry-run',
          copy: 'Preview which images will change and run the full upgrade flow without mutating anything.',
        },
        {
          title: 'Execute, then audit',
          copy: 'Real upgrade. One JSON record per run: status, services changed, backup taken, health probe outcomes.',
        },
      ],
    },
    proof: {
      kicker: 'A real session',
      h2: 'Read the run, not the terminal.',
      copy: 'Failures route through the same envelope as successes. Exit codes follow a documented table — agents branch on <code>.error.code</code>, never on regex.',
    },
    start: {
      kicker: 'Start here',
      h2: 'Install. Register. Hand it to your agent.',
      copy: 'Point shum at a server you already SSH into. The first <code>shum agent-help</code> call gives any agent everything it needs.',
      configLabel: 'Config',
      stateLabel: 'State',
      installCli: 'Install the CLI',
      fromSource: 'From source',
      installSkill: 'Install the Claude Code skill',
    },
    footer: {
      kicker: 'Resources',
      h2: 'Open-source. Apache 2.0. Built for self-hosted Linux.',
      resources: [
        {
          title: 'Agent contract',
          copy: 'Full agent contract: command surface, error codes, exit codes, failure-handling rules.',
        },
        {
          title: 'Repository',
          copy: 'Source, issues, releases, and implementation detail.',
        },
        {
          title: 'Quickstart',
          copy: 'Five-step quickstart for AI agents and humans alike.',
        },
        {
          title: 'Contributing',
          copy: 'Contribution standards and workflow expectations.',
        },
      ],
    },
  },
  ko: {
    brandCopy: '에이전트가 다루는 Compose CLI',
    githubLink: 'GitHub에서 보기',
    langToggle: 'English',
    nav: ['에이전트', '하네스', '플로우', '시작하기'],
    hero: {
      kicker: 'AI 코딩 에이전트와 이를 다루는 사람들을 위해',
      h1: 'AI 에이전트가 직접 다루는 Compose 업그레이드 CLI.',
      lead: 'shum은 원격 SSH 호스트에서 안전하고 복구 가능한 Docker Compose 업그레이드를 수행하는 CLI입니다. 모든 명령이 <code>--json</code>으로 응답하고, 오류는 안정적인 코드로 반환되며, 전체 인터페이스는 <code>shum agent-help</code> 한 번으로 로드됩니다. Claude Code, Codex, Gemini CLI, 또는 터미널에서 사용하세요.',
      primaryCta: '에이전트 계약 읽기',
      secondaryCta: 'GitHub에서 보기',
    },
    board: {
      kicker: '오류 엔벨로프',
      h2: '안정적이고 파싱 가능',
      status: 'stderr · exit 68',
      notes: {
        code: '패치 릴리스 간에도 안정적입니다. 메시지가 아니라 <code>.error.code</code>를 파싱하세요.',
        hint: '에이전트가 사용자에게 그대로 전달할 수 있는 운영 안내입니다.',
        details: '구조화된 컨텍스트: 어떤 별칭, 어떤 프로젝트, 어떤 아티팩트인지.',
      },
    },
    agents: {
      kicker: '에이전트를 위한 설계',
      h2: '세 가지 계약, 하나의 CLI.',
      copy: 'LLM은 학습을 통해 이미 SSH와 Docker를 알고 있습니다. shum은 그 위에 타입이 정의된 인터페이스를 제공하여, 에이전트가 그럴듯한 추측이 아니라 올바른 판단을 내리도록 합니다.',
      pillars: [
        {
          title: '모든 명령이 --json으로 응답합니다',
          copy: '조회 명령, 계획, 실행 모두 파싱 가능한 JSON을 반환합니다. 에이전트가 사람이 읽는 텍스트를 긁어올 필요가 없습니다.',
        },
        {
          title: '오류는 안정적인 코드를 담습니다',
          copy: '실패 시 shum은 타입이 정의된 엔벨로프를 stderr에 기록하고 문서화된 코드로 종료합니다. 코드는 공개 인터페이스의 일부로, 패치 릴리스에서 이름이 바뀌지 않습니다.',
        },
        {
          title: '인터페이스를 한 번에 로드합니다',
          copy: '`shum agent-help`는 모든 명령, 플래그, 오류 코드, JSON 구조 등 전체 CLI 인터페이스를 하나의 JSON 문서로 출력합니다. 세션 시작 시 한 번만 호출하면 됩니다.',
        },
      ],
    },
    harness: {
      kicker: '당신의 에이전트와 함께',
      h2: '주요 하네스를 위한 즉시 사용 가능한 설정.',
      copy: '각 예제는 간단한 드롭인 형태입니다: 지침 파일, 샘플 프롬프트, README. 그대로 사용하거나 기존 설정에 규칙을 병합하세요.',
      cards: [
        {
          tag: '스킬 + 슬래시 명령',
          blurb: '자연어("prod의 web 업그레이드")나 /shum-upgrade로 실행합니다. 표준 안전 업그레이드 플로우와 엄격한 실패 처리 규칙이 포함되어 있습니다.',
        },
        {
          tag: 'AGENTS.md 드롭인',
          blurb: 'Codex는 AGENTS.md를 계층적으로 읽습니다. 파일을 저장소에 넣거나 기존 파일에 규칙을 병합하세요.',
        },
        {
          tag: 'GEMINI.md 드롭인',
          blurb: 'Gemini CLI는 GEMINI.md를 계층적으로 로드하며 @file 임포트를 지원합니다. 기존 GEMINI.md에서 shum의 규칙을 참조하세요.',
        },
      ],
    },
    flow: {
      kicker: '운영 플로우',
      h2: '콜드 스타트부터 감사된 업그레이드까지 다섯 단계.',
      copy: '에이전트가 다루든 사람이 다루든 동일한 플로우입니다. CLI는 어느 쪽이든 같은 JSON을 반환합니다.',
      steps: [
        {
          title: '인터페이스 로드',
          copy: '세션당 한 번. 모든 명령, 플래그, 오류 코드, 출력 구조를 하나의 JSON 문서로.',
        },
        {
          title: '호스트와 프로젝트 탐색',
          copy: 'SSH 별칭이 곧 신원입니다. 호스트에서 이미 실행 중인 compose 프로젝트를 탐색합니다.',
        },
        {
          title: '정책 읽기',
          copy: '백업 명령, 복원 명령, 헬스 체크는 운영자가 아니라 프로젝트와 함께 이동합니다.',
        },
        {
          title: '계획 후 드라이런',
          copy: '어떤 이미지가 변경될지 미리 보고, 아무것도 변경하지 않은 채 전체 업그레이드 플로우를 실행합니다.',
        },
        {
          title: '실행 후 감사',
          copy: '실제 업그레이드. 실행마다 하나의 JSON 기록: 상태, 변경된 서비스, 수행된 백업, 헬스 프로브 결과.',
        },
      ],
    },
    proof: {
      kicker: '실제 세션',
      h2: '터미널이 아니라 실행 기록을 읽으세요.',
      copy: '실패도 성공과 동일한 엔벨로프를 거칩니다. 종료 코드는 문서화된 표를 따르며, 에이전트는 정규식이 아니라 <code>.error.code</code>로 분기합니다.',
    },
    start: {
      kicker: '여기서 시작',
      h2: '설치하고, 등록하고, 에이전트에게 넘기세요.',
      copy: '이미 SSH로 접속하는 서버에 shum을 연결하세요. 첫 <code>shum agent-help</code> 호출만으로 어떤 에이전트든 필요한 모든 것을 얻습니다.',
      configLabel: '설정',
      stateLabel: '상태',
      installCli: 'CLI 설치',
      fromSource: '소스에서 설치',
      installSkill: 'Claude Code 스킬 설치',
    },
    footer: {
      kicker: '리소스',
      h2: '오픈소스. Apache 2.0. 셀프 호스팅 Linux를 위해.',
      resources: [
        {
          title: '에이전트 계약',
          copy: '전체 에이전트 계약: 명령 인터페이스, 오류 코드, 종료 코드, 실패 처리 규칙.',
        },
        {
          title: '저장소',
          copy: '소스, 이슈, 릴리스, 구현 세부사항.',
        },
        {
          title: '퀵스타트',
          copy: 'AI 에이전트와 사람 모두를 위한 5단계 퀵스타트.',
        },
        {
          title: '기여하기',
          copy: '기여 표준과 워크플로우 기준.',
        },
      ],
    },
  },
}

const t = computed(() => messages[locale.value])

// ---------------------------------------------------------------------------
// Locale-independent data (code blocks, links, commands) merged with copy
// ---------------------------------------------------------------------------
const navHrefs = ['#agents', '#harness', '#flow', '#start']

const pillarCode = [
  {
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

const harnessMeta = [
  {
    name: 'Claude Code',
    href: 'https://github.com/imurodl/shum/tree/main/examples/agents/claude-code',
    install: 'cp -r .claude/skills/shum ~/.claude/skills/',
  },
  {
    name: 'OpenAI Codex',
    href: 'https://github.com/imurodl/shum/tree/main/examples/agents/codex',
    install: 'cp AGENTS.md ./AGENTS.md',
  },
  {
    name: 'Gemini CLI',
    href: 'https://github.com/imurodl/shum/tree/main/examples/agents/gemini',
    install: 'cp GEMINI.md ~/.gemini/GEMINI.md',
  },
]

const flowMeta = [
  { step: '01', command: 'shum agent-help | jq .', lang: 'bash' },
  {
    step: '02',
    command: 'shum host list --json\nshum project discover prod --json',
    lang: 'bash',
  },
  { step: '03', command: 'shum project policy show prod web --json', lang: 'bash' },
  {
    step: '04',
    command: 'shum project plan prod web --json\nshum project upgrade prod web --dry-run --json',
    lang: 'bash',
  },
  {
    step: '05',
    command: 'shum project upgrade prod web --json\nshum project run show <run-id> --json',
    lang: 'bash',
  },
]

const resourceHrefs = [
  'https://github.com/imurodl/shum/blob/main/AGENTS.md',
  'https://github.com/imurodl/shum',
  'https://github.com/imurodl/shum#quickstart-for-ai-agents',
  'https://github.com/imurodl/shum/blob/main/CONTRIBUTING.md',
]

const navLinks = computed(() =>
  t.value.nav.map((label, i) => ({ label, href: navHrefs[i] }))
)
const agentPillars = computed(() =>
  t.value.agents.pillars.map((p, i) => ({ ...p, ...pillarCode[i] }))
)
const harnessCards = computed(() =>
  t.value.harness.cards.map((c, i) => ({ ...c, ...harnessMeta[i] }))
)
const flowSteps = computed(() =>
  t.value.flow.steps.map((s, i) => ({ ...s, ...flowMeta[i] }))
)
const resourceCards = computed(() =>
  t.value.footer.resources.map((r, i) => ({ ...r, href: resourceHrefs[i] }))
)

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
  applyLocale(locale.value)

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
        <span class="brand-copy">{{ t.brandCopy }}</span>
      </a>

      <nav class="nav" aria-label="Primary">
        <a v-for="item in navLinks" :key="item.href" :href="item.href">{{ item.label }}</a>
      </nav>

      <div class="masthead-actions">
        <button
          type="button"
          class="lang-toggle"
          @click="toggleLocale"
          :aria-label="locale === 'en' ? 'Switch to Korean' : '영어로 전환'"
        >
          {{ t.langToggle }}
        </button>
        <a
          class="masthead-link"
          href="https://github.com/imurodl/shum"
          target="_blank"
          rel="noopener noreferrer"
        >
          {{ t.githubLink }}
        </a>
      </div>
    </header>

    <main class="page">
      <section class="hero" id="top" :ref="registerSection">
        <div class="hero-copy">
          <p class="section-kicker">{{ t.hero.kicker }}</p>
          <div class="hero-badge">
            <span class="board-status">v0.1.0</span>
            <span class="hero-badge-sep">·</span>
            <span class="hero-badge-license">Apache 2.0</span>
          </div>
          <h1>{{ t.hero.h1 }}</h1>
          <p class="hero-lead" v-html="t.hero.lead"></p>

          <div class="hero-actions">
            <a
              class="button button-primary"
              href="https://github.com/imurodl/shum/blob/main/AGENTS.md"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ t.hero.primaryCta }}
            </a>
            <a
              class="button button-secondary"
              href="https://github.com/imurodl/shum"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ t.hero.secondaryCta }}
            </a>
          </div>
        </div>

        <aside class="hero-board">
          <div class="board-header">
            <div>
              <p class="section-kicker">{{ t.board.kicker }}</p>
              <h2>{{ t.board.h2 }}</h2>
            </div>
            <span class="board-status">{{ t.board.status }}</span>
          </div>

          <div class="code" v-html="renderCode(agentPillars[1].code, 'json')" />

          <div class="board-notes">
            <article>
              <span>code</span>
              <p v-html="t.board.notes.code"></p>
            </article>
            <article>
              <span>hint</span>
              <p>{{ t.board.notes.hint }}</p>
            </article>
            <article>
              <span>details</span>
              <p>{{ t.board.notes.details }}</p>
            </article>
          </div>
        </aside>
      </section>

      <section class="agents" id="agents" :ref="registerSection">
        <div class="agents-intro">
          <p class="section-kicker">{{ t.agents.kicker }}</p>
          <h2>{{ t.agents.h2 }}</h2>
          <p>{{ t.agents.copy }}</p>
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
          <p class="section-kicker">{{ t.harness.kicker }}</p>
          <h2>{{ t.harness.h2 }}</h2>
          <p>{{ t.harness.copy }}</p>
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
          <p class="section-kicker">{{ t.flow.kicker }}</p>
          <h2>{{ t.flow.h2 }}</h2>
          <p>{{ t.flow.copy }}</p>
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
          <p class="section-kicker">{{ t.proof.kicker }}</p>
          <h2>{{ t.proof.h2 }}</h2>
          <p v-html="t.proof.copy"></p>
        </div>

        <div class="proof-board">
          <div class="code" v-html="renderCode(proofSession, 'bash')" />
        </div>
      </section>

      <section class="start" id="start" :ref="registerSection">
        <div class="start-copy">
          <p class="section-kicker">{{ t.start.kicker }}</p>
          <h2>{{ t.start.h2 }}</h2>
          <p v-html="t.start.copy"></p>
          <div class="start-meta">
            <p><strong>{{ t.start.configLabel }}</strong> <code>~/.config/shum</code></p>
            <p><strong>{{ t.start.stateLabel }}</strong> <code>~/.cache/shum</code></p>
          </div>
        </div>

        <div class="install-stack">
          <div class="install-card">
            <p class="install-note">{{ t.start.installCli }}</p>
            <div class="code" v-html="renderCode(installBlock, 'bash')" />
            <p class="install-note">{{ t.start.fromSource }}</p>
            <div class="code" v-html="renderCode(sourceInstallBlock, 'bash')" />
          </div>
          <div class="install-card">
            <p class="install-note">{{ t.start.installSkill }}</p>
            <div class="code" v-html="renderCode(skillInstallBlock, 'bash')" />
          </div>
        </div>
      </section>
    </main>

    <footer class="footer" :ref="registerSection">
      <div class="footer-intro">
        <p class="section-kicker">{{ t.footer.kicker }}</p>
        <h2>{{ t.footer.h2 }}</h2>
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
