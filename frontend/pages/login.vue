<script setup lang="ts">
const { login, allowRegistration } = useAuth()
const route = useRoute()
useHead({ title: 'Sign in' })

// Keep the destination when moving between the two auth pages.
const next = computed(() => safeNext(route.query.next) ?? '/')
const registerLink = computed(() =>
  route.query.next ? `/register?next=${encodeURIComponent(String(route.query.next))}` : '/register',
)
const email = ref('')
const password = ref('')
const error = ref('')
const errorBox = ref<HTMLElement | null>(null)
const submitting = ref(false)

async function handleSubmit() {
  if (submitting.value) return
  error.value = ''
  submitting.value = true
  try {
    await login(email.value, password.value)
    await navigateTo(next.value)
  } catch (e: any) {
    error.value = e?.data?.error || 'Could not sign in. Check your connection and try again.'
    await nextTick()
    errorBox.value?.focus()
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthShell title="Welcome back." subtitle="Sign in to your Dovod workspace.">
    <form class="auth-form" :aria-busy="submitting" @submit.prevent="handleSubmit">
      <div v-if="error" ref="errorBox" class="auth-error" role="alert" tabindex="-1">{{ error }}</div>
      <div class="auth-field">
        <label for="login-email" class="auth-label">Email</label>
        <input id="login-email" v-model.trim="email" name="email" type="email" required autocomplete="email" class="auth-input" placeholder="you@example.com" />
      </div>
      <div class="auth-field">
        <label for="login-password" class="auth-label">Password</label>
        <AuthPasswordInput id="login-password" v-model="password" autocomplete="current-password" />
      </div>
      <button type="submit" class="auth-button" :disabled="submitting">
        <span aria-live="polite">{{ submitting ? 'Signing in…' : 'Sign in' }}</span><span aria-hidden="true">{{ submitting ? '···' : '→' }}</span>
      </button>
    </form>
    <template #footer>
      <p v-if="allowRegistration" class="auth-footer">New to Dovod? <NuxtLink :to="registerLink" class="auth-link">Create an account</NuxtLink></p>
    </template>
  </AuthShell>
</template>
