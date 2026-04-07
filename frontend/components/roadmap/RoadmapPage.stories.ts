import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, computed } from 'vue'
import RoadmapStepNode from './RoadmapStepNode.vue'
import RoadmapRootNode from './RoadmapRootNode.vue'
import RoadmapEdgeLabel from './RoadmapEdgeLabel.vue'
import './RoadmapPageMock.css'

/**
 * Complete roadmap page mocks showing various graph topologies:
 * linear flows, branching decisions, parallel tracks, and convergence points.
 */

// --- Helper: Toolbar ---
const toolbarTemplate = (researchName: string, rmCode: string, rmTitle: string, completedKey: string) => `
  <div class="rm-toolbar">
    <div class="rm-toolbar-left">
      <button class="btn btn-sm rm-toolbar-back">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
        Back
      </button>
      <span class="rm-toolbar-title">${researchName}</span>
      <span class="rm-toolbar-sep"></span>
      <span class="rm-toolbar-roadmap-name">${rmCode} — ${rmTitle}</span>
    </div>
    <div class="rm-toolbar-right">
      <div class="rm-progress">
        <span class="rm-progress-label">{{ statusCounts['${completedKey}'] || 0 }}/{{ totalNodes }} ${completedKey}</span>
        <div class="rm-progress-bar">
          <div class="rm-progress-fill" :style="{ width: ((statusCounts['${completedKey}'] || 0) / totalNodes * 100) + '%' }"></div>
        </div>
      </div>
    </div>
  </div>
`

function countStatuses(nodes: any[]) {
  const counts: Record<string, number> = {}
  for (const n of nodes) {
    if (n.nodeType === 'root') continue
    const s = n.status || 'no_status'
    counts[s] = (counts[s] || 0) + 1
  }
  return counts
}

// =============================================================================
// Story 1: Linear learning path (simple chain)
// =============================================================================
const VueLearningPathComponent = defineComponent({
  name: 'VueLearningPath',
  components: { RoadmapStepNode, RoadmapRootNode, RoadmapEdgeLabel },
  setup() {
    const nodes = [
      { id: 'root', code: 'RM1', title: 'Vue 3 Learning Path', description: 'From zero to production-ready Vue 3 development', status: 'active', statuses: ['not_started', 'in_progress', 'completed'], nodeCount: 9, edgeCount: 9, nodeType: 'root' },
      { id: 'n1', code: 'N1', title: 'HTML & CSS Fundamentals', description: 'Semantic HTML, Flexbox, Grid, responsive design', nodeType: 'step', status: 'completed' },
      { id: 'n2', code: 'N2', title: 'JavaScript ES6+', description: 'Arrow functions, destructuring, promises, async/await', nodeType: 'step', status: 'completed' },
      { id: 'n3', code: 'N3', title: 'Fundamentals Complete', description: 'Core web skills mastered', nodeType: 'milestone', status: 'completed' },
      { id: 'n4', code: 'N4', title: 'Vue 3 Basics', description: 'Template syntax, reactivity, components, props, events', nodeType: 'step', status: 'in_progress' },
      { id: 'n5', code: 'N5', title: 'Composition API', description: 'ref, reactive, computed, watch, lifecycle hooks', nodeType: 'step', status: 'not_started' },
      { id: 'n6', code: 'N6', title: 'State Management', description: 'Pinia store patterns, composable-based state', nodeType: 'step', status: 'not_started' },
      { id: 'n7', code: 'N7', title: 'Testing & Deployment', description: 'Vitest, Vue Test Utils, CI/CD, Vercel', nodeType: 'step', status: 'not_started' },
      { id: 'n8', code: 'N8', title: 'Production Ready', description: 'Ship your first Vue 3 app', nodeType: 'milestone', status: 'not_started' },
    ]
    const edges = [
      { label: '', edgeType: 'default' },
      { label: 'next', edgeType: 'default' },
      { label: 'next', edgeType: 'success' },
      { label: 'next', edgeType: 'default' },
      { label: 'next', edgeType: 'default' },
      { label: 'next', edgeType: 'default' },
      { label: 'next', edgeType: 'default' },
      { label: 'ship it', edgeType: 'success' },
    ]
    const statusCounts = computed(() => countStatuses(nodes))
    const totalNodes = nodes.filter(n => n.nodeType !== 'root').length
    return { nodes, edges, statusCounts, totalNodes }
  },
  template: `
    <div class="roadmap-page-mock">
      ${toolbarTemplate('Frontend Research', 'RM1', 'Vue 3 Learning Path', 'completed')}
      <div class="rm-canvas">
        <div class="rm-graph-root"><RoadmapRootNode :data="nodes[0]" /></div>
        <div class="rm-connector"><svg width="2" height="32"><line x1="1" y1="0" x2="1" y2="32" stroke="var(--color-border-strong)" stroke-width="2" /></svg></div>
        <div class="rm-graph-flow">
          <template v-for="(node, i) in nodes.slice(1)" :key="node.id">
            <div class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div v-if="edges[i] && edges[i].label" class="rm-edge-label-wrapper">
                <svg width="2" height="12"><line x1="1" y1="0" x2="1" y2="12" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg>
                <RoadmapEdgeLabel :label="edges[i].label" :edge-type="edges[i].edgeType" />
                <svg width="2" height="12"><line x1="1" y1="0" x2="1" y2="12" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg>
              </div>
              <div v-else-if="i < nodes.length - 2" class="rm-connector-sm">
                <svg width="2" height="20"><line x1="1" y1="0" x2="1" y2="20" stroke="var(--color-border)" stroke-width="1" /></svg>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  `,
})

// =============================================================================
// Story 2: Branching decision tree (fork + converge)
// =============================================================================
const DecisionTreeComponent = defineComponent({
  name: 'DecisionTree',
  components: { RoadmapStepNode, RoadmapRootNode, RoadmapEdgeLabel },
  setup() {
    const rootNode = { id: 'root', code: 'RM3', title: 'Frontend Framework Decision', description: 'Evaluate and choose the right framework for the project', status: 'active', statuses: ['not_started', 'evaluating', 'decided'], nodeCount: 11, edgeCount: 13, nodeType: 'root' }
    const prereqNodes = [
      { id: 'n1', code: 'N1', title: 'Define Requirements', description: 'Performance needs, team size, SSR requirement, ecosystem', nodeType: 'step', status: 'decided' },
      { id: 'n2', code: 'N2', title: 'Framework Shortlist', description: 'Narrowed down to 3 candidates based on requirements', nodeType: 'milestone', status: 'decided' },
    ]
    const branchANodes = [
      { id: 'n3', code: 'N3', title: 'Evaluate React 19', description: 'Server components, suspense, concurrent features', nodeType: 'step', status: 'evaluating' },
      { id: 'n4', code: 'N4', title: 'Build React POC', description: 'Prototype dashboard with Next.js App Router', nodeType: 'step', status: 'not_started' },
    ]
    const branchBNodes = [
      { id: 'n5', code: 'N5', title: 'Evaluate Vue 3', description: 'Composition API, Nuxt 4, Volar DX', nodeType: 'step', status: 'evaluating' },
      { id: 'n6', code: 'N6', title: 'Build Vue POC', description: 'Prototype dashboard with Nuxt 4', nodeType: 'step', status: 'not_started' },
    ]
    const branchCNodes = [
      { id: 'n7', code: 'N7', title: 'Evaluate Svelte 5', description: 'Runes, compiled output, SvelteKit', nodeType: 'step', status: 'evaluating' },
      { id: 'n8', code: 'N8', title: 'Build Svelte POC', description: 'Prototype dashboard with SvelteKit', nodeType: 'step', status: 'not_started' },
    ]
    const convergeNodes = [
      { id: 'n9', code: 'N9', title: 'Compare Benchmarks', description: 'Bundle size, lighthouse score, DX survey, build time', nodeType: 'step', status: 'not_started' },
      { id: 'n10', code: 'N10', title: 'Team Vote', description: 'Present findings, collect preferences, make final call', nodeType: 'decision', status: 'not_started' },
      { id: 'n11', code: 'N11', title: 'Framework Chosen', description: 'Final decision documented and communicated', nodeType: 'milestone', status: 'not_started' },
    ]
    const allNodes = [rootNode, ...prereqNodes, ...branchANodes, ...branchBNodes, ...branchCNodes, ...convergeNodes]
    const statusCounts = computed(() => countStatuses(allNodes))
    const totalNodes = allNodes.filter(n => n.nodeType !== 'root').length
    return { rootNode, prereqNodes, branchANodes, branchBNodes, branchCNodes, convergeNodes, statusCounts, totalNodes }
  },
  template: `
    <div class="roadmap-page-mock">
      ${toolbarTemplate('Architecture Research', 'RM3', 'Frontend Framework Decision', 'decided')}
      <div class="rm-canvas">
        <div class="rm-graph-root"><RoadmapRootNode :data="rootNode" /></div>
        <div class="rm-connector"><svg width="2" height="24"><line x1="1" y1="0" x2="1" y2="24" stroke="var(--color-border-strong)" stroke-width="2" /></svg></div>

        <!-- Prerequisites (linear) -->
        <div class="rm-graph-flow">
          <div v-for="node in prereqNodes" :key="node.id" class="rm-graph-item">
            <RoadmapStepNode :data="node" />
            <div class="rm-connector-sm"><svg width="2" height="20"><line x1="1" y1="0" x2="1" y2="20" stroke="var(--color-border)" stroke-width="1" /></svg></div>
          </div>
        </div>

        <!-- Branch label -->
        <div class="rm-section-divider">
          <div class="rm-section-divider-line"></div>
          <span class="rm-section-divider-text">evaluate in parallel</span>
          <div class="rm-section-divider-line"></div>
        </div>

        <!-- 3-way branch -->
        <div class="rm-branch-row">
          <div class="rm-branch-col">
            <span class="rm-branch-label">Option A</span>
            <div v-for="node in branchANodes" :key="node.id" class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div class="rm-connector-sm"><svg width="2" height="16"><line x1="1" y1="0" x2="1" y2="16" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg></div>
            </div>
          </div>
          <div class="rm-branch-col">
            <span class="rm-branch-label">Option B</span>
            <div v-for="node in branchBNodes" :key="node.id" class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div class="rm-connector-sm"><svg width="2" height="16"><line x1="1" y1="0" x2="1" y2="16" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg></div>
            </div>
          </div>
          <div class="rm-branch-col">
            <span class="rm-branch-label">Option C</span>
            <div v-for="node in branchCNodes" :key="node.id" class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div class="rm-connector-sm"><svg width="2" height="16"><line x1="1" y1="0" x2="1" y2="16" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg></div>
            </div>
          </div>
        </div>

        <!-- Converge label -->
        <div class="rm-section-divider">
          <div class="rm-section-divider-line"></div>
          <span class="rm-section-divider-text">converge &amp; compare</span>
          <div class="rm-section-divider-line"></div>
        </div>

        <!-- Converge nodes -->
        <div class="rm-graph-flow">
          <div v-for="node in convergeNodes" :key="node.id" class="rm-graph-item">
            <RoadmapStepNode :data="node" />
            <div class="rm-connector-sm"><svg width="2" height="20"><line x1="1" y1="0" x2="1" y2="20" stroke="var(--color-border)" stroke-width="1" /></svg></div>
          </div>
        </div>
      </div>
    </div>
  `,
})

// =============================================================================
// Story 3: Parallel tracks with shared milestones (complex graph)
// =============================================================================
const ParallelTracksComponent = defineComponent({
  name: 'ParallelTracks',
  components: { RoadmapStepNode, RoadmapRootNode, RoadmapEdgeLabel },
  setup() {
    const rootNode = { id: 'root', code: 'RM4', title: 'Full-Stack Developer Roadmap', description: 'Parallel frontend and backend learning tracks converging at milestones', status: 'active', statuses: ['not_started', 'learning', 'practiced', 'mastered'], nodeCount: 14, edgeCount: 16, nodeType: 'root' }
    const sharedStart = [
      { id: 'n0', code: 'N1', title: 'Git & Terminal Basics', description: 'Version control, CLI navigation, SSH keys', nodeType: 'step', status: 'mastered' },
    ]
    const frontendTrack = [
      { id: 'n1', code: 'N2', title: 'HTML/CSS/JS', description: 'Semantic HTML, modern CSS, ES2024', nodeType: 'step', status: 'mastered' },
      { id: 'n2', code: 'N3', title: 'React or Vue', description: 'Pick one framework, build 3 projects', nodeType: 'decision', status: 'practiced' },
      { id: 'n3', code: 'N4', title: 'State & Routing', description: 'Client-side state management, SPA routing', nodeType: 'step', status: 'learning' },
      { id: 'n4', code: 'N5', title: 'Testing Frontend', description: 'Unit tests, component tests, E2E with Playwright', nodeType: 'step', status: 'not_started' },
    ]
    const backendTrack = [
      { id: 'n5', code: 'N6', title: 'Go or Node.js', description: 'Pick one backend language, learn fundamentals', nodeType: 'decision', status: 'mastered' },
      { id: 'n6', code: 'N7', title: 'REST API Design', description: 'HTTP methods, status codes, JSON:API, OpenAPI', nodeType: 'step', status: 'practiced' },
      { id: 'n7', code: 'N8', title: 'Database & SQL', description: 'PostgreSQL, migrations, indexing, query optimization', nodeType: 'step', status: 'learning' },
      { id: 'n8', code: 'N9', title: 'Auth & Security', description: 'JWT, OAuth2, CORS, rate limiting, OWASP', nodeType: 'step', status: 'not_started' },
    ]
    const milestoneCheckpoint = { id: 'n9', code: 'N10', title: 'Full-Stack Capable', description: 'Can build and deploy a complete web app independently', nodeType: 'milestone', status: 'not_started' }
    const advancedTrack = [
      { id: 'n10', code: 'N11', title: 'Docker & CI/CD', description: 'Containerization, GitHub Actions, automated deploys', nodeType: 'step', status: 'not_started' },
      { id: 'n11', code: 'N12', title: 'System Design Basics', description: 'Load balancing, caching, queues, microservices intro', nodeType: 'step', status: 'not_started' },
      { id: 'n12', code: 'N13', title: 'Monitoring & Observability', description: 'Logging, metrics, tracing, alerting', nodeType: 'info', status: '' },
      { id: 'n13', code: 'N14', title: 'Senior Engineer Ready', description: 'Portfolio of 5+ projects, mentoring experience, system design skills', nodeType: 'milestone', status: 'not_started' },
    ]
    const allNodes = [rootNode, ...sharedStart, ...frontendTrack, ...backendTrack, milestoneCheckpoint, ...advancedTrack]
    const statusCounts = computed(() => countStatuses(allNodes))
    const totalNodes = allNodes.filter(n => n.nodeType !== 'root').length
    return { rootNode, sharedStart, frontendTrack, backendTrack, milestoneCheckpoint, advancedTrack, statusCounts, totalNodes }
  },
  template: `
    <div class="roadmap-page-mock">
      ${toolbarTemplate('Career Development Research', 'RM4', 'Full-Stack Developer Roadmap', 'mastered')}
      <div class="rm-canvas">
        <div class="rm-graph-root"><RoadmapRootNode :data="rootNode" /></div>
        <div class="rm-connector"><svg width="2" height="24"><line x1="1" y1="0" x2="1" y2="24" stroke="var(--color-border-strong)" stroke-width="2" /></svg></div>

        <!-- Shared start -->
        <div class="rm-graph-flow">
          <div v-for="node in sharedStart" :key="node.id" class="rm-graph-item">
            <RoadmapStepNode :data="node" />
            <div class="rm-connector-sm"><svg width="2" height="20"><line x1="1" y1="0" x2="1" y2="20" stroke="var(--color-border)" stroke-width="1" /></svg></div>
          </div>
        </div>

        <div class="rm-section-divider">
          <div class="rm-section-divider-line"></div>
          <span class="rm-section-divider-text">learn in parallel</span>
          <div class="rm-section-divider-line"></div>
        </div>

        <!-- Two parallel tracks -->
        <div class="rm-branch-row">
          <div class="rm-branch-col">
            <span class="rm-branch-label">Frontend Track</span>
            <div v-for="node in frontendTrack" :key="node.id" class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div class="rm-connector-sm"><svg width="2" height="14"><line x1="1" y1="0" x2="1" y2="14" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg></div>
            </div>
          </div>
          <div class="rm-branch-col">
            <span class="rm-branch-label">Backend Track</span>
            <div v-for="node in backendTrack" :key="node.id" class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div class="rm-connector-sm"><svg width="2" height="14"><line x1="1" y1="0" x2="1" y2="14" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg></div>
            </div>
          </div>
        </div>

        <!-- Converge milestone -->
        <div class="rm-section-divider">
          <div class="rm-section-divider-line"></div>
          <span class="rm-section-divider-text">both tracks converge</span>
          <div class="rm-section-divider-line"></div>
        </div>

        <div class="rm-graph-flow">
          <div class="rm-graph-item">
            <RoadmapStepNode :data="milestoneCheckpoint" />
            <div class="rm-connector-sm"><svg width="2" height="20"><line x1="1" y1="0" x2="1" y2="20" stroke="var(--color-border)" stroke-width="1" /></svg></div>
          </div>
        </div>

        <div class="rm-section-divider">
          <div class="rm-section-divider-line"></div>
          <span class="rm-section-divider-text">advanced topics</span>
          <div class="rm-section-divider-line"></div>
        </div>

        <div class="rm-graph-flow">
          <div v-for="node in advancedTrack" :key="node.id" class="rm-graph-item">
            <RoadmapStepNode :data="node" />
            <div class="rm-connector-sm"><svg width="2" height="16"><line x1="1" y1="0" x2="1" y2="16" stroke="var(--color-border)" stroke-width="1" /></svg></div>
          </div>
        </div>
      </div>
    </div>
  `,
})

// =============================================================================
// Story 4: Marketing strategy with decision + multiple edges from one node
// =============================================================================
const MarketingRoadmapComponent = defineComponent({
  name: 'MarketingRoadmap',
  components: { RoadmapStepNode, RoadmapRootNode, RoadmapEdgeLabel },
  setup() {
    const rootNode = { id: 'root', code: 'RM2', title: 'Q2 Product Launch', description: 'Go-to-market plan for analytics dashboard', status: 'active', statuses: ['planned', 'approved', 'in_progress', 'launched'], nodeCount: 10, edgeCount: 12, nodeType: 'root' }
    const researchPhase = [
      { id: 'n1', code: 'N1', title: 'Market Research', description: 'Competitors, target segments, pricing validation', nodeType: 'step', status: 'launched' },
      { id: 'n2', code: 'N2', title: 'Messaging & Positioning', description: 'Value proposition, differentiators, elevator pitch', nodeType: 'step', status: 'approved' },
    ]
    const channelDecision = { id: 'n3', code: 'N3', title: 'Launch Channel Decision', description: 'Choose primary distribution channel', nodeType: 'decision', status: 'in_progress' }
    const channelANodes = [
      { id: 'n4', code: 'N4', title: 'Product Hunt Launch', description: 'Prepare assets, schedule launch day, coordinate upvotes', nodeType: 'step', status: 'planned' },
      { id: 'n5', code: 'N5', title: 'PH Follow-up Campaign', description: 'Email sequence to PH visitors, retargeting ads', nodeType: 'step', status: 'planned' },
    ]
    const channelBNodes = [
      { id: 'n6', code: 'N6', title: 'Content Marketing', description: 'SEO blog posts, case studies, comparison pages', nodeType: 'step', status: 'planned' },
      { id: 'n7', code: 'N7', title: 'Paid Ads Campaign', description: 'Google Ads, LinkedIn sponsored, retargeting', nodeType: 'step', status: 'planned' },
    ]
    const budgetInfo = { id: 'n8', code: 'N8', title: 'Budget: $15K allocated', description: 'Covers ads ($8K), content ($4K), tooling ($3K)', nodeType: 'info', status: '' }
    const finalPhase = [
      { id: 'n9', code: 'N9', title: 'Beta Program (50 users)', description: 'Onboard early adopters, collect NPS, iterate on feedback', nodeType: 'milestone', status: 'planned' },
      { id: 'n10', code: 'N10', title: 'Public Launch', description: 'Coordinated launch across all chosen channels', nodeType: 'milestone', status: 'planned' },
    ]
    const allNodes = [rootNode, ...researchPhase, channelDecision, ...channelANodes, ...channelBNodes, budgetInfo, ...finalPhase]
    const statusCounts = computed(() => countStatuses(allNodes))
    const totalNodes = allNodes.filter(n => n.nodeType !== 'root').length
    return { rootNode, researchPhase, channelDecision, channelANodes, channelBNodes, budgetInfo, finalPhase, statusCounts, totalNodes }
  },
  template: `
    <div class="roadmap-page-mock">
      ${toolbarTemplate('Product Analytics Research', 'RM2', 'Q2 Product Launch', 'launched')}
      <div class="rm-canvas">
        <div class="rm-graph-root"><RoadmapRootNode :data="rootNode" /></div>
        <div class="rm-connector"><svg width="2" height="24"><line x1="1" y1="0" x2="1" y2="24" stroke="var(--color-border-strong)" stroke-width="2" /></svg></div>

        <!-- Research phase -->
        <div class="rm-graph-flow">
          <div v-for="node in researchPhase" :key="node.id" class="rm-graph-item">
            <RoadmapStepNode :data="node" />
            <div class="rm-edge-label-wrapper">
              <svg width="2" height="8"><line x1="1" y1="0" x2="1" y2="8" stroke="var(--color-border)" stroke-width="1" /></svg>
              <RoadmapEdgeLabel label="next" edge-type="success" />
              <svg width="2" height="8"><line x1="1" y1="0" x2="1" y2="8" stroke="var(--color-border)" stroke-width="1" /></svg>
            </div>
          </div>

          <!-- Decision node -->
          <div class="rm-graph-item">
            <RoadmapStepNode :data="channelDecision" />
          </div>
        </div>

        <!-- Fork: two channel strategies -->
        <div class="rm-section-divider">
          <div class="rm-section-divider-line"></div>
          <span class="rm-section-divider-text">choose one or both</span>
          <div class="rm-section-divider-line"></div>
        </div>

        <div class="rm-branch-row">
          <div class="rm-branch-col">
            <span class="rm-branch-label">Product Hunt path</span>
            <div v-for="node in channelANodes" :key="node.id" class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div class="rm-connector-sm"><svg width="2" height="14"><line x1="1" y1="0" x2="1" y2="14" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg></div>
            </div>
          </div>
          <div class="rm-branch-col">
            <span class="rm-branch-label">Content + Ads path</span>
            <div v-for="node in channelBNodes" :key="node.id" class="rm-graph-item">
              <RoadmapStepNode :data="node" />
              <div class="rm-connector-sm"><svg width="2" height="14"><line x1="1" y1="0" x2="1" y2="14" stroke="var(--color-border)" stroke-width="1" stroke-dasharray="3 3" /></svg></div>
            </div>
          </div>
        </div>

        <!-- Info node (budget) — attached to the side -->
        <div class="rm-graph-flow">
          <div class="rm-graph-item">
            <RoadmapStepNode :data="budgetInfo" />
          </div>
        </div>

        <!-- Converge to final -->
        <div class="rm-section-divider">
          <div class="rm-section-divider-line"></div>
          <span class="rm-section-divider-text">both paths lead here</span>
          <div class="rm-section-divider-line"></div>
        </div>

        <div class="rm-graph-flow">
          <div v-for="(node, i) in finalPhase" :key="node.id" class="rm-graph-item">
            <RoadmapStepNode :data="node" />
            <div v-if="i === 0" class="rm-edge-label-wrapper">
              <svg width="2" height="8"><line x1="1" y1="0" x2="1" y2="8" stroke="var(--color-border)" stroke-width="1" /></svg>
              <RoadmapEdgeLabel label="feedback collected" edge-type="success" />
              <svg width="2" height="8"><line x1="1" y1="0" x2="1" y2="8" stroke="var(--color-border)" stroke-width="1" /></svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
})

// =============================================================================
// Meta + Exports
// =============================================================================
const meta: Meta<typeof VueLearningPathComponent> = {
  title: 'Pages/Roadmap',
  component: VueLearningPathComponent,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
}
export default meta
type Story = StoryObj<typeof VueLearningPathComponent>

export const LinearLearningPath: Story = {}

export const BranchingDecisionTree: Story = {
  render: () => ({
    components: { DecisionTreeComponent },
    template: '<DecisionTreeComponent />',
  }),
}

export const ParallelTracks: Story = {
  render: () => ({
    components: { ParallelTracksComponent },
    template: '<ParallelTracksComponent />',
  }),
}

export const MarketingStrategy: Story = {
  render: () => ({
    components: { MarketingRoadmapComponent },
    template: '<MarketingRoadmapComponent />',
  }),
}
