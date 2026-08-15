<template>
  <div v-if="pending">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <div v-else-if="question">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${researchSlug}` },
        { label: sessionTitle, to: `/research/${researchSlug}/session/${sessionId}` },
        { label: `Q${questionIndex + 1}` }
      ]" />
      <div class="question-header">
        <h1 class="page-title" v-html="renderRefs(question.text, researchSlug)"></h1>
        <div class="question-badges">
          <StatusBadge :status="question.status" />
          <StatusBadge v-if="question.priority" :status="question.priority" />
        </div>
      </div>
      <div class="question-meta">
        <span v-if="question.area" class="question-area">{{ question.area }}</span>
      </div>
    </div>

    <!-- Rationale -->
    <div v-if="question.rationale" class="card rationale-card">
      <h3 class="card-section-title">Rationale</h3>
      <div class="rationale-text markdown-content" v-html="linkRefs(parseMarkdownInline(normalizeContent(question.rationale)) as string, researchSlug)"></div>
    </div>

    <!-- Answer -->
    <div v-if="question.answer" class="card answer-card">
      <h3 class="card-section-title">Answer</h3>
      <div ref="answerEl" class="markdown-content" v-html="renderedAnswer"></div>
    </div>

    <EmptyState
      v-else-if="question.status === 'pending'"
      icon="&#x23F3;"
      title="Awaiting answer"
      description="This question hasn't been answered yet."
    />

    <!-- Cross-references from this answer -->
    <div v-if="hasRefs" class="crossrefs-block card no-print">
      <h3 class="crossrefs-title">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
        Cross-references
      </h3>
      <div class="crossrefs-list">
        <NuxtLink
          v-for="ref in outgoingRefs"
          :key="ref.target_entry_id || ref.target_ref"
          :to="refLink(ref)"
          class="crossref-item"
        >
          <span class="crossref-code">{{ ref.entry_code || ref.target_ref }}</span>
          <span v-if="ref.entry_title" class="crossref-entry-title">{{ ref.entry_title }}</span>
          <span v-if="!ref.resolved" class="crossref-unresolved">unresolved</span>
        </NuxtLink>
      </div>
    </div>

    <!-- Prev / Next question navigation -->
    <div v-if="allQuestions.length > 1" class="question-nav no-print">
      <NuxtLink v-if="prevQuestion" :to="`/research/${researchSlug}/session/${sessionId}/question/${prevQuestion.code || prevQuestion.id}`" class="btn btn-sm nav-btn">
        &larr; {{ truncate(prevQuestion.text, 50) }}
      </NuxtLink>
      <span v-else class="nav-placeholder"></span>
      <NuxtLink v-if="nextQuestion" :to="`/research/${researchSlug}/session/${sessionId}/question/${nextQuestion.code || nextQuestion.id}`" class="btn btn-sm nav-btn nav-next">
        {{ truncate(nextQuestion.text, 50) }} &rarr;
      </NuxtLink>
    </div>
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Question not found" />
</template>

<script setup lang="ts">
import { parseMarkdown, parseMarkdownInline } from '~/composables/useSafeMarkdown'
import { renderMermaidBlocks } from '~/composables/useMermaid'

const route = useRoute()
const id = route.params.id as string
const sessionId = route.params.sessionId as string
const questionId = route.params.questionId as string

// Research info
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Research')
const researchSlug = computed(() => researchData.value?.data?.research?.code || id)

// Session + questions
const { data, pending } = await useApi<{ data: any }>(`/api/researches/${id}/sessions/${sessionId}`)
const sessionTitle = computed(() => data.value?.data?.session?.title ?? 'Session')

// Flatten all questions from grouped structure
const allQuestions = computed(() => {
  const qs = data.value?.data?.questions ?? {}
  const flat: any[] = []
  for (const status of ['pending', 'in_progress', 'answered', 'deferred', 'skipped']) {
    if (qs[status]) flat.push(...qs[status])
  }
  return flat
})

const question = computed(() =>
  allQuestions.value.find((q: any) => q.id === questionId || q.code === questionId) ?? null
)
const questionIndex = computed(() =>
  allQuestions.value.findIndex((q: any) => q.id === questionId || q.code === questionId)
)

// Prev/next
const prevQuestion = computed(() => questionIndex.value > 0 ? allQuestions.value[questionIndex.value - 1] : null)
const nextQuestion = computed(() => questionIndex.value < allQuestions.value.length - 1 ? allQuestions.value[questionIndex.value + 1] : null)

// Rendered answer
const renderedAnswer = computed(() => {
  if (!question.value?.answer) return ''
  const html = parseMarkdown(normalizeContent(question.value.answer)) as string
  return linkRefs(html, researchSlug.value)
})

// Mermaid rendering
const answerEl = ref<HTMLElement | null>(null)
watch(renderedAnswer, () => {
  nextTick(() => {
    if (answerEl.value) renderMermaidBlocks(answerEl.value)
  })
})
onMounted(() => {
  if (answerEl.value) renderMermaidBlocks(answerEl.value)
})

// Cross-references from this question's answer
const { data: refsData } = useApi<{ outgoing: any[]; incoming: any[] }>(
  computed(() => question.value ? `/api/entries/${question.value.id}/crossrefs` : `/api/entries/__none__/crossrefs`)
)
const outgoingRefs = computed(() => refsData.value?.outgoing ?? [])
const hasRefs = computed(() => outgoingRefs.value.length > 0)

function refLink(ref: any): string {
  const rCode = ref.research_code || researchSlug.value
  const eCode = ref.entry_code || ref.target_ref
  return `/research/${rCode}/entry/${eCode}`
}

function truncate(text: string, max: number): string {
  return text.length > max ? text.slice(0, max) + '...' : text
}
</script>

<style scoped>
.question-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
}
.question-badges {
  display: flex;
  gap: var(--space-2);
  flex-shrink: 0;
}
.question-meta {
  display: flex;
  gap: var(--space-3);
  margin-top: var(--space-2);
}
.question-area {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface);
  padding: 0.15rem 0.5rem;
  border-radius: var(--radius-xs);
  border: 1px solid var(--color-border);
}
.card-section-title {
  font-size: var(--type-base);
  font-weight: var(--weight-semibold);
  margin-bottom: var(--space-3);
  letter-spacing: -0.01em;
}
.rationale-card {
  margin-bottom: var(--space-4);
}
.rationale-text {
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  line-height: 1.6;
}
.answer-card {
  padding: var(--space-8);
  border-radius: var(--radius-lg);
}

/* Cross-references (reuse entry page patterns) */
.crossrefs-block {
  margin-top: var(--space-6);
  padding: var(--space-6);
  border-radius: var(--radius-lg);
}
.crossrefs-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  margin: 0 0 var(--space-4) 0;
}
.crossrefs-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.crossref-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  text-decoration: none;
  color: var(--color-text);
  transition: background 0.15s;
}
.crossref-item:hover { background: var(--color-surface-hover); }
.crossref-code {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  flex-shrink: 0;
}
.crossref-entry-title {
  font-size: var(--type-sm);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.crossref-unresolved {
  font-size: var(--type-xs);
  color: var(--color-warning, var(--color-warning));
  font-style: italic;
  margin-left: auto;
}

/* Navigation */
.question-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--space-8);
  padding-top: var(--space-6);
  border-top: 1px solid var(--color-border);
  gap: var(--space-4);
}
.nav-btn {
  max-width: 45%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nav-next { margin-left: auto; text-align: right; }
.nav-placeholder { flex: 1; }

/* Print */
@media print {
  .question-badges { display: none; }
  .answer-card { padding: 0; border: none; }
  .rationale-card { padding: 0; border: none; }
  .question-area { background: none; border: none; padding: 0; }
}
</style>
