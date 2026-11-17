<script lang="ts">
  import type { Result } from "../../../bindings/github.com/collins/GoDBClient/internal/models/models";

  let { result }: { result: Result | null } = $props();

  function exportCsv() {
    if (!result) return;
    const cols = result.columns ?? [];
    const rows = result.rows ?? [];
    const esc = (v: unknown) => `"${String(v ?? "").replace(/"/g, '""')}"`;
    const lines = [cols.map(esc).join(","), ...rows.map((r) => (r ?? []).map(esc).join(","))];
    download(lines.join("\n"), "text/csv", "export.csv");
  }

  function exportJson() {
    if (!result) return;
    download(JSON.stringify({ columns: result.columns, rows: result.rows }, null, 2), "application/json", "export.json");
  }

  function download(content: string, type: string, name: string) {
    const blob = new Blob([content], { type });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = name;
    a.click();
    URL.revokeObjectURL(a.href);
  }

  function fmt(v: unknown): string {
    if (v === null || v === undefined) return "";
    if (typeof v === "object") return JSON.stringify(v);
    return String(v);
  }
</script>

<div class="grid-wrap">
  <div class="stats">
    <div class="stat"><span class="k">Rows</span><span class="v">{result?.rowCount ?? 0}</span></div>
    <div class="stat"><span class="k">Time</span><span class="v">{result?.durationMs ?? 0} ms</span></div>
    {#if result?.truncated}
      <div class="stat warn">truncated</div>
    {/if}
    <div class="spacer"></div>
    <button class="btn" onclick={exportCsv} disabled={!result || !(result.columns?.length)}>Export CSV</button>
    <button class="btn" onclick={exportJson} disabled={!result}>Export JSON</button>
  </div>

  <div class="scroll">
    {#if result && (result.columns?.length ?? 0) > 0}
      <table>
        <thead>
          <tr>
            <th class="rownum">#</th>
            {#each result.columns ?? [] as col, i (i)}
              <th title={col}>{col}</th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each result.rows ?? [] as row, ri (ri)}
            <tr>
              <td class="rownum">{ri + 1}</td>
              {#each (row ?? []) as cell, ci (ci)}
                <td title={fmt(cell)}>
                  {#if cell === null || cell === undefined}
                    <span class="null">NULL</span>
                  {:else if typeof cell === "number"}
                    <span class="num">{String(cell)}</span>
                  {:else if typeof cell === "boolean"}
                    <span class="bool" class:true={cell}>{String(cell)}</span>
                  {:else}
                    {fmt(cell)}
                  {/if}
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    {:else if result}
      <div class="empty-state">
        <div class="empty-title">Statement executed</div>
        <div class="empty-sub">{result.rowCount ?? 0} rows returned in {result.durationMs ?? 0} ms</div>
      </div>
    {:else}
      <div class="empty-state">
        <div class="empty-title">No results</div>
        <div class="empty-sub">Run a query or open a table to view data.</div>
      </div>
    {/if}
  </div>
</div>

<style>
  .grid-wrap { display: flex; flex-direction: column; height: 100%; min-height: 0; background: var(--bg); }
  .stats { height: 34px; display: flex; align-items: center; gap: 10px; padding: 0 10px; border-bottom: 1px solid var(--border); background: var(--panel); color: var(--muted-fg); font-size: 12px; }
  .stat { display: inline-flex; align-items: center; gap: 6px; }
  .k { color: var(--muted-fg); }
  .v { color: var(--fg); font-family: var(--mono); }
  .stat.warn { color: #fbbf24; background: rgba(245,158,11,.12); border: 1px solid rgba(245,158,11,.25); padding: 1px 6px; border-radius: 4px; font-size: 11px; }
  .spacer { flex: 1; }
  .btn { height: 24px; padding: 0 8px; background: var(--panel-2); color: var(--fg); border: 1px solid var(--border); border-radius: 5px; font-size: 11.5px; }
  .btn:hover { background: var(--hover); }
  .scroll { flex: 1; overflow: auto; min-height: 0; }
  table { width: max-content; min-width: 100%; border-collapse: separate; border-spacing: 0; font-size: 12.5px; font-family: var(--mono); }
  th, td { padding: 6px 10px; border-bottom: 1px solid var(--border-subtle); border-right: 1px solid var(--border-subtle); text-align: left; white-space: nowrap; max-width: 420px; overflow: hidden; text-overflow: ellipsis; }
  th { position: sticky; top: 0; z-index: 2; background: #1a202b; color: #c9d1d9; font-family: var(--sans); font-size: 11.5px; font-weight: 600; border-bottom: 1px solid var(--border); }
  td { color: var(--fg); }
  .rownum { width: 46px; min-width: 46px; max-width: 46px; text-align: right; color: var(--muted-fg); background: var(--bg-elev); position: sticky; left: 0; z-index: 1; border-right: 1px solid var(--border); }
  th.rownum { z-index: 3; background: #1a202b; }
  tbody tr:hover td { background: rgba(59,130,246,.06); }
  tbody tr:hover td.rownum { background: #18202c; }
  .null { display: inline-block; padding: 0 5px; border: 1px solid var(--border); border-radius: 4px; color: #6b7280; font-size: 10.5px; font-style: italic; font-family: var(--sans); }
  .num { color: #b5cea8; }
  .bool { color: #93c5fd; }
  .bool.true { color: #86efac; }
  .empty-state { height: 100%; min-height: 160px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px; color: var(--muted-fg); }
  .empty-title { color: var(--fg); font-weight: 500; font-size: 13px; }
  .empty-sub { font-size: 12px; }
</style>
