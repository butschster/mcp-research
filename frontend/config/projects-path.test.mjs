import { test } from 'node:test'
import assert from 'node:assert/strict'
import { projectsPath } from './projects-path.mjs'

test('standalone default and configurable hosted paths', () => {
  assert.equal(projectsPath(), '/')
  assert.equal(projectsPath('/projects/'), '/projects')
  assert.equal(projectsPath('/workspace/projects'), '/workspace/projects')
})
test('reject malformed, external, and reserved routes', () => {
  for (const path of ['//evil.test', 'https://evil.test', '/projects?x=1', '/:id', '/api/projects', '/research', '/login', '/how-it-works', '/projects/../login']) {
    assert.throws(() => projectsPath(path), /NUXT_PROJECTS_PATH/)
  }
})
