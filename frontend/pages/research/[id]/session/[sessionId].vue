<template>
  <div v-if="pending">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-progress"></div>
    <div class="skeleton-card skeleton-questions"></div>
  </div>

  <div v-else-if="session">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${id}` },
        { label: session.title }
      ]" />
      <div class="session-header">
        <h1 class="page-title">{{ session.title }}</h1>
        <StatusBadge :status="session.status" />
      </div>
      <p v-if="session.focus" class="card-meta mt-2">Focus: {{ session.focus }}</p>
    </div>

    <!-- Progress card -->
    <div class="card progress-card">
      <h3 class="card-section-title">Progress</h3>
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
    <div v-if="session.notes" class="card notes-card">
      <h3 class="card-section-title">Session Notes</h3>
      <p class="notes-text">{{ session.notes }}</p>
    </div>

    <!-- Questions -->
    <div class="card">
      <h3 class="card-section-title mb-4">Questions</h3>
      <QuestionList :questions="questions" />
    </div>
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Session not found" />
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string
const sessionId = route.params.sessionId as string

const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Research')

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
.session-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }
.card-section-title { font-size: var(--type-base); font-weight: 600; margin-bottom: var(--space-3); }
.progress-card { margin-bottom: var(--space-6); }
.notes-card { margin-bottom: var(--space-6); }
.notes-text { white-space: pre-wrap; color: var(--color-text-muted); font-size: var(--type-sm); }
.progress-stats {
  display: flex;
  gap: var(--space-6);
  font-size: var(--type-sm);
  margin-top: var(--space-2);
  flex-wrap: wrap;
}
.stat-answered { color: var(--color-success); }
.stat-pending  { color: var(--color-warning); }
.stat-skipped  { color: var(--color-error); }
.skeleton-card { background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); opacity: 0.5; }
.skeleton-header { height: 60px; margin-bottom: var(--space-4); }
.skeleton-progress { height: 100px; margin-bottom: var(--space-4); }
.skeleton-questions { height: 300px; }
</style>
