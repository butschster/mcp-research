/**
 * Agent-authored fields that contain markup.
 *
 * Thirteen components take a raw field — an entry description, a task title, a
 * question — run it through `renderRefs` and hand the result to `v-html`. Until
 * `renderRefs` escaped, all of that reached the DOM verbatim, and **no story in
 * this catalogue rendered a `<`**, so the catalogue could not show the bug. The
 * stories that use these strings exist to make the next regression visible on
 * sight.
 *
 * On the payloads:
 *
 *   - `<script>` inserted through `innerHTML` never executes, so it proves
 *     nothing on its own. It is here because it is what an author actually
 *     types when they mean "this is a string, not a tag", and it must render as
 *     one.
 *   - The `onerror` is the one that would fire. It replaces itself with the
 *     words `XSS EXECUTED` rather than calling `alert()`: a modal dialog blocks
 *     the docs page it appears on and says nothing a visible marker does not,
 *     and a story that has to be dismissed before the rest of the page can be
 *     read is a story people learn to skip.
 *   - Every string carries `[[E3]]`, because escaping that breaks the linking
 *     is not a fix. The reference has to survive as a link with the markup
 *     beside it as text.
 */

/** The `onerror` payload, alone, for a string that has no room for prose. */
export const markupImg = `<img src=x onerror="this.replaceWith('XSS EXECUTED')">`

/** An entry description. Long-form: this is the field EntryCard renders. */
export const markupDescription =
  `Escaping rules for author-supplied HTML: <b>bold</b>, ` +
  `<script>alert('xss')</script> and ${markupImg} all belong in the corpus. ` +
  `Compare with [[E3]].`

/** A question, as an interviewer would write it with a tag in the text. */
export const markupQuestionText =
  `Does the renderer treat <b>bold</b> in a question as markup, and what does ` +
  `it do with ${markupImg}? Related to [[E3]].`

/** A task title — one line, no wrapping expected. */
export const markupTaskTitle =
  `Escape <b>agent-authored</b> HTML: <script>alert(1)</script>, ${markupImg} — see [[E3]]`

/** A research name, for the search result list. */
export const markupResearchName = `<b>Rendering</b> pipeline ${markupImg}`
