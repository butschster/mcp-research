<template>
  <aside :class="['graph-sidebar', { collapsed }]">
    <button class="sidebar-toggle" @click="$emit('update:collapsed', !collapsed)" :title="collapsed ? 'Open panel' : 'Close panel'">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path v-if="collapsed" d="m9 18 6-6-6-6"/>
        <path v-else d="m15 18-6-6 6-6"/>
      </svg>
    </button>

    <template v-if="!collapsed">
      <div class="sidebar-header">
        <slot name="back">
          <span class="sidebar-title">Knowledge Graph</span>
        </slot>
      </div>

      <!-- Stats -->
      <div class="sidebar-section">
        <div class="sidebar-stats">
          <span>{{ nodeCount }} nodes</span>
          <span>{{ edgeCount }} edges</span>
        </div>
      </div>

      <!-- Nodes -->
      <div class="sidebar-section">
        <div class="sidebar-section-title">Nodes</div>
        <label
          v-for="nt in nodeTypes"
          :key="nt.key"
          class="sidebar-check"
        >
          <input
            type="checkbox"
            :checked="visibleNodeTypes.has(nt.key)"
            @change="$emit('toggle-node-type', nt.key)"
          />
          <span class="check-dot" :style="{ background: nt.color }"></span>
          <span class="check-label">{{ nt.label }}</span>
          <span class="check-count">{{ nodeCountByType[nt.key] || 0 }}</span>
        </label>
      </div>

      <!-- Edges -->
      <div class="sidebar-section">
        <div class="sidebar-section-title">Edges</div>
        <label
          v-for="et in edgeTypes"
          :key="et.key"
          class="sidebar-check"
        >
          <input
            type="checkbox"
            :checked="visibleEdgeTypes.has(et.key)"
            @change="$emit('toggle-edge-type', et.key)"
          />
          <span class="check-label">{{ et.label }}</span>
          <span class="check-count">{{ edgeCountByType[et.key] || 0 }}</span>
        </label>
      </div>

      <!-- Filters -->
      <div class="sidebar-section">
        <div class="sidebar-section-title">Filters</div>
        <label class="sidebar-check">
          <input type="checkbox" :checked="hideOrphans" @change="$emit('update:hideOrphans', !hideOrphans)" />
          <span class="check-label">Hide orphan nodes</span>
        </label>
        <label class="sidebar-check">
          <input type="checkbox" :checked="showArrows" @change="$emit('update:showArrows', !showArrows)" />
          <span class="check-label">Show edge direction</span>
        </label>
      </div>

      <!-- Focus depth -->
      <div class="sidebar-section">
        <div class="sidebar-section-title">
          Focus depth
          <span class="depth-value">{{ focusDepth }}</span>
        </div>
        <input
          type="range"
          min="1"
          max="5"
          :value="focusDepth"
          @input="$emit('update:focusDepth', Number(($event.target as HTMLInputElement).value))"
          class="depth-slider"
        />
        <div class="depth-labels">
          <span>1</span>
          <span>2</span>
          <span>3</span>
          <span>4</span>
          <span>5</span>
        </div>
        <div class="sidebar-hint">Right-click a node to focus</div>
        <button
          v-if="hasFocus"
          class="btn-clear-focus"
          @click="$emit('clear-focus')"
        >Clear focus</button>
      </div>
    </template>
  </aside>
</template>

<script setup lang="ts">
interface NodeTypeFilter {
  key: string
  label: string
  color: string
}

interface EdgeTypeFilter {
  key: string
  label: string
}

defineProps<{
  collapsed: boolean
  nodeTypes: NodeTypeFilter[]
  edgeTypes: EdgeTypeFilter[]
  visibleNodeTypes: Set<string>
  visibleEdgeTypes: Set<string>
  nodeCountByType: Record<string, number>
  edgeCountByType: Record<string, number>
  nodeCount: number
  edgeCount: number
  hideOrphans: boolean
  showArrows: boolean
  focusDepth: number
  hasFocus: boolean
}>()

defineEmits<{
  'update:collapsed': [value: boolean]
  'toggle-node-type': [key: string]
  'toggle-edge-type': [key: string]
  'update:hideOrphans': [value: boolean]
  'update:showArrows': [value: boolean]
  'update:focusDepth': [value: number]
  'clear-focus': []
}>()
</script>

<style scoped>
.graph-sidebar {
  width: 240px;
  flex-shrink: 0;
  background: rgba(0,0,0,0.5);
  border-right: 1px solid rgba(255,255,255,0.06);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  transition: width 0.2s;
  position: relative;
}

.graph-sidebar.collapsed {
  width: 36px;
  overflow: hidden;
}

.sidebar-toggle {
  position: absolute;
  top: 10px;
  right: 8px;
  background: none;
  border: none;
  color: rgba(255,255,255,0.4);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-xs);
  z-index: 2;
}
.sidebar-toggle:hover {
  color: rgba(255,255,255,0.8);
  background: rgba(255,255,255,0.06);
}

.sidebar-header {
  padding: 12px 14px 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sidebar-title {
  font-size: 14px;
  font-weight: var(--weight-semibold);
  color: rgba(255,255,255,0.85);
}

.sidebar-section {
  padding: 10px 14px;
  border-top: 1px solid rgba(255,255,255,0.04);
}

.sidebar-section-title {
  font-size: 10px;
  font-weight: var(--weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: rgba(255,255,255,0.35);
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-stats {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: rgba(255,255,255,0.45);
}

.sidebar-check {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  cursor: pointer;
  font-size: 12px;
  color: rgba(255,255,255,0.7);
}

.sidebar-check input[type="checkbox"] {
  appearance: none;
  width: 14px;
  height: 14px;
  border: 1.5px solid rgba(255,255,255,0.2);
  border-radius: var(--radius-xs);
  background: transparent;
  cursor: pointer;
  position: relative;
  flex-shrink: 0;
}

.sidebar-check input[type="checkbox"]:checked {
  background: rgba(255,255,255,0.15);
  border-color: rgba(255,255,255,0.4);
}

.sidebar-check input[type="checkbox"]:checked::after {
  content: '';
  position: absolute;
  top: 1px;
  left: 4px;
  width: 4px;
  height: 7px;
  border: solid rgba(255,255,255,0.9);
  border-width: 0 1.5px 1.5px 0;
  transform: rotate(45deg);
}

.check-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.check-label {
  flex: 1;
}

.check-count {
  font-size: 11px;
  color: rgba(255,255,255,0.25);
  font-variant-numeric: tabular-nums;
}

.depth-slider {
  width: 100%;
  appearance: none;
  height: 4px;
  background: rgba(255,255,255,0.1);
  border-radius: var(--radius-hair);
  outline: none;
  margin: 4px 0 2px;
}

.depth-slider::-webkit-slider-thumb {
  appearance: none;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--hue-5);
  cursor: pointer;
  border: 2px solid #1a1a2e;
}

.depth-slider::-moz-range-thumb {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--hue-5);
  cursor: pointer;
  border: 2px solid #1a1a2e;
}

.depth-labels {
  display: flex;
  justify-content: space-between;
  font-size: 9px;
  color: rgba(255,255,255,0.2);
  padding: 0 2px;
}

.depth-value {
  color: var(--hue-5);
  font-weight: var(--weight-bold);
  font-size: 12px;
}

.sidebar-hint {
  font-size: 10px;
  color: rgba(255,255,255,0.2);
  margin-top: 6px;
  font-style: italic;
}

.btn-clear-focus {
  display: block;
  width: 100%;
  margin-top: 8px;
  padding: 5px;
  background: rgba(167,139,250,0.1);
  border: 1px solid rgba(167,139,250,0.25);
  border-radius: 5px;
  color: var(--hue-5);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-clear-focus:hover {
  background: rgba(167,139,250,0.2);
}
</style>
