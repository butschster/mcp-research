import type { Meta, StoryObj } from '@storybook/vue3'
import ImportNoteGroup from './ImportNoteGroup.vue'
import { mockManyNoteItems, mockNoteItems } from '../../__mocks__/import'

/**
 * One collapsible row of the import ledger.
 *
 * Two tones and nothing in between. `attention` is something the person has to
 * decide about — a value the section refused, a required field the file did not
 * answer — and it opens itself, carries the warning colour and a filled glyph.
 * `note` is something to know: a key we read and did not use, a reference that
 * resolves to nothing. It arrives collapsed and stays grey.
 *
 * It deliberately does not remember its open state. `entry/Foldable.vue` does,
 * which is right for a panel that stands under an entry and wrong here: a group
 * collapsed for the last file must not arrive collapsed for the file where it
 * is the only warning there is.
 *
 * The root element is an `<li>`, so every story wraps it in a `<ul>` — dropping
 * it into a `<div>` renders, but the parent's list styling is what gives the
 * rows their separators.
 */
const meta: Meta<typeof ImportNoteGroup> = {
  title: 'Research/ImportNoteGroup',
  component: ImportNoteGroup,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template:
        '<ul style="max-width: 640px; margin: 0; padding: 0; list-style: none;"><story /></ul>',
    }),
  ],
  argTypes: {
    tone: { control: 'inline-radio', options: ['attention', 'note'] },
    label: { control: 'text' },
    items: { control: 'object' },
    defaultOpen: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof ImportNoteGroup>

/** The quiet tone, closed — how a `note` group arrives. The count is the whole
 *  summary: two keys, and opening it is optional. */
export const Note: Story = {
  args: {
    tone: 'note',
    label: '2 keys this section does not declare',
    items: mockNoteItems,
  },
}

/** The same group opened, showing the key/reason grid. Each row is a `<dt>` of
 *  monospaced key and an optional uppercase kind, against a `<dd>` that quotes
 *  the offending value before explaining itself. */
export const NoteOpen: Story = {
  args: { ...Note.args, defaultOpen: true } as Story['args'],
}

/** The loud tone: warning-coloured head, a filled `!` glyph, and open on
 *  arrival — the dialog passes `default-open` for exactly this tone, because a
 *  refusal nobody expanded is a refusal nobody read. */
export const Attention: Story = {
  args: {
    tone: 'attention',
    label: '3 values this section could not take',
    defaultOpen: true,
    items: [
      { kind: 'rejected', key: 'confidence', value: 'fairly high', reason: 'This field takes one of: low, medium, high.' },
      { kind: 'missing', key: 'owner', reason: 'Required by this section, and the file does not answer it.' },
      {
        kind: 'replaced',
        key: 'status',
        value: 'needs-review',
        reason: 'This research has no status by that name, so the entry was created as in_progress.',
      },
    ],
  },
}

/** Attention, collapsed. Only reachable by clicking the head shut, but worth
 *  staging: the warning colour and the glyph have to survive the collapse, or
 *  the row stops being distinguishable from a note the moment it is dismissed. */
export const AttentionCollapsed: Story = {
  args: { ...Attention.args, defaultOpen: false } as Story['args'],
}

/** One item. The label is singular — that is the caller's job, not the
 *  component's — and the count badge still shows, because a badge that vanishes
 *  at one makes the row jump when a second item appears. */
export const SingleItem: Story = {
  args: {
    tone: 'attention',
    label: '1 value this section could not take',
    defaultOpen: true,
    items: [
      { kind: 'rejected', key: 'reviewed_on', value: '3rd week of Q3', reason: 'Not a date this section can read. Expected YYYY-MM-DD.' },
    ],
  },
}

/** Twenty items — an Obsidian vault export, where every note carries the same
 *  seven housekeeping keys. The body has no scroller of its own: it grows, and
 *  the dialog around it is what scrolls. That is why this group ships closed. */
export const ManyItems: Story = {
  args: {
    tone: 'note',
    label: '20 keys this section does not declare',
    defaultOpen: true,
    items: mockManyNoteItems,
  },
}

/** A 120-character reason against a short key. The grid's first column is
 *  `minmax(0, auto)`, so the key column stays as narrow as its content and the
 *  reason takes the rest — the wrap happens in the reason, never in the key. */
export const LongReason: Story = {
  args: {
    tone: 'attention',
    label: '1 value this section could not take',
    defaultOpen: true,
    items: [
      {
        kind: 'rejected',
        key: 'status',
        value: 'needs-review',
        reason:
          'This research declares no status by that name, so the document was created as a draft and can be moved by hand afterwards.',
      },
    ],
  },
}

/** Cyrillic in both the key and the quoted value. Keys render in JetBrains
 *  Mono, which has no Cyrillic coverage, so these fall through to the system
 *  monospace — a legitimate difference in shape, not a broken glyph. */
export const CyrillicContent: Story = {
  args: {
    tone: 'note',
    label: '3 ключа прочитано, но не использовано как есть',
    defaultOpen: true,
    items: [
      { kind: 'заменено', key: 'рецензент', value: 'Пётр Бутенко', reason: 'Сохранено в документе, но не как поле секции.' },
      { key: 'источник', value: 'счета поставщиков за третий квартал', reason: 'Записано в метаданные без изменений.' },
      { key: 'уверенность', value: 'довольно высокая, но по Halyard мы видели только публичный прайс', reason: 'Поле принимает одно из: low, medium, high.' },
    ],
  },
}

/** At 768px and below the two-column grid becomes one: the key sits above its
 *  reason. Two columns of thirty characters each is not a table, it is two
 *  ragged strips. Switch the viewport to Mobile to see the other side. */
export const NarrowLayout: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
  args: {
    tone: 'attention',
    label: '3 values this section could not take',
    defaultOpen: true,
    items: Attention.args!.items,
  },
}

/** Both tones stacked, which is what the ledger actually renders. The
 *  `.note-group + .note-group` border is only visible here — a single group in
 *  isolation never draws it, so a change to it can only be reviewed on this
 *  story. */
export const BothTonesStacked: Story = {
  render: () => ({
    components: { ImportNoteGroup },
    setup: () => ({
      many: mockManyNoteItems.slice(0, 4),
      attention: Attention.args!.items,
      refs: [
        { key: '[[E44]]', reason: 'Nothing in this research answers to it. It appears 2 times, and is kept exactly as written.' },
        { key: '[[R9:E2]]', reason: 'Nothing in this research answers to it. It is kept exactly as written.' },
      ],
    }),
    template: `
      <ImportNoteGroup tone="attention" label="3 values this section could not take" :items="attention" :default-open="true" />
      <ImportNoteGroup tone="note" label="4 keys this section does not declare" :items="many" />
      <ImportNoteGroup tone="note" label="2 references do not resolve here" :items="refs" />
    `,
  }),
}
