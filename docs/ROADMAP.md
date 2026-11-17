# DataPulse Roadmap & Gap Analysis

Gap analysis and implementation phases comparing DataPulse to DBeaver and [dbdelve](https://github.com/ShayanAbbas1/dbdelve) (reference). Updated after Phase 1 cheap-wins implementation.

## What we already have

- **Connections:** create/test/save/edit/duplicate/delete profiles for Postgres, MySQL, SQLite, Oracle; passwords in OS keychain (never in JSON); connect/disconnect/reconnect; active-connection tracking.
- **Explorer tree:** connection → schemas/databases → tables/views → columns (lazy load, refresh, status dots); context actions (View Data, Generate SELECT, Copy Table Name, New SQL Query, Edit/Duplicate/Delete/Reconnect/Disconnect).
- **Workspace:** multi-tab SQL and table-data tabs, tab persistence/restore (tabs, SQL text, active tab via `localStorage`), shortcuts `Ctrl/Cmd+Enter` (run selection or statement at cursor), `Ctrl+Tab`/`Ctrl+Shift+Tab`, `Ctrl+W`, `Ctrl+T`.
- **SQL editor:** CodeMirror 6 with SQL highlighting, schema-aware autocomplete (tables/columns/keywords, per-connection cache), run selection or statement-at-cursor.
- **Query execution:** shared per-driver executor, cancel active query by connection, timeout/paging (dialect-aware wrap), destructive-query guard with `Inspect` + UI confirm dialog, `ExecContext` path for non-result statements.
- **Results:** paginated grid, row numbers, sticky header, NULL/type formatting, duration/row counts/truncation indicator, client-side column sorting, CSV/JSON export.
- **Table view:** Data (limit/offset pagination, refresh) and Structure (columns, primary keys, foreign keys, indexes).
- **Backend/core:** Wails v3 bridge services, driver registry + `Catalog`/`KV` interfaces, connection pool, keychain-backed secrets, profile store with directory creation, Postgres schema browsing fix, SQL driver registration/`SQLExec` support, SQLite in-memory executor guard tests.
- **Redis (backend only):** driver + `RedisService` bridge (Scan/KeyInfo/Get/Exec) exists; no frontend UI.

## Phase 1 — Cheap wins (completed)

1. **Cancel running query** — shared executors per driver, mutexed cancel registry, active query tracked by connection, Stop button in SQL tab. (`7494543`)
2. **Guard confirm dialog** — `QueryService.Inspect` + `models.GuardDecision` binding, inline UI confirmation before forced re-run; backend guard remains authoritative. (`7494543`, `735fa3d`)
3. **Indexes & foreign keys in Structure** — `getForeignKeys`/`getIndexes` surfaced in table Structure view (catalog methods already existed). (`735fa3d`)
4. **Connection management UI** — edit/duplicate/delete/reconnect/disconnect from sidebar; modal edit mode preserves existing keychain secret when password left blank. (`5f5c591`)
5. **Tab navigation** — `Ctrl+Tab`/`Ctrl+Shift+Tab` cycle, `Ctrl+W` close, `Ctrl+T` new SQL tab. (`5f5c591`)
6. **Results sorting** — client-side header-click sort (asc/desc/none, numeric-aware, NULLs last). (`8e9ce4f`)

## Phase 2 — History, export, and grid UX (next)

**Goal:** faster repeat workflows and better result inspection without schema-changing backend work.

7. **Query history (localStorage, per profile/connection)** — append executed statements with timestamp/duration/status/row count; history panel or palette; click to open/re-run; cap/prune entries. (Approved plan: localStorage first.)
8. **Saved queries/snippets** — named queries, per-profile grouping/folders in sidebar, rename/delete, quick insert/run; persist locally first.
9. **Grid filter bar & copy modes** — per-column/client-side filters (contains/equals/is-null/regex-lite), **Copy Rows As** (text, CSV+header, JSON, INSERT SQL for selected rows).
10. **More export formats** — INSERT SQL and TSV export; preserve NULLs/types; consider streaming/chunked export for large pages later.
11. **Cell/row value viewer** — side/bottom panel for full cell value with JSON pretty-print/highlight, copy value, show type/NULL distinctly.
12. **Query formatter** — format current statement/buffer (e.g. `Ctrl+Shift+F`) via SQL formatter library; keep user SQL otherwise untouched.

## Phase 3 — Metadata depth and execution tooling

**Goal:** DBeaver-style object insight and safer power-user execution.

13. **DDL view / “Show CREATE TABLE”** — new per-dialect catalog method(s) for DDL generation/retrieval (PG `pg_get_*`, MySQL `SHOW CREATE`, SQLite `sqlite_master.sql`, Oracle DBMS_METADATA fallback); show in Structure/DDL tab with copy.
14. **EXPLAIN / query plan pane** — EXPLAIN (and ANALYZE where safe/opt-in) as a separate action; render plan rows/tree; never mutate the user buffer silently.
15. **Sidebar/object coverage** — group views separately from tables; surface routines/functions/procedures, sequences, triggers (driver/catalog extensions); filter-as-you-type, collapse/expand all.
16. **Multi-statement execution** — split statements safely, execute sequentially with per-statement result chips (rows/time/error), Stop/Continue controls. (Previously deferred; requires careful guard/confirm semantics per statement.)
17. **Transactions UI (later in phase)** — explicit commit/rollback controls and auto-commit indicator for drivers/sessions that support it; guard destructive ops accordingly.

## Phase 4 — Redis and key-value UX

**Goal:** first-class Redis workflow parity with existing SQL UX where applicable.

18. **Redis KV browsing UI** — connection-mode UI using existing `RedisService`: pattern scan (cursor-based, never blocking `KEYS *`), key list with type/TTL/size, value viewer by type (string/list/set/hash/zset), pagination/chunked fetch.
19. **Redis command REPL** — command input with history, response rendering, safe defaults and confirmations for destructive commands.
20. **Key mutations (guarded)** — edit/delete/TTL actions with explicit confirmation and preview where feasible.

## Phase 5 — Data editing and import/export

**Goal:** controlled write workflows (deferred/high-risk).

21. **Inline/table data editing (deferred)** — edit cells only when rows are identifiable (PK); generate reviewed `UPDATE`/`INSERT`/`DELETE` statements, explicit apply/cancel; never implicit writes.
22. **Row insert/delete UX** — single-row operations with preview and transaction-aware execution where supported.
23. **Import data** — CSV/TSV/JSON import with column mapping and preview; execute reviewed statements/batch inserts.
24. **Advanced export** — streaming export beyond page limits, optional “export all rows” re-fetch with warnings/limits.

## Phase 6 — Platform, connectivity, and power features (large/later)

25. **ER diagrams** — schema visualization (FK graph), zoom/pan/find, export (SVG/PNG); persisted layout per schema/profile.
26. **SSH/SSM tunnels & TLS options** — profile-level tunnel/TLS configuration, secret handling, readiness/timeout UX (Phase 2 in PRD non-goals for v1).
27. **Connection modes & safety policy** — read-only/read-write/full modes, optional server-side read-only enforcement, stricter guard policies.
28. **Command palette** — scoped actions/objects/history/connections with keyboard-first navigation.
29. **Theme system** — light mode, additional themes, system appearance sync, accessible contrast tokens.
30. **Cross-tool import** — import connections from DBeaver (and others) including safe handling of secrets/settings.
31. **Plugins/extensions** — out-of-process or scripted extensions (long-term; design carefully).

## Backend/quality backlog (cross-cutting)

- **MySQL:** fix `ForeignKeys` catalog query (`information_schema.constraint_column_usage` → `key_column_usage` as planned) and add coverage.
- **SQLite:** quote identifiers safely in pragma/catalog paths; add SQLite integration catalog round-trip test.
- **Oracle:** verify placeholder style/privileges and pagination edge cases for supported versions.
- **Paging semantics:** decide long-term approach — current executor wraps user SQL for pagination vs. dbdelve’s “never rewrite, cap results as they arrive” model; document tradeoffs and handle multi-statement/DDL cases explicitly.
- **svelte-check / lint:** clean pre-existing frontend diagnostics/warnings and add CI checks (`go test ./...`, `go vet ./...`, lint, frontend check/build).
- **Error/typing polish:** reduce `any` usage, improve driver/profile typing, structured bridge errors for UI branching.
- **Performance:** virtualized grid for large result sets, incremental schema/column loading for very large databases.

## dbdelve specifics (not in scope now, but notable)

- Never rewrites user SQL for paging (caps rows at paint/receive time) — differs from our current wrapping approach.
- Per-profile password command / short-lived credential sourcing.
- Query plan misestimate flagging and richer explain analysis.
- Out-of-process plugin system and broader engine/ODBC reach.

## Recommended sequencing

1. **Phase 2** (history + saved queries + grid filter/copy/export/value viewer) — highest daily-use leverage.
2. **Phase 3.13–15** (DDL, EXPLAIN, richer sidebar objects) — admin/inspection depth.
3. **Phase 4** (Redis UI) — unlock existing backend capability.
4. **Phase 3.16 + Phase 5** — multi-statement and guarded editing/import (requires careful safety design).
5. **Phase 6** — connectivity, diagrams, palette/themes, plugins.
