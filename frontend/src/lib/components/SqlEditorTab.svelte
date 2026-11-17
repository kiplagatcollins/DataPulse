<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { EditorState } from "@codemirror/state";
  import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from "@codemirror/view";
  import { sql } from "@codemirror/lang-sql";
  import { oneDark } from "@codemirror/theme-one-dark";
  import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
  import { autocompletion, completionKeymap } from "@codemirror/autocomplete";
  import DataGrid from "./DataGrid.svelte";
  import { executeSql } from "../api";
  import type { Result } from "../../../bindings/github.com/collins/GoDBClient/internal/models/models";

  let { connId, initialSql = "SELECT 1;" }: { connId: string; initialSql?: string } = $props();

  let host: HTMLDivElement;
  let view: EditorView | null = null;
  let result = $state<Result | null>(null);
  let error = $state<string | null>(null);
  let running = $state(false);

  async function run() {
    if (!view) return;
    if (!connId) { error = "No connection selected. Connect to a database from the explorer."; return; }
    const sel = view.state.sliceDoc(view.state.selection.main.from, view.state.selection.main.to).trim();
    const q = sel || view.state.doc.toString();
    running = true; error = null;
    try { result = await executeSql(connId, q, 1, 500); }
    catch (e) { error = e instanceof Error ? e.message : String(e); result = null; }
    finally { running = false; }
  }

  onMount(() => {
    const runKey = keymap.of([{ key: "Mod-Enter", run: () => { void run(); return true; } }]);
    const state = EditorState.create({
      doc: initialSql,
      extensions: [
        lineNumbers(),
        highlightActiveLine(),
        highlightActiveLineGutter(),
        history(),
        keymap.of([...defaultKeymap, ...historyKeymap, ...completionKeymap]),
        runKey,
        autocompletion(),
        sql(),
        oneDark,
        EditorView.lineWrapping,
        EditorView.theme({ "&": { height: "100%" }, ".cm-scroller": { fontFamily: "var(--mono)" } }),
      ],
    });
    view = new EditorView({ state, parent: host });
  });
  onDestroy(() => view?.destroy());
</script>

<div class="sql-tab">
  <div class="toolbar">
    <button class="run-btn" onclick={run} disabled={running}><span class="run-icon">▶</span>{running ? "Running" : "Run"}</button>
    <span class="hint">Ctrl/Cmd+Enter</span>
    <div class="spacer"></div>
    <span class="ctx" class:ok={!!connId}>{connId ? "Connection ready" : "No connection selected"}</span>
  </div>
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
  .run-btn { height: 28px; display: inline-flex; align-items: center; gap: 6px; padding: 0 12px; border-radius: 6px; background: var(--accent); color: #fff; border: 1px solid var(--accent); font-size: 12.5px; font-weight: 500; }
  .run-btn:hover { background: var(--accent-hover); }
  .run-icon { font-size: 10px; }
  .hint { color: var(--muted-fg); font-size: 12px; }
  .spacer { flex: 1; }
  .ctx { font-size: 11.5px; color: var(--muted-fg); padding: 3px 8px; border: 1px solid var(--border); border-radius: 999px; background: var(--panel-2); }
  .ctx.ok { color: #86efac; border-color: rgba(34,197,94,.25); background: rgba(34,197,94,.08); }
  .split { flex: 1; min-height: 0; display: grid; grid-template-rows: minmax(180px, 42%) 1fr; }
  .editor-pane, .results-pane { min-height: 0; display: flex; flex-direction: column; }
  .editor-pane { border-bottom: 1px solid var(--border); }
  .pane-label { height: 26px; display: flex; align-items: center; padding: 0 10px; font-size: 11px; font-weight: 600; letter-spacing: .04em; text-transform: uppercase; color: var(--muted-fg); background: var(--bg-elev); border-bottom: 1px solid var(--border-subtle); }
  .editor { flex: 1; min-height: 0; overflow: auto; }
  .editor :global(.cm-editor) { height: 100%; font-size: 13px; }
  .results-body { flex: 1; min-height: 0; overflow: hidden; }
  .error { padding: 8px 10px; color: #fca5a5; background: rgba(239,68,68,.08); border-bottom: 1px solid rgba(239,68,68,.2); font-size: 12px; white-space: pre-wrap; }
</style>
