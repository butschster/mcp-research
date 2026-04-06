<template>
  <NuxtLink :to="`/research/${research.code || research.id}`" class="card research-card">
    <div class="card-header">
      <div class="card-title-row">
        <span v-if="research.code" class="short-code">{{ research.code }}</span>
        <h3 class="card-title">{{ research.name }}</h3>
      </div>
      <div class="card-header-actions">
        <button
          class="btn-icon"
          :title="research.status === 'archived' ? 'Restore from archive' : 'Archive'"
          @click.prevent.stop="toggleArchive"
        >
          <svg v-if="research.status === 'archived'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
        </button>
        <StatusBadge :status="research.status" />
      </div>
    </div>

    <p v-if="research.goal" class="card-meta goal-text">{{ research.goal }}</p>

    <div class="card-footer">
      <div v-if="research.tags?.length" class="tags-row">
        <span
          v-for="tag in research.tags"
          :key="tag"
          :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tag)}`]"
          @click.prevent.stop="emit('tagClick', tag)"
        >{{ tag }}</span>
      </div>
      <span v-if="research.updated_at" class="card-meta timestamp">
        {{ relativeTime(research.updated_at) }}
      </span>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
const props = defineProps<{
  research: {
    id: string
    code?: string
    name: string
    goal: string
    status: string
    tags: string[]
    updated_at?: string
  }
}>()

const emit = defineEmits<{ tagClick: [tag: string]; statusChanged: [] }>()

const { authFetch } = useAuth()
const config = useRuntimeConfig()
const base = config.public.apiBase || ''

async function toggleArchive() {
  const newStatus = props.research.status === 'archived' ? 'active' : 'archived'
  await authFetch(`${base}/api/researches/${props.research.id}`, {
    method: 'PUT',
    body: { status: newStatus },
  })
  emit('statusChanged')
}

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
</script>

<style scoped>
.research-card {
  display: flex;
  flex-direction: column;
  text-decoration: none;
  color: inherit;
}
.card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: var(--space-3); }
.card-header-actions { display: flex; align-items: center; gap: var(--space-2); }
.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: none;
  color: var(--color-text-muted);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  opacity: 0;
}
.research-card:hover .btn-icon { opacity: 1; }
.btn-icon:hover { background: var(--color-surface-hover); color: var(--color-text); }
.card-title-row { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.short-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.goal-text {
  margin-top: var(--space-2);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
}
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
  padding-top: var(--space-3);
  gap: var(--space-2);
}
.tags-row { display: flex; gap: var(--space-2); flex-wrap: wrap; }
.timestamp { white-space: nowrap; flex-shrink: 0; }
.tag-clickable { cursor: pointer; transition: all var(--transition-fast); }
.tag-clickable:hover { background: var(--color-primary-muted); color: var(--color-primary); }
</style>
