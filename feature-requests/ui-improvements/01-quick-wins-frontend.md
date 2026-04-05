# Track A — Quick Wins: Frontend-Only Improvements

**Scope:** Только изменения в `frontend/`. Нет изменений в Go.  
**Effort:** 1-4 часа на каждый пункт.  
**Priority:** Делать в порядке нумерации.

---

## QW-1: Active Sidebar Item — Left Border Accent

**Проблема:** Активный sidebar item использует только `background: var(--color-surface)` — почти неотличим от hover state.

**Файл:** `frontend/assets/css/main.css`

```css
/* БЫЛО */
.sidebar-item:hover,
.sidebar-item.active {
  background: var(--color-surface);
  color: var(--color-text);
}

/* СТАЛО */
.sidebar-item:hover {
  background: var(--color-surface);
  color: var(--color-text);
}
.sidebar-item.active {
  background: var(--color-surface);
  color: var(--color-text);
  border-left: 3px solid var(--color-primary);
  padding-left: calc(0.875rem - 3px); /* компенсируем border */
  font-weight: 500;
}
```

**Rationale:** Left border accent — стандартный паттерн для active nav item (Linear, GitHub, VS Code). Немедленно устраняет confusion "где я нахожусь".

---

## QW-2: Card Hover Enhancement

**Проблема:** `card:hover { border-color: var(--color-primary); }` — только border, нет depth.

**Файл:** `frontend/assets/css/main.css`

```css
/* СТАЛО */
.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.25rem;
  transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
}
.card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 1px rgba(56, 189, 248, 0.15), 0 4px 16px rgba(0, 0, 0, 0.3);
  background: color-mix(in srgb, var(--color-surface) 97%, var(--color-primary) 3%);
}
```

**Rationale:** Box-shadow создаёт ощущение глубины и интерактивности. Subtle tint от primary цвета при hover — премиальный dev-tool паттерн.

---

## QW-3: Read-Only Badge в Navigation

**Проблема:** Пользователи не понимают, что UI — read-only. Нет нигде явного указания.

**Файл:** `frontend/app.vue`

```vue
<!-- Добавить в nav-right, перед ConnectionStatus -->
<span class="readonly-badge">Read-only view</span>
```

```css
/* В frontend/assets/css/main.css */
.readonly-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  border: 1px solid var(--color-border);
  border-radius: 999px;
  padding: 0.125rem 0.5rem;
}
```

**Rationale:** Устраняет confusion у новых пользователей ("почему я не могу редактировать?"). Паттерн из Notion, Figma viewer mode.

---

## QW-4: Empty State — Actionable Content

**Проблема:** `EmptyState` показывает emoji + text, но не даёт никакого action.

**Файл:** `frontend/components/EmptyState.vue`

```vue
<template>
  <div class="empty-state">
    <div v-if="icon" class="empty-icon">{{ icon }}</div>
    <h3 class="empty-title">{{ title }}</h3>
    <p v-if="description" class="empty-desc">{{ description }}</p>
    <!-- НОВОЕ: copyable action -->
    <div v-if="command" class="empty-command">
      <code class="command-text">{{ command }}</code>
      <button class="copy-btn" @click="copy" :class="{ copied }">
        {{ copied ? '✓ Copied' : 'Copy' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon?: string
  title: string
  description?: string
  command?: string  // <- НОВЫЙ PROP
}>()

const copied = ref(false)
async function copy() {
  if (!props.command) return
  await navigator.clipboard.writeText(props.command)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>
```

**Использование в `pages/index.vue`:**
```vue
<EmptyState
  icon="🔬"
  title="No research projects yet"
  description="Type this into Claude to start your first research session:"
  command="Use the research/initialize prompt to create a new research project"
/>
```

**Добавить CSS:**
```css
.empty-command {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1rem;
  padding: 0.625rem 0.875rem;
  background: var(--color-surface-hover);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  font-size: 0.8125rem;
  max-width: 500px;
  text-align: left;
}
.command-text {
  flex: 1;
  color: var(--color-primary);
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.8125rem;
  word-break: break-all;
}
.copy-btn {
  flex-shrink: 0;
  padding: 0.25rem 0.625rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  color: var(--color-text-muted);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
}
.copy-btn:hover { color: var(--color-text); border-color: var(--color-primary); }
.copy-btn.copied { color: var(--color-success); border-color: var(--color-success); }
```

---

## QW-5: Progress Indicator на Sidebar-секциях

**Проблема:** В sidebar секций виден только entry count. Нет индикатора completion.

**Файл:** `frontend/pages/research/[id]/index.vue`

```vue
<!-- В блоке sidebar, заменить sidebar-item -->
<div
  v-for="section in sections"
  :key="section.id"
  :class="['sidebar-item', { active: activeSection === section.id }]"
  @click="activeSection = section.id"
>
  <div class="sidebar-item-content">
    <span class="sidebar-item-name">{{ section.display_name || section.name }}</span>
    <StatusBadge :status="section.status" />
  </div>
  <div class="sidebar-item-meta">
    <span class="card-meta">{{ section.entries_count }} entries</span>
  </div>
  <!-- Мини прогресс-бар если есть данные -->
  <div v-if="section.entries_count > 0" class="sidebar-progress">
    <div
      class="sidebar-progress-fill"
      :style="{ width: sectionProgressWidth(section) }"
    ></div>
  </div>
</div>
```

```css
.sidebar-item-content { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
.sidebar-item-name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.875rem; }
.sidebar-item-meta { font-size: 0.75rem; margin-top: 0.125rem; }
.sidebar-progress { height: 2px; background: var(--color-border); border-radius: 1px; margin-top: 0.375rem; overflow: hidden; }
.sidebar-progress-fill { height: 100%; background: var(--color-success); border-radius: 1px; transition: width 0.3s; }
```

---

## QW-6: Tag Color-Coding

**Проблема:** Все теги одного серого цвета. Нет семантики.

**Файл:** `frontend/assets/css/main.css`

```css
/* Базовый тег — без изменений */
.tag { ... }

/* Семантические варианты через CSS nth-of-type или data-attribute */
/* Быстрое решение: auto-colour по первой букве хеша тега */
```

**Файл:** `frontend/components/ResearchCard.vue` — заменить span тега:

```vue
<span
  v-for="tag in research.tags"
  :key="tag"
  :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tag)}`]"
  @click.prevent.stop="emit('tagClick', tag)"
>{{ tag }}</span>

<script>
function tagHue(tag: string): number {
  // Детерминированный цвет по строке тега
  const hash = [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0)
  return hash % 6 // 6 цветовых классов
}
</script>
```

```css
/* 6 семантических оттенков для тегов */
.tag-hue-0 { background: rgba(56,189,248,0.12); color: #38bdf8; }   /* blue */
.tag-hue-1 { background: rgba(52,211,153,0.12); color: #34d399; }   /* green */
.tag-hue-2 { background: rgba(251,191,36,0.12); color: #fbbf24; }   /* yellow */
.tag-hue-3 { background: rgba(248,113,113,0.12); color: #f87171; }  /* red */
.tag-hue-4 { background: rgba(167,139,250,0.12); color: #a78bfa; }  /* purple */
.tag-hue-5 { background: rgba(251,146,60,0.12); color: #fb923c; }   /* orange */
```

**Rationale:** Детерминированная раскраска — один тег всегда одного цвета. Лёгкое распознавание без ручной настройки.

---

## QW-7: Entry Breadcrumb — Секция в пути

**Проблема:** На странице entry breadcrumb показывает только Research → Entry. Нет секции.

**Файл:** `frontend/pages/research/[id]/entry/[entryId].vue`

Нужно передать section name в breadcrumb:
```vue
<Breadcrumbs :crumbs="[
  { label: 'Research', to: '/' },
  { label: research.name, to: `/research/${id}` },
  { label: entry.section_display_name || entry.section_name, to: `/research/${id}?section=${entry.section_id}` },
  { label: entry.title }
]" />
```

API уже возвращает section данные — нужно только включить их в breadcrumb.

---

## QW-8: Mobile Sidebar Fix

**Проблема:** На мобиле sidebar становится длинным списком перед контентом. `position: sticky` теряется.

**Файл:** `frontend/assets/css/main.css`

```css
@media (max-width: 768px) {
  .layout-sidebar {
    grid-template-columns: 1fr;
  }
  
  /* НОВОЕ: на мобиле sidebar — горизонтальный скролл */
  .sidebar {
    position: static;
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    overflow-x: auto;
    padding-bottom: 0.5rem;
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none; /* Firefox */
  }
  .sidebar::-webkit-scrollbar { display: none; }
  
  .sidebar-label { display: none; }
  
  .sidebar-item {
    flex-shrink: 0;
    white-space: nowrap;
    padding: 0.375rem 0.75rem;
    border-radius: 999px;
    border: 1px solid var(--color-border);
    font-size: 0.8125rem;
  }
  .sidebar-item.active {
    background: rgba(56,189,248,0.12);
    border-color: var(--color-primary);
    color: var(--color-primary);
    border-left: 1px solid var(--color-primary); /* override desktop active */
    padding-left: 0.75rem;
  }
  
  /* Прогресс-бар в sidebar item на мобиле — скрываем */
  .sidebar-progress { display: none; }
}
```

**Rationale:** Горизонтальный pill-scroll sidebar — паттерн из GitHub mobile, Linear mobile. Не занимает вертикальное место перед контентом.
