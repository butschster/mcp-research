<template>
  <EmptyState
    v-if="excluded"
    title="Not part of this link"
    description="The person who shared this research didn't include tasks. Ask them if you need them."
  >
    <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to the research</NuxtLink>
  </EmptyState>

  <div v-else class="tasks-page">
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: researchName, to: researchPath(slug) },
        { label: 'Tasks' },
      ]" />
      <div class="title-with-code">
        <span v-if="researchCode" class="short-code">{{ researchCode }}</span>
        <h1 class="page-title">Tasks</h1>
        <span class="task-counter">{{ tasks.length }}</span>
      </div>
    </div>

    <!-- The board and the detail modal both gate every write on canWrite, which
         the share lock holds false. Nothing here is draggable and nothing
         saves. -->
    <TasksKanbanBoard
      :columns="columns"
      :tasks="tasks"
      :research-slug="slug"
      @task-click="detailTask = $event"
    />

    <TasksTaskDetailModal
      :task="detailTask"
      :research-slug="slug"
      @close="detailTask = null"
    />
  </div>
</template>

<script setup lang="ts">
const { shareFetch, research, researchId, researchCode, include, slug } = useShare()
const researchName = computed(() => research.value?.name || 'Research')

const tasks = ref<any[]>([])
const detailTask = ref<any | null>(null)

// The same four the owner's board uses, so a visitor and the person who shared
// it are looking at the same board.
const columns = [
  { status: 'pending', label: 'Todo' },
  { status: 'in_progress', label: 'In Progress' },
  { status: 'completed', label: 'Done' },
  { status: 'failed', label: 'Rejected' },
]

// The entry point for this page is absent when the link excludes tasks, and the
// API 404s it. This is the third answer, for a URL that was typed by hand — and
// it has to say "not part of this link" rather than showing an empty board,
// which would assert that the research has no tasks.
const excluded = computed(() => !include.value.tasks)

async function load() {
  if (excluded.value) {
    tasks.value = []
    return
  }
  try {
    const res = await shareFetch<{ data: any[] }>(`/researches/${researchId.value}/tasks`)
    tasks.value = res.data ?? []
  } catch {
    tasks.value = []
  }
}

void load()

useResearchRealtime(
  () => slug.value,
  (event) => { if (event.entity === 'task') void load() },
  { onResync: () => void load(), researchId: () => researchId.value },
)
</script>
