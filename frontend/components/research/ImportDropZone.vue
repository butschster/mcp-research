<template>
  <div
    class="drop-zone"
    @dragenter.prevent="onEnter"
    @dragover.prevent="onOver"
    @dragleave.prevent="onLeave"
    @drop.prevent="onDrop"
  >
    <slot />

    <!-- At rest the pane looks exactly as it did. A dashed box standing there
         permanently turns a section into a file-upload screen, and this is a
         section. -->
    <Transition name="drop-fade">
      <div v-if="armed" class="drop-overlay" aria-hidden="true">
        <div class="drop-label">
        <p class="drop-headline">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12M7 10l5 5 5-5M4 20h16"/></svg>
          Drop to read it into {{ sectionName }}
        </p>
        <p class="drop-sub">One {{ extensions.join(" or ") }} file, up to {{ maxKb }} KB</p>
        </div>
      </div>
    </Transition>

    <input
      ref="picker"
      type="file"
      class="sr-only"
      tabindex="-1"
      aria-hidden="true"
      :accept="acceptAttr"
      @change="onPick"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * The drop target for one markdown file, and the picker for people who do not
 * drag.
 *
 * Nothing in the catalog does drag-and-drop of files: KanbanBoard's drag is
 * internal HTML5 DnD between columns, with no `DataTransfer.files` and no
 * arming. `pages/index.vue` does have a hidden file input for the JSON research
 * import — about eight lines of plumbing, shared with this and not worth
 * extracting until something needs it a third time.
 *
 * The three client-side checks here are courtesy, not enforcement — the server
 * refuses all three again. They exist so a 40 MB mis-drop is answered instantly
 * instead of after a minute of upload.
 */
const props = defineProps<{
  /** False for a viewer, and for the all-entries mode where there is no target. */
  enabled: boolean
  sectionName: string
  maxBytes: number
  /** From the server, so the hint and the check cannot drift from it. */
  extensions: readonly string[]
  reading: boolean
}>()

const emit = defineEmits<{
  file: [File]
  refuse: [{ message: string; retryable: boolean }]
}>()

const picker = ref<HTMLInputElement | null>(null)
const over = ref(false)
// A bare `dragleave` fires on every child boundary the pointer crosses, so the
// overlay strobes its way across a list of forty entries. Counting enter and
// leave is the standard answer and the only one that holds for nested content.
let depth = 0

const maxKb = computed(() => Math.round(props.maxBytes / 1024))
const acceptAttr = computed(() => [...props.extensions, 'text/markdown'].join(','))
const extPattern = computed(
  () => new RegExp(`(${props.extensions.map((e) => e.replace('.', '\\.')).join('|')})$`, 'i'),
)
const armed = computed(() => over.value && props.enabled)

function isFileDrag(e: DragEvent): boolean {
  return Array.from(e.dataTransfer?.types ?? []).includes('Files')
}

function onEnter(e: DragEvent) {
  if (!isFileDrag(e)) return
  depth += 1
  over.value = true
}

function onOver(e: DragEvent) {
  if (!isFileDrag(e)) return
  // Without this the browser navigates to the dropped file and replaces the
  // whole app — which is why the handler runs for a viewer too, and answers
  // with a refusal instead of a control they do not have.
  if (e.dataTransfer) e.dataTransfer.dropEffect = props.enabled ? 'copy' : 'none'
}

function onLeave(e: DragEvent) {
  if (!isFileDrag(e)) return
  depth = Math.max(0, depth - 1)
  if (depth === 0) over.value = false
}

function onDrop(e: DragEvent) {
  depth = 0
  over.value = false
  if (!isFileDrag(e)) return
  if (!props.enabled) {
    emit('refuse', { message: 'You have read-only access to this research.', retryable: false })
    return
  }
  accept(Array.from(e.dataTransfer?.files ?? []))
}

function onPick(e: Event) {
  const input = e.target as HTMLInputElement
  accept(Array.from(input.files ?? []))
  // Cleared so picking the same file twice in a row fires `change` the second
  // time — which is exactly what somebody does after fixing it in their editor.
  input.value = ''
}

function accept(files: File[]) {
  if (props.reading) return
  if (files.length === 0) return
  if (files.length > 1) {
    emit('refuse', {
      message: `One file at a time. You dropped ${files.length}.`,
      retryable: false,
    })
    return
  }
  const f = files[0]!
  if (!extPattern.value.test(f.name)) {
    emit('refuse', {
      message: `${f.name} is not a markdown file. Only ${props.extensions.join(' and ')} can be imported — rename it if it really is markdown.`,
      retryable: false,
    })
    return
  }
  if (f.size > props.maxBytes) {
    emit('refuse', {
      message: `${f.name} is ${Math.round(f.size / 1024)} KB, over the ${maxKb.value} KB limit.`,
      retryable: false,
    })
    return
  }
  emit('file', f)
}

/** The section header's button and the empty state both open the picker. */
defineExpose({ open: () => picker.value?.click() })
</script>

<style scoped>
.drop-zone { position: relative; }

.drop-overlay {
  position: absolute;
  /* Outside the pane's own padding, so the target reads as the whole column
     rather than as a box drawn inside it. */
  inset: calc(var(--space-3) * -1);
  /* Must stay BELOW --z-elevated: `.app-nav` is sticky at that level and this
     element is later in the tree, so at an equal z-index it wins on tree order
     and paints over the navigation. Beating the cards underneath is all it
     needs, and they are at `auto`. */
  z-index: var(--z-in-page);
  /* The list underneath stays visible and stays untouchable: a target that
     swallows the pointer is a target that eats the drop it was drawn for. */
  pointer-events: none;
  display: flex;
  flex-direction: column;
  /* Top-aligned, with the label stuck to the viewport below — see .drop-label. */
  align-items: center;
  justify-content: flex-start;
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-lg);
  background: var(--color-primary-muted);
}
/* Opacity only. The pane is the size of the page, and a box that large moving
   or scaling reads as a glitch rather than as an invitation. A <Transition>
   rather than a bare `transition`, because the element is `v-if` and a property
   nothing ever changes animates nothing. */
.drop-fade-enter-active,
.drop-fade-leave-active { transition: opacity var(--transition-fast); }
.drop-fade-enter-from,
.drop-fade-leave-to { opacity: 0; }

/* The pane can be four thousand pixels tall; the viewport cannot. Sticking the
   label to the viewport is what keeps "Drop to read it into Findings" on screen
   for a section with forty entries, where centring it in the pane put it
   fifteen hundred pixels below the fold. */
.drop-label {
  position: sticky;
  top: 35vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-6) 0;
}

.drop-headline {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  margin: 0;
  font-size: var(--type-base);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
}
.drop-sub {
  margin: 0;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
</style>
