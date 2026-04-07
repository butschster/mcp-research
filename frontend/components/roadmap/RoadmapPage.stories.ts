import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, computed, ref, reactive } from 'vue'
import RoadmapStepNode from './RoadmapStepNode.vue'
import RoadmapRootNode from './RoadmapRootNode.vue'
import RoadmapNodePopover from './RoadmapNodePopover.vue'

/**
 * Roadmap page stories using API-shaped JSON data.
 * Each story provides a mock API response (roadmap with nodes[] and edges[])
 * and renders it as the real page would.
 */

// --- API response shape (matches GET /api/roadmaps/{id}) ---
interface MockRoadmap {
  id: string
  code: string
  research_id: string
  title: string
  description: string
  statuses: string[]
  status: string
  nodes: Array<{
    id: string
    code: string
    title: string
    description: string
    node_type: string
    status: string
    position_x: number
    position_y: number
    parent_id: string
  }>
  edges: Array<{
    id: string
    source_node_id: string
    target_node_id: string
    label: string
    edge_type: string
  }>
}

// --- Reusable page component that renders any MockRoadmap ---
const RoadmapPageView = defineComponent({
  name: 'RoadmapPageView',
  components: { RoadmapStepNode, RoadmapRootNode, RoadmapNodePopover },
  props: {
    roadmap: { type: Object as () => MockRoadmap, required: true },
    researchName: { type: String, default: 'Research' },
  },
  setup(props) {
    const data = reactive({ ...props.roadmap, nodes: props.roadmap.nodes.map(n => ({ ...n })) })

    // Find root: node with no incoming edges
    const targets = new Set(data.edges.map(e => e.target_node_id))
    const rootId = data.nodes.find(n => !targets.has(n.id))?.id ?? data.nodes[0]?.id

    // Progress
    const lastStatus = data.statuses.length > 0 ? data.statuses[data.statuses.length - 1] : null
    const progress = computed(() => {
      const total = data.nodes.length
      const completed = lastStatus ? data.nodes.filter(n => n.status === lastStatus).length : 0
      return { total, completed, percent: total > 0 ? Math.round((completed / total) * 100) : 0 }
    })

    // Build adjacency for rendering order (BFS from root)
    function getOrderedNodes() {
      const adj = new Map<string, string[]>()
      for (const e of data.edges) {
        if (!adj.has(e.source_node_id)) adj.set(e.source_node_id, [])
        adj.get(e.source_node_id)!.push(e.target_node_id)
      }
      const visited = new Set<string>()
      const order: string[] = []
      const queue = [rootId]
      while (queue.length) {
        const id = queue.shift()!
        if (visited.has(id)) continue
        visited.add(id)
        order.push(id)
        for (const child of adj.get(id) ?? []) queue.push(child)
      }
      // Add any unvisited nodes
      for (const n of data.nodes) {
        if (!visited.has(n.id)) order.push(n.id)
      }
      return order
    }

    const nodeMap = computed(() => new Map(data.nodes.map(n => [n.id, n])))
    const orderedIds = getOrderedNodes()
    const rootNode = computed(() => nodeMap.value.get(rootId)!)
    const stepNodes = computed(() => orderedIds.filter(id => id !== rootId).map(id => nodeMap.value.get(id)!).filter(Boolean))

    // Edge lookup: source → edges[]
    const edgesBySource = computed(() => {
      const map = new Map<string, typeof data.edges>()
      for (const e of data.edges) {
        if (!map.has(e.source_node_id)) map.set(e.source_node_id, [])
        map.get(e.source_node_id)!.push(e)
      }
      return map
    })

    function getEdgeLabel(fromId: string, toId: string) {
      const edges = edgesBySource.value.get(fromId)
      return edges?.find(e => e.target_node_id === toId)
    }

    // Popover
    const selectedNode = ref<any>(null)
    const popoverPos = ref({ x: 0, y: 0 })

    function onNodeClick(node: any, event: MouseEvent) {
      if (node.id === rootId) return
      selectedNode.value = {
        id: node.id,
        title: node.title,
        description: node.description,
        nodeType: node.node_type,
        status: node.status,
      }
      popoverPos.value = { x: event.clientX + 12, y: event.clientY - 20 }
    }

    function onUpdateStatus(nodeId: string, status: string) {
      const n = data.nodes.find(n => n.id === nodeId)
      if (n) n.status = status
      selectedNode.value = null
    }

    return {
      data, rootId, rootNode, stepNodes, orderedIds,
      edgesBySource, getEdgeLabel, progress,
      selectedNode, popoverPos, onNodeClick, onUpdateStatus,
    }
  },
  template: `
    <div style="min-height:100vh;background:var(--color-bg);display:flex;flex-direction:column;">
      <!-- Toolbar -->
      <div style="display:flex;align-items:center;justify-content:space-between;padding:var(--space-3) var(--space-5);background:rgba(21,29,46,0.9);backdrop-filter:blur(12px);border-bottom:1px solid var(--color-border);gap:var(--space-4);">
        <div style="display:flex;align-items:center;gap:var(--space-3);">
          <button class="btn btn-sm" style="gap:var(--space-1);">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            Back
          </button>
          <span style="font-size:var(--type-sm);font-weight:600;color:var(--color-text);letter-spacing:-0.01em;">{{ researchName }}</span>
          <span style="width:1px;height:20px;background:var(--color-border-strong);"></span>
          <span style="font-size:var(--type-xs);color:var(--color-primary);font-weight:500;">{{ data.code }} — {{ data.title }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:var(--space-2);">
          <span style="font-size:var(--type-xs);color:var(--color-text-muted);font-variant-numeric:tabular-nums;">{{ progress.completed }}/{{ progress.total }}</span>
          <div style="width:100px;height:4px;background:var(--color-surface-hover);border-radius:2px;overflow:hidden;">
            <div :style="{ width: progress.percent + '%', height: '100%', background: 'rgba(107,203,119,0.8)', borderRadius: '2px', transition: 'width 0.3s' }"></div>
          </div>
          <span style="width:1px;height:20px;background:var(--color-border-strong);margin:0 var(--space-1);"></span>
          <button class="btn btn-sm active"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14"/><path d="m19 12-7 7-7-7"/></svg></button>
          <button class="btn btn-sm"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg></button>
          <span style="width:1px;height:20px;background:var(--color-border-strong);margin:0 var(--space-1);"></span>
          <button class="btn btn-sm">Auto layout</button>
          <button class="btn btn-sm"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg></button>
        </div>
      </div>

      <!-- Graph -->
      <div style="flex:1;display:flex;flex-direction:column;align-items:center;padding:var(--space-6) var(--space-4);overflow-y:auto;gap:var(--space-2);">
        <!-- Root -->
        <RoadmapRootNode :data="{
          code: data.code,
          title: data.title,
          description: data.description,
          status: data.status,
          statuses: data.statuses,
          nodeCount: data.nodes.length,
          edgeCount: data.edges.length,
        }" />

        <!-- Step nodes in traversal order -->
        <template v-for="(node, i) in stepNodes" :key="node.id">
          <div style="display:flex;flex-direction:column;align-items:center;" @click="onNodeClick(node, $event)">
            <RoadmapStepNode :data="{
              code: node.code,
              title: node.title,
              description: node.description,
              nodeType: node.node_type,
              status: node.status,
            }" />
          </div>
        </template>
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
    { id: 'n1', code: 'N1', title: 'Define Requirements', description: 'Performance needs, team size, SSR requirement, ecosystem', node_type: 'step', status: 'decided', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n2', code: 'N2', title: 'Framework Shortlist', description: 'Narrowed down to 3 candidates based on requirements', node_type: 'milestone', status: 'decided', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n3', code: 'N3', title: 'Evaluate React 19', description: 'Server components, suspense, concurrent features', node_type: 'step', status: 'evaluating', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n4', code: 'N4', title: 'Evaluate Vue 3', description: 'Composition API, Nuxt 4, Volar DX', node_type: 'step', status: 'evaluating', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n5', code: 'N5', title: 'Evaluate Svelte 5', description: 'Runes, compiled output, SvelteKit', node_type: 'step', status: 'evaluating', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n6', code: 'N6', title: 'Build React POC', description: 'Prototype dashboard with Next.js App Router', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n7', code: 'N7', title: 'Build Vue POC', description: 'Prototype dashboard with Nuxt 4', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n8', code: 'N8', title: 'Build Svelte POC', description: 'Prototype dashboard with SvelteKit', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n9', code: 'N9', title: 'Compare Benchmarks', description: 'Bundle size, lighthouse score, DX survey, build time', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n10', code: 'N10', title: 'Team Vote', description: 'Present findings, collect preferences, make final call', node_type: 'decision', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n11', code: 'N11', title: 'Framework Chosen', description: 'Final decision documented and communicated', node_type: 'milestone', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
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
    { id: 'n0', code: 'N1', title: 'Git & Terminal Basics', description: 'Version control, CLI navigation, SSH keys', node_type: 'step', status: 'mastered', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n1', code: 'N2', title: 'HTML/CSS/JS', description: 'Semantic HTML, modern CSS, ES2024', node_type: 'step', status: 'mastered', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n2', code: 'N3', title: 'React or Vue', description: 'Pick one framework, build 3 projects', node_type: 'decision', status: 'practiced', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n3', code: 'N4', title: 'State & Routing', description: 'Client-side state management, SPA routing', node_type: 'step', status: 'learning', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n4', code: 'N5', title: 'Testing Frontend', description: 'Unit tests, component tests, E2E with Playwright', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n5', code: 'N6', title: 'Go or Node.js', description: 'Pick one backend language, learn fundamentals', node_type: 'decision', status: 'mastered', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n6', code: 'N7', title: 'REST API Design', description: 'HTTP methods, status codes, JSON:API, OpenAPI', node_type: 'step', status: 'practiced', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n7', code: 'N8', title: 'Database & SQL', description: 'PostgreSQL, migrations, indexing, query optimization', node_type: 'step', status: 'learning', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n8', code: 'N9', title: 'Auth & Security', description: 'JWT, OAuth2, CORS, rate limiting, OWASP', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n9', code: 'N10', title: 'Full-Stack Capable', description: 'Can build and deploy a complete web app independently', node_type: 'milestone', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n10', code: 'N11', title: 'Docker & CI/CD', description: 'Containerization, GitHub Actions, automated deploys', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n11', code: 'N12', title: 'System Design Basics', description: 'Load balancing, caching, queues, microservices intro', node_type: 'step', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n12', code: 'N13', title: 'Monitoring & Observability', description: 'Logging, metrics, tracing, alerting', node_type: 'info', status: '', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n13', code: 'N14', title: 'Senior Engineer Ready', description: 'Portfolio of 5+ projects, mentoring, system design skills', node_type: 'milestone', status: 'not_started', position_x: 0, position_y: 0, parent_id: '' },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n0', target_node_id: 'n1', label: 'frontend track', edge_type: 'default' },
    { id: 'e2', source_node_id: 'n0', target_node_id: 'n5', label: 'backend track', edge_type: 'default' },
    { id: 'e3', source_node_id: 'n1', target_node_id: 'n2', label: 'next', edge_type: 'default' },
    { id: 'e4', source_node_id: 'n2', target_node_id: 'n3', label: 'next', edge_type: 'default' },
    { id: 'e5', source_node_id: 'n3', target_node_id: 'n4', label: 'next', edge_type: 'default' },
    { id: 'e6', source_node_id: 'n5', target_node_id: 'n6', label: 'next', edge_type: 'default' },
    { id: 'e7', source_node_id: 'n6', target_node_id: 'n7', label: 'next', edge_type: 'default' },
    { id: 'e8', source_node_id: 'n7', target_node_id: 'n8', label: 'next', edge_type: 'default' },
    { id: 'e9', source_node_id: 'n4', target_node_id: 'n9', label: 'converge', edge_type: 'success' },
    { id: 'e10', source_node_id: 'n8', target_node_id: 'n9', label: 'converge', edge_type: 'success' },
    { id: 'e11', source_node_id: 'n9', target_node_id: 'n10', label: 'next', edge_type: 'default' },
    { id: 'e12', source_node_id: 'n10', target_node_id: 'n11', label: 'next', edge_type: 'default' },
    { id: 'e13', source_node_id: 'n11', target_node_id: 'n13', label: 'next', edge_type: 'default' },
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
    { id: 'n1', code: 'N1', title: 'Market Research', description: 'Competitors, target segments, pricing validation', node_type: 'step', status: 'launched', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n2', code: 'N2', title: 'Messaging & Positioning', description: 'Value proposition, differentiators, elevator pitch', node_type: 'step', status: 'approved', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n3', code: 'N3', title: 'Launch Channel Decision', description: 'Choose primary distribution channel', node_type: 'decision', status: 'in_progress', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n4', code: 'N4', title: 'Product Hunt Launch', description: 'Prepare assets, schedule launch day', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n5', code: 'N5', title: 'Content Marketing', description: 'SEO blog posts, case studies, comparison pages', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n6', code: 'N6', title: 'Paid Ads Campaign', description: 'Google Ads, LinkedIn sponsored, retargeting', node_type: 'step', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n7', code: 'N7', title: 'Budget: $15K allocated', description: 'Covers ads ($8K), content ($4K), tooling ($3K)', node_type: 'info', status: '', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n8', code: 'N8', title: 'Beta Program (50 users)', description: 'Onboard early adopters, collect NPS, iterate', node_type: 'milestone', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
    { id: 'n9', code: 'N9', title: 'Public Launch', description: 'Coordinated launch across all chosen channels', node_type: 'milestone', status: 'planned', position_x: 0, position_y: 0, parent_id: '' },
  ],
  edges: [
    { id: 'e1', source_node_id: 'n1', target_node_id: 'n2', label: 'next', edge_type: 'success' },
    { id: 'e2', source_node_id: 'n2', target_node_id: 'n3', label: 'next', edge_type: 'default' },
    { id: 'e3', source_node_id: 'n3', target_node_id: 'n4', label: 'if PH', edge_type: 'default' },
    { id: 'e4', source_node_id: 'n3', target_node_id: 'n5', label: 'if content', edge_type: 'default' },
    { id: 'e5', source_node_id: 'n3', target_node_id: 'n6', label: 'if paid', edge_type: 'default' },
    { id: 'e6', source_node_id: 'n7', target_node_id: 'n6', label: 'budget for', edge_type: 'optional' },
    { id: 'e7', source_node_id: 'n4', target_node_id: 'n8', label: 'leads to', edge_type: 'success' },
    { id: 'e8', source_node_id: 'n5', target_node_id: 'n8', label: 'leads to', edge_type: 'success' },
    { id: 'e9', source_node_id: 'n6', target_node_id: 'n8', label: 'leads to', edge_type: 'success' },
    { id: 'e10', source_node_id: 'n8', target_node_id: 'n9', label: 'feedback OK', edge_type: 'success' },
    { id: 'e11', source_node_id: 'n8', target_node_id: 'n3', label: 'pivot', edge_type: 'warning' },
  ],
}

// =============================================================================
// Meta + Exports
// =============================================================================
const meta: Meta<typeof RoadmapPageView> = {
  title: 'Pages/Roadmap',
  component: RoadmapPageView,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
}
export default meta
type Story = StoryObj<typeof RoadmapPageView>

export const LinearLearningPath: Story = {
  args: {
    roadmap: vueLearningPath,
    researchName: 'Frontend Research',
  },
}

export const BranchingDecisionTree: Story = {
  args: {
    roadmap: frameworkDecision,
    researchName: 'Architecture Research',
  },
}

export const ParallelTracks: Story = {
  args: {
    roadmap: fullStackRoadmap,
    researchName: 'Career Development Research',
  },
}

export const MarketingStrategy: Story = {
  args: {
    roadmap: marketingLaunch,
    researchName: 'Product Analytics Research',
  },
}
