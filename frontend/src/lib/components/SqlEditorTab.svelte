<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { Compartment, EditorState } from "@codemirror/state";
  import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from "@codemirror/view";
  import { sql } from "@codemirror/lang-sql";
  import { oneDark } from "@codemirror/theme-one-dark";
  import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
  import { autocompletion, completionKeymap } from "@codemirror/autocomplete";
  import DataGrid from "./DataGrid.svelte";
  import { executeSql, getDatabases, getTables, getColumns, inspectSql, cancelQuery } from "../api";
  import type { Result } from "../../../bindings/github.com/collins/GoDBClient/internal/models/models";

  let { connId, initialSql = "SELECT 1;", onSqlChange = (_: string) => {} }: { connId: string; initialSql?: string; onSqlChange?: (s: string) => void } = $props();

  let host: HTMLDivElement;
  let view: EditorView | null = null;
  let result = $state<Result | null>(null);
  let error = $state<string | null>(null);
  let running = $state(false);
  let pendingConfirm = $state<{ sql: string; action: string } | null>(null);

  const sqlCompartment = new Compartment();
  const schemaCache = new Map<string, Promise<any> | any>();
  let changeTimer: number | null = null;

  function statementAtCursor(doc: string, pos: number): string {
    const before = doc.lastIndexOf(";", pos - 1);
    const after = doc.indexOf(";", pos);
    const start = before + 1;
    const end = after === -1 ? doc.length : after;
    const s = doc.slice(start, end).trim();
    if (s) return s;
    const altStart = doc.lastIndexOf(";", Math.max(0, start - 2)) + 1;
    const alt = doc.slice(altStart, start - 1).trim();
    return alt || doc.trim().replace(/;$/, "").trim();
  }

  function currentQuery(): string {
    if (!view) return "";
    const sel = view.state.sliceDoc(view.state.selection.main.from, view.state.selection.main.to).trim();
    return sel || statementAtCursor(view.state.doc.toString(), view.state.selection.main.head);
  }

  async function run(force = false) {
    if (!view) return;
    if (!connId) { error = "No connection selected. Connect to a database from the explorer."; return; }
    const q = currentQuery();
    if (!q) { error = "No query to run."; return; }
    error = null;
    if (!force) {
      try {
        const guard = await inspectSql(q);
        if (guard?.needsConfirm) { pendingConfirm = { sql: q, action: guard.action || "statement" }; return; }
      } catch {}
    }
    pendingConfirm = null;
    running = true;
    try { result = await executeSql(connId, q, 1, 500, force); }
    catch (e) { error = e instanceof Error ? e.message : String(e); result = null; }
    finally { running = false; }
  }

  async function stop() {
    if (!connId) return;
    try { await cancelQuery(connId); } catch {}
  }

  async function loadSchema(id: string) {
    if (!id) return null;
    if (schemaCache.has(id)) return schemaCache.get(id);
    const p = (async () => {
      const schemas: Record<string, Record<string, string[]>> = {};
      const dbs = await getDatabases(id);
      let tableCount = 0;
      for (const db of dbs) {
        const tables = await getTables(id, db);
        const tableMap: Record<string, string[]> = {};
        for (const t of tables) {
          if (tableCount >= 300) break;
          tableCount++;
          try {
            const cols = await getColumns(id, db, t.name);
            tableMap[t.name] = cols.map((c) => c.name);
          } catch { tableMap[t.name] = []; }
        }
        schemas[db] = tableMap;
        if (tableCount >= 300) break;
      }
      return schemas;
    })();
    schemaCache.set(id, p);
    try { const v = await p; schemaCache.set(id, v); return v; } catch { schemaCache.delete(id); return null; }
  }

  async function applySchema(id: string) {
    if (!view) return;
    const schemas = await loadSchema(id);
    if (!view) return;
    view.dispatch({ effects: sqlCompartment.reconfigure(sql(schemas ? { schemas, upperCaseKeywords: true } : { upperCaseKeywords: true })) });
  }

  function scheduleSqlChange(text: string) {
    if (changeTimer) window.clearTimeout(changeTimer);
    changeTimer = window.setTimeout(() => onSqlChange(text), 400);
  }

  onMount(() => {
    const runKey = keymap.of([{ key: "Mod-Enter", run: () => { void run(false); return true; } }]);
    const state = EditorState.create({
      doc: initialSql,
      extensions: [
        lineNumbers(),
        highlightActiveLine(),
        highlightActiveLineGutter(),
        history(),
        runKey,
        keymap.of([...completionKeymap, ...defaultKeymap, ...historyKeymap]),
        autocompletion(),
        sqlCompartment.of(sql({ upperCaseKeywords: true })),
        oneDark,
        EditorView.lineWrapping,
        EditorView.updateListener.of((u) => { if (u.docChanged) scheduleSqlChange(u.state.doc.toString()); }),
        EditorView.theme({ "&": { height: "100%" }, ".cm-scroller": { fontFamily: "var(--mono)" } }),
      ],
    });
    view = new EditorView({ state, parent: host });
    void applySchema(connId);
  });

  $effect(() => { const id = connId; if (view) void applySchema(id); });

  onDestroy(() => { if (changeTimer) window.clearTimeout(changeTimer); view?.destroy(); });
</script>

<div class="sql-tab">
  <div class="toolbar">
    {#if running}
      <button class="stop-btn" onclick={stop}>■ Stop</button>
    {:else}
      <button class="run-btn" onclick={() => run(false)}><span class="run-icon">▶</span>Run</button>
    {/if}
    <span class="hint">Ctrl/Cmd+Enter runs selection or statement</span>
    <div class="spacer"></div>
    <span class="ctx" class:ok={!!connId}>{connId ? "Connection ready" : "No connection selected"}</span>
  </div>
  {#if pendingConfirm}
    <div class="confirm">
      <span>Destructive <b>{pendingConfirm.action}</b> statement requires confirmation. Run anyway?</span>
      <button class="confirm-run" onclick={() => run(true)}>Run</button>
      <button class="confirm-cancel" onclick={() => (pendingConfirm = null)}>Cancel</button>
    </div>
  {/if}
  <div class="split">
    <div class="editor-pane">
      <div class="pane-label">Query</div>
      <div class="editor" bind:this={host}></div>
    </div>
    <div class="results-pane">
      <div class="pane-label">Results</div>
      {#if error}<div class="error">{error}</div>{/if}
      <div class="results-body"><DataGrid {result} /></div>
    </div>
  </div>
</div>

<style>
  .sql-tab { display: flex; flex-direction: column; height: 100%; min-height: 0; background: var(--bg); }
  .toolbar { height: 40px; display: flex; align-items: center; gap: 10px; padding: 0 10px; border-bottom: 1px solid var(--border); background: var(--panel); }
  .run-btn, .stop-btn { height: 28px; display: inline-flex; align-items: center; gap: 6px; padding: 0 12px; border-radius: 6px; color: #fff; font-size: 12.5px; font-weight: 500; }
  .run-btn { background: var(--accent); border: 1px solid var(--accent); }
  .run-btn:hover { background: var(--accent-hover); }
  .stop-btn { background: #b91c1c; border: 1px solid #dc2626; }
  .stop-btn:hover { background: #dc2626; }
  .run-icon { font-size: 10px; }
  .hint { color: var(--muted-fg); font-size: 12px; }
  .spacer { flex: 1; }
  .ctx { font-size: 11.5px; color: var(--muted-fg); padding: 3px 8px; border: 1px solid var(--border); border-radius: 999px; background: var(--panel-2); }
  .ctx.ok { color: #86efac; border-color: rgba(34,197,94,.25); background: rgba(34,197,94,.08); }
  .confirm { display: flex; align-items: center; gap: 10px; padding: 8px 10px; font-size: 12.5px; color: #fde68a; background: rgba(245,158,11,.1); border-bottom: 1px solid rgba(245,158,11,.25); }
  .confirm-run, .confirm-cancel { height: 24px; padding: 0 10px; border-radius: 5px; font-size: 12px; }
  .confirm-run { background: #b45309; color: #fff; border: 1px solid #d97706; }
  .confirm-cancel { background: var(--panel-2); color: var(--fg); border: 1px solid var(--border); }
  .split { flex: 1; min-height: 0; display: grid; grid-template-rows: minmax(180px, 42%) 1fr; }
  .editor-pane, .results-pane { min-height: 0; display: flex; flex-direction: column; }
  .editor-pane { border-bottom: 1px solid var(--border); }
  .pane-label { height: 26px; display: flex; align-items: center; padding: 0 10px; font-size: 11px; font-weight: 600; letter-spacing: .04em; text-transform: uppercase; color: var(--muted-fg); background: var(--bg-elev); border-bottom: 1px solid var(--border-subtle); }
  .editor { flex: 1; min-height: 0; overflow: auto; }
  .editor :global(.cm-editor) { height: 100%; font-size: 13px; }
  .results-body { flex: 1; min-height: 0; overflow: hidden; }
  .error { padding: 8px 10px; color: #fca5a5; background: rgba(239,68,68,.08); border-bottom: 1px solid rgba(239,68,68,.2); font-size: 12px; white-space: pre-wrap; }
</style>
