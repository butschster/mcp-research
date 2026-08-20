<template>
  <section v-if="groups.length" class="b-transcript">
    <div v-if="title" class="b-tr-head">
      <h3 class="b-tr-title">{{ title }}</h3>
      <span class="b-tr-count" data-no-mark>{{ turns.length }} {{ turns.length === 1 ? 'turn' : 'turns' }}</span>
    </div>

    <p v-if="truncated" class="b-tr-cut" role="status">
      {{ truncated }} further {{ truncated === 1 ? 'turn was' : 'turns were' }} not stored — this
      transcript is past the cap, and ends before the conversation did.
    </p>

    <div class="b-tr-groups">
      <div
        v-for="g in headGroups"
        :key="g.key"
        class="b-tr-group"
        :style="hueStyle(g.speaker)"
      >
        <p v-if="g.speaker" class="b-tr-speaker">
          <span class="b-tr-name">{{ g.speaker }}</span>
          <span v-if="g.lines[0]?.ts" class="b-tr-ts" data-no-mark>{{ g.lines[0]!.ts }}</span>
        </p>
        <p
          v-for="(line, i) in g.lines"
          :key="i"
          class="b-tr-line"
        ><span v-if="line.ts && !(i === 0 && g.speaker)" class="b-tr-ts b-tr-ts-inline" data-no-mark>{{ line.ts + ' ' }}</span><span v-html="inline(line.text)"></span></p>
      </div>
    </div>

    <!-- The tail is in the DOM even when folded, so find-in-page can reach it.
         `hidden="until-found"` lets the browser reveal it and tell us it did;
         where it is unsupported the attribute degrades to a plain `hidden`,
         which is the behaviour a fold would have had anyway. -->
    <div
      v-if="foldable"
      :id="tailId"
      ref="tail"
      class="b-tr-rest"
      :hidden="expanded ? undefined : ('until-found' as any)"
      @beforematch="onBeforeMatch"
    >
      <div class="b-tr-groups">
        <div
          v-for="(g, gi) in tailGroups"
          :key="g.key"
          class="b-tr-group"
          :style="hueStyle(g.speaker)"
          :tabindex="gi === 0 ? -1 : undefined"
          :ref="gi === 0 ? setFirstTail : undefined"
        >
          <p v-if="g.speaker" class="b-tr-speaker">
            <span class="b-tr-name">{{ g.speaker }}</span>
            <span v-if="g.lines[0]?.ts" class="b-tr-ts" data-no-mark>{{ g.lines[0]!.ts }}</span>
          </p>
          <p
            v-for="(line, i) in g.lines"
            :key="i"
            class="b-tr-line"
          ><span v-if="line.ts && !(i === 0 && g.speaker)" class="b-tr-ts b-tr-ts-inline" data-no-mark>{{ line.ts + ' ' }}</span><span v-html="inline(line.text)"></span></p>
        </div>
      </div>
    </div>

    <button
      v-if="foldable"
      ref="toggle"
      type="button"
      class="btn btn-sm b-tr-more"
      :aria-expanded="expanded"
      :aria-controls="tailId"
      @click="toggleFold"
    >
      {{ expanded ? 'Show fewer turns' : `Show all ${turns.length} turns` }}
    </button>
  </section>
</template>

<script setup lang="ts">
/**
 * A conversation held outside this tool, as structured turns.
 *
 * A component rather than a branch in BlockRenderer because it owns fold state
 * per instance — the line this repository draws is that a block which is markup
 * is a branch, and a block with state or network calls of its own is a
 * component.
 *
 * Two layout decisions worth not relitigating:
 *
 *   - **Name above an indented run, not a speaker column.** A 9rem column has no
 *     honest answer for `Александра Константиновна`: it wraps to three lines or
 *     truncates, and truncating the speaker destroys the one fact a transcript
 *     exists to carry. Consecutive turns by the same person are grouped under
 *     one name, which recovers most of a column's scan value at none of its cost
 *     — and there is one layout to build, test and print instead of two.
 *   - **Speaker colour is assigned by order of first appearance, not hashed.**
 *     Hashing the name (as tags do) collides for two speakers about one time in
 *     six, and two speakers in one colour is worse than no colour: by then the
 *     reader has learned to trust the hue. The map is computed once per document
 *     so the same person keeps their colour across two transcripts in one entry.
 */
import { renderInline } from '~/composables/useInlineMarkdown'

interface Turn {
  speaker?: string
  text: string
  ts?: string
}

const props = withDefaults(
  defineProps<{
    /** The block's stable id; also what data-block-id lands on via fallthrough. */
    blockId?: string
    title?: string
    turns?: Turn[]
    /** Research short code, so [[E3]] inside a turn resolves. */
    researchSlug?: string
    /** speaker name → 1..6, computed across the whole document by the renderer. */
    speakerHues?: Record<string, number>
    /**
     * Kept open regardless of length. The entry page sets it for a transcript
     * that holds an annotation: a mark inside a folded tail measures at zero and
     * its pin lands above the card.
     */
    forceExpanded?: boolean
    /** Turns the server had to drop. A cap that says nothing leaves a document
     *  that reads as a complete call and is not one. */
    truncated?: number
  }>(),
  {
    blockId: '', title: '', turns: () => [], researchSlug: '',
    speakerHues: () => ({}), forceExpanded: false, truncated: 0,
  }
)

/**
 * Fold past 30 — but only when there are more than 36 turns.
 *
 * The hysteresis is the point: without it a 32-turn transcript hides two turns
 * behind a button, which costs the reader a click to save six lines.
 */
const FOLD_AT = 30
const FOLD_OVER = 36

function inline(text: string): string {
  return renderInline(text, props.researchSlug)
}

interface Group {
  key: string
  speaker: string
  lines: Turn[]
}

/**
 * Consecutive turns by the same speaker become one group.
 *
 * A turn with no speaker continues the group above it rather than starting one:
 * `normTranscript` allows an unattributed line, and inventing "Speaker 1" would
 * assert something the document does not say.
 */
function group(turns: Turn[], offset: number): Group[] {
  const out: Group[] = []
  turns.forEach((t, i) => {
    const speaker = t.speaker || ''
    const last = out[out.length - 1]
    if (last && (speaker === '' || speaker === last.speaker)) {
      last.lines.push(t)
      return
    }
    out.push({ key: `${offset + i}:${speaker}`, speaker, lines: [t] })
  })
  return out
}

const foldable = computed(() => props.turns.length > FOLD_OVER)
// `forceExpanded` opens the fold; it does not remove it. Removing it meant a
// 500-turn transcript holding one mark could not be collapsed at all, and that
// dismissing the last mark collapsed 470 turns out from under the reader with
// no gesture from them. Opening is one-way: the reader closes it or nobody does.
const expanded = ref(props.forceExpanded)
watch(() => props.forceExpanded, (v) => {
  if (v) expanded.value = true
})

// Everything is one list when nothing folds; the split only exists to give the
// tail its own container.
const headTurns = computed(() => (foldable.value ? props.turns.slice(0, FOLD_AT) : props.turns))
const tailTurns = computed(() => (foldable.value ? props.turns.slice(FOLD_AT) : []))

const groups = computed(() => group(props.turns, 0))
const headGroups = computed(() => group(headTurns.value, 0))
const tailGroups = computed(() => group(tailTurns.value, FOLD_AT))

// A stored block always has an id; a hand-written payload or a story may not,
// and two of those sharing `tr-rest-x` would point one toggle at the other's
// tail.
const fallbackId = useId()
const tailId = computed(() => `tr-rest-${props.blockId || fallbackId}`)

function hueStyle(speaker: string) {
  const n = speaker ? props.speakerHues[speaker] : undefined
  return n ? { '--tr-hue': `var(--hue-${n})` } : {}
}

const toggle = ref<HTMLElement | null>(null)
const tail = ref<HTMLElement | null>(null)
const firstTailGroup = ref<HTMLElement | null>(null)
function setFirstTail(el: any) {
  firstTailGroup.value = (el as HTMLElement) || null
}

function onBeforeMatch() {
  // The browser is revealing the tail to show a find-in-page hit. Follow it
  // rather than fight it, or the next render hides the match again.
  expanded.value = true
}

async function toggleFold() {
  const opening = !expanded.value
  expanded.value = opening
  await nextTick()
  if (opening) {
    // The button has just been pushed down by everything that appeared above it.
    // Leaving focus on it strands a keyboard reader off-screen.
    firstTailGroup.value?.focus()
  } else {
    toggle.value?.focus()
    // Collapsing removes content above the button, so without this the page
    // appears to jump to somewhere the reader was not.
    tail.value?.parentElement?.scrollIntoView({ block: 'start' })
  }
}
</script>

<style scoped>
.b-transcript { display: flex; flex-direction: column; }

.b-tr-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: var(--space-3);
  padding-bottom: var(--space-2);
  margin-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border);
}
.b-tr-title {
  font-size: var(--type-md);
  font-weight: var(--weight-semibold);
  margin: 0;
  overflow-wrap: anywhere;
}
.b-tr-count {
  font-size: var(--type-2xs);
  color: var(--color-text-muted);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.b-tr-cut {
  margin: 0 0 var(--space-4);
  font-size: var(--type-2xs);
  color: var(--color-warning);
}

.b-tr-groups { display: flex; flex-direction: column; gap: var(--space-4); }
.b-tr-group { break-inside: avoid; }
.b-tr-group:focus { outline: none; }
.b-tr-group:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 4px; border-radius: var(--radius-sm); }

.b-tr-speaker {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--space-2);
  margin: 0 0 var(--space-1);
}
.b-tr-name {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  /* Falls back to the ordinary text colour, so a speaker the map does not know
     is quiet rather than invisible. */
  color: var(--tr-hue, var(--color-text));
  overflow-wrap: anywhere;
}
.b-tr-ts {
  font-size: var(--type-2xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
/* The space is inside the text, not only in the margin: `data-no-mark` keeps the
   stamp out of the annotation haystack, but Ctrl+C does not honour it, and
   `00:03:40And the gateway stays open.` is a broken quote out of a block whose
   job is quoting. `pre-wrap` preserves it. */
.b-tr-ts-inline { margin-right: var(--space-1); }

.b-tr-line {
  /* The indent is what makes this read as a column without being one. */
  padding-left: var(--space-4);
  margin: 0;
  line-height: var(--line-base);
  /* A turn's own line breaks survive normalization and mean something here. */
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.b-tr-line + .b-tr-line { margin-top: var(--space-2); }

.b-tr-rest { margin-top: var(--space-4); }
.b-tr-more { align-self: flex-start; margin-top: var(--space-4); }

@media print {
  /* A folded tail in print is content the reader paid for and did not get.
     `hidden="until-found"` is NOT `display: none` in a browser that supports it
     — the UA sheet gives it `content-visibility: hidden` instead — so undoing
     only the display would have printed 30 turns of 500 in exactly the browser
     the fold is built for. Both, and neither needs !important. */
  .b-tr-rest[hidden] { display: block; content-visibility: visible; }
  .b-tr-more { display: none; }
  /* The hue prints as an indistinguishable grey: a distinction that exists only
     on screen is worse than none. */
  .b-tr-name { color: #111; }
  .b-tr-group { break-inside: avoid; }
}
</style>
