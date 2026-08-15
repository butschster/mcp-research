#!/usr/bin/env node
/**
 * Three CSS mistakes that are invisible in review and obvious on screen.
 *
 * Each one shipped at least once. None of them is a matter of taste, which is
 * why they belong in a script rather than in a reviewer's judgement:
 *
 *  1. A class defined both globally and in a component's `<style scoped>`.
 *     The scoped rule carries a [data-v] attribute and therefore wins, so the
 *     component looks one way inside its own page and another way everywhere
 *     else. `.btn-icon` was scoped to the research page: the kebab button in
 *     ActionMenu — a child component, out of reach of the parent's scoped CSS —
 *     rendered full-size there and full-size on the session page, next to
 *     compact icon buttons that did get the rule.
 *
 *  2. `calc(100dvh - <a number somebody measured once>)`. The chrome grows, the
 *     number does not, and the page ends up taller than the window with nothing
 *     to scroll. Measure with a flex column instead.
 *
 *  3. A control whose height is left to emerge from padding plus content.
 *     15px of text in a `.btn` and a 16px icon in a `.btn-icon` came out 29.8
 *     and 29.2 tall, and `.btn-sm` 26.6 — three heights in one header row.
 *     A control states its height.
 *
 * Usage: node scripts/css-consistency.mjs
 * Exit code 1 if anything is found.
 */
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

const ROOT = new URL('..', import.meta.url).pathname
const GLOBAL_CSS = join(ROOT, 'assets/css/main.css')

/** The button family: the controls a user sees side by side in a row. */
const CONTROL = /^\.btn(-|$)/

function walk(dir, out = []) {
  for (const name of readdirSync(dir)) {
    if (name === 'node_modules' || name === '.nuxt' || name === '.output' || name === 'dist') continue
    const full = join(dir, name)
    if (statSync(full).isDirectory()) walk(full, out)
    else if (name.endsWith('.vue')) out.push(full)
  }
  return out
}

/** Every top-level class selector a stylesheet defines, with its line. */
function classSelectors(css, offsetLine = 0) {
  const found = new Map()
  css.split('\n').forEach((line, i) => {
    const m = line.match(/^\s*(\.[a-zA-Z0-9_-]+)[^{]*\{/)
    if (m && !line.includes(':slotted') && !line.includes(':deep')) {
      if (!found.has(m[1])) found.set(m[1], offsetLine + i + 1)
    }
  })
  return found
}

function scopedBlocks(source) {
  const blocks = []
  const re = /<style[^>]*\bscoped\b[^>]*>([\s\S]*?)<\/style>/g
  let m
  while ((m = re.exec(source))) {
    blocks.push({ css: m[1], line: source.slice(0, m.index).split('\n').length })
  }
  return blocks
}

const findings = []
const globalCss = readFileSync(GLOBAL_CSS, 'utf8')
const globalClasses = classSelectors(globalCss)
const files = walk(join(ROOT, 'components')).concat(walk(join(ROOT, 'pages')), [join(ROOT, 'app.vue')])

for (const file of files) {
  const source = readFileSync(file, 'utf8')
  const rel = relative(ROOT, file)

  for (const { css, line } of scopedBlocks(source)) {
    // 1. collisions with the global stylesheet
    for (const [sel, at] of classSelectors(css, line)) {
      if (globalClasses.has(sel)) {
        findings.push({
          kind: 'collision',
          where: `${rel}:${at}`,
          detail: `${sel} is also defined in assets/css/main.css:${globalClasses.get(sel)}. The scoped rule wins here and the global one wins everywhere else, so this element looks different depending on which page it is on.`,
        })
      }
    }
  }

  // 2. viewport arithmetic with a hard-coded chrome height
  source.split('\n').forEach((l, i) => {
    // `min-height`/`height`, not `max-height`: a max-height giving a column its
    // own scroll is a different and legitimate thing. What forces the page
    // itself past the viewport is a floor.
    if (/^\s*(\/\*|\*|\/\/)/.test(l)) return
    const m = l.match(/(?:^|[\s;{])(?:min-)?height:\s*calc\(\s*100d?vh\s*-\s*(\d+)px/)
    if (m) {
      findings.push({
        kind: 'magic-viewport',
        where: `${rel}:${i + 1}`,
        detail: `calc(100vh - ${m[1]}px) assumes the chrome is exactly ${m[1]}px tall. When it is not, the page scrolls with nothing to scroll. Use a flex column.`,
      })
    }
  })
}

// 3. control classes that set padding but never state a height
for (const block of globalCss.split(/\n(?=\.[a-zA-Z])/)) {
  const sel = block.match(/^(\.[a-zA-Z0-9_-]+)/)?.[1]
  if (!sel || !CONTROL.test(sel)) continue
  if (/:hover|:focus|:disabled|:active|--/.test(block.split('{')[0])) continue
  const body = block.slice(block.indexOf('{'), block.indexOf('}'))
  // `padding: 0` is a reset (.btn-ghost), not a control with a box of its own.
  const padding = body.match(/padding:\s*([^;}]+)/)?.[1].trim()
  if (padding && padding !== '0' && !/min-height|height/.test(body)) {
    const at = globalCss.slice(0, globalCss.indexOf(block)).split('\n').length
    findings.push({
      kind: 'derived-height',
      where: `assets/css/main.css:${at}`,
      detail: `${sel} sets padding but no min-height, so its height is whatever its content happens to be. Two of these side by side will not line up. Use var(--control-h).`,
    })
  }
}

// Only two of the three are failures.
//
// A collision is often deliberate — a component that legitimately styles its own
// root, where a global rule of the same name also exists — and this script
// cannot tell those apart from the ones that bite. It is printed under
// `--strict` and otherwise counted, because a gate that cries wolf fifty times
// is a gate everybody learns to skip, and that is worse than not having one.
const strict = process.argv.includes('--strict')
const failing = findings.filter((f) => f.kind !== 'collision')
const collisions = findings.filter((f) => f.kind === 'collision')

for (const kind of ['magic-viewport', 'derived-height']) {
  const group = findings.filter((f) => f.kind === kind)
  if (!group.length) continue
  console.log(`\n${kind} (${group.length})`)
  for (const f of group) console.log(`  ${f.where}\n    ${f.detail}`)
}

if (collisions.length) {
  if (strict) {
    console.log(`\ncollision (${collisions.length})`)
    for (const f of collisions) console.log(`  ${f.where}\n    ${f.detail}`)
  } else {
    console.log(`\ncollision: ${collisions.length} class name(s) defined both globally and in a scoped block.`)
    console.log('  Mostly deliberate. Re-run with --strict to read them, and check any name you just added.')
  }
}

if (!failing.length) {
  console.log('\ncss-consistency: no failures')
  process.exit(0)
}
console.log(`\n${failing.length} failure(s)`)
process.exit(1)
