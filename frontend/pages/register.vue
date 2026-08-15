<template>
  <AuthShell
    tagline="AI structures it. You explore it."
    cta="Create your account and start &rarr;"
    title="Create account"
    subtitle="Get started with Research"
  >
    <template #headline>Organize your<br>
          <span class="auth-hero__accent">knowledge</span><br>
          with AI precision.</template>

          <form @submit.prevent="handleSubmit" class="auth-form">
            <div v-if="error" class="auth-error" role="alert">{{ error }}</div>

            <label class="auth-label">
              Name
              <input
                v-model="name"
                type="text"
                required
                autocomplete="name"
                class="auth-input"
                placeholder="Your name"
              />
            </label>

            <label class="auth-label">
              Email
              <input
                v-model="email"
                type="email"
                required
                autocomplete="email"
                class="auth-input"
                placeholder="you@example.com"
              />
            </label>

            <label class="auth-label">
              Password
              <input
                v-model="password"
                type="password"
                required
                minlength="6"
                autocomplete="new-password"
                class="auth-input"
                placeholder="At least 6 characters"
              />
            </label>

            <button type="submit" class="auth-button" :disabled="submitting">
              {{ submitting ? 'Creating account...' : 'Create account' }}
            </button>
          </form>

    <template #footer>
          <p class="auth-footer">
            Already have an account? <NuxtLink :to="loginLink" class="auth-link">Sign in</NuxtLink>
          </p>
    </template>

  </AuthShell>
</template>

<script setup lang="ts">
const { register, allowRegistration } = useAuth()
const route = useRoute()

// An invitation carries both the destination and the address it was sent to,
// so the field is filled in for someone who has just been handed a link.
const next = computed(() => safeNext(route.query.next) ?? '/')
const loginLink = computed(() =>
  route.query.next ? `/login?next=${encodeURIComponent(String(route.query.next))}` : '/login',
)

const email = ref(typeof route.query.email === 'string' ? route.query.email : '')
const password = ref('')
const name = ref('')
const error = ref('')
const submitting = ref(false)

// Redirect if registration is disabled
// An invitation is its own authorization, so a closed server still lets its
// holder through — the token is checked again server-side.
const inviteToken = computed(() => (typeof route.query.invite === 'string' ? route.query.invite : ''))
if (!allowRegistration.value && !inviteToken.value) {
  navigateTo(route.query.next ? `/login?next=${encodeURIComponent(String(route.query.next))}` : '/login')
}


async function handleSubmit() {
  error.value = ''
  submitting.value = true
  try {
    await register(email.value, password.value, name.value, inviteToken.value || undefined)
    navigateTo(next.value)
  } catch (e: any) {
    error.value = e?.data?.error || e?.message || 'Registration failed'
  } finally {
    submitting.value = false
  }
}
</script>
