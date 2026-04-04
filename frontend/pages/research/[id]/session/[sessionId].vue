<template>
  <div v-if="pending">
    <div class="skeleton-card" style="height:60px;margin-bottom:1rem;"></div>
    <div class="skeleton-card" style="height:100px;margin-bottom:1rem;"></div>
    <div class="skeleton-card" style="height:300px;"></div>
  </div>

  <div v-else-if="session">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${id}` },
        { label: session.title }
      ]" />
      <div class="flex-between" style="margin-top:0.25rem;">
        <h1 class="page-title">{{ session.title }}</h1>
        <StatusBadge :status="session.status" />
      </div>
      <p v-if="session.focus" class="card-meta" style="margin-top:0.25rem;">
        Focus: {{ session.focus }}
      </p>
    </div>

    <!-- Progress card -->
    <div class="card" style="margin-bottom:1.5rem;">
      <h3 style="margin-bottom:0.75rem;">Progress</h3>
      <ProgressBar :value="progress.answered" :total="progress.total" />
      <div class="progress-stats">
        <span>Total: {{ progress.total }}</span>
        <span class="stat-answered">Answered: {{ progress.answered }}</span>
        <span class="stat-pending">Pending: {{ progress.pending }}</span>
        <span v-if="progress.deferred" class="card-meta">Deferred: {{ progress.deferred }}</span>
        <span v-if="progress.skipped" class="stat-skipped">Skipped: {{ progress.skipped }}</span>
      </div>
    </div>

    <!-- Notes card -->
    <div v-if="session.notes" class="card" style="margin-bottom:1.5rem;">
      <h3 style="margin-bottom:0.5rem;">Session Notes</h3>
      <p style="white-space:pre-wrap;color:var(--color-text-muted);">{{ session.notes }}</p>
    </div>

    <!-- Questions -->
    <div class="card">
      <h3 style="margin-bottom:1rem;">Questions</h3>
      <QuestionList :questions="questions" />
    </div>
  </div>

  <EmptyState v-else icon="🔍" title="Session not found" />
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string
const sessionId = route.params.sessionId as string

// Research name for breadcrumb
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Research')

// Session data
const { data, pending } = await useApi<{ data: any }>(`/api/sessions/${sessionId}`)

const session  = computed(() => data.value?.data?.session ?? data.value?.data?.Session)
const questions = computed(() => data.value?.data?.questions ?? data.value?.data?.Questions ?? {})
const progress  = computed(() => ({
  total:    data.value?.data?.progress?.total    ?? 0,
  answered: data.value?.data?.progress?.answered ?? 0,
  pending:  data.value?.data?.progress?.pending  ?? 0,
  deferred: data.value?.data?.progress?.deferred ?? 0,
  skipped:  data.value?.data?.progress?.skipped  ?? 0,
}))

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return
  if (['question', 'session'].includes(event.entity)) {
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const res = await $fetch<any>(`${base}/api/sessions/${sessionId}`)
    data.value = res
  }
})
</script>

<style scoped>
.flex-between  { display: flex; justify-content: space-between; align-items: center; }
.progress-stats {
  display: flex;
  gap: 1.5rem;
  font-size: 0.875rem;
  margin-top: 0.5rem;
  flex-wrap: wrap;
}
.stat-answered { color: var(--color-success); }
.stat-pending  { color: var(--color-warning); }
.stat-skipped  { color: var(--color-error); }
.skeleton-card { background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); animation: shimmer 1.5s infinite; }
@keyframes shimmer { 0%,100%{opacity:.6} 50%{opacity:1} }
</style>
