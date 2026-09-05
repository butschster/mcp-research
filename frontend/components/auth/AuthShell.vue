<script setup lang="ts">
import ThemeToggle from '~/components/ThemeToggle.vue'
import BrandLogo from '~/components/BrandLogo.vue'

withDefaults(defineProps<{
  title: string
  subtitle: string
  tagline?: string
}>(), {
  tagline: 'Work through questions with AI. Keep your sources, reasoning, and next steps in one place.',
})
</script>

<template>
  <div class="auth-page">
    <a href="#auth-main" class="skip-link">Skip to form</a>
    <header class="auth-header">
      <NuxtLink :to="{ name: 'index' }" class="auth-brand" aria-label="Dovod home"><BrandLogo /></NuxtLink>
      <div class="auth-header-actions"><span class="auth-header-note">A workspace for clear thinking.</span><ThemeToggle /></div>
    </header>
    <div class="auth-layout">
      <section class="auth-hero" aria-label="About Dovod">
        <div class="auth-hero-copy">
          <p class="auth-eyebrow">A question is a beginning.</p>
          <h2 class="auth-hero-headline"><slot name="headline">Think it<br><em>through.</em></slot></h2>
          <p class="auth-hero-tagline">{{ tagline }}</p>
        </div>
        <ol class="auth-process" aria-label="From question to decision">
          <li><span class="auth-process-number">01</span><div><h3>Ask a better question.</h3><p>Give your AI the context that matters.</p></div></li>
          <li><span class="auth-process-number">02</span><div><h3>Build on what you know.</h3><p>Bring sources and perspectives together.</p></div></li>
          <li><span class="auth-process-number">03</span><div><h3>Decide what comes next.</h3><p>Keep the reasoning behind each step.</p></div></li>
        </ol>
      </section>
      <main id="auth-main" class="auth-form-panel" tabindex="-1">
        <div class="auth-form-inner">
          <div class="auth-form-heading">
            <p class="auth-eyebrow">Your space to think</p>
            <h1 class="auth-title">{{ title }}</h1>
            <p class="auth-subtitle">{{ subtitle }}</p>
          </div>
          <slot />
          <slot name="footer" />
        </div>
      </main>
    </div>
    <footer class="auth-page-footer"><span>Good questions. Well-founded decisions.</span><span>Dovod</span></footer>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  flex-direction: column;
  min-height: 100dvh;
  width: 100%;
  padding: 2.5rem clamp(1.5rem, 5vw, 5rem) 1.75rem;
  background: var(--color-bg);
  color: var(--color-text);
}
.auth-header, .auth-page-footer { display: flex; align-items: center; justify-content: space-between; gap: var(--space-6); }
.auth-header { width: 100%; max-width: 82rem; margin: 0 auto; }
.auth-brand { display: inline-flex; width: 8rem; color: var(--color-primary); }
.auth-header-actions { display: flex; align-items: center; gap: var(--space-5); }
.auth-header-note { font-size: var(--type-xs); color: var(--color-text-muted); }
.auth-layout { display: grid; grid-template-columns: minmax(0, 1.15fr) minmax(0, 1fr); width: 100%; max-width: 76rem; margin: auto; flex: 1; align-items: center; gap: clamp(3rem, 7vw, 7rem); padding: 4rem 0; }
.auth-hero { padding-right: var(--space-8); }
.auth-eyebrow { font-size: var(--type-xs); font-weight: var(--weight-medium); color: var(--color-text-muted); margin-bottom: var(--space-5); }
.auth-hero-headline { font-family: var(--font-display); font-size: clamp(3.5rem, 5.5vw, 5.25rem); font-weight: 800; line-height: .99; letter-spacing: -.055em; text-wrap: balance; }
.auth-hero-headline :deep(em) { font-style: normal; color: var(--color-hero-accent); }
.auth-hero-tagline { font-size: var(--type-base); line-height: 1.7; max-width: 25rem; margin-top: var(--space-6); color: var(--color-text-muted); }
.auth-process { list-style: none; display: grid; gap: var(--space-5); margin-top: var(--space-10); max-width: 25rem; }
.auth-process li { display: grid; grid-template-columns: 2rem 1fr; gap: var(--space-4); border-top: 1px solid var(--color-border); padding-top: var(--space-4); }
.auth-process-number { font-family: 'JetBrains Mono', monospace; font-size: var(--type-2xs); color: var(--color-text-muted); padding-top: 3px; }
.auth-process h3 { font-size: var(--type-sm); font-weight: var(--weight-medium); letter-spacing: -.015em; }
.auth-process p { font-size: var(--type-xs); color: var(--color-text-muted); margin-top: 3px; }
.auth-form-panel { border-left: 1px solid var(--color-border); padding-left: clamp(2rem, 4vw, 4rem); }
.auth-form-inner { width: 100%; max-width: 25rem; margin: 0 auto; }
.auth-form-heading { margin-bottom: var(--space-8); }
.auth-form-heading .auth-eyebrow { margin-bottom: var(--space-3); }
.auth-title { font-family: var(--font-display); font-size: clamp(1.875rem, 2.6vw, 2.5rem); font-weight: 800; line-height: 1.12; letter-spacing: -.04em; text-wrap: balance; }
.auth-subtitle { color: var(--color-text-muted); font-size: var(--type-sm); line-height: 1.65; margin-top: var(--space-3); }
.auth-page :deep(.auth-form) { display: flex; flex-direction: column; gap: var(--space-5); width: 100%; }
.auth-page :deep(.auth-field) { display: flex; flex-direction: column; gap: var(--space-2); }
.auth-page :deep(.auth-label) { display: flex; flex-direction: column; gap: var(--space-2); font-size: var(--type-xs); font-weight: var(--weight-medium); }
.auth-page :deep(.auth-input) { width: 100%; min-height: 3.25rem; padding: var(--space-3) var(--space-4); background: var(--color-surface); color: var(--color-text); border: 1px solid var(--color-border-strong); border-radius: var(--radius-sm); font: inherit; font-size: 1rem; transition: border-color var(--transition-fast), box-shadow var(--transition-fast); }
.auth-page :deep(.auth-input:focus) { outline: none; border-color: var(--color-primary); box-shadow: var(--shadow-focus); }
.auth-page :deep(.auth-input::placeholder) { color: var(--color-text-muted); opacity: 1; }
.auth-page :deep(.auth-password .auth-input) { padding-right: 3.5rem; }
.auth-page :deep(.auth-hint) { font-size: var(--type-xs); line-height: 1.5; color: var(--color-text-muted); }
.auth-page :deep(.auth-button) { display: inline-flex; align-self: flex-start; justify-content: center; align-items: center; gap: var(--space-4); min-height: 3rem; margin-top: var(--space-2); padding: var(--space-3) var(--space-6); border: 1px solid transparent; border-radius: var(--radius-sm); background: var(--color-primary); color: var(--color-on-primary); font: inherit; font-size: var(--type-sm); font-weight: var(--weight-medium); cursor: pointer; transition: background var(--transition-fast), transform var(--transition-fast); }
.auth-page :deep(.auth-button:hover:not(:disabled)) { background: var(--color-primary-hover); }
.auth-page :deep(.auth-button:active:not(:disabled)) { transform: translateY(1px); }
.auth-page :deep(.auth-button:disabled) { background: var(--color-primary); opacity: .65; cursor: wait; }
.auth-page :deep(.auth-error) { padding: var(--space-3) var(--space-4); border: 1px solid rgba(var(--color-error-rgb), .3); border-radius: var(--radius-sm); background: rgba(var(--color-error-rgb), .08); color: var(--color-error); font-size: var(--type-sm); line-height: 1.6; overflow-wrap: anywhere; }
.auth-page :deep(.auth-footer) { padding-top: var(--space-6); margin-top: var(--space-6); border-top: 1px solid var(--color-border); font-size: var(--type-xs); color: var(--color-text-muted); line-height: 1.7; }
.auth-page :deep(.auth-link) { color: var(--color-primary); font-weight: var(--weight-semibold); text-decoration: underline; text-underline-offset: 3px; }
.auth-page :deep(.auth-link:hover) { color: var(--color-primary-hover); }
.auth-page-footer { width: 100%; max-width: 82rem; margin: 0 auto; padding-top: var(--space-5); border-top: 1px solid var(--color-border); font-size: var(--type-2xs); color: var(--color-text-muted); }
@media (max-width: 1023px) {
  .auth-layout { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: var(--space-10); padding: var(--space-12) 0; }
  .auth-hero { padding-right: 0; }
  .auth-hero-headline { font-size: 3.5rem; }
  .auth-form-panel { padding-left: var(--space-8); }
}
@media (max-width: 767px) {
  .auth-page { padding: 1.75rem 1.5rem 1.5rem; }
  .auth-header-note, .auth-hero { display: none; }
  .auth-brand { width: 7rem; }
  .auth-layout { display: flex; padding: 3.5rem 0; }
  .auth-form-panel { border: 0; padding: 0; width: 100%; }
  .auth-title { font-size: 2.25rem; }
  .auth-page-footer { font-size: var(--type-2xs); }
  .auth-page-footer span:last-child { display: none; }
}
</style>
