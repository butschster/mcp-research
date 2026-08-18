import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapStagesBoard from './RoadmapStagesBoard.vue'
import { STAGES, NODES, EDGES, UNASSIGNED_NODES } from './roadmap.fixtures'

// The stages board buckets raw nodes into declared stage columns via
// bucketByStage; a leftover "Unassigned" column appears only when it has
// content. Declared stages always render, even empty.
const meta: Meta<typeof RoadmapStagesBoard> = {
  title: 'Roadmap/StagesBoard',
  component: RoadmapStagesBoard,
  tags: ['autodocs'],
  decorators: [
    // The board is height:100% — give it a bounded, scrollable canvas.
    () => ({ template: '<div style="height: 520px; border: 1px solid var(--color-border); border-radius: var(--radius);"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapStagesBoard>

export const Populated: Story = {
  args: { stages: STAGES, nodes: NODES, edges: EDGES },
}

export const AllUnassigned: Story = {
  name: 'All unassigned (banner)',
  args: { stages: STAGES, nodes: UNASSIGNED_NODES, edges: EDGES },
}

export const EmptyStages: Story = {
  name: 'Empty stages (no nodes)',
  args: { stages: STAGES, nodes: [], edges: [] },
}

export const ManyColumns: Story = {
  name: '12+ columns',
  args: {
    stages: [
      'Intake', 'Triage', 'Discovery', 'Research', 'Design', 'Spec',
      'Build', 'Review', 'QA', 'Staging', 'Launch', 'Adoption', 'Retro', 'Archive',
    ],
    nodes: [
      ...NODES,
      { id: 'x1', code: 'N7', title: 'Retrospective', description: 'What we learned.', node_type: 'step', status: 'pending', stage: 'Retro' },
      { id: 'x2', code: 'N8', title: 'QA pass', description: 'Regression sweep.', node_type: 'step', status: 'in_progress', stage: 'QA' },
    ],
    edges: EDGES,
  },
}
