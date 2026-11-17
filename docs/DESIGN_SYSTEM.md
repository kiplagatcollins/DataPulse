# GoDBClient — Design System

Define the visual language so UI work stays consistent.

## 1. Overview

- **Brand:** GoDBClient
- **Direction:** dark-first, data-dense, calm developer tool — the UI recedes so data and SQL stand out.
- **Source of truth:** `frontend/src/lib/theme.css` (CSS custom properties defined here; Svelte components consume tokens only).

## 2. Principles

1. **Data over chrome** — density is a feature: grids, trees, and editors dominate; decoration is minimal.
2. **Recognition over recall** — database objects are identified by icon + color badge (per engine type), never by remembering connection names.
3. **One primary action per screen** — dialogs have exactly one filled button (Connect, Save, Run).
4. **Feedback loops** — every long operation (connect, query, scan) shows a progress/spinner state and a cancel affordance (Doherty threshold).
5. **Progressive disclosure** — connection dialogs hide advanced options (TLS, timeouts, Redis DB number) behind an expandable section.
6. **Accessible by default** — contrast ≥ 4.5:1 for text, visible focus rings, full keyboard navigation for tree and grid.

## 3. Colors

Use semantic tokens, not raw hex values in components.

| Token | Value | Usage |
|---|---|---|
| `--background` | `#0f1117` | App background |
| `--foreground` | `#e6e8ee` | Primary text |
| `--card` | `#171a23` | Panels: sidebar, grid header, dialogs |
| `--primary` | `#3b82f6` | Primary actions, active tab, selection |
| `--primary-foreground` | `#ffffff` | Text on primary |
| `--secondary` | `#10b981` | Success states, connected indicators |
| `--muted` | `#1e2230` | Subtle surfaces, zebra rows, code background |
| `--muted-foreground` | `#8b91a5` | Secondary text, placeholders, timestamps |
| `--accent` | `#f59e0b` | Highlights, warnings pending confirmation |
| `--destructive` | `#ef4444` | Errors, destructive actions, guard warnings |
| `--border` | `#262b3a` | Borders and dividers |
| `--ring` | `#60a5fa` | Focus rings |

Per-engine accent badges (used on tree nodes and connection cards only):

| Engine | Token | Value |
|---|---|---|
| PostgreSQL | `--db-postgres` | `#336791` |
| Oracle | `--db-oracle` | `#c74634` |
| MySQL/MariaDB | `--db-mysql` | `#f29111` |
| SQLite | `--db-sqlite` | `#0f80b0` |
| Redis | `--db-redis` | `#dc382d` |

Dark mode: default and only theme for v1 (tokens above). A light theme may be added later behind `[data-theme]`; components must not hardcode hex values so that swap stays trivial.

## 4. Typography

- **Font family:** Inter (UI)
- **Monospace:** JetBrains Mono (SQL editor, result cells, connection strings, Redis REPL)
- **Scale:**
  - 12px — table metadata, tree sublabels, badges
  - 13px — default body: grid cells, tree items, form controls
  - 14px — section headings in panels
  - 16px — dialog titles
  - 20px — empty-state headings
- **Weights:** 400 body, 500 controls and tree items, 600 headings
- **Rules:** monospace cells use `font-variant-numeric: tabular-nums` for column alignment; SQL editor uses 13px mono with 1.6 line height; never more than the two families above.

## 5. Spacing

- Base grid: 4px
- Common values: 4, 8, 12, 16, 24, 32
- App layout: fixed top bar 40px; left tree pane default 260px (resizable 180–420px); result pane fills remainder; status bar 28px.
- Section rhythm: 12px between form rows in dialogs, 24px between dialog sections.

## 6. Radius

- `sm`: 4px — inputs, badges, grid cells with backgrounds
- `md`: 8px — buttons, dropdowns, tree nodes on hover
- `lg`: 12px — dialogs, cards
- `full`: 9999px — status dots, avatars, chips
- Use `md` for buttons and inputs; `lg` for dialogs and floating panels.

## 7. Components

- **Button:** variants `primary` (filled blue), `secondary` (muted), `ghost` (toolbar icons), `destructive` (red, for confirm-modals only); sizes `sm` (24px, toolbars) and `default` (32px).
- **Input / Select:** 32px height, `sm` radius, `--border` border, `--ring` 2px focus ring; error state uses `--destructive` border + helper text below, never color alone.
- **Tree (metadata explorer):** 24px rows, engine-colored type badge, chevron disclosure; keyboard: ↑/↓ move, →/← expand/collapse, Enter opens.
- **QueryGrid (results):** sticky header row, zebra striping with `--muted`, `tabular-nums` for numerics, right-aligned numbers, `null` rendered as italic muted `NULL`, blob cells as `⌘ base64…`; 40px rows max, virtualized above 200 rows.
- **CodeEditor (SQL):** mono, current-line highlight, error underline from guard diagnostics, gutter shows query index for multi-statement input.
- **Modal / Dialog:** 480px default (560px for connection editor), `--card` surface on dimmed backdrop, Esc closes unless a guard confirmation (then explicit choice required).
- **Tabs:** 36px editor tabs with per-tab connection badge; active tab underlined with `--primary`.
- **Status bar:** connection state dot (secondary = connected, destructive = error), row count, query duration, page position.
- **Toast:** bottom-right, auto-dismiss 4s, errors persist until dismissed; `role="status"`.
- **RedisPanel:** key list + detail split; type badges reuse engine badge styling; REPL uses CodeEditor styling in single-line mode.

## 8. Icons

- **Set:** Lucide (svelte-lucide)
- **Sizes:** 14px inline (tree), 16px controls, 20px empty states
- Engine badges may use official logos as SVG only in tree/connection contexts; do not mix with Lucide glyphs in the same node.

## 9. Breakpoints

This is a desktop app — there are no mobile breakpoints. Layout responds to window size:

| Name | Min width | Layout rule |
|---|---|---|
| `compact` | 800px | Tree pane collapses to icon rail |
| `default` | 1024px | Tree 260px + editor/grid |
| `wide` | 1440px | Optional split result/bottom panel |

Minimum supported window: 800×560. Panes are resizable by drag with persisted sizes.

## 10. Accessibility

- Visible focus ring (`--ring`, 2px) on every interactive element, including grid cells and tree rows.
- Labels for every input; placeholder text is a hint, never the label.
- Guard warnings are text + icon + color — never color-only; `role="alertdialog"` for destructive confirmations.
- Status messages (query done, rows fetched) use `role="status"`.
- Full keyboard path: connection switcher (Ctrl+P), run query (Ctrl+Enter), next tab (Ctrl+Tab), focus tree (Ctrl+1), focus editor (Ctrl+2).
- Respect `prefers-reduced-motion` — disable row hover transitions and spinner animations.
- Test contrast on `--muted` surfaces (worst case), not just on `--background`.

## 11. Rules

- Use design tokens instead of raw colors — `var(--token)` only in components.
- Reuse existing primitives before creating new ones.
- The app is dark-first; any future light theme must ship token-complete in the same PR.
- Do not add a new UI library, font, or icon set without approval.
- Grid/tree changes must preserve keyboard navigation — it is part of the component contract, not a nice-to-have.
