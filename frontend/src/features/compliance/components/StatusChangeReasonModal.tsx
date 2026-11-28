import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Upload, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { cn } from '@/shared/lib/utils'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import { EvidenceFileChip } from './EvidenceFileChip'

const EVIDENCE_TYPES = ['application/pdf', 'image/png', 'image/jpeg', 'image/gif', 'image/webp']

/**
 * File picker for evidence: drag-and-drop or click, filtered by type. Files are
 * uploaded to the backend immediately; only the returned URLs get submitted
 * with the verdict.
 */
function EvidenceUpload({
  files,
  onUploaded,
  onRemove,
  disabled,
}: {
  files: string[]
  onUploaded: (urls: string[]) => void
  onRemove: (url: string) => void
  disabled?: boolean
}) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [uploading, setUploading] = useState(false)

  const addFiles = async (files: FileList | File[]) => {
    const list = Array.from(files).filter((f) => EVIDENCE_TYPES.includes(f.type))
    if (list.length === 0) {
      toast.error(
        t('compliance.evidence.types', { defaultValue: 'Only PDF, PNG, JPEG, GIF or WebP files are allowed' }),
      )
      return
    }
    setUploading(true)
    try {
      const urls = (await Promise.all(list.map((f) => complianceService.uploadFile(f)))).map(
        (r) => r.url,
      )
      onUploaded(urls)
    } catch (e) {
      toast.error(
        e instanceof ComplianceHttpError
          ? e.message
          : t('compliance.evidence.uploadError', { defaultValue: 'Failed to upload file' }),
      )
    } finally {
      setUploading(false)
    }
  }

  return (
    <>
      <input
        ref={inputRef}
        type="file"
        multiple
        accept={EVIDENCE_TYPES.join(',')}
        onChange={(e) => {
          if (e.target.files?.length) void addFiles(e.target.files)
          e.target.value = ''
        }}
        className="hidden"
      />
      <div
        onDragOver={(e) => {
          e.preventDefault()
          if (!disabled) setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(e) => {
          e.preventDefault()
          setDragging(false)
          if (!disabled && e.dataTransfer.files.length) void addFiles(e.dataTransfer.files)
        }}
        onClick={() => !disabled && !uploading && inputRef.current?.click()}
        className={cn(
          'flex cursor-pointer flex-col gap-2 rounded-md border border-dashed p-3 transition-colors',
          dragging ? 'border-primary bg-primary/5' : 'border-input bg-background/40 hover:border-primary/50',
          (disabled || uploading) && 'pointer-events-none opacity-60',
        )}
      >
        <div className="flex flex-col items-center justify-center gap-1.5 py-1.5 text-center">
          {uploading ? (
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          ) : (
            <Upload size={15} className="text-muted-foreground" />
          )}
          <p className="text-xs text-muted-foreground">
            {t('compliance.evidence.drop', { defaultValue: 'Drop files here or click to browse' })}
          </p>
          <p className="text-[10px] text-muted-foreground/70">
            {t('compliance.evidence.formats', { defaultValue: 'PDF, PNG, JPEG, GIF, WebP — max 10MB' })}
          </p>
        </div>
        {files.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {files.map((url) => (
              <div key={url} className="relative">
                <EvidenceFileChip url={url} className="w-48" />
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    onRemove(url)
                  }}
                  disabled={disabled}
                  aria-label={t('compliance.reasonModal.cancel', { defaultValue: 'Cancel' })}
                  className="absolute -right-1.5 -top-1.5 flex h-4 w-4 items-center justify-center rounded-full border border-border bg-popover text-muted-foreground shadow-sm hover:text-foreground disabled:opacity-50"
                >
                  <X size={9} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  )
}

export function StatusChangeReasonModal({
  onCancel,
  onConfirm,
  busy = false,
}: {
  onCancel: () => void
  onConfirm: (reason: string, evidences: string[]) => void
  busy?: boolean
}) {
  const { t } = useTranslation()
  const [reason, setReason] = useState('')
  const [evidences, setEvidences] = useState<string[]>([])
  const canSubmit = !busy && reason.trim().length > 0

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onCancel}
    >
      <div
        className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-5 py-4">
          <h2 className="text-base font-semibold">{t('compliance.reasonModal.title')}</h2>
          <button
            onClick={onCancel}
            aria-label={t('compliance.reasonModal.cancel')}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="space-y-4 px-5 py-4">
          <div>
            <p className="mb-2 text-xs text-muted-foreground">
              {t('compliance.reasonModal.description')}
            </p>
            <label className="mb-1 block text-xs font-medium text-foreground/80">
              {t('compliance.reasonModal.label')}
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={4}
              autoFocus
              placeholder={t('compliance.reasonModal.placeholder')}
              className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-foreground/80">
              {t('compliance.evidence.title', { defaultValue: 'Evidence (optional)' })}
            </label>
            <EvidenceUpload
              files={evidences}
              onUploaded={(urls) => setEvidences((prev) => [...prev, ...urls])}
              onRemove={(url) => setEvidences((prev) => prev.filter((u) => u !== url))}
              disabled={busy}
            />
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/20 px-5 py-3">
          <Button variant="outline" size="sm" onClick={onCancel} disabled={busy}>
            {t('compliance.reasonModal.cancel')}
          </Button>
          <Button size="sm" disabled={!canSubmit} onClick={() => onConfirm(reason.trim(), evidences)}>
            {t('compliance.reasonModal.confirm')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
