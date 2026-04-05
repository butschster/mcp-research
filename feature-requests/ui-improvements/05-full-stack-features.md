# Track B — Full-Stack Features

**Scope:** Новые фичи, требующие изменений и в Go, и в Vue.  
**Requires:** Backend API specs из `04-backend-api-specs.md`

---

## FS-1: Activity Feed Widget

**Depends:** BS-2 (Activity Log endpoint)

**Позиция:** На странице `research/[id]/index.vue`, между Tasks widget и sidebar layout.

**Новый компонент:** `frontend/components/ActivityFeed.vue`

```vue
<template>
  <div class="card activity-feed">
    <div class="feed-header">
      <h3 class="feed-title">Activity</h3>
      <span class="feed-meta card-meta">{{ events.length }} recent events</span>
    </div>
    <div class="feed-list">
      <div
        v-for="event in events"
        :key="event.id"
        class="feed-item"
        :class="`feed-${event.entity}`"
      >
        <span class="feed-icon">{{ entityIcon(event.entity) }}</span>
        <div class="feed-content">
          <span class="feed-title-text">{{ event.title }}</span>
          <span class="feed-time card-meta">{{ relativeTime(event.timestamp) }}</span>
        </div>
        <span class="feed-action-badge" :class="`feed-action-${event.action}`">
          {{ event.action }}
        </span>
      </div>
    </div>
    <button v-if="hasMore" class="btn feed-more" @click="loadMore">
      Load more
    </button>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ researchId: string }>()

const { data, refresh } = useApi<{ data: any[] }>(
  `/api/researches/${props.researchId}/activity?limit=10`
)
const events = computed(() => data.value?.data ?? [])
const hasMore = computed(() => events.value.length >= 10)

const ENTITY_ICONS: Record<string, string> = {
  entry: '📄', question: '❓', task: '✓', session: '💬', section: '📂', research: '🔬'
}
function entityIcon(entity: string) { return ENTITY_ICONS[entity] ?? '•' }

// Real-time: рефрешим feed при любом SSE событии
useRealtimeUpdates(async (event) => {
  if (event.research_id === props.researchId) await refresh()
})
</script>

<style scoped>
.activity-feed { margin-bottom: 1.5rem; }
.feed-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; }
.feed-list { display: flex; flex-direction: column; gap: 0; }
.feed-item {
  display: flex; align-items: flex-start; gap: 0.625rem;
  padding: 0.5rem 0; border-bottom: 1px solid var(--color-border);
}
.feed-item:last-child { border-bottom: none; }
.feed-icon { font-size: 0.875rem; flex-shrink: 0; margin-top: 0.125rem; }
.feed-content { flex: 1; min-width: 0; }
.feed-title-text { display: block; font-size: 0.875rem; font-weight: 500; }
.feed-time { display: block; font-size: 0.75rem; margin-top: 0.125rem; }
.feed-action-badge {
  font-size: 0.6875rem; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.05em; padding: 0.125rem 0.375rem; border-radius: 4px;
  flex-shrink: 0;
}
.feed-action-created { background: rgba(52,211,153,0.12); color: var(--color-success); }
.feed-action-updated { background: rgba(96,165,250,0.12); color: var(--color-info); }
.feed-action-answered { background: rgba(52,211,153,0.12); color: var(--color-success); }
.feed-action-completed { background: rgba(52,211,153,0.12); color: var(--color-success); }
.feed-more { width: 100%; margin-top: 0.75rem; justify-content: center; font-size: 0.8125rem; }
</style>
```

---

## FS-2: Research Stats Dashboard Widget

**Depends:** BS-6 (Stats endpoint)

**Позиция:** В верхней части `research/[id]/index.vue`, после header.

```vue
<template>
  <div v-if="stats" class="stats-grid">
    <div class="stat-card">
      <span class="stat-number">{{ stats.entries_by_section.reduce((s, x) => s + x.count, 0) }}</span>
      <span class="stat-label">Total entries</span>
    </div>
    <div class="stat-card">
      <span class="stat-number">{{ stats.questions_by_status.answered || 0 }}</span>
      <span class="stat-label">Questions answered</span>
    </div>
    <div class="stat-card" :class="{ 'stat-attention': stats.questions_by_status.pending > 0 }">
      <span class="stat-number">{{ stats.questions_by_status.pending || 0 }}</span>
      <span class="stat-label">Pending questions</span>
    </div>
    <div class="stat-card">
      <span class="stat-number">{{ stats.completion_pct }}%</span>
      <span class="stat-label">Complete</span>
      <ProgressBar :value="stats.completion_pct" :total="100" class="stat-progress" />
    </div>
  </div>
</template>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}
.stat-card {
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius); padding: 0.875rem 1rem;
  display: flex; flex-direction: column; gap: 0.25rem;
}
.stat-card.stat-attention { border-color: rgba(251,191,36,0.4); }
.stat-number { font-size: 1.75rem; font-weight: 700; line-height: 1; }
.stat-label { font-size: 0.75rem; color: var(--color-text-muted); }
.stat-progress { margin-top: 0.5rem; }

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
```

---

## FS-3: Full-text Search Integration

**Depends:** BS-3 (Search endpoint)

Расширение `SearchModal.vue` (из NF-1) для поиска по содержимому entries:

```vue
<!-- Добавить в search results -->
<div v-if="entryResults.length" class="result-group">
  <div class="result-group-label">Entries</div>
  <NuxtLink
    v-for="(e, i) in entryResults"
    :key="e.id"
    :to="`/research/${e.research_id}/entry/${e.id}`"
    :class="['result-item', { 'result-active': cursor === i + researchResults.length }]"
    @click="open = false"
  >
    <span class="result-icon">📄</span>
    <div class="result-content">
      <span class="result-title">{{ e.title }}</span>
      <span class="result-meta">
        {{ e.research_name }} › {{ e.section_name }}
      </span>
      <span class="result-highlight" v-html="e.highlight"></span>
    </div>
  </NuxtLink>
</div>

<script>
// Добавить в SearchModal
const { data: searchData } = await useFetch(
  computed(() => query.value.length >= 2 ? `/api/search?q=${encodeURIComponent(query.value)}` : null)
)
const entryResults = computed(() => searchData.value?.data?.entries ?? [])
</script>

<style>
.result-highlight {
  display: block; font-size: 0.75rem; color: var(--color-text-muted);
  margin-top: 0.125rem; font-style: italic;
}
</style>
```

---

## FS-4: Research Page — "Attention Required" Widget

**Проблема:** Пользователь не видит, что нужно его участие (pending questions, blocked tasks).

**Зависит от:** BS-1 (расширенные метаданные)

```vue
<!-- Добавить перед active session widget, только если есть pending -->
<div v-if="attentionItems.length" class="card attention-widget">
  <div class="attention-header">
    <span class="attention-icon">⚠️</span>
    <h3>Requires Your Attention</h3>
  </div>
  <div v-for="item in attentionItems" :key="item.id" class="attention-item">
    <span :class="['attention-type', `attention-${item.type}`]">{{ item.typeLabel }}</span>
    <span class="attention-text">{{ item.text }}</span>
    <NuxtLink v-if="item.link" :to="item.link" class="attention-action">View →</NuxtLink>
  </div>
</div>

<script>
const attentionItems = computed(() => {
  const items = []
  if (activeSession.value && sessionProgress.value?.pending > 0) {
    items.push({
      type: 'question',
      typeLabel: 'Question',
      text: `${sessionProgress.value.pending} questions need your answer`,
      link: `/research/${id}/session/${activeSession.value.id}`
    })
  }
  const blockedTasks = tasks.value.filter(t => t.status === 'blocked')
  blockedTasks.forEach(t => {
    items.push({ type: 'task', typeLabel: 'Blocked', text: t.title, link: null })
  })
  return items
})
</script>
```

```css
.attention-widget { border-color: rgba(251,191,36,0.3); margin-bottom: 1.5rem; }
.attention-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.75rem; }
.attention-header h3 { font-size: 1rem; }
.attention-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem 0; border-bottom: 1px solid var(--color-border); }
.attention-item:last-child { border-bottom: none; }
.attention-type { font-size: 0.6875rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; padding: 0.125rem 0.5rem; border-radius: 4px; flex-shrink: 0; }
.attention-question { background: rgba(96,165,250,0.12); color: var(--color-info); }
.attention-task { background: rgba(248,113,113,0.12); color: var(--color-error); }
.attention-text { flex: 1; font-size: 0.875rem; }
.attention-action { font-size: 0.8125rem; color: var(--color-primary); flex-shrink: 0; }
```
