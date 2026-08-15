import type { Meta, StoryObj } from '@storybook/vue3'
import PageHeader from './PageHeader.vue'

/**
 * The title of a page beside what that page lets you do to it.
 *
 * It replaces `PageHeaders.stories.ts`, which documented this pattern with
 * three stub components because the real one did not exist — the catalogue
 * admitting a component was missing. The layout had seven names across fourteen
 * pages, three of which had been moved into the global stylesheet and back out
 * twice; the second attempt broke the shared `/s/{token}` views, because those
 * deliberately render the owner's markup and a page-scoped rule does not reach
 * it. So the pages lose a header block rather than gaining a renamed class.
 */
const meta: Meta<typeof PageHeader> = {
  title: 'Layout/PageHeader',
  component: PageHeader,
  parameters: {
    docs: {
      description: {
        component:
          'Crumbs, a title that may carry a short code and a count, an actions '
          + 'cluster, and an optional lead. Everything but the title is optional.',
      },
    },
  },
}
export default meta

type Story = StoryObj<typeof PageHeader>

/** The plainest form: a root page with one action. */
export const Simple: Story = {
  render: () => ({
    components: { PageHeader },
    template: `
      <PageHeader title="Teams" lead="Teams own researches. Everyone in a team sees its researches.">
        <template #actions><button class="btn btn-sm btn-primary">+ New team</button></template>
      </PageHeader>
    `,
  }),
}

/** A research sub-page: crumbs, the research's code, and a count. */
export const WithCodeAndCount: Story = {
  render: () => ({
    components: { PageHeader },
    template: `
      <PageHeader
        :crumbs="[{ label: 'Research', to: '/' }, { label: 'Payment rails', to: '/research/R1' }, { label: 'Tasks' }]"
        code="R1"
        title="Tasks"
        :count="12"
      >
        <template #actions>
          <button class="btn btn-sm">Export</button>
          <button class="btn btn-sm btn-primary">+ New task</button>
        </template>
      </PageHeader>
    `,
  }),
}

/** A count of zero is a fact, so it renders; `undefined` is what hides it. */
export const ZeroCount: Story = {
  render: () => ({
    components: { PageHeader },
    template: `<PageHeader code="R1" title="Roadmaps" :count="0" />`,
  }),
}

/** No actions at all — the row collapses to the title. */
export const TitleOnly: Story = {
  render: () => ({
    components: { PageHeader },
    template: `<PageHeader title="Settings" />`,
  }),
}

/**
 * A long unbroken title next to a full actions cluster. This is the case the
 * seven hand-written versions of this row disagreed about, and the reason the
 * bar wraps rather than truncating: a research name is the agent's, not ours.
 */
export const LongTitleAndManyActions: Story = {
  render: () => ({
    components: { PageHeader },
    template: `
      <PageHeader
        :crumbs="[{ label: 'Research', to: '/' }, { label: 'Long' }]"
        code="R12"
        title="Evaluating pgvector against a dedicated vector database for retrieval over long research documents"
      >
        <template #actions>
          <span class="badge badge-active">active</span>
          <button class="btn btn-sm">Share</button>
          <button class="btn btn-sm">Export</button>
          <button class="btn btn-icon" aria-label="Tasks">T</button>
          <button class="btn btn-icon" aria-label="Graph">G</button>
        </template>
      </PageHeader>
    `,
  }),
}

/** Markup in the lead, which is why `lead` is a slot as well as a prop. */
export const LeadSlot: Story = {
  render: () => ({
    components: { PageHeader },
    template: `
      <PageHeader title="Northwind Ops">
        <template #lead>3 members &middot; You are an <strong>owner</strong></template>
      </PageHeader>
    `,
  }),
}
