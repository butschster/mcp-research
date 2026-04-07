<script setup lang="ts">
const { register, allowRegistration } = useAuth()

const email = ref('')
const password = ref('')
const name = ref('')
const error = ref('')
const submitting = ref(false)

// Redirect if registration is disabled
if (!allowRegistration.value) {
  navigateTo('/login')
}

async function handleSubmit() {
  error.value = ''
  submitting.value = true
  try {
    await register(email.value, password.value, name.value)
    navigateTo('/')
  } catch (e: any) {
    error.value = e?.data?.error || e?.message || 'Registration failed'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1 class="auth-title">Create account</h1>
      <p class="auth-subtitle">Get started with MCP Research</p>

      <form @submit.prevent="handleSubmit" class="auth-form">
        <div v-if="error" class="auth-error">{{ error }}</div>

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

      <p class="auth-footer">
        Already have an account? <NuxtLink to="/login" class="auth-link">Sign in</NuxtLink>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}
.auth-card {
  width: 100%;
  max-width: 400px;
  padding: var(--space-8);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
}
.auth-title {
  font-size: var(--text-2xl);
  font-weight: 600;
  margin-bottom: var(--space-1);
}
.auth-subtitle {
  color: var(--color-text-secondary);
  margin-bottom: var(--space-6);
}
.auth-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.auth-label {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  font-size: var(--text-sm);
  font-weight: 500;
}
.auth-input {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  background: var(--color-bg);
  color: var(--color-text);
  font-family: inherit;
}
.auth-input:focus {
  outline: 2px solid var(--color-primary);
  outline-offset: -1px;
}
.auth-button {
  padding: var(--space-2) var(--space-4);
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  margin-top: var(--space-2);
}
.auth-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.auth-error {
  padding: var(--space-2) var(--space-3);
  background: var(--color-danger-bg, #fef2f2);
  color: var(--color-danger, #dc2626);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
}
.auth-footer {
  margin-top: var(--space-4);
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}
.auth-link {
  color: var(--color-primary);
  text-decoration: none;
}
.auth-link:hover {
  text-decoration: underline;
}

/* Responsive */
@media (max-width: 480px) {
  .auth-card { padding: var(--space-5); margin: 0 var(--space-2); }
}
</style>
