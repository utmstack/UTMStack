import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Download, File as FileIcon, Loader2 } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { complianceService } from '../services/compliance-http.service'

const IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp']

type ChipState =
  | { kind: 'loading' }
  | { kind: 'image'; blobUrl: string }
  | { kind: 'file'; blobUrl: string; type: string }
  | { kind: 'error' }

export function EvidenceFileChip({ url, className }: { url: string; className?: string }) {
  const { t } = useTranslation()
  const [state, setState] = useState<ChipState>({ kind: 'loading' })
  const blobUrlRef = useState(() => ({ current: null as string | null }))[0]

  useEffect(() => {
    let cancelled = false
    complianceService
      .fetchFile(url)
      .then((blob) => {
        if (cancelled) return
        if (IMAGE_TYPES.includes(blob.type)) {
          const blobUrl = URL.createObjectURL(blob)
          blobUrlRef.current = blobUrl
          setState({ kind: 'image', blobUrl })
        } else {
          const blobUrl = URL.createObjectURL(blob)
          blobUrlRef.current = blobUrl
          setState({ kind: 'file', blobUrl, type: blob.type })
        }
      })
      .catch(() => !cancelled && setState({ kind: 'error' }))
    return () => {
      cancelled = true
      if (blobUrlRef.current) URL.revokeObjectURL(blobUrlRef.current)
      blobUrlRef.current = null
    }
  }, [url])

  const download = () => {
    const blobUrl = blobUrlRef.current
    if (!blobUrl) return
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = url.split('/').pop() || 'evidence'
    a.click()
  }

  const name = url.split('/').pop()

  if (state.kind === 'image') {
    return (
      <img
        src={state.blobUrl}
        alt={name}
        title={name}
        onClick={download}
        className={cn('h-14 w-14 shrink-0 cursor-pointer rounded-md border border-border bg-muted/30 object-cover', className)}
      />
    )
  }

  return (
    <button
      type="button"
      onClick={download}
      title={name}
      disabled={state.kind === 'loading'}
      className={cn(
        'flex items-center gap-2 rounded-md border border-border bg-background/40 px-2.5 py-1.5 text-left transition-colors hover:border-primary/50',
        state.kind === 'loading' && 'opacity-60',
        className,
      )}
    >
      {state.kind === 'loading' ? (
        <Loader2 size={15} className="shrink-0 animate-spin text-muted-foreground" />
      ) : state.kind === 'error' ? (
        <AlertTriangle size={15} className="shrink-0 text-amber-500" />
      ) : state.kind === 'file' && state.type === 'application/pdf' ? (
        <FileIcon size={15} className="shrink-0 text-red-500" />
      ) : (
        <FileIcon size={15} className="shrink-0 text-muted-foreground" />
      )}
      <span className="min-w-0 flex-1 truncate font-mono text-[11px] text-foreground/80">{name}</span>
      <Download size={12} className="shrink-0 text-muted-foreground" />
      {state.kind === 'error' && (
        <span className="shrink-0 text-[10px] text-amber-500">
          {t('compliance.evidence.downloadError', { defaultValue: 'Unreadable' })}
        </span>
      )}
    </button>
  )
}
