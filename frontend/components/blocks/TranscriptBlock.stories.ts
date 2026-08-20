import type { Meta, StoryObj } from '@storybook/vue3'
import TranscriptBlock from './TranscriptBlock.vue'
import { withoutShare } from '../../__mocks__/share'

const meta: Meta<typeof TranscriptBlock> = {
  title: 'Blocks/TranscriptBlock',
  component: TranscriptBlock,
  tags: ['autodocs'],
  // Share state is module state, so a story viewed after any `withShare()` one
  // inherits it: a task code chip would link into /s/{token}/ and [[E4]] would
  // render inert, both silently. This gives every story here a known start.
  decorators: [withoutShare()],
  argTypes: {
    blockId: { control: 'text' },
    title: { control: 'text' },
    turns: { control: 'object' },
    speakerHues: { control: 'object' },
    researchSlug: { control: 'text' },
    forceExpanded: { control: 'boolean' },
  },
  parameters: {
    docs: {
      description: {
        component:
          'A conversation held outside this tool — a call, a meeting, an interview — as ' +
          'structured turns. The speaker name sits above an indented run rather than in a ' +
          'column: a fixed column has no honest answer for a long name, and truncating the ' +
          'speaker destroys the one fact a transcript exists to carry. Consecutive turns by ' +
          'the same person are grouped under one name. Turn text carries the inline markdown ' +
          'subset and `[[E3]]` references. Past 36 turns the tail folds behind an expander, ' +
          'in the DOM as `hidden="until-found"` so find-in-page still reaches it.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof TranscriptBlock>

const HUES = { 'Пётр': 1, 'Аня': 2, Peter: 1, Anna: 2, Sam: 3 }

/** Waits for the block to render before a play function acts on it. */
async function waitFor(root: HTMLElement, selector: string): Promise<HTMLElement | null> {
  for (let i = 0; i < 50; i++) {
    const found = root.querySelector(selector) as HTMLElement | null
    if (found) return found
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  return null
}

const call = [
  { speaker: 'Пётр', text: 'Контур закрыт, наружу торчит только шлюз.', ts: '00:03:12' },
  { speaker: 'Пётр', text: 'И сканеру там встать негде.', ts: '00:03:40' },
  { speaker: 'Аня', text: 'Тогда сканер надо ставить внутрь — см. [[E4]].' },
  { speaker: 'Пётр', text: 'Хорошо.', ts: '00:05:02' },
]

export const Default: Story = {
  args: {
    blockId: 'tr000001',
    title: 'Созвон с инфраструктурой, 12 августа',
    turns: call,
    researchSlug: 'R1',
    speakerHues: HUES,
  },
}

/** No title: no header row and no rule. A lone right-aligned count under a rule
 *  would be chrome about nothing. */
export const NoTitle: Story = {
  args: { blockId: 'tr000002', turns: call, researchSlug: 'R1', speakerHues: HUES },
}

/** An unattributed turn continues the group above it. Inventing "Speaker 1"
 *  would assert something the document does not say. */
export const AnonymousTurns: Story = {
  args: {
    blockId: 'tr000003',
    title: 'Recording, speaker labels lost',
    turns: [
      { text: 'The perimeter is closed, only the gateway is exposed.', ts: '00:03:12' },
      { text: 'And the scanner has nowhere to sit.' },
      { speaker: 'Anna', text: 'Then it goes inside.' },
    ],
    researchSlug: 'R1',
    speakerHues: HUES,
  },
}

/** One speaker throughout — an interview answer, a monologue. */
export const SingleSpeaker: Story = {
  args: {
    blockId: 'tr000004',
    title: 'Voice memo',
    turns: [
      { speaker: 'Sam', text: 'Three things went wrong.', ts: '00:00:04' },
      { speaker: 'Sam', text: 'The migration ran on the wrong replica.' },
      { speaker: 'Sam', text: 'Nobody was watching the queue depth.' },
    ],
    researchSlug: 'R1',
    speakerHues: HUES,
  },
}

/** Long turns and a long name — the case a speaker column could not hold. */
export const LongNamesAndLines: Story = {
  args: {
    blockId: 'tr000005',
    title: 'Review board',
    turns: [
      {
        speaker: 'Александра Константиновна',
        text:
          'Если коротко: мы согласны на внутренний сканер, но только после того, как ' +
          'будет описан порядок ротации ключей и кто именно его исполняет в выходные.',
        ts: '01:12:40',
      },
      { speaker: 'Sam', text: 'Fine.' },
      {
        speaker: 'Александра Константиновна',
        text: 'И это должно быть записано\nв двух местах, а не в одном.',
      },
    ],
    researchSlug: 'R1',
    speakerHues: { 'Александра Константиновна': 5, Sam: 3 },
  },
}

/** 60 turns: the fold appears, and the button names the total. */
export const Folded: Story = {
  args: {
    blockId: 'tr000006',
    title: 'Long call',
    turns: Array.from({ length: 60 }, (_, i) => ({
      speaker: i % 2 === 0 ? 'Peter' : 'Anna',
      text: `Turn number ${i + 1}, saying something about the gateway.`,
      ts: `00:${String(Math.floor(i / 2)).padStart(2, '0')}:${String((i * 7) % 60).padStart(2, '0')}`,
    })),
    researchSlug: 'R1',
    speakerHues: HUES,
  },
}

/**
 * The same 60 turns after the reader pressed the button.
 *
 * Worth noticing what this is *not*: `forceExpanded` below drops the fold
 * altogether, so it never draws a toggle. This is the only story in which the
 * button says `Show fewer turns`, and the only one where the tail is revealed
 * rather than never folded. On expanding, focus moves to the first newly
 * revealed group — the button has just been pushed thousands of pixels down by
 * the content that appeared above it, and leaving focus there strands a
 * keyboard reader off-screen.
 */
export const Expanded: Story = {
  args: { ...(Folded.args as any), blockId: 'tr00000a' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const button = await waitFor(canvasElement, '.b-tr-more')
    button?.click()
  },
}

/** The same 60 turns, kept open with no toggle at all — what the entry page
 *  does for a transcript that holds an annotation (a mark inside a folded tail
 *  measures at zero and its pin lands above the card), and what print always
 *  does. */
export const ForceExpanded: Story = {
  args: { ...(Folded.args as any), blockId: 'tr000007', forceExpanded: true },
}

/** 36 turns — under the threshold, so nothing folds. The hysteresis exists so a
 *  32-turn transcript does not hide two turns behind a button. */
export const JustUnderTheFold: Story = {
  args: {
    blockId: 'tr000008',
    title: 'Thirty-six turns',
    turns: Array.from({ length: 36 }, (_, i) => ({
      speaker: i % 2 === 0 ? 'Peter' : 'Anna',
      text: `Turn number ${i + 1}.`,
    })),
    researchSlug: 'R1',
    speakerHues: HUES,
  },
}

/** No turns: the component renders nothing at all. The server drops such a
 *  block, so this only arrives from a hand-written payload. */
export const Empty: Story = {
  args: { blockId: 'tr000009', title: 'Nothing was said', turns: [], researchSlug: 'R1' },
}


/**
 * The whole palette, plus a speaker the map does not know.
 *
 * Hues are assigned by order of first appearance and wrap past six, so the
 * seventh speaker shares the first one's colour — the honest failure mode, and
 * the reason the name is always written out beside it. Colour is never the only
 * carrier here. The eighth speaker is absent from the map entirely and falls
 * back to the ordinary text colour: quiet rather than invisible.
 *
 * Not hashed, which is what tags do: `tagHue` is a sum of char codes mod six,
 * and two speakers in one call collide about one time in six. By the time that
 * happens the reader has learned to trust the hue, which makes a collision
 * worse than no colour at all.
 */
export const EverySpeakerHue: Story = {
  args: {
    blockId: 'tr00000b',
    title: 'Architecture review, all hands',
    turns: [
      { speaker: 'Anna', text: 'We start with the gateway.', ts: '00:00:10' },
      { speaker: 'Peter', text: 'The perimeter is closed.' },
      { speaker: 'Sam', text: 'The scanner has nowhere to sit — see [[E4]].' },
      { speaker: 'Мария', text: 'Тогда ставим внутрь.' },
      { speaker: 'Tomas', text: 'Who rotates the keys at the weekend?' },
      { speaker: 'Yuki', text: 'The on-call does, once the runbook exists.' },
      { speaker: 'Дмитрий', text: 'Seventh speaker — the palette wraps here.' },
      { speaker: 'A guest nobody logged', text: 'And I am not in the map at all.' },
    ],
    researchSlug: 'R1',
    speakerHues: { Anna: 1, Peter: 2, Sam: 3, 'Мария': 4, Tomas: 5, Yuki: 6, 'Дмитрий': 1 },
  },
}

/**
 * Five hundred turns — the cap the normalizer enforces.
 *
 * The fold holds it to thirty on screen, and the tail is in the DOM the whole
 * time as `hidden="until-found"`, so find-in-page still reaches turn 400 and
 * the component follows the browser's reveal rather than hiding the match
 * again.
 *
 * Known limitation, visible here: the fold counts turns, not height. Thirty
 * turns of five thousand characters each would still be a wall.
 */
export const AtTheTurnCap: Story = {
  args: {
    blockId: 'tr00000c',
    title: 'Full-day workshop',
    turns: Array.from({ length: 500 }, (_, i) => ({
      speaker: ['Anna', 'Peter', 'Sam'][i % 3],
      text: `Turn ${i + 1}. ${i % 5 === 0 ? 'A longer contribution about the gateway, the scanner and who is on call for the weekend rotation.' : 'Agreed.'}`,
      ts: `0${Math.floor(i / 120)}:${String(Math.floor(i / 2) % 60).padStart(2, '0')}:${String((i * 7) % 60).padStart(2, '0')}`,
    })),
    researchSlug: 'R1',
    speakerHues: { Anna: 1, Peter: 2, Sam: 3 },
  },
}

/**
 * A single unbroken string in a turn, and a URL nobody wrapped.
 *
 * A speaker name is clamped at 120 characters server-side; the text of a turn
 * is not, and `overflow-wrap: anywhere` is what keeps a pasted token from
 * scrolling the whole document sideways.
 */
export const UnbreakableStrings: Story = {
  args: {
    blockId: 'tr00000d',
    title: 'Pasted from the terminal',
    turns: [
      {
        speaker: 'Sam',
        text: 'The token is a1b2c3d4-e5f6-7890-abcd-ef1234567890-a1b2c3d4-e5f6-7890-abcd-ef1234567890',
        ts: '00:41:02',
      },
      {
        speaker: 'Anna',
        text: 'And it came from https://example.com/a/very/long/path/that/nobody/would/ever/type/by/hand?with=a&query=string&attached=true',
      },
    ],
    researchSlug: 'R1',
    speakerHues: { Anna: 1, Sam: 3 },
  },
}
