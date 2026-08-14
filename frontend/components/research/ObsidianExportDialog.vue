<template>
  <ModalOverlay :visible="visible" size="sm" labelledby="obsidian-export-title" @close="$emit('close')">
    <div class="dialog">
      <h3 id="obsidian-export-title" class="dialog-title">Obsidian export options</h3>

      <ul class="option-list">
        <li v-for="opt in options" :key="opt.key" class="option">
          <label class="option-label" :class="{ 'is-empty': opt.empty }">
            <input
              type="checkbox"
              :checked="state[opt.key]"
              :disabled="opt.empty"
              @change="toggle(opt.key, ($event.target as HTMLInputElement).checked)"
            />
            <span class="option-text">{{ opt.label }}</span>
            <span v-if="opt.count !== undefined" class="option-count">{{ opt.empty ? 'none' : opt.count }}</span>
          </label>
          <p v-if="opt.hint" class="option-hint">{{ opt.hint }}</p>
        </li>
      </ul>

      <div class="dialog-actions">
        <button class="btn btn-sm" @click="$emit('close')">Cancel</button>
        <button class="btn btn-sm btn-primary" @click="$emit('download', { ...state })">Download</button>
      </div>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
import type { ObsidianExportOptions } from '~/composables/useObsidianExport'
import { defaultObsidianOptions } from '~/composables/useObsidianExport'

/**
 * What goes into the vault, for the minority who want to change it.
 *
 * Options are deliberately **not** remembered between openings. A setting that
 * silently persists turns "why does my vault have no Sessions folder?" into an
 * unanswerable question, which is worse than re-ticking a box twice a year.
 */
const props = defineProps<{
  visible: boolean
  counts?: { sessions?: number; tasks?: number; roadmaps?: number }
}>()

defineEmits<{ download: [opts: ObsidianExportOptions]; close: [] }>()

const state = ref<ObsidianExportOptions>(defaultObsidianOptions())

// Every opening starts from the defaults, including one that follows a
// cancelled dialog.
watch(() => props.visible, (open) => {
  if (open) state.value = defaultObsidianOptions()
})

type Key = keyof ObsidianExportOptions

const options = computed(() => {
  const c = props.counts ?? {}
  return [
    { key: 'sessions' as Key, label: 'Sessions', count: c.sessions, empty: c.sessions === 0 },
    { key: 'tasks' as Key, label: 'Tasks', count: c.tasks, empty: c.tasks === 0 },
    { key: 'roadmaps' as Key, label: 'Roadmaps', count: c.roadmaps, empty: c.roadmaps === 0 },
    { key: 'html' as Key, label: 'Interactive HTML files', count: undefined, empty: false },
    {
      key: 'revisions' as Key,
      label: 'Revision history',
      count: undefined,
      empty: false,
      hint: 'A note per entry with who changed what, and when.',
    },
  ]
})

function toggle(key: Key, value: boolean) {
  state.value = { ...state.value, [key]: value }
}
</script>

<style scoped>
.dialog-title {
  font-size: var(--type-base);
  font-weight: 600;
  margin: 0 0 var(--space-4);
}
.option-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
.option-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  cursor: pointer;
}
/* An option with nothing behind it stays visible so the set does not change
   shape from one research to the next — it just cannot be acted on. */
.option-label.is-empty {
  color: var(--color-text-muted);
  cursor: default;
}
.option-text { flex: 1; }
.option-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}
.option-hint {
  margin: var(--space-1) 0 0 calc(13px + var(--space-2));
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.4;
}
.dialog-actions {
  display: flex;
  gap: var(--space-3);
  justify-content: flex-end;
  margin-top: var(--space-6);
}
</style>
