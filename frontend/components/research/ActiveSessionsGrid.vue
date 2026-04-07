<template>
  <div v-if="sessions.length" class="active-sessions-grid">
    <NuxtLink
      v-for="sess in sessions"
      :key="sess.id"
      :to="`/research/${researchSlug}/session/${sess.code || sess.id}`"
      class="card session-widget session-active"
    >
      <div class="session-widget-header">
        <div class="flex items-center gap-2">
          <span v-if="sess.code" class="short-code">{{ sess.code }}</span>
          <StatusBadge status="active" />
        </div>
      </div>
      <h3 class="session-title">{{ sess.title }}</h3>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  sessions: any[]
  researchSlug: string
}>()
</script>

<style scoped>
.active-sessions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}
.session-widget {
  display: block;
  text-decoration: none;
  color: inherit;
  margin-bottom: 0;
  border-color: var(--color-border);
  position: relative;
  overflow: hidden;
}
.session-active {
  border-color: rgba(108, 197, 224, 0.15);
}
.session-widget:hover { text-decoration: none; }
.session-widget::after {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 1px;
  background: linear-gradient(90deg, transparent, rgba(108, 197, 224, 0.3), transparent);
}
.session-widget-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: var(--space-1);
}
.session-title { font-size: var(--type-sm); font-weight: 600; letter-spacing: -0.01em; }
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
