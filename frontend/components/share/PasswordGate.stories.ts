import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import SharePasswordGate from './PasswordGate.vue'

/**
 * The one field between a visitor and a protected link.
 *
 * This is the only screen in the feature that stands between the reader and the
 * content, so it says who to ask rather than what went wrong: the visitor has no
 * account, no password reset and nobody to appeal to except the person who sent
 * them the link.
 *
 * Wrong and throttled are separate messages because they call for different
 * actions — check the password, or stop typing for a minute. Telling somebody
 * who has been rate-limited that their password is wrong makes them try harder,
 * which is exactly the wrong thing.
 *
 * The field is never cleared on a wrong answer. The failure mode here is
 * retyping a long password out of a chat message, not somebody reading it over a
 * shoulder.
 */
const meta: Meta<typeof SharePasswordGate> = {
  title: 'Share/PasswordGate',
  component: SharePasswordGate,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    busy: { control: 'boolean' },
    error: { control: 'select', options: ['', 'wrong', 'throttled'] },
    onSubmit: { action: 'submit' },
  },
}
export default meta
type Story = StoryObj<typeof SharePasswordGate>

/** As it arrives. The field takes focus on mount, and Open stays disabled until
 *  something is typed — there is nothing to submit before then. */
export const Idle: Story = {
  args: { busy: false, error: '' },
}

/** In flight. The field and the button both lock, so a second Enter cannot
 *  spend one of the few tries the server allows. */
export const Busy: Story = {
  args: { busy: true, error: '' },
}

/** The password did not match. Said under the field, as an alert, and the field
 *  keeps what was typed so a single wrong character can be fixed rather than
 *  retyped. */
export const WrongPassword: Story = {
  args: { busy: false, error: 'wrong' },
}

/** Too many tries. A different message from a wrong password, because the
 *  action it asks for is different: wait, then try the same password again. */
export const Throttled: Story = {
  args: { busy: false, error: 'throttled' },
}

/** Wired up: `hunter2` opens, anything else is refused, and a fourth attempt
 *  trips the throttle — the same three answers the server gives. */
export const Interactive: Story = {
  render: () => ({
    components: { SharePasswordGate },
    setup() {
      const busy = ref(false)
      const error = ref<'' | 'wrong' | 'throttled'>('')
      const opened = ref(false)
      const tries = ref(0)

      function submit(password: string) {
        busy.value = true
        error.value = ''
        setTimeout(() => {
          busy.value = false
          tries.value += 1
          if (password === 'hunter2') {
            opened.value = true
            return
          }
          error.value = tries.value >= 4 ? 'throttled' : 'wrong'
        }, 500)
      }

      return { busy, error, opened, submit }
    },
    template: `
      <div v-if="opened" style="padding: var(--space-6);">
        <p class="card-meta">Unlocked. The shell would now render the research.</p>
      </div>
      <SharePasswordGate v-else :busy="busy" :error="error" @submit="submit" />
    `,
  }),
}
