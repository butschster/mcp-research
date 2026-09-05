import { onBeforeUnmount, ref } from 'vue'

/**
 * Copy a string, say whether it worked, and reset after two seconds.
 *
 * This existed nine times by hand before it existed once. The copies had
 * already drifted: `EmptyState` — thirty-seven call sites, including the
 * "Continue R1" handoff command — called `navigator.clipboard.writeText`
 * with no `try`. On plain HTTP `navigator.clipboard` is simply absent, which
 * is a normal way to run this product on a LAN, so the button threw inside an
 * async click handler and did nothing, silently, with no message.
 *
 * `failed` is separate from `announcement` on purpose. A refusal has to be
 * visible as well as announced: a sighted mouse user pressing a button that
 * does nothing learns nothing from an `sr-only` live region. It also does not
 * expire — see `copy`.
 */
export function useCopyToClipboard(resetAfterMs = 2000) {
  const copied = ref(false)
  const failed = ref(false)
  const announcement = ref('')
  let timer: ReturnType<typeof setTimeout> | undefined

  async function copy(
    text: string,
    messages: { success?: string, failure?: string } = {},
  ) {
    const success = messages.success ?? 'Copied to the clipboard'
    const failure = messages.failure ?? 'Could not copy — select the text and copy it yourself'

    try {
      // Both halves matter: the API is absent outside a secure context, and
      // present-but-refused when the permission is denied.
      if (!navigator.clipboard?.writeText) throw new Error('no clipboard')
      await navigator.clipboard.writeText(text)
      copied.value = true
      failed.value = false
      announcement.value = success
    } catch {
      copied.value = false
      failed.value = true
      announcement.value = failure
    }

    clearTimeout(timer)
    // Success reverts on its own; a failure does not. The failure message is
    // often the text itself, offered for selecting by hand — taking it away
    // after two seconds removes the only fallback the reader had.
    if (copied.value) {
      timer = setTimeout(() => {
        copied.value = false
        announcement.value = ''
      }, resetAfterMs)
    }
  }

  /** Clear a standing failure, once the reader has done what it asked. */
  function dismiss() {
    failed.value = false
    announcement.value = ''
  }

  onBeforeUnmount(() => clearTimeout(timer))

  return { copied, failed, announcement, copy, dismiss }
}
