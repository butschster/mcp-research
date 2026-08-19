/**
 * What `POST /api/sections/{id}/import/preview` hands back — the parse report
 * for one markdown file, before anything is written.
 *
 * Shape follows `ImportPreview` in `composables/useSectionImport.ts` exactly.
 * The interesting part is `metadata_report` plus the four sibling arrays
 * (`fields`, `ignored`, `refused`, `unresolved_refs`): the dialog folds all
 * five into one ledger, so a fixture that fills only `metadata_report` exercises
 * a third of it.
 *
 * Bodies carry `[[E3]]`-style cross-references because that is what an author
 * writes, and because an unresolvable one is the whole point of the
 * `unresolved_refs` group. Nothing here is rendered as markdown — the dialog
 * shows the document as source, deliberately.
 */
import type { ImportPreview } from '../composables/useSectionImport'

const CLEAN_BODY = `## What a seat means to each vendor

Three vendors, three definitions, and the difference is worth about 20% of a
mid-market bill.

- **Northwind** counts every account that has logged in once in the billing
  month. Dormant accounts are free.
- **Cascade** counts every account that exists, dormant or not. See [[E3]] for
  the invoice we reconstructed.
- **Halyard** counts concurrent sessions, which is a different unit and is why
  their list price looks half of everyone else's.

The comparison in [[E7]] normalises all three onto Northwind's definition.

> Numbers are list price, not negotiated, current as of the third week of the
> quarter.
`

const NOISY_BODY = `# Ценообразование по местам

Разбор того, что каждый поставщик считает «местом», и почему счёт за одно и то
же количество людей отличается вдвое.

Сводная таблица — в [[E7]]. Исходные счета лежат в [[R9:E2]], а разбор
формулировки в договоре — в [[E44]], которого в этом исследовании нет.

| Vendor   | Unit               | List  |
|----------|--------------------|-------|
| Northwind| active account     | $18   |
| Cascade  | provisioned account| $14   |
| Halyard  | concurrent session | $9    |

См. также [[E44]] по поводу пункта 7.3.
`

/** Sixty lines of it, so the 40-line clamp has something to clamp. */
const LONG_BODY = Array.from({ length: 312 }, (_, i) => {
  if (i === 0) return '## Migration log, run by run'
  if (i % 12 === 1) return ''
  if (i % 12 === 2) return `### Run ${Math.ceil(i / 12)} — ${new Date(Date.UTC(2026, 3, i % 28 + 1)).toISOString().slice(0, 10)}`
  return `${String(i).padStart(4, '0')}  rows migrated, checksum ok, see [[E${(i % 9) + 1}]] for the schema this run assumed`
}).join('\n')

/** A file that needed nothing said about it — the one-glance confirm. */
export const mockCleanPreview: ImportPreview = {
  filename: 'seat-definition.md',
  title: 'What counts as a seat',
  title_source: 'frontmatter',
  description: 'The definition three vendors disagree on, and what it costs.',
  status: 'completed',
  tags: ['pricing', 'definitions', 'competitive'],
  body: CLEAN_BODY,
  body_lines: CLEAN_BODY.split('\n').length,
  metadata: { source: 'vendor pricing pages', confidence: 'high' },
  metadata_report: {
    stored: ['source', 'confidence'],
    spec_version: 4,
  },
}

/** One of everything the ledger can say, in a single file. Five groups: one
 *  `attention`, four `note`. */
export const mockNoisyPreview: ImportPreview = {
  filename: 'Ценообразование по местам.md',
  title: 'Ценообразование по местам — сравнение трёх поставщиков и разбор того, почему счёт за одинаковое число людей отличается вдвое',
  title_source: 'heading',
  description: 'Разбор трёх определений «места» и нормализация цен к одному из них.',
  status: 'in_progress',
  tags: ['pricing', 'competitive', 'перевод'],
  body: NOISY_BODY,
  body_lines: NOISY_BODY.split('\n').length,
  metadata: { source: 'счета поставщиков', reviewed: false },
  metadata_report: {
    stored: ['source', 'reviewed'],
    spec_version: 7,
    invalid_values: [
      {
        key: 'confidence',
        value: 'довольно высокая, но по Halyard мы видели только публичный прайс',
        reason: 'This field takes one of: low, medium, high.',
      },
      { key: 'reviewed_on', value: '3rd week of Q3', reason: 'Not a date this section can read. Expected YYYY-MM-DD.' },
    ],
    missing_required: ['owner'],
    unknown_keys: [
      { key: 'reviewer', value: 'Пётр Бутенко', reason: 'Kept in the document, not stored as a field.' },
      { key: 'obsidian-vault', value: 'research', reason: 'Kept in the document, not stored as a field.' },
    ],
  },
  fields: [
    {
      key: 'status',
      value: 'needs-review',
      applied: false,
      reason: 'This research has no status by that name, so the entry was created as in_progress and can be moved by hand.',
    },
    {
      key: 'tags',
      value: 'Перевод',
      applied: true,
      reason: 'Lowercased to “перевод” to match the eleven entries already tagged that way.',
    },
  ],
  ignored: [
    { key: 'code', value: 'E12', reason: 'Codes are assigned here, not carried in from a file.' },
    { key: 'research', value: 'R7', reason: 'The destination is the section this was dropped into.' },
    { key: 'created', value: '2026-06-02T09:00:00Z', reason: 'Timestamps are recorded at import.' },
  ],
  refused: [
    { key: 'session', value: 'SS3', reason: 'An import is not attributed to a session.' },
    { key: 'author', value: 'claude-opus-4', reason: 'Authorship is the account doing the import.' },
    { key: 'revisions', reason: 'History is written here and cannot be supplied.' },
  ],
  unresolved_refs: [
    { ref: 'E44', count: 2 },
    { ref: 'R9:E2', count: 1 },
  ],
  warnings: [
    'The file declared UTF-16; it was decoded as UTF-8, which matched. If the text below looks wrong, re-save it as UTF-8 and drop it again.',
  ],
}

/** No front matter and no heading, so the title is a guess made from the
 *  filename — the case somebody usually wants to correct before committing. */
export const mockFilenameTitlePreview: ImportPreview = {
  filename: 'notes-from-the-halyard-call-2026-07-14.md',
  title: 'Notes from the halyard call 2026 07 14',
  title_source: 'filename',
  status: 'draft',
  tags: [],
  body: 'Rough notes, unedited.\n\nThey price per concurrent session. Confirmed twice.\nAsk about the annual floor before quoting anything to finance.\n',
  body_lines: 5,
  metadata_report: { spec_version: 4 },
}

/** A 312-line document, so the well shows its first 40 and says so. */
export const mockLongBodyPreview: ImportPreview = {
  filename: 'migration-log.md',
  title: 'Migration log, run by run',
  title_source: 'heading',
  description: 'Every run of the seat-table migration, with checksums.',
  status: 'completed',
  tags: ['migration', 'log'],
  body: LONG_BODY,
  body_lines: LONG_BODY.split('\n').length,
  metadata_report: { stored: [], spec_version: 4 },
}

/** The items the ledger passes down to one group, at the two sizes that matter:
 *  the single warning, and the list nobody reads to the end. */
export const mockNoteItems = mockNoisyPreview.metadata_report.unknown_keys!

export const mockManyNoteItems = Array.from({ length: 20 }, (_, i) => ({
  kind: i % 3 === 0 ? 'rejected' : undefined,
  key: ['reviewer', 'obsidian-vault', 'aliases', 'cssclass', 'publish', 'permalink', 'weight'][i % 7] + (i > 6 ? `-${i}` : ''),
  value: i % 2 === 0 ? `value-${i}` : undefined,
  reason: 'Kept in the document, not stored as a field.',
}))
