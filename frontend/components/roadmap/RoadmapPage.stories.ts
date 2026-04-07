import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, h, ref, computed } from 'vue'
import RoadmapStepNode from './RoadmapStepNode.vue'
import RoadmapRootNode from './RoadmapRootNode.vue'
import RoadmapEdgeLabel from './RoadmapEdgeLabel.vue'
import './RoadmapPageMock.css'

/**
 * Complete roadmap page mock — shows a full roadmap graph layout
 * with root node, step/milestone/decision/info nodes, and edge labels.
 * This is a static representation of how the roadmap page renders.
 */

const mockNodes = [
  { id: 'root', code: 'RM1', title: 'Vue 3 Learning Path', description: 'From zero to production-ready Vue 3 development', status: 'active', statuses: ['not_started', 'in_progress', 'completed'], nodeCount: 9, edgeCount: 10, nodeType: 'root' },
  { id: 'n1', code: 'N1', title: 'HTML & CSS Fundamentals', description: 'Semantic HTML, Flexbox, Grid, responsive design, CSS variables', nodeType: 'step', status: 'completed' },
  { id: 'n2', code: 'N2', title: 'JavaScript ES6+', description: 'Arrow functions, destructuring, promises, async/await, modules', nodeType: 'step', status: 'completed' },
  { id: 'n3', code: 'N3', title: 'Fundamentals Complete', description: 'Core web development skills mastered', nodeType: 'milestone', status: 'completed' },
  { id: 'n4', code: 'N4', title: 'Vue 3 Basics', description: 'Template syntax, reactivity, components, props, events, slots', nodeType: 'step', status: 'in_progress' },
  { id: 'n5', code: 'N5', title: 'Composition API', description: 'ref, reactive, computed, watch, lifecycle hooks, composables', nodeType: 'step', status: 'not_started' },
  { id: 'n6', code: 'N6', title: 'Choose state management', description: 'Pinia vs composable-based state — evaluate for project needs', nodeType: 'decision', status: 'not_started' },
  { id: 'n7', code: 'N7', title: 'TypeScript recommended', description: 'TypeScript greatly improves DX with Vue 3 and Volar', nodeType: 'info', status: '' },
  { id: 'n8', code: 'N8', title: 'Testing & Deployment', description: 'Vitest, Vue Test Utils, CI/CD pipeline, Vercel/Netlify', nodeType: 'step', status: 'not_started' },
  { id: 'n9', code: 'N9', title: 'Production Ready', description: 'Ship your first Vue 3 application to production', nodeType: 'milestone', status: 'not_started' },
]

const mockEdges = [
  { id: 'e1', source: 'root', target: 'n1', label: '', edgeType: 'default' },
  { id: 'e2', source: 'n1', target: 'n2', label: 'next', edgeType: 'default' },
  { id: 'e3', source: 'n2', target: 'n3', label: 'next', edgeType: 'success' },
  { id: 'e4', source: 'n3', target: 'n4', label: 'next', edgeType: 'default' },
  { id: 'e5', source: 'n4', target: 'n5', label: 'next', edgeType: 'default' },
  { id: 'e6', source: 'n5', target: 'n6', label: 'next', edgeType: 'default' },
  { id: 'e7', source: 'n6', target: 'n8', label: 'either way', edgeType: 'default' },
  { id: 'e8', source: 'n7', target: 'n4', label: 'recommended', edgeType: 'optional' },
  { id: 'e9', source: 'n8', target: 'n9', label: 'ship it', edgeType: 'success' },
]

// Static layout component that mimics the roadmap page
const RoadmapPageMock = defineComponent({
  name: 'RoadmapPageMock',
  components: { RoadmapStepNode, RoadmapRootNode, RoadmapEdgeLabel },
  setup() {
    const researchName = 'Frontend Framework Research'
    const nodes = mockNodes
    const edges = mockEdges

    // Statuses for the progress bar
    const statusCounts = computed(() => {
      const counts: Record<string, number> = {}
      for (const n of nodes) {
        if (n.nodeType === 'root') continue
        const s = n.status || 'no_status'
        counts[s] = (counts[s] || 0) + 1
      }
      return counts
    })

    const totalNodes = computed(() => nodes.filter(n => n.nodeType !== 'root').length)

    return { researchName, nodes, edges, statusCounts, totalNodes }
  },
  template: `
    <div class="roadmap-page-mock">
      <!-- Toolbar -->
      <div class="rm-toolbar">
        <div class="rm-toolbar-left">
          <button class="btn btn-sm rm-toolbar-back">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            Back
          </button>
          <span class="rm-toolbar-title">{{ researchName }}</span>
          <span class="rm-toolbar-sep"></span>
          <span class="rm-toolbar-roadmap-name">RM1 — Vue 3 Learning Path</span>
        </div>
        <div class="rm-toolbar-right">
          <div class="rm-progress">
            <span class="rm-progress-label">{{ statusCounts['completed'] || 0 }}/{{ totalNodes }} completed</span>
            <div class="rm-progress-bar">
              <div class="rm-progress-fill" :style="{ width: ((statusCounts['completed'] || 0) / totalNodes * 100) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Graph area -->
      <div class="rm-canvas">
        <!-- Root node -->
        <div class="rm-graph-root">
          <RoadmapRootNode :data="nodes[0]" />
        </div>

        <!-- Arrow down from root -->
        <div class="rm-connector">
          <svg width="2" height="32"><line x1="1" y1="0" x2="1" y2="32" stroke="var(--color-border-strong)" stroke-width="2" /></svg>
        </div>

        <!-- Main flow -->
        <div class="rm-graph-flow">
          <template v-for="(node, i) in nodes.slice(1)" :key="node.id">
            <div class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <!-- Edge label to next node -->
              <div v-if="edges[i+1] && edges[i+1].label" class="rm-edge-label-wrapper">
                <svg width="2" height="16"><line x1="1" y1="0" x2="1" y2="16" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg>
                <RoadmapEdgeLabel :label="edges[i+1].label" :edge-type="edges[i+1].edgeType" />
                <svg width="2" height="16"><line x1="1" y1="0" x2="1" y2="16" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg>
              </div>
              <div v-else-if="i < nodes.length - 2" class="rm-connector-sm">
                <svg width="2" height="24"><line x1="1" y1="0" x2="1" y2="24" stroke="var(--color-border)" stroke-width="1" /></svg>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  `,
})

const meta: Meta<typeof RoadmapPageMock> = {
  title: 'Roadmap/RoadmapPage',
  component: RoadmapPageMock,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
}
export default meta
type Story = StoryObj<typeof RoadmapPageMock>

export const VueLearningPath: Story = {}

// Marketing strategy variant
const MarketingRoadmapMock = defineComponent({
  name: 'MarketingRoadmapMock',
  components: { RoadmapStepNode, RoadmapRootNode, RoadmapEdgeLabel },
  setup() {
    const nodes = [
      { id: 'root', code: 'RM2', title: 'Q2 Product Launch Strategy', description: 'Go-to-market plan for the new analytics dashboard', status: 'active', statuses: ['planned', 'approved', 'in_progress', 'launched'], nodeCount: 7, edgeCount: 7, nodeType: 'root' },
      { id: 'n1', code: 'N1', title: 'Market Research', description: 'Analyze competitors, identify target segments, validate pricing', nodeType: 'step', status: 'launched' },
      { id: 'n2', code: 'N2', title: 'Messaging & Positioning', description: 'Core value proposition, key differentiators, elevator pitch', nodeType: 'step', status: 'approved' },
      { id: 'n3', code: 'N3', title: 'Content Strategy', description: 'Blog posts, case studies, landing pages, email sequences', nodeType: 'step', status: 'in_progress' },
      { id: 'n4', code: 'N4', title: 'Launch Channel Decision', description: 'Product Hunt vs direct launch vs partner channels', nodeType: 'decision', status: 'planned' },
      { id: 'n5', code: 'N5', title: 'Beta Program', description: 'Invite 50 early adopters, collect feedback, iterate', nodeType: 'milestone', status: 'planned' },
      { id: 'n6', code: 'N6', title: 'Budget: $15K allocated', description: 'Approved budget covers ads, content, and tooling', nodeType: 'info', status: '' },
      { id: 'n7', code: 'N7', title: 'Public Launch', description: 'Coordinated launch across all channels', nodeType: 'milestone', status: 'planned' },
    ]
    const totalNodes = nodes.filter(n => n.nodeType !== 'root').length
    const launchedCount = nodes.filter(n => n.status === 'launched').length
    return { nodes, totalNodes, launchedCount }
  },
  template: `
    <div class="roadmap-page-mock">
      <div class="rm-toolbar">
        <div class="rm-toolbar-left">
          <button class="btn btn-sm rm-toolbar-back">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            Back
          </button>
          <span class="rm-toolbar-title">Product Analytics Research</span>
          <span class="rm-toolbar-sep"></span>
          <span class="rm-toolbar-roadmap-name">RM2 — Q2 Product Launch Strategy</span>
        </div>
        <div class="rm-toolbar-right">
          <div class="rm-progress">
            <span class="rm-progress-label">{{ launchedCount }}/{{ totalNodes }} launched</span>
            <div class="rm-progress-bar">
              <div class="rm-progress-fill" :style="{ width: (launchedCount / totalNodes * 100) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
      <div class="rm-canvas">
        <div class="rm-graph-root">
          <RoadmapRootNode :data="nodes[0]" />
        </div>
        <div class="rm-connector">
          <svg width="2" height="32"><line x1="1" y1="0" x2="1" y2="32" stroke="var(--color-border-strong)" stroke-width="2" /></svg>
        </div>
        <div class="rm-graph-flow">
          <div v-for="node in nodes.slice(1)" :key="node.id" class="rm-graph-item">
            <RoadmapStepNode :data="node" />
            <div class="rm-connector-sm">
              <svg width="2" height="20"><line x1="1" y1="0" x2="1" y2="20" stroke="var(--color-border)" stroke-width="1" /></svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
})

export const MarketingStrategy: Story = {
  render: () => ({
    components: { MarketingRoadmapMock },
    template: '<MarketingRoadmapMock />',
  }),
}
