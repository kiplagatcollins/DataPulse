<script lang="ts">
  import { onMount } from "svelte";
  import DataGrid from "./DataGrid.svelte";
  import { executeSql, getColumns, getForeignKeys, getIndexes } from "../api";
  import type { Result, ColumnInfo, ForeignKeyInfo } from "../../../bindings/github.com/collins/DataPulse/internal/models/models";

  let { connId, schema, table }: { connId: string; schema: string; table: string } = $props();

  let view = $state<"data" | "structure">("data");
  let result = $state<Result | null>(null);
  let columns = $state<ColumnInfo[]>([]);
  let fks = $state<ForeignKeyInfo[]>([]);
  let indexes = $state<string[]>([]);
  let error = $state<string | null>(null);
  let loading = $state(false);
  let limit = $state(100);
  let offset = $state(0);

  function q(name: string) { return `"${name.replace(/"/g, '""')}"`; }
  function ref() { return schema && schema !== "main" ? `${q(schema)}.${q(table)}` : q(table); }

  async function loadData() {
    loading = true; error = null;
    try { result = await executeSql(connId, `SELECT * FROM ${ref()} LIMIT ${limit} OFFSET ${offset};`, 1, limit); }
    catch (e) { error = e instanceof Error ? e.message : String(e); result = null; }
    finally { loading = false; }
  }
  async function loadStructure() {
    const [c, f, i] = await Promise.all([
      getColumns(connId, schema, table).catch(() => [] as ColumnInfo[]),
      getForeignKeys(connId, schema, table).catch(() => [] as ForeignKeyInfo[]),
      getIndexes(connId, schema, table).catch(() => [] as string[]),
    ]);
    columns = c; fks = f; indexes = i;
  }

  function prev() { offset = Math.max(0, offset - limit); void loadData(); }
  function next() { offset += limit; void loadData(); }
  onMount(() => { void loadData(); void loadStructure(); });
</script>

<div class="table-tab">
  <div class="bar">
    <div class="title"><span class="obj-icon">T</span><span class="name">{schema}.{table}</span></div>
    <div class="seg">
      <button class:active={view==='data'} onclick={() => view='data'}>Data</button>
      <button class:active={view==='structure'} onclick={() => view='structure'}>Structure</button>
    </div>
    <div class="spacer"></div>
    {#if view==='data'}
      <label class="ctl">Rows
        <select bind:value={limit} onchange={() => { offset = 0; void loadData(); }}>
          <option value={50}>50</option><option value={100}>100</option><option value={250}>250</option><option value={500}>500</option>
        </select>
      </label>
      <button class="ctl-btn" onclick={prev} disabled={offset===0 || loading}>Prev</button>
      <span class="meta">offset {offset}</span>
      <button class="ctl-btn" onclick={next} disabled={loading || (result?.rows?.length ?? 0) < limit}>Next</button>
      <button class="ctl-btn" onclick={loadData} disabled={loading}>{loading ? '…' : 'Refresh'}</button>
    {/if}
  </div>

  {#if error}<div class="error">{error}</div>{/if}

  {#if view === 'data'}
    <div class="content"><DataGrid {result} /></div>
  {:else}
    <div class="structure">
      <div class="section-title">Columns</div>
      <table>
        <thead><tr><th>Name</th><th>Type</th><th>Nullable</th><th>Primary Key</th><th>Default</th></tr></thead>
        <tbody>
          {#each columns as c}
            <tr>
              <td class="colname">{#if c.isPrimary}<span class="pk">PK</span>{/if}{c.name}</td>
              <td><span class="type">{c.type}</span></td>
              <td><span class:yes={c.nullable} class:no={!c.nullable}>{c.nullable ? 'YES' : 'NO'}</span></td>
              <td>{c.isPrimary ? 'YES' : '—'}</td>
              <td class="muted">{c.default ?? '—'}</td>
            </tr>
          {:else}
            <tr><td colspan="5" class="muted pad">No column information available.</td></tr>
          {/each}
        </tbody>
      </table>

      <div class="section-title">Foreign Keys</div>
      <table>
        <thead><tr><th>Name</th><th>Columns</th><th>References</th></tr></thead>
        <tbody>
          {#each fks as fk}
            <tr>
              <td class="colname">{fk.name || '—'}</td>
              <td class="type">{(fk.columns ?? []).join(', ') || '—'}</td>
              <td class="type">{fk.referencedSchema ? fk.referencedSchema + '.' : ''}{fk.referencedTable}{(fk.referencedColumns ?? []).length ? '(' + (fk.referencedColumns ?? []).join(', ') + ')' : ''}</td>
            </tr>
          {:else}
            <tr><td colspan="3" class="muted pad">No foreign keys.</td></tr>
          {/each}
        </tbody>
      </table>

      <div class="section-title">Indexes</div>
      <table>
        <thead><tr><th>Name</th></tr></thead>
        <tbody>
          {#each indexes as ix}
            <tr><td class="colname">{ix}</td></tr>
          {:else}
            <tr><td class="muted pad">No indexes.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .table-tab { display: flex; flex-direction: column; height: 100%; min-height: 0; }
  .bar { height: 40px; display: flex; align-items: center; gap: 10px; padding: 0 10px; border-bottom: 1px solid var(--border); background: var(--panel); }
  .title { display: flex; align-items: center; gap: 8px; font-weight: 600; font-size: 12.5px; }
  .obj-icon { width: 18px; height: 18px; border-radius: 4px; background: rgba(59,130,246,.15); color: #93c5fd; border: 1px solid rgba(59,130,246,.25); display: inline-flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 700; }
  .name { font-family: var(--mono); font-weight: 500; }
  .seg { display: inline-flex; background: var(--panel-2); border: 1px solid var(--border); border-radius: 6px; padding: 2px; margin-left: 6px; }
  .seg button { height: 24px; padding: 0 10px; border-radius: 4px; color: var(--muted-fg); font-size: 12px; }
  .seg button.active { background: var(--active); color: var(--fg); }
  .spacer { flex: 1; }
  .ctl { display: inline-flex; align-items: center; gap: 6px; color: var(--muted-fg); font-size: 12px; }
  .ctl-btn, select { height: 26px; padding: 0 8px; background: var(--input); color: var(--fg); border: 1px solid var(--border); border-radius: 5px; font-size: 12px; }
  .ctl-btn:hover { background: var(--hover); }
  .meta { color: var(--muted-fg); font-size: 12px; font-family: var(--mono); }
  .content { flex: 1; min-height: 0; }
  .error { padding: 8px 10px; color: #fca5a5; background: rgba(239,68,68,.08); border-bottom: 1px solid rgba(239,68,68,.2); font-size: 12px; }
  .structure { flex: 1; overflow: auto; padding-bottom: 16px; }
  .section-title { padding: 10px 12px 6px; font-size: 11px; font-weight: 700; letter-spacing: .05em; text-transform: uppercase; color: var(--muted-fg); border-top: 1px solid var(--border-subtle); margin-top: 4px; }
  .section-title:first-child { border-top: 0; margin-top: 0; }
  table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
  th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid var(--border); }
  th { position: sticky; top: 0; background: var(--panel-2); color: var(--muted-fg); font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: .04em; }
  .colname { font-family: var(--mono); display: flex; align-items: center; gap: 6px; }
  .pk { font-size: 9px; font-weight: 700; color: #fbbf24; background: rgba(245,158,11,.14); border: 1px solid rgba(245,158,11,.25); padding: 1px 4px; border-radius: 3px; }
  .type { font-family: var(--mono); color: #93c5fd; font-size: 12px; }
  .yes { color: #86efac; } .no { color: var(--muted-fg); }
  .muted { color: var(--muted-fg); }
  .pad { padding: 24px 12px; text-align: center; }
</style>
