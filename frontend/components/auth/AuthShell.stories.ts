import type { Meta, StoryObj } from '@storybook/vue3'
import AuthShell from './AuthShell.vue'

/**
 * The frame both auth pages are drawn in.
 *
 * It replaces `AuthPages.stories.ts`, which documented the two pages by
 * injecting a hand-transcribed copy of their stylesheet into the document —
 * a third copy of a design that already existed twice, written with raw values
 * instead of tokens, and therefore the copy most likely to be wrong. These
 * stories mount the real component, so what the catalogue shows is what ships.
 *
 * The pages own their form; the shell owns the hero, the card and the mount
 * animation. Everything passed through the default slot is *slotted* content,
 * which is why the form styles in the component are written with `:slotted()`
 * — a scoped rule does not otherwise reach markup the parent wrote.
 */
const meta: Meta<typeof AuthShell> = {
  title: 'Auth/AuthShell',
  component: AuthShell,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component:
          'Two panels: a hero that animates in on mount, and a form card. '
          + 'The page supplies the fields and the footer link.',
      },
    },
  },
}
export default meta

type Story = StoryObj<typeof AuthShell>

const form = (submit: string) => `
  <form class="auth-form" @submit.prevent>
    <label class="auth-label">
      Email
      <input type="email" class="auth-input" placeholder="you@example.com" />
    </label>
    <label class="auth-label">
      Password
      <input type="password" class="auth-input" placeholder="Your password" />
    </label>
    <button type="submit" class="auth-button">${submit}</button>
  </form>
`

/** Signing in — the shape the majority of visits take. */
export const SignIn: Story = {
  render: () => ({
    components: { AuthShell },
    template: `
      <AuthShell
        tagline="Let AI organize your knowledge."
        cta="Sign in to start researching →"
        title="Welcome back"
        subtitle="Sign in to continue to your research"
      >
        <template #headline>AI-driven<br><span class="auth-hero__accent">structured</span><br>research sessions.</template>
        ${form('Sign in')}
        <template #footer>
          <p class="auth-footer">Don't have an account? <a href="#" class="auth-link">Create one</a></p>
        </template>
      </AuthShell>
    `,
  }),
}

/** Registering, which carries a third field and different copy throughout. */
export const Register: Story = {
  render: () => ({
    components: { AuthShell },
    template: `
      <AuthShell
        tagline="Structured research, kept for you."
        cta="Create an account to begin →"
        title="Create your account"
        subtitle="Start organizing your research"
      >
        <template #headline>AI-driven<br><span class="auth-hero__accent">structured</span><br>research sessions.</template>
        <form class="auth-form" @submit.prevent>
          <label class="auth-label">Name<input class="auth-input" placeholder="Ada Lovelace" /></label>
          <label class="auth-label">Email<input type="email" class="auth-input" placeholder="you@example.com" /></label>
          <label class="auth-label">Password<input type="password" class="auth-input" placeholder="At least 8 characters" /></label>
          <button type="submit" class="auth-button">Create account</button>
        </form>
        <template #footer>
          <p class="auth-footer">Already have an account? <a href="#" class="auth-link">Sign in</a></p>
        </template>
      </AuthShell>
    `,
  }),
}

/**
 * A refused sign-in. The error carries `role="alert"`, which it did not until
 * recently — every other inline error in the product had it, and this was the
 * one place a failure announced nothing at all, on a screen where there is
 * nowhere else to go.
 */
export const WithError: Story = {
  render: () => ({
    components: { AuthShell },
    template: `
      <AuthShell
        tagline="Let AI organize your knowledge."
        cta="Sign in to start researching →"
        title="Welcome back"
        subtitle="Sign in to continue to your research"
      >
        <template #headline>AI-driven<br><span class="auth-hero__accent">structured</span><br>research sessions.</template>
        <form class="auth-form" @submit.prevent>
          <div class="auth-error" role="alert">Invalid email or password</div>
          <label class="auth-label">Email<input type="email" class="auth-input" value="ada@example.com" /></label>
          <label class="auth-label">Password<input type="password" class="auth-input" /></label>
          <button type="submit" class="auth-button">Sign in</button>
        </form>
      </AuthShell>
    `,
  }),
}

/**
 * Mid-submit. The button is the only thing that changes, which is deliberate:
 * disabling the fields as well would move focus and lose an in-flight
 * correction.
 */
export const Submitting: Story = {
  render: () => ({
    components: { AuthShell },
    template: `
      <AuthShell
        tagline="Let AI organize your knowledge."
        cta="Sign in to start researching →"
        title="Welcome back"
        subtitle="Sign in to continue to your research"
      >
        <template #headline>AI-driven<br><span class="auth-hero__accent">structured</span><br>research sessions.</template>
        <form class="auth-form" @submit.prevent>
          <label class="auth-label">Email<input type="email" class="auth-input" value="ada@example.com" /></label>
          <label class="auth-label">Password<input type="password" class="auth-input" value="secret" /></label>
          <button type="submit" class="auth-button" disabled>Signing in...</button>
        </form>
      </AuthShell>
    `,
  }),
}

/**
 * A very long hero line and a long error, which is where a two-panel layout
 * with a fixed hero usually comes apart.
 */
export const LongCopy: Story = {
  render: () => ({
    components: { AuthShell },
    template: `
      <AuthShell
        tagline="Let an agent organize the knowledge you have been meaning to write down for months."
        cta="Sign in to pick up where the agent left off →"
        title="Welcome back to your research workspace"
        subtitle="Sign in to continue to the sessions, entries and roadmaps your agent has been keeping"
      >
        <template #headline>AI-driven<br><span class="auth-hero__accent">structured</span><br>research sessions, kept in order.</template>
        <form class="auth-form" @submit.prevent>
          <div class="auth-error" role="alert">The server refused the sign-in because the account has been locked after too many attempts. Try again in fifteen minutes, or ask an owner to reset it.</div>
          <label class="auth-label">Email<input type="email" class="auth-input" /></label>
          <button type="submit" class="auth-button">Sign in</button>
        </form>
      </AuthShell>
    `,
  }),
}
