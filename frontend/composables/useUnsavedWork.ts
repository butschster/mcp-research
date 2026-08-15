/**
 * Whether anything on screen would be lost if the page were replaced.
 *
 * It exists for one caller: losing access to a research while its editor is
 * open and dirty. The notice that replaces the page body is the right answer to
 * a stale page and the wrong answer to an unsaved draft — the server has
 * already decided the save will fail, but that is no reason for the client to
 * throw the text away on the reader's behalf.
 *
 * A registry rather than a prop because the two ends are far apart: the editor
 * is several components inside a page, and the decision is taken in `app.vue`.
 */
const sources = new Set<() => boolean>()

/**
 * Declare that this component may be holding unsaved work.
 *
 * @param isDirty  Read at the moment the question is asked, not when it changes.
 */
export function useUnsavedWork(isDirty: () => boolean) {
  if (import.meta.server) return
  sources.add(isDirty)
  onUnmounted(() => sources.delete(isDirty))
}

/** Is anything on screen unsaved right now? */
export function hasUnsavedWork() {
  for (const isDirty of sources) {
    try {
      if (isDirty()) return true
    } catch {
      // A source that throws is a source that cannot vouch for itself; the safe
      // reading is that there is nothing to lose, since the alternative is a
      // page that can never be replaced.
    }
  }
  return false
}
