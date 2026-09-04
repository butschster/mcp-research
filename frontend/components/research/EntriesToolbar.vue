<template>
  <!--
    One row: the search box, the active tag, a handful of frequent tags, and a
    button that opens the rest.

    This replaces a tag cloud that stood above the list and wrapped as far as it
    needed to. In a research with many sections that was a hundred chips and half
    a screen before the first document. A cloud does two jobs — show what the
    research is about, and filter it — and at a hundred tags it has already
    stopped doing the first, so what is left is a filter, and a filter belongs
    in a toolbar.

    The row's height is the input's. Chips arriving after the entries land cause
    no layout shift, which is why there is no skeleton here.
  -->
  <!-- `no-print` sits on the controls, not the root: a printed page of
       seventeen documents out of ninety-two needs the chip that says why. -->
  <div ref="root" :class="['etb', 'cluster', { 'etb--searching': query.trim().length > 1 }]">
    <!-- The counter sits inside the field, over its right end, not beside it
         in the row: as a flex sibling it made the row wrap on the second
         keystroke of every search, and everything below jumped 38px. The
         field is the one element here with room to spare. -->
    <div class="etb-search no-print">
      <!-- `v-model`, not `:value` + `@input`: the directive is what holds the
           value still while an IME composes, so two romaji keystrokes do not
           become a search for text nobody submitted. -->
      <input
        v-model="queryModel"
        type="search"
        class="text-input etb-input"
        :placeholder="searchPlaceholder"
        :aria-label="searchLabel"
      />
      <span v-if="$slots.meta" class="etb-meta"><slot name="meta" /></span>
    </div>

    <!-- The active tag has one fixed place, right after the input, and leaves
         the quick row while it is on. With a popover in play the active tag is
         usually *not* one of the six most frequent, so a pressed chip in place
         would sometimes answer "what is filtering this list" and sometimes be
         nowhere on screen. One reflow beats one hunt. -->
    <!-- The count stays on the chip. Clicking `security 23` used to turn it
         into `security ×`, and the number — the size of the list now on
         screen, which nothing else on the pane states — was gone. -->
    <button
      v-if="modelValue"
      ref="activeChip"
      type="button"
      :class="['tag', `tag-hue-${tagHue(modelValue)}`, 'etb-chip', 'etb-chip--active']"
      :aria-label="`Clear tag filter: ${modelValue}`"
      :title="modelValue"
      @click="clear"
    ><span class="tag-text">{{ modelValue }}</span><span v-if="countMap[modelValue]" class="tag-count">{{ countMap[modelValue] }}</span><span class="etb-chip-x" aria-hidden="true">&times;</span></button>

    <TagList
      v-if="quickTags.length"
      ref="quickRow"
      class="etb-quick no-print"
      :tags="quickTags"
      :counts="countMap"
      clickable
      @tag-click="apply"
    />

    <!-- Always there when there are tags at all, even four. Which chips are
         visible is a fact about pixel width, so a button whose existence
         followed it would blink in and out as the pane resizes — and it is the
         only keyboard route to a tag the quick row does not show. -->
    <button
      v-if="sorted.length"
      ref="trigger"
      type="button"
      class="btn etb-more no-print"
      aria-haspopup="listbox"
      :aria-expanded="open ? 'true' : 'false'"
      :aria-controls="open && options.length ? listId : undefined"
      @click="open ? close() : openPanel()"
    >
      Tags
      <span class="btn-count">{{ sorted.length }}</span>
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="6 9 12 15 18 9"/></svg>
    </button>

    <!-- A combobox popup: not a dialog (no focus trap is promised) and not a
         menu (no menuitems). The filter input is the combobox and the list is
         its listbox, so a screen reader hears the arrow-key model this actually
         implements. Same doctrine ActionMenu wrote down about not announcing a
         widget you have not built. -->
    <!-- Enter animates through a keyframe on the panel itself; there is no
         leave animation on purpose. A <Transition> kept the panel in the DOM
         for 150ms after it closed, long enough for a Tab out of the filter box
         to land on the "Show all" button and then fall to <body> when the panel
         went. -->
    <div v-if="open" class="etb-panel no-print">
        <input
          ref="filterInput"
          v-model="filter"
          type="text"
          class="text-input etb-filter"
          role="combobox"
          :aria-expanded="options.length ? 'true' : 'false'"
          aria-autocomplete="list"
          aria-label="Filter tags"
          placeholder="Filter tags…"
          :aria-controls="options.length ? listId : undefined"
          :aria-activedescendant="options.length ? optionId(cursor) : undefined"
          @keydown="onKeydown"
          @input="cursor = 0"
        />
        <ul v-if="options.length" :id="listId" ref="listEl" role="listbox" aria-label="Tags" class="etb-list">
          <li
            v-for="(o, i) in options"
            :id="optionId(i)"
            :key="o.tag"
            role="option"
            :aria-selected="o.tag === modelValue ? 'true' : 'false'"
            :class="['etb-option', { 'is-cursor': i === cursor }]"
            :title="o.tag"
            @mousemove="cursor = i"
            @mousedown.prevent
            @click="choose(o.tag)"
          >
            <span class="etb-option-check" aria-hidden="true">{{ o.tag === modelValue ? '✓' : '' }}</span>
            <span class="etb-option-text">{{ o.tag }}</span>
            <span class="etb-option-count">{{ o.count }}</span>
          </li>
        </ul>
        <div v-else class="etb-none">
          <span>No tag matches “{{ filter }}”.</span>
          <button type="button" class="btn btn-sm" @click="showAll">Show all {{ snapshot.length }} tags</button>
        </div>
      </div>
    <!-- Mounted for the life of the toolbar; the no-match line above appears
         on the keystroke that empties the list, which is when a live region
         is least likely to be announced. -->
    <p class="sr-only" role="status">{{ open && !options.length ? `No tag matches “${filter}”.` : '' }}</p>
  </div>
</template>

<script lang="ts">
// Module-level, so two toolbars on one page (the story catalogue, a future
// split view) do not hand the same id to two listboxes.
let instances = 0
</script>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { tagHue } from '~/composables/useTagHue'

const props = withDefaults(
  defineProps<{
    /** Every tag on the surface with how many entries carry it. Any order. */
    tags: Array<{ tag: string; count: number }>
    /** The active tag, `''` for none. */
    modelValue: string
    query: string
    /** aria-label for the search input — names the research or section. */
    searchLabel: string
    searchPlaceholder?: string
    /** How many tags sit in the row as one-click chips. */
    quickCount?: number
  }>(),
  { searchPlaceholder: 'Search this research…', quickCount: 6 },
)

const emit = defineEmits<{
  'update:modelValue': [string]
  'update:query': [string]
}>()

const queryModel = computed({
  get: () => props.query,
  set: (v: string) => emit('update:query', v),
})

/* Sorted here rather than trusted from the endpoint, because section mode
   derives its counts client-side and both modes must agree on the order. Same
   order the endpoint uses: most entries first, then alphabetical. */
const sorted = computed(() =>
  [...props.tags].sort((a, b) => b.count - a.count || a.tag.localeCompare(b.tag)),
)
const countMap = computed(() => Object.fromEntries(sorted.value.map(t => [t.tag, t.count])))

/* A fixed six chip-shaped things. A count threshold shows nothing in the
   common research where every tag has count 1; "as many as fit" needs a
   measurement and changes the row as the pane resizes. The active chip counts
   as one of the six: pulling rank seven in to keep six *quick* chips made the
   row one chip wider than it was a moment ago, and at the width this pane
   actually gets that one chip was what pushed the Tags button onto a second
   line, alone. */
const quickTags = computed(() =>
  sorted.value
    .filter(t => t.tag !== props.modelValue)
    .slice(0, Math.max(props.quickCount - (props.modelValue ? 1 : 0), 0))
    .map(t => t.tag),
)

/* Applying a chip removes the button that was pressed: the tag leaves the
   quick row and reappears as the active chip. Left alone, focus fell to <body>
   and a keyboard user had to Tab to find out where they were. So focus follows
   the tag to its new chip; clearing hands it to the quick row, which is where
   the next filter is. */
const activeChip = ref<HTMLButtonElement | null>(null)
const quickRow = ref<{ $el: HTMLElement } | null>(null)

function apply(tag: string) {
  emit('update:modelValue', tag)
  nextTick(() => activeChip.value?.focus())
}

function clear() {
  emit('update:modelValue', '')
  nextTick(() => {
    const first = quickRow.value?.$el?.querySelector<HTMLElement>('button')
    ;(first ?? root.value?.querySelector<HTMLElement>('.etb-input'))?.focus()
  })
}

/* --- Popover ------------------------------------------------------------- */

const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const filterInput = ref<HTMLInputElement | null>(null)
const listEl = ref<HTMLElement | null>(null)

const open = ref(false)
const filter = ref('')
const cursor = ref(0)

/* The list is frozen while the popover is open. Entries and tags refetch on
   WebSocket events, and a list that re-sorts under the cursor is how a keyboard
   user applies a tag they did not choose. The popover filters the snapshot; it
   re-syncs on the next open. */
const snapshot = ref<Array<{ tag: string; count: number }>>([])
const options = computed(() => {
  const q = filter.value.trim().toLowerCase()
  return q ? snapshot.value.filter(t => t.tag.toLowerCase().includes(q)) : snapshot.value
})

const listId = `etb-list-${++instances}`
function optionId(i: number) { return `${listId}-opt-${i}` }

function openPanel() {
  snapshot.value = sorted.value
  filter.value = ''
  const applied = snapshot.value.findIndex(t => t.tag === props.modelValue)
  cursor.value = applied >= 0 ? applied : 0
  open.value = true
  nextTick(() => {
    filterInput.value?.focus()
    scrollCursorIntoView()
  })
}

function close(returnFocus = false) {
  if (!open.value) return
  open.value = false
  if (returnFocus) trigger.value?.focus()
}

function choose(tag: string) {
  emit('update:modelValue', tag === props.modelValue ? '' : tag)
  close(true)
}

function showAll() {
  filter.value = ''
  cursor.value = 0
  filterInput.value?.focus()
}

// The cursor resets on typing — as an input handler, not a watcher on
// `filter`. A watcher also fired when `openPanel()` cleared the last visit's
// text, and ran before the render, so it put the cursor back on option 0 after
// `openPanel()` had just placed it on the applied tag.

function moveCursor(to: number) {
  if (!options.value.length) return
  cursor.value = Math.min(Math.max(to, 0), options.value.length - 1)
  scrollCursorIntoView()
}

function scrollCursorIntoView() {
  nextTick(() => {
    listEl.value?.querySelector<HTMLElement>(`#${optionId(cursor.value)}`)
      ?.scrollIntoView({ block: 'nearest' })
  })
}

function onKeydown(e: KeyboardEvent) {
  switch (e.key) {
    case 'ArrowDown': e.preventDefault(); moveCursor(cursor.value + 1); break
    case 'ArrowUp': e.preventDefault(); moveCursor(cursor.value - 1); break
    case 'Home': e.preventDefault(); moveCursor(0); break
    case 'End': e.preventDefault(); moveCursor(options.value.length - 1); break
    case 'Enter': {
      e.preventDefault()
      const o = options.value[cursor.value]
      if (o) choose(o.tag)
      break
    }
    // Escape always closes, even with text in the box. A two-stage "clear, then
    // close" was considered; ActionMenu set the house behaviour and one control
    // disagreeing with it costs more than one extra keystroke.
    case 'Escape': e.preventDefault(); close(true); break
    // Tab leaves, and the popup goes with it — to the trigger, explicitly. Left
    // to the browser, the move is resolved before Vue removes the panel, and in
    // the no-match state the next tabbable is the "Show all" button inside it:
    // focus landed there and fell to <body> when the panel went.
    case 'Tab': e.preventDefault(); close(true); break
  }
}

// mirrors ActionMenu — the same document-level dismissal, because two call
// sites is too thin a case for a composable. Revisit at the third popover.
function onDocumentClick(e: MouseEvent) {
  if (open.value && root.value && !root.value.contains(e.target as Node)) close(false)
}
function onDocumentKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) close(true)
}
onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onDocumentKeydown)
})
onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onDocumentKeydown)
})

defineExpose({
  /** Closes the popover without moving focus — for the owner, when the surface
   *  under the toolbar changes (another section, another mode). */
  close: () => close(false),
})
</script>

<style scoped>
/* Every control in this row is pinned to one height. `.text-input` has none of
   its own (padding plus line-height comes to ~38px) and `.btn-sm` is 26px, so
   the two side by side were 12px apart. `.btn` already states `--control-h`;
   the input and the chips are told the same number here. */
.etb {
  --etb-h: var(--control-h);
  --etb-list-max: 20rem;
  position: relative;
  margin-bottom: var(--space-4);
}
.etb-search { position: relative; display: flex; flex: 1; min-width: 12rem; }
/* With a counter over its right end the field needs room for the text too:
   at the bare minimum the two overlapped. */
.etb-search:has(.etb-meta:not(:empty)) { min-width: 18rem; }
.etb-input {
  width: 100%;
  height: var(--etb-h);
  padding-block: 0;
}
/* Room for the counter, and for the browser's own clear button at the very
   end of a search field. */
.etb-search:has(.etb-meta:not(:empty)) .etb-input { padding-right: 8.5rem; }
.etb-meta {
  position: absolute;
  right: 2rem;
  top: 50%;
  transform: translateY(-50%);
  white-space: nowrap;
  pointer-events: none;
}
.etb-chip { min-height: var(--etb-h); max-width: var(--tag-max); gap: var(--space-1); flex: none; }
.etb-chip--active {
  border: none;
  background: var(--color-primary-muted);
  color: var(--color-primary);
  cursor: pointer;
  font-family: inherit;
}
.etb-chip--active:hover .etb-chip-x { opacity: 1; }
.etb-chip-x { font-size: var(--type-sm); line-height: 1; opacity: 0.7; transition: opacity var(--transition-fast); }

/* The quick chips are TagList's, out of reach of this scope; :deep is how a
   parent states a layout fact about a child's boxes. */
.etb-quick { flex: none; }
.etb-quick :deep(.tag) { min-height: var(--etb-h); }
/* The auto margin keeps the button on the right edge whether or not the row
   wraps — and the panel below is anchored to that same edge, so the two never
   part. Without it a wrapped button sat alone at the far left of line two
   with its panel opening 360px to the right of it. */
.etb-more { flex: none; margin-left: auto; }

/* Anchored to the toolbar's right edge, which is where the button always is
   (above). It cannot overflow a 600px pane and needs no flip logic. */
.etb-panel {
  position: absolute;
  top: calc(100% + var(--space-2));
  right: 0;
  width: min(320px, 100%);
  padding: var(--space-2);
  background: var(--color-surface-raised);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-2);
  /* In-page, not overlay: `.etb` opens no stacking context, so this value
     competes with the sticky nav (`--z-elevated`), and `--z-overlay` painted
     over it once the page scrolled. Cards keep `z-index: auto` under their
     entrance animation, so any positive value clears them. */
  z-index: var(--z-in-page);
}
.etb-filter { height: var(--etb-h); padding-block: 0; margin-bottom: var(--space-2); }
.etb-list {
  list-style: none;
  margin: 0;
  padding: 0;
  max-height: var(--etb-list-max);
  overflow-y: auto;
}
.etb-option {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  /* Stated, like every other height in this row. `min-height` plus padding
     plus an inherited 1.55 line-height came out at 39px, never 30. */
  height: var(--etb-h);
  padding: 0 var(--space-3);
  line-height: 1;
  border-left: 2px solid transparent;
  border-radius: var(--radius-xs);
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  cursor: pointer;
}
.etb-option.is-cursor { background: var(--color-surface-hover); color: var(--color-text); border-left-color: var(--color-primary); }
.etb-option[aria-selected='true'] { color: var(--color-text); }
.etb-option-check { width: 1em; flex: none; color: var(--color-primary); }
.etb-option-text { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.etb-option-count { flex: none; color: var(--color-text-muted); font-size: var(--type-xs); font-variant-numeric: tabular-nums; }
.etb-none {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}
/* A pasted 60-character tag name in the filter box is echoed here. */
.etb-none span { min-width: 0; overflow-wrap: anywhere; }

@keyframes etb-pop-in {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: none; }
}
.etb-panel { animation: etb-pop-in 0.15s ease; }

/* While a query stands the list below is already results, so the row may
   change shape without anything jumping: three quick chips make room for the
   wider search field and the active chip. */
.etb--searching .etb-quick :deep(.tag:nth-child(n + 4)) { display: none; }

/* Fewer quick chips as the window narrows. The pane is the window minus a
   240px sidebar, so at a 1150px window six chips filled the line and the
   Tags button wrapped onto one of its own; at 1024 four did the same. Whole
   chips are removed rather than clipped, and they are all still in the
   popover, so nothing is lost. Window width, not pane width, because CSS
   can measure only the former without a script. */
@media (max-width: 1250px) {
  .etb-quick :deep(.tag:nth-child(n + 5)) { display: none; }
}
@media (max-width: 1100px) {
  .etb-quick :deep(.tag:nth-child(n + 4)) { display: none; }
}
@media (max-width: 768px) {
  .etb { --etb-h: var(--control-h-touch); }
  .etb-panel { width: 100%; }
}
</style>
