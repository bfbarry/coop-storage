<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { useUser, useAuth } from '@clerk/vue'
import FileGrid from '@/components/FileGrid.vue'
import type { MetaObject } from '@/types/file'

const { user, isLoaded } = useUser()
const { getToken } = useAuth()

const files = ref<MetaObject[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const accountId = ref<number | null>(null)
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const API_BASE = 'http://localhost:7678'

async function fetchFiles() {
  if (accountId.value === null) return
  const res = await fetch(`${API_BASE}/metadata/home?owner_id=${accountId.value}`)
  if (!res.ok) throw new Error(`Failed to fetch files: ${res.status}`)
  files.value = await res.json()
}

watchEffect(async () => {
  if (!isLoaded.value) return
  if (!user.value) {
    loading.value = false
    return
  }

  loading.value = true
  error.value = null

  try {
    // Find or create the DB account for this Clerk user
    let account
    const accountRes = await fetch(`${API_BASE}/accounts?clerk_id=${encodeURIComponent(user.value.id)}`)
    if (accountRes.status === 404) {
      const createRes = await fetch(`${API_BASE}/accounts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          clerk_id: user.value.id,
          email: user.value.primaryEmailAddress?.emailAddress ?? '',
        }),
      })
      if (!createRes.ok) throw new Error('Failed to create account')
      account = await createRes.json()
    } else if (!accountRes.ok) {
      throw new Error(`Failed to look up account: ${accountRes.status}`)
    } else {
      account = await accountRes.json()
    }

    accountId.value = account.id
    await fetchFiles()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
    files.value = []
  } finally {
    loading.value = false
  }
})

async function handleUpload(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || accountId.value === null) return

  uploading.value = true
  error.value = null

  try {
    const token = await getToken.value({ skipCache: true })

    // 1. Get presigned upload URL
    const presignRes = await fetch(`${API_BASE}/upload/presign`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        filename: file.name,
        content_type: file.type || 'application/octet-stream',
        content_length: file.size,
      }),
    })
    if (!presignRes.ok) throw new Error(`Presign failed: ${presignRes.status}`)
    const { upload_url, object_key } = await presignRes.json()

    // 2. PUT directly to RustFS
    const putRes = await fetch(upload_url, {
      method: 'PUT',
      headers: { 'Content-Type': file.type || 'application/octet-stream' },
      body: file,
    })
    if (!putRes.ok) throw new Error(`Upload failed: ${putRes.status}`)

    // 3. Save metadata
    const metaRes = await fetch(`${API_BASE}/metadata`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        owner_id: accountId.value,
        object_key,
        name: file.name,
        file_type: file.type || 'application/octet-stream',
        is_file: true,
        version: 1,
      }),
    })
    if (!metaRes.ok) throw new Error(`Metadata save failed: ${metaRes.status}`)

    await fetchFiles()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Upload failed'
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function handleDownload(file: MetaObject) {
  if (!file.object_key) return
  try {
    const res = await fetch(`${API_BASE}/download/presign/${file.object_key}`)
    if (!res.ok) throw new Error(`Download presign failed: ${res.status}`)
    const { download_url } = await res.json()

    const a = document.createElement('a')
    a.href = download_url
    a.download = file.name
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Download failed'
  }
}
</script>

<template>
  <div class="file-explorer">
    <header class="explorer-header">
      <div class="header-left">
        <h1>My Files</h1>
        <p v-if="user" class="user-email">{{ user.primaryEmailAddress?.emailAddress }}</p>
      </div>
      <div class="header-right">
        <input ref="fileInput" type="file" hidden @change="handleUpload" />
        <button
          class="upload-btn"
          :disabled="uploading || !accountId"
          @click="fileInput?.click()"
        >
          {{ uploading ? 'Uploading…' : '+ Upload' }}
        </button>
      </div>
    </header>

    <div class="explorer-content">
      <div v-if="!isLoaded" class="loading">
        <div class="spinner" />
        <p>Starting up…</p>
      </div>

      <div v-else-if="!user" class="welcome">
        <p>Sign in to view your files</p>
      </div>

      <div v-else-if="error" class="error">
        <p>{{ error }}</p>
      </div>

      <div v-else>
        <div v-if="loading" class="loading">
          <div class="spinner" />
          <p>Loading files…</p>
        </div>
        <template v-else>
          <div class="file-count-bar">
            <span class="file-count">{{ files.length }} item{{ files.length !== 1 ? 's' : '' }}</span>
            <span class="hint">Double-click a file to download</span>
          </div>
          <FileGrid :files="files" @download="handleDownload" />
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.file-explorer {
  min-height: 100vh;
  background: #f5f5f5;
}

.explorer-header {
  background: white;
  padding: 20px 40px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left h1 { margin: 0 0 4px; color: #333; font-size: 28px; }
.user-email { margin: 0; color: #888; font-size: 14px; }

.upload-btn {
  padding: 10px 22px;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.upload-btn:hover:not(:disabled) { background: #43a047; }
.upload-btn:disabled { background: #a5d6a7; cursor: not-allowed; }

.explorer-content { padding: 20px 40px; }

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: #666;
}

.spinner {
  border: 4px solid #f3f3f3;
  border-top: 4px solid #4CAF50;
  border-radius: 50%;
  width: 50px;
  height: 50px;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error {
  text-align: center;
  padding: 40px 20px;
  color: #f44336;
  font-size: 16px;
  background: #ffebee;
  border-radius: 8px;
}

.welcome {
  text-align: center;
  padding: 80px 20px;
  color: #999;
  font-size: 20px;
}

.file-count-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 8px;
  padding: 0 20px;
}

.file-count {
  background: #e3f2fd;
  color: #1976d2;
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 600;
}

.hint { font-size: 13px; color: #aaa; }
</style>
