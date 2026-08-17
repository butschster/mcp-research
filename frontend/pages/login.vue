<template>
  <AuthShell
    tagline="Let AI organize your knowledge."
    cta="Sign in to start researching &rarr;"
    title="Welcome back"
    subtitle="Sign in to continue to your research"
  >
    <template #headline>AI-driven<br>
      <span class="auth-hero__accent">structured</span><br>
      research sessions.</template>

          <form @submit.prevent="handleSubmit" class="auth-form">
            <div v-if="error" class="auth-error" role="alert">{{ error }}</div>

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
                autocomplete="current-password"
                class="auth-input"
                placeholder="Your password"
              />
            </label>

            <button type="submit" class="auth-button" :disabled="submitting">
              {{ submitting ? 'Signing in...' : 'Sign in' }}
            </button>
          </form>

    <template #footer>
          <p v-if="allowRegistration" class="auth-footer">
            Don't have an account? <NuxtLink :to="registerLink" class="auth-link">Create one</NuxtLink>
          </p>
    </template>

  </AuthShell>
</template>

<script setup lang="ts">
const { login, authEnabled, allowRegistration } = useAuth()
const route = useRoute()

// Where to land afterwards. An invitation link sends people here and has to get
// them back, so signing in resumes the journey instead of ending it on the
// research list.
const next = computed(() => safeNext(route.query.next) ?? '/')
const registerLink = computed(() =>
  route.query.next ? `/register?next=${encodeURIComponent(String(route.query.next))}` : '/register',
)

const email = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)


async function handleSubmit() {
  error.value = ''
  submitting.value = true
  try {
    await login(email.value, password.value)
    navigateTo(next.value)
  } catch (e: any) {
    error.value = e?.data?.error || e?.message || 'Login failed'
  } finally {
    submitting.value = false
  }
}
</script>
