<template>
  <div class="centered-gate password-gate">
    <div class="password-card">
      <!-- A visitor who cannot read the prompt in this theme has no other
           control on the screen; the banner that normally carries the toggle
           is not rendered until the link is unlocked. -->
      <div class="gate-tools"><ThemeToggle size="sm" /></div>
      <div class="password-icon" aria-hidden="true">&#x1F512;</div>
      <h1 class="password-title">This link is password-protected</h1>
      <p class="card-meta">Ask the person who sent you the link for the password.</p>

      <form class="password-form" @submit.prevent="submit">
        <label class="field-label" for="share-password">Password</label>
        <input
          id="share-password"
          ref="field"
          v-model="password"
          type="password"
          class="text-input"
          autocomplete="current-password"
          :disabled="busy"
        />
        <p v-if="message" class="inline-error" role="alert">{{ message }}</p>
        <button class="btn btn-primary" type="submit" :disabled="busy || !password">
          {{ busy ? 'Opening…' : 'Open' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * The one field between a visitor and a protected link.
 *
 * The field is deliberately not cleared on a wrong answer: the failure mode
 * here is retyping a long password out of a chat message, not somebody reading
 * it over a shoulder.
 */
const props = defineProps<{
  busy?: boolean
  error?: '' | 'wrong' | 'throttled'
}>()

const emit = defineEmits<{ submit: [password: string] }>()

const password = ref('')
const field = ref<HTMLInputElement | null>(null)

const message = computed(() => {
  if (props.error === 'wrong') return "That password doesn't match. Check it with whoever sent you the link."
  if (props.error === 'throttled') return 'Too many tries. Wait a minute, then try again.'
  return ''
})

function submit() {
  if (!password.value || props.busy) return
  emit('submit', password.value)
}

onMounted(() => field.value?.focus())
</script>

<style scoped>
.password-card {
  position: relative;
  width: 100%;
  max-width: 380px;
  text-align: center;
}
.gate-tools {
  position: absolute;
  top: var(--space-3);
  right: var(--space-3);
}
.password-icon { font-size: 2rem; margin-bottom: var(--space-4); opacity: 0.7; }
.password-title {
  font-size: var(--type-lg);
  font-weight: var(--weight-semibold);
  margin-bottom: var(--space-2);
  line-height: var(--line-tight);
}
.password-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin-top: var(--space-5);
  text-align: left;
}
.password-form .btn { margin-top: var(--space-2); justify-content: center; }
</style>
