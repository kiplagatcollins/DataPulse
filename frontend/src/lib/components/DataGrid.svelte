<script lang="ts">
  import type { Result } from "../../../bindings/github.com/collins/DataPulse/internal/models/models";

  let { result }: { result: Result | null } = $props();

  type Filter = { op: "contains" | "equals" | "isnull" | "notnull" | "regex"; value: string };
  let sortCol = $state<number | null>(null);
  let sortDir = $state<"asc" | "desc">("asc");
  let filters = $state<Record<number, Filter>>({});
  let showFilters = $state(false);
  let selected = $state<{ r: number; c: number } | null>(null);
  let copyMsg = $state("");

  $effect(() => { result; sortCol = null; sortDir = "asc"; filters = {}; selected = null; });

  function toggleSort(i: number) {
    if (sortCol !== i) { sortCol = i; sortDir = "asc"; }
    else if (sortDir === "asc") sortDir = "desc";
    else { sortCol = null; sortDir = "asc"; }
  }

  function cmp(a: unknown, b: unknown): number {
    if (a === null || a === undefined) return b === null || b === undefined ? 0 : 1;
    if (b === null || b === undefined) return -1;
    if (typeof a === "number" && typeof b === "number") return a - b;
    const na = Number(a), nb = Number(b);
    if (!Number.isNaN(na) && !Number.isNaN(nb) && String(a).trim() !== "" && String(b).trim() !== "") return na - nb;
    return String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: "base" });
  }

  function cellText(v: unknown): string {
    if (v === null || v === undefined) return "";
    if (typeof v === "object") return JSON.stringify(v);
    return String(v);
  }

  function matches(v: unknown, f: Filter): boolean {
    if (f.op === "isnull") return v === null || v === undefined;
    if (f.op === "notnull") return !(v === null || v === undefined);
    const t = cellText(v).toLowerCase();
    const q = (f.value ?? "").toLowerCase();
    if (f.op === "equals") return t === q;
    if (f.op === "regex") { try { return new RegExp(f.value || "", "i").test(cellText(v)); } catch { return false; } }
    return t.includes(q);
  }

  let baseRows = $derived.by(() => {
    const rows = result?.rows ?? [];
    const entries = Object.entries(filters).map(([k, f]) => [Number(k), f] as const).filter(([, f]) => f.op === "isnull" || f.op === "notnull" || (f.value ?? "") !== "");
    if (entries.length === 0) return rows;
    return rows.filter((row) => entries.every(([ci, f]) => matches(row?.[ci], f)));
  });

  let viewRows = $derived.by(() => {
    if (sortCol === null) return baseRows;
    const dir = sortDir === "asc" ? 1 : -1;
    return [...baseRows].sort((ra, rb) => dir * cmp(ra?.[sortCol!], rb?.[sortCol!]));
  });

  function download(content: string, type: string, name: string) {
    const blob = new Blob([content], { type });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = name;
    a.click();
    URL.revokeObjectURL(a.href);
  }

  function csvEsc(v: unknown) { return `"${cellText(v).replace(/"/g, '""')}"`; }
  function tsvEsc(v: unknown) { return cellText(v).replace(/\t/g, " ").replace(/\r?\n/g, " "); }
  function sqlLit(v: unknown): string {
    if (v === null || v === undefined) return "NULL";
    if (typeof v === "number" && Number.isFinite(v)) return String(v);
    if (typeof v === "boolean") return v ? "TRUE" : "FALSE";
    const s = typeof v === "object" ? JSON.stringify(v) : String(v);
    return `'${s.replace(/'/g, "''")}'`;
  }

  function exportCsv() {
    if (!result) return;
    const cols = result.columns ?? [];
    const lines = [cols.map(csvEsc).join(","), ...viewRows.map((r) => (r ?? []).map(csvEsc).join(","))];
    download(lines.join("\n"), "text/csv", "export.csv");
  }
  function exportTsv() {
    if (!result) return;
    const cols = result.columns ?? [];
    const lines = [cols.map(tsvEsc).join("\t"), ...viewRows.map((r) => (r ?? []).map(tsvEsc).join("\t"))];
    download(lines.join("\n"), "text/tab-separated-values", "export.tsv");
  }
  function exportJson() {
    if (!result) return;
    download(JSON.stringify({ columns: result.columns, rows: viewRows }, null, 2), "application/json", "export.json");
  }
  function exportInsert() {
    if (!result) return;
    const cols = result.columns ?? [];
    const tbl = "table_name";
    const lines = viewRows.map((r) => `INSERT INTO ${tbl} (${cols.map((c) => `"${c.replace(/"/g, '""')}"`).join(", ")}) VALUES (${(r ?? []).map(sqlLit).join(", ")});`);
    download(lines.join("\n"), "text/plain", "export.sql");
  }

  async function copyAs(kind: "text" | "csv" | "json" | "insert") {
    if (!result) return;
    const cols = result.columns ?? [];
    let text = "";
    if (kind === "text") text = viewRows.map((r) => (r ?? []).map(cellText).join("\t")).join("\n");
    else if (kind === "csv") text = [cols.map(csvEsc).join(","), ...viewRows.map((r) => (r ?? []).map(csvEsc).join(","))].join("\n");
    else if (kind === "json") text = JSON.stringify(viewRows, null, 2);
    else text = viewRows.map((r) => `INSERT INTO table_name (${cols.map((c) => `"${c.replace(/"/g, '""')}"`).join(", ")}) VALUES (${(r ?? []).map(sqlLit).join(", ")});`).join("\n");
    try { await navigator.clipboard.writeText(text); copyMsg = "Copied"; }
    catch { copyMsg = "Copy failed"; }
    setTimeout(() => (copyMsg = ""), 1200);
  }

  function fmt(v: unknown): string { return cellText(v); }
  function pretty(v: unknown): string {
    if (v === null || v === undefined) return "NULL";
    if (typeof v === "string") { try { return JSON.stringify(JSON.parse(v), null, 2); } catch { return v; } }
    if (typeof v === "object") return JSON.stringify(v, null, 2);
    return String(v);
  }
  function setFilter(i: number, patch: Partial<Filter>) {
    const cur = filters[i] ?? { op: "contains", value: "" };
    const next = { ...cur, ...patch } as Filter;
    if (next.op !== "isnull" && next.op !== "notnull" && !next.value) { const { [i]: _, ...rest } = filters; filters = rest; }
    else filters = { ...filters, [i]: next };
  }
</script>

<div class="grid-wrap">
  <div class="stats">
    <div class="stat"><span class="k">Rows</span><span class="v">{viewRows.length}{viewRows.length !== (result?.rows?.length ?? 0) ? ` / ${result?.rows?.length ?? 0}` : ""}</span></div>
    <div class="stat"><span class="k">Time</span><span class="v">{result?.durationMs ?? 0} ms</span></div>
    {#if result?.truncated}<div class="stat warn">truncated</div>{/if}
    {#if copyMsg}<div class="stat ok">{copyMsg}</div>{/if}
    <div class="spacer"></div>
    <button class="btn" class:active={showFilters} onclick={() => (showFilters = !showFilters)}>Filter</button>
    <select class="copy-select" onchange={(e) => { const v = (e.currentTarget as HTMLSelectElement).value; if (v) { void copyAs(v as any); (e.currentTarget as HTMLSelectElement).value = ""; } }} disabled={!result}>
      <option value="">Copy as…</option>
      <option value="text">Text</option><option value="csv">CSV</option><option value="json">JSON</option><option value="insert">INSERT SQL</option>
    </select>
    <button class="btn" onclick={exportCsv} disabled={!result || !(result.columns?.length)}>CSV</button>
    <button class="btn" onclick={exportTsv} disabled={!result || !(result.columns?.length)}>TSV</button>
    <button class="btn" onclick={exportJson} disabled={!result}>JSON</button>
    <button class="btn" onclick={exportInsert} disabled={!result || !(result.columns?.length)}>INSERT</button>
  </div>

  <div class="main">
    <div class="scroll">
      {#if result && (result.columns?.length ?? 0) > 0}
        <table>
          <thead>
            <tr>
              <th class="rownum">#</th>
              {#each result.columns ?? [] as col, i (i)}
                <th title={col} class="sortable" onclick={() => toggleSort(i)}>
                  <span class="th-inner">{col}{#if sortCol === i}<span class="sort-ind">{sortDir === "asc" ? " ▲" : " ▼"}</span>{/if}</span>
                </th>
              {/each}
            </tr>
            {#if showFilters}
              <tr class="filter-row">
                <th class="rownum"></th>
                {#each result.columns ?? [] as _c, i (i)}
                  <th class="filter-cell">
                    <div class="filter-wrap">
                      <select value={filters[i]?.op ?? "contains"} onchange={(e) => setFilter(i, { op: (e.currentTarget as HTMLSelectElement).value as any })}>
                        <option value="contains">~</option><option value="equals">=</option><option value="regex">.*</option><option value="isnull">NULL</option><option value="notnull">!NULL</option>
                      </select>
                      <input placeholder="filter" value={filters[i]?.value ?? ""} disabled={filters[i]?.op === 'isnull' || filters[i]?.op === 'notnull'} oninput={(e) => setFilter(i, { value: (e.currentTarget as HTMLInputElement).value })} />
                    </div>
                  </th>
                {/each}
              </tr>
            {/if}
          </thead>
          <tbody>
            {#each viewRows as row, ri (ri)}
              <tr>
                <td class="rownum">{ri + 1}</td>
                {#each (row ?? []) as cell, ci (ci)}
                  <td title={fmt(cell)} class:selected={selected?.r === ri && selected?.c === ci} onclick={() => (selected = { r: ri, c: ci })}>
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
            {:else}
              <tr><td colspan={(result.columns?.length ?? 0) + 1} class="pad muted">No rows match filters.</td></tr>
            {/each}
          </tbody>
        </table>
      {:else if result}
        <div class="empty-state">
          <div class="empty-title">Statement executed</div>
          <div class="empty-sub">{result.rowCount ?? 0} rows affected in {result.durationMs ?? 0} ms</div>
        </div>
      {:else}
        <div class="empty-state">
          <div class="empty-title">No results</div>
          <div class="empty-sub">Run a query or open a table to view data.</div>
        </div>
      {/if}
    </div>

    {#if selected && viewRows[selected.r]}
      {@const col = result?.columns?.[selected.c] ?? `col ${selected.c + 1}`}
      {@const val = viewRows[selected.r]?.[selected.c]}
      <aside class="viewer">
        <div class="viewer-head">
          <div><div class="viewer-col">{col}</div><div class="viewer-meta">Row {selected.r + 1} · {val === null || val === undefined ? 'NULL' : typeof val}</div></div>
          <div class="viewer-actions">
            <button class="btn" onclick={() => navigator.clipboard.writeText(pretty(val))}>Copy</button>
            <button class="btn" onclick={() => (selected = null)}>Close</button>
          </div>
        </div>
        <pre class="viewer-body">{pretty(val)}</pre>
      </aside>
    {/if}
  </div>
</div>

<style>
  .grid-wrap { display: flex; flex-direction: column; height: 100%; min-height: 0; background: var(--bg); }
  .stats { height: 34px; display: flex; align-items: center; gap: 6px; padding: 0 10px; border-bottom: 1px solid var(--border); background: var(--panel); color: var(--muted-fg); font-size: 12px; flex-wrap: nowrap; overflow-x: auto; }
  .stat { display: inline-flex; align-items: center; gap: 6px; white-space: nowrap; }
  .k { color: var(--muted-fg); }
  .v { color: var(--fg); font-family: var(--mono); }
  .stat.warn { color: #fbbf24; background: rgba(245,158,11,.12); border: 1px solid rgba(245,158,11,.25); padding: 1px 6px; border-radius: 4px; font-size: 11px; }
  .stat.ok { color: #86efac; }
  .spacer { flex: 1; }
  .btn, .copy-select { height: 24px; padding: 0 8px; background: var(--panel-2); color: var(--fg); border: 1px solid var(--border); border-radius: 5px; font-size: 11.5px; }
  .btn:hover, .copy-select:hover { background: var(--hover); }
  .btn.active { border-color: var(--accent); color: #93c5fd; }
  .main { flex: 1; min-height: 0; display: flex; }
  .scroll { flex: 1; overflow: auto; min-height: 0; }
  table { width: max-content; min-width: 100%; border-collapse: separate; border-spacing: 0; font-size: 12.5px; font-family: var(--mono); }
  th, td { padding: 6px 10px; border-bottom: 1px solid var(--border-subtle); border-right: 1px solid var(--border-subtle); text-align: left; white-space: nowrap; max-width: 420px; overflow: hidden; text-overflow: ellipsis; }
  th { position: sticky; top: 0; z-index: 2; background: #1a202b; color: #c9d1d9; font-family: var(--sans); font-size: 11.5px; font-weight: 600; border-bottom: 1px solid var(--border); }
  th.sortable { cursor: pointer; user-select: none; }
  th.sortable:hover { background: #202838; }
  .th-inner { display: inline-flex; align-items: center; gap: 2px; }
  .sort-ind { color: var(--accent); font-size: 10px; }
  .filter-row th { top: 30px; z-index: 2; background: #161b23; padding: 4px 6px; }
  .filter-cell { max-width: 220px; }
  .filter-wrap { display: flex; gap: 4px; }
  .filter-wrap select, .filter-wrap input { height: 22px; background: var(--input); color: var(--fg); border: 1px solid var(--border); border-radius: 4px; font-size: 11px; }
  .filter-wrap select { width: 54px; padding: 0 2px; }
  .filter-wrap input { width: 100%; min-width: 60px; padding: 0 6px; }
  td { color: var(--fg); cursor: cell; }
  td.selected { outline: 1px solid var(--accent); background: rgba(59,130,246,.1) !important; }
  .rownum { width: 46px; min-width: 46px; max-width: 46px; text-align: right; color: var(--muted-fg); background: var(--bg-elev); position: sticky; left: 0; z-index: 1; border-right: 1px solid var(--border); cursor: default; }
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
  .pad { padding: 16px; text-align: center; }
  .muted { color: var(--muted-fg); }
  .viewer { width: 360px; max-width: 40%; border-left: 1px solid var(--border); background: var(--panel); display: flex; flex-direction: column; min-height: 0; }
  .viewer-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 8px; padding: 10px; border-bottom: 1px solid var(--border); }
  .viewer-col { font-weight: 600; font-size: 13px; font-family: var(--mono); }
  .viewer-meta { color: var(--muted-fg); font-size: 11.5px; margin-top: 2px; }
  .viewer-actions { display: flex; gap: 6px; }
  .viewer-body { flex: 1; margin: 0; padding: 10px; overflow: auto; white-space: pre-wrap; word-break: break-word; font-family: var(--mono); font-size: 12px; color: var(--fg); }
</style>
