<script lang="ts">
  import { onMount } from "svelte";
  import SidebarTree from "./lib/components/SidebarTree.svelte";
  import ConnectionModal from "./lib/components/ConnectionModal.svelte";
  import SqlEditorTab from "./lib/components/SqlEditorTab.svelte";
  import TableDataTab from "./lib/components/TableDataTab.svelte";
  import { getConnections, listActive } from "./lib/api";
  import { loadWorkspace, saveWorkspace, saveWorkspaceNow } from "./lib/workspace";
  import type { Profile } from "../bindings/github.com/collins/GoDBClient/internal/models/models";

  type Tab =
    | { id: string; title: string; kind: "sql"; connId: string; sql: string }
    | { id: string; title: string; kind: "table"; connId: string; schema: string; table: string };

  let connections = $state<Profile[]>([]);
  let activeIds = $state<string[]>([]);
  let showConn = $state(false);
  let seq = $state(1);
  const initialTab: Tab = { id: crypto.randomUUID(), title: "SQL 1", kind: "sql", connId: "", sql: "SELECT 1;" };
  let tabs = $state<Tab[]>([initialTab]);
  let activeTab = $state(initialTab.id);

  function snapshot() {
    return {
      version: 1 as const,
      tabs: tabs.map((t) => t.kind === "sql"
        ? { id: t.id, title: t.title, kind: "sql" as const, connId: t.connId, sql: t.sql }
        : { id: t.id, title: t.title, kind: "table" as const, connId: t.connId, schema: t.schema, table: t.table }),
      activeTab,
      seq,
    };
  }

  function persist() { saveWorkspace(snapshot()); }
  function persistNow() { saveWorkspaceNow(snapshot()); }

  $effect(() => { tabs; activeTab; seq; persist(); });

  async function refresh() {
    connections = await getConnections();
    try { activeIds = await listActive(); } catch { activeIds = []; }
    if (activeIds.length && tabs.some(t => !t.connId)) {
      tabs = tabs.map(t => t.connId ? t : ({ ...t, connId: activeIds[0] } as Tab));
    }
  }

  function restore() {
    const ws = loadWorkspace();
    if (!ws) return;
    const ids = new Set(connections.map((c) => c.id));
    const restored: Tab[] = [];
    for (const t of ws.tabs) {
      const connId = ids.has(t.connId) ? t.connId : "";
      if (t.kind === "sql") {
        restored.push({ id: t.id, title: t.title, kind: "sql", connId, sql: t.sql ?? "SELECT 1;" });
      } else if (t.schema && t.table) {
        restored.push({ id: t.id, title: t.title, kind: "table", connId, schema: t.schema, table: t.table });
      }
    }
    if (restored.length === 0) return;
    tabs = restored;
    seq = Math.max(ws.seq, restored.length);
    activeTab = restored.some((t) => t.id === ws.activeTab) ? ws.activeTab : restored[0].id;
  }

  function currentConn(): string {
    const t = tabs.find(x => x.id === activeTab);
    return t?.connId || activeIds[0] || "";
  }

  function newSql(initial = "", connId = "") {
    seq += 1;
    const t: Tab = { id: crypto.randomUUID(), title: `SQL ${seq}`, kind: "sql", connId: connId || currentConn(), sql: initial || "SELECT 1;" };
    tabs = [...tabs, t];
    activeTab = t.id;
  }

  function openQuery(connId: string) {
    newSql("", connId);
  }

  function openTable(connId: string, schema: string, table: string) {
    const t: Tab = { id: crypto.randomUUID(), title: `${schema}.${table}`, kind: "table", connId, schema, table };
    tabs = [...tabs, t];
    activeTab = t.id;
  }

  function generateSelect(connId: string, schema: string, table: string) {
    const ref = schema && schema !== "main" ? `"${schema}"."${table}"` : `"${table}"`;
    newSql(`SELECT *\nFROM ${ref}\nLIMIT 100;`);
    const t = tabs.find(x => x.id === activeTab);
    if (t && t.kind === "sql") t.connId = connId;
  }

  function updateSql(id: string, sql: string) {
    const i = tabs.findIndex((t) => t.id === id && t.kind === "sql");
    if (i < 0) return;
    const t = tabs[i];
    if (t.kind !== "sql" || t.sql === sql) return;
    tabs = tabs.map((tab) => tab.id === id && tab.kind === "sql" ? { ...tab, sql } : tab);
  }

  function closeTab(id: string) {
    const i = tabs.findIndex(t => t.id === id);
    if (i < 0) return;
    tabs = tabs.filter(t => t.id !== id);
    if (activeTab === id) activeTab = tabs[Math.max(0, i - 1)]?.id ?? "";
    if (tabs.length === 0) {
      seq = 1;
      const t: Tab = { id: crypto.randomUUID(), title: "SQL 1", kind: "sql", connId: currentConn(), sql: "SELECT 1;" };
      tabs = [t]; activeTab = t.id;
    }
  }

  onMount(() => {
    (async () => {
      await refresh();
      restore();
    })();
    const onHide = () => persistNow();
    window.addEventListener("pagehide", onHide);
    return () => window.removeEventListener("pagehide", onHide);
  });
</script>

<div class="ide">
  <header class="topbar">
    <div class="brand">
      <span class="brand-mark"></span>
      <span class="brand-name">GoDBClient</span>
    </div>
    <div class="topbar-actions">
      <button class="btn primary" onclick={() => (showConn = true)}>+ New Connection</button>
    </div>
  </header>

  <div class="body">
    <aside class="sidebar">
      <SidebarTree
        {connections}
        onNew={() => (showConn = true)}
        onRefresh={refresh}
        onOpenTable={openTable}
        onGenerateSelect={generateSelect}
        onOpenQuery={openQuery}
      />
    </aside>

    <main class="workspace">
      <div class="tabstrip">
        {#each tabs as t (t.id)}
          <div class="tab" class:active={t.id === activeTab} onclick={() => (activeTab = t.id)} role="button" tabindex="0" onkeydown={(e)=>e.key==='Enter'&&(activeTab=t.id)}>
            <span class="tab-icon">{t.kind === 'sql' ? 'SQL' : 'TBL'}</span>
            <span class="tab-title" title={t.title}>{t.title}</span>
            <button class="close" onclick={(e) => { e.stopPropagation(); closeTab(t.id); }} aria-label="Close tab">×</button>
          </div>
        {/each}
        <button class="add" onclick={() => newSql("")} title="New SQL tab">+</button>
      </div>

      <div class="pane">
        {#each tabs as t (t.id)}
          <section class="tabpane" class:show={t.id === activeTab}>
            {#if t.kind === "sql"}
              <SqlEditorTab connId={t.connId || currentConn()} initialSql={t.sql} onSqlChange={(s) => updateSql(t.id, s)} />
            {:else}
              <TableDataTab connId={t.connId} schema={t.schema} table={t.table} />
            {/if}
          </section>
        {/each}
      </div>

      <footer class="statusline">
        <span class="status-item"><span class="dot" class:on={activeIds.length > 0}></span>{activeIds.length} active</span>
        <span class="sep"></span>
        <span class="status-item">{connections.length} connections</span>
        <span class="sep"></span>
        <span class="status-item">{tabs.length} tabs</span>
      </footer>
    </main>
  </div>
</div>

<ConnectionModal bind:open={showConn} onSaved={refresh} />

<style>
  .ide { height: 100%; display: flex; flex-direction: column; background: var(--bg); color: var(--fg); }
  .topbar {
    height: 44px; display: flex; align-items: center; justify-content: space-between;
    padding: 0 12px; border-bottom: 1px solid var(--border); background: linear-gradient(180deg, #171c24, var(--panel));
  }
  .brand { display: flex; align-items: center; gap: 10px; font-weight: 600; font-size: 13px; }
  .brand-mark { width: 10px; height: 10px; border-radius: 3px; background: linear-gradient(135deg, #60a5fa, #2563eb); box-shadow: 0 0 0 3px rgba(59,130,246,.15); }
  .brand-name { letter-spacing: .2px; }
  .btn { border-radius: var(--radius); padding: 6px 12px; font-size: 12.5px; font-weight: 500; }
  .btn.primary { background: var(--accent); color: #fff; border: 1px solid var(--accent); }
  .btn.primary:hover { background: var(--accent-hover); }
  .body { flex: 1; min-height: 0; display: grid; grid-template-columns: 300px 1fr; }
  .sidebar { min-height: 0; overflow: hidden; border-right: 1px solid var(--border); background: var(--panel); }
  .workspace { min-width: 0; min-height: 0; display: flex; flex-direction: column; background: var(--bg); }
  .tabstrip {
    display: flex; align-items: flex-end; gap: 1px; padding: 6px 8px 0; height: 38px;
    border-bottom: 1px solid var(--border); background: var(--panel); overflow-x: auto; overflow-y: hidden;
  }
  .tab {
    height: 32px; display: flex; align-items: center; gap: 8px; padding: 0 10px; max-width: 220px;
    color: var(--muted-fg); background: transparent; border: 1px solid transparent; border-bottom: 0;
    border-radius: 6px 6px 0 0; cursor: pointer;
  }
  .tab:hover { color: var(--fg); background: var(--hover); }
  .tab.active { color: var(--fg); background: var(--bg); border-color: var(--border); position: relative; }
  .tab.active::after { content: ""; position: absolute; left: 0; right: 0; bottom: -1px; height: 1px; background: var(--bg); }
  .tab-icon { font-size: 9px; font-weight: 700; letter-spacing: .04em; color: #93c5fd; background: rgba(59,130,246,.15); border: 1px solid rgba(59,130,246,.25); padding: 1px 4px; border-radius: 3px; }
  .tab-title { font-size: 12.5px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .close, .add { color: var(--muted-fg); border-radius: 4px; display: inline-flex; align-items: center; justify-content: center; }
  .close { width: 18px; height: 18px; font-size: 15px; }
  .close:hover, .add:hover { background: rgba(255,255,255,.06); color: var(--fg); }
  .add { width: 28px; height: 28px; margin: 0 2px 4px; font-size: 18px; }
  .pane { flex: 1; min-height: 0; position: relative; }
  .tabpane { display: none; height: 100%; }
  .tabpane.show { display: block; height: 100%; }
  .statusline {
    height: 24px; display: flex; align-items: center; gap: 10px; padding: 0 10px;
    border-top: 1px solid var(--border); background: var(--panel); color: var(--muted-fg); font-size: 11.5px;
  }
  .status-item { display: inline-flex; align-items: center; gap: 6px; }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: #4b5563; }
  .dot.on { background: var(--ok); box-shadow: 0 0 0 2px rgba(34,197,94,.15); }
  .sep { width: 1px; height: 12px; background: var(--border); }
</style>
