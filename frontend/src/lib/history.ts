export type HistoryStatus = "ok" | "error" | "cancelled";

export type HistoryEntry = {
  id: string;
  connId: string;
  sql: string;
  ts: number;
  durationMs?: number;
  rowCount?: number;
  status: HistoryStatus;
  error?: string;
};

const KEY = "datapulse.history.v1";
const MAX = 300;

function read(): HistoryEntry[] {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? (parsed as HistoryEntry[]) : [];
  } catch {
    return [];
  }
}

function write(entries: HistoryEntry[]) {
  try {
    localStorage.setItem(KEY, JSON.stringify(entries.slice(0, MAX)));
  } catch {}
}

export function listHistory(connId?: string): HistoryEntry[] {
  const all = read();
  return connId ? all.filter((e) => e.connId === connId) : all;
}

export function addHistory(entry: Omit<HistoryEntry, "id" | "ts"> & { id?: string; ts?: number }) {
  const all = read();
  const next: HistoryEntry = {
    id: entry.id ?? crypto.randomUUID(),
    ts: entry.ts ?? Date.now(),
    connId: entry.connId,
    sql: entry.sql,
    durationMs: entry.durationMs,
    rowCount: entry.rowCount,
    status: entry.status,
    error: entry.error,
  };
  all.unshift(next);
  write(all);
  return next;
}

export function clearHistory(connId?: string) {
  if (!connId) { write([]); return; }
  write(read().filter((e) => e.connId !== connId));
}
