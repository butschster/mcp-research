export type ToastVariant = 'success' | 'error' | 'info'

export interface ToastAction {
  label: string
  /** Runs on click. The toast closes afterwards unless this returns false. */
  onClick: () => void | boolean
}

export interface Toast {
  id: number
  variant: ToastVariant
  title?: string
  message: string
  action?: ToastAction
  /** Milliseconds until it dismisses itself; 0 means it waits for the reader. */
  timeout: number
}

export interface ToastInput {
  variant?: ToastVariant
  title?: string
  message: string
  action?: ToastAction
  timeout?: number
}

/**
 * The app's notifications.
 *
 * State lives at module scope, so any component can raise one and a single host
 * renders them all — the same shape `useAuth` uses. Without that, every page
 * that wants to say "done" grows its own line of text somewhere in its layout,
 * which is how a confirmation ends up under a toolbar.
 *
 * Rules the defaults encode:
 *
 *   - **A success dismisses itself; a failure does not.** An error usually
 *     carries the only way to retry, and taking that away on a timer is how a
 *     reader loses the thread.
 *   - **Four at a time.** Past that the oldest goes, because a stack taller than
 *     the corner is a wall, not a notification.
 */
const items = ref<Toast[]>([])
const timers = new Map<number, { handle: ReturnType<typeof setTimeout>; endsAt: number }>()
let nextId = 1

const MAX_VISIBLE = 4
const DEFAULT_TIMEOUT: Record<ToastVariant, number> = {
  success: 5000,
  info: 6000,
  error: 0,
}

export function useToasts() {
  function push(input: ToastInput): number {
    const variant = input.variant ?? 'info'
    const toast: Toast = {
      id: nextId++,
      variant,
      title: input.title,
      message: input.message,
      action: input.action,
      timeout: input.timeout ?? DEFAULT_TIMEOUT[variant],
    }

    items.value = [...items.value, toast]
    while (items.value.length > MAX_VISIBLE) {
      dismiss(items.value[0]!.id)
    }
    if (toast.timeout > 0) arm(toast.id, toast.timeout)
    return toast.id
  }

  function dismiss(id: number) {
    const timer = timers.get(id)
    if (timer) {
      clearTimeout(timer.handle)
      timers.delete(id)
    }
    items.value = items.value.filter((t) => t.id !== id)
  }

  function clear() {
    for (const t of [...items.value]) dismiss(t.id)
  }

  /** Stops the countdown while a reader is hovering or tabbing through. */
  function hold() {
    for (const [id, timer] of timers) {
      clearTimeout(timer.handle)
      timers.set(id, { handle: timer.handle, endsAt: timer.endsAt })
    }
  }

  /** Restarts each countdown with whatever it had left. */
  function resume() {
    const now = Date.now()
    for (const [id, timer] of [...timers]) {
      arm(id, Math.max(1000, timer.endsAt - now))
    }
  }

  function success(message: string, title = 'Done') {
    return push({ variant: 'success', title, message })
  }

  function error(message: string, title = 'Something went wrong', action?: ToastAction) {
    return push({ variant: 'error', title, message, action })
  }

  return { toasts: items, push, dismiss, clear, hold, resume, success, error }
}

function arm(id: number, ms: number) {
  const existing = timers.get(id)
  if (existing) clearTimeout(existing.handle)
  timers.set(id, {
    handle: setTimeout(() => {
      timers.delete(id)
      items.value = items.value.filter((t) => t.id !== id)
    }, ms),
    endsAt: Date.now() + ms,
  })
}
