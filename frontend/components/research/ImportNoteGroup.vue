<template>
  <li class="note-group" :class="`tone-${tone}`">
    <button
      type="button"
      class="note-head"
      :aria-expanded="open"
      :aria-controls="bodyId"
      @click="open = !open"
    >
      <span v-if="tone === 'attention'" class="note-glyph" aria-hidden="true">!</span>
      <span class="note-label">{{ label }}</span>
      <span class="note-count count-chip">{{ items.length }}</span>
      <svg class="note-chevron" :class="{ 'is-open': open }" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m9 18 6-6-6-6"/></svg>
    </button>

    <div v-show="open" :id="bodyId" class="note-body">
      <dl class="note-list">
        <template v-for="(item, i) in items" :key="item.key + i">
          <dt class="note-key">
            <span v-if="item.kind" class="note-kind">{{ item.kind }}</span>
            <code>{{ item.key }}</code>
          </dt>
          <dd class="note-reason">
            <span v-if="item.value" class="note-value">“{{ item.value }}”</span>
            {{ item.reason }}
          </dd>
        </template>
      </dl>
    </div>
  </li>
</template>

<script setup lang="ts">
/**
 * One collapsible group in the import ledger.
 *
 * `entry/Foldable.vue` is the nearest thing and does not fit — but not because
 * of its memory: `rememberAs` is optional there and omitting it already starts
 * closed every time. It is the layout. Foldable is a `.card` with `--space-6`
 * of padding and a `--space-6` top margin, hardcoded, meant to stand under an
 * entry rather than be one row of five; its chevron points down, the panel
 * idiom, where a ledger row points right. Growing it a padding variant, a tone,
 * a default-open and a chevron direction is four props that never combine, on
 * three call sites that want none of them.
 */
const props = withDefaults(
  defineProps<{
    /** `attention` is something to act on; `note` is something to know. */
    tone?: 'attention' | 'note'
    label: string
    items: Array<{ kind?: string; key: string; value?: string; reason: string }>
    defaultOpen?: boolean
  }>(),
  { tone: 'note', defaultOpen: false },
)

const open = ref(props.defaultOpen)
const bodyId = useId()
</script>

<style scoped>
.note-group { list-style: none; }
.note-group + .note-group { border-top: 1px solid var(--color-border); }

.note-head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  width: 100%;
  /* Declared and actually binding. With the inherited 1.55 line-height the
     content alone was 23.25px and the min-height never applied, so the row
     height still emerged from the padding — which is the thing the declaration
     was there to prevent. */
  min-height: var(--control-h);
  line-height: 1.2;
  padding: var(--space-1) 0;
  background: none;
  border: none;
  text-align: left;
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  cursor: pointer;
}
.note-head:hover { color: var(--color-text); }
.note-head:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }

.tone-attention .note-head { color: var(--color-warning); }
.note-glyph {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--color-warning);
  color: var(--color-bg);
  font-size: 11px;
  font-weight: var(--weight-semibold);
}

.note-label { flex: 1; min-width: 0; }

/* `.count-chip` in system.css carries the look — it was the fifth copy of one
   seven-line rule when this arrived. Only the layout is local. */
.note-count { flex: none; }

.note-chevron { flex: none; transition: transform var(--transition-fast); }
.note-chevron.is-open { transform: rotate(90deg); }

.note-body { padding: 0 0 var(--space-3); }
.note-list {
  display: grid;
  grid-template-columns: minmax(0, auto) minmax(0, 1fr);
  gap: var(--space-1) var(--space-3);
  margin: 0;
  font-size: var(--type-xs);
}
.note-key {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  margin: 0;
}
.note-kind {
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-size: 10px;
  color: var(--color-text-faint);
}
.note-key code {
  font-family: 'JetBrains Mono', monospace;
  color: var(--color-text);
  overflow-wrap: anywhere;
}
.note-reason { margin: 0; color: var(--color-text-muted); overflow-wrap: anywhere; }
.note-value { color: var(--color-text); }

@media (max-width: 768px) {
  /* Two columns of thirty characters each is unreadable; the key goes above
     its reason, the same rule MetadataBlock uses at this width. */
  .note-list { grid-template-columns: minmax(0, 1fr); gap: 0; }
  .note-key { margin-top: var(--space-2); }
}
</style>
