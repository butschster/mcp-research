import type { Meta, StoryObj } from '@storybook/vue3'
import AuthShell from './AuthShell.vue'
import AuthPasswordInput from './AuthPasswordInput.vue'

const meta: Meta<typeof AuthShell> = {
  title: 'Auth/AuthShell',
  component: AuthShell,
  parameters: { layout: 'fullscreen' },
}
export default meta
type Story = StoryObj<typeof AuthShell>

function preview({ register = false, error = '', busy = false, long = false } = {}) {
  return {
    components: { AuthShell, AuthPasswordInput },
    setup: () => ({ password: ref('') }),
    template: `
      <AuthShell
        title="${long ? 'Welcome back to your Dovod workspace' : register ? 'Make space for your ideas.' : 'Welcome back.'}"
        subtitle="${register ? 'Create your Dovod account to get started.' : 'Sign in to your Dovod workspace.'}"
      >
        ${register ? '<template #headline>Start with<br><em>a question.</em></template>' : ''}
        <form class="auth-form" aria-busy="${busy}" @submit.prevent>
          ${error ? `<div class="auth-error" role="alert">${error}</div>` : ''}
          ${register ? '<div class="auth-field"><label for="story-name" class="auth-label">Your name</label><input id="story-name" class="auth-input" autocomplete="name" placeholder="Your name" required /></div>' : ''}
          <div class="auth-field"><label for="story-email" class="auth-label">Email</label><input id="story-email" type="email" class="auth-input" autocomplete="email" placeholder="you@example.com" required /></div>
          <div class="auth-field">
            <label for="story-password" class="auth-label">Password</label>
            <AuthPasswordInput id="story-password" v-model="password" autocomplete="${register ? 'new-password' : 'current-password'}" ${register ? ':minlength="6" describedby="story-password-hint"' : ''} />
            ${register ? '<p id="story-password-hint" class="auth-hint">Use at least 6 characters.</p>' : ''}
          </div>
          <button class="auth-button" type="submit" ${busy ? 'disabled' : ''}><span>${busy ? 'Signing in…' : register ? 'Create account' : 'Sign in'}</span><span aria-hidden="true">${busy ? '···' : '→'}</span></button>
        </form>
        <template #footer><p class="auth-footer">${register ? 'Already have an account? <a href="/login" class="auth-link">Sign in</a>' : 'New to Dovod? <a href="/register" class="auth-link">Create an account</a>'}</p></template>
      </AuthShell>
    `,
  }
}

export const SignIn: Story = { render: () => preview() }
export const Register: Story = { render: () => preview({ register: true }) }
export const WithError: Story = { render: () => preview({ error: 'Invalid email or password.' }) }
export const Submitting: Story = { render: () => preview({ busy: true }) }
export const LongCopy: Story = {
  render: () => preview({ long: true, error: 'Could not reach this Dovod workspace. Check your connection and try again. Your email and password are still in the form.' }),
}
