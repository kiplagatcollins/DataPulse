# Gap Analysis — dbdelve (reference) + DBeaver vs. DataPulse

As it stands today. Status per item verified against the codebase (commit `5f5c591`).
Cross-reference: [`ROADMAP.md`](ROADMAP.md) for phased implementation detail.

## What we already have

- **Connections** — create/test/save profiles (Postgres, MySQL, SQLite, Oracle); passwords in OS keychain (`internal/config/keychain.go`); connect/disconnect/reconnect.
- **Explorer tree** — connection → schema/databases → tables/views → columns; lazy load, refresh, status dots, context actions (`frontend/src/lib/components/SidebarTree.svelte`).
- **Tabs & workspace restore** — SQL and table-data tabs; persistence/restore via `localStorage` (`lib/workspace.ts`).
- **SQL editor** — CodeMirror 6, schema-aware completion (per-connection schema cache), `Ctrl+Enter` runs selection or statement-at-cursor (`SqlEditorTab.svelte`).
- **Query execution** — shared per-driver executor, cancel-by-connection, dialect-aware paging wrap, destructive-query guard (`internal/query/guard.go`, `Executor` in `internal/query/query.go`).
- **Results grid** — paginated, row numbers, NULL/type formatting, duration/truncation indicators, client-side sorting, CSV/JSON export (`DataGrid.svelte`).
- **Structure view** — columns, primary keys, foreign keys, indexes (`TableDataTab.svelte`; `MetadataService.Indexes/ForeignKeys`).
- **Guard (backend, authoritative)** — `ErrNeedsConfirm`, `QueryService.Inspect(sql)`, `models.GuardDecision`; inline confirm banner in UI before forced re-run.
- **Redis backend** — driver + `RedisService` bridge (`internal/bridge/services.go:245`); no frontend yet.

## Missing — cheap wins ✅ COMPLETED (was the one-stage recommendation)

1. ~~**Cancel running query**~~ — **done.** Shared executors per driver + mutexed cancel registry; Stop button in SQL tab. (`7494543`) *(the "per-call executor" defect is fixed)*
2. ~~**Guard confirm UI**~~ — **done.** `Inspect` + `GuardDecision` binding; inline confirmation banner (not a modal — open polish item). (`7494543`, `735fa3d`)
3. ~~**Indexes & FKs in tree/Structure**~~ — **done.** Catalog methods existed; surfaced in Structure view. (`735fa3d`)
4. ~~**Connection edit/delete/duplicate/reconnect UI**~~ — **done.** Sidebar context menu + modal edit mode; keychain secret preserved when password left blank. (`5f5c591`)
5. ~~**Tab navigation**~~ — **done.** `Ctrl/Cmd+Tab`(+Shift) cycle, `Ctrl+W` close, `Ctrl+T` new tab. (`5f5c591`)
6. ~~**Results sorting**~~ — **done.** Client-side header-click sort, numeric-aware, NULLs last; server `ORDER BY` still future. (`8e9ce4f`)

## Missing — medium effort (Phase 2 / 3)

7. **Query history** — localStorage first, per profile/connection; history panel/palette + re-run; cap/prune. *Not started — next up (ROADMAP Phase 2, item 7).*
8. **Saved queries** — named, per-profile folders in sidebar, rename/delete, quick insert/run. *Not started (Phase 2, item 8).*
9. **DDL view / Show CREATE TABLE** — needs new per-dialect catalog method (PG `pg_get_*`, MySQL `SHOW CREATE`, SQLite `sqlite_master.sql`, Oracle `DBMS_METADATA`). *Not started (Phase 3, item 13).*
10. **EXPLAIN pane** — prefix query with EXPLAIN (ANALYZE opt-in); render rows (parsed tree later). *Not started (Phase 3, item 14).*
11. **Cell/row value viewer** — full-value panel, JSON pretty-print/highlight, copy, type/NULL display; currently only `title` tooltip + CSS truncation. *Not started (Phase 2, item 11).*
12. **Grid filter bar + Copy Rows As** — per-column client-side filters; text / CSV+header / JSON / INSERT SQL for selected rows. *Not started (Phase 2, item 9).*
13. **More export formats** — INSERT SQL, TSV; streaming export for large sets. *Not started (Phase 2, item 10; Phase 5, item 24 for streaming).*
14. **Multi-statement execution** — split safely, sequential execution, per-statement result chips, Stop/Continue. *Not started (Phase 3, item 16; explicitly deferred).*
15. **Query formatter** — `Ctrl+Shift+F` via SQL formatter library. *Not started (Phase 2, item 12).*
16. **Redis KV browsing UI** — bridge exists (`RedisService`), zero references in `frontend/src`. *Not started (Phase 4, items 18–20).*

## Missing — large / later (Phases 5–6)

17. **Inline data editing** — edit cells → reviewed `UPDATE`/`DELETE` by PK; never implicit writes. *Deferred (Phase 5, item 21).*
18. **Platform features** — ER diagrams, CSV import, SSH/SSM tunnels, connection modes (read-only default), transactions UI (commit/rollback), command palette, theme system/light mode, views/procedures grouping, import connections from DBeaver. *All future (Phase 6, items 23–31).*

## dbdelve specifics DBeaver doesn't even have

- **Never rewriting user SQL for paging** — dbdelve caps rows client-side; **ours wraps SQL** (`pagingSQL`, `internal/query/query.go:137`). Open design decision; documented in ROADMAP backlog. *To revisit.*
- **Per-profile password command** — absent (`models.Profile` has no such field).
- **Plan misestimate flagging** — absent.
- **Plugin system** — absent (long-term).

## Known quality gaps (verified)

- MySQL `ForeignKeys` catalog query is buggy — uses `information_schema.constraint_column_usage`, which doesn't exist in MySQL; must use `key_column_usage` (ROADMAP backlog).
- Guard confirmation is an inline banner, not a modal dialog.
- `QueryGrid.svelte` is legacy/unused — candidate for deletion.
- No `svelte-check`/CI hygiene pass yet (pre-existing frontend diagnostics).

## Recommendation

The six cheap wins (1–6) are **done** — treat the one-stage recommendation as complete. Next: **Phase 2 in one push** — query history (7) + saved queries (8) + grid polish (filter bar, Copy Rows As, export formats — 12–13), then value viewer (11) and formatter (15) as filler. That matches the original sequencing: daily-use leverage first, then inspection depth (DDL/EXPLAIN), then Redis UI, then guarded editing.
