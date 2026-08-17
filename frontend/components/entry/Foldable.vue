<template>
  <div class="crossrefs-block card no-print">
    <button
      class="foldable-header"
      :aria-expanded="open"
      :aria-controls="bodyId"
      @click="toggle"
    >
      <span class="crossrefs-title foldable-title">
        <slot name="icon" />
        {{ title }}
        <span v-if="count" class="foldable-count">{{ count }}</span>
      </span>
      <svg
        class="foldable-chevron" :class="{ 'foldable-chevron--open': open }"
        width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
        stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"
      >
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </button>

    <div v-if="open" :id="bodyId" class="foldable-body">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * A panel below an entry that is closed until somebody asks for it.
 *
 * Cross-references, external links and related-by-tags used to render open, one
 * under the other. On an entry with a few of each they added more height than
 * the entry itself, and they are consulted occasionally rather than read — so
 * the default is closed and the header says how many are inside, which is
 * usually the whole answer.
 *
 * The choice is remembered per panel, for the browser, because somebody who
 * does use references should not have to reopen the same panel on every entry
 * they visit. `rememberAs` is the key; two panels sharing one would share the
 * state.
 */
const props = defineProps<{
  title: string
  /** Shown beside the title, so a closed panel still says whether it is empty. */
  count?: number
  /** localStorage key suffix. Omit to always start closed. */
  rememberAs?: string
}>()

const bodyId = `foldable-${useId()}`
const open = ref(false)

const storageKey = computed(() => (props.rememberAs ? `entry_panel_${props.rememberAs}` : ''))

onMounted(() => {
  if (!storageKey.value) return
  open.value = localStorage.getItem(storageKey.value) === 'open'
})

function toggle() {
  open.value = !open.value
  if (storageKey.value) {
    localStorage.setItem(storageKey.value, open.value ? 'open' : 'closed')
  }
}
</script>

<style scoped>
/* The panel's own box and title.
   The three consumers each define `.crossrefs-block` in their scoped block, and
   a parent's scoped rule does reach a child component's root — but `.crossrefs-
   title` sits inside this component, out of that reach, so it has to be stated
   here or the header loses its weight and spacing. */
.crossrefs-block {
  margin-top: var(--space-6);
  padding: var(--space-6);
  border-radius: var(--radius-lg);
}
.crossrefs-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  margin: 0;
}

.foldable-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  width: 100%;
  background: none;
  border: none;
  padding: 0;
  color: inherit;
  font: inherit;
  cursor: pointer;
  text-align: left;
}
.foldable-title { margin-bottom: 0; }
.foldable-count {
  font-size: var(--type-xs);
  font-weight: var(--weight-medium);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.05rem 0.35rem;
  border-radius: var(--radius-xs);
  font-variant-numeric: tabular-nums;
}
.foldable-chevron {
  flex: none;
  color: var(--color-text-muted);
  transition: transform var(--transition-fast);
}
.foldable-chevron--open { transform: rotate(180deg); }
.foldable-header:hover .foldable-chevron { color: var(--color-text); }
.foldable-body { margin-top: var(--space-4); }
</style>
