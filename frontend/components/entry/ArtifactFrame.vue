<template>
  <div
    ref="wrapRef"
    class="artifact-frame"
    :class="{ 'is-overlay': overlay }"
    :role="enlarged ? 'dialog' : undefined"
    :aria-modal="overlay ? 'true' : undefined"
    :aria-label="enlarged ? `Artifact${title ? ': ' + title : ''}` : undefined"
  >
    <!-- Enlarged, the artifact leaves behind whatever named it on the page — the
         html block's own caption sits outside this component — so the name comes
         with it, next to the only way back. -->
    <div v-if="enlarged" class="artifact-bar no-print">
      <p class="artifact-bar-title">
        <!-- The word "Artifact" is ours and fixed; everything after it is the
             author's. For an `entry_type: artifact` the title is lifted from
             the document's own <title>, so a bar that were nothing but the
             title would render "Sign in — session expired" in our surface
             colour, at our type scale, as the only host chrome on a screen the
             author otherwise owns. The label is what the reader can trust. -->
        <span class="artifact-bar-kind">Artifact</span>
        <span v-if="title" class="artifact-bar-name" :title="title">{{ title }}</span>
      </p>
      <button
        ref="collapseButton"
        type="button"
        class="artifact-btn"
        :title="exitLabel"
        :aria-label="exitLabel"
        @click="toggle"
      >
        <svg aria-hidden="true" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 4v5H4M20 9h-5V4M15 20v-5h5M4 15h5v5"/></svg>
      </button>
    </div>

    <iframe
      ref="frameRef"
      class="artifact-iframe"
      :title="title || 'Artifact'"
      sandbox="allow-scripts"
      referrerpolicy="no-referrer"
      :srcdoc="documentHtml"
      :style="enlarged ? undefined : { height: frameHeight + 'px' }"
      @load="onLoad"
    ></iframe>

    <!-- Recessive rather than hidden: an iframe swallows every pointer event
         inside it, so a control that only appears on hover of the content is a
         control nobody finds, and on a touch screen there is no hover at all. -->
    <button
      v-if="!enlarged"
      ref="expandButton"
      type="button"
      class="artifact-btn artifact-expand no-print"
      title="Fullscreen"
      aria-label="Fullscreen"
      @click="toggle"
    >
      <svg aria-hidden="true" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 9V4h5M15 4h5v5M20 15v5h-5M9 20H4v-5"/></svg>
    </button>

    <p v-if="sizingLooksStuck && !measured && !enlarged" class="artifact-hint no-print" role="status">
      Sizing the artifact… if it stays at this height, the document did not report
      its size and is shown at the fallback height.
    </p>

    <!-- Room for the page to say something about the frame from outside it —
         chiefly that its contents cannot be annotated. Kept outside the figure
         and out of print, so a caption never gains a line the author did not
         write. -->
    <div v-if="!enlarged && $slots.notice" class="artifact-notice no-print">
      <slot name="notice" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineSlots<{
  /** Shown under the frame, never inside it. */
  notice?: () => any
}>()

const props = withDefaults(
  defineProps<{
    /** The artifact's full HTML document. */
    html: string
    /** Accessible name for the frame. */
    title?: string
    /** Data handed to the artifact over postMessage once it loads. */
    bridgeData?: Record<string, unknown> | null
    /** Height used until the document reports its own, and when it never does. */
    fallbackHeight?: number
    /**
     * Upper bound in pixels. Generous by default so any real document still gets its
     * full height and no inner scrollbar; it exists only so an artifact cannot post
     * an absurd height and blow up the page — it knows the channel, so it can post
     * whatever it likes. Zero disables the bound entirely.
     */
    maxHeight?: number
  }>(),
  { title: '', bridgeData: null, fallbackHeight: 520, maxHeight: 60000 }
)

const frameRef = ref<HTMLIFrameElement | null>(null)
const wrapRef = ref<HTMLElement | null>(null)
const expandButton = ref<HTMLButtonElement | null>(null)
const collapseButton = ref<HTMLButtonElement | null>(null)
const frameHeight = ref(props.fallbackHeight)
const measured = ref(false)

/**
 * The sizing notice waits before it appears.
 *
 * It used to render immediately and disappear on the first height report, which
 * for a document that reports in 40ms is a flash — and it is a flex item, so it
 * took 30px of column with it on the way out, or 95px wrapped onto four lines on
 * a phone. Every artifact on the page jumped once, twice on a document with two.
 * A notice about something being slow has no business being the thing that
 * moves the page.
 */
const HINT_AFTER_MS = 1500
const sizingLooksStuck = ref(false)
let hintTimer: ReturnType<typeof setTimeout> | undefined

function armHint() {
  clearTimeout(hintTimer)
  sizingLooksStuck.value = false
  hintTimer = setTimeout(() => {
    sizingLooksStuck.value = !measured.value
  }, HINT_AFTER_MS)
}

// Safari still ships these prefixed.
type FsElement = HTMLElement & { webkitRequestFullscreen?: () => void }
type FsDocument = Document & { webkitFullscreenElement?: Element; webkitExitFullscreen?: () => void }

/** Set when the real Fullscreen API has this frame; `overlay` is the fallback. */
const fullscreen = ref(false)
const overlay = ref(false)
const enlarged = computed(() => fullscreen.value || overlay.value)

function fullscreenElement(): Element | null {
  const d = document as FsDocument
  return document.fullscreenElement ?? d.webkitFullscreenElement ?? null
}

/**
 * The fallback for a browser that will not give a div the screen — iOS Safari
 * grants fullscreen to video and nothing else.
 *
 * It fixes the wrapper where it stands rather than moving it to the body, which
 * is what the diagram viewer does. The diagram is an SVG and survives the move;
 * an iframe does not. Re-parenting an iframe reloads its document, so an
 * artifact somebody had filtered or scrolled would snap back to its initial
 * state at the exact moment they asked for a better look at it.
 *
 * `<Teleport>` and `ModalOverlay.vue` are the obvious simplifications and they
 * are wrong here for the same reason — both move the node.
 *
 * The cost is that a transformed ancestor contains `position: fixed`, and there
 * IS one: `.card:hover` lifts by 2px, and a block document renders inside
 * `.entry-content card` on both the document page and the shared one. So while
 * the pointer is over the card — which it is, having just clicked the button
 * inside it — the card would be the containing block and the overlay would fill
 * the card. Two rules answer that: `.entry-content:hover` keeps its transform
 * at none because a reading surface is not a control, and
 * `body.scroll-locked .card` cancels it on any card for the duration.
 */
function openOverlay() {
  overlay.value = true
  if (wrapRef.value) inertEverythingElse(wrapRef.value)
  handOverFocus()
  // The page behind a fixed overlay still scrolls under it, which reads as the
  // artifact having come loose from the document. The class is shared with
  // ModalOverlay rather than being ours: a second scroll-lock mechanism is how
  // one of them ends up clobbering the other.
  document.body.classList.add('scroll-locked')
}

function closeOverlay() {
  overlay.value = false
  releaseInert()
  document.body.classList.remove('scroll-locked')
  handOverFocus()
}

/**
 * Focus follows the control, because the control is destroyed by its own click.
 *
 * The corner button is `v-if="!enlarged"` and the bar's is a different element,
 * so activating either one removes the thing the keyboard was on and focus
 * falls to `<body>`: the next Tab restarts at the top of the page, behind an
 * artifact filling the screen. The diagram viewer avoids this by keeping one
 * button and swapping its glyph; here the two live in different places, so the
 * move has to be made explicitly.
 */
async function handOverFocus() {
  await nextTick()
  ;(enlarged.value ? collapseButton.value : expandButton.value)?.focus()
}

function toggle() {
  if (fullscreen.value) {
    const d = document as FsDocument
    ;(d.exitFullscreen ?? d.webkitExitFullscreen)?.call(document)
    return
  }
  if (overlay.value) return closeOverlay()

  const el = wrapRef.value as FsElement | null
  if (!el) return
  const request = el.requestFullscreen ?? el.webkitRequestFullscreen
  if (!request) return openOverlay()
  // A rejected promise here is a refusal, not a crash — take the overlay.
  const started = request.call(el) as Promise<void> | undefined
  started?.catch(openOverlay)
}

// fullscreenchange fires for the document, and the element may not be ours: a
// diagram elsewhere on the page going fullscreen must not make this frame think
// it has the screen.
function onFsChange() {
  const mine = !!wrapRef.value && fullscreenElement() === wrapRef.value
  if (mine === fullscreen.value) return
  fullscreen.value = mine
  handOverFocus()
}

/**
 * Escape is promised only on the path that can keep the promise.
 *
 * Real fullscreen is left by the UA, above the document, whatever has focus. The
 * overlay is left by our own window listener — and a key pressed inside a
 * cross-origin sandboxed iframe never crosses the document boundary, so the
 * moment the reader clicks into the artifact, Escape stops working. It comes
 * back as soon as focus returns to the page. Advertising a shortcut that dies on
 * first contact with the content is worse than not advertising it; the collapse
 * button is always there and cannot be painted over.
 */
const exitLabel = computed(() => (overlay.value ? 'Exit fullscreen' : 'Exit fullscreen (Esc)'))

/**
 * Everything outside the overlay is made inert for its duration.
 *
 * Without it the page behind a screen-filling overlay is still in the tab order
 * and still in the accessibility tree — Tab walks a nav nobody can see, and
 * `aria-modal` would be a claim the component does not honour. The usual answer
 * is to move the node out to the body and inert the app root, and this component
 * cannot move: re-parenting an iframe reloads it.
 *
 * So it inerts the siblings instead, at every level from the wrapper up to
 * `body`. The overlay's own ancestors stay reachable, which is what keeps the
 * overlay itself operable, and every other branch of the document goes dark.
 */
const inerted: HTMLElement[] = []

function inertEverythingElse(from: HTMLElement) {
  for (let node: HTMLElement | null = from; node && node !== document.body; node = node.parentElement) {
    for (const sibling of Array.from(node.parentElement?.children ?? [])) {
      if (sibling === node || !(sibling instanceof HTMLElement)) continue
      // Skip what is already inert for its own reasons, or releasing would
      // wake something a modal had put to sleep.
      if (sibling.hasAttribute('inert')) continue
      sibling.setAttribute('inert', '')
      inerted.push(sibling)
    }
  }
}

function releaseInert() {
  for (const el of inerted) el.removeAttribute('inert')
  inerted.length = 0
}

function onKeydown(e: KeyboardEvent) {
  // Real fullscreen leaves on Escape by itself; the overlay has to be told.
  if (e.key === 'Escape' && overlay.value) {
    e.preventDefault()
    closeOverlay()
  }
}

// A unique token per mount: the frame has an opaque origin (allow-scripts without
// allow-same-origin), so messages arrive with origin "null" and cannot be trusted
// by origin alone. Matching this token keeps other frames' messages out.
const channel = `artifact-${Math.random().toString(36).slice(2)}`

/**
 * The sandbox deliberately withholds allow-same-origin, so the parent cannot read
 * the frame's document to measure it. Instead we append a reporter to the document
 * we are about to render: it posts its height on load and on every resize.
 */
// The artifact is the one thing that keeps its own look when the page is printed
// to PDF: it is a document someone authored, not part of our chrome. Browsers
// drop background colours when printing, and the frame is a separate document we
// cannot reach with our stylesheet, so the rule is injected alongside the height
// reporter.
// Every element, not just the root. The property inherits on paper, but what a
// browser drops when printing is decided per box: a heading painted with a
// gradient behind clipped text loses the gradient and prints as nothing at all,
// while the body text beside it is untouched — which is exactly what "some of
// the text changes and the rest does not" looks like.
const printStyle = `
<style id="__print_fix">@media print {
  :root, body, *, *::before, *::after {
    print-color-adjust: exact !important;
    -webkit-print-color-adjust: exact !important;
  }
}</style>`

// The host's page background, duplicated here on purpose: it has to travel INTO
// the sandboxed document, which cannot read our custom properties.
const HOST_BACKGROUND = '#0c1220'


const shim = computed(() => printStyle + `
<script>(function(){
  var CH = ${JSON.stringify(channel)};
  var last = -1;

  var fix = document.getElementById('__print_fix');

  // An artifact that sets no background of its own was written for a dark host
  // and gets one, but only for print: on screen the frame behind it supplies it.
  try {
    var bg = getComputedStyle(document.body).backgroundColor;
    if (!bg || bg === 'transparent' || bg === 'rgba(0, 0, 0, 0)') {
      if (fix) {
        fix.textContent += '@media print { html, body { background: ${HOST_BACKGROUND} !important; } }';
      }
    }
  } catch (e) {}

  // A theme-aware artifact renders dark on screen and light on paper, because
  // printing stops matching prefers-color-scheme. On its own that would merely
  // be a light printout; combined with the dark surface the host frame paints
  // under it, it is dark text on a dark page — the document's own light palette
  // showing through our background.
  //
  // So whatever the dark query is applying right now is re-emitted for print.
  // Copying the rules rather than the resolved colours keeps this honest for
  // everything the query sets, not only custom properties.
  try {
    if (fix && window.matchMedia && matchMedia('(prefers-color-scheme: dark)').matches) {
      var carried = '';
      for (var i = 0; i < document.styleSheets.length; i++) {
        var rules;
        try { rules = document.styleSheets[i].cssRules; } catch (err) { continue; }
        if (!rules) continue;
        for (var j = 0; j < rules.length; j++) {
          var rule = rules[j];
          if (rule.type !== 4) continue;
          var cond = rule.conditionText || (rule.media && rule.media.mediaText) || '';
          if (!/prefers-color-scheme\\s*:\\s*dark/.test(cond)) continue;
          for (var k = 0; k < rule.cssRules.length; k++) carried += rule.cssRules[k].cssText + '\\n';
        }
      }
      if (carried) fix.textContent += '@media print {\\n' + carried + '}';
    }
  } catch (e) {}

  function measure() {
    var d = document.documentElement, b = document.body;
    var vals = [];
    if (d) {
      vals.push(d.scrollHeight, d.offsetHeight);
      if (d.getBoundingClientRect) vals.push(d.getBoundingClientRect().height);
    }
    if (b) {
      vals.push(b.scrollHeight, b.offsetHeight);
      // documentElement.scrollHeight can miss the body's own margins.
      var cs = window.getComputedStyle ? getComputedStyle(b) : null;
      var mt = cs ? parseFloat(cs.marginTop) || 0 : 0;
      var mb = cs ? parseFloat(cs.marginBottom) || 0 : 0;
      if (b.getBoundingClientRect) {
        vals.push(b.getBoundingClientRect().height + mt + mb);
      }
      vals.push(b.scrollHeight + mt + mb);
    }
    // +2px absorbs sub-pixel rounding, which otherwise leaves a scrollbar for
    // the sake of a fraction of a pixel.
    return Math.ceil(Math.max.apply(null, vals.concat([0]))) + 2;
  }

  function report(force) {
    var h = measure();
    if (!force && h === last) return;
    last = h;
    parent.postMessage({ channel: CH, type: 'height', height: h }, '*');
  }

  // Layout settles at different moments depending on what the artifact does:
  // parse, images, fonts, CSS animations. Report at each of them.
  document.addEventListener('DOMContentLoaded', report);
  window.addEventListener('load', function () { report(true); });
  window.addEventListener('resize', function () { report(true); });
  window.addEventListener('transitionend', report);
  window.addEventListener('animationend', report);
  if (document.fonts && document.fonts.ready && document.fonts.ready.then) {
    document.fonts.ready.then(function () { report(true); });
  }
  if (window.ResizeObserver) {
    var ro = new ResizeObserver(function () { report(false); });
    if (document.documentElement) ro.observe(document.documentElement);
    if (document.body) ro.observe(document.body);
  }
  if (window.MutationObserver && document.body) {
    new MutationObserver(function () { report(false); }).observe(document.body, {
      childList: true, subtree: true, characterData: true
    });
  }
  [0, 60, 250, 700, 1500].forEach(function (t) { setTimeout(function () { report(false); }, t); });

  window.addEventListener('message', function (e) {
    if (e.data && e.data.channel === CH && e.data.type === 'research-data') {
      window.researchData = e.data.payload;
      window.dispatchEvent(new CustomEvent('research-data', { detail: e.data.payload }));
      setTimeout(function () { report(true); }, 0);
    }
  });
})();<\/script>`)

const documentHtml = computed(() => {
  const html = props.html || ''
  // Put the reporter last so the artifact's own scripts have already defined
  // whatever they need; before </body> when there is one, appended otherwise.
  const close = html.lastIndexOf('</body>')
  if (close !== -1) {
    return html.slice(0, close) + shim.value + html.slice(close)
  }
  return html + shim.value
})

// A document sized in viewport units (`min-height: 100vh`) grows every time we grow
// the frame, because the frame *is* its viewport: grow → 100vh grows → it reports
// more → grow again. Counting consecutive increases stops that without capping a
// legitimately tall artifact.
const GROWTH_LIMIT = 20
let increases = 0
let frozen = false

function onMessage(e: MessageEvent) {
  // The artifact knows its channel — we injected it — so the token alone does not
  // prove the sender. Only the frame we created may drive our height.
  if (e.source !== frameRef.value?.contentWindow) return
  // Enlarged, the frame's height is the screen's and comes from CSS. Recording
  // what the document reports then would do two wrong things at once: a
  // `min-height: 100vh` artifact measures the screen and the frame would come
  // back that tall, and the growth counter would burn through its limit on a
  // resize that was never a loop.
  if (enlarged.value) return

  const data = e.data
  if (!data || data.channel !== channel) return
  if (data.type !== 'height' || typeof data.height !== 'number') return
  if (!Number.isFinite(data.height)) return

  let h = Math.max(Math.ceil(data.height), 80)
  if (props.maxHeight > 0) h = Math.min(h, props.maxHeight)

  if (h > frameHeight.value) {
    increases += 1
    if (increases > GROWTH_LIMIT) frozen = true
  } else {
    increases = 0
  }
  // Once frozen, still allow shrinking — that ends the loop rather than feeding it.
  if (frozen && h > frameHeight.value) return

  frameHeight.value = h
  measured.value = true
}

function onLoad() {
  if (props.bridgeData && frameRef.value?.contentWindow) {
    frameRef.value.contentWindow.postMessage(
      { channel, type: 'research-data', payload: props.bridgeData },
      '*'
    )
  }
}

watch(
  () => props.bridgeData,
  () => onLoad()
)

// Re-measure from scratch when the document itself changes.
watch(
  () => props.html,
  () => {
    measured.value = false
    frameHeight.value = props.fallbackHeight
    // `frozen` and `increases` too. Without them a document that once tripped
    // the viewport-unit guard leaves the frame capped at the fallback height
    // for every document that replaces it — which arrives by WebSocket, so
    // nobody connects the two.
    frozen = false
    increases = 0
    armHint()
  }
)

onMounted(() => {
  armHint()
  window.addEventListener('message', onMessage)
  document.addEventListener('fullscreenchange', onFsChange)
  document.addEventListener('webkitfullscreenchange', onFsChange)
  window.addEventListener('keydown', onKeydown)
})
onUnmounted(() => {
  window.removeEventListener('message', onMessage)
  document.removeEventListener('fullscreenchange', onFsChange)
  document.removeEventListener('webkitfullscreenchange', onFsChange)
  window.removeEventListener('keydown', onKeydown)
  clearTimeout(hintTimer)
  // A component torn down mid-overlay would otherwise leave the page unable to
  // scroll and half of it permanently inert, with nothing on screen to explain
  // why. Not closeOverlay(), which would chase focus onto a button that is
  // being unmounted.
  if (overlay.value) {
    releaseInert()
    document.body.classList.remove('scroll-locked')
  }
})
</script>

<style scoped>
/* Positioned so the expand button has something to sit in the corner of. */
.artifact-frame { position: relative; display: flex; flex-direction: column; gap: var(--space-2); }

.artifact-iframe {
  width: 100%;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-bg);
  display: block;
  /* No height transition: while it animates the frame is shorter than its
     content, which shows a scrollbar that then disappears. */
}

.artifact-notice {
  padding: var(--space-2) 0 0;
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
}

.artifact-hint {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin: 0;
}

/* --------------------------------------------------------------------------
   Fullscreen
   -------------------------------------------------------------------------- */

/* The numbers match `.mermaid-btn` on purpose: one `.entry-content` card can
   hold a diagram and an artifact, and two recessive controls over the same dark
   surface that differ by two pixels and brighten in opposite directions read as
   two different products. The glass is fused into the button here rather than
   living on a toolbar, because there is only one control. */
/* 26px matches `.mermaid-btn`, the other recessive control that floats over
   dark content — one `.entry-content` card can hold a diagram and an artifact,
   and two controls for the same gesture differing by two pixels read as two
   different products.

   The colours deliberately do not match, and the reason is the backdrop. A
   mermaid button is bare because its toolbar supplies the plate; this is one
   control and has to supply its own. And it floats over a document we did not
   write: the light one in the catalog put a muted glyph on a half-black plate
   at 1.08:1, a featureless grey square. So the plate is opaque enough to be a
   plate on any background, and the glyph is `--color-text` rather than muted. */
.artifact-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 26px;
  height: 26px;
  padding: 0;
  color: var(--color-text);
  background: rgba(0, 0, 0, 0.72);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  backdrop-filter: blur(4px);
  cursor: pointer;
}
.artifact-btn:hover { background: rgba(0, 0, 0, 0.88); }
/* Two rings, because the backdrop is a document we do not control and either
   ring alone disappears against some artifact: the accent is 1.96:1 on white. */
.artifact-btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
  box-shadow: 0 0 0 4px rgba(0, 0, 0, 0.6);
}

.artifact-expand {
  position: absolute;
  top: var(--space-2);
  right: var(--space-2);
  /* Above the frame, below anything the page floats over the article. */
  z-index: var(--z-in-page);
  /* Recessive, never absent. Hover is not a signal we can rely on here: the
     iframe eats pointer events over the content, and a touch screen has no
     hover at all — so this stays legible without being loud. */
  opacity: 0.7;
  transition: opacity var(--transition-fast);
}
.artifact-frame:hover .artifact-expand,
.artifact-expand:focus-visible { opacity: 1; }

/* The padding matches the roadmap and mindmap toolbars, the product's other two
   full-viewport takeovers. */
.artifact-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-5);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}
.artifact-bar-title {
  margin: 0;
  min-width: 0;
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  font-size: var(--type-sm);
}
.artifact-bar-kind {
  flex: none;
  font-size: var(--type-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-primary);
}
.artifact-bar-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: var(--weight-medium);
  color: var(--color-text);
}
.artifact-bar .artifact-btn { flex: none; }

/* The two ways of filling the screen differ only in how they get there. */
.artifact-frame:fullscreen,
.artifact-frame.is-overlay {
  position: fixed;
  inset: 0;
  z-index: var(--z-overlay);
  gap: 0;
  background: var(--color-bg);
}
.artifact-frame:fullscreen .artifact-iframe,
.artifact-frame.is-overlay .artifact-iframe {
  /* No inline height is set while enlarged — an inline declaration outranks
     every author rule, and the frame would keep the size it had on the page. */
  flex: 1;
  min-height: 0;
  border: none;
  border-radius: 0;
}

@media print {
  /* Chrome is not part of the document. `no-print` covers the markup; this
     covers a page printed while an artifact happens to be enlarged, which would
     otherwise print one artifact and nothing else. */
  .artifact-frame.is-overlay {
    position: static;
    inset: auto;
    z-index: auto;
  }
  /* Undoing the position was not enough. While enlarged there is no inline
     height — the screen supplies it — and the frame is `flex: 1; min-height: 0`
     in a column that has just become auto-height, which resolves to the
     iframe's intrinsic 150px. Measured: a 900px document printed as a 150px
     sliver. A definite height has to come back from somewhere. */
  .artifact-frame:fullscreen .artifact-iframe,
  .artifact-frame.is-overlay .artifact-iframe {
    flex: none;
    height: auto;
    min-height: 60vh;
  }
  /* No frame drawn around it: the artifact brings its own surface, and a border
     on top of that only competes with it. */
  .artifact-iframe { border: none; }
  /* The dark surface behind an artifact is painted by THIS element, not by the
     document inside it — most artifacts are written expecting a dark host and set
     no background of their own. Browsers drop an element's background when
     printing, which left light text on white paper. */
  .artifact-iframe {
    print-color-adjust: exact;
    -webkit-print-color-adjust: exact;
  }
  /* An iframe has no internal break points, so a split cuts the visual in half. */
  .artifact-frame { break-inside: avoid; }
}
</style>
