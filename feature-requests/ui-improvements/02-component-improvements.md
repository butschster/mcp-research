# Track A — Component Improvements

**Scope:** Рефакторинг и улучшение существующих компонентов. Только frontend.  
**Effort:** 2-8 часов на компонент.

---

## CI-1: ActivityIndicator — Новый компонент

**Проблема:** Нет визуального индикатора, что Claude сейчас работает с данными.

**Новый файл:** `frontend/components/ActivityIndicator.vue`

```vue
<template>
  <div v-if="active" class="activity-indicator">
    <span class="activity-dot"></span>
    <span class="activity-label">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  active: boolean
  label?: string
}>()
</script>

<style scoped>
.activity-indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: var(--color-text-muted);
  padding: 0.25rem 0.625rem;
  border-radius: 999px;
  background: rgba(56,189,248,0.08);
  border: 1px solid rgba(56,189,248,0.2);
}
.activity-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-primary);
  animation: pulse 1.5s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%       { opacity: 0.4; transform: scale(0.8); }
}
.activity-label { color: var(--color-primary); font-weight: 500; }
</style>
```

**Интеграция в `frontend/app.vue`:**
```vue
<div class="nav-right">
  <ActivityIndicator
    :active="hasRecentUpdate"
    label="Claude is working"
  />
  <ConnectionStatus />
</div>

<script setup>
// В useRealtimeUpdates callback — выставлять hasRecentUpdate на 5 сек
const hasRecentUpdate = ref(false)
let activityTimer: ReturnType<typeof setTimeout>

useRealtimeUpdates(() => {
  hasRecentUpdate.value = true
  clearTimeout(activityTimer)
  activityTimer = setTimeout(() => {
    hasRecentUpdate.value = false
  }, 5000)
})
</script>
```

---

## CI-2: ResearchCard — Расширение информации

**Проблема:** Карточка показывает только name + goal + tags + time. Нет прогресса.

**Файл:** `frontend/components/ResearchCard.vue`

```vue
<template>
  <NuxtLink :to="`/research/${research.id}`" class="card research-card">
    <div class="card-header">
      <h3 class="card-title">{{ research.name }}</h3>
      <StatusBadge :status="research.status" />
    </div>

    <p v-if="research.goal" class="card-meta goal-text">
      {{ research.goal }}
    </p>

    <!-- НОВОЕ: Stats row -->
    <div v-if="hasStats" class="card-stats">
      <span v-if="research.sections_count" class="stat-item">
        <span class="stat-icon">📂</span>
        {{ research.sections_count }} sections
      </span>
      <span v-if="research.entries_count" class="stat-item">
        <span class="stat-icon">📄</span>
        {{ research.entries_count }} entries
      </span>
      <span v-if="research.pending_questions" class="stat-item stat-attention">
        <span class="stat-icon">❓</span>
        {{ research.pending_questions }} pending
      </span>
    </div>

    <div class="card-footer">
      <div v-if="research.tags?.length" class="tags-row">
        <span
          v-for="tag in research.tags"
          :key="tag"
          :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tag)}`]"
          @click.prevent.stop="emit('tagClick', tag)"
        >{{ tag }}</span>
      </div>
      <span v-if="research.updated_at" class="card-meta timestamp">
        {{ relativeTime(research.updated_at) }}
      </span>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
const props = defineProps<{
  research: {
    id: string
    name: string
    goal: string
    status: string
    tags: string[]
    updated_at?: string
    sections_count?: number
    entries_count?: number
    pending_questions?: number
  }
}>()

const hasStats = computed(() =>
  props.research.sections_count || props.research.entries_count || props.research.pending_questions
)

function tagHue(tag: string): number {
  return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
}
// ... relativeTime без изменений
</script>

<style scoped>
.card-header   { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 0.375rem; }
.goal-text     { margin-top: 0; }
.card-stats    { display: flex; gap: 1rem; margin-top: 0.625rem; flex-wrap: wrap; }
.stat-item     { display: flex; align-items: center; gap: 0.25rem; font-size: 0.75rem; color: var(--color-text-muted); }
.stat-attention { color: var(--color-warning); }
.stat-icon     { font-size: 0.75rem; }
/* остальные стили без изменений */
</style>
```

**Note:** Поле `sections_count`, `entries_count`, `pending_questions` нужно добавить в API ответ `/api/researches` — это Track B задача (BS-1).

---

## CI-3: ConnectionStatus — Расширение до AI Status

**Проблема:** `ConnectionStatus` показывает только WebSocket (● connected/disconnected). Не показывает AI-активность.

**Файл:** `frontend/components/ConnectionStatus.vue`

```vue
<template>
  <div class="connection-status">
    <!-- WS индикатор — без изменений -->
    <span :class="['ws-dot', connected ? 'ws-connected' : 'ws-disconnected']"></span>
    <span class="ws-label card-meta">{{ connected ? 'Live' : 'Offline' }}</span>
    
    <!-- НОВОЕ: last event info -->
    <span v-if="lastEvent" class="last-event card-meta">
      · {{ lastEvent }}
    </span>
  </div>
</template>

<script setup lang="ts">
// ... существующий код без изменений

const lastEvent = ref('')
useRealtimeUpdates((event) => {
  const entityMap: Record<string, string> = {
    research: 'Project updated',
    entry: 'Entry added',
    question: 'Question answered',
    task: 'Task updated',
    session: 'Session updated',
  }
  lastEvent.value = entityMap[event.entity] || 'Updated'
  // Очищаем через 8 сек
  setTimeout(() => { lastEvent.value = '' }, 8000)
})
</script>
```

---

## CI-4: QuestionList — Счётчики в заголовке группы (closed state)

**Проблема:** Когда группа вопросов свёрнута — не видно сколько вопросов внутри.

**Файл:** `frontend/components/QuestionList.vue`

```vue
<!-- Существующий group-header — добавить count badge -->
<button class="group-header" @click="toggle(group.status)">
  <StatusBadge :status="group.status" />
  <span class="group-count">{{ group.questions.length }}</span>
  <!-- НОВОЕ: mini progress если есть answered дочерних -->
  <span v-if="group.status === 'pending'" class="group-priority-hint">
    <span v-if="highPriorityCount(group.questions)" class="high-priority-dot">
      {{ highPriorityCount(group.questions) }} high priority
    </span>
  </span>
  <span class="group-chevron" :class="{ open: isOpen(group.status) }">›</span>
</button>
```

```css
.group-count {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  border-radius: 999px;
  padding: 0.125rem 0.5rem;
  font-weight: 600;
}
.high-priority-dot {
  font-size: 0.75rem;
  color: var(--color-error);
}
```

---

## CI-5: ProgressBar — Label и tooltip

**Проблема:** `ProgressBar` — просто полоска без label. Нет percentage.

**Файл:** `frontend/components/ProgressBar.vue`

```vue
<template>
  <div class="progress-wrap">
    <div class="progress-bar" :title="`${pct}%`">
      <div
        class="progress-bar-fill"
        :style="{ width: `${pct}%` }"
        :class="progressClass"
      ></div>
    </div>
    <span v-if="showLabel" class="progress-label">{{ pct }}%</span>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  value: number
  total: number
  showLabel?: boolean
}>()

const pct = computed(() =>
  props.total > 0 ? Math.round((props.value / props.total) * 100) : 0
)

const progressClass = computed(() => {
  if (pct.value === 100) return 'fill-complete'
  if (pct.value >= 70) return 'fill-good'
  if (pct.value >= 30) return 'fill-mid'
  return 'fill-low'
})
</script>

<style scoped>
.progress-wrap { display: flex; align-items: center; gap: 0.5rem; }
.progress-bar  { flex: 1; height: 6px; background: var(--color-surface-hover); border-radius: 3px; overflow: hidden; }
.progress-bar-fill { height: 100%; border-radius: 3px; transition: width 0.3s; }
.fill-complete { background: var(--color-success); }
.fill-good     { background: var(--color-info); }
.fill-mid      { background: var(--color-warning); }
.fill-low      { background: var(--color-error); }
.progress-label { font-size: 0.75rem; color: var(--color-text-muted); min-width: 2.5rem; text-align: right; }
</style>
```

---

## CI-6: StatusBadge — Иконки к статусам

**Проблема:** Badges — только текст. Иконки ускоряют распознавание.

**Файл:** `frontend/components/StatusBadge.vue`

```vue
<template>
  <span :class="['badge', `badge-${props.status}`]">
    <span v-if="icon" class="badge-icon">{{ icon }}</span>
    {{ label }}
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{ status: string }>()

const STATUS_CONFIG: Record<string, { label: string; icon: string }> = {
  active:      { label: 'Active',      icon: '●' },
  completed:   { label: 'Completed',   icon: '✓' },
  archived:    { label: 'Archived',    icon: '○' },
  draft:       { label: 'Draft',       icon: '◌' },
  pending:     { label: 'Pending',     icon: '◷' },
  answered:    { label: 'Answered',    icon: '✓' },
  in_progress: { label: 'In Progress', icon: '▶' },
  deferred:    { label: 'Deferred',    icon: '→' },
  skipped:     { label: 'Skipped',     icon: '↷' },
  blocked:     { label: 'Blocked',     icon: '■' },
  failed:      { label: 'Failed',      icon: '×' },
  high:        { label: 'High',        icon: '↑' },
  medium:      { label: 'Medium',      icon: '•' },
  low:         { label: 'Low',         icon: '↓' },
}

const config = computed(() => STATUS_CONFIG[props.status] ?? { label: props.status, icon: '' })
const label = computed(() => config.value.label)
const icon = computed(() => config.value.icon)
</script>

<style scoped>
.badge-icon { margin-right: 0.25rem; }
</style>
```
