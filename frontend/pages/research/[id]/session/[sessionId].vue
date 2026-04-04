<template>
  <div v-if="pending" class="empty-state">Loading...</div>
  <div v-else-if="!session" class="empty-state">Session not found.</div>
  <div v-else>
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${id}` },
        { label: session.title },
      ]" />
      <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 0.25rem;">
        <h1 class="page-title">{{ session.title }}</h1>
        <StatusBadge :status="session.status" />
      </div>
      <p v-if="session.focus" class="card-meta" style="margin-top: 0.25rem;">Focus: {{ session.focus }}</p>
    </div>

    <!-- Progress -->
    <div class="card" style="margin-bottom: 1.5rem;">
      <h3 style="margin-bottom: 0.75rem;">Progress</h3>
      <div class="progress-bar" style="margin-bottom: 0.5rem;">
        <div class="progress-bar-fill" :style="{ width: progressPercent + '%' }" />
      </div>
      <div style="display: flex; gap: 1.5rem; font-size: 0.875rem;">
        <span>Total: {{ progress.total }}</span>
        <span style="color: var(--color-success);">Answered: {{ progress.answered }}</span>
        <span style="color: var(--color-warning);">Pending: {{ progress.pending }}</span>
        <span v-if="progress.deferred" style="color: var(--color-text-muted);">Deferred: {{ progress.deferred }}</span>
        <span v-if="progress.skipped" style="color: var(--color-error);">Skipped: {{ progress.skipped }}</span>
      </div>
    </div>

    <!-- Notes -->
    <div v-if="session.notes" class="card" style="margin-bottom: 1.5rem;">
      <h3 style="margin-bottom: 0.5rem;">Session Notes</h3>
      <p style="white-space: pre-wrap; color: var(--color-text-muted);">{{ session.notes }}</p>
    </div>

    <!-- Questions -->
    <div class="card">
      <h3 style="margin-bottom: 1rem;">Questions</h3>
      <QuestionList :questions="questions" />
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string
const sessionId = route.params.sessionId as string

const { data: researchData } = await useApi<{ data: { research: any } }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Project')

const { data, pending } = await useApi<{ data: { session: any; questions: Record<string, any[]>; progress: any } }>(`/api/sessions/${sessionId}`)

const session = computed(() => data.value?.data?.session ?? data.value?.data?.Session)
const questions = computed(() => data.value?.data?.questions ?? data.value?.data?.Questions ?? {})
const progress = computed(() => data.value?.data?.progress ?? data.value?.data?.Progress ?? { total: 0, answered: 0, pending: 0, deferred: 0, skipped: 0 })

const progressPercent = computed(() => {
  if (!progress.value.total) return 0
  return Math.round((progress.value.answered / progress.value.total) * 100)
})
</script>
