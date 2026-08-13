import type { Meta, StoryObj } from '@storybook/vue3'
import ArtifactFrame from './ArtifactFrame.vue'

const meta: Meta<typeof ArtifactFrame> = {
  title: 'Entry/ArtifactFrame',
  component: ArtifactFrame,
  tags: ['autodocs'],
  argTypes: {
    html: { control: 'text' },
    title: { control: 'text' },
    fallbackHeight: { control: 'number' },
    maxHeight: { control: 'number' },
  },
  parameters: {
    docs: {
      description: {
        component:
          'Renders an `entry_type: artifact` document inside `<iframe sandbox="allow-scripts">`. ' +
          'The sandbox withholds `allow-same-origin`, so the parent cannot measure the ' +
          'document; a height reporter is appended to the HTML instead and posts its size ' +
          'back over postMessage. `bridgeData` is delivered to the artifact after load as ' +
          '`window.researchData` plus a `research-data` event.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof ArtifactFrame>

const chart = `<!doctype html>
<html><head><meta charset="utf-8"><title>Throughput by model</title>
<style>
  body { margin:0; padding:20px; font-family:system-ui,sans-serif; background:#0c1220; color:#e2e8f0; }
  h1 { font-size:16px; margin:0 0 16px; }
  .row { display:grid; grid-template-columns:110px 1fr 50px; gap:10px; align-items:center;
         margin-bottom:10px; font-size:13px; }
  .track { height:20px; background:rgba(148,163,184,.1); border-radius:4px; overflow:hidden; }
  .fill { height:100%; background:#6cc5e0; }
  .num { text-align:right; color:#7f8ea3; }
</style></head>
<body>
<h1>Throughput, tokens/s</h1>
<div id="out"></div>
<script>
  const d = [['Llama 8B',96],['Mistral 12B',61],['Gemma 27B',34],['Qwen 32B',28]];
  const max = Math.max(...d.map(x => x[1]));
  document.getElementById('out').innerHTML = d.map(([n,v]) => \`
    <div class="row"><span>\${n}</span>
      <div class="track"><div class="fill" style="width:\${v/max*100}%"></div></div>
      <span class="num">\${v}</span></div>\`).join('');
<\/script>
</body></html>`

const plain = `<!doctype html>
<html><head><meta charset="utf-8"><title>Plain artifact</title>
<style>body{margin:0;padding:20px;font-family:system-ui,sans-serif;background:#fff;color:#111}</style>
</head><body><h1>No scripts here</h1><p>A static document still reports its height.</p></body></html>`

const usesBridge = `<!doctype html>
<html><head><meta charset="utf-8"><title>Bridge consumer</title>
<style>body{margin:0;padding:20px;font-family:system-ui,sans-serif;background:#0c1220;color:#e2e8f0}
code{color:#6cc5e0}</style></head>
<body>
<h1 style="font-size:16px;margin:0 0 12px">Data from the host</h1>
<pre id="out"><code>waiting for research-data…</code></pre>
<script>
  window.addEventListener('research-data', function (e) {
    document.getElementById('out').innerHTML =
      '<code>' + JSON.stringify(e.detail, null, 2) + '</code>';
  });
<\/script>
</body></html>`

const tall = `<!doctype html>
<html><head><meta charset="utf-8"><title>Very tall artifact</title>
<style>body{margin:0;padding:20px;font-family:system-ui,sans-serif;background:#0c1220;color:#e2e8f0}
div{padding:8px;border-bottom:1px solid rgba(148,163,184,.14)}</style></head>
<body><script>
  document.write(Array.from({length: 120}, (_, i) => '<div>Row ' + (i+1) + '</div>').join(''));
<\/script></body></html>`

export const WithChart: Story = {
  args: { html: chart, title: 'Throughput by model' },
}

export const StaticDocument: Story = {
  args: { html: plain, title: 'Plain artifact' },
}

export const ReceivesHostData: Story = {
  args: {
    html: usesBridge,
    title: 'Bridge consumer',
    bridgeData: {
      research: { code: 'R5', name: 'ITProtect LLM Platform' },
      entries: [
        { code: 'E12', title: 'Model selection criteria' },
        { code: 'E13', title: 'VRAM budget' },
      ],
    },
  },
}

export const ClampedByMaxHeight: Story = {
  args: { html: tall, title: 'Very tall artifact', maxHeight: 400 },
}

export const EmptyDocument: Story = {
  args: { html: '', title: 'Empty artifact', fallbackHeight: 200 },
}
