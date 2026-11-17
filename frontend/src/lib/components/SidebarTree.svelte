<script lang="ts">
  import { onMount } from "svelte";
  import type { Profile, TableInfo, ColumnInfo } from "../../../bindings/github.com/collins/GoDBClient/internal/models/models";
  import { connect, disconnect, getDatabases, getTables, getColumns, listActive } from "../api";
  import { listHistory, clearHistory, type HistoryEntry } from "../history";
  import { listSavedQueries, deleteSavedQuery, type SavedQuery } from "../savedQueries";

  let {
    connections,
    onNew,
    onRefresh,
    onOpenTable,
    onGenerateSelect,
    onOpenQuery,
    onEdit,
    onDuplicate,
    onDelete,
    onOpenSql,
  }: {
    connections: Profile[];
    onNew: () => void;
    onRefresh: () => Promise<void> | void;
    onOpenTable: (connId: string, schema: string, table: string) => void;
    onGenerateSelect: (connId: string, schema: string, table: string) => void;
    onOpenQuery: (connId: string) => void;
    onEdit: (p: Profile) => void;
    onDuplicate: (p: Profile) => void;
    onDelete: (p: Profile) => void | Promise<void>;
    onOpenSql: (connId: string, sql: string, title?: string) => void;
  } = $props();

  let active = $state<Set<string>>(new Set());
  let open = $state<Record<string, boolean>>({});
  let dbs = $state<Record<string, string[]>>({});
  let tables = $state<Record<string, TableInfo[]>>({});
  let cols = $state<Record<string, ColumnInfo[]>>({});
  let loading = $state<Record<string, boolean>>({});
  let errors = $state<Record<string, string>>({});
  let menu = $state<{ x: number; y: number; connId: string; schema?: string; table?: string } | null>(null);

  let history = $state<HistoryEntry[]>([]);
  let saved = $state<SavedQuery[]>([]);
  let showHistory = $state(true);
  let showSaved = $state(true);

  function refreshLists() { history = listHistory().slice(0, 50); saved = listSavedQueries().slice(0, 50); }
  async function refreshActive() { try { active = new Set(await listActive()); } catch {} }

  async function reconnect(id: string) {
    try { if (active.has(id)) await disconnect(id); await connect(id); }
    finally { await refreshActive(); void onRefresh(); }
  }

  async function toggleConn(c: Profile) {
    const k = `c:${c.id}`;
    open[k] = !open[k];
    if (!open[k]) return;
    try {
      if (!active.has(c.id)) await connect(c.id);
      await refreshActive();
      if (!dbs[c.id]) {
        loading[k] = true; errors[k] = "";
        dbs[c.id] = await getDatabases(c.id);
      }
    } catch (e) { errors[k] = e instanceof Error ? e.message : String(e); }
    finally { loading[k] = false; }
  }

  async function toggleDb(connId: string, db: string) {
    const k = `d:${connId}:${db}`;
    open[k] = !open[k];
    if (!open[k] || tables[k]) return;
    loading[k] = true; errors[k] = "";
    try { tables[k] = await getTables(connId, db); }
    catch (e) { errors[k] = e instanceof Error ? e.message : String(e); }
    finally { loading[k] = false; }
  }

  async function toggleTable(connId: string, schema: string, table: string) {
    const k = `t:${connId}:${schema}:${table}`;
    open[k] = !open[k];
    if (!open[k] || cols[k]) return;
    loading[k] = true;
    try { cols[k] = await getColumns(connId, schema, table); }
    catch { cols[k] = []; }
    finally { loading[k] = false; }
  }

  function ctx(e: MouseEvent, connId: string, schema?: string, table?: string) {
    e.preventDefault(); e.stopPropagation();
    menu = { x: e.clientX, y: e.clientY, connId, schema, table };
  }

  function connName(id: string) { return connections.find(c => c.id === id)?.name || "connection"; }
  function shortSql(s: string) { return s.replace(/\s+/g, " ").trim().slice(0, 60) || "(empty)"; }
  function timeAgo(ts: number) {
    const d = Date.now() - ts; if (d < 60000) return "now"; if (d < 3600000) return `${Math.floor(d/60000)}m`; if (d < 86400000) return `${Math.floor(d/3600000)}h`; return `${Math.floor(d/86400000)}d`;
  }

  $effect(() => { connections; void refreshActive(); });
  onMount(() => {
    refreshLists();
    const h = () => refreshLists(); const s = () => refreshLists();
    window.addEventListener("godbclient:history", h);
    window.addEventListener("godbclient:saved", s);
    return () => { window.removeEventListener("godbclient:history", h); window.removeEventListener("godbclient:saved", s); };
  });
</script>

<svelte:window onclick={() => (menu = null)} />

<div class="nav">
  <div class="nav-head">
    <div class="nav-title">Database Explorer</div>
    <div class="nav-actions">
      <button class="icon-btn accent" onclick={onNew} title="New connection">+</button>
      <button class="icon-btn" onclick={() => onRefresh()} title="Refresh">↻</button>
    </div>
  </div>

  <div class="nav-body">
    <div class="section">
      {#if connections.length === 0}
        <div class="empty">
          <div class="empty-icon">DB</div>
          <div class="empty-title">No connections</div>
          <div class="empty-sub">Create a connection to browse schemas, tables, and run queries.</div>
          <button class="btn-primary" onclick={onNew}>New Connection</button>
        </div>
      {:else}
        {#each connections as c (c.id)}
          <div class="node conn"
            onclick={() => toggleConn(c)}
            onkeydown={(e) => e.key==='Enter' && toggleConn(c)}
            oncontextmenu={(e) => ctx(e, c.id)}
            role="button" tabindex="0">
            <span class="caret">{open[`c:${c.id}`] ? "▾" : "▸"}</span>
            <span class="dot" class:on={active.has(c.id)} title={active.has(c.id) ? 'Connected' : 'Disconnected'}></span>
            <span class="label">{c.name}</span>
            <button class="mini-btn" title="New SQL query" onclick={(e) => { e.stopPropagation(); onOpenQuery(c.id); }}>SQL</button>
            <span class="pill">{String(c.driver)}</span>
          </div>

          {#if open[`c:${c.id}`]}
            {#if loading[`c:${c.id}`]}<div class="node meta" style="--d:1">Connecting…</div>{/if}
            {#if errors[`c:${c.id}`]}<div class="node err" style="--d:1">{errors[`c:${c.id}`]}</div>{/if}
            {#each (dbs[c.id] ?? []) as db}
              <div class="node" style="--d:1" onclick={() => toggleDb(c.id, db)} onkeydown={(e)=>e.key==='Enter'&&toggleDb(c.id,db)} role="button" tabindex="0">
                <span class="caret">{open[`d:${c.id}:${db}`] ? "▾" : "▸"}</span>
                <span class="kind schema">S</span>
                <span class="label">{db}</span>
              </div>

              {#if open[`d:${c.id}:${db}`]}
                {#if loading[`d:${c.id}:${db}`]}<div class="node meta" style="--d:2">Loading…</div>{/if}
                {#each (tables[`d:${c.id}:${db}`] ?? []) as t}
                  <div class="node" style="--d:2"
                    onclick={() => toggleTable(c.id, t.schema, t.name)}
                    ondblclick={() => onOpenTable(c.id, t.schema, t.name)}
                    oncontextmenu={(e) => ctx(e, c.id, t.schema, t.name)}
                    onkeydown={(e)=>e.key==='Enter'&&toggleTable(c.id,t.schema,t.name)}
                    role="button" tabindex="0">
                    <span class="caret">{open[`t:${c.id}:${t.schema}:${t.name}`] ? "▾" : "▸"}</span>
                    <span class="kind" class:view={t.type?.toLowerCase().includes('view')}>{t.type?.toLowerCase().includes('view') ? 'V' : 'T'}</span>
                    <span class="label">{t.name}</span>
                  </div>

                  {#if open[`t:${c.id}:${t.schema}:${t.name}`]}
                    {#if loading[`t:${c.id}:${t.schema}:${t.name}`]}<div class="node meta" style="--d:3">Loading…</div>{/if}
                    {#each (cols[`t:${c.id}:${t.schema}:${t.name}`] ?? []) as col}
                      <div class="node leaf" style="--d:3">
                        <span class="caret"></span>
                        <span class="kind col" class:pk={col.isPrimary}>{col.isPrimary ? 'PK' : 'C'}</span>
                        <span class="label">{col.name}</span>
                        <span class="type">{col.type}</span>
                      </div>
                    {/each}
                  {/if}
                {/each}
                {#if !loading[`d:${c.id}:${db}`] && (tables[`d:${c.id}:${db}`] ?? []).length === 0}
                  <div class="node meta" style="--d:2">No tables</div>
                {/if}
              {/if}
            {/each}
          {/if}
        {/each}
      {/if}
    </div>

    <div class="section panel">
      <div class="panel-head" onclick={() => (showHistory = !showHistory)} role="button" tabindex="0" onkeydown={(e)=>e.key==='Enter'&&(showHistory=!showHistory)}>
        <span class="caret">{showHistory ? "▾" : "▸"}</span><span class="panel-title">Query History</span>
        <button class="mini-btn ghost" onclick={(e) => { e.stopPropagation(); clearHistory(); refreshLists(); }}>Clear</button>
      </div>
      {#if showHistory}
        <div class="list">
          {#each history as h (h.id)}
            <div class="list-item" title={h.sql} onclick={() => onOpenSql(h.connId, h.sql)} role="button" tabindex="0" onkeydown={(e)=>e.key==='Enter'&&onOpenSql(h.connId,h.sql)}>
              <span class="status-dot" class:ok={h.status==='ok'} class:err={h.status==='error'}></span>
              <div class="list-main">
                <div class="list-text">{shortSql(h.sql)}</div>
                <div class="list-meta">{connName(h.connId)} · {timeAgo(h.ts)}{h.durationMs != null ? ` · ${h.durationMs}ms` : ''}</div>
              </div>
            </div>
          {:else}
            <div class="node meta" style="--d:0">No history yet</div>
          {/each}
        </div>
      {/if}
    </div>

    <div class="section panel">
      <div class="panel-head" onclick={() => (showSaved = !showSaved)} role="button" tabindex="0" onkeydown={(e)=>e.key==='Enter'&&(showSaved=!showSaved)}>
        <span class="caret">{showSaved ? "▾" : "▸"}</span><span class="panel-title">Saved Queries</span>
      </div>
      {#if showSaved}
        <div class="list">
          {#each saved as q (q.id)}
            <div class="list-item" title={q.sql} onclick={() => onOpenSql(q.connId, q.sql, q.name)} role="button" tabindex="0" onkeydown={(e)=>e.key==='Enter'&&onOpenSql(q.connId,q.sql,q.name)}>
              <span class="kind schema">Q</span>
              <div class="list-main">
                <div class="list-text">{q.name}</div>
                <div class="list-meta">{connName(q.connId)}</div>
              </div>
              <button class="del-btn" title="Delete" onclick={(e) => { e.stopPropagation(); deleteSavedQuery(q.id); refreshLists(); window.dispatchEvent(new Event('godbclient:saved')); }}>×</button>
            </div>
          {:else}
            <div class="node meta" style="--d:0">No saved queries</div>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  {#if menu}
    <div class="ctx" style:left={`${menu.x}px`} style:top={`${menu.y}px`}>
      {#if menu.table && menu.schema}
        <button onclick={() => { onOpenTable(menu!.connId, menu!.schema!, menu!.table!); menu = null; }}>View Data</button>
        <button onclick={() => { onGenerateSelect(menu!.connId, menu!.schema!, menu!.table!); menu = null; }}>Generate SELECT</button>
        <button onclick={() => { navigator.clipboard.writeText(menu!.table!); menu = null; }}>Copy Table Name</button>
      {:else}
        {@const p = connections.find(c => c.id === menu!.connId)}
        <button onclick={() => { onOpenQuery(menu!.connId); menu = null; }}>New SQL Query</button>
        <button onclick={() => { void reconnect(menu!.connId); menu = null; }}>{active.has(menu!.connId) ? 'Reconnect' : 'Connect'}</button>
        {#if active.has(menu!.connId)}<button onclick={() => { void (async () => { await disconnect(menu!.connId); await refreshActive(); })(); menu = null; }}>Disconnect</button>{/if}
        {#if p}<button onclick={() => { onEdit(p); menu = null; }}>Edit Connection</button>{/if}
        {#if p}<button onclick={() => { onDuplicate(p); menu = null; }}>Duplicate Connection</button>{/if}
        {#if p}<button class="danger" onclick={() => { onDelete(p); menu = null; }}>Delete Connection</button>{/if}
      {/if}
    </div>
  {/if}
</div>

<style>
  .nav { height: 100%; display: flex; flex-direction: column; font-size: 12.5px; }
  .nav-head { display: flex; align-items: center; justify-content: space-between; padding: 10px 10px 8px; border-bottom: 1px solid var(--border-subtle); }
  .nav-title { font-size: 11px; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; color: var(--muted-fg); }
  .nav-actions { display: flex; gap: 6px; }
  .nav-body { flex: 1; overflow: auto; padding: 6px; display: flex; flex-direction: column; gap: 6px; }
  .section { display: flex; flex-direction: column; }
  .panel { border-top: 1px solid var(--border-subtle); padding-top: 4px; }
  .panel-head { display: flex; align-items: center; gap: 6px; padding: 4px 6px; cursor: pointer; color: var(--muted-fg); border-radius: 5px; }
  .panel-head:hover { background: var(--hover); color: var(--fg); }
  .panel-title { font-size: 11px; font-weight: 700; letter-spacing: .05em; text-transform: uppercase; flex: 1; }
  .list { display: flex; flex-direction: column; gap: 2px; padding: 2px 0 4px; }
  .list-item { display: flex; align-items: center; gap: 8px; padding: 5px 6px; border-radius: 5px; cursor: pointer; }
  .list-item:hover { background: var(--hover); }
  .list-main { min-width: 0; flex: 1; }
  .list-text { font-size: 12px; color: var(--fg); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .list-meta { font-size: 10.5px; color: var(--muted-fg); }
  .status-dot { width: 7px; height: 7px; border-radius: 50%; background: #64748b; flex: 0 0 auto; }
  .status-dot.ok { background: var(--ok); }
  .status-dot.err { background: #ef4444; }
  .del-btn { width: 18px; height: 18px; border-radius: 4px; color: var(--muted-fg); font-size: 14px; display: inline-flex; align-items: center; justify-content: center; }
  .del-btn:hover { background: rgba(239,68,68,.15); color: #f87171; }
  .icon-btn { width: 24px; height: 24px; border-radius: 4px; border: 1px solid var(--border); color: var(--muted-fg); background: var(--panel-2); display: inline-flex; align-items: center; justify-content: center; font-size: 14px; }
  .icon-btn:hover { color: var(--fg); background: var(--hover); }
  .icon-btn.accent { background: var(--accent); border-color: var(--accent); color: #fff; }
  .empty { margin: 24px 8px; padding: 18px 14px; border: 1px dashed var(--border); border-radius: 8px; background: rgba(255,255,255,.01); display: flex; flex-direction: column; align-items: center; text-align: center; gap: 8px; }
  .empty-icon { width: 36px; height: 36px; border-radius: 8px; background: var(--panel-2); border: 1px solid var(--border); display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; color: var(--muted-fg); }
  .empty-title { color: var(--fg); font-weight: 600; }
  .empty-sub { color: var(--muted-fg); font-size: 12px; line-height: 1.4; }
  .btn-primary { margin-top: 4px; background: var(--accent); color: #fff; border: 0; border-radius: 6px; padding: 7px 12px; font-weight: 500; }
  .node { --d: 0; display: flex; align-items: center; gap: 6px; height: 28px; padding: 0 8px 0 calc(8px + (var(--d) * 14px)); border-radius: 5px; cursor: pointer; color: var(--fg); }
  .node:hover { background: var(--hover); }
  .node.conn { font-weight: 500; }
  .node.leaf { cursor: default; }
  .mini-btn { margin-left: auto; height: 18px; padding: 0 6px; border-radius: 4px; font-size: 10px; font-weight: 600; color: #93c5fd; background: rgba(59,130,246,.12); border: 1px solid rgba(59,130,246,.25); }
  .mini-btn:hover { background: rgba(59,130,246,.2); color: #bfdbfe; }
  .mini-btn.ghost { color: var(--muted-fg); background: transparent; border-color: transparent; font-weight: 500; }
  .node.meta, .node.err { cursor: default; color: var(--muted-fg); font-size: 12px; }
  .node.err { color: #f87171; }
  .caret { width: 10px; color: var(--muted-fg); font-size: 11px; flex: 0 0 auto; }
  .dot { width: 8px; height: 8px; border-radius: 50%; background: #4b5563; flex: 0 0 auto; }
  .dot.on { background: var(--ok); box-shadow: 0 0 0 2px rgba(34,197,94,.15); }
  .kind { min-width: 18px; height: 16px; padding: 0 4px; border-radius: 3px; font-size: 9px; font-weight: 700; display: inline-flex; align-items: center; justify-content: center; background: rgba(148,163,184,.12); color: #cbd5e1; border: 1px solid rgba(148,163,184,.18); flex: 0 0 auto; }
  .kind.schema { color: #c4b5fd; background: rgba(139,92,246,.14); border-color: rgba(139,92,246,.25); }
  .kind.view { color: #67e8f9; background: rgba(6,182,212,.14); border-color: rgba(6,182,212,.25); }
  .kind.col { color: #94a3b8; }
  .kind.pk { color: #fbbf24; background: rgba(245,158,11,.14); border-color: rgba(245,158,11,.25); }
  .label { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .pill, .type { margin-left: auto; color: var(--muted-fg); font-size: 10.5px; }
  .pill { text-transform: uppercase; background: var(--panel-2); border: 1px solid var(--border); padding: 1px 5px; border-radius: 4px; }
  .ctx { position: fixed; z-index: 200; min-width: 190px; background: #1a202b; border: 1px solid var(--border); border-radius: 8px; padding: 4px; box-shadow: var(--shadow); }
  .ctx button { display: block; width: 100%; text-align: left; background: transparent; border: 0; color: var(--fg); padding: 7px 10px; border-radius: 5px; font-size: 12.5px; }
  .ctx button:hover { background: var(--hover); }
  .ctx button.danger { color: #f87171; }
</style>