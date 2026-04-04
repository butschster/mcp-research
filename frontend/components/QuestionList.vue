<template>
  <div>
    <div v-for="group in groupedQuestions" :key="group.status" style="margin-bottom: 1.5rem;">
      <h4 style="margin-bottom: 0.75rem; display: flex; align-items: center; gap: 0.5rem;">
        <StatusBadge :status="group.status" />
        <span class="card-meta">({{ group.questions.length }})</span>
      </h4>
      <div v-for="q in group.questions" :key="q.id" :class="['question-item', { 'question-child': q.parent_id }]">
        <div style="display: flex; justify-content: space-between; align-items: start;">
          <div class="question-text">{{ q.text }}</div>
          <StatusBadge v-if="q.priority" :status="q.priority" />
        </div>
        <div v-if="q.area" class="card-meta">Area: {{ q.area }}</div>
        <div v-if="q.answer" class="question-answer">{{ q.answer }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Question {
  id: string
  text: string
  area: string
  priority: string
  answer: string
  parent_id: string
  status: string
}

const props = defineProps<{
  questions: Record<string, Question[]>
}>()

const statusOrder = ['pending', 'in_progress', 'answered', 'deferred', 'skipped']

const groupedQuestions = computed(() => {
  return statusOrder
    .filter(status => props.questions[status]?.length)
    .map(status => ({
      status,
      questions: props.questions[status],
    }))
})
</script>
