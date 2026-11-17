export type WorkspaceTab = {
  id: string;
  title: string;
  kind: "sql" | "table";
  connId: string;
  sql?: string;
  schema?: string;
  table?: string;
};

export type WorkspaceState = {
  version: 1;
  tabs: WorkspaceTab[];
  activeTab: string;
  seq: number;
};

const KEY = "datapulse.workspace.v1";

export function loadWorkspace(): WorkspaceState | null {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<WorkspaceState>;
    if (parsed.version !== 1 || !Array.isArray(parsed.tabs)) return null;
    const tabs = parsed.tabs.filter(
      (t): t is WorkspaceTab =>
        !!t && typeof t.id === "string" && (t.kind === "sql" || t.kind === "table") && typeof t.title === "string",
    );
    if (tabs.length === 0) return null;
    return {
      version: 1,
      tabs,
      activeTab: typeof parsed.activeTab === "string" ? parsed.activeTab : tabs[0].id,
      seq: typeof parsed.seq === "number" && parsed.seq > 0 ? parsed.seq : tabs.length,
    };
  } catch {
    return null;
  }
}

let saveTimer: number | null = null;
export function saveWorkspace(state: WorkspaceState) {
  if (saveTimer) window.clearTimeout(saveTimer);
  saveTimer = window.setTimeout(() => {
    try {
      localStorage.setItem(KEY, JSON.stringify(state));
    } catch {}
  }, 300);
}

export function saveWorkspaceNow(state: WorkspaceState) {
  if (saveTimer) {
    window.clearTimeout(saveTimer);
    saveTimer = null;
  }
  try {
    localStorage.setItem(KEY, JSON.stringify(state));
  } catch {}
}
