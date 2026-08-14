<template>
  <Teleport to="body">
    <div
      class="toast-host"
      role="region"
      aria-label="Notifications"
      @mouseenter="hold"
      @mouseleave="resume"
      @focusin="hold"
      @focusout="resume"
    >
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="['toast', `toast-${toast.variant}`]"
          :role="toast.variant === 'error' ? 'alert' : 'status'"
          aria-live="polite"
        >
          <span class="toast-icon" aria-hidden="true">
            <svg v-if="toast.variant === 'success'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            <svg v-else-if="toast.variant === 'error'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="13"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="11" x2="12" y2="16"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
          </span>

          <div class="toast-body">
            <p v-if="toast.title" class="toast-title">{{ toast.title }}</p>
            <p class="toast-message">{{ toast.message }}</p>
            <button v-if="toast.action" class="toast-action" @click="run(toast)">
              {{ toast.action.label }}
            </button>
          </div>

          <button class="toast-close" aria-label="Dismiss" @click="dismiss(toast.id)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { Toast } from '~/composables/useToasts'

/**
 * Renders every notification the app raises, in one corner.
 *
 * Mounted once, in the app shell. Nothing else should render toasts: two hosts
 * would stack two corners of the same messages.
 *
 * A countdown pauses while the pointer or the keyboard is inside, because a
 * message that vanishes as you reach for its button is worse than no button.
 */
const { toasts, dismiss, hold, resume } = useToasts()

function run(toast: Toast) {
  const keepOpen = toast.action?.onClick() === false
  if (!keepOpen) dismiss(toast.id)
}
</script>

<style scoped>
.toast-host {
  position: fixed;
  right: var(--space-5);
  bottom: var(--space-5);
  z-index: var(--z-toast);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  width: min(380px, calc(100vw - var(--space-8)));
  /* The host spans a column the toasts sit in; only the toasts themselves are
     clickable, so the page underneath keeps working. */
  pointer-events: none;
}
.toast {
  pointer-events: auto;
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-left: 3px solid var(--color-border-strong);
  border-radius: var(--radius);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.35);
}
/* The left edge carries the variant, so it survives a screenshot in greyscale
   and does not depend on the icon being noticed. */
.toast-success { border-left-color: var(--color-success); }
.toast-error { border-left-color: var(--color-error); }
.toast-info { border-left-color: var(--color-primary); }

.toast-icon {
  flex-shrink: 0;
  margin-top: 1px;
  display: inline-flex;
}
.toast-success .toast-icon { color: var(--color-success); }
.toast-error .toast-icon { color: var(--color-error); }
.toast-info .toast-icon { color: var(--color-primary); }

.toast-body { flex: 1; min-width: 0; }
.toast-title {
  margin: 0;
  font-size: var(--type-sm);
  font-weight: 600;
  color: var(--color-text);
}
.toast-message {
  margin: 0;
  font-size: var(--type-xs);
  line-height: 1.5;
  color: var(--color-text-muted);
  /* A filename has no spaces to break at. */
  overflow-wrap: anywhere;
}
.toast-action {
  margin-top: var(--space-2);
  padding: 0;
  background: none;
  border: none;
  font: inherit;
  font-size: var(--type-xs);
  color: var(--color-primary);
  text-decoration: underline;
  cursor: pointer;
}
.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  padding: 0;
  color: var(--color-text-muted);
  cursor: pointer;
  line-height: 1;
}
.toast-close:hover { color: var(--color-text); }

.toast-enter-active,
.toast-leave-active,
.toast-move {
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
.toast-leave-active { position: absolute; }

@media (prefers-reduced-motion: reduce) {
  .toast-enter-active,
  .toast-leave-active,
  .toast-move { transition: none; }
}
@media (max-width: 480px) {
  .toast-host {
    right: var(--space-3);
    left: var(--space-3);
    bottom: var(--space-3);
    width: auto;
  }
}
</style>
