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

const ready = ref(false)
onMounted(() => {
  requestAnimationFrame(() => {
    ready.value = true
  })
})

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

<template>
  <div class="auth-page">
    <!-- Left hero panel -->
    <div class="auth-hero">
      <div class="auth-hero__bg" aria-hidden="true">
        <div class="auth-hero__orb auth-hero__orb--1" />
        <div class="auth-hero__orb auth-hero__orb--2" />
        <div class="auth-hero__orb auth-hero__orb--3" />
        <div class="auth-hero__grid" />
      </div>

      <div class="auth-hero__content">
        <h2 :class="['auth-hero__headline', { 'is-visible': ready }]">
          Organize your<br>
          <span class="auth-hero__accent">knowledge</span><br>
          with AI precision.
        </h2>
        <p :class="['auth-hero__tagline', { 'is-visible': ready }]">
          AI structures it. You explore it.
        </p>
        <p :class="['auth-hero__cta', { 'is-visible': ready }]">
          Create your account and start &rarr;
        </p>
      </div>

      <div :class="['auth-hero__bottom', { 'is-visible': ready }]">
        <span class="auth-hero__bottom-logo">R</span>
        <span class="auth-hero__bottom-text">Research</span>
      </div>
    </div>

    <!-- Right form panel -->
    <div class="auth-form-panel">
      <div class="auth-form-inner">
        <div :class="['auth-form-card', { 'is-visible': ready }]">
          <div class="auth-form-logo">R</div>

          <h1 class="auth-title">Create account</h1>
          <p class="auth-subtitle">Get started with Research</p>

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
            Already have an account? <NuxtLink :to="loginLink" class="auth-link">Sign in</NuxtLink>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Page layout ────────────────────────────────────────── */
.auth-page {
  display: flex;
  min-height: 100dvh;
  width: 100vw;
}

/* ── Left hero panel ────────────────────────────────────── */
.auth-hero {
  display: none;
  position: relative;
  flex-direction: column;
  justify-content: space-between;
  width: 50%;
  padding: var(--space-12) var(--space-12);
  background: var(--color-bg-deep);
  overflow: hidden;
}

@media (min-width: 1024px) {
  .auth-hero { display: flex; }
}
@media (min-width: 1280px) {
  .auth-hero { width: 55%; padding: var(--space-16) var(--space-16); }
}

/* Animated background */
.auth-hero__bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.auth-hero__orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.35;
  will-change: transform;
}
.auth-hero__orb--1 {
  width: 500px; height: 500px;
  top: -10%; right: -10%;
  background: radial-gradient(circle, rgba(108, 197, 224, 0.5) 0%, transparent 70%);
  animation: orbDrift1 12s ease-in-out infinite;
}
.auth-hero__orb--2 {
  width: 400px; height: 400px;
  bottom: -5%; left: -5%;
  background: radial-gradient(circle, rgba(107, 157, 240, 0.4) 0%, transparent 70%);
  animation: orbDrift2 15s ease-in-out infinite;
}
.auth-hero__orb--3 {
  width: 300px; height: 300px;
  top: 40%; left: 30%;
  background: radial-gradient(circle, rgba(52, 211, 153, 0.3) 0%, transparent 70%);
  animation: orbDrift3 18s ease-in-out infinite;
}

.auth-hero__grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(ellipse 70% 60% at 50% 50%, black 30%, transparent 100%);
}

@keyframes orbDrift1 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(-30px, 20px) scale(1.05); }
  66% { transform: translate(20px, -15px) scale(0.95); }
}
@keyframes orbDrift2 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(25px, -20px) scale(1.08); }
  66% { transform: translate(-15px, 25px) scale(0.92); }
}
@keyframes orbDrift3 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(-20px, -30px) scale(1.1); }
}

/* Hero content */
.auth-hero__content {
  position: relative;
  z-index: 10;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  max-width: 520px;
}

.auth-hero__headline {
  font-family: 'Outfit', sans-serif;
  font-weight: var(--weight-bold);
  font-size: 3.25rem;
  line-height: 1.08;
  letter-spacing: -0.035em;
  color: #fff;
  opacity: 0;
  transform: translateY(16px);
  transition: opacity 0.6s ease, transform 0.6s ease;
  transition-delay: 0.2s;
}
.auth-hero__headline.is-visible {
  opacity: 1;
  transform: translateY(0);
}

.auth-hero__accent {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-info) 50%, var(--color-success) 100%);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.auth-hero__tagline {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  color: rgba(255, 255, 255, 0.3);
  margin-top: var(--space-6);
  opacity: 0;
  transform: translateY(12px);
  transition: opacity 0.5s ease, transform 0.5s ease;
  transition-delay: 0.5s;
}
.auth-hero__tagline.is-visible {
  opacity: 1;
  transform: translateY(0);
}

.auth-hero__cta {
  font-family: 'Outfit', sans-serif;
  font-size: var(--type-base);
  font-weight: var(--weight-medium);
  color: rgba(255, 255, 255, 0.5);
  margin-top: var(--space-12);
  opacity: 0;
  transform: translateY(12px);
  transition: opacity 0.5s ease, transform 0.5s ease;
  transition-delay: 0.7s;
}
.auth-hero__cta.is-visible {
  opacity: 1;
  transform: translateY(0);
}

/* Hero bottom */
.auth-hero__bottom {
  position: relative;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  color: rgba(255, 255, 255, 0.2);
  opacity: 0;
  transition: opacity 0.5s ease;
  transition-delay: 0.7s;
}
.auth-hero__bottom.is-visible { opacity: 1; }

.auth-hero__bottom-logo {
  width: 28px; height: 28px;
  border-radius: var(--radius-sm);
  background: rgba(108, 197, 224, 0.15);
  color: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-family: 'Outfit', sans-serif;
  font-weight: var(--weight-bold);
  font-size: var(--type-sm);
}

/* ── Right form panel ───────────────────────────────────── */
.auth-form-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
  padding: var(--space-6) var(--space-6);
}
@media (min-width: 640px) {
  .auth-form-panel { padding: var(--space-12) var(--space-12); }
}

.auth-form-inner {
  width: 100%;
  max-width: 380px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 480px;
}

.auth-form-card {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  opacity: 0;
  transform: translateY(16px);
  transition: opacity 0.6s ease, transform 0.6s ease;
  transition-delay: 0.15s;
}
.auth-form-card.is-visible {
  opacity: 1;
  transform: translateY(0);
}

.auth-form-logo {
  width: 48px; height: 48px;
  border-radius: var(--radius);
  background: var(--color-primary-muted);
  color: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-family: 'Outfit', sans-serif;
  font-weight: var(--weight-bold);
  font-size: var(--type-xl);
  margin-bottom: var(--space-8);
}

.auth-title {
  font-size: var(--type-xl);
  font-weight: var(--weight-bold);
  letter-spacing: -0.02em;
  margin-bottom: var(--space-2);
  text-align: center;
}

.auth-subtitle {
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  margin-bottom: var(--space-8);
  text-align: center;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  width: 100%;
}

.auth-label {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

.auth-input {
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius);
  font-size: var(--type-sm);
  background: var(--color-surface);
  color: var(--color-text);
  font-family: inherit;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.auth-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: var(--shadow-focus);
}
.auth-input::placeholder {
  color: var(--color-text-muted);
  opacity: 0.5;
}

.auth-button {
  position: relative;
  padding: var(--space-3) var(--space-4);
  background: linear-gradient(135deg, var(--color-primary), var(--color-info));
  background-size: 200% 200%;
  animation: btnShift 4s ease-in-out infinite;
  color: var(--color-bg);
  border: none;
  border-radius: var(--radius);
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  cursor: pointer;
  font-family: inherit;
  margin-top: var(--space-2);
  transition: transform 200ms ease, box-shadow 300ms ease;
  box-shadow: 0 0 20px -6px rgba(108, 197, 224, 0.25);
}
.auth-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 0 32px -4px rgba(108, 197, 224, 0.4);
}
.auth-button:active {
  transform: translateY(0);
}
.auth-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

@keyframes btnShift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.auth-error {
  padding: var(--space-3) var(--space-4);
  background: rgba(239, 107, 107, 0.08);
  border: 1px solid rgba(239, 107, 107, 0.2);
  color: var(--color-error);
  border-radius: var(--radius);
  font-size: var(--type-sm);
}

.auth-footer {
  margin-top: var(--space-6);
  text-align: center;
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}

.auth-link {
  color: var(--color-primary);
  text-decoration: none;
  font-weight: var(--weight-medium);
}
.auth-link:hover {
  text-decoration: underline;
}

/* ── Reduced motion ─────────────────────────────────────── */
@media (prefers-reduced-motion: reduce) {
  .auth-hero__orb,
  .auth-button { animation: none; }
  .auth-button:hover { transform: none; }
  .auth-hero__headline,
  .auth-hero__tagline,
  .auth-hero__cta,
  .auth-hero__bottom,
  .auth-form-card {
    opacity: 1;
    transform: none;
    transition: none;
  }
}

/* ── Mobile: single column ──────────────────────────────── */
@media (max-width: 1023px) {
  .auth-page { flex-direction: column; }
  .auth-form-panel { min-height: 100dvh; }
}
</style>
