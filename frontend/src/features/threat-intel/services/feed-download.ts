import { getStoredTokens } from '@/shared/lib/api-client'
import { IS_FEDERATION } from '@/shared/config/mode'
import { getCurrentInstanceId } from '@/shared/lib/current-instance'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'
const BASE = `${API_URL}/threat-intel`

function triggerDownload(filename: string, blob: Blob) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  setTimeout(() => URL.revokeObjectURL(url), 10_000)
}

export function feedFormatLabel(type: string): string {
  return type === 'daily' ? '.ndjson' : '.list.tar.gz'
}

async function authedBlob(path: string): Promise<Blob> {
  const headers: Record<string, string> = {}
  const tokens = getStoredTokens()
  if (tokens?.access_token) headers.Authorization = `Bearer ${tokens.access_token}`
  if (IS_FEDERATION) {
    const id = getCurrentInstanceId()
    if (id != null) headers['X-UTM-Instance'] = String(id)
  }
  const res = await fetch(`${BASE}${path}`, { headers })
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = JSON.parse(await res.text())
      const err = body?.error
      msg = (typeof err === 'string' ? err : err?.message) ?? msg
    } catch {
    }
    throw new Error(msg)
  }
  return res.blob()
}

export function downloadFeed(feed: { accuracy: string; type: string; name: string }) {
  const ext = feedFormatLabel(feed.type)
  return authedBlob(
    `/feeds/download/${encodeURIComponent(feed.accuracy)}/${encodeURIComponent(feed.type)}/${encodeURIComponent(feed.name)}`,
  ).then((blob) => triggerDownload(`${feed.name}${ext}`, blob))
}

export function downloadChecksum() {
  return authedBlob('/feeds/checksum').then((blob) => triggerDownload('threatwinds-checksum.txt', blob))
}
