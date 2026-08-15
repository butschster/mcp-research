<template>
  <div>
    <div class="empty-command secret-block">
      <input
        v-if="clipboardBroken"
        ref="fieldEl"
        class="command-text secret-input"
        :value="value"
        readonly
        @focus="($event.target as HTMLInputElement).select()"
      />
      <code v-else class="command-text">{{ value }}</code>
      <button v-if="!clipboardBroken" ref="buttonEl" class="copy-btn" :class="{ copied }" @click="copy">
        {{ copied ? '&#x2713; Copied' : 'Copy' }}
      </button>
    </div>
    <p v-if="clipboardBroken" class="modal-help">
      Copying isn't available on this connection — select the link above and copy it manually.
    </p>
    <p v-else-if="hint" class="modal-help">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
/**
 * A value the server will not show again, with a copy button.
 *
 * The clipboard fallback is why this is a component and not three lines of
 * markup: `navigator.clipboard` is simply absent on plain HTTP, which is a
 * normal way to run this product on a LAN, and a dead Copy button on the one
 * screen that shows a secret once is the worst place to discover that. The
 * behaviour existed in two places already and would have drifted between them.
 */
const props = defineProps<{
  value: string
  /** A line under the block, replaced by the clipboard-absent message. */
  hint?: string
  /** What the toast says on success. */
  toast?: string
}>()

const emit = defineEmits<{ copied: [] }>()

const copied = ref(false)
const clipboardBroken = ref(false)
const fieldEl = ref<HTMLInputElement | null>(null)
const buttonEl = ref<HTMLButtonElement | null>(null)

async function copy() {
  const { success } = useToasts()
  try {
    if (!navigator.clipboard) throw new Error('no clipboard')
    await navigator.clipboard.writeText(props.value)
    copied.value = true
    emit('copied')
    success(props.toast || 'Copied')
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    clipboardBroken.value = true
    await nextTick()
    fieldEl.value?.focus()
    fieldEl.value?.select()
  }
}

/** Lets the dialog put focus on the thing the reader came for. */
function focus() {
  // Not `?? `: `.focus()` returns undefined, so the right-hand side ran every
  // time. Harmless while the two are mutually exclusive, and a trap for the
  // next edit.
  if (buttonEl.value) buttonEl.value.focus()
  else fieldEl.value?.focus()
}

// copy() is exposed so the dialog's primary button can mean what it says.
defineExpose({ focus, copy })
</script>

<style scoped>
.secret-block { margin-top: 0; max-width: none; }
.secret-input { background: none; border: none; color: var(--color-primary); font-family: inherit; width: 100%; }
/* On plain HTTP this field is the only way to get the link, and the only
   focusable thing on the screen. Taking its ring away left a keyboard user with
   nothing to go on. */
.secret-input:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
</style>
