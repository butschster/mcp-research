import type { Meta, StoryObj } from '@storybook/vue3'
import SearchModal from './SearchModal.vue'
import { mockApiData } from '../__mocks__/api'
import { markupDescription, markupImg, markupResearchName } from '../__mocks__/markup'

const meta: Meta<typeof SearchModal> = {
  title: 'Search/SearchModal',
  component: SearchModal,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof SearchModal>

/**
 * Shows the trigger button in its default state.
 * The modal opens on click or Cmd+K.
 */
export const TriggerButton: Story = {}

/**
 * Demonstrates the open modal with search results.
 * Since SearchModal manages its own open state and fetches data internally,
 * use the play function to open it.
 */
export const OpenState: Story = {
  play: async ({ canvasElement }) => {
    const trigger = canvasElement.querySelector('.search-trigger') as HTMLElement
    if (trigger) trigger.click()
  },
}

/** Opens the modal and types `q`, once the reset-on-open watcher has run. */
function openAndType(q: string) {
  return async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const trigger = canvasElement.querySelector('.search-trigger') as HTMLElement | null
    trigger?.click()
    // `open` is watched, and the watcher clears the query and focuses the input.
    // Typing before it runs types into a field that is about to be emptied.
    await new Promise((resolve) => setTimeout(resolve, 0))
    await new Promise((resolve) => setTimeout(resolve, 0))
    const input = canvasElement.querySelector('.search-input') as HTMLInputElement | null
    if (!input) return
    input.value = q
    input.dispatchEvent(new Event('input', { bubbles: true }))
  }
}

const markupResults = {
  '/api/researches': {
    data: [
      {
        id: 'res_009',
        code: 'R9',
        name: markupResearchName,
        goal: 'How author-supplied HTML reaches the DOM',
        status: 'active',
        tags: ['rendering'],
      },
    ],
  },
  // The real endpoint is full-text search; this answers the one query the
  // stories type and returns nothing for anything else, so the empty-result
  // line is reachable.
  '/api/search': (url: string) => ({
    entries: /[?&]q=[^&]*render/i.test(url)
      ? [
          {
            id: 'ent_markup',
            code: 'E9',
            research_id: 'res_009',
            title: `Rendering ${markupImg} in a result row`,
            description: markupDescription,
            tags: ['rendering', 'escaping'],
          },
        ]
      : [],
  }),
}

/**
 * A result whose name and title contain markup, with the query highlighted.
 *
 * Three fields here go to `v-html` through `highlight()` — a research name, an
 * entry title, an entry description — and every one of them is written by
 * somebody else's agent. `highlight` now escapes the text **and** the query
 * before it builds the regex, in that order: escaping afterwards would eat the
 * `<mark>` it had just inserted.
 *
 * What to look for: the tags read as text, and the matched run is wrapped. If
 * escaping regresses, `XSS EXECUTED` appears in the result row.
 *
 * The query is escaped for a second reason, less obvious than the first — it is
 * interpolated into the replacement pattern's neighbourhood and typed by the
 * person reading the screen, so it is the one string here that is guaranteed to
 * contain whatever anyone felt like trying.
 *
 * One artifact this makes visible and does not fix: because the text is escaped
 * first, a query of `lt` or `amp` matches inside an entity and highlights
 * characters the reader never typed. It costs a wrong `<mark>`, not a wrong
 * escape.
 */
export const MarkupInResults: Story = {
  render: () => ({
    components: { SearchModal },
    setup() {
      mockApiData(markupResults)
    },
    template: '<SearchModal />',
  }),
  play: openAndType('render'),
}

/** The same list with a query that is itself markup. Nothing matches, so the
 *  point is the line underneath: `No results for "…"` prints the query, and it
 *  prints it as text. */
export const MarkupInQuery: Story = {
  render: () => ({
    components: { SearchModal },
    setup() {
      mockApiData(markupResults)
    },
    template: '<SearchModal />',
  }),
  play: openAndType('<img src=x onerror="alert(1)">'),
}
