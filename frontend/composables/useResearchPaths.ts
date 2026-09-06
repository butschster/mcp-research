/**
 * Where a link inside a research points.
 *
 * Every one of these used to be a template literal written by hand, in twenty
 * files. That was fine while there was one place a research could be read from;
 * a share link is a second, and hand-written `/research/…` in a component
 * rendered under `/s/{token}` is a link that drops an anonymous visitor onto the
 * login wall.
 *
 * Plain functions rather than a composable object: they are called from
 * `renderRefs`, which is a plain function applied to markdown, and from
 * templates. They read the same module state `useShare` owns, so a component
 * that knows nothing about sharing gets the right href by asking the same
 * question it always asked.
 */

function shareBase() {
  return shareActive() ? `/s/${shareToken()}` : ''
}

/** The research itself. */
export function researchPath(slug: string) {
  return shareActive() ? shareBase() : `/research/${slug}`
}

export function entryPath(slug: string, code: string) {
  return shareActive() ? `${shareBase()}/entry/${code}` : `/research/${slug}/entry/${code}`
}

export function sessionPath(slug: string, code: string) {
  return shareActive() ? `${shareBase()}/session/${code}` : `/research/${slug}/session/${code}`
}

export function roadmapPath(slug: string, code: string) {
  return shareActive() ? `${shareBase()}/roadmap/${code}` : `/research/${slug}/roadmap/${code}`
}

export function roadmapsPath(slug: string) {
  return shareActive() ? `${shareBase()}/roadmaps` : `/research/${slug}/roadmaps`
}

export function tasksPath(slug: string) {
  return shareActive() ? `${shareBase()}/tasks` : `/research/${slug}/tasks`
}

/**
 * One question, on the session that asked it.
 *
 * The owner has a page per question. A share does not — the session page is
 * the nearest thing that exists and it lists the question — so under a link
 * this lands on the session. A question node exists on a shared canvas only
 * when the link includes sessions, so the destination is always there.
 */
export function questionPath(slug: string, sessionCode: string, questionCode: string) {
  return shareActive()
    ? `${shareBase()}/session/${sessionCode}`
    : `/research/${slug}/session/${sessionCode}/question/${questionCode}`
}

/** The knowledge graph: documents, sections and the references between them. */
export function graphPath(slug: string) {
  return shareActive() ? `${shareBase()}/graph` : `/research/${slug}/graph`
}

/** The mind map. Both views draw content a link always carries, so both exist under a share. */
export function mindmapPath(slug: string) {
  return shareActive() ? `${shareBase()}/mindmap` : `/research/${slug}/mindmap`
}

/**
 * One task, on the board that holds it.
 *
 * The board has no page of its own per task — a task is a card and a modal — so
 * this is the board with that card opened, which is what `?task=T4` means there.
 * A `task_ref` row links here from its code chip, and so does a bare `[[T4]]`
 * in prose.
 */
export function taskPath(slug: string, code: string) {
  return `${tasksPath(slug)}?task=${encodeURIComponent(code)}`
}

/**
 * The queue of marks.
 *
 * Empty under a share, like `foreignResearchPath`: annotations are working
 * process — who doubted which sentence — and a link shares the document, not
 * the doubts about it. The caller renders nothing rather than a dead link.
 */
export function annotationsPath(slug: string) {
  return shareActive() ? '' : `/research/${slug}/annotations`
}

/** Personal document-update queue. A public link has no reader identity. */
export function updatesPath(slug: string) {
  return shareActive() ? '' : `/research/${slug}/updates`
}

export function exportPath(slug: string) {
  return shareActive() ? `${shareBase()}/export` : `/research/${slug}/export`
}

/**
 * A reference to another research entirely.
 *
 * Under a share there is no such destination — the link reaches one research
 * and confirming that a second exists is the thing it must not do — so this
 * returns an empty string and the caller renders plain text.
 */
export function foreignResearchPath(code: string) {
  return shareActive() ? '' : `/research/${code}`
}
