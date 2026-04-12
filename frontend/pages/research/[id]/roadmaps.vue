<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card skeleton-header"></div>
    <div class="roadmaps-grid-skeleton">
      <div v-for="i in 3" :key="i" class="skeleton-card" style="height: 160px;"></div>
    </div>
  </div>

  <div v-else-if="research" class="roadmaps-page">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: research.name, to: `/research/${researchSlug}` },
        { label: 'Roadmaps' }
      ]" />
      <div class="roadmaps-header">
        <div class="title-with-code">
          <span v-if="research.code" class="short-code">{{ research.code }}</span>
          <h1 class="page-title">Roadmaps</h1>
          <span v-if="roadmaps.length" class="task-counter">{{ roadmaps.length }}</span>
        </div>
      </div>
    </div>

    <!-- Roadmaps Grid -->
    <div v-if="roadmaps.length" class="roadmaps-grid">
      <NuxtLink
        v-for="rm in roadmaps"
        :key="rm.id"
        :to="`/research/${researchSlug}/roadmap/${rm.code || rm.id}`"
        class="roadmap-card-link"
      >
        <div class="card roadmap-card">
          <div class="rm-card-header">
            <div class="rm-card-title-row">
              <span v-if="rm.code" class="rm-card-code">{{ rm.code }}</span>
              <h3 class="rm-card-title">{{ rm.title }}</h3>
            </div>
            <StatusBadge :status="rm.status" />
          </div>
          <p v-if="rm.description" class="rm-card-desc">{{ rm.description }}</p>
          <div v-if="rm.statuses?.length" class="rm-card-statuses">
            <span v-for="s in rm.statuses" :key="s" class="rm-card-status-chip">{{ s }}</span>
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- Empty state -->
    <div v-else class="empty-state">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" style="opacity:0.3;">
        <circle cx="12" cy="12" r="3"/><circle cx="4" cy="6" r="2"/><circle cx="20" cy="6" r="2"/><circle cx="4" cy="18" r="2"/><circle cx="20" cy="18" r="2"/>
        <path d="M9.5 10.5 5.5 7.5"/><path d="M14.5 10.5l4-3"/><path d="M9.5 13.5 5.5 16.5"/><path d="M14.5 13.5l4 3"/>
      </svg>
      <p class="card-meta mt-4">No roadmaps yet</p>
      <p class="card-meta">Use the <code>roadmap_create</code> MCP tool to build a visual graph</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string

const { data: researchData, pending } = await useApi<{ data: any }>(`/api/researches/${id}`)
const research = computed(() => researchData.value?.data?.research)
const researchSlug = computed(() => research.value?.code || id)

const { data: roadmapsData } = await useApi<{ data: any[] }>(`/api/researches/${id}/roadmaps`)
const roadmaps = computed(() => roadmapsData.value?.data ?? [])

// Real-time updates
const config = useRuntimeConfig()
const rtBase = config.public.apiBase || ''
const { authFetch } = useAuth()

useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return
  if (event.entity === 'roadmap') {
    const res = await authFetch<{ data: any[] }>(`${rtBase}/api/researches/${id}/roadmaps`)
    roadmapsData.value = res
  }
})
</script>

<style scoped>
.roadmaps-page {
  max-width: 900px;
  margin: 0 auto;
}
.roadmaps-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}
.title-with-code {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}
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
.task-counter {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
}
.roadmaps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--space-4);
  margin-top: var(--space-4);
}
.roadmaps-grid-skeleton {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--space-4);
  margin-top: var(--space-4);
}
.roadmap-card-link {
  text-decoration: none;
  color: inherit;
}
.roadmap-card {
  display: flex;
  flex-direction: column;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.roadmap-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}
.rm-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}
.rm-card-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.rm-card-code {
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
.rm-card-title {
  font-size: var(--type-sm);
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
  line-height: 1.3;
}
.rm-card-desc {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.5;
  margin-top: var(--space-2);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.rm-card-statuses {
  display: flex;
  gap: var(--space-1);
  flex-wrap: wrap;
  margin-top: var(--space-3);
}
.rm-card-status-chip {
  font-size: 0.625rem;
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
  line-height: 1;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-6) 0;
  text-align: center;
}
.empty-state code {
  background: var(--color-surface-hover);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  font-size: var(--type-xs);
}
</style>
