<template>
  <div>
    <div class="page-header" style="display: flex; justify-content: space-between; align-items: center;">
      <h1 class="page-title">Research Projects</h1>
    </div>

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

    <div v-if="pending" class="empty-state">Loading...</div>
    <div v-else-if="!filteredResearches?.length" class="empty-state">
      <p>No research projects found.</p>
      <p class="card-meta" style="margin-top: 0.5rem;">
        Use the <code>research/initialize</code> prompt in Claude to create one.
      </p>
    </div>
    <div v-else class="grid grid-2">
      <ResearchCard
        v-for="r in filteredResearches"
        :key="r.id"
        :research="r"
        @tag-click="tagFilter = $event"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
const statusFilter = ref('')
const tagFilter = ref('')

const url = computed(() => {
  const base = '/api/researches'
  return statusFilter.value ? `${base}?status=${statusFilter.value}` : base
})

const { data, pending } = await useApi<{ data: any[]; count: number }>(url.value)

// Re-fetch when status changes
watch(statusFilter, async () => {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || ''
  const res = await $fetch<{ data: any[] }>(`${baseURL}${url.value}`)
  data.value = res as any
})

const researches = computed(() => data.value?.data ?? [])

const filteredResearches = computed(() => {
  if (!tagFilter.value) return researches.value
  return researches.value.filter((r: any) =>
    r.tags?.includes(tagFilter.value)
  )
})
</script>

<style scoped>
.active-tag-filter {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: rgba(56, 189, 248, 0.1);
  border: 1px solid rgba(56, 189, 248, 0.3);
  border-radius: var(--radius);
  font-size: 0.8125rem;
  color: var(--color-primary);
}
.tag-clear {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: 1rem;
  cursor: pointer;
  padding: 0 0.125rem;
  line-height: 1;
}
.tag-clear:hover {
  color: var(--color-text);
}
</style>
