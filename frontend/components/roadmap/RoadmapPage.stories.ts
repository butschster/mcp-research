import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, computed, ref, markRaw, onMounted, nextTick } from 'vue'
import { VueFlow, useVueFlow, Position } from '@vue-flow/core'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import dagre from '@dagrejs/dagre'
import RoadmapStepNode from './RoadmapStepNode.vue'
import RoadmapRootNode from './RoadmapRootNode.vue'
import RoadmapRefNode from './RoadmapRefNode.vue'
import RoadmapNodePopover from './RoadmapNodePopover.vue'

/**
 * Interactive roadmap page using real Vue Flow with dagre auto-layout.
 * Each story provides a mock API response and renders it as the real page would.
 */

// --- API response shape (matches GET /api/roadmaps/{id}) ---
interface MockRefData {
  title?: string; status?: string; code?: string; description?: string
  research_id?: string; section_name?: string; content?: string
  priority?: string; result?: string
  total_questions?: number; answered_questions?: number
  section_count?: number; entry_count?: number
}
interface MockNode {
  id: string; code: string; title: string; description: string
  node_type: string; status: string
  position_x: number; position_y: number; parent_id: string
  ref_type?: string; ref_id?: string; metadata?: string; ref_data?: MockRefData
}
interface MockEdge {
  id: string; source_node_id: string; target_node_id: string
  label: string; edge_type: string
}
interface MockRoadmap {
  id: string; code: string; research_id: string
  title: string; description: string
  statuses: string[]; status: string
  nodes: MockNode[]; edges: MockEdge[]
}

const EDGE_STYLES: Record<string, Record<string, any>> = {
  default: { stroke: 'rgba(140,150,170,0.5)', strokeWidth: 2 },
  success: { stroke: 'rgba(107,203,119,0.6)', strokeWidth: 2 },
  warning: { stroke: 'rgba(240,184,73,0.6)', strokeWidth: 2 },
  optional: { stroke: 'rgba(140,150,170,0.35)', strokeWidth: 1.5, strokeDasharray: '4 4' },
}

const NODE_SIZES: Record<string, { width: number; height: number }> = {
  'roadmap-root': { width: 400, height: 160 },
  'roadmap-step': { width: 320, height: 120 },
  'roadmap-ref': { width: 340, height: 140 },
}

function buildVueFlowGraph(data: MockRoadmap) {
  const targets = new Set(data.edges.map(e => e.target_node_id))
  const rootId = data.nodes.find(n => !targets.has(n.id))?.id ?? data.nodes[0]?.id

  // Build nodes
  const nodes = data.nodes.map(n => {
    const isRoot = n.id === rootId
    const hasRef = !isRoot && n.ref_type && n.ref_id

    let type: string
    if (isRoot) type = 'roadmap-root'
    else if (hasRef) type = 'roadmap-ref'
    else type = 'roadmap-step'

    let nodeData: Record<string, any>
    if (isRoot) {
      nodeData = { code: data.code, title: data.title, description: data.description, status: data.status, statuses: data.statuses, nodeCount: data.nodes.length, edgeCount: data.edges.length }
    } else {
      nodeData = { code: n.code, title: n.title, description: n.description, nodeType: n.node_type, status: n.status, refType: n.ref_type, refId: n.ref_id, metadata: n.metadata, refData: n.ref_data }
    }

    return {
      id: n.id,
      type,
      position: { x: 0, y: 0 },
      data: nodeData,
    }
  })

  // Build edges
  const edges = data.edges.map(e => ({
    id: e.id,
    source: e.source_node_id,
    target: e.target_node_id,
    type: 'smoothstep',
    label: e.label || undefined,
    style: EDGE_STYLES[e.edge_type] ?? EDGE_STYLES.default,
    labelStyle: { fill: 'var(--color-text-muted)', fontSize: '0.625rem' },
    labelBgStyle: { fill: 'var(--color-surface)', fillOpacity: 0.9 },
    labelBgPadding: [4, 2] as [number, number],
    labelBgBorderRadius: 3,
  }))

  // Dagre layout
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 80, ranksep: 140 })

  for (const node of nodes) {
    const size = NODE_SIZES[node.type] ?? NODE_SIZES['roadmap-step']
    g.setNode(node.id, { width: size.width, height: size.height })
  }
  for (const edge of edges) {
    g.setEdge(edge.source, edge.target)
  }
  dagre.layout(g)

  for (const node of nodes) {
    const pos = g.node(node.id)
    const size = NODE_SIZES[node.type] ?? NODE_SIZES['roadmap-step']
    node.position = { x: pos.x - size.width / 2, y: pos.y - size.height / 2 }
    ;(node as any).sourcePosition = Position.Bottom
    ;(node as any).targetPosition = Position.Top
  }

  return { nodes, edges }
}

// --- Page component ---
const RoadmapPageView = defineComponent({
  name: 'RoadmapPageView',
  components: { VueFlow, RoadmapStepNode, RoadmapRootNode, RoadmapRefNode, RoadmapNodePopover },
  props: {
    roadmap: { type: Object as () => MockRoadmap, required: true },
    researchName: { type: String, default: 'Research' },
  },
  setup(props) {
    const nodeTypes = {
      'roadmap-root': markRaw(RoadmapRootNode),
      'roadmap-step': markRaw(RoadmapStepNode),
      'roadmap-ref': markRaw(RoadmapRefNode),
    }

    const data = ref({ ...props.roadmap, nodes: props.roadmap.nodes.map(n => ({ ...n })) })
    const graph = computed(() => buildVueFlowGraph(data.value))

    // Progress
    const lastStatus = data.value.statuses.length > 0 ? data.value.statuses[data.value.statuses.length - 1] : null
    const progress = computed(() => {
      const total = data.value.nodes.length
      const completed = lastStatus ? data.value.nodes.filter(n => n.status === lastStatus).length : 0
      return { total, completed, percent: total > 0 ? Math.round((completed / total) * 100) : 0 }
    })

    // Popover
    const selectedNode = ref<any>(null)
    const popoverPos = ref({ x: 0, y: 0 })

    function onNodeClick({ node, event }: { node: any; event: MouseEvent }) {
      if (node.type === 'roadmap-root') return
      selectedNode.value = {
        id: node.id,
        title: node.data.title,
        description: node.data.description || '',
        nodeType: node.data.nodeType || 'step',
        status: node.data.status || '',
        refType: node.data.refType,
        refId: node.data.refId,
        refData: node.data.refData,
      }
      popoverPos.value = { x: event.clientX + 12, y: event.clientY - 20 }
    }

    function onUpdateStatus(nodeId: string, status: string) {
      const n = data.value.nodes.find(n => n.id === nodeId)
      if (n) n.status = status
      selectedNode.value = null
    }

    // Fit view on mount
    const { fitView } = useVueFlow()
    onMounted(() => {
      nextTick(() => fitView({ padding: 0.15, duration: 300 }))
    })

    return { nodeTypes, graph, progress, data, selectedNode, popoverPos, onNodeClick, onUpdateStatus, fitView }
  },
  template: `
    <div style="height:100vh;display:flex;flex-direction:column;background:var(--color-bg);">
      <!-- Toolbar -->
      <div style="display:flex;align-items:center;justify-content:space-between;padding:var(--space-3) var(--space-5);background:rgba(21,29,46,0.9);backdrop-filter:blur(12px);border-bottom:1px solid var(--color-border);gap:var(--space-4);flex-shrink:0;">
        <div style="display:flex;align-items:center;gap:var(--space-3);">
          <button class="btn btn-sm" style="gap:var(--space-1);">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m15 18-6-6 6-6"/></svg>
            Back
          </button>
          <span style="font-size:var(--type-sm);font-weight:600;color:var(--color-text);">{{ researchName }}</span>
          <span style="width:1px;height:20px;background:var(--color-border-strong);"></span>
          <span style="font-size:var(--type-xs);color:var(--color-primary);font-weight:500;">{{ data.code }} — {{ data.title }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:var(--space-2);">
          <span style="font-size:var(--type-xs);color:var(--color-text-muted);font-variant-numeric:tabular-nums;">{{ progress.completed }}/{{ progress.total }}</span>
          <div style="width:100px;height:4px;background:var(--color-surface-hover);border-radius:2px;overflow:hidden;">
            <div :style="{ width: progress.percent + '%', height: '100%', background: 'rgba(107,203,119,0.8)', borderRadius: '2px', transition: 'width 0.3s' }"></div>
          </div>
          <span style="width:1px;height:20px;background:var(--color-border-strong);margin:0 var(--space-1);"></span>
          <button class="btn btn-sm" @click="fitView({ padding: 0.15, duration: 300 })">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
            Fit
          </button>
        </div>
      </div>

      <!-- Vue Flow Canvas -->
      <div style="flex:1;min-height:0;">
        <VueFlow
          :nodes="graph.nodes"
          :edges="graph.edges"
          :node-types="nodeTypes"
          :default-viewport="{ x: 0, y: 0, zoom: 0.85 }"
          :min-zoom="0.1"
          :max-zoom="2"
          :fit-view-on-init="true"
          :nodes-draggable="true"
          :nodes-connectable="false"
          :pan-on-drag="true"
          :zoom-on-scroll="true"
          style="width:100%;height:100%;"
          @node-click="onNodeClick"
        />
      </div>

      <!-- Popover -->
      <RoadmapNodePopover
        v-if="selectedNode"
        :node="selectedNode"
        :statuses="data.statuses"
        :position="popoverPos"
        @update-status="onUpdateStatus"
        @close="selectedNode = null"
      />
    </div>
  `,
})

// =============================================================================
// Mock API responses
// =============================================================================

const vueLearningPath: MockRoadmap = {
  id: 'rm-1', code: 'RM1', research_id: 'r-1',
  title: 'Vue 3 Learning Path',
  description: 'From zero to production-ready Vue 3 development',
  statuses: ['not_started', 'in_progress', 'completed'],
  status: 'active',
  nodes: [
    { id: 'n1', code: 'N1', title: 'HTML & CSS Fundamentals', description: 'Semantic HTML, Flexbox, Grid, responsive design', node_type: 'step', status: 'completed', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n2', code: 'N2', title: 'JavaScript ES6+', description: 'Arrow functions, destructuring, promises, async/await', node_type: 'step', status: 'completed', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n3', code: 'N3', title: 'Fundamentals Complete', description: 'Core web skills mastered', node_type: 'milestone', status: 'completed', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n4', code: 'N4', title: 'Vue 3 Basics', description: 'Template syntax, reactivity, components, props, events', node_type: 'step', status: 'in_progress', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n5', code: 'N5', title: 'Composition API', description: 'ref, reactive, computed, watch, lifecycle hooks', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n6', code: 'N6', title: 'State Management', description: 'Pinia store patterns, composable-based state', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n7', code: 'N7', title: 'Testing & Deployment', description: 'Vitest, Vue Test Utils, CI/CD, Vercel', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n8', code: 'N8', title: 'Production Ready', description: 'Ship your first Vue 3 app', node_type: 'milestone', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n1', target_node_id: 'n2', label: 'next', edge_type: 'default' },
    { id: 'e2', source_node_id: 'n2', target_node_id: 'n3', label: '', edge_type: 'success' },
    { id: 'e3', source_node_id: 'n3', target_node_id: 'n4', label: 'next', edge_type: 'default' },
    { id: 'e4', source_node_id: 'n4', target_node_id: 'n5', label: 'next', edge_type: 'default' },
    { id: 'e5', source_node_id: 'n5', target_node_id: 'n6', label: 'next', edge_type: 'default' },
    { id: 'e6', source_node_id: 'n6', target_node_id: 'n7', label: 'next', edge_type: 'default' },
    { id: 'e7', source_node_id: 'n7', target_node_id: 'n8', label: 'ship it', edge_type: 'success' },
  ],
}

const frameworkDecision: MockRoadmap = {
  id: 'rm-3', code: 'RM3', research_id: 'r-2',
  title: 'Frontend Framework Decision',
  description: 'Evaluate and choose the right framework for the project',
  statuses: ['not_started', 'evaluating', 'decided'],
  status: 'active',
  nodes: [
    { id: 'n1', code: 'N1', title: 'Define Requirements', description: 'Performance, team size, SSR, ecosystem', node_type: 'step', status: 'decided', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n2', code: 'N2', title: 'Framework Shortlist', description: 'Narrowed to 3 candidates', node_type: 'milestone', status: 'decided', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n3', code: 'N3', title: 'Evaluate React 19', description: 'Server components, suspense, concurrent', node_type: 'step', status: 'evaluating', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n4', code: 'N4', title: 'Evaluate Vue 3', description: 'Composition API, Nuxt 4, Volar DX', node_type: 'step', status: 'evaluating', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n5', code: 'N5', title: 'Evaluate Svelte 5', description: 'Runes, compiled output, SvelteKit', node_type: 'step', status: 'evaluating', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n6', code: 'N6', title: 'Build React POC', description: 'Prototype with Next.js', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n7', code: 'N7', title: 'Build Vue POC', description: 'Prototype with Nuxt 4', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n8', code: 'N8', title: 'Build Svelte POC', description: 'Prototype with SvelteKit', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n9', code: 'N9', title: 'Compare Benchmarks', description: 'Bundle size, lighthouse, DX survey', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n10', code: 'N10', title: 'Team Vote', description: 'Present findings, make final call', node_type: 'decision', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n11', code: 'N11', title: 'Framework Chosen', description: 'Decision documented', node_type: 'milestone', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n1', target_node_id: 'n2', label: 'next', edge_type: 'default' },
    { id: 'e2', source_node_id: 'n2', target_node_id: 'n3', label: 'option A', edge_type: 'default' },
    { id: 'e3', source_node_id: 'n2', target_node_id: 'n4', label: 'option B', edge_type: 'default' },
    { id: 'e4', source_node_id: 'n2', target_node_id: 'n5', label: 'option C', edge_type: 'default' },
    { id: 'e5', source_node_id: 'n3', target_node_id: 'n6', label: 'build', edge_type: 'default' },
    { id: 'e6', source_node_id: 'n4', target_node_id: 'n7', label: 'build', edge_type: 'default' },
    { id: 'e7', source_node_id: 'n5', target_node_id: 'n8', label: 'build', edge_type: 'default' },
    { id: 'e8', source_node_id: 'n6', target_node_id: 'n9', label: 'results', edge_type: 'success' },
    { id: 'e9', source_node_id: 'n7', target_node_id: 'n9', label: 'results', edge_type: 'success' },
    { id: 'e10', source_node_id: 'n8', target_node_id: 'n9', label: 'results', edge_type: 'success' },
    { id: 'e11', source_node_id: 'n9', target_node_id: 'n10', label: 'next', edge_type: 'default' },
    { id: 'e12', source_node_id: 'n10', target_node_id: 'n11', label: 'chosen', edge_type: 'success' },
  ],
}

const fullStackRoadmap: MockRoadmap = {
  id: 'rm-4', code: 'RM4', research_id: 'r-3',
  title: 'Full-Stack Developer Roadmap',
  description: 'Parallel frontend and backend tracks converging at milestones',
  statuses: ['not_started', 'learning', 'practiced', 'mastered'],
  status: 'active',
  nodes: [
    { id: 'n0', code: 'N1', title: 'Git & Terminal Basics', description: 'Version control, CLI, SSH keys', node_type: 'step', status: 'mastered', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n1', code: 'N2', title: 'HTML/CSS/JS', description: 'Semantic HTML, modern CSS, ES2024', node_type: 'step', status: 'mastered', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n2', code: 'N3', title: 'React or Vue', description: 'Pick one, build 3 projects', node_type: 'decision', status: 'practiced', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n3', code: 'N4', title: 'State & Routing', description: 'Client state, SPA routing', node_type: 'step', status: 'learning', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n4', code: 'N5', title: 'Testing Frontend', description: 'Unit, component, E2E tests', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n5', code: 'N6', title: 'Go or Node.js', description: 'Pick backend language', node_type: 'decision', status: 'mastered', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n6', code: 'N7', title: 'REST API Design', description: 'HTTP, status codes, OpenAPI', node_type: 'step', status: 'practiced', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n7', code: 'N8', title: 'Database & SQL', description: 'PostgreSQL, migrations, indexing', node_type: 'step', status: 'learning', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n8', code: 'N9', title: 'Auth & Security', description: 'JWT, OAuth2, CORS, OWASP', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n9', code: 'N10', title: 'Full-Stack Capable', description: 'Build and deploy complete web app', node_type: 'milestone', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n10', code: 'N11', title: 'Docker & CI/CD', description: 'Containers, GitHub Actions', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n11', code: 'N12', title: 'System Design', description: 'Load balancing, caching, queues', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n12', code: 'N13', title: 'Monitoring', description: 'Logging, metrics, tracing', node_type: 'info', status: '', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n13', code: 'N14', title: 'Senior Engineer Ready', description: '5+ projects, mentoring, system design', node_type: 'milestone', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n0', target_node_id: 'n1', label: 'frontend', edge_type: 'default' },
    { id: 'e2', source_node_id: 'n0', target_node_id: 'n5', label: 'backend', edge_type: 'default' },
    { id: 'e3', source_node_id: 'n1', target_node_id: 'n2', label: '', edge_type: 'default' },
    { id: 'e4', source_node_id: 'n2', target_node_id: 'n3', label: '', edge_type: 'default' },
    { id: 'e5', source_node_id: 'n3', target_node_id: 'n4', label: '', edge_type: 'default' },
    { id: 'e6', source_node_id: 'n5', target_node_id: 'n6', label: '', edge_type: 'default' },
    { id: 'e7', source_node_id: 'n6', target_node_id: 'n7', label: '', edge_type: 'default' },
    { id: 'e8', source_node_id: 'n7', target_node_id: 'n8', label: '', edge_type: 'default' },
    { id: 'e9', source_node_id: 'n4', target_node_id: 'n9', label: 'converge', edge_type: 'success' },
    { id: 'e10', source_node_id: 'n8', target_node_id: 'n9', label: 'converge', edge_type: 'success' },
    { id: 'e11', source_node_id: 'n9', target_node_id: 'n10', label: '', edge_type: 'default' },
    { id: 'e12', source_node_id: 'n10', target_node_id: 'n11', label: '', edge_type: 'default' },
    { id: 'e13', source_node_id: 'n11', target_node_id: 'n13', label: '', edge_type: 'default' },
    { id: 'e14', source_node_id: 'n12', target_node_id: 'n11', label: 'reference', edge_type: 'optional' },
  ],
}

const marketingLaunch: MockRoadmap = {
  id: 'rm-2', code: 'RM2', research_id: 'r-4',
  title: 'Q2 Product Launch',
  description: 'Go-to-market plan for analytics dashboard',
  statuses: ['planned', 'approved', 'in_progress', 'launched'],
  status: 'active',
  nodes: [
    { id: 'n1', code: 'N1', title: 'Market Research', description: 'Competitors, segments, pricing', node_type: 'step', status: 'launched', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n2', code: 'N2', title: 'Messaging & Positioning', description: 'Value prop, differentiators', node_type: 'step', status: 'approved', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n3', code: 'N3', title: 'Channel Decision', description: 'Choose distribution channel', node_type: 'decision', status: 'in_progress', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n4', code: 'N4', title: 'Product Hunt', description: 'Assets, launch day, upvotes', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n5', code: 'N5', title: 'Content Marketing', description: 'SEO, case studies, landing pages', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n6', code: 'N6', title: 'Paid Ads', description: 'Google, LinkedIn, retargeting', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n7', code: 'N7', title: 'Budget: $15K', description: 'Ads $8K, content $4K, tools $3K', node_type: 'info', status: '', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n8', code: 'N8', title: 'Beta (50 users)', description: 'Early adopters, NPS, iterate', node_type: 'milestone', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n9', code: 'N9', title: 'Public Launch', description: 'All channels coordinated', node_type: 'milestone', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n1', target_node_id: 'n2', label: 'next', edge_type: 'success' },
    { id: 'e2', source_node_id: 'n2', target_node_id: 'n3', label: '', edge_type: 'default' },
    { id: 'e3', source_node_id: 'n3', target_node_id: 'n4', label: 'if PH', edge_type: 'default' },
    { id: 'e4', source_node_id: 'n3', target_node_id: 'n5', label: 'if content', edge_type: 'default' },
    { id: 'e5', source_node_id: 'n3', target_node_id: 'n6', label: 'if paid', edge_type: 'default' },
    { id: 'e6', source_node_id: 'n7', target_node_id: 'n6', label: 'budget', edge_type: 'optional' },
    { id: 'e7', source_node_id: 'n4', target_node_id: 'n8', label: '', edge_type: 'success' },
    { id: 'e8', source_node_id: 'n5', target_node_id: 'n8', label: '', edge_type: 'success' },
    { id: 'e9', source_node_id: 'n6', target_node_id: 'n8', label: '', edge_type: 'success' },
    { id: 'e10', source_node_id: 'n8', target_node_id: 'n9', label: 'feedback OK', edge_type: 'success' },
    { id: 'e11', source_node_id: 'n8', target_node_id: 'n3', label: 'pivot', edge_type: 'warning' },
  ],
}

const projectDashboard: MockRoadmap = {
  id: 'rm-5', code: 'RM5', research_id: 'r-5',
  title: 'Product Release Dashboard',
  description: 'Cross-research roadmap linking tasks, entries, sessions, and related researches',
  statuses: ['todo', 'in_progress', 'review', 'done'],
  status: 'active',
  nodes: [
    // Root is a plain step (no ref)
    { id: 'n1', code: 'N1', title: 'Research Phase', description: 'Gather requirements and analyze the problem space', node_type: 'milestone', status: 'done', position_x: 0, position_y: 0, parent_id: '' },
    // Entry ref: architecture doc
    { id: 'n2', code: 'N2', title: 'Architecture doc', description: '', node_type: 'step', status: 'done', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'entry', ref_id: 'entry-1',
      ref_data: { title: 'Microservices Architecture Overview', status: 'final', code: 'E3', section_name: 'Architecture', content: 'Service mesh provides transparent routing, load balancing, and observability between microservices...' },
    },
    // Session ref: user interviews
    { id: 'n3', code: 'N3', title: 'User interviews', description: '', node_type: 'step', status: 'in_progress', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'session', ref_id: 'session-1',
      ref_data: { title: 'UX Deep-dive Interviews', status: 'active', code: 'SS2', description: 'Understanding pain points', total_questions: 12, answered_questions: 8 },
    },
    // Research ref: related market analysis
    { id: 'n4', code: 'N4', title: 'Market analysis', description: '', node_type: 'info', status: '', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'research', ref_id: 'research-2',
      ref_data: { title: 'Competitive Market Analysis 2025', status: 'in_progress', code: 'R3', section_count: 5, entry_count: 23 },
    },
    // Regular decision node (no ref)
    { id: 'n5', code: 'N5', title: 'Choose tech stack', description: 'Based on research findings, pick the stack', node_type: 'decision', status: 'in_progress', position_x: 0, position_y: 0, parent_id: '' },
    // Task ref: implement auth
    { id: 'n6', code: 'N6', title: 'Implement auth', description: '', node_type: 'step', status: 'in_progress', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'task', ref_id: 'task-1',
      ref_data: { title: 'Implement OAuth2 + PKCE flow', status: 'in_progress', code: 'T2', priority: 'high' },
    },
    // Task ref: CI/CD pipeline
    { id: 'n7', code: 'N7', title: 'Set up CI/CD', description: '', node_type: 'step', status: 'done', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'task', ref_id: 'task-2',
      ref_data: { title: 'Configure GitHub Actions pipeline', status: 'done', code: 'T5', priority: 'medium', result: 'Build, test, deploy stages. Avg 3m20s.' },
    },
    // Task ref: security audit (critical)
    { id: 'n8', code: 'N8', title: 'Security audit', description: '', node_type: 'step', status: 'todo', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'task', ref_id: 'task-3',
      ref_data: { title: 'Patch session token vulnerability', status: 'todo', code: 'T1', priority: 'critical' },
    },
    // Question ref
    { id: 'n9', code: 'N9', title: 'Key question', description: '', node_type: 'step', status: 'done', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'question', ref_id: 'question-1',
      ref_data: { title: 'What are the main scalability bottlenecks?', status: 'answered', code: 'Q3', description: 'DB pooling and caching are the primary bottlenecks.' },
    },
    // Entry ref: API docs
    { id: 'n10', code: 'N10', title: 'API documentation', description: '', node_type: 'step', status: 'in_progress', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'entry', ref_id: 'entry-2',
      ref_data: { title: 'REST API Reference v2', status: 'draft', code: 'E12', section_name: 'Documentation', content: '21 endpoints across research, entry, session, task, and roadmap domains. All write endpoints require bearer auth...' },
    },
    // Final milestone (no ref)
    { id: 'n11', code: 'N11', title: 'Release v1.0', description: 'All tracks converge — ship it', node_type: 'milestone', status: 'todo', position_x: 0, position_y: 0, parent_id: '' },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n1', target_node_id: 'n2', label: 'document', edge_type: 'default' },
    { id: 'e2', source_node_id: 'n1', target_node_id: 'n3', label: 'interview', edge_type: 'default' },
    { id: 'e3', source_node_id: 'n1', target_node_id: 'n4', label: 'reference', edge_type: 'optional' },
    { id: 'e4', source_node_id: 'n2', target_node_id: 'n5', label: 'informs', edge_type: 'success' },
    { id: 'e5', source_node_id: 'n3', target_node_id: 'n5', label: 'informs', edge_type: 'success' },
    { id: 'e6', source_node_id: 'n9', target_node_id: 'n5', label: 'answers', edge_type: 'success' },
    { id: 'e7', source_node_id: 'n5', target_node_id: 'n6', label: 'build', edge_type: 'default' },
    { id: 'e8', source_node_id: 'n5', target_node_id: 'n7', label: 'infra', edge_type: 'default' },
    { id: 'e9', source_node_id: 'n6', target_node_id: 'n8', label: 'then', edge_type: 'default' },
    { id: 'e10', source_node_id: 'n6', target_node_id: 'n10', label: 'document', edge_type: 'optional' },
    { id: 'e11', source_node_id: 'n7', target_node_id: 'n11', label: '', edge_type: 'success' },
    { id: 'e12', source_node_id: 'n8', target_node_id: 'n11', label: '', edge_type: 'success' },
    { id: 'e13', source_node_id: 'n10', target_node_id: 'n11', label: '', edge_type: 'success' },
  ],
}

const mixedRefRoadmap: MockRoadmap = {
  id: 'rm-6', code: 'RM6', research_id: 'r-6',
  title: 'UX Research Plan',
  description: 'Interview-driven research with linked sessions and entries',
  statuses: ['planned', 'active', 'completed'],
  status: 'active',
  nodes: [
    { id: 'n1', code: 'N1', title: 'Define research goals', description: 'Identify key questions and target users', node_type: 'step', status: 'completed', position_x: 0, position_y: 0, parent_id: '' },
    // Session: initial exploration
    { id: 'n2', code: 'N2', title: 'Initial exploration', description: '', node_type: 'step', status: 'completed', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'session', ref_id: 'ss-1',
      ref_data: { title: 'Broad Exploration Session', status: 'completed', code: 'SS1', total_questions: 6, answered_questions: 6 },
    },
    // Entry: synthesis from first session
    { id: 'n3', code: 'N3', title: 'Synthesis notes', description: '', node_type: 'step', status: 'completed', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'entry', ref_id: 'e-1',
      ref_data: { title: 'Session 1 Synthesis', status: 'final', code: 'E1', section_name: 'Findings', content: 'Users primarily struggle with onboarding flow — 7 out of 8 participants mentioned confusion at the account setup step.' },
    },
    // Session: deep-dive
    { id: 'n4', code: 'N4', title: 'Deep-dive interviews', description: '', node_type: 'step', status: 'active', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'session', ref_id: 'ss-2',
      ref_data: { title: 'Onboarding Deep-dive', status: 'active', code: 'SS2', total_questions: 10, answered_questions: 4 },
    },
    // Task: create prototype
    { id: 'n5', code: 'N5', title: 'Build prototype', description: '', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'task', ref_id: 't-1',
      ref_data: { title: 'Create Figma prototype for new onboarding', status: 'todo', code: 'T3', priority: 'high' },
    },
    // Session: usability test
    { id: 'n6', code: 'N6', title: 'Usability testing', description: '', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'session', ref_id: 'ss-3',
      ref_data: { title: 'Prototype Usability Test', status: 'planned', code: 'SS3', total_questions: 8, answered_questions: 0 },
    },
    // Entry: final report
    { id: 'n7', code: 'N7', title: 'Final report', description: '', node_type: 'milestone', status: 'planned', position_x: 0, position_y: 0, parent_id: '',
      ref_type: 'entry', ref_id: 'e-2',
      ref_data: { title: 'UX Research Final Report', status: 'draft', code: 'E8', section_name: 'Reports' },
    },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n1', target_node_id: 'n2', label: 'explore', edge_type: 'default' },
    { id: 'e2', source_node_id: 'n2', target_node_id: 'n3', label: 'synthesize', edge_type: 'success' },
    { id: 'e3', source_node_id: 'n3', target_node_id: 'n4', label: 'deep-dive', edge_type: 'default' },
    { id: 'e4', source_node_id: 'n4', target_node_id: 'n5', label: 'prototype', edge_type: 'default' },
    { id: 'e5', source_node_id: 'n5', target_node_id: 'n6', label: 'test', edge_type: 'default' },
    { id: 'e6', source_node_id: 'n6', target_node_id: 'n7', label: 'report', edge_type: 'success' },
    { id: 'e7', source_node_id: 'n4', target_node_id: 'n7', label: 'if blocked', edge_type: 'warning' },
  ],
}

// =============================================================================
const meta: Meta<typeof RoadmapPageView> = {
  title: 'Pages/Roadmap',
  component: RoadmapPageView,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
}
export default meta
type Story = StoryObj<typeof RoadmapPageView>

export const LinearLearningPath: Story = {
  args: { roadmap: vueLearningPath, researchName: 'Frontend Research' },
}

export const BranchingDecisionTree: Story = {
  args: { roadmap: frameworkDecision, researchName: 'Architecture Research' },
}

export const ParallelTracks: Story = {
  args: { roadmap: fullStackRoadmap, researchName: 'Career Development' },
}

export const MarketingStrategy: Story = {
  args: { roadmap: marketingLaunch, researchName: 'Product Analytics' },
}

export const ProjectDashboardWithRefs: Story = {
  name: 'Project Dashboard (entity references)',
  args: { roadmap: projectDashboard, researchName: 'Product Release' },
}

export const UXResearchPlan: Story = {
  name: 'UX Research Plan (sessions + entries)',
  args: { roadmap: mixedRefRoadmap, researchName: 'Onboarding UX Research' },
}
