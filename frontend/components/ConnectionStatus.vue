<template>
  <div class="connection-status" :title="tooltip">
    <span class="ws-dot" :class="`ws-${state}`"></span>
    <span class="ws-label">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
const state = ref<'connected' | 'reconnecting' | 'disconnected'>('disconnected')
const label = computed(() => {
  if (state.value === 'connected') return 'Live'
  if (state.value === 'reconnecting') return 'Reconnecting…'
  return 'Offline'
})
const tooltip = computed(() => {
  if (state.value === 'connected') return 'Real-time updates active'
  if (state.value === 'reconnecting') return 'Reconnecting to server…'
  return 'Not connected'
})

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

function connect() {
  if (import.meta.server) return
  if (socket?.readyState === WebSocket.OPEN) return

  const config = useRuntimeConfig()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const base = config.public.apiBase || `${protocol}//${window.location.host}`
  const wsUrl = base.replace(/^http/, 'ws') + '/ws'

  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    state.value = 'connected'
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  }

  socket.onclose = () => {
    socket = null
    state.value = 'reconnecting'
    if (!reconnectTimer) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null
        connect()
      }, 3000)
    }
  }

  socket.onerror = () => socket?.close()
}

onMounted(connect)
onUnmounted(() => {
  socket?.close()
  if (reconnectTimer) clearTimeout(reconnectTimer)
})
</script>

<style scoped>
.connection-status {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  cursor: default;
}
.ws-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.ws-connected    { background: var(--color-success); }
.ws-reconnecting { background: var(--color-warning); }
.ws-disconnected { background: var(--color-text-muted); }
</style>
