import type { Meta, StoryObj } from '@storybook/vue3'
import CrossReferencesBlock from './CrossReferencesBlock.vue'
import {
  mockOutgoingRefs,
  mockIncomingRefs,
  mockOutgoingRef,
  mockOutgoingRefCrossResearch,
  mockOutgoingRefUnresolved,
  mockIncomingRef,
} from '../../__mocks__/crossref'
import { includeEntriesOnly, mockShareMeta, withShare, withoutShare } from '../../__mocks__/share'

/**
 * What this entry points at, and what points at it.
 *
 * Every row is a link, except where it must not be. Under a share link a
 * reference whose target is outside the shared research renders as inert
 * markup — same layout, no href, no hover, no cursor — because a distinct
 * "dead link" treatment is itself a statement that a destination exists.
 *
 * The row can only render what it is handed, and that is the larger half of the
 * problem: the payload for a foreign reference must arrive with its title and
 * its research name already absent, server-side. Suppressing the link does not
 * help if the name of another research is in the DOM.
 */
const meta: Meta<typeof CrossReferencesBlock> = {
  title: 'Entry/CrossReferencesBlock',
  component: CrossReferencesBlock,
  tags: ['autodocs'],
  // Share state is module state; this gives the ordinary stories a known
  // starting point rather than whatever the last story left behind. The
  // trade-offs are in __mocks__/share.ts.
  decorators: [withoutShare()],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof CrossReferencesBlock>

export const OutgoingOnly: Story = {
  args: {
    outgoing: mockOutgoingRefs,
    incoming: [],
    researchSlug: 'R1',
  },
}

export const IncomingOnly: Story = {
  args: {
    outgoing: [],
    incoming: [
      mockIncomingRef,
      { ...mockIncomingRef, source_id: 'ent_005', entry_code: 'E5', entry_title: 'Performance Optimization' },
    ],
    researchSlug: 'R1',
  },
}

export const BothDirections: Story = {
  args: {
    outgoing: [mockOutgoingRef, mockOutgoingRefCrossResearch],
    incoming: [mockIncomingRef],
    researchSlug: 'R1',
  },
}

export const WithUnresolved: Story = {
  args: {
    outgoing: [mockOutgoingRef, mockOutgoingRefUnresolved],
    incoming: [],
    researchSlug: 'R1',
  },
}

export const Empty: Story = {
  args: {
    outgoing: [],
    incoming: [],
    researchSlug: 'R1',
  },
}

/** A roadmap reference, `[[RM1]]`. It resolves to the roadmap page, and under a
 *  share only when the link includes roadmaps — see `SharedWithoutRoadmaps`. */
export const RoadmapReference: Story = {
  args: {
    outgoing: [
      {
        target_roadmap_id: 'rm_001',
        target_ref: 'RM1:N3',
        roadmap_code: 'RM1',
        roadmap_title: 'Rollout plan',
        research_code: 'R1',
        resolved: true,
      },
    ],
    incoming: [],
    researchSlug: 'R1',
  },
}

/**
 * Inside a share link. The reference to another entry of the same research is a
 * link into `/s/{token}/entry/…`; the reference to R2 is inert, and the
 * unresolved one is inert too — under a share, "unresolved" is what the server
 * returns for a target the visitor may not know exists.
 *
 * The R2 row still shows a title here, which is the server's job to strip and
 * not this component's. The story is deliberately wrong in that one respect, so
 * the leak has somewhere to be visible.
 */
export const InsideAShare: Story = {
  decorators: [withShare()],
  args: {
    outgoing: [mockOutgoingRef, mockOutgoingRefCrossResearch, mockOutgoingRefUnresolved],
    incoming: [mockIncomingRef],
    researchSlug: 'R7',
  },
}

/** A share that does not include roadmaps. The `[[RM1]]` row renders and goes
 *  nowhere: the reference was in the author's text, so hiding the row would edit
 *  the document, but the destination is not part of this link. */
export const SharedWithoutRoadmaps: Story = {
  decorators: [withShare({ ...mockShareMeta, include: includeEntriesOnly })],
  args: {
    outgoing: [
      {
        target_roadmap_id: 'rm_001',
        target_ref: 'RM1',
        roadmap_code: 'RM1',
        roadmap_title: 'Rollout plan',
        research_code: 'R7',
        resolved: true,
      },
      { ...mockOutgoingRef, research_code: 'R7' },
    ],
    incoming: [],
    researchSlug: 'R7',
  },
}

/** An incoming reference from another research, under a share. It must not
 *  arrive at all — the fact that something outside cites this research is
 *  itself information — and if it does, it renders inert rather than as a way
 *  in. */
export const SharedIncomingFromElsewhere: Story = {
  decorators: [withShare()],
  args: {
    outgoing: [],
    incoming: [
      { ...mockIncomingRef, research_code: 'R7', research_name: undefined },
      {
        source_id: 'ent_r2_009',
        source_type: 'entry',
        entry_code: 'E9',
        entry_title: 'Vendor shortlist',
        research_code: 'R2',
        research_name: 'Procurement 2026',
      },
    ],
    researchSlug: 'R7',
  },
}
