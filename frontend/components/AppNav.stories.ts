import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, ref, computed } from 'vue'

const AppNavStub = defineComponent({
  name: 'AppNavStub',
  props: {
    authEnabled: { type: Boolean, default: false },
    isAuthenticated: { type: Boolean, default: false },
    userName: { type: String, default: '' },
    userEmail: { type: String, default: '' },
    menuOpen: { type: Boolean, default: false },
  },
  setup(props) {
    const userMenuOpen = ref(props.menuOpen)
    const initial = computed(() => (props.userName || props.userEmail || '?')[0].toUpperCase())
    const displayName = computed(() => props.userName || props.userEmail)
    return { userMenuOpen, initial, displayName }
  },
  template: `
    <nav class="app-nav">
      <div class="container nav-inner" style="display:flex;align-items:center;justify-content:space-between;">
        <a href="/" class="logo" style="font-size:var(--type-lg);font-weight:700;color:var(--color-text);letter-spacing:-0.025em;text-decoration:none;">Research</a>
        <div class="nav-right" style="display:flex;align-items:center;gap:var(--space-3);">
          <!-- Search trigger stub -->
          <button class="search-trigger" style="display:flex;align-items:center;gap:var(--space-2);padding:var(--space-2) var(--space-3);background:var(--color-surface);border:1px solid var(--color-border);border-radius:var(--radius-sm);color:var(--color-text-muted);cursor:pointer;font-family:inherit;font-size:var(--type-sm);min-width:200px;">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
            <span style="flex:1;text-align:left;">Search...</span>
            <kbd style="font-size:0.625rem;padding:1px 5px;background:var(--color-surface-hover);border:1px solid var(--color-border);border-radius:3px;color:var(--color-text-muted);font-family:inherit;line-height:1.4;">⌘K</kbd>
          </button>

          <!-- Auth: user menu -->
          <template v-if="authEnabled && isAuthenticated">
            <div class="user-menu" style="position:relative;">
              <button class="user-menu-trigger" @click="userMenuOpen = !userMenuOpen" style="display:flex;align-items:center;gap:var(--space-2);background:none;border:1px solid transparent;color:var(--color-text-muted);cursor:pointer;font-family:inherit;font-size:var(--type-sm);padding:var(--space-1) var(--space-2);border-radius:var(--radius-sm);">
                <span style="width:22px;height:22px;border-radius:50%;background:var(--color-primary-muted);color:var(--color-primary);display:flex;align-items:center;justify-content:center;font-size:0.6875rem;font-weight:600;flex-shrink:0;">{{ initial }}</span>
                <span class="user-name" style="max-width:120px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ displayName }}</span>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
              </button>
              <div v-if="userMenuOpen" class="user-dropdown" style="position:absolute;right:0;top:calc(100% + var(--space-1));min-width:180px;background:var(--color-surface);border:1px solid var(--color-border-strong);border-radius:var(--radius);padding:var(--space-1);box-shadow:0 8px 32px rgba(0,0,0,0.3);z-index:100;">
                <a href="/settings" style="display:flex;align-items:center;gap:var(--space-2);width:100%;padding:var(--space-2) var(--space-3);border:none;background:none;color:var(--color-text-muted);font-size:var(--type-sm);cursor:pointer;border-radius:var(--radius-sm);text-decoration:none;">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
                  Settings
                </a>
                <div style="height:1px;background:var(--color-border);margin:var(--space-1) var(--space-2);"></div>
                <button style="display:flex;align-items:center;gap:var(--space-2);width:100%;padding:var(--space-2) var(--space-3);border:none;background:none;color:var(--color-text-muted);font-family:inherit;font-size:var(--type-sm);cursor:pointer;border-radius:var(--radius-sm);">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
                  Sign out
                </button>
              </div>
            </div>
          </template>

          <!-- No auth: read-only badge -->
          <span v-else-if="!authEnabled" class="readonly-badge">Read-only</span>
        </div>
      </div>
    </nav>
  `,
})

const meta: Meta<typeof AppNavStub> = {
  title: 'Layout/AppNav',
  component: AppNavStub,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
}
export default meta
type Story = StoryObj<typeof AppNavStub>

export const ReadOnly: Story = {
  args: {
    authEnabled: false,
  },
}

export const Authenticated: Story = {
  args: {
    authEnabled: true,
    isAuthenticated: true,
    userName: 'John Doe',
    userEmail: 'john@example.com',
  },
}

export const AuthenticatedMenuOpen: Story = {
  args: {
    authEnabled: true,
    isAuthenticated: true,
    userName: 'John Doe',
    userEmail: 'john@example.com',
    menuOpen: true,
  },
}

export const LongUserName: Story = {
  args: {
    authEnabled: true,
    isAuthenticated: true,
    userName: 'Alexander Konstantinovich',
    userEmail: 'alexander.konstantinovich@example.com',
    menuOpen: true,
  },
}

export const EmailOnly: Story = {
  args: {
    authEnabled: true,
    isAuthenticated: true,
    userName: '',
    userEmail: 'user@research.dev',
  },
}
