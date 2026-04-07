import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, ref, onMounted } from 'vue'

/**
 * Two-column authentication pages: Login and Register.
 * Left panel: animated hero with marketing copy.
 * Right panel: centered form card.
 * On mobile (<1024px): hero hidden, form full-screen.
 */

// Inject auth page styles once into the document head
const STYLE_ID = 'auth-page-stories-styles'
function ensureStyles() {
  if (document.getElementById(STYLE_ID)) return
  const style = document.createElement('style')
  style.id = STYLE_ID
  style.textContent = `
    .auth-page { display: flex; min-height: 100vh; width: 100%; }
    .auth-hero { display: none; position: relative; flex-direction: column; justify-content: space-between; width: 50%; padding: 3rem; background: #080e1e; overflow: hidden; }
    @media (min-width: 1024px) { .auth-hero { display: flex; } }
    @media (min-width: 1280px) { .auth-hero { width: 55%; padding: 4rem; } }
    .auth-hero__bg { position: absolute; inset: 0; overflow: hidden; }
    .auth-hero__orb { position: absolute; border-radius: 50%; filter: blur(80px); opacity: 0.35; will-change: transform; }
    .auth-hero__orb--1 { width: 500px; height: 500px; top: -10%; right: -10%; background: radial-gradient(circle, rgba(108,197,224,0.5) 0%, transparent 70%); animation: orbDrift1 12s ease-in-out infinite; }
    .auth-hero__orb--2 { width: 400px; height: 400px; bottom: -5%; left: -5%; background: radial-gradient(circle, rgba(107,157,240,0.4) 0%, transparent 70%); animation: orbDrift2 15s ease-in-out infinite; }
    .auth-hero__orb--3 { width: 300px; height: 300px; top: 40%; left: 30%; background: radial-gradient(circle, rgba(52,211,153,0.3) 0%, transparent 70%); animation: orbDrift3 18s ease-in-out infinite; }
    .auth-hero__grid { position: absolute; inset: 0; background-image: linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px); background-size: 60px 60px; mask-image: radial-gradient(ellipse 70% 60% at 50% 50%, black 30%, transparent 100%); }
    @keyframes orbDrift1 { 0%,100%{transform:translate(0,0) scale(1);} 33%{transform:translate(-30px,20px) scale(1.05);} 66%{transform:translate(20px,-15px) scale(0.95);} }
    @keyframes orbDrift2 { 0%,100%{transform:translate(0,0) scale(1);} 33%{transform:translate(25px,-20px) scale(1.08);} 66%{transform:translate(-15px,25px) scale(0.92);} }
    @keyframes orbDrift3 { 0%,100%{transform:translate(0,0) scale(1);} 50%{transform:translate(-20px,-30px) scale(1.1);} }
    .auth-hero__content { position: relative; z-index: 10; flex: 1; display: flex; flex-direction: column; justify-content: center; max-width: 520px; }
    .auth-hero__headline { font-family: 'Outfit', sans-serif; font-weight: 700; font-size: 3.25rem; line-height: 1.08; letter-spacing: -0.035em; color: #fff; }
    .auth-hero__accent { background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-info) 50%, var(--color-success) 100%); background-clip: text; -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
    .auth-hero__tagline { font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); color: rgba(255,255,255,0.3); margin-top: var(--space-6); }
    .auth-hero__cta { font-family: 'Outfit', sans-serif; font-size: var(--type-base); font-weight: 500; color: rgba(255,255,255,0.5); margin-top: var(--space-12); }
    .auth-hero__bottom { position: relative; z-index: 10; display: flex; align-items: center; gap: var(--space-3); font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); color: rgba(255,255,255,0.2); }
    .auth-hero__bottom-logo { width: 28px; height: 28px; border-radius: var(--radius-sm); background: rgba(108,197,224,0.15); color: var(--color-primary); display: flex; align-items: center; justify-content: center; font-family: 'Outfit', sans-serif; font-weight: 700; font-size: var(--type-sm); }
    .auth-form-panel { flex: 1; display: flex; align-items: center; justify-content: center; background: var(--color-bg); padding: 3rem; }
    .auth-form-inner { width: 100%; max-width: 380px; display: flex; flex-direction: column; justify-content: center; min-height: 480px; }
    .auth-form-card { width: 100%; display: flex; flex-direction: column; align-items: center; }
    .auth-form-logo { width: 48px; height: 48px; border-radius: var(--radius); background: var(--color-primary-muted); color: var(--color-primary); display: flex; align-items: center; justify-content: center; font-family: 'Outfit', sans-serif; font-weight: 800; font-size: var(--type-xl); margin-bottom: var(--space-8); }
    .auth-title { font-size: var(--type-xl); font-weight: 700; letter-spacing: -0.02em; margin-bottom: var(--space-2); text-align: center; }
    .auth-subtitle { color: var(--color-text-muted); font-size: var(--type-sm); margin-bottom: var(--space-8); text-align: center; }
    .auth-form { display: flex; flex-direction: column; gap: var(--space-4); width: 100%; }
    .auth-label { display: flex; flex-direction: column; gap: var(--space-1); font-size: var(--type-xs); font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; color: var(--color-text-muted); }
    .auth-input { padding: var(--space-3) var(--space-4); border: 1px solid var(--color-border-strong); border-radius: var(--radius); font-size: var(--type-sm); background: var(--color-surface); color: var(--color-text); font-family: inherit; transition: border-color 0.15s ease, box-shadow 0.15s ease; }
    .auth-input:focus { outline: none; border-color: var(--color-primary); box-shadow: 0 0 0 3px rgba(108,197,224,0.15); }
    .auth-input::placeholder { color: var(--color-text-muted); opacity: 0.5; }
    .auth-button { padding: var(--space-3) var(--space-4); background: linear-gradient(135deg, var(--color-primary), var(--color-info)); background-size: 200% 200%; animation: btnShift 4s ease-in-out infinite; color: var(--color-bg); border: none; border-radius: var(--radius); font-size: var(--type-sm); font-weight: 600; cursor: pointer; font-family: inherit; margin-top: var(--space-2); transition: transform 200ms ease, box-shadow 300ms ease; box-shadow: 0 0 20px -6px rgba(108,197,224,0.25); }
    .auth-button:hover { transform: translateY(-1px); box-shadow: 0 0 32px -4px rgba(108,197,224,0.4); }
    @keyframes btnShift { 0%,100%{background-position:0% 50%;} 50%{background-position:100% 50%;} }
    .auth-error { padding: var(--space-3) var(--space-4); background: rgba(239,107,107,0.08); border: 1px solid rgba(239,107,107,0.2); color: var(--color-error); border-radius: var(--radius); font-size: var(--type-sm); }
    .auth-footer { margin-top: var(--space-6); text-align: center; font-size: var(--type-sm); color: var(--color-text-muted); }
    .auth-link { color: var(--color-primary); text-decoration: none; font-weight: 500; }
    .auth-link:hover { text-decoration: underline; }
    @media (max-width: 1023px) { .auth-page { flex-direction: column; } .auth-form-panel { min-height: 100vh; } }
  `
  document.head.appendChild(style)
}

// --- Login Page Stub ---
const LoginPageStub = defineComponent({
  name: 'LoginPageStub',
  props: {
    showError: { type: Boolean, default: false },
    showRegistrationLink: { type: Boolean, default: true },
  },
  setup() {
    onMounted(ensureStyles)
    const email = ref('')
    const password = ref('')
    return { email, password }
  },
  template: `
    <div class="auth-page">
      <div class="auth-hero">
        <div class="auth-hero__bg" aria-hidden="true">
          <div class="auth-hero__orb auth-hero__orb--1"></div>
          <div class="auth-hero__orb auth-hero__orb--2"></div>
          <div class="auth-hero__orb auth-hero__orb--3"></div>
          <div class="auth-hero__grid"></div>
        </div>
        <div class="auth-hero__content">
          <h2 class="auth-hero__headline">AI-driven<br><span class="auth-hero__accent">structured</span><br>research sessions.</h2>
          <p class="auth-hero__tagline">Let AI organize your knowledge.</p>
          <p class="auth-hero__cta">Sign in to start researching &rarr;</p>
        </div>
        <div class="auth-hero__bottom">
          <span class="auth-hero__bottom-logo">R</span>
          <span>Research</span>
        </div>
      </div>
      <div class="auth-form-panel">
        <div class="auth-form-inner">
          <div class="auth-form-card">
            <div class="auth-form-logo">R</div>
            <h1 class="auth-title">Welcome back</h1>
            <p class="auth-subtitle">Sign in to continue to your research</p>
            <form class="auth-form" @submit.prevent>
              <div v-if="showError" class="auth-error">Invalid email or password. Please try again.</div>
              <label class="auth-label">
                Email
                <input v-model="email" type="email" class="auth-input" placeholder="you@example.com" />
              </label>
              <label class="auth-label">
                Password
                <input v-model="password" type="password" class="auth-input" placeholder="Your password" />
              </label>
              <button type="submit" class="auth-button">Sign in</button>
            </form>
            <p v-if="showRegistrationLink" class="auth-footer">
              Don't have an account? <a href="/register" class="auth-link">Create one</a>
            </p>
          </div>
        </div>
      </div>
    </div>
  `,
})

// --- Register Page Stub ---
const RegisterPageStub = defineComponent({
  name: 'RegisterPageStub',
  props: {
    showError: { type: Boolean, default: false },
  },
  setup() {
    onMounted(ensureStyles)
    const name = ref('')
    const email = ref('')
    const password = ref('')
    return { name, email, password }
  },
  template: `
    <div class="auth-page">
      <div class="auth-hero">
        <div class="auth-hero__bg" aria-hidden="true">
          <div class="auth-hero__orb auth-hero__orb--1"></div>
          <div class="auth-hero__orb auth-hero__orb--2"></div>
          <div class="auth-hero__orb auth-hero__orb--3"></div>
          <div class="auth-hero__grid"></div>
        </div>
        <div class="auth-hero__content">
          <h2 class="auth-hero__headline">Organize your<br><span class="auth-hero__accent">knowledge</span><br>with AI precision.</h2>
          <p class="auth-hero__tagline">AI structures it. You explore it.</p>
          <p class="auth-hero__cta">Create your account and start &rarr;</p>
        </div>
        <div class="auth-hero__bottom">
          <span class="auth-hero__bottom-logo">R</span>
          <span>Research</span>
        </div>
      </div>
      <div class="auth-form-panel">
        <div class="auth-form-inner">
          <div class="auth-form-card">
            <div class="auth-form-logo">R</div>
            <h1 class="auth-title">Create account</h1>
            <p class="auth-subtitle">Get started with Research</p>
            <form class="auth-form" @submit.prevent>
              <div v-if="showError" class="auth-error">An account with this email already exists.</div>
              <label class="auth-label">
                Name
                <input v-model="name" type="text" class="auth-input" placeholder="Your name" />
              </label>
              <label class="auth-label">
                Email
                <input v-model="email" type="email" class="auth-input" placeholder="you@example.com" />
              </label>
              <label class="auth-label">
                Password
                <input v-model="password" type="password" class="auth-input" placeholder="At least 6 characters" />
              </label>
              <button type="submit" class="auth-button">Create account</button>
            </form>
            <p class="auth-footer">
              Already have an account? <a href="/login" class="auth-link">Sign in</a>
            </p>
          </div>
        </div>
      </div>
    </div>
  `,
})

// --- Meta ---
const meta: Meta = {
  title: 'Pages/Auth',
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  decorators: [
    (story) => ({
      components: { story },
      template: '<div style="background: var(--color-bg); color: var(--color-text); font-family: \'Outfit\', system-ui, sans-serif;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj

export const Login: Story = {
  render: (args) => ({
    components: { LoginPageStub },
    setup() { return { args } },
    template: '<LoginPageStub v-bind="args" />',
  }),
  args: { showError: false, showRegistrationLink: true },
}

export const LoginWithError: Story = {
  render: (args) => ({
    components: { LoginPageStub },
    setup() { return { args } },
    template: '<LoginPageStub v-bind="args" />',
  }),
  args: { showError: true, showRegistrationLink: true },
}

export const LoginNoRegistration: Story = {
  render: (args) => ({
    components: { LoginPageStub },
    setup() { return { args } },
    template: '<LoginPageStub v-bind="args" />',
  }),
  args: { showError: false, showRegistrationLink: false },
}

export const Register: Story = {
  render: (args) => ({
    components: { RegisterPageStub },
    setup() { return { args } },
    template: '<RegisterPageStub v-bind="args" />',
  }),
  args: { showError: false },
}

export const RegisterWithError: Story = {
  render: (args) => ({
    components: { RegisterPageStub },
    setup() { return { args } },
    template: '<RegisterPageStub v-bind="args" />',
  }),
  args: { showError: true },
}
