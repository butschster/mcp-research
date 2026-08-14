import type { RealtimeState } from './useRealtimeUpdates'

/**
 * Turns three connection states into what a person should be told, and when.
 *
 * The whole design is the grace window. A closed lid, a wifi handover and a
 * `make build` restart all produce a gap of a second or four; a banner for each
 * of those is a banner nobody reads by Thursday. So nothing changes on screen
 * until the gap outlasts patience, nothing is announced until it outlasts
 * usefulness, and a connection that comes back on its own says nothing at all.
 */

/** Below this, a gap is invisible: the indicator still reads "Live". */
const GRACE_MS = 5_000
/** Past this, the reader is told — once — that the page has stopped updating. */
const ESCALATE_MS = 20_000
/** Connected for this long ends the episode, so a flapping link cannot re-alarm. */
const EPISODE_CLEAR_MS = 10_000

export type BannerState = 'live' | 'grace' | 'reconnecting' | 'stale' | 'offline'

const display = ref<BannerState>('live')

let degradedSince: number | null = null
let graceTimer: ReturnType<typeof setTimeout> | null = null
let escalateTimer: ReturnType<typeof setTimeout> | null = null
let recoveredTimer: ReturnType<typeof setTimeout> | null = null
let toastId: number | null = null
// What the standing toast is about, so a warning can be replaced by an error
// rather than suppressed by it.
let toastFor: 'stale' | 'offline' | 'auth' | null = null
let installed = false

function clearTimers() {
  for (const t of [graceTimer, escalateTimer]) if (t) clearTimeout(t)
  graceTimer = escalateTimer = null
}

export function useConnectionBanner() {
  const { state, reason, lastSyncedAt, retryNow } = useRealtimeStatus()
  const toasts = useToasts()
  const { logout } = useAuth()
  const route = useRoute()

  function escalate() {
    const kind: 'stale' | 'offline' | 'auth' =
      reason.value === 'auth' ? 'auth' : state.value === 'offline' ? 'offline' : 'stale'
    // Standing toast already says this. A worse verdict replaces it; the same
    // one does not repeat.
    if (toastId !== null && toastFor === kind) return
    if (toastId !== null) {
      toasts.dismiss(toastId)
      toastId = null
    }
    toastFor = kind

    if (kind === 'auth') {
      toastId = toasts.push({
        variant: 'error',
        title: 'Your session ended',
        message: 'You were signed out, or your access key was revoked. Sign in again to keep getting live updates.',
        timeout: 0,
        action: { label: 'Sign in again', onClick: () => logout(`/login?next=${encodeURIComponent(route.fullPath)}`) },
      })
      return
    }

    const since = clockTime(lastSyncedAt.value)
    toastId = toasts.push({
      variant: state.value === 'offline' ? 'error' : 'info',
      title: state.value === 'offline' ? 'Not receiving live updates' : 'Live updates are paused',
      message: state.value === 'offline'
        ? 'The connection to the server keeps failing. Your page will not update on its own until it is back.'
        : `Reconnecting${since ? ` since ${since}` : ''}. Anything changed since then is not on your screen yet.`,
      timeout: 0,
      // Returning false keeps the toast up. Without it, one press that does not
      // reconnect dismissed the only channel that reports the outage — and the
      // guard above then held `toastId` for the rest of the tab's life, so the
      // reader was never told again.
      action: { label: 'Try again', onClick: () => { retryNow(); return false } },
    })
  }

  function recovered() {
    // Silence is the reward for a healthy connection: if nothing was ever said,
    // nothing is said about it being over either.
    if (toastId === null) return
    toasts.dismiss(toastId)
    toastId = null
    toastFor = null
    toasts.push({ variant: 'success', title: 'Back in sync', message: 'Reconnected and refreshed what was on screen.' })
  }

  function evaluate(current: RealtimeState) {
    clearTimers()

    if (current === 'connected' || current === 'idle') {
      display.value = 'live'
      degradedSince = null
      if (recoveredTimer) clearTimeout(recoveredTimer)
      // The episode is not over the instant it reconnects — a link that flaps
      // every two seconds would otherwise raise a fresh alarm every cycle.
      recoveredTimer = setTimeout(recovered, EPISODE_CLEAR_MS)
      return
    }

    if (recoveredTimer) {
      clearTimeout(recoveredTimer)
      recoveredTimer = null
    }

    if (current === 'offline') {
      display.value = 'offline'
      escalate()
      return
    }

    // 'connecting' on first load is treated as a gap, not a failure: the footer
    // must never say Offline while the first handshake is still in flight.
    if (degradedSince === null) degradedSince = Date.now()
    const elapsed = Date.now() - degradedSince

    if (elapsed < GRACE_MS) {
      display.value = 'grace'
      graceTimer = setTimeout(() => evaluate(current), GRACE_MS - elapsed)
      return
    }
    display.value = 'reconnecting'
    if (elapsed < ESCALATE_MS) {
      escalateTimer = setTimeout(() => evaluate(current), ESCALATE_MS - elapsed)
      return
    }
    display.value = 'stale'
    escalate()
  }

  if (!installed && !import.meta.server) {
    installed = true
    watch(state, (s) => evaluate(s), { immediate: true })
    // A timer armed before the laptop slept fires late and would paint
    // "Reconnecting…" over a connection that is already back.
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') evaluate(state.value)
    })
  }

  return { display: readonly(display), state, reason, lastSyncedAt, retryNow }
}
