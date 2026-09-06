// Build-time only: the route and every named link are generated together.
export function projectsPath(value = '/') {
  const input = value.trim()
  if (!input || input.startsWith('//')) throw new Error('NUXT_PROJECTS_PATH must be an absolute path such as /projects')
  const path = input.replace(/\/+$/, '') || '/'
  if (path === '/') return path
  if (!/^\/[a-z][a-z0-9-]*(?:\/[a-z][a-z0-9-]*)*$/.test(path)) {
    throw new Error('NUXT_PROJECTS_PATH must be an absolute path such as /projects')
  }
  const reserved = new Set(['api', 'api-docs', 'research', 'login', 'register', 'settings', 'teams', 'templates', 'invite', 's', 'mcp', 'oauth', '.well-known', '_nuxt', 'brand', 'llms', 'how-it-works', 'self-host', 'website-assets'])
  if (reserved.has(path.split('/')[1])) throw new Error('NUXT_PROJECTS_PATH conflicts with a reserved route')
  return path
}
