// Browser-level regression tests without an additional test-runner dependency.
// Start Storybook first; requires Chrome and Node 22+ (built-in WebSocket).
// STORYBOOK_URL=http://127.0.0.1:6006 node scripts/test-memory-ui.mjs
import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'

const profile = await mkdtemp(join(tmpdir(), 'research-memory-chrome-'))
const chrome = spawn(process.env.CHROME_BIN || 'google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu', '--disable-dev-shm-usage',
  '--remote-debugging-port=0', `--user-data-dir=${profile}`, 'about:blank',
], { stdio: 'ignore' })
let launchError
chrome.on('error', error => { launchError = error })
let socket
try {
  async function until(fn, label) {
    const deadline = Date.now() + 30000
    while (Date.now() < deadline) {
      if (launchError) throw launchError
      const value = await fn()
      if (value) return value
      await delay(100)
    }
    throw new Error(`Timed out: ${label}`)
  }
  const port = await until(async () => {
    try { return (await readFile(join(profile, 'DevToolsActivePort'), 'utf8')).split('\n')[0] }
    catch { return null }
  }, 'Chrome startup')
  const pages = await (await fetch(`http://127.0.0.1:${port}/json`)).json()
  socket = new WebSocket(pages.find(page => page.type === 'page').webSocketDebuggerUrl)
  await new Promise((resolve, reject) => { socket.addEventListener('open', resolve, { once: true }); socket.addEventListener('error', reject, { once: true }) })
  let serial = 0
  const pending = new Map()
  socket.addEventListener('message', event => {
    const response = JSON.parse(event.data)
    const handler = pending.get(response.id)
    if (!handler) return
    pending.delete(response.id)
    clearTimeout(handler.timeout)
    if (response.error) handler.reject(new Error(JSON.stringify(response.error)))
    else handler.resolve(response.result)
  })
  function call(method, params = {}) {
    return new Promise((resolve, reject) => {
      const id = ++serial
      const timeout = setTimeout(() => { pending.delete(id); reject(new Error(`CDP timeout: ${method}`)) }, 15000)
      pending.set(id, { resolve, reject, timeout })
      socket.send(JSON.stringify({ id, method, params }))
    })
  }
  async function evaluate(expression) {
    const result = await call('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true })
    if (result.exceptionDetails) throw new Error(result.result?.description || result.exceptionDetails.text)
    return result.result.value
  }
  async function story(name) {
    await call('Page.navigate', { url: `${process.env.STORYBOOK_URL || 'http://127.0.0.1:6006'}/iframe.html?id=research-settings-memorylist--${name}&viewMode=story` })
    await until(() => evaluate('!!document.querySelector(".memory-list")'), `story ${name}`)
  }
  async function click(label, scope = '.memory-list') {
    await evaluate(`(() => { const button = [...document.querySelectorAll(${JSON.stringify(scope + ' button')})].find(b => b.textContent.trim() === ${JSON.stringify(label)}); if (!button || button.disabled) throw Error('Button unavailable: ' + ${JSON.stringify(label)}); button.focus(); button.click(); })()`)
    await delay(100)
  }
  async function fill(selector, value) {
    await evaluate(`(() => { const input = document.querySelector(${JSON.stringify(selector)}); if (!input) throw Error('Missing input'); input.value = ${JSON.stringify(value)}; input.dispatchEvent(new Event('input', { bubbles: true })); })()`)
    await delay(50)
  }
  async function key(key, code, modifiers = 0) {
    const virtualKeyCode = { Enter: 13, Tab: 9, Escape: 27 }[key]
    const event = { key, code, modifiers, windowsVirtualKeyCode: virtualKeyCode, nativeVirtualKeyCode: virtualKeyCode }
    await call('Input.dispatchKeyEvent', { type: key === 'Enter' ? 'keyDown' : 'rawKeyDown', ...event, ...(key === 'Enter' ? { text: '\r', unmodifiedText: '\r' } : {}) })
    await call('Input.dispatchKeyEvent', { type: 'keyUp', ...event })
    await delay(100)
  }
  await story('editable')
  assert.match(await evaluate('document.querySelector(".memory-list").textContent'), /Unknown author/)
  assert.equal(await evaluate('document.querySelector(".provenance a").getAttribute("href")'), '/research/R1/session/SS1')
  await click('Edit')
  assert.equal(await evaluate('document.activeElement === document.querySelector(".memory-item textarea")'), true, 'Edit focuses the draft')
  await click('Cancel')
  assert.equal(await evaluate('document.activeElement.textContent.trim()'), 'Edit', 'Cancel returns focus to Edit')
  await fill('.memory-list > form textarea', 'Browser-added note')
  await click('Add note')
  assert.equal(await evaluate('document.querySelectorAll(".memory-item").length'), 3)
  assert.equal(await evaluate('document.querySelector(".memory-list > form textarea").value'), '')
  await click('Edit')
  await fill('.memory-item textarea', 'Browser-edited note')
  await click('Save')
  assert.match(await evaluate('document.querySelector(".memory-item").textContent'), /Browser-edited note/)
  await evaluate('document.querySelector(".select-note input").click()')
  await click('Delete selected')
  assert.equal(await evaluate('document.querySelectorAll(".memory-item").length'), 3, 'delete needs confirmation')
  assert.match(await evaluate('document.querySelector("[role=dialog]").textContent'), /Delete memory notes/)
  assert.equal(await evaluate('document.querySelector("[role=dialog]").contains(document.activeElement)'), true, 'confirmation takes keyboard focus')
  await key('Tab', 'Tab', 8)
  assert.equal(await evaluate('document.activeElement.textContent.trim()'), 'Confirm deletion', 'Shift+Tab wraps to confirm')
  await key('Tab', 'Tab')
  assert.equal(await evaluate('document.activeElement.textContent.trim()'), 'Cancel', 'Tab stays inside dialog')
  await key('Escape', 'Escape')
  assert.equal(await evaluate('!!document.querySelector("[role=dialog]")'), false)
  assert.equal(await evaluate('document.activeElement.textContent.trim()'), 'Delete selected', 'Escape restores trigger focus')
  assert.equal(await evaluate('document.querySelectorAll(".memory-item").length'), 3, 'Escape does not delete')
  await key('Enter', 'Enter')
  await until(() => evaluate('!!document.querySelector("[role=dialog]")'), 'keyboard reopens deletion dialog')
  await click('Confirm deletion', '[role=dialog]')
  assert.equal(await evaluate('document.querySelectorAll(".memory-item").length'), 2)
  assert.match(await evaluate('document.querySelector(".memory-list").textContent'), /Browser-added note/, 'selected delete preserved a new note')
  await story('read-only')
  assert.equal(await evaluate('document.querySelectorAll(".memory-list textarea, .memory-list input").length'), 0)
  assert.deepEqual(await evaluate('[...document.querySelectorAll(".memory-list button")].map(b => b.textContent.trim())'), ['Refresh'])
  await story('conflict')
  await click('Edit')
  await fill('.memory-item textarea', 'Keep this draft after conflict')
  await click('Save')
  assert.match(await evaluate('document.querySelector("[role=alert]").textContent'), /changed/)
  assert.equal(await evaluate('document.querySelector(".memory-item textarea").value'), 'Keep this draft after conflict')
  await click('Load latest version')
  assert.equal(await evaluate('document.querySelector(".memory-item textarea").value'), 'Keep this draft after conflict', 'loading latest must not replace draft')
  assert.match(await evaluate('document.querySelector(".saved-version").textContent'), /Current saved text.*A colleague saved/s)
  await click('Save')
  assert.equal(await evaluate('document.querySelector(".memory-item .note-text").textContent'), 'Keep this draft after conflict')
  assert.equal(await evaluate('!!document.querySelector(".memory-item textarea")'), false, 'save after reload succeeds')
  assert.equal(await evaluate('document.activeElement.textContent.trim()'), 'Edit', 'successful save restores editor trigger focus')
  await click('Refresh')
  assert.equal(await evaluate('document.querySelector(".memory-item .note-text").textContent'), 'Keep this draft after conflict', 'draft was persisted with the latest version')
  await story('reload-deleted')
  await click('Edit')
  await fill('.memory-item textarea', 'Recover this deleted-note draft')
  await click('Refresh')
  assert.equal(await evaluate('document.querySelector(".memory-item textarea").value'), 'Recover this deleted-note draft')
  assert.equal(await evaluate('[...document.querySelectorAll(".memory-item button")].find(b => b.textContent.trim() === "Save").disabled'), true)
  assert.match(await evaluate('document.querySelector(".memory-list").textContent'), /note was deleted/)
  await click('Cancel')
  assert.equal(await evaluate('document.activeElement === document.querySelector(".memory-list > form textarea")'), true, 'deleted editor returns focus to new note')
  await story('error')
  await until(() => evaluate('!!document.querySelector("[role=alert]")'), 'save error story')
  assert.equal(await evaluate('document.querySelector(".memory-list > form textarea").value'), 'Keep this note while the connection recovers.')
  await story('pending')
  await until(() => evaluate('document.querySelector(".memory-list").getAttribute("aria-busy") === "true"'), 'pending story')
  assert.equal(await evaluate('[...document.querySelectorAll(".memory-list button, .memory-list textarea, .memory-list input")].every(control => control.disabled)'), true)
  await story('long-multilingual')
  await call('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true })
  assert.equal(await evaluate('document.documentElement.scrollWidth <= window.innerWidth'), true, 'mobile layout overflows')
  console.log('Memory UI: add/edit, keyboard focus, confirmed selected deletion, viewer access, conflict recovery, deleted drafts, errors, pending state and multilingual mobile layout passed.')
} finally {
  socket?.close()
  if (chrome.exitCode === null && chrome.signalCode === null && !launchError) {
    const exited = new Promise(resolve => chrome.once('exit', resolve))
    chrome.kill('SIGTERM')
    await exited
  }
  // This directory was created by this test, never supplied by the caller.
  await rm(profile, { recursive: true, force: true, maxRetries: 10, retryDelay: 200 })
}
