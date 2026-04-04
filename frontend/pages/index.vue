<template>
  <div>
    <div class="page-header flex-between">
      <h1 class="page-title">Research Projects</h1>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <select v-model="statusFilter">
        <option value="">All statuses</option>
        <option value="active">Active</option>
        <option value="completed">Completed</option>
        <option value="archived">Archived</option>
      </select>
      <span v-if="tagFilter" class="active-tag-filter">
        Tag: <strong>{{ tagFilter }}</strong>
        <button class="tag-clear" @click="tagFilter = ''">×</button>
      </span>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="skeleton-list">
      <div v-for="i in 4" :key="i" class="skeleton-card"></div>
    </div>

    <!-- List -->
    <div v-else-if="filtered.length" class="grid grid-2">
      <ResearchCard
        v-for="r in filtered"
        :key="r.id"
        :research="r"
        @tag-click="tagFilter = $event"
      />
    </div>

    <!-- Empty -->
    <EmptyState
      v-else
      icon="🔬"
      title="No research projects yet"
      description="Use the research/initialize prompt in Claude to create one."
    />
  </div>
</template>

<script setup lang="ts">
const statusFilter = ref('')
const tagFilter = ref('')

const apiUrl = computed(() =>
  statusFilter.value ? `/api/researches?status=${statusFilter.value}` : '/api/researches'
)

const { data, pending, refresh } = useApi<{ data: any[] }>(apiUrl.value)

watch(statusFilter, async () => {
  const config = useRuntimeConfig()
  const base = config.public.apiBase || ''
  const res = await $fetch<{ data: any[] }>(`${base}${apiUrl.value}`)
  data.value = res
})

const researches = computed(() => data.value?.data ?? [])

const filtered = computed(() =>
  tagFilter.value
    ? researches.value.filter((r: any) => r.tags?.includes(tagFilter.value))
    : researches.value
)

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.entity === 'research') {
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const res = await $fetch<{ data: any[] }>(`${base}${apiUrl.value}`)
    data.value = res
  }
})
</script>

<style scoped>
.flex-between { display: flex; justify-content: space-between; align-items: center; }
.skeleton-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(400px, 1fr)); gap: 1rem; }
.skeleton-card {
  height: 110px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  animation: shimmer 1.5s infinite;
}
@keyframes shimmer {
  0%, 100% { opacity: 0.6; }
  50%       { opacity: 1; }
}
.active-tag-filter {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.625rem;
  background: rgba(56,189,248,0.1);
  border: 1px solid rgba(56,189,248,0.3);
  border-radius: var(--radius);
  font-size: 0.8125rem;
  color: var(--color-primary);
}
.tag-clear {
  background: none;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  font-size: 0.875rem;
  padding: 0;
  line-height: 1;
}
</style>
