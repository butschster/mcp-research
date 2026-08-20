<template>
  <ModalOverlay :visible="visible" size="sm" labelledby="send-back-title" @close="$emit('cancel')">
    <ModalHeader title="Send back" title-id="send-back-title" @close="$emit('cancel')" />

    <div class="send-back">
      <p class="send-back__lead">
        {{ count === 1 ? 'This mark goes back to the agent.' : `${count} marks go back to the agent.` }}
        Say what is still wrong — the agent is required to read this before trying again, and a
        blank reason usually buys the same answer a second time.
      </p>

      <label class="send-back__label" :for="fieldId">Why is this not settled?</label>
      <textarea
        :id="fieldId"
        ref="field"
        v-model="reason"
        class="form-textarea"
        rows="3"
        placeholder="A vendor blog is not a measurement — this needs the benchmark itself."
      />

      <div class="send-back__actions">
        <button type="button" class="btn btn-sm" @click="$emit('cancel')">Cancel</button>
        <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="submit">
          {{ busy ? 'Sending…' : 'Send back' }}
        </button>
      </div>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
/**
 * Collecting the reason an answer was refused.
 *
 * This replaced `window.prompt`, which was the wrong instrument for the most
 * load-bearing free text in the feature: the agent is *required* to read this
 * before its next attempt. A native prompt blocks the thread, cannot be styled
 * or tested, is suppressible by the browser, and — inside the pass review,
 * which is itself a modal with a focus trap — fights the dialog it is called
 * from.
 *
 * Deliberately not a variant of ConfirmModal: that one has no input, and giving
 * it one would make every other confirmation carry a field it does not want.
 */
const props = withDefaults(defineProps<{
  visible: boolean
  /** How many marks this decision covers, so the sentence reads right for one. */
  count?: number
  busy?: boolean
}>(), { count: 1 })

const emit = defineEmits<{ confirm: [reason: string]; cancel: [] }>()

const reason = ref('')
const field = ref<HTMLTextAreaElement | null>(null)
const fieldId = useId()

function submit() {
  emit('confirm', reason.value.trim())
}

// A reason typed for one batch must never travel to the next.
watch(() => props.visible, async (open) => {
  if (!open) return
  reason.value = ''
  await nextTick()
  field.value?.focus()
})
</script>

<style scoped>
.send-back {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: var(--space-4);
}

.send-back__lead {
  margin: 0;
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}

.send-back__label {
  font-size: var(--type-2xs);
  font-weight: 600;
  color: var(--color-text-muted);
}

.send-back__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}
</style>
