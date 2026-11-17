<script lang="ts">
  interface Props {
    columns: string[];
    rows: (any[] | null)[];
  }

  let { columns, rows }: Props = $props();
</script>

<div class="query-grid">
  <table>
    <thead>
      <tr>
        {#each columns as column, i (i)}
          <th>{column}</th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each rows as row, i (i)}
        <tr>
          {#each row as value, j (j)}
            <td class:numeric={typeof value === "number"}>{value === null ? "NULL" : String(value)}</td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  .query-grid {
    width: 100%;
    overflow: auto;
    border: 1px solid var(--border);
    border-radius: var(--radius-md, 8px);
    background: var(--card);
  }

  table {
    border-collapse: collapse;
    width: 100%;
    font-size: 13px;
  }

  th, td {
    padding: 8px 12px;
    border-bottom: 1px solid var(--border);
    text-align: left;
    font-family: var(--font-mono, monospace);
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  th {
    position: sticky;
    top: 0;
    background: var(--muted);
    font-weight: 600;
  }

  tbody tr:nth-child(even) {
    background: var(--muted);
  }

  td.numeric {
    text-align: right;
  }

  td:empty {
    color: var(--muted-foreground);
  }
</style>
