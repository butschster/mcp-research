<template>
  <!-- A button only when pressing it does something. It used to be one always:
       a permanent footer tab stop on every page that announced "button" and did
       nothing on Enter, including in the healthy state, which is most of the
       time. When there is nothing to press this is a plain status region. -->
  <component
    :is="actionable ? 'button' : 'div'"
    :type="actionable ? 'button' : undefined"
    class="connection-status"
    :class="{ 'is-actionable': actionable }"
    :title="tooltip"
    @click="onClick"
  >
    <span class="ws-dot" :class="`ws-${dot}`" aria-hidden="true"></span>
    <span class="ws-label" aria-hidden="true">{{ label }}</span>
    <!-- The one accessible name, and the announcement. `.ws-label` is hidden
         below 768px and the dot carries no text at any width, so without this
         the state is a colour. `role="status"` is what makes a change to it
         spoken at all — the alternative was twenty seconds of silence before
         the escalation toast. -->
    <span class="sr-only" role="status">{{ spoken }}</span>
  </component>
</template>

<script setup lang="ts">
import { clockTime } from '~/composables/useRelativeTime'
import { useConnectionBanner } from '~/composables/useConnectionBanner'
import type { BannerState } from '~/composables/useConnectionBanner'
import type { RealtimeReason } from '~/composables/useRealtimeUpdates'

/**
 * Reports whether what is on screen is still being kept up to date.
 *
 * It used to answer that by opening its own WebSocket — a second connection,
 * without a token, alongside the real one. On a server with auth enabled that
 * socket was rejected every time, so the indicator sat on "Offline" while live
 * updates were in fact arriving perfectly well on the connection it knew
 * nothing about. It now reports the one connection the app actually has.
 *
 * It is told rather than asking: `useConnectionBanner()` owns the state machine
 * and `app.vue` hands the result down. Keeping the socket out of here is what
 * stopped the second connection from existing at all, and it is the only way
 * this indicator can be seen in each of its states without a server that
 * misbehaves on cue.
 */
const props = withDefaults(
  defineProps<{
    /** The banner state machine's verdict — five states, three appearances. */
    state: BannerState
    /** Meaningful only while offline; 'auth' is the one nothing can retry. */
    reason?: RealtimeReason
    /** Epoch ms of the last event or successful open. */
    lastSyncedAt?: number | null
  }>(),
  { reason: null, lastSyncedAt: null },
)

const emit = defineEmits<{ retry: [] }>()

const dot = computed(() => {
  switch (props.state) {
    case 'live':
    case 'grace': return 'connected'
    case 'reconnecting':
    case 'stale': return 'reconnecting'
    default: return 'disconnected'
  }
})

const label = computed(() => {
  if (dot.value === 'connected') return 'Live'
  if (dot.value === 'reconnecting') return 'Reconnecting…'
  return 'Offline'
})

const syncedAt = computed(() => clockTime(props.lastSyncedAt))

const tooltip = computed(() => {
  if (dot.value === 'connected') return 'Real-time updates active'
  if (props.reason === 'auth') return 'Your session ended — sign in again'
  // "since" is when this connection last heard anything — which on a quiet
  // research is not the same as when it was last known good, so it is only
  // worth saying once the connection is in doubt.
  const since = syncedAt.value ? ` Last change seen at ${syncedAt.value}.` : ''
  if (dot.value === 'reconnecting') return `Reconnecting to server…${since}`
  return `Not receiving live updates.${since}`
})

// The name ends with what pressing it does, because "Offline, button" tells a
// reader that something is wrong and not that anything can be done about it.
const spoken = computed(() => {
  const base = `Live updates: ${label.value}. ${tooltip.value}`
  return actionable.value ? `${base} Activate to reconnect now.` : base
})

// A credential that stopped being valid is not fixed by trying again, and a
// control that does nothing when pressed is worse than one that is not offered.
const actionable = computed(() => dot.value !== 'connected' && props.reason !== 'auth')

function onClick() {
  if (actionable.value) emit('retry')
}
</script>

<style scoped>
.connection-status {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: var(--type-xs);
  font-weight: 500;
  color: var(--color-text-muted);
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  padding: var(--space-1) var(--space-2);
  cursor: default;
}
.connection-status.is-actionable { cursor: pointer; }
@media (max-width: 768px) {
  /* The label is hidden at this width, leaving a target too small to hit. */
  .connection-status { min-height: 36px; min-width: 36px; justify-content: center; }
}
.connection-status.is-actionable:hover { border-color: var(--color-border); }
.connection-status:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
.ws-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  transition: background var(--transition-base);
}
.ws-connected {
  background: var(--color-success);
  box-shadow: 0 0 6px rgba(52, 211, 153, 0.5);
}
/* Shape carries the state as well as colour does. Below 768px the label is
   hidden and the dot is all there is; under prefers-reduced-motion the pulse
   goes too, which would leave green against amber at 6px — the deuteranope
   pair — as the entire signal. */
.ws-reconnecting {
  background: transparent;
  border: 2px solid var(--color-warning);
  animation: pulse-dot 1.5s ease-in-out infinite;
}
/* Hollow, and at full opacity: at 0.5 over the page ground this was about
   2.2:1, under the 3:1 minimum, on the one state that most needs to be seen. */
.ws-disconnected {
  background: transparent;
  border: 2px solid var(--color-text-muted);
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

@media (prefers-reduced-motion: reduce) {
  .ws-reconnecting { animation: none; }
}
</style>
