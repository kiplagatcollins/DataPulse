# AI Agents & Skills — Curated Field Guide

> Summarized from video/image content covering Claude Code workspaces, agentic AI, design/UX skills, and production engineering. Sources include [addyosmani/agent-skills](https://github.com/addyosmani/agent-skills) and trending skill repos.

---

## 1. Core Project Documentation Templates (addyosmani/agent-skills)

A set of four docs that give any AI coding agent (Claude Code, Cursor, Copilot) full project context:

### AGENTS.md — Agent Instructions
- **Purpose:** Rules and context for AI agents to produce consistent, high-quality code.
- **Before you start:** Read `PRD.md`, `DESIGN_SYSTEM.md`, `ARCHITECTURE.md`; check existing components and folder structure.
- **Rules:** TypeScript + Next.js 14+, follow the design system, reuse components, keep code modular, don't create unnecessary files.
- **Code:** Functional components, meaningful names, small components, accessible & semantic markup.
- **Security:** Never expose API keys; use env vars; validate all inputs; implement error handling.
- **Commands:** `npm install` / `npm run dev` / `npm run build` / `npm run test`.
- **When unsure:** Check `/docs`, follow existing patterns, ask before big changes.

### PRD.md — Product Requirements
- **Example product:** AI Portfolio Builder — "Build a stunning portfolio in minutes with AI."
- **Problem:** Creators struggle to showcase work persuasively (time, code limits, design skills).
- **Goal:** Launch a professional portfolio in <10 minutes with AI assistance.
- **Users:** Creators, students, freelancers, small business owners.
- **Core features:** AI website generation, one-click publishing, responsive mobile-first layouts.
- **Success metrics:** 10K+ users in 6 months, >80% setup completion, satisfaction >4.5/5.

### DESIGN_SYSTEM.md — UI Rules
- **Colors:** Primary `#3B82F6` (blue), Secondary `#10B981` (green), Accent `#F59E0B` (amber), dark-theme neutrals.
- **Typography:** Inter — H1 36px bold / H2 28px semibold / H3 22px medium / Body 16px / Caption 12px.
- **Spacing:** 4px grid (4, 8, 12, 16, 24, 32, 48).
- **Radius:** sm 4px, md 8px, lg 12px, full 9999px.
- **Components:** Buttons, inputs, cards, badges, modals. **Icons:** Lucide React.
- **Breakpoints:** sm 640 / md 768 / lg 1024 / xl 1280.

### ARCHITECTURE.md — System Structure
- **Stack:** Next.js 14 + React + Tailwind (frontend); Node.js / Next API routes (backend); Postgres (Supabase/Neon) + Prisma (DB); NextAuth/Clerk (auth); S3/Cloudflare R2 (storage); Vercel (deploy).
- **Structure:** `src/{app, components, lib, hooks, types}`, `public/`, `docs/`.
- **Data flow:** Frontend → API endpoint → business logic → response → UI state update.
- **Scalability:** Modular features, caching, background jobs, monitoring/alerts, auto-scaling.

---

## 2. Claude Code Workspace Anatomy

### Project layout
```
my_project/
├── CLAUDE.md              # Project memory
├── .claude/
│   ├── settings.json      # Permissions, hooks, env
│   ├── settings.local.json
│   ├── commands/          # Slash commands (review.md, deploy.md, test-all.md, bootstrap.md)
│   └── skills/            # Auto-activated skills
│       └── code-review/   # SKILL.md + scripts/ + references/ + assets/
├── agents/                # Subagent definitions (*.yml)
├── plugins/               # Bundled distributable setups
├── .mcp.json              # MCP server config
├── src/ tests/ docs/ scripts/
└── package.json .env.example Dockerfile README.md
```

### Key extension types
| Type | Role |
|---|---|
| **Skills** | Auto-activate on task match (`SKILL.md` + scripts + references + assets) |
| **Hooks** | Lifecycle event scripts — deterministic, not AI |
| **MCP** | External tool connections (GitHub, JIRA/Linear, Slack, PostgreSQL, Playwright, Filesystem) |
| **Subagents** | Delegated parallel work; isolated context, results only |
| **Agent Teams** | Multi-agent coordination via message passing |
| **Plugins** | Bundles of skills/agents/hooks/commands; distributable via marketplace |

### Hook events
`PreToolUse` (block before execution) · `PostToolUse` (auto-lint after writes) · `SessionStart` (load content / save summaries) · `PreCommit` (secret detection) · `Notification` (Slack/webhook alerts).

### Context management
- 0–50%: work freely · 50–70%: monitor · 70–80%: run `/compact` · 80%+: `/clear` mandatory.

### settings.json (key parts)
```json
{
  "permissions": { "allow": ["..."], "deny": ["..."] },
  "hooks": {
    "PreToolUse":  [{ "matcher": "Bash",  "hooks": [{ "type": "command", "command": "check-safety.sh" }] }],
    "PostToolUse": [{ "matcher": "Write", "hooks": [{ "type": "command", "command": "npm run lint" }] }]
  },
  "env": { "MAX_THINKING_TOKENS": "10000", "CLAUDE_AUTOCOMPACT_PCT_OVERRIDE": "50" }
}
```

### .mcp.json (key parts)
```json
{
  "mcpServers": {
    "github":   { "type": "stdio", "command": "npx", "args": ["-y", "@anthropic/mcp-github"],
                  "env": { "GITHUB_TOKEN": "${GITHUB_TOKEN}" } },
    "postgres": { "type": "stdio", "command": "npx", "args": ["-y", "@anthropic/mcp-postgres"],
                  "env": { "DATABASE_URL": "${DATABASE_URL}" } }
  }
}
```

### CLAUDE.md template
```markdown
# Project: My App
## Tech Stack
- Next.js 14, TypeScript, Tailwind / Supabase auth & DB / Prisma, tRPC
## Conventions
- Tests before code; conventional commits; never commit to main; lint + typecheck before PR
## Architecture
- src/components (UI) · src/services (business logic) · src/utils (shared helpers)
## Security
- No secrets in code or logs; validate inputs; parameterized queries only
```

### Getting started
```
npm i -g @anthropic-ai/claude-code
cd your_project && claude
```
Then: create `CLAUDE.md` with commands → add slash commands in `.claude/commands/` → configure MCP in `.mcp.json` → add skills as workflows grow.

### Best practices
Iterative development (start small, test often) · clear skill documentation · modular skill design · secrets via env vars · regular testing & auditing.

---

## 3. The Five-Layer Agent Development Kit

One stack, five layers (CLAUDE.md + Skills + Hooks + Subagents + Plugins):

| Layer | What | Notes |
|---|---|---|
| **L1 Memory** | `CLAUDE.md` (global at `~/.claude/`, project at `.claude/`) | Conventions, structure, repo expectations |
| **L2 Knowledge** | `skills/` — `SKILL.md` + scripts + templates | Description-matched, auto-invoked; on-demand docs |
| **L3 Guardrails** | `hooks/` — PreToolUse, PostToolUse, SessionStart | Block `rm -rf`, validate before run, auto-lint; deterministic git hooks |
| **L4 Delegation** | `subagents/` — code-reviewer, test-runner, explorer | Own context window, return results only; keeps main context clean |
| **L5 Distribution** | `plugins/` — manifest + marketplace | Bundles skills/agents/hooks/commands; one install aligns the whole team |

> "Skills provide expertise. Hooks enforce quality. Subagents delegate work."

---

## 4. Build a Marketing Team with Claude Agents & Skills

Top node: **CEO `claude-code` (the operating system)** — 20 departments of slash-command skills:

1. **Brand** — voice, visual identity, marketing psychology, AI-pattern cleaner
2. **Content** — blog, newsletter, SEO/GEO, case study, guides, customer stories
3. **Social** — tweets, LinkedIn, reply guy, Reddit opportunities, exec voice
4. **Video** — clipping, scripting, storyboarding, launch video, webinars
5. **Creative & Design** — frontend design, templates, brand system
6. **PR** — press release, pitching, share-of-voice, sentiment
7. **Demand Gen & ABM** — lead qualification, UTM generator, outreach sequences, paid ads, events
8. **Product Marketing** — ICP playbooks, messaging, one-pagers, sales decks, battlecards, GTM
9. **Website** — landing pages, A/B tests, site audits, SEO tracker, CRO
10. **Partnerships** — briefings, announcements, roundups
11. **Marketing Ops & AI** — brainstorming, skill creator/scout, OKR tracker, reporting
12. **Finance** — financial models, budget allocation, forecasting, ROI tracking
13. **Sales Enablement** — decks, objection handling, email templates, deal playbooks
14. **Customer Success** — onboarding flows, health scores, upsell, churn reduction
15. **Analytics** — dashboards, KPI tracking, cohort analysis, insight finder
16. **Events & Webinars** — planning, promo, registration, recaps
17. **Email** — writing, sequence building, segmentation, deliverability
18. **Paid Media** — campaign setup, audience research, ROAS tracking
19. **Market Research** — trend scouting, competitor reports, surveys
20. **Legal & Compliance** — privacy, terms, cookie policy, contract review

---

## 5. Slash-Command Personas & Workflows

Mental models, formatting frameworks, and persona switches — include the command in a prompt to instantly adapt tone, depth, and output format:

- **Response styles:** `/PRO` (expert) · `/FOUNDER` (strategic builder) · `/SKEPTIC` (risk-finder) · `/TEACHER` (patient guide)
- **Business & strategy:** `/PITCH` · `/ROADMAP` · `/MARKET` · `/SCALE` · `/GOALS`
- **Deep thinking & decisions:** `/DEVIL` · `/LONGGAME` · `/WORSTCASE` · `/FIRSTPRINCIPLES`
- **Creative & content:** `/NAME` · `/OPENERS` · `/STORY` · `/METAPHOR`
- **Organization & format:** `/OUTLINE` · `/CHECKLIST` · `/TABLE` · `/MATRIX` · `/JSON`
- **Coding:** `/CODE` · `/DEBUG` · `/OPTIMIZE` · `/REFACTOR` · `/TESTIT`
- **Analysis & systems:** `/DECIDE` · `/ROOT` · `/DATA` · `/AUTOMATE` · `/IMPROVE`

---

## 6. The Complete Agentic AI Cheatsheet

> Agentic AI = autonomous systems that **perceive, reason, plan, and act** to achieve goals with minimal human intervention.

### 10-phase roadmap (100 steps)
| Phase | Goal | Key steps |
|---|---|---|
| 1. What is Agentic AI | Understand it | Definition, agent vs chatbot, agent types, limitations |
| 2. Use Cases | Find high-value ones | Content, support, data, sales, marketing, DevOps, research, finance |
| 3. Core Components | Know how agents work | LLM (brain), memory, tools, planning, autonomy, feedback loop, guardrails, observability |
| 4. Tools & Frameworks | Choose the right stack | LLM providers, agent frameworks, vector DBs, orchestration, hosting, monitoring |
| 5. Design Agent | Design effectively | Objective, role, constraints, tools, workflow, error handling, memory, triggers, prompts |
| 6. Build & Test | Build reliably | Connect LLM, add tools, implement logic, validate outputs, simulate scenarios |
| 7. Deploy | Ship to production | Containerize, API, auth/security, webhooks, cron, rollout strategy |
| 8. Monitor | Ensure health | Logs, analytics, error tracking, cost monitoring, alerts, health checks |
| 9. Optimize | Make smarter | Prompt tuning, memory, latency, cost, A/B tests, human-in-the-loop, versioning |
| 10. Scale | Grow confidently | Multi-agent systems, swarms, load balancing, distributed memory, governance, compliance |

### The agent loop
**Perceive** (get data) → **Reason** (understand) → **Plan** (decide) → **Act** (execute) — all orbiting a **GOAL**.

### Agent types
Simple reflex · Model-based · Goal-based · Utility-based · Learning · Multi-agent systems.

### Key capabilities
Memory & recall · Tool use · Reasoning & planning · Decision making · Learning & adapting · Collaboration.

### Agent architecture
```
USER ↔ AGENT ↔ ENVIRONMENT
        │
   MEMORY + TOOLS
```

### Common use case examples
AI research assistant · sales outreach agent · code review agent · support ticket resolver · meeting summary agent · content repurposing agent · data insight generator.

### Example agent prompt
```
You are a helpful AI Agent.
Your goal is to {goal}.
You have access to: {tool1}, {tool2}...
Follow this plan: understand the task → plan the steps → use tools if needed → provide final answer.
```

### Prompt starters
- Act as a {role} AI agent.
- Your goal is to {goal}.
- Use the following tools: {tools}.
- Follow this plan: {steps}.
- Respond in {format}.

### Development checklist
Define clear goal · choose right model · design workflow · add tools & memory · test & validate · handle errors · deploy securely · monitor & improve.

### Quick tips
Start with an MVP · use strong prompts · leverage tools · test edge cases · monitor costs · iterate constantly.

### Essential commands (example)
`/run` (run the agent) · `/test` (test the workflow) · `/logs` (view agent logs) · `/tools` (list available tools) · `/memory` (show memory) · `/reset` (reset agent state).

### Recommended stack
- **LLMs:** OpenAI GPT-4o, Claude, Gemini, Llama, Mistral, Cohere
- **Frameworks:** LangChain, LlamaIndex, AutoGen, CrewAI, Haystack, LangGraph
- **Vector DBs:** Pinecone, Weaviate, Qdrant, Milvus, Chroma, Supabase Vector
- **Memory:** Redis, Astra DB, MongoDB, PostgreSQL, Supabase, Mem0
- **Integrations:** Google Search, Wolfram Alpha, Notion, Zapier, Make
- **Hosting:** Vercel, Railway, Render, Fly.io, AWS, Google Cloud
- **Observability:** LangSmith, Arize Phoenix, Helicone, W&B, PromptLayer, Grafana

### Evaluation metrics
Task success rate · accuracy · response time · cost per task · user satisfaction · hallucination rate · error rate.

### Common mistakes
❌ Unclear goals · poor prompts · too many tools · no error handling · ignoring testing · no monitoring · overcomplicating.

### Best practices & level-up path
Start small, iterate fast; define clear goals; use the right tools; validate rigorously; monitor continuously; keep humans in the loop.
**Path:** Learn basics → Build agents → Deploy & monitor → Optimize → Scale & automate.

---

## 7. UX Laws Your AI Agent Needs

- **Cognitive load:** Hick's law · Miller's law · Tesler's law · Occam's razor
- **Interaction:** Fitts's law · Jakob's law · Doherty threshold · Feedback loops
- **Memory:** Recognition over recall · Serial position
- **Perception (Gestalt):** Proximity · Similarity · Common region · Visual hierarchy
- **Engagement:** Aesthetic–usability effect · Von Restorff · Peak–end rule · Goal-gradient · Progressive disclosure
- **Robustness:** Postel's law

---

## 8. Font Pairings (12, by mood)

| Mood | Headings | Body |
|---|---|---|
| Modern & Clean | Playfair Display | Montserrat |
| Bold & Minimal | Montserrat | Lato |
| Elegant & Classic | Bodoni Moda | Poppins |
| Creative & Stylish | Pacifico | Montserrat |
| Strong & Modern | Lora | Open Sans |
| Minimal & Stylish | Helvetica Neue | Source Sans Pro |
| Vintage & Classic | Abril Fatface | Raleway |
| Tech & Modern | Rajdhani | Roboto |
| Soft & Friendly | Quicksand | Nunito |
| Luxury & Premium | Cormorant Garamond | Lato |
| Edgy & Bold | Anton | Muli |
| Fresh & Modern | Fraunces | Inter |

**Quick tips:** keep to 2–3 fonts · create contrast (display vs body) · establish hierarchy · use different weights · stay consistent.

---

## 9. Free UI/UX Learning Resources

| # | Resource | URL |
|---|---|---|
| 1 | UX Basics Study Guide | nngroup.com/articles/ux-basics-study-guide |
| 2 | UX Research Cheat Sheet | nngroup.com/articles/ux-research-cheat-sheet |
| 3 | 10 Usability Heuristics | nngroup.com/articles/ten-usability-heuristics |
| 4 | UX Writing Study Guide | nngroup.com/articles/ux-writing-study-guide |
| 5 | Web UX Study Guide | nngroup.com/articles/web-ux-study-guide |
| 6 | Apple Human Interface Guidelines | developer.apple.com/design/human-interface-guidelines |
| 7 | Apple Design Resources | developer.apple.com/design |
| 8 | Intro to Web Accessibility | w3.org/WAI/fundamentals-accessibility-intro |
| 9 | Digital Accessibility Foundations (free course) | w3.org/WAI/courses/foundations-course |
| 10 | GOV.UK Design System | design-system.service.gov.uk |
| 11 | IBM Carbon Design System | carbondesignsystem.com |

---

## 10. Trending Agent Skill Repos

### Fastest-growing (October 2026 — +230,679 stars combined)
1. Matt Pocock Skills (+35.3K)
2. Ponytail (+32.6K)
3. Security Audit (+31.4K)
4. Impeccable (+24.9K)
5. Anatomy (+23.1K)
6. Diagram Design (+20.5K)
7. Graphify (+17.1K)
8. Superpowers (+16.5K)
9. Taste Skill (+15.2K)
10. Hallmark (+14.1K)

### Notable repositories
`obra/superpowers` · `ComposioHQ/awesome-claude-skills` · `anthropics/skills` · `multica-ai/andrej-karpathy-skills` · `mattpocock/skills` · `nextlevelbuilder/ui-ux-pro-max-skill` · `JuliusBrussee/caveman` · `addyosmani/agent-skills` · `Leonxlnx/taste-skill` · `ayghri/i-have-adhd`

### SEO skills
`claude-seo` · `geo-seo-claude` · `schema` · `programmatic-seo` · `claude-blog`

### Also trending this week
`security-audit-skill` (multi-phase security audits with verified, machine-readable findings) · `agent-browser` · `astryx` · `graphify` · `claude-mem` · `Claude-Code-Game-Studios` · `herdr` · `gstack` · `claude-quickstarts` · `last30days-skill` (mvanhorn) · **Code Modernization** (point Claude at a legacy codebase and get: an understanding of what it does, a plan you approve, the modernized code, and proof the new code behaves like the old — any language, any kind of move).

### Notable skill author / org
Subash Chandra — `cloudflare/security-audit-skill`.

### Top 5 design skills (avoid generic AI-looking websites)
`ui-ux-pro-max` · `hallmark` · `emil-design-eng` · `shadcn` · `web-design-guidelines`

### 15 skills every professional coder should install
1. Frontend Design
2. Web App Testing
3. MCP Builder
4. React Best Practices
5. Web Design Guidelines
6. Vercel Deployment
7. Systematic Debugging
8. Test-Driven Development
9. Code Verification
10. Code Review & Quality
11. Code Simplification
12. API & Interface Design
13. Security & Hardening
14. CI/CD Automation
15. Performance Optimization (measure first, find real bottlenecks)

### gstack — a full dev workflow in one install
- **Lead Planner** — plans the whole job before writing code
- **Code Cartographer (Graphify)** — maps your project into a knowledge graph
- **Release Manager (GitHub CLI)** — opens PRs and ships code
- **Launch Engineer (Vercel)** — puts a site live in one command
- **Test Pilot (Playwright)** — drives a real browser and tests what it sees
- **Second Reviewer (Codex)** — a real second opinion on code
- **Session Monitor (Claude HUD)** — session visibility
- Plus: Fix N+1 queries, add DB indexes, cache with Redis, lazy-load images, optimistic UI

---

## 11. The "Vibe Coder" Production Checklist (~100 things nobody tells you)

### Database
Indexes · N+1 queries · connection pooling · migrations & rollbacks · backups (and testing restores) · transactions · race conditions · row locking · unique constraints · foreign keys · soft deletes · pagination.

### Infrastructure & Ops
Cache invalidation · cold starts · background jobs · dead letter queues · timezones · load testing · timeouts · retries with backoff · idempotency · queues · cron jobs · webhooks (+ retries, signatures) · debouncing · autoscaling · CDN · image optimization · bundle size.

### Observability & Reliability
Error tracking · structured logs · uptime monitoring · alerts · health checks · error boundaries · friendly error pages · status page · rollbacks · feature flags · kill switches · circuit breakers · fallbacks · postmortems.

### Security
Input validation · SQL injection · XSS · CSRF · CORS · SSRF · password hashing · session expiry · token rotation · OAuth scopes · 2FA · least privilege.

### Dev Workflow
Code review · tests · evals · preview deploys · seed data · dependency updates · git branches · CI/CD · staging vs prod · versioning · changelogs · infra as code · DNS · SSL renewal.

### Product & Business
Stripe webhooks · failed payments · refunds · proration · sales tax/VAT · invoices · usage limits · account deletion · data export · privacy policy · cookie consent · analytics · accessibility · SEO basics · email SPF/DKIM · upload limits.

### AI-Specific
Prompt injection · token costs · spend caps · context limits · model fallbacks · streaming · provider rate limits · embeddings · vector search · caching AI answers · hallucination checks.

---

## Sources & Further Reading

- https://github.com/addyosmani/agent-skills
- https://github.com/anthropics/skills
- https://www.marktechpost.com/2026/11/13/vercel-ui-sdk-agent-ui-behaviors/ (Vercel UI SDK — agent UI behaviors)
- https://github.com/mvanhorn/last30days-skill

---

## Key Takeaways

1. **Document for your agent** — AGENTS.md / PRD / design system / architecture docs make AI output consistent and high-quality.
2. **Layer your agent stack** — Memory (CLAUDE.md) → Knowledge (skills) → Guardrails (hooks) → Delegation (subagents) → Distribution (plugins).
3. **Agents are loops, not chats** — Perceive → Reason → Plan → Act, with memory, tools, and guardrails.
4. **Design matters** — UX laws, font pairing, and design-system skills prevent generic AI-looking output.
5. **Production is a checklist** — 100+ concerns (DB, security, observability, payments, AI costs) separate demos from products.
6. **Great agents are built, tested, and improved iteratively.**
