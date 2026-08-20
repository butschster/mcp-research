import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import AnnotationList from './AnnotationList.vue'
import type { Annotation, QueueFilters } from '../../composables/useAnnotations'
import {
  makeAnnotation,
  mockAnnotationQueue,
  mockAnnotationsAnswered,
} from '../../__mocks__/annotation'

/**
 * The list of marks, wherever it is shown — the document panel and the research
 * queue are the same component, because two would be two sets of empty states.
 *
 * The empty states are most of what this page is for. "Nothing is marked here",
 * "nothing is marked in this research" and "your filter matches nothing" are
 * three different messages, and showing the wrong one sends somebody looking for
 * a feature that is working. `emptyVariant` picks between the first two; the
 * third is chosen by the component itself the moment any filter is set, which is
 * why `EmptyBecauseFiltered` passes `research` and still gets the filtered copy.
 *
 * Grouping is not decoration either: a batch that is one document costs one read
 * of it, a batch scattered across six costs six. Within a group the rows are
 * ordered by block index, so they read in the order of the page, and an orphan —
 * which has no place — sinks to the end.
 */
const meta: Meta<typeof AnnotationList> = {
  title: 'Annotations/AnnotationList',
  component: AnnotationList,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
    loading: { control: 'boolean' },
    error: { control: 'text' },
    showFilters: { control: 'boolean' },
    grouped: { control: 'boolean' },
    selectable: { control: 'boolean' },
    dense: { control: 'boolean' },
    emptyVariant: { control: 'inline-radio', options: ['document', 'research', 'filtered'] },
    annotations: { control: false },
    onOpen: { action: 'open' },
    'onToggle-select': { action: 'toggle-select' },
    'onUpdate:filters': { action: 'update:filters' },
    onRetry: { action: 'retry' },
  },
  args: {
    annotations: mockAnnotationQueue,
    researchSlug: 'R1',
    grouped: true,
    loading: false,
    error: null,
  },
}
export default meta
type Story = StoryObj<typeof AnnotationList>

/**
 * The research queue: seven marks across two documents, each group headed by the
 * document's code and title and linking to it.
 */
export const GroupedByDocument: Story = {
  args: { annotations: mockAnnotationQueue, grouped: true },
}

/**
 * The document panel: one document's marks, so the group headers would repeat the
 * page you are already on. `grouped` off collapses them into a single list.
 */
export const Ungrouped: Story = {
  args: {
    annotations: mockAnnotationQueue.filter((a) => a.entry_id === 'ent_001'),
    grouped: false,
  },
}

/** With the filter row: status, kind and anchor state, all controlled by the caller. */
export const WithFilters: Story = {
  args: { showFilters: true, filters: {} },
}

/**
 * Filters actually applied, and actually filtering.
 *
 * The list does not filter — the caller refetches — so this story holds the
 * filter state and narrows the array itself, which is the round trip the
 * research page makes.
 */
export const InteractiveFilters: Story = {
  render: () => ({
    components: { AnnotationList },
    setup() {
      const filters = ref<QueueFilters>({ status: 'open' })
      const all = mockAnnotationQueue
      const shown = ref<Annotation[]>(all.filter((a) => a.status === 'open'))
      const apply = (next: QueueFilters) => {
        filters.value = next
        shown.value = all.filter((a) =>
          (!next.status || a.status === next.status)
          && (!next.kind || a.kind === next.kind)
          && (!next.anchor || (a.anchor?.state ?? 'anchored') === next.anchor))
      }
      return { filters, shown, apply }
    },
    template: `
      <AnnotationList
        :annotations="shown"
        research-slug="R1"
        grouped
        show-filters
        :filters="filters"
        @update:filters="apply"
      />
    `,
  }),
}

/** The batch surface: checkboxes on, notes hidden, two rows already picked. */
export const SelectableDense: Story = {
  args: {
    annotations: mockAnnotationsAnswered,
    selectable: true,
    dense: true,
    selectedIds: [mockAnnotationsAnswered[0]!.id, mockAnnotationsAnswered[2]!.id],
  },
}

/** Waiting on the server. Nothing else renders — not even an empty state that would lie. */
export const Loading: Story = {
  args: { annotations: [], loading: true },
}

/**
 * The request failed, and says so with a Retry beside it.
 *
 * The alternative — falling through to "No marks yet" — is a claim the component
 * is in no position to make, and it is the same claim a working empty research
 * makes, so nobody would know which one they were looking at.
 */
export const LoadFailed: Story = {
  args: { annotations: [], error: 'Could not load marks — the server returned 500.' },
}

/** Nothing marked in this document yet, phrased for the document panel. */
export const EmptyInDocument: Story = {
  args: { annotations: [], emptyVariant: 'document', grouped: false },
}

/** Nothing marked anywhere in the research, phrased for the queue. */
export const EmptyInResearch: Story = {
  args: { annotations: [], emptyVariant: 'research' },
}

/** The filtered-empty copy, asked for directly. */
export const EmptyFiltered: Story = {
  args: { annotations: [], emptyVariant: 'filtered', showFilters: true },
}

/**
 * `emptyVariant` still says `research`, but a filter is set — so the component
 * overrides it and says the filter matched nothing.
 *
 * That override is the whole reason the variant is a hint rather than an
 * instruction: the caller knows which surface it is, the component knows whether
 * anything was narrowed, and only the second one changes the message.
 */
export const EmptyBecauseFiltered: Story = {
  args: {
    annotations: [],
    emptyVariant: 'research',
    showFilters: true,
    filters: { status: 'answered', kind: 'dig' },
  },
}

/**
 * A filter set on the document panel, which keeps its own copy anyway: a person
 * looking at one document has just been told there is nothing in it, and
 * "nothing matches" would be the less useful of the two truths.
 */
export const EmptyInDocumentDespiteFilters: Story = {
  args: {
    annotations: [],
    emptyVariant: 'document',
    grouped: false,
    showFilters: true,
    filters: { kind: 'verify' },
  },
}

/**
 * Order inside a group.
 *
 * These four arrive newest-first, as the API returns them, and are drawn in block
 * order — A24, A22, A25, A23 — with the orphan last, because a mark with no place
 * in the document cannot claim one in the list.
 */
export const OrderedByPosition: Story = {
  args: {
    grouped: true,
    annotations: [
      makeAnnotation({ code: 'A22', anchor: { state: 'anchored', strategy: 'block+quote', confidence: 1, block_index: 4 } }),
      makeAnnotation({ code: 'A23', anchor: { state: 'orphaned', strategy: 'none', confidence: 0, block_index: -1 } }),
      makeAnnotation({ code: 'A24', anchor: { state: 'anchored', strategy: 'block+quote', confidence: 1, block_index: 1 } }),
      makeAnnotation({ code: 'A25', anchor: { state: 'moved', strategy: 'quote', confidence: 0.6, block_index: 9 } }),
    ],
  },
}

/** Marks whose document the payload did not name — the group heads itself "Untitled" and does not link. */
export const GroupWithoutEntryTitle: Story = {
  args: {
    grouped: true,
    researchSlug: '',
    annotations: [
      makeAnnotation({ code: 'A26', entry_code: undefined, entry_title: undefined }),
      makeAnnotation({ code: 'A27', entry_code: undefined, entry_title: undefined, entry_id: 'ent_001' }),
    ],
  },
}

/**
 * Forty marks across four documents. Each group's rows scroll inside their own
 * bounded box rather than the page growing to forty rows, which is what keeps
 * the filter row reachable.
 */
export const ManyMarks: Story = {
  render: () => ({
    components: { AnnotationList },
    setup() {
      const entries = [
        { id: 'ent_001', code: 'E1', title: 'Component Composition Patterns' },
        { id: 'ent_002', code: 'E2', title: 'Reactive State Management' },
        { id: 'ent_003', code: 'E3', title: 'Template Syntax Deep Dive' },
        { id: 'ent_004', code: 'E4', title: 'Slots and Render Functions' },
      ]
      const annotations = Array.from({ length: 40 }, (_, i) => {
        const entry = entries[i % entries.length]!
        return makeAnnotation({
          code: `A${100 + i}`,
          entry_id: entry.id,
          entry_code: entry.code,
          entry_title: entry.title,
          kind: (['verify', 'dig', 'disagree'] as const)[i % 3],
          status: (['open', 'answered', 'closed', 'dismissed'] as const)[i % 4],
          attempts: i % 5 === 0 ? 2 : 0,
          anchor: {
            state: (['anchored', 'drifted', 'moved', 'orphaned'] as const)[i % 4]!,
            strategy: 'block+quote',
            confidence: i % 4 === 2 ? 0.6 : 1,
            block_index: i,
          },
        })
      })
      return { annotations }
    },
    template: `<AnnotationList :annotations="annotations" research-slug="R1" grouped show-filters />`,
  }),
}
