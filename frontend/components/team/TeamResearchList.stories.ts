import type { Meta, StoryObj } from '@storybook/vue3'
import TeamResearchList from './TeamResearchList.vue'
import { mockResearch, mockResearchArchived, mockResearchCompleted, mockResearches } from '../../__mocks__/research'

/**
 * The work a team holds, on the team's own page — and the section that now
 * leads it.
 *
 * It used to be one conditional line at the bottom of the page
 * (`v-if="research_count > 0"`), which rendered nothing at all when the count
 * was zero. The single moment the product has to say *"there is no work in here
 * yet"* was the moment it went silent, and that is where the whole
 * create → invite → colleague-sees-nothing loop broke.
 *
 * `ResearchCard` covers the same nouns and is deliberately not reused: it is a
 * 400px grid card for the home page, where a research is the subject. Here it
 * is a supporting fact about a team, on a page whose other two sections are
 * rule lists — a grid of cards between them would be three layouts for three
 * lists of things.
 *
 * The empty case is not this component's: the page decides, because what to say
 * depends on whether the reader can move work in. An owner gets a button; a
 * viewer gets the owner's name.
 */
const meta: Meta<typeof TeamResearchList> = {
  title: 'Team/TeamResearchList',
  component: TeamResearchList,
  tags: ['autodocs'],
  argTypes: { researches: { control: 'object' } },
}
export default meta
type Story = StoryObj<typeof TeamResearchList>

/** The ordinary case: a few researches, mixed statuses. */
export const SeveralResearches: Story = {
  args: { researches: mockResearches },
}

/** One research — a team that has just had its first piece of work moved in. */
export const SingleResearch: Story = {
  args: { researches: [mockResearch] },
}

/**
 * Archived and completed work is listed, because the team page asks for every
 * status. A team holding eight archived researches is not an empty team, and
 * the research list's own default filter would have said it was.
 */
export const ArchivedAndCompleted: Story = {
  args: { researches: [mockResearchCompleted, mockResearchArchived] },
}

/** A goal that runs long. It truncates to one line: this is a list, not a
 *  reading surface, and a wrapping sentence breaks the row rhythm the page was
 *  rebuilt to get. */
export const LongGoal: Story = {
  args: {
    researches: [
      {
        ...mockResearch,
        goal: 'Investigate every observable difference between the Composition API and the Options API across component size, testability, tree-shaking and the shape of the resulting diffs in review, then write it up',
      },
    ],
  },
}

/** Cyrillic names at full length, which is the normal case in this product. */
export const CyrillicNames: Story = {
  args: {
    researches: [
      { ...mockResearch, code: 'R7', name: 'Исследование рынка систем управления знаниями', goal: 'Сравнить решения по стоимости владения' },
      { ...mockResearchCompleted, code: 'R8', name: 'Интеграция с внешними поставщиками данных', goal: '' },
    ],
  },
}

/** A research with no goal and no code yet — both cells hold their width so the
 *  rows below stay aligned. */
export const SparseRow: Story = {
  args: {
    researches: [{ id: 'res_new', name: 'Untitled research', status: 'active' }],
  },
}

/** Nothing to show. The component renders an empty rule set on purpose — the
 *  page owns the empty state, because only the page knows the reader's role. */
export const Empty: Story = {
  args: { researches: [] },
}
