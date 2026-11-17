# DataPulse — Agent Instructions

Guidelines for AI coding agents working on this project.

## Before You Start

1. Read `PRD.md` for product goals and non-goals (Phase 1 scope is firm).
2. Read `DESIGN_SYSTEM.md` for UI rules (dark-first tokens, component contracts).
3. Read `ARCHITECTURE.md` for structure and boundaries (core never imports Wails).
4. Inspect the folder structure and existing packages before creating files.
5. Reuse existing patterns and libraries; match the surrounding Go/Svelte style.

## Tech Stack

- **Backend:** Go 1.27 (strict, idiomatic), Wails v3.0.0 (`github.com/wailsapp/wails/v3`)
- **Frontend:** Svelte + Vite + TypeScript, `@wailsio/runtime`
- **Drivers:** `jackc/pgx/v5` (stdlib adapter), `go-sql-driver/mysql`, `modernc.org/sqlite`, `sijms/go-ora/v2` (pure Go, no Instant Client), `redis/go-redis/v9`
- **Secrets:** `zalando/go-keyring` — OS keychain only
- **Package manager:** Go modules (backend), bun/npm for `frontend/`
- **Test runner:** Go standard library, table-driven tests

## Commands

```bash
wails3 dev                     # dev server, hot reload (backend + frontend)
wails3 generate bindings       # REQUIRED after changing any bridge service signature
go test ./...                  # backend unit + integration tests
go vet ./...                   # static analysis (run before finishing)
golangci-lint run              # lint (once configured in M1)
wails3 build                   # production binary
```

## General Rules

- Make small, verifiable changes. Read before editing.
- Reuse existing components, helpers, and patterns before adding new ones.
- Keep code modular; core packages (`driver`, `connpool`, `config`, `metadata`, `query`) must stay UI-agnostic — **only `internal/bridge` may import Wails**.
- Do not create unnecessary files or duplicate logic; no stub packages.
- Do not add dependencies unless the current stack cannot solve the problem; ask before adding any.
- Keep business logic out of Svelte components; components call bindings and render state.

## Code Guidelines (Go)

- Accept `context.Context` as the first parameter on anything doing I/O; propagate, never `context.Background()` inside libraries.
- Wrap errors with `%w` and add action-oriented context ("connect postgres: …"); define sentinel errors for conditions the UI branches on (`ErrNeedsConfirm`, `ErrNotConnected`, `ErrTimeout`).
- Never panic in library code; return errors. Reserve panics for programmer bugs in tests.
- Use `database/sql` for SQL engines; dialect differences live inside the driver package, behind the `Driver`/`Catalog`/`KV` interfaces in `internal/driver/driver.go`.
- Table-driven tests; test files beside the code they cover (`*_test.go`).
- Exported symbols have doc comments; keep functions small; no unused parameters.

## Code Guidelines (Svelte/TS)

- Components consume tokens from `theme.css`; no raw hex values.
- All Go calls go through generated bindings via `lib/api.ts` wrappers — never call bindings inline from components.
- Types are strict; no `any` unless narrowly scoped and justified.
- Follow the component contracts in `DESIGN_SYSTEM.md` (keyboard navigation, focus rings are requirements).

## Boundaries

- `internal/driver/` — dialect knowledge, interfaces, registry; no Wails, no config imports.
- `internal/connpool/` — connection lifecycle only; no SQL parsing.
- `internal/config/` — profile persistence + keychain; passwords never in JSON.
- `internal/query/` — guard, paging, cancel registry; no Wails.
- `internal/bridge/` — thin Wails service structs; translation only, no business logic.
- `frontend/` — presentation + UI state; no direct database access (impossible by design — keep it that way).
- Do not cross layers: no driver-specific types outside the driver package; no SQL strings in bridge or frontend.

## Testing

- Add or update tests for every behavior change; cover error paths, not only happy paths.
- Guard tests must prove: every no-WHERE `UPDATE`/`DELETE` is blocked without `Confirm`, and passes with it.
- SQLite integration tests cover the full catalog/executor round-trip; use temp files, no daemons.
- Prefer table-driven tests when several inputs share one behavior.
- Before finishing, run: `go test ./...`, `go vet ./...`, and `wails3 generate bindings` if service signatures changed.

## Security

- Never commit or log passwords, DSNs with credentials, or keychain contents; DSNs are built in-memory and scrubbed before error wrapping.
- Secrets live only in the OS keychain; the profile JSON stores a reference, never a secret. Never add a plaintext fallback without an approved design change.
- Use parameterized queries exclusively; user-supplied identifiers (table/column names in catalog queries) are quoted with the driver's identifier-quoting helper, never string-concatenated raw.
- The destructive-query guard is a product requirement: never weaken it to make a test pass; extend tests and guard together.
- Validate all inputs crossing the bridge (page size clamps, enum-only driver types, length limits).
- Implement explicit error handling; do not swallow errors.

## Do Not Edit

- `frontend/bindings/` — generated by `wails3 generate bindings`
- `build/`, `node_modules/`, `frontend/dist/` — build output
- Lockfiles unless dependencies change
- `Taskfile.yml` structure unless the build actually breaks

## Definition of Done

- Behavior is implemented and covered by tests; `go test ./...` and `go vet ./...` pass.
- Bindings regenerated if any bridge signature changed.
- UI matches `DESIGN_SYSTEM.md` (tokens, keyboard access, feedback states).
- No new warnings, dead code, or unrelated edits.
- Docs updated when behavior, commands, or structure changes.

## Review Checklist

- [ ] Correctness and error handling (context propagated, errors wrapped)
- [ ] Layer boundaries respected (no Wails outside `bridge/`)
- [ ] No secrets in code, logs, or JSON
- [ ] Guard not weakened; tests cover the change
- [ ] Formatting, vet, and lint pass
- [ ] No unrelated changes

## Need Help?

- Check `ARCHITECTURE.md` section 12 (extension points) before proposing new packages.
- Follow established patterns in sibling driver packages — copy `internal/sqlite/` as the reference when adding behavior.
- Ask before large architectural changes or new dependencies.
