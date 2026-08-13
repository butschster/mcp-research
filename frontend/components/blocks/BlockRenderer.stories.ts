import type { Meta, StoryObj } from '@storybook/vue3'
import BlockRenderer from './BlockRenderer.vue'

const meta: Meta<typeof BlockRenderer> = {
  title: 'Blocks/BlockRenderer',
  component: BlockRenderer,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
  parameters: {
    docs: {
      description: {
        component:
          'Renders an `entry_type: blocks` document. Every block sits in the reading ' +
          'column; text fields carry a restricted markdown subset and `[[E3]]` ' +
          'references, escaped before any markup is introduced. The `html` block ' +
          'delegates to ArtifactFrame, which isolates the document in a sandboxed ' +
          'iframe sized to its own height.',
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
  body { margin:0; padding:18px; font-family:system-ui,sans-serif; background:#0c1220; color:#e2e8f0; }
  .row { display:grid; grid-template-columns:110px 1fr 46px; gap:10px; align-items:center; margin-bottom:8px; font-size:13px; }
  .track { height:18px; background:rgba(148,163,184,.1); border-radius:4px; overflow:hidden; }
  .fill { height:100%; background:#6cc5e0; }
  .num { text-align:right; color:#7f8ea3; }
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

// A code block whose language is mermaid is drawn, not printed: pan, zoom,
// fullscreen, and a link that reopens it in the live editor.
export const MermaidCodeBlock: Story = {
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
