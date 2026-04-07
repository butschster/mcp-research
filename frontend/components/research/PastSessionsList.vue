<template>
  <div v-if="sessions.length" class="past-sessions">
    <button class="btn-ghost past-sessions-toggle" @click="showPast = !showPast">
      Completed sessions
      <span class="past-sessions-count">{{ sessions.length }}</span>
      <span class="past-sessions-chevron" :class="{ open: showPast }">&rsaquo;</span>
    </button>
    <div v-show="showPast" class="past-sessions-list">
      <NuxtLink
        v-for="sess in sessions"
        :key="sess.id"
        :to="`/research/${researchSlug}/session/${sess.code || sess.id}`"
        class="past-session-item"
      >
        <span v-if="sess.code" class="short-code">{{ sess.code }}</span>
        <span class="past-session-title">{{ sess.title }}</span>
        <StatusBadge :status="sess.status" />
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  sessions: any[]
  researchSlug: string
}>()

const showPast = ref(false)
</script>

<style scoped>
.past-sessions { margin-bottom: var(--space-4); display: flex; flex-direction: column; align-items: flex-end; }
.past-sessions-toggle {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  padding: var(--space-2) 0;
}
.past-sessions-toggle:hover { color: var(--color-text); }
.past-sessions-count {
  font-size: var(--type-xs);
  background: var(--color-surface-hover);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
}
.past-sessions-chevron {
  margin-left: auto;
  font-size: var(--type-lg);
  color: var(--color-text-muted);
  transition: transform var(--transition-base);
  display: inline-block;
}
.past-sessions-chevron.open { transform: rotate(90deg); }
.past-sessions-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  margin-top: var(--space-2);
  width: 100%;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-2);
}
.past-session-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  text-decoration: none;
  color: inherit;
  font-size: var(--type-sm);
  transition: background var(--transition-fast);
}
.past-session-item:hover { background: var(--color-surface-hover); }
.past-session-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.short-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
</style>
