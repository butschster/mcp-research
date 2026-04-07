<template>
  <nav class="sidebar">
    <!-- All entries -->
    <div
      :class="['sidebar-item', { active: activeSection === '__all__' }]"
      @click="$emit('update:activeSection', '__all__')"
    >
      <div class="sidebar-item-content">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
        <span class="sidebar-item-name">All entries</span>
        <span class="sidebar-count">{{ totalEntryCount }}</span>
      </div>
    </div>

    <div class="sidebar-divider"></div>

    <!-- Per-section -->
    <div
      v-for="section in sections"
      :key="section.id"
      :class="['sidebar-item', { active: activeSection === section.id }]"
      @click="$emit('update:activeSection', section.id)"
    >
      <div class="sidebar-item-content">
        <span class="sidebar-item-name">{{ section.display_name || section.name }}</span>
        <span class="sidebar-count">{{ section.entries_count }}</span>
      </div>
      <div v-if="section.entries_count > 0" class="sidebar-progress">
        <div class="sidebar-progress-fill" :style="{ width: sectionProgressWidth(section) }"></div>
      </div>
    </div>

    <div class="sidebar-divider"></div>

    <!-- External links -->
    <div
      :class="['sidebar-item', { active: activeSection === '__links__' }]"
      @click="$emit('update:activeSection', '__links__')"
    >
      <div class="sidebar-item-content">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
        <span class="sidebar-item-name">External links</span>
        <span v-if="linksTotal" class="sidebar-count">{{ linksTotal }}</span>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
defineProps<{
  sections: any[]
  activeSection: string
  totalEntryCount: number
  linksTotal: number
}>()

defineEmits<{
  'update:activeSection': [value: string]
}>()

function sectionProgressWidth(section: any): string {
  if (section.status === 'completed') return '100%'
  if (section.status === 'active') return '50%'
  if (section.status === 'draft') return '10%'
  return '0%'
}
</script>

<style scoped>
.sidebar-item-content {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.sidebar-item-content svg {
  flex-shrink: 0;
  opacity: 0.5;
}
.sidebar-item.active .sidebar-item-content svg { opacity: 1; }
.sidebar-item-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--type-sm);
}
.sidebar-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  min-width: 1.2em;
  text-align: right;
}
.sidebar-divider {
  height: 1px;
  background: var(--color-border);
  margin: var(--space-1) var(--space-3);
}
</style>
