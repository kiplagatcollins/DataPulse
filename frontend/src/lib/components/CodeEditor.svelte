<script lang="ts">
  interface Props {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
    rows?: number;
  }

  let { value, onChange, placeholder = "", rows = 10 }: Props = $props();
</script>

<div class="code-editor">
  <textarea
    class="code-editor-area"
    value={value}
    oninput={(e) => onChange(e.currentTarget.value)}
    {placeholder}
    {rows}
    spellcheck="false"
    aria-label="sql"
  ></textarea>
</div>

<style>
  .code-editor {
    border: 1px solid var(--border);
    border-radius: var(--radius-md, 8px);
    background: var(--muted);
    overflow: hidden;
  }

  .code-editor-area {
    display: block;
    width: 100%;
    min-height: 160px;
    padding: 12px;
    border: none;
    outline: none;
    resize: vertical;
    background: transparent;
    color: var(--foreground);
    font-family: var(--font-mono, monospace);
    font-size: 13px;
    line-height: 1.6;
    font-variant-numeric: tabular-nums;
  }
</style>
