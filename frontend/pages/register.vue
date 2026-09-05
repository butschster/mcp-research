<script setup lang="ts">
const { register, allowRegistration } = useAuth()
const route = useRoute()
useHead({ title: 'Create an account' })

const next = computed(() => safeNext(route.query.next) ?? '/')
const loginLink = computed(() =>
  route.query.next ? `/login?next=${encodeURIComponent(String(route.query.next))}` : '/login',
)
const email = ref(typeof route.query.email === 'string' ? route.query.email : '')
const password = ref('')
const name = ref('')
const error = ref('')
const errorBox = ref<HTMLElement | null>(null)
const submitting = ref(false)

// An invitation still permits registration when public signup is closed.
const inviteToken = computed(() => typeof route.query.invite === 'string' ? route.query.invite : '')
if (!allowRegistration.value && !inviteToken.value) {
  navigateTo(loginLink.value)
}

async function handleSubmit() {
  if (submitting.value) return
  error.value = ''
  submitting.value = true
  try {
    await register(email.value, password.value, name.value, inviteToken.value || undefined)
    await navigateTo(next.value)
  } catch (e: any) {
    error.value = e?.data?.error || 'Could not create your account. Check your connection and try again.'
    await nextTick()
    errorBox.value?.focus()
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthShell title="Make space for your ideas." subtitle="Create your Dovod account to get started.">
    <template #headline>Start with<br><em>a question.</em></template>
    <form class="auth-form" :aria-busy="submitting" @submit.prevent="handleSubmit">
      <div v-if="error" ref="errorBox" class="auth-error" role="alert" tabindex="-1">{{ error }}</div>
      <div class="auth-field">
        <label for="register-name" class="auth-label">Your name</label>
        <input id="register-name" v-model.trim="name" name="name" type="text" required autocomplete="name" class="auth-input" placeholder="Your name" />
      </div>
      <div class="auth-field">
        <label for="register-email" class="auth-label">Email</label>
        <input id="register-email" v-model.trim="email" name="email" type="email" required autocomplete="email" class="auth-input" placeholder="you@example.com" />
      </div>
      <div class="auth-field">
        <label for="register-password" class="auth-label">Password</label>
        <AuthPasswordInput id="register-password" v-model="password" autocomplete="new-password" :minlength="6" describedby="password-hint" />
        <p id="password-hint" class="auth-hint">Use at least 6 characters.</p>
      </div>
      <button type="submit" class="auth-button" :disabled="submitting">
        <span aria-live="polite">{{ submitting ? 'Creating account…' : 'Create account' }}</span><span aria-hidden="true">{{ submitting ? '···' : '→' }}</span>
      </button>
    </form>
    <template #footer>
      <p class="auth-footer">Already have an account? <NuxtLink :to="loginLink" class="auth-link">Sign in</NuxtLink></p>
    </template>
  </AuthShell>
</template>
