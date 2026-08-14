interface WsEvent {
  type: string       // e.g. "research.updated", "entry.created"
  research_id: string
  entity_id: string
  entity: string     // "research", "section", "entry", "session", "question", "task"
}

type EventHandler = (event: WsEvent) => void

const handlers = new Set<EventHandler>()
let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

function connect() {
  if (import.meta.server) return
  if (socket?.readyState === WebSocket.OPEN) return

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const config = useRuntimeConfig()
  const base = config.public.apiBase || `${protocol}//${window.location.host}`
  // The token rides in the query string because a browser cannot set headers
  // on a WebSocket handshake — the same accommodation the SSE transport makes.
  // The hub decides what each connection may see from who it belongs to, so
  // without this an authenticated server sends nothing.
  const token = localStorage.getItem('auth_token')
  const wsUrl = base.replace(/^http/, 'ws') + '/ws' +
    (token ? `?token=${encodeURIComponent(token)}` : '')

  socket = new WebSocket(wsUrl)

  socket.onmessage = (msg) => {
    try {
      const event: WsEvent = JSON.parse(msg.data)
      handlers.forEach(h => h(event))
    } catch { /* ignore parse errors */ }
  }

  socket.onclose = () => {
    socket = null
    // Reconnect after 3 seconds
    if (!reconnectTimer) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null
        connect()
      }, 3000)
    }
  }

  socket.onerror = () => {
    socket?.close()
  }
}

/**
 * Subscribe to real-time WebSocket events.
 * Automatically connects on first use and reconnects on disconnect.
 *
 * @param handler - Called for every event. Filter by entity/type as needed.
 *
 * Usage:
 *   useRealtimeUpdates((event) => {
 *     if (event.entity === 'entry') refresh()
 *   })
 */
export function useRealtimeUpdates(handler: EventHandler) {
  if (import.meta.server) return

  handlers.add(handler)
  connect()

  onUnmounted(() => {
    handlers.delete(handler)
  })
}
