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
          <slot name="headline" />
        </h2>
        <p :class="['auth-hero__tagline', { 'is-visible': ready }]">{{ tagline }}</p>
        <p :class="['auth-hero__cta', { 'is-visible': ready }]">{{ cta }}</p>
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

          <h1 class="auth-title">{{ title }}</h1>
          <p class="auth-subtitle">{{ subtitle }}</p>

          <slot />
          <slot name="footer" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * The two-panel frame both auth pages are drawn in.
 *
 * It exists because `login.vue` and `register.vue` carried 361 byte-identical
 * lines of scoped CSS each — not "substantially similar", identical — plus a
 * third transcription inside `AuthPages.stories.ts` where the values had been
 * rewritten as literals instead of tokens. Three copies of one design, two of
 * which could drift silently and one of which already had.
 *
 * The pages keep their own form and their own submit; what they hand over is
 * the frame, the hero, and the mount animation, which is the part that was
 * duplicated and the part neither page had a reason to own.
 */
defineProps<{
  /** Hero: the line under the headline, and the line that points at the form. */
  tagline: string
  cta: string
  /** Form card: its heading and the sentence under it. */
  title: string
  subtitle: string
}>()

/* The entrance is staged rather than immediate: the panel is composited before
   anything moves, so the first frame is not a jump. */
const ready = ref(false)
onMounted(() => requestAnimationFrame(() => { ready.value = true }))
</script>

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
0%,
100% { transform: translate(0, 0) scale(1); }
33% { transform: translate(-30px, 20px) scale(1.05); }
66% { transform: translate(20px, -15px) scale(0.95); }
}
@keyframes orbDrift2 {
0%,
100% { transform: translate(0, 0) scale(1); }
33% { transform: translate(25px, -20px) scale(1.08); }
66% { transform: translate(-15px, 25px) scale(0.92); }
}
@keyframes orbDrift3 {
0%,
100% { transform: translate(0, 0) scale(1); }
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

:slotted(.auth-hero__accent) {
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
:slotted(.auth-form-panel) {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
  padding: var(--space-6) var(--space-6);
}
@media (min-width: 640px) {
:slotted(.auth-form-panel) { padding: var(--space-12) var(--space-12); }
}

:slotted(.auth-form-inner) {
  width: 100%;
  max-width: 380px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 480px;
}

:slotted(.auth-form-card) {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  opacity: 0;
  transform: translateY(16px);
  transition: opacity 0.6s ease, transform 0.6s ease;
  transition-delay: 0.15s;
}
:slotted(.auth-form-card.is-visible) {
  opacity: 1;
  transform: translateY(0);
}

:slotted(.auth-form-logo) {
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

:slotted(.auth-form) {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  width: 100%;
}

:slotted(.auth-label) {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

:slotted(.auth-input) {
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius);
  font-size: var(--type-sm);
  background: var(--color-surface);
  color: var(--color-text);
  font-family: inherit;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
:slotted(.auth-input:focus) {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: var(--shadow-focus);
}
:slotted(.auth-input::placeholder) {
  color: var(--color-text-muted);
  opacity: 0.5;
}

:slotted(.auth-button) {
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
:slotted(.auth-button:hover) {
  transform: translateY(-1px);
  box-shadow: 0 0 32px -4px rgba(108, 197, 224, 0.4);
}
:slotted(.auth-button:active) {
  transform: translateY(0);
}
:slotted(.auth-button:disabled) {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

@keyframes btnShift {
0%,
100% { background-position: 0% 50%; }
50% { background-position: 100% 50%; }
}

:slotted(.auth-error) {
  padding: var(--space-3) var(--space-4);
  background: rgba(239, 107, 107, 0.08);
  border: 1px solid rgba(239, 107, 107, 0.2);
  color: var(--color-error);
  border-radius: var(--radius);
  font-size: var(--type-sm);
}

:slotted(.auth-footer) {
  margin-top: var(--space-6);
  text-align: center;
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}

:slotted(.auth-link) {
  color: var(--color-primary);
  text-decoration: none;
  font-weight: var(--weight-medium);
}
:slotted(.auth-link:hover) {
  text-decoration: underline;
}

/* ── Reduced motion ─────────────────────────────────────── */
@media (prefers-reduced-motion: reduce) {
.auth-hero__orb,
:slotted(.auth-button) { animation: none; }
:slotted(.auth-button:hover) { transform: none; }
.auth-hero__headline,
.auth-hero__tagline,
.auth-hero__cta,
.auth-hero__bottom,
:slotted(.auth-form-card) {
    opacity: 1;
    transform: none;
    transition: none;
  }
}

/* ── Mobile: single column ──────────────────────────────── */
@media (max-width: 1023px) {
.auth-page { flex-direction: column; }
:slotted(.auth-form-panel) { min-height: 100dvh; }
}
</style>
