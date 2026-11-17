export type SavedQuery = {
  id: string;
  name: string;
  sql: string;
  connId: string;
  createdAt: number;
  updatedAt: number;
};

const KEY = "datapulse.savedQueries.v1";

function read(): SavedQuery[] {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? (parsed as SavedQuery[]) : [];
  } catch {
    return [];
  }
}

function write(items: SavedQuery[]) {
  try {
    localStorage.setItem(KEY, JSON.stringify(items));
  } catch {}
}

export function listSavedQueries(connId?: string): SavedQuery[] {
  const all = read().sort((a, b) => b.updatedAt - a.updatedAt);
  return connId ? all.filter((q) => q.connId === connId) : all;
}

export function saveQuery(input: { id?: string; name: string; sql: string; connId: string }): SavedQuery {
  const all = read();
  const now = Date.now();
  if (input.id) {
    const i = all.findIndex((q) => q.id === input.id);
    if (i >= 0) {
      all[i] = { ...all[i], name: input.name, sql: input.sql, connId: input.connId, updatedAt: now };
      write(all);
      return all[i];
    }
  }
  const q: SavedQuery = { id: crypto.randomUUID(), name: input.name, sql: input.sql, connId: input.connId, createdAt: now, updatedAt: now };
  all.unshift(q);
  write(all);
  return q;
}

export function deleteSavedQuery(id: string) {
  write(read().filter((q) => q.id !== id));
}

export function renameSavedQuery(id: string, name: string) {
  const all = read();
  const i = all.findIndex((q) => q.id === id);
  if (i >= 0) { all[i] = { ...all[i], name, updatedAt: Date.now() }; write(all); }
}
