# DataPulse — Product Requirements Document

Define what you are building, who it is for, and how you will know it works.

## 1. Product Overview

- **Product name:** DataPulse
- **Tagline:** A lightweight, native desktop client for all your databases.
- **Description:** DataPulse is a cross-platform desktop application for browsing, querying, and managing relational and key-value databases. It pairs a Go backend (Wails v3) with a Svelte UI, delivering DBeaver-style workflow — connection profiles, a metadata tree, and a SQL editor — in a small, fast, native-feeling app with no JVM or Electron runtime.

## 2. Problem

Existing universal database clients are heavyweight. DBeaver and its peers ship a JVM runtime, consume hundreds of MB of RAM, and take seconds to start, yet most developers only need: connect, browse schema, run a query, inspect results. JetBrains-style IDEs bolt database tools behind large subscriptions. Developers, analysts, and DBAs feel this cost every time they just want to "quickly check a table" — the tool overhead exceeds the task.

## 3. Goals

- **Primary outcome:** Connect to PostgreSQL, MySQL/MariaDB, SQLite, Oracle, and Redis and run useful queries within 10 seconds of first launch.
- **Secondary outcome:** Native desktop performance — cold start under 2s, result rendering of 100+ rows without visible lag, memory footprint under 150 MB idle.
- **Constraint that matters:** Local-first and private. Credentials live in the OS keychain; no telemetry; the app works fully offline.

## 4. Target Users

- **Primary user:** Backend/full-stack developers who juggle several databases daily and want a fast, keyboard-driven client.
- **Secondary user:** Data analysts and QA engineers who mostly browse tables and run read-only queries.
- **Optional buyer/admin:** Platform teams who want a safe (guard-railed) client they can standardize on.

## 5. Core Features

### Connection Manager

- Create, edit, duplicate, delete, and test connection profiles for PostgreSQL, MySQL/MariaDB, SQLite, Oracle, and Redis.
- Secrets (passwords) are stored only in the OS keychain (GNOME Keyring / macOS Keychain / Windows Credential Manager); the JSON profile file holds a reference, never a secret.
- Connection test uses a hard timeout (default 5s) and reports precise failure reasons (DNS, auth, TLS, timeout).

### Metadata Explorer

- Sidebar tree: connection → databases/schemas → tables & views → columns, foreign keys, indexes.
- Catalog data is fetched through a per-driver `Catalog` interface (information_schema for PG/MySQL, `sqlite_master`/PRAGMA for SQLite, Oracle data dictionary `ALL_*` views for Oracle, DB-scan for Redis), so the tree looks identical for every engine.
- Refresh-on-demand per node; no full-tree re-scan after DDL.

### SQL Query Editor & Executor

- Execute arbitrary SQL with server-side pagination (LIMIT/OFFSET dialect-aware, page size clamped 1–500, default 100).
- Rows return as JSON-safe structured data (`null`, numbers, text, timestamps as RFC3339, blobs as base64).
- Running queries are cancellable: every execution gets a query ID backed by a Go `context` cancel; results show duration and row counts.

### Safety Guard (destructive-query protection)

- `UPDATE`/`DELETE` without a `WHERE` clause is blocked with a structured warning; the user can acknowledge and re-run with explicit confirmation.
- `DROP`/`TRUNCATE`/`ALTER` require a softer confirmation.
- Connection timeouts, statement timeouts, and context cancellation are enforced at the driver layer, not the UI.

### Redis Key Browser + REPL

- Browse keys by pattern with cursor-based SCAN (no `KEYS *` on large datasets).
- Inspect a key's type, TTL, and length; view values formatted by type (string, list, set, hash, zset).
- Arbitrary command execution via a REPL pane.

## 6. Non-Goals (v1)

- SSH tunneling, TLS client certificates, and cloud SSO/IAM auth (Phase 2).
- Query history, saved snippets, export to CSV/JSON, ER diagrams, data editing grid (Phase 2+).
- NoSQL document UIs beyond Redis key browsing (MongoDB etc.).
- Plugin/marketplace system; multi-window workspaces.
- User accounts, sync, or any server component.

## 7. Constraints

- **Technical:** Go 1.26+ (Wails v3 requirement); pure-Go SQLite driver (`modernc.org/sqlite`) and pure-Go Oracle driver (`sijms/go-ora`) to keep CGO-free cross-compilation — no Oracle Instant Client required; `database/sql` is the canonical interface for SQL engines.
- **Privacy/security:** No secrets in config files, logs, or error messages; parameterized queries only; no network egress beyond the user's configured database hosts.
- **Platform:** Windows 10+/11, macOS 12+, Linux (GTK4/WebKitGTK 6.0 stack; legacy `-tags gtk3` for older distros). Desktop only in v1.

## 8. Success Metrics

- Time from install to first successful query: **< 10 minutes** including tool setup (**proposed**).
- Cold start to interactive UI: **< 2 s** on mid-range hardware (**proposed**).
- Idle RSS memory: **< 150 MB** (**proposed**).
- Destructive-query guard catch rate: 100% of no-WHERE `UPDATE`/`DELETE` blocked before confirmation (unit-tested) (**proposed**).

## 9. Current Status

- **Built:** Approved architecture and design documents (this set); no code yet.
- **In progress:** Milestone M1 — Wails v3 scaffold, module wiring, stub service + generated bindings.
- **Known limitations:** None yet; Wails v3 mobile support is out of scope for v1.

## 10. Open Questions

- Which keychain fallback (if any) for headless/Linux-minimal setups lacking a secret service?
- Oracle version floor: target 12c+ only (native `OFFSET/FETCH` pagination), or add ROWNUM subqueries for 11g?
- Should the query grid support inline cell editing in Phase 2, or jump straight to full row-edit forms?
- Confirm module path (`datapulse` vs GitHub-hosted path) before M1 scaffolding.
