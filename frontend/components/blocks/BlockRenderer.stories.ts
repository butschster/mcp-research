import type { Meta, StoryObj } from '@storybook/vue3'
import BlockRenderer from './BlockRenderer.vue'
import { withoutShare } from '../../__mocks__/share'

const meta: Meta<typeof BlockRenderer> = {
  title: 'Blocks/BlockRenderer',
  component: BlockRenderer,
  tags: ['autodocs'],
  // Share state is module state, so a story viewed after any `withShare()` one
  // inherits it: a task code chip would link into /s/{token}/ and [[E4]] would
  // render inert, both silently. This gives every story here a known start.
  decorators: [withoutShare()],
  argTypes: {
    blocks: { control: 'object' },
    researchSlug: { control: 'text' },
    entryId: { control: 'text' },
    readonly: { control: 'boolean' },
    marksMode: { control: 'inline-radio', options: ['all', 'open', 'off'] },
    // The page owns the task list and how its fetch went; the renderer only
    // passes both to the `task_ref` branch, which is why a story can draw any
    // state of that block without a server.
    tasks: { control: 'object' },
    tasksStatus: { control: 'select', options: ['idle', 'loading', 'ready', 'error', 'excluded'] },
    expandBlockIds: { control: 'object' },
    onRetryTasks: { action: 'retry-tasks' },
  },
  parameters: {
    docs: {
      description: {
        component:
          'Renders an `entry_type: blocks` document. Every block sits in the reading ' +
          'column; text fields carry a restricted markdown subset and `[[E3]]` ' +
          'references, escaped before any markup is introduced. The `html` block ' +
          'delegates to ArtifactFrame, which isolates the document in a sandboxed ' +
          'iframe sized to its own height. Two block types are components rather ' +
          'than branches, because they hold state of their own: `task_ref`, which ' +
          'writes a tick through to a real task, and `transcript`, which folds. The ' +
          'renderer also assigns the speaker colours, once across the whole document, ' +
          'so one person keeps one colour in every transcript in an entry.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof BlockRenderer>

const article = [
  { type: 'heading', data: { level: 2, text: 'Local model selection' } },
  {
    type: 'paragraph',
    data: {
      text: 'Throughput was measured on one card. See [[E3]] for the **method** and `ollama ps` output, plus the [upstream notes](https://example.com/notes).',
    },
  },
  {
    type: 'callout',
    data: { variant: 'warning', title: 'Numbers drift', text: 'Re-measure after a driver update.' },
  },
  { type: 'heading', data: { level: 3, text: 'Results' } },
  {
    type: 'table',
    data: {
      header: true,
      rows: [
        ['Model', 'tok/s', 'VRAM'],
        ['Llama 3.1 8B', '96', '6.2 GB'],
        ['Mistral 12B Q5', '61', '10.8 GB'],
        ['Qwen2.5 32B Q4', '28', '19.4 GB'],
      ],
    },
  },
  { type: 'list', data: { style: 'ordered', items: ['Pick a quantization', 'Measure', 'Compare against [[E4]]'] } },
  { type: 'quote', data: { text: 'Measure twice, quantize once.', cite: 'folk wisdom' } },
  { type: 'code', data: { language: 'bash', code: 'ollama run llama3.1:8b --verbose\n# eval rate: 96 tokens/s' } },
  { type: 'divider', data: {} },
  {
    type: 'paragraph',
    data: { text: 'A closing thought with *emphasis* and a relative link to [the roadmap](/research/R1).' },
  },
]

export const Article: Story = {
  args: { blocks: article, researchSlug: 'R1' },
}

export const EveryTextBlock: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      { type: 'heading', data: { level: 2, text: 'H2' } },
      { type: 'heading', data: { level: 3, text: 'H3' } },
      { type: 'heading', data: { level: 4, text: 'H4' } },
      { type: 'paragraph', data: { text: 'Paragraph with **bold**, *italic*, `code`, [[E1]].' } },
      { type: 'list', data: { style: 'unordered', items: ['unordered one', 'unordered two'] } },
      { type: 'list', data: { style: 'ordered', items: ['ordered one', 'ordered two'] } },
      { type: 'quote', data: { text: 'Quoted line.', cite: 'someone' } },
      { type: 'code', data: { language: 'go', code: 'fmt.Println("verbatim C:\\notes")' } },
      { type: 'divider', data: {} },
    ],
  },
}

export const AllCalloutVariants: Story = {
  args: {
    blocks: ['info', 'warning', 'success', 'danger'].map((variant) => ({
      type: 'callout',
      data: { variant, title: variant, text: `A ${variant} callout with [[E2]] inside.` },
    })),
    researchSlug: 'R1',
  },
}

const chartHtml = `<!doctype html>
<html><head><meta charset="utf-8"><title>Throughput</title>
<style>
  body { margin:0; padding:18px; font-family:system-ui,sans-serif; background:#14241d; color:#f6f3ec; }
  .row { display:grid; grid-template-columns:110px 1fr 46px; gap:10px; align-items:center; margin-bottom:8px; font-size:13px; }
  .track { height:18px; background:rgba(148,163,184,.1); border-radius:4px; overflow:hidden; }
  .fill { height:100%; background:#c7d9a7; }
  .num { text-align:right; color:#a7b6a6; }
</style></head>
<body><div id="out"></div>
<script>
  const d = [['Llama 8B',96],['Mistral 12B',61],['Qwen 32B',28]];
  const max = Math.max(...d.map(x => x[1]));
  document.getElementById('out').innerHTML = d.map(([n,v]) => \`
    <div class="row"><span>\${n}</span>
      <div class="track"><div class="fill" style="width:\${v/max*100}%"></div></div>
      <span class="num">\${v}</span></div>\`).join('');
<\/script>
</body></html>`

export const WithHtmlBlock: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      { type: 'paragraph', data: { text: 'Prose above the visual — impossible when the whole entry was one HTML page.' } },
      { type: 'html', data: { html: chartHtml, title: 'Throughput by model', caption: 'Rendered inside a sandboxed frame.' } },
      { type: 'paragraph', data: { text: 'And prose below it.' } },
    ],
  },
}

// The dedicated block type: a diagram is a figure with a caption, drawn with pan,
// zoom, fullscreen and a link back to the live editor.
export const MermaidBlock: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      { type: 'paragraph', data: { text: 'A diagram sits in the reading column like any other block.' } },
      {
        type: 'mermaid',
        data: {
          code: 'flowchart TD\n    A[entry_create] --> B{entry_type}\n    B -->|markdown| C[normalizeContent]\n    B -->|blocks| D[NormalizeBlockDocument]\n    B -->|artifact| E[ArtifactToBlockDocument]',
          caption: 'Where content goes on the way in — see [[E3]].',
        },
      },
    ],
  },
}

// The older spelling still draws, so documents written before the block type
// existed keep working.
export const MermaidAsCodeBlock: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      { type: 'paragraph', data: { text: 'A diagram sits in the reading column like any other block.' } },
      {
        type: 'code',
        data: {
          language: 'mermaid',
          code: 'flowchart TD\n    A[entry_create] --> B{entry_type}\n    B -->|markdown| C[normalizeContent]\n    B -->|blocks| D[NormalizeBlockDocument]\n    B -->|artifact| E[ArtifactToBlockDocument]',
        },
      },
      { type: 'code', data: { language: 'bash', code: 'make build-all   # ordinary code, printed as-is' } },
    ],
  },
}

// A source mermaid cannot parse keeps the source and offers the editor, which
// reports the syntax error.
export const BrokenMermaid: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      { type: 'code', data: { language: 'mermaid', code: 'flowchart TD\n    A --> ((broken' } },
    ],
  },
}

// The one block a reader writes to. Without an entryId the checkboxes render
// their state and do nothing, which is also what a read-only viewer sees.
export const Checklist: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      { type: 'paragraph', data: { text: 'Steps of the migration, ticked as they are done.' } },
      {
        id: 'a1b2c3d4',
        type: 'checklist',
        data: {
          title: 'Release runbook',
          items: [
            { key: 'k1', text: 'Back up the database' },
            { key: 'k2', text: 'Run the migration on a copy — see [[E3]]' },
            { key: 'k3', text: 'Announce the window' },
          ],
          state: { k1: true },
        },
      },
    ],
  },
}

export const ChecklistReadOnly: Story = {
  args: {
    ...Checklist.args,
    readonly: true,
  },
}

// With an entry id the boxes are live. Ticking here posts to that entry, so the
// story exists to show the enabled control, not to be clicked in anger.
export const ChecklistInteractive: Story = {
  args: {
    ...Checklist.args,
    entryId: 'demo-entry-id',
  },
}

// Two checklists in one document: the failure line belongs to the item that
// failed, not to every checklist on the page.
export const ChecklistTwoInOneDocument: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      {
        id: 'aaaa1111',
        type: 'checklist',
        data: { title: 'Prod', items: [{ key: 'k1', text: 'Back up' }], state: { k1: true } },
      },
      {
        id: 'bbbb2222',
        type: 'checklist',
        data: { title: 'Staging', items: [{ key: 'k1', text: 'Back up' }] },
      },
    ],
  },
}

// An empty checklist is dropped by the normalizer, so this can only be reached
// by a renderer bug — the story documents that it degrades quietly.
export const ChecklistEmpty: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [{ id: 'aaaa1111', type: 'checklist', data: { title: 'Nothing to do', items: [] } }],
  },
}

// Long unbreakable text must wrap rather than scroll the page sideways.
export const ChecklistLongToken: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      {
        id: 'aaaa1111',
        type: 'checklist',
        data: {
          items: [
            { key: 'k1', text: 'Revoke a1b2c3d4-e5f6-7890-abcd-ef1234567890-a1b2c3d4-e5f6-7890-abcd-ef1234567890' },
          ],
        },
      },
    ],
  },
}

// Text that tries to inject markup must render as text.
export const HostileText: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      { type: 'paragraph', data: { text: '<img src=x onerror="alert(1)"> and <script>alert(2)<\/script>' } },
      { type: 'paragraph', data: { text: 'A [bad link](javascript:alert(3)) keeps its label but loses the href.' } },
      { type: 'heading', data: { level: 2, text: '<b>not bold</b>' } },
    ],
  },
}

export const EmptyDocument: Story = {
  args: { blocks: [], researchSlug: 'R1' },
}

// A document with the two blocks that reach outside it: `task_ref` shows real
// tasks and writes a tick through to them, `transcript` holds a conversation
// that happened elsewhere. Both are branches here and components underneath —
// a block that is markup is a branch; a block with per-instance state or its
// own network calls is a component.
export const TaskRefAndTranscript: Story = {
  args: {
    researchSlug: 'R1',
    tasksStatus: 'ready',
    tasks: [
      { id: 'id1', code: 'T1', title: 'Close the perimeter', status: 'completed', priority: 'medium' },
      { id: 'id2', code: 'T4', title: 'Put the scanner inside the network', status: 'in_progress', priority: 'medium' },
      { id: 'id3', code: 'T7', title: 'Rotate the gateway certificates', status: 'blocked', priority: 'high' },
    ],
    blocks: [
      { id: 'head0001', type: 'heading', data: { level: 2, text: 'Where we are' } },
      {
        id: 'tr000001',
        type: 'transcript',
        data: {
          title: 'Infrastructure call, 12 August',
          turns: [
            { speaker: 'Peter', text: 'The perimeter is closed, only the gateway is exposed.', ts: '00:03:12' },
            { speaker: 'Anna', text: 'Then the scanner goes inside — see [[E4]].' },
          ],
        },
      },
      {
        id: 'tref0001',
        type: 'task_ref',
        data: { tasks: ['T1', 'T4', 'T7'], note: 'Close before Friday.', show_progress: true },
      },
    ],
  },
}

// Two transcripts in one document. The speaker colours are assigned by order of
// first appearance across the whole document, so the same person keeps a colour
// in both — hashing the name would have collided instead.
export const TwoTranscriptsShareSpeakerColours: Story = {
  args: {
    researchSlug: 'R1',
    blocks: [
      {
        id: 'tr000001',
        type: 'transcript',
        data: {
          title: 'Morning',
          turns: [
            { speaker: 'Anna', text: 'We start with the gateway.' },
            { speaker: 'Peter', text: 'Agreed.' },
          ],
        },
      },
      {
        id: 'tr000002',
        type: 'transcript',
        data: {
          title: 'Afternoon',
          turns: [
            { speaker: 'Peter', text: 'The gateway is done.' },
            { speaker: 'Anna', text: 'Then the scanner.' },
          ],
        },
      },
    ],
  },
}


// `GET /api/researches/{slug}/tasks` did not answer. The block still says what
// the document points at, and Try again re-asks the page — the renderer only
// forwards the emit, because the page owns the fetch.
export const TaskRefLoadFailed: Story = {
  args: {
    researchSlug: 'R1',
    tasksStatus: 'error',
    tasks: [],
    blocks: [
      { id: 'para0001', type: 'paragraph', data: { text: 'The work this leaves open:' } },
      {
        id: 'tref0001',
        type: 'task_ref',
        data: { tasks: ['T1', 'T4', 'T7'], note: 'Close before Friday.', show_progress: true },
      },
    ],
  },
}

// A share link published without tasks: the API 404s by design, so the block
// says so instead of offering a retry that cannot succeed.
export const TaskRefNotShared: Story = {
  args: {
    ...(TaskRefLoadFailed.args as any),
    tasksStatus: 'excluded',
  },
}

// A transcript long enough to fold, held open because the entry page found an
// annotation anchored inside it. A mark in a folded tail measures at zero and
// its pin lands above the card, so the page names the blocks its annotations
// live in and the renderer forces exactly those open. The share page passes an
// empty list — it shows no marks.
export const TranscriptExpandedByAnnotation: Story = {
  args: {
    researchSlug: 'R1',
    expandBlockIds: ['tr000042'],
    blocks: [
      {
        id: 'tr000042',
        type: 'transcript',
        data: {
          title: 'Infrastructure call, 12 August',
          turns: Array.from({ length: 44 }, (_, i) => ({
            speaker: i % 2 === 0 ? 'Peter' : 'Anna',
            text: `Turn ${i + 1}. ${i === 38 ? 'The sentence somebody marked lives down here, past the fold.' : 'Something about the gateway.'}`,
            ts: `00:${String(Math.floor(i / 2)).padStart(2, '0')}:${String((i * 7) % 60).padStart(2, '0')}`,
          })),
        },
      },
    ],
  },
}
