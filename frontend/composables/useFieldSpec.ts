/**
 * One definition of what a section declares and what counts as an answer.
 *
 * Three surfaces ask the same questions — the metadata block on a document, the
 * section table, and the field editor in settings — and "is this filled?" is
 * exactly the kind of predicate that gets written three slightly different ways
 * and then disagrees on screen. A document reading "complete" in the table and
 * "1 field missing" on its own page is the same class of bug as the stored
 * status disagreeing with the prose, which is the failure this whole feature
 * exists to end.
 */

export type FieldType = 'enum' | 'ref' | 'date' | 'text' | 'number' | 'url'

export interface FieldSpec {
  key: string
  label?: string
  type: FieldType
  required?: boolean
  repeated?: boolean
  options?: string[]
  help?: string
}

export interface MetadataStatus {
  missing_required?: string[]
  orphaned?: string[]
  issues?: { key: string; reason: string }[]
  complete?: boolean
  spec_version?: number
}

/** The label a person should see. Falling back to the key beats an empty cell. */
export function fieldLabel(f: FieldSpec): string {
  return f.label?.trim() || f.key
}

/**
 * A value counts as filled when somebody put something there — including an
 * explicit unknown, which is `null`.
 *
 * That distinction is the point: absent means nobody was asked, `null` means
 * somebody looked and could not say. A model, unlike a person, does not leave a
 * field blank when it does not know, so the writable unknown is what stops a
 * required field from manufacturing a fact.
 */
export function isFilled(value: unknown): boolean {
  if (value === null) return true
  if (value === undefined) return false
  if (typeof value === 'string') return value.trim() !== ''
  if (Array.isArray(value)) return value.length > 0
  return true
}

/** True only for the explicit unknown. */
export function isUnknown(value: unknown): boolean {
  return value === null
}

/**
 * How a value reads. Lists join with a comma; an unfilled field returns an
 * empty string so the caller can choose its own placeholder — the block says
 * "Not filled", the table an em dash, and those are different sentences.
 */
export function formatValue(value: unknown): string {
  if (value === null || value === undefined) return ''
  if (Array.isArray(value)) return value.map(v => formatValue(v)).filter(Boolean).join(', ')
  if (typeof value === 'number') return String(value)
  if (typeof value === 'string') return value
  return String(value)
}

/**
 * How a value should be handed to `renderRefs`.
 *
 * A `ref` is stored bare — `E47`, not `[[E47]]` — because the brackets are
 * syntax, not part of the code. So the brackets go back on for display, or the
 * one field type whose entire purpose is to point at something renders as inert
 * text while the same value is a live link in the Obsidian export.
 */
export function displayValue(f: FieldSpec, value: unknown): string {
  const text = formatValue(value)
  if (f.type !== 'ref' || !text) return text
  return text.split(', ').map(part => (part ? `[[${part}]]` : part)).join(', ')
}

/** Which declared, required fields this document does not answer. */
export function missingRequired(specs: FieldSpec[], values: Record<string, unknown>): string[] {
  return specs.filter(f => f.required && !isFilled(values?.[f.key])).map(f => f.key)
}

/**
 * Values kept under keys the section no longer declares. They are never
 * deleted — dropping a field is a decision about future collection, not a
 * verdict on what was already recorded — so they are shown apart.
 */
export function orphanedKeys(specs: FieldSpec[], values: Record<string, unknown>): string[] {
  const declared = new Set(specs.map(f => f.key))
  return Object.keys(values ?? {}).filter(k => !declared.has(k)).sort()
}

/**
 * The input a type wants. Everything single-line shares `.text-input` so a row
 * of controls comes out at one height.
 */
export function inputKind(f: FieldSpec): 'select' | 'date' | 'number' | 'text' {
  if (f.type === 'enum' && !f.repeated) return 'select'
  if (f.type === 'date') return 'date'
  if (f.type === 'number') return 'number'
  return 'text'
}

/**
 * Parses what a person typed back into the shape the field promises. Only the
 * repeated case needs splitting; everything else is handed over as typed and
 * validated by the server, which is the only place that decides.
 */
export function parseInput(f: FieldSpec, raw: string): unknown {
  const trimmed = raw.trim()
  if (trimmed === '') return ''
  if (f.repeated) {
    return trimmed.split(',').map(s => s.trim()).filter(Boolean)
  }
  if (f.type === 'number') {
    const n = Number(trimmed)
    return Number.isFinite(n) ? n : trimmed
  }
  return trimmed
}
