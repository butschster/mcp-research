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
          '`window.researchData` plus a `research-data` event.\n\n' +
          'An expand control sits in the frame\'s top-right corner at rest — recessive at ' +
          '`opacity: 0.55`, full on hover or keyboard focus, but never absent: the iframe ' +
          'swallows every pointer event over the document, so a control that only appears ' +
          'on hover of the content is a control nobody finds, and a touch screen has no ' +
          'hover at all. Enlarged, the frame grows a title bar carrying the `title` prop ' +
          'and the way back, because the caption that named the artifact on the page is ' +
          'left behind.\n\n' +
          'Enlarging happens two ways — the real Fullscreen API, and a `position: fixed` ' +
          'overlay for a browser that will not give a `div` the screen. Both are internal ' +
          'state with no prop behind them, so the stories below press the button.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof ArtifactFrame>

const chart = `<!doctype html>
<html><head><meta charset="utf-8"><title>Throughput by model</title>
<style>
  body { margin:0; padding:20px; font-family:system-ui,sans-serif; background:#14241d; color:#f6f3ec; }
  h1 { font-size:16px; margin:0 0 16px; }
  .row { display:grid; grid-template-columns:110px 1fr 50px; gap:10px; align-items:center;
         margin-bottom:10px; font-size:13px; }
  .track { height:20px; background:rgba(148,163,184,.1); border-radius:4px; overflow:hidden; }
  .fill { height:100%; background:#c7d9a7; }
  .num { text-align:right; color:#a7b6a6; }
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
<style>body{margin:0;padding:20px;font-family:system-ui,sans-serif;background:#14241d;color:#f6f3ec}
code{color:#c7d9a7}</style></head>
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
<style>body{margin:0;padding:20px;font-family:system-ui,sans-serif;background:#14241d;color:#f6f3ec}
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

// The frame is the document's viewport, so a document sized in vh reports more every
// time the host grows it. The host stops applying increases after 20 in a row, so
// this settles instead of growing without end.
const viewportSized = `<!doctype html>
<html><head><meta charset="utf-8"><title>Sized in vh</title>
<style>
  html,body { margin:0; }
  body { min-height:100vh; padding:20px; font-family:system-ui,sans-serif;
         background:#14241d; color:#f6f3ec; }
</style></head>
<body><h1 style="font-size:16px;margin:0">min-height: 100vh</h1>
<p style="color:#a7b6a6;font-size:13px">Grows the frame on every report until the host freezes it.</p>
</body></html>`

export const ViewportSizedDoesNotRunAway: Story = {
  args: { html: viewportSized, title: 'Sized in vh' },
}

/* ---------------------------------------------------------------------------
   Enlarging
   ---------------------------------------------------------------------------

   Neither enlarged state has a prop behind it — `fullscreen` and `overlay` are
   internal refs — so these stories press the control, which is also the only
   honest way to document a control.

   The overlay is `position: fixed; inset: 0`, which on an autodocs page would
   cover the autodocs page. Play functions do not run in docs unless a story
   asks (`docs.story.autoplay`), and none of these asks: open them in the canvas
   to see the enlarged state, or press the button yourself on the docs page. */

/**
 * Click expand, having first taken the real Fullscreen API off the element.
 *
 * Which of the two states the component lands in is not the story's to choose:
 * `requestFullscreen` needs a user gesture and a scripted click is not one, so
 * the request is refused and the component falls back to the overlay anyway.
 * Removing the method makes that the deterministic path — the `if (!request)`
 * branch — instead of a race against whichever browser is running the
 * catalogue, and it reproduces the case the fallback was written for: iOS
 * Safari, which grants fullscreen to video and to nothing else.
 *
 * The two states differ only in how they are entered. `.artifact-frame:fullscreen`
 * and `.artifact-frame.is-overlay` are one CSS rule, so what you see here is
 * what the Fullscreen API path looks like too.
 */
async function enlarge(canvasElement: HTMLElement) {
  const frame = canvasElement.querySelector('.artifact-frame')
  if (frame) {
    for (const method of ['requestFullscreen', 'webkitRequestFullscreen']) {
      Object.defineProperty(frame, method, { value: undefined, configurable: true })
    }
  }
  const expand = canvasElement.querySelector('.artifact-expand') as HTMLElement | null
  expand?.click()
  // Let Vue drop the inline height and paint the bar before anything is judged.
  await new Promise((resolve) => setTimeout(resolve, 0))
}

/**
 * Enlarged: the title bar appears above the frame with the artifact's name on
 * the left and the way back on the right, the expand control is gone (it is the
 * collapse button's job now), and the inline `height` is dropped so the frame
 * takes the screen from CSS rather than keeping the size it had on the page.
 *
 * The bar is `no-print`, and a page printed while an artifact happens to be
 * enlarged prints the whole page: the overlay's fixed positioning is undone in
 * the print stylesheet.
 */
export const Enlarged: Story = {
  args: { html: chart, title: 'Throughput by model' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await enlarge(canvasElement)
  },
}

/**
 * No title and no caption: the bar falls back to the word "Artifact", which is
 * also what the frame's accessible name becomes.
 *
 * Reachable only from a direct caller. `BlockRenderer` substitutes the same word
 * one level up (`:title="b.data.title || 'Artifact'"`), so a block document
 * never arrives here with an empty title — the fallback is the second of two.
 */
export const EnlargedWithoutATitle: Story = {
  args: { html: plain },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await enlarge(canvasElement)
  },
}

/**
 * A title longer than the bar. It truncates with an ellipsis on one line rather
 * than wrapping, so the collapse button keeps its place: the way out of
 * fullscreen must not move because somebody named their chart carefully.
 */
export const EnlargedWithALongTitle: Story = {
  args: {
    html: chart,
    title:
      'Пропускная способность локальных моделей на одной RTX 4090 — токенов в секунду, ' +
      'с учётом длины контекста и квантизации',
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await enlarge(canvasElement)
  },
}

// Far taller than any screen, and every row numbered so it is obvious which end
// you are looking at after a scroll.
const veryTall = `<!doctype html>
<html><head><meta charset="utf-8"><title>Migration checklist</title>
<style>
  body { margin:0; padding:20px; font-family:system-ui,sans-serif; background:#14241d; color:#f6f3ec; }
  h1 { font-size:16px; margin:0 0 12px; position:sticky; top:0; background:#14241d; padding:8px 0; }
  div { padding:7px 0; border-bottom:1px solid rgba(148,163,184,.14); font-size:13px; }
  span { color:#a7b6a6; font-variant-numeric:tabular-nums; margin-right:10px; }
</style></head>
<body><h1>Migration checklist — 200 steps</h1><script>
  document.write(Array.from({length: 200}, function (_, i) {
    return '<div><span>' + String(i + 1).padStart(3, '0') + '</span>Step ' + (i + 1) + '</div>';
  }).join(''));
<\/script></body></html>`

/**
 * A document far taller than the screen, enlarged.
 *
 * The frame must scroll internally rather than grow: `flex: 1; min-height: 0`
 * on the iframe is what holds it to the viewport, and the `min-height: 0` is
 * load-bearing — a flex item's default `min-height: auto` refuses to shrink
 * below its content, which for an iframe reporting 6000px means the collapse
 * button leaves the screen.
 *
 * Nothing is measured while enlarged, so the frame comes back to the page at the
 * height it left, then re-measures on the artifact's next report.
 */
export const TallDocumentScrollsWhenEnlarged: Story = {
  args: { html: veryTall, title: 'Migration checklist' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await enlarge(canvasElement)
  },
}

/**
 * The `min-height: 100vh` document, enlarged — the case the height guard in
 * `onMessage` exists for.
 *
 * Enlarged, the frame's viewport is the screen, so this artifact measures the
 * screen and reports it. Recording that would do two wrong things at once: the
 * frame would return to the page as tall as a monitor, and the runaway-growth
 * counter would burn through its twenty allowed increases on a resize that was
 * never a loop. So reports are dropped for as long as the frame is enlarged.
 *
 * Collapse it and the frame is the height it was before, not the screen's.
 */
export const ViewportSizedEnlarged: Story = {
  args: { html: viewportSized, title: 'Sized in vh' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await enlarge(canvasElement)
  },
}

// A clock and a click counter: both reset to zero if the document reloads, and
// the session id changes, so "did this survive?" is answerable at a glance.
const stateful = `<!doctype html>
<html><head><meta charset="utf-8"><title>Sampling playground</title>
<style>
  body { margin:0; padding:20px; font-family:system-ui,sans-serif; background:#14241d; color:#f6f3ec; }
  h1 { font-size:16px; margin:0 0 14px; }
  p { font-size:13px; margin:0 0 8px; color:#a7b6a6; }
  b { color:#c7d9a7; font-variant-numeric:tabular-nums; }
  button { font:inherit; font-size:13px; padding:6px 12px; border-radius:6px; cursor:pointer;
           color:#f6f3ec; background:rgba(148,163,184,.12); border:1px solid rgba(148,163,184,.28); }
</style></head>
<body>
<h1>Temperature sweep</h1>
<p>Document session <b id="sid">—</b>, alive for <b id="up">0.0</b>s</p>
<button id="go">Sampled <span id="n">0</span> times</button>
<p>Every number here lives in the frame. A reload would take all three back to their starting values.</p>
<script>
  document.getElementById('sid').textContent = Math.random().toString(36).slice(2, 8);
  var n = 0;
  document.getElementById('go').addEventListener('click', function () {
    n += 1;
    document.getElementById('n').textContent = n;
  });
  var t0 = Date.now();
  setInterval(function () {
    document.getElementById('up').textContent = ((Date.now() - t0) / 1000).toFixed(1);
  }, 100);
<\/script>
</body></html>`

/**
 * An artifact with state of its own, enlarged after it has been running for a
 * moment: the session id, the clock and the click count all carry across.
 *
 * That is the whole reason the overlay fixes the wrapper where it stands instead
 * of moving it to `body`, which is what the diagram viewer does. A diagram is an
 * SVG and survives the move; re-parenting an iframe reloads its document, so an
 * artifact somebody had filtered, sorted or scrolled would snap back to its
 * initial state at the exact moment they asked for a better look at it.
 *
 * Click the button a few times before pressing expand to see the count survive
 * too — the play function only waits for the clock.
 */
export const InteractiveStateSurvivesEnlarging: Story = {
  args: { html: stateful, title: 'Temperature sweep' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    // Long enough that a reset clock would be unmistakable.
    await new Promise((resolve) => setTimeout(resolve, 900))
    await enlarge(canvasElement)
  },
}

/**
 * The expand control at rest, over a light document.
 *
 * It is drawn on translucent black with a blur behind it rather than in the
 * page's own surface colour, because it floats over a document this component
 * did not write and cannot predict — white here, dark in every other story on
 * this page. Hover the frame, or tab to the button, to take it from `0.5` to
 * full.
 */
export const ExpandControlOverALightArtifact: Story = {
  args: { html: plain, title: 'Plain artifact' },
}

/**
 * The artifact running to the edges of the card that holds it.
 *
 * A document that is nothing but one artifact gets `.is-artifact` on its
 * `.entry-content` card, and the card gives up its padding entirely — the entry
 * page and the shared page both do this. The card keeps its border and its
 * rounding and clips the frame to them with `overflow: hidden`, which is why
 * the frame drops its own border and radius here rather than trying to match a
 * value it would have to carry a copy of.
 *
 * Only the frame goes flush. The html block's title and caption take the
 * padding back through `--entry-pad`, so they stay where the prose would be.
 *
 * Inert in every other story on this page: the rule is keyed on
 * `.entry-content.is-artifact` and Storybook has no such ancestor, by design.
 * This story supplies the real one, unmodified — faking it would document a
 * layout that exists nowhere.
 */
export const InsideTheEntryCard: Story = {
  args: { html: chart, title: 'Throughput by model' },
  parameters: { layout: 'padded' },
  render: (args: any) => ({
    components: { ArtifactFrame },
    setup: () => ({ args }),
    template: `
      <div class="entry-content card is-artifact">
        <figure class="b-figure" style="margin: 0;">
          <figcaption class="b-html-title">Throughput by model</figcaption>
          <ArtifactFrame v-bind="args" />
          <figcaption>Measured on a single RTX 4090 at 4-bit, 8k context.</figcaption>
        </figure>
      </div>
    `,
  }),
}
