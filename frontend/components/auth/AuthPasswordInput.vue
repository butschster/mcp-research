<script setup lang="ts">
defineProps<{
  id: string
  autocomplete: 'current-password' | 'new-password'
  minlength?: number
  describedby?: string
}>()
const value = defineModel<string>({ required: true })
const visible = ref(false)
</script>

<template>
  <div class="auth-password">
    <input :id="id" v-model="value" name="password" :type="visible ? 'text' : 'password'" :autocomplete="autocomplete" :minlength="minlength" :aria-describedby="describedby" required class="auth-input" />
    <button type="button" class="auth-password-toggle" :aria-label="visible ? 'Hide password' : 'Show password'" :aria-pressed="visible" :aria-controls="id" @click="visible = !visible">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z" /><circle cx="12" cy="12" r="3" /><path v-if="visible" d="m3 3 18 18" />
      </svg>
    </button>
  </div>
</template>

<style scoped>
.auth-password { position: relative; }
.auth-password .auth-input { width: 100%; padding-right: 3.5rem; }
.auth-password-toggle {
  position: absolute;
  inset: 2px 2px 2px auto;
  display: grid;
  place-items: center;
  width: 2.875rem;
  padding: 0;
  color: var(--color-text-muted);
  background: transparent;
  border: 0;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: color var(--transition-fast), background var(--transition-fast);
}
.auth-password-toggle:hover { color: var(--color-text); background: var(--color-surface-hover); }
</style>
