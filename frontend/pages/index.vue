<template>
  <div>
    <div class="page-header">
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
        <button class="tag-clear" @click="tagFilter = ''">&times;</button>
      </span>
    </div>

    <!-- Onboarding -->
    <GettingStartedBanner :has-researches="researches.length > 0" />

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
        @status-changed="refreshList"
      />
    </div>

    <!-- Empty -->
    <EmptyState
      v-else
      icon="&#x1F52C;"
      title="No research projects yet"
      description="Type this into Claude to start your first research session:"
      command="Use the research/initialize prompt to create a new research project"
    />
  </div>
</template>

<script setup lang="ts">
const { authFetch } = useAuth()
const config = useRuntimeConfig()
const base = config.public.apiBase || ''

const statusFilter = ref('active')
const tagFilter = ref('')

const apiUrl = computed(() =>
  statusFilter.value ? `/api/researches?status=${statusFilter.value}` : '/api/researches'
)

const { data, pending, refresh } = useApi<{ data: any[] }>(apiUrl.value)

async function refreshList() {
  const res = await authFetch<{ data: any[] }>(`${base}${apiUrl.value}`)
  data.value = res
}

watch(statusFilter, refreshList)

const researches = computed(() => data.value?.data ?? [])

const filtered = computed(() =>
  tagFilter.value
    ? researches.value.filter((r: any) => r.tags?.includes(tagFilter.value))
    : researches.value
)

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.entity === 'research') {
    const res = await authFetch<{ data: any[] }>(`${base}${apiUrl.value}`)
    data.value = res
  }
})
</script>

<style scoped>
.skeleton-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: var(--space-6);
}
.skeleton-card {
  height: 120px;
  border-radius: var(--radius);
}
.active-tag-filter {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) var(--space-3);
  background: var(--color-primary-muted);
  border: 1px solid rgba(108, 197, 224, 0.15);
  border-radius: var(--radius-sm);
  font-size: var(--type-xs);
  color: var(--color-primary);
}
.tag-clear {
  background: none;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  font-size: var(--type-sm);
  padding: 0;
  line-height: 1;
  opacity: 0.7;
  transition: opacity var(--transition-fast);
}
.tag-clear:hover { opacity: 1; }
</style>
