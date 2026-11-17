<script lang="ts">
  import type { Profile } from "../../../bindings/github.com/collins/GoDBClient/internal/models/models";
  import { saveConnection, testConnection } from "../api";

  let { open = $bindable(false), onSaved }: { open: boolean; onSaved: () => void | Promise<void> } = $props();

  function emptyProfile(): Profile {
    return { id: "", name: "", driver: "postgres" as any, host: "localhost", port: 5432, database: "postgres", username: "postgres", sslmode: "prefer", options: null };
  }

  let form = $state<Profile>(emptyProfile());
  let password = $state("");
  let sqlitePath = $state("");
  let status = $state<string | null>(null);
  let busy = $state(false);

  function applyDriverDefaults(driver: string) {
    form.driver = driver as any;
    if (driver === "postgres") { form.host = "localhost"; form.port = 5432; form.database = "postgres"; form.username = "postgres"; form.sslmode = "prefer"; }
    else if (driver === "mysql") { form.host = "localhost"; form.port = 3306; form.database = "mysql"; form.username = "root"; form.sslmode = ""; }
    else if (driver === "sqlite") { form.host = ""; form.port = 0; form.database = sqlitePath || ""; form.username = ""; form.sslmode = ""; }
    else if (driver === "oracle") { form.host = "localhost"; form.port = 1521; form.database = "xe"; form.username = "system"; form.sslmode = ""; }
  }

  function close() { if (!busy) { open = false; status = null; } }
  function onBackdropKey(e: KeyboardEvent) { if (e.key === "Escape") close(); }

  function buildProfile(): Profile {
    return { ...form, database: String(form.driver) === "sqlite" ? (sqlitePath || form.database || "") : form.database };
  }

  async function onTest() {
    busy = true; status = null;
    try { await testConnection(buildProfile(), password); status = "ok:Connection successful"; }
    catch (e) { status = "err:" + (e instanceof Error ? e.message : String(e)); }
    finally { busy = false; }
  }

  async function onSave() {
    busy = true; status = null;
    try {
      await saveConnection(buildProfile(), password);
      form = emptyProfile(); password = ""; sqlitePath = ""; open = false;
      await onSaved();
    } catch (e) { status = "err:" + (e instanceof Error ? e.message : String(e)); }
    finally { busy = false; }
  }
</script>

{#if open}
  <div class="modal-backdrop" role="button" tabindex="0" aria-label="Close dialog" onclick={close} onkeydown={onBackdropKey}>
    <div class="modal" role="dialog" aria-modal="true" aria-label="New Connection" tabindex="-1" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
      <header class="modal-head">
        <div>
          <h3>New Connection</h3>
          <p>Configure and save a database connection profile.</p>
        </div>
        <button class="icon-btn" onclick={close} aria-label="Close">×</button>
      </header>

      <div class="form-grid">
        <label class="field span-2"><span>Connection Name</span><input bind:value={form.name} placeholder="e.g. Production Postgres" /></label>
        <label class="field"><span>Driver</span>
          <select value={form.driver} onchange={(e) => applyDriverDefaults((e.target as HTMLSelectElement).value)}>
            <option value="postgres">PostgreSQL</option>
            <option value="mysql">MySQL</option>
            <option value="sqlite">SQLite</option>
            <option value="oracle">Oracle</option>
          </select>
        </label>
        <div class="field hint"><span>Driver Info</span><div class="hint-box">{String(form.driver) === 'sqlite' ? 'Local file-based database.' : 'Network database connection.'}</div></div>

        {#if String(form.driver) === "sqlite"}
          <label class="field span-2"><span>Database File Path</span><input bind:value={sqlitePath} placeholder="/absolute/path/to/database.sqlite" /></label>
        {:else}
          <label class="field"><span>Host</span><input bind:value={form.host} placeholder="localhost" /></label>
          <label class="field"><span>Port</span><input type="number" bind:value={form.port} /></label>
          <label class="field"><span>Database</span><input bind:value={form.database} /></label>
          <label class="field"><span>Username</span><input bind:value={form.username} /></label>
          <label class="field"><span>Password</span><input type="password" bind:value={password} placeholder="Stored in OS keychain" /></label>
          <label class="field"><span>SSL Mode</span><input bind:value={form.sslmode} placeholder="prefer / disable / require" /></label>
        {/if}
      </div>

      {#if status}
        <div class="status" class:ok={status.startsWith('ok:')} class:err={status.startsWith('err:')}>
          {status.replace(/^ok:|^err:/, '')}
        </div>
      {/if}

      <footer class="modal-actions">
        <button class="btn secondary" onclick={onTest} disabled={busy || !form.name}>Test Connection</button>
        <div class="spacer"></div>
        <button class="btn ghost" onclick={close} disabled={busy}>Cancel</button>
        <button class="btn primary" onclick={onSave} disabled={busy || !form.name}>Save Connection</button>
      </footer>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop { position: fixed; inset: 0; background: rgba(2, 6, 12, 0.7); backdrop-filter: blur(2px); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px; }
  .modal { width: 640px; max-width: 100%; background: var(--panel); border: 1px solid var(--border); border-radius: 10px; box-shadow: var(--shadow); overflow: hidden; }
  .modal-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; padding: 16px 18px; border-bottom: 1px solid var(--border); background: linear-gradient(180deg, #1a202b, var(--panel)); }
  .modal-head h3 { margin: 0; font-size: 15px; font-weight: 600; }
  .modal-head p { margin: 4px 0 0; font-size: 12px; color: var(--muted-fg); }
  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; padding: 16px 18px; }
  .span-2 { grid-column: 1 / -1; }
  .field { display: flex; flex-direction: column; gap: 5px; }
  .field > span { font-size: 11.5px; font-weight: 500; color: var(--muted-fg); }
  input, select { height: 34px; background: var(--input); color: var(--fg); border: 1px solid var(--border); border-radius: 6px; padding: 0 10px; }
  input::placeholder { color: #6b7280; }
  .hint-box { height: 34px; display: flex; align-items: center; padding: 0 10px; border: 1px dashed var(--border); border-radius: 6px; color: var(--muted-fg); font-size: 12px; }
  .status { margin: 0 18px 14px; padding: 8px 10px; border-radius: 6px; font-size: 12px; border: 1px solid var(--border); }
  .status.ok { color: #86efac; background: rgba(34,197,94,.08); border-color: rgba(34,197,94,.25); }
  .status.err { color: #fca5a5; background: rgba(239,68,68,.08); border-color: rgba(239,68,68,.25); }
  .modal-actions { display: flex; align-items: center; gap: 8px; padding: 12px 18px; border-top: 1px solid var(--border); background: rgba(0,0,0,.08); }
  .spacer { flex: 1; }
  .btn { height: 32px; padding: 0 12px; border-radius: 6px; font-size: 12.5px; font-weight: 500; border: 1px solid var(--border); }
  .btn.primary { background: var(--accent); border-color: var(--accent); color: #fff; }
  .btn.primary:hover { background: var(--accent-hover); }
  .btn.secondary { background: var(--panel-2); color: var(--fg); }
  .btn.secondary:hover { background: var(--hover); }
  .btn.ghost { background: transparent; color: var(--muted-fg); }
  .icon-btn { width: 28px; height: 28px; border-radius: 6px; border: 1px solid transparent; color: var(--muted-fg); font-size: 18px; display: inline-flex; align-items: center; justify-content: center; }
  .icon-btn:hover { background: var(--hover); color: var(--fg); border-color: var(--border); }
</style>
