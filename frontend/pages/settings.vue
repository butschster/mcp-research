<script setup lang="ts">
const { user, isAuthenticated, authEnabled } = useAuth()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || ''

// Redirect if not authenticated
if (authEnabled.value && !isAuthenticated.value) {
  navigateTo('/login')
}

// API Keys
interface APIKey {
  id: string
  name: string
  key_prefix: string
  last_used_at?: string
  created_at: string
}

const apiKeys = ref<APIKey[]>([])
const newKeyName = ref('')
const newKeyValue = ref('')
const keyError = ref('')

async function loadKeys() {
  const { token } = useAuth()
  try {
    const res = await $fetch<APIKey[]>(`${baseURL}/api/auth/api-keys`, {
      headers: { Authorization: `Bearer ${token.value}` },
    })
    apiKeys.value = res ?? []
  } catch { /* ignore */ }
}

async function createKey() {
  const { token } = useAuth()
  keyError.value = ''
  newKeyValue.value = ''
  try {
    const res = await $fetch<{ key: string; id: string; name: string; prefix: string }>(`${baseURL}/api/auth/api-keys`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: { name: newKeyName.value || 'Untitled' },
    })
    newKeyValue.value = res.key
    newKeyName.value = ''
    await loadKeys()
  } catch (e: any) {
    keyError.value = e?.data?.error || 'Failed to create key'
  }
}

async function deleteKey(id: string) {
  const { token } = useAuth()
  try {
    await $fetch(`${baseURL}/api/auth/api-keys/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token.value}` },
    })
    await loadKeys()
  } catch { /* ignore */ }
}

onMounted(() => {
  loadKeys()
})
</script>

<template>
  <div class="settings-page">
    <h1 class="page-title">Settings</h1>

    <div v-if="user" class="settings-section">
      <h2>Account</h2>
      <p class="card-meta">{{ user.email }}</p>
      <p v-if="user.name">{{ user.name }}</p>
    </div>

    <div class="settings-section">
      <h2>API Keys</h2>
      <p class="card-meta">Use API keys to authenticate MCP SSE connections and REST API requests.</p>

      <div v-if="newKeyValue" class="key-created">
        <strong>New API key created. Copy it now — it won't be shown again:</strong>
        <code class="key-value">{{ newKeyValue }}</code>
      </div>

      <div v-if="keyError" class="auth-error">{{ keyError }}</div>

      <form @submit.prevent="createKey" class="key-form">
        <input v-model="newKeyName" type="text" placeholder="Key name (optional)" class="auth-input" />
        <button type="submit" class="auth-button">Create key</button>
      </form>

      <table v-if="apiKeys.length" class="keys-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Key</th>
            <th>Last used</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="key in apiKeys" :key="key.id">
            <td>{{ key.name || '—' }}</td>
            <td><code>{{ key.key_prefix }}</code></td>
            <td class="card-meta">{{ key.last_used_at || 'Never' }}</td>
            <td><button class="delete-btn" @click="deleteKey(key.id)">Revoke</button></td>
          </tr>
        </tbody>
      </table>
      <p v-else class="card-meta">No API keys yet.</p>
    </div>

  </div>
</template>

<style scoped>
.settings-page { max-width: 700px; }
.page-title { font-size: var(--text-2xl); font-weight: 600; margin-bottom: var(--space-8); }
.settings-section {
  margin-bottom: var(--space-8);
  padding: var(--space-6);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
}
.settings-section h2 { font-size: var(--text-lg); font-weight: 600; margin-bottom: var(--space-2); }
.key-form { display: flex; gap: var(--space-2); margin: var(--space-4) 0; flex-wrap: wrap; }
.key-form .auth-input { flex: 1; min-width: 200px; }
.key-form .auth-button { white-space: nowrap; }
.key-created {
  padding: var(--space-3);
  background: var(--color-success-bg, #f0fdf4);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
  font-size: var(--text-sm);
}
.key-value {
  display: block;
  margin-top: var(--space-2);
  padding: var(--space-2);
  background: var(--color-bg);
  border-radius: var(--radius-sm);
  word-break: break-all;
  font-size: var(--text-xs);
}
.keys-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: var(--space-4);
  font-size: var(--text-sm);
}
.keys-table th, .keys-table td {
  padding: var(--space-2) var(--space-3);
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}
.keys-table th { font-weight: 500; color: var(--color-text-secondary); }
.delete-btn {
  background: none;
  border: none;
  color: var(--color-danger, #dc2626);
  cursor: pointer;
  font-size: var(--text-sm);
  font-family: inherit;
}
.delete-btn:hover { text-decoration: underline; }
.auth-error {
  padding: var(--space-2) var(--space-3);
  background: var(--color-danger-bg, #fef2f2);
  color: var(--color-danger, #dc2626);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  margin-bottom: var(--space-3);
}
.auth-input {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  background: var(--color-bg);
  color: var(--color-text);
  font-family: inherit;
}
.auth-button {
  padding: var(--space-2) var(--space-4);
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
}

/* Responsive */
@media (max-width: 768px) {
  .settings-section { padding: var(--space-4); }
  .keys-table { display: block; overflow-x: auto; -webkit-overflow-scrolling: touch; }
  .key-form { flex-direction: column; }
  .key-form .auth-input { min-width: 0; }
}
</style>
