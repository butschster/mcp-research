import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, h } from 'vue'
import { mockResearch, mockResearchCompleted, mockResearchArchived } from '../__mocks__/research'

// Stub ResearchCard to avoid useAuth/useRuntimeConfig dependency
const ResearchCardStub = defineComponent({
  name: 'ResearchCardStub',
  props: {
    research: { type: Object, required: true },
  },
  emits: ['tagClick', 'statusChanged'],
  setup(props, { emit }) {
    function tagHue(tag: string): number {
      return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
    }
    function relativeTime(iso: string): string {
      const diff = Date.now() - new Date(iso).getTime()
      const mins = Math.floor(diff / 60_000)
      if (mins < 1) return 'just now'
      if (mins < 60) return `${mins}m ago`
      const hrs = Math.floor(mins / 60)
      if (hrs < 24) return `${hrs}h ago`
      const days = Math.floor(hrs / 24)
      return `${days}d ago`
    }
    return { tagHue, relativeTime, emit }
  },
  template: `
    <a :href="'/research/' + (research.code || research.id)" class="card research-card" style="display:flex;flex-direction:column;text-decoration:none;color:inherit;">
      <div style="display:flex;justify-content:space-between;align-items:flex-start;gap:var(--space-3);">
        <div style="display:flex;align-items:center;gap:var(--space-2);min-width:0;">
          <span v-if="research.code" class="short-code" style="font-size:var(--type-xs);font-weight:600;color:var(--color-primary);background:var(--color-primary-muted);padding:0.15rem 0.4rem;border-radius:4px;font-family:'JetBrains Mono',monospace;flex-shrink:0;line-height:1;">{{ research.code }}</span>
          <h3 class="card-title">{{ research.name }}</h3>
        </div>
        <span :class="['badge', 'badge-' + research.status]">{{ research.status }}</span>
      </div>
      <p v-if="research.goal" class="card-meta" style="margin-top:var(--space-2);display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;line-height:1.5;">{{ research.goal }}</p>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-top:auto;padding-top:var(--space-3);gap:var(--space-2);">
        <div v-if="research.tags?.length" style="display:flex;gap:var(--space-2);flex-wrap:wrap;">
          <span
            v-for="tag in research.tags"
            :key="tag"
            :class="['tag', 'tag-hue-' + tagHue(tag)]"
            style="cursor:pointer;"
            @click.prevent.stop="emit('tagClick', tag)"
          >{{ tag }}</span>
        </div>
        <span v-if="research.updated_at" class="card-meta" style="white-space:nowrap;flex-shrink:0;">
          {{ relativeTime(research.updated_at) }}
        </span>
      </div>
    </a>
  `,
})

const meta: Meta<typeof ResearchCardStub> = {
  title: 'Cards/ResearchCard',
  component: ResearchCardStub,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof ResearchCardStub>

export const Active: Story = {
  args: { research: mockResearch },
}

export const Completed: Story = {
  args: { research: mockResearchCompleted },
}

export const Archived: Story = {
  args: { research: mockResearchArchived },
}

export const WithoutTags: Story = {
  args: {
    research: { ...mockResearch, tags: [] },
  },
}

export const ManyTags: Story = {
  args: {
    research: {
      ...mockResearch,
      tags: ['vue', 'react', 'angular', 'svelte', 'typescript', 'javascript', 'css', 'architecture'],
    },
  },
}

export const LongGoal: Story = {
  args: {
    research: {
      ...mockResearch,
      name: 'Comprehensive Frontend Framework Evaluation',
      goal: 'Evaluate and compare modern frontend frameworks including Vue 3, React 18, Angular 17, Svelte 5, and Solid.js across performance benchmarks, developer experience, ecosystem maturity, and enterprise readiness criteria.',
    },
  },
}

export const AllStatuses: Story = {
  render: () => ({
    components: { ResearchCardStub },
    template: `
      <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 1rem;">
        <ResearchCardStub
          v-for="r in researches"
          :key="r.id"
          :research="r"
        />
      </div>
    `,
    setup() {
      const researches = [mockResearch, mockResearchCompleted, mockResearchArchived]
      return { researches }
    },
  }),
}
