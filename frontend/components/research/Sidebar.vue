<template>
  <!--
    role="tablist" and focusable items.

    These were <div>s with @click: not reachable by keyboard, not announced.
    On the owner page that is an annoyance with workarounds — the search box,
    the mindmap, a typed URL. On a shared view it is terminal: the sidebar is
    the only navigation there is, the chrome is gone, and "All entries" and
    "External links" have no other entry point at all. A keyboard-only visitor
    could reach the first section and nothing else.
  -->
  <nav class="sidebar" role="tablist" aria-label="Sections">
    <!-- All entries -->
    <div
      role="tab"
      tabindex="0"
      :aria-selected="activeSection === '__all__'"
      :class="['sidebar-item', { active: activeSection === '__all__' }]"
      @click="$emit('update:activeSection', '__all__')"
      @keydown.enter.prevent="$emit('update:activeSection', '__all__')"
      @keydown.space.prevent="$emit('update:activeSection', '__all__')"
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
      role="tab"
      tabindex="0"
      :aria-selected="activeSection === section.id"
      :class="['sidebar-item', { active: activeSection === section.id }]"
      @click="$emit('update:activeSection', section.id)"
      @keydown.enter.prevent="$emit('update:activeSection', section.id)"
      @keydown.space.prevent="$emit('update:activeSection', section.id)"
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
      role="tab"
      tabindex="0"
      :aria-selected="activeSection === '__links__'"
      :class="['sidebar-item', { active: activeSection === '__links__' }]"
      @click="$emit('update:activeSection', '__links__')"
      @keydown.enter.prevent="$emit('update:activeSection', '__links__')"
      @keydown.space.prevent="$emit('update:activeSection', '__links__')"
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
/* Moved out of the global stylesheet: these style this component and
   nothing else, and they were a directory away from the markup they
   describe. What stays global is what three unrelated components share. */
.sidebar {
  position: sticky;
  top: calc(var(--space-4) + var(--nav-h));
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.sidebar-item {
  display: flex;
  flex-direction: column;
  padding: var(--space-2) var(--space-3);
  border-radius: 0;
  border-left: 2px solid transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color var(--transition-fast), border-color var(--transition-fast);
}
.sidebar-item:hover {
  color: var(--color-text);
}
.sidebar-item.active {
  color: var(--color-text);
  border-left-color: var(--color-primary);
}
.sidebar-item.active .sidebar-item-name {
  font-weight: var(--weight-medium);
}
.sidebar-progress {
  height: 2px;
  background: rgba(148, 163, 184, 0.08);
  border-radius: 1px;
  margin-top: var(--space-1);
  overflow: hidden;
}
.sidebar-progress-fill {
  height: 100%;
  background: var(--color-success);
  border-radius: 1px;
  transition: width 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}
/* The mobile shape: a horizontal strip instead of a rail. Its `@media`
   wrapper was lost when these rules were moved out of the global sheet, so
   for one commit the desktop sidebar was a scroller inside a 240px grid
   column with every progress bar hidden. */
@media (max-width: 768px) {
  .sidebar {
    position: static;
    flex-direction: row;
    gap: var(--space-1);
    overflow-x: auto;
    padding-bottom: var(--space-2);
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none;
  }
  .sidebar::-webkit-scrollbar { display: none; }
  .sidebar .sidebar-divider { display: none; }
  .sidebar-item {
    flex-shrink: 0;
    flex-direction: row;
    white-space: nowrap;
    padding: var(--space-2) var(--space-3);
    border-left: none;
    border-bottom: 2px solid transparent;
    font-size: var(--type-xs);
  }
  .sidebar-item.active {
    color: var(--color-primary);
    border-left-color: transparent;
    border-bottom-color: var(--color-primary);
  }
  .sidebar-progress { display: none; }
}

/* There was no focus style, because nothing here was focusable. */
.sidebar-item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
  border-radius: var(--radius-sm);
}
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
