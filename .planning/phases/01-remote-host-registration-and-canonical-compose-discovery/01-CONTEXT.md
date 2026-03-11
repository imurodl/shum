# Phase 1: Remote Host Registration and Canonical Compose Discovery - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Establish trusted SSH host registration and canonical Docker Compose discovery for self-hosted Linux machines. This phase covers how a user connects a host, how the tool discovers Compose projects, and how the tool presents the canonical rendered state before any upgrade workflow exists. Preflight checks, backup logic, upgrade execution, and rollback behavior are separate phases.

</domain>

<decisions>
## Implementation Decisions

### Host registration workflow
- v1 is SSH-alias-first: users should point the tool at an existing SSH config alias rather than re-entering host details into a custom workflow
- Strict SSH host-key verification is required in v1; there is no insecure or trust-on-first-use default path
- v1 assumes existing SSH keys and normal SSH agent workflows; password-based login and tool-managed keys are out of scope for this phase
- The SSH alias should be the primary host identity shown in tool output and history

### Discovery behavior
- Discovery should start with known Compose projects first rather than aggressively scanning the whole host by default
- Deeper scanning should be explicit and limited to user-chosen directories, not automatic whole-host crawling
- If the tool finds a Compose project that cannot be fully rendered, it should still surface it, but clearly mark it as blocked or incomplete
- If canonical rendering depends on profiles or environment context, the tool should require explicit user choice before treating the result as canonical

### Inspection experience
- Default discovery output should be a concise summary, not a wall of rendered details
- Human-readable terminal output is the default; structured JSON should be available via an explicit flag
- The inspect experience must always make project identity and trust/risk status obvious
- Deeper details should be exposed progressively through explicit drill-down commands or flags rather than one giant default dump

### Claude's Discretion
- Exact CLI command names and flag names
- Exact wording, formatting, and severity levels for blocked or ambiguous discovery states
- Exact table layout, color usage, and JSON schema shape
- How many drill-down commands or subviews exist, as long as the summary-first pattern stays intact

</decisions>

<specifics>
## Specific Ideas

- The tool should feel natural to self-hosters who already use SSH aliases in their terminal workflow
- Discovery should feel trustworthy, not magical; when the tool is unsure, it should say so clearly instead of guessing
- The first thing a user should understand from inspect output is "what project is this?" and "does the tool trust this as canonical?"

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- None yet — the repository currently contains planning artifacts only

### Established Patterns
- CLI-first product direction is already locked at the project level
- Public docs exist as a separate product surface and should not drive Phase 1 behavior
- Strict trust and safe-change positioning should shape Phase 1 messaging and defaults

### Integration Points
- Phase 1 should establish the host and project identity model that Phase 2 planning and safety gates will build on
- Canonical discovery output should produce the state later phases rely on for backup planning, digest planning, and rollback metadata
- Any persisted host/project inventory created here becomes the foundation for run history in later phases

</code_context>

<deferred>
## Deferred Ideas

- Password-based SSH authentication — possible later extension, not part of Phase 1
- Tool-managed SSH credentials — deferred until there is a strong reason to own credential workflows
- Whole-host filesystem scanning by default — deferred to avoid unsafe or noisy discovery behavior in v1
- Rich web UI for host and project inspection — outside this phase and outside the current CLI-first v1 focus

</deferred>

---
*Phase: 01-remote-host-registration-and-canonical-compose-discovery*
*Context gathered: 2026-03-11*
