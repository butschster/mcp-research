<template>
  <div class="artifact-frame">
    <iframe
      ref="frameRef"
      class="artifact-iframe"
      :title="title || 'Artifact'"
      sandbox="allow-scripts"
      referrerpolicy="no-referrer"
      :srcdoc="documentHtml"
      :style="{ height: frameHeight + 'px' }"
      @load="onLoad"
    ></iframe>

    <p v-if="!measured" class="artifact-hint no-print">
      Sizing the artifact… if it stays at this height, the document did not report
      its size and is shown at the fallback height.
    </p>
  </div>
</template>

<script setup lang="ts">
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
const frameHeight = ref(props.fallbackHeight)
const measured = ref(false)

// A unique token per mount: the frame has an opaque origin (allow-scripts without
// allow-same-origin), so messages arrive with origin "null" and cannot be trusted
// by origin alone. Matching this token keeps other frames' messages out.
const channel = `artifact-${Math.random().toString(36).slice(2)}`

/**
 * The sandbox deliberately withholds allow-same-origin, so the parent cannot read
 * the frame's document to measure it. Instead we append a reporter to the document
 * we are about to render: it posts its height on load and on every resize.
 */
const shim = computed(() => `
<script>(function(){
  var CH = ${JSON.stringify(channel)};
  var last = -1;

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
  }
)

onMounted(() => window.addEventListener('message', onMessage))
onUnmounted(() => window.removeEventListener('message', onMessage))
</script>

<style scoped>
.artifact-frame { display: flex; flex-direction: column; gap: var(--space-2); }

.artifact-iframe {
  width: 100%;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-bg);
  display: block;
  /* No height transition: while it animates the frame is shorter than its
     content, which shows a scrollbar that then disappears. */
}

.artifact-hint {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin: 0;
}

@media print {
  .artifact-iframe { border-color: #ddd; }
}
</style>
