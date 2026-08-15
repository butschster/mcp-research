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
        :to="roadmapPath(researchSlug, rm.code || rm.id)"
        class="roadmap-card-link"
      >
        <RoadmapCard :roadmap="rm" />
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

async function reloadRoadmaps() {
  roadmapsData.value = await authFetch<{ data: any[] }>(`${rtBase}/api/researches/${id}/roadmaps`)
}

useResearchRealtime(
  () => id,
  (event) => { if (event.entity === 'roadmap') reloadRoadmaps() },
  { researchId: () => research.value?.id, onResync: reloadRoadmaps },
)
</script>

<style scoped>
.roadmaps-page {
  /* width, not auto: an auto horizontal margin suppresses flex stretch, so this
     column would silently shrink to its content if it ever sat in a flex
     container. */
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
}
.roadmaps-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}
.roadmaps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(340px, 100%), 1fr));
  gap: var(--space-4);
  margin-top: var(--space-4);
}
.roadmaps-grid-skeleton {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(340px, 100%), 1fr));
  gap: var(--space-4);
  margin-top: var(--space-4);
}
.roadmap-card-link {
  text-decoration: none;
  color: inherit;
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
  border-radius: var(--radius-xs);
  font-size: var(--type-xs);
}
</style>
