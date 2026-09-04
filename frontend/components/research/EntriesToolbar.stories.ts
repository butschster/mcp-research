import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import EntriesToolbar from './EntriesToolbar.vue'

/**
 * The one-row search-and-tag filter above an entries list.
 *
 * It replaced a tag cloud that wrapped as far as it needed to — a hundred chips
 * in a research with many sections, half a screen before the first document.
 * Six most-frequent tags stay in the row as one-click chips; the rest live in a
 * popover with its own filter box. The active tag has one fixed place, right
 * after the input, and leaves the quick row while it is on.
 */
const meta: Meta<typeof EntriesToolbar> = {
  title: 'Research/EntriesToolbar',
  component: EntriesToolbar,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 800px; min-height: 420px"><story /></div>',
    }),
  ],
  // The toolbar is controlled: the owner holds both the query and the active
  // tag. The stories hold them here so the chips and the popover actually do
  // something when clicked.
  render: (args) => ({
    components: { EntriesToolbar },
    setup() {
      const tag = ref(args.modelValue ?? '')
      const query = ref(args.query ?? '')
      return { args, tag, query }
    },
    template: `
      <EntriesToolbar v-bind="args" v-model="tag" v-model:query="query">
        <template #meta>
          <span v-if="query.length === 1" class="card-meta">Keep typing…</span>
          <span v-else-if="query.length > 1" class="card-meta">7 matches</span>
        </template>
      </EntriesToolbar>
      <p class="card-meta" style="margin-top: 1rem">
        active tag: <strong>{{ tag || '—' }}</strong>, query: <strong>{{ query || '—' }}</strong>
      </p>
    `,
  }),
}
export default meta
type Story = StoryObj<typeof EntriesToolbar>

const few = [
  { tag: 'vue', count: 4 },
  { tag: 'composables', count: 2 },
  { tag: 'slots', count: 1 },
  { tag: 'performance', count: 1 },
]

/* A hundred and twenty tags with a realistic shape: a handful that are used a
   lot, a long tail used once. Sorting them is the component's job. */
const many = (() => {
  const head = [
    { tag: 'security', count: 12 }, { tag: 'auth', count: 9 }, { tag: 'api', count: 7 },
    { tag: 'database', count: 4 }, { tag: 'migration', count: 4 }, { tag: 'testing', count: 3 },
    { tag: 'frontend', count: 3 }, { tag: 'deployment', count: 2 },
  ]
  const tail = Array.from({ length: 112 }, (_, i) => ({ tag: `topic-${String(i + 1).padStart(3, '0')}`, count: 1 }))
  // Shuffled deterministically so the story proves the sort rather than the data.
  return [...tail, ...head].sort((a, b) => (a.tag.length * 7 + a.tag.charCodeAt(0)) % 13 - (b.tag.length * 7 + b.tag.charCodeAt(0)) % 13)
})()

const base = { searchLabel: 'Search Vue Research', modelValue: '', query: '' }

/** No tags at all: the row is the search input and nothing else. No disabled
 *  Tags button, no empty chip area. */
export const NoTags: Story = {
  args: { ...base, tags: [] },
}

/** Four tags. The Tags button is still there — its presence depends on there
 *  being tags, not on whether they all fit, so it cannot blink in and out as
 *  the pane resizes. */
export const FewTags: Story = {
  args: { ...base, tags: few },
}

/**
 * The reason the component exists. A hundred and twenty tags and the row is
 * still one control tall: six chips, `Tags 120`, and the rest behind the
 * button. Compare with what `EntriesView` did before — the same data was a
 * cloud twenty rows deep.
 */
export const ManyTags: Story = {
  args: { ...base, tags: many },
}

/**
 * The active tag is `topic-042`, which is nowhere near the top six. It sits in
 * its fixed place after the input with a `×`, and inside the popover it
 * carries a `✓`.
 *
 * The quick row drops to **five** while a tag is active — the active chip counts
 * as one of the six. Pulling rank seven in to keep six *quick* chips made the
 * row one chip wider than it was a moment ago, and at the width this pane gets
 * that chip was what pushed the Tags button onto a second line, alone.
 */
export const ActiveTagOutsideTopSix: Story = {
  args: { ...base, tags: many, modelValue: 'topic-042' },
}

/** An active tag with a query standing beside it — the two compose; the meta
 *  slot is where the owner reports `3 of 12 matches`. */
export const ActiveTagAndQuery: Story = {
  args: { ...base, tags: many, modelValue: 'security', query: 'token' },
}

/**
 * Agent-written tags run long. Each chip caps at 12rem with an ellipsis and
 * carries the full text in `title`; the popover option truncates the same way.
 * Nothing here may grow to two lines — the row's height rule depends on it.
 */
export const LongTagNames: Story = {
  args: {
    ...base,
    tags: [
      { tag: 'kubernetes-admission-controller-webhook-v2', count: 5 },
      { tag: 'observability-platform-migration-plan-q3', count: 4 },
      { tag: 'security', count: 3 },
      { tag: 'the-longest-tag-anyone-has-ever-typed-into-this-product-so-far', count: 2 },
      { tag: 'api', count: 1 },
    ],
    modelValue: 'observability-platform-migration-plan-q3',
  },
}

/** The narrowest pane the product has — around 600px, at a 1000px viewport
 *  with the sidebar open. The row wraps; the popover stays inside the pane. */
export const NarrowPane: Story = {
  decorators: [() => ({ template: '<div style="max-width: 600px; min-height: 420px"><story /></div>' })],
  args: { ...base, tags: many, modelValue: 'security' },
}

/** The popover, open. The filter box has focus, the applied tag carries a ✓
 *  and the cursor sits on it. Type to narrow, arrows to move, Enter to apply,
 *  Escape to leave. */
export const PopoverOpen: Story = {
  args: { ...base, tags: many, modelValue: 'auth' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, /^Tags/)
  },
}

/** A filter that matches no tag: one muted line and the way back. */
export const PopoverNoMatch: Story = {
  args: { ...base, tags: many },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, /^Tags/)
    const input = await waitFor(() => canvasElement.querySelector<HTMLInputElement>('input[role="combobox"]'))
    if (!input) return
    input.value = 'zzz'
    input.dispatchEvent(new Event('input', { bubbles: true }))
  },
}

/**
 * Every tag used exactly once — a young research, or a section whose documents
 * share no vocabulary yet.
 *
 * This is the state a count threshold would have shown nothing in, which is why
 * the quick row is a fixed six. With every count equal the sort falls through to
 * its alphabetical tie-break, so the six chips are the first six names.
 *
 * The chips carry bare names: `TagList` prints a count only above one. The
 * popover still prints the `1` on every row — a column of ones is itself the
 * fact that this distribution has no head.
 */
export const EveryTagCountOne: Story = {
  args: {
    ...base,
    tags: ['webhooks', 'retention', 'schema', 'billing', 'ingestion', 'sso', 'rate-limits', 'audit-log', 'onboarding']
      .map(tag => ({ tag, count: 1 })),
  },
}

/**
 * The same row inside a share link: the placeholder reads “Filter these
 * entries…” rather than “Search this research…”, because a visitor's search is
 * not the server's — `/api/search` is not on the share sub-mux, so `EntriesView`
 * filters the entries the page already holds, by title, description and tags,
 * and the box must not promise the bodies it is not reading.
 *
 * No `withShare()` decorator here on purpose: this component reads no share
 * state, it takes a placeholder. The component that calls `shareActive()` is
 * `EntriesView`, and `Research/EntriesView → InsideAShareFiltering` is where
 * that decision is catalogued.
 */
export const SharePlaceholder: Story = {
  args: {
    ...base,
    tags: few,
    searchLabel: 'Search Vue Research',
    searchPlaceholder: 'Filter these entries…',
    query: 'compos',
  },
}

/**
 * A phone. Below 768px every control moves to the touch height, the quick row
 * is cut to three chips, and the popover widens to the full pane.
 *
 * The three hidden chips are hidden, not removed from the data — they are still
 * in the popover behind `Tags 120`, which is the only reason dropping them from
 * the row loses nothing.
 */
export const Phone: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
  args: { ...base, tags: many, modelValue: 'security' },
}

/** Polls, because there is no `@storybook/test` in this project — same helper
 *  shape as `EntriesView.stories.ts`. */
async function clickButton(root: HTMLElement, label: RegExp): Promise<void> {
  const button = await waitFor(() =>
    Array.from(root.querySelectorAll('button')).find(b => label.test(b.textContent?.trim() ?? '')),
  )
  button?.click()
}

async function waitFor<T>(find: () => T | undefined | null): Promise<T | undefined> {
  for (let i = 0; i < 50; i++) {
    const found = find()
    if (found) return found
    await new Promise(resolve => setTimeout(resolve, 20))
  }
  return undefined
}
