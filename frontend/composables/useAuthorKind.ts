/**
 * One vocabulary for the four kinds of authorship.
 *
 * There were three phrasings of the same four strings in two files in the same
 * folder — the badge's `word` and `label`, and a menu header's own map — with
 * the badge rendered inches from the words. A reader comparing "agent" with "An
 * agent" with "written by an agent" is reading three decisions where one was
 * made.
 *
 * `word` is what the badge prints beside its glyph. `phrase` opens a sentence
 * ("An agent · 2 hours ago"). `label` is the screen-reader sentence.
 */
export interface AuthorKindWords {
  glyph: string
  word: string
  phrase: string
  label: string
}

export const AUTHOR_KINDS: Record<string, AuthorKindWords> = {
  human: { glyph: '●', word: 'person', phrase: 'A person', label: 'written by a person' },
  agent: { glyph: '◇', word: 'agent', phrase: 'An agent', label: 'written by an agent' },
  import: { glyph: '⇩', word: 'import', phrase: 'An import', label: 'written by an import' },
  restore: { glyph: '↺', word: 'restore', phrase: 'A restore', label: 'restored from an earlier revision' },
}

/**
 * An unknown kind still renders. A new `author_kind` added on the server must
 * not blank a document's provenance line on an older client — the kind itself
 * is the least wrong thing to show.
 */
export function authorKind(kind: string): AuthorKindWords {
  return AUTHOR_KINDS[kind] ?? {
    glyph: '◇',
    word: kind,
    phrase: kind,
    label: `written by ${kind}`,
  }
}
