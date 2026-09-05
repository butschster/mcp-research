<template>
  <EmptyState
    v-if="excluded"
    title="Not part of this link"
    description="The person who shared this project didn't include roadmaps. Ask them if you need them."
  >
    <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to project</NuxtLink>
  </EmptyState>

  <div v-else class="roadmaps-page">
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: researchName, to: researchPath(slug) },
        { label: 'Roadmaps' },
      ]" />
      <div class="title-with-code">
        <span v-if="researchCode" class="short-code">{{ researchCode }}</span>
        <h1 class="page-title">Roadmaps</h1>
        <span v-if="roadmaps.length" class="task-counter">{{ roadmaps.length }}</span>
      </div>
    </div>

    <div v-if="roadmaps.length" class="roadmaps-grid">
      <NuxtLink
        v-for="rm in roadmaps"
        :key="rm.id"
        :to="roadmapPath(slug, rm.code || rm.id)"
        class="roadmap-card-link"
      >
        <RoadmapCard :roadmap="rm" />
      </NuxtLink>
    </div>

    <EmptyState
      v-else
      icon="&#x1F5FA;"
      title="No roadmaps yet"
      description="Nothing has been mapped out in this project yet. This page updates by itself when it is."
    />
  </div>
</template>

<script setup lang="ts">
const { shareFetch, research, researchId, researchCode, include, slug } = useShare()
const researchName = computed(() => research.value?.name || 'Project')

const roadmaps = ref<any[]>([])

async function load() {
  try {
    const res = await shareFetch<{ data: any[] }>(`/researches/${researchId.value}/roadmaps`)
    roadmaps.value = res.data ?? []
  } catch {
    roadmaps.value = []
  }
}

// Absent entry point, 404 from the API, and this screen for a URL that was
// typed by hand. The empty-roadmaps state below would have claimed the research
// has none, which is a different and false statement.
const excluded = computed(() => !include.value.roadmaps)
if (!excluded.value) void load()

useResearchRealtime(
  () => slug.value,
  (event) => { if (event.entity === 'roadmap') void load() },
  { onResync: () => void load(), researchId: () => researchId.value },
)
</script>

<style scoped>
.roadmaps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(340px, 100%), 1fr));
  gap: var(--space-4);
}
.roadmap-card-link { text-decoration: none; }
</style>
