/** Canvas and Mermaid need resolved colors, not CSS var() strings. Read once
 * per theme change or render, never once per node in an animation frame. */
export function readThemeColors() {
  const style = getComputedStyle(document.documentElement)
  const get = (role: string) => style.getPropertyValue(`--${role}`).trim()
  return {
    background: get('color-bg'),
    surface: get('color-surface'),
    raised: get('color-surface-raised'),
    recessed: get('color-bg-deep'),
    text: get('color-text'),
    textRgb: get('color-text-rgb'),
    muted: get('color-text-muted'),
    border: get('color-border-strong'),
    primary: get('color-primary'),
    primaryRgb: get('color-primary-rgb'),
    success: get('color-success'),
    warning: get('color-warning'),
    warningRgb: get('color-warning-rgb'),
    error: get('color-error'),
    info: get('color-info'),
    violet: get('hue-5'),
    violetRgb: get('hue-5-rgb'),
    orange: get('hue-6'),
  }
}
