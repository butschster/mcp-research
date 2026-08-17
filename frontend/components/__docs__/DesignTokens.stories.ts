import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, h } from 'vue'

const DesignTokens = defineComponent({
  name: 'DesignTokens',
  setup() {
    return () => h('div', 'Design system tokens showcase')
  },
})

const meta: Meta<typeof DesignTokens> = {
  title: 'Design System/Tokens',
  component: DesignTokens,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof DesignTokens>

export const Colors: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Color Palette</h2>
        <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: var(--space-4);">
          <div v-for="color in colors" :key="color.name" style="display: flex; flex-direction: column; gap: var(--space-2);">
            <div :style="{ background: color.value, width: '100%', height: '64px', borderRadius: 'var(--radius)', border: '1px solid var(--color-border)' }"></div>
            <div>
              <div style="font-size: var(--type-sm); font-weight: 600;">{{ color.label }}</div>
              <code style="font-size: var(--type-xs); color: var(--color-text-muted);">{{ color.name }}</code>
            </div>
          </div>
        </div>
      </div>
    `,
    setup() {
      const colors = [
        { name: '--color-bg', label: 'Background', value: '#0c1220' },
        { name: '--color-surface', label: 'Surface', value: '#151d2e' },
        { name: '--color-surface-hover', label: 'Surface Hover', value: '#1e2940' },
        { name: '--color-primary', label: 'Primary', value: '#6cc5e0' },
        { name: '--color-text', label: 'Text', value: '#e2e8f0' },
        { name: '--color-text-muted', label: 'Text Muted', value: '#7f8ea3' },
        { name: '--color-success', label: 'Success', value: '#34d399' },
        { name: '--color-warning', label: 'Warning', value: '#f0b849' },
        { name: '--color-error', label: 'Error', value: '#ef6b6b' },
        { name: '--color-info', label: 'Info', value: '#6b9df0' },
        { name: '--color-primary-muted', label: 'Primary Muted', value: 'rgba(108, 197, 224, 0.10)' },
        { name: '--color-border', label: 'Border', value: 'rgba(148, 163, 184, 0.12)' },
        { name: '--color-border-strong', label: 'Border Strong', value: 'rgba(148, 163, 184, 0.22)' },
      ]
      return { colors }
    },
  }),
}

export const Typography: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Type Scale</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-5);">
          <div v-for="t in typeScale" :key="t.name" style="display: flex; align-items: baseline; gap: var(--space-6);">
            <code style="font-size: var(--type-xs); color: var(--color-text-muted); min-width: 120px;">{{ t.name }}</code>
            <span :style="{ fontSize: t.value, fontWeight: 600, letterSpacing: '-0.01em' }">The quick brown fox</span>
            <span style="font-size: var(--type-xs); color: var(--color-text-muted);">{{ t.value }}</span>
          </div>
        </div>
        <h3 style="font-size: var(--type-lg); font-weight: 600; margin: var(--space-8) 0 var(--space-4); letter-spacing: -0.01em;">Font Weights</h3>
        <div style="display: flex; flex-direction: column; gap: var(--space-3);">
          <div v-for="w in weights" :key="w.value" style="display: flex; align-items: baseline; gap: var(--space-6);">
            <code style="font-size: var(--type-xs); color: var(--color-text-muted); min-width: 120px;">{{ w.value }}</code>
            <span :style="{ fontWeight: w.value, fontSize: 'var(--type-lg)' }">{{ w.label }}</span>
          </div>
        </div>
      </div>
    `,
    setup() {
      const typeScale = [
        { name: '--type-4xl', value: '3.5rem' },
        { name: '--type-3xl', value: '2.75rem' },
        { name: '--type-2xl', value: '2rem' },
        { name: '--type-xl', value: '1.5rem' },
        { name: '--type-lg', value: '1.25rem' },
        { name: '--type-base', value: '1.0625rem' },
        { name: '--type-sm', value: '0.9375rem' },
        { name: '--type-xs', value: '0.875rem' },
      ]
      const weights = [
        { value: 400, label: 'Regular body text' },
        { value: 500, label: 'Medium weight for labels' },
        { value: 600, label: 'Semibold for subheadings' },
        { value: 700, label: 'Bold for headings' },
      ]
      return { typeScale, weights }
    },
  }),
}

export const Spacing: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Spacing Scale (4px base)</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-3);">
          <div v-for="s in spacings" :key="s.name" style="display: flex; align-items: center; gap: var(--space-4);">
            <code style="font-size: var(--type-xs); color: var(--color-text-muted); min-width: 100px;">{{ s.name }}</code>
            <div :style="{ width: s.value, height: '16px', background: 'var(--color-primary)', borderRadius: '2px', opacity: 0.7 }"></div>
            <span style="font-size: var(--type-xs); color: var(--color-text-muted);">{{ s.value }}</span>
          </div>
        </div>
      </div>
    `,
    setup() {
      const spacings = [
        { name: '--space-1', value: '0.25rem' },
        { name: '--space-2', value: '0.5rem' },
        { name: '--space-3', value: '0.75rem' },
        { name: '--space-4', value: '1rem' },
        { name: '--space-5', value: '1.25rem' },
        { name: '--space-6', value: '1.5rem' },
        { name: '--space-8', value: '2rem' },
        { name: '--space-10', value: '2.5rem' },
        { name: '--space-12', value: '3rem' },
        { name: '--space-16', value: '4rem' },
      ]
      return { spacings }
    },
  }),
}

export const Badges: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Status Badges</h2>
        <div style="display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: center;">
          <span v-for="s in statuses" :key="s" :class="['badge', 'badge-' + s]">{{ s }}</span>
        </div>
      </div>
    `,
    setup() {
      const statuses = [
        'active', 'completed', 'archived', 'draft', 'pending',
        'answered', 'in_progress', 'deferred', 'skipped', 'blocked',
        'failed', 'high', 'medium', 'low',
      ]
      return { statuses }
    },
  }),
}

export const Buttons: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Buttons</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-6);">
          <div>
            <h3 style="font-size: var(--type-sm); color: var(--color-text-muted); margin-bottom: var(--space-3); font-weight: 500;">Standard</h3>
            <div style="display: flex; gap: var(--space-3); align-items: center; flex-wrap: wrap;">
              <button class="btn">Default Button</button>
              <button class="btn btn-sm">Small Button</button>
              <button class="btn btn-ghost">Ghost Button</button>
            </div>
          </div>
          <div>
            <h3 style="font-size: var(--type-sm); color: var(--color-text-muted); margin-bottom: var(--space-3); font-weight: 500;">Disabled</h3>
            <div style="display: flex; gap: var(--space-3); align-items: center;">
              <button class="btn" disabled>Disabled</button>
              <button class="btn btn-sm" disabled>Disabled Small</button>
            </div>
          </div>
        </div>
      </div>
    `,
  }),
}

export const Cards: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Card Styles</h2>
        <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: var(--space-4);">
          <div class="card">
            <h3 class="card-title">Default Card</h3>
            <p class="card-meta" style="margin-top: var(--space-2);">Basic card with surface background and subtle border.</p>
          </div>
          <div class="card" style="border-color: var(--color-primary); border-left-width: 3px;">
            <h3 class="card-title">Accent Card</h3>
            <p class="card-meta" style="margin-top: var(--space-2);">Card with primary color left border accent.</p>
          </div>
          <div class="card" style="background: var(--color-surface-hover);">
            <h3 class="card-title">Hover Surface Card</h3>
            <p class="card-meta" style="margin-top: var(--space-2);">Card using the hover surface color as base.</p>
          </div>
        </div>
      </div>
    `,
  }),
}

export const Tags: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Tag Hues</h2>
        <div style="display: flex; flex-wrap: wrap; gap: var(--space-2);">
          <span v-for="tag in tags" :key="tag" :class="['tag', 'tag-hue-' + tagHue(tag)]">{{ tag }}</span>
        </div>
        <h3 style="font-size: var(--type-sm); color: var(--color-text-muted); margin: var(--space-6) 0 var(--space-3); font-weight: 500;">Hue distribution (0-5)</h3>
        <div style="display: flex; flex-wrap: wrap; gap: var(--space-2);">
          <span v-for="i in 6" :key="i" :class="['tag', 'tag-hue-' + (i - 1)]">hue-{{ i - 1 }}</span>
        </div>
      </div>
    `,
    setup() {
      function tagHue(tag: string): number {
        return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
      }
      const tags = ['vue', 'react', 'typescript', 'css', 'architecture', 'performance', 'testing', 'composables', 'pinia', 'state', 'design-system', 'frontend']
      return { tags, tagHue }
    },
  }),
}

export const Radii: Story = {
  render: () => ({
    template: `
      <div>
        <h2 style="font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-6); letter-spacing: -0.02em;">Border Radii</h2>
        <div style="display: flex; gap: var(--space-6); align-items: center; flex-wrap: wrap;">
          <div v-for="r in radii" :key="r.name" style="text-align: center;">
            <div :style="{ width: '80px', height: '80px', background: 'var(--color-surface)', border: '1px solid var(--color-border-strong)', borderRadius: r.value }"></div>
            <code style="font-size: var(--type-xs); color: var(--color-text-muted); display: block; margin-top: var(--space-2);">{{ r.name }}</code>
            <span style="font-size: var(--type-xs); color: var(--color-text-muted);">{{ r.value }}</span>
          </div>
        </div>
      </div>
    `,
    setup() {
      const radii = [
        { name: '--radius-sm', value: '6px' },
        { name: '--radius', value: '10px' },
        { name: '--radius-lg', value: '16px' },
      ]
      return { radii }
    },
  }),
}

/**
 * Rule lists — the three frames a list of rows can wear.
 *
 * `.card--list`, `.list-head` and `.list-empty` have no component of their own,
 * so this is the only place they can be looked at. The principle: a rule
 * between two rows is a divider and belongs to the rows; a rule at the top or
 * bottom is an edge and belongs to whatever frames the list.
 */
export const RuleLists: Story = {
  render: () => ({
    template: `
      <div style="display:flex;flex-direction:column;gap:var(--space-8);max-width:720px">

        <div>
          <p style="font-size:var(--type-xs);color:var(--color-text-muted);margin-bottom:var(--space-3)">
            A list card <strong>with a heading</strong> — the head owns the rule, the rows own the dividers,
            and the card's own border closes the bottom.
          </p>
          <div class="card card--list">
            <div class="list-head">
              <h3 class="card-section-title">Chosen for this research</h3>
              <p style="font-size:var(--type-xs);color:var(--color-text-muted)">
                Methodology somebody picked.
              </p>
            </div>
            <div class="data-rows">
              <div class="data-row">Systematic review</div>
              <div class="data-row">Five whys</div>
              <div class="data-row data-row--dead">Retired: bug triage</div>
            </div>
          </div>
        </div>

        <div>
          <p style="font-size:var(--type-xs);color:var(--color-text-muted);margin-bottom:var(--space-3)">
            A list card that is <strong>only the list</strong> — no head, no rule at either end,
            and the first and last rows take the card's corner.
          </p>
          <div class="card card--list">
            <div class="data-rows">
              <div class="data-row">Prefer the client's own words for a section title.</div>
              <div class="data-row">The 2019 dataset is the only one with region codes.</div>
            </div>
          </div>
        </div>

        <div>
          <p style="font-size:var(--type-xs);color:var(--color-text-muted);margin-bottom:var(--space-3)">
            <strong>Bounded</strong> — a list standing on a page with nothing framing it. Here the outer
            rules are the only marks saying where it starts and stops, so the first and last row carry them.
          </p>
          <div class="data-rows data-rows--bounded">
            <div class="data-row">Anna Kuznetsova</div>
            <div class="data-row">Dmitri Orlov</div>
          </div>
        </div>

        <div>
          <p style="font-size:var(--type-xs);color:var(--color-text-muted);margin-bottom:var(--space-3)">
            <strong>Empty</strong> — the sentence stands where the rows would be, and nothing draws around it.
          </p>
          <div class="card card--list">
            <div class="list-head">
              <h3 class="card-section-title">Available to attach</h3>
            </div>
            <p class="list-empty">Nothing left to attach — everything available is already on.</p>
          </div>
        </div>

      </div>`,
  }),
}
