<script lang="ts">
  interface Props {
    connected: boolean;
    rowCount: number;
    durationMs: number;
    pagePosition: string;
  }

  let { connected, rowCount, durationMs, pagePosition }: Props = $props();
</script>

<div class="status-bar" role="status">
  <span class="status-dot" data-connected={connected ? "true" : "false"}></span>
  <span class="status-text">{connected ? "connected" : "disconnected"}</span>
  <span class="status-sep">·</span>
  <span class="status-item">{rowCount} rows</span>
  <span class="status-sep">·</span>
  <span class="status-item">{durationMs}ms</span>
  <span class="status-sep">·</span>
  <span class="status-item">{pagePosition}</span>
</div>

<style>
  .status-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 12px;
    background: var(--card);
    border-top: 1px solid var(--border);
    font-size: 12px;
    color: var(--muted-foreground);
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 9999px;
    background: var(--destructive);
  }

  .status-dot[data-connected="true"] {
    background: var(--secondary);
  }
</style>
