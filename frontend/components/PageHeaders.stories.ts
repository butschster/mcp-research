import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent } from 'vue'
import Breadcrumbs from './Breadcrumbs.vue'
import StatusBadge from './StatusBadge.vue'

/**
 * Page headers as they appear on Research, Entry, Tasks, and Session pages.
 * These are inline page-level patterns (not standalone components),
 * documented here as a design reference.
 */

// --- Research Page Header ---
const ResearchPageHeader = defineComponent({
  name: 'ResearchPageHeader',
  components: { Breadcrumbs, StatusBadge },
  props: {
    code: { type: String, default: 'R1' },
    name: { type: String, default: 'Vue Component Architecture' },
    goal: { type: String, default: 'Investigate best practices for Vue 3 component design and composition patterns' },
    status: { type: String, default: 'active' },
    taskCount: { type: Number, default: 4 },
    showGoal: { type: Boolean, default: true },
  },
  template: `
    <div class="page-header">
      <Breadcrumbs :crumbs="[{ label: 'Research', to: '/' }, { label: name }]" />
      <div class="research-header" style="display:flex;justify-content:space-between;align-items:center;gap:var(--space-4);flex-wrap:wrap;">
        <div style="display:flex;align-items:center;gap:var(--space-3);flex-wrap:wrap;">
          <span v-if="code" style="font-size:var(--type-xs);font-weight:600;color:var(--color-primary);background:var(--color-primary-muted);padding:0.15rem 0.4rem;border-radius:4px;font-family:'JetBrains Mono',monospace;flex-shrink:0;line-height:1;">{{ code }}</span>
          <h1 class="page-title">{{ name }}</h1>
        </div>
        <div style="display:flex;align-items:center;gap:var(--space-3);flex-wrap:wrap;">
          <StatusBadge :status="status" />
          <a href="#" class="btn btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
            Tasks
            <span v-if="taskCount" style="font-size:0.75em;background:var(--color-surface-hover);padding:0.1rem 0.35rem;border-radius:3px;font-variant-numeric:tabular-nums;opacity:0.7;">{{ taskCount }}</span>
          </a>
          <a href="#" class="btn btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><circle cx="4" cy="6" r="2"/><circle cx="20" cy="6" r="2"/><circle cx="4" cy="18" r="2"/><circle cx="20" cy="18" r="2"/><path d="M9.5 10.5 5.5 7.5"/><path d="M14.5 10.5l4-3"/><path d="M9.5 13.5 5.5 16.5"/><path d="M14.5 13.5l4 3"/></svg>
            Mind map
          </a>
          <a href="#" class="btn btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Export
          </a>
          <button class="btn btn-sm" title="Show details">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          </button>
          <button class="btn btn-sm btn-danger" title="Archive">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
          </button>
        </div>
      </div>
      <p v-if="showGoal && goal" class="card-meta mt-2">{{ goal }}</p>
    </div>
  `,
})

// --- Entry Page Header ---
const EntryPageHeader = defineComponent({
  name: 'EntryPageHeader',
  components: { Breadcrumbs, StatusBadge },
  props: {
    code: { type: String, default: 'E1' },
    title: { type: String, default: 'Component Composition Patterns' },
    description: { type: String, default: 'Analysis of composable patterns in Vue 3' },
    status: { type: String, default: 'completed' },
    tags: { type: Array as () => string[], default: () => ['vue', 'composables'] },
    researchName: { type: String, default: 'Vue Component Architecture' },
    sectionName: { type: String, default: 'Fundamentals' },
  },
  setup(props) {
    function tagHue(tag: string): number {
      return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
    }
    return { tagHue }
  },
  template: `
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: '/research/R1' },
        { label: sectionName, to: '/research/R1?section=sec_001' },
        { label: title }
      ]" />
      <div style="display:flex;justify-content:space-between;align-items:flex-start;gap:var(--space-4);flex-wrap:wrap;">
        <div style="display:flex;align-items:center;gap:var(--space-3);flex-wrap:wrap;">
          <span v-if="code" style="font-size:var(--type-xs);font-weight:600;color:var(--color-primary);background:var(--color-primary-muted);padding:0.2rem 0.5rem;border-radius:4px;font-family:'JetBrains Mono',monospace;flex-shrink:0;line-height:1;">{{ code }}</span>
          <h1 class="page-title">{{ title }}</h1>
        </div>
        <div style="display:flex;gap:var(--space-2);align-items:center;flex-shrink:0;flex-wrap:wrap;">
          <StatusBadge :status="status" />
          <button class="btn btn-sm">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
            Copy
          </button>
        </div>
      </div>
      <p v-if="description" class="card-meta mt-2">{{ description }}</p>
      <div v-if="tags?.length" style="display:flex;gap:var(--space-2);flex-wrap:wrap;margin-top:var(--space-3);">
        <span v-for="tag in tags" :key="tag" :class="['tag', 'tag-hue-' + tagHue(tag)]">{{ tag }}</span>
      </div>
    </div>
  `,
})

// --- Tasks Page Header ---
const TasksPageHeader = defineComponent({
  name: 'TasksPageHeader',
  components: { Breadcrumbs },
  props: {
    code: { type: String, default: 'R1' },
    researchName: { type: String, default: 'Vue Component Architecture' },
    taskCount: { type: Number, default: 4 },
  },
  template: `
    <div class="page-header" style="margin-bottom:var(--space-3);padding-bottom:var(--space-2);">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: '/research/R1' },
        { label: 'Tasks' }
      ]" />
      <div style="display:flex;justify-content:space-between;align-items:center;gap:var(--space-4);flex-wrap:wrap;">
        <div style="display:flex;align-items:center;gap:var(--space-3);">
          <span v-if="code" style="font-size:var(--type-xs);font-weight:600;color:var(--color-primary);background:var(--color-primary-muted);padding:0.15rem 0.4rem;border-radius:4px;font-family:'JetBrains Mono',monospace;flex-shrink:0;line-height:1;">{{ code }}</span>
          <h1 class="page-title">Tasks</h1>
          <span style="font-size:var(--type-xs);color:var(--color-text-muted);background:var(--color-surface-hover);padding:0.15rem 0.5rem;border-radius:4px;font-variant-numeric:tabular-nums;">{{ taskCount }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:var(--space-3);">
          <button class="btn btn-sm btn-primary">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            New task
          </button>
        </div>
      </div>
    </div>
  `,
})

// --- Session Page Header ---
const SessionPageHeader = defineComponent({
  name: 'SessionPageHeader',
  components: { Breadcrumbs, StatusBadge },
  props: {
    title: { type: String, default: 'Deep Dive: Composition API' },
    status: { type: String, default: 'active' },
    focus: { type: String, default: 'Explore advanced patterns in Composition API usage' },
    researchName: { type: String, default: 'Vue Component Architecture' },
  },
  template: `
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: '/research/R1' },
        { label: title }
      ]" />
      <div style="display:flex;justify-content:space-between;align-items:center;gap:var(--space-4);flex-wrap:wrap;">
        <h1 class="page-title">{{ title }}</h1>
        <StatusBadge :status="status" />
      </div>
      <p v-if="focus" class="card-meta mt-2">Focus: {{ focus }}</p>
    </div>
  `,
})

// --- Meta & Stories ---

const meta: Meta = {
  title: 'Layout/PageHeaders',
  tags: ['autodocs'],
}
export default meta

type Story = StoryObj

export const ResearchHeader: Story = {
  render: (args) => ({
    components: { ResearchPageHeader },
    setup() { return { args } },
    template: '<ResearchPageHeader v-bind="args" />',
  }),
  args: {
    code: 'R1',
    name: 'Vue Component Architecture',
    goal: 'Investigate best practices for Vue 3 component design and composition patterns',
    status: 'active',
    taskCount: 4,
    showGoal: true,
  },
}

export const ResearchHeaderCompleted: Story = {
  render: (args) => ({
    components: { ResearchPageHeader },
    setup() { return { args } },
    template: '<ResearchPageHeader v-bind="args" />',
  }),
  args: {
    code: 'R2',
    name: 'State Management Patterns',
    goal: 'Compare Pinia, Vuex, and composable-based state management',
    status: 'completed',
    taskCount: 12,
    showGoal: true,
  },
}

export const ResearchHeaderLongTitle: Story = {
  render: (args) => ({
    components: { ResearchPageHeader },
    setup() { return { args } },
    template: '<ResearchPageHeader v-bind="args" />',
  }),
  args: {
    code: 'R5',
    name: 'Comprehensive Evaluation of Modern Frontend Frameworks and Build Tools in Enterprise Environments',
    goal: 'Evaluate and compare modern frontend frameworks including Vue 3, React 18, Angular 17, Svelte 5, and Solid.js.',
    status: 'active',
    taskCount: 23,
    showGoal: true,
  },
}

export const EntryHeader: Story = {
  render: (args) => ({
    components: { EntryPageHeader },
    setup() { return { args } },
    template: '<EntryPageHeader v-bind="args" />',
  }),
  args: {
    code: 'E1',
    title: 'Component Composition Patterns',
    description: 'Analysis of composable patterns in Vue 3',
    status: 'completed',
    tags: ['vue', 'composables'],
    researchName: 'Vue Component Architecture',
    sectionName: 'Fundamentals',
  },
}

export const EntryHeaderDraft: Story = {
  render: (args) => ({
    components: { EntryPageHeader },
    setup() { return { args } },
    template: '<EntryPageHeader v-bind="args" />',
  }),
  args: {
    code: 'E7',
    title: 'Reactive State Management',
    description: 'How ref, reactive, and computed work together',
    status: 'draft',
    tags: ['vue', 'reactivity', 'state'],
    researchName: 'Vue Component Architecture',
    sectionName: 'Advanced Patterns',
  },
}

export const EntryHeaderManyTags: Story = {
  render: (args) => ({
    components: { EntryPageHeader },
    setup() { return { args } },
    template: '<EntryPageHeader v-bind="args" />',
  }),
  args: {
    code: 'E12',
    title: 'Performance Optimization',
    description: 'Comprehensive guide to optimizing Vue 3 applications for production',
    status: 'active',
    tags: ['vue', 'performance', 'optimization', 'lazy-loading', 'tree-shaking', 'bundle-size', 'ssr'],
    researchName: 'Vue Component Architecture',
    sectionName: 'Advanced Patterns',
  },
}

export const TasksHeader: Story = {
  render: (args) => ({
    components: { TasksPageHeader },
    setup() { return { args } },
    template: '<TasksPageHeader v-bind="args" />',
  }),
  args: {
    code: 'R1',
    researchName: 'Vue Component Architecture',
    taskCount: 4,
  },
}

export const TasksHeaderEmpty: Story = {
  render: (args) => ({
    components: { TasksPageHeader },
    setup() { return { args } },
    template: '<TasksPageHeader v-bind="args" />',
  }),
  args: {
    code: 'R3',
    researchName: 'CSS Architecture Review',
    taskCount: 0,
  },
}

export const SessionHeader: Story = {
  render: (args) => ({
    components: { SessionPageHeader },
    setup() { return { args } },
    template: '<SessionPageHeader v-bind="args" />',
  }),
  args: {
    title: 'Deep Dive: Composition API',
    status: 'active',
    focus: 'Explore advanced patterns in Composition API usage',
    researchName: 'Vue Component Architecture',
  },
}

export const SessionHeaderCompleted: Story = {
  render: (args) => ({
    components: { SessionPageHeader },
    setup() { return { args } },
    template: '<SessionPageHeader v-bind="args" />',
  }),
  args: {
    title: 'Initial Exploration',
    status: 'completed',
    focus: '',
    researchName: 'State Management Patterns',
  },
}
