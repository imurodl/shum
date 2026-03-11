# Self-Host Upgrade Manager

## What This Is

An open-source, CLI-first reliability tool for self-hosters running Docker Compose apps on Linux. It connects to remote hosts over SSH, discovers Compose projects, and executes safer upgrade workflows with preflight checks, backup hooks, health verification, and rollback when an update fails. A public landing/docs site introduces the tool, explains the architecture, and provides installation and usage guidance.

## Core Value

Self-hosters can upgrade Docker Compose apps with confidence because every change is checked, backed up, verified, and reversible.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] User can connect a Linux host over SSH and discover Docker Compose projects on that host
- [ ] User can run a safe upgrade workflow with preflight checks, backup/snapshot hooks, post-upgrade health checks, and rollback on failure
- [ ] User can inspect the history and outcome of upgrade runs to understand exactly what happened during each change
- [ ] User can use a public landing/docs site to understand the product, install it, and follow guided setup and usage documentation

### Out of Scope

- Kubernetes support in v1 — Compose-first scope keeps the initial product focused and finishable
- systemd-only service management in v1 — adds a second operational model too early
- Full PaaS or app deployment platform — the goal is safe change management, not replacing existing hosting stacks
- AI-first automation as a core product pillar — not needed to make the tool useful or differentiated
- General-purpose monitoring/observability product — monitoring may integrate later, but it is not the primary job of v1

## Context

This project is being created as a serious showcase piece for a job search, but it should be a real open-source tool rather than a portfolio artifact. The intended builder profile is a full-stack engineer comfortable with Python, TypeScript, Linux, open source, and automation, and the project should visibly demonstrate those strengths. Current market research suggests avoiding saturated categories such as generic AI developer tools, broad internal developer portals, and generic self-hosted PaaS products. The strongest wedge is a systems-heavy product with clear operational value for self-hosters.

## Constraints

- **Audience**: Self-hosters running Linux and Docker Compose — the first version must solve a sharp problem for a specific operator profile
- **Scope**: CLI plus public docs site first — prioritize a strong OSS core before adding a broad management UI
- **Differentiation**: Not another PaaS, monitoring dashboard, or AI wrapper — the product needs a clearer wedge than those crowded categories
- **Implementation**: Language is not preselected — choose implementation details based on operational fit rather than stack aesthetics
- **Showcase**: The project must be technically impressive and explainable in interviews — architecture, tradeoffs, and reliability design should be visible

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Focus v1 on self-hosters using Docker Compose on Linux | Gives the project a concrete wedge and strong systems relevance | — Pending |
| Make safe upgrades the core value | Upgrade risk is a concrete, understandable operational pain point | — Pending |
| Ship CLI plus landing/docs site before a richer dashboard | Keeps the MVP tight while still providing a public-facing product presence | — Pending |
| Keep AI out of the core product definition | Avoids relying on a crowded trend instead of real operational value | — Pending |

---
*Last updated: 2026-03-11 after initialization*
