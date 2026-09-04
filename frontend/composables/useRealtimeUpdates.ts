export interface WsEvent {
  type: string              // e.g. "research.updated", "entry.created", "access.revoked"
  research_id: string
  entity_id: string
  entity: string            // "research", "section", "entry", "entry_view", "session", "question", "task", "team", "crossref"
  /** The same scope named the way the URLs are. Absent on events with no research. */
  research_code?: string
  /** Who caused it. Empty when auth is off, and for anything an agent wrote. */
  actor_user_id?: string
  /** Which tab caused it. This is the one to compare against your own. */
  actor_client_id?: string
  /** The entity this one hangs off, for entities that are not addressable on
   *  their own. An annotation event names the annotation in `entity_id`, which
   *  tells an open document page nothing — the entry it is attached to is what
   *  decides whether the event concerns what is on screen. Both identities
   *  travel, for the same reason `research_code` does. */
  parent_id?: string
  parent_code?: string
  /** Why, where the type alone does not say — chiefly on access.revoked. */
  reason?: string
  /** Human name of the entity, sent only where the recipient can no longer fetch it. */
  name?: string
  /** Server send time, in milliseconds. */
  at?: number
}

export type RealtimeState = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'offline'
export type RealtimeReason = 'auth' | 'gave_up' | null

type EventHandler = (event: WsEvent) => void
type ResyncHandler = () => void

const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30_000
/** Attempts before the connection is called lost rather than interrupted. */
const GIVE_UP_AFTER = 6
/** Server close code for a credential that stopped being valid. */
const CLOSE_UNAUTHORIZED = 4401

const handlers = new Set<EventHandler>()
const resyncHandlers = new Set<ResyncHandler>()

const state = ref<RealtimeState>('idle')
const reason = ref<RealtimeReason>(null)
const lastSyncedAt = ref<number | null>(null)

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let attempt = 0
let everConnected = false
let started = false
let openedWith: string | null = null
/**
 * The share link this tab is reading, if any.
 *
 * A share connection is a different credential in the same slot: it identifies
 * a scope rather than a person, and the server refuses it on the `?token=`
 * parameter. It is module state because `start()`'s watcher runs once for the
 * life of the tab and has to see it however late the shared page mounts.
 */
let shareCredential: { token: string; unlock: string } | null = null

/**
 * clientId identifies this tab, for the life of this tab.
 *
 * It is what a write is stamped with and what the resulting event is checked
 * against, because neither of the alternatives works: with auth off there is no
 * user id at all, so a browser write and an agent write are indistinguishable;
 * and with auth on, two tabs of one person share a user id, so one tab would
 * suppress a change the other made and must show.
 *
 * sessionStorage, not localStorage: a second tab must get a different id.
 */
const clientId = ((): string => {
  if (import.meta.server) return ''
  const existing = sessionStorage.getItem('ws_client_id')
  if (existing) return existing
  const id = (crypto.randomUUID?.() ?? `c${Date.now()}${Math.random().toString(36).slice(2)}`)
  sessionStorage.setItem('ws_client_id', id)
  return id
})()

function socketUrl(token: string | null) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const config = useRuntimeConfig()
  const base = config.public.apiBase || `${protocol}//${window.location.host}`
  const url = base.replace(/^http/, 'ws') + '/ws'
  // A share link opens its own kind of connection, scoped to one research.
  // `?share=` and `?token=` are different parameters on purpose: the server
  // must never be able to mistake one credential for the other.
  if (shareCredential) {
    const unlock = shareCredential.unlock ? `&unlock=${encodeURIComponent(shareCredential.unlock)}` : ''
    return `${url}?share=${encodeURIComponent(shareCredential.token)}${unlock}`
  }
  // The token rides in the query string because a browser cannot set headers
  // on a WebSocket handshake — the same accommodation the SSE transport makes.
  // The hub decides what each connection may see from who it belongs to, so
  // without this an authenticated server sends nothing.
  return url + (token ? `?token=${encodeURIComponent(token)}` : '')
}

/**
 * Point this tab's connection at a share link.
 *
 * Called by the shared-view shell. It reopens the socket, because a connection
 * already carrying somebody's session must not keep feeding a page that is now
 * showing an anonymous view of one research.
 */
export function useShareSocket(token: string, unlock: string) {
  if (import.meta.server) return
  const same = shareCredential?.token === token && shareCredential?.unlock === unlock
  shareCredential = { token, unlock }
  start()
  if (same && socket) return
  everConnected = false
  attempt = 0
  disconnect()
  connect(null)
}

/** Give the socket back to whoever is signed in, if anybody is. */
export function releaseShareSocket() {
  if (import.meta.server || !shareCredential) return
  shareCredential = null
  everConnected = false
  attempt = 0
  disconnect()
  const { authEnabled, token } = useAuth()
  connect(authEnabled.value ? (token.value ?? null) : null)
}

function connect(token: string | null) {
  if (import.meta.server) return
  // Any live socket counts, including one still shaking hands. The old guard
  // admitted only OPEN, so a second subscriber arriving mid-handshake — which
  // is every page, since app.vue subscribes a tick earlier — built a second
  // socket over the first. Both stayed connected and both fed the same handler
  // set, so every event was handled twice and every page refetched twice.
  if (socket) return

  clearReconnect()
  openedWith = token
  if (state.value !== 'offline') state.value = everConnected ? 'reconnecting' : 'connecting'

  const ws = new WebSocket(socketUrl(token))
  socket = ws

  ws.onopen = () => {
    attempt = 0
    state.value = 'connected'
    reason.value = null
    lastSyncedAt.value = Date.now()
    // Whatever happened while the socket was down was sent to nobody and is not
    // replayed — the only way back to the truth is to ask again. Without this a
    // three-second hiccup left the page quietly wrong for as long as it stayed
    // open, looking exactly like a page that was up to date.
    if (everConnected) resyncHandlers.forEach((h) => h())
    everConnected = true
  }

  ws.onmessage = (msg) => {
    try {
      const event: WsEvent = JSON.parse(msg.data)
      lastSyncedAt.value = Date.now()
      handlers.forEach((h) => h(event))
    } catch { /* ignore parse errors */ }
  }

  ws.onclose = (ev) => {
    if (socket === ws) socket = null
    if (!started) {
      state.value = 'idle'
      return
    }
    if (ev.code === CLOSE_UNAUTHORIZED) {
      // The server closed it because the credential stopped being valid.
      // Retrying with the same one only repeats the answer.
      state.value = 'offline'
      reason.value = 'auth'
      return
    }
    state.value = attempt >= GIVE_UP_AFTER ? 'offline' : 'reconnecting'
    if (state.value === 'offline') {
      reason.value = 'gave_up'
      // A handshake refused for want of a credential arrives here as a plain
      // 1006, indistinguishable from a dead network — so before settling on
      // "the server keeps failing" we ask the API which it was. Telling
      // somebody their connection is flaky when they have actually been signed
      // out sends them to the wrong remedy.
      void classifyGiveUp()
    }
    scheduleReconnect()
  }

  ws.onerror = () => ws.close()
}

async function classifyGiveUp() {
  // A share connection has no account to ask about, and `/api/auth/me` would
  // answer 401 for a link that is working perfectly well.
  if (shareCredential) return
  const token = openedWith
  if (!token) return
  try {
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    await $fetch(`${base}/api/auth/me`, { headers: { Authorization: `Bearer ${token}` } })
  } catch (e: any) {
    if (e?.response?.status === 401) reason.value = 'auth'
  }
}

function scheduleReconnect() {
  if (reconnectTimer) return
  // Exponential with jitter: a server restart used to bring every open tab back
  // in lockstep every three seconds, forever.
  const backoff = Math.min(RECONNECT_MAX_MS, RECONNECT_BASE_MS * 2 ** attempt)
  const delay = backoff * (0.7 + Math.random() * 0.6)
  attempt += 1
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    if (started) connect(openedWith)
  }, delay)
}

function clearReconnect() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

function disconnect() {
  clearReconnect()
  const ws = socket
  socket = null
  state.value = 'idle'
  reason.value = null
  if (ws) {
    // Detached first: onclose cannot tell a network drop from a close we asked
    // for, and left attached it would schedule a reconnect for the session we
    // are in the middle of ending.
    ws.onclose = null
    ws.onerror = null
    ws.close()
  }
}

/** Retry now instead of waiting out the backoff. */
function retryNow() {
  if (!started || socket) return
  if (state.value !== 'reconnecting' && state.value !== 'offline') return
  // An expired credential is not retryable by reconnecting; only signing in
  // again changes the answer.
  if (reason.value === 'auth') return
  attempt = 0
  reason.value = null
  clearReconnect()
  connect(openedWith)
}

/**
 * start wires the connection to who is signed in.
 *
 * The socket used to be opened once and never revisited, which cost two ways:
 * on a server with auth enabled it connected before login and spun on 401s, and
 * after `logout()` it stayed open — still identified as the previous user — so
 * the next account in the same tab received the previous one's events. Watching
 * the auth state is what ties the connection's identity to the session's.
 */
function start() {
  if (started || import.meta.server) return
  started = true

  const { authEnabled, token } = useAuth()

  watch(
    [authEnabled, token],
    ([enabled, tok]) => {
      // null means we have not asked the server yet; connecting now would only
      // guess at whether a credential is required.
      if (enabled === null) return

      // A shared page owns the connection while it is open. Signing out in
      // another tab must not close the socket a visitor is reading over.
      if (shareCredential) return

      const wanted = enabled ? (tok ?? null) : null
      if (enabled && !wanted) {
        // Nobody is signed in. That is a waiting room, not a failure.
        everConnected = false
        attempt = 0
        disconnect()
        return
      }
      if (socket && openedWith === wanted) return

      everConnected = false
      attempt = 0
      disconnect()
      connect(wanted)
    },
    { immediate: true },
  )

  // A laptop that was asleep comes back with a socket the browser has already
  // given up on, and possibly a long backoff still to wait out.
  window.addEventListener('online', retryNow)
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') retryNow()
  })
}

/**
 * This tab's id, with no side effects — safe to call from anywhere, including
 * from inside useAuth, where reaching back into the connection would recurse.
 */
export function realtimeClientId() {
  return clientId
}

/** Did this tab cause the event? */
export function isSelf(event: WsEvent) {
  return !!event.actor_client_id && event.actor_client_id === clientId
}

/** Did this person cause it, in some other tab? */
export function isSelfUser(event: WsEvent, userId?: string | null) {
  return !!userId && event.actor_user_id === userId && !isSelf(event)
}

/**
 * Subscribe to real-time events.
 *
 * @param handler   Called for every event that reaches this client. The server
 *                  only sends what the connection is allowed to see, so this is
 *                  about relevance, not permission.
 * @param options.onResync
 *                  Called after the connection comes back, because events sent
 *                  while it was down are gone. Refetch here.
 *
 * Usage:
 *   useRealtimeUpdates(
 *     (event) => { if (event.entity === 'entry') refresh() },
 *     { onResync: refresh },
 *   )
 */
export function useRealtimeUpdates(handler: EventHandler, options: { onResync?: ResyncHandler } = {}) {
  if (import.meta.server) return

  handlers.add(handler)
  const onResync = options.onResync
  if (onResync) resyncHandlers.add(onResync)
  start()

  onUnmounted(() => {
    handlers.delete(handler)
    if (onResync) resyncHandlers.delete(onResync)
  })
}

/** Connection state, for the chrome that reports it. */
export function useRealtimeStatus() {
  if (!import.meta.server) start()
  return {
    state: readonly(state),
    reason: readonly(reason),
    lastSyncedAt: readonly(lastSyncedAt),
    clientId,
    retryNow,
  }
}
