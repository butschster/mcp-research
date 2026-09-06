<template>
  <fieldset class="share-fieldset">
    <legend class="field-label">What the link shows</legend>
    <p class="modal-help">Sections, documents and cross-references &middot; always</p>
    <label class="check-row"><input type="checkbox" :checked="modelValue.roadmaps" @change="set('roadmaps', $event)" /> Roadmaps</label>
    <label class="check-row"><input type="checkbox" :checked="modelValue.sessions" @change="set('sessions', $event)" /> Interview sessions, with questions and answers</label>
    <label class="check-row"><input type="checkbox" :checked="modelValue.tasks" @change="set('tasks', $event)" /> Tasks</label>
    <label class="check-row"><input type="checkbox" :checked="modelValue.export" @change="set('export', $event)" /> Downloading the project as a file</label>
  </fieldset>
</template>

<script setup lang="ts">
import type { ShareInclude } from '~/composables/useShare'
/**
 * The four include flags, as checkboxes.
 *
 * One component for the create form and the edit form: two hand-maintained
 * copies of four checkboxes is how one of them ends up describing a flag the
 * other has renamed. The words here are the words `shareContents` uses on the
 * row and in the visitor's banner.
 */
const props = defineProps<{ modelValue: ShareInclude }>()
const emit = defineEmits<{ 'update:modelValue': [value: ShareInclude] }>()

function set(key: keyof ShareInclude, e: Event) {
  emit('update:modelValue', { ...props.modelValue, [key]: (e.target as HTMLInputElement).checked })
}
</script>

<style scoped>
/* `.check-row` is the product's checkbox line (system.css); a scoped copy here
   would put the include rows on a different beat from the password row that
   sits under them in the same dialog. */
.share-fieldset { border: none; padding: 0; margin: var(--space-4) 0; font-size: var(--type-sm); }
</style>
